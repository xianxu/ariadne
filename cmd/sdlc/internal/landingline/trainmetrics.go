package landingline

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

// Train measurements (#561): what a train's one full suite found that its
// cars' affected runs did not ("escapes"), and what splitting cost. They are
// evidence for the TL's manual cap decisions, never landing authority: no
// gate, reducer or admission rule reads them.

// The checks a car's readiness can rest on (#561).
const (
	ProofAffected = "affected"
	ProofSuite    = "suite"
	ProofUnknown  = "unknown"
)

// BoardingProof is the record that satisfied a car's readiness: the check,
// the car head it vouched, the commit carrying the record (the head, or a
// docs-only ancestor) and the status context. Stamp binds it to the vouch it
// was observed for: an older sdlc that rewrites the vouch changes At and
// carries this object forward as an unknown member, so the mismatch makes
// the proof unknown rather than stale evidence.
type BoardingProof struct {
	Check   string    `json:"check"`
	Head    string    `json:"head"`
	Source  string    `json:"source"`
	Context string    `json:"context"`
	Stamp   time.Time `json:"stamp"`
	// Extra holds members a later sdlc wrote; a rewrite with the same binding
	// carries them forward, and they never override a known member.
	Extra map[string]json.RawMessage `json:"-"`
}

func proofFields(p *BoardingProof) map[string]any {
	return map[string]any{"check": &p.Check, "head": &p.Head, "source": &p.Source, "context": &p.Context, "stamp": &p.Stamp}
}

// UnmarshalJSON is the line's strict, preserving codec for the nested object.
func (p *BoardingProof) UnmarshalJSON(data []byte) error {
	*p = BoardingProof{}
	extra, err := decodeStrict(data, proofFields(p))
	if err != nil {
		return fmt.Errorf("boarding proof: %w", err)
	}
	p.Extra = extra
	return nil
}

// MarshalJSON writes the known members and then Extra.
func (p BoardingProof) MarshalJSON() ([]byte, error) {
	type known BoardingProof
	base, err := json.Marshal(known(p))
	if err != nil {
		return nil, err
	}
	fields := proofFields(&p)
	return encodeWithExtra(base, p.Extra, func(k string) bool { _, ok := fields[k]; return ok }), nil
}

// sameBinding reports whether q proves the same thing as p: the same check,
// head and source. Only then may q inherit p's attached members.
func (p BoardingProof) sameBinding(q BoardingProof) bool {
	return p.Check == q.Check && p.Head == q.Head && p.Source == q.Source
}

// carryProof is a re-vouch's proof (#561): a new binding starts over and
// carries nothing; the same binding keeps the members a later sdlc attached.
func carryProof(prev, next *BoardingProof) *BoardingProof {
	if next == nil || prev == nil || !prev.sameBinding(*next) {
		return next
	}
	p := *next
	p.Extra = mergeExtra(prev.Extra, next.Extra)
	return &p
}

// AcceptedProof is r's proof when it is bound to r: a ready vouch, a known
// check, r's own head, a full source commit and r's own observation stamp.
// Anything else is unknown proof; it never changes whether r is ready.
func (r Readiness) AcceptedProof() (BoardingProof, bool) {
	p := r.Proof
	if !r.Ready || p == nil || (p.Check != ProofAffected && p.Check != ProofSuite) ||
		p.Head != r.Head || !fullOID.MatchString(p.Source) || !p.Stamp.Equal(r.At) {
		return BoardingProof{}, false
	}
	return *p, true
}

// MeasurementVersion is the measurement payload's format. A reader shows
// another version as unmeasured, never as a healthy zero.
const MeasurementVersion = 1

// Per-record bounds (#561): reaching one marks the record incomplete and
// counts what was dropped, never claiming completeness.
const (
	MaxMeasuredCars       = 16
	MaxMeasuredExecutions = 128
	MaxMeasuredFailures   = 256
	MaxMeasuredString     = 512
	MaxMeasurementBytes   = 256 << 10
)

// ExecutionKind names why a whole managed suite ran on a train.
type ExecutionKind string

const (
	ExecInitial       ExecutionKind = "initial"       // the train's own suite on its full prefix
	ExecBaseline      ExecutionKind = "baseline"      // main's, once, on the first failure
	ExecSplitProbe    ExecutionKind = "split_probe"   // a bisect prefix
	ExecReintegration ExecutionKind = "reintegration" // after main's code moved under the train
)

