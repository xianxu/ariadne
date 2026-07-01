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
	"embed"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// promptFS holds the per-category prompt templates as embedded markdown — the
// readable, single source of each prompt's prose (#153 M2). Extends the pattern
// architecture.md / code-review.md already use in this package. BuildPrompt loads
// a template and substitutes {{TOKEN}}s; the rendered output is golden-tested for
// byte-fidelity (golden_test.go).
//
//go:embed prompts/*.md
var promptFS embed.FS

// promptTemplate reads prompts/<category>.md. Panics on a missing template — a
// build-time bug, like helptext.MustGet.
func promptTemplate(c Category) string {
	b, err := promptFS.ReadFile("prompts/" + string(c) + ".md")
	if err != nil {
		panic("judge: prompt template missing: " + string(c) + ".md")
	}
	return string(b)
}

// promptSubstitutions builds the single Replacer mapping every {{TOKEN}} to its
// value. {{ARCH_BLOCK}} resolves to the at-plan lens for plan-quality, at-review
// otherwise. A .md uses only the tokens it needs; unused ones are simply absent.
// strings.Replacer is single-pass, so a value containing "{{…}}" (e.g. a diff) is
// inserted literally — no accidental re-substitution.
func promptSubstitutions(c Category, in PromptInput) *strings.Replacer {
	archLens := "at-review"
	if c == PlanQuality {
		archLens = "at-plan"
	}
	return strings.NewReplacer(
		"{{ARCH_BLOCK}}", ArchitectureBlock(archLens),
		"{{CONTRACT}}", ContractPreamble,
		"{{BOUNDARY_CONTRACT}}", BoundaryReviewContract,
		"{{CODE_REVIEW_BODY}}", CodeReviewBody(in),
		"{{DIFF}}", in.Diff,
		"{{CHANGED_ISSUES}}", strings.Join(in.ChangedIssues, "\n"),
		"{{ISSUE_CONTENT}}", in.IssueContent,
		"{{PLAN_CONTENT}}", orDefault(in.PlanContent, "(no separate plan file)"),
		"{{REF}}", orDefault(in.IssueRef, "<unknown>"),
		"{{MODEL}}", estimate.CurrentModel(),
	)
}

// renderTemplate is the uniform path: load the category's .md and substitute.
func renderTemplate(c Category, in PromptInput) string {
	return promptSubstitutions(c, in).Replace(promptTemplate(c))
}

// Category enumerates the supported judge checks. Names match the
// shell's CHECK_NAMES array verbatim so `make check-dry` and
// `sdlc judge dry` invoke the same prompt.
type Category string

const (
	DRY             Category = "dry"
	PURE            Category = "pure"
	Plan            Category = "plan"
	PlanQuality     Category = "plan-quality"
	EstimateQuality Category = "estimate-quality"
	Specs           Category = "specs"
	Lessons         Category = "lessons"
	MilestoneReview Category = "milestone-review"
)

// AllCategories returns every supported category in stable order. Used
// for --help enumeration and bulk-dispatch from push/merge in M5/M6.
//
// EstimateQuality is deliberately ABSENT: it is a change-code-time-only gate
// (invoked directly by runEstimateQualityJudge, not via IsValid), and enrolling
// it here would also run it at push/merge bulk-dispatch — wrong boundary (#117).
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
	case EstimateQuality:
		return "Check the ## Estimate derivation was applied (pre-implementation)"
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
	IssueRef      string   // e.g. "pair#31 M2" — repo-prefixed (for milestone-review / plan-quality)
	IssueContent  string   // full issue file text (for plan-quality, where we
	//   assess current state, not a diff)
	PlanContent string // optional separate plan file text (for plan-quality)
	// Boundary-review repo orientation (#137) — derived in cmd/sdlc from the live
	// git context and rendered into code-review.md so a fresh reviewer is anchored
	// to the ACTUAL repo, not a hardcoded "ariadne". Empty for non-review categories.
	Repo      string // repo name (git-root basename), e.g. "pair"
	RepoRoot  string // absolute git-root path
	IssueFile string // path to the issue file under review
	Boundary  string // "whole-issue close" | "milestone Mx close"
	RepoNote  string // base-vs-downstream orientation note
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
		return renderTemplate(category, in)

	case PURE:
		return renderTemplate(category, in)

	case Plan:
		return renderTemplate(category, in)

	case PlanQuality:
		return renderTemplate(category, in)

	case EstimateQuality:
		return renderTemplate(category, in)

	case Specs:
		return renderTemplate(category, in)

	case Lessons:
		// No agent invocation — just a reminder ping. Caller emits the
		// REMINDER: line directly so output classification recognizes it
		// as info, not failure.
		return ""

	case MilestoneReview:
		// Composition (milestone-review.md): the code-review procedure header
		// (CodeReviewBody, #137) + the at-review ARCH block + the block-first
		// BoundaryReviewContract (#147). One reviewer, both boundaries.
		return renderTemplate(category, in)
	}
	return ""
}

// LessonsReminder is the line `sdlc judge lessons` emits in place of an
// agent invocation. Matches pre-merge-checks.sh's emit_check_message
// for `lessons` so the output classifier picks it up as info.
const LessonsReminder = "REMINDER: Review workshop/lessons.md — capture any non-obvious patterns from this session."
