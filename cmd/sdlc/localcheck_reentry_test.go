package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A repo whose suite runs under the real helper, with a private lock file per
// test. Each run holds the slot only while its suite runs (#383), and the
// suite's nested acquire must pass through as reentrant on that run's holder.
func realLockLocalRunRepo(t *testing.T, recipe string) (localRunDeps, string, *ciFakeGH) {
	t.Helper()
	d, _, gh, _ := localRunRepo(t, "unused")
	helper, err := os.ReadFile(filepath.Join("..", "..", "scripts", "test-lock.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d.Root, "scripts", "test-lock.py"), helper, 0o755); err != nil {
		t.Fatal(err)
	}
	makefile := "test:\n\t@state=$$(python3 scripts/test-lock.py acquire --watch $$$$ --slots 2); test \"$$state\" = reentrant\n" + recipe
	if err := os.WriteFile(filepath.Join(d.Root, "Makefile"), []byte(makefile), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("add", "Makefile", "scripts/test-lock.py"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("commit", "-qm", "real lock fixture"); err != nil {
		t.Fatal(err)
	}
	mainOut, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	main := strings.TrimSpace(mainOut)
	gh.setHead(9, main)
	t.Setenv("WF_TEST_LOCK_FILE", filepath.Join(t.TempDir(), "suite.lock"))
	t.Setenv("WF_TEST_SUITES", "2")
	t.Setenv("WF_TEST_LOCK_TIMEOUT", "1")
	return d, main, gh
}

func TestSuiteBaselineRetainsRealLockOwner(t *testing.T) {
	d, main, gh := realLockLocalRunRepo(t, "\t@echo '--- FAIL: TestB'; if test -f introduced.go; then echo '--- FAIL: TestA'; fi; exit 1\n")
	if _, err := d.Git("switch", "-qc", "feature"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d.Root, "introduced.go"), []byte("package fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("add", "introduced.go"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("commit", "-qm", "introduced test failure"); err != nil {
		t.Fatal(err)
	}
	headOut, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(headOut)
	gh.setHead(9, head)
	check := localCheck{Name: suiteCheckName, Command: []string{"make", "test"}, TimeoutMinutes: 5, Weight: 2}
	merged, err := runLocalCheck(context.Background(), d, check, head)
	if err != nil || merged.passed() {
		t.Fatalf("merged run: record=%+v err=%v", merged, err)
	}
	refusal := diagnoseSuiteFailure(context.Background(), &localCheckFailure{Check: check, Run: merged, Head: head}, main, d)
	if !strings.Contains(refusal.Error(), "introduced: TestA") || !strings.Contains(refusal.Error(), "pre-existing: TestB") {
		t.Fatalf("baseline failed to reenter held lock: %v", refusal)
	}
	if gh.statuses[main] != nil {
		t.Fatalf("baseline posted a status: %v", gh.statuses[main])
	}
}

func TestRepeatedLocalChecksRetainRealLockOwner(t *testing.T) {
	d, main, gh := realLockLocalRunRepo(t, "\t@echo suite passed\n")
	if _, err := d.Git("switch", "-qc", "feature"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Git("commit", "--allow-empty", "-qm", "integration retry"); err != nil {
		t.Fatal(err)
	}
	headOut, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(headOut)
	gh.setHead(9, head)
	// Same check on a later integrated head, then a second named check: each
	// takes the slot afresh, and each suite's nested make sees its own holder.
	for _, step := range []struct{ name, sha string }{{suiteCheckName, main}, {suiteCheckName, head}, {"smoke", head}} {
		check := localCheck{Name: step.name, Command: []string{"make", "test"}, TimeoutMinutes: 5, Weight: 2}
		rec, err := runLocalCheck(context.Background(), d, check, step.sha)
		if err != nil || !rec.passed() {
			t.Fatalf("%s on %s: record=%+v err=%v", step.name, step.sha, rec, err)
		}
	}
}

// outsideAcquire is another suite asking the real helper for both slots: it
// answers acquired, timeout or reentrant. On acquired the slots stay held
// until the returned stop runs.
func outsideAcquire(t *testing.T, helper string) (string, func()) {
	t.Helper()
	owner := exec.Command("sleep", "60")
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	stop := func() { _ = owner.Process.Kill(); _ = owner.Wait() }
	cmd := exec.Command("python3", helper, "acquire", "--watch", fmt.Sprint(owner.Process.Pid), "--slots", "2")
	cmd.Env = withoutEnv(os.Environ(), "WF_TEST_LOCK_HELD_BY")
	out, err := cmd.Output()
	if err != nil {
		stop()
		t.Fatalf("outside acquire: %v", err)
	}
	return strings.TrimSpace(string(out)), stop
}

func withoutEnv(env []string, key string) []string {
	var out []string
	for _, kv := range env {
		if !strings.HasPrefix(kv, key+"=") {
			out = append(out, kv)
		}
	}
	return out
}

// blockingPostGH runs probe inside the status post, where merge's GitHub phase
// begins: everything after the suite runs here or later.
type blockingPostGH struct {
	*ciFakeGH
	probe func()
}

func (g blockingPostGH) PostStatus(ctx context.Context, repo, sha, statusContext, state, desc string) error {
	g.probe()
	return g.ciFakeGH.PostStatus(ctx, repo, sha, statusContext, state, desc)
}

// #383: the slot covers the suite's run and nothing after it. While the suite
// runs another waiter is not admitted; once the run reaches GitHub it is.
func TestSuiteSlotCoversOnlyTheRun(t *testing.T) {
	during := filepath.Join(t.TempDir(), "during")
	t.Setenv("PROBE_OUT", during)
	d, main, gh := realLockLocalRunRepo(t,
		"\t@env -u WF_TEST_LOCK_HELD_BY python3 scripts/test-lock.py acquire --watch $$$$ --slots 2 > \"$$PROBE_OUT\"\n")
	helper := filepath.Join(d.Root, "scripts", "test-lock.py")
	var after string
	d.GH = blockingPostGH{gh, func() {
		var stop func()
		after, stop = outsideAcquire(t, helper)
		stop()
	}}
	check := localCheck{Name: suiteCheckName, Command: []string{"make", "test"}, TimeoutMinutes: 5, Weight: 2}
	rec, err := runLocalCheck(context.Background(), d, check, main)
	if err != nil || !rec.passed() {
		t.Fatalf("run: record=%+v err=%v", rec, err)
	}
	if got, _ := os.ReadFile(during); strings.TrimSpace(string(got)) != "timeout" {
		t.Errorf("a waiter during the suite answered %q, want timeout (the run holds the slot)", got)
	}
	if after != "acquired" {
		t.Errorf("a waiter during the GitHub phase answered %q, want acquired (the run released the slot)", after)
	}
}

// #383: a second run in the same process queues for the slot like anyone else
// rather than passing through on the first run's hold.
func TestSuiteRerunQueuesForTheSlot(t *testing.T) {
	d, main, _ := realLockLocalRunRepo(t, "\t@echo suite passed\n")
	helper := filepath.Join(d.Root, "scripts", "test-lock.py")
	check := localCheck{Name: suiteCheckName, Command: []string{"make", "test"}, TimeoutMinutes: 5, Weight: 2}
	if rec, err := runLocalCheck(context.Background(), d, check, main); err != nil || !rec.passed() {
		t.Fatalf("run 1: record=%+v err=%v", rec, err)
	}
	state, stop := outsideAcquire(t, helper)
	if state != "acquired" {
		stop()
		t.Fatalf("after run 1 the slot should be free; outside acquire answered %q", state)
	}
	_, err := runLocalCheck(context.Background(), d, check, main)
	stop()
	if !errors.Is(err, errUnboundedSuite) {
		t.Fatalf("run 2 under an outside holder: err=%v, want the lock timeout (it must queue, not pass through)", err)
	}
	t.Setenv("WF_TEST_LOCK_TIMEOUT", "20")
	if rec, err := runLocalCheck(context.Background(), d, check, main); err != nil || !rec.passed() {
		t.Fatalf("run 3 once the outside holder left: record=%+v err=%v", rec, err)
	}
}
