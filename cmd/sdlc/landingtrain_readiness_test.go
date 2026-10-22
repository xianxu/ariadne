package main

import (
	"context"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// #516: a waiter can re-vouch while the conductor rechecks its own gates.
// Boarding must evaluate the sidecar against a clock sampled after reading it,
// while preserving the freshness checks for actually stale or future vouches.
func TestLandingTrainReadinessClockOrder(t *testing.T) {
	for _, tc := range []struct {
		name      string
		age       time.Duration // relative to the clock after the conductor's own gates
		wantTrain bool
	}{
		{name: "refreshed during own gates", wantTrain: true},
		{name: "stale", age: -landingline.MaxReadinessAge - time.Nanosecond},
		{name: "future", age: time.Nanosecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTrainSetup(t, "train")
			const secondBranch = "000002-second"
			s.pushCar(t, secondBranch, "second", 2, 2)
			ctx := context.Background()
			head, err := s.line.Enqueue(landingline.Ticket{Issue: 1, Branch: landingTestBranch})
			if err != nil {
				t.Fatal(err)
			}
			defer head.Release()
			waiter := s.queuedCar(t, secondBranch, 2, 2)
			s.blocker.Release()
			if err := s.line.Wait(ctx, head); err != nil {
				t.Fatal(err)
			}

			now := time.Now()
			initial := now
			oldNow := landingNow
			landingNow = func() time.Time { return now }
			t.Cleanup(func() { landingNow = oldNow })
			turn := &landingTurn{line: s.line, handle: head, log: &lineLog{line: s.line, stderr: io.Discard}, stderr: io.Discard}
			pr := s.gh.pr(landingTestBranch)
			self := func(context.Context) (landingline.Readiness, bool, error) {
				// This existing callback runs after maybeConductTrain starts deciding.
				// Replace the waiting car's sidecar synchronously: no load or sleeps
				// are needed to force the refresh between decision start and boarding.
				now = initial.Add(time.Millisecond)
				r := landingline.Readiness{Branch: secondBranch, Head: s.heads[secondBranch], PR: 2, Ready: true, At: now.Add(tc.age)}
				if err := s.line.Vouch(waiter, r); err != nil {
					t.Fatal(err)
				}
				return landingline.Readiness{Branch: landingTestBranch, Head: pr.HeadOID, PR: 1, Ready: true, At: now}, false, nil
			}
			target := landingTarget{Root: s.first, Remote: "upstream", Repo: "test/repo"}
			merged, trained, err := maybeConductTrain(ctx, execGitRunner{}, s.gh, s.gh, target, landingFlags(), turn, landingTestBranch, pr.HeadOID, pr, landingline.Schedule{Interval: time.Hour}, self, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			events := lineEvents(t, s.line)[landingline.EventTrain]
			if trained != tc.wantTrain {
				t.Fatalf("trained = %v, want %v; events: %+v", trained, tc.wantTrain, events)
			}
			if len(events) != 1 {
				t.Fatalf("want one train decision: %+v", events)
			}
			event := events[0]
			if !event.Departed.Equal(now) {
				t.Fatalf("departure = %s, want post-gate clock %s", event.Departed, now)
			}
			if !tc.wantTrain {
				if event.Result != "alone" || !slices.Equal(event.Cars, []int64{head.Ticket().Seq}) || s.gh.merges != 0 || len(s.gh.created) != 0 || merged.State != "OPEN" {
					t.Fatalf("invalid vouch boarded: PR %+v, merges %d, created %v, event %+v", merged, s.gh.merges, s.gh.created, event)
				}
				return
			}
			wantCars := []int64{head.Ticket().Seq, waiter.Ticket().Seq}
			if event.Result != string(landingline.PhaseLanded) || !slices.Equal(event.Cars, wantCars) || s.gh.merges != 1 || len(s.gh.created) != 1 || merged.State != "MERGED" {
				t.Fatalf("want one landed two-car train: PR %+v, merges %d, created %v, event %+v", merged, s.gh.merges, s.gh.created, event)
			}
			for _, branch := range []string{landingTestBranch, secondBranch} {
				if car := s.gh.pr(branch); car.State != "MERGED" {
					t.Fatalf("%s remains %s", branch, car.State)
				}
			}
		})
	}
}
