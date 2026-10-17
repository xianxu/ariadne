package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeSuiteLockHelper advertises the real argv contract without executing the
// acquisition body. Marker/argv effects belong only to an acquisition call.
func fakeSuiteLockHelper(acquireBody string) string {
	return "import sys\nif sys.argv[1:] == ['acquire', '--help']:\n    print('options:\\n  --watch WATCH  owner\\n  --slots SLOTS  weight')\n    sys.exit(0)\n" + acquireBody
}

// localRunRepo is a real git repository with one commit, and a fake suite-lock
// helper beside its Makefile.workflow that answers lockAnswer.
func localRunRepo(t *testing.T, lockAnswer string) (localRunDeps, string, *ciFakeGH, *bytes.Buffer) {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
		return string(out), err
	}
	must := func(args ...string) string {
		t.Helper()
		out, err := git(args...)
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(out)
	}
	must("init", "-q", "-b", "main")
	must("config", "user.email", "t@example.com")
	must("config", "user.name", "Test")
	write := func(rel, body string, mode os.FileMode) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
	}
	write("Makefile.workflow", "# fake\n", 0o644)
	write("scripts/test-lock.py", fakeSuiteLockHelper("print("+fmt.Sprintf("%q", lockAnswer)+")\n"), 0o755)
	write("marker-check.sh", "#!/bin/sh\necho ran > \"$1\"\n", 0o755)
	must("add", "-A")
	must("commit", "-q", "-m", "init")
	sha := must("rev-parse", "HEAD")
	gh := newCIFakeGH()
	gh.setHead(9, sha)
	var stderr bytes.Buffer
	return localRunDeps{GH: gh, Repo: "o/r", Root: root, Git: git, PR: 9, Stderr: &stderr, minute: time.Second}, sha, gh, &stderr
}

func TestRunLocalCheck(t *testing.T) {
	ctx := context.Background()
	check := func(argv ...string) localCheck {
		return localCheck{Name: "zellij", Command: argv, TimeoutMinutes: 30}
	}

	t.Run("pass is recorded on the sha", func(t *testing.T) {
		d, sha, gh, stderr := localRunRepo(t, "acquired")
		rec, err := runLocalCheck(ctx, d, check("sh", "-c", "echo ok; test -f marker-check.sh"), sha)
		if err != nil || !rec.passed() {
			t.Fatalf("%+v %v\n%s", rec, err, stderr)
		}
		if gh.statuses[sha]["sdlc-local/zellij"] != "success" {
			t.Fatalf("statuses %v", gh.statuses)
		}
		if b, _ := os.ReadFile(rec.LogPath); !strings.Contains(string(b), "ok") {
			t.Fatalf("log %q", b)
		}
		if !strings.Contains(stderr.String(), "sh -c echo ok") || !strings.Contains(stderr.String(), rec.LogPath) {
			t.Fatalf("command and log path not announced:\n%s", stderr)
		}
		if out, _ := d.Git("worktree", "list"); strings.Count(out, "\n") != 1 {
			t.Fatalf("worktree left behind:\n%s", out)
		}
		if len(gh.comments[9]) != 0 {
			t.Fatal("a passing run commented")
		}
	})

	t.Run("failure is recorded with a bounded comment", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		rec, err := runLocalCheck(ctx, d, check("sh", "-c", "i=0; while [ $i -lt 500 ]; do echo line $i; i=$((i+1)); done; exit 3"), sha)
		if err != nil || rec.passed() || rec.Exit != 3 {
			t.Fatalf("%+v %v", rec, err)
		}
		if gh.statuses[sha]["sdlc-local/zellij"] != "failure" || len(gh.comments[9]) != 1 {
			t.Fatalf("statuses %v comments %d", gh.statuses, len(gh.comments[9]))
		}
		c := gh.comments[9][0]
		if len(c) > failureTailBytes+1024 || !strings.Contains(c, "line 499") || strings.Contains(c, "line 100\n") {
			t.Fatalf("comment not a bounded tail (%d bytes)", len(c))
		}
	})

	t.Run("lock not taken: the check never runs", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "timeout")
		marker := filepath.Join(t.TempDir(), "ran")
		_, err := runLocalCheck(ctx, d, check("./marker-check.sh", marker), sha)
		if !errors.Is(err, errUnboundedSuite) {
			t.Fatalf("err %v", err)
		}
		if _, serr := os.Stat(marker); serr == nil {
			t.Fatal("check ran without the suite lock")
		}
		if len(gh.statuses[sha]) != 0 {
			t.Fatal("status posted for a run that never happened")
		}
	})

	t.Run("no Makefile.workflow refuses", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		if err := os.Remove(filepath.Join(d.Root, "Makefile.workflow")); err != nil {
			t.Fatal(err)
		}
		if _, err := runLocalCheck(ctx, d, check("true"), sha); !errors.Is(err, errUnboundedSuite) {
			t.Fatalf("err %v", err)
		}
	})

	t.Run("timeout is a recorded failure", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "skipped")
		c := check("sleep", "5")
		c.TimeoutMinutes = 1 // one second under the test's minute
		rec, err := runLocalCheck(ctx, d, c, sha)
		if err != nil || !rec.TimedOut || gh.statuses[sha]["sdlc-local/zellij"] != "failure" || !strings.Contains(rec.statusDescription(), "timed out") {
			t.Fatalf("%+v %v", rec, err)
		}
	})

	t.Run("an unrecorded pass is an error", func(t *testing.T) {
		d, sha, gh, _ := localRunRepo(t, "acquired")
		gh.known = map[string]bool{} // GitHub does not know the sha: 422
		if _, err := runLocalCheck(ctx, d, check("true"), sha); err == nil || !strings.Contains(err.Error(), "NOT recorded") {
			t.Fatalf("err %v", err)
		}
	})
}

