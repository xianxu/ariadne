package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
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
	r := newTrackerRepo(t, map[string]string{goodPath: goodCard, badPath: badCard},
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

	// Every reader of the composed records names the cause instead of reading
	// the card as absent (M1 review BR-3/BR-4).
	names := func(what string, err error) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "fingerprint") {
			t.Errorf("%s: want an error naming the cause, got %v", what, err)
		}
	}
	invalidateIssueRecords(ctx)
	errs.Reset()
	names("issue set-status", runSetStatus(ctx, &out, &errs, &setStatusFlags{Issue: 42, Status: "blocked", IssuesDir: "workshop/issues"}))
	_, err = historyFileIsTerminal(ctx, badDetailPath)
	names("archive terminal check", err)
	_, err = lookupIssueMeta(ctx, "#42", ".")
	names("project status lookup", err)
	_, err = overlayCardStatus(ctx, "workshop/issues", []issueFileRef{{Path: badDetailPath}})
	names("issue-file status overlay", err)
	_, err = fleet.LookupRepoIssues(ctx, ".", "000042")
	names("fleet branch lookup", err)
	env, err := openTrackerAt(ctx, ".")
	if err != nil {
		t.Fatal(err)
	}
	rs, err := loadIssueRecords(ctx, "workshop/issues", tracker.PreferFresh)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ownedCompletions(env, rs, "HEAD", "", false)
	names("landing completions", err)
	if _, _, _, warning := actualTrackerInputs(ctx, ".", "42"); !strings.Contains(warning, "card unreadable") {
		t.Errorf("actual: want a warning naming the unreadable card, got %q", warning)
	}
	if msg, died := expectDie(t, func() { guardIssueNotDone(ctx, io.Discard, badDetailPath, "42") }); !died || !strings.Contains(msg, "cannot confirm") {
		t.Errorf("the not-done guard must fail closed on an unreadable card: died=%v %q", died, msg)
	}
	msg, died := expectDie(t, func() {
		computeClose(io.Discard, &closeFlags{Context: ctx, Issue: 42, Milestone: "M1", Actual: "1", Verified: "x", IssuesDir: "workshop/issues", PlansDir: "workshop/plans"})
	})
	if !died || !strings.Contains(msg, "fingerprint") {
		t.Errorf("milestone close of the malformed card: died=%v %q", died, msg)
	}
	if inv, _ := fleetClaims(t, r.root); rowAt(t, inv, r.root).ClaimsState != fleet.ClaimsPartial ||
		!strings.Contains(rowAt(t, inv, r.root).ClaimsError, "#000042") || len(rowAt(t, inv, r.root).Claims) != 1 {
		t.Errorf("fleet inventory must report the claims read partial, naming #42: %+v", rowAt(t, inv, r.root))
	}
	out.Reset()
	if err := runIssueShow(ctx, &out, &errs, &issueShowFlags{IssuesDir: "workshop/issues"}, "42"); err != nil || !strings.Contains(out.String(), "card unreadable") {
		t.Errorf("issue show text: %v\n%s", err, out.String())
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
