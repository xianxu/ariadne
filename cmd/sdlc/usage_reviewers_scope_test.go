package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func reviewerTestScope(t *testing.T, root string, issue int) usageReviewerScope {
	t.Helper()
	if issue == 0 {
		return usageReviewerScope{}
	}
	repository, err := resolveUsageRepository(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	return usageReviewerScope{Repository: repository, Issue: issue}
}

func TestUsageReviewerDiagnosticScopeIsRepositoryQualified(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	peer := testfix.Repo(t, testfix.InitialCommit())
	for repo, origin := range map[string]string{root: "https://github.com/example/selected.git", peer: "git@github.com:example/foreign.git"} {
		testfix.Git(t, repo, "remote", "add", "origin", origin)
		testfix.Git(t, repo, "config", "branch.main.remote", "origin")
		testfix.Git(t, repo, "config", "branch.main.merge", "refs/heads/main")
	}
	selectedScope := reviewerTestScope(t, root, 365)
	foreignScope := reviewerTestScope(t, peer, 365)
	if selectedScope.Repository != "github.com/example/selected" || foreignScope.Repository != "github.com/example/foreign" {
		t.Fatalf("publication identities not qualified: %+v %+v", selectedScope, foreignScope)
	}
	foreign := usageAttributionRecord(t, "foreign-identity", usageAttributionReviewerID, root, 1, 1)
	foreign.Intent.Repository = foreignScope.Repository
	// The peer's identity is selected by an explicit scope, independently of the
	// root slice's order. Both #365 artifacts are valid historical ledgers.
	for repo, rounds := range map[string]int{root: 1, peer: 2} {
		path := filepath.Join(repo, "workshop/plans/000365-legacy-plan-gate.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		ledger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 365, Rounds: []gatestate.Round{{N: 1}}}
		if rounds == 2 {
			ledger.Rounds = append(ledger.Rounds, gatestate.Round{N: 2})
			ledger.ReviewerRuns = []reviewusage.Record{foreign}
		}
		if err := os.WriteFile(path, []byte(gatestate.Render(ledger, "repo")), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, roots := range [][]string{{root, peer}, {peer, root}} {
		for _, tc := range []struct {
			scope usageReviewerScope
			want  int64
		}{{selectedScope, 1}, {foreignScope, 2}, {usageReviewerScope{}, 3}} {
			records, coverage, err := loadUsageReviewers(context.Background(), roots, tc.scope)
			if err != nil {
				t.Fatal(err)
			}
			if coverage.Counts["legacy_reviewer_binding_gaps"] != tc.want {
				t.Fatalf("scope=%+v roots=%v want%d legacy gaps, got%+v", tc.scope, roots, tc.want, coverage)
			}
			if len(records) != 1 || records[0] != foreign {
				t.Fatalf("qualified diagnostics dropped exclusion identity: %+v", records)
			}
			sessions := []transcripts.SessionEvidence{usageAttributionSession(foreign.SessionID, root, "", false, 0)}
			bindings := bindUsageReviewers(sessions, records, usageAttributionAt(30))
			got := usageDetailAttribution(usage.Response{Agent: foreign.Intent.Agent, SourceID: foreign.SessionID, At: usageAttributionAt(15)}, usage.Attribution{Role: "worker", Slot: "repo:3", Include: true}, bindings, usageAuthority{Repository: selectedScope.Repository}, 365, true)
			if got.Role != "reviewer" || got.Include {
				t.Fatalf("foreign same-number review became selected worker/reviewer: %+v", got)
			}
		}
	}
	// A second checkout of the selected repository retains its repository scope;
	// publication URL spelling and the checkout path do not determine identity.
	checkout := filepath.Join(t.TempDir(), "alternate")
	testfix.Git(t, root, "worktree", "add", "-b", "alternate-review-scope", checkout)
	if got := reviewerTestScope(t, checkout, 365); got != selectedScope {
		t.Fatalf("alternate checkout changed repository scope: %+v want%+v", got, selectedScope)
	}
	path := filepath.Join(checkout, "workshop/plans/000365-checkout-plan-gate.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: 365, Rounds: []gatestate.Round{{N: 1}}}, "repo")), 0600); err != nil {
		t.Fatal(err)
	}
	_, coverage, err := loadUsageReviewers(context.Background(), []string{checkout, peer}, selectedScope)
	if err != nil {
		t.Fatal(err)
	}
	if coverage.Counts["legacy_reviewer_binding_gaps"] != 1 {
		t.Fatalf("alternate selected checkout's own gaps lost: %+v", coverage)
	}
}

func TestUsageReviewerDiagnosticScopeRequiresRepository(t *testing.T) {
	if _, _, err := loadUsageReviewers(context.Background(), nil, usageReviewerScope{Issue: 365}); err == nil {
		t.Fatal("unqualified issue scope admitted")
	}
}

func TestUsageReviewerDiagnosticIdentityFailureKeepsExclusionVisible(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	scope := reviewerTestScope(t, root, 365)
	foreign := usageAttributionRecord(t, "unqualified-root", usageAttributionReviewerID, root, 1, 1)
	for _, name := range []string{"first", "second"} {
		path := filepath.Join(root, "workshop/plans/000365-"+name+"-gate.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		ledger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 365, Rounds: []gatestate.Round{{N: 1}}}
		if name == "first" {
			ledger.ReviewerRuns = []reviewusage.Record{foreign}
		}
		if err := os.WriteFile(path, []byte(gatestate.Render(ledger, "repo")), 0600); err != nil {
			t.Fatal(err)
		}
	}
	original := resolveWorkspace
	lookups := 0
	resolveWorkspace = func(dir string) (workspace.Identity, error) {
		lookups++
		return workspace.Identity{}, errors.New("diagnostic identity unavailable")
	}
	t.Cleanup(func() { resolveWorkspace = original })
	records, coverage, err := loadUsageReviewers(context.Background(), []string{root}, scope)
	if err != nil {
		t.Fatal(err)
	}
	if coverage.Complete() || coverage.Counts["unavailable_reviewer_bindings"] != 1 || coverage.Counts["legacy_reviewer_binding_gaps"] != 0 {
		t.Fatalf("unknown diagnostic repository silently assigned or dropped: %+v", coverage)
	}
	if lookups != 1 || len(records) != 1 || records[0] != foreign {
		t.Fatalf("identity resolution not cached or exclusion lost: lookups=%d records=%+v", lookups, records)
	}
}
