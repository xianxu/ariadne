package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
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
	missing := estimateReconRefusal(noBlock, false)
	if missing == nil {
		t.Error("missing ## Estimate block should fail")
	} else if !strings.Contains(missing.Message, "estimate-source") {
		// #134: the missing-block error must point at the calibration source, not
		// just say "via the model".
		t.Errorf("missing-block message should point at `sdlc estimate-source`, got: %s", missing.Message)
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

func TestChangeCodeAgentDefault_PlanQualityUsesPairAgent(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	seenName := stubJudgeName(t)

	// PlansDir must be writable: since #187 the plan-quality gate persists a ledger.
	f := &changeCodeFlags{AgentExplicit: false, PlansDir: t.TempDir()}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "issue", "000001-issue.md", "## Spec\n\nx", ""); err != nil {
		t.Fatalf("runPlanQualityJudge: %v", err)
	}
	if *seenName != "codex" {
		t.Fatalf("plan-quality agent = %q, want codex", *seenName)
	}
}

func TestChangeCodeAgentDefault_EstimateQualityUsesPairAgent(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	seenName := stubJudgeName(t)

	f := &changeCodeFlags{AgentExplicit: false}
	content := "---\nid: 1\nstatus: working\nestimate_hours: 0.2\n---\n# T\n\n## Estimate\n\n```estimate\nmodel: estimate-logic-v2\nitem: smaller-go-module design=0.0 impl=0.2\ntotal: 0.2\n```\n"
	if err := runEstimateQualityJudge(ioDiscard(), ioDiscard(), f, "issue", content); err != nil {
		t.Fatalf("runEstimateQualityJudge: %v", err)
	}
	if *seenName != "codex" {
		t.Fatalf("estimate-quality agent = %q, want codex", *seenName)
	}
}

func TestChangeCodeAgentDefault_ExplicitAgentWins(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	seenName := stubJudgeName(t)

	f := &changeCodeFlags{Agent: "claude", AgentExplicit: true, PlansDir: t.TempDir()}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "issue", "000001-issue.md", "## Spec\n\nx", ""); err != nil {
		t.Fatalf("runPlanQualityJudge: %v", err)
	}
	if *seenName != "claude" {
		t.Fatalf("plan-quality agent = %q, want claude", *seenName)
	}
}

func TestChangeCodeAgentDefault_AgentCmdWins(t *testing.T) {
	t.Setenv("AGENT_CMD", "gemini")
	t.Setenv("PAIR_AGENT", "codex")
	seenName := stubJudgeName(t)

	f := &changeCodeFlags{AgentExplicit: false, PlansDir: t.TempDir()}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "issue", "000001-issue.md", "## Spec\n\nx", ""); err != nil {
		t.Fatalf("runPlanQualityJudge: %v", err)
	}
	if *seenName != "gemini" {
		t.Fatalf("plan-quality agent = %q, want gemini", *seenName)
	}
}

func stubJudgeName(t *testing.T) *string {
	t.Helper()
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	seenName := ""
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		seenName = name
		return []byte("VERDICT: CLEAN (confidence: high)\n"), nil
	}
	return &seenName
}

