package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestEstimateRefusal pins change-code's estimate gate (#113): the universal
// estimate requirement relocated from claim. A missing/empty/invalid
// estimate_hours refuses (estimate-present failure); a positive one passes;
// --no-estimate skips the gate entirely.
func TestEstimateRefusal(t *testing.T) {
	withEst := "---\nid: 000001\nstatus: working\nestimate_hours: 4\n---\n# T\n"
	noEst := "---\nid: 000001\nstatus: working\n---\n# T\n"

	if got := estimateRefusal(withEst, false); got != nil {
		t.Errorf("positive estimate should pass the gate, got %+v", *got)
	}
	if got := estimateRefusal(noEst, false); got == nil {
		t.Error("missing estimate should refuse, got nil")
	} else if got.Name != "estimate-present" {
		t.Errorf("failure name = %q, want estimate-present", got.Name)
	}
	// --no-estimate bypasses even a missing estimate.
	if got := estimateRefusal(noEst, true); got != nil {
		t.Errorf("--no-estimate should skip the gate, got %+v", *got)
	}
}

// TestEstimateReconRefusal pins change-code's estimate-reconciliation gate
// (#117): a reconciling ## Estimate block passes; a frontmatter/total mismatch or
// a missing block refuses; --no-estimate-recon skips the gate.
func TestEstimateReconRefusal(t *testing.T) {
	const estBlock = "## Estimate\n\n```estimate\n" +
		"model: estimate-logic-v2\n" +
		"familiarity: 1.0\n" +
		"item: greenfield-go-module design=0.3 impl=0.6\n" +
		"item: smaller-go-module design=0.2 impl=0.6\n" +
		"item: smaller-go-module design=0.2 impl=0.5\n" +
		"item: atlas-docs design=0.0 impl=0.2\n" +
		"item: milestone-review design=0.0 impl=0.6\n" +
		"design-buffer: 0.30\n" +
		"total: 3.4\n```\n"
	green := "---\nid: 1\nstatus: working\nestimate_hours: 3.4\n---\n# T\n\n" + estBlock
	if got := estimateReconRefusal(green, false); got != nil {
		t.Errorf("reconciling block should pass, got: %s", got.Message)
	}
	mismatch := "---\nid: 1\nstatus: working\nestimate_hours: 7\n---\n# T\n\n" + estBlock
	if estimateReconRefusal(mismatch, false) == nil {
		t.Error("estimate_hours 7 vs total 3.4 should fail")
	}
	noBlock := "---\nid: 1\nstatus: working\nestimate_hours: 3.4\n---\n# T\n\n## Spec\n\nx\n"
	if estimateReconRefusal(noBlock, false) == nil {
		t.Error("missing ## Estimate block should fail")
	}
	if estimateReconRefusal(noBlock, true) != nil {
		t.Error("--no-estimate-recon should skip the gate")
	}
}

// TestRunEstimateQualityJudge_SkipsWhenNoBlock pins the #117 M2-review fix: the
// estimate-quality judge must skip silently (no dispatch, no output) when the
// issue carries no ## Estimate block — otherwise inverting that guard would
// dispatch an LLM on every block-less issue.
func TestRunEstimateQualityJudge_SkipsWhenNoBlock(t *testing.T) {
	var out, errb bytes.Buffer
	f := &changeCodeFlags{DryRun: true}
	noBlock := "---\nid: 1\nstatus: working\nestimate_hours: 1\n---\n# T\n\n## Spec\n\nx\n"
	if err := runEstimateQualityJudge(&out, &errb, f, "t", noBlock); err != nil {
		t.Fatalf("expected nil (skip) for a block-less issue, got %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output when skipping, got %q", out.String())
	}
}

// TestPromptBranchingTTY pins the tty-prompt's character-mapping
// contract: a single-letter answer (case-insensitive) maps to the
// internal "yes" / "no" / cancel verbs. Drift here would silently
// confuse operators who type the wrong character.
func TestPromptBranchingTTY(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"w worktree", "w\n", "yes", false},
		{"y yes", "y\n", "yes", false},
		{"worktree word", "worktree\n", "yes", false},
		{"W uppercase", "W\n", "yes", false},
		{"m main", "m\n", "no", false},
		{"n no", "n\n", "no", false},
		{"in-place phrase", "in-place\n", "no", false},
		{"c cancel", "c\n", "", true},
		{"empty == cancel", "\n", "", true},
		{"unrecognized errors", "z\n", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stderr := &bytes.Buffer{}
			got, err := promptBranchingTTY(strings.NewReader(tt.input), stderr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

// TestIssueTitleFromContent pins the H1 extraction — used for the
// sizing-hint label. A missing H1 mustn't crash; it labels as "(no
// title)" so the hint still renders.
func TestIssueTitleFromContent(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"normal h1", "---\nfm\n---\n\n# My Title\n\nbody", "My Title"},
		{"h1 with trailing whitespace", "# Spaced  \n", "Spaced"},
		{"no h1 falls back", "no heading here\nat all\n", "(no title)"},
		{"h2 doesn't count", "## Subhead\nbody", "(no title)"},
		{"first h1 wins over later h1", "# First\nmid\n# Second\n", "First"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := issueTitleFromContent(tt.text); got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

// TestIsTTY_PipeIsNotTTY ensures the non-tty branch is taken when a
// test pipes stdin in — this drives the agent-protocol path and is a
// load-bearing assumption of the rest of the test suite.
func TestIsTTY_PipeIsNotTTY(t *testing.T) {
	// strings.Reader is not an *os.File → not a tty by construction.
	if isTTY(strings.NewReader("hi")) {
		t.Error("strings.Reader should not be a tty")
	}
}

// TestSentinelStable pins the sentinel value as a stable contract.
// The xx-sdlc skill grep's for it; any change here must land at the
// same time as the skill update.
func TestSentinelStable(t *testing.T) {
	if sentinelBranchingStrategy != "ASK_BRANCHING_STRATEGY" {
		t.Errorf("sentinel changed to %q — update xx-sdlc skill in lockstep",
			sentinelBranchingStrategy)
	}
	if askExitCode != 2 {
		t.Errorf("askExitCode changed to %d — update xx-sdlc skill in lockstep",
			askExitCode)
	}
}
