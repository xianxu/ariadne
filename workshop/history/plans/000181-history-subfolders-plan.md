# workshop/history subfolders — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The archive gains per-kind subfolders — new archives land in `workshop/history/issues/` and `workshop/history/plans/` (review sidecars ride with plans), ariadne's 258 existing flat files migrate in one `git mv` commit, resolution tolerates BOTH layouts (downstream repos keep working pre-migration), and no consumer hardcodes the layout (#181).

**Architecture:** Reads tolerant, writes strict. The subfolder convention is single-sourced as `vocab.ArchiveSubdirs(root) (issues, plans)` in `pkg/vocab` — Go-owned rather than cue-encoded, because (a) writers take a `--history-dir`/`WF_HISTORY_DIR` root override and must derive subdirs from an arbitrary root, and (b) changing `issue.cue`'s `discovery: archive` from a string to a struct would break downstream JSON consumers (parley sources discovery from the export). The cue keeps `archive: "workshop/history"` as the ROOT with a comment naming the Go-owned subfolder convention (#181). Readers (`familyFiles`, `NextID`, `isHistoryPath`) accept root + both subdirs; writers (`archiveDoneIssues`, `archivePlanArtifacts`) emit only into subdirs. ARCH-DRY: one derivation function, every consumer routes through it; ARCH-PURE: the derivation and path predicates stay pure, IO untouched at the seams.

**Sidecar decision (settled):** review sidecars archive to `history/plans/` — they are plans-dir residents today (`NNNNNN-*-review.md` live in `workshop/plans`), the family mover already sweeps them with plans in one glob, and a third subfolder buys nothing.

**Consumer map (verified against code):**

