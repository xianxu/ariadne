package landingline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"syscall"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
)

const trainControlFilename = "train-control.json"

// Missing state is distinct from unreadable state: legacy serial admission
// works without a record, but cannot bypass a corrupt safety decision.
func (l *Line) readTrainControlLocked() (*TrainControl, error) {
	f, err := os.Open(filepath.Join(l.dir, trainControlFilename))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxTicket+1))
	if err != nil {
		return nil, err
	}
	c, err := DecodeTrainControl(data)
	if err != nil {
		return nil, fmt.Errorf("landing train control unavailable; inspect %s: %w", f.Name(), err)
	}
	return &c, nil
}
func (l *Line) writeTrainControlLocked(c TrainControl) error {
	if err := validateTrainControl(c); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if len(data) > maxTicket {
		return errors.New("train control capacity exceeded; existing exclusions retained")
	}
	return writeAtomic(l.dir, filepath.Join(l.dir, trainControlFilename), c)
}

// Validate flock lifetime, never merely a PID. Called with the short line
// mutex; no network or repository lock is taken.
func (l *Line) headGenerationLocked() (*Ticket, error) {
	f, err := os.OpenFile(filepath.Join(l.dir, "head.lock"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := flock.Exclusive(f, true); errors.Is(err, syscall.EWOULDBLOCK) {
		return heldTicket(f), nil
	} else if err != nil {
		return nil, err
	}
	return nil, nil
}
func (l *Line) requireHeadLocked(h *Handle) error {
	if h == nil {
		return ErrTrainControlStale
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.head == nil {
		return ErrTrainControlStale
	}
	tk := heldTicket(h.head)
	if tk.Unreadable || tk.Seq != h.ticket.Seq || tk.Branch != h.ticket.Branch {
		return ErrTrainControlStale
	}
	return nil
}
func (l *Line) ownTrainLocked(h *Handle) (*TrainControl, error) {
	if err := l.requireHeadLocked(h); err != nil {
		return nil, err
	}
	c, err := l.readTrainControlLocked()
	if err != nil {
		return nil, err
	}
	if c.Train != h.ticket.Seq {
		return nil, ErrTrainControlStale
	}
	return c, nil
}

// BeginTrain replaces only a previous generation whose external cleanup is
// confirmed. Live dropped tickets carry their exclusions into the new record.
func (l *Line) BeginTrain(h *Handle, c TrainControl) error {
	return l.withMutex(func() error {
		if err := l.requireHeadLocked(h); err != nil {
			return err
		}
		if c.Train != h.ticket.Seq {
			return ErrTrainControlStale
		}
		live, err := l.scan()
		if err != nil {
			return err
		}
		if old, err := l.readTrainControlLocked(); err == nil {
			if old.Train == c.Train {
				return errors.New("train control already registered; update its current generation")
			}
			if old.Cleanup.Pending {
				return ErrTrainCleanupPending
			}
			kept, err := l.reconcileTrainDropsLocked(old, live)
			if err != nil {
				return err
			}
			c.Drops = kept
			c.Extra = mergeExtra(old.Extra, c.Extra)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		bySeq := map[int64]Ticket{}
		for _, t := range live {
			bySeq[t.Seq] = t
		}
		bySeq[h.ticket.Seq] = h.ticket
		excluded := map[int64]bool{}
		for _, d := range c.Drops {
			excluded[d.Member.Seq] = true
		}
		for _, m := range c.Members {
			if excluded[m.Seq] {
				return errors.New("retained dropped ticket cannot compose a successor train")
			}
			t, ok := bySeq[m.Seq]
			if !ok || t.Unreadable || t.Issue != m.Issue || t.Branch != m.Branch {
				return errors.New("train member does not match its live ticket")
			}
			if m.Seq != h.ticket.Seq && t.QueueRoot != "" && m.OwnerRoot != t.QueueRoot {
				return errors.New("train member owner differs from live ticket")
			}
		}
		return l.writeTrainControlLocked(c)
	})
}

func (l *Line) UpdateTrain(h *Handle, tr Train) (*TrainControl, error) {
	var out *TrainControl
	err := l.withMutex(func() error {
		c, err := l.ownTrainLocked(h)
		if err != nil {
			return err
		}
		if tr.ID != c.Train {
			return ErrTrainControlStale
		}
		if tr.Phase == PhaseMerging && !c.MergeStarted {
			return errors.New("merge phase requires StartTrainMerge")
		}
		if c.MergeStarted && !tr.MergeStarted {
			return errors.New("train update would clear irreversible merge marker")
		}
		bySeq := map[int64]TrainMember{}
		for _, m := range c.Members {
			bySeq[m.Seq] = m
		}
		members := make([]TrainMember, 0, len(tr.Cars))
		for _, car := range tr.Cars {
			m, ok := bySeq[car.Seq]
			if !ok || m.Car() != car {
				return errors.New("train update changed a pinned member")
			}
			members = append(members, m)
		}
		c.Phase = tr.Phase
		c.MergeStarted = c.MergeStarted || tr.MergeStarted
		c.Members = members
		if err := l.writeTrainControlLocked(*c); err != nil {
			return err
		}
		out = c
		return nil
	})
	return out, err
}

// ApplyTrainDrops seals the full accepted batch with the same mutex that
// accepts requests. A racing second request is either in this transition or
// observes its terminal phase; no accepted request falls between snapshots.
// The conductor must cancel and join the obsolete effect before calling it.
func (l *Line) ApplyTrainDrops(h *Handle, tr Train) (Train, []Effect, *TrainControl, error) {
	var next Train
	var effects []Effect
	var out *TrainControl
	err := l.withMutex(func() error {
		c, err := l.ownTrainLocked(h)
		if err != nil {
			return err
		}
		if tr.ID != c.Train {
			return ErrTrainControlStale
		}
		if c.MergeStarted {
			return ErrTrainMergeStarted
		}
		n, work, err := Transition(tr, CarsDropped{Drops: c.CurrentDrops()})
		if err != nil {
			return err
		}
		c.Phase = n.Phase
		// CarsDropped shares Aborted's terminal transition; no car remains
		// composed after its held/release effects have been committed.
		c.Members = nil
		if err := l.writeTrainControlLocked(*c); err != nil {
			return err
		}
		next, effects, out = n, work, c
		return nil
	})
	if err != nil {
		return tr, nil, nil, err
	}
	return next, effects, out, nil
}

// UpdateTrainPublication fences before push/create, including bisect heads.
func (l *Line) UpdateTrainPublication(h *Handle, p TrainPublication) (*TrainControl, error) {
	var out *TrainControl
	err := l.withMutex(func() error {
		c, err := l.ownTrainLocked(h)
		if err != nil {
			return err
		}
		if len(c.CurrentDrops()) > 0 {
			return ErrTrainDropPending
		}
		if c.Publication.Started && !p.Started {
			return errors.New("publication-started marker cannot be cleared")
		}
		if c.Publication.Started && (p.Repository != c.Publication.Repository || p.Remote != c.Publication.Remote || p.Branch != c.Publication.Branch) {
			return errors.New("train publication identity changed")
		}
		p.Extra = mergeExtra(c.Publication.Extra, p.Extra)
		c.Publication = p
		if err := l.writeTrainControlLocked(*c); err != nil {
			return err
		}
		out = c
		return nil
	})
	return out, err
}

func (l *Line) RequestTrainDrop(train int64, m TrainMember, reason string) (DropReceipt, error) {
	var receipt DropReceipt
	err := l.withMutex(func() error {
		c, err := l.readTrainControlLocked()
		if err != nil {
			return err
		}
		// A recorded merge-start/terminal state can refuse exactly even
		// after its head left. Only a prospective new acceptance requires
		// the matching live head; an accepted retry keeps its receipt.
		next, d, again, err := decideTrainDrop(*c, train, m, reason, l.o.Now())
		if err != nil {
			return err
		}
		if !again {
			head, err := l.headGenerationLocked()
			if err != nil {
				return err
			}
			if head == nil || head.Unreadable || head.Seq != c.Train || c.Train != train {
				return ErrTrainControlStale
			}
		}
		if !again {
			if err := l.writeTrainControlLocked(next); err != nil {
				return err
			}
		}
		receipt = DropReceipt{Control: next, Drop: d, Repeated: again}
		live, err := l.scan()
		if err == nil {
			_, err = l.reconcileTrainDropsLocked(&next, live)
		}
		receipt.NotificationError = err
		return nil
	})
	return receipt, err
}

func (l *Line) StartTrainMerge(h *Handle, tr Train) (*TrainControl, error) {
	var out *TrainControl
	err := l.withMutex(func() error {
		c, err := l.ownTrainLocked(h)
		if err != nil {
			return err
		}
		next, err := decideTrainMerge(*c, tr)
		if err != nil {
			return err
		}
		if err := l.writeTrainControlLocked(next); err != nil {
			return err
		}
		out = &next
		return nil
	})
	return out, err
}
func (l *Line) ReadTrainControl() (*TrainControl, error) {
	var out *TrainControl
	err := l.withMutex(func() error { var err error; out, err = l.readTrainControlLocked(); return err })
	return out, err
}
func (l *Line) TrainControlStatus() (*TrainControl, *Ticket, error) {
	var out *TrainControl
	var head *Ticket
	err := l.withMutex(func() error {
		var err error
		out, err = l.readTrainControlLocked()
		if err != nil {
			return err
		}
		head, err = l.headGenerationLocked()
		return err
	})
	return out, head, err
}

// The admitted head records positive external proof only for the exact
// obligation it inspected, preserving all still-live accepted exclusions.
func (l *Line) FinishTrainControl(h *Handle, train int64, expected TrainCleanup) error {
	return l.withMutex(func() error {
		if err := l.requireHeadLocked(h); err != nil {
			return err
		}
		c, err := l.readTrainControlLocked()
		if err != nil {
			return err
		}
		if c.Train != train || !reflect.DeepEqual(c.Cleanup, expected) {
			return errors.New("train cleanup obligation changed; re-read and confirm exact cleanup")
		}
		c.Cleanup.Pending = false
		return l.writeTrainControlLocked(*c)
	})
}

// Reconciliation runs under line.lock, never nesting the public Settle.
// Every accepted notification is materialized before candidate selection.
func (l *Line) reconcileTrainDropsLocked(c *TrainControl, live []Ticket) ([]TrainDrop, error) {
	alive := map[int64]bool{}
	for _, t := range live {
		alive[t.Seq] = true
	}
	// Admission removes the conductor's ticket. Its head.lock, rather than
	// tickets/, is the lifetime resource for an accepted self-drop.
	head, err := l.headGenerationLocked()
	if err != nil {
		return nil, err
	}
	var kept []TrainDrop
	for _, d := range c.Drops {
		isHead := head != nil && !head.Unreadable && head.Seq == d.Member.Seq
		// A fast dropped waiter can consume Held and release its ticket
		// before the active conductor's watcher observes the safety write.
		// Keep that generation's decision until its terminal batch seal.
		activeDecision := head != nil && !head.Unreadable && head.Seq == c.Train && d.Train == c.Train && !c.Phase.Terminal()
		if !alive[d.Member.Seq] && !isHead && !activeDecision {
			continue
		}
		kept = append(kept, d)
		if !alive[d.Member.Seq] {
			continue
		}
		old := l.outcomeOf(d.Member.Seq)
		if old != nil && old.Kind == OutcomeHeld && old.Train == d.Train && old.Reason == d.Reason {
			continue
		}
		o := Outcome{Version: OutcomeVersion, Seq: d.Member.Seq, Train: d.Train, Kind: OutcomeHeld, Reason: d.Reason, At: l.o.Now()}
		if old != nil {
			o.Extra = old.Extra
		}
		if err := writeAtomic(filepath.Join(l.dir, "outcomes"), outcomePath(l.dir, d.Member.Seq), o); err != nil {
			return nil, fmt.Errorf("drop accepted; held notification pending for ticket %d: %w", d.Member.Seq, err)
		}
	}
	return kept, nil
}
func (l *Line) reconcileAdmissionDropsLocked(live []Ticket) ([]Ticket, error) {
	c, err := l.readTrainControlLocked()
	if errors.Is(err, os.ErrNotExist) {
		return live, nil
	}
	if err != nil {
		return nil, err
	}
	kept, err := l.reconcileTrainDropsLocked(c, live)
	if err != nil {
		return nil, err
	}
	if len(kept) != len(c.Drops) {
		c.Drops = kept
		if err := l.writeTrainControlLocked(*c); err != nil {
			return nil, err
		}
	}
	excluded := map[int64]bool{}
	for _, d := range kept {
		excluded[d.Member.Seq] = true
	}
	return slices.DeleteFunc(slices.Clone(live), func(t Ticket) bool { return excluded[t.Seq] }), nil
}
