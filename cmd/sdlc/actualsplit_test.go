package main

import (
	"bytes"
	"context"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// splitFixture is issue #9 worked 00:00–00:50 in one session, with commits at
// 00:10, 00:15 (M1), 00:40 and 00:50 (M2). rounds builds the boundary gate
// ledger from a minute→RFC3339 clock; it is written under plans, relative to
// the repo.
func splitFixture(t *testing.T, plans string, rounds func(ts func(int) string) []gatestate.Round) string {
	t.Helper()
	// Recent times: CommitWindow ignores commits older than WindowCapDays.
	base := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Hour)
	ts := func(m int) string { return base.Add(time.Duration(m) * time.Minute).Format(time.RFC3339) }
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	if err := os.MkdirAll("workshop/issues", 0o755); err != nil {
		t.Fatal(err)
	}
	issue := "---\nid: 000009\nstatus: working\nstarted: " + ts(0) + "\n---\n\n# x\n"
	if err := os.WriteFile("workshop/issues/000009-x.md", []byte(issue), 0o644); err != nil {
		t.Fatal(err)
	}
	commit := func(iso, msg string) {
		t.Helper()
		if err := os.WriteFile("f", []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", msg}} {
			cmd := exec.Command("git", args...)
			cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+iso, "GIT_COMMITTER_DATE="+iso)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", args, err, out)
			}
		}
	}
	commit(ts(10), "#9 M1: a")
	commit(ts(15), "#9 M1: b")
	commit(ts(40), "#9 M2: c")
	commit(ts(50), "#9 M2: d")

	l := gatestate.Ledger{Gate: "boundary-review", IssueNum: 9, IDPrefix: "BR", Rounds: rounds(ts)}
	if err := writeBoundaryGateLedger(filepath.Join(dir, plans), "000009-x.md", l, dir); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	top := strings.TrimSpace(testfix.Git(t, dir, "rev-parse", "--show-toplevel"))
	tdir := filepath.Join(home, ".claude", "projects", strings.NewReplacer("/", "-", ".", "-").Replace(top))
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	var lines []string
	for m := 0; m <= 50; m += 5 {
		lines = append(lines, `{"timestamp":"`+ts(m)+`","type":"user","message":{"content":"x"}}`)
	}
	if err := os.WriteFile(filepath.Join(tdir, "s.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return top
}

// m1PassedFirstRound: M1 passed its only review round (dispatched 00:20,
// finalized 00:25) after its last commit at 00:15.
func m1PassedFirstRound(ts func(int) string) []gatestate.Round {
	return []gatestate.Round{{N: 1, Timestamp: ts(25), Dispatched: ts(20), Agent: "claude", Boundary: "M1",
		Recipe: "milestone-review", Reviewed: strings.Repeat("e", 40)}}
}

// #234: the measured actual is build + converge, and M1's verification after
// its last commit (00:15 → 00:25) is not counted. A milestone's increment is
// its own phases (E1).
func TestComputeActualSplitsAtBoundaries(t *testing.T) {
	top := splitFixture(t, "workshop/plans", m1PassedFirstRound)
	path := filepath.Join(top, "workshop/issues/000009-x.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(raw, []byte("\n## Plan\n- [x] M1 — finalized\n- [ ] M2 — executable\n")...), 0600); err != nil {
		t.Fatal(err)
	}
	stamp := strings.TrimSpace(testfix.Capture(t, top, "show", "-s", "--format=%aI", "HEAD"))
	t.Setenv("GIT_AUTHOR_DATE", stamp)
	t.Setenv("GIT_COMMITTER_DATE", stamp)
	testfix.Git(t, top, "add", ".")
	testfix.Git(t, top, "commit", "-qm", "#9: declare executable M2")
	hours := func(m float64) float64 { return m / 60 }
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-6 }
	for _, c := range []struct {
		milestone       string
		build, converge float64
	}{
		{"", hours(15 + 25), 0},
		{"M1", hours(15), 0},
		{"M2", hours(25), 0},
	} {
		res := computeActualFor(context.Background(), top, "", "9", c.milestone, "")
		if res.Status != actualMeasured || res.Split == nil {
			t.Fatalf("%q: status %v split %+v detail %s", c.milestone, res.Status, res.Split, res.Detail)
		}
		if !near(res.Split.Build, c.build) || !near(res.Split.Converge, c.converge) || !near(res.Hours, c.build+c.converge) {
			t.Errorf("%q: hours %v split %+v, want build %v converge %v", c.milestone, res.Hours, res.Split, c.build, c.converge)
		}
		if res.Split.Source != "dispatch" {
			t.Errorf("%q: source %q, want dispatch", c.milestone, res.Split.Source)
		}
	}
}

