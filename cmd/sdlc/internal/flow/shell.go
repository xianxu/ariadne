package flow

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
)

// Size is what the close-time shell check measures of a review window (#231).
type Size struct {
	AddedLines int      // insertions in code files (churn.IsCodeFile) — git's `+` count, as churn reports it
	Milestones []string // Mx rows in the Plan
	// EarlierFullReview: an earlier round of this close already ran the full
	// review (a REWORK). Sticky, so a fix that shrinks the diff cannot send the
	// next round back to the small-diff recipe (#231 BR-18).
	EarlierFullReview bool
	LedgerErr         error // the boundary ledger could not be read
}

// Measure sizes a window from its numstat rows — a rename is its destination —
// summing the insertions of code files only. Pure.
func Measure(stats []churn.FileStat, milestones []string) Size {
	size := Size{Milestones: milestones}
	for _, st := range stats {
		if churn.IsCodeFile(st.Path) {
			size.AddedLines += st.Insertions
		}
	}
	return size
}

// Crossings returns one reason per limit the window crosses, in a fixed order,
// or nil when it stays inside the shell. An unreadable ledger is a crossing
// too: it fails toward the full flow rather than past it.
func (s Size) Crossings() []string {
	var out []string
	if s.AddedLines > MaxAddedLines {
		out = append(out, fmt.Sprintf("%d added lines in code files (limit %d)", s.AddedLines, MaxAddedLines))
	}
	if len(s.Milestones) > 0 {
		out = append(out, "the Plan has Mx milestones: "+strings.Join(s.Milestones, ", "))
	}
	if s.EarlierFullReview {
		out = append(out, "an earlier round of this close already ran the full review")
	}
	if s.LedgerErr != nil {
		out = append(out, fmt.Sprintf("the boundary ledger is unreadable (%v)", s.LedgerErr))
	}
	return out
}

// Upgrade is the record a gate writes when a quick issue crosses the shell:
// full, inferred — whoever had declared quick — with the contract hashes gone,
// since only the quick flow uses them. The only way a gate moves a flow; there
// is no downgrade.
func Upgrade(Flow) Flow { return Flow{kind: Full, provenance: Inferred} }
