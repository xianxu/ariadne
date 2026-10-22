package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestNoncanonicalSideNamingPolicy(t *testing.T) {
	for _, tc := range []struct{ name, want string }{
		{"i/000768-x.md.side", "i/000768-x.side"},
		{"p/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md", "p/000768-darwin-pty-final-output-plan.side"},
		{"p/000768-review.md.side/deep/nested/draft", "p/000768-review.side"},
		{"p/000768-plan.side/sub/a.md.side/q.md", ""},
		{"p/000768-plan.side/q.md", ""},
		{"p/000757-old.md.side/draft", ""},
		{"h/plans/000768-old.md.side/draft", ""},
		{"p/0007680-other.md.side/draft", ""},
		{"other/000768-x.md.side/draft", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateArchiveSideNames([]string{tc.name}, []string{"i/000768-x.md"}, "i", "p")
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), "rename it to "+tc.want)) {
				t.Fatalf("got %v; expected canonical %q", err, tc.want)
			}
		})
	}
}

func TestNoncanonicalSideCloseReadsRootsOnce(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			root := t.TempDir()
			i, p := filepath.Join(root, "custom-issues"), filepath.Join(root, "custom-plans")
			for j := 0; j < n; j++ {
				mkArtifact(t, filepath.Join(i, fmt.Sprintf("%06d-x.md", j+1)), "issue")
				mkArtifact(t, filepath.Join(p, fmt.Sprintf("%06d-x.side/q.md", j+1)), "side")
			}
			calls, entries := 0, 0
			read := func(dir string) ([]os.DirEntry, error) {
				calls++
				x, err := os.ReadDir(dir)
				entries += len(x)
				return x, err
			}
			issue := filepath.Join(i, "000768-x.md")
			if err := checkCloseSideNames(issue, p, read); err != nil {
				t.Fatal(err)
			}
			if calls != 2 || entries != 2*n {
				t.Fatalf("reads=%d entries=%d; want 2 reads and %d entries", calls, entries, 2*n)
			}
			mkArtifact(t, filepath.Join(p, "000768-x.md.side/q.md"), "bad")
			if err := checkCloseSideNames(issue, p, os.ReadDir); err == nil || !strings.Contains(err.Error(), filepath.Join(p, "000768-x.side")) {
				t.Fatalf("absolute configured root: %v", err)
			}
			if err := checkCloseSideNames(issue, p, func(string) ([]os.DirEntry, error) { return nil, os.ErrPermission }); !errors.Is(err, os.ErrPermission) {
				t.Fatalf("unreadable root accepted: %v", err)
			}
		})
	}
}

func TestNoncanonicalSideCloseRefusesBeforeReview(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		t.Run(map[bool]string{false: "untracked", true: "tracked"}[tracked], func(t *testing.T) {
			issues := closeRepo(t, 768)
			bad := "workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md"
			mkArtifact(t, bad, "retained draft\n")
			if tracked {
				testfix.Git(t, ".", "add", ".")
				testfix.Git(t, ".", "commit", "-qm", "#768: draft")
			}
			before := gitx.Capture("status", "--porcelain=v1", "--untracked-files=all")
			calls, _ := stubJudge(t, "Review-Verdict: SHIP"+emptyBoundaryFindings)
			var out bytes.Buffer
			msg, died := expectDie(t, func() {
				_ = runCloseWithReview(io.Discard, &out, &closeFlags{Issue: 768, Actual: "1", Verified: "fixture", NoAtlas: true, IssuesDir: issues})
			})
			if !died || !strings.Contains(msg, "000768-darwin-pty-final-output-plan.side") || *calls != 0 {
				t.Fatalf("expected canonical rename before review: calls=%d output=%s", *calls, msg+out.String())
			}
			if gitx.Capture("status", "--porcelain=v1", "--untracked-files=all") != before {
				t.Fatal("refusal changed fixture")
			}
		})
	}
}

func TestNoncanonicalSideLandingPublishRefuses(t *testing.T) {
	git, base := publishRepo(t)
	git("switch", "-c", "issue")
	writeIssueStatus(t, git, 69, "codecomplete", "#69 close")
	bad := "workshop/plans/000069-x-plan.md.side/sub/draft.md"
	mkArtifact(t, bad, "retained\n")
	git("add", ".")
	git("commit", "-qm", "side draft")
	pr := landingPR{HeadOID: gitx.Capture("rev-parse", "HEAD"), BaseOID: base}
	err := runLandingPublishGate(context.Background(), pr, "workshop/issues", "workshop/plans", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "000069-x-plan.side") {
		t.Fatalf("malformed side bundle passed pinned-head preflight: %v", err)
	}
	// A mutable checkout cannot hide the bad folder at the selected commit.
	if err := os.Remove(bad); err != nil {
		t.Fatal(err)
	}
	gates := productionCarGates(execGitRunner{}, nil, landingTarget{}, &mergeFlags{IssuesDir: "workshop/issues", PlansDir: "workshop/plans"})
	if err := gates.publish(context.Background(), pr); err == nil || !strings.Contains(err.Error(), "000069-x-plan.side") {
		t.Fatalf("train readiness lost pinned naming check: %v", err)
	}
	// Canonical rename with its bytes retained makes the ordinary gate pass.
	mkArtifact(t, "workshop/plans/000069-x-plan.side/sub/draft.md", "retained\n")
	git("add", ".")
	git("commit", "-qm", "canonical side folder")
	pr.HeadOID = gitx.Capture("rev-parse", "HEAD")
	if err := gates.publish(context.Background(), pr); err != nil {
		t.Fatalf("canonical control refused: %v", err)
	}
}

func TestNoncanonicalSideArchiveRefusesBeforePublication(t *testing.T) {
	root, origin, pr := laFixture(t)
	bad := "workshop/plans/000246-landing-plan.md.side/shared-owner-draft.md"
	laWrite(t, root, bad, "retained draft\n")
	pr.HeadOID = laCommit(t, root, "side draft")
	pr.MergeOID, pr.LandedOID = pr.HeadOID, pr.HeadOID
	testfix.Git(t, root, "push", "-q", "origin", "HEAD:main")
	before := laGit(t, origin, "rev-parse", "main")
	if err := laArchive(root, pr); err == nil || !strings.Contains(err.Error(), "000246-landing-plan.side") {
		t.Fatalf("parent-only archive accepted: %v", err)
	}
	if laGit(t, origin, "rev-parse", "main") != before {
		t.Fatal("refusal published an archive")
	}
	if b, err := os.ReadFile(filepath.Join(root, bad)); err != nil || string(b) != "retained draft\n" {
		t.Fatalf("source changed: %q %v", b, err)
	}
}
