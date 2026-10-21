package landingline

import (
	"fmt"
	"slices"
)

// A running train is an explicit state machine (ARCH-ORDER): the conductor
// (cmd/sdlc/landingtrain.go) feeds Transition the outcome of each effect it
// executed and executes the effects it returns. Nothing else changes a
// Train's state, so its legal histories are the ones Transition accepts.

// Phase is where a train is.
type Phase string

const (
	PhaseForming       Phase = "forming"       // admitted head, choosing cars
	PhaseIntegrating   Phase = "integrating"   // building the train branch
	PhaseTesting       Phase = "testing"       // a suite on the full train
	PhaseBisecting     Phase = "bisecting"     // suites on prefixes, finding a culprit
	PhaseAttributing   Phase = "attributing"   // fresh failure attribution before suite blame
	PhaseReintegrating Phase = "reintegrating" // main moved: merging it into the train
	PhaseChecking      Phase = "checking"      // final prefix landing checks after its suite
	PhaseLanding       Phase = "landing"       // waiting for CI on the train PR
	PhaseMerging       Phase = "merging"       // the train PR's merge was sent
	PhaseLanded        Phase = "landed"        // terminal: prefix 1..Land is on main
	PhaseAbandoned     Phase = "abandoned"     // terminal: nothing landed, PR closed
	PhaseRefused       Phase = "refused"       // terminal: the train's cars refused
	// PhaseUnconfirmed is terminal: the conductor stopped while its merge's
	// outcome was unknown. No record may claim abandoned or landed without
	// an observation (#449); the cars go back to wait and each one reads
	// its own PR's state.
	PhaseUnconfirmed Phase = "merge-unconfirmed"
)

// Terminal reports a phase no event leaves.
func (p Phase) Terminal() bool {
	return p == PhaseLanded || p == PhaseAbandoned || p == PhaseRefused || p == PhaseUnconfirmed
}

// MaxReintegrations bounds main moving under a train, as integrationLimit
// bounds a single landing.
const MaxReintegrations = 3

// SuiteOutcome is a recorded suite's verdict on one prefix. A known flake
// whose rerun passed (#394) is Pass: it never splits a train.
type SuiteOutcome int

const (
	SuitePass    SuiteOutcome = iota
	SuiteFail                 // an unexplained failure: split
	SuiteMainRed              // main's baseline fails too: no car is to blame
)

// Train is a running train's state.
type Train struct {
	ID    int64 // the conductor's ticket
	Phase Phase
	Cars  []Car // in admission order; prefix k is Cars[:k]
	// Testing is the prefix whose suite is running; Land the prefix that will
	// land once CI is green.
	Testing, Land      int
	AssessLo, AssessHi int // exact adjacent pair awaiting attribution
	// Published is the prefix the train PR's head is at (0: not pushed).
	Published    int
	Results      map[int]bool
	Splits       int
	Culprit      int64
	Reintegrated int
	Probed       bool
	// Released are cars sent back to wait, ready (a split's remainder): the
	// next train departs at once for them.
	Released []int64
}

// TrainEvent is an outcome the conductor observed.
type TrainEvent interface{ trainEvent() }

type (
	// Boarded: Board chose these cars and each one's remote tip was re-read
	// equal to its vouched head (a car whose tip moved is not in Cars).
	Boarded struct{ Cars []Car }
	// Built: the train branch merged Kept in order; Conflicts did not merge.
	Built struct {
		Kept      []int64
		Conflicts map[int64]string
	}
	// SuiteDone: the recorded suite on prefix Prefix finished.
	SuiteDone struct {
		Prefix  int
		Outcome SuiteOutcome
	}
	// SplitAssessed carries fresh attribution for the requested adjacent pair.
	SplitAssessed struct {
		Lo, Hi  int
		Verdict SplitAttribution
	}
	// LandingChecksDone is the selected final-prefix checks' observed result.
	LandingChecksDone struct {
		Prefix int
		Passed bool
		Detail string
	}
	// CIDone: CI on the train PR's head finished.
	CIDone struct {
		Green  bool
		Detail string
	}
	// MainMoved: main is no longer the train's base (seen before or by the merge).
	MainMoved struct{}
	// Reintegrated: main is merged into the train again; NeedSuite when the
	// change is code (a docs-only move keeps the suite record).
	Reintegrated struct{ NeedSuite bool }
	// Merged: the train PR is merged (confirmed on GitHub). Each car proves
	// its own landing later, through its PR's state, so no commit rides here.
	Merged struct{}
	// MergeUnknown: the merge request's outcome was not observed.
	MergeUnknown struct{}
	// Probed: GitHub was read after MergeUnknown.
	Probed struct{ Merged bool }
	// Aborted: the conductor hit an error it cannot carry on from (an IO
	// failure, a reintegration conflict). Every car still riding goes back
	// to wait, ready; a merge that may have landed is each car's to see.
	Aborted struct{ Reason string }
)

