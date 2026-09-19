package flow

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Size is what the shell measures (#231): of a review window at close, and of
// the issue alone at change-code, where there is no diff yet (AddedLines 0).
type Size struct {
	AddedLines  int      // insertions in code files (churn.IsCodeFile) — git's `+` count, as churn reports it
	DesignLines int      // the design's length (DesignLines)
	Milestones  []string // Mx rows in the Plan
	// EarlierFullReview: an earlier round of this close already ran the full
	// review (a REWORK). Sticky, so a fix that shrinks the diff cannot send the
	// next round back to the small-diff recipe (#231 BR-18).
	EarlierFullReview bool
	LedgerErr         error // the boundary ledger could not be read
}

// Measure sizes a window from its numstat rows — a rename is its destination —
// summing the insertions of code files only, plus the design's length and the
// Plan's Mx rows. change-code passes no stats: the same sizing, before a diff
// exists, so the entry-time inference and the close-time check are one check.
// Pure.
func Measure(stats []churn.FileStat, designLines int, milestones []string) Size {
	size := Size{DesignLines: designLines, Milestones: milestones}
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
	if s.DesignLines > MaxDesignLines {
		out = append(out, fmt.Sprintf("a design of %d lines (limit %d: %s)", s.DesignLines, MaxDesignLines, DesignRule))
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

// GrewPastReview says why a quick issue's final diff can no longer ship on its
// small-diff review — "" while its added lines stay within
// MaxAddedLinesAfterReview. The publish check asks it of the diff as published,
// fixes made after the verdict included. Pure.
func (s Size) GrewPastReview() string {
	if s.AddedLines <= MaxAddedLinesAfterReview {
		return ""
	}
	return fmt.Sprintf("%d added lines in code files since its small-diff review ran (limit %d, twice the shell's %d)",
		s.AddedLines, MaxAddedLinesAfterReview, MaxAddedLines)
}

// designSections are the issue sections DesignRule counts, beside the durable
// plan: where brainstorming lands the design (Spec) and where its steps go
// (Plan). Problem, Log and Revisions grow with the work, not the design.
var designSections = []string{"Spec", "Plan"}

// DesignLines is the design's length: the lines of the issue's design sections
// (fence-aware, through issue.SectionBody) plus the durable plan's. plan is ""
// when there is none. Pure.
func DesignLines(body, plan string) int {
	n := lineCount(plan)
	for _, h := range designSections {
		if s, ok := issue.SectionBody(body, h); ok {
			n += lineCount(s)
		}
	}
	return n
}

// lineCount counts lines as `wc -l` would, less blank lines at either end — the
// gap between a heading and its text is not design.
func lineCount(s string) int {
	s = strings.Trim(s, " \t\r\n")
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// Upgrade is the record a gate writes when a quick issue crosses the shell:
// full, inferred — whoever had declared quick — with the contract hashes gone,
// since only the quick flow uses them. The only way a gate moves a flow; there
// is no downgrade.
func Upgrade(Flow) Flow { return Flow{kind: Full, provenance: Inferred} }
