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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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
	if code := codeSurfacePaths(w.Conflicted); len(code) > 0 && strings.TrimSpace(fullSuite) == "" {
		return fmt.Sprintf("--fixed-then-ship: this window resolves a conflict with main in code (%s).\n"+
			"  Only the review narrows; the full test suite still runs (operator decision, #320).\n"+
			"  Run it (`make test`), then re-run with --full-suite '<evidence>'.", strings.Join(code, ", "))
	}
	return ""
}

// codeSurfacePaths keeps the paths publishGateHasCodeSurface counts as code — the one
// filter the publish gate's arms and this refusal share, so they agree on what code is.
func codeSurfacePaths(paths []string) []string {
	var code []string
	for _, p := range paths {
		if publishGateHasCodeSurface([]string{p}) {
			code = append(code, p)
		}
	}
	return code
}

// contractSections are the issue sections a fix must not change beyond recording the
// fix (#183 BR-1): the contract and its acceptance oracle.
var contractSections = []string{"Spec", "Revisions", "Done when"}

// maxContractNoteSection bounds one section's text in a review note, so a long Spec
// cannot crowd the prompt; the reviewer reads the rest from the issue file.
const maxContractNoteSection = 4000

// contractChangeNotes compares the contract sections of the issue body the close
// review read (before) with the body now (after) and returns one review note per
// changed section, carrying both texts (#183 BR-1). The reviewer's window excludes the
// issues directory, so without this note it could never see a scope change in the
// issue body, and the fix-review's REWORK-on-scope-change rule could not fire. Section
// text is whitespace-squashed for the comparison, so re-wrapping is no change. Pure.
func contractChangeNotes(before, after string) []string {
	var notes []string
	for _, sec := range contractSections {
		b, _ := issue.SectionBody(before, sec)
		a, _ := issue.SectionBody(after, sec)
		if strings.Join(strings.Fields(b), " ") == strings.Join(strings.Fields(a), " ") {
			continue
		}
		notes = append(notes, fmt.Sprintf("`## %s` changed since the close review (the issue body is outside the window's diff). "+
			"Judge whether this records the fix or changes scope (scope → REWORK).\n  BEFORE:\n%s\n  AFTER:\n%s",
			sec, indentNote(clipNote(b)), indentNote(clipNote(a))))
	}
	return notes
}

func clipNote(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(empty)"
	}
	if len(s) > maxContractNoteSection {
		return s[:maxContractNoteSection] + " … (clipped; read the issue file)"
	}
	return s
}

func indentNote(s string) string {
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
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
const doneWhenLogTag = " (done-when: contract moved since change-code; judged by the close review)"
