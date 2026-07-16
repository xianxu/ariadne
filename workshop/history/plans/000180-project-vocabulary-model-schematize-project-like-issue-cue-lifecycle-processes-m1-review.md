# Boundary Review — ariadne#180 (milestone M1)

| field | value |
|-------|-------|
| issue | 180 — project vocabulary model: schematize project like issue (cue + lifecycle + processes) |
| repo | ariadne |
| issue file | workshop/issues/000180-project-vocabulary-model-schematize-project-like-issue-cue-lifecycle-processes.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 84f1e1286f066761ec63f5f96e34c1a4311a6705^..HEAD |
| command | sdlc milestone-close --issue 180 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-07-16T11:35:15-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 delivers what the plan's M1 row claims, and delivers it well: `project.cue` models the funnel with compiled baseline guards and vet-enforced laws, `pkg/vocab.Project()` mirrors the issue binding exactly, the third-noun helper extraction is genuinely behavior-preserving (existing pins untouched), and the kind-keyed `ArchiveSubdir` migration is complete — my shadow-sweep of every archive-layout consumer (push/merge writers, interrupted-archive recovery, `assessDirty`, `resolve` family globs, `state` drift hint, `NextID`, and `judge`'s pathspecs, where git's default `*` matches across `/` so review windows still see subfoldered files) found each one deriving through the single function or read-tolerant of the flat layout. Nothing blocks the boundary; the findings below are cheap. One caveat on confidence: the Bash harness was broken in this review session (sandboxed and unsandboxed shells both failed before command execution), so I could not independently re-run `go test`/`vet_test.sh` — verification is by close reading against the implementor's logged bare green run.

**1. Strengths**
- `construct/vocabulary/testdata/project_invalid.cue` is self-contained by design so vet fails on the enum conflict rather than a vacuous missing-reference — the exact trap the plan's chunk review flagged, executed correctly.
- `pkg/vocab/conformance_test.go:125-131` pins the *deliberate absence* of `paused→done`, not just the presences — negative-space testing that will keep M3/M4's close verb honest.
- `TestArchiveSubdir_SingleDerivationPoint` (`pkg/vocab/vocab_test.go:168`) plus `TestRecoverInterruptedArchive_SubfolderLayout` and `TestResolveRun_SubfolderedAndFlatArchive` cover the three ways the layout change could regress: hand-concatenation, crash recovery, and read-path resolution across both layouts.
- The `#181`-comment in `construct/vocabulary/issue.cue` honestly documents *why* the subdir derivation is Go-owned rather than cue-modeled (flag-override roots, downstream JSON stability) — the kind of rationale that prevents a future "just model it in cue" regression.
- `cmd/vocabulary` needed zero edits — project auto-registers via `resolveVocab`'s filename merge, confirming the noun layer's design paid off at the third noun.

**2. Critical findings** — none.

**3. Important findings**
- **Stale local vocabulary materialization** — `construct/generated/vocabulary/` (gitignored, but the *served* face: `.claude/skills/xx-vocabulary` symlinks to it) still lists "Defined nouns: issue, pensive, verdict" and has no `project.json`; `.source-sha` no longer matches the source set, so `vocabulary check` will report STALE. Agents reading the live skill won't see the project noun while M2–M4 build on it. Fix: run `make weave` (one command) before crossing the boundary.

**4. Minor findings**
- Plan M1.2's Files list claims `pkg/vocab/lifecycle_test.go`; it was not created (helpers are covered indirectly via the three nouns' existing pins — defensible, but the plan should stop claiming the file; see §7).
- `archiveJoinRE` (`pkg/vocab/vocab_test.go:209`) only catches roots whose variable name contains "history" — `filepath.Join(d.Archive, "issues")` would evade the guard (`resolve.go` routes correctly today; the blind spot is future-facing).
- `merge.go:415,418` dry-run output and `push.go:210,376` success lines still print the archive *root* while real writes land in the `issues/` subdir — cosmetic message drift.
- merge.go:618-622's issues-subdir derivation duplicates push.go's `archiveDoneIssues` leg — acknowledged in-code as pre-existing two-write-site debt (ARCH-DRY); worth a consolidation ticket, not this boundary.
- `vocab.go:12-19` import grouping (`path/filepath` isolated above the stdlib group) is unidiomatic.
- `TransitionFor` returns a pointer into the shared model's slice — a caller could mutate the singleton. Consistent with the package's exposed-map convention, but note it when the guard runner consumes it in M3.

**5. Test coverage notes**
Coverage matches the failure modes this diff could ship: layout pins from arbitrary roots, the source-scan derivation guard, subfolder recovery through the porcelain parser, both-layout resolution ordering, `NextID` over the archived subdir, and a full conformance mirror for the project noun. I hand-verified `pkg/vocab/project.json` field-by-field against `project.cue` (categories, when, discovery, scaffold, all 11 lifecycle edges with `guards: []` where the cue omits them) — consistent; and the dogfood instance `workshop/projects/project-management-primitive.md` conforms to `#Project` as written (status `ideation` pre-baseline, so the compiled deadline guard correctly doesn't bite). Gap: nothing exercises `ProjectModel.LegalTransitions`/`StatusNames` directly, but they're one-line delegations to helpers pinned elsewhere.

**6. Architectural notes** (each ARCH marker worked explicitly)
- **ARCH-DRY: pass.** The lifecycle-helper extraction lands exactly at the rule-of-three trigger and both existing nouns now delegate. Residual duplication is the acknowledged merge/push archive-writer pair (Minor above).
- **ARCH-PURE: pass.** Everything new in `pkg/vocab` is pure and tested without IO (embed is compile-time); the cmd-layer changes stay in the existing thin IO seams; the one IO-reading test is a deliberate source-scan meta-guard.
- **ARCH-PURPOSE: pass for this boundary, with a watch item.** Shadow-sweep of the archive-layout consumers found no hand-maintained restatement left behind. But `construct/datatype/project.md:28` still hand-maintains the *old* project enum (`active|paused|done|dropped`) that the model now contradicts — and the new dogfood instance carries `status: ideation`, which the current prose calls invalid. The demotion is scheduled in-issue at M5, so it's a sequenced consumer, not an abandoned one — do not let M5 slip out of this issue, or M1 recreates the exact gap the issue exists to close.

**7. Plan revision recommendations**
Add one `## Revisions` entry to `workshop/plans/000180-project-vocabulary-model-plan.md`: M1.2's Files list drops `pkg/vocab/lifecycle_test.go` (helpers are pinned via the issue/verdict/project model tests — behavior-preserving extraction verified by unchanged pins), and M1.4's "9 call sites" reconciles to the 11 actually migrated (9 non-test + 2 test, per the issue's Estimate). Optionally note in the same entry that the plan-Revisions' dogfood delta ("mvp_scope [#180, #171], #182 explicitly out") was superseded same-day by the scope event moving #182 into mvp_scope — the project file's Log records it, but the plan currently states the pre-event scope as current.
