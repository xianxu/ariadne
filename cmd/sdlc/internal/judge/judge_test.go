package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
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

func TestBoundaryReviewInspectionContract(t *testing.T) {
	p := BuildPrompt(MilestoneReview, PromptInput{ReviewWindow: "PINNED-MANIFEST"})
	normalized := strings.Join(strings.Fields(p), " ")
	for _, want := range []string{
		"PINNED-MANIFEST",
		"Run at least the stat and name-status recipes",
		"return REWORK",
		"repository, pinned objects, or a required read-only command is unavailable",
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("boundary review inspection contract missing %q:\n%s", want, p)
		}
	}
	if strings.Contains(p, "Read the diff against") {
		t.Error("boundary review procedure still assumes patch bytes are embedded")
	}
}

// TestEstimateQuality_NotInBulkDispatch pins the #117 decision: the estimate-
// quality judge is a change-code-time-only gate. It must stay out of
// AllCategories() (which drives push/merge bulk dispatch) and IsValid (standalone
// `sdlc judge` invocation), so it can never silently run at the wrong boundary.
func TestEstimateQuality_NotInBulkDispatch(t *testing.T) {
	for _, c := range AllCategories() {
		if c == EstimateQuality {
			t.Fatal("EstimateQuality must NOT be in AllCategories() — it would enroll in push/merge bulk dispatch (#117)")
		}
	}
	if IsValid(string(EstimateQuality)) {
		t.Error("EstimateQuality should not be standalone-valid (change-code-time only)")
	}
}

func TestBuildPrompt_EstimateQuality(t *testing.T) {
	p := BuildPrompt(EstimateQuality, PromptInput{IssueRef: "ariadne#117", IssueContent: "## Estimate\nblock"})
	for _, want := range []string{"ariadne#117", "estimate-logic-v2", "fabricated", ContractPreamble} {
		if !strings.Contains(p, want) {
			t.Errorf("EstimateQuality prompt missing %q", want)
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
		"ARCH-DRY",              // renders the principle from the registry (#75)
		"Don't Repeat Yourself", // from the embedded architecture.md
		"Do NOT modify any files",
		"DIFF_CONTENT",
		"CLEAN   = no ARCH-DRY violations.",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q\n%s", want, p)
		}
	}
}

// #75: architecture.md is the single source — it carries both markers and both
// lenses, and is embedded verbatim into every prompt that needs it.
func TestArchitectureRegistry_Content(t *testing.T) {
	for _, want := range []string{"ARCH-DRY", "ARCH-PURE", "ARCH-PURPOSE", "ARCH-MOCK", "ARCH-CONSTRAINTS", "at-plan", "at-review", "principle:"} {
		if !strings.Contains(ArchitectureRegistry, want) {
			t.Errorf("ArchitectureRegistry missing %q", want)
		}
	}
}

type architectureClauseContract struct {
	label     string
	canonical string
	required  []string
}

