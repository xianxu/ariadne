# Dynamic Skills (weave) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the execution approach (superpowers-subagent-driven-development or superpowers-executing-plans). Steps use checkbox (`- [ ]`) syntax.

**Goal:** Let `weave compile` regenerate a skill package at compile time by exec'ing a tracked, language-neutral `.dynamic-skill` script — and make the datatype skill its first consumer (a `cmd/datatype` binary that writes `SKILL.md` with a live datatype-noun list, fixing the discovery miss in #111).

**Architecture:** Two pieces. **(M1) The general weave mechanism** — a new injected `Runner` exec seam + a generate stage that runs *first* in `weave compile` (before `walk.Walk`), derives its scan set from the existing `skill <dir>` intents, and execs each owned package's executable `.dynamic-skill` (cwd = package dir; non-zero exit fails the compile). **(M2) The first consumer** — `cmd/datatype` (`go:embed` prose template + filename enumeration of `construct/datatype/*.md`), the datatype package's `.dynamic-skill`, the committed regenerated `SKILL.md`, a golden drift guard, and atlas/target reconciliation.

**Tech Stack:** Go 1.26, `cmd/weave` (internal: walk, weavefs, plan, intent, golden), `text/template` + `embed`, `os/exec`.

Full design + rationale: the `## Spec` of `workshop/issues/000111-make-datatype-discovery-better.md`. This plan is the build sequence.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `isExecutable(mode)` | `cmd/weave/internal/walk/dynamic.go` | new |
| `renderSkill(names []string) string` | `cmd/datatype` | new |

(As-built: `typeNames(datatypeDir string) ([]string, error)` reads the dir, so it
is an IO seam reader — see Integration points — not a pure entity. `renderSkill`
takes the resolved names and is pure; its signature simplified from the planned
`(tmpl, []typeName) (string, error)` to use the package-level `go:embed` template
and a collision-proof placeholder string-replace.)

- **isExecutable** — the pure kernel: `mode&0o111 != 0`, table-tested without IO.
- **renderSkill / typeNames** — `typeNames` maps a `construct/datatype/` listing to sorted type names (filename without `.md`, the authoritative convention; ignores `type:`/`name:` frontmatter). `renderSkill` executes the embedded `text/template` with those names → the `SKILL.md` string. Both pure → byte-identical output across runs (the drift guard depends on determinism).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `DynamicSkillDirs(fs, layers)` | `cmd/weave/internal/walk/dynamic.go` | new | filesystem (`fs.ReadDir`/`fs.Stat`) |
| `typeNames(datatypeDir)` | `cmd/datatype/datatype.go` | new | filesystem (`os.ReadDir`) |
| `Runner` | `cmd/weave/internal/weavefs` (new file) or `cmd/weave` | new | `os/exec` |
| generate stage | `cmd/weave/main.go` (`run`) | modified | Runner + weavefs |
| `cmd/datatype` main | `cmd/datatype/main.go` | new | filesystem (read `construct/datatype/`, write `SKILL.md`) |
| CI drift guard | a make target / CI step | new | `weave compile` + `git diff --exit-code` |

- **Runner** — `type Runner interface { Run(dir string, argv []string) error }`; production `execRunner` wraps `exec.Command` (sets `cmd.Dir = dir`, streams stderr, returns error on non-zero). **Separate from `weavefs.FS`** (FS stays filesystem-only by documented stance). Injected into the generate stage so tests use a fake Runner (records calls / returns a canned error) — no real binary, no real compile.
  - **Failure:** non-zero exit → `run()` returns the error (compile fails loudly).
