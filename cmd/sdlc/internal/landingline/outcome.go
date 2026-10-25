package landingline

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// A train's conductor tells each waiting car what happened to it through
// outcomes/<seq> (#449): the car is still a waiter (it holds its ticket) while
// the train runs, and its own merge process does its own cleanup.

// OutcomeKind is what a train did with a car.
type OutcomeKind string

const (
	OutcomeBoarded  OutcomeKind = "boarded"  // on a running train: keep waiting
	OutcomeLanded   OutcomeKind = "landed"   // its PR is on main: clean up
	OutcomeHeld     OutcomeKind = "held"     // culprit or conflict: leave, refused
	OutcomeRefused  OutcomeKind = "refused"  // the whole train was refused
	OutcomeReleased OutcomeKind = "released" // back to waiting, still ready
)

// OutcomeVersion is the sidecar format; a reader refuses another.
const OutcomeVersion = 1

// Outcome is one car's word from its train.
type Outcome struct {
	Version int                        `json:"version"`
	Seq     int64                      `json:"seq"`
	Train   int64                      `json:"train"`
	Kind    OutcomeKind                `json:"kind"`
	Reason  string                     `json:"reason,omitempty"`
	PR      int                        `json:"pr,omitempty"` // the train PR
	At      time.Time                  `json:"at"`
	Extra   map[string]json.RawMessage `json:"-"`
}

func outcomeFields(o *Outcome) map[string]any {
	return map[string]any{"version": &o.Version, "seq": &o.Seq, "train": &o.Train, "kind": &o.Kind,
		"reason": &o.Reason, "pr": &o.PR, "at": &o.At}
}

// DecodeOutcome parses an outcome with the line's strict, preserving codec.
func DecodeOutcome(data []byte) (Outcome, error) {
	var o Outcome
	extra, err := jsonstrict.DecodeObject(data, outcomeFields(&o))
	if err != nil {
		return Outcome{}, fmt.Errorf("outcome: %w", err)
	}
	o.Extra = extra
	return o, nil
}

func (o Outcome) MarshalJSON() ([]byte, error) {
	type known Outcome
	base, err := json.Marshal(known(o))
	if err != nil {
		return nil, err
	}
	fields := outcomeFields(&o)
	return jsonstrict.EncodeWithExtra(base, o.Extra, func(k string) bool { _, ok := fields[k]; return ok }), nil
}

// Valid reports whether o is a well-formed outcome for ticket seq.
func (o Outcome) Valid(seq int64) error {
	switch {
	case o.Version != OutcomeVersion:
		return fmt.Errorf("outcome version %d, want %d", o.Version, OutcomeVersion)
	case o.Seq != seq:
		return fmt.Errorf("outcome for ticket %d filed as %d", o.Seq, seq)
	case o.Train <= 0:
		return errors.New("outcome without a train")
	case o.At.IsZero():
		return errors.New("outcome without a time")
	}
	switch o.Kind {
	case OutcomeBoarded, OutcomeLanded, OutcomeHeld, OutcomeRefused, OutcomeReleased:
		return nil
	}
	return fmt.Errorf("unknown outcome %q", o.Kind)
}

// Final reports an outcome that ends the car's wait.
func (o Outcome) Final() bool {
	return o.Kind == OutcomeLanded || o.Kind == OutcomeHeld || o.Kind == OutcomeRefused
}

// Settled is Wait's return when a train ended this car's wait: landed (the
// caller cleans up its merged PR) or held/refused (the caller fails with
// Reason). Release still applies: the caller leaves the line.
type Settled struct{ Outcome Outcome }

func (s *Settled) Error() string {
	return fmt.Sprintf("landing train %d: %s: %s", s.Outcome.Train, s.Outcome.Kind, s.Outcome.Reason)
}

// ErrGone is a settle for a ticket that is no longer in line.
var ErrGone = errors.New("landing line: the car is no longer waiting")

// Settle records a car's outcome under the mutex, only while its ticket is
// live: a car that died meanwhile gets none (its rerun of sdlc merge resumes
// from GitHub's state instead). A later sdlc's members on the previous
// outcome are carried forward.
func (l *Line) Settle(seq int64, o Outcome) error {
	o.Version, o.Seq = OutcomeVersion, seq
	if o.At.IsZero() {
		o.At = l.o.Now()
	}
	if err := o.Valid(seq); err != nil {
		return err
	}
	return l.withMutex(func() error {
		live, err := l.scan()
		if err != nil {
			return err
		}
		alive := false
		for _, t := range live {
			alive = alive || t.Seq == seq
		}
		if !alive {
			return ErrGone
		}
		path := outcomePath(l.dir, seq)
		if old, err := os.ReadFile(path); err == nil && len(old) <= maxTicket {
			if prev, err := DecodeOutcome(old); err == nil && prev.Seq == seq {
				o.Extra = mergeExtra(prev.Extra, o.Extra)
			}
		}
		return writeAtomic(filepath.Join(l.dir, "outcomes"), path, o)
	})
}

// outcomeOf reads a car's outcome; none, or one that doesn't parse, is nil.
func (l *Line) outcomeOf(seq int64) *Outcome {
	data, err := os.ReadFile(outcomePath(l.dir, seq))
	if err != nil || len(data) > maxTicket {
		return nil
	}
	o, err := DecodeOutcome(data)
	if err != nil || o.Valid(seq) != nil {
		return nil
	}
	return &o
}

// clearOutcome removes a released car's outcome so the next train starts
// from none; done under the mutex so it never removes a newer one.
func (l *Line) clearOutcome(seq int64, kind OutcomeKind) error {
	return l.withMutex(func() error {
		if o := l.outcomeOf(seq); o != nil && o.Kind == kind {
			return os.Remove(outcomePath(l.dir, seq))
		}
		return nil
	})
}

func outcomePath(dir string, seq int64) string {
	return filepath.Join(dir, "outcomes", fmt.Sprintf("%020d", seq))
}

func mergeExtra(prev, next map[string]json.RawMessage) map[string]json.RawMessage {
	if len(prev) == 0 {
		return next
	}
	out := make(map[string]json.RawMessage, len(prev)+len(next))
	for k, v := range prev {
		out[k] = v
	}
	for k, v := range next {
		out[k] = v
	}
	return out
}

// writeAtomic writes v as JSON to path by temporary file then rename.
func writeAtomic(dir, path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}

// sidecarSeq parses a sidecar's name; ok is false for anything else.
func sidecarSeq(name string) (int64, bool) {
	seq, err := strconv.ParseInt(name, 10, 64)
	return seq, err == nil
}