// Execution outcomes. Unknown is a run that did not start, was cancelled or
// gave no result: it has no cost and leaves coverage incomplete.
const (
	ExecPassed  = "passed"
	ExecFailed  = "failed"
	ExecUnknown = "unknown"
)

// Car resolutions: unresolved (""), covered by a verified passing prefix, or
// confirmed culprit by exact adjacent-prefix attribution.
const (
	ResolutionPassed  = "passed"
	ResolutionEscaped = "escaped"
)

// MeasuredCar is one original car of a departed train.
type MeasuredCar struct {
	Seq        int64  `json:"seq"`
	Issue      int    `json:"issue,omitempty"`
	Head       string `json:"head"`
	Proof      string `json:"proof"`
	Resolution string `json:"resolution,omitempty"`
}

// EscapeFailure is one exact, unregistered failure in a package the culprit
// car changed.
type EscapeFailure struct {
	Seq  int64  `json:"seq"`
	Test string `json:"test"`
}

// ExecutionSummary is one managed whole-suite execution. Missing timing is
// an unknown cost, never zero.
type ExecutionSummary struct {
	Kind         ExecutionKind `json:"kind"`
	Outcome      string        `json:"outcome"`
	Attempt      string        `json:"attempt,omitempty"`
	RunSeconds   *float64      `json:"run_seconds,omitempty"`
	QueueSeconds *float64      `json:"queue_seconds,omitempty"`
}

// MeasureOverflow counts what a bound dropped.
type MeasureOverflow struct {
	Cars       int `json:"cars,omitempty"`
	Executions int `json:"executions,omitempty"`
	Failures   int `json:"failures,omitempty"`
}

// Measurement record kinds: one start once departure identities are fixed,
// one final on every controlled exit.
const (
	MeasureRecordStart = "start"
	MeasureRecordFinal = "final"
)

// TrainMeasurement is one train's measurement record.
type TrainMeasurement struct {
	Version          int                `json:"version"`
	Train            int64              `json:"train"`
	Record           string             `json:"record"`
	Departed         time.Time          `json:"departed"`
	Phase            string             `json:"phase,omitempty"`
	Cars             []MeasuredCar      `json:"cars"`
	Failures         []EscapeFailure    `json:"failures,omitempty"`
	Executions       []ExecutionSummary `json:"executions,omitempty"`
	SplitStart       time.Time          `json:"split_start,omitzero"`
	SplitWallSeconds *float64           `json:"split_wall_seconds,omitempty"`
	Overflow         *MeasureOverflow   `json:"overflow,omitempty"`
	Incomplete       []string           `json:"incomplete,omitempty"`
}

func (m TrainMeasurement) clone() TrainMeasurement {
	m.Cars = slices.Clone(m.Cars)
	m.Failures = slices.Clone(m.Failures)
	m.Executions = slices.Clone(m.Executions)
	m.Incomplete = slices.Clone(m.Incomplete)
	if m.Overflow != nil {
		o := *m.Overflow
		m.Overflow = &o
	}
	if m.SplitWallSeconds != nil {
		v := *m.SplitWallSeconds
		m.SplitWallSeconds = &v
	}
	return m
}

// MeasurementEvent is a fact the conductor observed for the tracker.
type MeasurementEvent interface{ measurementEvent() }

type (
	// MeasureDeparture fixes the original cars and their proofs.
	MeasureDeparture struct {
		Train    int64
		Departed time.Time
		Cars     []MeasuredCar
	}
	// MeasureStartAppended is the start record's append result.
	MeasureStartAppended struct{ Err error }
	// MeasureExecution is one managed whole-suite execution's returned
	// evidence. Finished is when its outcome was known (zero: unknown).
	MeasureExecution struct {
		Summary  ExecutionSummary
		Finished time.Time
	}
	// MeasurePrefixPassed: a verified passing suite covered these cars.
	MeasurePrefixPassed struct{ Seqs []int64 }
	// MeasureCulprit: exact adjacent-prefix attribution confirmed Seq as the
	// culprit for these exact failures.
	MeasureCulprit struct {
		Seq      int64
		Failures []string
	}
	// MeasureFinal is the train's controlled exit, with its terminal phase.
	MeasureFinal struct {
		At    time.Time
		Phase string
	}
)

func (MeasureDeparture) measurementEvent()     {}
func (MeasureStartAppended) measurementEvent() {}
func (MeasureExecution) measurementEvent()     {}
func (MeasurePrefixPassed) measurementEvent()  {}
func (MeasureCulprit) measurementEvent()       {}
func (MeasureFinal) measurementEvent()         {}

