// issuelist.go — `sdlc issue list` filtering (#348): a pure predicate over
// IssueState built from the flags.
package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/repoenum"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// issueListFilter is the AND of its parts; within --area and --severity the
// values are alternatives (OR).
type issueListFilter struct {
	status     string
	areas      []string // segment prefixes
	severities []string
}

// newIssueListFilter validates the flags into a filter.
func newIssueListFilter(status string, areas, severities []string) (issueListFilter, error) {
	if status != "" && !isValidStatus(status) {
		return issueListFilter{}, fmt.Errorf("invalid status %q (valid: %s)", status, strings.Join(vocab.Issue().AllStatuses(), ", "))
	}
	spec := vocab.Issue().RepoEnums.Fields["area"]
	for _, a := range areas {
		if !repoenum.ValidValue(spec, a) {
			return issueListFilter{}, fmt.Errorf("--area %q is not a /-path of lowercase segments", a)
		}
	}
	valid := vocab.Issue().Enums["severity"]
	for _, s := range severities {
		if !slices.Contains(valid, s) {
			return issueListFilter{}, fmt.Errorf("--severity %q is not one of %s", s, strings.Join(valid, ", "))
		}
	}
	return issueListFilter{status: status, areas: areas, severities: severities}, nil
}

func (f issueListFilter) matches(is IssueState) bool {
	if f.status != "" && is.Status != f.status {
		return false
	}
	if len(f.severities) > 0 && !slices.Contains(f.severities, is.Severity) {
		return false
	}
	if len(f.areas) > 0 {
		return slices.ContainsFunc(is.Area, func(a string) bool {
			return slices.ContainsFunc(f.areas, func(p string) bool { return repoenum.MatchPrefix(p, a) })
		})
	}
	return true
}
