package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestLandingDuplicateGateConfiguredTarget(t *testing.T) {
	root := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	// Only upstream exists; origin/main cannot serve as an accidental fallback.
	remote := filepath.Join(t.TempDir(), "upstream.git")
	testfix.Git(t, root, "init", "--bare", "-q", "-b", "main", remote)
	testfix.Git(t, root, "remote", "add", "upstream", remote)
	testfix.Git(t, root, "switch", "-c", "remote-main")
	mkArtifact(t, filepath.Join(root, "workshop/issues/000001-remote.md"), "remote reservation\n")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "remote claimant")
	testfix.Git(t, root, "push", "upstream", "HEAD:main")
	testfix.Git(t, root, "fetch", "--no-tags", "--refmap=", "upstream", "main")
	main := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "FETCH_HEAD"))
	testfix.Git(t, root, "switch", "-c", "issue", base)
	mkArtifact(t, filepath.Join(root, "workshop/issues/000001-local.md"), "local reservation\n")
	testfix.Git(t, root, "add", ".")
	testfix.Git(t, root, "commit", "-qm", "local claimant")
	if err := runLandingDuplicateGate(main, "workshop/issues", "workshop/history", execGitRunner{}); err == nil || !strings.Contains(err.Error(), "000001") {
		t.Fatalf("missed configured target collision: %v", err)
	}
	if err := runLandingDuplicateGate(base, "workshop/issues", "workshop/history", execGitRunner{}); err != nil {
		t.Fatalf("safe pinned target: %v", err)
	}
	if err := runLandingDuplicateGate(main, "workshop/issues", "workshop/history", &lsTreeFailRunner{}); err == nil {
		t.Fatal("silently skipped failed tree read")
	}
	if err := runLandingDuplicateGate(strings.Repeat("f", 40), "workshop/issues", "workshop/history", execGitRunner{}); err == nil {
		t.Fatal("accepted missing main commit")
	}
	if err := runLandingDuplicateGate(main, "../outside", "workshop/history", execGitRunner{}); err == nil {
		t.Fatal("silently skipped invalid configured root")
	}
}

func TestLandingPublishGateIndependentlyPublishedBody(t *testing.T) {
	for _, delta := range []string{"code", "docs", "none"} {
		t.Run(delta, func(t *testing.T) {
			git, base := publishRepo(t)
			git("switch", "-c", "issue")
			writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
			closeHead := gitx.Capture("rev-parse", "HEAD")
			body, err := os.ReadFile(issuePathFor(69))
			if err != nil {
				t.Fatal(err)
			}
			git("switch", "-c", "independent-publication", base)
			mkArtifact(t, issuePathFor(69), string(body))
			git("add", ".")
			git("commit", "-qm", "independent issue body publication")
			prBase := gitx.Capture("rev-parse", "HEAD")
			git("switch", "issue")
			if delta == "code" {
				commitCode(t, git, "late.go")
			}
			if delta == "docs" {
				mkArtifact(t, "atlas/postclose.md", "bookkeeping\n")
				git("add", ".")
				git("commit", "-qm", "post-close docs")
			}
			pr := landingPR{HeadOID: gitx.Capture("rev-parse", "HEAD"), BaseOID: prBase}
			// The prior body-diff enumeration demonstrably overlooks this owned close.
			omitted, err := mergedCodecompleteIssues(prBase, "workshop/issues")
			if err != nil || len(omitted) != 0 {
				t.Fatalf("fixture not independently published: %v %v", omitted, err)
			}
			err = runLandingPublishGate(pr, "workshop/issues", io.Discard)
			if delta == "code" {
				if err == nil || !strings.Contains(err.Error(), "landed after `sdlc close`") {
					t.Fatalf("post-close code escaped review: %v", err)
				}
			} else if err != nil {
				t.Fatalf("safe %s refused: %v (close %s)", delta, err, closeHead)
			}
			pr.HeadOID = base
			if err := runLandingPublishGate(pr, "workshop/issues", io.Discard); err == nil {
				t.Fatal("gate ran on different checkout head")
			}
		})
	}
}
