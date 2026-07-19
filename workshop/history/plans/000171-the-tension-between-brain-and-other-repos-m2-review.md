# Boundary Review — ariadne#171 (milestone M2)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | f1d6ca6c9d675dff95cc62516d19c4dc42bfd247..HEAD |
| command | sdlc milestone-close --issue 171 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-07-17T17:38:03-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers what it claims: `FindByIssueRef`'s single-brain-dir, refuse-on-multiple lookup is cleanly replaced by a fleet-wide, scope-aware `DiscoverByIssueRef`, the close gate loops all matches through the existing tick/upsert helpers, `resolveRepoDir` is rebuilt on the shared walk without behavior change, and the atlas is updated in-range. The shadow-sweep is clean — no code consumer of the removed function remains. Nothing blocks shipping, but three Important items should be fixed cheaply before the boundary: the brain repo is detected by hardcoded basename instead of the codebase's canonical `.brain/config.md` predicate, the `RepoDir` field that M3's peer-commit scoping will consume is never asserted by any test, and the plan's ticked Step 5 claims two edge-case tests (symlink dedup, unreadable file) that were not written. Confidence is medium, not high, because **I could not execute the test suite**: the Bash harness is broken in this review session (every command fails with `EPERM … mkdir ~/.claude/session-env/<id>` before launch; a delegated subagent hit the identical failure). The main agent must run `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/` and confirm green before finalizing — my review is static-analysis only.

## 1. Strengths

- **Plan-then-apply discipline extended, not eroded.** `projectEdit` carries `oldText`, and both `applyClose` and the TOCTOU snapshot `validate()` loop the slice (`cmd/sdlc/close.go:1114-1121`) — the #139 "no writes before a finalizing verdict" invariant survives the single→many generalization intact, including the mid-review-mutation guard (and the updated `TestCloseCommand_ProjectChangedDuringBoundaryReview_DoesNotFinalize` still pins it).
- **The `SiblingRepoDirs` extraction is honest ARCH-DRY.** Keeping it deliberately unfiltered and placing the skip-list in `isFleetSibling` (used only by the fleet glob) makes `resolveRepoDir` *literally* behavior-identical — a better resolution than the plan's own text, which had the skip-list inside the shared helper and claimed identity at the same time. `TestSiblingRepoDirs_ReturnsAllDirs` pins the no-filtering contract explicitly.
- **Vocab-derived paths.** `workshop/projects` / `workshop/history/projects` come from `vocab.Project().Discovery()` + `vocab.ArchiveSubdir` (`discover.go:78-80`), not literals — consistent with the #163/#181 single-derivation guard (whose source test would have caught a literal).
- **The scope parameter does the job it was added for.** `dropTerminalLegacy` closes the exact re-tick hazard the plan-quality review flagged (a `done` brain-legacy record under `ActiveOnly`), with both directions tested (`TestDiscoverByIssueRef_DropsTerminalLegacyUnderActiveOnly`).
- **Docs gate satisfied**: `atlas/workflow/sdlc-binary.md` documents the new discovery surface in-range; README already states the `workshop/projects/` residency, and no new user-typed surface was added, so no README change is owed.

## 2. Critical findings

None found statically — but tests were not executed in this review (harness failure above), so the green-suite claim is unverified. Treat "run the suite" as a mandatory pre-close step, not a formality.

## 3. Important findings

1. **Brain detected by basename, not the canonical predicate** — `discover.go:116` (`if base == "brain"`). The constitution (§1) and the codebase both define brain as "`.brain/config.md` exists" (`repoguard.go:70`, `migrate.go:259`); this is a third, divergent restatement of that predicate (ARCH-DRY). Two failure modes: (a) a brain repo checked out under any other name gets its legacy `data/project` silently unscanned *and* its `workshop/projects` treated as a normal fleet home — which in M3 becomes a peer auto-commit into an auto-committing brain repo, exactly what #176 forbids; (b) a non-brain repo that happens to be named `brain` never has its canonical home scanned. Fix sketch: `if _, err := os.Stat(filepath.Join(repoDir, ".brain", "config.md")); err == nil { scan(repoDir, filepath.Join("data","project"), true); continue }`.
2. **`ProjectMatch.RepoDir`/`Repo` are never asserted** — every discover test checks counts and `Legacy` only, and the one close-level multi-match test (`TestRunClose_UpdatesAllMatchingProjects`, `close_finalize_test.go:297`) puts both projects in the *same* repo, so `projectEdit.repoDir` is exercised only for the trivial self-repo case. `RepoDir` is precisely the field M3's peer-write commit scoping consumes; a wrong value ships silently today and detonates in M3. Add `RepoDir`/`Repo` assertions to `TestDiscoverByIssueRef_AllMatchesAcrossPeers`, and note the plan's Chunk 3 Step 1 specified the close test with two projects "in two peer repos" — the delivered same-repo variant weakens that.
3. **Plan Step 5 (Chunk 2) is ticked but partially delivered** — it enumerates "a symlinked duplicate path counted once; an unreadable file skipped" as required edge-case tests; neither exists in `discover_test.go` (grep confirms). The dedup code (`discover.go:87-93`) and best-effort read-skip (`discover.go:94-97`) are live but unpinned. Either add the two small tests or un-tick/annotate the step (requirements traceability).
4. **Plan/code contradiction on skip-list placement** — plan Task 2.1 Step 1 and the Integration-points bullet ("`siblingRepoDirs` … applying only the spurious-sibling skip-list", plan line 75) contradict the shipped design (unfiltered `SiblingRepoDirs`, skip-list in `isFleetSibling`). The code is the better design; the plan needs a `## Revisions` entry so it stops describing a helper that doesn't exist as specified (see §7).

