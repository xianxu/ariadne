package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// The #252 binary must run the legacy workflow unchanged in any repository
// without a tracker (no remote issue-tracker branch, no cutover marker): that
// is what lets #252 ship before any repository cuts over. Found by the A3 soak:
// `issue new` fetched the absent tracker ref and died.
func TestLegacyRepositoryIssueNewAllocatesFromFiles(t *testing.T) {
	r := legacyRepo(t)
	out, err := slotRun(t, r.root, "issue", "new", "legacy mode still works", "--slug", "legacy-new")
	if err != nil {
		t.Fatalf("issue new in a legacy repository: %v\n%s", err, out)
	}
	if !strings.Contains(r.git("ls-files", "workshop/issues"), "000004-legacy-new.md") {
		t.Fatalf("legacy allocation should give 000004 (max of issues/ and history/ + 1):\n%s\n%s", r.git("ls-files", "workshop/issues"), out)
	}
	if r.git("ls-remote", "origin", "refs/heads/issue-tracker") != "" {
		t.Fatal("legacy issue new created a tracker")
	}
}

// The legacy lifecycle, run end to end by the #252 binary in a repository
// without a tracker: the proof that #252 can ship before any cutover. Each
// verb behaves as at pre-252-freeze; the tracker-only verbs refuse with the
// legacy way; the publish gate needs no tracker and no upstream.
func TestLegacyRepositoryFullLifecycle(t *testing.T) {
	r := legacyRepo(t)
	path := "workshop/issues/000004-cycle.md"
	mustSlotRun(t, r.root, "issue", "new", "legacy cycle", "--slug", "cycle")
	r.git("fetch", "-q", "origin")
	if r.git("ls-tree", "--name-only", "origin/main", "--", path) == "" {
		t.Fatal("legacy issue new did not publish its reservation to main")
	}
	mustSlotRun(t, r.root, "claim", "--issue", "4")
	r.git("fetch", "-q", "origin")
	if !strings.Contains(r.git("show", "origin/main:"+path), "status: working") {
		t.Fatal("legacy claim did not publish the working status to main")
	}
	// Legacy publication copies commits to main, so the resting branch diverges
	// from origin/main (the #249 problem the tracker removes): no fast-forward.
	mustSlotRun(t, r.root, "start-plan", "--issue", "4")
	if got := r.git("branch", "--show-current"); got != "main" {
		t.Fatalf("legacy start-plan moved the checkout to %s", got)
	}
	for _, args := range [][]string{
		{"issue", "set-title", "--issue", "4", "renamed"},
		{"issue", "move-detail", "--issue", "4"},
	} {
		if out, err := slotRun(t, r.root, args...); err == nil || !strings.Contains(out+err.Error(), "has not cut over") {
			t.Errorf("sdlc %s in a legacy repository: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	raw, err := os.ReadFile(filepath.Join(r.root, path))
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, path, strings.Replace(string(raw), "## Problem\n", "## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Scratch\n", 1))
	mustSlotRun(t, r.root, "change-code", "--issue", "4", "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon", "--flow", "quick")
	branch := r.git("branch", "--show-current")
	if branch == "main" {
		t.Fatal("legacy change-code did not create the issue branch")
	}
	if r.git("log", "--format=%s", "--grep=^#4: issue-sync: spec/plan at change-code", "HEAD") == "" {
		t.Fatal("legacy change-code did not commit the accepted design (its step-7 sync)")
	}
	writeRepoFile(t, r.root, "cmd/cycle.go", "package cycle\n")
	r.git("add", "cmd/cycle.go")
	r.git("commit", "-qm", "#4: implement")
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	mustSlotRun(t, r.root, "close", "--issue", "4", "--verified", "legacy cycle", "--actual", "1", "--no-atlas")
	if !strings.Contains(readRepoFileOr(t, r.root, path), "status: codecomplete") {
		t.Fatal("legacy close did not record codecomplete in the details")
	}
	// Landing is not scripted here: legacy publication of this file-shaped
	// history conflicts exactly as the pre-#252 binary does (the #249 pain the
	// tracker removes); the differential run in #252's Log proves the new binary
	// matches the old one verb for verb, and the legacy merge tests cover landing.
	if r.git("ls-remote", "origin", "refs/heads/issue-tracker") != "" {
		t.Fatal("the legacy lifecycle created a tracker")
	}
}

// Legacy pr/push/merge need no upstream for main: the transfer guard decides
// the repository is legacy before opening any tracker environment.
func TestLegacyGatesNeedNoUpstream(t *testing.T) {
	r := legacyRepo(t)
	r.git("branch", "--unset-upstream", "main")
	t.Chdir(r.root)
	if err := guardTransferredDetails(t.Context()); err != nil {
		t.Fatalf("legacy transfer guard without an upstream: %v", err)
	}
}

func readRepoFileOr(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return ""
	}
	return string(raw)
}

// Restored from pre-252-freeze (#252 M2 had moved them onto the tracker):
// legacy allocation and set-status under the repo lock.
func TestLegacyRepoLockConcurrentIssueNewSerializesAllocation(t *testing.T) {
	issues, history := newTestDirs(t)
	lock := newSerializingTestLock()
	restore := stubRepoLockAcquire(t, lock.acquire)
	defer restore()

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, title := range []string{"First concurrent issue", "Second concurrent issue"} {
		wg.Add(1)
		go func(title string) {
			defer wg.Done()
			_, stderr, err := executeSDLCTestCommand(
				"issue", "new", title,
				"--issues-dir", issues,
				"--history-dir", history,
			)
			if err != nil {
				errs <- err
				return
			}
			if strings.Contains(stderr, "index.lock") {
				errs <- &testError{"unexpected git index lock failure: " + stderr}
			}
		}(title)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	matches, err := filepath.Glob(filepath.Join(issues, "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(matches)
	if len(matches) != 2 {
		t.Fatalf("created %d issues, want 2: %v", len(matches), matches)
	}
	got := []string{filepath.Base(matches[0]), filepath.Base(matches[1])}
	if !strings.HasPrefix(got[0], "000001-") || !strings.HasPrefix(got[1], "000002-") {
		t.Fatalf("issue files should allocate distinct sequential IDs, got %v", got)
	}
	joined := strings.Join(got, "\n")
	for _, wantTitle := range []string{"first-concurrent-issue", "second-concurrent-issue"} {
		if !strings.Contains(joined, wantTitle) {
			t.Fatalf("issue files missing %q: %v", wantTitle, got)
		}
	}
	if waits := lock.waitMessages(); waits == "" || !strings.Contains(waits, "waiting for sdlc repo lock held by") {
		t.Fatalf("expected lock wait message, got %q", waits)
	}
}

func TestLegacyRepoLockSetStatusMutationWaits(t *testing.T) {
	issues, _ := newTestDirs(t)
	path := filepath.Join(issues, "000001-status.md")
	if err := os.WriteFile(path, []byte("---\nid: 000001\nstatus: open\nupdated: 2026-06-27\n---\n\n# Status\n\n## Log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lock := newSerializingTestLock()
	restore := stubRepoLockAcquire(t, lock.acquire)
	defer restore()

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := executeSDLCTestCommand("issue", "set-status", "working", "--issue", "1", "--issues-dir", issues, "--force")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "status: working") {
		t.Fatalf("status not updated:\n%s", data)
	}
	if waits := lock.waitMessages(); waits == "" || !strings.Contains(waits, "pid 777") {
		t.Fatalf("expected wait message with holder pid, got %q", waits)
	}
}
