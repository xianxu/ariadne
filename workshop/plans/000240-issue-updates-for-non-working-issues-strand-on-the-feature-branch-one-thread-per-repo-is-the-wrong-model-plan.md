# Non-issue workshop trunk publication implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish one explicitly selected, dirty non-issue workshop prose file to main while another issue branch is active, without carrying that prose in the issue branch.

**Architecture:** Add `sdlc trunk publish --path P` around the existing `gitx.TrunkFile.UpdateMany` CAS transaction. A pure path/three-body decision selects only approved prose and refuses a main path that moved from the branch merge base; a thin command reads local Git/worktree state, publishes exact bytes, then restores an unchanged local file to the branch's HEAD. One path per invocation makes cleanup a single transition. Retry observes main and finishes a prior successful push without another write.

**Tech Stack:** Go, Cobra, `cmd/sdlc/internal/gitx`, real bare-Git test fixtures.

---

## Core concepts

| PURE entity | Lives in | Status |
|---|---|---|
| `TrunkProsePath` eligibility | `cmd/sdlc/trunkpublish.go` | new |
| `prosePublishDecision` | `cmd/sdlc/trunkpublish.go` | new |

`TrunkProsePath` is lexical and accepts only `.md` paths below `workshop/parley/`, `workshop/pensive/`, `workshop/targets/`, `workshop/projects/`, plus exact `workshop/lessons.md`. It rejects issue, plan, atlas, code, base-layer, asset, traversal, and absolute paths. Filesystem inspection separately verifies regular files, including no symlink in the path. The explicit allowlist is intentional: code, settings and atlas require issue review. A path must be modified only in the working tree, not already committed on this issue branch or staged in the index; this lets successful publication remove its dirty copy without rewriting history (ARCH-PURPOSE).

`prosePublishDecision(local, branchBase, main, presence)` yields already-present, write, or refuse. It extends the existing body comparison in `cmd/sdlc/republish.go:37` only for presence-aware prose semantics; share the equal-body rule rather than duplicate it (ARCH-DRY). It permits a new local file only if absent at the branch merge base and main. It permits an existing local edit only if main still has the branch-base bytes; identical local/main bytes are already-present. A main path that moved refuses even when only another line changed; the user must merge/reconcile deliberately. Tests are pure (ARCH-PURE).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| `runTrunkPublish` | `cmd/sdlc/trunkpublish.go` | new | working tree, Git queries, `gitx.TrunkFile` |
| `finishTrunkPublish` | `cmd/sdlc/trunkpublish.go` | new | local path and index cleanup |
| `gitx.TrunkFile.UpdateMany` | `cmd/sdlc/internal/gitx/updatemany.go` | reused | Git commit-tree, fetch, CAS push, bounded retry |

The command pins the selected path and local bytes before publishing. It verifies a regular file and rejects staged or branch-committed changes to it. On each CAS attempt it compares the pinned branch-base tree and current trunk view, rechecks the worktree identity, and constructs one `TrunkWrite` with `ExactBytes`. It uses the same stateful Git seam and real bare-Git fixtures as other SDLC publication tests (`cmd/sdlc/internal/gitx/updatemany.go:75`, ARCH-MOCK). An uncertain push returns the existing uncertainty error and leaves local bytes; rerunning first checks whether main carries the local bytes and then cleans. A rejected CAS retries only if unrelated main paths moved; a selected-path conflict refuses. The command serializes within the existing SDLC transaction lock, while the remote CAS governs other workspaces (ARCH-ORDER).

After confirmed publication, `finishTrunkPublish` cleans the one path only if its current local bytes and index state still match the pinned read. A tracked path is restored from HEAD; an untracked path is removed. A cleanup failure before the effect leaves the original bytes for retry. If cleanup succeeded but its reply was lost, a retry sees a clean path and returns no-op; it does not infer ownership of main's unrelated changes. Other dirty paths and the issue branch's commits/index are untouched. No durable record family is added: the main commit is the artifact, and normal Git history retains it (ARCH-FUNERAL). Work is a human-invoked, one-path transaction, not a watcher. The command accepts files up to 1 MiB (prose envelope; larger files refuse before reading fully), uses `gitx.BoundRemote` for a 2-minute deadline and command cancellation, and the existing three-attempt remote CAS budget. An exceeded limit fails before push and leaves the local file untouched (ARCH-CONSTRAINTS).

## Task 1: pin the path and decision contract

**Files:** `cmd/sdlc/trunkpublish.go`, `cmd/sdlc/trunkpublish_test.go`

- [x] Write `TrunkProsePath` tests with arbitrary path components (including traversal and separators) and `prosePublishDecision` tests over adversarial present/absent/value combinations; assert the specific refusal reason.
- [x] Run targeted tests to prove they fail before implementation.
- [x] Implement the pure allowlist and decision; run the targeted tests green.
- [x] Commit the tested core with the command as `#240: publish selected workshop prose to main`.

## Task 2: publish and finish one explicit set

**Files:** `cmd/sdlc/trunkpublish.go`, `cmd/sdlc/trunkpublish_test.go`, `cmd/sdlc/main.go`, `cmd/sdlc/internal/recovery/catalog.go`

- [x] Write `runTrunkPublish` real bare-Git tests that mutate main and the working tree independently; assert no branch/index effects and rejection before write when the selected path moved. The reused `gitx.UpdateMany` seam's existing tests cover CAS races and unknown push responses.
- [x] Write `finishTrunkPublish` fault-injection tests for interrupted cleanup and concurrent local edits; mechanically compare pinned bytes and index state before any restore/removal, then rerun the production command to prove convergence or a safe refusal.
- [x] Implement the Cobra verb with the existing transaction lock and `TrunkFile.UpdateMany`. Keep Git/process calls in the thin command; inject only the publication seam needed to reproduce outcomes. Register recovery semantics with concrete proof names.
- [x] Run targeted package tests, `go test ./cmd/sdlc -run 'TestRecoveryContractsAreProven|TestTrunkProsePath|TestProsePublishDecision|TestTrunkPublish|TestReadLocalProse' -count=1 -timeout 10m`; commit as `#240: publish selected workshop prose to main`.

## Task 3: document, verify and review

**Files:** `atlas/workflow/sdlc-binary.md`, `atlas/workflow/workspace-branching.md`, `workshop/issues/000240-*.md`, `workshop/lessons.md` only if a review finds a transferable rule.

- [x] Document `sdlc issue publish --issue` for issue details, `sdlc trunk publish --path` for non-issue workshop prose, and the branch-reviewed home for atlas/code/settings. Keep `atlas/index.md` current if adding a file (none planned).
- [x] Run `gofmt`, targeted tests, `make -k`, and the local suite via `sdlc local-check run --name suite` after the tree is stable; inspect logs and compare the patch with main.
- [ ] Resolve review findings as a class, rerun affected checks, tick the final close row only when fulfilled, and run `sdlc close --issue 240 --verified <evidence>`; publish via `sdlc pr` and `sdlc merge` after CI.

## Revisions

### 2026-10-14 — plan-quality round 1

PQ-1: one path per invocation removes partial batch cleanup; the retry state is
explicit above. PQ-2: named functions now carry adversarial-input strategy, and
lexical eligibility is separated from filesystem inspection. PQ-3: the issue
has one close boundary. The review's minor envelope note is addressed too.

### 2026-10-14 — implementation validation scope

The command's real-Git tests cover its path-level decision and local cleanup;
the reused `gitx.UpdateMany` suite remains the proof for CAS retry and unknown
remote responses. This avoids duplicating the lower seam's stateful fake tests.
