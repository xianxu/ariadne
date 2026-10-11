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
    - "n": 2
      timestamp: "2026-10-10T18:01:18-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: contractChangeNotes compares the issue body at review.HumanBase with now; TestFixedThenShip_ScopeChangeReachesTheReviewer fails without it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Split into two assertions at closefixed_test.go:123-129; the reviewed-head equality is now checked.
          round: 2
        - id: BR-3
          disposition: addressed
          note: FixReview is in AllInjectedCategories (prompts.go:138), so process-manual renders it; regeneration deferred with a recorded reason in Revisions.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The tag now reads "judged by the close review", followed by the verdict annotation; the Revisions entry records the change.
          round: 2
        - id: BR-5
          disposition: addressed
          note: codeSurfacePaths is shared by publishDeltaRank, classifyPublishDelta and fixedThenShipRefusal.
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: reviewanchor.go:156-161 still filters code paths by hand instead of calling codeSurfacePaths
          detail: 'Rule: every path-list-to-code-paths filter goes through codeSurfacePaths, so the only per-path call to publishGateHasCodeSurface is inside it. reviewanchor.go is the one remaining copy (older code, outside the window).'
          family: code-surface-predicate-dup
          round: 2
        - id: BR-7
          severity: Minor
          title: clipNote cuts at byte 4000 and can split a multi-byte character
          detail: Back up to the nearest character boundary (utf8.RuneStart) before s[:max], so a clipped Spec containing em-dashes or arrows stays valid UTF-8.
          family: byte-slice-utf8
          round: 2
      recipe: milestone-review
      reviewed: 0369e343aa665a4fbe0ca401be503dd4ce99c05c
      blocked: false
    - "n": 3
      timestamp: "2026-10-10T18:30:08-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: not-addressed
          note: reviewanchor.go:156-160 at 4a5f5730 still loops publishGateHasCodeSurface([]string{p}) by hand; replace with codeSurfacePaths(d.Paths). Minor, non-blocking.
          round: 3
        - id: BR-7
          disposition: not-addressed
          note: closefixed.go clipNote still returns s[:maxContractNoteSection]; back up to a utf8.RuneStart boundary. Minor, non-blocking.
          round: 3
      recipe: fix-review
      reviewed: 4a5f5730e9c4f04c31678f317334752ed13dec90
      blocked: false
    - "n": 4
      timestamp: "2026-10-10T18:31:56-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: reviewanchor.go:155 now calls codeSurfacePaths(d.Paths); every other publishGateHasCodeSurface caller outside the helper uses it as a whole-list boolean or a non-code conflict-marker check, so the rule holds across the tree.
          round: 4
        - id: BR-7
          disposition: addressed
          note: clipNote backs up to a utf8.RuneStart boundary; the new TestContractChangeNotes case puts é across the 4000-byte cut and fails on the old s[:max].
          round: 4
      recipe: fix-review
      reviewed: 972baa66aa67bfc7527d4f087e0708e32ff4eaed
      blocked: false
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

## Round 2 — 2026-10-10T18:01:18-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — contractChangeNotes compares the issue body at review.HumanBase with now; TestFixedThenShip_ScopeChangeReachesTheReviewer fails without it.
- BR-2 — addressed — Split into two assertions at closefixed_test.go:123-129; the reviewed-head equality is now checked.
- BR-3 — addressed — FixReview is in AllInjectedCategories (prompts.go:138), so process-manual renders it; regeneration deferred with a recorded reason in Revisions.
- BR-4 — addressed — The tag now reads "judged by the close review", followed by the verdict annotation; the Revisions entry records the change.
- BR-5 — addressed — codeSurfacePaths is shared by publishDeltaRank, classifyPublishDelta and fixedThenShipRefusal.

### Raised

- **BR-6** [Minor] `code-surface-predicate-dup` reviewanchor.go:156-161 still filters code paths by hand instead of calling codeSurfacePaths
  Rule: every path-list-to-code-paths filter goes through codeSurfacePaths, so the only per-path call to publishGateHasCodeSurface is inside it. reviewanchor.go is the one remaining copy (older code, outside the window).
- **BR-7** [Minor] `byte-slice-utf8` clipNote cuts at byte 4000 and can split a multi-byte character
  Back up to the nearest character boundary (utf8.RuneStart) before s[:max], so a clipped Spec containing em-dashes or arrows stays valid UTF-8.

## Round 3 — 2026-10-10T18:30:08-07:00 (claude) — passed

### Disposed

- BR-6 — not-addressed — reviewanchor.go:156-160 at 4a5f5730 still loops publishGateHasCodeSurface([]string{p}) by hand; replace with codeSurfacePaths(d.Paths). Minor, non-blocking.
- BR-7 — not-addressed — closefixed.go clipNote still returns s[:maxContractNoteSection]; back up to a utf8.RuneStart boundary. Minor, non-blocking.

## Round 4 — 2026-10-10T18:31:56-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — reviewanchor.go:155 now calls codeSurfacePaths(d.Paths); every other publishGateHasCodeSurface caller outside the helper uses it as a whole-list boolean or a non-code conflict-marker check, so the rule holds across the tree.
- BR-7 — addressed — clipNote backs up to a utf8.RuneStart boundary; the new TestContractChangeNotes case puts é across the 4000-byte cut and fails on the old s[:max].

## Open findings

(none — every finding has been disposed)
