//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Both real runs own the same checkout and check but different heads. The
// real helper queues them behind a private capacity-one flock: cancelling the
// newer lease must not recreate the older run's FIFO ticket or signal it.
// SDLC_RUNID_CLI=/absolute/bin/sdlc reuses the private fixture against an
// actual built CLI; GitHub is stubbed and no fleet checks are touched.
func TestLocalCheckCancelRunIDPreservesQueuedSibling(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "unused")
	helper, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d.Root, "scripts", "test-lock.py"), helper, 0700); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) string {
		t.Helper()
		out, err := d.Git(args...)
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(out)
	}
	if cli := os.Getenv("SDLC_RUNID_CLI"); cli != "" {
		if !filepath.IsAbs(cli) {
			t.Fatal("SDLC_RUNID_CLI must name an absolute executable")
		}
		wireLocalCheckSignalCLI(t, d.Root)
		decl := `{"version":1,"checks":[{"name":"suite","command":["sh","-c","echo $$ > \"$SDLC_LEASE_CHECK_PID\"; exec sleep 60"],"timeout_minutes":1,"weight":1}]}`
		if err := os.WriteFile(filepath.Join(d.Root, ".sdlc", "local-checks.json"), []byte(decl), 0600); err != nil {
			t.Fatal(err)
		}
		git("add", ".sdlc")
	}
	git("add", "scripts/test-lock.py")
	git("commit", "-qm", "use real private queue helper")
	olderSHA := git("rev-parse", "HEAD")
	if os.Getenv("SDLC_RUNID_CLI") != "" {
		git("push", "-q", "origin", "HEAD:main")
	}
	git("commit", "--allow-empty", "-qm", "second tested head")
	newerSHA := git("rev-parse", "HEAD")
	if os.Getenv("SDLC_RUNID_CLI") != "" {
		git("push", "-q", "origin", "HEAD:run-id-new")
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	runs := localCheckRunsDir(logDir)
	private := t.TempDir()
	lockPath := filepath.Join(private, "suite.lock")
	t.Setenv("WF_TEST_LOCK_FILE", lockPath)
	t.Setenv("WF_TEST_SUITES", "1")
	t.Setenv("WF_TEST_LOCK", "")
	t.Setenv("WF_TEST_LOCK_TIMEOUT", "")
	t.Setenv("WF_TEST_LOCK_HELD_BY", "")
	blocker, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { blocker.Close() })
	if err := flock.Exclusive(blocker, false); err != nil {
		t.Fatal(err)
	}

	older := startRunIDChild(t, d.Root, olderSHA)
	oldID, oldLease := awaitRunIDLease(t, runs, older, olderSHA)
	tickets := lockPath + ".tickets"
	oldTicket := awaitRunIDTickets(t, tickets, 1, older)[0]
	oldTicketInfo, err := os.Stat(oldTicket)
	if err != nil {
		t.Fatal(err)
	}
	oldTicketBytes, err := os.ReadFile(oldTicket)
	if err != nil {
		t.Fatal(err)
	}
	oldLeaseInfo, err := os.Stat(oldLease)
	if err != nil {
		t.Fatal(err)
	}
	oldLeaseBytes, err := os.ReadFile(oldLease)
	if err != nil {
		t.Fatal(err)
	}
	newer := startRunIDChild(t, d.Root, newerSHA)
	newID, newLease := awaitRunIDLease(t, runs, newer, newerSHA)
	awaitRunIDTickets(t, tickets, 2, older, newer)
	assertRunIDChildAlive(t, older)
	assertRunIDChildAlive(t, newer)
	assertRunIDHeld(t, oldLease)
	assertRunIDHeld(t, newLease)

	// Files capture child output directly; no concurrent bytes.Buffer readers.
	for _, item := range []struct {
		child *runIDChild
		id    string
	}{{older, oldID}, {newer, newID}} {
		if !strings.Contains(item.child.output(t), item.id) {
			t.Errorf("run does not expose its lease ID %s: %s", item.id, item.child.output(t))
		}
	}
	notice := newer.output(t)
	for _, want := range []string{oldID, abbrevSHA(olderSHA), abbrevSHA(newerSHA), "sdlc local-check cancel --run " + oldID} {
		if !strings.Contains(notice, want) {
			t.Errorf("second queued run lacks duplicate notice %q:\n%s", want, notice)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
	// The notice offers an exact action; it must not have replaced the older
	// run automatically. The sibling's stable ticket below also proves this.
	assertRunIDChildAlive(t, older)
	assertRunIDChildAlive(t, newer)

	t.Chdir(d.Root)
	cancel := func(args ...string) (string, error) {
		if cli := os.Getenv("SDLC_RUNID_CLI"); cli != "" {
			cmd := exec.Command(cli, append([]string{"local-check", "cancel"}, args...)...)
			cmd.Dir = d.Root
			output, err := cmd.CombinedOutput()
			return string(output), err
		}
		var output bytes.Buffer
		cmd := NewLocalCheckCmd()
		cmd.SetOut(&output)
		cmd.SetErr(&output)
		cmd.SetArgs(append([]string{"cancel"}, args...))
		err := cmd.Execute()
		return output.String(), err
	}
	foreignID, foreignPID, foreignDone := holdRunIDForeignLease(t, runs, newerSHA)
	controls := [][]string{
		{"--run", foreignID},
		{"--run", "lease-missing"},
		{"--run", "../" + oldID},
		{"--run", "./" + oldID},
		{"--run", filepath.Join(runs, oldID)},
		{"--run", oldID + "/nested"},
		{"--run", oldID + "\\nested"},
		{"--run", "*"},
		{"--run", ""},
		{"--run", newID, "--name", "suite"},
	}
	for _, args := range controls {
		output, err := cancel(args...)
		if err == nil || strings.Contains(err.Error(), "unknown flag") {
			t.Fatalf("selector %q must be refused by cancellation selection: err=%v output=%s", args, err, output)
		}
		assertRunIDChildAlive(t, older)
		assertRunIDChildAlive(t, newer)
		assertRunIDHeld(t, oldLease)
		assertRunIDHeld(t, newLease)
		select {
		case err := <-foreignDone:
			t.Fatalf("refused selector stopped foreign process: %v", err)
		default:
		}
		if err := syscall.Kill(foreignPID, 0); err != nil {
			t.Fatalf("refused selector signalled foreign lease's process: %v", err)
		}
	}

	output, err := cancel("--run", newID)
	if err != nil {
		t.Fatalf("cancel newer run %s: %v\n%s", newID, err, output)
	}
	awaitRunIDCancelled(t, newer)
	if _, err := os.Stat(newLease); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cancelled newer lease remains: %v", err)
	}
	assertRunIDChildAlive(t, older)
	assertRunIDHeld(t, oldLease)
	assertRunIDHeld(t, oldTicket)
	for _, saved := range []struct {
		path string
		info os.FileInfo
		raw  []byte
	}{
		{oldLease, oldLeaseInfo, oldLeaseBytes},
		{oldTicket, oldTicketInfo, oldTicketBytes},
	} {
		info, err := os.Stat(saved.path)
		if err != nil || !os.SameFile(saved.info, info) {
			t.Fatalf("survivor lease/ticket replaced or removed: %s (%v)", saved.path, err)
		}
		raw, err := os.ReadFile(saved.path)
		if err != nil || !bytes.Equal(raw, saved.raw) {
			t.Fatalf("survivor lease/ticket changed: %s (%v)", saved.path, err)
		}
	}
	if paths := awaitRunIDTickets(t, tickets, 1, older); paths[0] != oldTicket {
		t.Fatalf("survivor requeued under another ticket: %v; want %s", paths, oldTicket)
	}
	if _, err := os.Stat(older.checkPID); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("blocked survivor ran before capacity released: %v", err)
	}
	if err := syscall.Flock(int(blocker.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatal(err)
	}
	// The child check's PID marker exists only after the real helper admits it.
	deadline := time.Now().Add(testfix.HangCeiling)
	for {
		assertRunIDChildAlive(t, older)
		if raw, _ := os.ReadFile(older.checkPID); len(raw) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("survivor was not admitted after capacity released:\n%s", older.output(t))
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Retain a by-name compatibility control on the surviving, now-running job.
	output, err = cancel("--name", "suite")
	if err != nil {
		t.Fatalf("cancel admitted survivor by name: %v\n%s", err, output)
	}
	awaitRunIDCancelled(t, older)
	if _, err := os.Stat(oldLease); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("survivor lease remains after cancellation: %v", err)
	}
}

