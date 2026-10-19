package landingline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
//	ready/<seq>      a waiting merge's readiness vouch (#449); its own
//	                 directory, since scan deletes every unlocked file in
//	                 tickets/. Swept once tickets/<seq> is gone.
//	outcomes/<seq>   what a train did with a waiting car (#449), written by
//	                 the train's conductor; swept the same way.
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
	// OnAdmit reports an admission and the older tickets it went ahead of. It
	// runs outside the line mutex, so it may call Record.
	OnAdmit func(head Ticket, displaced []Ticket)
	// Admission (#449) is called once per admission attempt, before the line
	// mutex is taken, so its slow inputs (the train setting) are read outside
	// it. It returns the rule for this attempt, which the line applies under
	// the mutex, after taking head.lock, to each live ticket with the events
	// log read there: the departure ledger is decided on what the previous
	// head wrote before releasing head.lock, never on a read from before it.
	// The rule says whether the ticket may be admitted now, given its
	// readiness vouch (nil when it has none) and whether it holds a grant.
	// The head is the first admissible ticket in Order; nil admits every
	// ticket, the serial line of #393.
	Admission func() AdmissionRule
	// OnHeld reports why this waiter, though first in line, is not admitted
	// (a train not due, not ready), when that changes.
	OnHeld func(why string)
	// OnOutcome reports a train's word on this waiter (#449) when it changes.
	OnOutcome func(o Outcome)
}

// Line is one repository's landing line.
type Line struct {
	dir string
	o   Options
}

// AdmissionRule decides one ticket's admission from the events log as read
// under the line mutex.
type AdmissionRule func(events []Event, t Ticket, r *Readiness, granted bool) (bool, string)

// ErrNotWaiting is a vouch from a merge that is no longer waiting (admitted
// or released): not a failure, the vouching simply ends.
var ErrNotWaiting = errors.New("landing line: only a waiting merge vouches")

// Handle is one merge's place in the line. Its files are guarded by mu: a
// merge's vouching goroutine writes through it while Wait admits it.
type Handle struct {
	mu     sync.Mutex
	ticket Ticket
	path   string
	ready  string   // its readiness sidecar's path
	out    string   // its outcome sidecar's path
	file   *os.File // the ticket's flock while waiting
	head   *os.File // head.lock's flock once admitted
}

func (h *Handle) Ticket() Ticket { return h.ticket }

// Open prepares dir as a line; it creates nothing but directories.
func Open(dir string, o Options) (*Line, error) {
	for _, sub := range []string{"tickets", "ready", "outcomes"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return nil, err
		}
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
		if _, err := os.Lstat(path); err == nil {
			err = fmt.Errorf("landing line: ticket %d already exists; the counter in %s went backwards", seq, filepath.Join(l.dir, "seq"))
			f.Close()
			os.Remove(f.Name())
			return err
		}
		if _, err := f.Write(data); err == nil {
			err = os.Rename(f.Name(), path)
		}
		if err != nil {
			f.Close()
			os.Remove(f.Name())
			return err
		}
		h = &Handle{ticket: t, path: path, ready: readyPath(l.dir, seq), out: outcomePath(l.dir, seq), file: f}
		return nil
	})
	return h, err
}

// Wait returns once h is the head: head.lock is free and Order puts h first
// among the live tickets. Priority is read on every poll, so a grant made
// while h waits counts at the next admission; a running head is never
// interrupted. A cancelled ctx returns its error; the caller still releases.
func (l *Line) Wait(ctx context.Context, h *Handle) error {
	last, lastWhy := -1, ""
	var lastOutcome OutcomeKind
	var lastTrain int64
	for {
		// A train's word comes first: a landed or refused car leaves the
		// line, a released one waits again from no outcome.
		if o := l.outcomeOf(h.ticket.Seq); o != nil {
			if (o.Kind != lastOutcome || o.Train != lastTrain) && l.o.OnOutcome != nil {
				l.o.OnOutcome(*o)
			}
			lastOutcome, lastTrain = o.Kind, o.Train
			switch {
			case o.Final():
				return &Settled{Outcome: *o}
			case o.Kind == OutcomeReleased:
				if err := l.clearOutcome(h.ticket.Seq, OutcomeReleased); err != nil {
					return err
				}
			}
		}
		pos, head, why, err := l.tryAdmit(h)
		if err != nil {
			return err
		}
		if pos == 0 {
			return nil
		}
		if pos != last && l.o.OnWait != nil {
			l.o.OnWait(pos, head)
		}
		if why != lastWhy && why != "" && l.o.OnHeld != nil {
			l.o.OnHeld(why)
		}
		last, lastWhy = pos, why
		if err := l.o.Sleep(ctx, l.o.Poll); err != nil {
			return err
		}
	}
}