type measurePhase int

const (
	measureEmpty measurePhase = iota
	measureStarting
	measureCollecting
	measureFinalized
)

// MeasurementTracker accumulates one train's measurement. Its state is
// unexported: only AdvanceMeasurement changes it, so the conductor supplies
// facts and appends the snapshots it returns, and never edits counters.
type MeasurementTracker struct {
	phase    measurePhase
	m        TrainMeasurement
	attempts map[string]bool
	final    MeasureFinal
	pending  []string // incompleteness noted before departure
}

// AdvanceMeasurement is the tracker's one transition. It returns the record
// to append, if any (the start snapshot on departure, the final one on the
// first finalization), and a diagnostic for a rejected event. An illegal or
// contradictory fact before finalization is retained as sticky
// incompleteness; nothing is invented. A finalized tracker never changes.
func AdvanceMeasurement(t *MeasurementTracker, e MeasurementEvent) (*TrainMeasurement, error) {
	if t.phase == measureFinalized {
		if f, ok := e.(MeasureFinal); ok && f.At.Equal(t.final.At) && f.Phase == t.final.Phase {
			return nil, nil
		}
		return nil, fmt.Errorf("train measurement %d: %T after finalization rejected", t.m.Train, e)
	}
	illegal := func() (*TrainMeasurement, error) {
		t.incomplete(fmt.Sprintf("illegal %T while %s", e, t.phaseName()))
		return nil, nil
	}
	switch e := e.(type) {
	case MeasureDeparture:
		if t.phase != measureEmpty {
			return illegal()
		}
		t.m = TrainMeasurement{Version: MeasurementVersion, Train: e.Train, Departed: e.Departed}
		t.attempts = map[string]bool{}
		t.phase = measureStarting
		seen := map[int64]bool{}
		for _, c := range e.Cars {
			if seen[c.Seq] {
				t.incomplete(fmt.Sprintf("duplicate car %d", c.Seq))
				continue
			}
			seen[c.Seq] = true
			if len(t.m.Cars) >= MaxMeasuredCars {
				t.overflow().Cars++
				continue
			}
			if c.Proof != ProofAffected && c.Proof != ProofSuite {
				c.Proof = ProofUnknown
			}
			c.Head, c.Resolution = bounded(c.Head), ""
			t.m.Cars = append(t.m.Cars, c)
		}
		for _, why := range t.pending {
			t.incomplete(why)
		}
		t.pending = nil
		s := t.snapshot(MeasureRecordStart)
		return &s, nil
	case MeasureStartAppended:
		if t.phase != measureStarting {
			return illegal()
		}
		if e.Err != nil {
			t.incomplete("start record not written: " + e.Err.Error())
		}
		t.phase = measureCollecting
	case MeasureExecution:
		if t.phase != measureCollecting {
			return illegal()
		}
		x := e.Summary
		if x.Attempt != "" {
			if t.attempts[x.Attempt] {
				return nil, nil // the same execution observed twice counts once
			}
			t.attempts[x.Attempt] = true
		}
		switch x.Kind {
		case ExecInitial, ExecBaseline, ExecSplitProbe, ExecReintegration:
		default:
			return illegal()
		}
		switch x.Outcome {
		case ExecPassed, ExecFailed:
		case ExecUnknown:
			t.incomplete(string(x.Kind) + " execution gave no result")
			x.RunSeconds, x.QueueSeconds = nil, nil
		default:
			return illegal()
		}
		if !validSeconds(x.RunSeconds) || !validSeconds(x.QueueSeconds) {
			x.RunSeconds, x.QueueSeconds = nil, nil
		}
		x.Attempt = bounded(x.Attempt)
		if len(t.m.Executions) >= MaxMeasuredExecutions {
			t.overflow().Executions++
		} else {
			t.m.Executions = append(t.m.Executions, x)
		}
		if x.Kind == ExecInitial && x.Outcome == ExecFailed && t.m.SplitStart.IsZero() {
			if e.Finished.IsZero() {
				t.incomplete("split start time unknown")
			}
			t.m.SplitStart = e.Finished
		}
	case MeasurePrefixPassed:
		if t.phase != measureCollecting {
			return illegal()
		}
		for _, seq := range e.Seqs {
			i := t.car(seq)
			switch {
			case i < 0:
				t.incomplete(fmt.Sprintf("passing prefix names unknown car %d", seq))
			case t.m.Cars[i].Resolution == ResolutionEscaped:
				t.incomplete(fmt.Sprintf("car %d both escaped and passed", seq))
			default:
				t.m.Cars[i].Resolution = ResolutionPassed
			}
		}
	case MeasureCulprit:
		if t.phase != measureCollecting {
			return illegal()
		}
		i := t.car(e.Seq)
		switch {
		case i < 0:
			t.incomplete(fmt.Sprintf("culprit names unknown car %d", e.Seq))
			return nil, nil
		case len(e.Failures) == 0:
			t.incomplete(fmt.Sprintf("culprit %d has no exact failure", e.Seq))
			return nil, nil
		case t.m.Cars[i].Resolution == ResolutionPassed:
			t.incomplete(fmt.Sprintf("car %d both passed and escaped", e.Seq))
			return nil, nil
		case t.m.Cars[i].Resolution == ResolutionEscaped:
			return nil, nil // repeated confirmation of one boarding counts once
		}
		t.m.Cars[i].Resolution = ResolutionEscaped
		for _, f := range e.Failures {
			if f == "" || len(f) > MaxMeasuredString || len(t.m.Failures) >= MaxMeasuredFailures {
				t.overflow().Failures++
				continue
			}
			t.m.Failures = append(t.m.Failures, EscapeFailure{Seq: e.Seq, Test: f})
		}
	case MeasureFinal:
		if t.phase != measureStarting && t.phase != measureCollecting {
			return illegal()
		}
		if t.phase == measureStarting {
			t.incomplete("start record unacknowledged")
		}
		if !t.m.SplitStart.IsZero() && !e.At.IsZero() && !e.At.Before(t.m.SplitStart) {
			wall := e.At.Sub(t.m.SplitStart).Seconds()
			t.m.SplitWallSeconds = &wall
		}
		t.m.Phase = bounded(e.Phase)
		t.final, t.phase = e, measureFinalized
		s := t.snapshot(MeasureRecordFinal)
		return &s, nil
	default:
		return illegal()
	}
	return nil, nil
}

