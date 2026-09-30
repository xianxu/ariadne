---
gate: boundary-review
issue: 272
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-29T21:12:59-07:00"
      agent: claude
      recipe: small-diff-review
      blocked: false
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-09-29T21:25:17-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: 'Descendant exemption fails once the parent advances: start-plan refuses the parent of an unlanded child'
          detail: unlandedSharedBase exempts ref only when tip is an ancestor of ref. If parent 000009 commits again after child 000010 forked, the merge base is the fork point beyond main, so the parent is refused as carrying the child's work, which contradicts the Done-when clause that a branch an unlanded descendant was built on still passes. The test case at planningbranch_test.go:189 never commits on the parent after the fork. This is the only instance of the family in the window. Fix by deciding who owns main..merge-base (commit issue tags, or the recorded fork point) and add the parent-advanced regression case.
          family: stack-ownership-by-ancestry
          round: 2
        - id: BR-2
          severity: Minor
          title: merge-base failure is treated as unrelated histories and passes the guard
          detail: planningbranch.go:137-140 returns (false, nil) on any merge-base error. Only exit 1 with empty output means no common ancestor; other errors should propagate. This is the only swallow site in the diff.
          family: silent-error-swallow
          round: 2
        - id: BR-3
          severity: Minor
          title: Switch path fetches main twice; the already-on-branch path adds a new fetch
          detail: preparePlanningBranch pins main at line 52, then refuseUnlandedBase calls env.main.Snapshot() again. Pass the pinned ref in. Note the already-on-branch path now needs the network.
          family: redundant-main-snapshot
          round: 2
      recipe: small-diff-review
      blocked: true
    - "n": 3
      timestamp: "2026-09-29T21:30:07-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Ownership is now read from #N tags; the case "a child built on this branch, which then advanced" pins the parent-advanced scenario and passes.'
          round: 3
        - id: BR-2
          disposition: addressed
          note: The merge-base call was removed; rev-list, for-each-ref and log errors all propagate from refuseUnlandedBase.
          round: 3
        - id: BR-3
          disposition: addressed
          note: The switch path passes pinned in. The already-on-branch Snapshot is needed because Done-when checks against fresh main.
          round: 3
      findings:
        - id: BR-4
          severity: Minor
          title: 'A #N mention anywhere in this branch''s own commit subject counts it as another issue''s work'
          detail: 'issueTagRE matches anywhere in the subject, so a commit "#9: prep hook for #10", shared with an unlanded 000010 branch built on 000009, refuses #9. This breaks the Done-when clause for a descendant built on this branch. Fix: skip commits whose subject also carries the planning issue''s own tag. This is the only instance in the window.'
          family: commit-owner-leading-tag
          round: 3
        - id: BR-5
          severity: Minor
          title: issueBranchRE repeats issueFamilyRE (migrate.go:178) with an added capture group
          detail: 'ARCH-DRY: use one NNNNNN- prefix regex with a capture group for both. These are the only two instances.'
          family: shared-issue-id-regex
          round: 3
        - id: BR-6
          severity: Minor
          title: The switch-to-existing-branch path has no passing test case
          detail: Every passing case runs on the already-on-branch path; the switch path is tested only for refusal. Add one passing switch case, e.g. an existing branch that merged main, checked out from rest.
          family: two-mode-clause-coverage
          round: 3
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#272 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-29T21:12:59-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-09-29T21:25:17-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `stack-ownership-by-ancestry` Descendant exemption fails once the parent advances: start-plan refuses the parent of an unlanded child
  unlandedSharedBase exempts ref only when tip is an ancestor of ref. If parent 000009 commits again after child 000010 forked, the merge base is the fork point beyond main, so the parent is refused as carrying the child's work, which contradicts the Done-when clause that a branch an unlanded descendant was built on still passes. The test case at planningbranch_test.go:189 never commits on the parent after the fork. This is the only instance of the family in the window. Fix by deciding who owns main..merge-base (commit issue tags, or the recorded fork point) and add the parent-advanced regression case.
- **BR-2** [Minor] `silent-error-swallow` merge-base failure is treated as unrelated histories and passes the guard
  planningbranch.go:137-140 returns (false, nil) on any merge-base error. Only exit 1 with empty output means no common ancestor; other errors should propagate. This is the only swallow site in the diff.
- **BR-3** [Minor] `redundant-main-snapshot` Switch path fetches main twice; the already-on-branch path adds a new fetch
  preparePlanningBranch pins main at line 52, then refuseUnlandedBase calls env.main.Snapshot() again. Pass the pinned ref in. Note the already-on-branch path now needs the network.

## Round 3 — 2026-09-29T21:30:07-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Ownership is now read from #N tags; the case "a child built on this branch, which then advanced" pins the parent-advanced scenario and passes.
- BR-2 — addressed — The merge-base call was removed; rev-list, for-each-ref and log errors all propagate from refuseUnlandedBase.
- BR-3 — addressed — The switch path passes pinned in. The already-on-branch Snapshot is needed because Done-when checks against fresh main.

### Raised

- **BR-4** [Minor] `commit-owner-leading-tag` A #N mention anywhere in this branch's own commit subject counts it as another issue's work
  issueTagRE matches anywhere in the subject, so a commit "#9: prep hook for #10", shared with an unlanded 000010 branch built on 000009, refuses #9. This breaks the Done-when clause for a descendant built on this branch. Fix: skip commits whose subject also carries the planning issue's own tag. This is the only instance in the window.
- **BR-5** [Minor] `shared-issue-id-regex` issueBranchRE repeats issueFamilyRE (migrate.go:178) with an added capture group
  ARCH-DRY: use one NNNNNN- prefix regex with a capture group for both. These are the only two instances.
- **BR-6** [Minor] `two-mode-clause-coverage` The switch-to-existing-branch path has no passing test case
  Every passing case runs on the already-on-branch path; the switch path is tested only for refusal. Add one passing switch case, e.g. an existing branch that merged main, checked out from rest.

## Open findings

- **BR-4** [Minor] `commit-owner-leading-tag` A #N mention anywhere in this branch's own commit subject counts it as another issue's work
- **BR-5** [Minor] `shared-issue-id-regex` issueBranchRE repeats issueFamilyRE (migrate.go:178) with an added capture group
- **BR-6** [Minor] `two-mode-clause-coverage` The switch-to-existing-branch path has no passing test case