var constraintsClauseContracts = []architectureClauseContract{
	{
		label:     "principle",
		canonical: "Runtime behavior is part of the architecture. Before choosing a mechanism, identify the small set of external constraints that can materially shape it: latency, workload/input scale and growth, CPU, memory, disk/network IO, concurrency, target environment and co-tenancy, and overload behavior. Make consequential expectations explicit instead of leaving them as hidden assumptions.",
		required: []string{
			"Runtime behavior is part of the architecture.",
			"identify the small set of external constraints",
			"latency, workload/input scale and growth, CPU, memory, disk/network IO, concurrency, target environment and co-tenancy, and overload behavior",
			"Make consequential expectations explicit",
		},
	},
	{
		label:     "at-plan",
		canonical: "Classify the workload and interaction path (for example keystroke, UI response, startup/shutdown, online request, batch, or training/inference), then give each relevant constraint a budget/range, basis (measured fact, requirement, domain-informed assumption, or operator choice), and bounded behavior when exceeded. Mark irrelevant categories `N/A`; do not fill a ceremonial checklist or invent universal defaults. Make an educated initial estimate when useful, but confirm material uncertainty with the operator.",
		required: []string{
			"Classify the workload and interaction path",
			"keystroke, UI response, startup/shutdown, online request, batch, or training/inference",
			"budget/range, basis (measured fact, requirement, domain-informed assumption, or operator choice), and bounded behavior when exceeded",
			"Mark irrelevant categories `N/A`",
			"do not fill a ceremonial checklist or invent universal defaults",
			"confirm material uncertainty with the operator",
		},
	},
	{
		label:     "at-review",
		canonical: "Check that the implementation enforces the declared operating envelope and that representative measurements or tests exercise the relevant environment and workload. Flag blocking optional work on a critical UI path, unbounded concurrency or fan-out, repeated expensive work that should be cached or incremental, resource monopolization, unsupported performance claims, and behavior that silently operates outside the stated bounds.",
		required: []string{
			"implementation enforces the declared operating envelope",
			"representative measurements or tests exercise the relevant environment and workload",
			"blocking optional work on a critical UI path",
			"unbounded concurrency or fan-out",
			"repeated expensive work that should be cached or incremental",
			"resource monopolization",
			"unsupported performance claims",
			"silently operates outside the stated bounds",
		},
	},
}

func architectureEntry(registry, marker string) (string, error) {
	heading := "## " + marker
	if strings.Count(registry, heading) != 1 {
		return "", errors.New("architecture marker heading must occur exactly once")
	}
	start := strings.Index(registry, heading)
	entry := registry[start:]
	if next := strings.Index(entry[len(heading):], "\n## ARCH-"); next >= 0 {
		entry = entry[:len(heading)+next]
	}
	return entry, nil
}

func architectureClause(entry, label string) (string, error) {
	prefix := "- **" + label + ":**"
	if strings.Count(entry, prefix) != 1 {
		return "", errors.New("architecture clause label must occur exactly once")
	}
	start := strings.Index(entry, prefix)
	clause := entry[start:]
	if next := strings.Index(clause[len(prefix):], "\n- **"); next >= 0 {
		clause = clause[:len(prefix)+next]
	}
	return strings.Join(strings.Fields(clause), " "), nil
}

func constraintsContractViolations(registry string) []string {
	entry, err := architectureEntry(registry, "ARCH-CONSTRAINTS")
	if err != nil {
		return []string{err.Error()}
	}
	var violations []string
	for _, contract := range constraintsClauseContracts {
		clause, err := architectureClause(entry, contract.label)
		if err != nil {
			violations = append(violations, contract.label+": "+err.Error())
			continue
		}
		expected := strings.Join(strings.Fields("- **"+contract.label+":** "+contract.canonical), " ")
		if clause != expected {
			violations = append(violations, contract.label+": clause is not canonical affirmative text")
		}
		for _, required := range contract.required {
			if !strings.Contains(clause, required) {
				violations = append(violations, contract.label+": missing "+required)
			}
		}
	}
	return violations
}

func validConstraintsRegistryForTest() string {
	var b strings.Builder
	b.WriteString("## ARCH-CONSTRAINTS — fixture\n\n")
	for _, contract := range constraintsClauseContracts {
		fmt.Fprintf(&b, "- **%s:** %s\n", contract.label, contract.canonical)
	}
	return b.String()
}

func TestArchitectureRegistry_ConstraintsContract(t *testing.T) {
	if violations := constraintsContractViolations(ArchitectureRegistry); len(violations) > 0 {
		t.Fatalf("ARCH-CONSTRAINTS contract violations:\n  %s", strings.Join(violations, "\n  "))
	}
}

