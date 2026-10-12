---
gate: boundary-review
issue: 309
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T12:25:10-07:00"
      dispatched: "2026-10-12T11:59:20-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Legacy merge path never verifies the checks gh reports belong to the pinned head
          detail: merge.go:552-566 pins head from local origin/<branch> but gh pr checks reads GitHub's current PR head; right after a push the PR object can lag, old-head green checks pass the gate, and --match-head-commit only protects if GitHub has not yet synced. Compare PR headRefOid to head (treat mismatch as pending) and add a fake test.
          family: checks-bound-to-merged-head
          round: 1
        - id: BR-2
          severity: Minor
          title: Dropped M1 dry-run bullet is recorded only in the issue Log, not the plan Revisions
          detail: Add a Revisions entry so the plan stops claiming the dry-run behaviour.
          family: plan-revisions-record-scope-changes
          round: 1
        - id: BR-3
          severity: Minor
          title: no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
          detail: 'gatesig.go:163 matches "PR N has|reports" and "could not read", but not "CI gate: read <head>''s tree" or "CI gate: cannot pin the merge".'
          family: gate-catalog-covers-all-refusals
          round: 1
        - id: BR-4
          severity: Minor
          title: exitErr test helper builds the exit code via rune arithmetic, valid only for 0-9
          detail: cigate_test.go:68; use strconv.Itoa.
          family: test-helper-correctness
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T12:56:15-07:00"
      dispatched: "2026-10-12T12:43:31-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: cigate.go:216-230 reads the head before and after the checks; reverting the comparison turns TestRunCIGate and TestRunMerge_LaggingPRHeadRefuses red.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The plan's Revisions section now records the dropped dry-run bullet (2026-10-12, M1 implementation).
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: 'The prefix pattern covers the class, but no classifier test pins any no-ci refusal: reverting gatesig.go to the 80160793 pattern leaves processmanual tests green. Add TestClassifyOutputLine cases for every refusal string, plus negatives for the CI checks: success line and the ACK.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: cigate_test.go:72 now uses strconv.Itoa.
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      reviewed: 1f915d6065bf307057fd2c7c4a88d6bdcfe67b13
      blocked: false
---

# Gate ledger — ariadne#309 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T12:25:10-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `checks-bound-to-merged-head` Legacy merge path never verifies the checks gh reports belong to the pinned head
  merge.go:552-566 pins head from local origin/<branch> but gh pr checks reads GitHub's current PR head; right after a push the PR object can lag, old-head green checks pass the gate, and --match-head-commit only protects if GitHub has not yet synced. Compare PR headRefOid to head (treat mismatch as pending) and add a fake test.
- **BR-2** [Minor] `plan-revisions-record-scope-changes` Dropped M1 dry-run bullet is recorded only in the issue Log, not the plan Revisions
  Add a Revisions entry so the plan stops claiming the dry-run behaviour.
- **BR-3** [Minor] `gate-catalog-covers-all-refusals` no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
  gatesig.go:163 matches "PR N has|reports" and "could not read", but not "CI gate: read <head>'s tree" or "CI gate: cannot pin the merge".
- **BR-4** [Minor] `test-helper-correctness` exitErr test helper builds the exit code via rune arithmetic, valid only for 0-9
  cigate_test.go:68; use strconv.Itoa.

## Round 2 — 2026-10-12T12:56:15-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — cigate.go:216-230 reads the head before and after the checks; reverting the comparison turns TestRunCIGate and TestRunMerge_LaggingPRHeadRefuses red.
- BR-2 — addressed — The plan's Revisions section now records the dropped dry-run bullet (2026-10-12, M1 implementation).
- BR-3 — not-addressed — The prefix pattern covers the class, but no classifier test pins any no-ci refusal: reverting gatesig.go to the 80160793 pattern leaves processmanual tests green. Add TestClassifyOutputLine cases for every refusal string, plus negatives for the CI checks: success line and the ACK.
- BR-4 — addressed — cigate_test.go:72 now uses strconv.Itoa.

## Open findings

- **BR-3** [Minor] `gate-catalog-covers-all-refusals` no-ci RefusalPat misses the head-tree-read and cannot-pin refusals
