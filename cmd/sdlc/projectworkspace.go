package main

import (
	"fmt"
	"os"
	"path/filepath"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// projectWorkspaceRoots selects concrete content paths without equating the
// authority of independent clones. Local-only repositories are first-class
// roots; same-named fleet copies are shadowed, never duplicate write targets.
func projectWorkspaceRoots(id workspace.Identity) ([]projectdoc.ProjectRoot, error) {
	repos, err := workspaceContentRepos(id)
	if err != nil {
		return nil, err
	}
	eligible, err := projectdoc.FleetRepoDirs(id.FleetRoot)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{id.Repo: true}
	for _, dir := range eligible {
		names[filepath.Base(dir)] = true
	}
	local, err := environmentRepos(id)
	if err != nil {
		return nil, err
	}
	for _, r := range local {
		names[r.Name] = true
	}
	var roots []projectdoc.ProjectRoot
	for _, r := range repos {
		if !names[r.Name] {
			continue
		}
		if _, err := os.Lstat(filepath.Join(r.Root, ".git")); err == nil {
			v, err := workspace.NormalizeVantage(execGitRunner{}, r.Root)
			if err != nil {
				return nil, fmt.Errorf("project repository %s: %w", r.Root, err)
			}
			// The caller and the verified host are intentional worktree roots. Other
			// linked checkouts in the canonical fleet do not become extra project homes.
			if r.Root != id.WorktreeRoot && v.WorktreeRoot != v.PrimaryRoot && (id.EnvironmentHost == nil || r.Root != id.EnvironmentHost.WorktreeRoot) {
				continue
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		roots = append(roots, projectdoc.ProjectRoot{RepoDir: r.Root, Repo: r.Name})
	}
	return roots, nil
}
