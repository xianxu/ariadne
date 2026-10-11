package main

import (
	"context"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// splitFixture is issue #9 worked 00:00–00:50 in one session. M1 passed its
// only review round (dispatched 00:20, finalized 00:25) after its last commit
// at 00:15; M2's work runs on to 00:50.
func splitFixture(t *testing.T) string {
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

	l := gatestate.Ledger{Gate: "boundary-review", IssueNum: 9, IDPrefix: "BR", Rounds: []gatestate.Round{{
		N: 1, Timestamp: ts(25), Dispatched: ts(20),
		Agent: "claude", Boundary: "M1", Recipe: "milestone-review", Reviewed: strings.Repeat("e", 40),
	}}}
	if err := writeBoundaryGateLedger(filepath.Join(dir, "workshop/plans"), "000009-x.md", l, dir); err != nil {
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

// #234: the measured actual is build + converge, and M1's verification after
// its last commit (00:15 → 00:25) is not counted. A milestone's increment is
// its own phases (E1).
func TestComputeActualSplitsAtBoundaries(t *testing.T) {
	top := splitFixture(t)
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
		res := computeActualFor(context.Background(), top, "", "9", c.milestone)
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