func TestRunLocalCheckRecordsEveryAttemptBeforeStatusPost(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	check := localCheck{Name: "suite", Command: []string{"sh", "-c", "echo '--- FAIL: TestA (0.01s)'; exit 2"}, TimeoutMinutes: 30}
	first, err := runLocalCheck(context.Background(), d, check, sha)
	if err != nil || first.Exit != 2 {
		t.Fatalf("failure run: %+v %v", first, err)
	}
	failedBytes, err := os.ReadFile(first.LogPath)
	if err != nil || !strings.Contains(string(failedBytes), "TestA") {
		t.Fatalf("failure log: %q %v", failedBytes, err)
	}
	check.Command = []string{"sh", "-c", "echo retry passed"}
	gh.known = map[string]bool{} // status POST fails after the saved run
	second, err := runLocalCheck(context.Background(), d, check, sha)
	if err == nil || !strings.Contains(err.Error(), "NOT recorded") || !second.passed() {
		t.Fatalf("retry: %+v %v", second, err)
	}
	root := filepath.Join(d.Root, ".git", "sdlc", "test-stats")
	rows, warnings, err := readSuiteAttempts(root, 100)
	if err != nil || len(warnings) != 0 || len(rows) != 2 {
		t.Fatalf("attempts: %d %v %v", len(rows), warnings, err)
	}
	var failed, passed bool
	for _, row := range rows {
		if row.SHA != sha || row.Repo != d.Repo || row.LogSHA256 == "" {
			t.Fatalf("incomplete recovery record: %+v", row)
		}
		if row.Exit == 2 {
			failed = true
		}
		if row.Exit == 0 {
			passed = true
		}
	}
	if !failed || !passed {
		t.Fatalf("missing attempt: %+v", rows)
	}
	if b, err := os.ReadFile(first.LogPath); err != nil || string(b) != string(failedBytes) {
		t.Fatalf("retry overwrote failed log: %q %v", b, err)
	}
}

func TestRunLocalCheckStatsInitializationFailureStillRunsAndPostsStatus(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	storeRoot := filepath.Join(filepath.Dir(logDir), "test-stats")
	if err := os.MkdirAll(storeRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(storeRoot, "inflight"), []byte("blocked"), 0o644); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "ran")
	check := localCheck{Name: "suite", Command: []string{"./marker-check.sh", marker}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("check blocked by optional stats: %+v %v\n%s", rec, err, stderr)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("check did not run: %v", err)
	}
	if gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("status not posted: %v", gh.statuses)
	}
	if rec.LogPath != filepath.Join(logDir, sha, "suite.log") {
		t.Fatalf("legacy log fallback path: %s", rec.LogPath)
	}
	if _, err := os.Stat(rec.LogPath); err != nil {
		t.Fatalf("fallback log missing: %v", err)
	}
	if !strings.Contains(stderr.String(), "attempt") {
		t.Fatalf("stats failure not reported: %s", stderr)
	}
}

func TestRunLocalCheckStatsPublicationFailureStillPostsStatus(t *testing.T) {
	d, sha, gh, stderr := localRunRepo(t, "acquired")
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	storeRoot := filepath.Join(filepath.Dir(logDir), "test-stats")
	if err := os.MkdirAll(storeRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(storeRoot, "attempts"), []byte("blocked"), 0o644); err != nil {
		t.Fatal(err)
	}
	check := localCheck{Name: "suite", Command: []string{"sh", "-c", "echo ran"}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, check, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("check blocked by optional stats: %+v %v\n%s", rec, err, stderr)
	}
	if gh.statuses[sha]["sdlc-local/suite"] != "success" {
		t.Fatalf("status not posted: %v", gh.statuses)
	}
	if b, err := os.ReadFile(rec.LogPath); err != nil || !strings.Contains(string(b), "ran") {
		t.Fatalf("run log unavailable: %q %v", b, err)
	}
	if !strings.Contains(stderr.String(), "attempt result was not fully saved") {
		t.Fatalf("stats failure not reported: %s", stderr)
	}
}

