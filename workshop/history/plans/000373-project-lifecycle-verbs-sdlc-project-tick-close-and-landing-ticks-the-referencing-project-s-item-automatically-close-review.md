# Boundary Review — ariadne#373 (whole-issue close)

| field | value |
|-------|-------|
| issue | 373 — Project lifecycle verbs: sdlc project tick/close, and landing ticks the referencing project's item automatically |
| repo | ariadne |
| issue file | workshop/issues/000373-project-lifecycle-verbs-sdlc-project-tick-close-and-landing-ticks-the-referencing-project-s-item-automatically.md |
| boundary | whole-issue close |
| milestone | — |
| window | 72bf044e19b3be76bc22851008aaf9fbf2f30dc8..6a45043f308811eb127044b83ebcdfbb788713e1 |
| command | sdlc close --issue 373 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T03:55:50-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the #373 diff now: I've read the core tick path and both merge hooks, and next I'm checking close, status and the tests.

Still reviewing: checking how the new MVP gate fits the gate-signature catalog, then I'll run the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The code does what the Spec asks. `project.TickLandedIssue` is pure and idempotent. The landing tick goes to each project's origin main through one compare-and-swap commit (`gitx.TrunkFile`). It runs from both the durable and the legacy merge paths, and it never fails a merge. `MVPProgress` is now the one definition of "MVP done": the `mvp-landed` guard, `project status` and `project close` all use it. Closing a linked ops record now updates the real file and moves the link, which fixes a real orphaning bug. The build passes, and all targeted tests pass: `internal/project`, `processmanual`, `recovery`, and the `cmd/sdlc` close, status, tick, landing and gates tests (288s). Three cheap fixes stand before SHIP:
- **README is stale.** It shows `sdlc project close --slug example`, which now fails without `--verified`, and it never mentions `project tick`.
- **The TL-facing help contradicts the code.** `cmd/tl-lead/help.md` still describes the old ownership and archive rules.
- **The tick runs under the merge lock.** It does network IO to other repos' origins with no timeout while holding that lock.

Running `project status` against the real ops file from this review checkout could not resolve sibling repos ("no sibling repo matches"). That is a limit of where the review runs, not a code defect. I accept the Log's real-file dry-run claims, which are specific.

### 1. Strengths
- `cmd/sdlc/internal/project/project.go:59-101`: the open-row match (`openIssueRow`) is now shared with the close-time tick. `TickLandedIssue`'s test pins every case that matters: `[-]` untouched, `[.]` ticked, a PR suffix added to already-ticked rows, `PR #410` not mistaken for `#41`, and a second run returning identical bytes.
- `cmd/sdlc/projectlandtick.go:42-60`: the `landing_tick` policy is read again from origin's bytes inside the transform, so a stale local copy never authorizes a write to a TL-owned file. Origin's bytes are parsed before anything is written; a copy that fails to parse is refused (ARCH-SECURE).
- `publishProjectEdit` (`projectlandtick.go:156`) folds the brain-repo refusal and peer publish into one helper, which `publishPeerProject` now reuses (ARCH-DRY).
- Tests use real temp repos with bare origins. They cover the Done-when's cross-repo landing on both the durable path (`TestLandingTicksPeerProject`) and the legacy path. They also check that the peer checkout's HEAD, index and worktree are untouched, that a rerun publishes nothing, and that a TL-owned project needs to opt in (ARCH-MOCK).
- `projectclose.go:76-129`: the gates are collected instead of returning at the first one, so `--dry-run` reports every refusal while a real close still refuses on the first.

### 2. Critical findings
None.

### 3. Important findings
- **README is not updated** (`README.md:468-480`).
  - The example `sdlc project close --slug example` now fails, because `--verified` is required.
  - `--no-retro` and `--drop` examples sit next to it; `--drop` still works without `--verified`.
  - `sdlc project tick`, `close --dry-run`, `--verified`, the `mvp:`/`done_when:` output of `status`, `landing_tick`, and `--force` now also waiving the MVP gate are all missing.
  - Fix: add `--verified '…'` to the close example, and add a `project tick --issue … [--pr N] [--dry-run]` line with a note on `landing_tick`.
