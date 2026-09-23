package main

import (
	"fmt"
	"os"
	"path/filepath"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// projectWorkspaceOverlays resolves Git topology at the caller boundary. Fleet
// projects belong to primary checkouts, except that the caller's checkout
// supplies its own repo's content. Linked worktrees alongside primaries must
// not become separate project owners or peer-write targets.
func projectWorkspaceOverlays(identity workspace.Identity) ([]projectdoc.CheckoutOverlay, error) {
	overlay := projectdoc.CheckoutOverlay{PrimaryRoot: identity.PrimaryRoot, WorktreeRoot: identity.WorktreeRoot, Repo: identity.Repo}
	siblings, err := projectdoc.FleetRepoDirs(identity.FleetRoot)
	if err != nil {
		return nil, err
	}
	for _, dir := range siblings {
		if dir == identity.PrimaryRoot {
			continue
		}
		if dir == identity.WorktreeRoot {
			overlay.ExcludeRoots = append(overlay.ExcludeRoots, dir)
			continue
		}
		// Legacy project homes need no Git identity. A .git entry, including a
		// broken one, instead promises Git evidence and failures must surface.
		if _, err := os.Lstat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		vantage, err := workspace.NormalizeVantage(execGitRunner{}, dir)
		if err != nil {
			return nil, fmt.Errorf("project fleet identity: %w", err)
		}
		if vantage.WorktreeRoot != vantage.PrimaryRoot {
			overlay.ExcludeRoots = append(overlay.ExcludeRoots, dir)
		}
	}
	return []projectdoc.CheckoutOverlay{overlay}, nil
}
