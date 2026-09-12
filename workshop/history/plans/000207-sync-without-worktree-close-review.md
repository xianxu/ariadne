# Boundary Review — ariadne#207 (whole-issue close)

| field | value |
|-------|-------|
| issue | 207 — Publish issue files without a main worktree |
| repo | ariadne |
| issue file | workshop/issues/000207-sync-without-worktree.md |
| boundary | whole-issue close |
| milestone | — |
| window | ff1b52b6311ce92bd014c626f1b089fc01aa7860..c9106d153ae2ad7307e8f62100d559c5b83b5e61 |
| command | sdlc close --issue 207 |
| reviewer | claude |
| timestamp | 2026-09-11T11:04:04-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The main change is sound. The old worktree route and its three guards are gone, `gitx.TrunkFile` is used, and the collision decision is a pure function called from inside the per-attempt `prepare`. I reverted each claimed fix in a scratch copy to confirm its test goes red: BR-1, BR-3, BR-4 (`rc` reset), BR-5, hoisting the id-space read out of the loop, and making `UpdateMany` reuse its first set all fail their tests. What blocks shipping is the re-allocate arm, the riskiest part of the Spec. It fails in ways I reproduced against a scratch copy of HEAD (the repo was left untouched):

- **`issue new` prints a deleted path.** After a re-allocation, stdout names the old file, which `finish()` has already removed.
- **Orphans survive a successful publish.** The BR-4 fix resets `rc` on every attempt, which drops the list of files earlier attempts wrote. So a two-attempt publish leaves a full duplicate of the issue behind, and the retry picks a second new id even when the first was still free.

Several claimed fixes are also incomplete: BR-3's first half, BR-7's `issue-sync.md`, and the "by fake" half of BR-2.

**Build and tests:** build and vet are clean. The full `cmd/sdlc` suite is green in a git-initialized scratch copy, except the known ariadne#210 failure, which is also red at base.

## 1. Strengths
- **The collision check is really inside the retry loop.** `refIDSpace` is read inside `prepare` (`synctrunk.go:124`), and `TestSyncViaTrunk_ReallocatesOnMidRetryCollision` goes red when that read is hoisted out. `TestUpdateMany_PathsAreDerivedPerAttempt` also runs against a real bare origin with a real peer push.
- **`TrunkWrite.Delete` is its own field.** `TestUpdateMany_WholeSetEarlyReturn` pins the harder case: a pending delete with matching writes must still commit.
- **BR-8's real-git tests are genuine.** They cover a dirty, mid-rebase main worktree left untouched, a `pre-receive` hook that declines the push after the local write, and a byte-identical round-trip under `.gitattributes`.
- **`decideCollision` is pure and covers all three outcomes.** It is table-tested with no IO, including the case where both our path and a foreign path are on the trunk.
- **One default subject for both arms.** `defaultSyncSubject` is shared, and the fix is pinned against the bare origin's actual commit.

## 2. Critical
**C1. `issue new` stdout names a file that no longer exists after a re-allocation** (`issue.go:361`). `shown` is computed before the sync, and nothing passes the new path back to the caller. Reproduced by letting a peer publish `000002` after allocation: stdout was `workshop/issues/000002-mine.md`, but only `000003-mine.md` exists. Stdout is the documented path contract, so an agent that writes to it recreates the colliding id.
- *Fix:* return the final path from `syncIssuesToMain`/`syncViaTrunk` and print it.
- *Test:* run `runIssueNew` with a `claimRunner` whose allocation fetch fails, so it allocates from a stale ref while `UpdateMany`'s own fetch sees the peer.

**C2. Files written by earlier attempts are forgotten** (`synctrunk.go:123`, `:165`, `:180`). `rc = nil` runs at the start of every attempt, which throws away `rc.written`. Two consequences:
- `finish()` can no longer remove candidates that rejected attempts wrote.
- `unionIDSpace`'s local scan sees those candidates, so the retry re-allocates again.

Reproduced with the collision present on both attempts: the id went 700→701→702 (701 was free both times), with two contradictory "now 0007xx" announcements. `000701-mine.md` was left as an untracked full copy of the issue with a valid id, which a later bare `sdlc claim` would publish as a second issue. `TestReallocation_FinishRemovesOldAndOrphans` passes only because it builds `written` by hand.
- *Fix:* keep the candidate list in a closure variable outside the per-attempt reset, and leave those files out of the local union (or reuse the previous candidate id if it is still free on the new base).

## 3. Important
- **I1. Two re-allocations in one pass delete a published file** (`synctrunk.go:175-179`, BR-3 part (a)). `rc` holds only one old/new pair. Reproduced: `000701-a.md` was published and then deleted by `finish()`, while `000700-b.md` stayed behind. The test #188 asked for was never written. Reachability through current callers is low.
- **I2. Spec step 5 is not delivered.** When retries run out (`updatemany.go:106`) the error is the generic "trunk moved under 3 attempts"; `rc`/`Foreign` are dropped on the error path, so no colliding paths are named. The Done-when "an unrelated-path retry succeeds without re-allocating" has no test, and C2 shows it is currently false.
- **I3. Concurrent edits to the same issue are now silently overwritten.** Content is replaced last-writer-wins, and that includes a stale double-claim. The old `CONFLICT BEHAVIOR` promise ("refuses") was removed from `claim.md:52` with no replacement statement. The Spec's premise that the merge-base check existed "only because the route edits a shared checkout" is wrong: it detected exactly this. At minimum, document the loss in the help text and atlas and file a follow-up.
- **I4. `unionIDSpace` silently ignores a failed local scan** (`synctrunk.go:204`). Its justification comment is false: the push's compare-and-swap cannot see local-only ids, so it will not reject a clash with one. It is also a second union with a different failure policy from `allocateIssueID`, which returns the error, and `nextFreeID` (fill the first gap) is a second allocation rule beside `issue.NextID` (max+1) (`ARCH-DRY`).
- **I5. Stale docs remain.** `atlas/workflow/issue-sync.md:20-42` still describes the worktree route and conflict detector (BR-7 named this file). `changecode.go:257` and `sdlc-binary.md:69` still describe copying into the main worktree. `sdlc-binary.md:94` has a spliced, broken sentence.
- **I6. `UpdateMany`/`commitSetAndPush` duplicate `Update`/`commitAndPush`.** `Update`, `Read` and `ReadDegraded` now have no production consumers. The retry-policy tests (declined hook, attempt limit, offline, file mode) only exercise the dead path. Rebuild `Update` as a thin wrapper over `UpdateMany` (`ARCH-DRY`).
- **I7. A quoted path is treated as a deletion** (`synctrunk.go:133`). Reproduced: git quotes `000300-café.md` in its output, `ReadFile` reports not-exist, the path becomes a `Delete` of a nonsense name, nothing is published, and the arm prints `synced`. The old arm failed loudly on the same read (`ARCH-SECURE`).

## 4. Minor
- `FailedPublishKeepsTheOriginalFile` cannot fail: it has no collision, so `rc` is nil. It stays green with the BR-2 fix reverted.
- `TrunkView.Ref()` returns the tracking ref's name, which can move, rather than the pinned `base` SHA (`ARCH-ORDER`).
- `refIDSpace`'s `ls-tree` resolves paths against the current directory (verified), so the new guard is blind from a subdirectory. This predates the window, and the `resolveIDDirs` comment claiming otherwise is false.
- If a publish fails after a re-allocation, `issue new` commits the old colliding id and recommends `issue sync --issue <old> --push`, which will refuse.
- `idFrontmatterRE` matches any line that starts with `id:`, not just the frontmatter.
- `decideCollision` does not notice two files in the same publish set sharing one id.

