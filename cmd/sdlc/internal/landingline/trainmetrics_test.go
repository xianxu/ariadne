package landingline

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func secs(v float64) *float64 { return &v }

func twoCarDeparture() MeasureDeparture {
	return MeasureDeparture{Train: 7, Departed: tt0, Cars: []MeasuredCar{
		{Seq: 7, Issue: 1, Head: oid('a'), Proof: ProofAffected},
		{Seq: 8, Issue: 2, Head: oid('b'), Proof: ProofAffected},
	}}
}

// #561: a two-car escape, as the conductor feeds it: departure, start
// append, failed initial, baseline, probe passing car 1, attribution of car 2.
func TestAdvanceMeasurementTwoCarEscape(t *testing.T) {
	var tr MeasurementTracker
	start, err := AdvanceMeasurement(&tr, twoCarDeparture())
	if err != nil || start == nil || start.Record != MeasureRecordStart || len(start.Cars) != 2 {
		t.Fatalf("start %+v %v", start, err)
	}
	steps := []MeasurementEvent{
		MeasureStartAppended{},
		MeasureExecution{Summary: ExecutionSummary{Kind: ExecInitial, Outcome: ExecFailed, Attempt: "a1", RunSeconds: secs(600), QueueSeconds: secs(60)}, Finished: tt0.Add(10 * time.Minute)},
		MeasureExecution{Summary: ExecutionSummary{Kind: ExecBaseline, Outcome: ExecPassed, Attempt: "a2", RunSeconds: secs(500), QueueSeconds: secs(0)}, Finished: tt0.Add(20 * time.Minute)},
		MeasureExecution{Summary: ExecutionSummary{Kind: ExecSplitProbe, Outcome: ExecPassed, Attempt: "a3", RunSeconds: secs(400), QueueSeconds: secs(30)}, Finished: tt0.Add(30 * time.Minute)},
		MeasurePrefixPassed{Seqs: []int64{7}},
		MeasureCulprit{Seq: 8, Failures: []string{"example.com/m/leaf.TestF"}},
	}
	for _, e := range steps {
		if s, err := AdvanceMeasurement(&tr, e); s != nil || err != nil {
			t.Fatalf("%T: snapshot %v err %v", e, s, err)
		}
	}
	final, err := AdvanceMeasurement(&tr, MeasureFinal{At: tt0.Add(40 * time.Minute), Phase: "landed"})
	if err != nil || final == nil || final.Record != MeasureRecordFinal || len(final.Incomplete) != 0 {
		t.Fatalf("final %+v %v", final, err)
	}
	if final.Cars[0].Resolution != ResolutionPassed || final.Cars[1].Resolution != ResolutionEscaped || len(final.Failures) != 1 ||
		final.SplitWallSeconds == nil || *final.SplitWallSeconds != 30*60 || len(final.Executions) != 3 {
		t.Fatalf("final %+v", final)
	}
	// A finalized tracker never changes: the same finalization is a no-op,
	// anything else is rejected and the snapshot stays frozen.
	if s, err := AdvanceMeasurement(&tr, MeasureFinal{At: tt0.Add(40 * time.Minute), Phase: "landed"}); s != nil || err != nil {
		t.Fatalf("repeat final: %v %v", s, err)
	}
	if _, err := AdvanceMeasurement(&tr, MeasurePrefixPassed{Seqs: []int64{8}}); err == nil {
		t.Fatal("a fact after finalization was accepted")
	}
	if tr.m.Cars[1].Resolution != ResolutionEscaped {
		t.Fatal("finalized snapshot changed")
	}
	// The start snapshot is a copy: later accumulation did not reach it.
	if start.Cars[1].Resolution != "" || len(start.Executions) != 0 {
		t.Fatalf("start snapshot aliased: %+v", start)
	}
}

