// Package judge wraps the "fresh-context LLM check against a diff"
// pattern that ariadne has historically run as `scripts/pre-merge-checks.sh`.
//
// The package provides:
//
//   - Categories — the named principle/sanity checks (dry, pure, plan,
//     specs, lessons) plus milestone-review for the post-milestone
//     fresh-eyes pass per AGENTS.md §3.
//   - Prompt construction per category, ported byte-faithfully from the
//     shell's build_prompt heredocs.
//   - Output classification (clean / info / failure) ported from
//     scripts/lib.sh's is_clean_check_output / is_info_check_output.
//   - Subprocess dispatch via an agent CLI (claude, codex, gemini).
//     The Run shim lets tests inject fakes; production execs the binary.
//
// Anti-collusion property (per pensive): every Run call spawns a fresh
// subprocess with no inherited session state. The doer cannot rationalize
// its own work; the judge sees only the diff + prompt.
package judge

import (
	"fmt"
	"strings"
)

// Category enumerates the supported judge checks. Names match the
// shell's CHECK_NAMES array verbatim so `make check-dry` and
// `sdlc judge dry` invoke the same prompt.
type Category string

const (
	DRY             Category = "dry"
	PURE            Category = "pure"
	Plan            Category = "plan"
	PlanQuality     Category = "plan-quality"
	Specs           Category = "specs"
	Lessons         Category = "lessons"
	MilestoneReview Category = "milestone-review"
)

// AllCategories returns every supported category in stable order. Used
// for --help enumeration and bulk-dispatch from push/merge in M5/M6.
func AllCategories() []Category {
	return []Category{DRY, PURE, Plan, PlanQuality, Specs, Lessons, MilestoneReview}
}

// IsValid reports whether s names a known category.
func IsValid(s string) bool {
	for _, c := range AllCategories() {
		if string(c) == s {
			return true
		}
	}
	return false
}

// Label returns a human-readable description for the category, matching
// the shell's CHECK_LABELS entries.
func (c Category) Label() string {
	switch c {
	case DRY:
		return "Check DRY principle"
	case PURE:
		return "Check PURE principle"
	case Plan:
		return "Check issue plan completeness"
	case PlanQuality:
		return "Check plan executability (pre-implementation)"
	case Specs:
		return "Check atlas/README sync"
	case Lessons:
		return "Check for lessons to capture"
	case MilestoneReview:
		return "Post-milestone code review (AGENTS.md §3)"
	}
	return string(c)
}

// NeedsAgent reports whether the category invokes the LLM. `lessons`
// is just a reminder ping — no diff, no agent.
func (c Category) NeedsAgent() bool {
	return c != Lessons
}

// AllowedTools returns the tool allowlist for this category's agent
// invocation. ALL judges are READ-ONLY reviewers (#62 M2): they report
// findings; the main agent — which has full session context — applies the
// fixes, commits, and re-runs. A gate that mutates the tree (the old Specs
// auto-edit) could pass while leaving uncommitted changes, stranding the
// subsequent merge; read-only removes that failure mode by construction.
// (Bash stays for read-only inspection — grep, go vet — matching the other
// review categories' long-standing posture.)
func (c Category) AllowedTools() string {
	return "Read,Grep,Glob,Bash"
}

// PromptInput is the data each category's prompt template consumes.
// Callers populate the fields relevant to the category they invoke;
// unused fields are ignored.
type PromptInput struct {
	Diff          string   // unified diff of the review window
	ChangedIssues []string // paths to changed issue files (for `plan`)
	Base, Head    string   // refs that bound the window (for milestone-review)
	IssueRef      string   // e.g. "ariadne#31 M2" (for milestone-review / plan-quality)
	IssueContent  string   // full issue file text (for plan-quality, where we
	//   assess current state, not a diff)
	PlanContent string // optional separate plan file text (for plan-quality)
}

