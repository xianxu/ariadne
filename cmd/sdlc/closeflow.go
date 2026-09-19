package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// closeFlowOutcome is what a close does with the issue's flow (#231).
type closeFlowOutcome struct {
	flow      flow.Flow // the flow this close runs under — and records, when upgraded
	crossings []string  // non-empty: a quick issue left the shell and is upgraded to full
	size      flow.Size // the measured window (quick flow only)
}

func (o closeFlowOutcome) upgraded() bool { return len(o.crossings) > 0 }

// category is the boundary-review recipe this close dispatches: the small-diff
// review on the quick flow, the full milestone-review otherwise — an upgraded
// issue is full by then, so it needs no special case.
func (o closeFlowOutcome) category() judge.Category {
	if o.flow.Kind() == flow.Quick {
		return judge.SmallDiffReview
	}
	return judge.MilestoneReview
}

// decideCloseFlow stays quick or upgrades, from the recorded quick flow and the
// window's measured size. Pure; crossing the shell never refuses.
func decideCloseFlow(rec flow.Flow, size flow.Size) closeFlowOutcome {
	if c := size.Crossings(); len(c) > 0 {
		return closeFlowOutcome{flow: flow.Upgrade(rec), crossings: c, size: size}
	}
	return closeFlowOutcome{flow: rec, size: size}
}

// closeFlowLine is the one line a quick-flow close prints about its flow.
func closeFlowLine(o closeFlowOutcome) string {
	if o.upgraded() {
		return "flow: quick → full — outside the quick-flow shell (" + strings.Join(o.crossings, "; ") +
			"); this close runs the full boundary review"
	}
	return fmt.Sprintf("flow: quick — inside the shell (%d added lines in code files, a design of %d lines); "+
		"this close runs the small-diff review", o.size.AddedLines, o.size.DesignLines)
}

// closeFlowStep is close's flow step, run inside computeClose on the window the
// atlas gate and the boundary review already use (#58). An issue with no record,
// or a full one, closes exactly as before. On the quick flow it runs the two
// Done-when checks (issue close only — they guard the final acceptance review)
// and measures the window against the shell. It WRITES NOTHING: an upgrade is
// composed into the issue text and recorded by applyClose at finalize, so a
// REWORK leaves the issue as it was (#139); what keeps the upgrade across rounds
// is the boundary ledger, whose rounds record the recipe they ran — an earlier
// full-review round is a crossing (earlierFullReview, #231 BR-18).
func closeFlowStep(stderr io.Writer, f *closeFlags, mode, issuePath, fm, body, windowBase, windowHead string) closeFlowOutcome {
	rec, err := flow.Recorded(fm)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("flow record unreadable (%v) — closing as the full flow", err))
	}
	if rec == nil || rec.Kind() != flow.Quick {
		var fl flow.Flow
		if rec != nil {
			fl = *rec
		}
		return closeFlowOutcome{flow: fl}
	}
	if mode == "issue" {
		checkQuickDoneWhen(stderr, f, *rec, body)
	}
	plan := readOptionalPlanFile(f.plansDir(), strings.TrimSuffix(filepath.Base(issuePath), filepath.Ext(issuePath)))
	size := measureCloseWindow(stderr, windowBase, windowHead, body, plan)
	size.EarlierFullReview, size.LedgerErr = earlierFullReview(f.plansDir(), filepath.Base(issuePath), f.Issue, f.Milestone)
	o := decideCloseFlow(*rec, size)
	cinfo(stderr, closeFlowLine(o))
	return o
}

// checkQuickDoneWhen runs the quick flow's two deterministic checks before its
// one review: Done-when must have a bullet (it is the review's only oracle), and
// must have moved if the contract did since change-code.
func checkQuickDoneWhen(stderr io.Writer, f *closeFlags, rec flow.Flow, body string) {
	if err := flow.DoneWhenPresent(body); err != nil {
		// No --no-<gate> flag: only --force waives the review's missing oracle.
		if !f.Force {
			die(stderr, err.Error())
		}
		cwarn(stderr, "--force: closing a quick-flow issue with no `## Done when` bullet")
	}
	if err := flow.DoneWhenFresh(rec, body); err != nil {
		if !f.skip("done-when-fresh") {
			die(stderr, err.Error())
		}
		cwarn(stderr, "--no-done-when-fresh (or --force): skipping the Done-when freshness check — rationale in --verified")
	}
}

// measureCloseWindow gathers the shell's facts: the numstat of the window, the
// design's length as it stands now (the issue body and the durable plan — a
// design that grew past its limit during the work crosses here), and the Plan's
// Mx rows (through the fence-filtered body, like every plan-item reader).
func measureCloseWindow(stderr io.Writer, windowBase, windowHead, body, plan string) flow.Size {
	var stats []churn.FileStat
	if windowBase != "" {
		st, err := windowFileStats(windowBase, windowHead)
		if err != nil {
			// Same stance as the atlas gate's name diff: a broken window diff is a
			// hard stop, never a silently empty measurement.
			die(stderr, fmt.Sprintf("window numstat %s..%s failed: %v", shortSHA(windowBase), abbrevSHA(windowHead), err))
		}
		stats = st
	}
	var milestones []string
	if plan, ok := issue.PlanItemsBody(body); ok {
		milestones = issue.MilestonesInPlanOrder(plan)
	}
	return flow.Measure(stats, flow.DesignLines(body, plan), milestones)
}

// windowFileStats is the numstat of a window, one row per changed file. Shared
// with the churn report (churnForWindow), which reads the same final diff.
func windowFileStats(base, head string) ([]churn.FileStat, error) {
	span := base + ".." + head
	// -z: without it git quotes a path holding non-ASCII, a double quote or a
	// backslash, and the quoted path no longer reads as the doc or test it is —
	// its lines count as code, upgrading a small change for lines that do not
	// count (pinned by TestCloseQuotedDocAndTestNamesDoNotCount). Every git
	// listing a gate reads as data uses -z (#231 M2 review).
	out, err := gitx.RunGit("diff", "--numstat", "-z", span)
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat -z %s: %w", span, err)
	}
	return churn.ParseNumstatZ(string(out)), nil
}

// earlierFullReview reports whether an earlier round of this boundary already
// ran the full review — read from the boundary ledger, which every round
// persists, REWORK included. Pure over the ledger (fullRoundIn); this is its IO.
func earlierFullReview(plansDir, issueFileName string, issueNum int, boundary string) (bool, error) {
	l, err := readBoundaryGateLedger(plansDir, issueFileName, issueNum)
	if err != nil {
		return false, err
	}
	return fullRoundIn(gatestate.FilterBoundary(l, boundary).Rounds), nil
}

// fullRoundIn: did any of these boundary rounds run the full review? A round
// that never ran (NoCap) ran nothing; an unstamped round predates #231, when
// milestone-review was the only boundary recipe, so it was the full review.
func fullRoundIn(rounds []gatestate.Round) bool {
	for _, r := range rounds {
		if !r.NoCap && r.Recipe != string(judge.SmallDiffReview) {
			return true
		}
	}
	return false
}
