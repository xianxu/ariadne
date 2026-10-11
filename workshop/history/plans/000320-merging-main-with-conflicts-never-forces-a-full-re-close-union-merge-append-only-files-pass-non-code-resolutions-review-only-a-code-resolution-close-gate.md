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
    - "n": 2
      timestamp: "2026-10-10T17:50:13-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: TestCompileEnsuresGitattributes (cmd/weave/main_test.go) runs a real compile twice and pins the exact block; passes.
          round: 2
        - id: BR-2
          disposition: addressed
          note: splitIgnore errors now say "weave generated block" (managed_ignore.go:35,41,55), neutral across both files.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Caveat now in the UnionMergeAttributes doc comment (gitattributes.go:14-16) and atlas/workflow/weave.md.
          round: 2
        - id: BR-4
          disposition: addressed
          note: d.Markers populated at publishgate.go:223 and refused in classifyPublishDelta; the integration subtest "committed with its markers refuses" goes red without it.
          round: 2
      recipe: milestone-review
      reviewed: 94dbc3d8a9f7450d44e1a04240df8ca1e55aff22
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

## Round 2 — 2026-10-10T17:50:13-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — TestCompileEnsuresGitattributes (cmd/weave/main_test.go) runs a real compile twice and pins the exact block; passes.
- BR-2 — addressed — splitIgnore errors now say "weave generated block" (managed_ignore.go:35,41,55), neutral across both files.
- BR-3 — addressed — Caveat now in the UnionMergeAttributes doc comment (gitattributes.go:14-16) and atlas/workflow/weave.md.
- BR-4 — addressed — d.Markers populated at publishgate.go:223 and refused in classifyPublishDelta; the integration subtest "committed with its markers refuses" goes red without it.

## Open findings

(none — every finding has been disposed)