// #561: sticky incompleteness. A failed start append, an execution without
// a result, a contradiction or an illegal event leave valid facts but never
// a complete record; nothing is invented.
func TestAdvanceMeasurementStickyIncomplete(t *testing.T) {
	for name, events := range map[string][]MeasurementEvent{
		"start append failed":     {MeasureStartAppended{Err: errors.New("disk full")}},
		"unknown execution":       {MeasureStartAppended{}, MeasureExecution{Summary: ExecutionSummary{Kind: ExecInitial, Outcome: ExecUnknown}}},
		"contradiction":           {MeasureStartAppended{}, MeasureCulprit{Seq: 7, Failures: []string{"p.TestX"}}, MeasurePrefixPassed{Seqs: []int64{7}}},
		"unknown car":             {MeasureStartAppended{}, MeasurePrefixPassed{Seqs: []int64{99}}},
		"illegal order":           {MeasureExecution{Summary: ExecutionSummary{Kind: ExecInitial, Outcome: ExecPassed}}, MeasureStartAppended{}},
		"no start ack":            {},
		"culprit without failure": {MeasureStartAppended{}, MeasureCulprit{Seq: 8}},
	} {
		t.Run(name, func(t *testing.T) {
			var tr MeasurementTracker
			if _, err := AdvanceMeasurement(&tr, twoCarDeparture()); err != nil {
				t.Fatal(err)
			}
			for _, e := range events {
				if _, err := AdvanceMeasurement(&tr, e); err != nil {
					t.Fatalf("%T: %v", e, err)
				}
			}
			final, _ := AdvanceMeasurement(&tr, MeasureFinal{At: tt0.Add(time.Hour), Phase: "refused"})
			if final == nil || len(final.Incomplete) == 0 {
				t.Fatalf("complete after %s: %+v", name, final)
			}
			for _, c := range final.Cars {
				if c.Resolution == ResolutionEscaped && len(final.Failures) == 0 {
					t.Fatalf("an escape without its failure: %+v", final)
				}
			}
		})
	}
}

// #561 PQ-2: arbitrary event orders never resolve a car without a passing
// prefix or culprit fact, never cost an execution that was not observed,
// never change after finalization, and never shed incompleteness.
func TestAdvanceMeasurementRandomSequences(t *testing.T) {
	rng := rand.New(rand.NewSource(561))
	pool := []MeasurementEvent{
		twoCarDeparture(), MeasureStartAppended{}, MeasureStartAppended{Err: errors.New("x")},
		MeasureExecution{Summary: ExecutionSummary{Kind: ExecInitial, Outcome: ExecFailed, Attempt: "i", RunSeconds: secs(1), QueueSeconds: secs(1)}, Finished: tt0},
		MeasureExecution{Summary: ExecutionSummary{Kind: ExecSplitProbe, Outcome: ExecUnknown, Attempt: "p"}},
		MeasureExecution{Summary: ExecutionSummary{Kind: "bogus", Outcome: ExecPassed}},
		MeasurePrefixPassed{Seqs: []int64{7}}, MeasureCulprit{Seq: 8, Failures: []string{"p.TestY"}},
		MeasureCulprit{Seq: 7, Failures: []string{"p.TestZ"}},
		MeasureFinal{At: tt0.Add(time.Hour), Phase: "landed"},
	}
	for round := 0; round < 2000; round++ {
		var tr MeasurementTracker
		var frozen *TrainMeasurement
		sticky := false
		observed := map[string]bool{}
		resolvedBy := map[int64]bool{}
		for range 12 {
			e := pool[rng.Intn(len(pool))]
			s, _ := AdvanceMeasurement(&tr, e)
			if tr.phase != measureFinalized {
				switch e := e.(type) {
				case MeasureExecution:
					if tr.phase == measureCollecting {
						observed[e.Summary.Attempt] = true
					}
				case MeasurePrefixPassed:
					for _, s := range e.Seqs {
						resolvedBy[s] = true
					}
				case MeasureCulprit:
					resolvedBy[e.Seq] = true
				}
			}
			if s != nil && s.Record == MeasureRecordFinal {
				frozen = s
			}
			if len(tr.m.Incomplete) > 0 {
				sticky = true
			} else if sticky {
				t.Fatalf("round %d: incompleteness was shed", round)
			}
		}
		for _, c := range tr.m.Cars {
			if c.Resolution != "" && !resolvedBy[c.Seq] {
				t.Fatalf("round %d: car %d resolved without a fact", round, c.Seq)
			}
		}
		for _, x := range tr.m.Executions {
			if !observed[x.Attempt] {
				t.Fatalf("round %d: execution %q costed without an observation", round, x.Attempt)
			}
		}
		if frozen != nil {
			got, _ := json.Marshal(tr.m)
			again, _ := json.Marshal(frozen.clone())
			var a, b TrainMeasurement
			_ = json.Unmarshal(got, &a)
			_ = json.Unmarshal(again, &b)
			a.Record, b.Record = "", ""
			x, _ := json.Marshal(a)
			y, _ := json.Marshal(b)
			if string(x) != string(y) {
				t.Fatalf("round %d: finalized state changed:\n%s\n%s", round, x, y)
			}
		}
	}
}

