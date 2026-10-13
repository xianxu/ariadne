# Boundary Review — ariadne#265 (whole-issue close)

| field | value |
|-------|-------|
| issue | 265 — weave robustness: atomic compile, symlinked Makefile, legacy include, dependency writes |
| repo | ariadne |
| issue file | workshop/issues/000265-weave-robustness.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5655861efb765f10a5e5e9bd514dc3743c23fc56..4e2fbb31019f4a8bebbed683df4cc281c5081d86 |
| command | sdlc close --issue 265 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T18:34:54-07:00 |
| verdict | REWORK |

## Review

Checking whether `prepareManaged` writes anything, then testing a suspected false refusal when a path moves from a data mount to a generated output.

Prepare is write-free, but I confirmed a false refusal: a path that moves from a data mount to a generated output is now refused on every compile. Running the package tests next.

These failures come from the environment: TMPDIR points inside the review checkout. Re-running the new compile tests with a normal TMPDIR.

The harness blocks `git init` outside its own directories, so the compile-level tests can't run here. Last check: does `ProducedPathSet` see generated (staged) outputs in the collision check?

```verdict
verdict: REWORK
confidence: medium
```

**Verdict: REWORK.** Items 3 and 4 are correct, and the prepare/commit split of `ApplyManaged` is clean. But the new check-everything-first order in `compilePrepared` checks the artifacts pass against the state *before* the data passes run. When a path moves from a data mount to a generated output, that check refuses every compile, even though the old order handled the move without error. Nothing inside weave can clear it, because the data pass that would retire the old mount never runs. A related problem: the collision guard only compares exact paths, so a mount over a *parent* directory of a generated output still produces a partial write before the artifacts pass refuses. That breaks the "refused compile writes nothing" promise. Both come from one cause: the artifacts check doesn't account for what the data pass is about to change.

**Inspection limits:** I could not run the compile-level tests in `cmd/weave/startup_test.go` in this checkout. The review harness refuses `git init` outside its own directories, and with TMPDIR inside the checkout, `rev-parse` fails. The exception is `TestCompileWritesNothingInMountlessDependency`, which ran and passes on the head and on the reverted `main.go`, as the plan expects for something #264 already fixed. So I could not confirm by reverting that the item-1 test fails without the fix. The plan-level evidence below comes from a scratch test I ran and then deleted. All `internal/...` packages pass.

1. **Strengths**
   - `cmd/weave/internal/plan/ownership.go:155-272`: `prepareManaged` only reads; every write is in `commitManaged`, and `ApplyManaged` and `CheckManaged` share one inspection routine (ARCH-DRY pass). I checked that `materializeManaged` doesn't write either.
   - `cmd/weave/internal/startup/tools.go:17-27`: switching to `Lstat` with a refusal that names the link is the right small fix, and the test confirms the linked Makefile's recipe never runs.
   - `cmd/weave/internal/plan/apply.go:126-130`: `TrimPrefix("-")` plus a switch covers both `include` and `-include` for both forms, and the test covers indented lines too.
   - The atlas was updated for the new compile order and the symlink refusal.

2. **Critical**
   - **`cmd/weave/main.go:567-585` — false refusal that weave cannot recover from (ARCH-PURPOSE).**
     - `CheckManaged(root, actions, ScopeArtifacts)` reads the inventory before the root data pass runs. Any wanted artifact path still recorded under `data` fails with "output X belongs to data" (`ownership.go:212-214`).
     - In the old order, the data pass retired that entry first, so the artifacts pass succeeded.
     - **Reproduced at plan level:** after `ApplyManaged(data, [Symlink x.md])`, `CheckManaged(artifacts, [WriteFile x.md])` returns `output x.md belongs to data`. The old sequence, `ApplyManaged(data, nil)` then `ApplyManaged(artifacts, …)`, returns nil.
     - Every re-run fails the same way. Recovering means editing `ownership.json` by hand.
     - **Fix sketch:** prepare the artifacts pass against the root data pass's projected state. For example, give `prepareManaged` an optional projected inventory (the data pass's `other+wanted`) and the data pass's produced paths. That also makes the ad hoc collision loop unnecessary (same family as the next finding).

