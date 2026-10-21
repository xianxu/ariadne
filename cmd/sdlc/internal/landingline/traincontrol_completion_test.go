package landingline

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// The lock commit, rather than the conductor's earlier poll, determines
// whether a completion or an accepted drop owns the next transition.
func TestTrainCompletionAndDropOrdering(t *testing.T) {
	cases := []struct {
		name     string
		prepare  func(*Train) TrainEvent
		removed  bool
		terminal bool
	}{
		{name: "bisect membership reduction", removed: true, prepare: func(tr *Train) TrainEvent {
			tr.Phase = PhaseAttributing
			tr.AssessLo, tr.AssessHi = 1, 2
			return SplitAssessed{Lo: 1, Hi: 2, Verdict: SplitAttribution{Kind: AttributionAttributable, Reason: "affected test"}}
		}},
		{name: "enter attribution", prepare: func(tr *Train) TrainEvent {
			tr.Phase, tr.Testing = PhaseBisecting, 1
			tr.Results = map[int]bool{2: false}
			return SuiteDone{Prefix: 1, Outcome: SuitePass}
		}},
		{name: "terminal refusal", terminal: true, prepare: func(tr *Train) TrainEvent {
			tr.Testing = 2
			return SuiteDone{Prefix: 2, Outcome: SuiteMainRed}
		}},
		{name: "ordinary phase completion", prepare: func(tr *Train) TrainEvent {
			tr.Testing = 2
			return SuiteDone{Prefix: 2, Outcome: SuitePass}
		}},
	}
	for _, scenario := range cases {
		for _, dropFirst := range []bool{true, false} {
			order := "completion first"
			if dropFirst {
				order = "drop first"
			}
			t.Run(scenario.name+"/"+order, func(t *testing.T) {
				l, h, _, control := controlFixture(t)
				before := NewTrain(control.Train)
				before.Phase = PhaseTesting
				for _, m := range control.Members {
					before.Cars = append(before.Cars, m.Car())
				}
				event := scenario.prepare(&before)
				if _, err := l.UpdateTrain(h, before); err != nil {
					t.Fatal(err)
				}
				completed, _, err := Transition(before, event)
				if err != nil {
					t.Fatal(err)
				}
				member := control.Members[1]
				if dropFirst {
					if _, err := l.RequestTrainDrop(control.Train, member, "drop won"); err != nil {
						t.Fatal(err)
					}
					path := filepath.Join(l.dir, trainControlFilename)
					accepted, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					got, err := l.UpdateTrain(h, completed)
					if !errors.Is(err, ErrTrainDropPending) || got != nil {
						t.Fatalf("completion overwrote accepted drop: control=%+v err=%v", got, err)
					}
					retained, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					if string(accepted) != string(retained) {
						t.Fatal("refused completion mutated control")
					}
					assertCompletionDropApplied(t, l, h, before, member)
					return
				}
				if _, err := l.UpdateTrain(h, completed); err != nil {
					t.Fatal(err)
				}
				if scenario.removed || scenario.terminal {
					if _, err := l.RequestTrainDrop(control.Train, member, "completion lost"); err == nil {
						t.Fatal("removed or terminal member accepted after completion committed")
					}
					persisted, err := l.ReadTrainControl()
					if err != nil {
						t.Fatal(err)
					}
					if len(persisted.CurrentDrops()) != 0 || persisted.Cleanup.Pending {
						t.Fatal("completion-first refusal changed drop obligations")
					}
					if scenario.terminal {
						return
					}
					// The bisected-away member is no longer composed. Its surviving
					// conductor remains droppable against the committed shorter prefix.
					member = control.Members[0]
				}
				if _, err := l.RequestTrainDrop(control.Train, member, "surviving member"); err != nil {
					t.Fatal(err)
				}
				assertCompletionDropApplied(t, l, h, completed, member)
			})
		}
	}
}

func assertCompletionDropApplied(t *testing.T, l *Line, h *Handle, tr Train, member TrainMember) {
	t.Helper()
	next, effects, sealed, err := l.ApplyTrainDrops(h, tr)
	if err != nil {
		t.Fatalf("apply drop against authoritative membership: %v", err)
	}
	if next.Phase != PhaseAbandoned || sealed.Phase != PhaseAbandoned || !sealed.Cleanup.Pending {
		t.Fatalf("drop batch not sealed: %+v %+v", next, sealed)
	}
	var held, released []int64
	closed := 0
	for _, effect := range effects {
		switch e := effect.(type) {
		case Hold:
			held = append(held, e.Seq)
		case Release:
			released = append(released, e.Seqs...)
		case ClosePR:
			closed++
		case Merge:
			t.Fatal("accepted drop emitted merge")
		}
	}
	var expected []int64
	for _, car := range tr.Cars {
		if car.Seq != member.Seq {
			expected = append(expected, car.Seq)
		}
	}
	if !slices.Equal(held, []int64{member.Seq}) || !slices.Equal(released, expected) || closed != 1 {
		t.Fatalf("dropped/surviving identities changed: held=%v released=%v close=%d", held, released, closed)
	}
}
