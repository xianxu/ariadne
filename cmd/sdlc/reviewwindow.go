package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

type boundaryGitRunner interface {
	RunGit(args ...string) ([]byte, error)
}

type liveBoundaryGit struct{}

func (liveBoundaryGit) RunGit(args ...string) ([]byte, error) {
	return gitx.RunGit(args...)
}

type boundaryReviewManifestRequest struct {
	BaseRef, HeadRef      string
	IssuesDir, HistoryDir string
	IssueNum              int
	PlansDir              string
}

// resolveBoundaryReviewManifest is the IO shell for the pure judge manifest.
// It pins symbolic refs to commits and validates repository/tracker paths before
// any reviewer subprocess is launched.
func resolveBoundaryReviewManifest(git boundaryGitRunner, req boundaryReviewManifestRequest) (judge.ReviewWindowManifest, error) {
	root, err := runBoundaryGitValue(git, "repository root", "rev-parse", "--show-toplevel")
	if err != nil {
		return judge.ReviewWindowManifest{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return judge.ReviewWindowManifest{}, fmt.Errorf("resolve review repository root: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("not a directory")
		}
		return judge.ReviewWindowManifest{}, fmt.Errorf("review repository root unavailable %q: %w", root, err)
	}

	base, err := resolveBoundaryCommit(git, "base", req.BaseRef)
	if err != nil {
		return judge.ReviewWindowManifest{}, err
	}
	issuesDir, err := repositoryRelativeReviewDir(root, req.IssuesDir, "issues")
	if err != nil {
		return judge.ReviewWindowManifest{}, err
	}
	historyDir, err := repositoryRelativeReviewDir(root, req.HistoryDir, "history")
	if err != nil {
		return judge.ReviewWindowManifest{}, err
	}
	m := judge.ReviewWindowManifest{
		RepoRoot: root, BaseSHA: base,
		IssuesDir: issuesDir, HistoryDir: historyDir,
	}
	if req.HeadRef == "" {
		m.WorkingTree = true
		m.AmbientHeadSHA, err = resolveBoundaryCommit(git, "ambient HEAD", "HEAD")
	} else {
		m.HeadSHA, err = resolveBoundaryCommit(git, "head", req.HeadRef)
	}
	if err != nil {
		return judge.ReviewWindowManifest{}, err
	}

	if req.IssueNum > 0 {
		m.IssueFile, err = resolveReviewIssuePath(root, req.IssuesDir, req.IssueNum)
		if err != nil {
			return judge.ReviewWindowManifest{}, err
		}
		m.PlanFile = optionalReviewPlanPath(root, req.PlansDir, m.IssueFile)
	}
	if _, err := judge.RenderReviewWindow(m); err != nil {
		return judge.ReviewWindowManifest{}, fmt.Errorf("invalid review window: %w", err)
	}
	return m, nil
}

func repositoryRelativeReviewDir(root, configured, label string) (string, error) {
	if strings.TrimSpace(configured) == "" {
		return "", fmt.Errorf("review %s directory is empty", label)
	}
	abs := filepath.Clean(configured)
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(root, abs)
	} else if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", fmt.Errorf("make review %s directory repository-relative: %w", label, err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("review %s directory %q must be inside repository %q", label, configured, root)
	}
	return filepath.ToSlash(rel), nil
}

func resolveBoundaryCommit(git boundaryGitRunner, label, ref string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", fmt.Errorf("review %s ref is empty", label)
	}
	return runBoundaryGitValue(git, label+" commit", "rev-parse", "--verify", ref+"^{commit}")
}

func runBoundaryGitValue(git boundaryGitRunner, label string, args ...string) (string, error) {
	out, err := git.RunGit(args...)
	if err != nil {
		return "", fmt.Errorf("resolve review %s with git %s: %w", label, strings.Join(args, " "), err)
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return "", fmt.Errorf("resolve review %s with git %s: empty output", label, strings.Join(args, " "))
	}
	return value, nil
}

func resolveReviewIssuePath(root, issuesDir string, issueNum int) (string, error) {
	searchDir := issuesDir
	if !filepath.IsAbs(searchDir) {
		searchDir = filepath.Join(root, searchDir)
	}
	path, err := issueFilePath(searchDir, issueNum)
	if err != nil {
		return "", fmt.Errorf("resolve review issue: %w", err)
	}
	if filepath.IsAbs(issuesDir) {
		return path, nil
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", fmt.Errorf("make review issue path repository-relative: %w", err)
	}
	return rel, nil
}

func optionalReviewPlanPath(root, plansDir, issueFile string) string {
	displayPath, statPath := reviewPlanPaths(root, plansDir, issueFile)
	if displayPath == "" {
		return ""
	}
	if info, err := os.Stat(statPath); err != nil || info.IsDir() {
		return ""
	}
	return displayPath
}

func reviewPlanPaths(root, plansDir, issueFile string) (displayPath, statPath string) {
	if plansDir == "" || issueFile == "" {
		return "", ""
	}
	name := strings.TrimSuffix(filepath.Base(issueFile), filepath.Ext(issueFile)) + "-plan.md"
	displayPath = filepath.Join(plansDir, name)
	statPath = displayPath
	if !filepath.IsAbs(statPath) {
		statPath = filepath.Join(root, statPath)
	}
	return displayPath, statPath
}
