package actualcorrection

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func testSpec() Spec {
	return Spec{Version: 1, Token: "correction-one", Repository: "repo-one", IssueID: "000335", CompletionToken: "close-one", EvidenceCommit: strings.Repeat("a", 40), ReviewedHEAD: strings.Repeat("b", 40), ExpectedCardBlob: strings.Repeat("c", 40), CandidateCardBlob: strings.Repeat("d", 40), MeasurementDigest: strings.Repeat("a", 64), Reason: "repair measured attribution", CalibrationDestination: "/tmp/calibration.tsv", TargetHashes: []string{strings.Repeat("b", 64)}, PreviousActual: "0.05", NewActual: "2.57", CalibrationRowDigest: strings.Repeat("c", 64), MirrorProjectionDigest: strings.Repeat("d", 64)}
}

func confirmation(s State, kind EventKind) Event {
	return Event{Kind: kind, Spec: s.Spec, TrackerCommit: s.TrackerCommit, CandidateCardBlob: s.Spec.CandidateCardBlob, CalibrationRowDigest: s.Spec.CalibrationRowDigest, MirrorProjectionDigest: s.Spec.MirrorProjectionDigest}
}

func step(t *testing.T, s State, e Event, phase Phase, effect Effect) State {
	t.Helper()
	next, got, err := Transition(s, e)
	if err != nil || next.Phase != phase || got != effect {
		t.Fatalf("%s in %s = %s, %s, %v; want %s, %s", e.Kind, s.Phase, next.Phase, got, err, phase, effect)
	}
	if !reflect.DeepEqual(next.Spec, s.Spec) {
		t.Fatal("transition changed immutable spec")
	}
	return next
}

func TestTransitionInterruptedCorrection(t *testing.T) {
	s := State{Spec: testSpec(), Phase: Prepared}
	s = step(t, s, Event{Kind: Apply}, Prepared, PublishTracker)
	s = step(t, s, confirmation(s, TrackerUncertain), UncertainTracker, ProbeTracker)
	s = step(t, s, Event{Kind: Repeat}, UncertainTracker, ProbeTracker)
	e := confirmation(s, TrackerConfirmed)
	e.TrackerCommit = strings.Repeat("e", 40)
	s = step(t, s, e, CalibrationPending, AppendCalibration)
	s = step(t, s, Event{Kind: Cancel}, CalibrationPending, None)
	s = step(t, s, Event{Kind: Repeat}, CalibrationPending, AppendCalibration)
	s = step(t, s, confirmation(s, CalibrationConfirmed), MirrorPending, PublishMirror)
	s = step(t, s, Event{Kind: Repeat}, MirrorPending, PublishMirror)
	s = step(t, s, confirmation(s, MirrorConfirmed), Complete, Finalize)
	step(t, s, Event{Kind: Repeat}, Complete, None)
}

func TestTransitionRefusesMismatchedConfirmation(t *testing.T) {
	s := State{Spec: testSpec(), Phase: Prepared}
	e := confirmation(s, TrackerConfirmed)
	e.TrackerCommit = strings.Repeat("e", 40)
	states := []State{s, step(t, s, e, CalibrationPending, AppendCalibration)}
	states = append(states, step(t, states[1], confirmation(states[1], CalibrationConfirmed), MirrorPending, PublishMirror))
	for i, state := range states {
		kind := []EventKind{TrackerConfirmed, CalibrationConfirmed, MirrorConfirmed}[i]
		t.Run(string(kind), func(t *testing.T) {
			fields := []string{"Token", "Repository", "IssueID", "CompletionToken", "EvidenceCommit", "ReviewedHEAD", "ExpectedCardBlob", "CandidateCardBlob", "MeasurementDigest", "Reason", "CalibrationDestination", "PreviousActual", "NewActual", "CalibrationRowDigest", "MirrorProjectionDigest"}
			for _, field := range fields {
				t.Run(field, func(t *testing.T) {
					event := confirmation(state, kind)
					if kind == TrackerConfirmed {
						event.TrackerCommit = strings.Repeat("e", 40)
					}
					value := reflect.ValueOf(&event.Spec).Elem().FieldByName(field)
					value.SetString(value.String() + "foreign")
					assertRefused(t, state, event)
				})
			}
			for _, mutate := range []func(*Event){func(e *Event) { e.Spec.Version++ }, func(e *Event) { e.Spec.Exclude = !e.Spec.Exclude }, func(e *Event) { e.Spec.TargetHashes = []string{strings.Repeat("f", 64)} }, func(e *Event) { e.CandidateCardBlob = strings.Repeat("f", 40) }, func(e *Event) { e.TrackerCommit = "" }} {
				event := confirmation(state, kind)
				if kind == TrackerConfirmed {
					event.TrackerCommit = strings.Repeat("e", 40)
				}
				mutate(&event)
				assertRefused(t, state, event)
			}
			if kind != TrackerConfirmed {
				event := confirmation(state, kind)
				event.TrackerCommit = strings.Repeat("f", 40)
				assertRefused(t, state, event)
			}
			if kind == CalibrationConfirmed {
				event := confirmation(state, kind)
				event.CalibrationRowDigest = strings.Repeat("f", 64)
				assertRefused(t, state, event)
			}
			if kind == MirrorConfirmed {
				event := confirmation(state, kind)
				event.MirrorProjectionDigest = strings.Repeat("f", 64)
				assertRefused(t, state, event)
			}
		})
	}
}

