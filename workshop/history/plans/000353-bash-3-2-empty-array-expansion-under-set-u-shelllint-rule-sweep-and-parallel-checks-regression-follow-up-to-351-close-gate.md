---
gate: boundary-review
issue: 353
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T19:21:16-07:00"
      dispatched: "2026-10-13T19:19:42-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: MAX_PARALLEL_CHECKS positive-integer refusal has no automated test
          detail: scripts/parallel-checks.sh:99-103 refuses 0/non-numeric values (a Done-when item) but was verified only by a manual stub run; add a scripts test asserting exit 1 for MAX_PARALLEL_CHECKS=0 and =abc (local-gates label applies).
          family: behavior-change-without-regression-test
          round: 1
        - id: BR-2
          severity: Minor
          title: Plan checklist still describes default-deny; code flags only measured-unsafe operators
          detail: Add a Revisions entry recording the narrowed operator set (bare, ^, ,) per the 3.2.57 measurement noted in the Log.
          family: plan-text-drifts-from-code
          round: 1
        - id: BR-3
          severity: Minor
          title: isShell reimplements the shebang-field loop now provided by shebangNames
          detail: pkg/shelllint/repo_test.go:103-118 could call shebangNames(src, "sh","bash","zsh","dash","ksh") (ARCH-DRY).
          family: duplicated-helper-logic
          round: 1
        - id: BR-4
          severity: Minor
          title: Nounset detection ignores -u on the shebang line or bash -u invocation
          detail: This misses findings rather than over-flagging; no tracked script currently enables -u that way.
          family: nounset-detection-coverage
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T19:26:01-07:00"
      dispatched: "2026-10-13T19:24:06-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: scripts/test/parallel-checks.test.sh asserts exit 1 for 0/abc/-2; removing the validation turns 3/6 cases red (verified), and 20-ci-workflow-tests runs it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The Revisions entry dated 2026-10-13 records the narrowed operator set (bare, ^, ,), matching array.go unguardedArrays.
          round: 2
        - id: BR-3
          disposition: addressed
          note: isShell now calls shebangNames(src, "sh","bash","zsh","dash","ksh") at repo_test.go:95-105.
          round: 2
        - id: BR-4
          disposition: addressed
          note: shebangNounset regex in array.go plus TestArrayFindingsNeedsNounset cases for "#!/bin/bash -eu" and "env -S bash -u"; a caller's bash -u is noted as statically invisible.
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: ae1b29a6c0fbc177443e36687e706f04f1a2d8a0
      blocked: false
---

# Gate ledger — ariadne#353 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T19:21:16-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `behavior-change-without-regression-test` MAX_PARALLEL_CHECKS positive-integer refusal has no automated test
  scripts/parallel-checks.sh:99-103 refuses 0/non-numeric values (a Done-when item) but was verified only by a manual stub run; add a scripts test asserting exit 1 for MAX_PARALLEL_CHECKS=0 and =abc (local-gates label applies).
- **BR-2** [Minor] `plan-text-drifts-from-code` Plan checklist still describes default-deny; code flags only measured-unsafe operators
  Add a Revisions entry recording the narrowed operator set (bare, ^, ,) per the 3.2.57 measurement noted in the Log.
- **BR-3** [Minor] `duplicated-helper-logic` isShell reimplements the shebang-field loop now provided by shebangNames
  pkg/shelllint/repo_test.go:103-118 could call shebangNames(src, "sh","bash","zsh","dash","ksh") (ARCH-DRY).
- **BR-4** [Minor] `nounset-detection-coverage` Nounset detection ignores -u on the shebang line or bash -u invocation
  This misses findings rather than over-flagging; no tracked script currently enables -u that way.

## Round 2 — 2026-10-13T19:26:01-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — scripts/test/parallel-checks.test.sh asserts exit 1 for 0/abc/-2; removing the validation turns 3/6 cases red (verified), and 20-ci-workflow-tests runs it.
- BR-2 — addressed — The Revisions entry dated 2026-10-13 records the narrowed operator set (bare, ^, ,), matching array.go unguardedArrays.
- BR-3 — addressed — isShell now calls shebangNames(src, "sh","bash","zsh","dash","ksh") at repo_test.go:95-105.
- BR-4 — addressed — shebangNounset regex in array.go plus TestArrayFindingsNeedsNounset cases for "#!/bin/bash -eu" and "env -S bash -u"; a caller's bash -u is noted as statically invisible.

## Open findings

(none — every finding has been disposed)