- **generate stage** — runs in `run()` **after `walk.Walk`** (so the parsed `skill <dir>` intents exist to reuse — `walk.Walk` reads only `construct/deps`/`base.manifest`/`prose`, never `SKILL.md`) and **before `planActions`/`GatherSkills`** (so the regenerated `SKILL.md` is what discovery reads). **Leaf-layer-only:** scans only `layers[len-1]`'s `skill` intents — never an ancestor's (weave iterates ancestor layers at real paths, so leaf-scoping is the byte-pristine guarantee, not the no-symlinks fact). Skipped in `--dry-run`, `golden`, `verify-complete` (no tree mutation). For each selected dir, `Runner.Run(dir, [".dynamic-skill"])`.
- **CI drift guard** (NOT a golden in-memory compare — impossible for an opaque marker whose `--output` weave can't redirect): a CI step runs `weave compile` then `git diff --exit-code` on the generated skill files; a stale committed `SKILL.md` fails the check.

**Test surface:** the selection + `typeNames`/`renderSkill` are pure, unit-tested without IO. The generate stage is tested with a **fake Runner** + a tempdir fixture skill dir containing an executable `.dynamic-skill` (assert it's invoked with cwd=dir; assert non-zero → compile error; assert adapted excluded; assert read-only paths skip it). `cmd/datatype` tested against a tempdir `construct/datatype/` fixture (assert sorted names in output, deterministic re-run identical).

---

## Chunk 1 — M1: the weave dynamic-skill mechanism

### Task 1.1 — `Runner` exec seam
**Files:** Create `cmd/weave/internal/weavefs/runner.go` (or `cmd/weave/runner.go`); Test: `..._test.go`
- [ ] Write failing test: a fake Runner records `(dir, argv)`; an `execRunner` integration test runs `/bin/sh -c 'exit 3'` in a tempdir and asserts a non-nil error, and `exit 0` → nil.
- [ ] Run → FAIL.
- [ ] Implement `Runner` interface + `execRunner` (`cmd.Dir=dir`, inherit stderr, map non-zero → error). 
- [ ] Run → PASS. Commit: `#111 M1: weave Runner exec seam`.

### Task 1.2 — dynamic-skill selection (pure, leaf-only)
**Files:** Create `cmd/weave/internal/walk/dynamic.go`; Test: `dynamic_test.go`
- [ ] Failing test: given the **leaf** layer's `skill construct/local` + `skill construct/adapted` + `internal skill construct/skill` intents AND an **ancestor** layer also declaring a `skill` dir with an executable marker, plus a fake listing where `construct/local/datatype/.dynamic-skill` is executable, `construct/adapted/foo/.dynamic-skill` is executable, the ancestor's marker is executable, and others absent/non-exec → selection returns only the **leaf's** `construct/local/datatype` (adapted excluded; **ancestor markers NOT returned**; non-exec ignored).
- [ ] Run → FAIL.
- [ ] Implement: take the **leaf** layer only (`layers[len(layers)-1]`), iterate its `skill`-intent dirs (skip the `construct/adapted` dir), list subdirs, keep those whose `.dynamic-skill` exists AND has an executable bit. Pure over an injected listing (the caller supplies the real listing via weavefs). Reuse the leaf's already-parsed intents — do not re-parse the manifest or hardcode dirs.
- [ ] Run → PASS. Commit: `#111 M1: leaf-only dynamic-skill selection from skill intents (adapted excluded)`.

### Task 1.3 — generate stage wired into `run()`
**Files:** Modify `cmd/weave/main.go` (`run`); Test: `cmd/weave/*_test.go`
- [ ] Failing test: drive `generateDynamicSkills(layers, fs, runner)` with a fake Runner + tempdir fixture; assert the leaf fixture's executable `.dynamic-skill` is invoked with cwd = its dir; assert a Runner error aborts the compile (non-zero); assert the `--dry-run`/golden/verify-complete paths do NOT invoke it.
- [ ] Run → FAIL.
- [ ] Implement: call `generateDynamicSkills(layers, fs, execRunner)` in `run()` **after `walk.Walk`** (so `layers`/intents exist) and **before `planActions`/`GatherSkills`** (so the regenerated `SKILL.md` is read by discovery); guard it out of the read-only paths.
- [ ] Run → PASS; `go test ./cmd/weave/... && go vet` green. Commit: `#111 M1: generate stage (after walk, before GatherSkills) execs leaf dynamic-skills`.

### Task 1.4 — M1 milestone close
- [ ] `go build ./... && go test ./cmd/weave/... && go vet ./cmd/weave/...` green (paste evidence).
- [ ] `sdlc milestone-close --issue 111 --milestone M1 ...` (auto fresh-context review; fix Critical/Important; `--no-atlas` if M1 adds no user surface — the surface is the datatype consumer in M2; rationale in `--verified`).

## Chunk 2 — M2: cmd/datatype consumer + migration + drift guard + atlas

### Task 2.1 — `cmd/datatype` (pure render + thin IO main)
**Files:** Create `cmd/datatype/main.go`, `cmd/datatype/datatype.go` (pure), `cmd/datatype/SKILL.md.tmpl` (the current 178-line prose, with the description's noun-list as a `{{...}}` action), `cmd/datatype/datatype_test.go`
- [ ] Failing test: `typeNames` over a tempdir fixture (`a.md, c.md, b.md`) → `[a b c]` sorted; `renderSkill(tmpl, names)` contains the joined list in the description and is byte-identical on re-run.
- [ ] Run → FAIL.
- [ ] Implement: `//go:embed SKILL.md.tmpl`; `typeNames` (read `--datatype-dir`, default `construct/datatype`, filenames sans `.md`, sorted); `renderSkill`; `main` parses `--output <dir>` (+ optional `--datatype-dir`), writes `<dir>/SKILL.md`. Move the current `construct/local/datatype/SKILL.md` body verbatim into `SKILL.md.tmpl`, replacing the truncated `…known frontmatter type:` tail with `…known frontmatter type: {{ .Types }}` (names joined `, `).
- [ ] Run → PASS. Commit: `#111 M2: cmd/datatype — go:embed template + filename enumeration`.

### Task 2.2 — make datatype a dynamic skill + regenerate (committed)
**Files:** Create `construct/local/datatype/.dynamic-skill` (executable); regenerate `construct/local/datatype/SKILL.md`
- [ ] Write `.dynamic-skill` (`#!/bin/sh` + `go run ../../../cmd/datatype --output .`), `chmod +x`.
- [ ] Run `make weave` (or `sdlc`-built weave) → confirm it execs `.dynamic-skill` and `construct/local/datatype/SKILL.md` regenerates with the live noun list in the description; `git diff` shows only the description tail changing; re-run → clean tree (idempotent).
- [ ] Confirm the skill still lowers (`.claude/skills/xx-datatype` resolves; `weave skills` lists datatype). 
- [ ] Commit: `#111 M2: datatype becomes a dynamic skill (committed regenerated SKILL.md)`.

### Task 2.3 — CI drift guard
**Files:** a make target (`Makefile.workflow`) and/or CI step
- [ ] Add a `weave-drift-check` (name TBD) make target: run `weave compile` then `git diff --exit-code -- '**/SKILL.md'` (or the generated skill paths) — non-zero if a committed generated file is stale vs regeneration. (Regenerate-then-diff, NOT an in-memory golden compare: the opaque marker's `--output` can't be redirected by weave.)
- [ ] Manually verify: touch a fake `construct/datatype/zzz.md`, run the target → fails (stale); `make weave` + commit → passes. Then remove the fake.
- [ ] Commit: `#111 M2: CI drift guard (weave compile + git diff --exit-code)`.

### Task 2.4 — atlas + targets
**Files:** `atlas/workflow/weave.md`, `workshop/targets/base-layer-mechanics.md`, `workshop/targets/skill-system.md`; check `atlas/index.md`
- [ ] Document the dynamic-skill convention (executable `.dynamic-skill`, generate stage, the bounded exec-seam reversal, committed-codegen + drift guard). Reconcile the "weave is filesystem-only / no exec" assertions to "filesystem + a narrow `.dynamic-skill` exec seam."
- [ ] Commit: `#111 M2: atlas + targets — dynamic skills`.

### Task 2.5 — M2 verify + issue close
- [ ] `go build ./... && go test ./... && go vet ./...` green; `make weave` clean tree; `make harness-check` green (skill still lowers to every harness).
- [ ] Manual proof of the fix: the regenerated description ends with the datatype nouns (so "make a continuation" would trigger xx-datatype). Record in Log.
- [ ] `sdlc close --issue 111 --milestone M2 ...` (end-of-issue integration review; fix Critical/Important) → `sdlc pr` → `sdlc merge`.

---

## Risks & non-goals
- **Exec-seam reversal** is deliberate and bounded ("run a package's `.dynamic-skill`", injected + tested) — not a return to open-ended exec. Atlas/targets reconciled (Task 2.4).
- **Bootstrapping:** the generated `SKILL.md` is committed, so fresh clones + read-only weave paths see it; no chicken-and-egg. The drift guard keeps it honest.
- **Non-goal:** `construct/adapted` dynamic skills; per-repo local datatype lists in the description; migrating datatype's runtime helpers into `cmd/datatype` subcommands (static-prose generation only — see Spec "Out of scope").

## Revisions

### 2026-06-17 — plan-quality findings folded in (3 blocking, all paper-fixes)
- **Ordering (Task 1.3):** generate runs **after `walk.Walk`, before `planActions`/`GatherSkills`** — not "first / before walk." `walk.Walk` doesn't read `SKILL.md`; the `skill` intents are a product of the walk, so reusing them (DRY) requires running after it.
- **Leaf-only (Task 1.2 + integration point):** scan only `layers[len-1]`'s intents; an ancestor's marker must never be exec'd (weave iterates ancestor layers at real paths — the no-symlinks fact is insufficient). Test asserts an ancestor marker is NOT invoked.
- **Drift guard (Task 2.3):** a CI `weave compile` + `git diff --exit-code`, not an in-memory `golden` compare (the opaque marker's `--output` can't be redirected by weave).

### 2026-06-17 — M1 boundary review (FIX-THEN-SHIP, no Critical)
- **Pure/IO table corrected:** `DynamicSkillDirs` is an FS-injected seam reader (does `fs.ReadDir`/`fs.Stat`, matching `GatherSkills`), not pure — moved to Integration points. The only pure entity in `dynamic.go` is `isExecutable`. (Code structure was already right — the table label was wrong.)
- **Reinforced:** Task 2.4 (atlas + `base-layer-mechanics` + `skill-system` reconciliation) MUST land in M2 before merge — a live exec seam must not propagate downstream while the docs still claim "weave is filesystem-only." Treated as an M2 hard gate.

### 2026-06-17 — M2 integration review (FIX-THEN-SHIP, no Critical) — fixes folded
- **Drift guard wired into CI** (was an Important gap): the `make weave-drift-check` target was human-only — `scripts/merge-checks.d/` had no entry, so neither CI nor the pre-push hook ran it. Added `scripts/merge-checks.d/30-weave-drift.sh` (reuses the make target; no-op pass where the target is absent). Verified: clean→pass, injected fake datatype→fail.
- **Production exec path now e2e-tested** (was an Important gap): `TestCompileRunsDynamicSkills` drives `run(..., dryRun=false)` over a fixture with a real executable `.dynamic-skill` → asserts the sentinel appears (proves the real `ExecRunner` runs the *relative* `./.dynamic-skill` with cwd = package dir; the fake-runner + absolute-`/bin/sh` units didn't cover it).
- **Pure/IO table corrected again:** `typeNames` is an IO seam reader (`os.ReadDir`), moved to Integration points; only `isExecutable`/`renderSkill` are pure. `renderSkill` signature reconciled to as-built.
