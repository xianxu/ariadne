# Boundary Review — ariadne#155 (whole-issue close)

| field | value |
|-------|-------|
| issue | 155 — weave: fresh-bootstrapped derivative silently under-compiles (missing base.manifest breaks transitive walk) |
| repo | ariadne |
| issue file | workshop/issues/000155-weave-fresh-bootstrapped-derivative-silently-under-compiles-missing-base-manifest-breaks-transitive-walk.md |
| boundary | whole-issue close |
| milestone | — |
| window | 968ab7eeefc222c350cb22ac613db6361eefcb83..HEAD |
| command | sdlc close --issue 155 |
| reviewer | claude |
| timestamp | 2026-07-06T10:22:58-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All green: gofmt clean, vet clean, full suite passes. I have a complete picture.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Both fixes land as specified and the purpose is fully delivered: a fresh-bootstrapped derivative can no longer silently under-compile. Fix 1 (loud-fail) is correctly placed in the shared `pkg/layergraph/discoverEdges`, so it covers all three `Walk` consumers (weave, datatype, vocabulary — each propagates the error with context) from one source; Fix 2 (`weave link` seeds `base.manifest`) is idempotent and single-sourced. I confirmed `ParseDeps` yields only `substrate` rows, so the loud-fail cannot regress `data` deps. Tests are green (gofmt/vet clean, full suite passes), and `TestLinkSeedMakesRepoTraversable` is a genuine end-to-end proof (the seeded manifest actually parses and composes a 3-layer chain, not just a string check). Nothing blocks SHIP; the findings are a stale doc comment plus two cheap coverage/doc gaps.

### 1. Strengths
- **Loud-fail placed at the single shared source** (`pkg/layergraph/walk.go:61-77`) rather than in `weave verify-complete` as the Spec's Option 1 loosely suggested — this is the better call: `cmd/datatype/main.go:115` and `cmd/vocabulary/resolve.go:27` share the same footgun and both now propagate the error with `walk layer graph from %s: %w`. ARCH-DRY, well-reasoned.
- **Present/absent split is precise** (`pathExists`, `pkg/layergraph/walk.go:129-136`): a present-but-manifest-less substrate is a broken layer edge (loud), while an absent peer (partial checkout) keeps the legitimate silent present-skip. `TestWalkAbsentSubstrateSilentlySkipped` pins the distinction.
- **The exact reported footgun is regression-tested** — `TestWalkChainBrokenByManifestlessIntermediate` reproduces kbench→kaggle→metis with the manifest-less mid and asserts the error names the broken intermediate.
- **`TestLinkSeedMakesRepoTraversable` proves the seed, not just the file** — it walks *through* a link-seeded mid to a 3-layer chain, so it exercises `intent.ParseManifest` on the seeded content (a broken seed would fail `loadLayer`).
- **`seededBaseManifest` is a pure string function**, single-sourced (`cmd/weave/main.go:297`), and the seed never clobbers a hand-authored manifest (`TestLinkSeedNeverClobbersExisting`). Clean ARCH-PURE separation: `ensureBaseManifest` takes injected `fs`+`out`.

### 2. Critical findings
None.

### 3. Important findings
None blocking.

### 4. Minor findings
- **Stale doc comment on `runLink`** (`cmd/weave/main.go:226-227`): "Read-only on everything but the one deps file." is now false — `runLink` also writes `construct/base.manifest` via `ensureBaseManifest`. This is a base-layer file that propagates downstream; update the comment to name both writes.
- **`buildLink` `Short` help omits the seed** (`cmd/weave/main.go:207`): still reads "Record `substrate <path>` (verbatim) in this repo's construct/deps" though `link` now also seeds `base.manifest`. Runtime output announces the seed, so low impact, but the one-liner under-describes the command's effect.
- Seed header interpolates the raw `substratePath` into `#` comment lines — harmless (comment-stripped by `ParseDeps`/`ParseManifest`), noted only for completeness; a path with an embedded newline is not a realistic CLI arg.

### 5. Test coverage notes
- **The "repair a pre-#155 repo" path is exercised but unasserted.** The code explicitly commits to seeding even when the deps row is already present (`cmd/weave/main.go:268-270`), and `TestLinkIdempotent` (main_test.go:807) hits that path — but it only asserts `construct/deps` is unchanged, never that `base.manifest` was seeded. A regression that re-added an early `return nil` before `ensureBaseManifest` would pass every test. Cheap fix: extend `TestLinkIdempotent` (or add a case) to assert the manifest is seeded when the row pre-exists but the manifest is absent. Low real-world impact (only affects re-linking an already-broken repo), hence Minor.
- Otherwise coverage is strong: walk error path, absent-skip, chain-break repro, seed-when-absent, never-clobber, and end-to-end traversability are all pinned.

### 6. Architectural notes for upcoming work
- **ARCH-DRY — pass.** Fix consolidated at the shared walk; seed content single-sourced in `seededBaseManifest`. Note the deliberate residual: existing #95-cutover repos keep their own hand-authored manifests (not derived from the seed) — acceptable, since the seed governs *new* repos and the header comment says so.
- **ARCH-PURE — pass.** Logic sits behind injected `FS`/`io.Writer`; `seededBaseManifest` is pure. Walk tests are INTEGRATION-level over `t.TempDir()`/`OSFS{}`, appropriate for the IO seam.
- **ARCH-PURPOSE — pass (shadow-sweep clean).** Both preferred options (loud-fail + seed) shipped, not the cheap subset; all three Done-when items are delivered and tested. The loud-fail's single source reaches every consumer, so no consumer is left as a "documentation-only" restatement of the layer contract.

### 7. Plan revision recommendations
None — the plan's two checklist items match the code, and the Core-concepts-equivalent entities (`discoverEdges`, `pathExists`, `ensureBaseManifest`, `seededBaseManifest`) all exist at the stated paths with the claimed behavior. The atlas (`atlas/workflow/weave.md`) was updated in-range for the new rule + seed companion, satisfying the docs gate.
