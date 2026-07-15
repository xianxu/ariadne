// repoguard.go — the off-workflow guard family (#176, from the #172 friction
// audit): the sdlc LIFECYCLE refuses to run where the workflow state says it
// shouldn't. Two guards:
//
//  1. guardSpineRepo — repo identity. A brain repo (`.brain/config.md`, the
//     constitution's own canonical test) is a capture repo: its own merged
//     AGENTS.md used to invite sdlc, so the binary owns the gate (#69 pattern —
//     the charter must not live only in one agent's memory). A repo without
//     workshop/issues isn't an SDLC repo at all. Wired into exactly the
//     lifecycle verbs (claim, start-plan, change-code, milestone-close, close,
//     merge, push — processmanual.WorkflowVerbs; the drift test enumerates it).
//     Reads (estimate-source, actual, state, process-manual, issue …) stay
//     unguarded by construction — sdlc legitimately READS brain (calibration
//     docs, ledger).
//
//  2. guardIssueNotDone — issue state. `done` is terminal (issue.cue); working
//     a done issue was un-gated until close-time (only the reclose guard fired,
//     at the very end). start-plan/change-code now refuse up front.
//
// Escape hatch: WF_SPINE_GUARD=off (an env, not 7 new per-verb flags) — it
// cwarn-ACKs so the #172 friction instrument can measure bypasses.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

const spineGuardBypassACK = "spine repo guard bypassed (WF_SPINE_GUARD=off) — say why in your commit/log"

func brainGuardMsg(repoName string) string {
	return fmt.Sprintf("%s is a brain (capture repo) — the sdlc lifecycle doesn't run here.\n"+
		"  Captures go through the datatype flows (project/roadmap/pensive) + plain git;\n"+
		"  engineering work belongs in a peer work repo.\n"+
		"  (Emergency override: WF_SPINE_GUARD=off — it is logged.)", repoName)
}

func notSDLCRepoMsg(issuesDir string) string {
	return fmt.Sprintf("not an SDLC repo — no %s/ here.\n"+
		"  Run from the work repo's root, or scaffold the tracker with `sdlc issue new`.", issuesDir)
}

func issueDoneMsg(issueStr string) string {
	return fmt.Sprintf("issue #%s is status: done — terminal (issue.cue).\n"+
		"  Open a new issue referencing #%s for follow-on work;\n"+
		"  a deliberate reopen is `sdlc set-status` (off the modeled lifecycle — say why).", issueStr, issueStr)
}

// guardSpineRepo refuses lifecycle verbs in a brain or non-SDLC repo. Called
// first in each lifecycle verb's RunE — before flag validation, so the refusal
// (not a confusing downstream error) is what the agent sees.
func guardSpineRepo(stderr io.Writer) {
	if os.Getenv("WF_SPINE_GUARD") == "off" {
		cwarn(stderr, spineGuardBypassACK)
		return
	}
	repoTop, err := gitx.RepoTopLevel()
	if err != nil {
		return // not a git repo — the verb's own error explains better
	}
	if _, serr := os.Stat(filepath.Join(repoTop, ".brain", "config.md")); serr == nil {
		die(stderr, brainGuardMsg(filepath.Base(repoTop)))
	}
	// A WF_ISSUES_DIR override is honored as the verbs read it (cwd-relative);
	// the default is anchored at the repo TOP so the guard is correct from any
	// subdirectory (package tests run with cwd inside cmd/sdlc).
	issuesDir := os.Getenv("WF_ISSUES_DIR")
	if issuesDir == "" {
		issuesDir = filepath.Join(repoTop, "workshop", "issues")
	}
	if _, serr := os.Stat(issuesDir); serr != nil {
		die(stderr, notSDLCRepoMsg("workshop/issues"))
	}
}

// guardIssueNotDone refuses start-plan/change-code on a terminal issue. Reads
// the status from the issue file; unreadable/unparsable files are left to the
// verb's own error path (this guard only decides the done question).
func guardIssueNotDone(stderr io.Writer, issuePath, issueStr string) {
	raw, err := os.ReadFile(issuePath)
	if err != nil {
		return
	}
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return
	}
	if status, _ := issue.GetField(fm, "status"); status == "done" {
		die(stderr, issueDoneMsg(issueStr))
	}
}
