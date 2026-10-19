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

func TestIntegrationStateIssueOnlyProgress(t *testing.T) {
	for _, ready := range []bool{false, true} {
		s := newIntegrationState("tested", landingPR{HeadOID: "tested"}, false)
		s.attempts = integrationLimit
		next, effect, err := transitionIntegration(s, integrationEvent{kind: integrationMainObserved, main: "issue-main", issueOnly: true})
		if err != nil || effect != integrationGates || next.attempts != integrationLimit {
			t.Fatalf("metadata observation at exhausted budget: %+v %s %v", next, effect, err)
		}
		next, effect, err = transitionIntegration(next, integrationEvent{kind: integrationGatesChecked, main: "issue-main", issueOnly: true, ready: ready})
		if ready {
			if err != nil || effect != integrationLand || next.head != "tested" {
				t.Fatalf("ready direct merge: %+v %s %v", next, effect, err)
			}
			continue
		}
		if err != nil || effect != integrationMerge {
			t.Fatalf("strict fallback: %+v %s %v", next, effect, err)
		}
		for _, ev := range []integrationEvent{{kind: integrationMerged, head: "integrated"}, {kind: integrationReplayAccepted}, {kind: integrationPushRegistered, pr: landingPR{HeadOID: "integrated"}}} {
			next, effect, err = transitionIntegration(next, ev)
			if err != nil {
				t.Fatal(err)
			}
		}
		if next.attempts != integrationLimit {
			t.Fatalf("issue-only integration charged: %+v", next)
		}
		_, effect, err = transitionIntegration(next, integrationEvent{kind: integrationGatesChecked, main: "code-main"})
		if err != nil || effect != integrationRefuse {
			t.Fatalf("code bypassed exhausted budget: %s %v", effect, err)
		}
	}
}

func TestIntegrationStatePendingChargeSurvivesIssueOnlyPushDrift(t *testing.T) {
	for _, metadataIntegration := range []bool{false, true} {
		s := newIntegrationState("old", landingPR{HeadOID: "old"}, false)
		events := []integrationEvent{{kind: integrationGatesChecked, main: "target", issueOnly: metadataIntegration}, {kind: integrationMerged, head: "merged"}, {kind: integrationReplayAccepted}}
		for _, ev := range events {
			var err error
			s, _, err = transitionIntegration(s, ev)
			if err != nil {
				t.Fatal(err)
			}
		}
		next, effect, err := transitionIntegration(s, integrationEvent{kind: integrationMainObserved, main: "new-record", issueOnly: true})
		if err != nil || effect != integrationPush || next.main != "target" {
			t.Fatalf("metadata reset pending merge or changed its target: %+v %s %v", next, effect, err)
		}
		next, _, err = transitionIntegration(next, integrationEvent{kind: integrationPushRegistered, pr: landingPR{HeadOID: "merged"}})
		want := 1
		if metadataIntegration {
			want = 0
		}
		if err != nil || next.attempts != want {
			t.Fatalf("metadata=%v charge=%d want %d err=%v", metadataIntegration, next.attempts, want, err)
		}
	}
}

func TestIntegrationStateRepeatedMetadataFallbackAndResume(t *testing.T) {
	s := newIntegrationState("tested", landingPR{HeadOID: "tested"}, false)
	s.attempts = 1
	for n := 0; n < 4; n++ {
		for _, event := range []integrationEvent{
			{kind: integrationMainObserved, main: "records", issueOnly: true},
			{kind: integrationGatesChecked, main: "records", issueOnly: true},
			{kind: integrationMainObserved, main: "records", issueOnly: true},
			{kind: integrationMerged, head: "next"},
			{kind: integrationReplayAccepted},
			{kind: integrationPushRegistered, pr: landingPR{HeadOID: "next"}},
		} {
			var err error
			s, _, err = transitionIntegration(s, event)
			if err != nil {
				t.Fatal(err)
			}
		}
		if s.attempts != 1 {
			t.Fatalf("iteration %d charged metadata: %+v", n, s)
		}
	}
	for _, metadata := range []bool{true, false} {
		s = newIntegrationState("local", landingPR{HeadOID: "lease"}, true)
		for _, event := range []integrationEvent{
			{kind: integrationMainObserved, main: "integrated", included: true, issueOnly: metadata},
			{kind: integrationReplayAccepted},
			{kind: integrationPushRegistered, pr: landingPR{HeadOID: "local"}},
		} {
			var err error
			s, _, err = transitionIntegration(s, event)
			if err != nil {
				t.Fatal(err)
			}
		}
		want := 1
		if metadata {
			want = 0
		}
		if s.attempts != want {
			t.Fatalf("resume metadata=%v attempts=%d", metadata, s.attempts)
		}
	}
}
