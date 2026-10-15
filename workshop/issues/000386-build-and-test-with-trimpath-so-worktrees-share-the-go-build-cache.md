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

- [ ] Run the unchanged full suite with GOFLAGS=-trimpath through the normal machine gate. Capture exact head, output, duration and any failures. Repair demonstrated source-path assumptions with the existing cwd/Git/module-root helpers; add a regression for each discovered problem. No speculative path refactor.
- [ ] Make the owner Makefile export effective Go flags plus -trimpath by default, retaining environment/go-env flags (currently -p=3). Add .sdlc/local-checks.json declaring required suite argv [make, test], weight 2 and timeout 60 minutes: the existing execution settings, with trimming inherited from the tested Makefile. Add an execution regression in scripts/test/trimpath_test.go exercising real owner build/test recipes with a temporary tracing Go/Python fixture, checking trimpath plus preserved flags. Run RED/GREEN and remove the default in a temporary mutation to prove this test fails.
- [ ] Measure one unchanged package (pkg/vocab) in two clean temporary Git worktrees of the same pinned commit, using a private shared cache per mode. Compare untrimmed A/B versus trimmed A/B: archive cache path/build ID, second-worktree compile actions, elapsed time and cache growth. Keep Go version/flags/head in the Log; do not clear the machine's cache. Remove owned worktrees and private caches after capturing results.
- [ ] Document the defaults and direct-Go opt-in in README and the existing build/test atlas section. Publish a PR, run the recorded full suite on its pushed head with trimming enabled by default, and record evidence. Obtain operator verification for the protected .sdlc declaration, then close through mandatory review and land on SHIP.

### Design and verification details

Recommended approach: set the repository's effective GOFLAGS once in its owner Makefile, then let both owner tool builds and the existing Python sharder inherit it (ARCH-DRY). Preserve configured flags rather than replacing the operator's -p cap. Explicit command-line Make overrides remain explicit overrides. A second copy of -trimpath in every Go subprocess would duplicate policy; a global go-env write would exceed this repository's scope.

The recorded suite currently uses fallback make test because no .sdlc/local-checks.json exists. The new declaration keeps the same argv, required enforcement, weight and timeout; it makes the repo's suite explicit without changing sdlc's configuration authority. The pre-landing runner reads the declaration from main, so verification of this PR still uses main's equivalent fallback and the candidate's Makefile. The operator must apply local-gates-verified after inspecting the concrete PR; the agent never applies it.

Core concepts: no new domain entity or lifecycle. Integration surfaces are the owner Makefile (modified, Go subprocess environment), .sdlc/local-checks.json (new, existing recorded-check schema) and scripts/test/trimpath_test.go (new regression through real Make recipes). Existing fake executable fixtures persist argv/environment observations in a private test folder; the two-worktree measurement is live Go-cache conformance. No generic build wrapper or cache manager is introduced.

The path inventory found no runtime.Caller/Callers/FuncForPC/go-build source lookup. Existing repoRootForTest/realRepoRoot use Git/cwd, and package helpers walk to go.mod. Full-suite execution is the acceptance proof, including nested builds and stage VCS metadata checks (ARCH-PURPOSE).

ARCH-CONSTRAINTS: normal weighted suite gate and existing Go parallelism cap remain authoritative; measurement is sequential on a single representative package with isolated caches, not an extra suite. Record actual bytes/timing rather than promise a percentage. ARCH-SECURE: fixtures/caches are private temporary paths; no credentials or global Go settings are changed. ARCH-ORDER: no new persistent state machine; worktrees/caches are measurement-scoped resources removed after commands finish. ARCH-FUNERAL: ordinary Go cache pruning remains Go-owned; experiment caches/worktrees are removed and only measured evidence stays in this issue. Existing source-path fixes, if any are demonstrated, stay at their existing integration seam (ARCH-PURE/ARCH-MOCK).

## Log

### 2026-10-15

Verified operator-approved assignment receipt 60709edf-831b-4bf9-b980-453d4e272323, claimed from ariadne:5 and entered start-plan. Source baseline is current main at issue-branch creation. Effective Go toolchain is go1.27.1, GOFLAGS=-p=3, GOCACHE=/Volumes/AgentScratch/go-build. Scratch currently has 184 GiB free; no shared-cache cleanup is part of this work.

Read-only source inventory (including a delegated independent sweep) found no compiler-embedded Go source-path dependency. Existing Git/cwd/module-root helpers cover source guards. Owner Makefile builds tools and weave, while Makefile.local invokes scripts/test-shard.py; neither currently enables trimpath. Weave's release builder already explicitly uses it. The generic dev-alias emitter invokes direct go build and is a separate cross-repo policy surface; the proposed change is the issue's owner-Makefile and recorded-suite scope. Include this distinction in the plan report for TL review.

Plan is inside the quick-flow shell unless demonstrated fixes expand it. No implementation edits or suite started before the committed plan is sent to the hub.

## Revisions

- 2026-10-15 — Clarify repository-scoped default flag inheritance and equivalent explicit recorded-suite settings, preserve existing Go flags, and include a measured untrimmed control. Original purpose and Done-when unchanged.
