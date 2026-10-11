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
    - "n": 2
      timestamp: "2026-10-10T17:16:29-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'ed89620a: helptext says "the branch''s own merge commits (BranchPoint..HEAD)", Scope doc says "a merge in BranchPoint..HEAD", atlas says "its own merge commits (branch point..HEAD)" and its verb list now includes closes, matching leadRE and the branchCommits filter.'
          round: 2
      recipe: small-diff-review
      reviewed: ed89620af14a2b72377ccdf49f9d10c588a00a23
      blocked: false
---

# Gate ledger — ariadne#317 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T17:14:29-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-matches-code` Help text and Scope doc say merges are never scoped boundaries; only BranchPoint..HEAD merges are dropped
  helptext/active-time.md "With --branch-point, merge commits are not boundaries" and commit.go Scope doc "a merge is never a scoped boundary". A main-side merge outside the branch range whose lead names the issue is still kept. Reword both to "the branch's own merges" (as the atlas does), or drop merges by parent count for every window commit. Same family: the atlas verb list omits `closes`, which leadRE accepts.

## Round 2 — 2026-10-10T17:16:29-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — ed89620a: helptext says "the branch's own merge commits (BranchPoint..HEAD)", Scope doc says "a merge in BranchPoint..HEAD", atlas says "its own merge commits (branch point..HEAD)" and its verb list now includes closes, matching leadRE and the branchCommits filter.

## Open findings

(none — every finding has been disposed)
