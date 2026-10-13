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
    - "n": 2
      timestamp: "2026-10-12T18:53:56-07:00"
      dispatched: "2026-10-12T18:40:03-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Projected inventory plus retired-path skip; reverting both turns TestCompileMovesPathFromDataMountToGeneratedOutput red. The skip alone is not pinned (new finding).
          round: 2
        - id: BR-2
          disposition: addressed
          note: Overlap check at/above/below replaces the exact-path loop; removing it reddens both compile collision tests, removing the descendant scan reddens TestCheckManagedSeesEarlierPassOutputs.
          round: 2
        - id: BR-3
          disposition: addressed
          note: ownership.go CheckManaged doc now says publication stages are reclaimed only by an apply.
          round: 2
        - id: BR-4
          disposition: addressed
          note: 'Moot: the per-owner loop was replaced by CheckManaged; prepareManaged computes ProducedPathSet once.'
          round: 2
      findings:
        - id: BR-5
          severity: Important
          title: Retired-path skip in prepareManaged (ownership.go:268-273) is reached by no committed test
          detail: Reverting only the within(id.Path, retired) skip leaves every committed test green, because the BR-1 transition test uses a non-strict manifest symlink that the projected inventory alone handles. A throwaway test (data mount at construct/generated/example, then a generator writing construct/generated/example/SKILL.md) passes at head and fails with the skip disabled ("non-directory parent"), so it is a permanent refusal. Commit that compile-level test.
          family: fix-branch-without-regression-fixture
          round: 2
        - id: BR-6
          severity: Minor
          title: Check-time projection restates commitManaged's retirement rule instead of sharing it
          detail: prepareManaged's after loop encodes retire = previous and not produced, the same rule commitManaged applies, so a change to one can drift from the other. Share a helper, or cross-reference the two in a comment. Belongs only loosely to the family (a duplicated derivation, not a recomputation).
          family: loop-invariant-recompute
          round: 2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 3
      timestamp: "2026-10-12T19:07:28-07:00"
      dispatched: "2026-10-12T18:56:14-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: TestCompileMovesDataMountToGeneratedParent (cmd/weave/startup_test.go:279) fails with the within(id.Path, retired) skip disabled ("non-directory parent construct/generated/example") and passes at head; verified by revert in the pinned checkout.
          round: 3
        - id: BR-6
          disposition: addressed
          note: managedPass.retiring() (ownership.go:156) is the single retirement rule, used by prepareManaged's projection (ownership.go:226) and commitManaged's retirement loop (ownership.go:390).
          round: 3
      recipe: milestone-review
      isolation: checkout
      reviewed: 23728ab3a28a815633f4ed6a27624a8b4f140719
      blocked: false
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

## Round 2 — 2026-10-12T18:53:56-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Projected inventory plus retired-path skip; reverting both turns TestCompileMovesPathFromDataMountToGeneratedOutput red. The skip alone is not pinned (new finding).
- BR-2 — addressed — Overlap check at/above/below replaces the exact-path loop; removing it reddens both compile collision tests, removing the descendant scan reddens TestCheckManagedSeesEarlierPassOutputs.
- BR-3 — addressed — ownership.go CheckManaged doc now says publication stages are reclaimed only by an apply.
- BR-4 — addressed — Moot: the per-owner loop was replaced by CheckManaged; prepareManaged computes ProducedPathSet once.

### Raised

- **BR-5** [Important] `fix-branch-without-regression-fixture` Retired-path skip in prepareManaged (ownership.go:268-273) is reached by no committed test
  Reverting only the within(id.Path, retired) skip leaves every committed test green, because the BR-1 transition test uses a non-strict manifest symlink that the projected inventory alone handles. A throwaway test (data mount at construct/generated/example, then a generator writing construct/generated/example/SKILL.md) passes at head and fails with the skip disabled ("non-directory parent"), so it is a permanent refusal. Commit that compile-level test.
- **BR-6** [Minor] `loop-invariant-recompute` Check-time projection restates commitManaged's retirement rule instead of sharing it
  prepareManaged's after loop encodes retire = previous and not produced, the same rule commitManaged applies, so a change to one can drift from the other. Share a helper, or cross-reference the two in a comment. Belongs only loosely to the family (a duplicated derivation, not a recomputation).

## Round 3 — 2026-10-12T19:07:28-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — TestCompileMovesDataMountToGeneratedParent (cmd/weave/startup_test.go:279) fails with the within(id.Path, retired) skip disabled ("non-directory parent construct/generated/example") and passes at head; verified by revert in the pinned checkout.
- BR-6 — addressed — managedPass.retiring() (ownership.go:156) is the single retirement rule, used by prepareManaged's projection (ownership.go:226) and commitManaged's retirement loop (ownership.go:390).

## Open findings

(none — every finding has been disposed)
