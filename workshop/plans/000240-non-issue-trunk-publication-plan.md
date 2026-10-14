# Non-issue workshop trunk publication implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish explicitly selected, dirty non-issue workshop prose to main while another issue branch is active, without carrying that prose in the issue branch.

**Architecture:** Add `sdlc trunk publish --path P[,P…]` around the existing `gitx.TrunkFile.UpdateMany` CAS transaction. A pure path/three-body decision selects only approved prose and refuses a main path that moved from the branch merge base; a thin command reads local Git/worktree state, publishes exact bytes, then restores only unchanged selected local files to the branch's HEAD. Retry observes main and finishes a prior successful push without another write.

**Tech Stack:** Go, Cobra, `cmd/sdlc/internal/gitx`, real bare-Git test fixtures.

---

## Core concepts

| PURE entity | Lives in | Status |
|---|---|---|
| `TrunkProsePath` eligibility | `cmd/sdlc/trunkpublish.go` | new |
| `prosePublishDecision` | `cmd/sdlc/trunkpublish.go` | new |

`TrunkProsePath` accepts only ordinary Markdown files directly or recursively below `workshop/parley/`, `workshop/pensive/`, `workshop/targets/`, `workshop/projects/`, plus the exact `workshop/lessons.md`. It rejects issue, plan, atlas, code, base-layer, asset, symlink, traversal, and absolute paths. The explicit allowlist is intentional: code, settings and atlas require issue review. A path must be modified only in the working tree, not already committed on this issue branch or staged in the index; this lets successful publication remove its dirty copy without rewriting history (ARCH-PURPOSE).

`prosePublishDecision(local, branchBase, main, presence)` yields already-present, write, or refuse. It permits a new local file only if absent at the branch merge base and main. It permits an existing local edit only if main still has the branch-base bytes; identical local/main bytes are already-present. A main path that moved refuses even when only another line changed; the user must merge/reconcile deliberately. Tests are pure and cover all presence/value combinations (ARCH-PURE).

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| `runTrunkPublish` | `cmd/sdlc/trunkpublish.go` | new | working tree, Git queries, `gitx.TrunkFile` |
| `gitx.TrunkFile.UpdateMany` | `cmd/sdlc/internal/gitx/updatemany.go` | reused | Git commit-tree, fetch, CAS push, bounded retry |

The command pins the selected paths and local bytes before publishing. It verifies regular files and rejects staged or branch-committed changes to those paths. On each CAS attempt it compares the pinned branch-base tree and current trunk view, rechecks the allowlist and worktree file identity, and constructs one `TrunkWrite` with `ExactBytes`. It never publishes partial selection. It uses the same stateful Git seam and real bare-Git fixtures as other SDLC publication tests (ARCH-DRY, ARCH-MOCK). An uncertain push returns the existing uncertainty error and leaves local bytes; rerunning first checks whether main carries the pinned bytes and then cleans. A rejected CAS retries only if unrelated main paths moved; a selected-path conflict refuses. The command serializes within the existing SDLC transaction lock, while the remote CAS governs other workspaces (ARCH-ORDER).

After confirmed publication, each selected path is cleaned only if its current local bytes and index state still match the pinned read. A tracked path is restored from HEAD; an untracked path is removed. If any cleanup fails, the command reports main's confirmed state and the affected local path, so a retry can finish. Other dirty paths and the issue branch's commits/index are untouched. No durable record family is added: the main commit is the artifact, and normal Git history retains it (ARCH-FUNERAL). Work is a human-invoked, bounded batch of explicit paths, not a watcher; the transaction uses the existing three-attempt remote CAS budget (ARCH-CONSTRAINTS).

## Task 1: pin the path and decision contract

**Files:** `cmd/sdlc/trunkpublish.go`, `cmd/sdlc/trunkpublish_test.go`

- [ ] Write table tests for the exact allowlist, traversal/symlink paths, and the decision's present/absent/changed-main cases. Make each refusal assert its own reason.
- [ ] Run targeted tests to prove they fail before implementation.
- [ ] Implement the pure allowlist and decision; run the targeted tests green.
- [ ] Commit the tested core as `#240 M1: define trunk prose publication contract`.

## Task 2: publish and finish one explicit set

**Files:** `cmd/sdlc/trunkpublish.go`, `cmd/sdlc/trunkpublish_test.go`, `cmd/sdlc/main.go`, `cmd/sdlc/internal/recovery/catalog.go`

- [ ] Write real bare-Git tests from an issue branch with dirty parley/project/lessons files and unrelated dirty code. Assert main receives exactly selected bytes in one commit, while HEAD, issue branch, index and unrelated work remain unchanged and selected local files become clean.
- [ ] Add refusal tests for a staged selection, a branch-committed selected path, a selected path moved on main, path escapes, and a non-regular file. Assert main and local state remain unchanged on refusal.
- [ ] Add a test that makes a CAS attempt race an unrelated main change, then succeeds on retry; add an uncertain push test whose first call leaves local bytes and whose repeat recognizes main and cleans them.
- [ ] Implement the Cobra verb with the existing transaction lock and `TrunkFile.UpdateMany`. Keep Git/process calls in the thin command; inject only the publication seam needed to reproduce outcomes. Register recovery semantics with concrete proof names.
- [ ] Run targeted package tests, `go test ./cmd/sdlc/... -run 'TestTrunkPublish|TestProsePublish' -count=1 -timeout 60m`; commit as `#240 M1: publish selected workshop prose to main`.

## Task 3: document, verify and review

**Files:** `atlas/workflow/sdlc-binary.md`, `atlas/workflow/workspace-branching.md`, `workshop/issues/000240-*.md`, `workshop/lessons.md` only if a review finds a transferable rule.

- [ ] Document `sdlc issue publish --issue` for issue details, `sdlc trunk publish --path` for non-issue workshop prose, and the branch-reviewed home for atlas/code/settings. Keep `atlas/index.md` current if adding a file (none planned).
- [ ] Run `gofmt`, targeted tests, `make -k`, and the local suite via `sdlc local-check run --name suite` after the tree is stable; inspect logs and compare the patch with main.
- [ ] Tick M1, record verification and review result in the issue Log, then run `sdlc milestone-close --issue 240 --milestone M1`.
- [ ] Resolve review findings as a class, rerun affected checks, tick the final close row only when fulfilled, and run `sdlc close --issue 240 --verified <evidence>`; publish via `sdlc pr` and `sdlc merge` after CI.
