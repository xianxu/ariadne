package gatestate

import (
	"fmt"
	"strings"
)

// RenderPriorFindings renders the block the next round's prompt embeds — the mechanism
// that turns a memoryless reviewer into a converging one (ariadne#187 A2).
//
// Two lists, and both matter. OPEN findings must be disposed before anything new is
// raised, so the judge can see its earlier work was addressed instead of re-deriving an
// absolute bar. ALREADY-DISPOSED findings are listed precisely so they are NOT re-raised
// at a lower severity, which is the descent pattern the postmortem observed.
//
// Since ariadne#194 M3 it also carries the FAMILY vocabulary and, when a family already
// has prior findings, the escalation that changes the reviewer's recommendation from
// "fix this instance" to "state the rule". Rendering the vocabulary is not decoration:
// FamilyCounts keys on the slug, so a reviewer that cannot see which slugs are in play
// coins a fresh one every round, every count stays 1, and the escalation never fires.
//
// Pure: no clock, no IO.
func RenderPriorFindings(l Ledger) string { return RenderPriorFindingsScoped(l, l) }

// RenderPriorFindingsScoped renders the block from TWO views of the same ledger:
// `scoped` supplies the findings this boundary must dispose of, `full` supplies the
// family counts.
//
// The split is the whole point of one ledger per issue (ariadne#194 M3 review, BR-20).
// FamilyCounts on the scoped view means a family can never be seen recurring ACROSS
// milestones — and cross-milestone recurrence is exactly the tools#1 evidence this
// feature was built from, where one missing rule spanned M1's rounds and the close
// review. Worse, at the whole-issue close the scope is "" so every milestone round drops
// out and the vocabulary renders empty, which is what the M3 prompt actually did.
//
// Plan-quality has one boundary by construction and passes the same ledger twice.
func RenderPriorFindingsScoped(scoped, full Ledger) string {
	l := scoped
	if len(l.Rounds) == 0 {
		first := "This is the FIRST round of this gate at this boundary — there are no prior\n" +
			"findings to dispose of. Review the work on its merits."
		// ...but a family raised at an EARLIER boundary still applies. Returning here
		// unconditionally is what made the M3 prompt read "FIRST round" while three
		// milestones of families sat in the ledger (BR-20).
		if counts := FamilyCounts(full); len(counts) > 0 {
			return first + "\n\n" + renderFamilyVocabulary(counts) + renderFamilyEscalation(counts)
		}
		return first
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d prior round(s) of this gate have already run. Their findings follow.\n", len(l.Rounds))

	open := OpenFindings(l)
	b.WriteString("\nOPEN FINDINGS — you MUST dispose of every one of these before raising anything new:\n\n")
	if len(open) == 0 {
		// "All disposed" and "none were ever recorded" are different states, and only one
		// of them is a clean slate. A round that hit a protocol error can persist with no
		// findings, so an unconditional "every prior finding has been disposed" would make
		// a positive claim the ledger cannot support — in the artifact whose whole purpose
		// is not losing findings.
		if everRaised(l) {
			b.WriteString("  (none — every prior finding has been disposed)\n")
		} else {
			b.WriteString("  (none — no findings have been recorded for this gate yet)\n")
		}
	}
	for _, f := range open {
		fmt.Fprintf(&b, "  - id: %s  [%s] %s\n", f.ID, f.Severity, f.Title)
		if f.Detail != "" {
			fmt.Fprintf(&b, "      %s\n", strings.ReplaceAll(f.Detail, "\n", "\n      "))
		}
	}

	// Disposed findings, with the state that settled them — last disposition wins, so
	// this reflects the current judgment, not every historical one. A live dispute is NOT
	// listed here: it is not settled by the judge, and the section below asks for an answer.
	type disposed struct{ id, state, title string }
	var settled []disposed
	closed := closedSet(l)
	state := latestStates(l)
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if closed[f.ID] && state[f.ID] != Disputed {
				settled = append(settled, disposed{f.ID, state[f.ID], f.Title})
			}
		}
	}
	if len(settled) > 0 {
		b.WriteString("\nALREADY DISPOSED — do NOT re-raise these, at any severity:\n\n")
		for _, s := range settled {
			fmt.Fprintf(&b, "  - id: %s  (%s) %s\n", s.id, s.state, s.title)
		}
	}

	b.WriteString(renderLiveDisputes(LiveDisputes(l)))

	b.WriteString("\nIf a disposed finding is genuinely still wrong, dispose it `not-addressed`\n")
	b.WriteString("by its id — do not raise it again as new.\n")
	if counts := FamilyCounts(full); len(counts) > 0 {
		b.WriteString("\n")
		b.WriteString(renderFamilyVocabulary(counts))
		b.WriteString(renderFamilyEscalation(counts))
	}

	return strings.TrimRight(b.String(), "\n")
}

// everRaised reports whether ANY round has recorded a finding. It distinguishes "all
// disposed" from "never recorded" — states an open-set count of zero cannot tell apart,
// and which a prompt must not conflate (see RenderPriorFindings).
func everRaised(l Ledger) bool {
	for _, r := range l.Rounds {
		if len(r.New) > 0 {
			return true
		}
	}
	return false
}

// renderLiveDisputes renders the findings the implementer declined, with the rationale and
// the counter-measurement, and asks the judge to answer each against that evidence
// (ariadne#202). Before this the rebuttal had no channel into the next round, so the judge
// re-derived the same mistake or the finding timed out by the round cap. Pure.
func renderLiveDisputes(live []LiveDispute) string {
	if len(live) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nDISPUTED BY THE IMPLEMENTER — answer each against its evidence. Re-run the\n")
	b.WriteString("measurement if you can. Dispose it `withdrawn` if the evidence holds, or\n")
	b.WriteString("`not-addressed` with a note that answers the evidence (not the rationale alone).\n")
	b.WriteString("A dispute you leave unanswered stands.\n\n")
	for _, d := range live {
		fmt.Fprintf(&b, "  - id: %s  [%s] %s\n", d.Finding.ID, d.Finding.Severity, d.Finding.Title)
		fmt.Fprintf(&b, "      rationale: %s\n", strings.ReplaceAll(d.Dispute.Rationale, "\n", "\n        "))
		fmt.Fprintf(&b, "      evidence:  %s\n", strings.ReplaceAll(d.Dispute.Evidence, "\n", "\n        "))
	}
	return b.String()
}
