package landingline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

// A line directory holds:
//
//	line.lock        the allocation mutex: every scan and admission runs under it
//	head.lock        the admission token, flock-held by the admitted head; it
//	                 holds the head's ticket as JSON while held
//	seq              the next ticket number
//	tickets/<seq>    one per waiting merge, flock-held by that merge
//
// Every flock dies with its process, so a dead waiter's ticket is pruned by
// the next scan and a dead head frees head.lock: the line keeps no state that
// a crash can leave inconsistent. A process is either waiting (it holds its
// ticket) or the head (it holds head.lock and its ticket is gone).

// Options are the line's injected clock, sleep and grants; zero values take
// the production defaults.
type Options struct {
	Now    func() time.Time
	Sleep  func(context.Context, time.Duration) error
	Poll   time.Duration
	Grants func() map[int]Grant
	// OnWait reports a waiter's position (1 = next) when it changes.
	OnWait func(position int, head *Ticket)
	// OnAdmit reports an admission and the older tickets it went ahead of.
	OnAdmit func(head Ticket, displaced []Ticket)
}

// Line is one repository's landing line.
type Line struct {
	dir string
	o   Options
}

// Handle is one merge's place in the line.
type Handle struct {
	ticket Ticket
	path   string
	file   *os.File // the ticket's flock while waiting
	head   *os.File // head.lock's flock once admitted
}

func (h *Handle) Ticket() Ticket { return h.ticket }

// Open prepares dir as a line; it creates nothing but directories.
func Open(dir string, o Options) (*Line, error) {
	if err := os.MkdirAll(filepath.Join(dir, "tickets"), 0o755); err != nil {
		return nil, err
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Sleep == nil {
		o.Sleep = sleepCtx
	}
	if o.Poll <= 0 {
		o.Poll = 2 * time.Second
	}
	if o.Grants == nil {
		o.Grants = func() map[int]Grant { return nil }
	}
	return &Line{dir: dir, o: o}, nil
}

// Enqueue takes the next ticket. The ticket file appears already locked: it is
// written and locked under a temporary name, then renamed into place.
func (l *Line) Enqueue(t Ticket) (*Handle, error) {
	var h *Handle
	err := l.withMutex(func() error {
		seq, err := l.nextSeq()
		if err != nil {
			return err
		}
		t.Seq, t.Enqueued = seq, l.o.Now()
		if t.PID == 0 {
			t.PID = os.Getpid()
		}
		data, err := json.Marshal(t)
		if err != nil {
			return err
		}
		f, err := os.CreateTemp(filepath.Join(l.dir, "tickets"), ".tmp-")
		if err != nil {
			return err
		}
		if err := flock.Exclusive(f, true); err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}
		path := filepath.Join(l.dir, "tickets", fmt.Sprintf("%020d", seq))
		if _, err := f.Write(data); err == nil {
			err = os.Rename(f.Name(), path)
		}
		if err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}
		h = &Handle{ticket: t, path: path, file: f}
		return nil
	})
	return h, err
}

// Wait returns once h is the head: head.lock is free and Order puts h first
// among the live tickets. Priority is read on every poll, so a grant made
// while h waits counts at the next admission; a running head is never
// interrupted. A cancelled ctx returns its error; the caller still releases.
func (l *Line) Wait(ctx context.Context, h *Handle) error {
	last := -1
	for {
		pos, head, err := l.tryAdmit(h)
		if err != nil {
			return err
		}
		if pos == 0 {
			return nil
		}
		if pos != last && l.o.OnWait != nil {
			l.o.OnWait(pos, head)
		}
		last = pos
		if err := l.o.Sleep(ctx, l.o.Poll); err != nil {
			return err
		}
	}
}

