package landingline

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"
)

func cars(n int) []Car {
	var out []Car
	for i := 1; i <= n; i++ {
		out = append(out, Car{Seq: int64(i), Issue: 100 + i, Branch: "b", Head: oid('a'), PR: i})
	}
	return out
}

// drive feeds events in order and returns the train and every effect, failing
// on a rejected event.
func drive(t *testing.T, tr Train, events ...TrainEvent) (Train, []Effect) {
	t.Helper()
	var all []Effect
	for _, ev := range events {
		var effects []Effect
		var err error
		tr, effects, err = Transition(tr, ev)
		if err != nil {
			t.Fatalf("%T rejected: %v", ev, err)
		}
		all = append(all, effects...)
	}
	return tr, all
}

func built(n int) Built {
	return Built{Kept: carSeqs(cars(n))}
}

func wantEffects(t *testing.T, got []Effect, want ...Effect) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("effects\n got %#v\nwant %#v", got, want)
	}
}

func TestTrainLandsAllCarsOnOneSuite(t *testing.T) {
	tr, effects := drive(t, NewTrain(1),
		Boarded{Cars: cars(3)}, built(3), SuiteDone{Prefix: 3, Outcome: SuitePass}, CIDone{Green: true}, Merged{OID: oid('m')})
	wantEffects(t, effects,
		Build{Cars: cars(3)}, Publish{Prefix: 3}, RunSuite{Prefix: 3}, WaitCI{Prefix: 3}, Merge{Prefix: 3}, LandCars{Seqs: []int64{1, 2, 3}})
	if tr.Phase != PhaseLanded || tr.Splits != 0 {
		t.Fatalf("train %+v", tr)
	}
}

func TestTrainHoldsAConflictingCar(t *testing.T) {
	tr, effects := drive(t, NewTrain(1), Boarded{Cars: cars(3)},
		Built{Kept: []int64{1, 3}, Conflicts: map[int64]string{2: "conflicts with car 1"}})
	wantEffects(t, effects, Build{Cars: cars(3)}, Hold{Seq: 2, Reason: "conflicts with car 1"}, Publish{Prefix: 2}, RunSuite{Prefix: 2})
	if got := carSeqs(tr.Cars); !slices.Equal(got, []int64{1, 3}) {
		t.Fatalf("cars %v", got)
	}
	tr, _ = drive(t, NewTrain(1), Boarded{Cars: cars(1)}, Built{Conflicts: map[int64]string{1: "conflicts with main"}})
	if tr.Phase != PhaseAbandoned {
		t.Fatalf("every car conflicted: %s", tr.Phase)
	}
}

func TestTrainSplitHoldsTheCulpritAndLandsThePrefix(t *testing.T) {
	// Four cars; car 3 breaks the suite. Prefix 2 passes, prefix 3 fails.
	tr, effects := drive(t, NewTrain(1), Boarded{Cars: cars(4)}, built(4),
		SuiteDone{Prefix: 4, Outcome: SuiteFail},
		SuiteDone{Prefix: 2, Outcome: SuitePass},
		SuiteDone{Prefix: 3, Outcome: SuiteFail})
	wantEffects(t, effects[3:],
		RunSuite{Prefix: 2}, RunSuite{Prefix: 3},
		Hold{Seq: 3, Reason: "the train's suite fails from car 3 (#103) on; the cars before it pass"},
		Release{Seqs: []int64{4}}, Publish{Prefix: 2}, WaitCI{Prefix: 2})
	if tr.Phase != PhaseLanding || tr.Land != 2 || tr.Culprit != 3 || tr.Splits != 2 || !slices.Equal(tr.Released, []int64{4}) {
		t.Fatalf("train %+v", tr)
	}
	tr, effects = drive(t, tr, CIDone{Green: true}, Merged{OID: oid('m')})
	wantEffects(t, effects, Merge{Prefix: 2}, LandCars{Seqs: []int64{1, 2}})
}

func TestTrainWhoseFirstCarFailsLandsNothing(t *testing.T) {
	tr, effects := drive(t, NewTrain(1), Boarded{Cars: cars(2)}, built(2),
		SuiteDone{Prefix: 2, Outcome: SuiteFail}, SuiteDone{Prefix: 1, Outcome: SuiteFail})
	wantEffects(t, effects[3:], RunSuite{Prefix: 1},
		Hold{Seq: 1, Reason: "the train's suite fails from car 1 (#101) on; the cars before it pass"},
		Release{Seqs: []int64{2}}, ClosePR{})
	if tr.Phase != PhaseAbandoned {
		t.Fatalf("phase %s", tr.Phase)
	}
}

