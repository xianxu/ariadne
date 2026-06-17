package activetime

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

// --- differential parity: Go Compute vs the actual Python script ---

var pyHoursRE = regexp.MustCompile(`(?m)^\s*#(\d+):\s+([0-9]+(?:\.[0-9]+)?)\s+hr\b`)

// TestParityAgainstPython runs the real active-time-v3.py and Go Compute over
// IDENTICAL crafted fixtures (a temp git repo + temp transcript dir spanning a
// prefix segment, a multi-issue commit segment, and a mention-only suffix) and
// asserts the per-issue HOURS match to 2 decimals — the precision sdlc actual
// reports. This is the M1 parity gate in committed, repeatable form. Skips when
// python3 or the script is unavailable.
func TestParityAgainstPython(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	// Locate the script relative to this package: ../../../../construct/local/issues
	script, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "construct", "local", "issues", "active-time-v3.py"))
	if err != nil || !fileExists(script) {
		t.Skipf("active-time-v3.py not found at %s", script)
	}

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

	since, until := "2026-03-01T00:00:00Z", "2026-03-02T00:00:00Z"
	opts := Options{
		Dirs: []string{tdir}, GitRepo: repo, SinceISO: since, UntilISO: until,
		Issues: []string{"8", "10"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	}

	// Go.
	res, err := Compute(opts)
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Measured {
		t.Fatalf("want Measured, got %v", res.Status)
	}

	// Python — identical args.
	cmd := exec.Command("python3", script,
		"--dir", tdir, "--git-repo", repo, "--since", since, "--until", until,
		"--issue", "8", "--issue", "10",
		"--commit-weight", "1.0", "--threshold-min", "15", "--include-assistant")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("python3 active-time-v3.py failed: %v", err)
	}
	pyHours := map[string]float64{}
	for _, m := range pyHoursRE.FindAllStringSubmatch(string(out), -1) {
		h, _ := strconv.ParseFloat(m[2], 64)
		pyHours[m[1]] = h
	}
	if len(pyHours) == 0 {
		t.Fatalf("parsed no per-issue hours from python output:\n%s", out)
	}

	// Compare numeric-issue hours to 2 decimals.
	round2 := func(x float64) float64 { return float64(int64(x*100+0.5)) / 100 }
	for _, iss := range []string{"8", "10"} {
		goH := round2(res.PerIssue[iss] / 60)
		pyH := round2(pyHours[iss])
		if goH != pyH {
			t.Fatalf("parity mismatch #%s: go=%.2fh python=%.2fh\npython output:\n%s",
				iss, goH, pyH, out)
		}
		t.Logf("parity #%s: go=%.2fh python=%.2fh ✓", iss, goH, pyH)
	}
	// No issue printed by python should be missing from Go.
	for iss, pyH := range pyHours {
		if round2(res.PerIssue[iss]/60) != round2(pyH) {
			t.Fatalf("parity mismatch #%s: go=%.2fh python=%.2fh", iss, res.PerIssue[iss]/60, pyH)
		}
	}
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}
