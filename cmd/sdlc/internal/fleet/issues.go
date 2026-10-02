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
	// StaleStatus: the status came from the last-fetched tracker (#252).
	StaleStatus bool
}

// IssueLookup returns every same-repository issue matching a six-digit ID.
type IssueLookup func(id string) ([]IssueRecord, error)

// LookupRepoIssues reports one repository's record for a six-digit issue ID.
// With a tracker (#252) the card's status is declared; without one (pre-
// migration) each active details file's status is. Duplicate or unreadable
// details, and invalid statuses, discard partial results. Stable order.
func LookupRepoIssues(ctx context.Context, repoRoot, id string) ([]IssueRecord, error) {
	records := make([]IssueRecord, 0)
	parsedID, _, validID := issue.ParseFilename(id + "-.md")
	if !validID || parsedID != id {
		return records, nil
	}
	rs, err := repoRecords(ctx, repoRoot)
	if err != nil {
		return records, err
	}
	ref := filepath.Base(filepath.Clean(repoRoot)) + "#" + id
	for _, rec := range rs.All() {
		if rec.ID != id {
			continue
		}
		if rec.CardErr != nil {
			return make([]IssueRecord, 0), fmt.Errorf("read same-repo issue %s: %w", ref, rec.CardErr)
		}
		if rec.DetailErr != nil {
			return make([]IssueRecord, 0), fmt.Errorf("read same-repo issue %q: %w", rec.DetailPath, rec.DetailErr)
		}
		status := rec.Status()
		if status == "" || !containsString(vocab.Issue().AllStatuses(), status) {
			return make([]IssueRecord, 0), fmt.Errorf("validate same-repo issue %q: invalid or missing status %q", rec.DetailPath, status)
		}
		records = append(records, IssueRecord{Ref: ref, DeclaredStatus: status, StaleStatus: rs.Stale && rec.Card != nil})
		if rec.Card != nil {
			break // one card is one record, whatever details copies exist
		}
	}
	return records, nil
}

// repoRecords loads (once per process) a repository's composed issue records.
// Fleet inventory is a read-only view: a stale tracker read is acceptable.
var repoRecords = func() func(context.Context, string) (tracker.Records, error) {
	var mu sync.Mutex
	cache := map[string]tracker.Records{}
	return func(ctx context.Context, repoRoot string) (tracker.Records, error) {
		mu.Lock()
		defer mu.Unlock()
		if rs, ok := cache[repoRoot]; ok {
			return rs, nil
		}
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
	StaleStatus    bool   `json:"stale_status,omitempty"`
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
		StaleStatus:    matches[0].StaleStatus,
	}
	if err := validateIssueAssociation(association); err != nil {
		return associations, fmt.Errorf("validate same-repo issue association %s: %w", id, err)
	}
	return append(associations, association), nil
}

// LookupRepoClaims reads one repository's claims from the same cached tracker
// load the branch-prefix lookup uses (one read per repository, #288). Each
// recorded worktree is re-canonicalized with the helper that produced the
// inventory's tree paths, so placement compares like with like; a path that no
// longer exists keeps its recorded spelling (it matches no row: dangling).
//
// A checkout with no tracker cutover marker and no fetched tracker has no
// claims: it is absent without probing its remote (#288 BR-11), so inventory
// contacts only the remotes of repositories that use the tracker — one fetch
// each, the same read their branch-prefix lookups already make.
func LookupRepoClaims(ctx context.Context, repoRoot string) RepoClaims {
	if cut, err := tracker.CutOver(repoRoot); err != nil {
		return RepoClaims{State: ClaimsUnknown, Error: err.Error()}
	} else if !cut {
		return RepoClaims{State: ClaimsAbsent}
	}
	rs, err := repoRecords(ctx, repoRoot)
	if err != nil {
		return RepoClaims{State: ClaimsUnknown, Error: err.Error()}
	}
	return repoClaimsFrom(rs, filepath.Base(filepath.Clean(repoRoot)), func(p string) string {
		if c, err := canonicalPath(p); err == nil {
			return c
		}
		return p
	})
}

// repoClaimsFrom derives a repository's claim read from its composed records.
func repoClaimsFrom(rs tracker.Records, repoName string, canon func(string) string) RepoClaims {
	switch {
	case !rs.Tracker && !rs.Stale:
		return RepoClaims{State: ClaimsAbsent}
	case !rs.Tracker && rs.FetchErr == nil:
		return RepoClaims{State: ClaimsUnknown, Error: "could not confirm whether the publication remote has an issue tracker (unreachable, none fetched here, no cutover marker)"}
	case !rs.Tracker:
		return RepoClaims{State: ClaimsUnknown, Error: "tracker unreachable and never fetched here: " + rs.FetchErr.Error()}
	}
	out := RepoClaims{State: ClaimsPresent}
	for _, rec := range rs.All() {
		if rec.Duplicate {
			continue
		}
		ref := repoName + "#" + rec.ID
		if rec.CardErr != nil {
			out.Unreadable = append(out.Unreadable, ref)
			continue
		}
		if rec.Card == nil {
			continue
		}
		card := ClaimCard{Ref: ref, Status: rec.Status(), Revision: rec.Card.BlobOID}
		if c, ok, err := issue.CardClaimant(rec.Card.Raw); err != nil {
			out.Unreadable = append(out.Unreadable, ref)
			continue
		} else if ok {
			pub := ClaimantFrom(c)
			pub.Worktree = canon(pub.Worktree)
			card.Claimant = &pub
		}
		out.Cards = append(out.Cards, card)
	}
	var reasons []string
	if rs.Stale {
		out.State = ClaimsStale
		reasons = append(reasons, "tracker unreachable, answered from the last fetch: "+errString(rs.FetchErr))
	}
	if len(out.Unreadable) > 0 {
		out.State = ClaimsPartial
		reasons = append(reasons, unreadableError(out.Unreadable))
	}
	out.Error = strings.Join(reasons, "; ")
	return out
}

func errString(err error) string {
	if err == nil {
		return "no reason recorded"
	}
	return err.Error()
}