## 5. Test coverage
- **Pinned (verified by reverting):** BR-1, BR-3, BR-4, BR-5, the loop placement, and per-attempt derivation. BR-2 is pinned only by the real-git test.
- **Missing:**
  - collision present on two attempts (would catch C2)
  - two files re-allocated in one pass (I1)
  - `runIssueNew` end-to-end with a re-allocation (C1)
  - `UpdateMany` running out of retries, declined hook, offline, file mode
  - the unrelated-path retry Done-when
  - quoted paths (I7)
- No real-git test exercises re-allocation across a real retry; only the fake does.

## 6. Architecture
| Principle | Verdict | Notes |
|---|---|---|
| ARCH-DRY | flag | I4, I6. Reusing `refIDSpace`, `issueIDFromPath` and `defaultSyncSubject` is good. |
| ARCH-PURE | pass, with a note | The `prepare` closure still mixes decisions, `os.WriteFile` and `rc` changes, and C2 and I1 both live there. A pure `plan(...) → (TrunkWrite, []realloc)` would make them unit-testable. |
| ARCH-PURPOSE | flag | The `realloc-ordering-contract` family is back: BR-2 and BR-4, now C2 and I1. The per-attempt, per-file states have never been enumerated. `free-id-space-incomplete` repeats too. |
| ARCH-MOCK | pass, with a note | Real bare-origin tests exist. `fakePublisher` shares the seam but models neither the compare-and-swap nor the no-op case. |
| ARCH-CONSTRAINTS | pass | Bounded at 3 attempts; git calls grow linearly with files per attempt. |
| ARCH-SECURE | flag | I7 and the cwd-relative read. Arguments are passed as separate argv and `--end-of-options` is used, which is good. |
| ARCH-ORDER | flag | C2, I1 and the `TrunkView` ref. Replace `rc` with a per-file record returned by the last attempt, plus a candidate list that is never reset. |

## 7. Plan revisions
1. Spec step 5 (the retry-exhaustion message) and the Done-when about unrelated-path retries are not delivered. Mark them undelivered until fixed.
2. Correct the Spec's statement that the merge-base check existed only for shared-checkout safety, and record that same-issue edits are now last-writer-wins.
3. Spec step 3: a successful multi-attempt publish also leaves a duplicate today. Once fixed, state that exactly one file (the new path) survives.
4. The Log says BR-2 was checked "by fake and by real git" and that every BR-1–13 is addressed. Correct both: BR-2 is pinned by real git only, and BR-3(a) and BR-7 remain open.

```findings
findings:
  - id: new
    severity: Critical
    family: stale-machine-output
    title: |
      issue new prints the pre-reallocation path on stdout; that file was deleted by finish()
    detail: |
      issue.go:361 prints `shown`, computed before the sync; syncIssuesToMain returns no
      final path. Reproduced end-to-end: stdout workshop/issues/000002-mine.md, which is
      gone; only 000003-mine.md exists. Surface the final path and print it.
  - id: new
    severity: Critical
    family: realloc-ordering-contract
    title: |
      per-attempt rc reset forgets earlier candidates: orphan duplicate survives, retry re-reallocates
    detail: |
      synctrunk.go:123 drops rc.written, so finish() cannot remove rejected attempts'
      candidates, and unionIDSpace counts them. Reproduced: 700->701->702 with 701 free,
      and an untracked full-duplicate 000701-mine.md left for a bare `claim` to publish.
  - id: new
    severity: Important
    family: realloc-ordering-contract
    title: |
      two re-allocations in one pass mispair rc; finish() deletes a just-published file
    detail: |
      BR-3(a) only half-fixed. Reproduced: 000701-a.md published then removed, 000700-b.md
      (colliding) kept. rc is a single pair; the requested two-file test was not written.
  - id: new
    severity: Important
    family: donewhen-not-delivered
    title: |
      Spec step 5 exhaustion message and the unrelated-retry Done-when are undelivered/untested
    detail: |
      updatemany.go:106 emits a generic contention error and rc/Foreign are dropped on
      error. No test covers "unrelated-path retry succeeds without re-allocating"; C2
      shows it currently re-allocates.
  - id: new
    severity: Important
    family: deleted-guard-unreplaced
    title: |
      concurrent same-issue trunk edits (incl. stale double-claim) are now silently overwritten
    detail: |
      The merge-base check detected these; the Spec deleted it on a wrong premise and
      claim.md's CONFLICT BEHAVIOR dropped the promise without stating last-writer-wins.
      Document the loss and file a lost-update follow-up.
  - id: new
    severity: Important
    family: free-id-space-incomplete
    title: |
      unionIDSpace silently swallows a failed local scan on a false CAS justification
    detail: |
      synctrunk.go:204. The CAS cannot see local-only ids, so it will not reject. It is a
      second union with a different failure policy from allocateIssueID, and nextFreeID is
      a second allocation rule beside issue.NextID (ARCH-DRY).
  - id: new
    severity: Important
    family: stale-user-facing-docs
    title: |
      atlas/workflow/issue-sync.md still documents the deleted worktree route (BR-7 claimed addressed)
    detail: |
      issue-sync.md:20-42; also changecode.go:257 and sdlc-binary.md:69 describe copying
      into the main worktree, and sdlc-binary.md:94 has a spliced sentence.
  - id: new
    severity: Important
    family: parallel-implementation
    title: |
      UpdateMany/commitSetAndPush duplicate Update/commitAndPush; Update has no production consumer
    detail: |
      Retry-policy tests (declined hook, exhaustion, offline, mode) cover only the dead
      path. Re-express Update as an adapter over UpdateMany.
  - id: new
    severity: Important
    family: failed-read-as-benign-state
    title: |
      a git-quoted path reads as not-exist, becomes a Delete, and the arm prints synced
    detail: |
      Reproduced with 000300-café.md: nothing published, `synced` emitted. Classify deletion
      from git status (-z, --name-status) and error on an unexplained not-exist read.
  - id: new
    severity: Minor
    family: tautological-test
    title: |
      FailedPublishKeepsTheOriginalFile stays green with the BR-2 fix reverted
    detail: |
      No collision, so rc is nil. Only the real-git test pins BR-2, which contradicts the
      Log's "by fake and by real git".
  - id: new
    severity: Minor
    family: snapshot-not-pinned
    title: |
      TrunkView.Ref returns the mutable tracking-ref name rather than the resolved base SHA
  - id: new
    severity: Minor
    family: cwd-relative-git-read
    title: |
      refIDSpace ls-tree is cwd-relative, so the collision guard is blind from a subdirectory
    detail: |
      Pre-existing, and the resolveIDDirs comment claims otherwise. Fix with --full-tree or
      by reading through TrunkView in the repo root.
  - id: new
    severity: Minor
    family: realloc-ordering-contract
    title: |
      issue new fallback after a failed realloc publish commits the old id and gives refusing advice
  - id: new
    severity: Minor
    family: identity-rewrite-scope
    title: |
      idFrontmatterRE matches any line-start id:, not only the frontmatter block
  - id: new
    severity: Minor
    family: free-id-space-incomplete
    title: |
      decideCollision ignores two files in the same publish set sharing one id
```

---

