package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The real helper holds its lock until this test process exits. Each test has
// a private lock file, so later suite runs in the same process must reenter it.
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
	// Same check on a later integrated head, then a second named check: both
	// reacquire after the first run's watcher still holds the machine lock.
	for _, step := range []struct{ name, sha string }{{suiteCheckName, main}, {suiteCheckName, head}, {"smoke", head}} {
		check := localCheck{Name: step.name, Command: []string{"make", "test"}, TimeoutMinutes: 5, Weight: 2}
		rec, err := runLocalCheck(context.Background(), d, check, step.sha)
		if err != nil || !rec.passed() {
			t.Fatalf("%s on %s: record=%+v err=%v", step.name, step.sha, rec, err)
		}
	}
}
