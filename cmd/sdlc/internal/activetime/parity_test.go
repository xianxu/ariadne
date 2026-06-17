package activetime

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitInit builds a throwaway repo with deterministic, offset-explicit commit
// dates (UTC) so segment boundaries align with the UTC transcript timestamps.
func gitInit(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	run := func(env []string, args ...string) {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(), env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run(nil, "init", "-q")
	run(nil, "config", "user.email", "t@t")
	run(nil, "config", "user.name", "t")
	run(nil, "config", "commit.gpgsign", "false")
	return repo
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
// segment, a multi-issue commit segment, and a mention-only suffix, and asserts
// the per-issue minutes against GOLDEN values.
//
// Those golden values were established by the M1 parity gate: a differential
// test that ran the real active-time-v3.py over THESE EXACT fixtures and
// confirmed Go == Python (#8=0.46h, #10=0.21h; see issue #110 Log). The Python
// oracle is deleted with the script in M2, so this freezes its verdict as a
// permanent regression guard — any future drift in the attribution math fails
// here. Derivation (commit-weight 1.0, threshold 15, prefix-weight = commit-weight):
//
//	prefix  [10:00,10:30) active 15  → #8 +15           (anchored #8)
//	commit  [10:30,11:30) active 15  → #8 +7.5, #10 +7.5 (anchored #10,#8 split)
//	suffix  [11:30,11:46) active 10  → #8 +5,   #10 +5   (no anchor, mention 1:1)
//	totals: #8 = 27.5min, #10 = 12.5min
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
	golden := map[string]float64{"8": 27.5, "10": 12.5}
	for iss, want := range golden {
		if !approx(res.PerIssue[iss], want) {
			t.Errorf("#%s = %.4f min, want golden %.1f (Python-verified)", iss, res.PerIssue[iss], want)
		}
	}
	if !approx(res.TotalActive, 40) {
		t.Errorf("total active = %.4f, want 40 min", res.TotalActive)
	}
}