type runIDChild struct {
	cmd                  *exec.Cmd
	done                 chan error
	outputPath, checkPID string
	finished             bool
}

func startRunIDChild(t *testing.T, root, sha string) *runIDChild {
	t.Helper()
	dir := t.TempDir()
	out, err := os.Create(filepath.Join(dir, "child.log"))
	if err != nil {
		t.Fatal(err)
	}
	c := &runIDChild{done: make(chan error, 1), outputPath: out.Name(), checkPID: filepath.Join(dir, "check.pid")}
	c.cmd = exec.Command(os.Args[0], "-test.run=^TestLocalCheckLeaseRunChild$")
	if cli := os.Getenv("SDLC_RUNID_CLI"); cli != "" {
		c.cmd = exec.Command(cli, "local-check", "run", "--name", "suite", "--rev", sha)
	}
	processgroup.Configure(c.cmd)
	c.cmd.Dir = root
	c.cmd.WaitDelay = time.Second
	c.cmd.Env = append(os.Environ(), "SDLC_LEASE_ROOT="+root, "SDLC_LEASE_SHA="+sha, "SDLC_LEASE_CHECK_PID="+c.checkPID)
	c.cmd.Stdout, c.cmd.Stderr = out, out
	if err := c.cmd.Start(); err != nil {
		out.Close()
		t.Fatal(err)
	}
	go func() { c.done <- c.cmd.Wait() }()
	t.Cleanup(func() {
		checkPID, checkStarted := 0, ""
		if !c.finished {
			if raw, _ := os.ReadFile(c.checkPID); len(raw) > 0 {
				checkPID, _ = strconv.Atoi(strings.TrimSpace(string(raw)))
				if checkPID > 1 {
					checkStarted, _ = processStartTime(checkPID)
				}
			}
			_ = c.cmd.Process.Signal(syscall.SIGTERM)
			select {
			case <-c.done:
				c.finished = true
			case <-time.After(testfix.HangCeiling):
				_ = processgroup.Terminate(c.cmd, true)
				select {
				case <-c.done:
					c.finished = true
				case <-time.After(testfix.HangCeiling):
					t.Error("run-id child did not exit")
				}
			}
		}
		if checkPID > 1 && checkStarted != "" {
			if now, err := processStartTime(checkPID); err == nil && now == checkStarted {
				_ = syscall.Kill(checkPID, syscall.SIGKILL)
			}
		}
		out.Close()
	})
	return c
}