func assertRefused(t *testing.T, s State, e Event) {
	t.Helper()
	next, effect, err := Transition(s, e)
	if err == nil || effect != None || !reflect.DeepEqual(next, s) {
		t.Fatalf("invalid event changed state/effect: %+v, %s, %v", e, effect, err)
	}
}

func TestTransitionOrderingAndGenerationChange(t *testing.T) {
	prepared := State{Spec: testSpec(), Phase: Prepared}
	tracker := confirmation(prepared, TrackerConfirmed)
	tracker.TrackerCommit = strings.Repeat("e", 40)
	calibration := step(t, prepared, tracker, CalibrationPending, AppendCalibration)
	mirror := step(t, calibration, confirmation(calibration, CalibrationConfirmed), MirrorPending, PublishMirror)
	complete := step(t, mirror, confirmation(mirror, MirrorConfirmed), Complete, Finalize)
	for _, s := range []State{prepared, {Spec: testSpec(), Phase: UncertainTracker}, calibration, mirror, complete} {
		step(t, s, Event{Kind: Cancel}, s.Phase, None)
		for _, kind := range []EventKind{TrackerConfirmed, CalibrationConfirmed, MirrorConfirmed, TrackerUncertain} {
			if (kind == TrackerConfirmed || kind == TrackerUncertain) && (s.Phase == Prepared || s.Phase == UncertainTracker) || kind == CalibrationConfirmed && s.Phase == CalibrationPending || kind == MirrorConfirmed && s.Phase == MirrorPending {
				continue
			}
			assertRefused(t, s, confirmation(s, kind))
		}
		if s.Phase != Complete {
			assertRefused(t, s, Event{Kind: GenerationChanged})
		}
	}
}

func TestValidateStrictSchema(t *testing.T) {
	s := State{Spec: testSpec(), Phase: Prepared}
	if err := Validate(s); err != nil {
		t.Fatal(err)
	}
	mutations := []func(*State){func(s *State) { s.Spec.Version = 2 }, func(s *State) { s.Spec.Token = "" }, func(s *State) { s.Spec.IssueID = "0" }, func(s *State) { s.Spec.Reason = " " }, func(s *State) { s.Spec.ExpectedCardBlob = "bad" }, func(s *State) { s.Spec.MeasurementDigest = "bad" }, func(s *State) { s.Spec.TargetHashes = nil }, func(s *State) { s.Spec.TargetHashes = append(s.Spec.TargetHashes, s.Spec.TargetHashes[0]) }, func(s *State) { s.Spec.NewActual = "NaN" }, func(s *State) { s.Spec.NewActual = "-1" }, func(s *State) { s.Spec.Exclude = true }, func(s *State) { s.Phase = "unknown" }, func(s *State) { s.Phase = Complete }, func(s *State) { s.Spec.PreviousActual = "NaN" }}
	for i, mutate := range mutations {
		bad := s
		bad.Spec.TargetHashes = append([]string(nil), s.Spec.TargetHashes...)
		mutate(&bad)
		if Validate(bad) == nil {
			t.Errorf("accepted malformed state %d: %+v", i, bad)
		}
	}
	s.Spec.Exclude = true
	s.Spec.NewActual = "N/A"
	if err := Validate(s); err != nil {
		t.Fatal(err)
	}
	// A controller may retain its known candidate commit before confirmation.
	s.TrackerCommit = strings.Repeat("e", 40)
	if err := Validate(s); err != nil {
		t.Fatal(err)
	}
}

// Enumerate cancellation/restart counts at each effect boundary. Replaying the
// durable state never advances a phase until the exact effect is confirmed.
func TestTransitionDeterministicRestartSequences(t *testing.T) {
	for schedule := 0; schedule < 64; schedule++ {
		s := State{Spec: testSpec(), Phase: Prepared}
		for boundary := 0; boundary < 3; boundary++ {
			phase := []Phase{Prepared, CalibrationPending, MirrorPending}[boundary]
			effect := []Effect{PublishTracker, AppendCalibration, PublishMirror}[boundary]
			for restart := 0; restart < (schedule>>(2*boundary))&3; restart++ {
				s = step(t, s, Event{Kind: Repeat}, phase, effect)
				s = step(t, s, Event{Kind: Cancel}, phase, None)
				assertRefused(t, s, Event{Kind: GenerationChanged})
				raw, err := json.Marshal(s)
				if err != nil {
					t.Fatal(err)
				}
				var recovered State
				if err = json.Unmarshal(raw, &recovered); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(s, recovered) {
					t.Fatal("recovery changed state")
				}
				s = recovered
			}
			e := confirmation(s, []EventKind{TrackerConfirmed, CalibrationConfirmed, MirrorConfirmed}[boundary])
			if boundary == 0 {
				e.TrackerCommit = strings.Repeat("e", 40)
			}
			s = step(t, s, e, []Phase{CalibrationPending, MirrorPending, Complete}[boundary], []Effect{AppendCalibration, PublishMirror, Finalize}[boundary])
		}
		step(t, s, Event{Kind: Repeat}, Complete, None)
	}
}