3. **Important**
   - **`cmd/weave/main.go:574-579` — the collision guard only compares exact paths, so parent/child overlaps still write before refusing.**
     - Example: a data mount at `construct/generated` and a generated output at `construct/generated/x/SKILL.md`. The artifacts check passes, because no symlink parent exists yet.
     - The data apply then commits its symlink, inventory and `.gitignore`, and the artifacts apply fails in `safeParents`. That is the non-atomic refusal this issue set out to remove.
     - **Same family:** the false refusal above. Both are the artifacts check ignoring the data pass's effects, and the projected-state fix covers both.
   - **No test for the transition case.** Add a compile-level test where a data mount is first applied at path X, then the deps drop it and a generator produces X. The compile should succeed.

4. **Minor**
   - `main.go:573`: `ProducedPathSet(actions)` is recomputed inside the loop. It only runs for root, but it could be hoisted.
   - `CheckManaged` skips `ReclaimPublications`, while `ApplyManaged` reclaims before preparing. This is harmless today, but the doc comment says it runs "every refusal ApplyManaged would", which isn't quite true.

5. **Test coverage**
   - Each Done-when clause has a test: atomic refusal, mount-less dependency, symlink refusal, and indirect include.
   - Both states are covered for items 3 and 4. Regular Makefiles are covered by existing tests, and so is the direct include.
   - What's missing is the case where the check order changes the result (the cross-pass transition above).

6. **Architecture**
   - **ARCH-DRY: pass.** One shared prepare routine. The collision loop is a second, partial copy of overlap logic that `prepareManaged`/`safeParents` already have; it should go away once the projected-state fix lands.
   - **ARCH-PURE: pass.** Prepare is a read-only inspection and commit holds all the writes. The orchestration in `main.go` is sensible.
   - **ARCH-PURPOSE: flag.** "Check everything before writing" is only atomic if the check models what earlier passes will change. Right now it models each pass separately, which gives both false refusals and missed refusals.

7. **Plan revisions**
   - Add a `## Revisions` entry: "The artifacts check is prepared against the root data pass's projected inventory and produced paths, not the current state on disk. This replaces the exact-path collision check and covers data→artifact path transitions."
   - The Log's statement that "each pass is inspected twice" should also say that the second inspection is the authoritative one, and that the first must not be stricter than it.

```findings
findings:
  - id: new
    severity: Critical
    family: cross-pass-check-ignores-prior-pass-effects
    title: |
      Up-front artifacts check refuses a data-to-artifact path transition the old order converged (permanent, unrecoverable refusal)
    detail: |
      main.go:581 CheckManaged(root, artifacts) reads the pre-data inventory; a wanted path still recorded under scope data fails "output X belongs to data" (ownership.go:212), though the root data pass would retire it first. Reproduced at plan level: CheckManaged errors, while ApplyManaged(data,nil) then ApplyManaged(artifacts) succeeds. Since the data pass never runs, every re-run refuses. Fix: prepare artifacts against the data pass's projected inventory/produced paths; add a compile-level transition test.
  - id: new
    severity: Important
    family: cross-pass-check-ignores-prior-pass-effects
    title: |
      Data-vs-artifact collision guard is exact-path only; parent/child overlaps still commit the data pass before the artifacts refusal
    detail: |
      main.go:574-579 compares ProducedPathSet keys only. A data mount at construct/generated with a generated construct/generated/x/SKILL.md passes both checks (no symlink parent exists yet), then data apply writes its mount, inventory and .gitignore, and artifacts apply fails in safeParents, so the refusal is not atomic. Family: the same root cause as the transition false refusal; the projected-state prepare covers both and replaces this ad hoc loop (ARCH-DRY).
  - id: new
    severity: Minor
    family: doc-claim-overstates-behavior
    title: |
      CheckManaged doc says it runs every refusal ApplyManaged would, but it skips ReclaimPublications
    detail: |
      ownership.go:131 vs ApplyManaged:144. Harmless today; reword or note the omission.
  - id: new
    severity: Minor
    family: loop-invariant-recompute
    title: |
      ProducedPathSet(actions) recomputed inside the per-owner loop in compilePrepared
    detail: |
      main.go:573; hoist above the loop (moot if the collision loop is replaced).
```
