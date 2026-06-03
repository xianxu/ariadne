package judge

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// execCommand is exposed as a var so the test can swap to a fake if
// needed; for now, all tests use the real exec.
var execCommand = exec.Command

func TestIsValid(t *testing.T) {
	for _, name := range []string{"dry", "pure", "plan", "specs", "lessons", "milestone-review"} {
		if !IsValid(name) {
			t.Errorf("IsValid(%q) = false; want true", name)
		}
	}
	for _, name := range []string{"", "dryrun", "Lessons", "milestone"} {
		if IsValid(name) {
			t.Errorf("IsValid(%q) = true; want false", name)
		}
	}
}

func TestCategoryAllowedTools(t *testing.T) {
	// #62 M2: ALL judges are read-only reviewers — none get Edit/Write, incl.
	// Specs (which used to auto-edit docs and could strand a merge with a dirty
	// tree). They report; the main agent applies fixes.
	for _, c := range AllCategories() {
		got := c.AllowedTools()
		if got != "Read,Grep,Glob,Bash" {
			t.Errorf("%s.AllowedTools() = %q, want read-only Read,Grep,Glob,Bash", c, got)
		}
		if strings.Contains(got, "Edit") || strings.Contains(got, "Write") {
			t.Errorf("%s.AllowedTools() contains a write tool: %q", c, got)
		}
	}
}

func TestCategoryNeedsAgent(t *testing.T) {
	if Lessons.NeedsAgent() {
		t.Error("Lessons should not need an agent")
	}
	for _, c := range []Category{DRY, PURE, Plan, Specs, MilestoneReview} {
		if !c.NeedsAgent() {
			t.Errorf("%s should need an agent", c)
		}
	}
}

func TestBuildPrompt_DRY(t *testing.T) {
	p := BuildPrompt(DRY, PromptInput{Diff: "DIFF_CONTENT"})
	for _, want := range []string{
		"DRY (Don't Repeat Yourself) violations",
		`No DRY violations found.`,
		"Do NOT modify any files",
		"DIFF_CONTENT",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q\n%s", want, p)
		}
	}
}

func TestBuildPrompt_Plan_ListsIssues(t *testing.T) {
	p := BuildPrompt(Plan, PromptInput{
		Diff:          "DIFF",
		ChangedIssues: []string{"workshop/issues/000031.md", "workshop/issues/000042.md"},
	})
	if !strings.Contains(p, "workshop/issues/000031.md\nworkshop/issues/000042.md") {
		t.Errorf("Plan prompt should list changed issues:\n%s", p)
	}
}

func TestBuildPrompt_Lessons_Empty(t *testing.T) {
	if got := BuildPrompt(Lessons, PromptInput{}); got != "" {
		t.Errorf("Lessons should produce empty prompt, got %q", got)
	}
}

// TestBuildPrompt_PlanQuality_HasContract pins the executability
// review's key invariants: the issue ref, the failure-modes list,
// the VERDICT line, and the issue content all reach the agent. Drift
// in any of these silently weakens the gate.
func TestBuildPrompt_PlanQuality_HasContract(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{
		IssueRef:     "ariadne#99",
		IssueContent: "ISSUE_FILE_BODY",
		PlanContent:  "SEPARATE_PLAN_BODY",
	})
	for _, want := range []string{
		"ariadne#99",
		"Is this plan executable as-written",
		"Vague checklist items",
		"Undeclared cross-issue",
		"Mismatched estimate vs scope",
		"VERDICT: CLEAN | INFO | FAILURE",
		"ISSUE_FILE_BODY",
		"SEPARATE_PLAN_BODY",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("PlanQuality prompt missing %q:\n%s", want, p)
		}
	}
}

// TestBuildPrompt_PlanQuality_NoSeparatePlan covers the common case
// where the plan lives inline in the issue file and there's no
// separate workshop/plans/ file. The prompt should still build, and
// the "Plan file" section should show a stand-in marker rather than
// an empty fenced block.
func TestBuildPrompt_PlanQuality_NoSeparatePlan(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{
		IssueRef:     "ariadne#100",
		IssueContent: "INLINE_PLAN_BODY",
	})
	if !strings.Contains(p, "(no separate plan file)") {
		t.Errorf("expected stand-in marker when PlanContent is empty:\n%s", p)
	}
	if !strings.Contains(p, "INLINE_PLAN_BODY") {
		t.Errorf("expected issue content in prompt:\n%s", p)
	}
}

