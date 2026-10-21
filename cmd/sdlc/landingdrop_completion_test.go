package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// The effect's completion has been computed, but a real storage transaction
// accepts a drop before its commit. No scheduler timing or production hook
// determines which side of the fence wins.
func TestLandingDropPreemptsComputedCompletion(t *testing.T) {
	for _, kind := range []string{"bisect", "terminal", "phase", "abort"} {
		t.Run(kind, func(t *testing.T) {
			c, tr, cars := trainCompletionFixture(t)
			var ev landingline.TrainEvent
			switch kind {
			case "bisect":
				tr.Phase, tr.Testing = landingline.PhaseBisecting, 1
				tr.Results = map[int]bool{2: false, 3: false}
				ev = landingline.SuiteDone{Prefix: 1, Outcome: landingline.SuitePass}
			case "terminal":
				ev = landingline.SuiteDone{Prefix: 3, Outcome: landingline.SuiteMainRed}
			case "phase":
				ev = landingline.SuiteDone{Prefix: 3, Outcome: landingline.SuitePass}
			case "abort":
				ev = landingline.Aborted{Reason: "effect failed"}
			}
			control, err := c.line.UpdateTrain(c.handle, tr)
			if err != nil {
				t.Fatal(err)
			}
			next, _, err := landingline.Transition(tr, ev)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.line.RequestTrainDrop(c.id, control.Members[1], "completion drop"); err != nil {
				t.Fatal(err)
			}
			stopped := false
			stop := func() { stopped = true }
			if kind == "abort" {
				err = c.abortTrain(&tr, cars, stop, errors.New("effect failed"))
			} else {
				err = c.commitTrainCompletion(&tr, next, cars, stop)
			}
			var reform *trainReformError
			if !errors.As(err, &reform) {
				t.Fatalf("accepted drop did not re-form survivors: %v", err)
			}
			if !stopped || tr.Phase != landingline.PhaseAbandoned || len(tr.Cars) != 0 || !slices.Equal(tr.Released, []int64{cars[0].Seq, cars[2].Seq}) || c.selfEnd != landingline.OutcomeReleased {
				t.Fatalf("pre-completion riders lost: stopped=%v train=%+v self=%s", stopped, tr, c.selfEnd)
			}
			events, malformed, err := c.line.Events()
			if err != nil || malformed != 0 {
				t.Fatalf("outcome audit: malformed=%d err=%v", malformed, err)
			}
			outcomes := map[int64]string{}
			for _, event := range events {
				if event.Kind == landingline.EventCar {
					if outcomes[event.Seq] != "" {
						t.Fatalf("car %d settled twice", event.Seq)
					}
					outcomes[event.Seq] = event.Result
				}
			}
			if len(outcomes) != 3 || outcomes[cars[0].Seq] != string(landingline.OutcomeReleased) || outcomes[cars[1].Seq] != string(landingline.OutcomeHeld) || outcomes[cars[2].Seq] != string(landingline.OutcomeReleased) {
				t.Fatalf("obsolete completion effects changed outcomes: %v", outcomes)
			}
			stored, err := c.line.ReadTrainControl()
			if err != nil || stored.Cleanup.Pending || stored.Phase != landingline.PhaseAbandoned || len(stored.Members) != 0 || len(stored.CurrentDrops()) != 1 {
				t.Fatalf("drop cleanup was not positively sealed: %+v %v", stored, err)
			}
		})
	}
}

func TestLandingAbortCommitsBeforeEffects(t *testing.T) {
	c, tr, cars := trainCompletionFixture(t)
	cause := errors.New("native effect failed")
	if err := c.abortTrain(&tr, cars, func() {}, cause); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	stored, err := c.line.ReadTrainControl()
	if err != nil || stored.Phase != landingline.PhaseAbandoned || len(stored.Members) != 0 {
		t.Fatalf("abort effects ran without their terminal commit: %+v %v", stored, err)
	}
}

func trainCompletionFixture(t *testing.T) (*conductor, landingline.Train, []landingline.Car) {
	t.Helper()
	s := newTrainSetup(t, "train")
	s.blocker.Release()
	var handles []*landingline.Handle
	var cars []landingline.Car
	for i := 1; i <= 3; i++ {
		h, err := s.line.Enqueue(landingline.Ticket{Issue: i, Branch: fmt.Sprintf("completion-car-%d", i), QueueRoot: s.first})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(h.Release)
		handles = append(handles, h)
		cars = append(cars, landingline.Car{Seq: h.Ticket().Seq, Issue: i, Branch: h.Ticket().Branch, Head: s.heads[landingTestBranch], PR: i})
	}
	if err := s.line.Wait(context.Background(), handles[0]); err != nil {
		t.Fatal(err)
	}
	c := &conductor{ctx: context.Background(), r: landingTransportRunner{}, gh: s.gh, ci: s.gh, line: s.line, handle: handles[0], id: cars[0].Seq, self: cars[0].Seq,
		t: landingTarget{Root: s.first, Remote: "upstream", Repo: "test/repo"}, f: landingFlags(), stderr: io.Discard,
		log: &lineLog{line: s.line, stderr: io.Discard}, settled: map[int64]bool{}}
	if err := c.beginTrainControl(cars); err != nil {
		t.Fatal(err)
	}
	tr := landingline.NewTrain(c.id)
	tr.Cars, tr.Phase, tr.Testing = cars, landingline.PhaseTesting, 3
	if _, err := s.line.UpdateTrain(c.handle, tr); err != nil {
		t.Fatal(err)
	}
	return c, tr, cars
}
