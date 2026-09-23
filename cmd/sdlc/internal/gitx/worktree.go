package gitx

import "github.com/xianxu/ariadne/pkg/workspace"

type Worktree = workspace.Worktree

func ParseWorktrees(b []byte) ([]Worktree, error) { return workspace.ParseWorktrees(b) }
