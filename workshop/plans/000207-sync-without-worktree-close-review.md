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