func TestTrainOfOneFailingIsHeldWithoutASplitSuite(t *testing.T) {
	tr, effects := drive(t, NewTrain(1), Boarded{Cars: cars(1)}, built(1), SuiteDone{Prefix: 1, Outcome: SuiteFail})
	if tr.Phase != PhaseAbandoned || tr.Splits != 0 {
		t.Fatalf("train %+v", tr)
	}
	if !slices.ContainsFunc(effects, func(e Effect) bool { h, ok := e.(Hold); return ok && h.Seq == 1 }) {
		t.Fatalf("effects %#v", effects)
	}
}

func TestTrainMainRedRefusesEveryCar(t *testing.T) {
	tr, effects := drive(t, NewTrain(1), Boarded{Cars: cars(2)}, built(2), SuiteDone{Prefix: 2, Outcome: SuiteMainRed})
	wantEffects(t, effects[3:], Refuse{Seqs: []int64{1, 2}, Reason: "main is red: its baseline suite fails too"}, ClosePR{})
	if tr.Phase != PhaseRefused {
		t.Fatal(tr.Phase)
	}
}

func TestTrainCIRedRefusesTheWholeTrain(t *testing.T) {
	tr, effects := drive(t, NewTrain(1), Boarded{Cars: cars(2)}, built(2), SuiteDone{Prefix: 2, Outcome: SuitePass}, CIDone{Detail: "merge-check"})
	wantEffects(t, effects[len(effects)-2:], Refuse{Seqs: []int64{1, 2}, Reason: "the train's CI is red: merge-check"}, ClosePR{})
	if tr.Phase != PhaseRefused {
		t.Fatal(tr.Phase)
	}
}

func TestTrainMainMovingReintegrates(t *testing.T) {
	landing, _ := drive(t, NewTrain(1), Boarded{Cars: cars(2)}, built(2), SuiteDone{Prefix: 2, Outcome: SuitePass})
	t.Run("a docs-only move keeps the suite", func(t *testing.T) {
		tr, effects := drive(t, landing, MainMoved{}, Reintegrated{NeedSuite: false}, CIDone{Green: true})
		wantEffects(t, effects, Reintegrate{Prefix: 2}, Publish{Prefix: 2}, WaitCI{Prefix: 2}, Merge{Prefix: 2})
		if tr.Reintegrated != 1 {
			t.Fatal(tr.Reintegrated)
		}
	})
	t.Run("a code move reruns the suite on the new head", func(t *testing.T) {
		tr, effects := drive(t, landing, MainMoved{}, Reintegrated{NeedSuite: true})
		wantEffects(t, effects, Reintegrate{Prefix: 2}, Publish{Prefix: 2}, RunSuite{Prefix: 2})
		if tr.Phase != PhaseTesting || len(tr.Results) != 0 {
			t.Fatalf("train %+v", tr)
		}
	})
	t.Run("main moving under the merge reintegrates too", func(t *testing.T) {
		tr, _ := drive(t, landing, CIDone{Green: true}, MainMoved{})
		if tr.Phase != PhaseReintegrating {
			t.Fatal(tr.Phase)
		}
	})
	t.Run("bounded", func(t *testing.T) {
		tr := landing
		for range MaxReintegrations {
			tr, _ = drive(t, tr, MainMoved{}, Reintegrated{})
		}
		tr, effects := drive(t, tr, MainMoved{})
		if tr.Phase != PhaseRefused {
			t.Fatalf("phase %s effects %#v", tr.Phase, effects)
		}
	})
}

func TestTrainLostMergeReplyIsProbed(t *testing.T) {
	merging, _ := drive(t, NewTrain(1), Boarded{Cars: cars(2)}, built(2), SuiteDone{Prefix: 2, Outcome: SuitePass}, CIDone{Green: true})
	t.Run("it had merged", func(t *testing.T) {
		tr, effects := drive(t, merging, MergeUnknown{}, Probed{Merged: true, OID: oid('m')})
		wantEffects(t, effects, Probe{}, LandCars{Seqs: []int64{1, 2}})
		if tr.Phase != PhaseLanded {
			t.Fatal(tr.Phase)
		}
	})
	t.Run("it had not: one retry, then the cars go back", func(t *testing.T) {
		tr, effects := drive(t, merging, MergeUnknown{}, Probed{}, MergeUnknown{}, Probed{})
		wantEffects(t, effects, Probe{}, Merge{Prefix: 2}, Probe{}, Release{Seqs: []int64{1, 2}}, ClosePR{})
		if tr.Phase != PhaseAbandoned {
			t.Fatal(tr.Phase)
		}
	})
}

