package landingline

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// The landing train (#449): on an interval, the line's head lands every car
// that vouched readiness with it, in one integration and one suite. This file
// is the pure part of when a train departs and who boards it; trainstate.go
// is what a running train does, bisect.go how a failing one is split.

// Schedule is a repository's train setting. Interval 0 is the serial line of
// #393: every head departs at once and rides alone.
type Schedule struct {
	Interval time.Duration
	MaxCars  int
}

// DefaultMaxCars bounds one train; a log2 bisect of 8 costs at most 3 suites.
const DefaultMaxCars = 8

// MaxReadinessAge is how long a car's vouch stays current. A waiting car
// re-vouches on every poll that its gates still hold; one older than this is
// a car that stopped looking, and it does not board.
const MaxReadinessAge = 5 * time.Minute

// Off reports whether trains are off: the serial line, unchanged.
func (s Schedule) Off() bool { return s.Interval <= 0 }

// DepartureDue reports whether a train may depart at now, and why:
//   - trains are off (the serial line);
//   - no train has run yet in the log's window;
//   - the last train's next departure has come (its departure plus the
//     interval, or at once when it left ready cars behind).
//
// The last train comes from the events log, so the schedule survives any
// process's death and needs no daemon. A fleet-blocking grant is not a
// departure: it is admitted alone, outside the train (the admission rule).
func DepartureDue(now time.Time, s Schedule, events []Event) (bool, string) {
	if s.Off() {
		return true, "trains are off"
	}
	next, ok := NextDeparture(events)
	if !ok {
		return true, "first train"
	}
	if !now.Before(next) {
		return true, "train due"
	}
	return false, fmt.Sprintf("next train at %s", next.Format(time.RFC3339))
}

// NextDeparture is the latest train event's NextDue; false when the log holds
// no train.
func NextDeparture(events []Event) (time.Time, bool) {
	var next time.Time
	found := false
	for _, e := range events {
		if e.Kind == EventTrain && !e.NextDue.IsZero() {
			next, found = e.NextDue, true
		}
	}
	return next, found
}

// NextDue is when the train after one that departed at departed may leave:
// at once when this one left ready cars behind (cap, split remainder), else
// one interval on.
func NextDue(departed time.Time, s Schedule, leftReady bool) time.Time {
	if leftReady {
		return departed
	}
	return departed.Add(s.Interval)
}

// ReadinessVersion is the sidecar format; a reader refuses another.
const ReadinessVersion = 1

// Readiness is a waiting car's vouch, written by the car's own merge process
// after its own gates passed for Head: the reviewed anchor unchanged, a branch
// suite record and green CI on that exact commit. Alone marks a car that must
// ride by itself (it touches local-gates paths, so the train PR's guard would
// fail). The conductor trusts none of it beyond a claim: it re-reads the
// car's remote tip and boards only a car whose tip is Head.
type Readiness struct {
	Version int       `json:"version"`
	Seq     int64     `json:"seq"`
	Branch  string    `json:"branch"`
	Head    string    `json:"head"`
	PR      int       `json:"pr"`
	Ready   bool      `json:"ready"`
	Alone   bool      `json:"alone,omitempty"`
	Reason  string    `json:"reason,omitempty"`
	At      time.Time `json:"at"`
	// Extra holds members a later sdlc wrote that this one doesn't know.
	// They survive a rewrite of the same ticket's vouch (ARCH-EVOLVE).
	Extra map[string]json.RawMessage `json:"-"`
}

// readinessFields are the known members, by exact name.
func readinessFields(r *Readiness) map[string]any {
	return map[string]any{
		"version": &r.Version, "seq": &r.Seq, "branch": &r.Branch, "head": &r.Head, "pr": &r.PR,
		"ready": &r.Ready, "alone": &r.Alone, "reason": &r.Reason, "at": &r.At,
	}
}

// DecodeReadiness parses a sidecar with the line's strict, preserving codec.
func DecodeReadiness(data []byte) (Readiness, error) {
	var r Readiness
	extra, err := decodeStrict(data, readinessFields(&r))
	if err != nil {
		return Readiness{}, fmt.Errorf("readiness: %w", err)
	}
	r.Extra = extra
	return r, nil
}

// MarshalJSON writes the known members and then Extra, so a rewrite carries
// a later version's members forward. Extra never overrides a known member.
func (r Readiness) MarshalJSON() ([]byte, error) {
	type known Readiness // drops the method, keeps the tags
	base, err := json.Marshal(known(r))
	if err != nil {
		return nil, err
	}
	fields := readinessFields(&r)
	return encodeWithExtra(base, r.Extra, func(k string) bool { _, ok := fields[k]; return ok }), nil
}

