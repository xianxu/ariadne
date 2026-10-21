package main

import (
	"fmt"
	"io"
	"path"
	"strings"
)

// Only issue records, not the broader docs-only suite-reuse surface (#472).
// Keep the policy in one place; ops/ is deliberately not admitted.
var issueRecordRoots = []string{"workshop/issues/", "workshop/issue-cards/", "workshop/plans/", "workshop/history/"}

func issueRecordOnlyPaths(paths []string) bool {
	if len(paths) == 0 {
		return false
	}
	for _, p := range paths {
		if path.Clean(p) != p || strings.HasPrefix(p, "/") || !strings.HasSuffix(p, ".md") {
			return false
		}
		allowed := false
		for _, root := range issueRecordRoots {
			allowed = allowed || strings.HasPrefix(p, root)
		}
		if !allowed {
			return false
		}
	}
	return true
}

// Enumerate each commit's changes, including every merge parent and rename
// source. A net tree diff would hide code added and then reverted.
func mainMovementIssueOnly(r gitRunner, root, head, main string) (bool, error) {
	if !landingOIDValid(head) || !landingOIDValid(main) {
		return false, fmt.Errorf("main movement requires full commit IDs")
	}
	raw, err := r.GitInDir(root, "log", "--format=", "--name-only", "-z", "--no-renames", "--root", "-m", head+".."+main, "--")
	if err != nil {
		return false, fmt.Errorf("read main movement: %w", err)
	}
	if len(raw) == 0 {
		return false, nil
	}
	if raw[len(raw)-1] != 0 {
		return false, fmt.Errorf("main movement lacks a NUL terminator")
	}
	var paths []string
	for _, p := range strings.Split(string(raw), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return issueRecordOnlyPaths(paths), nil
}

// Each observation carries the same path rule at every integration boundary.
// Recovery charges the actual interrupted merge, not later main changes.
func observeIntegrationMain(r gitRunner, t landingTarget, s integrationState, main string, stderr io.Writer) (integrationEvent, error) {
	included, err := integrationAncestor(r, t.Root, main, s.head)
	if err != nil {
		return integrationEvent{}, err
	}
	event := integrationEvent{kind: integrationMainObserved, main: main, included: included}
	origin, target := s.head, main
	if s.resume && s.phase == integrationGates {
		origin = s.pr.HeadOID
		target, err = landingGit(r, t.Root, "rev-parse", "--verify", s.head+"^2^{commit}")
	} else if included {
		return event, nil
	}
	if err == nil {
		event.issueOnly, err = mainMovementIssueOnly(r, t.Root, origin, target)
	}
	if err != nil {
		fmt.Fprintf(stderr, "issue-record optimization unavailable (%v); treating main movement as ordinary\n", err)
	}
	return event, nil
}

func readinessAllowsUnintegrated(p landingPR, observed landingReadiness, main string) (bool, error) {
	if observed.Repo != p.Repo || observed.Number != p.Number || observed.HeadRef != p.HeadRef || observed.HeadOID != p.HeadOID || observed.BaseRef != p.BaseRef || observed.State != "OPEN" {
		return false, fmt.Errorf("PR identity changed while observing landing readiness; preserve checkout and retry")
	}
	return observed.BaseTargetOID == main && observed.MergeStateStatus == "CLEAN", nil
}