func TestArchitectureRegistry_ConstraintsContractMutants(t *testing.T) {
	valid := validConstraintsRegistryForTest()
	atPlanPredicate := constraintsClauseContracts[1].required[0]
	mutants := map[string]string{
		"deleted predicate": strings.Replace(valid, atPlanPredicate, "", 1),
		"predicate moved to principle": strings.Replace(
			strings.Replace(valid, atPlanPredicate, "", 1),
			"- **principle:**", "- **principle:** "+atPlanPredicate, 1),
		"predicate negated": strings.Replace(valid, atPlanPredicate, "Do not classify the workload and interaction path", 1),
	}
	for name, mutant := range mutants {
		t.Run(name, func(t *testing.T) {
			if violations := constraintsContractViolations(mutant); len(violations) == 0 {
				t.Fatal("mutant unexpectedly satisfies ARCH-CONSTRAINTS contract")
			}
		})
	}
}

func TestArchitectureRegistry_ConstraintsContractRejectsNegatedPredicates(t *testing.T) {
	valid := validConstraintsRegistryForTest()
	for _, contract := range constraintsClauseContracts {
		for _, required := range contract.required {
			t.Run(contract.label+"/"+required, func(t *testing.T) {
				mutant := strings.Replace(valid, required, "Do not "+required, 1)
				if violations := constraintsContractViolations(mutant); len(violations) == 0 {
					t.Fatal("case-preserving negation unexpectedly satisfies ARCH-CONSTRAINTS contract")
				}
			})
		}
	}
}

func TestArchitectureRegistry_ConstraintsContractRejectsSeparatedNegatedPredicates(t *testing.T) {
	valid := validConstraintsRegistryForTest()
	for _, contract := range constraintsClauseContracts {
		for _, required := range contract.required {
			t.Run(contract.label+"/"+required, func(t *testing.T) {
				mutant := strings.Replace(valid, required, "Do not ever "+required, 1)
				if violations := constraintsContractViolations(mutant); len(violations) == 0 {
					t.Fatal("semantically negated predicate unexpectedly satisfies ARCH-CONSTRAINTS contract")
				}
			})
		}
	}
}

func TestArchitectureRegistry_ConstraintsStructureFailsClosed(t *testing.T) {
	valid := validConstraintsRegistryForTest()
	mutants := map[string]string{
		"missing entry":    strings.Replace(valid, "ARCH-CONSTRAINTS", "ARCH-OTHER", 1),
		"duplicate entry":  valid + valid,
		"missing clause":   strings.Replace(valid, "- **at-review:**", "- **review:**", 1),
		"duplicate clause": valid + "- **at-plan:** duplicate\n",
	}
	for name, mutant := range mutants {
		t.Run(name, func(t *testing.T) {
			if violations := constraintsContractViolations(mutant); len(violations) == 0 {
				t.Fatal("malformed registry unexpectedly satisfies ARCH-CONSTRAINTS contract")
			}
		})
	}
}

// #75: the registry is delivered into all four architecture-aware prompts —
// plan-quality (at-plan) + milestone-review/dry/pure (at-review). Editing the one
// file updates them all; a missing embed is silent architectural drift.
func TestArchitectureRegistry_EmbeddedInPrompts(t *testing.T) {
	in := PromptInput{Diff: "D", IssueRef: "r#1", IssueContent: "I", Base: "a", Head: "b"}
	for _, c := range []Category{PlanQuality, MilestoneReview, DRY, PURE} {
		if !strings.Contains(BuildPrompt(c, in), ArchitectureRegistry) {
			t.Errorf("%s prompt does not embed ArchitectureRegistry (#75)", c)
		}
	}
	// Negative: prompts NOT wired for architecture must not carry the block — an
	// accidental future embed into the wrong prompt is caught.
	for _, c := range []Category{Plan, Specs} {
		if strings.Contains(BuildPrompt(c, in), ArchitectureRegistry) {
			t.Errorf("%s prompt should NOT embed ArchitectureRegistry (only the 4 architecture-aware prompts)", c)
		}
	}
	// Lens labels reach the right consumers.
	if !strings.Contains(BuildPrompt(PlanQuality, in), "at-plan") {
		t.Error("plan-quality should render the at-plan lens")
	}
	if !strings.Contains(BuildPrompt(MilestoneReview, in), "at-review") {
		t.Error("milestone-review should render the at-review lens")
	}
}