func (Boarded) trainEvent()           {}
func (Built) trainEvent()             {}
func (SplitAssessed) trainEvent()     {}
func (SuiteDone) trainEvent()         {}
func (LandingChecksDone) trainEvent() {}
func (CIDone) trainEvent()            {}
func (MainMoved) trainEvent()         {}
func (Reintegrated) trainEvent()      {}
func (Merged) trainEvent()            {}
func (MergeUnknown) trainEvent()      {}
func (Probed) trainEvent()            {}
func (Aborted) trainEvent()           {}

// Effect is work the conductor executes; its outcome comes back as an Event.
type Effect interface{ trainEffect() }

type (
	AssessSplit      struct{ Lo, Hi int } // validate exact failure, registry and candidate impact
	Build            struct{ Cars []Car } // merge main, then each car's head, in order
	Publish          struct{ Prefix int } // push the train branch at prefix, open or update its PR
	RunLandingChecks struct{ Prefix int } // selected final-prefix checks, never probes
	RunSuite         struct{ Prefix int } // a recorded suite on prefix's commit
	WaitCI           struct{ Prefix int }
	Merge            struct{ Prefix int } // merge the train PR, pinned to its head
	Probe            struct{}             // read the train PR's state
	Reintegrate      struct{ Prefix int } // merge current main into the train
	Hold             struct {             // a car leaves, refused with reason (culprit, conflict)
		Seq    int64
		Reason string
	}
	Release struct{ Seqs []int64 } // cars go back to waiting, still ready
	Refuse  struct {               // cars leave, refused with reason
		Seqs   []int64
		Reason string
	}
	LandCars struct{ Seqs []int64 } // these cars' PRs are on main: they clean up
	ClosePR  struct{}               // close the train PR and delete its branch
)

func (AssessSplit) trainEffect()      {}
func (Build) trainEffect()            {}
func (Publish) trainEffect()          {}
func (RunLandingChecks) trainEffect() {}
func (RunSuite) trainEffect()         {}
func (WaitCI) trainEffect()           {}
func (Merge) trainEffect()            {}
func (Probe) trainEffect()            {}
func (Reintegrate) trainEffect()      {}
func (Hold) trainEffect()             {}
func (Release) trainEffect()          {}
func (Refuse) trainEffect()           {}
func (LandCars) trainEffect()         {}
func (ClosePR) trainEffect()          {}

// NewTrain is a train its conductor just formed.
func NewTrain(id int64) Train { return Train{ID: id, Phase: PhaseForming, Results: map[int]bool{}} }

