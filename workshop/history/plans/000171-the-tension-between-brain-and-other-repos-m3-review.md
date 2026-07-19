# Boundary Review — ariadne#171 (milestone M3)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 368362fd4cbad2ed8ab5d950e368e0c746bf68c6..HEAD |
| command | sdlc milestone-close --issue 171 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-07-17T21:01:38-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The review is complete. All checks done statically (Bash was unavailable the whole session — the known EPERM failure mode). Final report:

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M3 delivers what the plan's Chunk 3 promised, and delivers it well: a genuinely pure decision core (`planPeerWrites`) with a complete decision-table unit test, a thin shell over the existing `gitRunner` seam, and real multi-repo git fixtures — including the cross-repo close test that the M2 review explicitly deferred to this boundary. Every failure path degrades in the safe direction (report-only or warn-and-continue; the close never fails on a peer-write outcome), matching the Done-when clause "committed (scoped) or loudly reported." Nothing is Critical. What keeps this from SHIP: one behavioral-safety gap (pre-existing *unstaged* edits to a peer's project file get silently absorbed into the auto-commit — the exact hazard class the staged-index refusal exists to prevent), plus the fact that I could not execute the test suite in this session (Bash was broken with the documented `~/.claude/session-env` EPERM), so my confidence is static-analysis-grade: **the main agent must run `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/...` and record green in the close bundle.**

## 1. Strengths

