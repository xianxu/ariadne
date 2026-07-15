# Archive All Issue Artifacts Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When an issue is archived to `workshop/history/`, also move every `workshop/plans/NNNNNN-*` artifact (durable plan + boundary-review sidecars) that shares its id prefix, in both the `sdlc merge` and `sdlc push` archive paths.

**Architecture:** A single pure-ish helper `archivePlanArtifacts(issueBase, …)` globs and moves the id-prefixed plan artifacts and returns `preparedArchiveMove`s; both `archiveDoneIssues` (push) and `archiveDoneIssuesInDir` (merge) call it right after they move the issue file, so the moved plans ride the existing `archiveAddArgs` → commit path unchanged (ARCH-DRY — one mover, two callers, one commit mechanism). The push **interrupted-archive recovery** parser (`preparedArchiveMoves`) is extended to recognize plan moves (deleted-from-plans + added-to-history, paired by basename, *without* the issue-only terminal-frontmatter gate) so a crash mid-archive still recovers cleanly rather than erroring "unrelated changes" (Root Cause — don't regress existing robustness).

**Tech Stack:** Go; table tests colocated in `cmd/sdlc/` using `t.TempDir()` (no process fakes needed — pure filesystem + git-status string parsing); SDLC atlas docs.

---

## Core concepts

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `issueIDPrefix` | `cmd/sdlc/push.go` | new |
| `isPlanPath` | `cmd/sdlc/push.go` | new |
| `preparedArchiveMoves` | `cmd/sdlc/push.go:272` | modified |

- **issueIDPrefix** — `issueIDPrefix(issueBase string) string` returns the leading 6-digit id (`"000143"`) of an issue filename, or `""` if it doesn't match `NNNNNN-*`. The single source for "which plan artifacts belong to this issue" (the glob key `id+"-*"`).
  - **DRY rationale:** The 6-digit-prefix convention is currently implicit in globs (`[0-9]{6}-*.md`); this names it once so the plan-artifact glob and any future id-keyed lookup share it.
- **isPlanPath** — `isPlanPath(path, plansDir string) bool`: `filepath.Dir(path)==filepath.Clean(plansDir) && issueFilename(base)`. The plans-dir counterpart to the existing `isIssuePath`/`isHistoryPath` (reuses `issueFilename`, the `NNNNNN-*.md` matcher).
  - **DRY rationale:** Mirrors `isIssuePath` exactly with a different dir; no new filename grammar.
- **preparedArchiveMoves** *(modified)* — the push recovery parser. Extended so a `half` tracks whether its deletion came from issues or plans; finalization accepts `(deleted-from-plans + added-to-history)` as a plan move with **no** terminal-frontmatter check, while issue moves keep the existing terminal gate. The non-terminal-history early-`other` rejection is **deferred to finalization** (a history addition's issue-vs-plan nature is only known once its paired deletion is seen).
  - **Relationships:** consumes `git status --porcelain` text; produces `[]preparedArchiveMove` + leftover `other` lines. 1 issue ↔ N plan artifacts, paired independently by basename.
  - **Future extensions:** if a third artifact dir under `workshop/` ever becomes per-issue, the same source-tagged `half` generalizes.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `archivePlanArtifacts` | `cmd/sdlc/push.go` | new | filesystem (glob + rename) |
| `archiveDoneIssues` | `cmd/sdlc/push.go:468` | modified | archive loop |
| `archiveDoneIssuesInDir` | `cmd/sdlc/merge.go:548` | modified | archive loop |
| `recoverInterruptedArchive` | `cmd/sdlc/push.go:232` | modified | git status |
| `pushFlags` / `mergeFlags` | push.go / merge.go | modified | CLI flags |

- **archivePlanArtifacts** — `archivePlanArtifacts(issueBase, plansFull, historyFull, recPlansDir, recHistoryDir string) ([]preparedArchiveMove, error)`. Globs `plansFull/<id>-*`, `os.Rename`s each into `historyFull`, and records each move with the caller's path convention (`recPlansDir`/`recHistoryDir`). `plansFull==recPlansDir` for push (cwd-relative); merge passes the mainPath-joined `*Full` for rename and the mainPath-relative `rec*` for git. Returns no error and an empty slice when the issue has no plan (glob matches nothing).
  - **Injected into:** called by both archive loops immediately after the issue `os.Rename`, its moves appended to the same `moves` slice → `archiveAddArgs` stages both halves, the existing commit lands them.
  - **Future extensions:** a `--no-artifact-sweep` escape hatch if ever needed (YAGNI now).
- **archiveDoneIssues / archiveDoneIssuesInDir** *(modified)* — each gains a `plansDir` param and, per archived issue, appends `archivePlanArtifacts(...)`'s moves.
- **recoverInterruptedArchive** *(modified)* — passes `f.PlansDir` into `preparedArchiveMoves`.
- **pushFlags / mergeFlags** *(modified)* — add `PlansDir string` (`envOr("WF_PLANS_DIR", "workshop/plans")`), threaded to the archive + recovery calls.

**Test surface.** Filesystem helpers tested against `t.TempDir()`; `preparedArchiveMoves` is a pure string→struct parser tested with crafted `git status` text (issue-only, issue+plan, plan-without-issue, no-plan). No mocks — the archive loop's only IO is glob/rename/read, exercised directly.

---

## Design decisions

- **D1 — id-prefix is the membership key.** A plan artifact belongs to issue `NNNNNN` iff its filename starts `NNNNNN-`. Covers `…-plan.md` and every `…-{close,m<x>}-review.md` sidecar without enumerating suffixes. The 6-digit id can't collide (`000143-*` ≠ `001430-*`).
- **D2 — one mover, both callers (ARCH-DRY).** `archivePlanArtifacts` is the only place that globs+moves plans; push and merge differ only in the path convention they pass in, mirroring how they already differ for the issue move itself.
- **D3 — moved plans ride the existing commit.** Appending to the `moves` slice means `archiveAddArgs` + the existing `git add -- … && git commit` stages and commits them; no new commit logic. Issues with no plan append nothing (no-op).
- **D4 — recovery stays robust (Root Cause).** Teaching `preparedArchiveMoves` about plan moves (source-tagged `half`, terminal gate only for issue halves) prevents a mid-archive `sdlc push` crash from stranding the user with an "unrelated changes" refusal. Without this, the new forward-path moves would silently break the recovery the binary already guarantees.
- **D5 — only archived issues' plans move.** Plan moves are produced per-archived-issue in the forward path, and in recovery only when paired with an id-matching deletion — never for a still-open issue.

---

## Chunk 1: Forward path (merge + push) + flags

### Task 1: `archivePlanArtifacts` helper + id prefix

**Files:** Modify `cmd/sdlc/push.go`; Test `cmd/sdlc/archiveartifacts_test.go` (new)

- [ ] **Step 1: Write the failing test** — temp `plans/` with `000143-x-plan.md` + `000143-x-close-review.md` + an unrelated `000999-y-plan.md`; assert `archivePlanArtifacts("000143-x.md", plans, history, "plans", "history")` moves exactly the two `000143-*` files to `history/`, leaves `000999-*`, returns 2 moves with the recorded rel paths, and that a no-plan issue id returns 0 moves with no error.
- [ ] **Step 2: Run → fails** (undefined `archivePlanArtifacts`).
- [ ] **Step 3: Implement** `issueIDPrefix` + `archivePlanArtifacts` (glob `plansFull/<id>-*`, sort, `MkdirAll(historyFull)`, `os.Rename`, append `preparedArchiveMove{IssuePath: filepath.Join(recPlansDir, base), HistoryPath: filepath.Join(recHistoryDir, base)}`).
- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit** — `#143: archivePlanArtifacts — move id-prefixed plan artifacts`

### Task 2: Wire into push + merge archive loops + flags

**Files:** Modify `cmd/sdlc/push.go` (`archiveDoneIssues:468`, `pushFlags`, call site `:184`, recovery call `:237`), `cmd/sdlc/merge.go` (`archiveDoneIssuesInDir:548`, `mergeFlags`, call site `:480`)

- [ ] **Step 1:** Add `PlansDir string` to `pushFlags` + `mergeFlags` with `envOr("WF_PLANS_DIR", "workshop/plans")`; add the flag registration mirroring `--history-dir`.
- [ ] **Step 2:** Give `archiveDoneIssues` + `archiveDoneIssuesInDir` a `plansDir` param; after the issue `os.Rename`/append, call `archivePlanArtifacts(base, …)` and append its moves. Push passes `plansDir, historyDir, plansDir, historyDir`; merge passes `filepath.Join(mainPath, plansDir), filepath.Join(mainPath, historyDir), plansDir, historyDir`.
- [ ] **Step 3:** Update the two call sites + the `recoverInterruptedArchive` → `preparedArchiveMoves` call to pass `f.PlansDir` (param added in Chunk 2; until then pass through, compiles once the signature lands — do Chunk 2 Task 3 first if needed, or stub).
- [ ] **Step 4: Test** — extend with an `archiveDoneIssues`/`archiveDoneIssuesInDir` temp-repo test: a done issue with a plan + sidecar → both issue and plan artifacts land in `history/` and are in the returned moves; an open issue's plan is untouched.
- [ ] **Step 5:** `go build ./cmd/sdlc/` + `go test ./cmd/sdlc/ -run Archive` → PASS. **Commit** — `#143: sweep plan artifacts in push + merge archive`

## Chunk 2: Recovery robustness

### Task 3: Teach `preparedArchiveMoves` about plan moves

**Files:** Modify `cmd/sdlc/push.go` (`preparedArchiveMoves:272`, `recoverInterruptedArchive:232`, add `isPlanPath`)

- [ ] **Step 1: Write the failing test** — `preparedArchiveMoves` with crafted status text containing an issue move (`D issues/000143-x.md` + `?? history/000143-x.md` [terminal frontmatter fixture]) **and** a plan move (`D plans/000143-x-plan.md` + `?? history/000143-x-plan.md` [no frontmatter]); assert BOTH come back as moves and `other` is empty. Add a negative: a stray non-terminal history file with no paired deletion → `other`.
- [ ] **Step 2: Run → fails** (plan move lands in `other`).
- [ ] **Step 3: Implement** — add `isPlanPath`; extend `half` with `srcIsPlan bool`; in the deletion branch tag plan deletions via `isPlanPath`; for history additions store the half and **defer** the terminal check; in finalization accept `srcDeleted && historyAdded` as a move, applying `historyFileIsTerminal` only when `!srcIsPlan`. Thread `plansDir` through `recoverInterruptedArchive` → `preparedArchiveMoves`.
- [ ] **Step 4: Run → PASS**, then `go test ./cmd/sdlc/...` (full) green — existing `preparedArchiveMoves` tests must still pass (issue-only recovery unchanged).
- [ ] **Step 5: Commit** — `#143: recover plan-artifact moves on interrupted archive`

## Chunk 3: Docs + verify

### Task 4: Atlas

**Files:** Modify `atlas/workflow/artifact-hierarchy.md` (the "Archived with issue" claim is now literally true — confirm wording), `atlas/workflow/sdlc-binary.md` (archive-step description, if it enumerates what moves)

- [ ] **Step 1:** Verify/adjust the "Archived with issue" lines so they explicitly include plan + review sidecars; note the id-prefix sweep at `merge`/`push`.
- [ ] **Step 2: Commit** — `#143: atlas — archive sweep now covers plan artifacts`

### Task 5: End-to-end verify

- [ ] Build the binary; in a throwaway temp git repo with a done issue + a `000NNN-*-plan.md` + a `000NNN-*-close-review.md`, run the archive path (or a focused harness) and confirm all three land in `history/` and are committed together. Record evidence in `## Log`.

---

## Done-when mapping

| Issue Done-when | Delivered by |
|---|---|
| merge + push move `workshop/plans/NNNNNN-*` into history, committed together | Tasks 1–2 (D2, D3) |
| issue with no plan archives with no error | Task 1 (glob no-match) |
| plans for not-archived issues untouched | Tasks 1–2 (D5) |
| atlas "archived with issue" accurate | Task 4 |
| tests: plan+sidecar moved+committed / no-plan no-op / mixed-batch isolation | Tasks 1–3 |

## Non-goals

- No change to the archive *commit message* or the gh-issue-close behavior.
- No new escape-hatch flag (YAGNI).
- Merge has no interrupted-recovery parser (only push does); not adding one.