// #561: bounds mark overflow and incompleteness rather than dropping silently.
func TestAdvanceMeasurementBounds(t *testing.T) {
	var cars []MeasuredCar
	for i := range MaxMeasuredCars + 2 {
		cars = append(cars, MeasuredCar{Seq: int64(i + 1), Head: oid('a'), Proof: ProofAffected})
	}
	var tr MeasurementTracker
	_, _ = AdvanceMeasurement(&tr, MeasureDeparture{Train: 1, Departed: tt0, Cars: cars})
	_, _ = AdvanceMeasurement(&tr, MeasureStartAppended{})
	var fails []string
	for i := range MaxMeasuredFailures + 3 {
		fails = append(fails, fmt.Sprintf("p.Test%d", i))
	}
	_, _ = AdvanceMeasurement(&tr, MeasureCulprit{Seq: 1, Failures: fails})
	final, _ := AdvanceMeasurement(&tr, MeasureFinal{At: tt0, Phase: "landed"})
	if len(final.Cars) != MaxMeasuredCars || final.Overflow == nil || final.Overflow.Cars != 2 || final.Overflow.Failures != 3 || len(final.Incomplete) == 0 {
		t.Fatalf("%+v", final)
	}
	raw, err := EncodeMeasurement(*final)
	if err != nil || len(raw) > MaxMeasurementBytes {
		t.Fatalf("%d bytes %v", len(raw), err)
	}
}

func measureEvent(t *testing.T, m TrainMeasurement) Event {
	t.Helper()
	raw, err := EncodeMeasurement(m)
	if err != nil {
		t.Fatal(err)
	}
	return Event{Kind: EventTrainMeasure, At: m.Departed, Train: m.Train, Measurement: raw}
}

// escapeTrain is a finished two-car train whose second car escaped.
func escapeTrain(t *testing.T, id int64, at time.Time) []Event {
	t.Helper()
	var tr MeasurementTracker
	d := twoCarDeparture()
	d.Train, d.Departed = id, at
	start, _ := AdvanceMeasurement(&tr, d)
	_, _ = AdvanceMeasurement(&tr, MeasureStartAppended{})
	_, _ = AdvanceMeasurement(&tr, MeasureExecution{Summary: ExecutionSummary{Kind: ExecInitial, Outcome: ExecFailed, Attempt: "i", RunSeconds: secs(600), QueueSeconds: secs(0)}, Finished: at.Add(10 * time.Minute)})
	_, _ = AdvanceMeasurement(&tr, MeasureExecution{Summary: ExecutionSummary{Kind: ExecBaseline, Outcome: ExecPassed, Attempt: "b", RunSeconds: secs(300), QueueSeconds: secs(60)}, Finished: at.Add(20 * time.Minute)})
	_, _ = AdvanceMeasurement(&tr, MeasurePrefixPassed{Seqs: []int64{7}})
	_, _ = AdvanceMeasurement(&tr, MeasureCulprit{Seq: 8, Failures: []string{"p.TestA", "p.TestB"}})
	final, _ := AdvanceMeasurement(&tr, MeasureFinal{At: at.Add(30 * time.Minute), Phase: "landed"})
	return []Event{measureEvent(t, *start), measureEvent(t, *final),
		{Kind: EventTrain, At: at.Add(30 * time.Minute), Train: id, Seq: id, Result: "landed", Departed: at}}
}

