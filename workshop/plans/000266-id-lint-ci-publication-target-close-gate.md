---
gate: boundary-review
issue: 266
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T21:52:08-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: The --remote path's fail-closed case (marked repo, origin without an issue-tracker branch, exit 2) has no test
          detail: Spec, atlas/workflow/ci-merge-check.md:98 and the script comment all claim a marked repository whose origin carries no tracker still exits 2. It holds today only because TrunkFile.fetch fails and becomes offlineError, then degraded. Add a case to TestDuplicateIDCheckRunsInADetachedCICheckout that deletes origin's issue-tracker branch and asserts exit 2, so a tolerant tracker read cannot silently turn CI green.
          family: fail-closed-path-untested
          round: 1
        - id: BR-2
          severity: Minor
          title: Done-when clause 1 (parley.nvim PR run green) can only be verified after merge
          detail: Plan box 3 is unchecked and the Log explains why. Record this in a Revisions entry or in close --verified so the close does not claim an unobserved Done-when.
          family: done-when-unobserved-at-close
          round: 1
        - id: BR-3
          severity: Minor
          title: Duplicate empty "### 2026-09-28" heading in the issue Log
          family: log-hygiene
          round: 1
      recipe: small-diff-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-28T21:54:14-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: trackedlegacy_test.go deletes origin's issue-tracker and asserts exit 2 plus COULD NOT RUN with the stale fetched ref present; passes at head.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Done-when split into close-time and post-merge clauses, plus a Revisions entry explaining the post-merge evidence.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The issue Log now has a single 2026-09-28 heading.
          round: 2
      findings:
        - id: BR-4
          severity: Minor
          title: Revisions entry cites Done-when clause numbers that are inverted after the reorder
          detail: The Revisions entry calls the parley.nvim PR run clause 1 and says close attests clause 2. The final commit reordered Done-when so the regression test is clause 1 and the parley run is clause 2. Swap the numbers. This is the only instance in the window.
          family: issue-prose-cross-reference-drift
          round: 2
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#266 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T21:52:08-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `fail-closed-path-untested` The --remote path's fail-closed case (marked repo, origin without an issue-tracker branch, exit 2) has no test
  Spec, atlas/workflow/ci-merge-check.md:98 and the script comment all claim a marked repository whose origin carries no tracker still exits 2. It holds today only because TrunkFile.fetch fails and becomes offlineError, then degraded. Add a case to TestDuplicateIDCheckRunsInADetachedCICheckout that deletes origin's issue-tracker branch and asserts exit 2, so a tolerant tracker read cannot silently turn CI green.
- **BR-2** [Minor] `done-when-unobserved-at-close` Done-when clause 1 (parley.nvim PR run green) can only be verified after merge
  Plan box 3 is unchecked and the Log explains why. Record this in a Revisions entry or in close --verified so the close does not claim an unobserved Done-when.
- **BR-3** [Minor] `log-hygiene` Duplicate empty "### 2026-09-28" heading in the issue Log

## Round 2 — 2026-09-28T21:54:14-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — trackedlegacy_test.go deletes origin's issue-tracker and asserts exit 2 plus COULD NOT RUN with the stale fetched ref present; passes at head.
- BR-2 — addressed — Done-when split into close-time and post-merge clauses, plus a Revisions entry explaining the post-merge evidence.
- BR-3 — addressed — The issue Log now has a single 2026-09-28 heading.

### Raised

- **BR-4** [Minor] `issue-prose-cross-reference-drift` Revisions entry cites Done-when clause numbers that are inverted after the reorder
  The Revisions entry calls the parley.nvim PR run clause 1 and says close attests clause 2. The final commit reordered Done-when so the regression test is clause 1 and the parley run is clause 2. Swap the numbers. This is the only instance in the window.

## Open findings

- **BR-4** [Minor] `issue-prose-cross-reference-drift` Revisions entry cites Done-when clause numbers that are inverted after the reorder