// tryAdmit admits h (position 0) or reports its position and the head.
func (l *Line) tryAdmit(h *Handle) (int, *Ticket, error) {
	pos := 0
	var current *Ticket
	err := l.withMutex(func() error {
		live, err := l.scan()
		if err != nil {
			return err
		}
		grants := l.o.Grants()
		order := Order(live, grants)
		pos = 1
		for _, t := range order {
			if t.Seq == h.ticket.Seq {
				break
			}
			pos++
		}
		head, err := os.OpenFile(filepath.Join(l.dir, "head.lock"), os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		if err := flock.Exclusive(head, true); err != nil {
			current = readTicket(head)
			head.Close()
			if errors.Is(err, syscall.EWOULDBLOCK) {
				return nil
			}
			return err
		}
		next, displaced, ok := NextHead(live, grants)
		if !ok || next.Seq != h.ticket.Seq {
			head.Close()
			return nil
		}
		data, _ := json.Marshal(h.ticket)
		if err := head.Truncate(0); err == nil {
			_, err = head.WriteAt(data, 0)
		}
		if err != nil {
			head.Close()
			return err
		}
		h.head = head
		h.dropTicket()
		pos = 0
		if l.o.OnAdmit != nil {
			l.o.OnAdmit(h.ticket, displaced)
		}
		return nil
	})
	return pos, current, err
}

// Status reports the admitted head (nil when none) and the waiting tickets in
// admission order.
func (l *Line) Status() (*Ticket, []Ticket, error) {
	var head *Ticket
	var waiting []Ticket
	err := l.withMutex(func() error {
		live, err := l.scan()
		if err != nil {
			return err
		}
		waiting = Order(live, l.o.Grants())
		f, err := os.OpenFile(filepath.Join(l.dir, "head.lock"), os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := flock.Exclusive(f, true); errors.Is(err, syscall.EWOULDBLOCK) {
			head = readTicket(f)
		} else if err != nil {
			return err
		}
		return nil
	})
	return head, waiting, err
}

// Release leaves the line, as head or waiter; it is idempotent.
func (h *Handle) Release() {
	if h.head != nil {
		h.head.Close()
		h.head = nil
	}
	h.dropTicket()
}

func (h *Handle) dropTicket() {
	if h.file != nil {
		os.Remove(h.path)
		h.file.Close()
		h.file = nil
	}
}

// scan returns the live tickets, removing any a dead process left: a ticket
// file nobody holds a lock on. A held ticket whose JSON doesn't parse still
// counts, with its number from its name, so it is never silently skipped.
func (l *Line) scan() ([]Ticket, error) {
	dir := filepath.Join(l.dir, "tickets")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var live []Ticket
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			continue // removed meanwhile
		}
		err = flock.Exclusive(f, true)
		if err == nil {
			os.Remove(path) // its owner is gone
			f.Close()
			continue
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			f.Close()
			return nil, err
		}
		if strings.HasPrefix(e.Name(), ".tmp-") {
			f.Close()
			continue // a temporary file is never a ticket
		}
		t := readTicket(f)
		f.Close()
		seq, perr := strconv.ParseInt(e.Name(), 10, 64)
		if t == nil {
			if perr != nil {
				continue
			}
			t = &Ticket{Seq: seq}
		}
		live = append(live, *t)
	}
	return live, nil
}

func (l *Line) nextSeq() (int64, error) {
	path := filepath.Join(l.dir, "seq")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	n++
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.FormatInt(n, 10)+"\n"), 0o644); err != nil {
		return 0, err
	}
	return n, os.Rename(tmp, path)
}

func (l *Line) withMutex(fn func() error) error {
	f, err := os.OpenFile(filepath.Join(l.dir, "line.lock"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := flock.Exclusive(f, false); err != nil {
		return err
	}
	return fn()
}

func readTicket(f *os.File) *Ticket {
	data := make([]byte, 4096)
	n, _ := f.ReadAt(data, 0)
	var t Ticket
	if n == 0 || json.Unmarshal(data[:n], &t) != nil {
		return nil
	}
	return &t
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