// A review sent back with REWORK at 00:20, then fixed by the commits at 00:40
// and 00:50: the boundary is still open, so converge runs from the first
// dispatch to the window end.
func TestComputeActualCountsConvergeAfterRework(t *testing.T) {
	top := splitFixture(t, "workshop/plans", func(ts func(int) string) []gatestate.Round {
		return []gatestate.Round{{N: 1, Timestamp: ts(25), Dispatched: ts(20), Agent: "claude", Recipe: "milestone-review"}}
	})
	res := computeActualFor(context.Background(), top, "", "9", "", "")
	if res.Split == nil || math.Abs(res.Split.Build-20.0/60) > 1e-6 || math.Abs(res.Split.Converge-30.0/60) > 1e-6 {
		t.Fatalf("split %+v (%v %s), want build 20m, converge 30m", res.Split, res.Status, res.Detail)
	}
}

// BR-1: the split reads the rounds from the plans dir the close persists them
// to (--plans-dir), not only the default.
func TestComputeActualReadsTheGivenPlansDir(t *testing.T) {
	top := splitFixture(t, "elsewhere/plans", m1PassedFirstRound)
	res := computeActualFor(context.Background(), top, "", "9", "", filepath.Join(top, "elsewhere/plans"))
	if res.Split == nil || res.Split.Source != "dispatch" || math.Abs(res.Hours-40.0/60) > 1e-6 {
		t.Fatalf("given plans dir: hours %v split %+v (%s)", res.Hours, res.Split, res.Detail)
	}
	if res := computeActualFor(context.Background(), top, "", "9", "", ""); res.Split == nil || res.Split.Source != "no-review" {
		t.Fatalf("the default plans dir has no ledger here, want no-review: %+v", res.Split)
	}
}

// #234 PQ-1: the row's round metadata comes from the ledger as persisted at
// row-write time, after this close's own round. A first-round pass measured
// before any dispatch (no-review) is a dispatch-cut row once its round is
// there; an older unstamped round makes the cut a finalize fallback.
func TestRowSplitMetaAndSource(t *testing.T) {
	stamped := gatestate.Round{N: 1, Timestamp: "2026-10-10T17:00:00-07:00", Dispatched: "2026-10-10T16:55:00-07:00", Recipe: "small-diff-review", Reviewed: "abc"}
	unstamped := gatestate.Round{N: 1, Timestamp: "2026-10-09T10:00:00-07:00", Boundary: "M1"}
	seed := gatestate.Round{N: 1, Timestamp: "2026-10-09T10:00:00-07:00", NoCap: true}
	measured := &actualSplit{Build: 1, Source: activetime.SplitNoReview}
	cases := []struct {
		name   string
		rounds []gatestate.Round
		split  *actualSplit
		n      int
		recipe string
		source string
	}{
		{"first-round pass", []gatestate.Round{stamped}, measured, 1, "small-diff-review", "dispatch"},
		{"typed --actual", []gatestate.Round{stamped}, nil, 1, "small-diff-review", "manual"},
		{"no review ran", []gatestate.Round{seed}, measured, 0, "", "no-review"},
		{"an unstamped earlier round", []gatestate.Round{unstamped, stamped}, measured, 2, "milestone-review+small-diff-review", "finalize"},
		// The split cannot use a round whose finalize time does not parse, so
		// the row does not count it either (one rule, cutTimes).
		{"a round for every boundary", []gatestate.Round{{N: 1, Timestamp: "2026-10-10T16:00:00-07:00", Boundary: gatestate.BoundaryAll}, stamped}, measured, 1, "small-diff-review", "dispatch"},
		{"an unparseable round", []gatestate.Round{{N: 1, Timestamp: "yesterday", Dispatched: "2026-10-10T16:55:00-07:00"}, stamped}, measured, 1, "small-diff-review", "dispatch"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			meta := rowSplitMeta(gatestate.Ledger{Rounds: c.rounds})
			if meta.Rounds != c.n || meta.Recipe != c.recipe {
				t.Errorf("meta = %+v, want %d rounds, recipe %q", meta, c.n, c.recipe)
			}
			if got := rowSplitSource(c.split, meta); got != c.source {
				t.Errorf("source = %q, want %q", got, c.source)
			}
		})
	}
}

