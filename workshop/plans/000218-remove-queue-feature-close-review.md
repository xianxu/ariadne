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

---

## Re-review — 2026-09-09T13:04:13-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 218 — Remove the sdlc queue verb and workshop/queue.md |
| repo | ariadne |
| issue file | workshop/issues/000218-remove-queue-feature.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6828a2aa0d3d069347c55daa097d9c027d68af38..6c04fa4ab3ba638a2fe2a7a8f5f232b70ec451bf |
| command | sdlc close --issue 218 |
| reviewer | claude |
| timestamp | 2026-09-09T13:04:13-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The removal itself is complete, correct, and independently verified — I did not take the commit messages or the prior review sidecar at face value. `go build ./...` and `go vet ./cmd/sdlc/...` are clean; `go test ./cmd/sdlc/internal/... -count=1` is fully green; the full `go test ./cmd/sdlc/ -count=1` ran in 120s and failed on exactly one test, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (`fleet_plan_test.go:14`), the pre-existing ariadne#210 archived-plan path, precisely as the issue predicts. `sdlc --help` lists no `queue`, `datatype list` offers no `queue`, all six deleted paths are gone, and a fleet sweep across the 15 sibling ariadne-styled repos found no dangling `construct/datatype/queue.md`, `workshop/queue.md`, or `sdlc queue` reference, so the base-layer deletion propagates cleanly. The `TrunkFile` "behavior and coverage unchanged" claim holds under a real check — the `^func ` set in `trunkfile_test.go` is byte-identical before and after (31 functions, zero deleted) — and I confirmed all six re-anchor rows landed, including the `PreservesFileMode` comment that round 2 said was missing. What holds SHIP back is one thing, and it is the same thing as last round: the Done-when's identifier sweep, the issue's own headline close-gate evidence, still does not produce the output it claims. The `grep -rn` → `git grep` swap fixed the *tool* but not the *exclusion set*, and this very window commits three gate sidecars under `workshop/plans/000218-*` that the exclusions do not cover — measured at HEAD it returns **5 lines, not 1**. The property the clause asserts is nonetheless true: with an id-glob exclusion I measured exactly the one deliberate atlas line.

### 1. Strengths

- **`cmd/sdlc/repoguard.go:9-14`** — the "exactly the lifecycle verbs" claim is now substantively true. I checked all 8 `guardSpineRepo` call sites (`claim`, `start-plan`, `change-code`, `milestone-close`, `close`, `project close`, `merge`, `push`) against `processmanual.WorkflowVerbs()` and they match exactly; `TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs` runs all 8 subtests and passes. Pointing at `grep -rn 'guardSpineRepo('` as the authority rather than the comment is the right instinct.
- **Coverage-loss assertion survives scrutiny (`workshop/issues/000218-remove-queue-feature.md:143-148`).** A green suite genuinely cannot prove this, so I checked it directly: `repoguard_test.go:18-24` derives its verb list from `WorkflowVerbs()`, builds the real cobra tree, and the charter-message pin at `:54` *is* the guard-first ordering assertion — a `--issue is required` die would fail it. The surviving behavioral pin is stronger than the deleted source-inspecting one, which #209's own review had already shown to be mutation-blind. No gap.
- **ARCH-ORDER — pass, and a strength.** The queue's CAS/intent-replay state machine went out wholesale while `TrunkFile`'s ordering seam survived intact: `TestTrunkFile_RetryReRunsTransformOnMovedBase` and `TestTrunkFile_OneFetchPerAttempt` still push a peer commit from *inside* the transform, which is a real injected interleaving rather than a sample of size one.
- **ARCH-CONSTRAINTS — pass.** `workshop/issues/000218-remove-queue-feature.md:216-236` diagnoses the suite hang as `close_test.go:131` taking the production `.git/sdlc.lock` against a 30-minute `DefaultWaitTimeout`, rather than calling it slowness, and files it as ariadne#219 instead of absorbing it. Naming the envelope and the repro is exactly the at-review lens.
- **`gitx.FirstLine` correctly kept** (`trunkfile.go:227`, consumed at `issueids.go:113`, unit-tested at `trunkfile_test.go:449`) — the one piece of the #209 surface that was never queue-specific.

### 2. Critical findings

None.

### 3. Important findings

**BR-3 (not-addressed) — Done-when identifier sweep still inert, 4th round in family `verification-not-executable`.** `workshop/issues/000218-remove-queue-feature.md:178-187`. Per the escalation rule I am **not** proposing the instance fix as the deliverable — here is the rule.

The measured facts first. At HEAD, the clause as written returns 5 lines: the one intended `atlas/workflow/sdlc-binary.md:652`, plus four in `workshop/plans/000218-remove-queue-feature-close-review.md` (`:5`, `:23`, `:44`, `:69`). The exclusions stop at `':!workshop/issues/000218*'` and `':!workshop/issues/000219*'`, but this same window commits three gate sidecars under `workshop/plans/000218-*`. Family prevalence: round 1 (PQ-5) the command was alias-dependent; round 2 (BR-3) it shipped unfixed with that diagnosis recorded; round 3 the tool was corrected but the exclusion set was reasoned about rather than re-run.

