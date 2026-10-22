package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

type operatingFixture struct{ root, remote, target string }

func operatingGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func operatingRepo(t *testing.T) operatingFixture {
	t.Helper()
	d := t.TempDir()
	// Agent identity belongs to sdlc in production. Supply that executable
	// boundary explicitly so these Git fixtures never use a host installation.
	bin := filepath.Join(d, "bin")
	if err := os.Mkdir(bin, 0755); err != nil {
		t.Fatal(err)
	}
	identity := "#!/bin/sh\n[ \"$#\" -eq 1 ] && [ \"$1\" = agent-id ] || exit 64\nprintf '%s\\n' 'Agent: test/workingdoc/fixture'\n"
	if err := os.WriteFile(filepath.Join(bin, "sdlc"), []byte(identity), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	f := operatingFixture{filepath.Join(d, "main"), filepath.Join(d, "origin.git"), "workshop/workingdoc/selected.md"}
	operatingGit(t, d, "init", "--bare", f.remote)
	operatingGit(t, d, "init", "-b", "main", f.root)
	operatingGit(t, f.root, "config", "user.name", "Test")
	operatingGit(t, f.root, "config", "user.email", "test@example.com")
	operatingWrite(t, f.root, "unrelated.md", []byte("unchanged\n"))
	operatingGit(t, f.root, "add", ".")
	operatingGit(t, f.root, "commit", "-m", "base")
	operatingGit(t, f.root, "remote", "add", "origin", f.remote)
	operatingGit(t, f.root, "push", "-u", "origin", "main")
	return f
}

func operatingWrite(t *testing.T, root, path string, data []byte) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func operatingIndex(t *testing.T, root string) []byte {
	t.Helper()
	p := operatingGit(t, root, "rev-parse", "--git-path", "index")
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestOperatingPublishExactPayloadAndRetry(t *testing.T) {
	f := operatingRepo(t)
	payload := []byte("  exact\x00 bytes\n\n")
	operatingWrite(t, f.root, f.target, []byte("independent staged bytes\n"))
	operatingGit(t, f.root, "add", f.target)
	operatingWrite(t, f.root, f.target, []byte("newer editor text\n"))
	unrelated := operatingGit(t, f.root, "ls-files", "--stage", "--", "unrelated.md")
	commit, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, payload, "")
	if err != nil {
		t.Fatal(err)
	}
	if trailer := operatingGit(t, f.root, "show", "-s", "--format=%(trailers:key=Agent,valueonly)", commit); trailer != "test/workingdoc/fixture" {
		t.Fatalf("publication agent trailer = %q", trailer)
	}
	cmd := exec.Command("git", "-C", f.remote, "show", commit+":"+f.target)
	blob, err := cmd.Output()
	if err != nil || !bytes.Equal(blob, payload) {
		t.Fatalf("blob=%q err=%v", blob, err)
	}
	if got := operatingGit(t, f.root, "diff", "--cached", "--name-only", "HEAD"); got != "" {
		t.Fatalf("successful publication left staged changes: %s", got)
	}
	if got := operatingGit(t, f.root, "ls-files", "--stage", "--", "unrelated.md"); got != unrelated {
		t.Fatal("unrelated index entry changed")
	}
	live, _ := os.ReadFile(filepath.Join(f.root, f.target))
	if string(live) != "newer editor text\n" {
		t.Fatalf("live=%q", live)
	}
	again, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, payload, "")
	if err != nil || again != commit {
		t.Fatalf("retry=%s %v", again, err)
	}
	if n := operatingGit(t, f.root, "rev-list", "--count", "HEAD"); n != "2" {
		t.Fatalf("duplicate commits: %s", n)
	}
}

