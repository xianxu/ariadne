package main

import (
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// churnForWindow measures churn over [baseLong, HEAD] — the SAME window the boundary
// review and the atlas gate use. Callers pass boundaryWindowBase's result rather than
// re-deriving it (ARCH-DRY), which is what makes the reported churn provably cover the
// commits that were actually reviewed.
//
// Two git reads, and they are deliberately different queries:
//
//   - `diff --numstat base..HEAD` — the NET result: what survived to be reviewed.
//   - `log --numstat --format= base..HEAD` — every commit's insertions, so a file
//     rewritten five times counts five times. That difference IS the rework signal.
//
// An EMPTY base is not an error: boundaryWindowBase returns "" on a docs-only window
// with no `#N` commit, and a close there must still succeed. It yields a zero report.
//
// A BAD base is an error, and reporting it is why this uses gitx.RunGit rather than
// gitx.Capture. Capture flattens any failure to "" (internal/gitx/window.go:50-56, whose
// own doc warns against exactly this use), so a bogus SHA would render
// `churn: prod 0 / test 0 / …` — indistinguishable from a genuinely empty window, in the
// one number introduced to answer "which gates earn their cost". The error goes up; the
// CALLER warns and zeroes.
func churnForWindow(baseLong string) (churn.Report, error) {
	if baseLong == "" {
		return churn.Report{}, nil
	}
	span := baseLong + "..HEAD"

	finalOut, err := gitx.RunGit("diff", "--numstat", span)
	if err != nil {
		return churn.Report{}, fmt.Errorf("git diff --numstat %s: %w", span, err)
	}
	commitOut, err := gitx.RunGit("log", "--numstat", "--format=", span)
	if err != nil {
		return churn.Report{}, fmt.Errorf("git log --numstat %s: %w", span, err)
	}

	final := churn.ParseNumstat(string(finalOut))
	commitTotal := churn.TotalInsertions(churn.ParseNumstat(string(commitOut)))
	return churn.Summarize(final, commitTotal), nil
}
