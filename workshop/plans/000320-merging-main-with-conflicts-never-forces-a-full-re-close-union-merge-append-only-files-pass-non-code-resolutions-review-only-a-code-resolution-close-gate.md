---
gate: boundary-review
issue: 320
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T17:23:58-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: EnsureGitattributes is tested via plan.Apply only; no planActions/ApplyManaged test asserts .gitattributes is written
          detail: By reading, the ApplyManaged default branch passes the action through to Apply, but weave compile never ran (sandbox) and no compile-level test pins the planActions wiring.
          family: wiring-untested-at-entrypoint
          round: 1
        - id: BR-2
          severity: Minor
          title: splitIgnore/managedIgnoreText errors say "gitignore block" when the bad block is in .gitattributes
          family: reused-helper-misleading-name
          round: 1
        - id: BR-3
          severity: Minor
          title: The "land lessons.md compactions alone" caveat for union merge is only in the issue Log, not in UnionMergeAttributes' doc comment or atlas
          family: union-merge-invariant-undocumented
          round: 1
        - id: BR-4
          severity: Minor
          title: A non-code conflict resolution that still contains conflict markers passes the publish gate
          family: unreviewed-resolution-content
          round: 1
      recipe: milestone-review
      reviewed: aca4448eae6b42f1c50b1d0beefe53c5b90cd734
      blocked: false
---

# Gate ledger — ariadne#320 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T17:23:58-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `wiring-untested-at-entrypoint` EnsureGitattributes is tested via plan.Apply only; no planActions/ApplyManaged test asserts .gitattributes is written
  By reading, the ApplyManaged default branch passes the action through to Apply, but weave compile never ran (sandbox) and no compile-level test pins the planActions wiring.
- **BR-2** [Minor] `reused-helper-misleading-name` splitIgnore/managedIgnoreText errors say "gitignore block" when the bad block is in .gitattributes
- **BR-3** [Minor] `union-merge-invariant-undocumented` The "land lessons.md compactions alone" caveat for union merge is only in the issue Log, not in UnionMergeAttributes' doc comment or atlas
- **BR-4** [Minor] `unreviewed-resolution-content` A non-code conflict resolution that still contains conflict markers passes the publish gate

## Open findings

- **BR-1** [Minor] `wiring-untested-at-entrypoint` EnsureGitattributes is tested via plan.Apply only; no planActions/ApplyManaged test asserts .gitattributes is written
- **BR-2** [Minor] `reused-helper-misleading-name` splitIgnore/managedIgnoreText errors say "gitignore block" when the bad block is in .gitattributes
- **BR-3** [Minor] `union-merge-invariant-undocumented` The "land lessons.md compactions alone" caveat for union merge is only in the issue Log, not in UnionMergeAttributes' doc comment or atlas
- **BR-4** [Minor] `unreviewed-resolution-content` A non-code conflict resolution that still contains conflict markers passes the publish gate