// #69: ArchitectureMarkers is the single extraction site for ARCH-* names —
// shared by the {{ARCH_STAR}} substitution and the AGENTS.md drift guard.
func TestArchitectureMarkers(t *testing.T) {
	markers := ArchitectureMarkers()
	want := []string{"ARCH-DRY", "ARCH-PURE", "ARCH-PURPOSE", "ARCH-MOCK", "ARCH-CONSTRAINTS"}
	if len(markers) != len(want) {
		t.Fatalf("ArchitectureMarkers() = %v, want %v", markers, want)
	}
	for i, w := range want {
		if markers[i] != w {
			t.Errorf("marker[%d] = %q, want %q (registry order)", i, markers[i], w)
		}
	}
}

// #69: code-review.md is the one embedded boundary-review procedure; CodeReviewBody
// substitutes the window fields and expands {{ARCH_STAR}} from the live registry.
func TestCodeReviewBody_Renders(t *testing.T) {
	if strings.TrimSpace(codeReviewTemplate) == "" {
		t.Fatal("code-review.md embed is empty")
	}
	body := CodeReviewBody(PromptInput{
		IssueRef: "pair#72 M1", Base: "BASE_SHA", Head: "HEAD_SHA",
		Repo: "pair", RepoRoot: "/w/pair", IssueFile: "workshop/issues/000072-x.md",
		Boundary: "milestone M1 close",
		RepoNote: "a downstream repo built on the ariadne base layer",
	})
	for _, want := range []string{
		"pair#72 M1",                  // {{ISSUE_REF}} — repo-prefixed, not hardcoded ariadne (#137)
		"Base: BASE_SHA",              // {{BASE}}
		"Head: HEAD_SHA",              // {{HEAD}}
		"pair",                        // {{REPO}}
		"/w/pair",                     // {{REPO_ROOT}}
		"workshop/issues/000072-x.md", // {{ISSUE_FILE}}
		"milestone M1 close",          // {{BOUNDARY}}
		"downstream repo",             // {{REPO_NOTE}}
		"ARCH-DRY, ARCH-PURE, ARCH-PURPOSE, ARCH-MOCK, ARCH-CONSTRAINTS", // {{ARCH_STAR}} enumerated from the registry (full set, not a substring — asserts the consumer derives the new marker)
		"Core concepts cross-check",
		"```verdict",                               // {{VERDICT_BLOCK}} — the structured handoff (#147)
		"verdict: <SHIP | FIX-THEN-SHIP | REWORK>", // tokens rendered from vocab.Verdict().Emitted()
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered body missing %q", want)
		}
	}
	// No placeholder survives the render.
	if strings.Contains(body, "{{") {
		t.Errorf("unsubstituted placeholder remains in rendered body:\n%s", body)
	}
}

// #69 guardrail: the procedure must CITE the markers, not re-inline the principle
// bodies — those live once in architecture.md and arrive co-present via
// ArchitectureBlock. If a principle's defining phrase leaks into code-review.md,
// the registry has stopped being the sole definition site (ARCH-DRY).
func TestCodeReviewTemplate_DoesNotInlinePrincipleBodies(t *testing.T) {
	for _, body := range []string{
		"Reuse before adding", // ARCH-DRY principle line
		"One source of truth", // ARCH-DRY principle line
		`thin "glue" layer`,   // ARCH-PURE principle line
	} {
		if strings.Contains(codeReviewTemplate, body) {
			t.Errorf("code-review.md inlines registry principle text %q — cite the marker instead", body)
		}
	}
	// It must still reference the markers via the substitution token.
	if !strings.Contains(codeReviewTemplate, "{{ARCH_STAR}}") {
		t.Error("code-review.md should reference ARCH-* via the {{ARCH_STAR}} token")
	}
}

