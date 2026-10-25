package main

import (
	"strings"
	"time"
)

// affectedRerun is what a known-flake child of an affected run must run
// besides the flaky tests (#586): the selection's groups, under the repo's
// affected command. Nothing records which groups the failed attempt reached,
// so every selected group runs again; Go packages do not, since go test
// reports every package and the classifier refuses any failure it cannot name.
type affectedRerun struct{ Command, GoPkgs, Groups []string }

// affectedEnvCommand is the affected command under its fixed env contract
// (#473): sorted, space-separated lists passed as argv through env(1).
func affectedEnvCommand(command, goPkgs, groups []string) []string {
	return append([]string{"env",
		"SDLC_AFFECTED_GO_PKGS=" + strings.Join(goPkgs, " "),
		"SDLC_AFFECTED_GROUPS=" + strings.Join(groups, " "),
	}, command...)
}

// affectedRerunBatch appends the groups command to a flake rerun batch when
// the selection has groups; the suite's batch (a nil rerun) is unchanged.
func affectedRerunBatch(rerun [][]string, a *affectedRerun) [][]string {
	if a == nil || len(a.Groups) == 0 {
		return rerun
	}
	return append(append([][]string(nil), rerun...), affectedEnvCommand(a.Command, nil, a.Groups))
}

// rerunEnvelope is a known-flake child's slot weight and timeout (#586 D6):
// a few named tests take one slot under the rerun cap; appended groups are
// the affected check's own workload, bounded by its declaration.
func rerunEnvelope(c localCheck, minute time.Duration) (weight int, timeout time.Duration) {
	declared := time.Duration(c.TimeoutMinutes) * minute
	if c.Rerun == nil || len(c.Rerun.Groups) == 0 {
		return 1, min(knownFlakeRerunTimeout, declared)
	}
	return max(1, c.Weight), declared
}
