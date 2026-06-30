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
func CodeReviewBody(in PromptInput) string {
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
		"{{ARCH_STAR}}", strings.Join(ArchitectureMarkers(), ", "),
	)
	return r.Replace(codeReviewTemplate)
}

// orDefault returns s, or def when s is empty.
func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