// TestPlanQuality_RegisteredInCategories pins that the new category
// participates in AllCategories + IsValid — bulk-dispatchers iterate
// the list, and an omission would silently skip the gate.
func TestPlanQuality_RegisteredInCategories(t *testing.T) {
	if !IsValid(string(PlanQuality)) {
		t.Error("PlanQuality should be a valid category")
	}
	found := false
	for _, c := range AllCategories() {
		if c == PlanQuality {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("PlanQuality missing from AllCategories: %v", AllCategories())
	}
}

func TestBuildPrompt_MilestoneReview_HasContract(t *testing.T) {
	p := BuildPrompt(MilestoneReview, PromptInput{
		IssueRef: "ariadne#31 M3",
		Base:     "9e8625e",
		Head:     "d7789e0",
		Diff:     "DIFF",
	})
	for _, want := range []string{
		"ariadne#31 M3",
		"Base: 9e8625e",
		"Head: d7789e0",
		"Critical (must fix before next milestone)",
		"anti-collusion property",
		"Core concepts cross-check",
		"PURE: tests run without IO",
		"Atlas update gate",
		"Plan revision recommendations",
		"THE VERY FIRST LINE of your response MUST be",
		"SHIP | FIX-THEN-SHIP | REWORK",
		"Strengths:",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("milestone-review prompt missing %q", want)
		}
	}
}

func TestParseVerdict(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   Verdict
	}{
		{
			"ship plain",
			"SHIP (confidence: high)\n\nSummary: …",
			VerdictShip,
		},
		{
			"fix-then-ship",
			"FIX-THEN-SHIP (confidence: medium)\nbody …",
			VerdictFixThenShip,
		},
		{
			"rework low confidence",
			"REWORK (confidence: low)\n",
			VerdictRework,
		},
		{
			"leading blank lines",
			"\n\nSHIP (confidence: high)\n",
			VerdictShip,
		},
		{
			"indented verdict",
			"  SHIP (confidence: high)",
			VerdictShip,
		},
		{
			"no confidence parenthetical still parses",
			"REWORK\nfurther notes …",
			VerdictRework,
		},
		{
			"first non-empty line is prose, not a verdict",
			"Looks fine to me, no major findings.\nSHIP\n",
			VerdictUnknown,
		},
		{
			// The #56 M1 failure: a markdown title + `## Verdict` header
			// before the verdict. Structural lines are skipped.
			"markdown title and header before emphasized verdict",
			"# Post-Milestone Code Review — ariadne#56 M1\n\n## 1. Verdict\n\n**SHIP** (confidence: high)\n\nSummary…",
			VerdictShip,
		},
		{
			"emphasized verdict on first line",
			"**FIX-THEN-SHIP** (confidence: medium)\n",
			VerdictFixThenShip,
		},
		{
			"verdict written as a heading",
			"## REWORK (confidence: low)\n",
			VerdictRework,
		},
		{
			// Precision guard: a token at line-start in prose (followed by
			// more words, not a paren/EOL) is NOT a verdict.
			"ship-prefixed prose at line start is not a verdict",
			"SHIP-blocking issues remain in the parser.\n",
			VerdictUnknown,
		},
		{
			// The #56 M3 failure: the reviewer narrates investigation prose
			// before the verdict line. The confidence-qualified fallback
			// catches it even though the leading scan stops at the prose.
			"prose preamble then confidence-qualified verdict",
			"I have enough to render the verdict. Let me confirm one detail first.\n\nThe renderer emits deps from --deps only.\n\nFIX-THEN-SHIP (confidence: high)\n\nM3 is a docs milestone…",
			VerdictFixThenShip,
		},
		{
			// Precision still holds: prose preamble + a *bare* later token
			// (no confidence paren) is NOT a verdict.
			"prose preamble then bare token is not a verdict",
			"Looks reasonable overall.\n\nSHIP\n",
			VerdictUnknown,
		},
		{
			"completely empty",
			"",
			VerdictUnknown,
		},
		{
			"only whitespace",
			"   \n\t\n",
			VerdictUnknown,
		},
		{
			"ship-like prose without anchor",
			"This will ship after one tweak.\n",
			VerdictUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseVerdict(tt.output); got != tt.want {
				t.Errorf("ParseVerdict(%q) = %s, want %s", tt.output, got, tt.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   Outcome
	}{
		// Structured VERDICT line (preferred path; what Plan + Specs
		// prompts now ask subagents to emit). Tolerant on whitespace
		// and the optional confidence parenthetical.
		{"verdict clean", "VERDICT: CLEAN\n\nAll good.", Clean},
		{"verdict info", "VERDICT: INFO (confidence: high)\n\nMinor nits only.", Info},
		{"verdict failure", "VERDICT: FAILURE\n\nUnchecked-but-done items found.", Failure},
		{"verdict with leading whitespace", "   VERDICT: CLEAN\nbody", Clean},
		{"verdict after blank line", "\n\nVERDICT: CLEAN\nbody", Clean},

		// #70 regression: the verdict behind a PREAMBLE. These all returned
		// Failure before (the parser only checked the first non-empty line, so a
		// title / prose line dropped it to the legacy grep → Failure → blocked the
		// merge). The robust scan now finds the VERDICT: line anywhere.
		{"#70: clean behind a markdown title + NOTE", "# Code Review\n\nVERDICT: CLEAN\n\n[NOTE] a minor nit, non-blocking.", Clean},
		{"#70: clean behind prose preamble", "I've reviewed the diff in full.\n\nVERDICT: CLEAN\nfindings: none.", Clean},
		{"#70: failure behind a title still fails", "# Specs Review\nVERDICT: FAILURE\nstale atlas ref.", Failure},
		{"#70: info behind preamble passes", "## Plan Review\n\nVERDICT: INFO (confidence: high)\nminor suggestion only.", Info},
		// Cross-family tokens map through the contract (a SHIP-family token on a
		// VERDICT: line classifies by blocking-ness).
		{"ship token → non-blocking (info)", "VERDICT: SHIP (confidence: high)\n…", Info},
		{"fix-then-ship token → non-blocking", "VERDICT: FIX-THEN-SHIP\naddress then ship.", Info},
		{"rework token → blocking (failure)", "VERDICT: REWORK\nneeds rework.", Failure},
		{"block token → failure", "VERDICT: BLOCK\nhard stop.", Failure},

		// Real-world repro from pair#23 close: judge approved in prose
		// but the prompt didn't yet ask for a VERDICT line, so the
		// output starts with a markdown header. Falls through to the
		// legacy grep → no sentinel match → Failure. This case is the
		// motivation for the verdict-line migration; once a prompt is
		// migrated, the agent emits VERDICT and this path stops firing.
		{"legacy prose approval still falls through", "# TPM Review\n## Status: Looks good", Failure},

		// Legacy grep fallback path — unchanged contract.
		{"clean dry", "No DRY violations found.", Clean},
		{"clean pure", "No PURE violations found.", Clean},
		{"clean specs", "Everything is in sync.", Clean},
		{"clean plan", "No issue files changed", Clean},
		{"info lessons", LessonsReminder, Info},
		{"failure with content", "Found 2 violations:\n- foo.go:42 duplicates bar.go:99", Failure},
		{"failure empty", "", Failure},
		{"failure whitespace", "   \n\t\n", Failure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.output); got != tt.want {
				t.Errorf("Classify(%q) = %s, want %s", tt.output, got, tt.want)
			}
		})
	}
}

