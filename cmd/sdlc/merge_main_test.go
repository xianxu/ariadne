package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestRunMerge_LegacyReleasesRepoLockDuringLocalCheck(t *testing.T) {
	signalDir := t.TempDir()
	started, release := filepath.Join(signalDir, "started"), filepath.Join(signalDir, "release")
	t.Cleanup(func() { _ = os.WriteFile(release, nil, 0o600) })
	command := fmt.Sprintf("touch %q; i=0; while [ ! -e %q ] && [ \"$i\" -lt 600 ]; do sleep 0.05; i=$((i+1)); done; test -e %q", started, release, release)
	gh, _ := localChecksMergeRepo(t, command, 0, "")
	cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
	cmd.SetContext(context.Background())
	var locks atomic.Int32
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	f.Lock = func(run func() error) error {
		locks.Add(1)
		return withRequiredRepoTransactionLock(cmd, run)
	}
	finished := make(chan error, 1)
	go func() { finished <- runMerge(io.Discard, io.Discard, f) }()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(started); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("local check did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if locks.Load() == 0 {
		t.Fatal("merge never acquired the repo lock")
	}
	other := markManualLockCommand(&cobra.Command{Use: "claim"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	other.SetContext(ctx)
	if err := withRequiredRepoTransactionLock(other, func() error { return nil }); err != nil {
		t.Fatalf("another command could not acquire repo lock during local check: %v", err)
	}
	if err := os.WriteFile(release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-finished:
		if err != nil || gh.prMergeCalls != 1 {
			t.Fatalf("merge err=%v calls=%d", err, gh.prMergeCalls)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("merge did not finish")
	}
}

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

func TestRunMerge_LegacyRefusesHeadMovedDuringChecks(t *testing.T) {
	root := tempRepo(t)
	before := git(t, root, "rev-parse", "HEAD")
	g := &legacyMovingMainGH{e2eGH: e2eGH{openPR: "42"}}
	g.onChecks = func() { git(t, root, "commit", "--allow-empty", "-qm", "concurrent local edit") }
	swapMergeDeps(t, g, nil)
	cmd := markManualLockCommand(&cobra.Command{Use: "merge"})
	cmd.SetContext(context.Background())
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history", Lock: func(run func() error) error {
		return withRequiredRepoTransactionLock(cmd, run)
	}}
	msg, died := expectDie(t, func() { _ = runMerge(io.Discard, io.Discard, f) })
	if !died || !strings.Contains(msg, "HEAD moved") || g.prMergeCalls != 0 {
		t.Fatalf("died=%v merges=%d message=%s", died, g.prMergeCalls, msg)
	}
	if got := git(t, root, "rev-parse", "HEAD"); got == before {
		t.Fatal("fixture did not move HEAD")
	}
}

func TestRunMerge_LegacyRefusesHeadMovedAfterPublishGate(t *testing.T) {
	root := tempRepo(t)
	before := git(t, root, "rev-parse", "HEAD")
	g := &e2eGH{openPR: "42"}
	swapMergeDeps(t, g, func(context.Context, string, string, io.Writer) error {
		git(t, root, "commit", "--allow-empty", "-qm", "concurrent published commit")
		git(t, root, "push", "origin", "feature")
		return nil
	})
	f := &mergeFlags{Yes: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { _ = runMerge(io.Discard, io.Discard, f) })
	if !died || !strings.Contains(msg, "HEAD moved") || g.prMergeCalls != 0 {
		t.Fatalf("died=%v merges=%d message=%s", died, g.prMergeCalls, msg)
	}
	if got := git(t, root, "rev-parse", "HEAD"); got == before {
		t.Fatal("fixture did not move HEAD")
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