func (c *runIDChild) output(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(c.outputPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func assertRunIDChildAlive(t *testing.T, c *runIDChild) {
	t.Helper()
	select {
	case err := <-c.done:
		c.finished = true
		t.Fatalf("queued survivor exited unexpectedly: %v\n%s", err, c.output(t))
	default:
	}
	if err := c.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("run-id child is not alive: %v", err)
	}
}

func awaitRunIDCancelled(t *testing.T, c *runIDChild) {
	t.Helper()
	select {
	case err := <-c.done:
		c.finished = true
		if err == nil || !strings.Contains(c.output(t), "interrupted") {
			t.Fatalf("run did not take managed cancellation: %v\n%s", err, c.output(t))
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatalf("cancelled run did not unwind:\n%s", c.output(t))
	}
}

func awaitRunIDLease(t *testing.T, dir string, c *runIDChild, sha string) (string, string) {
	t.Helper()
	deadline := time.Now().Add(testfix.HangCeiling)
	for {
		assertRunIDChildAlive(t, c)
		for _, p := range leaseFiles(t, dir) {
			raw, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			var rec runLease
			if json.Unmarshal(raw, &rec) == nil && rec.PID == c.cmd.Process.Pid && rec.SHA == sha {
				assertRunIDHeld(t, p)
				return strings.TrimSuffix(filepath.Base(p), ".json"), p
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("run did not publish lease on %s:\n%s", sha, c.output(t))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func awaitRunIDTickets(t *testing.T, dir string, want int, children ...*runIDChild) []string {
	t.Helper()
	deadline := time.Now().Add(testfix.HangCeiling)
	for {
		for _, c := range children {
			assertRunIDChildAlive(t, c)
		}
		paths, err := filepath.Glob(filepath.Join(dir, "[0-9]*"))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) == want {
			ready := true
			for _, p := range paths {
				f, err := os.OpenFile(p, os.O_RDWR, 0)
				if err != nil {
					ready = false
					break
				}
				lockErr := flock.Exclusive(f, true)
				raw, readErr := os.ReadFile(p)
				var ticket struct {
					PID int `json:"pid"`
				}
				ready = ready && errors.Is(lockErr, syscall.EWOULDBLOCK) && readErr == nil && json.Unmarshal(raw, &ticket) == nil && ticket.PID > 1
				f.Close()
			}
			if ready {
				return paths
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("want %d live FIFO tickets; got %v", want, paths)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func assertRunIDHeld(t *testing.T, path string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := flock.Exclusive(f, true); !errors.Is(err, syscall.EWOULDBLOCK) {
		t.Fatalf("expected original live holder on %s: %v", path, err)
	}
}

// A genuine live process plus an already-held lease supplies an independent
// ownership refusal control without queuing a third suite or signalling the test.
func holdRunIDForeignLease(t *testing.T, dir, sha string) (string, int, <-chan error) {
	t.Helper()
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { defer close(done); done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		select {
		case <-done:
		case <-time.After(testfix.HangCeiling):
			t.Error("foreign process did not exit")
		}
	})
	started, err := processStartTime(cmd.Process.Pid)
	if err != nil || started == "" {
		t.Fatalf("foreign process identity: %q %v", started, err)
	}

	rec := runLease{Root: canonicalRoot(t.TempDir()), Check: "suite", SHA: sha, PID: cmd.Process.Pid, PIDStarted: started, Started: time.Now().UTC()}
	id, release, err := holdRunLease(dir, rec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)
	return id, cmd.Process.Pid, done
}

func TestDuplicateRunLeaseNoticeControls(t *testing.T) {
	current := runLease{Root: "/ours", Check: "suite", SHA: strings.Repeat("a", 40)}
	for _, tc := range []struct {
		name, root, check, sha string
		held, want             bool
	}{
		{"different head", "/ours", "suite", strings.Repeat("b", 40), true, true},
		{"same head", "/ours", "suite", current.SHA, true, false},
		{"other check", "/ours", "e2e", strings.Repeat("b", 40), true, false},
		{"foreign root", "/other", "suite", strings.Repeat("b", 40), true, false},
		{"stale lease", "/ours", "suite", strings.Repeat("b", 40), false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			rec := runLease{Root: tc.root, Check: tc.check, SHA: tc.sha}
			id, release, err := holdRunLease(dir, rec)
			if err != nil {
				t.Fatal(err)
			}
			if tc.held {
				t.Cleanup(release)
			} else {
				release()
				raw, err := json.Marshal(rec)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, id+".json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			var output bytes.Buffer
			noticeDuplicateRunLeases(dir, "new-run", current, &output, openRunLease)
			if got := strings.Contains(output.String(), "cancel --run "+id); got != tc.want {
				t.Fatalf("notice=%q; want duplicate=%v", output.String(), tc.want)
			}
			if tc.held {
				assertRunIDHeld(t, filepath.Join(dir, id+".json"))
			}
		})
	}
}

// One census read per live file at N and 2N rejects rescanning every lease
// for each sibling. Notices are a batch observation, with no subprocess per row.
func TestDuplicateRunLeaseNoticeLinearReads(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			dir := t.TempDir()
			current := runLease{Root: "/ours", Check: "suite", SHA: strings.Repeat("a", 40)}
			for i := 0; i < n; i++ {
				_, release, err := holdRunLease(dir, current)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(release)
			}
			reads := 0
			read := func(p string) (*os.File, runLease, bool, error) { reads++; return openRunLease(p) }
			var output bytes.Buffer
			noticeDuplicateRunLeases(dir, "new-run", current, &output, read)
			if reads != n || output.Len() != 0 {
				t.Fatalf("%d leases: reads=%d notice=%s", n, reads, output.String())
			}
		})
	}
}
