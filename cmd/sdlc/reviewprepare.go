package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
)

const reviewPreparationPath = ".sdlc/review-prepare.json"

// readReviewPreparation reads only immutable Git objects, never the worktree.
// Resolving the supplied revision once also pins callers that pass a branch.
func readReviewPreparation(repoRoot, head string) (*reviewprepare.Config, error) {
	fail := func(err error) (*reviewprepare.Config, error) {
		return nil, fmt.Errorf("read %s at %s: %w", reviewPreparationPath, head, err)
	}
	pinned, err := gitx.RunGit("-C", repoRoot, "rev-parse", "--verify", "--end-of-options", head+"^{commit}")
	if err != nil {
		return fail(err)
	}
	for _, path := range []string{".sdlc", reviewPreparationPath} {
		raw, err := gitx.RunGit("-C", repoRoot, "ls-tree", "-z", "--full-tree", strings.TrimSpace(string(pinned)), "--", path)
		if err != nil {
			return fail(err)
		}
		if len(raw) == 0 {
			return nil, nil
		}
		header, name, ok := strings.Cut(strings.TrimSuffix(string(raw), "\x00"), "\t")
		fields := strings.Fields(header)
		if !ok || name != path || len(fields) != 3 {
			return fail(fmt.Errorf("invalid Git tree entry"))
		}
		if path == ".sdlc" {
			if fields[0] != "040000" || fields[1] != "tree" {
				return fail(fmt.Errorf(".sdlc must be a directory"))
			}
			continue
		}
		if (fields[0] != "100644" && fields[0] != "100755") || fields[1] != "blob" {
			return fail(fmt.Errorf("declaration must be a regular file"))
		}
		blob, err := gitx.RunGit("-C", repoRoot, "cat-file", "blob", fields[2])
		if err != nil {
			return fail(err)
		}
		config, err := reviewprepare.Parse(blob)
		if err != nil {
			return fail(err)
		}
		return &config, nil
	}
	return nil, nil
}
