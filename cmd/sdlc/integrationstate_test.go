package main

import "testing"

func TestIntegrationStateOwnsTransitions(t *testing.T) {
	s := newIntegrationState("old", landingPR{HeadOID: "old"}, false)
	apply := func(e integrationEvent, want integrationStep) {
		t.Helper()
		next, effect, err := transitionIntegration(s, e)
		if err != nil || effect != want {
			t.Fatalf("event=%+v effect=%s err=%v", e, effect, err)
		}
		s = next
	}
	apply(integrationEvent{kind: integrationMainObserved, main: "main"}, integrationMerge)
	apply(integrationEvent{kind: integrationMerged, head: "merged"}, integrationReplay)
	if s.head != "merged" || s.oldHead != "old" || s.attempts != 0 {
		t.Fatalf("after merge: %+v", s)
	}
	apply(integrationEvent{kind: integrationReplayAccepted}, integrationPush)
	apply(integrationEvent{kind: integrationPushRegistered, pr: landingPR{HeadOID: "merged"}}, integrationGates)
	if s.attempts != 1 || s.resume || s.pr.HeadOID != s.head {
		t.Fatalf("after push: %+v", s)
	}
	apply(integrationEvent{kind: integrationGatesChecked, main: "main", included: true}, integrationLand)
	if _, _, err := transitionIntegration(s, integrationEvent{kind: integrationMerged, head: "wrong"}); err == nil {
		t.Fatal("accepted merge after landing")
	}
}

func TestIntegrationStateRecoveryAndBound(t *testing.T) {
	for _, resume := range []bool{false, true} {
		s := newIntegrationState("local", landingPR{HeadOID: "remote"}, resume)
		s, e, err := transitionIntegration(s, integrationEvent{kind: integrationMainObserved, main: "main"})
		if err != nil {
			t.Fatal(err)
		}
		if resume {
			if e != integrationReplay || s.oldHead != "remote" {
				t.Fatalf("resume %+v %s", s, e)
			}
		} else {
			s, e, err = transitionIntegration(s, integrationEvent{kind: integrationMergeFailed})
			if err != nil || e != integrationAbort {
				t.Fatalf("conflict %s %v", e, err)
			}
			continue
		}
		s, e, err = transitionIntegration(s, integrationEvent{kind: integrationReplayRejected})
		if err != nil || e != integrationReset {
			t.Fatalf("replay %s %v", e, err)
		}
	}
	s := newIntegrationState("h", landingPR{HeadOID: "h"}, false)
	for n := 0; n < integrationLimit; n++ {
		for _, ev := range []integrationEvent{{kind: integrationMainObserved, main: "m"}, {kind: integrationMerged, head: "new"}, {kind: integrationReplayAccepted}, {kind: integrationPushRegistered, pr: landingPR{HeadOID: "new"}}} {
			var err error
			s, _, err = transitionIntegration(s, ev)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	_, e, err := transitionIntegration(s, integrationEvent{kind: integrationGatesChecked, main: "next"})
	if err != nil || e != integrationRefuse {
		t.Fatalf("bound %s %v", e, err)
	}
	s = newIntegrationState("h", landingPR{HeadOID: "h"}, true)
	s, _, _ = transitionIntegration(s, integrationEvent{kind: integrationMainObserved, main: "m"})
	s, _, _ = transitionIntegration(s, integrationEvent{kind: integrationReplayAccepted})
	s, e, err = transitionIntegration(s, integrationEvent{kind: integrationPushFailed})
	if err != nil || e != integrationRefuse || s.head != "h" || s.attempts != 0 {
		t.Fatalf("uncertain push %+v %s %v", s, e, err)
	}
}

func TestIntegrationStateMainDriftBeforePushRequestsReset(t *testing.T) {
	s := newIntegrationState("old", landingPR{HeadOID: "old"}, false)
	for _, event := range []integrationEvent{
		{kind: integrationMainObserved, main: "main-1", included: false},
		{kind: integrationMerged, head: "merged"},
		{kind: integrationReplayAccepted},
	} {
		var err error
		s, _, err = transitionIntegration(s, event)
		if err != nil {
			t.Fatal(err)
		}
	}
	s, effect, err := transitionIntegration(s, integrationEvent{kind: integrationMainObserved, main: "main-2", included: false})
	if err != nil || effect != integrationStep("retry-reset") || s.oldHead != "old" || s.head != "merged" {
		t.Fatalf("pre-push drift state=%+v effect=%s err=%v", s, effect, err)
	}
	s, effect, err = transitionIntegration(s, integrationEvent{kind: integrationRecoveryCompleted})
	if err != nil || effect != integrationGates || s.head != "old" || s.attempts != 1 {
		t.Fatalf("retry after reset state=%+v effect=%s err=%v", s, effect, err)
	}
}