// #561: the fold. One escape over two resolved affected boardings; a legacy
// train is unmeasured; a start without a final is incomplete with its cars
// unresolved; an identical duplicate folds once; a conflicting one is
// incomplete; an unsupported version is unmeasured; the window is half-open.
func TestAggregateTrainMeasurements(t *testing.T) {
	from, to := tt0, tt0.Add(7*24*time.Hour)
	events := escapeTrain(t, 7, tt0.Add(time.Hour))
	events = append(events, events[1]) // identical duplicate final
	events = append(events, Event{Kind: EventTrain, Train: 20, Seq: 20, Result: "landed", Departed: tt0.Add(2 * time.Hour)})
	events = append(events, Event{Kind: EventTrain, Train: 21, Seq: 21, Result: "alone", Departed: tt0.Add(2 * time.Hour)})
	var tr MeasurementTracker
	d := twoCarDeparture()
	d.Train, d.Departed = 30, tt0.Add(3*time.Hour)
	start, _ := AdvanceMeasurement(&tr, d)
	events = append(events, measureEvent(t, *start))
	conflicted := escapeTrain(t, 40, tt0.Add(4*time.Hour))
	other := conflicted[1]
	other.Measurement = json.RawMessage(strings.Replace(string(other.Measurement), `"phase":"landed"`, `"phase":"refused"`, 1))
	events = append(events, append(conflicted, other)...)
	events = append(events, Event{Kind: EventTrainMeasure, At: tt0.Add(5 * time.Hour), Train: 50, Measurement: json.RawMessage(`{"version":2,"train":50}`)})
	events = append(events, escapeTrain(t, 60, to)...) // outside: the window is half-open
	s := AggregateTrainMeasurements(events, from, to)
	want := TrainFilterStats{Trains: 5, Measured: 1, Incomplete: 2, Unmeasured: 2,
		AffectedBoardings: 6, ResolvedAffected: 2, Escaped: 1, UnresolvedAffected: 4, EscapeFailures: 2,
		InitialRuns: 1, BaselineRuns: 1, SplitRunSeconds: 300, SplitQueueSeconds: 60, SplitTrains: 1, SplitWallSeconds: 20 * 60}
	got := s
	got.Version, got.From, got.To, got.Text = 0, time.Time{}, time.Time{}, ""
	if got != want {
		t.Fatalf("stats\n got %+v\nwant %+v", got, want)
	}
	for _, phrase := range []string{"train filter (7d)", "escapes 1/2 resolved affected boardings (50.0%)", "4 unresolved", "trains 5: 1 measured, 2 incomplete, 2 unmeasured", "1 extra suite(s) (1 baseline, 0 probe) 5m run, 1m queue, 20m wall"} {
		if !strings.Contains(s.Text, phrase) {
			t.Errorf("text lacks %q: %s", phrase, s.Text)
		}
	}
	if empty := AggregateTrainMeasurements(nil, from, to); !strings.Contains(empty.Text, "escapes 0/0 resolved affected boardings (rate unknown)") {
		t.Fatalf("an empty window is not unknown: %s", empty.Text)
	}
}

// #561 cost model: the fold visits each event once plus each payload's
// bounded members once; N and 2N trains cost exactly twice as much.
func TestTrainMetricAggregationWorkIsLinear(t *testing.T) {
	work := func(n int) int {
		var events []Event
		for i := range n {
			events = append(events, escapeTrain(t, int64(100+i), tt0.Add(time.Duration(i)*time.Minute))...)
		}
		visits := 0
		aggregateTrainMeasurements(events, tt0, tt0.Add(7*24*time.Hour), func() { visits++ })
		return visits
	}
	w8, w16 := work(8), work(16)
	if w16 != 2*w8 {
		t.Fatalf("visits %d at 8 trains, %d at 16: not linear", w8, w16)
	}
	if per := w8 / 8; per > 3+1+2+2+2 {
		t.Fatalf("%d visits per train: more than one pass", per)
	}
}

