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

---

## Re-review — 2026-09-09T13:13:24-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 218 — Remove the sdlc queue verb and workshop/queue.md |
| repo | ariadne |
| issue file | workshop/issues/000218-remove-queue-feature.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6828a2aa0d3d069347c55daa097d9c027d68af38..a9871748bbc951f002c5224c960df5835c3993ec |
| command | sdlc close --issue 218 |
| reviewer | claude |
| timestamp | 2026-09-09T13:13:24-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The removal itself is complete and I verified it independently rather than from the commit messages: `go build ./...` clean, `go test ./cmd/sdlc/internal/... -count=1` fully green (gitx 16.2s), `sdlc --help` lists no `queue`, `datatype list` offers no `queue`, `construct/generated/datatype/SKILL.md` was regenerated (zero `queue` hits, mtime after the deletion) with the working tree clean, and no peer repo under `/Users/xianxu/workspace/` carries a dangling `construct/datatype/queue.md` or `workshop/queue.md`, so the base-layer deletion propagates without stranding a symlink. The "behavior and coverage unchanged" claim on `TrunkFile` holds under a real check — the `^func ` set in `trunkfile_test.go` is byte-identical across the window (31 funcs, 27 tests), only the fixture string changed — and the "coverage loss: none" assertion checks out: `project close` **is** in `GateCatalog` → `SpineVerbs` → `WorkflowVerbs`, so all 8 `guardSpineRepo` call sites are enumerated by `TestGuardSpineRepo_BrainRefusesAllLifecycleVerbs`, and the deleted `subcommandGuardSource` was hardcoded to read `queue.go` with queue-only markers. BR-3 is genuinely fixed — I ran the documented `git grep` verbatim and got exactly one line. What holds SHIP back is not code: **three of the four open findings were not acted on at all.** BR-7's second site and BR-8's site are untouched, BR-9's fix removed the list but left the count nine lines below in the same comment block — which now explicitly claims "No list or count appears here on purpose." And a derived sweep for BR-8's own class finds five live sites where the hand-typed enumeration named three, two of which this window *introduced* by regressing correct future tense to false present tense.

## 1. Strengths

- **`cmd/sdlc/internal/gitx/trunkfile_test.go`** — the rename is a verified no-op on behavior. I diffed the `^func ` set across `6828a2a..a987174`: identical, 31 functions. `queue.md` → `note.md` and `queue: …` → `trunk: …` throughout, including the peer-push helper (`:360`) and the sibling-survival assertion (`:315`). That is the right way to do a rename inside a removal.
- **The `git grep` conversion of the Done-when sweep** (`:184-197`) is a real fix, not a re-word. Ran verbatim under bash: one line, `atlas/workflow/sdlc-binary.md:652`, exactly as claimed. Scoping the exclusions by issue **id** rather than path (`':!*000218*'`) is the correct generalization — I confirmed it tolerates `workshop/plans/000218-*-close-review.md`, which the close itself generates and which therefore cannot exist when the clause is tested beforehand.
- **`cmd/sdlc/repoguard.go:9-15`** — replacing the verb list with `grep -rn "guardSpineRepo(" cmd/sdlc/*.go` as the authority plus the `WorkflowVerbs` derivation is the right structural fix (ARCH-DRY). I verified the derivation is sound: 8 call sites, 8 subtests, exact match.
- **`workshop/lessons.md:36-38`** — correct disposition for a lesson whose example was deleted: keep the pattern, tell the reader the cited code is gone, say why the citation names the shape.
- **`workshop/issues/000207-sync-without-worktree.md:177-180`** — amending an *open* peer issue that cited the deleted consumer, in the same round, is ARCH-PURPOSE class-not-instance discipline actually applied (see M-1 for a formatting artifact in the edit).

## 2. Critical findings

None. No correctness bug, no behavior drift, no silent error swallowing in the diff.

## 3. Important findings

**I-1 — the `boundary-claim-premature` enumeration is hand-typed and now measurably incomplete; two of its named sites were never fixed.** BR-8 stated the rule correctly and named three sites. Measured at HEAD: the Done-when clause (`:159-163`) is fixed; `workshop/issues/000218-remove-queue-feature.md:304` (`- [x] Remove workshop/queue.md from origin/main.`) and `atlas/workflow/sdlc-binary.md:649` ("Its consumer is `ariadne#207`") are untouched. Two further sites were **introduced by this window**, regressing tense in the direction the rule forbids — `cmd/sdlc/internal/gitx/trunkfile.go:452` and `trunkfile_test.go:721` both changed `"that ariadne#207 **will point** at arbitrary paths"` → `"ariadne#207 **points** this at arbitrary repo paths"`. A fifth is `workshop/issues/000207-sync-without-worktree.md:178` ("this issue is now its only one"). Ground truth: `git grep 'TrunkFile' -- '*.go' | grep -v _test.go` is empty outside `trunkfile.go` — zero production callers. See the "Plan revision recommendations" section for the runnable sweep this should be derived from instead.

