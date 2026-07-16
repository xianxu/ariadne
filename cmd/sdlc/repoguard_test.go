package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processmanual"
)

// lifecycleVerbs enumerates the guarded verb set from the SAME catalog the
// friction instrument uses (processmanual.WorkflowVerbs) — a new lifecycle verb
// added there automatically demands the guard here (#176 drift protection).
func lifecycleVerbs() []string {
	var out []string
	for v := range processmanual.WorkflowVerbs() {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func writeBrainMarker(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".brain", "config.md"), []byte("# brain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// #176: EVERY lifecycle verb refuses in a brain repo, with the charter message,
// BEFORE any verb-specific validation (the message pin enforces guard-first
// placement — a "--issue is required" die would fail this test).
func TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs(t *testing.T) {
	dir := hermeticRepo(t)
	writeBrainMarker(t, dir)
	for _, verb := range lifecycleVerbs() {
		t.Run(verb, func(t *testing.T) {
			root := buildRoot()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(strings.Fields(verb))
			msg, died := expectDie(t, func() { _ = root.Execute() })
			if !died {
				t.Fatalf("`sdlc %s` in a brain repo must die", verb)
			}
			if !strings.Contains(msg, "capture repo") || !strings.Contains(msg, "lifecycle") {
				t.Errorf("`sdlc %s` refusal must carry the charter, got %q", verb, msg)
			}
		})
	}
}

// A repo without workshop/issues (and no WF_ISSUES_DIR override) is not an
// SDLC repo — lifecycle verbs refuse with the next-action instead of a
// confusing downstream error.
func TestGuardSpineRepo_NonSDLCRepoRefuses(t *testing.T) {
	hermeticRepo(t) // no workshop/issues created
	t.Setenv("WF_ISSUES_DIR", "")
	root := buildRoot()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"claim"})
	msg, died := expectDie(t, func() { _ = root.Execute() })
	if !died || !strings.Contains(msg, "not an SDLC repo") {
		t.Fatalf("claim in a non-SDLC repo: died=%v msg=%q", died, msg)
	}
}

// A normal SDLC repo passes the guard silently — the verb proceeds to its own
// validation (claim without --issue dies with ITS message, not the guard's).
func TestGuardSpineRepo_SDLCRepoPasses(t *testing.T) {
	dir := hermeticRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_ISSUES_DIR", "")
	root := buildRoot()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"claim"})
	msg, died := expectDie(t, func() { _ = root.Execute() })
	if died && (strings.Contains(msg, "capture repo") || strings.Contains(msg, "not an SDLC repo")) {
		t.Fatalf("guard fired in a normal SDLC repo: %q", msg)
	}
}

// WF_SPINE_GUARD=off is the documented emergency hatch: the guard steps aside
// with a cwarn ACK (measurable by the #172 instrument) and the verb proceeds.
func TestGuardSpineRepo_EnvBypassACKs(t *testing.T) {
	dir := hermeticRepo(t)
	writeBrainMarker(t, dir)
	t.Setenv("WF_SPINE_GUARD", "off")
	root := buildRoot()
	var errBuf bytes.Buffer
	root.SetOut(io.Discard)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"claim"})
	msg, died := expectDie(t, func() { _ = root.Execute() })
	if died && strings.Contains(msg, "capture repo") {
		t.Fatalf("guard must step aside under WF_SPINE_GUARD=off, got %q", msg)
	}
	if !strings.Contains(errBuf.String(), "spine repo guard bypassed") {
		t.Errorf("env bypass must ACK loudly, stderr: %q", errBuf.String())
	}
}

// #176 done-issue guard: start-plan/change-code on a terminal issue refuse with
// the new-issue/reopen next-action (only re-close was guarded before).
func TestGuardIssueNotDone(t *testing.T) {
	dir := hermeticRepo(t)
	issuesDir := filepath.Join(dir, "workshop", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doneIssue := "---\nid: 000005\nstatus: done\ndeps: []\ncreated: 2026-07-01\nupdated: 2026-07-01\nestimate_hours: 1\n---\n\n# t\n\n## Problem\n\nx\n\n## Spec\n\nx\n\n## Done when\n\n- x\n\n## Plan\n\n- [x] x\n\n## Log\n\n### 2026-07-01\n"
	if err := os.WriteFile(filepath.Join(issuesDir, "000005-t.md"), []byte(doneIssue), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_ISSUES_DIR", "")
	for _, verb := range []string{"start-plan", "change-code"} {
		t.Run(verb, func(t *testing.T) {
			root := buildRoot()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs([]string{verb, "--issue", "5"})
			msg, died := expectDie(t, func() { _ = root.Execute() })
			if !died || !strings.Contains(msg, "done") || !strings.Contains(msg, "new issue") {
				t.Fatalf("`sdlc %s` on a done issue: died=%v msg=%q", verb, died, msg)
			}
		})
	}
}

// The false-positive arm the guard must never hit: non-done statuses pass —
// a guard firing on a live issue would block all work.
func TestGuardIssueNotDone_WorkingIssuePasses(t *testing.T) {
	dir := hermeticRepo(t)
	issuesDir := filepath.Join(dir, "workshop", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	workingIssue := "---\nid: 000006\nstatus: working\ndeps: []\ncreated: 2026-07-01\nupdated: 2026-07-01\nestimate_hours: 1\n---\n\n# t\n\n## Problem\n\nx\n\n## Spec\n\nx\n\n## Done when\n\n- x\n\n## Plan\n\n- [ ] x\n\n## Log\n\n### 2026-07-01\n"
	if err := os.WriteFile(filepath.Join(issuesDir, "000006-t.md"), []byte(workingIssue), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_ISSUES_DIR", "")
	// Guard layer: a working issue passes for both consumers (change-code's
	// command tree can't be driven past the guard here — its own downstream
	// gates exit via bare exitWithCode, which expectDie can't intercept).
	issuePath := filepath.Join(issuesDir, "000006-t.md")
	msg, died := expectDie(t, func() { guardIssueNotDone(io.Discard, issuePath, "6") })
	if died {
		t.Fatalf("done-guard fired on a WORKING issue: %q", msg)
	}
	// Command tree: start-plan completes normally on a working issue.
	root := buildRoot()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"start-plan", "--issue", "6"})
	msg, died = expectDie(t, func() { _ = root.Execute() })
	if died && strings.Contains(msg, "is status: done") {
		t.Fatalf("done-guard fired through start-plan on a WORKING issue: %q", msg)
	}
}

// #172-instrument guard: none of the new lines (all four) may collide with a
// GateCatalog pattern. (Deliberate flip side: the instrument also does not
// COUNT these — no catalog row; see repoguard.go's follow-up note.)
func TestSpineGuardLinesNoGatesigCollision(t *testing.T) {
	for _, line := range []string{
		brainGuardMsg("brain"),
		notSDLCRepoMsg("workshop/issues"),
		spineGuardBypassACK,
		issueDoneMsg("5"),
	} {
		assertNoGatesigCollision(t, line)
	}
}
