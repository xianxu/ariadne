package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/project"
	"github.com/xianxu/ariadne/pkg/workspace"
)

type workspaceRepo struct{ Name, Root string }

// selectWorkspaceRepo chooses content, never common-directory ownership.
func selectWorkspaceRepo(token string, current workspaceRepo, repos []workspaceRepo) (workspaceRepo, error) {
	if token == "" || token == current.Name {
		return current, nil
	}
	var matches []workspaceRepo
	for _, r := range repos {
		if r.Name == token {
			return r, nil
		}
		if strings.HasPrefix(strings.ToLower(r.Name), strings.ToLower(token)) {
			matches = append(matches, r)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) == 0 {
		return workspaceRepo{}, fmt.Errorf("no sibling repo matches %q", token)
	}
	names := make([]string, len(matches))
	for i, r := range matches {
		names[i] = r.Name
	}
	sort.Strings(names)
	return workspaceRepo{}, fmt.Errorf("ambiguous repo %q: matches %s", token, strings.Join(names, ", "))
}

// environmentRepos includes only verified direct repository children. A .git
// entry promises topology; broken evidence must never select a global fallback.
func environmentRepos(id workspace.Identity) ([]workspaceRepo, error) {
	if id.EnvironmentHost == nil {
		return nil, nil
	}
	entries, err := os.ReadDir(id.EnvironmentRoot)
	if err != nil {
		return nil, err
	}
	var out []workspaceRepo
	for _, entry := range entries {
		if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		dir := filepath.Join(id.EnvironmentRoot, entry.Name())
		if _, err := os.Lstat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		v, err := workspace.NormalizeVantage(execGitRunner{}, dir)
		if err != nil {
			return nil, fmt.Errorf("environment repository %s: %w", dir, err)
		}
		if v.WorktreeRoot != dir || v.EnvironmentRoot != id.EnvironmentRoot || v.EnvironmentHost == nil || v.EnvironmentHost.RepoIdentity != id.EnvironmentHost.RepoIdentity {
			return nil, fmt.Errorf("environment repository %s has conflicting Git identity", dir)
		}
		out = append(out, workspaceRepo{filepath.Base(dir), dir})
	}
	return out, nil
}

func workspaceContentRepos(id workspace.Identity) ([]workspaceRepo, error) {
	dirs, err := project.SiblingRepoDirs(id.FleetRoot)
	if err != nil {
		return nil, err
	}
	byName := map[string]workspaceRepo{}
	for _, dir := range dirs {
		byName[filepath.Base(dir)] = workspaceRepo{filepath.Base(dir), dir}
	}
	local, err := environmentRepos(id)
	if err != nil {
		return nil, err
	}
	for _, r := range local {
		byName[r.Name] = r
	}
	byName[id.Repo] = workspaceRepo{id.Repo, id.WorktreeRoot}
	out := make([]workspaceRepo, 0, len(byName))
	for _, r := range byName {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
