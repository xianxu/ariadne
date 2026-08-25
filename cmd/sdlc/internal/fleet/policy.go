package fleet

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ResolvePolicy(policy PolicyCapabilityValue, paths CanonicalPaths) PolicyResult {
	rel, err := filepath.Rel(paths.WorktreeRoot, paths.Requested)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return policyResultFailure(DiagnosticPathOutsideRepo, paths.Requested, policy.PolicyVersion, "requested path is outside the repository worktree")
	}

	key := ""
	switch policy.KeyKind {
	case "repo":
		key = paths.RepoIdentity
	case "worktree":
		key = paths.WorktreeRoot
	case "declared-root":
		key = resolveDeclaredRoot(paths.WorktreeRoot, rel, policy.Roots)
		if key == "" {
			return policyResultFailure(DiagnosticOutsideDeclaredScope, paths.Requested, policy.PolicyVersion, "requested path is outside every declared admission root")
		}
	default:
		return policyResultFailure(DiagnosticInvalidPolicy, paths.Requested, policy.PolicyVersion, fmt.Sprintf("unknown normalized key kind %q", policy.KeyKind))
	}

	capacity := policy.Capacity
	if policy.Capacity.Limit != nil {
		limit := *policy.Capacity.Limit
		capacity.Limit = &limit
	}
	return PolicyResult{
		OK: true,
		Value: &PolicyResultValue{
			PolicyVersion: policy.PolicyVersion,
			PolicyDigest:  policy.PolicyDigest,
			RepoIdentity:  paths.RepoIdentity,
			AdmissionKey:  key,
			Capacity:      capacity,
			OnCapacity:    policy.OnCapacity,
		},
	}
}

func resolveDeclaredRoot(worktreeRoot, rel string, rules []string) string {
	if rel == "." {
		return ""
	}
	relParts := strings.Split(filepath.ToSlash(rel), "/")
	for _, rule := range rules {
		prefix := strings.TrimSuffix(rule, "/*")
		prefixParts := strings.Split(prefix, "/")
		if len(relParts) <= len(prefixParts) {
			continue
		}
		matched := true
		for i := range prefixParts {
			if relParts[i] != prefixParts[i] {
				matched = false
				break
			}
		}
		if matched && relParts[len(prefixParts)] != "" {
			parts := append(append([]string{}, prefixParts...), relParts[len(prefixParts)])
			return filepath.Join(append([]string{worktreeRoot}, parts...)...)
		}
	}
	return ""
}

func policyResultFailure(code, requested string, version int, message string) PolicyResult {
	return PolicyResult{
		Diagnostic: &PolicyDiagnostic{
			Code:          code,
			Message:       message,
			Path:          requested,
			PolicyVersion: &version,
		},
	}
}