func (t *MeasurementTracker) phaseName() string {
	return [...]string{"empty", "starting", "collecting", "finalized"}[t.phase]
}

func (t *MeasurementTracker) incomplete(why string) {
	why = bounded(why)
	if t.phase == measureEmpty {
		t.pending = append(t.pending, why)
		return
	}
	if !slices.Contains(t.m.Incomplete, why) && len(t.m.Incomplete) < MaxMeasuredExecutions {
		t.m.Incomplete = append(t.m.Incomplete, why)
	}
}

func (t *MeasurementTracker) overflow() *MeasureOverflow {
	if t.m.Overflow == nil {
		t.m.Overflow = &MeasureOverflow{}
		t.incomplete("overflow: a per-train bound dropped evidence")
	}
	return t.m.Overflow
}

func (t *MeasurementTracker) car(seq int64) int {
	return slices.IndexFunc(t.m.Cars, func(c MeasuredCar) bool { return c.Seq == seq })
}

func (t *MeasurementTracker) snapshot(record string) TrainMeasurement {
	s := t.m.clone()
	s.Record = record
	return s
}

func bounded(s string) string {
	if len(s) > MaxMeasuredString {
		return s[:MaxMeasuredString]
	}
	return s
}

func validSeconds(v *float64) bool {
	return v == nil || (*v >= 0 && !math.IsNaN(*v) && !math.IsInf(*v, 0))
}

// EncodeMeasurement is the record's bytes for the log, bounded: an encoding
// over MaxMeasurementBytes is replaced by a minimal record that says so.
func EncodeMeasurement(m TrainMeasurement) (json.RawMessage, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if len(data) <= MaxMeasurementBytes {
		return data, nil
	}
	min := TrainMeasurement{Version: m.Version, Train: m.Train, Record: m.Record, Departed: m.Departed, Phase: m.Phase,
		Overflow:   &MeasureOverflow{Cars: len(m.Cars), Executions: len(m.Executions), Failures: len(m.Failures)},
		Incomplete: []string{fmt.Sprintf("record of %d bytes exceeded %d", len(data), MaxMeasurementBytes)}}
	return json.Marshal(min)
}

