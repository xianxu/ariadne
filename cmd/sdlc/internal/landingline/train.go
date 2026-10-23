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
	// MaxWait bounds a ready car's wait (#475): a train leaves once the
	// oldest ready car has waited this long. 0: no bound.
	MaxWait time.Duration
	// Offset places the repository on the machine plan's wall-clock grid
	// (#475); nil keeps the interval from the last departure.
	Offset *time.Duration
}

// Waiters summarizes the line's ready cars at one admission pass: how many
// hold a current vouch, and the oldest of their first ready vouches.
type Waiters struct {
	Ready       int
	OldestSince time.Time
}

// DefaultMaxCars bounds one train whose repository sets no train_max_cars:
// every adopter starts with two cars, so a failing train splits once and an
// escape is cheap to attribute (#561). Raising it is the manual, measured
// decision FormatTrainRampPolicy states; an explicit cap is never changed.
const DefaultMaxCars = 2

// MaxReadinessAge is how long a car's vouch stays current. A waiting car
// re-vouches on every poll that its gates still hold; one older than this is
// a car that stopped looking, and it does not board.
const MaxReadinessAge = 5 * time.Minute

// Off reports whether trains are off: the serial line, unchanged.
func (s Schedule) Off() bool { return s.Interval <= 0 }

// DepartureDue reports whether a train may depart at now, and why:
//   - trains are off (the serial line);
//   - no train has run yet in the log's window;
//   - N cars are ready (a full train leaves early, #475);
//   - the oldest ready car has waited MaxWait (#475);
//   - the next departure has come: the grid's next slot after the last
//     train, or the interval after it without a machine plan, or at once
//     when the last departure left ready cars behind.
//
// The departure ledger is the events log, so the schedule survives any
// process's death and needs no daemon. An early departure leaves the grid
// alone, and a lone or fleet-blocking landing is not a departure: it never
// pushes the next train back (#475), it can only bring it forward when it
// left ready cars behind.
func DepartureDue(now time.Time, s Schedule, events []Event, w Waiters) (bool, string) {
	if s.Off() {
		return true, "trains are off"
	}
	next, why, ok := nextDeparture(events, s)
	if !ok {
		return true, "first train"
	}
	if s.MaxCars > 0 && w.Ready >= s.MaxCars {
		return true, fmt.Sprintf("train due (full: %d ready)", w.Ready)
	}
	deadline, waiting := waitDeadline(s, w)
	if waiting && !now.Before(deadline) {
		return true, fmt.Sprintf("train due (a car has waited %s)", s.MaxWait)
	}
	if !now.Before(next) {
		return true, "train due (" + why + ")"
	}
	if waiting && deadline.Before(next) {
		next, why = deadline, "max wait"
	}
	return false, fmt.Sprintf("next train at %s (%s)", next.Format("2006-01-02 15:04"), why)
}

// waitDeadline is when the oldest ready car will have waited MaxWait; false
// when no bound is set or no car is ready.
func waitDeadline(s Schedule, w Waiters) (time.Time, bool) {
	if s.MaxWait <= 0 || w.Ready == 0 || w.OldestSince.IsZero() {
		return time.Time{}, false
	}
	return w.OldestSince.Add(s.MaxWait), true
}

// departedOf is when a recorded departure left: Departed, or At for a record
// written before Departed existed.
func departedOf(e Event) time.Time {
	if !e.Departed.IsZero() {
		return e.Departed
	}
	return e.At
}

// leftReady reports whether a record's NextDue is "at once": its departure
// left ready cars behind.
func leftReady(e Event) bool { return !e.NextDue.After(departedOf(e)) }

// nextDeparture is the next train's earliest departure and why. A
// conductor's train sets it: the grid's next slot after its departure, read
// at this moment so a changed plan offset applies at once, or the NextDue it
// recorded without a plan. Any newer record that left ready cars behind (a
// train cut short, or a lone landing) brings it forward to then. A lone
// landing's record is never a schedule otherwise, including one written by an
// sdlc before #475 with the interval after it.
func nextDeparture(events []Event, s Schedule) (time.Time, string, bool) {
	var next time.Time
	why, found := "", false
	for _, e := range events {
		if e.Kind != EventTrain || e.NextDue.IsZero() {
			continue
		}
		switch {
		case leftReady(e):
			next, why, found = e.NextDue, "cars left ready", true
		case e.Result == "alone":
			continue // not a departure (#475)
		case s.Offset != nil:
			next, why, found = Slots{Interval: s.Interval, Offset: *s.Offset}.Next(departedOf(e)), "slot", true
		default:
			next, why, found = e.NextDue, "interval", true
		}
	}
	return next, why, found
}