| Consumer | Today | Change |
|----------|-------|--------|
| `issue.cue` `discovery.archive` | `"workshop/history"` (root) | UNCHANGED (comment names the convention) |
| `pkg/vocab` | `Discovery.Archive` string | + `ArchiveSubdirs(root)` func |
| `familyFiles` (resolve.go:225) | globs Home, Plans, Archive | + globs both subdirs (root kept — legacy tolerance) |
| `archiveDoneIssues` (push.go:563) | dest = `historyDir/<base>` | dest = `<issues subdir>/<base>` |
| `archiveDoneIssuesInDir` (merge.go:608) | OWN issue-dest logic (merge.go:629 rename + :639 recorded rel path — a DUPLICATE of push's, latent ARCH-DRY debt) | both legs → issues subdir; consider extracting the shared dest computation while there |
| `archivePlanArtifacts` (push.go:281) | dest = `historyFull/<base>` | dest = `<plans subdir>/<base>` (both callers: push + merge.go:608) |
| `isHistoryPath` (push.go:477) | EXACT dir == historyDir | dir ∈ {root, issues, plans} (feeds `isTrackerPath`/`assessDirty` + `recoverInterruptedArchive`) |
| `issue.NextID` (scaffold.go:29) | reads issuesDir + historyDir | + reads `history/issues` subdir (root kept — legacy) |
| `assessDirty`/`isTrackerPath` (merge.go:157) | via isHistoryPath | inherits the fix — no direct change |
| `collectDiff` (milestoneclose.go:587), `state.detectDrift`, startplan assessDirty | prefix/exact usage via the same predicates or root-only paths | verify at implementation; expected no-change (prefix semantics) or inherits isHistoryPath |
| ariadne's 258 flat files | mixed in root | one-commit `git mv` migration (159 → issues/, 99 → plans/) |

**Downstream:** peers get the behavior when their sdlc rebuilds (the shell function rebuilds per call); reads tolerate their flat legacy archives indefinitely, so migration there is optional tidiness — noted in atlas, no propagate required.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `ArchiveSubdirs` | `pkg/vocab/vocab.go` | new |
| `isHistoryPath` | `cmd/sdlc/push.go` | modified |
| `NextID` | `cmd/sdlc/internal/issue/scaffold.go` | modified |

- **`ArchiveSubdirs(root string) (issues, plans string)`** — THE convention, one place: `root/issues`, `root/plans`. Every consumer (writers with flag-overridden roots, readers via `Discovery.Archive`) derives through it; nothing concatenates the literals elsewhere (guard test: grep-style source check à la #163, plus unit test).
  - **DRY rationale:** the root+convention split exists precisely so a `--history-dir` override and the vocab default produce consistent layouts from one function.
  - **Future extensions:** `history/projects/` (#180) is a third return or a kind-keyed variant — one function edit.
- **`isHistoryPath`** — accepts a path whose dir is the root OR either subdir (root kept: pre-migration trees + downstream). Still requires `issueFilename`.
- **`NextID`** — scans `issuesDir`, `historyDir` (legacy), and the issues subdir. (Plans subdir unnecessary — plan files carry the same ids as their issues.)

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `familyFiles` | `cmd/sdlc/resolve.go` | modified | filesystem glob |
| `archiveDoneIssues` / `archivePlanArtifacts` | `cmd/sdlc/push.go` (+ merge.go caller) | modified | os.Rename + git staging |
| migration commit | `workshop/history/**` | one-time | git mv |
| `issue.cue` comment | `construct/vocabulary/issue.cue` | modified (comment only) | — |

**Test surface:** `ArchiveSubdirs` unit; `TestIsHistoryPath` table (root, issues/, plans/, non-history, non-issue-filename); `NextID` across layouts (max id in a subdir wins); `familyFiles`/resolve fixture with a family split across BOTH layouts (flat legacy + subfoldered) resolving as one family; archive e2e — existing push/merge archive tests updated to expect subfolder dests (they exist: the merge_e2e/push tests assert `HistoryPath`), plus `recoverInterruptedArchive` with subfolder dests; migration verified by `sdlc resolve` of known archived ids + clean `git status` + `git diff --check`.

---

## Tasks (single-pass, plain checkboxes)

### Task 1: `vocab.ArchiveSubdirs` + predicate updates (TDD)

- [ ] Failing tests: `TestArchiveSubdirs` (pkg/vocab); `TestIsHistoryPath` (NEW — isHistoryPath is only indirectly tested today) table: root, issues/, plans/, non-history, non-issue-filename; `TestNextID_SubfolderLayout` (max id found in `history/issues/`); run red.
- [ ] Guard test (#163 pattern, spec Done-when #2): a source-level check that no non-test Go file concatenates `history/issues` / `history/plans` literals outside `ArchiveSubdirs` — every consumer derives.
- [ ] Implement: `ArchiveSubdirs` in pkg/vocab; `isHistoryPath` accepts the three dirs (derive subdirs via `vocab.ArchiveSubdirs(historyDir)`); `NextID` adds the issues subdir to its scan list.
- [ ] Green; commit `#181: vocab.ArchiveSubdirs + layout-tolerant predicates`.

### Task 2: readers — resolve across both layouts (TDD)

- [ ] Failing test: resolve fixture with `000031-x.md` in `history/issues/`, `000031-x-plan.md` in `history/plans/`, and a legacy `000032-*` family flat in root — both resolve fully (family complete, ordering intact).
- [ ] Implement: `familyFiles` adds the two subdir globs (dirs list = Home, Plans, Archive, ArchiveSubdirs(Archive) ×2 — de-dupe survives overlap).
- [ ] Green; commit `#181: resolve tolerates flat + subfoldered archives`.

### Task 3: writers — archive into subfolders (TDD)

- [ ] Failing tests: update existing archive expectations to `history/issues/<base>` / `history/plans/<base>` — the verified inventory: push_test.go `TestArchiveDoneIssues_MovesAndClosesGH` / `TestPushPublishSequence_CodecompleteFlippedThenArchived` / `TestPreparedArchiveMoves*` / `TestRecoverInterruptedArchiveCommitsAndPushes` / `TestArchiveAddArgs`; archiveartifacts_test.go (all 8); merge_test.go `TestAssessDirty` / `TestArchiveDoneIssuesInDir_MovesTerminalAndRecordsRelativePaths` / `TestRunMerge_DirtyTrackerFile_Proceeds`; merge_e2e_test.go `TestRunMerge_CodecompleteFlippedToDoneAndArchived` + the resume test. `assessDirty` gains BOTH a dirty `history/issues/...` case AND a `history/plans/NNNNNN-x-plan.md` case (the plan-family one is the real regression: flat history plan files are Tracker today via issueFilename; without the predicate fix they'd become Blocking and refuse merges).
- [ ] Implement: `archiveDoneIssues` dest → issues subdir; `archiveDoneIssuesInDir` (merge.go:629 rename dest + :639 recorded rel path) → issues subdir — its issue-dest logic DUPLICATES push's (two write sites; extract shared computation if cheap); `archivePlanArtifacts` history dest → plans subdir (both its `historyFull` and `recHistoryDir` legs — derive inside via ArchiveSubdirs so callers stay root-typed); mkdirs per subdir.
- [ ] Green; full `go test ./cmd/sdlc/... ./pkg/...`; commit `#181: archive writes land in history/{issues,plans}/`.

### Task 4: migrate ariadne's 258 files + docs

- [ ] One commit: `git mv` — `*-plan.md` + `*-review.md` → `history/plans/` (99), the rest → `history/issues/` (159). Verify: `sdlc resolve` on 3 known archived ids (e.g. #175, #63, #12) returns complete families; `git status` clean post-commit; `git diff --check`.
- [ ] Docs: `issue.cue` archive comment names the Go-owned convention; atlas (`sdlc-binary.md` resolve/discovery prose + `pre-merge-checks.md`/`issue-lifecycle.md` archive mentions if stale); helptext sweep — 7 files mention workshop/history (push, merge, close, state, fetch, resolve, judge) + `.claude/skills/xx-issues/SKILL.md:32`; `Makefile:6` WF_HISTORY_DIR stays root-typed (unchanged); note downstream repos migrate lazily (reads tolerate flat).
- [ ] Mutation-check (#63): point `archiveDoneIssues` back at the root dest → the push e2e must go red; SAME check on merge's `archiveDoneIssuesInDir` dest (two write sites, two mutants); restore.
- [ ] Tick issue Plan; Log (ARCH-DRY: one derivation function; reads-tolerant/writes-strict rationale); close per the #174 protocol.

### Notes for the implementer

- `recoverInterruptedArchive` pairs porcelain-deleted issue paths with history files — it consumes `isHistoryPath`, so Task 1's fix should carry it; verify its test fixture explicitly.
- `preparedArchiveMove.HistoryPath` is staged verbatim (`git add -- <path>`) — subfolder paths flow through staging automatically; the merge caller records mainPath-relative paths, keep the rel/full distinction straight (push.go:281's doc comment explains it).
- `state.detectDrift`'s message says "move to workshop/history/" — cosmetically update to the issues subdir.
- Don't touch `stripCodeFences`-family or judge diff-collection unless a test proves prefix semantics break (expected: they don't — history exclusion is by path prefix).

## Review notes (fresh-eyes, 2026-07-15)

Verified by the plan reviewer against code (all folded above): consumer map
complete after adding merge.go:629/:639 (merge's own issue-dest logic — a
pre-existing push/merge duplication, ARCH-DRY debt this change can retire);
collectDiff empirically layout-agnostic (git default pathspec wildcards
CROSS `/` — `workshop/history/*.md` matches subdirs; only `:(glob)` magic
stops at slashes); recoverInterruptedArchive pairs by exact basename, so a
basename-preserving relayout is recovery-safe; migration classification is
exact (53 plan + 46 review = 99; 159 issues; the two frontmatter-bearing
plan docs are genuine plans — filename-suffix, not frontmatter, is the
reliable discriminator).

## Revisions

### 2026-07-15 — close-review (FIX-THEN-SHIP) deltas

1. **I1 fixed in-boundary:** `TestRecoverInterruptedArchive_SubfolderLayout`
   pins interrupted-archive recovery under the new layout (issue half via the
   terminal gate + plan-sidecar half), the promised-but-missed test-surface
   line. The flat-fixture recovery tests stay as legacy-tolerance pins.
2. Minors applied: merge's ArchiveSubdirs/MkdirAll hoisted out of the
   per-issue loop; state.go's drift hint derives via ArchiveSubdirs (no
   format-string subdir embed); the source-guard regex broadened to catch
   any history-rooted `filepath.Join(..., "issues"/"plans")` regardless of
   variable name plus `%s/issues`-style embeds; stale flat links in
   pensive/continuation swept to subdir paths.
3. Deferred knowingly: the push/merge issue-dest duplication consolidation
   (three near-identical computations counting archivePlanArtifacts) — next
   touch of the archive path should extract one helper.
