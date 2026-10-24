package refresh

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
)

// HelperResult describes confirmed source progress, never a running waiter's revision.
type HelperResult struct{ State, Reason, Before, After string }

// helperAction is the pure eligibility rule; ancestry comes from captured Git evidence.
func helperAction(branch, head, target string, isAncestor bool) string {
	if branch != "refs/heads/main" {
		return "off-main"
	}
	if head == target {
		return "current"
	}
	if !isAncestor {
		return "diverged"
	}
	return "refresh"
}

// RefreshHelper updates one already-scoped supplier. The caller holds its
// environment setup lease and supplies a descriptor-inheriting Git runner.
func RefreshHelper(ctx context.Context, supplier string, g acquire.GitRunner) HelperResult {
	skipped := func(why string) HelperResult { return HelperResult{State: "skipped", Reason: why} }
	s, err := observe(ctx, g, supplier)
	if err != nil {
		return skipped("checkout is dirty, detached, in an active operation, or cannot be verified")
	}
	if s.branch != "refs/heads/main" {
		return skipped("off-main branch " + s.branch)
	}
	if s.common != filepath.Join(supplier, ".git") || s.gitdir != s.common {
		return skipped("supplier is not an independent ordinary checkout")
	}
	if _, err := g.Run(ctx, supplier, "fetch", "--no-recurse-submodules", "--no-tags", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
		return skipped("fetch origin/main unavailable")
	}
	s.target, err = oid(ctx, g, supplier, "refs/remotes/origin/main")
	if err != nil {
		return skipped("origin/main target unavailable")
	}
	isAncestor, err := ancestor(ctx, g, supplier, s.head, s.target)
	if err != nil {
		return skipped("ancestry unavailable")
	}
	action := helperAction(s.branch, s.head, s.target, isAncestor)
	if action == "diverged" {
		return skipped("main is ahead or divergent")
	}
	if err := revalidate(ctx, g, s); err != nil {
		return skipped("checkout changed before refresh; inspect it")
	}
	if action == "current" {
		return HelperResult{State: "current", Before: s.head, After: s.head}
	}
	rows, err := targetDeclarations(ctx, g, s)
	if err != nil || !reflect.DeepEqual(rows, s.rows) {
		return skipped("dependency declarations changed or unavailable; run ordinary refresh explicitly")
	}
	if err := ignoredCollisions(ctx, g, s, false); err != nil {
		return skipped("ignored-file collision check refused; inspect supplier")
	}
	if err := revalidate(ctx, g, s); err != nil {
		return skipped("checkout changed before fast-forward; inspect it")
	}
	_, updateErr := g.Run(ctx, supplier, "-c", "submodule.recurse=false", "-c", "merge.autoStash=false", "merge", "--ff-only", "--no-edit", "--no-overwrite-ignore", s.target)
	// An error does not prove no effect. Keep uncertain endings and never roll back.
	now, readErr := observe(ctx, g, supplier)
	result := HelperResult{State: "uncertain", Reason: "fast-forward ending unconfirmed; inspect Git status", Before: s.head, After: now.head}
	expected := s
	expected.head = s.target
	if updateErr == nil && readErr == nil && sameStart(expected, now) {
		result.State, result.Reason = "refreshed", ""
	}
	return result
}

func (r HelperResult) Line(supplier string) string {
	if r.State == "current" {
		return ""
	}
	if r.State == "refreshed" {
		return fmt.Sprintf("suite helper refreshed %q: %s -> %s", supplier, r.Before, r.After)
	}
	return fmt.Sprintf("suite helper %s %q: %s; observed HEAD=%q; repair: inspect this checkout, reconcile it explicitly, then retry the original command", r.State, supplier, r.Reason, r.After)
}