// tryAdmit admits h (position 0) or reports its position, the head, and why
// h is not admitted though first in line.
func (l *Line) tryAdmit(h *Handle) (int, *Ticket, string, error) {
	pos := 0
	var current *Ticket
	var displaced []Ticket
	why := ""
	// The rule's inputs (the schedule, the events log) are read here, before
	// the line mutex: no slow read runs while every waiter is blocked on it.
	var admissible AdmissionRule
	if l.o.Admission != nil {
		admissible = l.o.Admission()
	}
	err := l.withMutex(func() error {
		live, err := l.scan()
		if err != nil {
			return err
		}
		if !h.waiting() {
			return errors.New("landing line: this ticket was released")
		}
		grants := l.o.Grants()
		order := Order(live, grants)
		pos = 0
		for i, t := range order {
			if t.Seq == h.ticket.Seq {
				pos = i + 1
			}
		}
		if pos == 0 {
			return errors.New("landing line: this ticket is no longer in line")
		}
		head, err := os.OpenFile(filepath.Join(l.dir, "head.lock"), os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		if err := flock.Exclusive(head, true); err != nil {
			current = heldTicket(head)
			head.Close()
			if errors.Is(err, syscall.EWOULDBLOCK) {
				return nil
			}
			return err
		}
		candidates := live
		if admissible != nil {
			// head.lock is held: the previous head has left, so its train
			// event is in the log this reads.
			events, _, err := l.Events()
			if err != nil {
				events = nil // unreadable: the first train, never a stall
			}
			ready := l.readReady()
			candidates = nil
			for _, t := range live {
				var r *Readiness
				if v, ok := ready[t.Seq]; ok {
					r = &v
				}
				_, granted := grants[t.Issue]
				ok, reason := admissible(events, t, r, granted)
				if ok {
					candidates = append(candidates, t)
				} else if t.Seq == h.ticket.Seq {
					why = reason
				}
			}
		}
		next, ok := NextHead(candidates, grants)
		if !ok || next.Seq != h.ticket.Seq {
			head.Close()
			return nil
		}
		passed := Passed(live, grants, next)
		data, _ := json.Marshal(h.ticket)
		if err := head.Truncate(0); err == nil {
			_, err = head.WriteAt(data, 0)
		}
		if err != nil {
			head.Close()
			return err
		}
		h.mu.Lock()
		h.head = head
		h.dropLocked()
		h.mu.Unlock()
		pos, displaced = 0, passed
		return nil
	})
	// After the mutex: the hook may log, which takes the mutex itself (flock is
	// per open file, so a nested take in one process would deadlock).
	if err == nil && pos == 0 && l.o.OnAdmit != nil {
		l.o.OnAdmit(h.ticket, displaced)
	}
	return pos, current, why, err
}

// Vouch records a waiting merge's readiness (#449), replacing its last one.
// It is written under the mutex, by temporary file then rename, so a reader
// sees a whole vouch or none.
func (l *Line) Vouch(h *Handle, r Readiness) error {
	r.Version, r.Seq = ReadinessVersion, h.ticket.Seq
	if r.At.IsZero() {
		r.At = l.o.Now()
	}
	if err := r.Valid(h.ticket.Seq); err != nil {
		return err
	}
	return l.withMutex(func() error {
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.file == nil {
			return ErrNotWaiting
		}
		// A later sdlc may have written members this one doesn't know: keep
		// them (ARCH-EVOLVE). A previous vouch that doesn't parse keeps nothing.
		if old, err := os.ReadFile(h.ready); err == nil && len(old) <= maxTicket {
			if prev, err := DecodeReadiness(old); err == nil && prev.Seq == h.ticket.Seq {
				r.Extra = mergeExtra(prev.Extra, r.Extra)
			}
		}
		return writeAtomic(filepath.Join(l.dir, "ready"), h.ready, r)
	})
}

// Readiness returns the live tickets' vouches. A sidecar that doesn't parse,
// or doesn't match its name, is no vouch: never read as ready.
func (l *Line) Readiness() (map[int64]Readiness, error) {
	var out map[int64]Readiness
	err := l.withMutex(func() error {
		if _, err := l.scan(); err != nil {
			return err
		}
		out = l.readReady()
		return nil
	})
	return out, err
}

// readReady reads ready/ under the mutex.
func (l *Line) readReady() map[int64]Readiness {
	out := map[int64]Readiness{}
	entries, err := os.ReadDir(filepath.Join(l.dir, "ready"))
	if err != nil {
		return out
	}
	for _, e := range entries {
		seq, ok := sidecarSeq(e.Name())
		if !ok {
			continue
		}
		data, err := os.ReadFile(filepath.Join(l.dir, "ready", e.Name()))
		if err != nil || len(data) > maxTicket {
			continue
		}
		r, err := DecodeReadiness(data)
		if err != nil || r.Valid(seq) != nil {
			continue
		}
		out[seq] = r
	}
	return out
}

func readyPath(dir string, seq int64) string {
	return filepath.Join(dir, "ready", fmt.Sprintf("%020d", seq))
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
			head = heldTicket(f)
		} else if err != nil {
			return err
		}
		return nil
	})
	return head, waiting, err
}

