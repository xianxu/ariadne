package fleet

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

const IssueProvenanceBranchPrefix = "branch-prefix"

// IssueRecord is one same-repository lookup result for a six-digit issue ID.
type IssueRecord struct {
	Ref            string
	DeclaredStatus string
}

// IssueLookup returns every same-repository issue matching a six-digit ID.
type IssueLookup func(id string) ([]IssueRecord, error)

// IssueAssociation is issue metadata carried alongside measured tree facts.
type IssueAssociation struct {
	Ref            string `json:"ref"`
	DeclaredStatus string `json:"declared_status"`
	Provenance     string `json:"provenance"`
}

// AssociateBranchIssue associates only a whole issue-prefixed branch with one
// unambiguous issue in the same repository. No-association results are non-nil
// so the JSON contract renders [] rather than null.
func AssociateBranchIssue(branch string, lookup IssueLookup) ([]IssueAssociation, error) {
	associations := make([]IssueAssociation, 0)
	if lookup == nil || strings.Contains(branch, "\\") {
		return associations, nil
	}

	candidate := branch
	if slash := strings.IndexByte(candidate, '/'); slash >= 0 {
		candidate = candidate[:slash]
	}
	// ParseFilename intentionally accepts paths for legacy callers. Isolate the
	// leading branch component before calling it so a slash-prefixed lookalike can
	// never be rescued by its basename; an issue prefix must begin the whole branch.
	id, _, ok := issue.ParseFilename(candidate + ".md")
	if !ok {
		return associations, nil
	}

	matches, err := lookup(id)
	if err != nil {
		return associations, fmt.Errorf("lookup issue %s: %w", id, err)
	}
	if len(matches) != 1 {
		return associations, nil
	}
	return append(associations, IssueAssociation{
		Ref:            matches[0].Ref,
		DeclaredStatus: matches[0].DeclaredStatus,
		Provenance:     IssueProvenanceBranchPrefix,
	}), nil
}
