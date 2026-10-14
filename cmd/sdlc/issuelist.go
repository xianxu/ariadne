// issuelist.go — `sdlc issue list` filtering (#348): a pure predicate over
// IssueState built from the flags.
package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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
	for _, a := range areas {
		if err := issue.CheckAreaSyntax(a); err != nil {
			return issueListFilter{}, fmt.Errorf("--area: %w", err)
		}
	}
	for _, s := range severities {
		if s == "" {
			return issueListFilter{}, errors.New("--severity: empty value")
		}
		if err := issue.CheckSeverity(s); err != nil {
			return issueListFilter{}, fmt.Errorf("--severity: %w", err)
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
