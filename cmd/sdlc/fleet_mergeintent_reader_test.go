package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func TestFleetMergeIntentsUsesSharedStoreDespiteBrokenMember(t *testing.T) {
	isolateFleetIntegrationGit(t)
	base := canonicalRoot(t.TempDir())
	root := newFleetE2ERepo(t, base, "z-healthy", "")
	broken := filepath.Join(base, "a-prunable")
	runFleetIntegrationGit(t, root, "worktree", "add", "-b", "000624-broken", broken)
	if err := os.Remove(filepath.Join(broken, ".git")); err != nil {
		t.Fatal(err)
	}
	common := filepath.Join(root, ".git")
	line, err := landingLineDir(execGitRunner{}, root)
	if err != nil {
		t.Fatal(err)
	}
	hold, err := landingline.OpenIntents(line).Hold(landingline.Intent{Branch: "main", Head: git(t, root, "rev-parse", "HEAD"), Root: root, PID: 4242, StartedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if err := hold.Release(true); err != nil {
		t.Fatal(err)
	}
	got, err := collectFleetMergeIntents(t.Context(), fleet.Inventory{Rows: []fleet.TreeRow{fleetIntentRow(broken, common), fleetIntentRow(root, common)}}, observeFleetMergeIntents, func() string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	row := findFleetIntegrationRow(t, got, root)
	if row.MergeIntentsState != "complete" || len(row.MergeIntents) != 1 || row.MergeIntents[0].State != "orphaned" {
		t.Fatalf("broken member hid accessible common store: %+v", row)
	}
}

func TestFleetMergeIntentsReaderRetainsEarlierObservations(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	root := roots[1]
	dir, err := landingLineDir(mergeRunner, root)
	if err != nil {
		t.Fatal(err)
	}
	store := landingline.OpenIntents(dir)
	head := git(t, root, "rev-parse", "HEAD")
	for _, branch := range []string{"000624-first", "000624-second"} {
		git(t, root, "branch", branch)
		hold, err := store.Hold(landingline.Intent{Branch: branch, Head: head, Root: canonicalRoot(root), PID: 123, StartedAt: time.Now().UTC()})
		if err != nil {
			t.Fatal(err)
		}
		if err := hold.Release(true); err != nil {
			t.Fatal(err)
		}
	}
	ordered, err := store.List()
	if err != nil || len(ordered) != 2 {
		t.Fatalf("private store: %v %v", ordered, err)
	}
	failBranch := ordered[1].Intent.Branch
	cause := errors.New("branch authority unavailable")
	r := landingHookRunner{gitRunner: mergeRunner, before: func(_ string, args []string) error {
		if strings.Contains(strings.Join(args, " "), "refs/heads/"+failBranch) {
			return cause
		}
		return nil
	}}
	got, err := observeMergeIntents(t.Context(), r, root, func(mergeIntentReport) bool { return true })
	if !errors.Is(err, cause) || len(got) != 2 {
		t.Fatalf("lost partial results: got=%+v err=%v", got, err)
	}
	found := false
	for _, report := range got {
		if report.Branch == failBranch {
			found = true
			if report.State != "unreadable" || report.Reentry != "" || !strings.Contains(report.Note, cause.Error()) {
				t.Fatalf("failed branch probe invented authority: %+v", report)
			}
			if strings.Contains(formatMergeIntent(report), "remove it") {
				t.Fatal("a failed Git probe must not advise removing a valid recovery intent")
			}
		}
	}
	if !found {
		t.Fatal("lost failed observation")
	}
}

func TestFleetMergeIntentsCanonicalPhysicalIdentity(t *testing.T) {
	isolateFleetIntegrationGit(t)
	base := t.TempDir()
	root := filepath.Join(base, "sample")
	runFleetIntegrationGit(t, base, "init", "-b", "main", root)
	runFleetIntegrationGit(t, root, "config", "user.name", "fixture")
	runFleetIntegrationGit(t, root, "config", "user.email", "fixture@example.test")
	runFleetIntegrationGit(t, root, "commit", "--allow-empty", "-qm", "initial")
	head := strings.TrimSpace(string(runFleetIntegrationGit(t, root, "rev-parse", "HEAD")))
	line, err := landingLineDir(execGitRunner{}, root)
	if err != nil {
		t.Fatal(err)
	}
	physical := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(line), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(physical, line); err != nil {
		t.Fatal(err)
	}
	hold, err := landingline.OpenIntents(line).Hold(landingline.Intent{Branch: "main", Head: head, Root: canonicalRoot(root), PID: 123, StartedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	defer hold.Release(false)
	reports, err := observeFleetMergeIntents(context.Background(), root, "")
	if err != nil || len(reports) != 1 {
		t.Fatalf("observe: %v %v", reports, err)
	}
	want, err := filepath.EvalSymlinks(reports[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].Identity != want {
		t.Fatalf("identity %q is not physical %q", reports[0].Identity, want)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("private target must not be followed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(line, "intents", "invalid")); err != nil {
		t.Fatal(err)
	}
	reports, err = observeFleetMergeIntents(context.Background(), root, "")
	if err != nil || len(reports) != 2 {
		t.Fatalf("invalid entry: %+v %v", reports, err)
	}
	for _, report := range reports {
		if report.State == "unreadable" && report.Path != filepath.Join(canonicalRoot(physical), "intents", "invalid") {
			t.Fatalf("rejected symlink's diagnostic followed its target: %+v", report)
		}
	}
}

func TestFleetMergeIntentsBranchProbeGrowth(t *testing.T) {
	isolateFleetIntegrationGit(t)
	for _, n := range []int{8, 16} {
		base := t.TempDir()
		root := filepath.Join(base, "sample")
		runFleetIntegrationGit(t, base, "init", "-b", "main", root)
		runFleetIntegrationGit(t, root, "config", "user.name", "fixture")
		runFleetIntegrationGit(t, root, "config", "user.email", "fixture@example.test")
		runFleetIntegrationGit(t, root, "commit", "--allow-empty", "-qm", "initial")
		head := strings.TrimSpace(string(runFleetIntegrationGit(t, root, "rev-parse", "HEAD")))
		line, err := landingLineDir(execGitRunner{}, root)
		if err != nil {
			t.Fatal(err)
		}
		store := landingline.OpenIntents(line)
		for i := 0; i < n; i++ {
			branch := fmt.Sprintf("000624-case-%d", i)
			runFleetIntegrationGit(t, root, "branch", branch)
			hold, err := store.Hold(landingline.Intent{Branch: branch, Head: head, Root: canonicalRoot(root), PID: 4242, StartedAt: time.Now().UTC()})
			if err != nil {
				t.Fatal(err)
			}
			if err := hold.Release(true); err != nil {
				t.Fatal(err)
			}
		}
		calls, probes := 0, 0
		r := landingHookRunner{gitRunner: execGitRunner{}, before: func(_ string, args []string) error {
			calls++
			if len(args) > 1 && args[0] == "rev-parse" && args[1] == "--verify" {
				probes++
			}
			return nil
		}}
		got, err := observeMergeIntentsWithPause(t.Context(), r, root, func(mergeIntentReport) bool { return true }, func() string { return "" })
		if err != nil || len(got) != n || probes != n || calls != n+1 {
			t.Fatalf("N=%d reports=%d Git=%d probes=%d err=%v", n, len(got), calls, probes, err)
		}
	}
}