// #234 E3: the close line for a row without an estimate shows no ratio.
func TestCalibrationLine(t *testing.T) {
	quick := estimate.LedgerRow{Issue: "ariadne#321", Actual: 0.17, BuildHours: 0.17, BoundaryRounds: 1, SplitSource: "dispatch"}
	got := calibrationLine(quick, "trusted-window")
	if strings.Contains(got, "×") || !strings.Contains(got, "no estimate — not a calibration point") || !strings.Contains(got, "build 0.17 / converge 0.00") {
		t.Errorf("quick close line: %q", got)
	}
	full := estimate.LedgerRow{Issue: "ariadne#234", Estimate: 2.7, Actual: 3, SplitSource: "manual"}
	if got := calibrationLine(full, "trusted-window"); !strings.Contains(got, "ratio 0.9×") || strings.Contains(got, "build") {
		t.Errorf("typed-actual line: %q", got)
	}
}

// pair#737: M2's four REWORK rounds remained in history after its executable
// Plan row was retired. M3 must measure before its first review is dispatched.
func TestComputeActualRetiredBoundaryBeforeReplacementReview(t *testing.T) {
	top := splitFixture(t, "workshop/plans", func(ts func(int) string) []gatestate.Round {
		rounds := m1PassedFirstRound(ts)
		for i := 0; i < 4; i++ {
			rounds = append(rounds, gatestate.Round{N: i + 2, Boundary: "M2", Dispatched: ts(27 + i*3), Timestamp: ts(29 + i*3), Agent: "claude", Recipe: "milestone-review"})
		}
		return rounds
	})
	path := filepath.Join(top, "workshop/issues/000009-x.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("\n## Plan\n\n- [x] M1 — closed\n- M2 — superseded; history retained\n- [ ] M3 — replacement\n")...)
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	stamp := strings.TrimSpace(testfix.Capture(t, top, "show", "-s", "--format=%aI", "HEAD"))
	t.Setenv("GIT_AUTHOR_DATE", stamp)
	t.Setenv("GIT_COMMITTER_DATE", stamp)
	testfix.Git(t, top, "add", ".")
	testfix.Git(t, top, "commit", "-qm", "#9: retire M2; implement M3")
	got := computeActualFor(context.Background(), top, "", "9", "M3", "")
	if got.Status != actualMeasured || math.Abs(got.Hours-25.0/60) > 1e-6 {
		t.Fatalf("retired M2 swallowed M3: status=%v hours=%v split=%+v detail=%s", got.Status, got.Hours, got.Split, got.Detail)
	}
}

func TestActualRetiredBoundaryDiagnostic(t *testing.T) {
	var out bytes.Buffer
	printActualMeasurement(&out, actualResult{Issue: "9", Window: "2026-01-01T00:00:00Z → 2026-01-01T01:00:00Z", Hours: 0})
	if !strings.Contains(out.String(), "selected phase evidence unavailable") {
		t.Fatalf("silent zero diagnostic: %s", out.String())
	}
}

func TestComputeActualRequestedMilestoneRequiresPlan(t *testing.T) {
	for _, tc := range []struct {
		name   string
		rounds func(func(int) string) []gatestate.Round
	}{
		{"empty", func(func(int) string) []gatestate.Round { return nil }},
		{"finalized only", m1PassedFirstRound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			top := splitFixture(t, "workshop/plans", tc.rounds)
			res := computeActualFor(context.Background(), top, "", "9", "M2", "")
			if res.Status != actualError || !strings.Contains(res.Detail, "Plan unavailable") {
				t.Fatalf("unauthorized M2 measured: status=%v hours=%v detail=%s", res.Status, res.Hours, res.Detail)
			}
		})
	}
}
