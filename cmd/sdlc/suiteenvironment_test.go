package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const workDirFailure = "go: creating work dir: mkdir /blocked/go-build123: permission denied"

func environmentCheck() localCheck {
	return localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo '" + workDirFailure + "'; exit 1"}, TimeoutMinutes: 1}
}

func TestSuiteEnvironmentFailureDoesNotPoisonRepair(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "skipped")
	rec, err := runLocalCheck(context.Background(), d, environmentCheck(), sha)
	if err == nil || !strings.Contains(err.Error(), "environment") || !strings.Contains(err.Error(), "GOTMPDIR") {
		t.Fatalf("expected environment refusal: rec=%+v err=%v\n%s", rec, err, stderr)
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("environment error published code failure: %v %v", gh.statuses, gh.comments)
	}
	log, readErr := os.ReadFile(rec.LogPath)
	if readErr != nil || !strings.Contains(string(log), workDirFailure) || !strings.HasSuffix(rec.LogPath, ".environment.log") {
		t.Fatalf("environment diagnostic not retained: %s %q %v", rec.LogPath, log, readErr)
	}
	logDir, _ := localCheckLogDir(d)
	rows, warnings, readErr := readSuiteAttempts(suiteStoreRoot(logDir), 100)
	if readErr != nil || len(warnings) != 0 || len(rows) != 0 {
		t.Fatalf("environment became suite evidence: %+v %v %v", rows, warnings, readErr)
	}
	legacy, warnings := readLegacySuiteLogs(filepath.Dir(filepath.Dir(logDir)), d.Repo, nil)
	if len(legacy) != 0 || len(warnings) != 0 {
		t.Fatalf("environment became legacy suite evidence: %+v %v", legacy, warnings)
	}
	c := environmentCheck()
	c.Command = []string{"true"}
	if repaired, err := runLocalCheck(context.Background(), d, c, sha); err != nil || !repaired.passed() || gh.statuses[sha][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("same-head repaired environment blocked: %+v %v\n%s", repaired, err, stderr)
	}
}

func TestSuiteEnvironmentBaselineIsNotRedMain(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "skipped")
	log := filepath.Join(t.TempDir(), "train.log")
	if err := os.WriteFile(log, []byte("--- FAIL: TestActual\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	check := environmentCheck()
	failed := localRunRecord{Exit: 1, LogPath: log}
	err := diagnoseSuiteFailure(context.Background(), &localCheckFailure{Check: check, Run: failed, Head: sha}, sha, d)
	if err == nil || !strings.Contains(err.Error(), "environment") || !strings.Contains(err.Error(), "GOTMPDIR") || strings.Contains(err.Error(), "main is red") {
		t.Fatalf("ordinary baseline misclassified: %v", err)
	}
	c := &conductor{ctx: context.Background(), heads: []string{sha}, base: sha, suite: check, local: d}
	c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return failed, nil }
	if ev, err := c.runSuite(1); err == nil || ev != nil || !strings.Contains(err.Error(), "environment") || !strings.Contains(err.Error(), "GOTMPDIR") {
		t.Fatalf("train baseline blamed code: %+v %v", ev, err)
	}
	if gh.statuses[sha] != nil || len(gh.comments[9]) != 0 {
		t.Fatalf("baseline environment posted status/comment: %v %v", gh.statuses, gh.comments)
	}
}

func TestScratchFallbackOverridesInheritedTemp(t *testing.T) {
	d, sha, _, stderr := localRunRepo(t, "skipped")
	root := t.TempDir()
	blocked := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	// A regular-file path fails even for a privileged test process: no live
	// sandbox permission or chmod assumption is involved.
	t.Setenv("WF_TEST_SCRATCH", blocked)
	t.Setenv("TMPDIR", filepath.Join(blocked, "temp"))
	t.Setenv("GOTMPDIR", filepath.Join(blocked, "go"))
	c := localCheck{Name: "temp-fixture", Command: []string{"sh", "-c", `test -d "$TMPDIR" && test -d "$GOTMPDIR" && printf ok > "$TMPDIR/probe" && printf ok > "$GOTMPDIR/probe"`}, TimeoutMinutes: 1}
	if rec, err := runLocalCheck(context.Background(), d, c, sha); err != nil || !rec.passed() {
		t.Fatalf("fallback inherited invalid temp env: %+v %v\n%s", rec, err, stderr)
	}
	for _, variable := range []string{"GOTMPDIR", "TMPDIR"} {
		if !strings.Contains(stderr.String(), variable) {
			t.Fatalf("replacement note lacks %s:\n%s", variable, stderr)
		}
	}
}
