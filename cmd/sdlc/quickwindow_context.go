package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// The publish gate only needs the base of a whole-issue quick review. Gather
// exactly those facts under the caller's context, then reuse the pure window
// decision. This avoids an uncancellable read in the general review planner.
func quickBoundaryWindowBase(ctx context.Context, issueID string) (string, error) {
	out, _, err := gitx.RunGitContext(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("quick review HEAD: %w", err)
	}
	head := strings.TrimSpace(string(out))
	base, stale, baseErr := reviewMainBase(ctx, head)
	if err := ctx.Err(); err != nil {
		return "", err
	}
	facts := windowFacts{Head: head, MainBase: base, MainStale: stale}
	if baseErr != nil {
		facts.MainBaseErr = baseErr.Error()
	}
	if base == "" {
		entries, err := gitx.LogReverseContext(ctx)
		if err != nil {
			return "", fmt.Errorf("quick review history: %w", err)
		}
		for _, e := range entries {
			if !strings.Contains(e.Subject, "#"+issueID) {
				continue
			}
			out, _, err := gitx.RunGitContext(ctx, "rev-parse", "--verify", e.SHA+"^")
			if err == nil {
				facts.BranchStart = strings.TrimSpace(string(out))
			} else if ctx.Err() != nil {
				return "", ctx.Err()
			} else {
				facts.BranchStart = e.SHA
			}
			break
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return planReviewWindow(facts).Base, nil
}

func quickWindowFileStats(ctx context.Context, base, head string) ([]churn.FileStat, error) {
	span := base + ".." + head
	out, _, err := gitx.RunGitContext(ctx, "diff", "--numstat", "-z", span)
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat -z %s: %w", span, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return churn.ParseNumstatZ(string(out)), nil
}

// Kept separate from the general close-window path so other review callers
// continue to get their existing window measurement and error semantics.