## Re-review — 2026-09-11T11:46:11-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 207 — Publish issue files without a main worktree |
| repo | ariadne |
| issue file | workshop/issues/000207-sync-without-worktree.md |
| boundary | whole-issue close |
| milestone | — |
| window | ff1b52b6311ce92bd014c626f1b089fc01aa7860..d64e6c352470ed219d1404d2c6bf6cd88439ebaf |
| command | sdlc close --issue 207 |
| reviewer | claude |
| timestamp | 2026-09-11T11:46:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

Build and vet are clean on a scratch copy of HEAD. The targeted suites (`cmd/sdlc`, `internal/gitx`) pass once `TestRunIssueNew_FromGitHubFillsProblem` is skipped. That test calls `die()` when there is no GitHub remote, which exits the test binary, so any run that includes it stops early with a bare `FAIL`. The two full-suite failures come from files missing in the scratch copy, not from this window. I reverted 16 claimed fixes one at a time: 10 turned a test red, 5 left the suite green, and one mutation did not compile. BR-1 and BR-3 are fixed and pinned. Three things stop a clean SHIP. BR-2's fix has no test that fails without it, and a retry still steps over its own earlier candidate: I reproduced 700→702 with 701 free. Both doc sites BR-7 named are unchanged. And there is a new instance of a family that has come up before: from any subdirectory, `issue new`, `claim` and `issue sync` publish nothing and print `[ok] No issue changes to sync.` That is the exact failure in #207's Problem section, reproduced with the HEAD binary. None of this needs a redesign, but BR-2 stays open as Critical in the ledger until a test pins it.

## 1. Strengths
- **Per-attempt and cross-attempt state are now separate.** `publishResult` resets `reallocs` on each attempt and keeps `candidates` and `collisions` across attempts, and the field comments say why (`synctrunk.go:49-62`). This is the right shape for ARCH-ORDER.
- **The `newTrunkPublisher` seam (`synctrunk.go:40`) gives `runIssueNew` real end-to-end coverage.** Reverting BR-1 or BR-13 turns a test red.
- **`Update` is now a thin adapter over `UpdateMany` (`trunkfile.go:337-349`).** There is one CAS loop, and `trunkfile_test.go` is unchanged and green, so its retry-policy tests now cover the live path.
- **The real-git tests are strong:**
  - a `pre-receive` hook confirms the local write happens before the push;
  - a dirty, mid-rebase main worktree is left untouched;
  - a fresh clone's file is byte-identical under `.gitattributes`;
  - non-ASCII names are covered through both `diff` and `ls-files`.
- **`decideCollision` stays pure, with three named outcomes.** BR-15's refusal of two files sharing one id is pinned.

## 2. Critical
**BR-2 is not addressed** (`synctrunk.go:188`, `:196`, `:314`).
- **The fix has no test.** Adding `res.candidates = nil` next to the per-attempt `reallocs` reset leaves every targeted test green. Every multi-attempt test either has no collision on attempt 1, or calls `syncViaTrunkWithRealloc`, which never runs `finish()` (`StaleReallocationIsNotCarriedAcrossAttempts`). The code does currently remove the stale candidate; I confirmed that with my own test, but nothing in the suite guards it.
- **The retry still re-allocates over its own candidate.** With the collision present on both attempts, attempt 1 writes `000701-mine.md`. `unionIDSpace`'s local scan then counts that file as a reservation, so attempt 2 publishes 702 while 701 is free on the trunk.
- *Fix:* leave `res.candidates` out of the local union. Then add a `syncViaTrunk` test with a `fakePublisher{rerun: 1}` and the collision on both attempts, asserting `NewID == 701` and that exactly one local file survives. That test goes red under either regression.

## 3. Important
- **New, family `cwd-relative-git-read`: publishing from a subdirectory silently does nothing.** Reproduced with the HEAD binary from `docs/sub` on a feature branch against a bare origin: exit 0, `No issue changes to sync.`, the file left untracked, and origin unchanged. It also reproduces at base on main, so it predates this issue.
  - Cause: `execGitRunner.Git` runs in the process cwd (`runner.go:36`). The pathspecs in `changedIssueFiles` (`claim.go:348-351`) and `issueFilesForID`'s glob therefore resolve against the cwd, and `ls-files` returns cwd-relative paths that `prepare` then joins onto the repo root.
  - `issue.go:270-273` claims the opposite, that the sync's pathspecs are read against the repo root.
  - The family-repeat rule is in the findings block below.
- **BR-7 is not addressed.**
  - Both sites it named are unchanged: `changecode.go:257-260` still describes copying into the main worktree, and the broken sentence at `sdlc-binary.md:94-97` is still broken.
  - Stale siblings nobody swept: `issue.go:270-273`, `issue.go:354-357`, `sdlc-binary.md:104-108` and `issue_test.go:202-204` all still describe the "no worktree on main" trigger or worktree copies.

## 4. Minor
- **BR-11:** the fix is correct (`updatemany.go:88` passes the base SHA), but pointing the view back at the tracking ref leaves every test green.
- **BR-14:** removing the frontmatter scoping (`reallocate.go:62`) leaves every test green. The frontmatter `id:` always matches first; the untested case is a frontmatter block with no `id:` plus a prose `id:` line.
- **BR-9's fallback guard (`synctrunk.go:211-219`) is never exercised.** Disabling it leaves everything green. The quoting bug itself is pinned by the `-z` tests.
- **BR-4:** resetting `collisions` on each attempt leaves everything green, because the exhaustion test runs a single attempt.
- **The lost-update docs don't name the follow-up.** `issue-sync.md:56` cites "(ariadne#207 BR-5)" and `claim.md` names nothing, while the Done-when says ariadne#222 carries the design.
- **BR-6 residual:** `nextFreeID` (first free id above ours) is still a second allocation rule beside `issue.NextID` (highest + 1). It can reuse a gap; this is rarely reachable.
- `appendUnique` duplicates what `slices.Contains` already does.

## 5. Test coverage notes
| Mutation (fix reverted) | Result |
|---|---|
| BR-1 path printed on stdout, BR-3 single-record, BR-6 swallowed scan, BR-9 `ls-files -z`, BR-10 cleanup on failure, BR-12 `--full-tree`, BR-13 advice, BR-13 error wrapping, BR-15 two files at one id, BR-4 exhaustion message | caught |
| BR-2 candidates reset per attempt, BR-9 not-exist guard, BR-11 tracking-ref name, BR-14 frontmatter scope, BR-4 collisions reset per attempt | **survived** |

Re-allocation across a real retry still has no real-git test; only the fake covers it. Every test runs from the repo root, which is how the subdirectory gap went unnoticed.

## 6. Architecture
| Marker | Verdict | Notes |
|---|---|---|
| ARCH-DRY | pass, with a note | BR-8 is fixed and `refIDSpace` is consumed. `nextFreeID` vs `NextID` remains. |
| ARCH-PURE | flag (note) | The `prepare` closure still mixes pure decisions with `os.WriteFile` and result-state changes, and BR-2's residue lives there. A pure `planPublish(...) → (TrunkWrite, []reallocation)` would make it unit-testable. |
| ARCH-PURPOSE | flag | Three fixes covered the named instance rather than the class: the cwd class, the stale-docs class, and BR-2 (orphans fixed, but its own candidates still count as reservations). |
| ARCH-MOCK | pass | The real bare-origin tests are good. The fake shares the seam but does not model the CAS. |
| ARCH-CONSTRAINTS | pass | Retries are bounded at 3, and git calls grow linearly with the number of files. |
| ARCH-SECURE | pass, with a note | Arguments go as argv arrays, with `--end-of-options` and `-z`. The not-exist guard exists but is never exercised. |
| ARCH-ORDER | flag | The cross-attempt fields are exactly the ones no test observes across attempts. `beforePrepare` could inject the ordering, but no test uses it with a persistent collision. |

