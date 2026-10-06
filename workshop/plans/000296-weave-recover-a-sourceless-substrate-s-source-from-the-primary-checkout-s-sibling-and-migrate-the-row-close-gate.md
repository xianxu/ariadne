---
gate: boundary-review
issue: 296
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-06T09:18:03-07:00"
      agent: claude
      recipe: milestone-review
      blocked: false
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-10-06T09:24:02-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: README omits that compile/dependencies in a primary checkout rewrite construct/deps and that slots recover sourceless rows
          detail: README.md:80-85 still says to record a source by hand to restore a missing one. Compile now edits a tracked file in a primary checkout, and slots recover from the primary-side sibling. Both are user-visible and should be described there; only atlas/ was updated.
          family: readme-user-surface-sync
          round: 2
        - id: BR-2
          severity: Important
          title: The "own checkout with a remote origin" check is written twice (recoverSource and remoteOrigin) with different reporting
          detail: acquire.go:398-427 and migrate.go:80-91 both check own toplevel, origin, parse, and not-file, using different git readers and path comparisons. remoteOrigin reduces this to a bool, so the migration warning wrongly says "has no remote origin" for local, credentialed or invalid origins, and it bypasses the injected Client.Git seam. Export one acquire.Client.RemoteOrigin(ctx, dir) (Source, error) that returns the reason, and use it in both places.
          family: shared-helper-not-extracted
          round: 2
        - id: BR-3
          severity: Minor
          title: The error outside a numbered environment is extended, but the plan says it stays unchanged
          detail: The code adds "(no source declared; no numbered environment to recover it from)". Either note this in the plan's Revisions or keep the old message.
          family: plan-code-drift
          round: 2
        - id: BR-4
          severity: Minor
          title: migrateSourceless runs the primaryCheckout git probes on every compile, even with no sourceless rows
          family: avoidable-repeated-work
          round: 2
        - id: BR-5
          severity: Minor
          title: finishRestore takes eight parameters and passes Restore's err straight through
          family: function-signature-shape
          round: 2
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#296 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-06T09:18:03-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-10-06T09:24:02-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `readme-user-surface-sync` README omits that compile/dependencies in a primary checkout rewrite construct/deps and that slots recover sourceless rows
  README.md:80-85 still says to record a source by hand to restore a missing one. Compile now edits a tracked file in a primary checkout, and slots recover from the primary-side sibling. Both are user-visible and should be described there; only atlas/ was updated.
- **BR-2** [Important] `shared-helper-not-extracted` The "own checkout with a remote origin" check is written twice (recoverSource and remoteOrigin) with different reporting
  acquire.go:398-427 and migrate.go:80-91 both check own toplevel, origin, parse, and not-file, using different git readers and path comparisons. remoteOrigin reduces this to a bool, so the migration warning wrongly says "has no remote origin" for local, credentialed or invalid origins, and it bypasses the injected Client.Git seam. Export one acquire.Client.RemoteOrigin(ctx, dir) (Source, error) that returns the reason, and use it in both places.
- **BR-3** [Minor] `plan-code-drift` The error outside a numbered environment is extended, but the plan says it stays unchanged
  The code adds "(no source declared; no numbered environment to recover it from)". Either note this in the plan's Revisions or keep the old message.
- **BR-4** [Minor] `avoidable-repeated-work` migrateSourceless runs the primaryCheckout git probes on every compile, even with no sourceless rows
- **BR-5** [Minor] `function-signature-shape` finishRestore takes eight parameters and passes Restore's err straight through

## Open findings

- **BR-1** [Important] `readme-user-surface-sync` README omits that compile/dependencies in a primary checkout rewrite construct/deps and that slots recover sourceless rows
- **BR-2** [Important] `shared-helper-not-extracted` The "own checkout with a remote origin" check is written twice (recoverSource and remoteOrigin) with different reporting
- **BR-3** [Minor] `plan-code-drift` The error outside a numbered environment is extended, but the plan says it stays unchanged
- **BR-4** [Minor] `avoidable-repeated-work` migrateSourceless runs the primaryCheckout git probes on every compile, even with no sourceless rows
- **BR-5** [Minor] `function-signature-shape` finishRestore takes eight parameters and passes Restore's err straight through