var fullOID = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// Valid reports whether r is a well-formed vouch for ticket seq: a parsed
// sidecar from another process is a claim to check, never evidence (ARCH-SECURE).
func (r Readiness) Valid(seq int64) error {
	switch {
	case r.Version != ReadinessVersion:
		return fmt.Errorf("readiness version %d, want %d", r.Version, ReadinessVersion)
	case r.Seq != seq:
		return fmt.Errorf("readiness for ticket %d filed as %d", r.Seq, seq)
	case r.Branch == "":
		return fmt.Errorf("readiness without a branch")
	case r.Ready && !fullOID.MatchString(r.Head):
		return fmt.Errorf("readiness head %q is not a full commit id", r.Head)
	case r.Ready && r.PR <= 0:
		return fmt.Errorf("readiness without a PR")
	case r.At.IsZero():
		return fmt.Errorf("readiness without a time")
	}
	return nil
}

// Current reports whether r vouches a car ready at now.
func (r Readiness) Current(now time.Time) bool {
	return r.Ready && !r.At.After(now) && now.Sub(r.At) <= MaxReadinessAge
}

// For reports why r cannot vouch for ticket t, or "" when it can: a vouch is
// another process's claim, so it must name the ticket's own branch and be
// current. Board and the admission rule share this one check.
func (r Readiness) For(t Ticket, now time.Time) string {
	switch {
	case r.Branch != t.Branch:
		return fmt.Sprintf("vouch names branch %q, the ticket %q", r.Branch, t.Branch)
	case !r.Ready:
		return "not ready: " + r.Reason
	case !r.Current(now):
		return "vouch is stale"
	}
	return ""
}

// Car is one boarded PR: the ticket and the exact head it vouched.
type Car struct {
	Seq    int64  `json:"seq"`
	Issue  int    `json:"issue"`
	Branch string `json:"branch"`
	Head   string `json:"head"`
	PR     int    `json:"pr"`
}

// Boarding is who rides this train and who stays.
type Boarding struct {
	Cars []Car
	// LeftReady is how many current, ready waiters did not board (cap, or
	// behind an Alone car): the next train then departs at once.
	LeftReady int
	// Why is one line per waiter that did not board, for the status view.
	Why map[int64]string
}

// Board picks the train behind the head, in admission order: every waiter
// whose vouch is current, up to the schedule's cap. A waiter holding a
// fleet-blocking grant never boards: it lands alone, outside any train, and
// is admitted next (it counts as left ready, so that is at once). An Alone car rides by
// itself: as the head it boards nobody; as a waiter it ends this train's
// boarding (it and everyone behind it keep their place for the next one). A
// waiter that is not ready is passed over, not a stop: it never held anyone.
func Board(head Car, headAlone bool, waiters []Ticket, ready map[int64]Readiness, grants map[int]Grant, s Schedule, now time.Time) Boarding {
	b := Boarding{Cars: []Car{head}, Why: map[int64]string{}}
	if s.Off() || headAlone {
		for _, t := range waiters {
			if r, ok := ready[t.Seq]; ok && r.For(t, now) == "" {
				b.LeftReady++
			}
		}
		return b
	}
	max := s.MaxCars
	if max <= 0 {
		max = DefaultMaxCars
	}
	stopped := false
	for _, t := range waiters {
		r, ok := ready[t.Seq]
		if !ok {
			b.Why[t.Seq] = "not vouched yet"
			continue
		}
		if why := r.For(t, now); why != "" {
			b.Why[t.Seq] = why
			continue
		}
		if _, granted := grants[t.Issue]; granted {
			b.LeftReady++
			b.Why[t.Seq] = "fleet-blocking: lands alone, outside the train"
			continue
		}
		switch {
		case stopped:
			b.LeftReady++
			b.Why[t.Seq] = "waits for the next train"
		case r.Alone:
			stopped = true
			b.LeftReady++
			b.Why[t.Seq] = "rides alone: " + r.Reason
		case len(b.Cars) >= max:
			stopped = true
			b.LeftReady++
			b.Why[t.Seq] = fmt.Sprintf("train full (%d cars)", max)
		default:
			b.Cars = append(b.Cars, Car{Seq: t.Seq, Issue: t.Issue, Branch: t.Branch, Head: r.Head, PR: r.PR})
		}
	}
	return b
}
