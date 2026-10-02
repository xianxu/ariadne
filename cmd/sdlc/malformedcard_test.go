package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
)

// #288: one malformed tracker card is quarantined, not fatal. Every other
// card's verbs proceed; the bad card's own verbs refuse naming the cause;
// read views report it unreadable/unknown (never absent); publishing refuses
// repo-wide (an unreadable card may hide a handoff); and its ID is never
// reallocated.
func TestOneMalformedCardDoesNotBlockOthers(t *testing.T) {
	goodPath, goodCard, goodDetailPath, goodDetail := seededIssue(t, "000041", "good")
	badPath, badCard, badDetailPath, badDetail := seededIssue(t, "000042", "bad")
	// A claimant carrying a raw machine ID, never a fingerprint: ParseCard
	// rejects it, so before #288 every tracker read failed.
	badCard = strings.Replace(badCard, "status: open", "status: working\nclaimant:\n    operator: a\n    machine: RAW-MACHINE-ID\n    machine_name: m\n    worktree: /w\n    repository: r", 1)
	newTrackerRepo(t, map[string]string{goodPath: goodCard, badPath: badCard},
		map[string]string{goodDetailPath: goodDetail, badDetailPath: badDetail})
	ctx := context.Background()

	var out, errs bytes.Buffer
	if err := runClaim(ctx, &out, &errs, claimFlagsFor(41)); err != nil {
		t.Fatalf("claiming a readable card beside a malformed one: %v\n%s", err, errs.String())
	}
	invalidateIssueRecords(ctx)
	errs.Reset()
	if err := runClaim(ctx, &out, &errs, claimFlagsFor(42)); err == nil || !strings.Contains(err.Error()+errs.String(), "unreadable card") || !strings.Contains(err.Error()+errs.String(), "fingerprint") {
		t.Fatalf("claiming the malformed card must refuse naming the cause: %v\n%s", err, errs.String())
	}

	if o := observeIssue(t, 42, ""); o.Card.State != observe.Unknown || !strings.Contains(o.Card.Error, "unreadable card") {
		t.Fatalf("issue show --json of the malformed card: %+v", o.Card)
	}
	if o := observeIssue(t, 41, ""); o.Card.Status != "working" || o.Assignment.Relation != observe.RelationThisWorkspace {
		t.Fatalf("issue show --json of the good card: %+v %+v", o.Card, o.Assignment)
	}

	states, _, err := listIssueStates(ctx, "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]IssueState{}
	for _, s := range states {
		seen[s.ID] = s
	}
	if seen["000042"].Status != "unreadable" || !strings.Contains(seen["000042"].Unreadable, "fingerprint") || seen["000041"].Status != "working" {
		t.Fatalf("issue list/state: %+v", states)
	}
	if d := detectDrift(states, "workshop/history", func(string) (string, string, bool) { return "", "", false }); !driftMentions(d, "000042", "card unreadable") {
		t.Fatalf("state drift does not surface the malformed card: %+v", d)
	}

	if err := guardTransferredDetails(ctx); err == nil || !strings.Contains(err.Error(), "#000042 is malformed") {
		t.Fatalf("publishing must refuse while a card is unreadable: %v", err)
	}

	out.Reset()
	errs.Reset()
	if err := runIssueNew(ctx, &out, &errs, newIssueFlags(), []string{"After The Bad One"}); err != nil {
		t.Fatalf("issue new: %v\n%s", err, errs.String())
	}
	if got := strings.TrimSpace(out.String()); !strings.Contains(got, "000043-") {
		t.Fatalf("issue new reused or skipped the quarantined ID: %q", got)
	}
}

func driftMentions(findings []DriftFinding, id, text string) bool {
	for _, f := range findings {
		if f.Issue == id && strings.Contains(f.Message, text) {
			return true
		}
	}
	return false
}
