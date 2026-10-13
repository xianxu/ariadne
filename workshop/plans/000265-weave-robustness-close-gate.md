---
gate: boundary-review
issue: 265
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T18:34:54-07:00"
      dispatched: "2026-10-12T18:23:00-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: Up-front artifacts check refuses a data-to-artifact path transition the old order converged (permanent, unrecoverable refusal)
          detail: 'main.go:581 CheckManaged(root, artifacts) reads the pre-data inventory; a wanted path still recorded under scope data fails "output X belongs to data" (ownership.go:212), though the root data pass would retire it first. Reproduced at plan level: CheckManaged errors, while ApplyManaged(data,nil) then ApplyManaged(artifacts) succeeds. Since the data pass never runs, every re-run refuses. Fix: prepare artifacts against the data pass''s projected inventory/produced paths; add a compile-level transition test.'
          family: cross-pass-check-ignores-prior-pass-effects
          round: 1
        - id: BR-2
          severity: Important
          title: Data-vs-artifact collision guard is exact-path only; parent/child overlaps still commit the data pass before the artifacts refusal
          detail: 'main.go:574-579 compares ProducedPathSet keys only. A data mount at construct/generated with a generated construct/generated/x/SKILL.md passes both checks (no symlink parent exists yet), then data apply writes its mount, inventory and .gitignore, and artifacts apply fails in safeParents, so the refusal is not atomic. Family: the same root cause as the transition false refusal; the projected-state prepare covers both and replaces this ad hoc loop (ARCH-DRY).'
          family: cross-pass-check-ignores-prior-pass-effects
          round: 1
        - id: BR-3
          severity: Minor
          title: CheckManaged doc says it runs every refusal ApplyManaged would, but it skips ReclaimPublications
          detail: ownership.go:131 vs ApplyManaged:144. Harmless today; reword or note the omission.
          family: doc-claim-overstates-behavior
          round: 1
        - id: BR-4
          severity: Minor
          title: ProducedPathSet(actions) recomputed inside the per-owner loop in compilePrepared
          detail: main.go:573; hoist above the loop (moot if the collision loop is replaced).
          family: loop-invariant-recompute
          round: 1
      recipe: small-diff-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#265 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T18:34:54-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `cross-pass-check-ignores-prior-pass-effects` Up-front artifacts check refuses a data-to-artifact path transition the old order converged (permanent, unrecoverable refusal)
  main.go:581 CheckManaged(root, artifacts) reads the pre-data inventory; a wanted path still recorded under scope data fails "output X belongs to data" (ownership.go:212), though the root data pass would retire it first. Reproduced at plan level: CheckManaged errors, while ApplyManaged(data,nil) then ApplyManaged(artifacts) succeeds. Since the data pass never runs, every re-run refuses. Fix: prepare artifacts against the data pass's projected inventory/produced paths; add a compile-level transition test.
- **BR-2** [Important] `cross-pass-check-ignores-prior-pass-effects` Data-vs-artifact collision guard is exact-path only; parent/child overlaps still commit the data pass before the artifacts refusal
  main.go:574-579 compares ProducedPathSet keys only. A data mount at construct/generated with a generated construct/generated/x/SKILL.md passes both checks (no symlink parent exists yet), then data apply writes its mount, inventory and .gitignore, and artifacts apply fails in safeParents, so the refusal is not atomic. Family: the same root cause as the transition false refusal; the projected-state prepare covers both and replaces this ad hoc loop (ARCH-DRY).
- **BR-3** [Minor] `doc-claim-overstates-behavior` CheckManaged doc says it runs every refusal ApplyManaged would, but it skips ReclaimPublications
  ownership.go:131 vs ApplyManaged:144. Harmless today; reword or note the omission.
- **BR-4** [Minor] `loop-invariant-recompute` ProducedPathSet(actions) recomputed inside the per-owner loop in compilePrepared
  main.go:573; hoist above the loop (moot if the collision loop is replaced).

## Open findings

- **BR-1** [Critical] `cross-pass-check-ignores-prior-pass-effects` Up-front artifacts check refuses a data-to-artifact path transition the old order converged (permanent, unrecoverable refusal)
- **BR-2** [Important] `cross-pass-check-ignores-prior-pass-effects` Data-vs-artifact collision guard is exact-path only; parent/child overlaps still commit the data pass before the artifacts refusal
- **BR-3** [Minor] `doc-claim-overstates-behavior` CheckManaged doc says it runs every refusal ApplyManaged would, but it skips ReclaimPublications
- **BR-4** [Minor] `loop-invariant-recompute` ProducedPathSet(actions) recomputed inside the per-owner loop in compilePrepared
