package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
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
	return fmt.Sprintf("flow: quick — inside the shell (%d code files, %d added lines); this close runs the small-diff review",
		len(o.size.CodeFiles), o.size.AddedLines)
}

// closeFlowStep is close's flow step, run inside computeClose on the window the
// atlas gate and the boundary review already use (#58). An issue with no record,
// or a full one, closes exactly as before. On the quick flow it runs the two
// Done-when checks (issue close only — they guard the final acceptance review)
// and measures the window against the shell. It WRITES NOTHING: an upgrade is
// composed into the issue text and recorded by applyClose at finalize, so a
// REWORK leaves the issue as it was (#139) and the re-close re-derives it.
func closeFlowStep(stderr io.Writer, f *closeFlags, mode, fm, body, windowBase, windowHead string, diffFiles []string) closeFlowOutcome {
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
	o := decideCloseFlow(*rec, measureCloseWindow(stderr, windowBase, windowHead, diffFiles, body))
	cinfo(stderr, closeFlowLine(o))
	return o
}

// checkQuickDoneWhen runs the quick flow's two deterministic checks before its
// one review: Done-when must have a bullet (it is the review's only oracle), and
// must have moved if the contract did since change-code.
func checkQuickDoneWhen(stderr io.Writer, f *closeFlags, rec flow.Flow, body string) {
	if err := flow.DoneWhenPresent(body); err != nil {
		if !f.skip("done-when") {
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
// shared-surface declaration as committed at its base and head, and the Plan's
// Mx rows (through the fence-filtered body, like every plan-item reader).
func measureCloseWindow(stderr io.Writer, windowBase, windowHead string, diffFiles []string, body string) flow.Size {
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
	surfaces, serr := committedSurfaces(windowBase, windowHead)
	var milestones []string
	if plan, ok := issue.PlanItemsBody(body); ok {
		milestones = issue.MilestonesInPlanOrder(plan)
	}
	return flow.Measure(diffFiles, stats, surfaces, serr, milestones)
}

// windowFileStats is the numstat of a window, one row per changed file. Shared
// with the churn report (churnForWindow), which reads the same final diff.
func windowFileStats(base, head string) ([]churn.FileStat, error) {
	span := base + ".." + head
	out, err := gitx.RunGit("diff", "--numstat", span)
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat %s: %w", span, err)
	}
	return churn.ParseNumstat(string(out)), nil
}

// committedSurfaces is the union of the shared-surface declaration as committed
// at the window's base and at its head. Never the working tree: an uncommitted
// edit cannot loosen the shell, and a branch that deletes a pattern still sees
// it through the base (#231 PQ-5). An absent declaration is simply no surfaces;
// an unreadable one is an error the shell turns into a crossing.
func committedSurfaces(base, head string) (flow.Surfaces, error) {
	var all flow.Surfaces
	for _, rev := range []string{base, head} {
		if rev == "" {
			continue
		}
		text, ok, err := declarationAt(rev)
		if err != nil {
			return all, err
		}
		if !ok {
			continue
		}
		s, err := flow.ParseSurfaces(text)
		if err != nil {
			return all, err
		}
		all = flow.Union(all, s)
	}
	return all, nil
}

// declarationAt reads the declaration at a revision. ls-tree separates "absent"
// from "unreadable" — the distinction a failed `git show` alone cannot make.
func declarationAt(rev string) (string, bool, error) {
	out, err := gitx.RunGit("ls-tree", "--full-tree", "--name-only", rev, "--", flow.DeclarationPath)
	if err != nil {
		return "", false, fmt.Errorf("ls-tree %s: %w", rev, err)
	}
	if strings.TrimSpace(string(out)) == "" {
		return "", false, nil
	}
	b, err := gitx.RunGit("show", rev+":"+flow.DeclarationPath)
	if err != nil {
		return "", false, fmt.Errorf("show %s:%s: %w", rev, flow.DeclarationPath, err)
	}
	return string(b), true, nil
}