// #561 PQ-3: the nested proof is strict and preserving: duplicate, null,
// mistyped and trailing members refuse; a future member survives an older
// proof-aware rewrite and never overrides a known one.
func TestBoardingProofCodec(t *testing.T) {
	good := `{"check":"affected","head":"` + oid('a') + `","source":"` + oid('a') + `","context":"sdlc-local/affected","stamp":"2026-10-19T09:00:00Z","later":{"x":1}}`
	var p BoardingProof
	if err := json.Unmarshal([]byte(good), &p); err != nil || p.Check != ProofAffected || string(p.Extra["later"]) != `{"x":1}` {
		t.Fatalf("%+v %v", p, err)
	}
	out, _ := json.Marshal(p)
	if !strings.Contains(string(out), `"later":{"x":1}`) {
		t.Fatalf("future member dropped: %s", out)
	}
	p.Extra["check"] = json.RawMessage(`"suite"`)
	if out, _ := json.Marshal(p); strings.Count(string(out), `"check"`) != 1 || !strings.Contains(string(out), `"check":"affected"`) {
		t.Fatalf("extra overrode a known member: %s", out)
	}
	for _, bad := range []string{
		`{"check":"affected","check":"suite"}`,
		`{"check":null}`,
		`{"check":1}`,
		`{"check":"affected"} {}`,
		`[]`,
	} {
		if err := json.Unmarshal([]byte(bad), &p); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
	// Readiness carries it as a known member through its own codec.
	r := Readiness{Version: ReadinessVersion, Seq: 3, Branch: "b", Head: oid('a'), PR: 3, Ready: true, At: tt0,
		Proof: &BoardingProof{Check: ProofSuite, Head: oid('a'), Source: oid('c'), Context: "sdlc-local/suite", Stamp: tt0}}
	data, _ := json.Marshal(r)
	back, err := DecodeReadiness(data)
	if err != nil || back.Proof == nil || back.Proof.Source != oid('c') {
		t.Fatalf("%s: %+v %v", data, back, err)
	}
	if p, ok := back.AcceptedProof(); !ok || p.Check != ProofSuite {
		t.Fatalf("bound proof not accepted: %+v", back.Proof)
	}
	if _, err := DecodeReadiness([]byte(strings.Replace(string(data), `"check":"suite"`, `"check":"suite","check":"affected"`, 1))); err == nil {
		t.Fatal("a duplicate nested member decoded")
	}
}

// #561: an older sdlc's rewrite (At changes, proof carried as an unknown
// member) or another head leaves the proof unknown; readiness itself is
// unaffected.
func TestAcceptedProofBinding(t *testing.T) {
	base := Readiness{Version: ReadinessVersion, Seq: 3, Branch: "b", Head: oid('a'), PR: 3, Ready: true, At: tt0,
		Proof: &BoardingProof{Check: ProofAffected, Head: oid('a'), Source: oid('a'), Context: "sdlc-local/affected", Stamp: tt0}}
	if _, ok := base.AcceptedProof(); !ok {
		t.Fatal("bound proof rejected")
	}
	for name, mutate := range map[string]func(*Readiness){
		"older rewrite": func(r *Readiness) { r.At = tt0.Add(time.Minute) },
		"another head":  func(r *Readiness) { r.Head = oid('b') },
		"unknown check": func(r *Readiness) { p := *r.Proof; p.Check = "e2e"; r.Proof = &p },
		"short source":  func(r *Readiness) { p := *r.Proof; p.Source = "abc"; r.Proof = &p },
		"not ready":     func(r *Readiness) { r.Ready = false },
		"no proof":      func(r *Readiness) { r.Proof = nil },
	} {
		r := base
		mutate(&r)
		if _, ok := r.AcceptedProof(); ok {
			t.Errorf("%s: proof accepted", name)
		}
		if name != "not ready" && r.Valid(3) != nil && name != "another head" {
			t.Errorf("%s: proof affected readiness validity: %v", name, r.Valid(3))
		}
	}
}

// #561 PQ-3: Vouch stamps the proof with its own observation, keeps a later
// sdlc's nested members across a re-vouch of the same binding, starts a new
// binding over, and an older sdlc's rewrite (proof carried as an unknown
// member under a new At) leaves the proof unbound, never the vouch invalid.
func TestVouchProofBinding(t *testing.T) {
	dir := t.TempDir()
	l := openLine(t, dir, testOptions(nil))
	h := enqueue(t, l, 1)
	seq := h.Ticket().Seq
	proof := func(check, source string) *BoardingProof {
		return &BoardingProof{Check: check, Head: oid('a'), Source: source, Context: "sdlc-local/" + check}
	}
	read := func() Readiness {
		t.Helper()
		data, _ := os.ReadFile(readyPath(dir, seq))
		r, err := DecodeReadiness(data)
		if err != nil {
			t.Fatalf("%s: %v", data, err)
		}
		return r
	}
	if err := l.Vouch(h, Readiness{Branch: "b", Head: oid('a'), PR: 1, Ready: true, At: tt0, Proof: proof(ProofAffected, oid('a'))}); err != nil {
		t.Fatal(err)
	}
	if p, ok := read().AcceptedProof(); !ok || !p.Stamp.Equal(tt0) {
		t.Fatalf("first vouch: %+v %v", p, ok)
	}
	// A later sdlc attached a nested member; the same binding keeps it.
	data, _ := os.ReadFile(readyPath(dir, seq))
	data = []byte(strings.Replace(string(data), `"context":"sdlc-local/affected"`, `"context":"sdlc-local/affected","selection":"3 packages"`, 1))
	os.WriteFile(readyPath(dir, seq), data, 0o644)
	if err := l.Vouch(h, Readiness{Branch: "b", Head: oid('a'), PR: 1, Ready: true, At: tt0.Add(time.Minute), Proof: proof(ProofAffected, oid('a'))}); err != nil {
		t.Fatal(err)
	}
	r := read()
	if p, ok := r.AcceptedProof(); !ok || string(p.Extra["selection"]) != `"3 packages"` || !p.Stamp.Equal(tt0.Add(time.Minute)) {
		t.Fatalf("same binding: %+v %v", r.Proof, ok)
	}
	// A new binding (suite on a docs-only ancestor) carries nothing.
	if err := l.Vouch(h, Readiness{Branch: "b", Head: oid('a'), PR: 1, Ready: true, At: tt0.Add(2 * time.Minute), Proof: proof(ProofSuite, oid('c'))}); err != nil {
		t.Fatal(err)
	}
	if r := read(); r.Proof == nil || r.Proof.Extra != nil || r.Proof.Check != ProofSuite {
		t.Fatalf("new binding kept old members: %+v", r.Proof)
	}
	// An older sdlc's rewrite: At moves, the proof rides as an unknown member.
	data, _ = os.ReadFile(readyPath(dir, seq))
	data = []byte(strings.Replace(string(data), `"at":"`+tt0.Add(2*time.Minute).Format(time.RFC3339)+`"`, `"at":"`+tt0.Add(3*time.Minute).Format(time.RFC3339)+`"`, 1))
	os.WriteFile(readyPath(dir, seq), data, 0o644)
	r = read()
	if r.Valid(seq) != nil || !r.Ready {
		t.Fatalf("the proof affected readiness: %v", r.Valid(seq))
	}
	if _, ok := r.AcceptedProof(); ok {
		t.Fatal("a proof from another observation was accepted")
	}
	// A vouch with no proof (CI waived) drops the old one.
	if err := l.Vouch(h, Readiness{Branch: "b", Head: oid('a'), PR: 1, Ready: true, At: tt0.Add(4 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	if r := read(); r.Proof != nil {
		t.Fatalf("an unproven vouch kept a proof: %+v", r.Proof)
	}
}
