package judge

import (
	_ "embed"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// codeReviewTemplate is the embedded code-review.md — the single source of the
// SDLC boundary-review *procedure* (#69): the quality/testing/requirements/
// readiness checklist + ariadne's Core-concepts cross-check, Atlas gate, severity
// buckets, and the SHIP|FIX-THEN-SHIP|REWORK verdict. It is the "essence" the
// binary renders at every boundary (milestone-close and close), reconciled from
// the adapted superpowers code-reviewer.md.
//
// Layering: this file owns the *procedure*; architecture.md owns the
// *principles*. The procedure refers to ARCH-* markers (via {{ARCH_STAR}}); it
// must NOT inline the principle bodies — those are delivered once, co-present,
// by ArchitectureBlock at dispatch (a guardrail test pins this).
//
//go:embed code-review.md
var codeReviewTemplate string

// CodeReviewBody renders the boundary-review procedure for one window: it
// substitutes the repo-orientation anchors (#137: repo name/root, issue ref +
// file, boundary kind, base/head, base-vs-downstream note — all from PromptInput)
// and expands {{ARCH_STAR}} to the live ARCH-* marker list (from the registry, via
// ArchitectureMarkers — so the enumerated checklist tracks the registry with no
// hardcoding). Pure string templating (ARCH-PURE); the architecture block + output
// contract + diff are appended by the caller (BuildPrompt).
//
// markers is the ARCH-* set this review applies ({{ARCH_STAR}}): every marker on
// the full flow, the registry's quick-flow set on the small-diff review (#231).
func CodeReviewBody(in PromptInput, markers []string) string {
	ref := orDefault(in.IssueRef, "<unknown>")
	r := strings.NewReplacer(
		"{{ISSUE_REF}}", ref,
		"{{BASE}}", in.Base,
		"{{HEAD}}", in.Head,
		"{{REPO}}", orDefault(in.Repo, "<unknown-repo>"),
		"{{REPO_ROOT}}", in.RepoRoot,
		"{{ISSUE_FILE}}", in.IssueFile,
		"{{BOUNDARY}}", orDefault(in.Boundary, "a development boundary"),
		"{{REPO_NOTE}}", in.RepoNote,
		"{{VERDICT_BLOCK}}", vocab.Verdict().RenderBlockInstruction(),
		"{{ARCH_STAR}}", strings.Join(markers, ", "),
		"{{REVERT_CHECK}}", revertCheck(in.Isolated),
	)
	return r.Replace(codeReviewTemplate)
}

// revertCheck is the revert-the-fix instruction. Only a dispatch that really
// runs the reviewer in a disposable checkout (#204: close and milestone-close)
// may tell it to revert in place; every other caller still runs in the
// operator's checkout and keeps the scratch-copy instruction.
func revertCheck(isolated bool) string {
	if isolated {
		return `Where cheap, revert the fix and confirm
    the regression goes red: you run in a disposable checkout of the pinned
    head, removed after this review, so revert and test there (foreground).`
	}
	return `Where cheap, revert the fix in a scratch
    copy and confirm the regression goes red.`
}

// orDefault returns s, or def when s is empty.
func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