func TestOperatingRefusesUnsafeState(t *testing.T) {
	for _, kind := range []string{"branch", "dirty", "staged", "symlink", "escape", "remote-conflict", "unrelated-ahead"} {
		t.Run(kind, func(t *testing.T) {
			f := operatingRepo(t)
			target := f.target
			switch kind {
			case "branch":
				operatingGit(t, f.root, "checkout", "-b", "feature")
			case "dirty":
				operatingWrite(t, f.root, "unrelated.md", []byte("edit"))
			case "staged":
				operatingWrite(t, f.root, "unrelated.md", []byte("edit"))
				operatingGit(t, f.root, "add", "unrelated.md")
			case "symlink":
				if err := os.Symlink(t.TempDir(), filepath.Join(f.root, "workshop")); err != nil {
					t.Fatal(err)
				}
			case "escape":
				target = "../outside.md"
			case "remote-conflict":
				operatingWrite(t, f.root, target, []byte("other"))
				operatingGit(t, f.root, "add", target)
				operatingGit(t, f.root, "commit", "-m", "other")
				operatingGit(t, f.root, "push", "origin", "main")
			case "unrelated-ahead":
				operatingWrite(t, f.root, "unrelated.md", []byte("edit"))
				operatingGit(t, f.root, "add", ".")
				operatingGit(t, f.root, "commit", "-m", "unrelated")
			}
			before := operatingGit(t, f.remote, "rev-parse", "main")
			if _, err := publishOperatingWorkingdoc(context.Background(), f.root, target, []byte("wanted"), ""); err == nil {
				t.Fatal("accepted unsafe publication")
			}
			if after := operatingGit(t, f.remote, "rev-parse", "main"); after != before {
				t.Fatal("remote changed")
			}
		})
	}
}

func TestOperatingInterruptedPushRetry(t *testing.T) {
	for _, phase := range []string{"after-commit", "after-push"} {
		t.Run(phase, func(t *testing.T) {
			f := operatingRepo(t)
			data := []byte("snapshot\n")
			operatingWrite(t, f.root, f.target, data)
			stop := errors.New("interrupted")
			_, err := publishOperatingWorkingdocHook(context.Background(), f.root, f.target, data, "", func(p string) error {
				if p == phase {
					return stop
				}
				return nil
			})
			if !errors.Is(err, stop) {
				t.Fatalf("got %v", err)
			}
			retained := operatingGit(t, f.root, "rev-parse", "HEAD")
			commit, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, data, "")
			if err != nil || commit != retained {
				t.Fatalf("retry=%s retained=%s err=%v", commit, retained, err)
			}
			if n := operatingGit(t, f.root, "rev-list", "--count", "HEAD"); n != "2" {
				t.Fatalf("duplicate: %s", n)
			}
		})
	}
}

