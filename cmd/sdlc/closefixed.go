// closefixed.go — `sdlc close --fixed-then-ship` (#183): the narrow re-close. After a
// finalized close review, the fixes made after its verdict (and any resolution of a
// conflict with main) are reviewed as the interdiff since the head that review read,
// by the fix-review recipe, instead of a full re-review of the branch patch. Plus the
// quick flow's Done-when freshness decision (F3), which defers to whichever review
// runs. Everything here is pure; computeClose and the dispatch apply it.
package main

import (
	"fmt"
	"strings"
)

// fixedThenShipRefusal decides whether a --fixed-then-ship close may proceed, over
// the planned fix window, whether the quick flow's shell (measured on the branch
// patch) upgraded the issue, and the --full-suite evidence. "" proceeds; anything
// else is the refusal, naming its next action. Pure.
func fixedThenShipRefusal(issueStr string, w reviewWindow, upgraded bool, fullSuite string) string {
	full := fmt.Sprintf("sdlc close --issue %s --verified '<evidence>'", issueStr)
	switch {
	case w.Kind == windowNone:
		return fmt.Sprintf("--fixed-then-ship has no fix window: %s.\n"+
			"  It reviews the changes since a finalized close review; without one, run the full close:\n  %s",
			orPlaceholder(w.Note, "no commit anchors it"), full)
	case upgraded:
		return fmt.Sprintf("--fixed-then-ship: the branch patch has left the quick-flow shell, so the full review is due.\n  %s", full)
	}
	if code := conflictedCode(w.Conflicted); len(code) > 0 && strings.TrimSpace(fullSuite) == "" {
		return fmt.Sprintf("--fixed-then-ship: this window resolves a conflict with main in code (%s).\n"+
			"  Only the review narrows; the full test suite still runs (operator decision, #320).\n"+
			"  Run it (`make test`), then re-run with --full-suite '<evidence>'.", strings.Join(code, ", "))
	}
	return ""
}

// conflictedCode is the conflicted paths with code surface — the publish gate's own
// predicate, so the gate and this refusal agree on what counts as code.
func conflictedCode(conflicted []string) []string {
	var code []string
	for _, p := range conflicted {
		if publishGateHasCodeSurface([]string{p}) {
			code = append(code, p)
		}
	}
	return code
}

// fixedThenShipNotes are the review notes a fix window carries: the conflict
// resolutions the reviewer must judge. Pure.
func fixedThenShipNotes(w reviewWindow) []string {
	if len(w.Conflicted) == 0 {
		return nil
	}
	return []string{"conflict resolutions (main rewrote lines the close review read; the resolution was never reviewed): " +
		strings.Join(w.Conflicted, ", ")}
}

// fixedThenShipLogTag is the close log line's marker for a narrow re-close, with the
// full-suite evidence when given. Pure.
func fixedThenShipLogTag(fullSuite string) string {
	if s := strings.TrimSpace(fullSuite); s != "" {
		return " (fixed-then-ship; full suite: " + s + ")"
	}
	return " (fixed-then-ship)"
}

// doneWhenAction is what close does with the quick flow's Done-when freshness check.
type doneWhenAction int

const (
	doneWhenFresh  doneWhenAction = iota // the contract and Done-when agree with change-code's record
	doneWhenDefer                        // stale, and a review runs: the reviewer judges it (F3)
	doneWhenRefuse                       // stale, and no review runs: the deterministic refusal stands
	doneWhenWaived                       // stale, no review, --no-done-when-fresh (or --force)
)

// doneWhenDecision (F3, #183): a moved contract with an unmoved Done-when is either a
// reframe or a fix record (a review-fix Revisions entry), and only a reader of the
// diff can tell which. So when a review runs, it judges; with no review, the
// deterministic refusal stands, and a review that did not run never finalizes, so it
// cannot pass the check either (requirement 7). Pure.
func doneWhenDecision(stale, reviewRuns, waived bool) doneWhenAction {
	switch {
	case !stale:
		return doneWhenFresh
	case reviewRuns:
		return doneWhenDefer
	case waived:
		return doneWhenWaived
	}
	return doneWhenRefuse
}

// doneWhenReviewNote is the note a deferred freshness check hands the reviewer.
const doneWhenReviewNote = "the contract (`## Spec` / `## Revisions`) changed since change-code while `## Done when` did not. " +
	"Judge whether `## Done when` still covers what this diff builds; if it does not, raise an Important finding. " +
	"State the judgement on its own line: `Done-when: current — <why>` or `Done-when: stale — <why>`."

// doneWhenLogTag is the close log line's record of a deferred check (the verdict
// annotation that follows it names the judging review).
const doneWhenLogTag = " (done-when: contract moved since change-code; judged by the review)"
