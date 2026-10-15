package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

const detailsBody = "## Problem\n\nFiling takes three calls.\n\n## Spec\n\nOne call.\n\n## Done when\n\n- one call files it\n"

func detailsFile(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "details.md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// #371: the card and the whole body publish in the one creation commit, and
// the checkout is not written.
func TestIssueNewDetailsPublishesBodyInOneCall(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "ariadne:0")
	mainBefore, head, status := r.originMain(), r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
	f := newIssueFlags()
	f.Details = detailsFile(t, detailsBody)
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"One Call"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	detail := r.mainFile("workshop/issues/000008-one-call.md")
	for _, want := range []string{"## Spec\n\nOne call.\n", "## Done when\n\n- one call files it\n", "## Log\n\n### "} {
		if !strings.Contains(detail, want) {
			t.Errorf("details on main lack %q:\n%s", want, detail)
		}
	}
	if !strings.Contains(r.card("workshop/issue-cards/000008-one-call.md"), "Filing takes three calls.") {
		t.Error("the card lacks the Problem")
	}
	if n := strings.TrimSpace(r.git("-C", r.origin, "rev-list", "--count", mainBefore+"..main")); n != "1" {
		t.Errorf("main moved by %s commits, want the one creation commit", n)
	}
	assertCheckoutUntouched(t, r, head, status)
}

// #371: a bad details file refuses before anything is reserved.
func TestIssueNewDetailsRefusedBeforeReserving(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "ariadne:0")
	mainBefore, trackerBefore := r.originMain(), r.git("-C", r.origin, "rev-parse", "issue-tracker")
	for name, body := range map[string]string{
		"unknown section": "## Design\nx\n",
		"title":           "# T\n## Spec\nx\n",
	} {
		f := newIssueFlags()
		f.Details = detailsFile(t, body)
		var out, errs bytes.Buffer
		err := runIssueNew(context.Background(), &out, &errs, f, []string{"Bad Body"})
		if err == nil || !strings.Contains(err.Error(), "nothing was reserved") {
			t.Errorf("%s: err %v, want a refusal before reserving", name, err)
		}
	}
	f := newIssueFlags()
	f.Details, f.FromGitHub = detailsFile(t, detailsBody), 42
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, f, nil); err == nil || !strings.Contains(err.Error(), "--from-github") {
		t.Errorf("Problem from both --details and --from-github: err %v", err)
	}
	if r.originMain() != mainBefore || r.git("-C", r.origin, "rev-parse", "issue-tracker") != trackerBefore {
		t.Error("a refused filing moved main or the tracker")
	}
}

// #371: --details - reads stdin; --dry-run shows the merged body.
func TestIssueNewDetailsStdinDryRun(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "ariadne:0")
	prev := issueNewStdin
	issueNewStdin = strings.NewReader(detailsBody)
	t.Cleanup(func() { issueNewStdin = prev })
	f := newIssueFlags()
	f.Details, f.DryRun = "-", true
	mainBefore := r.originMain()
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"From Stdin"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if !strings.Contains(out.String(), "## Spec\n\nOne call.\n") {
		t.Errorf("dry-run lacks the stdin body:\n%s", out.String())
	}
	if r.originMain() != mainBefore {
		t.Error("dry-run published")
	}
}

// #371: --no-claim from a numbered slot files the issue unclaimed.
func TestIssueNewNoClaimFromSlot(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "pair:1")
	f := newIssueFlags()
	f.NoClaim = true
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"For Later"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if _, claimed, err := issue.CardClaimant([]byte(r.card("workshop/issue-cards/000008-for-later.md"))); err != nil || claimed {
		t.Fatalf("claimed=%v err=%v, want unclaimed", claimed, err)
	}
	if !strings.Contains(errs.String(), "--no-claim") {
		t.Errorf("no unclaimed notice:\n%s", errs.String())
	}
}
