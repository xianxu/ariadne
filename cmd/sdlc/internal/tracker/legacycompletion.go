package tracker

import (
	"fmt"
)

// LegacyCloseEvidence is a shell-observed source-file version. Position is its
// immutable newest-first Git history position; caller order has no authority.
// HasReview identifies a review-bearing close; the shell validates its review
// and saved measurement objects after this pure generation selection.
type LegacyCloseEvidence struct {
	Position                              int
	Commit, Path, Status, Actual, Started string
	HasReview                             bool
}
type LegacyCloseSelection struct{ TerminalCommit, CloseCommit string }

// SelectLegacyClose chooses the final terminal generation and its latest
// review-bearing close without turning metadata edits into new closes.
func SelectLegacyClose(observed []LegacyCloseEvidence, actual, started string) (LegacyCloseSelection, error) {
	var out LegacyCloseSelection
	states := make([]LegacyCloseEvidence, len(observed))
	for _, state := range observed {
		if state.Position < 0 || state.Position >= len(states) || state.Commit == "" || states[state.Position].Commit != "" {
			return out, fmt.Errorf("ambiguous legacy history ordering")
		}
		states[state.Position] = state
	}
	if len(states) == 0 || states[0].Status != "done" {
		return out, fmt.Errorf("terminal generation is not proven")
	}
	terminalIndex := 0
	for terminalIndex+1 < len(states) && states[terminalIndex+1].Status == "done" {
		terminalIndex++
	}
	doneRunIndex := terminalIndex
	// A real archive/rename transition is a later end bound than a done
	// record at its old path. Never predate that transition with old metadata.
	for i := 1; i <= doneRunIndex; i++ {
		if states[i].Path != states[0].Path {
			terminalIndex = i - 1
			break
		}
	}
	if states[terminalIndex].Actual != actual || states[terminalIndex].Started != started {
		return out, fmt.Errorf("terminal actual/started differs from adopted generation")
	}
	out.TerminalCommit = states[terminalIndex].Commit
	for i := doneRunIndex + 1; i < len(states); i++ {
		if states[i].Status == "done" {
			return out, fmt.Errorf("ambiguous reopened terminal generation")
		}
	}
	for i, s := range states {
		if !s.HasReview || (s.Status != "codecomplete" && s.Status != "done") {
			continue
		}
		if s.Actual != actual || s.Started != started {
			return out, fmt.Errorf("latest close actual/started differs from terminal generation")
		}
		for _, later := range states[:i] {
			if later.Status != "codecomplete" && later.Status != "done" {
				return out, fmt.Errorf("issue reopened after its close")
			}
		}
		out.CloseCommit = s.Commit
		return out, nil
	}
	return out, nil
}
