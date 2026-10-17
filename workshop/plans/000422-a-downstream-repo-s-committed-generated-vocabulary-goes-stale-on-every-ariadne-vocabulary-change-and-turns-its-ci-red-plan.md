---
collaboration-mode: discussion
---

# Shipped generated outputs Implementation Plan (ariadne#422)

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A consumer can ship a weave-generated file in git (parley's `construct/generated/vocabulary/issue.json`) without an ariadne vocabulary change turning its cold CI red, while a hand edit to that file is still refused.

**Architecture:** The consumer declares the outputs it ships with a new manifest row, `ship PATH`. For each declared staged output, weave also writes `PATH.weave-sha256`, a committed digest of what it wrote. Ownership in `prepareManaged` gains one rule, for the output and its sidecar as a pair: when a declared output's current bytes match its current sidecar, weave owns both, even with no inventory (a cold checkout) or a stale one (a `git pull` that brought another slot's committed refresh). Undeclared outputs behave exactly as today. Ignoring stays as it is: the consumer un-ignores the shipped pair with authored negations after weave's block, the mechanism parley already uses for issue.json (`managedIgnoreText` puts the block first so authored negations win).

**Tech Stack:** Go, `cmd/weave` (intent parser, plan lowering, managed ownership), `pkg/weaveownership` (digest).