// TestContractDoc_InSyncWithTokens is the #70 M2 drift guard: the human schema
// doc (construct/judge-output-contract.md) must list exactly the tokens the Go
// source of truth (ContractTokens) defines. A token added to one and not the
// other silently diverges the "both reference one contract" promise.
func TestContractDoc_InSyncWithTokens(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "construct", "judge-output-contract.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract doc %s: %v", path, err)
	}
	doc := string(data)
	known := map[string]bool{}
	for _, tok := range ContractTokens {
		known[tok] = true
		if !strings.Contains(doc, "`"+tok+"`") {
			t.Errorf("contract doc missing token `%s` (drift from contract.go ContractTokens)", tok)
		}
	}
	// Reverse direction: every token the doc's table row leads with (`| `TOKEN` |`)
	// must be a real ContractToken — so a stray doc token also fails.
	for _, line := range strings.Split(doc, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "| `") {
			continue
		}
		rest := line[strings.Index(line, "`")+1:]
		tok := rest[:strings.Index(rest, "`")]
		if tok != "" && !known[tok] {
			t.Errorf("contract doc table lists `%s`, not in contract.go ContractTokens (drift)", tok)
		}
	}
	if !strings.Contains(doc, "VERDICT: <TOKEN>") {
		t.Error("contract doc missing the `VERDICT: <TOKEN>` format line both sides depend on")
	}
}

// #128 drift guard (was #75's per-marker enumeration check): the ARCH-* definitions
// no longer live in AGENTS.md's narrative — #128 single-sourced them behind
// `sdlc arch-principles` so the constitution stops RESTATING (and silently drifting
// from) the registry. The narrative's remaining contract is to ROUTE there + keep
// marker awareness, so this guards both survive edits. The marker SET stays guarded
// by TestArchitectureMarkers + the command's own test (cmd/sdlc
// TestRunArchPrinciples_RendersRegistry asserts the command derives every marker from
// architecture.md) — that's where "every consumer derives" is now enforced.
func TestArchitecture_NarrativeRoutesToArchPrinciples(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	agents := string(data)
	if !strings.Contains(agents, "sdlc arch-principles") {
		t.Error("AGENTS.md Core Design Principles should route to `sdlc arch-principles` (the single source for ARCH-*)")
	}
	// Marker awareness stays in the constitution (so non-gate work knows ARCH-*
	// exist and to cite them) even though the definitions moved to the command.
	if !strings.Contains(agents, "ARCH-") {
		t.Error("AGENTS.md should retain ARCH-* marker awareness + the cite-the-marker instruction")
	}
}

