package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestPreparedReviewReadSetRejectsDirectory(t *testing.T) {
	if _, err := captureCloseReviewArtifact(t.TempDir()); err == nil {
		t.Fatal("directory read failure was treated as an absent artifact")
	}
}

func TestPreparedReviewReadSetBranchAndLedgers(t *testing.T) {
	for _, change := range []string{"branch", "boundary-ledger", "plan-seed"} {
		t.Run(change, func(t *testing.T) {
			dir := closeRepo(t, 69)
			plans := filepath.Join("workshop", "plans")
			if err := os.MkdirAll(plans, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "000069-x.md")
			head := strings.TrimSpace(git(t, "", "rev-parse", "HEAD"))
			snap, err := captureCloseReviewSnapshot(closeResult{issuePath: path, issueText: readIssue(t, dir)}, head, "", plans)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "branch":
				git(t, "", "switch", "-c", "same-head-different-branch")
			case "boundary-ledger":
				if err := os.WriteFile(boundaryGatePath(plans, filepath.Base(path)), []byte("concurrent boundary round"), 0644); err != nil {
					t.Fatal(err)
				}
			case "plan-seed":
				if err := os.WriteFile(planGatePath(plans, filepath.Base(path)), []byte("concurrent plan findings"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := snap.validate(); err == nil {
				t.Fatalf("%s changed but review stayed valid", change)
			}
		})
	}
}

func TestPreparedReviewEveryVerdictValidatesBeforeWrites(t *testing.T) {
	for _, verdict := range []judge.Verdict{judge.VerdictShip, judge.VerdictRework, judge.VerdictUnknown, judge.VerdictNotRun} {
		t.Run(string(verdict), func(t *testing.T) {
			issues := closeRepo(t, 69)
			plans := filepath.Join("workshop", "plans")
			if err := os.MkdirAll(plans, 0755); err != nil {
				t.Fatal(err)
			}
			before := readIssue(t, issues)
			p := boundaryReviewParams{IssueNum: 69, IssuesDir: issues, PlansDir: plans, Head: strings.TrimSpace(git(t, "", "rev-parse", "HEAD"))}
			f := closeFlagsFor(issues)
			f.Force = true
			called := false
			err := finalizeBoundaryReview(io.Discard, io.Discard, f, closeResult{}, reviewResult{Verdict: verdict, Output: "review output", Agent: "test"}, p, func() (string, error) { called = true; return "", errors.New("concurrent ledger changed") })
			if !called || err == nil || !strings.Contains(err.Error(), "stale") {
				t.Errorf("stale verdict: validated=%v err=%v", called, err)
			}
			entries, readErr := os.ReadDir(plans)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if len(entries) > 0 {
				t.Errorf("stale verdict persisted artifacts: %v", entries)
			}
			if after := readIssue(t, issues); after != before {
				t.Error("stale verdict changed issue")
			}
		})
	}
}

func TestPreparedReviewInterruptedCannotPersistEvenForced(t *testing.T) {
	for _, kind := range []string{"context", "dispatch-error"} {
		t.Run(kind, func(t *testing.T) {
			issues := closeRepo(t, 69)
			plans := filepath.Join("workshop", "plans")
			if err := os.MkdirAll(plans, 0755); err != nil {
				t.Fatal(err)
			}
			p := boundaryReviewParams{IssueNum: 69, IssuesDir: issues, PlansDir: plans}
			review := reviewResult{Verdict: judge.VerdictShip, Output: "SHIP", Agent: "test"}
			if kind == "context" {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				p.Context = ctx
			} else {
				review.DispatchError = errors.New("reviewer could not start")
			}
			f := closeFlagsFor(issues)
			f.Force = true
			if err := finalizeBoundaryReview(io.Discard, io.Discard, f, closeResult{}, review, p, nil); err == nil {
				t.Fatal("forced interrupted review finalized")
			}
			entries, err := os.ReadDir(plans)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) > 0 {
				t.Fatalf("interrupted review persisted: %v", entries)
			}
		})
	}
}

func TestPreparedReviewExactAndArtifactDecisions(t *testing.T) {
	issues := closeRepo(t, 69)
	path := filepath.Join(issues, "000069-x.md")
	prepared, err := capturePreparedReview("", []reviewArtifact{{path: path, present: true, text: readIssue(t, issues)}}, filepath.Join("workshop", "plans", "not-yet-created.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := prepared.validateExact(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("unrelated.md", []byte("documentation"), 0600); err != nil {
		t.Fatal(err)
	}
	git(t, "", "add", "--", "unrelated.md")
	git(t, "", "commit", "-qm", "unrelated documentation")
	if err := prepared.validateIdentityAndArtifacts(); err != nil {
		t.Fatalf("artifact comparison rejected unrelated docs: %v", err)
	}
	if err := prepared.validateExact(); err == nil {
		t.Fatal("exact planning review accepted changed HEAD")
	}
}

func TestPreparedReviewCancellationDuringValidationCannotPersist(t *testing.T) {
	issues := closeRepo(t, 69)
	plans := filepath.Join("workshop", "plans")
	if err := os.MkdirAll(plans, 0755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := boundaryReviewParams{Context: ctx, IssueNum: 69, IssuesDir: issues, PlansDir: plans}
	f := closeFlagsFor(issues)
	f.Force = true
	err := finalizeBoundaryReview(io.Discard, io.Discard, f, closeResult{}, reviewResult{Verdict: judge.VerdictShip, Output: "SHIP"}, p, func() (string, error) { cancel(); return "", nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("validation-time cancellation lost: %v", err)
	}
	entries, err := os.ReadDir(plans)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 0 {
		t.Fatalf("cancelled validation persisted artifacts: %v", entries)
	}
}

func TestPreparedReviewPureComparison(t *testing.T) {
	baseline := preparedReview{root: "/repo", gitDir: "/repo/.git", branch: "refs/heads/work", head: "abc", artifacts: []reviewArtifact{{path: "/repo/issue.md", present: true, text: "prepared body"}, {path: "/repo/plan.md"}}}
	cases := []struct {
		name   string
		change func(*preparedReview)
		valid  bool
	}{
		{"unchanged", func(*preparedReview) {}, true},
		{"head policy delegated", func(s *preparedReview) { s.head = "def" }, true},
		{"repository", func(s *preparedReview) { s.root = "/other" }, false},
		{"worktree", func(s *preparedReview) { s.gitDir = "/repo/.git/worktrees/other" }, false},
		{"branch", func(s *preparedReview) { s.branch = "refs/heads/other" }, false},
		{"artifact identity", func(s *preparedReview) { s.artifacts[0].path = "/other/issue.md" }, false},
		{"disappeared", func(s *preparedReview) { s.artifacts[0].present = false; s.artifacts[0].text = "" }, false},
		{"appeared empty", func(s *preparedReview) { s.artifacts[1].present = true }, false},
		{"content", func(s *preparedReview) { s.artifacts[0].text = "changed" }, false},
		{"set", func(s *preparedReview) { s.artifacts = s.artifacts[:1] }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			current := baseline
			current.artifacts = append([]reviewArtifact(nil), baseline.artifacts...)
			tc.change(&current)
			if err := comparePreparedReview(baseline, current); (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
}

func TestPreparedReviewDispatchPreflightFailureCannotPersist(t *testing.T) {
	issues := closeRepo(t, 69)
	plans := filepath.Join("workshop", "plans")
	if err := os.MkdirAll(plans, 0755); err != nil {
		t.Fatal(err)
	}
	p := boundaryReviewParams{IssueNum: 69, IssuesDir: issues, PlansDir: plans, Head: strings.TrimSpace(git(t, "", "rev-parse", "HEAD"))}
	r := closeResult{issuePath: filepath.Join(issues, "000069-x.md"), issueText: readIssue(t, issues)}
	// An absent review-window base is a preflight error, not a completed round.
	if err := reviewThenFinalize(io.Discard, io.Discard, closeFlagsFor(issues), r, p); err == nil {
		t.Fatal("missing review window accepted")
	}
	entries, err := os.ReadDir(plans)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > 0 {
		t.Fatalf("review that never dispatched persisted artifacts: %v", entries)
	}
}