// Release leaves the line, as head or waiter; it is idempotent.
func (h *Handle) Release() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.head != nil {
		h.head.Close()
		h.head = nil
	}
	h.dropLocked()
}

func (h *Handle) waiting() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.file != nil
}

// dropLocked removes the ticket and its vouch; the caller holds h.mu.
func (h *Handle) dropLocked() {
	if h.file != nil {
		os.Remove(h.path)
		os.Remove(h.ready)
		os.Remove(h.out)
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
		// The file name is the ticket's number; its JSON can't move it in line.
		seq, perr := strconv.ParseInt(e.Name(), 10, 64)
		if perr != nil {
			continue // not a ticket name: written by something else
		}
		if t == nil {
			t = &Ticket{Unreadable: true}
		}
		t.Seq = seq
		live = append(live, *t)
	}
	l.sweepSidecars(live)
	return live, nil
}

// sweepSidecars removes the vouches and outcomes whose ticket is gone: a
// dead waiter's, or one a crash left between dropping a ticket and removing
// its sidecars.
func (l *Line) sweepSidecars(live []Ticket) {
	alive := map[int64]bool{}
	for _, t := range live {
		alive[t.Seq] = true
	}
	for _, sub := range []string{"ready", "outcomes"} {
		dir := filepath.Join(l.dir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			seq, ok := sidecarSeq(e.Name())
			if !ok && !strings.HasPrefix(e.Name(), ".tmp-") {
				continue // not ours
			}
			if ok && alive[seq] {
				continue
			}
			// A temporary file outlives only a crash mid-write: writers hold
			// the mutex, so none is in flight while scan runs.
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

func (l *Line) nextSeq() (int64, error) {
	path := filepath.Join(l.dir, "seq")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	var n int64
	if text := strings.TrimSpace(string(data)); text != "" {
		if n, err = strconv.ParseInt(text, 10, 64); err != nil || n < 0 {
			return 0, fmt.Errorf("landing line: %s is not a ticket counter (%q); fix or remove it", path, text)
		}
	}
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

// heldTicket reads a held lock's ticket; a held lock whose contents don't parse
// is still a head, so it reads as present but unreadable.
func heldTicket(f *os.File) *Ticket {
	if t := readTicket(f); t != nil {
		return t
	}
	return &Ticket{Unreadable: true}
}

// maxTicket bounds a ticket read; a ticket is one small JSON object, so a file
// past this is not one and reads as unreadable.
const maxTicket = 1 << 20

func readTicket(f *os.File) *Ticket {
	data, err := io.ReadAll(io.NewSectionReader(f, 0, maxTicket))
	var t Ticket
	if err != nil || len(data) == 0 || json.Unmarshal(data, &t) != nil {
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
