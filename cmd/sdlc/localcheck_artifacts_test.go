package main

import (
	"context"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunnerArtifactsRetainScratchAndLogUntilReconciled(t *testing.T) {
	privateScratch := t.TempDir()
	privateTemp := filepath.Join(privateScratch, "tmp")
	t.Setenv("WF_TEST_SCRATCH", privateScratch)
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	d, sha, _, _ := localRunRepo(t, "unused")
	d.normalTempDir = func() (string, error) { return privateTemp, nil }
	d.runnerCleanup = func(*exec.Cmd) func() error {
		return func() error { return &runnerCleanupUnconfirmed{Detail: "ending unknown"} }
	}
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "vm", Resource: "runner", Command: []string{"sh", "-c", `echo "$TMPDIR"; echo evidence > "$TMPDIR/input"`}, TimeoutMinutes: 30}, sha)
	if err == nil {
		t.Fatal("unknown ending passed")
	}
	raw, err := os.ReadFile(rec.LogPath)
	if err != nil {
		t.Fatal(err)
	}
	scratch := strings.SplitN(string(raw), "\n", 2)[0]
	relative, relErr := filepath.Rel(privateTemp, scratch)
	if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		t.Fatalf("fixture scratch escaped private root: %q", scratch)
	}
	t.Cleanup(func() { _ = os.RemoveAll(scratch) })
	if _, err := os.Stat(filepath.Join(scratch, "input")); err != nil {
		t.Fatalf("scratch deleted on unknown return: %v", err)
	}
	trees, err := d.Git("worktree", "list", "--porcelain")
	if err != nil {
		t.Fatal(err)
	}
	wt := ""
	for _, line := range strings.Split(trees, "\n") {
		if strings.HasPrefix(line, "worktree ") && strings.TrimPrefix(line, "worktree ") != d.Root {
			wt = strings.TrimPrefix(line, "worktree ")
		}
	}
	if wt == "" {
		t.Fatal("unknown checkout absent")
	}
	t.Cleanup(func() { _, _ = d.Git("worktree", "unlock", wt); _, _ = d.Git("worktree", "remove", "--force", wt) })
	logDir := filepath.Dir(rec.LogPath)
	root := filepath.Dir(filepath.Dir(logDir))
	// Simulate dead original owner and aged artifacts. No PID liveness proof may
	// discard descendants' inputs while the shared authority remains locked.
	if err := os.WriteFile(filepath.Join(logDir, "owner"), []byte("999999999"), 0600); err != nil {
		t.Fatal(err)
	}
	tempAttempt := filepath.Join(root, "attempts", ".attempt-"+filepath.Base(logDir)+"-fixture")
	if err := os.MkdirAll(filepath.Dir(tempAttempt), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tempAttempt, []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	for _, p := range []string{scratch, logDir, wt, tempAttempt} {
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}
	var swept strings.Builder
	sweepStaleTemp(privateTemp, true, time.Now(), func(int) bool { return false }, &swept)
	if !strings.Contains(swept.String(), "[scratch sweep] kept "+scratch) {
		t.Fatalf("a protected run folder was not logged as kept:\n%s", swept.String())
	}
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	sweepStaleLocalWorktrees(d)
	for _, p := range []string{filepath.Join(scratch, "input"), rec.LogPath, wt, tempAttempt} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("collector deleted unknown artifact %s: %v", p, err)
		}
	}
	if _, err := d.Git("worktree", "unlock", wt); err != nil {
		t.Fatal(err)
	}
	sweepStaleTemp(privateTemp, true, time.Now(), func(int) bool { return false }, io.Discard)
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tempAttempt); !os.IsNotExist(err) {
		t.Fatalf("reconciled temporary attempt retained: %v", err)
	}
	if _, err := os.Stat(scratch); !os.IsNotExist(err) {
		t.Fatalf("reconciled scratch retained: %v", err)
	}
	if _, err := os.Stat(rec.LogPath); !os.IsNotExist(err) {
		t.Fatalf("reconciled log retained: %v", err)
	}
}

func TestRunnerRefusesUnavailableAttemptWriter(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "unused")
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	store := suiteStoreRoot(logDir)
	if err := os.MkdirAll(store, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "inflight"), []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "launched")
	_, err = runLocalCheck(context.Background(), d, localCheck{Name: "vm", Resource: "runner", Command: []string{"sh", "-c", `echo launched > "$1"`, "fixture", marker}, TimeoutMinutes: 30}, sha)
	if err == nil || !strings.Contains(err.Error(), "attempt") {
		t.Fatalf("writer unavailable accepted: %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("runner launched without writer: %v", err)
	}
}

