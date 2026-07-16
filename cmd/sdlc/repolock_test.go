package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestRepoLockCommandMetadata(t *testing.T) {
	root := buildRoot()
	mutating := [][]string{
		{"claim"},
		{"change-code"},
		{"close"},
		{"issue", "new"},
		{"issue", "set-status"},
		{"project", "close"},
		{"set-status"},
		{"fetch"},
		{"merge"},
		{"milestone-close"},
		{"pr"},
		{"push"},
	}
	for _, path := range mutating {
		cmd := mustFindCommand(t, root, path...)
		if !commandNeedsRepoLock(cmd) {
			t.Fatalf("%v should require repo lock", path)
		}
	}
	for _, path := range [][]string{{"close"}, {"milestone-close"}} {
		cmd := mustFindCommand(t, root, path...)
		if commandAutoWrapsRepoLock(cmd) {
			t.Fatalf("%v should be manually lock-scoped, not whole-command wrapped", path)
		}
	}

	readOnly := [][]string{
		{"issue", "list"},
		{"issue", "show"},
		{"issue", "validate"},
		{"state"},
		{"start-plan"},
		{"actual"},
		{"active-time"},
		{"judge"},
		{"arch-principles"},
		{"estimate-source"},
	}
	for _, path := range readOnly {
		cmd := mustFindCommand(t, root, path...)
		if commandNeedsRepoLock(cmd) {
			t.Fatalf("%v should not require repo lock", path)
		}
	}
	if commandNeedsRepoLock(root) {
		t.Fatal("root command should not require repo lock")
	}
	// propagate-base mutates downstream repos through git -C <peer>; it is
	// intentionally outside this checkout's git-common-dir lock.
}

func TestLockedCommandFilesAvoidRawOSExit(t *testing.T) {
	for _, path := range []string{
		"claim.go",
		"changecode.go",
		"close.go",
		"fetch.go",
		"issue.go",
		"merge.go",
		"milestoneclose.go",
		"pr.go",
		"push.go",
		"setstatus.go",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "os.Exit(") {
			t.Fatalf("%s uses raw os.Exit inside a locked command body; use exitWithCode so lock cleanup runs", path)
		}
	}
}

func mustFindCommand(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()
	cur := root
	for _, part := range path {
		var next *cobra.Command
		for _, child := range cur.Commands() {
			if child.Name() == part {
				next = child
				break
			}
		}
		if next == nil {
			t.Fatalf("command path %v missing segment %q", path, part)
		}
		cur = next
	}
	return cur
}

func TestWithRepoTransactionLockSkipsReadOnlyCommand(t *testing.T) {
	cmd := &cobra.Command{Use: "state"}
	calls := 0
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		calls++
		return func() error { return nil }, nil
	})
	defer restore()

	ran := false
	if err := withRepoTransactionLock(cmd, func() error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("withRepoTransactionLock err: %v", err)
	}
	if !ran {
		t.Fatal("read-only command did not run")
	}
	if calls != 0 {
		t.Fatalf("read-only command acquired lock %d time(s)", calls)
	}
}

func TestWithRepoTransactionLockAcquiresAndReleasesMutatingCommand(t *testing.T) {
	cmd := markMutatingCommand(&cobra.Command{Use: "claim"})
	var acquired, released int
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		acquired++
		return func() error {
			released++
			return nil
		}, nil
	})
	defer restore()

	if err := withRepoTransactionLock(cmd, func() error { return nil }); err != nil {
		t.Fatalf("withRepoTransactionLock err: %v", err)
	}
	if acquired != 1 || released != 1 {
		t.Fatalf("acquired/released = %d/%d, want 1/1", acquired, released)
	}
}

func TestWithRequiredRepoTransactionLockAcquiresManualCommand(t *testing.T) {
	cmd := markManualLockCommand(&cobra.Command{Use: "close"})
	var acquired, released int
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		acquired++
		return func() error {
			released++
			return nil
		}, nil
	})
	defer restore()

	if err := withRequiredRepoTransactionLock(cmd, func() error { return nil }); err != nil {
		t.Fatalf("withRequiredRepoTransactionLock err: %v", err)
	}
	if acquired != 1 || released != 1 {
		t.Fatalf("acquired/released = %d/%d, want 1/1", acquired, released)
	}
}

func TestWithRepoTransactionLockIsContextReentrantOnly(t *testing.T) {
	cmd := markMutatingCommand(&cobra.Command{Use: "claim"})
	var acquired int
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		acquired++
		return func() error { return nil }, nil
	})
	defer restore()

	if err := withRepoTransactionLock(cmd, func() error {
		nested := markMutatingCommand(&cobra.Command{Use: "issue new"})
		nested.SetContext(cmd.Context())
		return withRepoTransactionLock(nested, func() error { return nil })
	}); err != nil {
		t.Fatalf("nested withRepoTransactionLock err: %v", err)
	}
	if acquired != 1 {
		t.Fatalf("inherited nested context should acquire once, got %d", acquired)
	}

	independent := markMutatingCommand(&cobra.Command{Use: "issue new"})
	independent.SetContext(context.Background())
	if err := withRepoTransactionLock(independent, func() error { return nil }); err != nil {
		t.Fatalf("independent withRepoTransactionLock err: %v", err)
	}
	if acquired != 2 {
		t.Fatalf("independent command context should acquire again, got %d", acquired)
	}
}