// BuildPrompt renders the prompt for one category. Returns "" for
// categories that don't invoke an agent (lessons).
//
// Wording is preserved byte-faithfully from pre-merge-checks.sh's
// build_prompt heredocs so the agent behavior matches the shell version.
// Drift between this prompt and the shell version is a bug — they
// describe one contract.
func BuildPrompt(category Category, in PromptInput) string {
	switch category {
	case DRY:
		return fmt.Sprintf(`You are a code reviewer. Review the following diff for DRY (Don't Repeat Yourself) violations.
Look for: duplicated logic, copy-pasted code blocks, functions that could be consolidated,
repeated patterns that should be extracted into shared helpers.

Report any violations you find with file paths and line numbers. Suggest how to fix them.
Do NOT modify any files. Only report.

%s

Tokens for this check:
  CLEAN   = no DRY violations.
  FAILURE = duplicated logic that should be consolidated.

Diff:
%s
`, ContractPreamble, in.Diff)

	case PURE:
		return fmt.Sprintf(`You are a code reviewer. Review the following diff for PURE principle adherence.
The PURE principle means: write the majority of code as pure functions (no side effects, deterministic),
then use minimal "glue" code to integrate with UI and IO.

Look for: business logic mixed with IO, functions that could be pure but aren't,
side effects that could be moved to the boundary.

Report any violations with file paths and line numbers. Suggest how to refactor.
Do NOT modify any files. Only report.

%s

Tokens for this check:
  CLEAN   = no PURE violations.
  FAILURE = business logic mixed with IO that should move to the boundary.

Diff:
%s
`, ContractPreamble, in.Diff)

	case Plan:
		changedList := strings.Join(in.ChangedIssues, "\n")
		return fmt.Sprintf(`You are a project management reviewer (TPM). You don't know technical details.
Only review the issue files that changed in this diff — do NOT review other issues.

For each changed issue file, check:
1. Does it have a filled-in Plan section with checklist items?
2. Are plan checklist items that appear done (based on the diff and git log) still unchecked?
3. Does the Log section have entries documenting what was done?
4. Is the status frontmatter correct (should it be "done")?

Do NOT modify any files.
If a checklist item looks completed based on the diff, say so and recommend checking it off.

%s

Tokens for this check:
  CLEAN   = no issues; ready to ship.
  INFO    = informational/non-blocking notes only (minor nits, stylistic).
  FAILURE = issues that must be addressed before shipping (unchecked-but-done
            items, missing log entries, wrong status frontmatter, etc.).

After the VERDICT line: a 1-paragraph summary explaining it, then any findings.

Changed issue files:
%s

Diff:
%s
`, ContractPreamble, changedList, in.Diff)

	case PlanQuality:
		ref := in.IssueRef
		if ref == "" {
			ref = "<unknown>"
		}
		planSection := in.PlanContent
		if planSection == "" {
			planSection = "(no separate plan file)"
		}
		return fmt.Sprintf(`You are a senior engineer reviewing an issue's plan BEFORE implementation begins.
Issue: %s

Read the Spec + Plan (and the separate plan file if present below). Your job is
to answer one question: **Is this plan executable as-written, or does it need
refinement before someone starts changing code?**

Common failure modes to flag:

  - Vague checklist items ("do the thing", "implement it", "handle errors")
    that don't name files, functions, or concrete behaviors.
  - Missing test surface — the Plan changes code but doesn't say what
    behavior the tests will pin.
  - Undefined acceptance criteria — Done-when is empty or just paraphrases
    the title.
  - Undeclared cross-issue or cross-repo deps — Plan touches files owned by
    another in-flight workstream without acknowledging it.
  - Scope creep risk — Plan mixes the stated change with unrelated cleanup
    that should be its own issue.
  - Mismatched estimate vs scope — estimate_hours wildly disagrees with the
    visible scope (e.g., 0.5h for what looks like 8h of work).

%s

Tokens for this check:
  CLEAN   = plan is concrete, testable, scoped; safe to start coding.
  INFO    = plan is workable but has minor non-blocking suggestions.
  FAILURE = plan has at least one of the failure modes above; address before
            starting implementation.

After the VERDICT line: a 1-paragraph summary of the verdict followed by specific
findings (quote the vague items, name the missing test surface, etc.). Be
concrete — vague approval is as harmful as vague plans.

Issue file:
---
%s
---

Plan file (if separate):
---
%s
---
`, ref, ContractPreamble, in.IssueContent, planSection)

	case Specs:
		return fmt.Sprintf(`You are a READ-ONLY documentation reviewer. Compare the code changes in the diff below against:
1. The spec files in atlas/
2. README.md

Those files are not meant to be comprehensive — atlas/ is a practical pointer for future developers and agents to the sketch of functionalities, history, and intention; details live in the code. Do NOT flag documentation that is fine, and do NOT ask for over-specification.

DO NOT EDIT ANY FILES. You are a gate, not a doer: report stale/incorrect docs precisely (file:line + what's out of sync + the fix needed) and let the main agent — which has full session context — apply them, commit, and re-run. (Editing here would let a passing gate leave the tree dirty and strand the merge — #62.)

%s

Tokens for this check:
  CLEAN   = atlas + README are in sync with the diff; nothing to change.
  INFO    = only minor / optional suggestions; nothing stale that blocks.
  FAILURE = stale or incorrect documentation that must be fixed before shipping
            (the main agent fixes, commits, and re-runs).

After the VERDICT line: a 1-paragraph summary explaining it, followed by the list
of stale spots (file:line + the concrete fix) so the main agent can apply them.

Diff:
%s
`, ContractPreamble, in.Diff)

	case Lessons:
		// No agent invocation — just a reminder ping. Caller emits the
		// REMINDER: line directly so output classification recognizes it
		// as info, not failure.
		return ""

	case MilestoneReview:
		ref := in.IssueRef
		if ref == "" {
			ref = "<unknown>"
		}
		return fmt.Sprintf(`You are conducting a post-milestone code review for %s.
Base: %s   Head: %s

Read the diff against the issue's plan + spec. Focus on:

  Critical (must fix before next milestone)
    - correctness bugs
    - behavior drift from stated contracts (look for ports of existing
      scripts where byte-faithfulness was promised)
    - crashes / panics on unexpected input
    - silent error swallowing where the source raised

  Important (fix before next milestone if cheap)
    - API design of newly-introduced internal packages (downstream
      milestones will consume them; surface stable?)
    - missing test coverage that would catch the kind of bug shipped
    - inconsistent error handling philosophy across the diff

  Minor (note for future)
    - style nits, naming, comment density
    - performance only if hot-path

  Core concepts cross-check (if the plan has a Core concepts table):
    The plan should list entities in a greppable table — name, kind
    (PURE/INTEGRATION), file location, status (new/modified/deleted).
    For each row:
      - Verify the entity exists at the stated path (grep the diff or
        filesystem).
      - PURE: tests run without IO (no exec, net, mutable fs). If tests
        need mocks to run, it isn't really PURE — flag Critical and
        recommend promoting it to INTEGRATION.
      - INTEGRATION: injected into pure callers, not invoked directly
        from business logic.
      - "modified" / "deleted" status: the diff shows the expected
        change/removal at the stated location.
    Any contradiction between table and code = Critical finding, plus
    a plan-revision recommendation (a "## Revisions" entry on the plan
    so it stops claiming what the code doesn't deliver).

  Atlas update gate (per AGENTS.md §8):
    The milestone should update atlas/ entries for any new architectural
    surface, flow, or terminology introduced. Scan the diff for evidence
    of new surface — new entity types, new subcommands, new conventions,
    new file-tree locations. Any present without corresponding atlas/
    changes in the same range = Important finding ("atlas update appears
    missing for <surface>").

%s

Tokens for milestone review:
  SHIP          = ready; ship it.
  FIX-THEN-SHIP = ship after addressing the findings (non-blocking at the gate;
                  fix them before the next milestone).
  REWORK        = blocking; needs rework before shipping.

After the VERDICT line: a 1-paragraph summary explaining it — what worked, what
blocks SHIP if it isn't — followed by:
  1. Strengths: 2-5 specific things done well (file:line where useful).
     Affirm the validated approaches so the operator knows what's
     confirmed-good ground. Empty acceptable for trivial milestones.
  2. Critical findings (file:line + fix sketch); empty if none.
  3. Important findings (same format).
  4. Minor findings (terse one-liners).
  5. Test coverage notes.
  6. Architectural notes for upcoming work.
  7. Plan revision recommendations: list specific "## Revisions" entries
     the plan needs (empty if the plan still matches the code).

You have no prior session context — that is the anti-collusion property.
Verify behavior against documented contracts directly; do not take the
implementor's word in commit messages or docs at face value.

Tools: read-only. Do not modify code.

Diff:
%s
`, ref, in.Base, in.Head, ContractPreamble, in.Diff)
	}
	return ""
}

// LessonsReminder is the line `sdlc judge lessons` emits in place of an
// agent invocation. Matches pre-merge-checks.sh's emit_check_message
// for `lessons` so the output classifier picks it up as info.
const LessonsReminder = "REMINDER: Review workshop/lessons.md — capture any non-obvious patterns from this session."
