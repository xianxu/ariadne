package landingline

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// grantBox stands in for the grants log: the test changes it while waiters
// read it on every poll.
type grantBox struct {
	mu sync.Mutex
	m  map[int]Grant
}

func (b *grantBox) set(m map[int]Grant) { b.mu.Lock(); b.m = m; b.mu.Unlock() }
func (b *grantBox) get() map[int]Grant  { b.mu.Lock(); defer b.mu.Unlock(); return b.m }

// fastOptions polls every millisecond; grants (nil: none) is read on every
// poll, like the production log.
func fastOptions(grants *grantBox) Options {
	o := Options{Poll: time.Millisecond}
	if grants != nil {
		o.Grants = grants.get
	}
	return o
}

func openLine(t *testing.T, dir string, o Options) *Line {
	t.Helper()
	l, err := Open(dir, o)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func enqueue(t *testing.T, l *Line, issue int) *Handle {
	t.Helper()
	h, err := l.Enqueue(Ticket{Issue: issue, Branch: "b"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.Release)
	return h
}

// admitAsync waits for h in a goroutine and reports on the channel.
func admitAsync(l *Line, h *Handle) <-chan error {
	done := make(chan error, 1)
	go func() { done <- l.Wait(context.Background(), h) }()
	return done
}

func admitted(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("not admitted")
	}
}

func notAdmitted(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("admitted early (err %v)", err)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestLineOneHeadAtATime(t *testing.T) {
	l := openLine(t, t.TempDir(), fastOptions(nil))
	a, b := enqueue(t, l, 1), enqueue(t, l, 2)
	admitted(t, admitAsync(l, a))
	second := admitAsync(l, b)
	notAdmitted(t, second)
	head, waiting, err := l.Status()
	if err != nil || head == nil || head.Seq != a.Ticket().Seq || len(waiting) != 1 {
		t.Fatalf("status head %v waiting %v err %v", head, waiting, err)
	}
	a.Release()
	admitted(t, second)
}

func TestLineFIFOAndSeqOrder(t *testing.T) {
	l := openLine(t, t.TempDir(), fastOptions(nil))
	first := enqueue(t, l, 1)
	second := enqueue(t, l, 2)
	if second.Ticket().Seq <= first.Ticket().Seq {
		t.Fatalf("seq %d after %d", second.Ticket().Seq, first.Ticket().Seq)
	}
	// The younger ticket waits even when it asks first.
	younger := admitAsync(l, second)
	notAdmitted(t, younger)
	admitted(t, admitAsync(l, first))
	first.Release()
	admitted(t, younger)
}

func TestLineGrantReordersButNeverPreempts(t *testing.T) {
	grants := &grantBox{}
	l := openLine(t, t.TempDir(), fastOptions(grants))
	head := enqueue(t, l, 1)
	admitted(t, admitAsync(l, head))
	normal, blocking := enqueue(t, l, 2), enqueue(t, l, 3)
	n, b := admitAsync(l, normal), admitAsync(l, blocking)
	grants.set(map[int]Grant{3: grant(3, time.Now(), 4)})
	// The granted ticket waits for the running head: no preemption.
	notAdmitted(t, b)
	head.Release()
	admitted(t, b)
	notAdmitted(t, n)
	blocking.Release()
	admitted(t, n)
}

func TestLineIndependentDirectories(t *testing.T) {
	a := openLine(t, t.TempDir(), fastOptions(nil))
	b := openLine(t, t.TempDir(), fastOptions(nil))
	ha, hb := enqueue(t, a, 1), enqueue(t, b, 1)
	admitted(t, admitAsync(a, ha))
	admitted(t, admitAsync(b, hb))
}

func TestLineCancelledWaitLeavesTheLine(t *testing.T) {
	l := openLine(t, t.TempDir(), fastOptions(nil))
	head := enqueue(t, l, 1)
	admitted(t, admitAsync(l, head))
	waiter := enqueue(t, l, 2)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- l.Wait(ctx, waiter) }()
	cancel()
	if err := <-done; err == nil {
		t.Fatal("cancelled wait returned nil")
	}
	waiter.Release()
	if _, waiting, _ := l.Status(); len(waiting) != 0 {
		t.Fatalf("released waiter still listed: %v", waiting)
	}
}

// The helper process holds a ticket (or the head) until it is killed, so the
// kernel, not cleanup code, releases its flocks.
func TestHelperHoldTicket(t *testing.T) {
	dir := os.Getenv("LANDINGLINE_HELPER_DIR")
	if dir == "" {
		t.Skip("helper process only")
	}
	l, err := Open(dir, Options{Poll: 5 * time.Millisecond})
	if err != nil {
		os.Exit(2)
	}
	h, err := l.Enqueue(Ticket{Issue: 99, Branch: "helper"})
	if err != nil {
		os.Exit(2)
	}
	if os.Getenv("LANDINGLINE_HELPER_MODE") == "head" {
		if err := l.Wait(context.Background(), h); err != nil {
			os.Exit(2)
		}
	}
	os.Stdout.WriteString("ready\n")
	time.Sleep(time.Hour) // until killed; select{} would end it as a deadlock
	h.Release()
}

func startHelper(t *testing.T, dir, mode string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldTicket$")
	cmd.Env = append(os.Environ(), "LANDINGLINE_HELPER_DIR="+dir, "LANDINGLINE_HELPER_MODE="+mode)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil || line != "ready\n" {
		t.Fatalf("helper not ready: %q %v", line, err)
	}
	return cmd
}

func kill(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if err := cmd.Process.Signal(syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
}

func TestLineDeadWaiterIsPruned(t *testing.T) {
	dir := t.TempDir()
	helper := startHelper(t, dir, "ticket")
	l := openLine(t, dir, fastOptions(nil))
	mine := enqueue(t, l, 1)
	// The helper's ticket is older, so it is next while it lives.
	done := admitAsync(l, mine)
	notAdmitted(t, done)
	kill(t, helper)
	admitted(t, done)
	if left, _ := filepath.Glob(filepath.Join(dir, "tickets", "*")); len(left) != 0 {
		t.Fatalf("tickets left after admission: %v", left)
	}
}

func TestLineDeadHeadReleasesTheLine(t *testing.T) {
	dir := t.TempDir()
	helper := startHelper(t, dir, "head")
	l := openLine(t, dir, fastOptions(nil))
	if head, _, err := l.Status(); err != nil || head == nil || head.Issue != 99 {
		t.Fatalf("helper is not the head: %v %v", head, err)
	}
	done := admitAsync(l, enqueue(t, l, 1))
	notAdmitted(t, done)
	kill(t, helper)
	admitted(t, done)
}

func TestLineAdmitHookReportsDisplacement(t *testing.T) {
	grants := &grantBox{}
	var gotHead Ticket
	var gotDisplaced []Ticket
	o := fastOptions(grants)
	o.OnAdmit = func(head Ticket, displaced []Ticket) { gotHead, gotDisplaced = head, displaced }
	l := openLine(t, t.TempDir(), o)
	busy := enqueue(t, l, 1)
	admitted(t, admitAsync(l, busy))
	older, granted := enqueue(t, l, 2), enqueue(t, l, 3)
	grants.set(map[int]Grant{3: grant(3, time.Now(), 4)})
	busy.Release()
	admitted(t, admitAsync(l, granted))
	if gotHead.Issue != 3 || len(gotDisplaced) != 1 || gotDisplaced[0].Seq != older.Ticket().Seq {
		t.Fatalf("admit hook head %v displaced %v", gotHead, gotDisplaced)
	}
}
