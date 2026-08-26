package fleet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// LookupRepoIssues reads every exact issue-ID match from one repository's
// active issue home. It reuses the canonical filename and frontmatter grammars,
// returns stable filename order, and discards partial results on any read or
// parse error.
func LookupRepoIssues(repoRoot, id string) ([]IssueRecord, error) {
	records := make([]IssueRecord, 0)
	parsedID, _, validID := issue.ParseFilename(id + "-.md")
	if !validID || parsedID != id {
		return records, nil
	}
	home := filepath.Join(repoRoot, "workshop", "issues")
	entries, err := os.ReadDir(home)
	if err != nil {
		if os.IsNotExist(err) {
			return records, nil
		}
		return records, fmt.Errorf("read same-repo issue home %q: %w", home, err)
	}
	paths := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		candidateID, _, ok := issue.ParseFilename(entry.Name())
		if ok && candidateID == id {
			paths = append(paths, filepath.Join(home, entry.Name()))
		}
	}
	sort.Strings(paths)
	refID := strings.TrimLeft(id, "0")
	if refID == "" {
		refID = "0"
	}
	ref := filepath.Base(filepath.Clean(repoRoot)) + "#" + refID
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return make([]IssueRecord, 0), fmt.Errorf("read same-repo issue %q: %w", path, err)
		}
		frontmatter, _, err := issue.Parse(string(raw))
		if err != nil {
			return make([]IssueRecord, 0), fmt.Errorf("parse same-repo issue %q: %w", path, err)
		}
		status, _ := issue.GetField(frontmatter, "status")
		records = append(records, IssueRecord{Ref: ref, DeclaredStatus: status})
	}
	return records, nil
}

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