Architectural note for later: `changedIssueFiles` scans only `IssuesDir`, so an uncommitted move out of it (a hand-made archive, say) publishes a bare Delete. The trunk would then hold that id in neither directory.

## 7. Plan revision recommendations
1. Spec lines 59-61 still say the merge-base check existed "only because the current route edits a shared checkout". Add a Revisions entry: it also caught concurrent edits to the same issue, that guarantee is gone, and ariadne#222 tracks it.
2. The Log's "Twelve mutations run, twelve caught" and the BR-2 narrative need correcting: the candidate persistence is not pinned, and the 700→702 skip remains.
3. The Done-when about `--full-tree` holds only for id-space reads. Either widen it to "publishing from a subdirectory lands the file", or mark it undelivered.
4. Spec step 1: state that the local id-space union excludes this publish's own candidates.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Reverting the stdout path fix fails TestRunIssueNew_PrintsThePathItActuallyPublished.
  - id: BR-2
    disposition: not-addressed
    note: |
      Resetting res.candidates per attempt leaves every test green (no test runs finish() after a two-attempt re-allocation), and the retry still re-allocates over its own candidate: reproduced 700 to 702 with 701 free, because unionIDSpace counts it.
  - id: BR-3
    disposition: addressed
    note: |
      Collapsing reallocs to a single record fails TestSyncViaTrunk_TwoReallocationsInOnePublish.
  - id: BR-4
    disposition: addressed
    note: |
      The exhaustion message is pinned and the unrelated-retry test exists; accumulating collisions across attempts is unpinned (resetting them per attempt leaves the suite green).
  - id: BR-5
    disposition: addressed
    note: |
      claim.md and issue-sync.md state last-writer-wins and #222 exists; neither doc names #222, and the Spec premise is uncorrected.
  - id: BR-6
    disposition: addressed
    note: |
      The swallowed local scan is pinned by RefusesWhenTheLocalIDScanFails; nextFreeID (first free id above ours) still diverges from NextID (highest + 1), with low reachability.
  - id: BR-7
    disposition: not-addressed
    note: |
      changecode.go:257-260 and the broken sentence at sdlc-binary.md:94-97 are unchanged; unswept siblings at issue.go:270-273, issue.go:354-357, sdlc-binary.md:104-108 and issue_test.go:202-204.
  - id: BR-8
    disposition: addressed
    note: |
      Update is an adapter over UpdateMany, commitAndPush is gone, and trunkfile_test.go is unchanged and green.
  - id: BR-9
    disposition: addressed
    note: |
      -z on all three queries is pinned by real-git and stub tests; the not-on-disk-and-not-tracked guard (synctrunk.go:211-219) is never exercised.
  - id: BR-10
    disposition: addressed
    note: |
      Calling finish() on the failure path fails both the fake and the real-git test.
  - id: BR-11
    disposition: not-addressed
    note: |
      The code pins the base SHA, but pointing the view back at trackingRef() leaves every test green; no test moves the ref inside prepare.
  - id: BR-12
    disposition: addressed
    note: |
      Removing --full-tree fails TestRefIDSpace_IsNotBlindFromASubdirectory; the same family recurs elsewhere as a new finding.
  - id: BR-13
    disposition: addressed
    note: |
      Both the advice branch and the errIDTaken wrapping are caught when reverted.
  - id: BR-14
    disposition: not-addressed
    note: |
      Removing the frontmatter scoping leaves every test green; add a frontmatter block without id: plus a prose id: line, which must refuse.
  - id: BR-15
    disposition: addressed
    note: |
      Disabling the same-id check fails TestSyncViaTrunk_RefusesTwoFilesClaimingOneID.
findings:
  - id: new
    severity: Important
    family: cwd-relative-git-read
    title: |
      issue new, claim and issue sync publish nothing from a subdirectory and report ok
    detail: |
      This is the 2nd finding in family cwd-relative-git-read. Reproduced with the HEAD
      binary from docs/sub on a feature branch against a bare origin: "[ok] No issue
      changes to sync.", exit 0, the file left untracked, and origin unchanged. The id is
      never reserved, which is the failure in #207's own Problem. Also reproduced at base
      on main, so both arms had it. execGitRunner.Git runs in the process cwd
      (runner.go:36). The changedIssueFiles pathspecs and the issueFilesForID glob resolve
      against the cwd, and ls-files returns cwd-relative paths that prepare then joins
      onto the repo root. issue.go:270-273 asserts the opposite.
      RULE: every git call on the issue-sync and id path runs with its working directory
      pinned to the repo top level. Build the gitRunner rooted there, the way
      NewTrunkFile refuses an empty dir, and join every path that reaches os.* onto that
      same root. A per-call --full-tree is an instance of the rule, not the rule.
      Measured prevalence: 2 sites patched one at a time (refIDSpace, pathTrackedAtHEAD);
      at least 5 unpatched on the same path (three changedIssueFiles queries at
      claim.go:348-351, the issueFilesForID glob, the syncPathspec add/commit). Pin it
      with a test that publishes with cwd set to a subdirectory.