func TestWithRepoTransactionLockRegistersDieCleanup(t *testing.T) {
	cmd := markMutatingCommand(&cobra.Command{Use: "claim"})
	var released int
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		return func() error {
			released++
			return nil
		}, nil
	})
	defer restore()

	if err := withRepoTransactionLock(cmd, func() error {
		runDieCleanups()
		return nil
	}); err != nil {
		t.Fatalf("withRepoTransactionLock err: %v", err)
	}
	if released != 1 {
		t.Fatalf("die cleanup + normal defer released %d times, want exactly 1", released)
	}
}

func TestWrapRepoLockCommandsDoesNotWrapManualLockCommand(t *testing.T) {
	var acquired int
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		acquired++
		return func() error { return nil }, nil
	})
	defer restore()

	root := &cobra.Command{Use: "root"}
	manualRan := false
	manual := markManualLockCommand(&cobra.Command{
		Use:  "close",
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			manualRan = true
			return nil
		},
	})
	root.AddCommand(manual)
	wrapRepoLockCommands(root)
	root.SetArgs([]string{"close"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute manual err: %v", err)
	}
	if !manualRan {
		t.Fatal("manual command did not run")
	}
	if acquired != 0 {
		t.Fatalf("manual command should not be whole-command wrapped, acquired %d time(s)", acquired)
	}
	if !commandNeedsRepoLock(manual) {
		t.Fatal("manual command should still be registered as needing repo lock")
	}
}

func TestWrapRepoLockCommandsWrapsRunE(t *testing.T) {
	var acquired int
	restore := stubRepoLockAcquire(t, func(*cobra.Command) (func() error, error) {
		acquired++
		return func() error { return nil }, nil
	})
	defer restore()

	root := &cobra.Command{Use: "root"}
	mutatingRan := false
	mutating := markMutatingCommand(&cobra.Command{
		Use:  "mutate",
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			mutatingRan = true
			return nil
		},
	})
	root.AddCommand(mutating)
	wrapRepoLockCommands(root)
	root.SetArgs([]string{"mutate"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute mutating err: %v", err)
	}
	if !mutatingRan || acquired != 1 {
		t.Fatalf("mutatingRan/acquired = %v/%d, want true/1", mutatingRan, acquired)
	}

	readOnlyRan := false
	root = &cobra.Command{Use: "root"}
	root.AddCommand(&cobra.Command{
		Use:  "state",
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			readOnlyRan = true
			return nil
		},
	})
	wrapRepoLockCommands(root)
	root.SetArgs([]string{"state"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute read-only err: %v", err)
	}
	if !readOnlyRan || acquired != 1 {
		t.Fatalf("readOnlyRan/acquired = %v/%d, want true/1", readOnlyRan, acquired)
	}
}

func TestRepoLockGitCommonDirResolvesRelativePathFromSubdir(t *testing.T) {
	dir := t.TempDir()
	git(t, "", "init", "-b", "main", dir)
	subdir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(subdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	got, err := repoLockGitCommonDir()
	if err != nil {
		t.Fatalf("repoLockGitCommonDir err: %v", err)
	}
	want, err := filepath.EvalSymlinks(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("repoLockGitCommonDir = %q, want %q", got, want)
	}
}

func TestRepoLockGitCommonDirSharedAcrossLinkedWorktree(t *testing.T) {
	dir := t.TempDir()
	wt := filepath.Join(t.TempDir(), "feature-worktree")
	git(t, "", "init", "-b", "main", dir)
	git(t, dir, "config", "user.email", "e2e@example.com")
	git(t, dir, "config", "user.name", "e2e")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "README.md")
	git(t, dir, "commit", "-m", "seed")
	git(t, dir, "worktree", "add", "-b", "feature", wt)

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(wt); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	got, err := repoLockGitCommonDir()
	if err != nil {
		t.Fatalf("repoLockGitCommonDir err: %v", err)
	}
	want, err := filepath.EvalSymlinks(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("linked worktree common dir = %q, want %q", got, want)
	}
}

func stubRepoLockAcquire(t *testing.T, fn func(*cobra.Command) (func() error, error)) func() {
	t.Helper()
	orig := repoLockAcquireForCommand
	repoLockAcquireForCommand = fn
	return func() { repoLockAcquireForCommand = orig }
}

func TestRepoLockConcurrentIssueNewSerializesAllocation(t *testing.T) {
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

func TestRepoLockSetStatusMutationWaits(t *testing.T) {
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

type serializingTestLock struct {
	sem   chan struct{}
	mu    sync.Mutex
	waits bytes.Buffer
}

func newSerializingTestLock() *serializingTestLock {
	return &serializingTestLock{sem: make(chan struct{}, 1)}
}

func (l *serializingTestLock) acquire(cmd *cobra.Command) (func() error, error) {
	select {
	case l.sem <- struct{}{}:
		time.Sleep(20 * time.Millisecond)
	default:
		msg := "waiting for sdlc repo lock held by pid 777: sdlc issue new\n"
		_, _ = cmd.ErrOrStderr().Write([]byte(msg))
		l.mu.Lock()
		_, _ = l.waits.WriteString(msg)
		l.mu.Unlock()
		l.sem <- struct{}{}
	}
	return func() error {
		<-l.sem
		return nil
	}, nil
}

func (l *serializingTestLock) waitMessages() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.waits.String()
}

func executeSDLCTestCommand(args ...string) (stdout, stderr string, err error) {
	root := buildRoot()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	root.SetContext(context.Background())
	err = root.Execute()
	return out.String(), errOut.String(), err
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
