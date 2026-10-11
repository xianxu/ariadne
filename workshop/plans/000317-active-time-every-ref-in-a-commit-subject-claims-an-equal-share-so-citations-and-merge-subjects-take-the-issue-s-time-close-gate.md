---
gate: boundary-review
issue: 317
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T17:14:29-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Help text and Scope doc say merges are never scoped boundaries; only BranchPoint..HEAD merges are dropped
          detail: 'helptext/active-time.md "With --branch-point, merge commits are not boundaries" and commit.go Scope doc "a merge is never a scoped boundary". A main-side merge outside the branch range whose lead names the issue is still kept. Reword both to "the branch''s own merges" (as the atlas does), or drop merges by parent count for every window commit. Same family: the atlas verb list omits `closes`, which leadRE accepts.'
          family: doc-claim-matches-code
          round: 1
      recipe: small-diff-review
      reviewed: 57bb1c283b62ec2b90139fada28f40d2d98a2f83
      blocked: false
---

# Gate ledger — ariadne#317 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T17:14:29-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-matches-code` Help text and Scope doc say merges are never scoped boundaries; only BranchPoint..HEAD merges are dropped
  helptext/active-time.md "With --branch-point, merge commits are not boundaries" and commit.go Scope doc "a merge is never a scoped boundary". A main-side merge outside the branch range whose lead names the issue is still kept. Reword both to "the branch's own merges" (as the atlas does), or drop merges by parent count for every window commit. Same family: the atlas verb list omits `closes`, which leadRE accepts.

## Open findings

- **BR-1** [Minor] `doc-claim-matches-code` Help text and Scope doc say merges are never scoped boundaries; only BranchPoint..HEAD merges are dropped