- **ARCH-PURE: pass, and genuinely.** `planPeerWrites` (`cmd/sdlc/peerwrite.go:53`) is pure — data in, decisions out, deterministic sort — and `peerwrite_test.go`'s decision table runs with zero IO and zero mocks. The M3 delta of snapshotting `IsBrain` into `RepoGitState` in the shell (rather than calling `gitx.IsBrainRepo` from the planner) is exactly the right move to keep the planner pure.
- **ARCH-DRY: pass.** Reuses the existing `gitRunner` seam (`runner.go:24`), single-sources the brain predicate via `gitx.IsBrainRepo` (`peerwrite.go:96`), and threads one package-level `closeRunner` through all four `applyClose` callsites (including `milestoneclose.go:163`, which the plan's caller list had missed — caught and documented in `## Revisions`).
- **Fail-safe throughout.** Unknown state → report-only; unborn HEAD or non-git dir → `diff --cached` fails → treated as staged → report-only; git add/commit failure → `cwarn` and continue (`peerwrite.go:105-125`, no `die` in the path). The conservative default direction is consistent.
- **Pathspec-scoped commit with honest assertions.** `git commit -m … -- <files>` plus tests asserting exactly one file touched (`peerwrite_apply_test.go:170-173`) and a clean tree afterward — the TOCTOU belt-and-braces the Revisions entry claims is actually pinned.
- **The deferred M2 debt was paid.** `TestClose_PeerProjectCommitted` is the promised end-to-end "matched project lives in a different repo than the closing one" test, and `TestClose_CurrentRepoProjectNotAutoCommitted` pins the current-repo-rides-the-close-commit invariant. Atlas (`atlas/workflow/sdlc-binary.md`) was updated in-window, and the plan's `## Revisions` honestly records five deltas from the sketch.

## 2. Critical findings

None.

## 3. Important findings

- **`cmd/sdlc/peerwrite.go:74` — unstaged edits to the peer's project file are absorbed into the auto-commit.** The refusal checks branch + *staged index* only. If a peer is on main with a clean index but another session has uncommitted working-tree edits to the very project file being ticked, `computeClose` reads those edits from disk, `applyClose` writes them back merged with the tick, and the scoped commit publishes them under `project: close-time update (…)` — precisely the "absorb another session's work" hazard the `HasStagedChanges` refusal articulates (`peerwrite.go:75`). The code matches the plan's letter (`Branch == "main" && !HasStagedChanges`) but not its safety intent. Fix sketch: detect pre-write dirt on the target files — either read `RepoGitState` *before* the `os.WriteFile` loop in `applyClose` and add a `git diff --quiet HEAD -- <relpaths>` check per peer, or (no reordering needed) compare `projectEdit.oldText` against `git show HEAD:<rel>` and set a `TargetFilesDirty bool` on `RepoGitState`; either becomes a fifth report-only row in the pure planner. Add the corresponding decision-table case and an end-to-end fixture. If you judge this out of M3's reviewed scope, it needs at minimum a named follow-up in the issue Log — not silence.

## 4. Minor findings

- `peerwrite.go:60` — the manual `NextAction` doesn't shell-quote paths (`cd %s`, `git add %s`); a repo path with spaces yields a broken command (the `-m` message is correctly `%q`-quoted).
- `peerwrite.go:91` — when `rev-parse` fails (e.g. a non-git sibling dir holding a project file), the trimmed *error text* becomes `Branch`, so the reason reads `is on branch "fatal: not a git repository…"`. Safe direction, garbled message; blank the branch on error and let the unknown/off-main reason speak.
- `peerwrite.go:69` — the brain report-only `NextAction` tells the operator to hand `git commit` into brain, but brain's charter is the nous auto-commit rhythm; the suggested command contradicts the very reason cited. Consider a brain-specific next-action ("nous will sweep it") instead of the generic manual command.
- `workshop/plans/000171-cross-repo-project-lift-plan.md:110` — the milestone-map row still reads `- [ ] M3` while the issue's `## Plan` M3 row is `[x]` and all Chunk 3 steps are ticked; tick it in the close bundle.
- `printCloseDryRun` doesn't preview the peer-commit decision — already documented as a future extension in the plan's `PeerWriteDecision` entry, so no action; noting for completeness.
- Hardcoded `"main"` means a peer with a `master` default branch can never auto-commit (permanent report-only). Conservative-safe and fine for this fleet; worth a comment or config if the fleet ever diverges.

## 5. Test coverage notes

- Covered well: the full planner decision table (commit / off-main / staged / brain / unknown / current-repo-omitted / empty), determinism, real-git `readRepoGitState` (staged flip, branch, brain predicate), shell commit + report-only against real repos, and three end-to-end `runClose` fleet scenarios.
- Gaps: (a) the unstaged-dirty-target-file scenario (the Important above); (b) `applyPeerWrites`'s git-failure path (warn-and-continue) is structurally evident but unpinned — a stub runner that errors would cheaply pin "close never fails"; (c) peer commits are exercised only via issue-close (`runClose`); the milestone-close mode reaches the same `applyClose` so risk is low, but no test proves it.
- **This session could not execute anything** — the suite result must come from the main agent before the boundary is crossed.

## 6. Architectural notes for upcoming work

- **ARCH-PURPOSE: pass for this boundary.** M3's slice of the Done-when ("peer-repo write committed (scoped) or loudly reported") is fully delivered; the remaining consumers of the purpose (M4 navigation, M5 residency docs, M6 migration) are genuine planned milestones, not disguised deferrals of M3's point.
- The per-test-file local git helper (`gitIn`, `initFleetRepo`) is now the ~12th private copy of the same fixture idiom across `cmd/sdlc` tests. Not this diff's debt, but a shared `internal` test fixture package is overdue — a candidate side-quest or issue.
- M4's `sdlc project find`/`resolve` will consume `DiscoverByIssueRef(ActiveAndArchive)`; the `ProjectMatch.Status` future-extension noted at M2 may become live there — keep it in mind before adding a second per-consumer re-parse.

## 7. Plan revision recommendations

The plan's 2026-07-17 M3 `## Revisions` entry already reconciles the shipped deltas accurately (brain-in-state, `closingRef`, unknown row, `runClose` stdout, pathspec scoping) — no correction needed there. Two small items:
- Tick the milestone-map row `- [ ] M3` (line 110) to match the issue Plan and the ticked Chunk 3.
- If the unstaged-dirty-file gap is fixed at this boundary, append it to the M3 Revisions entry as delta 6 (working-tree-dirty target files → report-only); if deferred instead, record the deferral decision in the issue `## Log` so the plan doesn't silently claim the safety property is complete.
