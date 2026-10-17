---
collaboration-mode: discussion
---

# Shipped generated outputs Implementation Plan (ariadne#422)

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A consumer can ship a weave-generated file in git (parley's `construct/generated/vocabulary/issue.json`) without an ariadne vocabulary change turning its cold CI red, while a hand edit to that file is still refused.

**Architecture:** The consumer declares the outputs it ships with a new manifest row, `ship PATH`. For each declared staged output, weave also writes `PATH.weave-sha256`, a committed digest of what it wrote. Ownership in `prepareManaged` gains one rule: an occupied declared output whose current bytes match its sidecar counts as weave's own, even with no inventory (a cold checkout) or a stale one (a `git pull` that brought a newer committed copy). Declared paths and their sidecars stay out of weave's generated `.gitignore` block. Undeclared outputs behave exactly as today.

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
| `prepareManaged` sidecar rule | `cmd/weave/internal/plan/ownership.go` | modified | filesystem reads of the destination and its sidecar |
| `managedIgnore` keep set | `cmd/weave/internal/plan/managed_ignore.go` | modified | `.gitignore` |
| compile wiring | `cmd/weave/main.go` | modified | the generation stage, stdout |

- **`prepareManaged` sidecar rule** — after the existing `own` loop: `if !own && sidecar[id.Path] != "" { own, e = vouched(fs, root, id.Path, sidecar[id.Path]) }`. `vouched` Lstats both files (regular files only, through `safeParents`) and compares `digest(current) == strings.TrimSpace(sidecar bytes)`. A missing or unreadable sidecar does not vouch (falls through to today's switch: refuse cold, displace warm). The sidecar itself is a plain (non-strict) artifacts output, so an occupied stale sidecar is overwritten as today's non-strict outputs are.
  - **Injected into:** `weavefs.FS` (tests use `weavefs.OSFS{}` over `t.TempDir()`, as every ownership test does).
- **`managedIgnore` keep set** — `managedIgnore(fs, root, old, ids, migrate, keep map[string]bool)` skips `keep` paths when emitting entries; `prepareManaged` builds `keep` from the shipped wrappers (the path and its sidecar).
- **compile wiring** — after collecting `generated` (main.go ~548): `generated, unmatched := plan.ShipStaged(generated, plan.ShippedPaths(layers))`, then `weave: warning: ship row %s names no generated output` per unmatched path. The existing `weave: regenerated tracked output %s: commit it` line (printManagedReport) already tells the operator to commit the refreshed copy and sidecar.

**Durable state (ARCH-CONSTRAINTS at-plan):** one sidecar file per declared output. weave writes it, the consumer's git tracks it, and it changes only when the output's bytes change. When the `ship` row goes, the sidecar is no longer produced, and weave's existing retirement removes it (it is inventoried like any output). Size: 65 bytes each.

## Chunk 1: declaration and staging

### Task 1: `ship` manifest verb

**Files:** Modify `cmd/weave/internal/intent/intent.go` (Kind const, String), `cmd/weave/internal/intent/manifest.go` (`kindByVerb`), `cmd/weave/internal/plan/plan.go` (lowering switch: `case intent.Ship:` no action, with a comment), `cmd/weave/internal/golden/completeness.go` (`coverIntent`, `verbName`). Test: `cmd/weave/internal/intent/intent_test.go`.

- [ ] Failing test `TestParseManifestShip`: `internal ship construct/generated/vocabulary/issue.json` parses to `Intent{Kind: Ship, Visibility: Internal, Source: "construct/generated/vocabulary/issue.json", Target: same}`.
- [ ] `go test -run TestParseManifestShip ./cmd/weave/internal/intent/` → FAIL (unknown verb skipped).
- [ ] Add `Ship` to the Kind iota, `"ship": Ship` to `kindByVerb`, the no-op lowering case and the golden cases.
- [ ] `go test ./cmd/weave/internal/intent/ ./cmd/weave/internal/plan/ ./cmd/weave/internal/golden/` → PASS.

### Task 2: `ShippedPaths`, `ShipStaged`, `SidecarPath`

**Files:** Create `cmd/weave/internal/plan/shipped.go`, `cmd/weave/internal/plan/shipped_test.go`; modify `cmd/weave/internal/plan/staged.go` (`Sidecar` field, keyed literals).

- [ ] Failing tests:
  - `ShippedPaths` selects the leaf's `internal ship` row and an ancestor's `export ship` row, drops an ancestor's `internal ship`, and refuses `../x` and an absolute path.
  - `ShipStaged` over `stagedFixture`'s actions with `issue.json` declared: the issue.json wrapper has `Sidecar == ".../issue.json.weave-sha256"`, one extra `WriteFile` holds `digest("generated issue.json")+"\n"`, the other outputs are unchanged, and a declared `construct/generated/vocabulary/missing.json` comes back unmatched.
- [ ] Run → FAIL (undefined).
- [ ] Implement (pure; no FS).
- [ ] `go test -run 'Shipped|ShipStaged' ./cmd/weave/internal/plan/` → PASS.
- [ ] Commit `#422: weave: ship manifest verb and sidecar staging`.

## Chunk 2: ownership and ignore

### Task 3: sidecar vouches for a declared output

**Files:** Modify `cmd/weave/internal/plan/ownership.go` (`prepareManaged` builds `sidecar` and `keep` maps from wrappers before `materializeManaged`, as `strict` is; `vouched` helper; `keep` stored on `managedPass` and passed to `managedIgnore`). Test: `cmd/weave/internal/plan/shipped_test.go`.

- [ ] Failing tests, each with `stagedFixture` + `ShipStaged(…, issue.json)` and `managedApply(…, ScopeArtifacts)`:
  - **Cold, the parley case:** no inventory; committed `issue.json` = "old", sidecar = digest("old"). Apply succeeds; issue.json = "generated issue.json", sidecar = its digest, `Regenerated` holds issue.json.
  - **Cold, hand edit:** issue.json = "edited", sidecar = digest("old"). Apply refuses with "occupied by something weave did not write"; issue.json is unchanged.
  - **Cold, no sidecar:** refuses as today.
  - **Cold, undeclared:** same bytes as the parley case but no `ship` row → refuses as today.
  - **Warm, pulled copy:** apply once (inventory records the first output), then overwrite issue.json and its sidecar with a consistent pair (a `git pull` of another slot's refresh) → apply regenerates with no displacement (`Displaced` empty).
  - **Warm, hand edit:** inventory present, issue.json edited, sidecar stale → displaced, as today.
- [ ] Run → the parley and pulled-copy cases FAIL.
- [ ] Implement the rule and `vouched`.
- [ ] `go test ./cmd/weave/internal/plan/` → PASS.

### Task 4: shipped paths stay out of the generated ignore block

**Files:** Modify `cmd/weave/internal/plan/managed_ignore.go` (`keep` param), its callers. Test: `shipped_test.go`.

- [ ] Failing test: after the parley-case apply, `.gitignore`'s weave block lists SKILL.md and `.source-sha` but neither `issue.json` nor `issue.json.weave-sha256`.
- [ ] Implement, run → PASS; existing `TestManagedIgnore*` and `TestGeneratedGitignoreEntriesAreExact` still pass.
- [ ] Commit `#422: weave: a committed sidecar vouches for a shipped output`.

## Chunk 3: compile wiring, e2e, atlas

### Task 5: compile applies `ship` rows

**Files:** Modify `cmd/weave/main.go` (~548: `ShipStaged` + unmatched warning). Test: `cmd/weave/startup_test.go`, modelled on `TestCompileRefusalLeavesCheckoutUntouched`.

- [ ] Failing e2e `TestCompileRegeneratesShippedOutputCold`: a base layer with a dynamic skill writing `SKILL.md`; a leaf whose manifest says `internal ship construct/generated/example/SKILL.md`, with an occupied committed `SKILL.md` = "old" and `SKILL.md.weave-sha256` = digest("old"), and no inventory. `run(...)` succeeds, `SKILL.md` holds the generated bytes, the sidecar its digest. A second case: no `ship` row → the existing refusal. A third: a `ship` row naming a path nothing generates → the warning line.
- [ ] Run → FAIL; implement; `go test -run 'TestCompile' ./cmd/weave/` → PASS.
- [ ] Commit `#422: weave: compile ships declared outputs`.

### Task 6: atlas and help

**Files:** `atlas/workflow/weave.md` (ownership section: the `ship` row, the sidecar, the vouching rule, why), the manifest format comment in `construct/base.manifest` (verb list).

- [ ] Document; commit `#422: atlas: shipped generated outputs`.

### Task 7: verification

- [ ] `go test ./cmd/weave/...` and `go vet ./cmd/weave/...`.
- [ ] Full `make test` in the background before the close.
- [ ] Live, against parley's real case, without writing parley: copy parley's tracked tree into a scratch dir (`git archive`), add the `ship` row and a sidecar of the committed issue.json, and run this branch's weave compile cold (no inventory) against current ariadne → issue.json regenerates, and the refresh line prints. Record in the issue Log.
- [ ] Hand parley#371 (filed, depends on #422) to the TL for routing after landing.
