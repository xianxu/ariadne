package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestIsValidStatus(t *testing.T) {
	for _, s := range vocab.Issue().AllStatuses() {
		if !isValidStatus(s) {
			t.Errorf("isValidStatus(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "WORKING", "done!", "completed", "in-progress"} {
		if isValidStatus(s) {
			t.Errorf("isValidStatus(%q) = true, want false", s)
		}
	}
}

func TestCheckTransitionGuards_RefusesDone(t *testing.T) {
	fm := "id: 000001\nstatus: codecomplete\nestimate_hours: 2\n"
	body := "# Title\n"
	// #160: done is reachable only from codecomplete (the merge edge); set-status
	// still refuses it, routing through the publish flow (close → merge/push).
	err := checkTransitionGuards("codecomplete", "done", fm, body)
	if err == nil {
		t.Fatal("expected refusal")
	}
	if !strings.Contains(err.Error(), "sdlc close") || !strings.Contains(err.Error(), "sdlc merge") {
		t.Errorf("error should redirect through the publish flow: %q", err.Error())
	}
	// working → done is now illegal (routes through codecomplete) — Guard 0 refuses.
	illegal := checkTransitionGuards("working", "done", fm, body)
	if illegal == nil || !strings.Contains(illegal.Error(), "illegal transition") {
		t.Errorf("working → done should be an illegal transition now: %v", illegal)
	}
}

// TestCheckTransitionGuards_RefusesCodecomplete pins #160: only `sdlc close`
// writes codecomplete, so set-status refuses it and points at close.
func TestCheckTransitionGuards_RefusesCodecomplete(t *testing.T) {
	fm := "id: 000001\nstatus: working\nestimate_hours: 2\n"
	body := "# Title\n"
	err := checkTransitionGuards("working", "codecomplete", fm, body)
	if err == nil {
		t.Fatal("expected refusal for → codecomplete")
	}
	if !strings.Contains(err.Error(), "sdlc close") {
		t.Errorf("error should redirect to sdlc close: %q", err.Error())
	}
}

// TestCheckTransitionGuards_WorkingNoLongerNeedsEstimate pins the #113
// decoupling: → working is now estimate-free (claim is a cheap lock; the
// estimate gate moved to `sdlc change-code`). All three shapes — missing,
// empty, and present estimate_hours — must flip cleanly.
func TestCheckTransitionGuards_WorkingNoLongerNeedsEstimate(t *testing.T) {
	body := "# Title\n"
	cases := []struct {
		name string
		fm   string
	}{
		{"missing estimate", "id: 000001\nstatus: open\n"},
		{"empty estimate", "id: 000001\nstatus: open\nestimate_hours:\n"},
		{"present estimate", "id: 000001\nstatus: open\nestimate_hours: 3.5\n"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkTransitionGuards("open", "working", tt.fm, body); err != nil {
				t.Errorf("open → working should be estimate-free now, got: %v", err)
			}
		})
	}
}

func TestCheckTransitionGuards_ReopenNeedsLogEntry(t *testing.T) {
	fm := "id: 000001\nstatus: done\n"
	today := todayIso()

	// No Log section at all → refused.
	if err := checkTransitionGuards("done", "working", fm, "# T\n"); err == nil {
		t.Error("expected refusal: no Log section at all")
	}

	// Log section exists but no today entry → refused.
	bodyOld := "# T\n\n## Log\n\n- 2025-01-01: started\n"
	if err := checkTransitionGuards("done", "working", fm, bodyOld); err == nil {
		t.Error("expected refusal: Log lacks today's entry")
	}

	// Log section with today's entry → OK.
	bodyOK := "# T\n\n## Log\n\n- " + today + ": reopened — found regression\n"
	if err := checkTransitionGuards("done", "working", fm, bodyOK); err != nil {
		t.Errorf("expected ok for today-entry, got: %v", err)
	}

	// Log entry as ### Header instead of bullet — still recognized.
	bodyHeader := "# T\n\n## Log\n\n### " + today + "\nreopened — context\n"
	if err := checkTransitionGuards("done", "working", fm, bodyHeader); err != nil {
		t.Errorf("expected ok for today-header, got: %v", err)
	}
}

func TestCheckTransitionGuards_NormalTransitions(t *testing.T) {
	// Every model-legal transition that isn't →done (routes to close) or a
	// done-reopen (needs a log entry) must pass the guards cleanly — including
	// the #122 M4 additions.
	fm := "id: 000001\nstatus: x\n"
	legal := [][2]string{
		{"open", "working"},    // claim
		{"working", "blocked"}, // block
		{"blocked", "working"}, // unblock
		{"working", "wontfix"}, // abandon mid-flight
		{"working", "punt"},    // defer mid-flight
		// #122 M4 additions:
		{"open", "wontfix"},    // triage-reject unstarted
		{"open", "punt"},       // triage-defer unstarted
		{"punt", "working"},    // resume a deferred
		{"wontfix", "working"}, // reconsider a rejected
		{"blocked", "wontfix"}, // abandon while blocked
		{"blocked", "punt"},    // defer while blocked
		// #160 codecomplete edges that must pass cleanly (→codecomplete and →done
		// are refused by Guards 1b/1 and tested separately; reopen from codecomplete
		// is NOT a done-reopen, so Guard 2 must not demand a log entry):
		{"codecomplete", "working"}, // reopen / rework after drift
		{"codecomplete", "wontfix"}, // abandon late
		{"codecomplete", "punt"},    // defer late
	}
	for _, tr := range legal {
		if err := checkTransitionGuards(tr[0], tr[1], fm, "# T\n"); err != nil {
			t.Errorf("%s → %s should be legal, got: %v", tr[0], tr[1], err)
		}
	}
}

// TestCheckTransitionGuards_IllegalRejected pins the #122 M4 lifecycle gate:
// a transition the model doesn't declare is refused with a message naming the
// illegal edge + the --force escape.
func TestCheckTransitionGuards_IllegalRejected(t *testing.T) {
	fm := "id: 000001\nstatus: x\n"
	illegal := [][2]string{
		{"open", "blocked"}, // claim first
		{"open", "done"},    // can't close an unstarted issue (no actuals)
		{"done", "wontfix"}, // done only reopens to working
		{"punt", "wontfix"}, // terminal→terminal not modeled
	}
	for _, tr := range illegal {
		err := checkTransitionGuards(tr[0], tr[1], fm, "# T\n")
		if err == nil {
			t.Errorf("%s → %s should be rejected by the lifecycle gate", tr[0], tr[1])
			continue
		}
		if !strings.Contains(err.Error(), "illegal transition") || !strings.Contains(err.Error(), "--force") {
			t.Errorf("%s → %s: want an 'illegal transition … --force' message, got: %v", tr[0], tr[1], err)
		}
	}
}

// TestStatusDecision_ForceBypassesLifecycleGate: --force lets an illegal
// transition through (the operator's logged escape hatch).
func TestStatusDecision_ForceBypassesLifecycleGate(t *testing.T) {
	card := []byte(openCard7)
	if _, _, err := statusDecision(card, "", "blocked", false, "2026-09-25", "now", nil); err == nil {
		t.Fatal("open→blocked should be refused without --force")
	}
	out, prev, err := statusDecision(card, "", "blocked", true, "2026-09-25", "now", nil)
	if err != nil || prev != "open" || !strings.Contains(string(out), "status: blocked") {
		t.Fatalf("--force: %s %s %v", out, prev, err)
	}
}

func TestLogHasEntryToday_VariousShapes(t *testing.T) {
	today := "2026-05-25"
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"bullet entry", "## Log\n\n- 2026-05-25: thing\n", true},
		{"date header", "## Log\n\n### 2026-05-25\n", true},
		{"date in middle of paragraph", "## Log\n\nWork done on 2026-05-25.\n", true},
		{"no log section", "## Plan\n- [ ] x\n", false},
		{"log without today", "## Log\n\n- 2025-12-31: old\n", false},
		// today appears only after another ## section → must not count.
		{"date past next section", "## Log\n\n- old: stuff\n\n## Plan\n- 2026-05-25 in plan\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := logHasEntryToday(c.body, today); got != c.want {
				t.Errorf("logHasEntryToday(...) = %v, want %v", got, c.want)
			}
		})
	}
}

