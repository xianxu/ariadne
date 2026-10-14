---
gate: boundary-review
issue: 348
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T18:20:00-07:00"
      dispatched: "2026-10-13T18:13:50-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: issue validate fails every file when no main ref exists, even files with no area/severity
          detail: 'issueValidateDeclaration runs unconditionally (issue.go:122) and classificationProblems turns any declErr into a problem for every file (issue.go:176). Reproduced: in a master-only repo the base binary exits 0 and head exits 1. The existing TestRunIssueValidate fails on a clean checkout of the pinned head (no main ref) because it reads the real repo''s refs via repoRootOf(cwd). Fix: read the declaration lazily, only for files or cards that carry an area. Inject the reader and make the test hermetic.'
          family: classification-check-gates-unclassified-input
          round: 1
        - id: BR-2
          severity: Important
          title: 'Declarable areas such as 2026, 1e5, true and null cannot be written: flowList/Render emit them unquoted, so they become non-string YAML scalars'
          detail: 'The segment rule allows digit-led and keyword segments. flowList (repoenums.go:660) and Render (scaffold.go:242) emit them unquoted, and validateRepoEnumNode requires !!str, so set-area and issue new refuse a declared value. FieldList (gates) decodes the same value as a string, so the writers and the gates disagree. Fix: quote non-!!str values, or tighten the segment rule in issue.cue. Add a test.'
          family: writer-gate-encoding-divergence
          round: 1
        - id: BR-3
          severity: Important
          title: README update appears missing for issue set-area/set-severity, the --area/--severity flags and .sdlc/vocabulary.json
          family: docs-gate-readme
          round: 1
        - id: BR-4
          severity: Minor
          title: checkClassification does not refuse duplicate areas; slices.Compact only removes adjacent ones, so set-area a b a fails later in SetCardField
          family: writer-validation-parity
          round: 1
        - id: BR-5
          severity: Minor
          title: Severity membership and area syntax are each checked in three places (card.go, checkClassification, newIssueListFilter); consolidate them into one helper
          family: arch-dry-shared-validator
          round: 1
        - id: BR-6
          severity: Minor
          title: 'state.go: withClassification was inserted under withOwnership''s doc comment, so withOwnership lost its doc'
          family: doc-comment-placement
          round: 1
        - id: BR-7
          severity: Minor
          title: repoenum.ValidValue recompiles the path regex on every call; cache it per spec
          family: hot-path-recompile
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#348 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T18:20:00-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `classification-check-gates-unclassified-input` issue validate fails every file when no main ref exists, even files with no area/severity
  issueValidateDeclaration runs unconditionally (issue.go:122) and classificationProblems turns any declErr into a problem for every file (issue.go:176). Reproduced: in a master-only repo the base binary exits 0 and head exits 1. The existing TestRunIssueValidate fails on a clean checkout of the pinned head (no main ref) because it reads the real repo's refs via repoRootOf(cwd). Fix: read the declaration lazily, only for files or cards that carry an area. Inject the reader and make the test hermetic.
- **BR-2** [Important] `writer-gate-encoding-divergence` Declarable areas such as 2026, 1e5, true and null cannot be written: flowList/Render emit them unquoted, so they become non-string YAML scalars
  The segment rule allows digit-led and keyword segments. flowList (repoenums.go:660) and Render (scaffold.go:242) emit them unquoted, and validateRepoEnumNode requires !!str, so set-area and issue new refuse a declared value. FieldList (gates) decodes the same value as a string, so the writers and the gates disagree. Fix: quote non-!!str values, or tighten the segment rule in issue.cue. Add a test.
- **BR-3** [Important] `docs-gate-readme` README update appears missing for issue set-area/set-severity, the --area/--severity flags and .sdlc/vocabulary.json
- **BR-4** [Minor] `writer-validation-parity` checkClassification does not refuse duplicate areas; slices.Compact only removes adjacent ones, so set-area a b a fails later in SetCardField
- **BR-5** [Minor] `arch-dry-shared-validator` Severity membership and area syntax are each checked in three places (card.go, checkClassification, newIssueListFilter); consolidate them into one helper
- **BR-6** [Minor] `doc-comment-placement` state.go: withClassification was inserted under withOwnership's doc comment, so withOwnership lost its doc
- **BR-7** [Minor] `hot-path-recompile` repoenum.ValidValue recompiles the path regex on every call; cache it per spec

## Open findings

- **BR-1** [Critical] `classification-check-gates-unclassified-input` issue validate fails every file when no main ref exists, even files with no area/severity
- **BR-2** [Important] `writer-gate-encoding-divergence` Declarable areas such as 2026, 1e5, true and null cannot be written: flowList/Render emit them unquoted, so they become non-string YAML scalars
- **BR-3** [Important] `docs-gate-readme` README update appears missing for issue set-area/set-severity, the --area/--severity flags and .sdlc/vocabulary.json
- **BR-4** [Minor] `writer-validation-parity` checkClassification does not refuse duplicate areas; slices.Compact only removes adjacent ones, so set-area a b a fails later in SetCardField
- **BR-5** [Minor] `arch-dry-shared-validator` Severity membership and area syntax are each checked in three places (card.go, checkClassification, newIssueListFilter); consolidate them into one helper
- **BR-6** [Minor] `doc-comment-placement` state.go: withClassification was inserted under withOwnership's doc comment, so withOwnership lost its doc
- **BR-7** [Minor] `hot-path-recompile` repoenum.ValidValue recompiles the path regex on every call; cache it per spec
