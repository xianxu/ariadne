package main

import (
	"strings"
	"testing"
)

func TestFilterIssueQueueAttemptsOnlyProvenForeign(t *testing.T) {
	own := SuiteAttempt{ID: "aabb", SHA: queueGraphOID(0)}
	foreign := SuiteAttempt{ID: "ccdd", SHA: queueGraphOID(1)}
	unknown := SuiteAttempt{ID: "eeff", SHA: queueGraphOID(2)}
	members := map[string]queueHeadMembership{own.SHA: {Known: true, Issues: []int{548, 3}}, foreign.SHA: {Known: true, Issues: []int{3}}}
	for _, attempts := range [][]SuiteAttempt{{own, foreign}, {foreign, own}} {
		got, err := filterIssueQueueAttempts(attempts, members, 548)
		if err != nil || len(got) != 1 || got[0].ID != own.ID {
			t.Fatalf("got=%+v err=%v", got, err)
		}
	}
	if _, err := filterIssueQueueAttempts([]SuiteAttempt{foreign, unknown}, members, 548); err == nil || !strings.Contains(err.Error(), unknown.ID) || !strings.Contains(err.Error(), unknown.SHA) {
		t.Fatalf("unknown missing identity: %v", err)
	}
	foreign.ID = own.ID
	if _, err := filterIssueQueueAttempts([]SuiteAttempt{own, foreign}, members, 548); err == nil {
		t.Fatal("foreign filter hid duplicate identity")
	}
}