func ioDiscard() *bytes.Buffer {
	return &bytes.Buffer{}
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

// TestIsTTY_RealNonTerminalFilesAreNotTTY is the #141 regression: isTTY must use
// a real terminal probe, not an os.ModeCharDevice check. /dev/null is a CHARACTER
// DEVICE but not a terminal — the old check returned true for it, so an agent
// whose stdin is /dev/null was mistaken for interactive and merge's fail-fast
// confirm gate was silently skipped (judges ran, then the prompt aborted). All of
// these real *os.File readers must report false.
func TestIsTTY_RealNonTerminalFilesAreNotTTY(t *testing.T) {
	devnull, err := os.Open(os.DevNull) // "/dev/null" — a char device, NOT a tty
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	t.Cleanup(func() { devnull.Close() })
	if isTTY(devnull) {
		t.Errorf("%s is a char device but not a terminal — isTTY must be false (#141)", os.DevNull)
	}

	regular := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(regular, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	rf, err := os.Open(regular)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rf.Close() })
	if isTTY(rf) {
		t.Error("a regular file is not a terminal — isTTY must be false")
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pr.Close(); pw.Close() })
	if isTTY(pr) {
		t.Error("an os.Pipe read end is not a terminal — isTTY must be false")
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

// ── #187: the stateful plan gate ───────────────────────────────────────────

// TestGateOrderPlanBeforeEstimate pins B1: the estimate is a FUNCTION of the plan, so it
// is never demanded before plan-quality has accepted the plan. Costing an unapproved plan
// is waste by construction — pair#127 re-derived its estimate five times, four of them
// forced by plan changes that were still in flight.
//
// This guard is real rather than a restatement because runChangeCode ITERATES the same
// declaration changeCodeGateOrder reads: reordering the literal reorders execution.
func TestGateOrderPlanBeforeEstimate(t *testing.T) {
	order := changeCodeGateOrder()
	idx := func(name string) int {
		for i, g := range order {
			if g == name {
				return i
			}
		}
		t.Fatalf("gate %q missing from the order %v", name, order)
		return -1
	}
	if idx("structural") > idx("plan-quality") {
		t.Error("structural must run before plan-quality (it is free — no subprocess)")
	}
	for _, est := range []string{"estimate", "estimate-recon", "estimate-quality"} {
		if idx(est) < idx("plan-quality") {
			t.Errorf("%s must run AFTER plan-quality (#187 B1), order = %v", est, order)
		}
	}
}

// stubJudgeSeq installs a fake judge returning a SEQUENCE of outputs (round 1, round 2, …)
// and captures the prompt each round received. Extends the existing stubJudgeName seam
// rather than standing up a second fake (ARCH-DRY).
func stubJudgeSeq(t *testing.T, outputs ...string) (calls *int, prompts *[]string) {
	t.Helper()
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	n := 0
	var seen []string
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		// The prompt is the last argument for every supported agent CLI.
		if len(args) > 0 {
			seen = append(seen, args[len(args)-1])
		}
		out := outputs[len(outputs)-1]
		if n < len(outputs) {
			out = outputs[n]
		}
		n++
		return []byte(out), nil
	}
	return &n, &seen
}

func findingsReply(verdict, block string) string {
	return "VERDICT: " + verdict + " (confidence: high)\n\nprose\n\n```findings\n" + block + "```\n"
}

// Round 1 raises a Critical: the gate refuses and the ledger records the finding.
// Round 2 disposes it and raises only a Minor: the gate PASSES.
//
// This is the issue's HEADLINE Done-when — "a second change-code on a plan whose
// Critical/Important findings were addressed passes without new blocking findings at a
// lower severity" — and the whole reason the feature exists.
func TestPlanQualityConvergesAcrossRounds(t *testing.T) {
	dir := t.TempDir()
	f := &changeCodeFlags{Issue: 187, PlansDir: dir}
	calls, prompts := stubJudgeSeq(t,
		findingsReply("FAILURE", "findings:\n  - id: new\n    severity: Critical\n    title: seam in wrong layer\n"),
		findingsReply("INFO", "dispose:\n  - id: PQ-1\n    disposition: addressed\nfindings:\n  - id: new\n    severity: Minor\n    title: naming\n"),
	)

	// Round 1 blocks.
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue v1", "plan v1"); err == nil {
		t.Fatal("round 1 with an open Critical must block")
	}
	l, err := readPlanGateLedger(dir, "000187-x.md", 187)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if len(l.Rounds) != 1 || len(l.Rounds[0].New) != 1 || l.Rounds[0].New[0].ID != "PQ-1" {
		t.Fatalf("round 1 not recorded: %+v", l)
	}
	if !l.Rounds[0].Blocked {
		t.Error("round 1 should be recorded as blocked")
	}

	// Round 2 passes — the plan changed, so no pass-through; the Critical is disposed.
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue v2", "plan v2"); err != nil {
		t.Fatalf("round 2 should PASS after the Critical was addressed, got %v", err)
	}
	if *calls != 2 {
		t.Errorf("judge invoked %d times, want 2", *calls)
	}

	// Round 2's prompt must have carried round 1's finding — the convergence mechanism.
	if len(*prompts) < 2 || !strings.Contains((*prompts)[1], "PQ-1") {
		t.Error("round 2's prompt did not carry the prior finding id")
	}
	if !strings.Contains((*prompts)[1], "seam in wrong layer") {
		t.Error("round 2's prompt did not carry the prior finding text")
	}

	l, _ = readPlanGateLedger(dir, "000187-x.md", 187)
	if len(l.Rounds) != 2 {
		t.Fatalf("want 2 recorded rounds, got %d", len(l.Rounds))
	}
	open := gatestate.OpenFindings(l)
	if len(open) != 1 || open[0].Severity != "Minor" {
		t.Errorf("open findings = %+v, want just the Minor", open)
	}
}