- **`cmd/tl-lead/help.md:85-92` contradicts the new behavior** (ARCH-PURPOSE: one consumer wasn't updated).
  - It says "Archive: at `sdlc project close`, move the directory to teams/<team>/history/projects/<slug>/ and drop the symlink". The code now updates the real file in place and moves the link into `workshop/history/projects/`, and the team folder stays.
  - It says "workers never write it. Tick a Breakdown row … only after you have verified the merge". This ignores the landing tick, the `landing_tick: true` opt-in, and the `sdlc project tick` command that merge now prints for the TL.
  - Fix: update the guidance to match `construct/datatype/project.md` § Ownership.
- **The landing tick holds the merge lock during network IO with no timeout** (ARCH-CONSTRAINTS / ARCH-ORDER; `landing.go:625`, `merge.go:748-750`).
  - `tickOnLanding` runs inside `f.withLock(...)` and fetches from and pushes to other repos' origins through `peerTrunk` and `commandContext`, with no timeout. The landing's own git calls go through `boundedLandingGit`, which has one.
  - The tick is optional and only warns on failure, yet a hung peer origin would hold this repo's merge lock and block other slots.
  - TrunkFile's attempt limit caps retries, not a hung fetch or push.
  - Fix: run the tick after the locked closure returns (durable path), or after the cleanup steps (legacy path), with a context deadline.

### 4. Minor findings
- `commitLinkedProjectClose` (`projectclose.go:286`): if the final `os.Remove(livePath)` fails, the real file is already closed and both links exist. A rerun then refuses with "archived project already exists". Nothing rolls back.
- `mentionsPR` compiles its regex on every row (`project.go:99`). It's cheap, but `strings.Contains` with a word-boundary check, or a cached pattern, would do.
- The MVP gate's `--force` waiver has no row in `processmanual/gatesig.go`, unlike the retro waiver. Its acknowledgement text "--force: closing with the MVP incomplete" is invisible to the friction catalog.
- `runProjectTick` doesn't count dry-run "Would skip" outcomes as missed, so a dry run exits 0 even when the real run would fail.
- Discovery takes two routes: `landedProjectMatches` for merge and `discoverProjectsInScope` for the hand-run tick. Both end in `DiscoverInRoots`, but they resolve the repo name differently (`identity.Repo` vs the resolved target's `Repo`).

### 5. Test coverage notes
- The Done-when cross-repo test exists on both merge paths. The Log's mutation claims are plausible given the assertions: for example, `TestProjectCloseLinkedRecord` checks that the archive is a link and that the real file was updated.
- No test covers a merge hook that finds zero matching projects, or own-repo ticking through the legacy merge path where `mainPath` is behind origin afterwards. The latter behaves correctly, but nothing pins it.
- `TestTickLandedIssue` doesn't cover a row with trailing whitespace. The trim handles it.

### 6. Architectural notes
Pass or flag for each principle:
- **ARCH-DRY: pass.** Shared matcher, `MVPProgress`, `publishProjectEdit`.
- **ARCH-PURE: pass.** `TickLandedIssue`, `MVPProgress`, `LandingTick` and `renderMVP` are pure and tested without IO.
- **ARCH-PURPOSE: flag.** The tl-lead help wasn't updated.
- **ARCH-MOCK: pass.** Real git against bare origins; GitHub is untouched and the PR number is an input.
- **ARCH-CONSTRAINTS: flag.** Unbounded IO under the lock.
- **ARCH-SECURE: pass.** Origin's bytes are parsed and the policy re-read on them; `--pr` is validated.
- **ARCH-ORDER: pass,** with the Minor partial-failure note on linked close. The compare-and-swap commit plus idempotence covers retries and a resumed landing.
- **ARCH-FUNERAL: pass.** One commit per landing; the team folder outliving the close is declared out of scope.

Flipping `TLLandingTickDefault` later is a one-line change, and the tests reference the constant, so they keep up with it.

### 7. Plan revision recommendations
- `## Revisions`: Core concepts lists `landingTickTransform(ref, pr)` and `tickLanded(ctx, ref, pr)`. The code has `landingTickTransform(repoName, issueID, pr, auto)`, `tickLandedProjects`, `tickOnLanding` and `publishProjectEdit`. Update the rows.
- `## Revisions`: the durable plan's Task checkboxes are all still `[ ]` while the issue's Plan shows them all `[x]`. Tick them.
- `## Revisions`: note that `--force` on `project close` now also waives the MVP gate.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: docs-surface-not-swept
    title: |
      README project section is stale: close example now fails without --verified; project tick, --dry-run, landing_tick undocumented
    detail: |
      README.md:468 shows `sdlc project close --slug example`, which now refuses (--verified required, projectclose.go:117). The new `project tick` verb, close --dry-run/--verified, status mvp/done_when output, landing_tick, and --force waiving the MVP gate are absent from README.
  - id: new
    severity: Important
    kind: defect
    family: docs-surface-not-swept
    title: |
      cmd/tl-lead/help.md contradicts the new linked-close and landing-tick behavior
    detail: |
      help.md:85-92 tells TLs to move the team directory to teams/<team>/history/projects/ and drop the symlink at project close, and that workers never write project.md / TLs tick rows by hand. The code now updates the real file in place and moves the link into workshop/history/projects (folder stays), and merge ticks rows (opt-in via landing_tick) or prints `sdlc project tick`. ARCH-PURPOSE: a consumer not swept.
  - id: new
    severity: Important
    kind: defect
    family: optional-work-under-critical-lock
    title: |
      Landing tick does unbounded peer-origin network IO while holding the merge lock
    detail: |
      tickOnLanding runs inside f.withLock (landing.go:625, merge.go:750) and uses peerTrunk with commandContext (no deadline), unlike the landing's boundedLandingGit. A hung peer fetch/push blocks other slots' merges for warn-only work. Move it after the locked closure and give it a timeout (ARCH-CONSTRAINTS).
  - id: new
    severity: Minor
    kind: defect
    family: partial-failure-recovery
    title: |
      commitLinkedProjectClose leaves both links if removing the live link fails; rerun then refuses
    detail: |
      projectclose.go:308: the real file is already closed and the archive link exists; a retry hits "archived project already exists".
  - id: new
    severity: Minor
    kind: defect
    family: waiver-catalog-coverage
    title: |
      MVP-gate --force waiver has no processmanual gatesig row
    detail: |
      The project close retro waiver is cataloged (gatesig.go:128) but the new "--force: closing with the MVP incomplete" acknowledgement is not, so transcript friction analysis can't attribute it.
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      Durable plan's Core-concepts signatures and Task checkboxes don't match the delivered code
    detail: |
      Plan names landingTickTransform(ref, pr) and tickLanded(ctx, ref, pr). Code has landingTickTransform(repoName, issueID, pr, auto), tickLandedProjects, tickOnLanding and publishProjectEdit. The plan's Tasks are all unchecked while the issue shows them done.
```

---

## Re-review — 2026-10-15T05:10:33-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 373 — Project lifecycle verbs: sdlc project tick/close, and landing ticks the referencing project's item automatically |
| repo | ariadne |
| issue file | workshop/issues/000373-project-lifecycle-verbs-sdlc-project-tick-close-and-landing-ticks-the-referencing-project-s-item-automatically.md |
| boundary | whole-issue close |
| milestone | — |
| window | 72bf044e19b3be76bc22851008aaf9fbf2f30dc8..145f753912eecde709fe32b0a2a25931d1a30dcb |
| command | sdlc close --issue 373 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T05:10:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Let me look at the round-1 fix commit first.

Checking the remaining prior fixes: whether the BR-4 rollback test actually fails without the fix, and whether the legacy merge tick test exercises the lock.

Now reverting the BR-4 fix in the disposable checkout to see whether its test goes red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The round-1 fixes are in the right places and the code holds up. The landing tick now runs after the merge lock on both merge paths, with a 2-minute deadline that reaches the peer git operations (`peerTrunk` → `ResolvePublicationTarget` / `NewTrunkFileContext`). The README and the tl-lead help now describe the new close and tick behaviour. `--no-mvp` is in the catalog, and the plan has a Revisions entry plus checked Tasks. What stops a plain SHIP is test evidence for two of the fixes. Moving the tick out of the lock (BR-3) changes behaviour, and no test fails if that change is undone. For BR-4, I reverted the fix in this disposable checkout (`commitLinkedProjectClose` back to its pre-fix body) and `go test ./cmd/sdlc -run TestProjectClose` still passed. The new rollback test only forces a rename failure, and the old code already handled that case. Both gaps are cheap to close. The wider suites passed at head: `./cmd/sdlc` with the close/tick/landing filters, `internal/project`, and `internal/processmanual`.

1. **Strengths**
   - `cmd/sdlc/landing.go:556-632`, `cmd/sdlc/merge.go:682-783`: the tick now sits outside `withLock` on both merge paths and runs only after a successful landing (`err == nil && merged`). The rest of the landing still runs under the lock.
   - `cmd/sdlc/projectlandtick.go:228`: `context.WithTimeout(commandContext(ctx), landingTickTimeout)` bounds every git call to a peer origin, because `peerTrunk` passes the context through.
   - `cmd/sdlc/projectlandtick.go:43-61`: the policy is read again from origin's bytes, and a copy that won't parse is refused rather than rewritten (ARCH-SECURE).
   - `dryRunTick` now returns `ok`, so `project tick --dry-run` exits with the same status the real run would.
   - The `mentionsPR` rewrite drops the per-call `regexp.MustCompile` and keeps the rule that `PR #12` does not match `PR #123`.

2. **Critical:** none.

3. **Important**
   - BR-3 is `not-addressed`, but only for lack of a regression test. The code is correct. The test infrastructure already exists (`merge_main_test.go:68`, a `f.Lock` wrapper around `withRequiredRepoTransactionLock`). A cheap test: wrap `f.Lock` so that when `run()` returns, it asserts the peer's bare origin has **not** yet received the tick commit, then assert after `runMerge` returns that it has. Do this for both the durable and the legacy paths. That test fails on the pre-fix code.

4. **Minor**
   - BR-4 is `not-addressed` for the same reason: no test fails without the fix. Fix: make the `projectCloseRenameFn` stub assert `os.Lstat(livePath)` is already gone before it calls `os.Rename`. That pins the commit-point ordering, and the pre-fix code fails it. Separately, at `projectclose.go:318` the result of `_ = os.Symlink(liveTarget, livePath)` is thrown away. If restoring the link fails, the record is open but has no live link, and the error doesn't say so. Join it into the returned error.
   - New finding, family `waiver-catalog-coverage` (2nd in family). Rule: every `GateCatalog` row's `AckPat` and `RefusalPat` must be pinned by a test that triggers the real waiver and refusal and matches the pattern against the text actually emitted. The new `no-mvp` row has neither test. Its sibling `no-retro` is pinned at `projectclose_test.go:42`.

5. **Test coverage notes**
   - Tests exist for the TickLandedIssue cases, the peer CAS publish, the legacy and durable tick paths, status, and close (`--verified`, MVP refusal, `--dry-run`, linked records).
   - Missing: a check that the tick runs outside the lock, the `--no-mvp` waive path, and the BR-4 ordering.
   - The `landingTickTimeout` deadline is untested. That's acceptable given it simply passes the context through.

6. **Architecture**
   - **ARCH-DRY: pass.** `publishProjectEdit` is shared with close's peer publish.
   - **ARCH-PURE: pass.** `TickLandedIssue` and `MVPProgress` are pure; IO stays in `projectlandtick.go`.
   - **ARCH-MOCK: pass.** Tests run against real bare-repo origins behind the `gitx.TrunkFile` seam.
   - **ARCH-CONSTRAINTS: pass.** The tick is out of the lock and bounded at 2 minutes.
   - **ARCH-SECURE: pass.** Origin's copy is parsed, and an unparseable copy is refused.
   - **ARCH-ORDER: flag (Minor).** The linked-close commit point is now explicit, but neither the ordering nor the swallowed restore error is tested (see BR-4).
   - **ARCH-PURPOSE: pass.** The docs consumers BR-2 named have been updated: README, tl-lead help, helptext, and the datatype doc.
   - **ARCH-FUNERAL: pass.** The only new durable output is commits; the stage file is removed by a `defer`.

7. **Plan revisions:** none beyond the round-1 entry, which records the real names and the move out of the lock.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:468-499 now shows close --dry-run/--verified, project tick, --no-mvp/--force, status mvp/done_when, and landing_tick; matches projectclose.go and projectlandtick.go.
  - id: BR-2
    disposition: addressed
    note: |
      cmd/tl-lead/help.md:85-96 now describes the printed project tick command, the landing_tick opt-in, and an in-place close that moves only the link; matches commitLinkedProjectClose.
  - id: BR-3
    disposition: not-addressed
    note: |
      Code is fixed (tick after withLock on both paths, 2m ctx deadline reaches peerTrunk), but no test fails without it. Add a Lock-wrapper test asserting the peer origin has no tick commit when run() returns, on both durable and legacy paths.
  - id: BR-4
    disposition: not-addressed
    note: |
      Verified: with the fix reverted, TestProjectClose* still passes, because the old code already rolled back a rename failure. Pin the ordering with a rename stub that asserts the live link is gone. Also join the error from _ = os.Symlink restore (projectclose.go:318) into the returned error.
  - id: BR-5
    disposition: addressed
    note: |
      gatesig.go:131-134 adds the no-mvp row; gatesig_test GateFlagNames/GateFlagsFor pin it. Pattern-vs-emitted-text pinning is raised separately below.
  - id: BR-6
    disposition: addressed
    note: |
      Plan Tasks are checked, and the 2026-10-15 Revisions entry records landingTickTransform(repoName, issueID, pr, auto), tickLandedProjects, tickOnLanding, and publishProjectEdit.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: waiver-catalog-coverage
    title: |
      The no-mvp waiver's AckPat/RefusalPat are not pinned against the text close actually emits
    detail: |
      2nd in family. Rule: every GateCatalog row's AckPat and RefusalPat must be matched by a test against the real waiver and refusal output. No test runs --no-mvp or matches "--no-mvp (or --force): closing with the MVP incomplete"; no-retro is pinned at projectclose_test.go:42. Sweep every catalog row, not only this one.
```

---

## Re-review — 2026-10-15T05:53:47-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 373 — Project lifecycle verbs: sdlc project tick/close, and landing ticks the referencing project's item automatically |
| repo | ariadne |
| issue file | workshop/issues/000373-project-lifecycle-verbs-sdlc-project-tick-close-and-landing-ticks-the-referencing-project-s-item-automatically.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7bfdff6ee3dfc9141dc97525f865c5316b156595..0aa64f88819a74197e75d1b0f6f99a032006ab1b |
| command | sdlc close --issue 373 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T05:53:47-07:00 |
| verdict | SHIP |

## Review

Prior findings look addressed so far. Next I'm running the package tests to confirm.

```verdict
verdict: SHIP
confidence: high
```

All three open findings are fixed, and each fix has a regression test that pins the behaviour. I found nothing new that blocks the close.

- **BR-3 (tick under the merge lock):** both merge paths now run the tick after the lock is released (`cmd/sdlc/landing.go:609-617`, `cmd/sdlc/merge.go:778-783`). The whole tick also has a 2-minute limit (`landingTickTimeout`, `cmd/sdlc/projectlandtick.go:217-226`). That limit reaches each git call: `peerTrunk` uses `NewTrunkFileContext(ctx)` and `runIn` respects `ctx`.
- **BR-4 (linked close leaves both links):** the linked close now removes the live link before the rename that commits the close. If the rename fails, it puts the live link back and removes the archive link.
- **BR-7 (no-mvp patterns unpinned):** a table-driven test now checks both `project close` gate rows against the text close actually prints.

The targeted tests (`go test ./cmd/sdlc -run 'ProjectClose|ProjectTick|TickOnLanding|LandingTicks|LegacyLandingTicks'`) pass, as do the `internal/project` and `internal/processmanual` packages.

1. **Strengths**
   - There is now one shared definition of "MVP done", `projectdoc.MVPProgress` (`cmd/sdlc/internal/project/mvp.go`). The mvp-landed guard, `project status` and `project close` all use it, and an unreadable ref is never counted as landed (ARCH-DRY, ARCH-SECURE).
   - `TickLandedIssue` is a pure function and can be run twice safely: `mentionsPR` won't count PR #70 as PR #7. The IO side is a thin layer that commits each project edit to origin in a single check-then-write step. It re-reads the `landing_tick` policy from origin's copy, not from a possibly stale local one (`cmd/sdlc/projectlandtick.go:42-58`).
   - `publishProjectEdit` merges the publish code from close and the landing tick into one place, including the brain-repo refusal (`cmd/sdlc/peerproject.go:81-88`).
   - `tickOutsideLock` checks the "tick runs outside the lock" rule from inside the lock itself, on both the durable and the legacy merge paths. If the tick moved back inside the lock, this test would fail.
   - A dry-run close checks every gate and reports all of them, not just the first refusal (`cmd/sdlc/projectclose.go:77-129`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none raised as findings.
   - Note only: if the durable merge's locked step fails after the PR has already merged (for example, archiving fails), `runDurableMerge` returns before `tickOnLanding` runs. It prints no `sdlc project tick` retry hint. The operator can still run the tick by hand.

5. **Test coverage notes**
   - **Hand-run tick:** one test covers publishing, the dry run, rerunning safely, and leaving every checkout untouched (`TestProjectTick_PublishesEveryCitingProjectOnOrigin`).
   - **TL-owned projects:** a TL-owned project is skipped unless it opts in (`TestTickOnLanding_TLOwnedIsOptIn`).
   - **Failure path:** a failed tick warns and names the retry command (`TestTickOnLanding_FailureWarnsWithRetry`).
   - **Linked close:** the success path is covered, with the commit-point check (`TestProjectCloseLinkedRecord`), and so is rollback followed by a successful rerun (`TestProjectCloseLinkedRecordRollsBack`).
   - **Gap:** no test exercises the 2-minute tick timeout itself. That is acceptable, because the lock-scope fix was the part that mattered.

6. **Architecture checks**
   - **ARCH-DRY: pass.** The shared helpers are `MVPProgress`, `openIssueRow`/`refIsIssue`, `publishProjectEdit` and `discoverProjectsInScope`.
   - **ARCH-PURE: pass.** The tick edit, MVP classification and `LandingTick` policy are pure functions. The IO is injected (`projectCloseTerminalFn`, `TrunkFile`).
   - **ARCH-PURPOSE: pass.** The automatic landing tick works on both merge paths and across repos, and both `project tick` and `project close` are delivered. Both `project close` gate rows are covered by catalog-matching tests.
   - **ARCH-MOCK: pass.** Tests run against real local git origins through the same `TrunkFile` code path production uses.
   - **ARCH-CONSTRAINTS: pass.** The tick runs outside the lock with a time limit. Finding the projects is bounded by the fleet's root directories.
   - **ARCH-SECURE: pass.** Origin's bytes are re-parsed before editing, and anything that won't parse is refused. An invalid `landing_tick` type is an error, not a guess. The brain repo is never written.
   - **ARCH-ORDER: pass.** The publish is a check-then-write commit on origin, so a rerun changes nothing. The linked close has one explicit commit point, and every earlier failure rolls back.
   - **ARCH-FUNERAL: pass.** The change creates no new durable file family. The tick writes commits to origin, and the linked close moves the link into the existing archive.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      tickOnLanding now runs after f.withLock in both landing.go:609-617 and merge.go:778-783, with a 2m context.WithTimeout that reaches TrunkFile git calls; the tickOutsideLock test asserts origin is untouched at lock release on both merge paths.
  - id: BR-4
    disposition: addressed
    note: |
      commitLinkedProjectClose removes the live link before the rename and restores it on failure; TestProjectCloseLinkedRecordRollsBack forces the rename to fail and checks the link is restored, no archive link is left, and a rerun succeeds; TestProjectCloseLinkedRecord asserts the live link is gone at the commit point.
  - id: BR-7
    disposition: addressed
    note: |
      TestProjectCloseGatesMatchTheCatalog loops over every project close catalog row (no-retro, no-mvp) and checks RefusalPat against the real refusal error and AckPat, via the new assertGatesigAcknowledges, against the real waiver stderr. Rows owned by other commands predate this issue.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f1f1e8c7-fcc1-4370-97a8-00b3dcf00c1d","repository":"github.com/xianxu/ariadne","issue_id":"000373","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/23dee52d3448d231","reviewed_head":"0aa64f88819a74197e75d1b0f6f99a032006ab1b","slot":"ariadne:4","requested_at":"2026-10-15T12:51:29.117694Z"},"phase":"finished","pid":36147,"started_at":"2026-10-15T12:51:29.373872Z","finished_at":"2026-10-15T12:53:46.777512Z","session_id":"19080286-ab7d-4d91-b47c-d6dfa784910b","outcome":"returned"}] -->