func TestTrainRejectsIllegalEvents(t *testing.T) {
	testing_, _ := drive(t, NewTrain(1), Boarded{Cars: cars(3)}, built(3))
	landed, _ := drive(t, testing_, SuiteDone{Prefix: 3, Outcome: SuitePass}, CIDone{Green: true}, Merged{})
	for name, c := range map[string]struct {
		tr Train
		ev TrainEvent
	}{
		"merged before the suite":     {testing_, Merged{}},
		"a stale suite result":        {testing_, SuiteDone{Prefix: 2, Outcome: SuitePass}},
		"CI before the suite":         {testing_, CIDone{Green: true}},
		"boarding twice":              {testing_, Boarded{Cars: cars(1)}},
		"anything after landing":      {landed, MainMoved{}},
		"an empty boarding":           {NewTrain(1), Boarded{}},
		"reintegrated without a move": {testing_, Reintegrated{}},
	} {
		got, effects, err := Transition(c.tr, c.ev)
		if err == nil || effects != nil || !reflect.DeepEqual(got, c.tr) {
			t.Errorf("%s: accepted (err %v, effects %#v)", name, err, effects)
		}
	}
}

// TestTrainInvariantsOverRandomHistories drives random event sequences
// through the production Transition and checks invariants stated
// independently of it:
//   - a car lands only inside a prefix whose suite passed on the head that
//     lands (or a docs-only reintegration of it), never a held car;
//   - every car ends exactly once: landed, held, refused or released;
//   - a terminal train accepts nothing.
func TestTrainInvariantsOverRandomHistories(t *testing.T) {
	rng := rand.New(rand.NewPCG(449, 1))
	for run := range 2000 {
		n := 1 + rng.IntN(6)
		tr, _, err := Transition(NewTrain(1), Boarded{Cars: cars(n)})
		if err != nil {
			t.Fatal(err)
		}
		ended := map[int64]string{}
		passed := map[int]bool{} // prefixes whose suite passed on the current head
		for step := 0; step < 40 && !tr.Phase.Terminal(); step++ {
			ev := randomEvent(rng, tr)
			next, effects, err := Transition(tr, ev)
			if err != nil {
				continue // a rejected event changes nothing (checked above)
			}
			if sd, ok := ev.(SuiteDone); ok && sd.Outcome == SuitePass {
				passed[sd.Prefix] = true
			}
			if ri, ok := ev.(Reintegrated); ok && ri.NeedSuite {
				passed = map[int]bool{}
			}
			for _, e := range effects {
				switch e := e.(type) {
				case LandCars:
					if !passed[len(e.Seqs)] {
						t.Fatalf("run %d: landed prefix %d with no passing suite (history %v)", run, len(e.Seqs), passed)
					}
					for _, s := range e.Seqs {
						end(t, ended, s, "landed")
					}
				case Hold:
					end(t, ended, e.Seq, "held")
				case Refuse:
					for _, s := range e.Seqs {
						end(t, ended, s, "refused")
					}
				case Release:
					for _, s := range e.Seqs {
						end(t, ended, s, "released")
					}
				}
			}
			tr = next
		}
		if tr.Phase.Terminal() {
			if len(ended) != n {
				t.Fatalf("run %d: a %s train of %d cars ended %d of them: %v", run, tr.Phase, n, len(ended), ended)
			}
			if _, _, err := Transition(tr, MainMoved{}); err == nil {
				t.Fatalf("run %d: a %s train accepted an event", run, tr.Phase)
			}
		}
	}
}

func end(t *testing.T, ended map[int64]string, seq int64, how string) {
	t.Helper()
	if was, ok := ended[seq]; ok {
		t.Fatalf("car %d ended twice: %s, then %s", seq, was, how)
	}
	ended[seq] = how
}

// randomEvent proposes a plausible next outcome for tr's phase, sometimes an
// illegal one, so rejection is exercised too.
func randomEvent(rng *rand.Rand, tr Train) TrainEvent {
	outcomes := []SuiteOutcome{SuitePass, SuitePass, SuiteFail, SuiteMainRed}
	switch {
	case rng.IntN(10) == 0:
		return []TrainEvent{Merged{}, CIDone{Green: true}, MainMoved{}, SuiteDone{Prefix: rng.IntN(8)}}[rng.IntN(4)]
	case tr.Phase == PhaseIntegrating:
		kept := []int64{}
		conflicts := map[int64]string{}
		for _, c := range tr.Cars {
			if rng.IntN(5) == 0 {
				conflicts[c.Seq] = "conflict"
			} else {
				kept = append(kept, c.Seq)
			}
		}
		return Built{Kept: kept, Conflicts: conflicts}
	case tr.Phase == PhaseTesting || tr.Phase == PhaseBisecting:
		return SuiteDone{Prefix: tr.Testing, Outcome: outcomes[rng.IntN(len(outcomes))]}
	case tr.Phase == PhaseLanding:
		return []TrainEvent{CIDone{Green: true}, CIDone{Green: true}, CIDone{Detail: "red"}, MainMoved{}}[rng.IntN(4)]
	case tr.Phase == PhaseReintegrating:
		return Reintegrated{NeedSuite: rng.IntN(2) == 0}
	case tr.Phase == PhaseMerging:
		return []TrainEvent{Merged{}, MergeUnknown{}, Probed{Merged: rng.IntN(2) == 0}, MainMoved{}}[rng.IntN(4)]
	}
	return MainMoved{}
}
