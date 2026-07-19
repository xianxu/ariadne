# Boundary Review — ariadne#171 (milestone M6)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | milestone M6 |
| milestone | M6 |
| window | 5f8182fbfd8af7fd0849e9a4f2a751499fe8bee1..HEAD |
| command | sdlc milestone-close --issue 171 --milestone M6 |
| reviewer | claude |
| timestamp | 2026-07-17T22:58:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M6 is a data-op milestone and the data op itself is verified sound: all four terminal records sit at their operator-confirmed destinations (`nous`/`kbench`/`metis` `workshop/history/projects/`), the three peer commits exist on each repo's `main` with exactly the claimed hashes (`34005ce`, `1f25273`, `8a4676c`) and the plan's commit-message template, frontmatter conforms to `#Project` under the M1 relaxed guard (verified against `construct/vocabulary/project.cue` — open struct, `done` exempt from baseline; `closed:` dates match each record's own body), and brain's `data/project/` holds only the active `metis-v2-experiment-algebra`. What keeps this from SHIP is an evidence-integrity problem: the Log's Step-6 verification claim that `project find --issue metis#2` returned "BOTH the archived metis-v1 and the brain-legacy metis-v2 (flagged)" is not reproducible against the current fleet state — brain's metis-v2 contains no `[metis#2<non-digit>` marker anywhere (its refs are all `metis#18`–`metis#33`), so `containsIssueMarker` (`cmd/sdlc/internal/project/discover.go:143`) cannot match it. The boundary record claims a live verification of the multi-match/legacy-flag path that the code and data say did not happen as described. (Process note: Bash was broken in this session — harness-level EPERM — so everything above was verified via direct file reads and peer-repo reflogs rather than command execution.)

**1. Strengths**

- Migration commits are clean and durable: one commit per destination repo, greppable message matching the plan template verbatim, all on `main` (nous/kbench at tip; metis has normal work on top). The Log's recorded hashes match the reflogs exactly — no hand-waved evidence here.
- Schema honesty carried through: `closed:` added only where absent and sourced from each record's own body (charon-launch-push `2026-05-04` matches its own line 10; kaggle `2026-07-02`; metis-v1 `2026-07-07`), and no `deadline`/`planned_finish` fabricated — precisely the M1 design intent, and `shared-brain` correctly left untouched (already had `closed: 2026-06-02`).
- The charon#13 limitation is correctly diagnosed against the real code path: `discoverProjectsForRef` fails at `resolveRepoDir` (`cmd/sdlc/projectfind.go:51`) when the ref's repo isn't a checked-out sibling — documenting it as a navigation limitation, not a discovery gap, is accurate.
- The manual-move-vs-`sdlc migrate` divergence is justified in the plan with concrete mechanism-level reasons (rewriteRefs would localize the qualified refs; dest-vantage verification fails closed on historical refs) — a deliberate, recorded exception, not drift.
- Brain end-state exactly matches the reconciled Done-when: only the active metis-v2 remains, preserving the "no terminal SDLC records in brain" thesis without archiving a live record mid-flight.

**2. Critical findings** — none.

**3. Important findings**

- **Unreproducible verification evidence in the boundary record** (`workshop/plans/000171-cross-repo-project-lift-plan.md:879-881`, note 2 of the M6 log entry). The claim "`metis#2` returns BOTH the archived `metis-v1` and the brain-legacy `metis-v2` (flagged) — the designed multi-match" cannot be produced by the shipped code against the current fleet: `metis-v1` matches via its `[metis#2]` markers (lines 45/78/218 of the migrated file), but brain's `metis-v2-experiment-algebra.md` has zero occurrences of `[metis#2` followed by a non-digit (all its refs are `metis#18+`; the only bracketed one is `[metis#18`/`[metis#22` inside scope lists, both boundary-rejected). No single ref exists today that multi-matches both files, since v1's refs (1–9) and v2's (18+) are disjoint. Consequence: the live multi-match + legacy-flag behavior the Log presents as verified was not verified (it is unit-pinned by M4's `TestProjectFind_FleetWideArchiveInclusive`, but the Log claims a live run). Fix sketch: correct the Log/plan note to what actually ran; if a live legacy-flag check is wanted, use a ref genuinely bracketed in metis-v2 (e.g. `metis#18`) — that returns only the brain record with `(legacy)`, which is the flag path; drop the multi-match claim or demonstrate it with a fixture.

**4. Minor findings**

- The issue's `## Plan` M6 row (`workshop/issues/000171-...md:121`) is already `[x]` yet the window diff contains no issue-file change — the tick was committed at or before the M5-close base commit, i.e. before M6 executed. Bookkeeping-order slip; worth a lessons glance if it was bundled into M5's "boundary bookkeeping".
- Plan Step 3 names `sdlc project validate` as the primary validation command; no such subcommand exists (only `sdlc issue validate` — the parenthetical `bin/vocabulary` path is the real one). The Log doesn't record which tool actually ran. I re-derived conformance from `project.cue` and all four pass, but the evidence line should name the actual command.
- The M6 plan row says "committed both sides" while the brain side is staged, riding nous's auto-commit (as Step 5 designed). Disclosed honestly in the Log — just confirm at issue close that the sweep actually landed the deletion commit.
- `metis-v1.md`'s `sources: [workshop/pensive/2026-07-03-01-...]` is a repo-relative pointer to what is actually `brain/workshop/pensive/...` — dangling from its new metis vantage (contrast metis-v2's repo-qualified `metis/workshop/pensive/...`). Accepted by the verbatim-freeze design; noting for anyone who later chases the pointer. Same class: `[metis#2]: ../../metis/...` reference-link at line 218 doesn't resolve from the new depth.

**5. Test coverage notes**

No code changed in this window, so no new tests owed. The behaviors M6 leans on are pinned from earlier milestones (`TestProjectFind_FleetWideArchiveInclusive` covers archived + legacy-flagged matches; `TestContainsIssueMarker` covers the ID boundary). One knock-on from the Important finding: the plan's Manual Verification item 4 ("close an issue referenced by the still-resident metis-v2 and confirm the deprecation warning") must use a ref actually bracketed in metis-v2 — `metis#18` qualifies; `metis#2` does not.

**6. Architectural notes** (ARCH sweep)

- **ARCH-DRY: pass.** No code; the migration reuses the existing `#181` per-kind archive layout (nous had already migrated its history to per-kind subfolders — reflog line 526 — so `workshop/history/projects/` is consistent fleet-wide) and the one shared discovery walker.
- **ARCH-PURE: pass (n/a).** Data-only window; the pure/IO split was M2/M3's surface and is untouched.
- **ARCH-PURPOSE: pass.** Shadow-sweep of the migration's consumers: all four records derive from their new homes (destination files present + committed; discovery's archive scope reaches them mechanically — `[kaggle#1 M1]` and `[metis#2]` markers verified in the migrated files); brain retains only the active record per the reconciled Done-when; no hand-maintained copy of the four remains anywhere. The deferred item (metis-v2's own migration at its close) is a genuine separable follow-up, not the issue's purpose. The one purpose-adjacent gap is the evidence item above — the *verification* under-delivered, not the migration.

**7. Plan revision recommendations**

One `## Revisions` entry on `workshop/plans/000171-cross-repo-project-lift-plan.md`: correct M6 log note (2) — `metis#2` returns only the archived `metis-v1`; the brain-legacy multi-match as described is not producible (v1/v2 ref ranges are disjoint), record the command(s) actually run, and if live legacy-flag evidence is wanted, substitute a `metis#18`-class query. Optionally amend Step 3 to name the validation tool that actually exists/ran instead of `sdlc project validate`.
