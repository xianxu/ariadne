# Boundary Review — ariadne#218 (whole-issue close)

| field | value |
|-------|-------|
| issue | 218 — Remove the sdlc queue verb and workshop/queue.md |
| repo | ariadne |
| issue file | workshop/issues/000218-remove-queue-feature.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6828a2aa0d3d069347c55daa097d9c027d68af38..b5be04de60c2445bed6a79402334178e94234aea |
| command | sdlc close --issue 218 |
| reviewer | claude |
| timestamp | 2026-09-09T12:52:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The removal itself is complete and correct — I verified it independently rather than from the commit message: `go build ./...` and `go vet ./cmd/sdlc/...` are clean, `go test ./cmd/sdlc/internal/... -count=1` is fully green, and `go test ./cmd/sdlc/ -count=1` (155s, plain shell, lock free) fails on exactly one test — `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` at `fleet_plan_test.go:14`, the pre-existing ariadne#210 archived-plan path, precisely as the issue predicts. The `TrunkFile` "behavior and coverage unchanged" claim holds under a real check (the `^func ` set in `trunkfile_test.go` is identical before and after — 27 tests, zero deleted), the coverage-loss assertion about the deleted guard tests checks out (`TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs` runs 8 subtests off `processmanual.WorkflowVerbs()` and pins guard-first ordering), and a fleet-wide sweep found no peer repo carrying a `construct/datatype/queue.md` symlink, a `workshop/queue.md`, or an `sdlc queue` reference — so the base-layer deletion propagates without dangling links. What holds SHIP back is not the code: the Done-when's identifier sweep — the issue's own headline close-gate evidence — is **inert as written** and returns 99 lines rather than the claimed one, because `grep -rn … .` emits `./`-prefixed paths while all three `grep -v` filters anchor at `^workshop/`/`^construct/`. This is the plan gate's own PQ-5, disposed `not-addressed` in round 4 with that exact diagnosis, shipped unfixed. The substance is right (the corrected sweep does return exactly the one deliberate atlas line, which I ran), so the fix is a three-character edit to the recorded command.

## 1. Strengths

- **`cmd/sdlc/internal/gitx/trunkfile_test.go`** — the fixture rename is a genuine no-op on behavior. Diffing `^func ` lines across the window is empty: the same 27 test functions, only the `queue.md` → `note.md` string changed. That is what makes "behavior and coverage unchanged" a verified claim rather than a hopeful one, and it is the right way to do a rename inside a removal.
- **`cmd/sdlc/repoguard.go:9-11`** — restoring "exactly the lifecycle verbs" is substantively true now. I checked all 8 `guardSpineRepo` call sites against `WorkflowVerbs()` (SpineVerbs from `GateCatalog` + claim/start-plan) and they match exactly; `migrate`'s deliberate absence is still documented at `migrate.go:460`.
- **`workshop/lessons.md:36-38`** — exactly the right disposition for a lesson whose example was deleted: keep the general pattern, tell the reader the cited code is gone, and say *why* the citation names the shape rather than the file. A reader will not hunt for `fakeTrunk`.
- **`atlas/workflow/sdlc-binary.md:646-657`** — the re-anchored TrunkFile section explains why the primitive outlived its first consumer (shared namespace with concurrent writers vs. an attention list with none) instead of just deleting the sentence. That is the one legitimate residue the sweep is designed to expect.
- **`workshop/issues/000207-sync-without-worktree.md:177-180`** — amending an *open* peer issue that cited the deleted consumer, in the same round, is the ARCH-PURPOSE class-not-instance discipline actually applied.

## 2. Critical findings

None.

## 3. Important findings

**I-1 — `workshop/issues/000218-remove-queue-feature.md:169-183`: the Done-when sweep command cannot produce the result the clause claims.**

Measured, running the command byte-for-byte as recorded:

```
$ grep -rn "sdlc queue\|internal/queue\|NewQueueCmd\|datatype/queue\|helptext/queue" \
    --include='*.go' --include='*.md' . \
  | grep -v "^workshop/history/" | grep -v "^workshop/issues/00021[89]" \
  | grep -v "^construct/generated/" | wc -l
99                     # 86 of them under ./workshop/history/
```