// The pass-through is what makes B1 a net win: an unchanged plan must not re-dispatch,
// so an estimate-gate failure costs milliseconds on retry rather than a fresh judge round.
func TestPlanQualityPassThroughOnUnchangedContent(t *testing.T) {
	dir := t.TempDir()
	f := &changeCodeFlags{Issue: 187, PlansDir: dir}
	calls, _ := stubJudgeSeq(t, findingsReply("CLEAN", "findings: []\n"))

	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan"); err != nil {
		t.Fatalf("round 1: %v", err)
	}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan"); err != nil {
		t.Fatalf("round 2 (unchanged): %v", err)
	}
	if *calls != 1 {
		t.Errorf("judge invoked %d times, want 1 — unchanged content must not re-dispatch", *calls)
	}
	l, _ := readPlanGateLedger(dir, "000187-x.md", 187)
	if len(l.Rounds) != 1 {
		t.Errorf("a passed-through invocation must not add a round, got %d", len(l.Rounds))
	}
}

// An EDITED plan must re-dispatch — the short-circuit must never cache away a real review.
func TestPlanQualityRedispatchesWhenContentChanges(t *testing.T) {
	dir := t.TempDir()
	f := &changeCodeFlags{Issue: 187, PlansDir: dir}
	calls, _ := stubJudgeSeq(t, findingsReply("CLEAN", "findings: []\n"))

	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan"); err != nil {
		t.Fatalf("round 1: %v", err)
	}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan EDITED"); err != nil {
		t.Fatalf("round 2: %v", err)
	}
	if *calls != 2 {
		t.Errorf("judge invoked %d times, want 2 — an edited plan must be re-reviewed", *calls)
	}
}

// A short-circuit after a BLOCKING round would let a refused plan walk through unchanged.
func TestPlanQualityNoPassThroughAfterBlockingRound(t *testing.T) {
	dir := t.TempDir()
	f := &changeCodeFlags{Issue: 187, PlansDir: dir}
	calls, _ := stubJudgeSeq(t,
		findingsReply("FAILURE", "findings:\n  - id: new\n    severity: Critical\n    title: seam\n"))

	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan"); err == nil {
		t.Fatal("round 1 must block")
	}
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan"); err == nil {
		t.Fatal("an unchanged BLOCKED plan must still block, never pass through")
	}
	if *calls != 2 {
		t.Errorf("judge invoked %d times, want 2 — a blocking round must not be cached", *calls)
	}
}

