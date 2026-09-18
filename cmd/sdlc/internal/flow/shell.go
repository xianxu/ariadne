package flow

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
)

// Size is what the close-time shell check measures of a review window (#231).
type Size struct {
	CodeFiles   []string // changed code files (churn.IsCodeFile), in diff order
	AddedLines  int      // insertions in those files — git's `+` count, as churn reports it
	Surfaces    []string // changed paths that match a declared shared surface
	Milestones  []string // Mx rows in the Plan
	SurfacesErr error    // the shared-surface declaration could not be read
	// EarlierFullReview: an earlier round of this close already ran the full
	// review (a REWORK). Sticky, so a fix that shrinks the diff cannot send the
	// next round back to the small-diff recipe (#231 BR-18).
	EarlierFullReview bool
	LedgerErr         error // the boundary ledger could not be read
}

// Measure sizes a window. files is the window's changed paths as git reports
// them with rename detection on — a rename is its destination — which is what
// the window CHANGED, so code files are counted over it (a binary, which numstat
// cannot count, still counts as a file). touched is every path on either side of
// the window with rename detection off, which is what the window TOUCHED: moving
// a file out of a shared surface, or renaming the declaration itself, touches it,
// so surfaces are matched over touched (#231 BR-24). stats are the numstat rows,
// summed over code files only. Pure.
func Measure(files, touched []string, stats []churn.FileStat, s Surfaces, surfacesErr error, milestones []string) Size {
	size := Size{Milestones: milestones, SurfacesErr: surfacesErr}
	code := map[string]bool{}
	for _, f := range files {
		if churn.IsCodeFile(f) {
			size.CodeFiles = append(size.CodeFiles, f)
			code[f] = true
		}
	}
	for _, f := range touched {
		if s.Match(f) {
			size.Surfaces = append(size.Surfaces, f)
		}
	}
	for _, st := range stats {
		if code[st.Path] {
			size.AddedLines += st.Insertions
		}
	}
	return size
}

// Crossings returns one reason per limit the window crosses, in a fixed order,
// or nil when it stays inside the shell. An unreadable declaration is a
// crossing too: it fails toward the full flow rather than past it.
func (s Size) Crossings() []string {
	var out []string
	if n := len(s.CodeFiles); n > MaxCodeFiles {
		out = append(out, fmt.Sprintf("%d code files changed (limit %d): %s", n, MaxCodeFiles, capList(s.CodeFiles)))
	}
	if s.AddedLines > MaxChangedLines {
		out = append(out, fmt.Sprintf("%d added lines in code files (limit %d)", s.AddedLines, MaxChangedLines))
	}
	if len(s.Surfaces) > 0 {
		out = append(out, "declared shared surface touched: "+capList(s.Surfaces))
	}
	if len(s.Milestones) > 0 {
		out = append(out, "the Plan has Mx milestones: "+strings.Join(s.Milestones, ", "))
	}
	if s.SurfacesErr != nil {
		out = append(out, fmt.Sprintf("the shared-surface declaration is unreadable (%v)", s.SurfacesErr))
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

// listCap bounds how many paths a crossing names: the reason lands on one Log
// line, and a hundred-file diff needs its count, not its inventory.
const listCap = 5

func capList(paths []string) string {
	if len(paths) <= listCap {
		return strings.Join(paths, ", ")
	}
	return fmt.Sprintf("%s, and %d more", strings.Join(paths[:listCap], ", "), len(paths)-listCap)
}