// Transition is the train's one state change. A rejected event (wrong phase,
// a stale suite, anything after a terminal phase) returns an error and t
// unchanged: the conductor stops rather than guess.
func Transition(t Train, ev TrainEvent) (Train, []Effect, error) {
	if t.Phase.Terminal() {
		return t, nil, fmt.Errorf("train %d is %s; %T ignored", t.ID, t.Phase, ev)
	}
	n := t
	n.Results = clone(t.Results)
	n.Cars = slices.Clone(t.Cars)
	n.Released = slices.Clone(t.Released)
	switch e := ev.(type) {
	case Aborted:
		if t.Phase == PhaseForming {
			n.Phase = PhaseAbandoned
			return n, nil, nil
		}
		n.Released = append(n.Released, carSeqs(n.Cars)...)
		effects := []Effect{}
		if len(n.Cars) > 0 {
			effects = append(effects, Release{Seqs: carSeqs(n.Cars)})
		}
		n.Cars = nil
		if t.Phase == PhaseMerging {
			// The merge may have landed: the PR is left for the next
			// conductor's reclaim, and each released car sees its own PR.
			n.Phase = PhaseUnconfirmed
			return n, effects, nil
		}
		n.Phase = PhaseAbandoned
		return n, append(effects, ClosePR{}), nil
	case Boarded:
		if t.Phase != PhaseForming || len(e.Cars) == 0 {
			break
		}
		n.Cars = slices.Clone(e.Cars)
		n.Phase = PhaseIntegrating
		return n, []Effect{Build{Cars: slices.Clone(e.Cars)}}, nil
	case Built:
		if t.Phase != PhaseIntegrating || !partitions(t.Cars, e) {
			break
		}
		var effects []Effect
		var kept []Car
		for _, c := range n.Cars {
			if reason, conflict := e.Conflicts[c.Seq]; conflict {
				effects = append(effects, Hold{Seq: c.Seq, Reason: reason})
			} else if slices.Contains(e.Kept, c.Seq) {
				kept = append(kept, c)
			}
		}
		n.Cars = kept
		if len(kept) == 0 {
			n.Phase = PhaseAbandoned
			return n, append(effects, ClosePR{}), nil
		}
		n.Phase, n.Testing = PhaseTesting, len(kept)
		effects = append(effects, n.publish(len(kept))...)
		return n, append(effects, RunSuite{Prefix: len(kept)}), nil
	case SuiteDone:
		if (t.Phase != PhaseTesting && t.Phase != PhaseBisecting) || e.Prefix != t.Testing {
			break
		}
		if e.Outcome == SuiteMainRed {
			n.Phase = PhaseRefused
			return n, []Effect{Refuse{Seqs: carSeqs(n.Cars), Reason: "main is red: its baseline suite fails too"}, ClosePR{}}, nil
		}
		n.Results[e.Prefix] = e.Outcome == SuitePass
		if t.Phase == PhaseTesting && e.Outcome == SuitePass {
			n.Phase, n.Land = PhaseChecking, e.Prefix
			return n, append(n.publish(e.Prefix), RunLandingChecks{Prefix: e.Prefix}), nil
		}
		next, culprit, lo, done := Bisect(len(n.Cars), n.Results)
		if !done {
			n.Phase, n.Testing = PhaseBisecting, next
			n.Splits++
			return n, []Effect{RunSuite{Prefix: next}}, nil
		}
		n.Phase, n.AssessLo, n.AssessHi = PhaseAttributing, lo, culprit
		return n, []Effect{AssessSplit{Lo: lo, Hi: culprit}}, nil
	case SplitAssessed:
		if t.Phase != PhaseAttributing || e.Lo != t.AssessLo || e.Hi != t.AssessHi || e.Verdict.Reason == "" {
			break
		}
		switch e.Verdict.Kind {
		case AttributionAttributable:
			return split(n, e.Hi, e.Lo, e.Verdict.Reason)
		case AttributionUnattributed, AttributionUnverified:
			// Stop unresolved cars; Release would silently retrain them.
			effects := []Effect{Refuse{Seqs: carSeqs(n.Cars[e.Lo:]), Reason: e.Verdict.Reason}}
			return finishSplit(n, e.Lo, effects)
		default:
			break
		}
	case LandingChecksDone:
		if t.Phase != PhaseChecking || e.Prefix != t.Land {
			break
		}
		if !e.Passed {
			n.Phase = PhaseRefused
			return n, []Effect{Refuse{Seqs: carSeqs(n.Cars[:n.Land]), Reason: "the train's landing checks failed: " + e.Detail}, ClosePR{}}, nil
		}
		n.Phase = PhaseLanding
		return n, []Effect{WaitCI{Prefix: n.Land}}, nil
	case CIDone:
		if t.Phase != PhaseLanding {
			break
		}
		if !e.Green {
			n.Phase = PhaseRefused
			return n, []Effect{Refuse{Seqs: carSeqs(n.Cars[:n.Land]), Reason: "the train's CI is red: " + e.Detail}, ClosePR{}}, nil
		}
		n.Phase = PhaseMerging
		return n, []Effect{Merge{Prefix: n.Land}}, nil
	case MainMoved:
		if t.Phase != PhaseLanding && t.Phase != PhaseMerging {
			break
		}
		if t.Reintegrated >= MaxReintegrations {
			n.Phase = PhaseRefused
			return n, []Effect{Refuse{Seqs: carSeqs(n.Cars[:n.Land]), Reason: fmt.Sprintf("main moved %d times under the train", MaxReintegrations+1)}, ClosePR{}}, nil
		}
		n.Reintegrated++
		n.Phase = PhaseReintegrating
		return n, []Effect{Reintegrate{Prefix: n.Land}}, nil
	case Reintegrated:
		if t.Phase != PhaseReintegrating {
			break
		}
		if e.NeedSuite {
			// A new head: the old prefix results describe other commits.
			n.Results = map[int]bool{}
			n.Cars = n.Cars[:n.Land]
			n.Phase, n.Testing = PhaseTesting, n.Land
			n.Published = 0 // the reintegrated head is a new commit
			return n, append(n.publish(n.Land), RunSuite{Prefix: n.Land}), nil
		}
		n.Published = 0
		n.Phase = PhaseChecking
		return n, append(n.publish(n.Land), RunLandingChecks{Prefix: n.Land}), nil
	case Merged:
		if t.Phase != PhaseMerging {
			break
		}
		n.Phase = PhaseLanded
		return n, []Effect{LandCars{Seqs: carSeqs(n.Cars[:n.Land])}}, nil
	case MergeUnknown:
		if t.Phase != PhaseMerging {
			break
		}
		return n, []Effect{Probe{}}, nil
	case Probed:
		if t.Phase != PhaseMerging {
			break
		}
		switch {
		case e.Merged:
			n.Phase = PhaseLanded
			return n, []Effect{LandCars{Seqs: carSeqs(n.Cars[:n.Land])}}, nil
		case !t.Probed:
			n.Probed = true
			return n, []Effect{Merge{Prefix: n.Land}}, nil
		default:
			// Confirmed unmerged twice: give the cars back rather than loop.
			n.Phase = PhaseAbandoned
			n.Released = append(n.Released, carSeqs(n.Cars[:n.Land])...)
			return n, []Effect{Release{Seqs: carSeqs(n.Cars[:n.Land])}, ClosePR{}}, nil
		}
	}
	return t, nil, fmt.Errorf("train %d: %T is not legal while %s", t.ID, ev, t.Phase)
}

