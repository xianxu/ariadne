package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func emptyInitRepo(t *testing.T) *trackerRepo {
	t.Helper()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if canonical, err := filepath.EvalSymlinks(root); err == nil {
		root = canonical
	}
	testfix.Git(t, root, "config", "core.hooksPath", t.TempDir())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	writeRepoFile(t, root, "workshop/issues/.gitkeep", "")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "initial setup scaffold")
	testfix.Git(t, root, "push", "-qu", "origin", "main")
	return &trackerRepo{t: t, root: root, origin: origin}
}

func TestIssueInitStartsTrackerWorkflow(t *testing.T) {
	r := emptyInitRepo(t)
	slotClaimant(t, "ariadne:3")
	if _, stderr, err := executeSDLCTestCommand("issue", "init"); err != nil {
		t.Fatalf("init: %v\n%s", err, stderr)
	}
	marker, err := os.ReadFile(filepath.Join(r.root, tracker.CutoverMarkerPath))
	if err != nil {
		t.Fatal(err)
	}
	root, err := tracker.ParseCutoverMarker(marker)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(testfix.Capture(t, r.origin, "rev-list", "--max-parents=0", "issue-tracker")); got != root {
		t.Fatalf("tracker root %q want %q", got, root)
	}
	before, status := r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
	if _, stderr, err := executeSDLCTestCommand("issue", "init"); err != nil {
		t.Fatalf("repeat init: %v\n%s", err, stderr)
	}
	assertCheckoutUntouched(t, r, before, status)
	if _, stderr, err := executeSDLCTestCommand("issue", "new", "First work"); err != nil {
		t.Fatalf("first issue: %v\n%s", err, stderr)
	}
	assertCheckoutUntouched(t, r, before, status)
	if _, stderr, err := executeSDLCTestCommand("start-plan", "--issue", "1"); err != nil {
		t.Fatalf("first planning: %v\n%s", err, stderr)
	}
	if got := r.git("branch", "--show-current"); got != "000001-first-work" {
		t.Fatalf("planning branch %q", got)
	}
}

func TestIssueInitExcludesBrainAndDeclinedScaffold(t *testing.T) {
	for _, brain := range []bool{true, false} {
		t.Run(map[bool]string{true: "brain", false: "declined"}[brain], func(t *testing.T) {
			r := emptyInitRepo(t)
			if brain {
				writeRepoFile(t, r.root, ".brain/config.md", "brain\n")
			} else {
				r.git("rm", "workshop/issues/.gitkeep")
			}
			r.git("add", ".")
			r.git("commit", "-qm", "setup exclusion")
			r.git("push", "-q")
			before := r.originMain()
			if _, stderr, err := executeSDLCTestCommand("issue", "init"); err != nil {
				t.Fatalf("skip: %v\n%s", err, stderr)
			}
			if r.originMain() != before || testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker") != "" {
				t.Fatal("excluded repository was initialized")
			}
		})
	}
}

func TestIssueInitRefusesExistingLegacyInventory(t *testing.T) {
	r := emptyInitRepo(t)
	writeRepoFile(t, r.root, "workshop/issues/000001-existing.md", "---\nid: 000001\nstatus: open\ncreated: 2026-10-21\n---\n\n# Existing legacy issue\n")
	r.git("add", ".")
	r.git("commit", "-qm", "legacy issue")
	r.git("push", "-q")
	before := r.originMain()
	_, stderr, err := executeSDLCTestCommand("issue", "init")
	if err == nil || !strings.Contains(err.Error()+stderr, "issue migrate") {
		t.Fatalf("nonempty init: %v\n%s", err, stderr)
	}
	if r.originMain() != before || testfix.Capture(t, r.origin, "branch", "--list", "issue-tracker") != "" {
		t.Fatal("nonempty init mutated remote")
	}
}

func TestLegacySetupHintBeforeFirstIssueAndBarePlanning(t *testing.T) {
	r := emptyInitRepo(t)
	for _, args := range [][]string{{"start-plan"}, {"issue", "new", "First work", "--dry-run"}} {
		out, stderr, err := executeSDLCTestCommand(args...)
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, stderr)
		}
		text := out + stderr
		if !strings.Contains(text, "sdlc issue init") || strings.Contains(text, "sdlc issue migrate") {
			t.Fatalf("empty local inventory did not hint init: %v\n%s", args, text)
		}
	}
	writeRepoFile(t, r.root, "workshop/history/issues/000001-old.md", "old\n")
	var out strings.Builder
	runStartPlan(context.Background(), &out, 0)
	if !strings.Contains(out.String(), "sdlc issue migrate") {
		t.Fatalf("existing history did not hint migrate: %s", out.String())
	}
}

