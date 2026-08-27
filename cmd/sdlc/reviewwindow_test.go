package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

type fakeBoundaryGit struct {
	root    string
	refs    map[string]string
	failRef string
	calls   [][]string
}

func (f *fakeBoundaryGit) RunGit(args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	if strings.Join(args, " ") == "rev-parse --show-toplevel" {
		if f.root == "" {
			return nil, errors.New("not a repository")
		}
		return []byte(f.root + "\n"), nil
	}
	if len(args) == 3 && args[0] == "rev-parse" && args[1] == "--verify" {
		ref := strings.TrimSuffix(args[2], "^{commit}")
		if ref == f.failRef {
			return nil, errors.New("unknown revision")
		}
		if sha := f.refs[ref]; sha != "" {
			return []byte(sha + "\n"), nil
		}
	}
	return nil, errors.New("unexpected git argv: " + strings.Join(args, " "))
}

func TestResolveBoundaryReviewManifest_PinsRefsThroughRunner(t *testing.T) {
	root := t.TempDir()
	fake := &fakeBoundaryGit{
		root: root,
		refs: map[string]string{"main~2": testReviewBaseSHA, "topic": testReviewHeadSHA},
	}

	m, err := resolveBoundaryReviewManifest(fake, boundaryReviewManifestRequest{
		BaseRef: "main~2", HeadRef: "topic", IssuesDir: "custom/issues", HistoryDir: "custom/history",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.RepoRoot != root || m.BaseSHA != testReviewBaseSHA || m.HeadSHA != testReviewHeadSHA || m.WorkingTree {
		t.Fatalf("resolved manifest = %+v", m)
	}
	wantCalls := [][]string{
		{"rev-parse", "--show-toplevel"},
		{"rev-parse", "--verify", "main~2^{commit}"},
		{"rev-parse", "--verify", "topic^{commit}"},
	}
	if got := formatGitCalls(fake.calls); got != formatGitCalls(wantCalls) {
		t.Fatalf("git calls:\n%s\nwant:\n%s", got, formatGitCalls(wantCalls))
	}
}

func TestResolveBoundaryReviewManifest_OmittedHeadPinsAmbientHead(t *testing.T) {
	root := t.TempDir()
	fake := &fakeBoundaryGit{
		root: root,
		refs: map[string]string{"base": testReviewBaseSHA, "HEAD": testReviewHeadSHA},
	}

	m, err := resolveBoundaryReviewManifest(fake, boundaryReviewManifestRequest{
		BaseRef: "base", IssuesDir: "workshop/issues", HistoryDir: "workshop/history",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !m.WorkingTree || m.HeadSHA != "" || m.AmbientHeadSHA != testReviewHeadSHA {
		t.Fatalf("working-tree manifest = %+v", m)
	}
}

func TestResolveBoundaryReviewManifest_ValidatesIssueAndFindsOptionalPlan(t *testing.T) {
	root := t.TempDir()
	issues := "custom/issues"
	plans := "custom/plans"
	if err := os.MkdirAll(filepath.Join(root, issues), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, plans), 0o755); err != nil {
		t.Fatal(err)
	}
	issue := filepath.Join(issues, "000162-review-window.md")
	plan := filepath.Join(plans, "000162-review-window-plan.md")
	if err := os.WriteFile(filepath.Join(root, issue), []byte("# issue\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, plan), []byte("# plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fake := &fakeBoundaryGit{root: root, refs: map[string]string{"base": testReviewBaseSHA, "head": testReviewHeadSHA}}

	m, err := resolveBoundaryReviewManifest(fake, boundaryReviewManifestRequest{
		BaseRef: "base", HeadRef: "head", IssuesDir: issues, HistoryDir: "custom/history", IssueNum: 162, PlansDir: plans,
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.IssueFile != issue || m.PlanFile != plan {
		t.Fatalf("tracker paths = issue %q plan %q, want %q and %q", m.IssueFile, m.PlanFile, issue, plan)
	}
}

func TestResolveBoundaryReviewManifest_FailsClosedOnGitOrTrackerErrors(t *testing.T) {
	root := t.TempDir()
	baseFake := func() *fakeBoundaryGit {
		return &fakeBoundaryGit{root: root, refs: map[string]string{"base": testReviewBaseSHA, "head": testReviewHeadSHA}}
	}

	tests := map[string]struct {
		fake *fakeBoundaryGit
		req  boundaryReviewManifestRequest
	}{
		"repository": {fake: &fakeBoundaryGit{}, req: boundaryReviewManifestRequest{BaseRef: "base", HeadRef: "head", IssuesDir: "i", HistoryDir: "h"}},
		"base ref":   {fake: func() *fakeBoundaryGit { f := baseFake(); f.failRef = "bad"; return f }(), req: boundaryReviewManifestRequest{BaseRef: "bad", HeadRef: "head", IssuesDir: "i", HistoryDir: "h"}},
		"issue path": {fake: baseFake(), req: boundaryReviewManifestRequest{BaseRef: "base", HeadRef: "head", IssuesDir: "missing", HistoryDir: "h", IssueNum: 162}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := resolveBoundaryReviewManifest(tc.fake, tc.req); err == nil {
				t.Fatal("resolver accepted unavailable review input")
			}
		})
	}
}

func TestResolveBoundaryReviewManifest_ConformsToRealGit(t *testing.T) {
	dir := testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "HEAD"))
	if err := os.WriteFile("tracked.txt", []byte("reviewed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, dir, "add", "tracked.txt")
	testfix.Git(t, dir, "commit", "-q", "-m", "reviewed")
	head := strings.TrimSpace(testfix.Capture(t, dir, "rev-parse", "HEAD"))

	committed, err := resolveBoundaryReviewManifest(liveBoundaryGit{}, boundaryReviewManifestRequest{
		BaseRef: base, HeadRef: "HEAD", IssuesDir: "workshop/issues", HistoryDir: "workshop/history",
	})
	if err != nil {
		t.Fatal(err)
	}
	if committed.BaseSHA != base || committed.HeadSHA != head || committed.WorkingTree {
		t.Fatalf("committed manifest = %+v", committed)
	}

	working, err := resolveBoundaryReviewManifest(liveBoundaryGit{}, boundaryReviewManifestRequest{
		BaseRef: base, IssuesDir: "workshop/issues", HistoryDir: "workshop/history",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !working.WorkingTree || working.AmbientHeadSHA != head {
		t.Fatalf("working manifest = %+v", working)
	}
}

func formatGitCalls(calls [][]string) string {
	lines := make([]string, len(calls))
	for i, call := range calls {
		lines[i] = strings.Join(call, " ")
	}
	return strings.Join(lines, "\n")
}

const (
	testReviewBaseSHA = "1111111111111111111111111111111111111111"
	testReviewHeadSHA = "2222222222222222222222222222222222222222"
)