func TestRunnerArtifactOwnersIndependent(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "unused")
	var dirs, trees []string
	for i := 0; i < 2; i++ {
		wt, cleanup, err := detachedWorktree(d.Git, sha)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = d.Git("worktree", "unlock", wt); cleanup() })
		if _, err := d.Git("worktree", "lock", wt); err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		if err := attachRunnerArtifactOwner(dir, wt); err != nil {
			t.Fatal(err)
		}
		dirs = append(dirs, dir)
		trees = append(trees, wt)
	}
	if _, err := d.Git("worktree", "unlock", trees[0]); err != nil {
		t.Fatal(err)
	}
	if runnerArtifactProtected(dirs[0]) || !runnerArtifactProtected(dirs[1]) {
		t.Fatal("one unlock affected the other owner")
	}
	for _, raw := range []string{"", "bad", "gitdir: relative", "gitdir: " + strings.Repeat("x", 5000)} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, runnerArtifactOwnerFile), []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if !runnerArtifactProtected(dir) {
			t.Fatalf("malformed ownership proved absence: %q", raw)
		}
	}
	for _, n := range []int{8, 16} {
		calls := 0
		read := func(p string) ([]byte, error) { calls++; return os.ReadFile(p) }
		for i := 0; i < n; i++ {
			if !runnerArtifactProtectedWithRead(dirs[1], read) {
				t.Fatal("held owner lost")
			}
		}
		if calls != n {
			t.Fatalf("%d candidates required%d owner reads", n, calls)
		}
	}
}

func TestRunnerPanicRetainsUnknownArtifacts(t *testing.T) {
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	d, sha, _, _ := localRunRepo(t, "unused")
	scratch := ""
	d.runnerCleanup = func(cmd *exec.Cmd) func() error {
		return func() error {
			for _, e := range cmd.Env {
				if strings.HasPrefix(e, "TMPDIR=") {
					scratch = strings.TrimPrefix(e, "TMPDIR=")
				}
			}
			panic("cleanup ended unexpectedly")
		}
	}
	t.Cleanup(func() {
		if scratch != "" {
			_ = os.RemoveAll(scratch)
		}
	})
	crashed := false
	func() {
		defer func() { crashed = recover() != nil }()
		_, _ = runLocalCheck(context.Background(), d, localCheck{Name: "vm", Resource: "runner", Command: []string{"sh", "-c", `echo evidence > "$TMPDIR/input"`}, TimeoutMinutes: 30}, sha)
	}()
	if !crashed {
		t.Fatal("fixture did not panic")
	}
	trees, _ := d.Git("worktree", "list", "--porcelain")
	for _, line := range strings.Split(trees, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			wt := strings.TrimPrefix(line, "worktree ")
			if wt != d.Root {
				t.Cleanup(func() { _, _ = d.Git("worktree", "unlock", wt); _, _ = d.Git("worktree", "remove", "--force", wt) })
			}
		}
	}
	if _, err := os.Stat(filepath.Join(scratch, "input")); err != nil {
		t.Fatalf("panic erased scratch: %v", err)
	}
	root, _ := queueStateRoot()
	snap, err := (&queuegate.Store{Root: root}).Snapshot(context.Background())
	if err != nil || snap.Complete || len(snap.Active) != 1 || snap.Active[0].State != "unknown" {
		t.Fatalf("panic claimed completion: %+v %v", snap, err)
	}
	logDir, _ := localCheckLogDir(d)
	rows, err := os.ReadDir(filepath.Join(suiteStoreRoot(logDir), "inflight"))
	if err != nil || len(rows) != 1 {
		t.Fatalf("panic lost log ownership: %v %v", rows, err)
	}
}

// Direct private collector conformance also permits safe mutation testing:
// no runtime-selected root or command preparation is reachable from this test.
func TestRunnerCollectorsRespectPrivateAuthority(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "unused")
	wt, cleanup, err := detachedWorktree(d.Git, sha)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = d.Git("worktree", "unlock", wt); cleanup() })
	if _, err := d.Git("worktree", "lock", wt); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	scratch := filepath.Join(root, "sdlc-run-vm-999999999-123")
	if err := os.Mkdir(scratch, 0700); err != nil {
		t.Fatal(err)
	}
	if err := attachRunnerArtifactOwner(scratch, wt); err != nil {
		t.Fatal(err)
	}
	store := t.TempDir()
	writer, err := beginSuiteAttempt(store)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(writer.LogPath, []byte("evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	logDir := filepath.Dir(writer.LogPath)
	if err := attachRunnerArtifactOwner(logDir, wt); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logDir, "owner"), []byte("999999999"), 0600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	for _, p := range []string{scratch, logDir} {
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}
	sweepStaleTemp(root, true, time.Now(), func(int) bool { return false }, io.Discard)
	if err := pruneSuiteAttempts(store, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{scratch, writer.LogPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("unknown artifact retired: %s: %v", p, err)
		}
	}
}
