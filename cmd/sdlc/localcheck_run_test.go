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
	write("scripts/test-lock.py", "import sys\nif sys.argv[1:] == ['acquire', '--help']:\n    print('options:\\n  --watch WATCH  owner\\n  --slots SLOTS  weight')\nelse:\n    print("+fmt.Sprintf("%q", lockAnswer)+")\n", 0o755)
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