// No findings block ⇒ warn loudly, fall back to the verdict token, and STILL persist a
// round. Dropping it would freeze len(Rounds) at 0 for a CLI that never emits the fence:
// the round cap could never fire and gate_rounds would report 0 for the priciest sessions.
func TestPlanQualityFallsBackWithoutFindingsBlock(t *testing.T) {
	dir := t.TempDir()
	f := &changeCodeFlags{Issue: 187, PlansDir: dir}
	stubJudgeSeq(t, "VERDICT: FAILURE (confidence: high)\n\nno block here\n")
	var errb bytes.Buffer

	if err := runPlanQualityJudge(ioDiscard(), &errb, f, "n", "000187-x.md", "issue", "plan"); err == nil {
		t.Fatal("a FAILURE verdict without a block must still block")
	}
	if !strings.Contains(errb.String(), "no valid ```findings block") {
		t.Errorf("must warn loudly about the protocol miss, got: %s", errb.String())
	}
	l, _ := readPlanGateLedger(dir, "000187-x.md", 187)
	if len(l.Rounds) != 1 {
		t.Fatalf("a protocol-miss round must still be persisted, got %d rounds", len(l.Rounds))
	}
	if l.Rounds[0].ProtocolError == "" {
		t.Error("the persisted round must record the protocol error")
	}
}

// --force is a GLOBAL bypass, so it must be recorded ONLY when this gate actually blocked
// — otherwise the accepted-vs-forced count over-reports overrides.
func TestPlanQualityForceRecordedOnlyWhenBlocking(t *testing.T) {
	t.Run("blocking round records the rationale", func(t *testing.T) {
		dir := t.TempDir()
		f := &changeCodeFlags{Issue: 187, PlansDir: dir, Force: "shipping the hotfix"}
		stubJudgeSeq(t, findingsReply("FAILURE", "findings:\n  - id: new\n    severity: Critical\n    title: seam\n"))
		_ = runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan")
		l, _ := readPlanGateLedger(dir, "000187-x.md", 187)
		if l.Rounds[0].Forced != "shipping the hotfix" {
			t.Errorf("Forced = %q, want the rationale", l.Rounds[0].Forced)
		}
	})
	t.Run("passing round records no force", func(t *testing.T) {
		dir := t.TempDir()
		f := &changeCodeFlags{Issue: 187, PlansDir: dir, Force: "forced past something else"}
		stubJudgeSeq(t, findingsReply("CLEAN", "findings: []\n"))
		_ = runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan")
		l, _ := readPlanGateLedger(dir, "000187-x.md", 187)
		if l.Rounds[0].Forced != "" {
			t.Errorf("Forced = %q on a PASSING round; --force is global and must not be attributed here", l.Rounds[0].Forced)
		}
	})
}

// A corrupt ledger must halt the gate, not silently start over — starting over would
// re-open every disposed finding while looking like it worked.
func TestPlanQualityHaltsOnCorruptLedger(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(planGatePath(dir, "000187-x.md"), []byte("---\n:::bad: [yaml\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := &changeCodeFlags{Issue: 187, PlansDir: dir}
	calls, _ := stubJudgeSeq(t, findingsReply("CLEAN", "findings: []\n"))
	if err := runPlanQualityJudge(ioDiscard(), ioDiscard(), f, "n", "000187-x.md", "issue", "plan"); err == nil {
		t.Fatal("a corrupt ledger must halt the gate")
	}
	if *calls != 0 {
		t.Error("must not dispatch the judge when the ledger is unreadable")
	}
}

// roundCapFromEnv reads the operator override and ignores nonsense.
func TestRoundCapFromEnv(t *testing.T) {
	if got := roundCapFromEnv(); got != gatestate.DefaultRoundCap {
		t.Errorf("unset = %d, want %d", got, gatestate.DefaultRoundCap)
	}
	t.Setenv("WF_PLAN_ROUND_CAP", "7")
	if got := roundCapFromEnv(); got != 7 {
		t.Errorf("WF_PLAN_ROUND_CAP=7 = %d, want 7", got)
	}
	for _, bad := range []string{"0", "-1", "banana"} {
		t.Setenv("WF_PLAN_ROUND_CAP", bad)
		if got := roundCapFromEnv(); got != gatestate.DefaultRoundCap {
			t.Errorf("WF_PLAN_ROUND_CAP=%q = %d, want the default", bad, got)
		}
	}
}