// NextDeparture is the next train's earliest departure as the log records
// it without a machine plan (status and history); false when the log holds
// no departure.
func NextDeparture(events []Event) (time.Time, bool) {
	next, _, ok := nextDeparture(events, Schedule{})
	return next, ok
}

// NextDue is when the train after one that departed at departed may leave:
// at once when this one left ready cars behind (cap, split remainder), else
// the grid's next slot after it (#475), or one interval on without a
// machine plan.
func NextDue(departed time.Time, s Schedule, leftReady bool) time.Time {
	if leftReady {
		return departed
	}
	if s.Offset != nil {
		return Slots{Interval: s.Interval, Offset: *s.Offset}.Next(departed)
	}
	return departed.Add(s.Interval)
}

// AloneNextDue is the NextDue a lone (or granted, or refused) landing
// records: none, so it never moves the next train (#475), unless it left
// ready cars behind, which brings that train forward to now.
func AloneNextDue(departed time.Time, leftReady bool) time.Time {
	if leftReady {
		return departed
	}
	return time.Time{}
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
	// Since is the car's first ready vouch at this head (#475): a re-vouch
	// keeps it, a new head or a not-ready vouch starts it over. A sidecar
	// from before it reads as At.
	Since time.Time `json:"since,omitzero"`
	// SinceStamp is the At of the vouch that wrote Since. An older sdlc that
	// re-vouches for any reason (a new head, not ready, ready again) rewrites
	// At but carries Since and SinceStamp forward as unknown members, so the
	// mismatch marks that Since stale: it is trusted only on a vouch written
	// by an sdlc that maintains it (#475).
	SinceStamp time.Time `json:"since_stamp,omitzero"`
	// Proof is the record that satisfied the car's CI gate (#561), for train
	// measurements only: AcceptedProof trusts it only when bound to this
	// vouch, and it never decides readiness.
	Proof *BoardingProof `json:"proof,omitempty"`
	// Extra holds members a later sdlc wrote that this one doesn't know.
	// They survive a rewrite of the same ticket's vouch (ARCH-EVOLVE).
	Extra map[string]json.RawMessage `json:"-"`
}

// readinessFields are the known members, by exact name.
func readinessFields(r *Readiness) map[string]any {
	return map[string]any{
		"version": &r.Version, "seq": &r.Seq, "branch": &r.Branch, "head": &r.Head, "pr": &r.PR,
		"ready": &r.Ready, "alone": &r.Alone, "reason": &r.Reason, "at": &r.At, "since": &r.Since, "since_stamp": &r.SinceStamp,
		"proof": &r.Proof,
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

// firstReady is when r's car first vouched ready at its head: Since, or At
// for a sidecar written before Since existed.
func (r Readiness) firstReady() time.Time {
	if !r.Since.IsZero() && r.SinceStamp.Equal(r.At) {
		return r.Since
	}
	return r.At
}

// carrySince keeps a re-vouch's first-ready time: same head, ready before
// and now. Anything else starts it over (a new head, or not ready).
func carrySince(prev *Readiness, r Readiness) time.Time {
	if !r.Ready {
		return time.Time{}
	}
	if prev != nil && prev.Ready && prev.Head == r.Head && prev.Branch == r.Branch {
		return prev.firstReady()
	}
	return r.At
}

// ReadyWaiters summarizes the live tickets whose vouch holds at now.
func ReadyWaiters(live []Ticket, ready map[int64]Readiness, now time.Time) Waiters {
	var w Waiters
	for _, t := range live {
		r, ok := ready[t.Seq]
		if !ok || r.For(t, now) != "" {
			continue
		}
		w.Ready++
		if s := r.firstReady(); w.OldestSince.IsZero() || s.Before(w.OldestSince) {
			w.OldestSince = s
		}
	}
	return w
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