func TestPruneLocalCheckLogs(t *testing.T) {
	dir := t.TempDir()
	base := time.Now().Add(-time.Hour)
	for i := 0; i < 25; i++ {
		p := filepath.Join(dir, fmt.Sprintf("sha%02d", i))
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	pruneLocalCheckLogs(dir, 20)
	entries, _ := os.ReadDir(dir)
	if len(entries) != 20 || entries[0].Name() != "sha05" {
		t.Fatalf("kept %d, oldest %s", len(entries), entries[0].Name())
	}
}

func TestHashAndTailKeepsHashForLongOutputLine(t *testing.T) {
	p := filepath.Join(t.TempDir(), "suite.log")
	content := []byte(strings.Repeat("x", 2<<20) + "\nlast\n")
	if err := os.WriteFile(p, content, 0o600); err != nil {
		t.Fatal(err)
	}
	hash, tail, err := hashAndTail(p, 2)
	if err != nil {
		t.Fatal(err)
	}
	if hash != hashBytes(content) || len(tail) != 2 || tail[1] != "last" {
		t.Fatalf("hash/tail mismatch: %s, %d lines", hash, len(tail))
	}
}

func TestSuiteAgentFieldsKeepsKnownAgentWithoutModel(t *testing.T) {
	agent, model, effort := suiteAgentFields("sdlc\n")
	if agent == nil || *agent != "sdlc" || model != nil || effort != nil {
		t.Fatalf("identity: %v %v %v", agent, model, effort)
	}
}

func TestRunLocalCheckMeasuresLockWaitAndExactHeadAgent(t *testing.T) {
	d, _, gh, stderr := localRunRepo(t, "acquired")
	helper := filepath.Join(d.Root, "scripts", "test-lock.py")
	body := fakeSuiteLockHelper("import time\ntime.sleep(0.12)\nprint('acquired')\n")
	if err := os.WriteFile(helper, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := d.Git("add", "scripts/test-lock.py"); err != nil {
		t.Fatalf("add: %v %s", err, out)
	}
	if out, err := d.Git("commit", "-q", "-m", "lock wait", "-m", "Agent: codex/gpt-6-sol/medium"); err != nil {
		t.Fatalf("commit: %v %s", err, out)
	}
	out, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(out)
	gh.setHead(9, sha)
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"true"}, TimeoutMinutes: 30}, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("run: %+v %v", rec, err)
	}
	rows, warnings, err := readSuiteAttempts(filepath.Join(d.Root, ".git", "sdlc", "test-stats"), 10)
	if err != nil || len(warnings) != 0 || len(rows) != 1 {
		t.Fatalf("rows: %+v %v %v; log=%s; stderr=%s", rows, warnings, err, rec.LogPath, stderr.String())
	}
	a := rows[0]
	if a.QueueSeconds == nil || *a.QueueSeconds < 0.1 {
		t.Fatalf("queue wait not measured around lock: %+v", a.QueueSeconds)
	}
	if a.QueuedAt == nil || a.AcquiredAt == nil || !a.AcquiredAt.After(*a.QueuedAt) || a.AcquiredAt.After(a.StartedAt) {
		t.Fatalf("queue endpoints do not precede the run: %+v", a)
	}
	if a.SHA != sha || a.Agent == nil || *a.Agent != "codex" || a.Model == nil || *a.Model != "gpt-6-sol" || a.Effort == nil || *a.Effort != "medium" {
		t.Fatalf("exact-head identity: %+v", a)
	}
}

func TestRunLocalCheckBaselineIsSavedWithoutPostingStatus(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	d.SkipRecord = true
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"true"}, TimeoutMinutes: 30}, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("baseline: %+v %v", rec, err)
	}
	if len(gh.statuses[sha]) != 0 {
		t.Fatalf("baseline posted status: %+v", gh.statuses[sha])
	}
	rows, warnings, err := readSuiteAttempts(filepath.Join(d.Root, ".git", "sdlc", "test-stats"), 10)
	if err != nil || len(warnings) != 0 || len(rows) != 1 || !rows[0].MainBaseline {
		t.Fatalf("saved baseline: %+v %v %v", rows, warnings, err)
	}
}