`grep -rn … .` prefixes every path with `./`, so none of the three `^`-anchored filters ever matches. The clause asserts "exactly one line". Fix: anchor the filters at `^\./` (or replace the `.` target with explicit directories). With that one change the command returns exactly the one line the clause names — `atlas/workflow/sdlc-binary.md:652` — so only the recorded artifact is wrong, not the underlying removal. This is `verification-not-executable`'s **3rd** instance in this issue (PQ-1, PQ-5, now this), and PQ-5's round-4 note already stated the diagnosis verbatim; shipping the clause unfixed means the close gate's `--verified` evidence points at a command a future reader will re-run and see fail.

## 4. Minor findings

- **`workshop/issues/000218-remove-queue-feature.md:130`** — the re-anchor table lists `trunkfile_test.go:15,:83` but the diff correctly re-anchored a **third** comment at `trunkfile_test.go:719` (`TestTrunkFile_PreservesFileMode`, the same "fine for a queue" sentence as `trunkfile.go:452`). The code did more than the table; PQ-6 asked for exactly this row and it is still absent. Table needs the row, not the code a change.
- **`workshop/issues/000218-remove-queue-feature.md:129` and the Estimate rationale** — "(36 sites)" matches no countable set. The pre-change file has **34** `queue.md` occurrences on 34 lines, and 42 total case-insensitive `queue` tokens. This is `unbacked-count-claim` recurring after PQ-3 corrected "37 tests" and the estimate note corrected "~20" → "36"; the estimate derivation leans on the wrong number.
- **`cmd/sdlc/repoguard.go:10-11`** — the restored parenthetical enumerates seven verbs (`claim, start-plan, change-code, milestone-close, close, merge, push`) as if it were `WorkflowVerbs`, but `project close` is in that set (`internal/processmanual/gatesig.go:112`), is guarded (`projectclose.go:36`), and the drift test enumerates it — I confirmed the 8th subtest `project_close` runs and passes. Since making this sentence read true is the issue's stated purpose, add `project close` to the list.
- **`workshop/issues/000218-remove-queue-feature.md:154` / `:281`** — "`workshop/queue.md` is gone from `origin/main`" is ticked, but it is still present on `origin/main` (`a722b52`, confirmed against `git ls-remote`). The deletion is committed on the branch and the branch is a clean descendant of main with no competing commits touching that path, so the merge will remove it — but the clause is not satisfiable at *this* boundary. Re-word to "deleted on the branch; verify post-merge."

## 5. Test coverage notes

The deleted-coverage question is the right substitute for a test-strategy line on a removal, and the issue's answer survives scrutiny. `TestQueueCmd_WriteVerbsAreSpineGuarded` / `subcommandGuardSource` were the tree's only source-inspecting guard-ordering tests, but the surviving behavioral pin is stronger than what was lost: `repoguard_test.go:39-57` constructs the real cobra tree per verb, asserts the refusal message carries the charter, and that message assertion *is* the guard-first ordering pin (a `--issue is required` die would fail it). It enumerates from `WorkflowVerbs()`, so a new lifecycle verb demands the guard automatically. No gap.

`gitx` remains covered against a real bare origin (ARCH-MOCK), and the removal of `FuzzDocRoundTrip`'s corpus is correct — the hand-editable parser it fuzzed is gone, so there is no untrusted-input surface left to guard.

## 6. Architectural notes

- **ARCH-DRY — pass.** No duplication introduced; `gitx.FirstLine` survives with its single definition and its `issueids.go:113` delegating consumer, exactly as the Spec's Keep list requires.
- **ARCH-PURE — pass.** The pure core (`internal/queue`) and its IO shell (`cmd/sdlc/queue.go`) were removed together, leaving no orphaned half. `TrunkFile`'s transform seam is untouched.
- **ARCH-PURPOSE — flag (I-1, and the two table Minors).** The shadow-sweep of the "queue" concept genuinely comes out clean when run correctly: identifiers, prose in `lessons.md`, the open ariadne#207, atlas, `datatype list`, and every peer repo. But the *rule* the gate extracted twice — "the enumeration must be derived from a runnable sweep whose entire residue is dispositioned" — was adopted as prose and not as a working command. The class was named; the enumeration that would enforce it was not written.
- **ARCH-MOCK — pass.** Real throwaway git repos with a real bare origin, unchanged.
- **ARCH-CONSTRAINTS — pass, with a carry-forward.** The removal strictly reduces runtime. Worth carrying to #207: `TrunkFile`'s `fetch`/`push` still take no `context.WithTimeout` (worst case 3 fetches + 3 pushes unbounded), a residue #209's M1 review accepted only because a library was not an interactive path. Deleting the queue verb removed the interactive consumer, so #207 is now where that decision lands.
- **ARCH-SECURE — pass.** Deletion only; no new input parsing, no credential surface. `path`/`msg` still reach git as argv elements.
- **ARCH-ORDER — pass.** The CAS/retry interleaving tests are byte-identical apart from the fixture name, so the seam a test uses to choose the interleaving (`TestTrunkFile_RetryReRunsTransformOnMovedBase`, `…OneFetchPerAttempt`) is intact — this diff removes no ordering oracle.