```

---

## Re-review — 2026-09-11T12:20:19-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 207 — Publish issue files without a main worktree |
| repo | ariadne |
| issue file | workshop/issues/000207-sync-without-worktree.md |
| boundary | whole-issue close |
| milestone | — |
| window | ff1b52b6311ce92bd014c626f1b089fc01aa7860..1649ab12df12d0ba555679832d08e51bbb2b72ff |
| command | sdlc close --issue 207 |
| reviewer | claude |
| timestamp | 2026-09-11T12:20:19-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core of #207 holds up. Issue files publish from a feature branch with no worktree on main, the collision decision runs inside the per-attempt `prepare`, and `issue new` re-allocates correctly from a subdirectory (reproduced with the HEAD binary). BR-2 is fixed: I reverted each half in a scratch clone at HEAD and a test went red every time.

Two things block a clean SHIP:
- **BR-16, the cwd-relative family:** round 2 again patched individual call sites rather than the rule, and the site-level fix broke two paths that work at base. I checked this against a base-built binary and with a test.
- **A new Important:** the trunk arm hardcodes `workshop/history`, so under `WF_HISTORY_DIR` the collision guard cannot see archived ids.

BR-11 and BR-14 are correct in the code, but no test fails without them. Build and `go vet` are clean. The `cmd/sdlc` and `gitx` suites are green except for three failures that are not from this window:
- `TestProjectCloseRejectsDuplicateLogicalMVPScopeRefs` is also red at base.
- `TestFleetPlanHas…` and `TestProcessManualCmd_WiredInRoot` fail the same way on an unmutated clone, because files that are not in git are missing there.

`TestRunIssueNew_FromGitHubFillsProblem` was skipped, as in round 2.

## 1. Strengths
- **BR-2 is fixed at the root.** `publishResult` keeps per-attempt `reallocs` separate from the cross-attempt `candidates`/`collisions` (`synctrunk.go:49-62`). The retry also excludes its own earlier candidates from the id space (`synctrunk.go:200-224`). All three reverts were caught (table in §5).
- **The subdirectory test uses real git** (`issuesync_test.go:1119`). It runs `syncIssuesToMain` against a bare origin through both the git-query path and the `PublishExisting` glob path. Reverting either site turns it red.
- **`claimRunnerStub.GitInDir` now falls back to `Git`** (`claim_test.go:51-54`). The stub now behaves like real git instead of answering one call and not the other.
- **BR-7's named sites and their siblings are all rewritten.** The sentence at `sdlc-binary.md:94-100` is repaired, and `changecode.go:257-262` now bases the decision on what would be published.

## 2. Critical
None.

## 3. Important

**I1 — BR-16 is not addressed. The rule was not applied, and the site-level fix broke two working paths.**

Reproduced from `docs/sub` against a bare origin:

| scenario | base | HEAD |
|---|---|---|
| `issue sync --issue 1` (no push), relative dir | `[ok] No issue changes`, exit 0 (silent no-op) | `git add` exit 128 |
| same, **absolute** `--issues-dir` | committed, exit 0 | `git add` exit 128 (**regression**) |
| `claim --issue 1 --no-start`, on main | not run | `git add` exit 128 |
| `claim --issue 1` (start flip), feature branch | code unchanged | `no issue file matches` (`setstatus.go:118`) |
| `resolveBranchName`, absolute dir (test) | `000005-x` | `exists in glob but is not a readable regular file` (**regression**) |
| `issue new`, feature branch | not run | publishes correctly |

Why each row fails:
- **In-place arm:** `syncInPlace` still runs `git add` and `git commit` in the process cwd (`claim.go:251`, `:261`). BR-16 named this site explicitly.
- **The two regressions:** `issueFilesForID` changed what it returns, from paths in the form the caller gave to repo-relative paths. Its two consumers were not updated: `syncPathspec` feeds the result to a cwd-relative `git add`, and `resolveBranchName` calls `os.Stat` on it (`branchcreate.go:65`).
- **Swallowed error:** `issueFilesForID` also turns a `RepoTopLevel` error into "no match" (`issuefiles.go:136-139`).
- **The shape of the class:** the repo root is now computed separately at five sites on one publish path.

The fix is the rule, not more sites:
- Resolve the id directories once, at the dispatch. `resolveIDDirs` already returns the typed value (`idDirs{Top, Rel, Abs}`).
- Nothing below `runClaim` or `syncIssuesToMain` should accept a raw `IssuesDir` string or run git without `Dir = Top`.
- Wrap the runner so its `Git` calls `GitInDir(Top, …)`.
- Pass `idDirs` to `changedIssueFiles`, `syncPathspec`, `issueFilesForID` and `locateIssueFile`.
- Pin it by extending the subdirectory test to: no-push `issue sync`, `claim` on main, `claim --issue N` with the start flip, and an absolute `--issues-dir`.

**I2 (new) — the trunk arm ignores the configured history directory.** This is the 3rd finding in family `free-id-space-incomplete`.
- **Cause:** `synctrunk.go:182` passes the literal `"workshop/history"`, and `claimFlags` has no `HistoryDir`. Meanwhile `issue new` allocates with `f.HistoryDir` (`issue.go:281`), and `merge` archives into `WF_HISTORY_DIR`.
- **Verified with `WF_HISTORY_DIR=archive`:**
  - Re-allocation moved 700 to 701 while `archive/000701-shipped.md` was already on the trunk.
  - A republication beside `archive/000700-shipped.md` was published with no refusal.
- **Impact:** this is pair#179's worst case (a live issue and an archived one sharing an id). The Done-when "sync/claim refuse on collision" does not hold under a supported configuration.
- **The rule:** every consumer of the id space derives from one resolution of the configured directories. The same resolve-once change as I1 carries both fixes.

## 4. Minor
- **Leftovers from the deleted worktree route** (new, 2nd finding in `stale-user-facing-docs`):
  - `sdlc-binary.md:233` describes `claim`'s main-worktree precheck as if it still exists.
  - `claim.go:133-137` refers to a "main-worktree route", "worktree hunt" and "network pull".
  - `claim.go:326` is an empty section header.
  - `mustGitOutput` (`claim.go:419`) has no callers; its only caller at base was the deleted merge-base check.
- **BR-11:** the code is correct, but pointing the view back at the tracking ref (m6) leaves both suites green.
- **BR-14:** the code is correct, but removing the frontmatter scoping (m7) leaves everything green. `TestRewriteIdentity_LeavesAProseIDLineAlone` puts the frontmatter `id:` first, so an unscoped regex passes it too. The missing fixture is a frontmatter block with no `id:` followed by a prose `id: 000999` line, which must refuse.
- `issue-sync.md:56` and `claim.md:63-68` still don't point at ariadne#222.
- `changecode.go:261` is an unwrapped comment line of about 110 columns.

## 5. Test coverage notes

| mutation (fix reverted) | result |
|---|---|
| m1: drop the own-candidate exclusion (`synctrunk.go:204`) | caught by `RetryReusesItsOwnCandidateID` |
| m2: reset candidates on each attempt (`:188`) | caught by the same test |
| m3: `finish()` stops removing rejected candidates (`:78`) | caught by `FinishRemovesOldAndOrphans` |
| m4: `changedIssueFiles` back to the cwd (`claim.go:366`) | caught by `PublishesFromASubdirectory` |
| m5: the `issueFilesForID` glob back to the cwd (`issuefiles.go:142`) | caught by the same test |
| m6: TrunkView uses the tracking-ref name (`updatemany.go:88`) | **survived** (BR-11) |
| m7: `idFrontmatterRE` unscoped (`reallocate.go:62`) | **survived** (BR-14) |
| m8: `pathTrackedAtHEAD` back to the cwd (`synctrunk.go:305`) | survived, harmlessly: `--full-tree` already anchors this `ls-tree` to the repo root |

Every subdirectory test exercises the trunk arm. None covers the in-place arm, the start flip, or an absolute dir, which is exactly where I1 lives. No test sets `WF_HISTORY_DIR`.

## 6. Architecture

| marker | verdict | notes |
|---|---|---|
| ARCH-DRY | flag (minor) | The repo root is resolved at `claim.go:150` and `:354`, `synctrunk.go:170` and `:301`, and `issuefiles.go:136`. `repoRel` (`issuefiles.go:116`) re-implements `gitx.InsideRoot` (`inside.go:31`) with a different symlink policy. The fixture now resolves symlinks itself (`issue_test.go:426-430`), which suggests the two policies should be one. |
| ARCH-PURE | pass, with a note | Unchanged from round 2: `prepare` mixes the decision with `os.WriteFile` and changes to the result. A pure `planPublish(...)` would let the BR-2 exclusion be tested without git. |
| ARCH-PURPOSE | flag | I1: this is the third round of site fixes in `cwd-relative-git-read`, and this round's fix broke two working paths. I2: the issue's central guarantee, the collision guard, does not cover the configured archive. |
| ARCH-MOCK | pass | Tests run against a real bare origin, and the subdirectory test uses real git. |
| ARCH-CONSTRAINTS | pass | Retries are bounded at 3. About six `rev-parse` subprocesses per publish is off the hot path. |
| ARCH-SECURE | flag | The raw `IssuesDir` string travels below the verb boundary untyped. Its consumers read it against three different bases: the cwd, git's root, and the symlink-resolved root. `idDirs` is the typed value this principle asks for; it exists but isn't passed down. This is the shared cause of I1 and I2. |
| ARCH-ORDER | pass, with a note | Per-attempt and cross-attempt state are now explicit. BR-11 has no test that could catch it, because no test moves the ref inside `prepare` and `fakePublisher` does not model the compare-and-swap. |

## 7. Plan revision recommendations
1. **Done-when "Publishing works from a SUBDIRECTORY":** this holds for the trunk arm only. It fails for `claim` and `issue sync --push` on main. Either deliver it and pin it, or mark it partial.
2. **Round-2 Log entry:** add that the in-place arm and the start flip are still cwd-relative, and that absolute `--issues-dir` regressed in `issue sync` and in `change-code` branch naming.
3. **Spec, re-allocate step 1:** say that the id space covers the configured issues and history dirs, or record the current literal as a known limit.
4. **BR-7's "Swept by derivation":** name the list of terms searched. Two prose sites and a dead helper escaped it.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      m1 and m2 caught by RetryReusesItsOwnCandidateID, m3 by FinishRemovesOldAndOrphans; the 700 to 702 walk is gone.
  - id: BR-7
    disposition: addressed
    note: |
      Every named site and every round-2 sibling rewritten and read-verified; the remaining leftovers are raised as a new Minor.
  - id: BR-11
    disposition: not-addressed
    note: |
      Code pins the base SHA, but m6 (view back on trackingRef()) leaves both suites green; no test moves the ref inside prepare.
  - id: BR-14
    disposition: not-addressed
    note: |
      m7 (scoping removed) is green; LeavesAProseIDLineAlone puts the frontmatter id first, so it cannot tell. Needs frontmatter-without-id plus a prose id line, which must refuse.
  - id: BR-16
    disposition: not-addressed
    note: |
      Three more sites patched, not the rule. syncInPlace add/commit (claim.go:251,261) still run in the cwd:
      no-push issue sync and claim on main exit 128 from a subdirectory (HEAD binary). startOnClaim via
      locateIssueFile (setstatus.go:118) makes claim --issue N die. Regressions vs base: issueFilesForID now
      returns repo-relative paths, so an absolute --issues-dir from a subdirectory fails in issue sync
      (base committed) and in resolveBranchName (base named the branch). Root resolved at 5 sites; idDirs not threaded.
findings:
  - id: new
    severity: Important
    family: free-id-space-incomplete
    title: |
      trunk arm hardcodes workshop/history and ignores WF_HISTORY_DIR, so the collision guard cannot see archived ids
    detail: |
      This is the 3rd finding in family free-id-space-incomplete. synctrunk.go:182 passes a literal, and
      claimFlags has no HistoryDir, so claim, issue sync and issue new's sync (issue.go:334) cannot pass on
      the dir that allocateIssueID (issue.go:281) and merge honour. Verified with WF_HISTORY_DIR=archive:
      re-allocation went 700 to 701 onto archive/000701-shipped.md, and a republication beside
      archive/000700-shipped.md published with no refusal. RULE: every id-space consumer derives from ONE
      resolution of the configured dirs, resolveIDDirs(IssuesDir, HistoryDir) at the dispatch, passed down
      as idDirs; no call site names a directory literal. Prevalence: 1 literal, 3 verbs that cannot pass it
      on. The same resolve-once fix covers BR-16.
  - id: new
    severity: Minor
    family: stale-user-facing-docs
    title: |
      leftovers from the deleted worktree route survived the sweep: two prose sites, a dead helper, an empty header
    detail: |
      This is the 2nd finding in family stale-user-facing-docs. sdlc-binary.md:233 describes claim's
      main-worktree precheck as if it still exists; claim.go:133-137 refers to a main-worktree route,
      worktree hunt and network pull; claim.go:326 is an empty section header; mustGitOutput (claim.go:419)
      has no callers, since its only caller at base was the deleted merge-base check. RULE: a deletion
      sweep searches for the deleted symbols AND the words used to describe them in every spelling
      (main-worktree, worktree hunt, precheck, merge-base), plus helpers left with no callers, and records
      that list in the Log. Prevalence: 2 prose sites, 1 dead helper, 1 empty header.
```