// #70: the robust scan finds the VERDICT: token anywhere, tolerating preamble +
// markdown markup; ok=false only when there's genuinely no VERDICT: line.
func TestParseVerdictToken(t *testing.T) {
	cases := []struct {
		name string
		in   string
		tok  string
		ok   bool
	}{
		{"first line", "VERDICT: CLEAN\nbody", "CLEAN", true},
		{"behind title", "# Review\n\nVERDICT: SHIP (confidence: high)\n…", "SHIP", true},
		{"behind prose", "I looked at everything.\nVERDICT: FAILURE\n…", "FAILURE", true},
		{"emphasized", "**VERDICT: REWORK**\n…", "REWORK", true},
		{"lowercase prefix", "verdict: clean\n", "CLEAN", true},
		{"fix-then-ship", "VERDICT: FIX-THEN-SHIP\n", "FIX-THEN-SHIP", true},
		{"no verdict line", "# Review\nLooks good to me.\n", "", false},
		{"bare token is not a VERDICT line", "SHIP (confidence: high)\n", "", false},
		// #70 M1 review (I1): a judge reviewing THIS parser quotes the contract;
		// a line that starts `VERDICT: <token>` then continues as PROSE must NOT
		// match (the trailing precision guard). The wrong-token-capture case
		// (`…CLEAN…actually FAILURE`) is the dangerous one — it would have
		// spuriously passed a gate.
		{"prose continuation after token rejected", "VERDICT: BLOCK is the generic hard block.\n", "", false},
		{"wrong-token prose quote rejected", "VERDICT: CLEAN means no issues, but actually FAILURE here.\n", "", false},
		{"mid-line quote already safe", "the VERDICT: CLEAN line is what the parser reads.\n", "", false},
		{"emphasized verdict still matches", "**VERDICT: SHIP**\n", "SHIP", true},
		{"confidence-qualified still matches", "VERDICT: REWORK (confidence: low)\n", "REWORK", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tok, ok := ParseVerdictToken(c.in)
			if tok != c.tok || ok != c.ok {
				t.Errorf("ParseVerdictToken(%q) = (%q,%v), want (%q,%v)", c.in, tok, ok, c.tok, c.ok)
			}
		})
	}
}

