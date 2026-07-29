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
// Pure: no clock, no IO.
func RenderPriorFindings(l Ledger) string {
	if len(l.Rounds) == 0 {
		return "This is the FIRST round of this gate for this issue — there are no prior\n" +
			"findings to dispose of. Review the plan on its merits."
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
	// this reflects the current judgment, not every historical one.
	type disposed struct{ id, state, title string }
	var settled []disposed
	closed := closedSet(l)
	state := map[string]string{}
	for _, r := range l.Rounds {
		for _, d := range r.Dispositions {
			state[d.ID] = d.State
		}
	}
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if closed[f.ID] {
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

	b.WriteString("\nIf a disposed finding is genuinely still wrong, dispose it `not-addressed`\n")
	b.WriteString("by its id — do not raise it again as new.\n")
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
