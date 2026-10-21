package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Opt-in read-only proof against captured live PR metadata and an isolated
// clone of its repository. The clone must include main, issue-tracker and the
// PR heads; origin names the real repository. This never merges, archives or
// changes completion cards. Tracker fetches update only the disposable clone.
// Set SDLC_LIVE_ARCHIVE_FIXTURE to a JSON file containing Root and Cases, each
// case containing Issue (integer) and PR (landingPR fields captured via gh).
func TestLiveLandingArchiveConformance(t *testing.T) {
	filename := os.Getenv("SDLC_LIVE_ARCHIVE_FIXTURE")
	if filename == "" {
		t.Skip("set SDLC_LIVE_ARCHIVE_FIXTURE for read-only archive conformance")
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Root  string
		Cases []struct {
			Issue int
			PR    landingPR
		}
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(fixture.Root) || len(fixture.Cases) == 0 {
		t.Fatal("absolute isolated clone Root and nonempty Cases required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	tip := laGit(t, fixture.Root, "rev-parse", "refs/remotes/origin/main")
	head, index := laGit(t, fixture.Root, "rev-parse", "HEAD"), laGit(t, fixture.Root, "ls-files", "--stage")
	for _, tc := range fixture.Cases {
		pr := tc.PR
		landed, ok := landedCommit(pr, func(oid string) bool {
			_, err := archiveReadContext(ctx, fixture.Root, "merge-base", "--is-ancestor", oid, tip)
			return err == nil
		})
		if !ok {
			t.Fatalf("#%d: integration not on pinned main", tc.Issue)
		}
		pr.LandedOID = landed
		dirs, err := resolveArchiveRoots(fixture.Root, "workshop/issues", "workshop/plans", "workshop/history")
		if err != nil {
			t.Fatal(err)
		}
		selected, err := selectArchiveLandingIssues(ctx, fixture.Root, pr, dirs)
		if err != nil || len(selected) != 1 {
			t.Fatalf("#%d: owned completions=%d err=%v", tc.Issue, len(selected), err)
		}
		if !strings.HasPrefix(filepath.Base(selected[0].path), fmt.Sprintf("%06d-", tc.Issue)) {
			t.Fatalf("wrong owned issue: %s", selected[0].path)
		}
		done, err := landingArchiveComplete(ctx, fixture.Root, tip, pr.Repo, pr, "workshop/issues", "workshop/plans", "workshop/history")
		if err != nil || !done {
			t.Fatalf("#%d PR #%d: complete=%v err=%v", tc.Issue, pr.Number, done, err)
		}
		t.Logf("#%d PR #%d: accepted exact durable archive at main %s", tc.Issue, pr.Number, tip)
	}
	if laGit(t, fixture.Root, "rev-parse", "HEAD") != head || laGit(t, fixture.Root, "ls-files", "--stage") != index {
		t.Fatal("proof changed clone HEAD or index")
	}
}