---

## Re-review — 2026-09-11T12:52:26-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 207 — Publish issue files without a main worktree |
| repo | ariadne |
| issue file | workshop/issues/000207-sync-without-worktree.md |
| boundary | whole-issue close |
| milestone | — |
| window | ff1b52b6311ce92bd014c626f1b089fc01aa7860..168c36aaa8ed176e203eeb57ae5834b85bf2d1d8 |
| command | sdlc close --issue 207 |
| reviewer | claude |
| timestamp | 2026-09-11T12:52:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The four prior-round fixes I could reach are real, not paper: I rebuilt the HEAD binary and reproduced `issue new` / `claim` / `issue sync` (both arms) publishing correctly from `docs/sub` against a bare origin, and `WF_HISTORY_DIR=archive` now both blocks re-allocation onto an archived id and refuses a republication beside one — and four independent mutations (unanchor `locateIssueFile`, unpin `changedIssueFiles`, unanchor `issueFilesForID`, hardcode `workshop/history`) each turn a test red, so BR-16 and BR-17 are pinned rather than asserted. What keeps this off SHIP is one Important finding and one prior finding that does not survive its own mutation check: the collision guard decides against the raw trunk read alone, so a same-id slug rename (`git mv 000700-old.md 000700-new.md`) is refused even though the same atomic publish already carries `Delete(old)`, and restoring an archived issue is refused with "already published under a different name" printing two *identical* filenames plus advice that cannot clear it — both reproduced with the HEAD binary. BR-14's guard is correct in code but its test passes with the fix reverted, so it is unpinned. The suite is otherwise green; the one failing test (`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`) is pre-existing — the plan it reads was archived by `dfeba9c`, which predates this window's base.

## 1. Strengths

- `cmd/sdlc/synctrunk_test.go:200` — `ReallocatesOnMidRetryCollision` seeds the peer **inside** the retry window via `beforePrepare` and asserts the final set carries the re-derived path. This is the assertion that actually distinguishes "decision inside the CAS loop" from "decision hoisted out", and it is the one the Spec was rewritten around.
- `cmd/sdlc/issuesync_test.go:952` — the `pre-receive` hook writes a witness file only if the candidate exists at push time. That converts an ordering claim ("local write precedes the push") into an observable, and it also asserts the trunk did not move, the original survived, the candidate was removed, and `synced` was not printed. Four distinct failure modes, one fixture.
- `cmd/sdlc/claim.go:221-240` — `syncPaths{Root, Dirs}` resolved once at the dispatch and threaded is the correct answer to BR-16/BR-17, not a third patch. The deliberate exception (`locateIssueFile` anchoring internally, `setstatus.go:119`) is recorded in the code rather than hidden.
- `cmd/sdlc/internal/gitx/updatemany.go:24-31` — `TrunkWrite.Delete` as its own field makes removal representable instead of collapsing into "unchanged"; `setMatchesTrunk` then correctly gates the whole-set early return on both halves.
- `cmd/sdlc/internal/gitx/trunkfile.go:337` — collapsing `Update` into a `UpdateMany` adapter removed the duplicate retry loop and `commitAndPush` entirely; `refIDSpace` (`issueids.go:158`) remains the only trunk id-space reader, confirmed by sweeping every `ls-tree` call site in `cmd/`.

## 2. Critical findings

