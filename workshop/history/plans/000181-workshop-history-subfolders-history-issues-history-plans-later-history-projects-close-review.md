# Boundary Review — ariadne#181 (whole-issue close)

| field | value |
|-------|-------|
| issue | 181 — workshop/history subfolders: history/issues, history/plans, later history/projects |
| repo | ariadne |
| issue file | workshop/issues/000181-workshop-history-subfolders-history-issues-history-plans-later-history-projects.md |
| boundary | whole-issue close |
| milestone | — |
| window | 804de760841354dc7c3d7d7b5ed48dc730920e43..HEAD |
| command | sdlc close --issue 181 |
| reviewer | claude |
| timestamp | 2026-07-15T17:00:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All verification is done. Composing the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This boundary delivers what #181 committed to, and delivers it the right way: the layout convention is single-sourced in `vocab.ArchiveSubdirs`, writers emit strictly into `history/{issues,plans}/`, readers tolerate the pre-#181 flat layout, and the 258-file migration is real and complete on disk (I verified: 0 flat `.md` left in `workshop/history/`, 159 in `issues/`, 99 in `plans/`). The consumer sweep is thorough — resolve, NextID, isHistoryPath (and through it assessDirty/isTrackerPath and the recovery pairing), push + merge writers, collectDiff confirmed layout-agnostic by pathspec semantics, plus atlas/helptext/cue docs. What keeps this from SHIP outright is one Important test-coverage gap the plan itself promised (interrupted-archive recovery with subfolder dests) and a small cluster of Minors. Caveat: this review harness could not execute Bash (session-env EPERM even with sandbox disabled), so I could not run `go test` myself — findings are from reading the code and test files, not from a green run; the main agent should confirm the suite is green.

**1. Strengths**

- `ArchiveSubdirs` (pkg/vocab/vocab.go:106) is exactly the right shape: derives from an arbitrary root (so `--history-dir` overrides stay consistent), documented rationale for Go-owned vs cue-encoded (downstream JSON compat), and a named extension point for #180's `history/projects/`. ARCH-DRY done properly.
- The #163-pattern source guard (`TestArchiveSubdirs_SingleDerivationPoint`) turns the "no hardcoded layout" Done-when into an enforced invariant rather than a review-time hope.
- `TestAssessDirty`'s new `history/plans/` case (cmd/sdlc/merge_test.go:93) pins the real latent regression — a migrated plan file flipping from Tracker to Blocking and refusing merges — and the comment says why. This is the highest-value test in the diff.
- Reads-tolerant/writes-strict is applied consistently: `familyFiles` globs root + both subdirs with de-dupe (resolve.go:229-244), `isHistoryPath` accepts all three dirs with the depth check pinned (`.../issues/deeper/...` → false), `NextID` keeps the flat root in its scan. Downstream repos genuinely migrate lazily.
- The test sweep replaced substring assertions with exact-arg checks where nesting made substrings vacuous (archiveartifacts_test.go:114-119) — the exact gotcha the Log records.

**2. Critical findings**

None.

**3. Important findings**

- **Missing test: interrupted-archive recovery with subfolder dests** — cmd/sdlc/push_test.go:203-287. `TestPreparedArchiveMovesDetectsUnstagedMove`, `TestPreparedArchiveMovesRejectsNonTerminalHistoryFile`, and `TestRecoverInterruptedArchiveCommitsAndPushes` all still use only flat `workshop/history/000036-*.md` paths — yet the writers now produce subfolder paths, so a real interrupted `sdlc push` leaves `?? workshop/history/issues/...` porcelain, and that composition (porcelain parse → basename pairing → `historyFileIsTerminal` read of the subdir path; plus the plan-sidecar leg `?? workshop/history/plans/...-plan.md` with its terminal-gate bypass) is exercised nowhere. The plan's test surface explicitly listed "plus `recoverInterruptedArchive` with subfolder dests". Fix sketch: clone `TestRecoverInterruptedArchiveCommitsAndPushes` with `workshop/history/issues/000036-done.md` + a `workshop/history/plans/000036-x-plan.md` half. Cheap; catches exactly the class of bug this diff could ship in its production-recovery path.

**4. Minor findings**

- cmd/sdlc/state.go:320 — `"move to %s/issues/"` hand-concatenates the subdir name in a format string, outside `ArchiveSubdirs`, and the guard test's literal patterns can't see it (ARCH-DRY, display-only). Either derive the string or accept it as a known cosmetic escape.
- The guard test's patterns (vocab_test.go) only match `filepath.Join(historyDir, "issues")` by that exact variable name — `filepath.Join(historyFull, "issues")` or format-string embeds slip through (state.go:320 is a live example). Consider a broader regex.
- cmd/sdlc/merge.go:630-633 — `ArchiveSubdirs` + `MkdirAll` run inside the per-issue loop; hoist above it.
- The push/merge issue-dest duplication survives (merge.go:626-629 comment acknowledges it; the plan said "extract if cheap"). Acceptable deferral, but it's now three near-identical dest computations counting `archivePlanArtifacts` — next touch should consolidate.
- Stale flat links to migrated files remain outside atlas: `workshop/pensive/2026-07-14-01-...md:7,224,226`, `workshop/pensive/2026-06-17-01-...md:7`, `workshop/continuation/20260701T162301-...md:82` all reference `workshop/history/000NNN-*` paths that no longer exist. Low-signal archival docs, but the frontmatter `references:` lists are now dangling.

**5. Test coverage notes**

Good: predicate table test (`TestIsHistoryPath` — new direct coverage for a previously indirect predicate), `NextID` subfolder case, resolve fixture covering both layouts as complete ordered families, assessDirty both-subdir cases, e2e merge/push archive assertions all moved to subdir expectations, untracked-sidecar staging contract re-pinned with exact-arg matching. Gap: the recovery path (Important above). Unverifiable from here: the Log's mutation-check claims (two mutants) and a green `go test ./cmd/sdlc/... ./pkg/...` — I could not execute commands; main agent should confirm.

**6. Architectural notes**

- **ARCH-DRY: pass with two blemishes** — single derivation point + guard test is the right structure; the state.go format-string escape and the retained push/merge dest duplication are the residue (both flagged above).
- **ARCH-PURE: pass** — `ArchiveSubdirs`, `isHistoryPath`, `assessDirty` are pure and directly unit-tested without IO; the writers stay thin IO seams; no new mock-dependent "pure" entities.
- **ARCH-PURPOSE: pass** — shadow-sweep of consumers: push writer, merge writer, plan-artifact writer, resolve reader, NextID, recovery predicate chain, collectDiff (pathspec-agnostic, confirmed), drift message, cue comment, atlas, helptext — every consumer derives or is genuinely layout-agnostic; the migration (the point of the issue) actually happened. `history/projects/` is correctly deferred to #180 ("later" per the Spec — separable, not the purpose).
- For #180: `ArchiveSubdirs`'s two-value return will need widening (third return vs kind-keyed map) — the kind-keyed variant scales better if more kinds appear; decide there, but the comment already flags it.

**7. Plan revision recommendations**

The plan matches the code except one test-surface line it promised and didn't deliver: "plus `recoverInterruptedArchive` with subfolder dests". Preferred resolution is to add the test (Important above) rather than revise the plan. If it's consciously skipped instead, add a `## Revisions` entry to `workshop/plans/000181-history-subfolders-plan.md` stating the recovery-path subfolder case is deferred and why the predicate-level test is deemed sufficient.
