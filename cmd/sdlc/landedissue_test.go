package main

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// landedFakeGH holds GitHub's PR records by head branch (#483), behind the
// same landingGH seam production reads.
type landedFakeGH struct {
	stubGH
	prs map[string][]landingPR
	err error
}

func (g landedFakeGH) LandingPRs(_ context.Context, _ string, branch string) ([]landingPR, error) {
	return g.prs[branch], g.err
}
func (g landedFakeGH) LandingMerge(context.Context, string, int, string) error {
	return errors.New("landedFakeGH: no merges")
}

func TestLandedPR(t *testing.T) {
	merged := landingPR{Number: 406, State: "MERGED"}
	for _, tc := range []struct {
		name string
		prs  []landingPR
		want bool
	}{
		{"none", nil, false},
		{"merged", []landingPR{merged}, true},
		{"closed only", []landingPR{{Number: 1, State: "CLOSED"}}, false},
		{"merged and open", []landingPR{merged, {Number: 407, State: "OPEN"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, got := landedPR(tc.prs); got != tc.want {
				t.Fatalf("landedPR = %v, want %v", got, tc.want)
			}
		})
	}
}

// landedS09 builds pair#622's shape on issue #9 (#483): the canonical branch
// was squash-merged as PR 406 and deleted, and a follow-up branch cut from
// main holds only the live-acceptance evidence. It returns the merged head.
func landedS09(t *testing.T) (*trackerRepo, map[string]string, string) {
	t.Helper()
	r, paths, _ := startedHere(t)
	head := r.git("rev-parse", "HEAD")
	squash := r.git("commit-tree", r.git("rev-parse", "HEAD^{tree}"), "-p", r.originMain(), "-m", "#9: s09 (#406)")
	r.git("push", "-q", "origin", squash+":refs/heads/main")
	r.git("fetch", "-q", "origin")
	r.git("switch", "-q", "-c", "000009-live-acceptance", "origin/main")
	r.git("commit", "-q", "--allow-empty", "-m", "#9: live acceptance attested by the operator")
	r.git("branch", "-q", "-D", s09Branch)
	// The configured URL is GitHub; Git's local rewrite keeps it offline.
	r.git("config", "remote.origin.url", "https://github.com/test/repo.git")
	r.git("config", "url."+r.origin+".insteadOf", "https://github.com/test/repo.git")
	swapLandedGH(t, landedFakeGH{prs: map[string][]landingPR{
		s09Branch: {{Number: 406, State: "MERGED", HeadRef: s09Branch, HeadOID: head}},
	}})
	return r, paths, head
}

func swapLandedGH(t *testing.T, gh ghCaller) {
	t.Helper()
	previous := ghClient
	ghClient = gh
	t.Cleanup(func() { ghClient = previous })
}

func closeCheck(t *testing.T, paths map[string]string) error {
	t.Helper()
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return requireCloseIssueBranch(env, "000009", paths["000009"])
}

func TestCloseAdmitsFollowUpBranchOfLandedIssue(t *testing.T) {
	_, paths, _ := landedS09(t)
	if err := closeCheck(t, paths); err != nil {
		t.Fatalf("pair#622 shape: close from the follow-up branch refused: %v", err)
	}
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if note := closeBranchRecoveryNote(env, paths["000009"]); !strings.Contains(note, "follow-up branch 000009-live-acceptance") {
		t.Fatalf("close note misdescribes the follow-up branch: %q", note)
	}
}

func TestCloseRefusesFollowUpCarryingAnotherIssue(t *testing.T) {
	r, paths, _ := landedS09(t)
	r.git("switch", "-q", "-c", "000010-other", "origin/main")
	r.git("commit", "-q", "--allow-empty", "-m", "#10: unlanded other work")
	r.git("switch", "-q", "-c", "000009-stacked")
	if err := closeCheck(t, paths); err == nil || !strings.Contains(err.Error(), "000010-other") {
		t.Fatalf("a follow-up carrying #10's unlanded work must refuse: %v", err)
	}
}

func TestCloseRefusalNamesTheLandedRecovery(t *testing.T) {
	r, paths, _ := landedS09(t)
	r.git("switch", "-q", "-c", "acceptance", "origin/main")
	err := closeCheck(t, paths)
	if err == nil || !strings.Contains(err.Error(), "already landed (PR #406)") || !strings.Contains(err.Error(), "sdlc issue new '<title>' --deps 9") {
		t.Fatalf("an unprefixed branch must refuse with the landed recovery: %v", err)
	}
}

func TestCloseKeepsTodaysRefusalWhenNotLanded(t *testing.T) {
	for name, gh := range map[string]landedFakeGH{
		"open PR":  {prs: map[string][]landingPR{s09Branch: {{Number: 407, State: "OPEN", HeadRef: s09Branch}}}},
		"gh error": {err: errors.New("gh unreachable")},
	} {
		t.Run(name, func(t *testing.T) {
			_, paths, _ := landedS09(t)
			swapLandedGH(t, gh)
			err := closeCheck(t, paths)
			if err == nil || strings.Contains(err.Error(), "already landed") {
				t.Fatalf("not landed must keep today's refusal: %v", err)
			}
		})
	}
}

func TestLocalCanonicalAheadOfMergedHeadIsNotLanded(t *testing.T) {
	r, _, head := landedS09(t)
	r.git("switch", "-q", "-c", s09Branch, head)
	r.git("commit", "-q", "--allow-empty", "-m", "#9: new work after the landing")
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, landed := issueBranchLanded(env, s09Branch); landed {
		t.Fatal("unlanded local commits on the canonical branch read as landed")
	}
	r.git("reset", "-q", "--hard", head)
	if _, landed := issueBranchLanded(env, s09Branch); !landed {
		t.Fatal("a local canonical branch at the merged head reads as unlanded")
	}
}

func TestStartPlanAndChangeCodeRefuseLandedIssue(t *testing.T) {
	r, paths, _ := landedS09(t)
	var out bytes.Buffer
	err := startPlanBranch(context.Background(), &out, 9)
	if err == nil || !strings.Contains(err.Error(), "already landed (PR #406)") || !strings.Contains(err.Error(), "--deps 9") {
		t.Fatalf("start-plan on a landed issue: %v", err)
	}
	stem := strings.TrimSuffix(filepath.Base(paths["000009"]), ".md")
	_, err = refreshChangeCodeMirror(&changeCodeFlags{Issue: 9}, stem, filepath.Join(r.root, "workshop/issues", filepath.Base(paths["000009"])))
	if err == nil || !strings.Contains(err.Error(), "already landed (PR #406)") || !strings.Contains(err.Error(), "--deps 9") {
		t.Fatalf("change-code on a landed issue: %v", err)
	}
}