func TestIssueInitPrerequisitesHaveNoPublicationEffects(t *testing.T) {
	for _, kind := range []string{"dirty tracked", "untracked", "non-main", "no remote", "unpublished main", "local ahead", "diverged", "unborn"} {
		t.Run(kind, func(t *testing.T) {
			r := emptyInitRepo(t)
			switch kind {
			case "dirty tracked":
				writeRepoFile(t, r.root, "workshop/issues/.gitkeep", "retain edits\n")
			case "untracked":
				writeRepoFile(t, r.root, "notes.txt", "retain notes\n")
			case "non-main":
				r.git("switch", "-qc", "feature")
			case "no remote":
				r.git("remote", "remove", "origin")
			case "unpublished main":
				testfix.Git(t, r.origin, "update-ref", "refs/heads/retained", r.originMain())
				testfix.Git(t, r.origin, "update-ref", "-d", "refs/heads/main")
			case "local ahead", "diverged":
				writeRepoFile(t, r.root, "new.txt", "local work\n")
				r.git("add", ".")
				r.git("commit", "-qm", "local ahead")
				if kind == "diverged" {
					clone := filepath.Join(t.TempDir(), "other")
					testfix.Git(t, "", "clone", "-q", r.origin, clone)
					writeRepoFile(t, clone, "other.txt", "other work\n")
					testfix.Git(t, clone, "add", ".")
					testfix.Git(t, clone, "commit", "-qm", "remote ahead")
					testfix.Git(t, clone, "push", "-q")
				}
			case "unborn":
				r.git("update-ref", "-d", "refs/heads/main")
			}
			beforeRefs := testfix.Capture(t, r.origin, "show-ref")
			beforeStatus := r.git("status", "--porcelain", "--untracked-files=all")
			if _, _, err := executeSDLCTestCommand("issue", "init"); err == nil {
				t.Fatal("missing prerequisite accepted")
			}
			if got := testfix.Capture(t, r.origin, "show-ref"); got != beforeRefs {
				t.Fatalf("refused setup changed remote refs: %q -> %q", beforeRefs, got)
			}
			if got := r.git("status", "--porcelain", "--untracked-files=all"); got != beforeStatus {
				t.Fatal("refused setup changed working bytes or index status")
			}
		})
	}
}

func TestIssueInitAfterLaterIssuePreservesCheckoutAndCard(t *testing.T) {
	r := emptyInitRepo(t)
	if _, _, err := executeSDLCTestCommand("issue", "init"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := executeSDLCTestCommand("issue", "new", "Later work", "--no-claim"); err != nil {
		t.Fatal(err)
	}
	head := r.git("rev-parse", "HEAD")
	remote := r.originMain()
	trackerTip := testfix.Capture(t, r.origin, "rev-parse", "issue-tracker")
	if _, _, err := executeSDLCTestCommand("issue", "init"); err != nil {
		t.Fatal(err)
	}
	if r.git("rev-parse", "HEAD") != head || r.originMain() != remote || testfix.Capture(t, r.origin, "rev-parse", "issue-tracker") != trackerTip {
		t.Fatal("completed init changed later issue or checkout")
	}
}

func TestLegacySetupHintSuppressesUnknownBrainAndDeclined(t *testing.T) {
	for _, kind := range []string{"brain", "declined", "unreadable"} {
		t.Run(kind, func(t *testing.T) {
			r := emptyInitRepo(t)
			switch kind {
			case "brain":
				writeRepoFile(t, r.root, ".brain/config.md", "brain\n")
			case "declined":
				r.git("rm", "workshop/issues/.gitkeep")
				if err := os.MkdirAll(filepath.Join(r.root, "workshop/history"), 0755); err != nil {
					t.Fatal(err)
				}
			case "unreadable":
				r.git("rm", "workshop/issues/.gitkeep")
				// A regular file at the discovery root is an unknown inventory,
				// portable even when tests run with permission to read mode 000.
				writeRepoFile(t, r.root, "workshop/issues", "not a directory\n")
			}
			var out strings.Builder
			legacySetupAdvice(&out, "workshop/issues", "workshop/history")
			if out.Len() != 0 {
				t.Fatalf("unjustified setup advice: %s", out.String())
			}
		})
	}
}

func TestIssueInitInventoryProcessGrowth(t *testing.T) {
	count := func(files, branches int) int {
		t.Helper()
		r := emptyInitRepo(t)
		for i := 0; i < files; i++ {
			writeRepoFile(t, r.root, fmt.Sprintf("source/file-%d.txt", i), "unrelated\n")
		}
		r.git("add", ".")
		r.git("commit", "--allow-empty", "-qm", "unrelated source")
		r.git("push", "-q")
		tree, parent := r.git("rev-parse", "HEAD^{tree}"), r.git("rev-parse", "HEAD")
		for i := 0; i < branches; i++ {
			// Distinct tips exercise per-tip metadata reads rather than letting
			// identical-commit deduplication hide their process growth.
			tip := r.git("commit-tree", tree, "-p", parent, "-m", fmt.Sprintf("unrelated tip %d", i))
			r.git("branch", fmt.Sprintf("unrelated-%d", i), tip)
		}
		trace := filepath.Join(t.TempDir(), "git.trace")
		t.Setenv("GIT_TRACE", trace)
		if _, err := runInitialization(context.Background(), &gitInitializationEffects{}); err != nil {
			t.Fatal(err)
		}
		t.Setenv("GIT_TRACE", "")
		raw, err := os.ReadFile(trace)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Count(string(raw), "built-in: git ")
	}
	files8, files16 := count(8, 0), count(16, 0)
	branches8, branches16 := count(0, 8), count(0, 16)
	t.Logf("Git process counts: files 8/16=%d/%d, branches 8/16=%d/%d", files8, files16, branches8, branches16)
	if files8 == 0 || files16 != files8 {
		t.Fatalf("unrelated file growth spawned Git processes: %d -> %d", files8, files16)
	}
	if branches8 == 0 || branches16 > 2*branches8 {
		t.Fatalf("branch scan grew worse than linearly: %d -> %d", branches8, branches16)
	}
}