func TestOperatingMutationBoundaries(t *testing.T) {
	for _, kind := range []string{"target", "head-before", "head-after", "index-hook"} {
		t.Run(kind, func(t *testing.T) {
			f := operatingRepo(t)
			payload := []byte("validated\n")
			operatingWrite(t, f.root, f.target, payload)
			remote := operatingGit(t, f.remote, "rev-parse", "main")
			hook := func(phase string) error {
				if kind == "target" && phase == "before-commit" {
					operatingWrite(t, f.root, f.target, []byte("new editor bytes\n"))
				}
				if (kind == "head-before" && phase == "before-commit") || (kind == "head-after" && phase == "after-commit") {
					operatingWrite(t, f.root, "unrelated.md", []byte("racing commit"))
					operatingGit(t, f.root, "add", "unrelated.md")
					operatingGit(t, f.root, "commit", "-m", "race")
				}
				return nil
			}
			if kind == "index-hook" {
				// A real Git hook can alter the private commit tree. Verification must catch it.
				hookPath := filepath.Join(f.root, ".git", "hooks", "pre-commit")
				if err := os.WriteFile(hookPath, []byte("#!/bin/sh\nprintf injected > unrelated.md\ngit add unrelated.md\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			commit, err := publishOperatingWorkingdocHook(context.Background(), f.root, f.target, payload, "", hook)
			if kind == "target" {
				if err != nil {
					t.Fatal(err)
				}
				if got := operatingGit(t, f.remote, "show", commit+":"+f.target); got != "validated" {
					t.Fatalf("payload=%q", got)
				}
				b, _ := os.ReadFile(filepath.Join(f.root, f.target))
				if string(b) != "new editor bytes\n" {
					t.Fatalf("live=%q", b)
				}
			} else {
				if err == nil {
					t.Fatal("accepted mutation")
				}
				if got := operatingGit(t, f.remote, "rev-parse", "main"); got != remote {
					t.Fatal("pushed mutated commit")
				}
			}
		})
	}
}

func TestOperatingPinnedBaselineResettle(t *testing.T) {
	f := operatingRepo(t)
	first := []byte("first snapshot\n")
	operatingWrite(t, f.root, f.target, first)
	baseline, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, first, "")
	if err != nil {
		t.Fatal(err)
	}
	second := []byte("second snapshot\n")
	operatingWrite(t, f.root, f.target, second)
	// The original real index still has no selected entry. Publication must
	// isolate it again and compare remote authority to the pinned blob.
	commit, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, second, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if commit == baseline {
		t.Fatal("resettle did not publish")
	}
	if got := operatingGit(t, f.remote, "show", commit+":"+f.target); got != "second snapshot" {
		t.Fatalf("blob=%q", got)
	}
	if _, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, []byte("third"), baseline); err == nil {
		t.Fatal("stale baseline accepted distinct remote content")
	}
	if got := operatingGit(t, f.remote, "rev-parse", "main"); got != commit {
		t.Fatal("conflict changed remote")
	}
}

func TestOperatingRejectedPushKeepsCommitAndIndex(t *testing.T) {
	f := operatingRepo(t)
	hook := filepath.Join(f.remote, "hooks", "pre-receive")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	data := []byte("retry snapshot\n")
	operatingWrite(t, f.root, f.target, data)
	index := operatingIndex(t, f.root)
	remote := operatingGit(t, f.remote, "rev-parse", "main")
	retained, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, data, "")
	if err == nil || retained == "" || retained == remote {
		t.Fatalf("commit=%s err=%v", retained, err)
	}
	if !bytes.Equal(index, operatingIndex(t, f.root)) {
		t.Fatal("rejected push changed real index")
	}
	if got := operatingGit(t, f.remote, "rev-parse", "main"); got != remote {
		t.Fatal("rejected push changed remote")
	}
	// Retry with another projection must preserve the retained distinct commit.
	if _, err = publishOperatingWorkingdoc(context.Background(), f.root, f.target, []byte("different"), ""); err == nil {
		t.Fatal("distinct retry accepted")
	}
	if got := operatingGit(t, f.root, "rev-parse", "HEAD"); got != retained {
		t.Fatal("distinct retry replaced commit")
	}
	if err = os.Remove(hook); err != nil {
		t.Fatal(err)
	}
	commit, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, data, "")
	if err != nil || commit != retained {
		t.Fatalf("retry=%s err=%v", commit, err)
	}
}

func TestOperatingCancellationBeforeCommit(t *testing.T) {
	f := operatingRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	head := operatingGit(t, f.root, "rev-parse", "HEAD")
	_, err := publishOperatingWorkingdocHook(ctx, f.root, f.target, []byte("payload"), "", func(phase string) error {
		if phase == "before-commit" {
			cancel()
		}
		return nil
	})
	if err == nil {
		t.Fatal("cancellation accepted")
	}
	if got := operatingGit(t, f.root, "rev-parse", "HEAD"); got != head {
		t.Fatal("cancelled operation committed")
	}
	if got := operatingGit(t, f.remote, "rev-parse", "main"); got != head {
		t.Fatal("cancelled operation pushed")
	}
}

func TestOperatingRetryConfirmsPublicationAfterRemoteAdvances(t *testing.T) {
	f := operatingRepo(t)
	data := []byte("published\n")
	operatingWrite(t, f.root, f.target, data)
	_, err := publishOperatingWorkingdocHook(context.Background(), f.root, f.target, data, "", func(phase string) error {
		if phase == "after-push" {
			return errors.New("lost acknowledgment")
		}
		return nil
	})
	if err == nil {
		t.Fatal("missing injected interruption")
	}
	peer := filepath.Join(t.TempDir(), "peer")
	operatingGit(t, filepath.Dir(peer), "clone", "-b", "main", f.remote, peer)
	operatingGit(t, peer, "config", "user.name", "Peer")
	operatingGit(t, peer, "config", "user.email", "peer@example.com")
	operatingWrite(t, peer, "unrelated.md", []byte("later remote change\n"))
	operatingGit(t, peer, "add", "unrelated.md")
	operatingGit(t, peer, "commit", "-m", "later")
	operatingGit(t, peer, "push", "origin", "main")
	remote := operatingGit(t, f.remote, "rev-parse", "main")
	commit, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, data, "")
	if err != nil || commit != remote {
		t.Fatalf("retry=%s err=%v", commit, err)
	}
}