**The rule that covers all four:** *a Done-when clause that records a command's expected output must have that command executed verbatim, at HEAD, in the same action that writes the clause — and any self-exclusion must be an artifact-ID glob, never a hand-listed directory path.* The id-glob half is not stylistic: an issue's own artifacts land in at least two directories (`workshop/issues/NNNNNN*` and `workshop/plans/NNNNNN*`), and gate sidecars materialize in the very window that records the clause, so a directory-anchored exclusion is stale the moment the next gate writes. Every round of this family has been the author *predicting* the output instead of *reading* it.

Applying the rule here yields `':!workshop/history' ':!*000218*' ':!*000219*'`, which I ran: exactly 1 line, the intended atlas note. Since no gate enforces this, it belongs in `workshop/lessons.md` — this is the not-code-enforced case where a lessons entry earns its place.

### 4. Minor findings

- **BR-7 (not-addressed, partial).** `workshop/issues/000218-remove-queue-feature.md:297` — the Plan checkbox still reads `- [x] Remove workshop/queue.md from origin/main.` BR-7 named both the Done-when clause and this checkbox; only the clause was rewritten. I re-fetched and confirmed `git ls-tree origin/main workshop/queue.md` still returns blob `4abe545`. This is the 2nd live instance in family `boundary-claim-premature`, so the deliverable is the rule, not the tick: *no artifact may state in present or perfect tense anything that becomes true only after a step outside this boundary — the merge, or a future issue's implementation; write the tense the boundary can verify and name the step that makes it true.* The enumeration for #218 is three sites: the Done-when clause (fixed round 3), this checkbox, and `atlas/workflow/sdlc-binary.md:649` below. Sweep all three, not the one named.
- **New — `atlas/workflow/sdlc-binary.md:649` asserts a consumer that does not exist yet** (family `boundary-claim-premature`, same rule as above). "Its consumer is `ariadne#207`" reads as present state, but I verified `gitx.TrunkFile` has **zero** production callers at HEAD — `git grep 'TrunkFile' -- '*.go'` outside `trunkfile.go` and `_test.go` is empty. The issue's Spec is honest about this ("zero consumers *today*"); the atlas, which is what a future reader actually greps, is not.
- **New — `cmd/sdlc/repoguard.go:23` "an env, not 7 new per-verb flags"** is now inconsistent with the 8 verbs listed at `:9-11` in the same comment block. This is the 2nd finding in family `prose-enumeration-drift`, so again the rule rather than the number: *a comment must not hand-restate a derived set or its cardinality; either point at the derivation or pin the prose with a drift test.* The `7` was already stale before #218 (`project close` gained the guard in `c5b2096`, #180 M4) — the round-2 fix corrected the list one sentence above and left the count below, which is the instance-not-class pattern (ARCH-PURPOSE). The honest resolution is to delete the enumeration and the count, keeping only the `grep -rn 'guardSpineRepo('` pointer the comment already added.
- **BR-5 (withdrawn).** I retract round 2's finding: "36" *does* match a countable set. On the pre-change file, `grep -c -i queue cmd/sdlc/internal/gitx/trunkfile_test.go` = **36** (lines containing a case-insensitive `queue`; 42 raw tokens, 34 `queue.md`). The Estimate's retained "36" at `:268` is accurate. What now needs correcting is the *remediation*: the table row at `:134` asserts "neither matched a countable set," which is false for 36 and was written on my predecessor's mistaken measurement.

### 5. Test coverage notes

`trunkfile_test.go`'s function set is unchanged (31, identical), so the "behavior and coverage unchanged" contract holds under a real check rather than a green run. The `queue.md` → `note.md` fixture rename is uniform — no residual `queue` token anywhere in `cmd/sdlc/internal/gitx/*.go` — and `go vet ./cmd/sdlc/...` typechecks test files, so a dangling `NewQueueCmd`/`internal/queue` reference in any `_test.go` would have surfaced. The one coverage question a deletion raises (does anything still pin guard-first ordering?) is answered by `repoguard_test.go`, which I ran and inspected rather than trusting. No new tests are warranted: the remaining findings are all artifact-prose defects, and the rule they need is a lessons entry, not a test.

### 6. Architectural notes

