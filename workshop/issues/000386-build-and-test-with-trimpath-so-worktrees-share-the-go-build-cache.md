---
id: 000386
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '5977f781c4c2c2030c27a33a555605a8a41cba1f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T10:52:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c3576436", done: "7f5d6529"}
---

# Build and test with -trimpath so worktrees share the Go build cache

## Problem

On 2026-10-15 the shared Go build cache (`GOCACHE=/Volumes/AgentScratch/go-build`) filled the 186 GB scratch volume
(184 GB), failing every slot's suite. Main multiplier: builds embed each worktree's absolute source path, so the same
package is cached separately for each of ~20 slot worktrees; after a `go clean -cache`, ~30 GB came back within
minutes. The operator approved building with `-trimpath`.

## Spec

- Run this repo's full suite with `GOFLAGS=-trimpath`; fix every test that depends on absolute source paths (for
  example locating fixtures via `runtime.Caller(0)` or embedding `__FILE__`-like paths), using module-relative or
  `go list`-derived roots instead.
- Then make `-trimpath` the default for this repo's builds and its recorded suite (Makefile and `.sdlc` suite
  command, as allowed by the local-gates guard), not via a global `go env -w`.
- Note the cache effect: build the same package in two worktrees and confirm one cache entry (or measure the cache
  growth for two worktrees before and after).

## Done when

- The recorded full suite passes with `-trimpath` on by default.
- A measurement shows cache entries shared across two worktrees for an unchanged package.

## Plan

- [x] Run the unchanged full suite with GOFLAGS=-trimpath through the normal machine gate. Capture exact head, output, duration and any failures. Repair demonstrated source-path assumptions with the existing cwd/Git/module-root helpers; add a regression for each discovered problem. No speculative path refactor.
- [x] Make the owner Makefile export effective Go flags plus -trimpath by default, retaining environment/go-env flags (currently -p=3). Add .sdlc/local-checks.json declaring required suite argv [make, test], weight 2 and timeout 60 minutes: the existing execution settings, with trimming inherited from the tested Makefile. Add an execution regression in scripts/test/trimpath_test.go exercising real owner build/test recipes with a temporary tracing Go/Python fixture, checking trimpath plus preserved flags. Run RED/GREEN and remove the default in a temporary mutation to prove this test fails.
- [x] Add -trimpath to construct/dev-aliases.sh's existing emitter only, including its example. Extend the existing emitted-function execution fixture in construct/scripts/test/dev-aliases.test.sh; run RED/GREEN and remove the emitted flag once to prove the regression.
- [ ] Measure one unchanged package (pkg/vocab) in two clean temporary Git worktrees of the same pinned commit, using a private shared cache per mode. Compare untrimmed A/B versus trimmed A/B: archive cache path/build ID, second-worktree compile actions, elapsed time and cache growth. Keep Go version/flags/head in the Log; do not clear the machine's cache. Remove owned worktrees and private caches after capturing results.
- [ ] Document the defaults and direct-Go opt-in in README and the existing build/test atlas section. Publish a PR, run the recorded full suite on its pushed head with trimming enabled by default, and record evidence.

After verification, close through mandatory review, obtain operator verification for the protected .sdlc declaration, integrate current main and land on SHIP. Verification checkboxes do not claim their own later close/landing success.

### Design and verification details

Recommended approach: set the repository's effective GOFLAGS once in its owner Makefile, then let both owner tool builds and the existing Python sharder inherit it (ARCH-DRY). Preserve configured flags rather than replacing the operator's -p cap. Explicit command-line Make overrides remain explicit overrides. A second copy of -trimpath in every Go subprocess would duplicate policy; a global go-env write would exceed this repository's scope.

The recorded suite currently uses fallback make test because no .sdlc/local-checks.json exists. The new declaration keeps the same argv, required enforcement, weight and timeout; it makes the repo's suite explicit without changing sdlc's configuration authority. The pre-landing runner reads the declaration from main, so verification of this PR still uses main's equivalent fallback and the candidate's Makefile. The operator must apply local-gates-verified after inspecting the concrete PR; the agent never applies it.

Core concepts: no new domain entity or lifecycle. Integration surfaces are the owner Makefile (modified, Go subprocess environment), .sdlc/local-checks.json (new, existing recorded-check schema) scripts/test/trimpath_test.go (new regression through real Make recipes), and construct/dev-aliases.sh (modified emitted Go argv, existing execution fixture). Existing fake executable fixtures persist argv/environment observations in a private test folder; the two-worktree measurement is live Go-cache conformance. No generic build wrapper or cache manager is introduced.

The path inventory found no runtime.Caller/Callers/FuncForPC/go-build source lookup. Existing repoRootForTest/realRepoRoot use Git/cwd, and package helpers walk to go.mod. Full-suite execution is the acceptance proof, including nested builds and stage VCS metadata checks (ARCH-PURPOSE).

