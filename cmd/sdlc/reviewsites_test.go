package main

import (
	"bytes"
	"strings"
	"testing"
)

// #409: every review site chooses its reviewer through the repo's policy, not
// only the boundary review. Each runs in a repo configured `other` with a
// claude session, so the policy's answer is codex; a site that bypassed the
// policy would review with claude.
func otherPolicyClaudeSession(t *testing.T) {
	t.Helper()
	t.Setenv("AGENT_CMD", "")
	t.Setenv("CODEX_CI", "")
	t.Setenv("CODEX_THREAD_ID", "")
	t.Setenv("PAIR_AGENT", "claude")
	boundaryWindow(t) // a committed repository, the working directory
	writeReviewConfig(t, ".", `{"boundary_reviewer": "other"}`)
}

func TestPlanQualityChoosesThroughThePolicy(t *testing.T) {
	otherPolicyClaudeSession(t)
	seen := stubJudgeName(t)
	f := &changeCodeFlags{PlansDir: t.TempDir()}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "issue", "000001-issue.md", "## Spec\n\nx", ""); err != nil {
		t.Fatalf("runPlanQualityJudge: %v", err)
	}
	if *seen != "codex" {
		t.Fatalf("plan-quality reviewed by %q, want codex (the policy's other family)", *seen)
	}
}

func TestEstimateQualityChoosesThroughThePolicy(t *testing.T) {
	otherPolicyClaudeSession(t)
	seen := stubJudgeName(t)
	content := "---\nid: 1\nstatus: working\nestimate_hours: 0.2\n---\n# T\n\n## Estimate\n\n```estimate\nmodel: estimate-logic-v2\nitem: smaller-go-module design=0.0 impl=0.2\ntotal: 0.2\n```\n"
	if err := runEstimateQualityJudge(ioDiscard(), ioDiscard(), &changeCodeFlags{}, "issue", content); err != nil {
		t.Fatalf("runEstimateQualityJudge: %v", err)
	}
	if *seen != "codex" {
		t.Fatalf("estimate-quality reviewed by %q, want codex", *seen)
	}
}

func TestSdlcJudgeChoosesThroughThePolicy(t *testing.T) {
	otherPolicyClaudeSession(t)
	var stdout, stderr bytes.Buffer
	f := &judgeFlags{Base: "HEAD", Head: "HEAD", DryRun: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	if err := runJudge(&stdout, &stderr, "dry", f); err != nil {
		t.Fatalf("runJudge: %v\n%s", err, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "codex exec") {
		t.Fatalf("sdlc judge would not run codex:\n%s", got)
	}
}