## 4. Minor findings

- `close.go:314-320` — `closeResult`'s #139 doc comment is now fused onto `type projectEdit`; `projectEdit`'s godoc opens with "closeResult bundles everything applyClose needs" and `closeResult` is undocumented. Split the comment block.
- Applied/warn messages identify projects by `filepath.Base(m.Path)` only (`close.go:600,602,648`); with fleet-wide multi-match, same-named files in different repos are indistinguishable — prefix with `m.Repo`.
- `dropTerminalLegacy` re-reads and re-parses files `scan` just read; the plan's own "Future extensions: add a Status/FM field to ProjectMatch" is the fix. Cold path — fine to defer.
- Milestone close over N matched projects `die`s on the *first* file missing a detail-block anchor (`close.go:628`), so multi-membership fixes surface one re-run at a time. UX note for M3/M4.
- `filepath.Glob` error discarded (`discover.go:85`) — only malformed-pattern risk; consistent with the removed code.

## 5. Test coverage notes

- **Not executed** — Bash is broken in this review session (EPERM creating `~/.claude/session-env/<id>` before any command launches; reproduced with sandbox bypass and via a fresh subagent). The main agent must run `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/` before recording this verdict as satisfied. Static reading found no compile hazards (`os` still used in `resolve.go`; `vocab.ArchiveSubdir`/`ArchiveProjects`/`IsTerminal` all exist; `project` import added to `resolve.go`).
- Delivered coverage is otherwise solid: cross-peer all-match, ActiveOnly-vs-ActiveAndArchive, terminal-legacy drop (both directions), stale-sibling skip, zero matches, same-repo double membership, unfiltered `SiblingRepoDirs`, close-level multi-tick, and the multi-project TOCTOU guard.
- Gaps: `RepoDir` assertions (finding 3.2), symlink-dedup + unreadable-skip (finding 3.3), and no test of a close whose matched project lives in a *different* repo than the closing one.

## 6. Architectural notes (ARCH-* pass/flag)

- **ARCH-DRY: flag (one item).** The sibling walk, archive-path derivation, and tick/upsert reuse are all genuinely single-sourced — pass on those. The hardcoded `"brain"` basename is a third hand-rolled brain predicate alongside `repoguard.go` and `migrate.go` (finding 3.1); consider exporting one predicate M3 can also use before it starts committing into peer repos.
- **ARCH-PURE: pass, with a labeling quibble.** `DiscoverByIssueRef` is filesystem-reading discovery; its tests are mock-free and deterministic against temp dirs (the repo's established `resolve_test.go` pattern), so the spirit holds. But the plan's Core Concepts table lists it under "Pure entities" while its tests require a mutable fs — it is an fs-integration seam, not PURE by this repo's review definition. Reclassify in the plan (see §7) rather than change code. The `computeClose` (pure planning) / `applyClose` (writes) split is preserved and cleanly extended.
- **ARCH-PURPOSE: pass.** M2's boundary scope — discovery, all-match close wiring, legacy deprecation warning — is fully delivered; the shadow-sweep found no surviving `FindByIssueRef` consumer, and the deferred consumers (`find`/`resolve` project kind/parley) are M4's declared scope, not the purpose smuggled into a follow-up. For M3: today `computeClose` writes peer files with *no* commit mechanics — the plan says so, but until M3 lands, a close can leave an uncommitted edit in a peer repo's working tree; don't let M3 slip behind M4/M5 in execution order.

## 7. Plan revision recommendations

Add one `## Revisions` entry (2026-07-17, M2 boundary review) to `workshop/plans/000171-cross-repo-project-lift-plan.md` covering:

1. **Skip-list placement delta**: `SiblingRepoDirs` ships unfiltered (exact `resolveRepoDir` identity, pinned by `TestSiblingRepoDirs_ReturnsAllDirs`); the spurious-sibling skip-list lives in `isFleetSibling` inside `DiscoverByIssueRef`. Supersedes Task 2.1 Step 1's and the Integration-points bullet's description.
2. **Classification fix**: move `DiscoverByIssueRef` (and the `ProjectMatch` producer path) from "Pure entities" to the integration/seam table, or annotate it as deterministic-given-fs — its tests require a real filesystem.
3. **Chunk 2 Step 5 honesty**: either record the symlink-dedup and unreadable-file tests as added (after writing them) or annotate the step as partially delivered.
4. (If adopted) the brain-predicate change from basename to `.brain/config.md` stat.
