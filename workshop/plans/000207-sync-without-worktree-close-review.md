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
