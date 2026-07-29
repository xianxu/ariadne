package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// estimateTimingRE finds any co-occurrence of "estimate" and "start-plan" within 80
// characters — the SEMANTIC sweep, not a list of known literals.
//
// This distinction is the whole point (#187 round-4 finding). A guard that checked a
// hand-listed set of strings passed clean while five live surfaces still told agents the
// estimate is set at start-plan: helptext/issue.md, helptext/set-status.md,
// atlas/workflow/issue-lifecycle.md (twice), and atlas/workflow/sdlc-binary.md. A literal
// sweep cannot find what it does not already know about. It is the #167 lesson applied —
// a prose policy with many consumers drifts unless a test reads all of them.
var estimateTimingRE = regexp.MustCompile(`(?i)estimate.{0,80}start-plan|start-plan.{0,80}estimate`)

// estimateTimingAllowed are the surfaces verified to carry NO timing claim: they mention
// both words incidentally. Each entry is a decision, not a suppression — re-audit if the
// file's prose changes.
var estimateTimingAllowed = map[string]string{
	"cmd/sdlc/helptext/estimate.md":        "documents the ## Estimate block grammar; states no timing",
	"cmd/sdlc/helptext/estimate-source.md": "points at the calibration source; states no timing",
	"cmd/sdlc/startplan.go":                "the nudge itself; asserted by TestEstimateNudge in startplan_test.go",
	"cmd/sdlc/startplan_test.go":           "asserts the nudge's wording",
	"cmd/sdlc/estimatetiming_test.go":      "this guard",
	"atlas/workflow/sdlc-binary.md":        "gate-order table + estimate-source seam; retimed, asserted below",
}

// TestEstimateTimingConsistency pins #187 B2 across every surface that tells an agent WHEN
// to derive the estimate. After B1 the answer is "after the plan clears plan-quality" —
// never "at start-plan".
func TestEstimateTimingConsistency(t *testing.T) {
	root := repoRootForTest(t)
	for _, rel := range trackedFilesForTest(t, root) {
		if strings.HasPrefix(rel, "workshop/") || estimateTimingAllowed[rel] != "" {
			continue
		}
		switch filepath.Ext(rel) {
		case ".md", ".go", ".cue", ".sh":
		default:
			continue
		}
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(body), "\n") {
			if !estimateTimingRE.MatchString(line) {
				continue
			}
			if strings.Contains(line, "after the plan clears plan-quality") ||
				strings.Contains(line, "not claim") && strings.Contains(line, "plan-quality") {
				continue
			}
			t.Errorf("%s:%d still ties the estimate to start-plan (#187 B2):\n  %s", rel, i+1, strings.TrimSpace(line))
		}
	}
}

// TestEstimateTimingStatedPositively pins the other half: the surfaces that OWN the claim
// must state it, so the sweep above can't be satisfied by deleting every mention.
func TestEstimateTimingStatedPositively(t *testing.T) {
	root := repoRootForTest(t)
	for _, rel := range []string{
		"AGENTS.base.md",
		"cmd/sdlc/helptext/change-code.md",
		"cmd/sdlc/helptext/issue.md",
		"cmd/sdlc/helptext/set-status.md",
		"atlas/workflow/issue-lifecycle.md",
		// start-plan.md is here and NOT in the sweep above for a structural reason: the
		// sweep needs "estimate" and "start-plan" within 80 chars ON ONE LINE, and in
		// this file the identifier is the filename, not the prose — so the guard built to
		// catch exactly this class was blind to its most obvious surface (M1 review I3).
		// The positive assertion is what covers it; a revert fails here.
		"cmd/sdlc/helptext/start-plan.md",
	} {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if !strings.Contains(string(body), "plan-quality") {
			t.Errorf("%s must state the new estimate timing (after plan-quality clears)", rel)
		}
	}
	// The gate-order table must show plan-quality BEFORE the estimate gates.
	body, err := os.ReadFile(filepath.Join(root, "atlas/workflow/sdlc-binary.md"))
	if err != nil {
		t.Fatal(err)
	}
	row := ""
	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, "Planning → implementation gate") {
			row = line
		}
	}
	if row == "" {
		t.Fatal("sdlc-binary.md has no change-code gate row")
	}
	if strings.Index(row, "plan-quality") > strings.Index(row, "estimate (#113)") {
		t.Errorf("the atlas gate-order row still lists estimate before plan-quality (#187 B1):\n  %s", row)
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skipf("not in a git repo: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// trackedFilesForTest lists git-tracked paths, so the sweep covers the real repo rather
// than a hand-listed set — and so it automatically covers files added later.
func trackedFilesForTest(t *testing.T, root string) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("git ls-files failed: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}
