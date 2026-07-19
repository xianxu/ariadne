# Boundary Review — ariadne#171 (milestone M1)

| field | value |
|-------|-------|
| issue | 171 — the tension between brain and other repos |
| repo | ariadne |
| issue file | workshop/issues/000171-the-tension-between-brain-and-other-repos.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | a838786903a29b251167b744511040bffeff24fa^..HEAD |
| command | sdlc milestone-close --issue 171 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-07-17T17:09:28-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** M1's substance is delivered correctly and completely: the compiled baseline guard in `construct/vocabulary/project.cue:92` now exempts `done` (and only `done` — `committed`/`executing`/`paused` still require the baseline), the vet gate gained both a positive fixture (`project_done_no_baseline.json` validates) and a negative control (`project_executing_no_baseline.json` still rejected), and the datatype prose rows (`construct/datatype/project.md:29-30`) were updated to match, exactly per the plan's Task 1.2 Step 3. Every consumer of the guard derives from the cue (the `vocabulary validate-instance` gate reads the `.cue` directly; `pkg/vocab/project.json` never contained the guard since `#`-definitions don't export), so the model stays single-sourced. The one gap: mtime evidence shows the gitignored-but-served `construct/generated/vocabulary/` materialization (`.source-sha` freshness stamp) was not re-woven after the cue edit — the identical Important finding #180's M1 review raised, and the durable plan's M1 chunk never scheduled the `make weave` step despite the lessons.md entry from that episode. Cheap to fix before the boundary; nothing blocks. One environment caveat: the Bash tool is broken in this review session (harness `session-env` mkdir failure even unsandboxed), so I verified the vet/test semantics statically rather than executing them — the close commit should carry the actual run evidence.

## 1. Strengths

- **The negative control is the right test.** `vet_test.sh:48-50` pins that an `executing` record without a baseline is *still* rejected — the exact over-relaxation bug this diff could have shipped. The plan only sketched the positive case; the implementor added the guard against the opposite failure.
- **Durable fixtures over the plan's mktemp sketch.** Landing the fixtures in `construct/vocabulary/testdata/` (alongside the existing invalid-model fixtures, which the export already excludes) is a strict improvement on the plan's throwaway `mktemp` approach — the cases survive as regression tests.
- **The cue comment (`project.cue:86-91`) carries the full why** — pre-baseline-era honesty, the executing-path argument for why new projects still carry a baseline, and the `dropped` rationale — so the guard's shape is self-explaining at its single source.
- **Prose updated in the same commit, correctly scoped** (`construct/datatype/project.md:29-30`): "required for committed/executing/paused … Not required for `done` — a record archived from the pre-baseline era may lack it (#171)". No retired `active` forms introduced, so the `TestProjectProseCitesModel` drift binding stays green statically.
- **The side-quest fix (50c9e3f) is sound**: `TestProjectCloseRecordsFogAndArchives` now pins `projectTodayFn` instead of depending on the real clock — a genuine date-flakiness fix, properly labeled `side-quest:`.

## 2. Critical findings

None.

## 3. Important findings

- **`construct/generated/vocabulary/` is stale relative to the edited cue — run `make weave` before crossing the boundary** (repeat of #180 M1's Important finding; ARCH-DRY, the served face must derive). The `.source-sha` stamp is a sha256 over the raw source `.cue` text (`cmd/vocabulary/stamp.go:20-31`), so the M1 edit to `project.cue` invalidates it regardless of the export being content-identical; mtime ordering confirms the stamp predates the cue edit, so `vocabulary check --output construct/generated/vocabulary` will report STALE and `make check`-style gates go red. Fix: one `make weave` run. Root-cause note: the durable plan's M1 chunk regenerates `pkg/vocab/project.json` but omits the weave step entirely — despite the 2026-07-16 continuation's explicit lesson ("`make weave` belongs in this boundary's consumer sweep"). See plan-revision recommendation below.

## 4. Minor findings

- Durable-plan tracking lags: the plan doc's M1 checkbox row (`workshop/plans/000171-cross-repo-project-lift-plan.md:108`) and all Chunk 1 step checkboxes remain `- [ ]` while the issue's `## Plan` M1 row is ticked. Tick the executed steps (or note the plan doc tracks at milestone granularity) so the two surfaces agree.
- Plan Task 1.2 Step 1's stated expectation is impossible: `#Project` is a CUE `#`-definition and doesn't `cue export`, so regenerating `pkg/vocab/project.json` is a byte-identical no-op — there is no "guard-condition change" diff to see. Harmless, but a future reader will hunt for a missing diff.

## 5. Test coverage notes

- The two failure modes this diff could ship — under-relaxing (legacy `done` still rejected, forcing fabricated dates at M6) and over-relaxing (live records escaping the baseline requirement) — are both pinned in `vet_test.sh:45-50`. That is the right coverage for a schema-guard change.
- `dropped`-without-baseline has no fixture; it was never required and the risk is nil, so I don't ask for one.
- I could not execute `bash construct/vocabulary/vet_test.sh` or `go test ./pkg/vocab/...` (Bash tool broken in this review environment); semantics were verified statically against the cue and the fixtures. The implementor should ensure the milestone-close log records an actual green run.

## 6. Architectural notes for upcoming work

- **ARCH-DRY: pass**, one caveat. The guard changed at its single source and consumers derive. But the datatype prose's baseline-requirement row is a hand-maintained restatement the `TestProjectProseCitesModel` drift test does *not* bind (it binds statuses/sections/authority-citation only) — it was updated correctly this time, by discipline. If M5's prose work touches this file anyway, consider whether the requirement row should drop the restatement or the drift test should grow a clause.
- **ARCH-PURE: pass** (n/a for a schema change; the shell test is deterministic; the side-quest fix moved a test *off* the real clock — the right direction).
- **ARCH-PURPOSE: pass.** M1's purpose is to unblock M6's honest legacy conversion without fabricated dates; the relax plus negative control delivers exactly that. Shadow-sweep of guard consumers: cue validator (derives), `vet_test.sh` (updated), datatype prose (updated), `pkg/vocab` embedded JSON (guard never present, N/A), generated served face (export unchanged; freshness stamp is the Important finding). The full issue purpose (M2–M6) is properly sequenced in the plan, not deferred away.
- The staleness finding is now the **third** occurrence of "generated vocabulary face lags a cue edit" (#180 M1, the lessons entry, now here). The recurring fix is mechanical — consider wiring `vocabulary check` into a deterministic gate (`sdlc push`/`merge`, next to the instance-conformance gate) so this class stops depending on reviewer catch. Worth a small issue rather than an M1 scope creep.

## 7. Plan revision recommendations

Append one `## Revisions` entry to `workshop/plans/000171-cross-repo-project-lift-plan.md`:

- **2026-07-17 — M1 executed; two corrections.** (1) Task 1.2 Step 1's expected `pkg/vocab/project.json` diff does not exist: `#Project` is a `#`-definition and doesn't export, so regeneration is a verified no-op; the real regeneration deliverable at any cue-touching boundary is `make weave` (the `construct/generated/vocabulary/` served face + `.source-sha` stamp), which Chunk 1 omitted — add it as an explicit step (done for M1 as part of the FIX-THEN-SHIP fixes). (2) Fixtures landed as durable `construct/vocabulary/testdata/project_{done,executing}_no_baseline.json` with a negative control, superseding Step 1's mktemp sketch.
