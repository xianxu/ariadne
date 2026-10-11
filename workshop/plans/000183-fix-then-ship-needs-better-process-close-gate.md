---
gate: boundary-review
issue: 183
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T17:44:00-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: 'The fix-review''s REWORK-on-scope-change rule can''t fire: every window command excludes workshop/issues/'
          detail: 'fix-review.md point 4 tells the reviewer to return REWORK when the interdiff changes Spec, Revisions or Done when, but reviewCommands (internal/judge/reviewwindow.go) excludes the issues directory from stat, names, full and targeted, so the reviewer never sees a before/after of the issue body. Fix: compare the contract-section hashes at the reviewed head and at HEAD without the model, put the result in ReviewNotes with a section diff, and add a test where Spec changes after the close review.'
          family: reviewer-input-not-in-window
          round: 1
        - id: BR-2
          severity: Minor
          title: closefixed_test.go:117 boolean precedence means the reviewed-head equality is never checked
          detail: A || B && C parses as A || (B && C), so the condition only checks that Reviewed is non-empty. Parenthesize it or split it into two assertions.
          family: test-assertion-precedence
          round: 1
        - id: BR-3
          severity: Minor
          title: process-manual not regenerated although Done when says it catalogs fix-review
          family: generated-doc-deferred
          round: 1
        - id: BR-4
          severity: Minor
          title: Log tag reads "judged by the review" where the Spec says "judged by the <verdict> review"
          family: spec-wording-drift
          round: 1
        - id: BR-5
          severity: Minor
          title: The publish gate's code-arm loop re-implements conflictedCode; share one code-paths helper
          family: code-surface-predicate-dup
          round: 1
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#183 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T17:44:00-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `reviewer-input-not-in-window` The fix-review's REWORK-on-scope-change rule can't fire: every window command excludes workshop/issues/
  fix-review.md point 4 tells the reviewer to return REWORK when the interdiff changes Spec, Revisions or Done when, but reviewCommands (internal/judge/reviewwindow.go) excludes the issues directory from stat, names, full and targeted, so the reviewer never sees a before/after of the issue body. Fix: compare the contract-section hashes at the reviewed head and at HEAD without the model, put the result in ReviewNotes with a section diff, and add a test where Spec changes after the close review.
- **BR-2** [Minor] `test-assertion-precedence` closefixed_test.go:117 boolean precedence means the reviewed-head equality is never checked
  A || B && C parses as A || (B && C), so the condition only checks that Reviewed is non-empty. Parenthesize it or split it into two assertions.
- **BR-3** [Minor] `generated-doc-deferred` process-manual not regenerated although Done when says it catalogs fix-review
- **BR-4** [Minor] `spec-wording-drift` Log tag reads "judged by the review" where the Spec says "judged by the <verdict> review"
- **BR-5** [Minor] `code-surface-predicate-dup` The publish gate's code-arm loop re-implements conflictedCode; share one code-paths helper

## Open findings

- **BR-1** [Important] `reviewer-input-not-in-window` The fix-review's REWORK-on-scope-change rule can't fire: every window command excludes workshop/issues/
- **BR-2** [Minor] `test-assertion-precedence` closefixed_test.go:117 boolean precedence means the reviewed-head equality is never checked
- **BR-3** [Minor] `generated-doc-deferred` process-manual not regenerated although Done when says it catalogs fix-review
- **BR-4** [Minor] `spec-wording-drift` Log tag reads "judged by the review" where the Spec says "judged by the <verdict> review"
- **BR-5** [Minor] `code-surface-predicate-dup` The publish gate's code-arm loop re-implements conflictedCode; share one code-paths helper