// #70: milestone ParseVerdict now accepts the unified `VERDICT:` prefix (M2's
// migrated prompt) AND the legacy bare token (back-compat), including behind a
// preamble — which is what fixes the `unknown` verdict seen on #68's M2 review.
func TestParseVerdict_VerdictPrefix(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Verdict
	}{
		{"verdict prefix ship", "VERDICT: SHIP (confidence: high)\n…", VerdictShip},
		{"verdict prefix behind title", "# Post-Milestone Review\n\nVERDICT: FIX-THEN-SHIP\n…", VerdictFixThenShip},
		{"verdict prefix behind prose (the #68 unknown case)", "I have everything I need.\nVERDICT: REWORK\nreasons…", VerdictRework},
		{"legacy bare token still parses", "SHIP (confidence: high)\nbody", VerdictShip},
		{"no verdict at all → unknown", "Looks reasonable.\nNo verdict emitted.", VerdictUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ParseVerdict(c.in); got != c.want {
				t.Errorf("ParseVerdict(%q) = %s, want %s", c.in, got, c.want)
			}
		})
	}
}

func TestBuildArgs_Claude(t *testing.T) {
	name, args, err := BuildArgs(DispatchOptions{
		Agent:        AgentClaude,
		Prompt:       "review this",
		AllowedTools: "Read,Grep",
	})
	if err != nil {
		t.Fatal(err)
	}
	if name != "claude" {
		t.Errorf("name = %q want claude", name)
	}
	want := []string{"-p", "--allowedTools", "Read,Grep", "--permission-mode", "bypassPermissions", "review this"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Errorf("args = %v\nwant: %v", args, want)
	}
}

func TestBuildArgs_Codex_SandboxAddsFullAuto(t *testing.T) {
	_, args, _ := BuildArgs(DispatchOptions{Agent: AgentCodex, Prompt: "p", IsSandbox: true})
	if args[1] != "--full-auto" {
		t.Errorf("expected --full-auto in sandbox mode, got %v", args)
	}
}

func TestBuildArgs_Gemini_NoSandboxFlag(t *testing.T) {
	_, args, _ := BuildArgs(DispatchOptions{Agent: AgentGemini, Prompt: "p"})
	for _, a := range args {
		if a == "--yolo" {
			t.Errorf("--yolo should not be present without sandbox flag: %v", args)
		}
	}
}

func TestBuildArgs_UnknownAgent(t *testing.T) {
	_, _, err := BuildArgs(DispatchOptions{Agent: "bogus", Prompt: "p"})
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}

func TestBuildArgs_DefaultIsClaude(t *testing.T) {
	name, _, err := BuildArgs(DispatchOptions{Prompt: "p", AllowedTools: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	if name != "claude" {
		t.Errorf("empty Agent should default to claude, got %q", name)
	}
}

func TestFormatCommandLine_Quoting(t *testing.T) {
	cmd, err := FormatCommandLine(DispatchOptions{
		Agent:        AgentClaude,
		Prompt:       "this has spaces and 'quotes'",
		AllowedTools: "Read",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cmd, `'this has spaces and '\''quotes'\'''`) {
		t.Errorf("prompt not properly shell-quoted:\n%s", cmd)
	}
}

func TestDispatch_FakeRun_CapturesOutput(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	var gotName string
	var gotArgs []string
	Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = args
		return []byte("No DRY violations found.\n"), nil
	}
	out, err := Dispatch(context.Background(), DispatchOptions{
		Agent:        AgentClaude,
		Prompt:       "review",
		AllowedTools: "Read",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "claude" {
		t.Errorf("name = %q", gotName)
	}
	if gotArgs[len(gotArgs)-1] != "review" {
		t.Errorf("last arg should be the prompt, got %v", gotArgs)
	}
	if Classify(out) != Clean {
		t.Errorf("output should classify as Clean, got %s", Classify(out))
	}
}

func TestDispatch_LaunchError_Surfaces(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("exec: command not found")
	}
	_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "x", AllowedTools: "Read"})
	if err == nil {
		t.Error("expected error when Run fails to launch")
	}
}

// Regression for M3 review I3: non-zero exit from the subprocess (whether
// with or without output) should NOT be a Dispatch error — it's a finding
// for Classify to interpret. Real launch failures still error.
func TestDispatch_ExitErrorWithEmptyOutput_NotAnError(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	// Real *exec.ExitError requires a started process. Easiest path:
	// spawn `false` (always exits 1) via the actual exec package, no
	// args needed.
	Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return realExec("false")
	}
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "x", AllowedTools: "Read"})
	if err != nil {
		t.Errorf("non-zero exit should not surface as Dispatch error, got %v", err)
	}
	if Classify(out) != Failure {
		t.Errorf("empty output should classify as Failure, got %s", Classify(out))
	}
}

// realExec runs `false` (or any always-exit-non-zero binary) so we get a
// genuine *exec.ExitError. Wrapped here to keep the test's intent clear.
func realExec(name string) ([]byte, error) {
	return execCommand(name).CombinedOutput()
}
