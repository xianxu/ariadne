package main

import (
	"github.com/xianxu/ariadne/pkg/workspace"
	"path/filepath"
)

// resolveWorkspace is the IO seam for commands that need repository identity.
// WorktreeRoot remains the owner of local code and artifacts; PrimaryRoot is
// never substituted for it. Tests may inject identity into pure filesystem IO.
var resolveWorkspace = func(dir string) (workspace.Identity, error) {
	if dir == "" {
		dir = "."
	}
	return workspace.Resolve(execGitRunner{}, dir, "")
}

func workspaceRepoName(dir string) (string, error) {
	identity, err := resolveWorkspace(dir)
	if err != nil {
		return "", err
	}
	return identity.Repo, nil
}

func defaultWorkspacePath(root, path string, explicit bool) string {
	if explicit || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