// TestAgentPromptsEmbedContract pins the #70 M2 unification: every
// agent-emitting category embeds the one ContractPreamble verbatim, so the
// output format is a single source of truth (no per-prompt paraphrase to drift).
func TestAgentPromptsEmbedContract(t *testing.T) {
	in := PromptInput{Diff: "D", IssueRef: "r#1", IssueContent: "I", Base: "a", Head: "b",
		ChangedIssues: []string{"workshop/issues/000001.md"}}
	for _, c := range AllCategories() {
		if !c.NeedsAgent() {
			continue // Lessons is the documented REMINDER: exception
		}
		want := ContractPreamble
		if c == MilestoneReview {
			want = BoundaryReviewContract // #147: block-first contract for the boundary review
		}
		if !strings.Contains(BuildPrompt(c, in), want) {
			t.Errorf("%s prompt does not embed its output contract (verdict format drift)", c)
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
		// #175 forward fix: flag over-split Mx plans at design time.
		"Over-split milestones",
		"review boundary",
		"plain checkboxes",
		"VERDICT: <TOKEN>", // the shared contract format (#70 M2)
		"CLEAN   = plan is concrete",
		"ISSUE_FILE_BODY",
		"SEPARATE_PLAN_BODY",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("PlanQuality prompt missing %q:\n%s", want, p)
		}
	}

	// #187 B1: the estimate gates now run AFTER plan-quality, so at this point the
	// estimate legitimately does not exist yet. A "mismatched estimate vs scope" ask
	// here would demand the agent cost a plan nobody has accepted — the waste-by-
	// construction B2 removes. This assertion replaces the old one that required it.
	for _, forbidden := range []string{"Mismatched estimate vs scope", "estimate_hours"} {
		if strings.Contains(p, forbidden) {
			t.Errorf("PlanQuality prompt must not mention the estimate (#187 B1), found %q", forbidden)
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
		"Critical (must fix before crossing the boundary)",
		"anti-collusion property",
		"Core concepts cross-check",
		"PURE: tests run without IO",
		"Docs update gate",
		"Plan revision recommendations",
		"VERDICT: <TOKEN>",                     // unified contract format (#70 M2 — was bare "SHIP | …")
		"```verdict",                           // the authoritative structured handoff block (#147)
		"FIX-THEN-SHIP  ship after addressing", // token gloss rendered from the model's `when`
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

func TestResolveAgentCLI_Precedence(t *testing.T) {
	tests := []struct {
		name        string
		explicit    string
		explicitSet bool
		env         AgentDefaultEnv
		want        AgentCLI
	}{
		{
			name:        "explicit claude beats pair codex",
			explicit:    "claude",
			explicitSet: true,
			env:         AgentDefaultEnv{PairAgent: "codex"},
			want:        AgentClaude,
		},
		{
			name:        "explicit bogus remains bogus for dispatch validation",
			explicit:    "bogus",
			explicitSet: true,
			env:         AgentDefaultEnv{PairAgent: "codex"},
			want:        AgentCLI("bogus"),
		},
		{
			name: "agent cmd gemini beats pair codex",
			env:  AgentDefaultEnv{AgentCmd: "gemini", PairAgent: "codex"},
			want: AgentGemini,
		},
		{
			name: "agent cmd bogus remains bogus for dispatch validation",
			env:  AgentDefaultEnv{AgentCmd: "bogus", PairAgent: "codex"},
			want: AgentCLI("bogus"),
		},
		{
			name: "pair codex selects codex",
			env:  AgentDefaultEnv{PairAgent: "codex"},
			want: AgentCodex,
		},
		{
			name: "codex ci signal selects codex",
			env:  AgentDefaultEnv{CodexCI: "1"},
			want: AgentCodex,
		},
		{
			name: "codex thread signal selects codex",
			env:  AgentDefaultEnv{CodexThreadID: "019f"},
			want: AgentCodex,
		},
		{
			name: "unknown pair falls through to codex signal",
			env:  AgentDefaultEnv{PairAgent: "unknown", CodexCI: "1"},
			want: AgentCodex,
		},
		{
			name: "empty env falls back to claude",
			env:  AgentDefaultEnv{},
			want: AgentClaude,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAgentCLI(tt.explicit, tt.explicitSet, tt.env)
			if got != tt.want {
				t.Errorf("ResolveAgentCLI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCurrentAgentDefaultEnv_ReadsProcessSignals(t *testing.T) {
	t.Setenv("AGENT_CMD", "gemini")
	t.Setenv("PAIR_AGENT", "codex")
	t.Setenv("CODEX_CI", "1")
	t.Setenv("CODEX_THREAD_ID", "thread")
	t.Setenv("CLAUDECODE", "1")

	got := CurrentAgentDefaultEnv()
	if got.AgentCmd != "gemini" {
		t.Errorf("AgentCmd = %q", got.AgentCmd)
	}
	if got.PairAgent != "codex" {
		t.Errorf("PairAgent = %q", got.PairAgent)
	}
	if got.CodexCI != "1" {
		t.Errorf("CodexCI = %q", got.CodexCI)
	}
	if got.CodexThreadID != "thread" {
		t.Errorf("CodexThreadID = %q", got.CodexThreadID)
	}
	if got.ClaudeCode != "1" {
		t.Errorf("ClaudeCode = %q", got.ClaudeCode)
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
	for _, tc := range []struct {
		agent AgentCLI
		name  string
	}{
		{AgentClaude, "claude"},
		{AgentCodex, "codex"},
		{AgentGemini, "gemini"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			orig := Run
			defer func() { Run = orig }()
			var gotName string
			var gotArgs []string
			Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
				gotName = name
				gotArgs = args
				return ProcessOutput{
					Stdout: []byte("No DRY violations found.\n"),
					Stderr: []byte("diagnostic from " + name + "\n"),
				}, nil
			}
			var diagnostics bytes.Buffer
			out, err := Dispatch(context.Background(), DispatchOptions{
				Agent: tc.agent, Prompt: "review", AllowedTools: "Read", Stderr: &diagnostics,
			})
			if err != nil {
				t.Fatal(err)
			}
			if gotName != tc.name {
				t.Errorf("name = %q, want %q", gotName, tc.name)
			}
			if gotArgs[len(gotArgs)-1] != "review" {
				t.Errorf("last arg should be the prompt, got %v", gotArgs)
			}
			if Classify(out) != Clean {
				t.Errorf("output should classify as Clean, got %s", Classify(out))
			}
			if strings.Contains(out, "diagnostic") {
				t.Errorf("semantic output contains stderr: %q", out)
			}
			if got := diagnostics.String(); got != "diagnostic from "+tc.name+"\n" {
				t.Errorf("diagnostic sink = %q", got)
			}
		})
	}
}

func TestDispatch_LaunchError_Surfaces(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
		return ProcessOutput{Stderr: []byte("launch diagnostic\n")}, errors.New("exec: command not found")
	}
	var diagnostics bytes.Buffer
	_, err := Dispatch(context.Background(), DispatchOptions{
		Agent: AgentClaude, Prompt: "x", AllowedTools: "Read", Stderr: &diagnostics,
	})
	if err == nil {
		t.Fatal("expected error when Run fails to launch")
	}
	// #138: the diagnostic names the attempted agent + the owner bin/ on PATH, so
	// a launch failure is debuggable from the error string alone.
	for _, want := range []string{"claude", "owner bin", "PATH="} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("launch error missing %q: %v", want, err)
		}
	}
	if got := diagnostics.String(); got != "launch diagnostic\n" {
		t.Errorf("launch stderr = %q", got)
	}
}

func TestDispatch_ExitErrorForwardsDiagnosticsAndReturnsStdout(t *testing.T) {
	orig := Run
	defer func() { Run = orig }()
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
		return ProcessOutput{
			Stdout: []byte("VERDICT: FAILURE (confidence: high)\n"),
			Stderr: []byte("non-zero diagnostic\n"),
		}, &exec.ExitError{}
	}
	var diagnostics bytes.Buffer
	out, err := Dispatch(context.Background(), DispatchOptions{
		Agent: AgentClaude, Prompt: "x", AllowedTools: "Read", Stderr: &diagnostics,
	})
	if err != nil {
		t.Fatalf("non-zero exit surfaced as Dispatch error: %v", err)
	}
	if out != "VERDICT: FAILURE (confidence: high)\n" {
		t.Errorf("semantic output = %q", out)
	}
	if got := diagnostics.String(); got != "non-zero diagnostic\n" {
		t.Errorf("non-zero stderr = %q", got)
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
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
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
func realExec(name string) (ProcessOutput, error) {
	cmd := execCommand(name)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return ProcessOutput{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, err
}

// TestCodeReviewSeveritiesMatchModel pins code-review.md's severity buckets to the
// `finding` model (#187): the boundary review and the plan gate share ONE taxonomy, so a
// severity renamed in finding.cue cannot leave the boundary-review prose behind.
//
// The `s + " ("` shape matches code-review.md's bucket headings ("Critical (must fix
// before crossing the boundary)"), so this fails on a rename rather than merely on the
// word appearing somewhere in the prose.
func TestCodeReviewSeveritiesMatchModel(t *testing.T) {
	body := CodeReviewBody(PromptInput{})
	for _, s := range vocab.Finding().Severities() {
		if !strings.Contains(body, s+" (") {
			t.Errorf("code-review.md does not name severity %q from the finding model", s)
		}
	}
}

// TestPlanQualityPromptRendersFindingModel pins that the plan-quality prompt renders its
// accepted severity/disposition set FROM the model (#187), so finding.cue stays the single
// source and the prompt cannot drift from what the parser will accept.
func TestPlanQualityPromptRendersFindingModel(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{})
	m := vocab.Finding()
	for _, s := range m.Severities() {
		if !strings.Contains(p, s) {
			t.Errorf("plan-quality prompt omits severity %q", s)
		}
	}
	// AllDispositions(), not .Dispositions — the latter is map[string][]string since the
	// closing/open partition landed.
	for _, d := range m.AllDispositions() {
		if !strings.Contains(p, d) {
			t.Errorf("plan-quality prompt omits disposition %q", d)
		}
	}
	if !strings.Contains(p, "```findings") {
		t.Error("plan-quality prompt must show the fenced findings block")
	}
}

// TestPlanQualityPromptCarriesPriorFindings pins A2's actual mechanism: the prior-round
// block must reach the rendered prompt. Without this the gate is stateless no matter what
// the ledger records.
func TestPlanQualityPromptCarriesPriorFindings(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{PriorFindings: "SENTINEL-PRIOR-BLOCK"})
	if !strings.Contains(p, "SENTINEL-PRIOR-BLOCK") {
		t.Error("PriorFindings did not reach the rendered plan-quality prompt")
	}
	// Round 1 must announce itself rather than rendering an empty section.
	first := BuildPrompt(PlanQuality, PromptInput{})
	if !strings.Contains(first, "(no prior rounds)") {
		t.Error("an empty PriorFindings should render the explicit no-prior-rounds default")
	}
}

// TestPlanQualityPromptDemandsStrategyNotEnumeration pins C1's semantics in the prompt
// text: the gate must ask for test FUNCTIONS + strategy and explicitly reject enumerated
// case lists. This is the plumbing check; the behavioral check is the pair#127 replay.
func TestPlanQualityPromptDemandsStrategyNotEnumeration(t *testing.T) {
	p := BuildPrompt(PlanQuality, PromptInput{})
	for _, want := range []string{
		"functions", "one line of strategy", "fuzz it",
		"enumerated list of test cases", "line-numbered inventory",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("plan-quality prompt missing the C1 ask %q", want)
		}
	}
}

// TestCodeReviewCarriesPlanGateForward pins #187 A3's safety argument. The plan gate stops
// blocking on Minor and round-cap-demoted findings ONLY because the boundary review picks
// them up from the ledger. That claim is made in Decide's doc comment, the plan-quality
// prompt, and the change-code helptext — if this pointer were ever dropped, all three
// would become false and the demotion would become a silent loss.
// #194 M2 moved the MECHANISM without weakening the invariant. The reviewer no longer
// reads `-plan-gate.md` off disk (two id namespaces, one output fence, no rule for
// disposing an id this ledger never issued); instead the still-open plan-gate findings
// are SEEDED into the boundary ledger and arrive through the same prior-findings block
// as everything else — see persistBoundaryRound / seedFromPlanGate, and
// TestBoundaryReview_SeedsDeferredPlanGateFindings which pins the seeding itself.
//
// What this test still guards is the prompt half: the reviewer must be ASKED to dispose
// of prior findings. Without that instruction the seeded findings arrive and are ignored,
// and the plan gate's demotion becomes the silent loss it was never allowed to be.
func TestBoundaryReviewIsAskedToDisposePriorFindings(t *testing.T) {
	in := goldenInput
	in.PriorFindings = "### BR-7 (Minor, round 1)\ncarried from plan-quality PQ-3"
	prompt := BuildPrompt(MilestoneReview, in)
	for _, want := range []string{"BR-7", "carried from plan-quality PQ-3", "dispose", "```findings"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("the boundary reviewer must be shown prior findings and asked to dispose them; missing %q", want)
		}
	}
}