**I-2 — no `workshop/lessons.md` rule for the mistake that cost three of the four rounds** (AGENTS.md §4). The lesson is not code-enforced and no gate catches it: a verification recorded in an artifact ran green for the author only because `grep` was aliased to `ugrep` in that shell, and returned 99–103 lines for anyone else. `lessons.md` gained only a parenthetical about deleted code in this window; I checked for `alias`/`ugrep`/`git grep`/`runnable command` and there is no entry covering it. The existing "what a guard COMPUTES vs what it takes on faith" entry (`:65`) covers derived-set restatement but not shell-dependence. **This is the 3rd finding in family `verification-not-executable`** — earlier rounds fixed instances (BR-1, BR-3), so the deliverable here is the rule, and lessons.md is exactly where a cross-issue non-code-enforced rule belongs. Suggested: *a verification clause must be a command whose output is identical under any operator's shell — no aliased binary, no shell-local function; prefer `git grep` over `grep -rn` for repo sweeps (deterministic relative paths, gitignore-aware), and record the measured output alongside the command.*

## 4. Minor findings

- **M-1 — `workshop/issues/000207-sync-without-worktree.md:178-181`:** the inserted italic parenthetical captured the sentence that followed the insertion point. `…the choice is the caller's.)* So the Done-when clause *"…"* needs amending:` — "So the Done-when clause…" now reads as part of the aside instead of continuing the paragraph above it. Split the line after `caller's.)*`.
- **M-2** — `atlas/workflow/sdlc-binary.md:635` still tags `TrunkFile` as `#209` in the file-tree listing, and `:646` heads the section `(gitx.TrunkFile, #209)`. Defensible as provenance rather than a consumer claim, so not folded into I-1 — but if the section is edited for I-1, consider whether `#209` should read `#209, removed #218`.

## 5. Test coverage notes

- The kind of bug this diff could ship — a dangling reference or a lost test — is covered by evidence I re-derived, not asserted: identical `^func` set in `trunkfile_test.go`, 8/8 guard subtests off `WorkflowVerbs()`, clean build with the registration removed, `datatype list` reflecting the deleted prototype.
- No new tests are warranted. The one gap is unpinnable by a test and correctly handled as prose: `lessons.md:36-38` tells the reader the cited fixtures are gone.
- Non-`.go`/`.md` surfaces were outside the documented sweep. I checked them separately (`*.sh`, `*.cue`, `*.tmpl`, `*.json`, `Makefile*`, `construct/base.manifest`): every hit is a generic BFS-queue variable in `bootstrap.sh` / `list-peers.sh` / `layergraph/walk.go` / `projectstatus.go`, plus one `fleet_policy_invalid_action.json` fixture using `"queue"` as an invalid `onCapacity` action. Clean.
- `construct/base.manifest` correctly does not enumerate datatype prototypes (`:174-179` — per-layer ownership, DAG-merged union), so deleting `construct/datatype/queue.md` needs no manifest edit.

## 6. Architectural notes

