package fleet

import (
	"path/filepath"
	"testing"
)

func TestResolvePolicyKeyProperties(t *testing.T) {
	repo := filepath.Join(string(filepath.Separator), "fleet", "kbench")
	identity := filepath.Join(repo, ".git")
	bounded := 1

	tests := []struct {
		name      string
		policy    PolicyCapabilityValue
		worktree  string
		paths     []string
		wantEqual bool
		wantCode  string
	}{
		{
			name:      "repo nested paths share one key",
			policy:    policyValue("repo", nil, Capacity{Kind: "bounded", Limit: &bounded}, "reject"),
			worktree:  repo,
			paths:     []string{repo, filepath.Join(repo, "competition", "a")},
			wantEqual: true,
		},
		{
			name:      "one worktree shares one key",
			policy:    policyValue("worktree", nil, Capacity{Kind: "bounded", Limit: &bounded}, "provision-worktree"),
			worktree:  filepath.Join(repo, ".worktrees", "one"),
			paths:     []string{filepath.Join(repo, ".worktrees", "one"), filepath.Join(repo, ".worktrees", "one", "nested")},
			wantEqual: true,
		},
		{
			name:      "declared root nested paths share one key",
			policy:    policyValue("declared-root", []string{"competition/*"}, Capacity{Kind: "bounded", Limit: &bounded}, "reject"),
			worktree:  repo,
			paths:     []string{filepath.Join(repo, "competition", "a"), filepath.Join(repo, "competition", "a", "nested")},
			wantEqual: true,
		},
		{
			name:      "distinct declared roots do not collide",
			policy:    policyValue("declared-root", []string{"competition/*"}, Capacity{Kind: "bounded", Limit: &bounded}, "reject"),
			worktree:  repo,
			paths:     []string{filepath.Join(repo, "competition", "a"), filepath.Join(repo, "competition", "b")},
			wantEqual: false,
		},
		{
			name:     "outside declared scope fails closed",
			policy:   policyValue("declared-root", []string{"competition/*"}, Capacity{Kind: "bounded", Limit: &bounded}, "reject"),
			worktree: repo,
			paths:    []string{filepath.Join(repo, "notes")},
			wantCode: DiagnosticOutsideDeclaredScope,
		},
		{
			name:     "inconsistent canonical paths fail as invalid normalized policy input",
			policy:   policyValue("repo", nil, Capacity{Kind: "unbounded"}, ""),
			worktree: repo,
			paths:    []string{filepath.Join(string(filepath.Separator), "elsewhere")},
			wantCode: DiagnosticInvalidPolicy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make([]PolicyResult, 0, len(tt.paths))
			for _, requested := range tt.paths {
				results = append(results, ResolvePolicy(tt.policy, CanonicalPaths{
					RepoIdentity: identity,
					RepoRoot:     repo,
					WorktreeRoot: tt.worktree,
					Requested:    requested,
				}))
			}
			if tt.wantCode != "" {
				got := results[0]
				if got.OK || got.Diagnostic == nil || got.Diagnostic.Code != tt.wantCode {
					t.Fatalf("ResolvePolicy() = %+v, want diagnostic %q", got, tt.wantCode)
				}
				return
			}
			for _, got := range results {
				if !got.OK || got.Value == nil {
					t.Fatalf("ResolvePolicy() = %+v, want success", got)
				}
				if got.Value.PolicyVersion != tt.policy.PolicyVersion || got.Value.PolicyDigest != tt.policy.PolicyDigest || got.Value.Capacity.Kind != tt.policy.Capacity.Kind || got.Value.OnCapacity != tt.policy.OnCapacity {
					t.Fatalf("ResolvePolicy() did not preserve normalized policy: got=%+v policy=%+v", got.Value, tt.policy)
				}
			}
			if len(results) == 2 {
				equal := results[0].Value.AdmissionKey == results[1].Value.AdmissionKey
				if equal != tt.wantEqual {
					t.Fatalf("admission keys equal=%v, want %v: %q / %q", equal, tt.wantEqual, results[0].Value.AdmissionKey, results[1].Value.AdmissionKey)
				}
			}
		})
	}
}

func policyValue(kind string, roots []string, capacity Capacity, action string) PolicyCapabilityValue {
	if roots == nil {
		roots = []string{}
	}
	return PolicyCapabilityValue{
		PolicyVersion: 1,
		PolicyDigest:  testPolicyDigest,
		KeyKind:       kind,
		Roots:         roots,
		Capacity:      capacity,
		OnCapacity:    action,
	}
}
