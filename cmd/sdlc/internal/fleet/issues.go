package fleet

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const IssueProvenanceBranchPrefix = "branch-prefix"

// IssueRecord is one same-repository lookup result for a six-digit issue ID.
type IssueRecord struct {
	Ref            string
	DeclaredStatus string
}

// IssueLookup returns every same-repository issue matching a six-digit ID.
type IssueLookup func(id string) ([]IssueRecord, error)

// LookupRepoIssues reports one repository's record for a six-digit issue ID.
// With a tracker (#252) the card's status is declared; without one (pre-
// migration) each active details file's status is. Duplicate or unreadable
// details, and invalid statuses, discard partial results. Stable order.
func LookupRepoIssues(repoRoot, id string) ([]IssueRecord, error) {
	records := make([]IssueRecord, 0)
	parsedID, _, validID := issue.ParseFilename(id + "-.md")
	if !validID || parsedID != id {
		return records, nil
	}
	rs, err := repoRecords(repoRoot)
	if err != nil {
		return records, err
	}
	ref := filepath.Base(filepath.Clean(repoRoot)) + "#" + id
	for _, rec := range rs.All() {
		if rec.ID != id {
			continue
		}
		if rec.DetailErr != nil {
			return make([]IssueRecord, 0), fmt.Errorf("read same-repo issue %q: %w", rec.DetailPath, rec.DetailErr)
		}
		status := rec.Status()
		if status == "" || !containsString(vocab.Issue().AllStatuses(), status) {
			return make([]IssueRecord, 0), fmt.Errorf("validate same-repo issue %q: invalid or missing status %q", rec.DetailPath, status)
		}
		records = append(records, IssueRecord{Ref: ref, DeclaredStatus: status})
		if rec.Card != nil {
			break // one card is one record, whatever details copies exist
		}
	}
	return records, nil
}

// repoRecords loads (once per process) a repository's composed issue records.
// Fleet inventory is a read-only view: a stale tracker read is acceptable.
var repoRecords = func() func(string) (tracker.Records, error) {
	var mu sync.Mutex
	cache := map[string]tracker.Records{}
	return func(repoRoot string) (tracker.Records, error) {
		mu.Lock()
		defer mu.Unlock()
		if rs, ok := cache[repoRoot]; ok {
			return rs, nil
		}
		ctx := context.Background()
		repo, err := tracker.RepositoryForCheckout(ctx, repoRoot)
		if err != nil {
			return tracker.Records{}, err
		}
		rs, err := tracker.LoadRecords(ctx, repo, filepath.Join(repoRoot, vocab.Issue().Discovery().Home), tracker.PreferFresh)
		if err != nil {
			return tracker.Records{}, fmt.Errorf("read same-repo issues of %q: %w", repoRoot, err)
		}
		cache[repoRoot] = rs
		return rs, nil
	}
}()

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
	association := IssueAssociation{
		Ref:            matches[0].Ref,
		DeclaredStatus: matches[0].DeclaredStatus,
		Provenance:     IssueProvenanceBranchPrefix,
	}
	if err := validateIssueAssociation(association); err != nil {
		return associations, fmt.Errorf("validate same-repo issue association %s: %w", id, err)
	}
	return append(associations, association), nil
}
