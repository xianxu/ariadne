package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// runLandingDuplicateGate reads the already-fetched configured destination. It
// neither fetches origin nor silently waives an unreadable collision check. It
// measures the pinned head being landed, the same window as the instance gate.
func runLandingDuplicateGate(mainOID, headOID, issuesDir, historyDir string, r gitRunner) error {
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
	head, err := refIDSpace(headOID, dirs, r)
	if err != nil {
		return fmt.Errorf("landing duplicate-id HEAD: %w", err)
	}
	baseOID, err := landingMergeBase(r, mainOID, headOID)
	if err != nil {
		return fmt.Errorf("landing duplicate-id %w", err)
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

// landingMergeBase is the branch point the landing gates measure the branch from:
// merge-base(main, ref). Unavailable fails closed — no gate may fall back to main's
// tip, whose window holds every change main made since the branch was cut (#322).
func landingMergeBase(r gitRunner, mainOID, ref string) (string, error) {
	out, err := r.Git("merge-base", mainOID, ref)
	base := strings.TrimSpace(string(out))
	if err != nil || !landingOIDValid(base) {
		return "", fmt.Errorf("merge-base unavailable: %v: %s", err, out)
	}
	return base, nil
}

// runLandingInstanceGate validates the instance files the branch itself changed:
// merge-base(main, head)..head. It used to diff main's tip against the working
// tree, so an issue main archived after the branch point read as ADDED by the
// branch and refused the landing on a file the branch never touched (#322).
// root is the tree checked out at headOID ("" = the current directory).
func runLandingInstanceGate(root, mainOID, headOID string, gates []nounGate, stdout, stderr io.Writer, r gitRunner) error {
	base, err := landingMergeBase(r, mainOID, headOID)
	if err != nil {
		return fmt.Errorf("landing instance-conformance gate %w", err)
	}
	return validateChangedInstancesFn(root, base, headOID, gates, stdout, stderr)
}

// Ownership follows close ancestry, not a body diff: an independently published
// issue copy on main must not erase the original branch's review obligations.
func runLandingPublishGate(ctx context.Context, pr landingPR, issuesDir, plansDir string, stderr io.Writer) error {
	if !landingOIDValid(pr.HeadOID) || !landingOIDValid(pr.BaseOID) {
		return fmt.Errorf("landing publish gate requires pinned PR head and base")
	}
	head, err := gitx.GitOutputContext(ctx, "rev-parse", "--verify", "HEAD")
	if err != nil {
		return fmt.Errorf("landing publish gate read checkout HEAD: %w", err)
	}
	if strings.TrimSpace(string(head)) != pr.HeadOID {
		return fmt.Errorf("landing publish gate checkout HEAD differs from selected PR head")
	}
	if err := guardTransferredDetailsFn(ctx); err != nil {
		return err
	}
	root, err := gitx.RepoTopLevelContext(ctx)
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
	rs, err := loadIssueRecords(ctx, filepath.Join(root, rel), tracker.Fresh)
	if err != nil {
		return err
	}
	if rs.Tracker {
		env, err := openTrackerAt(ctx, root)
		if err != nil {
			return err
		}
		if err := refuseReopenedCompletions(env, rs, pr.HeadOID, pr.BaseOID); err != nil {
			return err
		}
	}
	selected, err := selectLandingIssues(ctx, root, pr, rel)
	if err != nil {
		return err
	}
	planRel, err := gitx.InsideRoot(root, resolvePlansDir(plansDir))
	if err != nil {
		return err
	}
	tree, err := archiveTreeContext(ctx, root, pr.HeadOID, landingArchiveRoots{issues: rel, plans: planRel})
	if err != nil {
		return err
	}
	var paths, issues []string
	for name := range tree {
		paths = append(paths, name)
	}
	for _, item := range selected {
		issues = append(issues, item.path)
	}
	if err := validateArchiveSideNames(paths, issues, rel, planRel); err != nil {
		return err
	}
	entries := make([]publishIssue, 0, len(selected))
	for _, item := range selected {
		entries = append(entries, publishIssue{Path: item.path, Anchor: item.anchor})
	}
	return validatePublishAnchors(ctx, entries, stderr)
}
