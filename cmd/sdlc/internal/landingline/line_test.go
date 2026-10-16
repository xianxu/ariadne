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

// grantBox stands in for the grants log, which waiters read on every poll.
type grantBox struct {
	mu sync.Mutex
	m  map[int]Grant
}

func (b *grantBox) set(m map[int]Grant) { b.mu.Lock(); b.m = m; b.mu.Unlock() }
func (b *grantBox) get() map[int]Grant  { b.mu.Lock(); defer b.mu.Unlock(); return b.m }

func testOptions(grants *grantBox) Options {
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

// step is one admission attempt, the unit Wait repeats: the tests drive it
// directly, so each ordering is chosen by the test, not by the scheduler.
func step(t *testing.T, l *Line, h *Handle) int {
	t.Helper()
	pos, _, err := l.tryAdmit(h)
	if err != nil {
		t.Fatal(err)
	}
	return pos
}

func wantPos(t *testing.T, l *Line, h *Handle, want int) {
	t.Helper()
	if got := step(t, l, h); got != want {
		t.Fatalf("ticket %d: position %d, want %d (0 = admitted)", h.Ticket().Seq, got, want)
	}
}

func TestLineOneHeadAtATime(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	a, b := enqueue(t, l, 1), enqueue(t, l, 2)
	wantPos(t, l, a, 0)
	wantPos(t, l, b, 1) // next, but the head holds the line
	head, waiting, err := l.Status()
	if err != nil || head == nil || head.Seq != a.Ticket().Seq || len(waiting) != 1 {
		t.Fatalf("status head %v waiting %v err %v", head, waiting, err)
	}
	a.Release()
	wantPos(t, l, b, 0)
}

func TestLineFIFOAndSeqOrder(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	first, second := enqueue(t, l, 1), enqueue(t, l, 2)
	if second.Ticket().Seq <= first.Ticket().Seq {
		t.Fatalf("seq %d after %d", second.Ticket().Seq, first.Ticket().Seq)
	}
	wantPos(t, l, second, 2) // asking first doesn't jump the line
	wantPos(t, l, first, 0)
	first.Release()
	wantPos(t, l, second, 0)
}

func TestLineGrantReordersButNeverPreempts(t *testing.T) {
	grants := &grantBox{}
	l := openLine(t, t.TempDir(), testOptions(grants))
	head := enqueue(t, l, 1)
	wantPos(t, l, head, 0)
	normal, blocking := enqueue(t, l, 2), enqueue(t, l, 3)
	wantPos(t, l, blocking, 2)
	grants.set(map[int]Grant{3: grant(3, time.Now(), 4)})
	wantPos(t, l, blocking, 1) // first in line, yet the running head keeps it
	wantPos(t, l, normal, 2)
	head.Release()
	wantPos(t, l, normal, 2) // the granted ticket goes first
	wantPos(t, l, blocking, 0)
	blocking.Release()
	wantPos(t, l, normal, 0)
}

func TestLineIndependentDirectories(t *testing.T) {
	a := openLine(t, t.TempDir(), testOptions(nil))
	b := openLine(t, t.TempDir(), testOptions(nil))
	wantPos(t, a, enqueue(t, a, 1), 0)
	wantPos(t, b, enqueue(t, b, 1), 0)
}

func TestLineReleasedTicketCantWait(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	h := enqueue(t, l, 1)
	h.Release()
	if err := l.Wait(context.Background(), h); err == nil {
		t.Fatal("a released ticket waited")
	}
}

func TestLineTicketNumberComesFromItsName(t *testing.T) {
	l := openLine(t, t.TempDir(), testOptions(nil))
	h := enqueue(t, l, 1)
	// Rewrite the held ticket's JSON with a lie, then with garbage: the name
	// still decides its number, and a held ticket that doesn't parse still counts.
	for _, body := range []string{`{"seq": 999, "issue": 1}`, `{not json`} {
		if err := os.WriteFile(h.path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		_, waiting, err := l.Status()
		if err != nil || len(waiting) != 1 || waiting[0].Seq != h.Ticket().Seq {
			t.Fatalf("%s: waiting %v err %v", body, waiting, err)
		}
	}
}

// Wait is the poll loop over step: it reports position changes and admits once
// the head leaves. Its sleep is injected, so the head's release happens at a
// chosen poll rather than after real time passes.
func TestLineWaitAdmitsAfterTheHeadReleases(t *testing.T) {
	var positions []int
	o := testOptions(nil)
	o.OnWait = func(pos int, _ *Ticket) { positions = append(positions, pos) }
	var head *Handle
	polls := 0
	o.Sleep = func(ctx context.Context, d time.Duration) error {
		polls++
		if polls == 2 {
			head.Release()
		}
		return nil
	}
	l := openLine(t, t.TempDir(), o)
	head = enqueue(t, l, 1)
	wantPos(t, l, head, 0)
	if err := l.Wait(context.Background(), enqueue(t, l, 2)); err != nil {
		t.Fatal(err)
	}
	if polls != 2 || len(positions) != 1 || positions[0] != 1 {
		t.Fatalf("polls %d positions %v", polls, positions)
	}
}

func TestLineCancelledWaitLeavesTheLine(t *testing.T) {
	o := testOptions(nil)
	o.Sleep = func(ctx context.Context, d time.Duration) error { return context.Canceled }
	l := openLine(t, t.TempDir(), o)
	wantPos(t, l, enqueue(t, l, 1), 0)
	waiter := enqueue(t, l, 2)
	if err := l.Wait(context.Background(), waiter); err == nil {
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
	l := openLine(t, dir, testOptions(nil))
	mine := enqueue(t, l, 1)
	wantPos(t, l, mine, 2) // the live helper's older ticket is next
	kill(t, helper)
	wantPos(t, l, mine, 0)
	if left, _ := filepath.Glob(filepath.Join(dir, "tickets", "*")); len(left) != 0 {
		t.Fatalf("tickets left after admission: %v", left)
	}
}

func TestLineDeadHeadReleasesTheLine(t *testing.T) {
	dir := t.TempDir()
	helper := startHelper(t, dir, "head")
	l := openLine(t, dir, testOptions(nil))
	if head, _, err := l.Status(); err != nil || head == nil || head.Issue != 99 {
		t.Fatalf("helper is not the head: %v %v", head, err)
	}
	mine := enqueue(t, l, 1)
	wantPos(t, l, mine, 1)
	kill(t, helper)
	wantPos(t, l, mine, 0)
}

func TestLineAdmitHookReportsDisplacement(t *testing.T) {
	grants := &grantBox{}
	var gotHead Ticket
	var gotDisplaced []Ticket
	o := testOptions(grants)
	o.OnAdmit = func(head Ticket, displaced []Ticket) { gotHead, gotDisplaced = head, displaced }
	l := openLine(t, t.TempDir(), o)
	busy := enqueue(t, l, 1)
	wantPos(t, l, busy, 0)
	older, granted := enqueue(t, l, 2), enqueue(t, l, 3)
	grants.set(map[int]Grant{3: grant(3, time.Now(), 4)})
	busy.Release()
	wantPos(t, l, granted, 0)
	if gotHead.Issue != 3 || len(gotDisplaced) != 1 || gotDisplaced[0].Seq != older.Ticket().Seq {
		t.Fatalf("admit hook head %v displaced %v", gotHead, gotDisplaced)
	}
}
