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
}

// Measure sizes a window. files is the window's changed paths (so a binary,
// which numstat cannot count, still counts as a file); stats are its numstat
// rows, summed over code files only. Pure.
func Measure(files []string, stats []churn.FileStat, s Surfaces, surfacesErr error, milestones []string) Size {
	size := Size{Milestones: milestones, SurfacesErr: surfacesErr}
	code := map[string]bool{}
	for _, f := range files {
		if churn.IsCodeFile(f) {
			size.CodeFiles = append(size.CodeFiles, f)
			code[f] = true
		}
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
		out = append(out, fmt.Sprintf("%d code files changed (limit %d): %s", n, MaxCodeFiles, strings.Join(s.CodeFiles, ", ")))
	}
	if s.AddedLines > MaxChangedLines {
		out = append(out, fmt.Sprintf("%d added lines in code files (limit %d)", s.AddedLines, MaxChangedLines))
	}
	if len(s.Surfaces) > 0 {
		out = append(out, "declared shared surface touched: "+strings.Join(s.Surfaces, ", "))
	}
	if len(s.Milestones) > 0 {
		out = append(out, "the Plan has Mx milestones: "+strings.Join(s.Milestones, ", "))
	}
	if s.SurfacesErr != nil {
		out = append(out, fmt.Sprintf("the shared-surface declaration is unreadable (%v)", s.SurfacesErr))
	}
	return out
}

// Upgrade is the record a gate writes when a quick issue crosses the shell:
// full, inferred — whoever had declared quick — with the contract hashes gone,
// since only the quick flow uses them. The only way a gate moves a flow; there
// is no downgrade.
func Upgrade(Flow) Flow { return Flow{kind: Full, provenance: Inferred} }