// split ends a bisect: hold the culprit, release the cars behind it to the
// next train (at once), and land the good prefix, or abandon when it's empty.
func split(n Train, culprit, lo int, reason string) (Train, []Effect, error) {
	c := n.Cars[culprit-1]
	n.Culprit = c.Seq
	effects := []Effect{Hold{Seq: c.Seq, Reason: reason}}
	if rest := carSeqs(n.Cars[culprit:]); len(rest) > 0 {
		n.Released = append(n.Released, rest...)
		effects = append(effects, Release{Seqs: rest})
	}
	return finishSplit(n, lo, effects)
}

func finishSplit(n Train, lo int, effects []Effect) (Train, []Effect, error) {
	if lo == 0 {
		n.Phase = PhaseAbandoned
		n.Cars = nil
		return n, append(effects, ClosePR{}), nil
	}
	n.Cars = n.Cars[:lo]
	n.Phase, n.Land = PhaseChecking, lo
	effects = append(effects, n.publish(lo)...)
	return n, append(effects, RunLandingChecks{Prefix: lo}), nil
}

// publish moves the train PR's head to prefix k when it is not there yet (a
// split lands a shorter prefix than was pushed).
func (t *Train) publish(k int) []Effect {
	if t.Published == k {
		return nil
	}
	t.Published = k
	return []Effect{Publish{Prefix: k}}
}

// partitions reports whether b accounts for every boarded car exactly once:
// kept, in boarding order, or conflicting. A car in neither would leave the
// train with no outcome, so such a report is rejected, never guessed at.
func partitions(cars []Car, b Built) bool {
	if len(b.Kept)+len(b.Conflicts) != len(cars) {
		return false
	}
	k := 0
	for _, c := range cars {
		_, conflict := b.Conflicts[c.Seq]
		switch {
		case conflict && k < len(b.Kept) && b.Kept[k] == c.Seq:
			return false
		case conflict:
		case k < len(b.Kept) && b.Kept[k] == c.Seq:
			k++
		default:
			return false
		}
	}
	return k == len(b.Kept)
}

func carSeqs(cars []Car) []int64 {
	out := make([]int64, len(cars))
	for i, c := range cars {
		out[i] = c.Seq
	}
	return out
}

func clone(m map[int]bool) map[int]bool {
	out := make(map[int]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
