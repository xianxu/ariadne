package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTypedAutomaticRerunRetainsInflightOnUnconfirmedCleanup(t *testing.T) {
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	d, sha, _, stderr := localRunRepo(t, "acquired")
	wt, cleanup, err := detachedWorktree(d.Git, sha)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = d.Git("worktree", "unlock", wt); cleanup() })
	if _, err := d.Git("worktree", "lock", wt); err != nil {
		t.Fatal(err)
	}
	d.FlakeRegistry = gateRegistry(t, gateKnown)
	d.flakeSHA, d.flakeContract = sha, strings.Repeat("a", 64)
	d.StrictGroupCleanup, d.ordinaryTypedOwner = true, true
	d.runnerLifetime = &runnerArtifactLifetime{retirable: true}
	scriptRerun(t, "printf '%s\\n' typed-child-raw-evidence")
	var retainedScratch string
	t.Cleanup(func() {
		if retainedScratch != "" {
			_ = os.RemoveAll(retainedScratch)
		}
	})
	cleanupObserved := false
	d.runnerCleanup = func(cmd *exec.Cmd) func() error {
		realJoin := configureRunnerCleanup(cmd)
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "TMPDIR=") {
				retainedScratch = strings.TrimPrefix(env, "TMPDIR=")
			}
		}
		return func() error {
			if err := realJoin(); err != nil {
				return err
			}
			cleanupObserved = true
			return &runnerCleanupUnconfirmed{Detail: "fixture cannot attest completion"}
		}
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	store := suiteStoreRoot(logDir)
	_, err = knownFlakeStep(context.Background(), d, suiteGateRun{
		check: localCheck{Name: suiteCheckName, TimeoutMinutes: 30}, sha: sha, wt: wt,
		helper: filepath.Join(d.Root, "scripts", "test-lock.py"), logDir: logDir, storeRoot: store,
		failedID: "aabbccdd", failedFacts: SuiteLogFacts{Failed: true, Failures: []string{gateKnown}}, repo: d.Repo,
	})
	var uncertain *runnerCleanupUnconfirmed
	if !errors.As(err, &uncertain) || !cleanupObserved {
		t.Fatalf("expected post-child cleanup uncertainty, got %v; observed=%v\n%s", err, cleanupObserved, stderr)
	}
	if d.runnerLifetime.mayRetire() {
		t.Fatal("unconfirmed child cleanup released the parent's inputs")
	}
	entries, err := os.ReadDir(filepath.Join(store, "inflight"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("typed child raw inflight disappeared: %v; entries=%v", err, entries)
	}
	inflight := filepath.Join(store, "inflight", entries[0].Name())
	raw, err := os.ReadFile(filepath.Join(inflight, "run.log"))
	if err != nil || !strings.Contains(string(raw), "typed-child-raw-evidence") {
		t.Fatalf("raw child evidence lost: %q %v", raw, err)
	}
	if _, err := os.Stat(filepath.Join(inflight, runnerArtifactOwnerFile)); err != nil {
		t.Fatalf("child owner marker missing: %v", err)
	}
	if !runnerArtifactProtected(inflight) {
		t.Fatal("retained child is not protected by the exact parent checkout lock")
	}
	rows, warnings, err := readSuiteAttempts(store, 100)
	if err != nil || len(rows) != 0 || len(warnings) != 0 {
		t.Fatalf("uncertain child published an attempt: rows=%+v warnings=%v err=%v", rows, warnings, err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(inflight, old, old); err != nil {
		t.Fatal(err)
	}
	if err := pruneSuiteAttempts(store, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(inflight, "run.log")); err != nil {
		t.Fatalf("collector erased protected typed child: %v", err)
	}
	// Confirm the protection is the existing worktree lock, not an unrelated
	// lease or a new marker policy introduced specifically for automatic retry.
	if _, err := d.Git("worktree", "unlock", wt); err != nil {
		t.Fatal(err)
	}
	if runnerArtifactProtected(inflight) {
		t.Fatal("child marker did not reference the exact parent checkout")
	}
}