- **ARCH-DRY — flag.** `cmd/sdlc/repoguard.go:24` still reads `(an env, not 7 new per-verb flags)` while the guarded set is 8. This is BR-9's exact site, disposed `not-addressed` below. The regression relative to round 3 is that `:12-15` now asserts *"No list or count appears here on purpose"* — a self-refuting comment is worse than a merely stale one, because it invites a reader to trust the block.
- **ARCH-PURE — pass.** `internal/queue` (pure core) and `cmd/sdlc/queue.go` (IO shell) were removed together; nothing was left half-wired. `gitx.FirstLine` correctly survives with its live consumer (`issueids.go:113`, called at `:131` and `:372`) and its pure unit test.
- **ARCH-PURPOSE — flag.** The shadow-sweep on the removal itself passes: every consumer (verb, package, helptext, datatype, seeded file, `main.go` registration, both atlas sites, `lessons.md`, open #207) derives from the deletion, with one deliberate documented residue. The flag is on the *finding-answering* axis: BR-8 named a class and an enumeration, and this round delivered neither the enumeration's remaining members nor a derivation for it (I-1).
- **ARCH-MOCK — pass, net improvement.** `trunkfile_test.go` runs against a real throwaway repo + real bare origin through `internal/testfix` (`:16-18`); `runGitIn` is the injected seam for failure injection. Deleting `fakeTrunk` — a stateless double that `lessons.md:5-25` was written about — removes the codebase's worst instance of this anti-pattern.
- **ARCH-CONSTRAINTS — pass.** The `go test ./cmd/sdlc/` envelope violation (`close_test.go:131` takes the production `.git/sdlc.lock`, so an in-transaction run waits `DefaultWaitTimeout` = 30min) is diagnosed rather than hand-waved, filed as ariadne#219 (issue file confirmed present, `status: open`), and the issue explicitly declines to carry a full-suite run as gate evidence with a stated substitute. Correct handling of a pre-existing constraint.
- **ARCH-SECURE — pass.** `internal/queue`'s fuzz-guarded round-trip parser existed because the queue file was hand-editable input the process did not produce; removing the consumer removes the untrusted surface with it, and the fuzz corpus went with it rather than being orphaned. No credentials touched; no test reaches real user state (`trunkFixture` is `t.TempDir()`-scoped).
- **ARCH-ORDER — pass.** The interleaving seam is intact and injectable: `pushPeerLine` fires *inside* the transform (`trunkfile_test.go:183`, `:594`) so a moved base is deterministically reproducible, `RetryExhaustionSurfacesGitRejection` pins the 3-attempt bound, `OneFetchPerAttempt` pins fetch cardinality, and `runGitIn` injection pins the non-retryable fast-fail. These are not sample-of-one tests.

## 7. Plan revision recommendations

The Spec's re-anchor enumeration (`:130-137`) already carries the right disclaimer ("hand-typed and therefore not authoritative"), and this round proves the disclaimer was accurate rather than defensive. Add a `## Revisions` entry that replaces the hand-typed premature-claim list with the derivation — this sweep is measured, not proposed:

```
git grep -n 'ariadne#207' -- '*.go' '*.md' ':!workshop/history' ':!workshop/plans/000218*'
grep -n 'origin/main' workshop/issues/000218-remove-queue-feature.md
```

Residue at HEAD, fully dispositioned — sweep 1 returns 5 non-#218 lines: `trunkfile.go:22` ("what **lets** ariadne#207 consume this" — capability, KEEP), `trunkfile.go:325` ("the same bound ariadne#207 **specs**" — the spec exists, KEEP), `trunkfile.go:452` (FIX), `trunkfile_test.go:721` (FIX), `atlas/workflow/sdlc-binary.md:649` (FIX). Plus `workshop/issues/000207-sync-without-worktree.md:178` (FIX), which sweep 1 excludes and which a second pass over `000207*` catches. Sweep 2 returns 3 lines: `:71` (Spec, declarative of what to delete — KEEP), `:161` (corrected Done-when — KEEP), `:304` (FIX). The Revisions entry should record that the class measured **6** sites where round 3 named 3, and that two of the six were created by this window's own re-anchoring commit.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Ran the documented `git grep` verbatim under bash: exactly one line, atlas/workflow/sdlc-binary.md:652.
  - id: BR-7
    disposition: not-addressed
    note: |
      Done-when :159-163 fixed; the Plan checkbox at :304 still claims removal from origin/main.
  - id: BR-8
    disposition: not-addressed
    note: |
      atlas/workflow/sdlc-binary.md:649 unchanged; TrunkFile still has zero production callers at HEAD.
  - id: BR-9
    disposition: not-addressed
    note: |
      List removed from :9-15, but :24 still reads "not 7 new per-verb flags" against 8 guarded verbs.
findings:
  - id: new
    severity: Important
    family: boundary-claim-premature
    title: |
      The premature-claim enumeration is hand-typed; 5 live sites measured against the 3 it names, 2 of them introduced by this window
    detail: |
      This is the 5th and 6th instance in family `boundary-claim-premature` (measured
      prevalence on ariadne#218: 6 sites, 1 fixed, 5 live). The RULE was already stated
      correctly by BR-8 and is not the problem; the ENUMERATION is, because it is
      hand-typed — the same defect the Spec already confesses for the re-anchor table at
      :124. Do not fix the five sites one by one. Derive them:
      `git grep -n 'ariadne#207' -- '*.go' '*.md' ':!workshop/history' ':!workshop/plans/000218*'`
      plus a pass over `workshop/issues/000207*` and
      `grep -n 'origin/main' workshop/issues/000218-remove-queue-feature.md`, then disposition
      the entire residue in a `## Revisions` entry. Live sites: trunkfile.go:452 and
      trunkfile_test.go:721 (both REGRESSED in commit b5be04d from correct future tense
      "ariadne#207 will point at" to false present tense "ariadne#207 points this at");
      atlas/workflow/sdlc-binary.md:649; workshop/issues/000218-remove-queue-feature.md:304;
      workshop/issues/000207-sync-without-worktree.md:178. Ground truth:
      `git grep TrunkFile -- '*.go' | grep -v _test.go` is empty outside trunkfile.go itself.
  - id: new
    severity: Important
    family: verification-not-executable
    title: |
      No lessons.md rule for the shell-dependent verification that cost three of four rounds
    detail: |
      This is the 3rd finding in family `verification-not-executable`; BR-1 and BR-3 fixed
      instances, so the deliverable here is the rule and its durable home. AGENTS.md
      section 4 requires it and no gate enforces it. I checked lessons.md for
      alias/ugrep/"git grep"/"runnable command" — nothing covers shell-dependence; the
      nearest entry (:65, "what a guard COMPUTES vs takes on faith") covers derived-set
      restatement instead. Proposed rule: a verification clause must be a command whose
      output is identical under any operator's shell — no aliased binary, no shell-local
      function. Prefer `git grep` over `grep -rn` for repo sweeps (deterministic
      repo-relative paths, gitignore-aware), and record the measured output next to the
      command so a reader can tell a passing check from an inert one.
  - id: new
    severity: Minor
    family: prose-edit-orphans-sentence
    title: |
      The ariadne#207 amendment captured the following sentence into its parenthetical
    detail: |
      workshop/issues/000207-sync-without-worktree.md:178-181. The inserted italic aside
      ends `...the choice is the caller's.)* So the Done-when clause *"..."* needs amending:`
      on one line, so the sentence that continued the paragraph above now reads as part of
      the aside. Split after `caller's.)*`. Introduced by b5be04d.
```

---

## Re-review — 2026-09-09T13:22:15-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 218 — Remove the sdlc queue verb and workshop/queue.md |
| repo | ariadne |
| issue file | workshop/issues/000218-remove-queue-feature.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6828a2aa0d3d069347c55daa097d9c027d68af38..11749d5d36061d348750f4fef7363f1440e98d74 |
| command | sdlc close --issue 218 |
| reviewer | claude |
| timestamp | 2026-09-09T13:22:15-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The removal itself is complete and verifiable: `go build ./...` clean, `go test ./cmd/sdlc/internal/... -count=1` green, `sdlc --help` and `datatype list` both free of `queue`, the documented identifier sweep run verbatim returns exactly the one deliberate atlas line, `git grep -i queue -- cmd/sdlc/internal/gitx/` is empty, and the working tree is clean (so the datatype regeneration touched nothing tracked). The Spec's two load-bearing assertions both check out independently: `TrunkFile` has zero production callers at HEAD (matching the corrected atlas prose), and the deleted `subcommandGuardSource` guard-ordering test covered queue subcommands *only* — guard-first placement for the eight real lifecycle verbs is behaviorally pinned by `repoguard_test.go:39` off `processmanual.WorkflowVerbs()`. What blocks SHIP is entirely prose residue from the prior round: three of six prior findings are still live, including one (BR-9) whose fix wrote the governing rule into the very comment block that still violates it four lines down, and one (BR-10) whose adopted derivation has a blind spot that drops a site the hand list had caught. All three are single-line edits.

### 1. Strengths

- **`cmd/sdlc/repoguard.go:9-15`** — the guard comment now points at `grep -rn "guardSpineRepo(" cmd/sdlc/*.go` plus `processmanual.WorkflowVerbs` as the derivation, and states *why* no list appears, naming the c5b2096 drift. That is the correct shape of fix for the family.
- **`workshop/lessons.md:5-45`** — the "Don't write a description where the referent is available" synthesis is the real root cause, and the table mapping eight apparently-unrelated families onto one defect is the most useful artifact this issue produced. The `"I verified it" is a property of your context` corollary discharges BR-11 exactly.
- **The `## Done when` sweep clause (`:183-197`)** is now genuinely executable and its stated residue is exact — I ran it verbatim in a plain shell and got one line, the one it names.
- **`cmd/sdlc/internal/gitx/trunkfile_test.go`** — the `queue.md` → `note.md` / `queue:` → `trunk:` rename is complete and behavior-neutral: the diff removes no test function, touches no assertion, and `ARCH-MOCK`'s real-bare-origin fixture is unchanged. The gitx suite passes in 16.07s.
- **The "coverage loss: none" claim is checkable and checks out.** At 6828a2a, all four `subcommandGuardSource` call sites were queue subcommands; nothing else in the tree used it.

### 2. Critical findings

None.

### 3. Important findings

**BR-10 — not-addressed.** `workshop/issues/000207-sync-without-worktree.md:178-179` still reads *"ariadne#218 removed the queue verb that was `TrunkFile`'s first consumer, so **this issue is now its only one**"* — one of the five live sites BR-10 measured, and it directly contradicts the atlas line corrected in the same commit (`atlas/workflow/sdlc-binary.md:648`, "no consumer in the tree today"). `git grep TrunkFile -- '*.go' | grep -v _test.go` outside `trunkfile.go` is empty, so #207 is not a consumer.

The substantive part is *why* it survived: the derivation the `## Log` adopted (`git grep -n 'ariadne#207' …`) **structurally cannot reach a site that refers to #207 as "this issue"** — a file naming itself doesn't match its own id. The hand list BR-10 supplied caught it; the derivation that replaced the hand list dropped it. Fix the derivation, not the sentence: the enumeration for a removal is `git grep -n -i 'queue' -- workshop/issues/ ':!*000218*'` (returns exactly 000207:178-179 today) **∪** the `ariadne#207` sweep, and the `## Log` should record that a self-referential artifact is the known blind spot of an id-based grep.

### 4. Minor findings

- **BR-9 — not-addressed.** `cmd/sdlc/repoguard.go:24` still says `(an env, not 7 new per-verb flags)`. `grep -rn "guardSpineRepo(" cmd/sdlc/*.go` gives eight call sites (changecode, claim, close, merge, milestoneclose, projectclose, push, startplan), so the count is wrong — and it sits ten lines below the sentence this commit added: *"A comment cannot hand-restate a derived set — or its cardinality — and stay true."* The rule was written and its own file was not swept: instance, not class (`ARCH-PURPOSE`). Fix: `(an env, not one flag per guarded verb)`. While there, note the cited grep self-matches twice (the comment line at :10 and the func def at :65) — `git grep -n 'guardSpineRepo(cmd' -- 'cmd/sdlc/*.go'` returns only real call sites.
- **BR-12 — not-addressed.** `workshop/issues/000207-sync-without-worktree.md:181` is byte-identical to when it was raised; the file has not been touched since b5be04d. Split the line after `caller's.)*`.
- **NEW — `atlas/index.md:13`, 2nd in family `prose-edit-orphans-sentence`.** Deleting the trailing `— and \`queue\`, … #209` item took the em-dash that closed `migrate`'s gloss with it, leaving `…and the read-only \`resolve\`/\`open\` ref resolver #144, \`migrate\` — cross-repo artifact move with ref rewrite #179)` — a list whose last item trails after the conjunction, so `#144, \`migrate\`` can be misread as part of the resolver. Per the escalation rule I am not asking for this instance to be patched in isolation: the rule is *a prose deletion is checked by re-reading the enclosing sentence end-to-end, not the deleted span* — deletions that are locally clean orphan the conjunction, the closing dash, or the following clause. Measured prevalence on ariadne#218: 2 (this and BR-12), both introduced by b5be04d, both invisible in a diff hunk because the hunk only shows what left.
- **NEW — `ARCH-DRY`, `cmd/sdlc/internal/gitx/trunkfile.go:450-452` and `trunkfile_test.go:719-721.`** The mode-preservation rationale is duplicated verbatim across the two files. It cost four identical edits in this one issue: both copies carried the stale "fine for a queue" phrasing, both were re-anchored in b5be04d, both regressed to the premature "#207 points this at", and both were re-fixed in 11749d5. The test comment should carry only what the test pins and defer the *why* (`// why: see TrunkFile.commitAndPush's mode-preservation comment`).

### 5. Test coverage notes

No behavior changed, so there is nothing new to pin, and the boundary correctly relies on the surviving suite rather than adding a test that asserts an absence. The two claims that a green suite *cannot* show were both stated in the Spec and both independently verified above (zero `TrunkFile` production callers; `subcommandGuardSource` was queue-only). I did **not** run `go test ./cmd/sdlc/` — this review runs inside the `sdlc close` transaction, which holds `.git/sdlc.lock`, and `close_test.go:131` would park on it for `DefaultWaitTimeout`. That is ariadne#219, correctly diagnosed as pre-existing; the producible-anywhere list stands in for it, and since this diff only deletes tests it strictly reduces runtime.

### 6. Architectural notes

`ARCH-PURE` pass — a pure package (`internal/queue`) and its IO shell are removed together, and `gitx` keeps its `runGitIn` injection seam. `ARCH-MOCK` pass — the fixture still drives a real bare origin with the real git binary; the rename touches names only. `ARCH-SECURE` pass — this removes the tree's only hand-editable trunk-resident input along with its fuzz corpus; nothing new parses untrusted data. `ARCH-CONSTRAINTS` pass — runtime strictly decreases, and the one envelope defect in view is filed as #219 rather than absorbed. `ARCH-ORDER` pass — the CAS interleaving coverage survives intact (`RetryReRunsTransformOnMovedBase`, `OneFetchPerAttempt`, `RetryExhaustionSurfacesGitRejection`, `NonRetryablePushFailsFast`), each forcing a specific ordering through the injectable seam. Worth carrying into #207: removing the queue removes the only *intent-replaying* caller, so "mergeability is the caller's property" is now exercised only by append transforms — #207 should restore a caller-side test that a colliding id re-allocates rather than fast-forwards, which is the amendment 000207:181 already calls for. `ARCH-DRY` and `ARCH-PURPOSE` flagged above.

### 7. Plan revision recommendations

- Append to the `## Log` derived-sweep entry: the `ariadne#207` grep's blind spot (a file that names itself "this issue"), the supplementary command `git grep -n -i 'queue' -- workshop/issues/ ':!*000218*'`, and the disposition of `000207-sync-without-worktree.md:178`.
- Per AGENTS.md §1, mid-stream artifact revisions belong in a `## Revisions` section with timestamp + reason + delta; the last two rounds folded them into `## Log`. Not blocking, but the issue has now accumulated three rounds of revision history with no `## Revisions` heading.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Done-when :159-163 and Plan :304 both re-worded to branch-tense; verified git ls-tree a722b52 still has the file and merge-base(HEAD,origin/main)==a722b52.
  - id: BR-8
    disposition: addressed
    note: |
      atlas :648-651 now "no consumer in the tree today"; confirmed zero non-test TrunkFile callers outside trunkfile.go.
  - id: BR-9
    disposition: not-addressed
    note: |
      repoguard.go:24 still reads "not 7 new per-verb flags" while guardSpineRepo has 8 call sites — the same comment block that now forbids restating a derived set's cardinality.
  - id: BR-10
    disposition: not-addressed
    note: |
      000207-sync-without-worktree.md:178 ("this issue is now its only one") is undispositioned; the adopted ariadne#207 grep cannot reach a file that names itself "this issue".
  - id: BR-11
    disposition: addressed
    note: |
      lessons.md:31-45 states the shell-independence rule, the git grep preference, and the record-the-output requirement.
  - id: BR-12
    disposition: not-addressed
    note: |
      000207-sync-without-worktree.md:181 is unchanged since b5be04d; the sentence still starts on the aside's closing line.
findings:
  - id: new
    severity: Minor
    family: prose-edit-orphans-sentence
    title: |
      atlas/index.md:13 — removing the trailing queue item orphaned the conjunction and migrate's closing em-dash
    detail: |
      2nd in this family (with BR-12), both from b5be04d. Do not patch this
      instance alone: the rule is that a prose deletion is verified by re-reading
      the enclosing sentence end-to-end, not the deleted span, because a diff hunk
      shows only what left. The list now reads "…and the read-only resolve/open ref
      resolver #144, `migrate` — … #179)", so `migrate` trails after the
      conjunction and can be misread as part of the resolver.
  - id: new
    severity: Minor
    family: duplicated-rationale-comment
    title: |
      The mode-preservation rationale is duplicated verbatim in trunkfile.go:450-452 and trunkfile_test.go:719-721
    detail: |
      ARCH-DRY. One idea, two copies, four identical edits inside this single issue —
      both carried the stale "fine for a queue" text, both were re-anchored in
      b5be04d, both regressed to the premature "#207 points this at", and both were
      re-fixed in 11749d5. The test comment should state what the test pins and
      defer the why to the production comment it is testing.
```
