package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"strings"
	"testing"
	"time"
)

// #284: the claims views group owned issues by slot and by operator, with ages
// relative to now; unowned issues are left out.
func TestClaimViewsAndProse(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2026-10-07T15:00:00-07:00")
	issues := []IssueState{
		{ID: "000284", Status: "working", Title: "claims", Owner: &IssueOwner{Operator: "Xian", Workspace: "ariadne:2", Worktree: "/w/2"}, ClaimedAt: "2026-10-07T12:00:00-07:00"},
		{ID: "000285", Status: "open", Title: "guard", Owner: &IssueOwner{Operator: "Xian", Worktree: "/w/plain"}, ClaimedAt: "2026-10-04T15:00:00-07:00"},
		{ID: "000286", Status: "open", Title: "pushes", Owner: &IssueOwner{Operator: "Pat", Workspace: "ariadne:2", Worktree: "/w/2"}},
		{ID: "000287", Status: "open", Title: "unowned"},
	}
	bySlot, byOperator := claimViews(issues, now)
	if len(bySlot) != 2 || bySlot[0].key != "/w/plain" || bySlot[1].key != "ariadne:2" || strings.Join(bySlot[1].issues, ",") != "#284 (3h),#286" {
		t.Fatalf("by slot: %+v", bySlot)
	}
	if len(byOperator) != 2 || byOperator[1].key != "Xian" || strings.Join(byOperator[1].issues, ",") != "#284 (3h),#285 (3d)" {
		t.Fatalf("by operator: %+v", byOperator)
	}
	var out bytes.Buffer
	if err := renderProseAt(&out, State{Issues: issues}, now); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[ariadne:2, claimed 3h ago]", "Claims by slot:", "  ariadne:2  #284 (3h), #286", "Claims by operator:", "  Pat  #286"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("prose lacks %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "#287 (") {
		t.Fatal("an unowned issue appears in the claims views")
	}
}

// #284: state reads each owned issue's owner and claim time — the claim
// commit's, not a later card write's — and reports nothing for unowned ones.
func TestStateReportsOwnerAndClaimTime(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	claimCommit := r.git("log", "-1", "--format=%cI", "refs/remotes/origin/issue-tracker")
	time.Sleep(1100 * time.Millisecond) // the later write must not share the claim's second
	var out bytes.Buffer
	if err := startPlanBranch(context.Background(), &out, 9); err != nil {
		t.Fatal(err)
	}
	issues, _, err := listIssueStates(context.Background(), "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	if err := fillClaimTimes(context.Background(), r.root, issues); err != nil {
		t.Fatal(err)
	}
	seen := map[string]IssueState{}
	for _, i := range issues {
		seen[i.ID] = i
	}
	nine := seen["000009"]
	if nine.Owner == nil || nine.Owner.Worktree != canonRoot(r.root) {
		t.Fatalf("#9's owner: %+v", nine.Owner)
	}
	want, _ := time.Parse(time.RFC3339, claimCommit)
	if got, err := time.Parse(time.RFC3339, nine.ClaimedAt); err != nil || !got.Equal(want) {
		t.Fatalf("#9 claimed at %q, want the claim commit's %s", nine.ClaimedAt, claimCommit)
	}
	if ten := seen["000010"]; ten.Owner != nil || ten.ClaimedAt != "" {
		t.Fatalf("#10 is unowned: %+v", ten)
	}
	raw, err := json.Marshal(nine)
	if err != nil || !strings.Contains(string(raw), `"owner":{`) || !strings.Contains(string(raw), `"claimed_at":"`) {
		t.Fatalf("JSON: %s %v", raw, err)
	}
}

// #284 BR-32: a failed claim-age read is reported in state's drift, not lost
// to the drift computation that follows it.
func TestStateReportsAFailedClaimAgeRead(t *testing.T) {
	claimSetRepo(t)
	claimFor(t, 9)
	prev := claimTimesOf
	t.Cleanup(func() { claimTimesOf = prev })
	claimTimesOf = func(*tracker.Repository, map[string]bool) (map[string]time.Time, error) {
		return nil, errors.New("injected history failure")
	}
	var out bytes.Buffer
	if err := runState(context.Background(), &out, &stateFlags{JSON: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "claim ages unavailable: injected history failure") {
		t.Fatalf("the failure is not in state's drift:\n%s", out.String())
	}
}
