package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// runLandingDuplicateGate reads the already-fetched configured destination. It
// neither fetches origin nor silently waives an unreadable collision check.
func runLandingDuplicateGate(mainOID, issuesDir, historyDir string, r gitRunner) error {
	if !landingOIDValid(mainOID) {
		return fmt.Errorf("landing duplicate-id gate requires pinned main commit")
	}
	dirs, err := resolveIDDirs(issuesDir, historyDir)
	if err != nil {
		return err
	}
	trunk, err := refIDSpace(mainOID, dirs, r)
	if err != nil {
		return fmt.Errorf("landing duplicate-id target: %w", err)
	}
	head, err := refIDSpace("HEAD", dirs, r)
	if err != nil {
		return fmt.Errorf("landing duplicate-id HEAD: %w", err)
	}
	out, err := r.Git("merge-base", mainOID, "HEAD")
	baseOID := strings.TrimSpace(string(out))
	if err != nil || !landingOIDValid(baseOID) {
		return fmt.Errorf("landing duplicate-id merge-base unavailable: %v: %s", err, out)
	}
	base, err := refIDSpace(baseOID, dirs, r)
	if err != nil {
		return fmt.Errorf("landing duplicate-id base: %w", err)
	}
	clashes := renderClashes(head, base, trunk)
	if len(clashes) > 0 {
		return fmt.Errorf("landing would introduce %d duplicate issue ID(s) against configured main %s:\n%s", len(clashes), shortSHA(mainOID), strings.Join(clashes, "\n"))
	}
	return nil
}

// Ownership follows close ancestry, not a body diff: an independently published
// issue copy on main must not erase the original branch's review obligations.
func runLandingPublishGate(pr landingPR, issuesDir string, stderr io.Writer) error {
	if !landingOIDValid(pr.HeadOID) || !landingOIDValid(pr.BaseOID) {
		return fmt.Errorf("landing publish gate requires pinned PR head and base")
	}
	head, err := gitx.RunGit("rev-parse", "--verify", "HEAD")
	if err != nil || strings.TrimSpace(string(head)) != pr.HeadOID {
		return fmt.Errorf("landing publish gate checkout HEAD differs from selected PR head")
	}
	root, err := gitx.RepoTopLevel()
	if err != nil {
		return err
	}
	rel, err := gitx.InsideRoot(root, issuesDir)
	if err != nil {
		return err
	}
	if rel == "." {
		return fmt.Errorf("landing publish gate requires a scoped issues directory")
	}
	selected, err := selectLandingIssues(root, pr, rel)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(selected))
	for _, item := range selected {
		paths = append(paths, item.path)
	}
	return validatePublishIssues(paths, stderr)
}