func TestOperatingCancellationStopsGitHookDescendants(t *testing.T) {
	f := operatingRepo(t)
	d := t.TempDir()
	ready := filepath.Join(d, "ready")
	blocked := filepath.Join(d, "blocked")
	if err := syscall.Mkfifo(ready, 0600); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(blocked, 0600); err != nil {
		t.Fatal(err)
	}
	// RDWR keeps the reader open before the hook arrives; readiness is an
	// actual hook write, independent of process launch/instrumentation time.
	signal, err := os.OpenFile(ready, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer signal.Close()
	childFile := filepath.Join(d, "child")
	hook := "#!/bin/sh\ncat '" + blocked + "' &\nprintf '%s' $! > '" + childFile + "'\nprintf r > '" + ready + "'\nwait\n"
	if err = os.WriteFile(filepath.Join(f.root, ".git", "hooks", "pre-commit"), []byte(hook), 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := publishOperatingWorkingdoc(ctx, f.root, f.target, []byte("snapshot"), ""); done <- e }()
	arrived := make(chan error, 1)
	go func() { b := make([]byte, 1); _, e := signal.Read(b); arrived <- e }()
	deadline, _ := t.Deadline()
	if deadline.IsZero() {
		deadline = time.Now().Add(time.Minute)
	}
	guard := time.NewTimer(time.Until(deadline))
	defer guard.Stop()
	select {
	case err = <-arrived:
		if err != nil {
			t.Fatal(err)
		}
	case err = <-done:
		t.Fatalf("publisher stopped before hook readiness: %v", err)
	case <-guard.C:
		t.Fatal("hook never reached readiness")
	}
	pidBytes, err := os.ReadFile(childFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	cancel()
	// Git's child inherits the output pipes. Killing Git alone takes the
	// one-second WaitDelay and reports ErrWaitDelay; killing its group closes
	// them and reports the cancellation promptly without surviving work.
	select {
	case err = <-done:
		if err == nil || strings.Contains(err.Error(), "WaitDelay") {
			t.Fatalf("cancellation=%v", err)
		}
	case <-guard.C:
		t.Fatal("cancellation did not return")
	}
	state, _ := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "stat=").Output()
	if strings.TrimSpace(string(state)) != "" && !strings.HasPrefix(strings.TrimSpace(string(state)), "Z") {
		t.Fatalf("Git descendant %d remains runnable: %s", pid, state)
	}
}

func TestOperatingPublicationPreservesConcurrentIndex(t *testing.T) {
	for _, kind := range []string{"changed", "locked"} {
		t.Run(kind, func(t *testing.T) {
			f := operatingRepo(t)
			data := []byte("snapshot\n")
			operatingWrite(t, f.root, f.target, data)
			var concurrent []byte
			commit, err := publishOperatingWorkingdocHook(context.Background(), f.root, f.target, data, "", func(phase string) error {
				if phase != "before-index" {
					return nil
				}
				if kind == "changed" {
					operatingWrite(t, f.root, "unrelated.md", []byte("newly staged\n"))
					operatingGit(t, f.root, "add", "unrelated.md")
				} else {
					lock := filepath.Join(f.root, ".git", "index.lock")
					if e := os.WriteFile(lock, []byte("another writer"), 0600); e != nil {
						t.Fatal(e)
					}
					t.Cleanup(func() { _ = os.Remove(lock) })
				}
				concurrent = operatingIndex(t, f.root)
				return nil
			})
			var warning interface{ PublishedCommit() string }
			if err == nil || !errors.As(err, &warning) || warning.PublishedCommit() != commit {
				t.Fatalf("commit=%s warning=%v", commit, err)
			}
			instruction := "git restore --staged --source " + commit + " -- ':(literal)" + f.target + "'"
			if !strings.Contains(err.Error(), instruction) {
				t.Fatalf("missing reconciliation instruction %q: %v", instruction, err)
			}
			if got := operatingGit(t, f.remote, "rev-parse", "main"); got != commit {
				t.Fatal("warning is not confirmed publication")
			}
			if !bytes.Equal(concurrent, operatingIndex(t, f.root)) {
				t.Fatal("concurrent index overwritten")
			}
			if kind == "locked" {
				b, e := os.ReadFile(filepath.Join(f.root, ".git", "index.lock"))
				if e != nil || string(b) != "another writer" {
					t.Fatalf("another writer's lock changed: %q %v", b, e)
				}
			}
			live, _ := os.ReadFile(filepath.Join(f.root, f.target))
			if !bytes.Equal(live, data) {
				t.Fatal("live target changed")
			}
		})
	}
}

func TestOperatingExpectedCommitRejectsAnnotatedTagObject(t *testing.T) {
	f := operatingRepo(t)
	operatingWrite(t, f.root, f.target, []byte("baseline"))
	operatingGit(t, f.root, "add", f.target)
	operatingGit(t, f.root, "commit", "-m", "baseline")
	operatingGit(t, f.root, "push", "origin", "main")
	operatingGit(t, f.root, "tag", "-a", "baseline", "-m", "tag")
	tag := operatingGit(t, f.root, "rev-parse", "baseline")
	if _, err := publishOperatingWorkingdoc(context.Background(), f.root, f.target, []byte("desired"), tag); err == nil {
		t.Fatal("accepted tag object as expected commit")
	}
}

func TestOperatingRefreshBeforeTargetPreparation(t *testing.T) {
	for _, kind := range []string{"behind", "dirty", "divergent", "selected-edit", "selected-conflict"} {
		t.Run(kind, func(t *testing.T) {
			f := operatingRepo(t)
			peer := filepath.Join(t.TempDir(), "peer")
			operatingGit(t, filepath.Dir(peer), "clone", "-b", "main", f.remote, peer)
			operatingGit(t, peer, "config", "user.name", "Peer")
			operatingGit(t, peer, "config", "user.email", "peer@example.com")
			remotePath := "unrelated.md"
			if kind == "selected-conflict" {
				remotePath = f.target
			}
			operatingWrite(t, peer, remotePath, []byte("remote advancement\n"))
			operatingGit(t, peer, "add", remotePath)
			operatingGit(t, peer, "commit", "-m", "remote")
			operatingGit(t, peer, "push", "origin", "main")
			if kind == "dirty" || kind == "divergent" {
				operatingWrite(t, f.root, "unrelated.md", []byte("local change\n"))
				if kind == "divergent" {
					operatingGit(t, f.root, "add", "unrelated.md")
					operatingGit(t, f.root, "commit", "-m", "divergent")
				}
			}
			if strings.HasPrefix(kind, "selected-") {
				operatingWrite(t, f.root, f.target, []byte("preserve live selected bytes\n"))
			}
			if kind == "selected-edit" {
				operatingGit(t, f.root, "add", f.target)
			}
			selectedEntry := operatingGit(t, f.root, "ls-files", "--stage", "--", f.target)
			head := operatingGit(t, f.root, "rev-parse", "HEAD")
			index := operatingIndex(t, f.root)
			err := refreshOperatingWorkingdoc(context.Background(), f.root, f.target)
			if kind == "behind" || kind == "selected-edit" {
				if err != nil {
					t.Fatal(err)
				}
				if got := operatingGit(t, f.root, "rev-parse", "HEAD"); got != operatingGit(t, f.remote, "rev-parse", "main") {
					t.Fatal("main not refreshed")
				}
			} else {
				if err == nil {
					t.Fatal("unsafe refresh accepted")
				}
				if got := operatingGit(t, f.root, "rev-parse", "HEAD"); got != head {
					t.Fatal("unsafe refresh moved HEAD")
				}
				if !bytes.Equal(index, operatingIndex(t, f.root)) {
					t.Fatal("unsafe refresh altered index")
				}
			}
			if strings.HasPrefix(kind, "selected-") {
				b, _ := os.ReadFile(filepath.Join(f.root, f.target))
				if string(b) != "preserve live selected bytes\n" {
					t.Fatal("selected live file overwritten")
				}
				if got := operatingGit(t, f.root, "ls-files", "--stage", "--", f.target); got != selectedEntry {
					t.Fatal("refresh altered selected index entry")
				}
			}
		})
	}
}

func TestOperatingPublicationWarningQuotesSelectedPath(t *testing.T) {
	w := &operatingPublicationWarning{commit: strings.Repeat("a", 40), target: "workshop/workingdoc/operator's.md", cause: errors.New("index changed")}
	if !strings.Contains(w.Error(), " -- ':(literal)workshop/workingdoc/operator'\"'\"'s.md'") {
		t.Fatalf("unsafe reconciliation command: %s", w.Error())
	}
}