// TestStatusDecision_StampsStartedOnce pins #116: open→working stamps an
// engagement anchor; an existing stamp is never overwritten.
func TestStatusDecision_StampsStartedOnce(t *testing.T) {
	out, _, err := statusDecision([]byte(openCard7), "", "working", false, "2026-09-25", "2026-06-18T10:00:00-07:00", nil)
	if err != nil || !strings.Contains(string(out), "started: 2026-06-18T10:00:00-07:00") || !strings.Contains(string(out), "updated: 2026-09-25") {
		t.Fatalf("open→working: %s %v", out, err)
	}
	stamped := strings.Replace(openCard7, "status: open", "status: open\nstarted: 2025-01-01T00:00:00-07:00", 1)
	out, _, err = statusDecision([]byte(stamped), "", "working", false, "2026-09-25", "2099-12-31T23:59:59-07:00", nil)
	if err != nil || strings.Contains(string(out), "2099") || !strings.Contains(string(out), "started: 2025-01-01T00:00:00-07:00") {
		t.Fatalf("existing started moved: %s %v", out, err)
	}
}

// TestRunSetStatus_UpdatesCardAndHonorsDryRun: the status lives on the card;
// a dry run decides without publishing, and the resting branch is untouched.
func TestRunSetStatus_UpdatesCardAndHonorsDryRun(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var stdout, stderr bytes.Buffer
	f := &setStatusFlags{Issue: 9, Status: "working", IssuesDir: "workshop/issues", DryRun: true}
	if err := runSetStatus(context.Background(), &stdout, &stderr, f); err != nil {
		t.Fatal(err)
	}
	if r.card(cardPath) != card || !strings.Contains(stdout.String(), "Would update") {
		t.Fatalf("dry run published or was silent: %s", stdout.String())
	}
	f.DryRun = false
	if err := runSetStatus(context.Background(), &stdout, &stderr, f); err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	if got := r.card(cardPath); !strings.Contains(got, "status: working") || !strings.Contains(got, "updated: "+todayIso()) {
		t.Fatalf("card:\n%s", got)
	}
	if r.git("status", "--porcelain") != "" {
		t.Fatal("set-status edited the resting branch")
	}
	f.Status = "codecomplete"
	if err := runSetStatus(context.Background(), &stdout, &stderr, f); err == nil || !strings.Contains(err.Error(), "sdlc close") {
		t.Fatalf("→ codecomplete must route to close: %v", err)
	}
}