## 7. Plan revision recommendations

Add a `## Revisions` entry to `workshop/issues/000218-remove-queue-feature.md` covering:

1. **Sweep command corrected** (I-1) — filters re-anchored to `^\./…`; record the measured before/after (99 lines → 1) so the next reader sees the clause was executed, not asserted.
2. **Re-anchor table completed** — add the `gitx/trunkfile_test.go:719` row (`TestTrunkFile_PreservesFileMode` comment), which the code re-anchored and the table omits; note that PQ-6's rule (table derived from the sweep) is what the row satisfies.
3. **Fixture-count corrected** — `(36 sites)` → 34 `queue.md` occurrences (42 total `queue` tokens); note in the Estimate that the `cross-cutting-refactor` rationale cited the wrong figure.
4. **`origin/main` clause re-scoped** — `workshop/queue.md` is deleted on the branch and lands on `origin/main` at merge; state the post-merge check rather than ticking it at close.

```findings
findings:
  - id: new
    severity: Important
    family: verification-not-executable
    title: |
      Done-when identifier sweep is inert — returns 99 lines, not the claimed one
    detail: |
      workshop/issues/000218-remove-queue-feature.md:169-183. `grep -rn ... .` emits
      ./-prefixed paths, so all three `grep -v "^workshop/..."` / `"^construct/..."`
      filters never match. Measured 99 lines (86 under ./workshop/history/) against a
      clause asserting "exactly one line". Anchoring the filters at `^\./` makes the
      command return exactly the one deliberate atlas line it names, so the removal is
      sound and only the recorded evidence is wrong. 3rd in family; PQ-5 round 4 stated
      this diagnosis verbatim and it shipped unfixed.
  - id: new
    severity: Minor
    family: stale-anchor-sweep
    title: |
      Re-anchor table omits trunkfile_test.go:719, which the code did re-anchor
    detail: |
      The Spec table at line 130 lists only `:15,:83`, but the diff correctly
      re-anchored the TestTrunkFile_PreservesFileMode comment at trunkfile_test.go:719
      (the same "fine for a queue" sentence as trunkfile.go:452). PQ-6 asked for this
      row. Code is right; the table under-claims and needs the row.
  - id: new
    severity: Minor
    family: unbacked-count-claim
    title: |
      "36 sites" matches no countable set in trunkfile_test.go
    detail: |
      workshop/issues/000218-remove-queue-feature.md:129 and the Estimate's
      cross-cutting-refactor rationale both cite 36. Measured on the pre-change file:
      34 `queue.md` occurrences on 34 lines; 42 total case-insensitive `queue` tokens.
      Recurrence of PQ-3's family after "37 tests" was dropped and "~20" was corrected
      to "36".
  - id: new
    severity: Minor
    family: prose-enumeration-drift
    title: |
      repoguard.go's restored "exactly the lifecycle verbs" list omits `project close`
    detail: |
      cmd/sdlc/repoguard.go:10-11 spells WorkflowVerbs as seven verbs, but
      `project close` is in the set (internal/processmanual/gatesig.go:112), calls
      guardSpineRepo (projectclose.go:36), and the drift test enumerates it — I ran
      TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs and the `project_close` subtest
      runs and passes. The set-level claim "exactly" is true; the inline list of
      members is incomplete, in the very sentence this issue set out to make read true.
  - id: new
    severity: Minor
    family: boundary-claim-premature
    title: |
      "workshop/queue.md is gone from origin/main" is ticked but only true post-merge
    detail: |
      Line 154 and plan item 281. The file is still on origin/main (a722b52, confirmed
      against `git ls-remote`); the deletion is committed on the branch, which is a
      clean descendant of main with no competing commits on that path, so the merge
      will remove it. The clause is not satisfiable at this boundary — re-word to
      "deleted on the branch; verify post-merge."
```