**Why not the other options** (issue Spec): option 1 (stop committing) breaks parley's plugin installs, which read the vocabulary from the installed tree with no weave run (parley#208). Option 3 (a fleet check naming stale consumers) reports the breakage but doesn't prevent it. A blanket "tracked file = weave's" rule would drop the authored-replacement protection (ARCH-SECURE; ariadne#368's spec, superseded by this issue).

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `intent.Ship` | `cmd/weave/internal/intent/intent.go`, `manifest.go` | new |
| `ShippedPaths` | `cmd/weave/internal/plan/shipped.go` | new |
| `ShipStaged` | `cmd/weave/internal/plan/shipped.go` | new |
| `SidecarPath` | `cmd/weave/internal/plan/shipped.go` | new |
| `stagedOutput` | `cmd/weave/internal/plan/staged.go` | modified |

- **`intent.Ship`** — the `ship PATH` manifest verb: a declaration, lowering to no filesystem action (like `skill`). Visibility follows `intent.Selected`, so `internal ship` applies only on the consumer's own walk.
  - **Relationships:** N rows per manifest → a set of leaf-relative paths.
  - **DRY rationale:** reuses the manifest's visibility rule; no second config file.
  - **Future extensions:** other shipped artifact kinds (a skill a plugin bundles) reuse the same row.
- **`ShippedPaths(layers []walk.Layer) []string`** — the selected `ship` rows' cleaned, safe-relative sources, with a refusal for an unsafe path.
- **`ShipStaged(actions []Action, shipped []string) (out []Action, unmatched []string)`** — for each staged `WriteFile` whose path is declared, it marks the wrapper (`stagedOutput.Sidecar = SidecarPath(path)`) and appends a plain `WriteFile{Path: sidecar, Content: digest(content)+"\n", Mode: 0644}`. A declared path no staged output produces is returned as unmatched, so the caller can warn about a typo.
- **`SidecarPath(p) = p + ".weave-sha256"`** — the one place the sidecar name is defined.
- **`stagedOutput`** gains `Sidecar string` (empty: not shipped). Positional literals in `staged.go` become keyed.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `prepareManaged` pair rule | `cmd/weave/internal/plan/ownership.go` | modified | filesystem reads of the destination and its sidecar |
| compile wiring | `cmd/weave/main.go` | modified | the generation stage, stdout |

- **`prepareManaged` pair rule** — before the `wanted` loop, for each shipped wrapper (collected from the actions, as `strict` is, before `materializeManaged`): `vouched(fs, root, output, sidecar)` Lstats both (regular files only, through `safeParents`) and compares `digest(output bytes) == strings.TrimSpace(sidecar bytes)`. A vouched pair puts both paths in `paired`; in the loop, `own = own || paired[id.Path]`. So the output is overwritten (not refused cold, not displaced warm) and so is its sidecar (not displaced when the inventory recorded an older one). An unvouched pair (a hand edit, a missing or malformed sidecar) falls through to today's switch: refuse cold, displace warm.
  - **Injected into:** `weavefs.FS` (tests use `weavefs.OSFS{}` over `t.TempDir()`, as every ownership test does).
- **compile wiring** — after collecting `generated` (main.go ~548): `generated, unmatched := plan.ShipStaged(generated, plan.ShippedPaths(layers))`, then `weave: warning: ship row %s names no generated output` per unmatched path. The existing `weave: regenerated tracked output %s: commit it` line (printManagedReport) already tells the operator to commit the refreshed copy and sidecar.

**Durable state (ARCH-FUNERAL at-plan):** one 65-byte sidecar per declared output. weave writes it, the consumer's git tracks it, and it changes only when the output's bytes change. When the `ship` row goes, the sidecar is no longer produced: on a warm checkout weave's retirement removes it (it is inventoried); on a cold one it isn't inventoried, so it stays in git until the consumer's commit that drops the row also runs `git rm` on it (said in the atlas entry).

## Chunk 1: declaration and staging

### Task 1: `ship` manifest verb

**Files:** Modify `cmd/weave/internal/intent/intent.go` (Kind const, String), `cmd/weave/internal/intent/manifest.go` (`kindByVerb`), `cmd/weave/internal/plan/plan.go` (lowering switch: `case intent.Ship:` no action, with a comment), `cmd/weave/internal/golden/completeness.go` (`coverIntent`, `verbName`). Test: `cmd/weave/internal/intent/intent_test.go`.

- [x] Failing test `TestParseManifestShip`: `internal ship construct/generated/vocabulary/issue.json` parses to `Intent{Kind: Ship, Visibility: Internal, Source: "construct/generated/vocabulary/issue.json", Target: same}`.
- [x] `go test -run TestParseManifestShip ./cmd/weave/internal/intent/` → FAIL (unknown verb skipped).
- [x] Add `Ship` to the Kind iota, `"ship": Ship` to `kindByVerb`, the no-op lowering case and the golden cases.
- [x] `go test ./cmd/weave/internal/intent/ ./cmd/weave/internal/plan/ ./cmd/weave/internal/golden/` → PASS.

### Task 2: `ShippedPaths`, `ShipStaged`, `SidecarPath`

**Files:** Create `cmd/weave/internal/plan/shipped.go`, `cmd/weave/internal/plan/shipped_test.go`; modify `cmd/weave/internal/plan/staged.go` (`Sidecar` field, keyed literals).

- [x] Failing tests. Strategy: `ShippedPaths` over (row visibility × leaf/ancestor × safe/unsafe path), guard = only `Selected` safe rows; `ShipStaged` over (declared/undeclared/unmatched path), guard = exactly one sidecar per declared staged `WriteFile`, holding `digest(content)+"
"`, and unmatched paths returned.
- [x] Run → FAIL (undefined).
- [x] Implement (pure; no FS).
- [x] `go test -run 'Shipped|ShipStaged' ./cmd/weave/internal/plan/` → PASS.
- [x] Commit `#422: weave: ship manifest verb and sidecar staging`.

## Chunk 2: ownership

### Task 3: a vouched output and sidecar pair is weave's own

**Files:** Modify `cmd/weave/internal/plan/ownership.go` (`prepareManaged` collects the shipped wrappers before `materializeManaged`, as `strict` is, and builds `paired` from `vouched`; `vouched` helper). Test: `cmd/weave/internal/plan/shipped_test.go`.

- [x] Failing tests (`stagedFixture` + `ShipStaged`, `managedApply` in ScopeArtifacts). Strategy: the matrix (output bytes: generated / committed-old / hand-edited) × (sidecar: matching / stale / missing) × (inventory: none / recorded older pair) × (declared / undeclared), guard = overwrite only when the pair vouches or the inventory owns it; otherwise refuse cold, displace warm, bytes untouched. Named rows: the parley case (cold, old pair → regenerated, both rewritten); the pulled copy (warm, newer committed pair → regenerated, `Displaced` empty); cold hand edit → refused.
- [x] Run → the parley and pulled-copy rows FAIL.
- [x] Implement the rule and `vouched`.
- [x] `go test ./cmd/weave/internal/plan/` → PASS.

- [x] Commit `#422: weave: a vouched output and sidecar pair is weave's own`.

## Chunk 3: compile wiring, e2e, atlas

### Task 4: compile applies `ship` rows

**Files:** Modify `cmd/weave/main.go` (~548: `ShipStaged` + unmatched warning). Test: `cmd/weave/startup_test.go`, modelled on `TestCompileRefusalLeavesCheckoutUntouched`.

- [x] Failing e2e `TestCompileRegeneratesShippedOutputCold` (base layer with a dynamic skill writing `SKILL.md`; leaf manifest `internal ship construct/generated/example/SKILL.md`; no inventory). Strategy: (committed pair vouches / no `ship` row / `ship` row naming nothing generated), guard = regenerate / today's refusal / the warning line.
- [x] Run → FAIL; implement; `go test -run 'TestCompile' ./cmd/weave/` → PASS.
- [x] Commit `#422: weave: compile ships declared outputs`.

### Task 5: atlas and help

**Files:** `atlas/workflow/weave.md` (ownership section: the `ship` row, the sidecar, the vouching rule, why), the manifest format comment in `construct/base.manifest` (verb list).

- [x] Document; commit `#422: atlas: shipped generated outputs`.

### Task 6: verification

- [x] `go test ./cmd/weave/...` and `go vet ./cmd/weave/...`.
- [x] Full `make test` in the background before the close.
- [x] Live, against parley's real case, without writing parley: copy parley's tracked tree into a scratch dir (`git archive`), add the `ship` row and a sidecar of the committed issue.json, and run this branch's weave compile cold (no inventory) against current ariadne → issue.json regenerates, and the refresh line prints. Record in the issue Log.
- [x] Hand parley#371 (filed, depends on #422) to the TL for routing after landing.

## Revisions

- 2026-10-17, plan-quality round 1: PQ-1 (Important). The sidecar is inventoried after the first apply, so a pulled
  newer sidecar would be displaced. The rule now covers the pair: a vouched output and sidecar are both weave's own.
  Minors: dropped the `.gitignore` keep set (Task 4) in favour of the existing authored-negation mechanism, the one
  parley already uses and parley#371 specifies; the removal note is scoped to warm checkouts and filed under
  ARCH-FUNERAL; test lists compressed to strategy lines.
- 2026-10-17, close review round 1 (SHIP, 3 Minors, all fixed): `ShippedPaths` takes `[]layer.Layer` and returns
  `([]string, error)` (an unsafe `ship` path refuses), not `walk.Layer` → `[]string` as Core concepts says. The
  unmatched-row warning names the reason ("matches no file this compile generates (only generated regular files
  ship)"), covering a symlink output and an exported row reaching a consumer that doesn't generate it. A symlinked
  sidecar is tested (refused cold).