None.

## 3. Important findings

**The collision guard decides against the trunk snapshot alone — not against the id space this publish will leave behind** (`cmd/sdlc/synctrunk.go:248`, `cmd/sdlc/issuecollision.go:56`). **This is the 4th finding in family `free-id-space-incomplete`.** Earlier rounds fixed instances (BR-3 local unpublished ids, BR-15 two files in one set, BR-17 the configured history dir). Do not fix these two instances — fix the rule.

Reproduced with the HEAD binary:

```
git mv workshop/issues/000700-old-slug.md workshop/issues/000700-new-slug.md
sdlc claim --issue 700
  ==> Publishing issue changes to the trunk:
      workshop/issues/000700-new-slug.md
      workshop/issues/000700-old-slug.md          <- the Delete is IN the set
  Error: issue id 000700 is already published under a different name: ...
         Rename one side by hand and re-run
```

and, restoring an archived issue (`workshop/history/000700-seven.md` → `workshop/issues/000700-seven.md`):

```
  Error: issue id 000700 is already published under a different name:
        ours:  workshop/issues/000700-seven.md
        trunk: workshop/history/000700-seven.md   <- same name, different dir
```

Both are permanent for `claim` and `issue sync --push`: re-running produces the identical refusal, and the advice names an action the operator has already taken. The publish is atomic, so applying the rename would leave exactly one path at id 700.

**RULE:** `decideCollision` must be given the id space **as this publish will leave it** — `trunk ∪ local`, plus this set's `Write`, minus this set's `Delete` — with our own artifact recognised by `(id, slug)` rather than by exact path, so a file moved between the issues dir and the history dir is not read as a foreign claimant. Compute it once, in one place, and pass it in; `space` as a raw `refIDSpace` result is the shared defect behind all four findings. The cheapest structural form is a pure `planPublish(changed, contents, idSpace, firstPublication) -> (TrunkWrite, []reallocation, error)` extracted out of the 90-line `prepare` closure (ARCH-PURE) — the delete/rename/realloc interaction is exactly the combination a table-driven pure test enumerates and the current IO-bound closure cannot. Measured prevalence: 4 instances of one rule, 3 already fixed individually.

## 4. Minor findings

- **`resolveBranchName` is now half-anchored** (`cmd/sdlc/branchcreate.go:58-64` vs `:100`): the `--issue` glob was root-anchored this round while its sibling `listUntrackedIssues` still runs `r.Git("ls-files" … issuesDir+"/")` in the process cwd. **4th finding in family `cwd-relative-git-read`** — do not patch the site. Rule (unchanged from BR-16/BR-17): each verb resolves root+dirs once at its own dispatch and threads them; no call site re-resolves or names a literal. Unswept siblings measured: `listUntrackedIssues` (branchcreate.go:100), `findIssueFileByName` (changecode.go:485,489), `scanIssueFiles`' diff pathspec and glob (issuefiles.go:35-42), `milestoneclose.go:682`'s literal `"workshop/history"`. `change-code` from a subdirectory fails either way, so nothing regressed — but the mixed state is what the rule exists to prevent.
- **`sdlc claim --help` gives two different flag lists.** `--history-dir` was registered (claim.go:100) but not added to the hand-written FLAGS block in `cmd/sdlc/helptext/claim.md`; cobra's auto "Flags:" section below it does list it. **3rd finding in family `stale-user-facing-docs`** — rule: the FLAGS block is a hand-maintained restatement of the cobra flag set (ARCH-PURPOSE: a restatement that doesn't derive is a deferred consumer). Either derive it or add a test asserting every registered flag name appears in its helptext.
- **`repoRel` duplicates `gitx.InsideRoot`** (issuefiles.go:103 vs internal/gitx/inside.go:31) — same job, and the copy's containment test (`strings.HasPrefix(rel, "..")`) is weaker than `gitx.Escapes`. **2nd finding in family `parallel-implementation`** — rule: extend the existing helper (teach `InsideRoot` to resolve the root too) rather than adding a second one beside it.
- **Two doc comments were detached from their functions** by insertions in this window: `syncInPlace`'s comment now documents `type syncPaths` (claim.go:210-225), and `issueFilesForID`'s now documents `repoRel` (issuefiles.go:103-111). Both functions are left undocumented.
- `os.WriteFile(abs, newData, 0o644)` (synctrunk.go:261) drops the original file's mode on a re-allocation, where `commitSetAndPush` is careful to preserve it from the base tree.
- `setMatchesTrunk` → `readFrom` → `refPresent` re-runs `rev-parse` on the same immutable base SHA once per path, and `commitSetAndPush` adds an `ls-tree` per path for the mode — ~6 git processes per file per attempt (ARCH-CONSTRAINTS: repeated work that can be hoisted). Fine at N=1–3; worth hoisting if a no-`--issue` `claim` ever carries a large changed set.
- `var _ = errors.Is` (synctrunk_test.go:132) is a leftover no-op.
- The `hash-object --path` "BOUNDED CLAIM" comment about `.gitattributes` resolving from the working tree was deleted with `commitAndPush` and not carried into `commitSetAndPush`; the limitation still applies.

## 5. Test coverage notes

