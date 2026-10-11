---
gate: boundary-review
issue: 321
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T17:34:14-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Atlas wording omits that lead-less tracker commits are also dropped unscoped
          detail: With Scope.Issue set, a tracker-only commit with no lead (a neutral boundary) is now filtered too, since names is false. The atlas and Scope comment describe only the other-issue case.
          family: doc-claim-precision
          round: 1
        - id: BR-2
          severity: Minor
          title: besideHead rev-lists the entire tracker history on every measurement
          detail: Not window-bounded; negligible now and git log already reads it all, but worth noting if the tracker grows.
          family: unbounded-history-read
          round: 1
      recipe: small-diff-review
      reviewed: a975fa97db1c2aab24ad6a5bd31181ad87d7b431
      blocked: false
---

# Gate ledger — ariadne#321 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T17:34:14-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-precision` Atlas wording omits that lead-less tracker commits are also dropped unscoped
  With Scope.Issue set, a tracker-only commit with no lead (a neutral boundary) is now filtered too, since names is false. The atlas and Scope comment describe only the other-issue case.
- **BR-2** [Minor] `unbounded-history-read` besideHead rev-lists the entire tracker history on every measurement
  Not window-bounded; negligible now and git log already reads it all, but worth noting if the tracker grows.

## Open findings

- **BR-1** [Minor] `doc-claim-precision` Atlas wording omits that lead-less tracker commits are also dropped unscoped
- **BR-2** [Minor] `unbounded-history-read` besideHead rev-lists the entire tracker history on every measurement