ARCH-CONSTRAINTS: normal weighted suite gate and existing Go parallelism cap remain authoritative; measurement is sequential on a single representative package with isolated caches, not an extra suite. Record actual bytes/timing rather than promise a percentage. ARCH-SECURE: fixtures/caches are private temporary paths; no credentials or global Go settings are changed. ARCH-ORDER: no new persistent state machine; worktrees/caches are measurement-scoped resources removed after commands finish. ARCH-FUNERAL: ordinary Go cache pruning remains Go-owned; experiment caches/worktrees are removed and only measured evidence stays in this issue. Existing source-path fixes, if any are demonstrated, stay at their existing integration seam (ARCH-PURE/ARCH-MOCK).

## Log

### 2026-10-15

Verified operator-approved assignment receipt 60709edf-831b-4bf9-b980-453d4e272323, claimed from ariadne:5 and entered start-plan. Source baseline is current main at issue-branch creation. Effective Go toolchain is go1.27.1, GOFLAGS=-p=3, GOCACHE=/Volumes/AgentScratch/go-build. Scratch currently has 184 GiB free; no shared-cache cleanup is part of this work.

Read-only source inventory (including a delegated independent sweep) found no compiler-embedded Go source-path dependency. Existing Git/cwd/module-root helpers cover source guards. Owner Makefile builds tools and weave, while Makefile.local invokes scripts/test-shard.py; neither currently enables trimpath. Weave's release builder already explicitly uses it. The generic dev-alias emitter invokes direct go build and is a separate cross-repo policy surface; the proposed change is the issue's owner-Makefile and recorded-suite scope. Include this distinction in the plan report for TL review.

Plan is inside the quick-flow shell unless demonstrated fixes expand it. No implementation edits or suite started before the committed plan is sent to the hub.

### 2026-10-15 — approved scope and baseline verification

Verified TL approval receipt 7cc52a6b-ec83-4141-a1b8-14ef6a6ca6f6: plan 09917b82 approved, with the generic dev-alias emitter also in scope because it builds every peer owner in each worktree. Keep that production change to the flag itself. TL will request the operator's local-gates label once the concrete PR is available.

Unchanged-source baseline at 09917b82befd712f7d2fa96cc8f40b047d153d88 passed with GOFLAGS='-p=3 -trimpath' (existing cap retained): make test exit 0, 639.2s execution after 694s gate wait, 12 shards / 1485 cmd/sdlc tests and 62 other packages, zero failures. Log: /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne386-baseline-w33e6s6c.log. No demonstrated absolute-source-path dependency, so no speculative path-helper change. Checkout stayed unchanged through the suite.

### 2026-10-15 — default flags and targeted verification

change-code admitted the inferred quick flow. New recipe-execution regressions first failed because owner builds/sharder lacked trimpath and the recorded declaration was absent (0.681s). Owner Makefile now reads effective Go flags, appends trimpath only when absent, and exports them; explicit command-line Make overrides retain their normal semantics. The existing suite command/budgets are explicit in .sdlc/local-checks.json. Focused GREEN passed (0.585s). Removing only the default append produced intended RED with missing trimpath across tool/sharder/recorded paths (0.725s); restored before GREEN.

Delegated emitter change touched only its Go argv/example plus its existing tests. Existing shape and executed-argv assertions gave RED 44/4 before the flag, GREEN 48/0 after it, RED 44/4 when only the emitted flag was removed, then restored GREEN 48/0. Inspected the diff independently. Fake Go executables now find the -o operand rather than relying on its old argument position.

Combined script tests exposed portable-makefile.test.sh's fake treating the new go env GOFLAGS query as an extra build. Updated that fake's query response and asserted each build's effective trimpath; preserved the exact owned-build population assertion. No production workaround. Combined scripts/test package then passed (57.989s), including all construct shell fixtures under three output environments. make tools weave-build passed for all seven owner binaries; go version -m bin/sdlc confirms -trimpath=true while retaining VCS metadata. README/atlas describe defaults, direct-Go use and reloading emitted shell functions.

## Revisions

- 2026-10-15 — Clarify repository-scoped default flag inheritance and equivalent explicit recorded-suite settings, preserve existing Go flags, and include a measured untrimmed control. Original purpose and Done-when unchanged.

- 2026-10-15T11:19-07:00 — TL approval explicitly includes construct/dev-aliases.sh's flag-only emitter change across peer owners. Supersedes the proposed emitter exclusion in the initial Log; current Plan and integration inventory include the existing emitter/test seam. Done-when unchanged.

- 2026-10-15 — Separate completed implementation/verification checkboxes from later close, operator-label and landing actions, avoiding a self-dependent close precondition. No acceptance or scope change.
