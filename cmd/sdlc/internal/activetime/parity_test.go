package activetime

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// gitInit builds a throwaway repo with deterministic, offset-explicit commit
// dates (UTC) so segment boundaries align with the UTC transcript timestamps.
func gitInit(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	return testfix.Repo(t)
}

func gitCommit(t *testing.T, repo, isoDate, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(repo, "f"), []byte(isoDate), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "f"}, {"commit", "-q", "-m", msg}} {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_DATE="+isoDate, "GIT_COMMITTER_DATE="+isoDate)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

// --- #68 loud-fail guards (ported from test_active_time_v3.py) against real git ---

func TestComputeGuardCommitsButZeroEvents(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-03-01T12:00:00+00:00", "#1 did the work")
	res, err := Compute(Options{
		Dirs: []string{t.TempDir()}, GitRepo: repo, // empty transcript dir
		SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-06-01T00:00:00Z",
		Issues: []string{"1"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != TelemetryGap {
		t.Fatalf("commits-but-0-events must be TelemetryGap, got %v", res.Status)
	}
}

func TestComputeGuardEmptyWindow(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-03-01T12:00:00+00:00", "#1 work")
	// Query a window that excludes the commit → 0 commits, 0 events.
	res, err := Compute(Options{
		Dirs: []string{t.TempDir()}, GitRepo: repo,
		SinceISO: "2024-01-01T00:00:00Z", UntilISO: "2024-02-01T00:00:00Z",
		Issues: []string{"1"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != EmptyWindow {
		t.Fatalf("out-of-window commit must be EmptyWindow, got %v", res.Status)
	}
}

// --- golden attribution regression (frozen from the Python parity gate) ---

// TestAttributionGolden runs Go Compute over a crafted fixture spanning a prefix
// run, multi-issue commit-boundary runs, and suffix work, and asserts the
// per-issue minutes against the current global-boundary attribution model.
//
// This test used to freeze the Python segment-anchored oracle from #110. #92
// intentionally replaces that model with source-scoped activity runs claimed by
// nearby issue commits. Derivation (commit-weight 1.0, threshold 15,
// prefix-weight = commit-weight):
//
//	10:00→10:15 active 15 → #8
//	10:15→10:30 active 15 → #8
//	10:45→11:00 active 15 → #8/#10 split
//	11:00→11:15 active 15 → #8/#10 split
//	11:35→11:45 active 10 → #8/#10 split via previous commit fallback
//	totals: #8 = 50min, #10 = 20min
func TestAttributionGolden(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-03-01T10:30:00+00:00", "#8 first bit")
	gitCommit(t, repo, "2026-03-01T11:30:00+00:00", "#10 second (#8)")

	tdir := t.TempDir()
	writeJSONL(t, filepath.Join(tdir, "s.jsonl"), []string{
		`{"timestamp":"2026-03-01T10:00:00Z","type":"user","message":{"content":"#8 planning"}}`,
		`{"timestamp":"2026-03-01T10:15:00Z","type":"assistant","message":{"content":[{"type":"text","text":"on #8"}]}}`,
		`{"timestamp":"2026-03-01T10:45:00Z","type":"user","message":{"content":"#10 mid"}}`,
		`{"timestamp":"2026-03-01T11:00:00Z","type":"user","message":{"content":"#8 more"}}`,
		`{"timestamp":"2026-03-01T11:35:00Z","type":"user","message":{"content":"#10 wrap"}}`,
		`{"timestamp":"2026-03-01T11:45:00Z","type":"user","message":{"content":"#8 done"}}`,
	})

	res, err := Compute(Options{
		Dirs: []string{tdir}, GitRepo: repo,
		SinceISO: "2026-03-01T00:00:00Z", UntilISO: "2026-03-02T00:00:00Z",
		Issues: []string{"8", "10"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Measured {
		t.Fatalf("want Measured, got %v", res.Status)
	}
	golden := map[string]float64{"8": 50, "10": 20}
	for iss, want := range golden {
		if !approx(res.PerIssue[iss], want) {
			t.Errorf("#%s = %.4f min, want golden %.1f", iss, res.PerIssue[iss], want)
		}
	}
	if !approx(res.TotalActive, 70) {
		t.Errorf("total active = %.4f, want 70 min", res.TotalActive)
	}
}

// TestComputeFillsAgentSpanEndToEnd drives a >15-min Agent dispatch→return span
// through the full Compute path (real git + transcript) and asserts the result
// reflects the full span — the headline #118 behavior ("measures ship wall-clock,
// not ~15 min"). This is the integration-seam coverage the pure unit tests can't
// give: it proves the event.go→compute.go→buildSegments span wiring, not just the
// math. Without span-fill the 30-min subagent run between the dispatch event and
// the (dropped) tool_result return contributes nothing to its segment — the old
// engine would report ~5 min here, not 35.
func TestComputeFillsAgentSpanEndToEnd(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-03-01T10:40:00+00:00", "#8 done")

	tdir := t.TempDir()
	writeJSONL(t, filepath.Join(tdir, "s.jsonl"), []string{
		`{"timestamp":"2026-03-01T10:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		// Agent dispatch (assistant tool_use) — an event AND the span start.
		`{"timestamp":"2026-03-01T10:05:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"A1","name":"Agent","input":{}}]}}`,
		// return 30 min later — dropped as an event, but closes the span (a >15-min
		// subagent run the old engine truncates/misses).
		`{"timestamp":"2026-03-01T10:35:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"A1","content":"done"}]}}`,
		`{"timestamp":"2026-03-01T10:45:00Z","type":"user","message":{"content":"#8 done"}}`,
	})

	res, err := Compute(Options{
		Dirs: []string{tdir}, GitRepo: repo,
		SinceISO: "2026-03-01T00:00:00Z", UntilISO: "2026-03-02T00:00:00Z",
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Measured {
		t.Fatalf("want Measured, got %v", res.Status)
	}
	// [10:00,10:40) segment: 5-min gap (10:00→10:05) ∪ the 30-min span [10:05,10:35]
	// = 35 min, anchored by the #8 commit at segEnd. The bare-gap engine would give
	// ~5 (no second event in-segment), so 35 proves the span is filled in full.
	if !approx(res.TotalActive, 35) {
		t.Fatalf("want TotalActive 35 (span filled), got %.4f (un-filled ⇒ ~5)", res.TotalActive)
	}
	if !approx(res.PerIssue["8"], 35) {
		t.Fatalf("want 35 min → #8 (commit-anchored), got %.4f", res.PerIssue["8"])
	}
}
