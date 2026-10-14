package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// validateBoundaryResponse admits a complete reviewer handoff against precisely the
// ledger view it was shown. Pure: rejected attempts never become ledger rounds.
func validateBoundaryResponse(output string, scoped gatestate.Ledger) error {
	verdict := judge.ParseVerdict(output)
	if verdict == judge.VerdictUnknown {
		return fmt.Errorf("missing known boundary verdict")
	}
	report, ok := gatestate.ParseFindingsBlock(output)
	if !ok {
		return fmt.Errorf("missing valid findings block")
	}
	round := gatestate.AssignIDs(scoped, report, gatestate.HighestRound(scoped)+1, "", "")
	if _, err := gatestate.ApplyChecked(scoped, round); err != nil {
		return err
	}
	dispositions := map[string]gatestate.Disposition{}
	for _, d := range report.Dispositions {
		dispositions[d.ID] = d
	}
	for _, live := range gatestate.LiveDisputes(scoped) {
		if live.Finding.Kind != "unverifiable" {
			continue
		}
		d, answered := dispositions[live.Finding.ID]
		if !answered || strings.TrimSpace(d.Note) == "" {
			return fmt.Errorf("unverifiable dispute %s requires a judge disposition with a note answering its evidence", live.Finding.ID)
		}
	}
	if verdict != judge.VerdictRework {
		return nil
	}
	m := vocab.Finding()
	for _, f := range report.New {
		if (f.ID == "" || f.ID == "new") && m.Blocks(f.Severity) && strings.TrimSpace(f.Title) != "" && strings.TrimSpace(f.Detail) != "" {
			return nil
		}
	}
	for _, r := range scoped.Rounds {
		for _, f := range r.New {
			if d, ok := dispositions[f.ID]; ok && !m.Closes(d.State) && m.Blocks(f.Severity) {
				return nil
			}
		}
	}
	return fmt.Errorf("REWORK requires a new blocking finding with title and detail, or an explicit not-addressed disposition of a known blocking finding")
}
