package judge

import (
	"strings"
	"testing"
)

const (
	testBaseSHA = "1111111111111111111111111111111111111111"
	testHeadSHA = "2222222222222222222222222222222222222222"
)

func TestRenderReviewWindow_CommittedRangeUsesPinnedArgv(t *testing.T) {
	m := ReviewWindowManifest{
		RepoRoot:   "/tmp/repo with ' quote",
		BaseSHA:    testBaseSHA,
		HeadSHA:    testHeadSHA,
		IssuesDir:  "custom/issues",
		HistoryDir: "custom/history",
		IssueFile:  "custom/issues/000162-review.md",
		PlanFile:   "custom/plans/000162-review-plan.md",
	}

	got, err := RenderReviewWindow(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"mode: committed range",
		"base: " + testBaseSHA,
		"head: " + testHeadSHA,
		"issue file: custom/issues/000162-review.md",
		"plan file: custom/plans/000162-review-plan.md",
		`stat: git -C '/tmp/repo with '"'"' quote' diff --stat ` + testBaseSHA + ` ` + testHeadSHA + ` -- ':!custom/issues/' ':!custom/history/'`,
		`names: git -C '/tmp/repo with '"'"' quote' diff --name-status ` + testBaseSHA + ` ` + testHeadSHA + ` -- ':!custom/issues/' ':!custom/history/'`,
		`full: git -C '/tmp/repo with '"'"' quote' diff ` + testBaseSHA + ` ` + testHeadSHA + ` -- ':!custom/issues/' ':!custom/history/'`,
		`targeted: git -C '/tmp/repo with '"'"' quote' diff ` + testBaseSHA + ` ` + testHeadSHA + ` -- '<path-from-name-status>' ':!custom/issues/' ':!custom/history/'`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered manifest missing %q:\n%s", want, got)
		}
	}
}

func TestRenderReviewWindow_WorkingTreeNamesAmbientHeadAndScope(t *testing.T) {
	m := ReviewWindowManifest{
		RepoRoot:       "/tmp/repo",
		BaseSHA:        testBaseSHA,
		AmbientHeadSHA: testHeadSHA,
		WorkingTree:    true,
		IssuesDir:      "workshop/issues",
		HistoryDir:     "workshop/history",
	}

	got, err := RenderReviewWindow(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"mode: working tree",
		"ambient HEAD: " + testHeadSHA,
		"includes committed-after-base plus staged and unstaged tracked changes; excludes untracked files",
		`full: git -C /tmp/repo diff ` + testBaseSHA + ` -- ':!workshop/issues/' ':!workshop/history/'`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered manifest missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "diff "+testBaseSHA+" "+testHeadSHA) {
		t.Fatalf("working-tree command must not pin ambient HEAD as a diff argument:\n%s", got)
	}
}

func TestRenderReviewWindow_RejectsStructurallyInvalidManifests(t *testing.T) {
	valid := ReviewWindowManifest{
		RepoRoot:   "/tmp/repo",
		BaseSHA:    testBaseSHA,
		HeadSHA:    testHeadSHA,
		IssuesDir:  "workshop/issues",
		HistoryDir: "workshop/history",
	}

	mutants := map[string]ReviewWindowManifest{
		"relative root":          withReviewRoot(valid, "relative/repo"),
		"symbolic base":          withReviewBase(valid, "HEAD~1"),
		"missing committed head": withReviewHead(valid, ""),
		"working tree has head":  withWorkingTreeHead(valid, testHeadSHA),
		"absolute issues dir":    withReviewIssuesDir(valid, "/tmp/issues"),
	}
	for name, mutant := range mutants {
		t.Run(name, func(t *testing.T) {
			if _, err := RenderReviewWindow(mutant); err == nil {
				t.Fatal("RenderReviewWindow accepted invalid manifest")
			}
		})
	}
}

func withReviewIssuesDir(m ReviewWindowManifest, issuesDir string) ReviewWindowManifest {
	m.IssuesDir = issuesDir
	return m
}

func withReviewRoot(m ReviewWindowManifest, root string) ReviewWindowManifest {
	m.RepoRoot = root
	return m
}

func withReviewBase(m ReviewWindowManifest, base string) ReviewWindowManifest {
	m.BaseSHA = base
	return m
}

func withReviewHead(m ReviewWindowManifest, head string) ReviewWindowManifest {
	m.HeadSHA = head
	return m
}

func withWorkingTreeHead(m ReviewWindowManifest, head string) ReviewWindowManifest {
	m.WorkingTree = true
	m.HeadSHA = head
	m.AmbientHeadSHA = testHeadSHA
	return m
}