// todayIso returns time.Now() formatted as YYYY-MM-DD — same format
// the production code uses. (Time injection would be cleaner; for M4
// we match the existing close.go posture of calling time.Now()
// directly.)
func todayIso() string {
	return time.Now().Format("2006-01-02")
}

// #277: set-status into working records the workspace like claim does — on an
// open card and on a reopen of unattributed or own work — and refuses when
// another workspace owns the card, even forced.
func TestStatusDecisionRecordsOrRefusesTheClaimant(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	out, _, err := statusDecision([]byte(openCard7), "", "working", false, "2026-10-01", "2026-10-01T09:00:00-07:00", &me)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok, _ := issue.CardClaimant(out); !ok || got != me {
		t.Fatalf("open → working not stamped:\n%s", out)
	}
	blocked := bytes.Replace(out, []byte("status: working"), []byte("status: blocked"), 1)
	if again, _, err := statusDecision(blocked, "", "working", false, "2026-10-01", "2026-10-01T09:00:00-07:00", &me); err != nil || !bytes.Contains(again, []byte("status: working")) {
		t.Fatalf("own blocked → working: %v", err)
	}
	other := me
	other.Worktree = "/w/b"
	for _, force := range []bool{false, true} {
		if _, _, err := statusDecision(blocked, "", "working", force, "2026-10-01", "2026-10-01T09:00:00-07:00", &other); err == nil || !strings.Contains(err.Error(), "`sdlc reclaim`") {
			t.Fatalf("foreign entry (force=%v) = %v", force, err)
		}
	}
	if out, _, err := statusDecision(blocked, "", "open", true, "2026-10-01", "2026-10-01T09:00:00-07:00", &other); err != nil || !bytes.Contains(out, []byte("status: open")) {
		t.Fatalf("leaving working is not an ownership question: %v", err)
	}
}
