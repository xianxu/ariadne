package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

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
	got, err := observeMergeIntents(r, root, func(mergeIntentReport) bool { return true })
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
}