- **ARCH-DRY — pass**, with the `prose-enumeration-drift` finding above as the one hand-maintained restatement left standing.
- **ARCH-PURE — pass.** The pure core (`internal/queue`: `Line`/`Doc`/`Intent.Apply` + fuzz corpus) and its IO shell (`cmd/sdlc/queue.go`) were removed together, leaving no orphaned half. `TrunkFile`'s `Update(path, msg, transform)` seam is untouched.
- **ARCH-PURPOSE — flag**, on the sweep clause only: the round-3 response fixed the tool the finding named while an enumerable sibling (the `workshop/plans/` sidecars) sat in the same commit. The *removal's* own shadow-sweep passes — I enumerated every consumer (verb registration, helptext, datatype prototype, atlas verb table, atlas section, atlas index, repoguard comment, lessons citation, the open ariadne#207 at `:178-180`) and every one derives or is updated; the ~20 surviving `queue` tokens in the tree are all generic senses (BFS worklists in `projectstatus.go:199`, `layergraph/walk.go`, ARCH-ORDER's own "cancel / queue / preempt / ignore" at `judge/architecture.md:170`, the `onCapacity:"queue"` fleet-policy enum).
- **ARCH-MOCK — pass.** `trunkFixture` still stands up a real throwaway repo against a real bare origin; the ARCH-MOCK rationale comment survived the rename. The `fakeTrunk` double that `lessons.md` indicts went out with the verb, and the lessons entry now says so rather than leaving a reader hunting.
- **ARCH-CONSTRAINTS — pass** (see Strengths).
- **ARCH-SECURE — pass.** The change strictly *reduces* write surface: a verb that committed and pushed to `origin/main` from any checkout is gone, along with its spine-guarded write subcommands. No credentials, no new untrusted-input parse. The hand-editable-file fuzz guard went out with the file it guarded, which is correct.
- **ARCH-ORDER — pass** (see Strengths).

For upcoming work: ariadne#207 now carries the whole justification for ~660 lines of `gitx.TrunkFile` with no production caller. If #207 slips, that dead code needs a decision, not a default.

### 7. Plan revision recommendations

The issue needs one `## Revisions` entry covering four edits, timestamped with the reason "close review round 3 — evidence clauses corrected":

1. `:186` — replace the sweep exclusions with `':!workshop/history' ':!*000218*' ':!*000219*'` (verified: returns exactly the one atlas line), and state the id-glob rule as the reason, not the paths.
2. `:297` — reword the Plan checkbox to match the Done-when's corrected tense: "Delete `workshop/queue.md` on the branch; it leaves `origin/main` at merge."
3. `:134` — drop "and neither matched a countable set." 36 is `grep -c -i queue` on the pre-change file; the Estimate's use of it at `:268` is correct and should stay.
4. A `## Log` line recording that `atlas/workflow/sdlc-binary.md:649` and `cmd/sdlc/repoguard.go:23` were swept as the remaining members of their families, not as one-off fixes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Clause now uses git grep with principled exclusions and states the expected residue plus the raw-grep numbers.
  - id: BR-2
    disposition: addressed
    note: |
      Table demoted to non-authoritative and PreservesFileMode row added; I verified all six rows against the diff.
  - id: BR-3
    disposition: not-addressed
    note: |
      Tool fixed (grep -rn to git grep) but not the exclusion set; measured 5 lines at HEAD, not 1.
  - id: BR-4
    disposition: addressed
    note: |
      The PreservesFileMode row covers it; the code re-anchor is present at trunkfile_test.go:718-721.
  - id: BR-5
    disposition: withdrawn
    note: |
      Retracted — grep -c -i queue on the pre-change trunkfile_test.go is exactly 36, so the count was checkable.
  - id: BR-6
    disposition: addressed
    note: |
      All 8 guardSpineRepo call sites match WorkflowVerbs exactly; all 8 subtests run and pass.
  - id: BR-7
    disposition: not-addressed
    note: |
      Done-when clause rewritten, but the Plan checkbox at :297 that BR-7 also named still claims origin/main.
findings:
  - id: new
    severity: Minor
    family: boundary-claim-premature
    title: |
      atlas says "Its consumer is ariadne#207" while TrunkFile has zero production callers at HEAD
    detail: |
      atlas/workflow/sdlc-binary.md:649. Verified: git grep 'TrunkFile' over non-test .go files
      outside internal/gitx/trunkfile.go is empty — ~660 lines of production code with no caller.
      2nd live instance in this family (with BR-7's Plan checkbox), so the deliverable is the rule:
      no artifact may state in present or perfect tense what becomes true only after a step outside
      this boundary (the merge, or a future issue's implementation); write the verifiable tense and
      name the step. Measured enumeration for 218 is three sites — the Done-when clause (fixed round
      3), the Plan checkbox at :297, and this atlas line. Sweep all three.
  - id: new
    severity: Minor
    family: prose-enumeration-drift
    title: |
      repoguard.go:23 still says "not 7 new per-verb flags" while :9-11 now lists 8 guarded verbs
    detail: |
      The round-2 fix corrected the verb list one sentence above and left the cardinality below, in
      the same comment block, in the commit whose stated purpose was making that block read true.
      The 7 was already stale before 218 — project close gained the guard in c5b2096 (#180 M4). 2nd
      in family, so the rule rather than the number: a comment must not hand-restate a derived set
      or its cardinality; either point at the derivation or pin the prose with a drift test. The
      comment already added the grep -rn 'guardSpineRepo(' pointer — delete the list and the count
      and keep only that.
```