- Mutation-verified this round: unanchoring `locateIssueFile`, `changedIssueFiles`, and `issueFilesForID`, and replacing `f.historyDir()` with the literal, each turn a named test red. BR-16 and BR-17 are genuinely pinned.
- **BR-14 is not pinned.** `TestRewriteIdentity_LeavesAProseIDLineAlone` (reallocate_test.go:112) passes with the whole `frontmatterSpan` restriction removed, because its fixture has a frontmatter `id:` and the regex's first match is that line either way. The fixture that separates them: frontmatter present but **without** an `id:` line, plus `id: 000999` in the body — fixed code refuses, unfixed code silently rewrites prose and renames.
- **BR-11 is unpinned too**, but has no cheap mutation target: replacing the resolved base SHA with `t.trackingRef()` leaves both packages green, and only a concurrent fetch between `resolve` and `prepare` can distinguish them. The property is readable off the type (`base` is a SHA from `resolve`), so it cannot silently do nothing — noted rather than re-raised.
- The rename and archive-restore paths in the Important finding have no test at any layer; add them to `issuecollision_test.go` as pure cases once the decision takes the post-publish id space.
- The suite is red at HEAD on `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` for a pre-existing reason (ariadne#210). Any `--verified` evidence should say so explicitly rather than quoting a green run.

## 6. Architectural notes

- **ARCH-DRY** — flag (Minor): `repoRel` vs `gitx.InsideRoot`. Otherwise strong: one id-space reader, one `defaultSyncSubject`, one retry loop.
- **ARCH-PURE** — flag (folded into the Important finding): `decideCollision`/`nextFreeID`/`rewriteIdentity` are properly pure and IO-free in tests, but the `prepare` closure (synctrunk.go:180-280) mixes file reads, git queries, candidate writes and the publish-set decision in one body. The uncaught rename/delete interaction lives exactly there.
- **ARCH-PURPOSE** — pass on the headline purpose (publishing needs no worktree; verified end-to-end), flag on the shadow sweep: the `--history-dir` helptext restatement doesn't derive, and the cwd-relative class is swept on the sync/id path but half-swept in `branchcreate.go`.
- **ARCH-MOCK** — pass. Real bare origins via `testfix`, a declining `pre-receive` hook for the ordering property, and a `fakePublisher` that models the retry rather than a stateless call recorder. `ViewOf` exists precisely so consumer tests can drive a real read.
- **ARCH-CONSTRAINTS** — pass with a note: attempts bounded at 3, candidates bounded by the changed set, no unbounded fan-out; the per-file git-process count is the only repeated-work item.
- **ARCH-SECURE** — pass. `--end-of-options` and `--` separate caller-supplied refs and pathspecs structurally; `-z` on every query removed the quoted-path misread; an unexplained not-exist errors instead of being published as a deletion; `resolveIDDirs` refuses dirs outside the repo. Minor: candidate files are written 0644 regardless of the original's mode.
- **ARCH-ORDER** — pass. `publishResult`'s per-attempt vs across-attempt split is written down at the field level and both halves are covered (stale-realloc-not-carried, candidate-reused-on-retry). Note for later: it is still four independent fields whose legal combinations are implied rather than enumerated; if a fifth arrives, collapse to a tagged state.

## 7. Plan revision recommendations

- **Done-when, "ariadne#188 closes as superseded"** — not satisfied at this boundary, and deliberately so: Plan step 6 was reworded to "record which bullet shipped here (`b7f08ec`)", leaving #188's close to #188's own lifecycle. Revise the Done-when line to match the Plan, or it will read as an unmet exit criterion at close.
- **Spec, "Same path, same slug, is never a collision"** — the code implements same *path*; the archive-restore reproduction shows same *slug* at a different path is treated as foreign. Decide which the Spec means and say so, then make the code match.
- **Spec, the `issue sync`/`claim` refusal row** — add the carve-out the Important finding asks for: a foreign trunk path that this same publish deletes (a slug rename) is not a collision, and the refusal text must not tell an operator to do the thing they just did.
- **Done-when** — add the two cases now missing: a same-id slug rename publishes as one atomic delete+write, and restoring an archived issue republishes rather than refusing.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      updatemany.go passes the resolved base SHA into TrunkView; no test pins it (the mutation to trackingRef leaves both packages green) because only a concurrent fetch can distinguish them.
  - id: BR-14
    disposition: not-addressed
    note: |
      Code is correct but unpinned — TestRewriteIdentity_LeavesAProseIDLineAlone still passes with frontmatterSpan removed; needs a fixture whose frontmatter has no id: line while the body does.
  - id: BR-16
    disposition: addressed
    note: |
      Reproduced fixed from docs/sub with the HEAD binary for issue new, claim and issue sync (both arms); three mutations each turn a named test red.
  - id: BR-17
    disposition: addressed
    note: |
      WF_HISTORY_DIR=archive verified end-to-end: allocation skips the archived id and republication beside it refuses naming the archived path; the literal-dir mutation is caught.
  - id: BR-18
    disposition: addressed
    note: |
      mustGitOutput, the empty header and both prose sites are gone; residual "no worktree hunt" phrasing in claim.go:14 and sdlc-binary.md:122 no longer distinguishes the arms but is cosmetic.
findings:
  - id: new
    severity: Important
    family: free-id-space-incomplete
    title: |
      the collision guard reads the raw trunk, ignoring this publish's own deletes and same-slug moves, so a slug rename and an archive-restore are refused permanently
    detail: |
      This is the 4th finding in family free-id-space-incomplete. Earlier rounds fixed instances
      (BR-3 local unpublished ids, BR-15 two files in one set, BR-17 the configured history dir).
      Do NOT fix these two instances. Reproduced with the HEAD binary against a bare origin:
      `git mv 000700-old-slug.md 000700-new-slug.md` then `claim --issue 700` prints both paths in
      the publish set and still refuses, although the set carries Delete(old) and the commit is
      atomic; and restoring workshop/history/000700-seven.md to workshop/issues/ refuses with
      "already published under a different name" while printing two identical filenames. Both are
      permanent for claim and `issue sync --push`, and the advice ("Rename one side by hand and
      re-run") names an action already taken. RULE: decideCollision must receive the id space as
      this publish will LEAVE it — trunk union local, plus this set's Write, minus this set's
      Delete — with our own artifact keyed on (id, slug) rather than exact path, computed once and
      passed in. Extracting a pure planPublish(changed, contents, idSpace, firstPublication) out of
      the prepare closure (synctrunk.go:180-280) is what makes the delete/rename/realloc combinations
      testable at all (ARCH-PURE). Measured prevalence: 4 instances of one rule, 3 fixed individually.
  - id: new
    severity: Minor
    family: cwd-relative-git-read
    title: |
      resolveBranchName is half-anchored — the --issue glob moved to the repo root while its sibling ls-files still runs in the process cwd
    detail: |
      This is the 4th finding in family cwd-relative-git-read. Do NOT patch the site. branchcreate.go:58-64
      now root-anchors issueFilesForID while listUntrackedIssues (branchcreate.go:100) still passes a
      relative pathspec to r.Git in the cwd, so one function answers from two different trees. RULE
      (unchanged from BR-16/BR-17): each verb resolves root + configured dirs ONCE at its own dispatch
      and threads them; no call site re-resolves or names a literal. Unswept siblings measured:
      listUntrackedIssues, findIssueFileByName (changecode.go:485,489), scanIssueFiles' diff pathspec
      and glob (issuefiles.go:35-42), and milestoneclose.go:682's literal "workshop/history".
      change-code is broken from a subdirectory before and after, so nothing regressed — the mixed
      state is the finding.
  - id: new
    severity: Minor
    family: stale-user-facing-docs
    title: |
      claim --help prints two different flag lists — the hand-written FLAGS block omits the new --history-dir that cobra's auto section shows
    detail: |
      This is the 3rd finding in family stale-user-facing-docs. Do NOT just add the line. --history-dir
      was registered on claim (claim.go:100) in this window; cmd/sdlc/helptext/claim.md's FLAGS block
      was not updated, and `sdlc claim --help` therefore lists the flag once and omits it once. RULE:
      the FLAGS block is a hand-maintained restatement of the cobra flag set, i.e. a consumer that does
      not derive (ARCH-PURPOSE) — either render it from the flag set or add a test asserting every
      registered flag name appears in that verb's helptext. Prevalence: 1 flag today across 30 helptext
      files, none of which is checked against its command.
  - id: new
    severity: Minor
    family: parallel-implementation
    title: |
      repoRel duplicates gitx.InsideRoot with a weaker containment test
    detail: |
      This is the 2nd finding in family parallel-implementation. issuefiles.go:103 adds a second
      "make p relative to the repo root" helper beside internal/gitx/inside.go:31, differing only in
      resolving symlinks on the root as well and in returning bool instead of error — and its escape
      check (strings.HasPrefix(rel, "..")) is weaker than gitx.Escapes. RULE: when an existing helper
      is nearly right, extend it (teach InsideRoot to resolve the root) rather than adding a sibling;
      two helpers answering one question drift, and the divergence is a bug in whichever lacks the guard.
  - id: new
    severity: Minor
    family: doc-comment-anchoring
    title: |
      two doc comments were detached from their functions by insertions in this window
    detail: |
      claim.go:210-225 — syncInPlace's doc comment now documents `type syncPaths`, and syncInPlace
      (claim.go:251) has none; issuefiles.go:103-111 — issueFilesForID's comment now documents
      repoRel, and issueFilesForID (issuefiles.go:113) has none. Both read as documented while godoc
      attaches the prose to the wrong symbol. Insert new declarations after the documented function,
      not between the comment and its subject.
```
