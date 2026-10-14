package main

import (
	"context"
	"io"
	"strings"
	"testing"
)

// The legacy path never integrates automatically. Check main before spending
// suite time and again after the checks, without altering either branch tip.
func TestRunMerge_LegacyRefusesMovedMain(t *testing.T) {
	for _, when := range []string{"before", "during-checks"} {
		t.Run(when, func(t *testing.T) {
			root := tempRepo(t)
			remote := git(t, root, "remote", "get-url", "origin")
			other := t.TempDir()
			git(t, "", "clone", "-q", remote, other)
			git(t, other, "config", "user.name", "test")
			git(t, other, "config", "user.email", "test@example.com")
			advance := func() {
				procedureWrite(t, other, "main-change.go", "package main\n")
				git(t, other, "add", ".")
				git(t, other, "commit", "-qm", "main advanced")
				git(t, other, "push", "origin", "main")
			}
			head := git(t, root, "rev-parse", "HEAD")
			g := &legacyMovingMainGH{e2eGH: e2eGH{openPR: "42"}}
			if when == "before" {
				advance()
			} else {
				g.onChecks = advance
			}
			swapMergeDeps(t, g, nil)
			f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
			msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
			if !died || !strings.Contains(msg, "main moved") || !strings.Contains(msg, "git merge origin/main") || g.prMergeCalls != 0 {
				t.Fatalf("died=%v merges=%d message=%s", died, g.prMergeCalls, msg)
			}
			if got := git(t, root, "rev-parse", "HEAD"); got != head {
				t.Fatalf("branch changed %s -> %s", head, got)
			}
			if got := git(t, remote, "rev-parse", "refs/heads/feature"); got != head {
				t.Fatalf("remote branch changed: %s", got)
			}
		})
	}
}

type legacyMovingMainGH struct {
	e2eGH
	onChecks func()
}

func (g *legacyMovingMainGH) PRChecks(ctx context.Context, repo string, n int) ([]prCheck, error) {
	if g.onChecks != nil {
		f := g.onChecks
		g.onChecks = nil
		f()
	}
	return g.e2eGH.PRChecks(ctx, repo, n)
}
