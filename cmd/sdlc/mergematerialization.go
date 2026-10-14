package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// The remote PR has already merged. Fetch once, validate the exact incoming
// tree, and fast-forward to that immutable commit. Resting main never gains a
// local merge commit; divergence needs operator reconciliation.
func materializeMergedMain(ctx context.Context, r gitRunner, root, issuesDir string, inPlace bool) (bool, error) {
	atMain := !inPlace
	out, err := r.GitInDir(root, "fetch", "-q", "origin", "main")
	if err != nil {
		return atMain, fmt.Errorf("fetch merged main: %w\n%s; fetch origin/main then retry", err, out)
	}
	out, err = r.GitInDir(root, "rev-parse", "--verify", "FETCH_HEAD^{commit}")
	if err != nil {
		return atMain, fmt.Errorf("resolve fetched main: %w; fetch origin/main then retry", err)
	}
	target := strings.TrimSpace(string(out))
	out, err = r.GitInDir(root, "rev-parse", "--verify", "refs/heads/main^{commit}")
	if err != nil {
		return atMain, fmt.Errorf("resolve resting main: %w; restore the main branch before retrying", err)
	}
	rest := strings.TrimSpace(string(out))
	reader, err := gitx.NewTrunkFileContext(ctx, root, "origin", "main")
	if err != nil {
		return atMain, err
	}
	for _, ref := range []string{rest, target} {
		if err := verifyAssetMaterialization(ctx, root, reader.ViewOf(ref), issuesDir, rest, target); err != nil {
			return atMain, err
		}
	}
	if inPlace {
		if out, err = r.GitInDir(root, "switch", "main"); err != nil {
			return false, fmt.Errorf("switch to main: %w\n%s", err, out)
		}
		atMain = true
		if err := verifyMaterializedAssets(root, reader.ViewOf(rest), issuesDir); err != nil {
			return true, err
		}
	}
	if out, err = r.GitInDir(root, "merge", "--ff-only", target); err != nil {
		return atMain, fmt.Errorf("fast-forward main: %w\n%s; reconcile main with the fetched commit %s before retrying", err, out, target)
	}
	return atMain, verifyMaterializedAssets(root, reader.ViewOf(target), issuesDir)
}
