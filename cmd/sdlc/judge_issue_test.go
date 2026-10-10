package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// planningJudgeRepo builds a repo holding issue #189's shape: an issue with a
// distinctive ## Spec, a durable plan, and a plan-quality ledger with one prior
// finding. It returns the issue path, the plan path and the ledger path.
func planningJudgeRepo(t *testing.T) (issuePath, planPath, ledgerPath string) {
	t.Helper()
	testfix.Repo(t, testfix.Chdir(), testfix.InitialCommit())
	for _, d := range []string{"workshop/issues", "workshop/plans"} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	issuePath = filepath.Join("workshop/issues", "000069-x.md")
	issueBody := "---\nid: 000069\nstatus: working\n---\n\n# x\n\n## Spec\n\nSPEC-SENTINEL the judge must see.\n\n" +
		"## Plan\n\n- [ ] do it\n\n## Log\n"
	planPath = filepath.Join("workshop/plans", "000069-x-plan.md")
	for p, body := range map[string]string{issuePath: issueBody, planPath: "# plan\n\nPLAN-SENTINEL body.\n"} {
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ledger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 69, IDPrefix: "PQ", Rounds: []gatestate.Round{{
		N: 1, Timestamp: "2026-10-10T00:00:00Z", Agent: "claude",
		New: []gatestate.Finding{{ID: "PQ-1", Severity: "Important", Title: "PRIOR-FINDING-SENTINEL", Round: 1}},
	}}}
	if err := writePlanGateLedger("workshop/plans", "000069-x.md", ledger, "ariadne"); err != nil {
		t.Fatal(err)
	}
	return issuePath, planPath, planGatePath("workshop/plans", "000069-x.md")
}

func planningJudgeFlags(dryRun bool) *judgeFlags {
	return &judgeFlags{
		Base: "HEAD", Head: "HEAD", DryRun: dryRun, Issue: 69,
		IssuesDir: "workshop/issues", HistoryDir: "workshop/history", PlansDir: "workshop/plans",
	}
}

// promptBetween cuts a dry-run's prompt out of its framing.
func promptBetween(t *testing.T, out, header string) string {
	t.Helper()
	_, rest, ok := strings.Cut(out, header+"\n")
	if !ok {
		t.Fatalf("dry-run output lacks %q:\n%s", header, out)
	}
	prompt, _, ok := strings.Cut(rest, "\n── command (would invoke) ──")
	if !ok {
		t.Fatalf("dry-run output lacks the command footer:\n%s", out)
	}
	return prompt
}

// #189: the manual plan-quality judge rendered an EMPTY issue. It now renders the
// prompt change-code sends, byte for byte, and only reads the ledger.
func TestJudgePlanQuality_DryRunRendersChangeCodePrompt(t *testing.T) {
	issuePath, planPath, ledgerPath := planningJudgeRepo(t)
	ledgerBefore, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if err := runJudge(&stdout, &stderr, "plan-quality", planningJudgeFlags(true)); err != nil {
		t.Fatalf("runJudge: %v\n%s", err, stderr.String())
	}
	got := promptBetween(t, stdout.String(), "── prompt ──")
	for _, want := range []string{"SPEC-SENTINEL the judge must see.", "PLAN-SENTINEL body.", "PRIOR-FINDING-SENTINEL", "ariadne#69"} {
		if !strings.Contains(got, want) {
			t.Errorf("plan-quality prompt missing %q:\n%s", want, got)
		}
	}

	issueBytes, _ := os.ReadFile(issuePath)
	planBytes, _ := os.ReadFile(planPath)
	var ccOut, ccErr bytes.Buffer
	cf := &changeCodeFlags{Issue: 69, IssuesDir: "workshop/issues", PlansDir: "workshop/plans", DryRun: true}
	if err := runPlanQualityJudge(&ccOut, &ccErr, cf, "000069-x", issuePath, string(issueBytes), string(planBytes)); err != nil {
		t.Fatalf("runPlanQualityJudge: %v", err)
	}
	if want := promptBetween(t, ccOut.String(), "── plan-quality prompt ──"); got != want {
		t.Errorf("judge and change-code render different plan-quality prompts\n--- judge ---\n%s\n--- change-code ---\n%s", got, want)
	}

	if after, _ := os.ReadFile(ledgerPath); !bytes.Equal(after, ledgerBefore) {
		t.Errorf("a manual plan-quality render wrote the ledger:\nbefore:\n%s\nafter:\n%s", ledgerBefore, after)
	}
}

// A live plan-quality run would be a review nobody records; it refuses and names the
// verb that owns the gate, before any dispatch.
func TestJudgePlanQuality_LiveRunRefuses(t *testing.T) {
	planningJudgeRepo(t)
	calls, _ := stubJudge(t, "No findings.")
	var stdout, stderr bytes.Buffer
	msg, died := expectDie(t, func() { _ = runJudge(&stdout, &stderr, "plan-quality", planningJudgeFlags(false)) })
	if !died {
		t.Fatalf("live plan-quality ran instead of refusing:\n%s", stdout.String())
	}
	if !strings.Contains(msg, "sdlc change-code") || !strings.Contains(msg, "--dry-run") {
		t.Errorf("refusal does not point at change-code and --dry-run: %q", msg)
	}
	if *calls != 0 {
		t.Errorf("refused run still dispatched %d time(s)", *calls)
	}
}

// An issue-less plan-quality run refuses by the --issue check, not by an empty render.
func TestJudgePlanQuality_RefusesWithoutIssue(t *testing.T) {
	planningJudgeRepo(t)
	f := planningJudgeFlags(true)
	f.Issue = 0
	var stdout, stderr bytes.Buffer
	msg, died := expectDie(t, func() { _ = runJudge(&stdout, &stderr, "plan-quality", f) })
	if !died {
		t.Fatalf("plan-quality rendered without --issue:\n%s", stdout.String())
	}
	if !strings.Contains(msg, "--issue") {
		t.Errorf("refusal is not the --issue check: %q", msg)
	}
}