// TrainFilterStats is the measured filter quality over a window: the one
// fold status and the weekly artifact share (#561).
type TrainFilterStats struct {
	Version int       `json:"version"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
	// Trains are every multi-car train seen in the window: Measured have a
	// matched start and final, Incomplete have one without the other, a
	// conflict or a sticky incompleteness, Unmeasured have no usable
	// measurement (legacy trains, unsupported versions).
	Trains     int `json:"trains"`
	Measured   int `json:"measured"`
	Incomplete int `json:"incomplete"`
	Unmeasured int `json:"unmeasured"`
	// The escape rate is Escaped / ResolvedAffected; a zero denominator is
	// unknown. Unresolved affected boardings and other proofs sit beside it.
	AffectedBoardings  int `json:"affected_boardings"`
	ResolvedAffected   int `json:"resolved_affected"`
	Escaped            int `json:"escaped"`
	UnresolvedAffected int `json:"unresolved_affected"`
	SuiteBoardings     int `json:"suite_boardings"`
	UnknownProof       int `json:"unknown_proof"`
	EscapeFailures     int `json:"escape_failures"`
	// Managed whole-suite executions by kind. The split's extra suites are
	// baseline plus split probes; reintegration is a separate extra cost.
	InitialRuns       int     `json:"initial_runs"`
	BaselineRuns      int     `json:"baseline_runs"`
	SplitProbeRuns    int     `json:"split_probe_runs"`
	ReintegrationRuns int     `json:"reintegration_runs"`
	UnknownRuns       int     `json:"unknown_runs"`
	SplitRunSeconds   float64 `json:"split_run_seconds"`
	SplitQueueSeconds float64 `json:"split_queue_seconds"`
	ReintegrationRunSeconds float64 `json:"reintegration_run_seconds"`
	// SplitWallSeconds is elapsed recovery, first failed initial outcome to
	// the train's exit, including registered targeted reruns.
	SplitTrains      int     `json:"split_trains"`
	SplitWallSeconds float64 `json:"split_wall_seconds"`
	CostUnknown      int     `json:"cost_unknown"`
	Text             string  `json:"text"`
}

// TrainFilterStatsVersion is the stats payload's format.
const TrainFilterStatsVersion = 1

// measuredTrain joins one train's records.
type measuredTrain struct {
	terminal      bool
	at            time.Time
	start, final  *TrainMeasurement
	startRaw      json.RawMessage
	finalRaw      json.RawMessage
	conflict      bool
	unsupported   bool
}

// AggregateTrainMeasurements folds the line's events over [from, to): one
// pass over the events, then one over the trains it found.
func AggregateTrainMeasurements(events []Event, from, to time.Time) TrainFilterStats {
	return aggregateTrainMeasurements(events, from, to, func() {})
}

// aggregateTrainMeasurements is the fold, with visit counting each event and
// payload it touches (the linear-work tests count it).
func aggregateTrainMeasurements(events []Event, from, to time.Time, visit func()) TrainFilterStats {
	trains := map[int64]*measuredTrain{}
	var order []int64
	get := func(id int64) *measuredTrain {
		mt := trains[id]
		if mt == nil {
			mt = &measuredTrain{}
			trains[id] = mt
			order = append(order, id)
		}
		return mt
	}
	for _, e := range events {
		visit()
		switch e.Kind {
		case EventTrain:
			if e.Result == "alone" || e.Train == 0 {
				continue
			}
			mt := get(e.Train)
			mt.terminal = true
			if mt.at.IsZero() {
				mt.at = e.Departed
				if mt.at.IsZero() {
					mt.at = e.At
				}
			}
		case EventTrainMeasure:
			if e.Train == 0 {
				continue
			}
			mt := get(e.Train)
			if mt.at.IsZero() {
				mt.at = e.At // until a payload names the departure
			}
			var head struct {
				Version  int       `json:"version"`
				Record   string    `json:"record"`
				Departed time.Time `json:"departed"`
			}
			if json.Unmarshal(e.Measurement, &head) != nil || head.Version != MeasurementVersion {
				mt.unsupported = true
				continue
			}
			var m TrainMeasurement
			if json.Unmarshal(e.Measurement, &m) != nil || m.Train != e.Train {
				mt.unsupported = true
				continue
			}
			mt.at = m.Departed
			slot, raw := &mt.start, &mt.startRaw
			if m.Record == MeasureRecordFinal {
				slot, raw = &mt.final, &mt.finalRaw
			} else if m.Record != MeasureRecordStart {
				mt.unsupported = true
				continue
			}
			switch {
			case *slot == nil:
				*slot, *raw = &m, e.Measurement
			case string(*raw) != string(e.Measurement):
				mt.conflict = true
			}
		}
	}
	s := TrainFilterStats{Version: TrainFilterStatsVersion, From: from, To: to}
	for _, id := range order {
		mt := trains[id]
		visit()
		if mt.at.Before(from) || !mt.at.Before(to) {
			continue
		}
		s.Trains++
		switch {
		case mt.start == nil && mt.final == nil:
			s.Unmeasured++
			continue
		case mt.conflict || mt.unsupported || mt.start == nil || mt.final == nil || len(mt.final.Incomplete) > 0:
			s.Incomplete++
		default:
			s.Measured++
		}
		m := mt.final
		if m == nil || mt.conflict {
			m = mt.start // its cars, all unresolved
		}
		resolved := mt.final != nil && !mt.conflict
		for _, c := range m.Cars {
			visit()
			switch c.Proof {
			case ProofAffected:
				s.AffectedBoardings++
				switch {
				case !resolved || c.Resolution == "":
					s.UnresolvedAffected++
				case c.Resolution == ResolutionEscaped:
					s.ResolvedAffected++
					s.Escaped++
				default:
					s.ResolvedAffected++
				}
			case ProofSuite:
				s.SuiteBoardings++
			default:
				s.UnknownProof++
			}
		}
		if !resolved {
			continue
		}
		for _, f := range m.Failures {
			visit()
			if i := slices.IndexFunc(m.Cars, func(c MeasuredCar) bool { return c.Seq == f.Seq }); i >= 0 && m.Cars[i].Proof == ProofAffected {
				s.EscapeFailures++
			}
		}
		for _, x := range m.Executions {
			visit()
			switch x.Kind {
			case ExecInitial:
				s.InitialRuns++
			case ExecBaseline:
				s.BaselineRuns++
			case ExecSplitProbe:
				s.SplitProbeRuns++
			case ExecReintegration:
				s.ReintegrationRuns++
			}
			if x.Outcome == ExecUnknown {
				s.UnknownRuns++
				continue
			}
			if x.Kind == ExecInitial {
				continue
			}
			if x.RunSeconds == nil || x.QueueSeconds == nil {
				s.CostUnknown++
				continue
			}
			if x.Kind == ExecReintegration {
				s.ReintegrationRunSeconds += *x.RunSeconds
			} else {
				s.SplitRunSeconds += *x.RunSeconds
			}
			s.SplitQueueSeconds += *x.QueueSeconds
		}
		if !m.SplitStart.IsZero() {
			s.SplitTrains++
			if m.SplitWallSeconds == nil {
				s.CostUnknown++
			} else {
				s.SplitWallSeconds += *m.SplitWallSeconds
			}
		}
	}
	s.Text = FormatTrainFilterStats(s)
	return s
}

// FormatTrainFilterStats is the one human line for status and the weekly
// artifact: the escape fraction with its denominator, coverage beside it,
// and the split's actual cost.
func FormatTrainFilterStats(s TrainFilterStats) string {
	days := int(math.Round(s.To.Sub(s.From).Hours() / 24))
	rate := "rate unknown"
	if s.ResolvedAffected > 0 {
		rate = fmt.Sprintf("%.1f%%", 100*float64(s.Escaped)/float64(s.ResolvedAffected))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "train filter (%dd): escapes %d/%d resolved affected boardings (%s); %d unresolved, %d on suite, %d unknown proof",
		days, s.Escaped, s.ResolvedAffected, rate, s.UnresolvedAffected, s.SuiteBoardings, s.UnknownProof)
	fmt.Fprintf(&b, "; trains %d: %d measured, %d incomplete, %d unmeasured", s.Trains, s.Measured, s.Incomplete, s.Unmeasured)
	fmt.Fprintf(&b, "; split cost %d extra suite(s) (%d baseline, %d probe) %s run, %s queue, %s wall over %d split train(s); %d reintegration suite(s) %s",
		s.BaselineRuns+s.SplitProbeRuns, s.BaselineRuns, s.SplitProbeRuns, minutes(s.SplitRunSeconds), minutes(s.SplitQueueSeconds),
		minutes(s.SplitWallSeconds), s.SplitTrains, s.ReintegrationRuns, minutes(s.ReintegrationRunSeconds))
	if s.CostUnknown > 0 || s.UnknownRuns > 0 {
		fmt.Fprintf(&b, "; %d cost(s) unknown, %d run(s) without a result", s.CostUnknown, s.UnknownRuns)
	}
	return b.String()
}

func minutes(seconds float64) string {
	return fmt.Sprintf("%dm", int(math.Round(seconds/60)))
}
