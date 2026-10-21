//go:build unix

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// This helper uses stdin/stdout barriers rather than elapsed time to force the
// directory -> parent / parent -> directory ordering across real flock owners.
func TestResumeStoreProcessHelper(t *testing.T) {
	mode := os.Getenv("SDLC_RESUME_STORE_HELPER")
	if mode == "" {
		return
	}
	root := os.Getenv("SDLC_RESUME_STORE_ROOT")
	input := bufio.NewScanner(os.Stdin)
	wait := func() {
		t.Helper()
		if !input.Scan() || input.Text() != "continue" {
			t.Fatal("missing subprocess barrier")
		}
	}
	switch mode {
	case "staging":
		lock, err := lockResumeDirectory(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Close()
		tmp, err := os.CreateTemp(filepath.Join(root, "attempts"), ".resume-*")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.WriteString("uncommitted proof"); err != nil {
			t.Fatal(err)
		}
		if err := tmp.Sync(); err != nil {
			t.Fatal(err)
		}
		tmp.Close()
		fmt.Println("staged")
		wait()
	case "directory":
		lock, err := lockResumeDirectory(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Close()
		fmt.Println("locked")
		wait()
	case "parent", "publisher":
		lock, err := lockResumeParent(context.Background(), root, "aa")
		if err != nil {
			t.Fatal(err)
		}
		defer lock.Close()
		fmt.Println("locked")
		wait()
		if mode == "publisher" {
			in := resumeCoreFixture(t)
			in.Parent = lock.Attempt
			in.ParentRecordSHA256 = resumeRecordDigest(lock.Raw)
			in.ParentLogSHA256 = lock.Attempt.LogSHA256
			raw, err := json.Marshal(resumeCoreLink(in))
			if err != nil {
				t.Fatal(err)
			}
			// The parent remains locked until publication finishes. A pruner
			// waiting for it while owning the directory would form a cycle.
			fmt.Println("publishing")
			if err := publishResumeEvidence(root, lock, raw); err != nil {
				t.Fatal(err)
			}
			fmt.Println("published")
		}
	case "pruner":
		now, err := time.Parse(time.RFC3339Nano, os.Getenv("SDLC_RESUME_STORE_NOW"))
		if err != nil {
			t.Fatal(err)
		}
		if err := pruneSuiteAttemptsWithHook(root, now, func() {
			fmt.Println("snapshot")
			wait()
		}); err != nil {
			t.Fatal(err)
		}
		fmt.Println("pruned")
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}
}

func TestResumeStoreProcessStagingCrashCleanup(t *testing.T) {
	root := t.TempDir()
	publisher := startResumeStoreProcess(t, root, "staging", time.Now().UTC())
	publisher.await(t, "staged")
	entries, err := os.ReadDir(filepath.Join(root, "attempts"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("staged files: %v %v", entries, err)
	}
	path := filepath.Join(root, "attempts", entries[0].Name())
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("active stage removed")
	}
	if err := processgroup.Terminate(publisher.cmd, true); err != nil {
		t.Fatal(err)
	}
	select {
	case <-publisher.done:
	case <-time.After(testfix.HangCeiling):
		t.Fatal("crashed publisher did not exit")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("crash fixture did not leave staging residue")
	}
	if err := pruneSuiteAttempts(root, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("crashed stage retained: %v", err)
	}
}

type resumeStoreProcess struct {
	cmd    *exec.Cmd
	input  io.WriteCloser
	lines  chan string
	done   chan struct{}
	stderr bytes.Buffer
	err    error
}

func startResumeStoreProcess(t *testing.T, root, mode string, now time.Time) *resumeStoreProcess {
	t.Helper()
	p := &resumeStoreProcess{lines: make(chan string, 16), done: make(chan struct{})}
	p.cmd = exec.Command(os.Args[0], "-test.run=^TestResumeStoreProcessHelper$")
	processgroup.Configure(p.cmd)
	p.cmd.WaitDelay = time.Second
	p.cmd.Env = append(os.Environ(), "SDLC_RESUME_STORE_HELPER="+mode, "SDLC_RESUME_STORE_ROOT="+root, "SDLC_RESUME_STORE_NOW="+now.Format(time.RFC3339Nano))
	p.cmd.Stderr = &p.stderr
	var err error
	p.input, err = p.cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := p.cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := p.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		s := bufio.NewScanner(output)
		for s.Scan() {
			p.lines <- s.Text()
		}
		close(p.lines)
		p.err = p.cmd.Wait()
		close(p.done)
	}()
	t.Cleanup(func() {
		_ = p.input.Close()
		_ = processgroup.Terminate(p.cmd, true)
		select {
		case <-p.done:
		case <-time.After(testfix.HangCeiling):
			t.Error("resume-store helper was not reaped")
		}
	})
	return p
}

func (p *resumeStoreProcess) await(t *testing.T, want string) {
	t.Helper()
	select {
	case got, ok := <-p.lines:
		if !ok || got != want {
			t.Fatalf("helper barrier = %q (open %v), want %q", got, ok, want)
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatalf("resume-store helper stuck before %q", want)
	}
}

func (p *resumeStoreProcess) proceed(t *testing.T) {
	t.Helper()
	if _, err := io.WriteString(p.input, "continue\n"); err != nil {
		t.Fatal(err)
	}
}

func (p *resumeStoreProcess) wait(t *testing.T) {
	t.Helper()
	select {
	case <-p.done:
		if p.err != nil {
			t.Fatalf("resume-store helper: %v\n%s", p.err, p.stderr.String())
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("resume-store helper did not exit")
	}
}

func TestResumeStoreProcessSnapshotPublisherPreservesNewLink(t *testing.T) {
	root := t.TempDir()
	parent := resumeStoreParent(t, root, "aa")
	now := resumeCoreFixture(t).Parent.FinishedAt.Add(40 * 24 * time.Hour)
	old := now.Add(-40 * 24 * time.Hour)
	if err := os.Chtimes(parent, old, old); err != nil {
		t.Fatal(err)
	}
	publisher := startResumeStoreProcess(t, root, "publisher", now)
	publisher.await(t, "locked")
	pruner := startResumeStoreProcess(t, root, "pruner", now)
	pruner.await(t, "snapshot")
	// Snapshot contains only the old parent. The publisher owns that parent
	// while waiting for the directory, so legacy pruning must skip its inode.
	publisher.proceed(t)
	publisher.await(t, "publishing")
	pruner.proceed(t)
	pruner.await(t, "pruned")
	pruner.wait(t)
	publisher.await(t, "published")
	publisher.wait(t)
	for _, id := range []string{"aa", "bb"} {
		if _, err := os.Stat(filepath.Join(root, "attempts", id+".json")); err != nil {
			t.Fatalf("snapshot race lost %s: %v", id, err)
		}
	}
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"aa", "bb"} {
		if _, err := os.Stat(filepath.Join(root, "attempts", id+".json")); err != nil {
			t.Fatalf("pending publication member %s was lost: %v", id, err)
		}
	}
}

func TestResumeStoreProcessPruneSkipsLockedTerminalParent(t *testing.T) {
	root, link, now := retentionFixture(t, true)
	owner := startResumeStoreProcess(t, root, "parent", now)
	owner.await(t, "locked")
	pruner := startResumeStoreProcess(t, root, "pruner", now)
	pruner.await(t, "snapshot")
	pruner.proceed(t)
	pruner.await(t, "pruned")
	pruner.wait(t)
	for _, id := range []string{link.RerunOf, link.ID, link.ChildID} {
		if _, err := os.Stat(filepath.Join(root, "attempts", id+".json")); err != nil {
			t.Fatalf("locked bundle lost %s: %v", id, err)
		}
	}
	owner.proceed(t)
	owner.wait(t)
	if err := pruneSuiteAttempts(root, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "attempts", link.ID+".json")); !os.IsNotExist(err) {
		t.Fatalf("idle terminal bundle was not retired: %v", err)
	}
}

func TestResumeStoreProcessDirectoryLockOwnerDeath(t *testing.T) {
	root := t.TempDir()
	resumeStoreParent(t, root, "aa")
	owner := startResumeStoreProcess(t, root, "directory", time.Time{})
	owner.await(t, "locked")
	// A strict nonblocking probe must return while the other process still
	// owns the flock. Only the enclosing hang ceiling bounds a broken probe.
	type result struct {
		lock *resumeDirectoryLock
		err  error
	}
	probe := make(chan result, 1)
	go func() { l, err := tryLockResumeDirectory(root); probe <- result{l, err} }()
	select {
	case got := <-probe:
		if got.lock != nil {
			got.lock.Close()
			t.Fatal("directory probe acquired another process's lock")
		}
		if got.err == nil {
			t.Fatal("directory probe did not report contention")
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("nonblocking directory probe waited for its owner")
	}
	if err := owner.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-owner.done:
	case <-time.After(testfix.HangCeiling):
		t.Fatal("killed directory owner was not reaped")
	}
	lock, err := tryLockResumeDirectory(root)
	if err != nil {
		t.Fatalf("kernel did not release dead owner's directory flock: %v", err)
	}
	defer lock.Close()
	if err := lock.verify(); err != nil {
		t.Fatal(err)
	}
}
