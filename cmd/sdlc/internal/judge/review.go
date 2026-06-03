package judge

import (
	_ "embed"
	"strings"
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
// substitutes the issue ref + base/head, and expands {{ARCH_STAR}} to the live
// ARCH-* marker list (from the registry, via ArchitectureMarkers — so the
// enumerated checklist tracks the registry with no hardcoding). Pure string
// templating (ARCH-PURE); the architecture block + output contract + diff are
// appended by the caller (BuildPrompt).
func CodeReviewBody(issueRef, base, head string) string {
	r := strings.NewReplacer(
		"{{ISSUE_REF}}", issueRef,
		"{{BASE}}", base,
		"{{HEAD}}", head,
		"{{ARCH_STAR}}", strings.Join(ArchitectureMarkers(), ", "),
	)
	return r.Replace(codeReviewTemplate)
}
