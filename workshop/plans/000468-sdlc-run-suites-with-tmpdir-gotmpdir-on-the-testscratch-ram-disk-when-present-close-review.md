# Boundary Review — ariadne#468 (whole-issue close)

| field | value |
|-------|-------|
| issue | 468 — sdlc: run suites with TMPDIR/GOTMPDIR on the TestScratch RAM disk when present |
| repo | ariadne |
| issue file | workshop/issues/000468-sdlc-run-suites-with-tmpdir-gotmpdir-on-the-testscratch-ram-disk-when-present.md |
| boundary | whole-issue close |
| milestone | — |
| window | 931a4ad6898a9883e1324701e92860bca401f2bc..f0927a34254769c8179bd8aa501cc88587ff3398 |
| command | sdlc close --issue 468 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T11:10:16-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've read the full diff and the issue. Now I'm checking the local-check names in the repo config against the run-folder regex.

The package tests pass. I'm running the new `cmd/sdlc` scratch tests now, then I'll write up the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. `runUnderSlot` is the single place every suite starts, and it now points each run's `TMPDIR` and `GOTMPDIR` at its own folder on the RAM disk. The mount check, the 2 GiB floor and the fallback note live in one shared package, `internal/scratchdisk`, which both the local-check code and `reviewbox` use. The run's own checkout and review checkouts follow the same rule, and when the disk refuses the write they fall back, with tests for both. The added tests passed on the pinned head: the scratchdisk, reviewbox and golint packages, plus the six new `cmd/sdlc` scratch tests. Two things stand between this and SHIP. First, the sweep for leftover run folders never matches a folder made by a check whose name contains a hyphen, and check-name validation allows hyphens. So a crashed run of such a check leaves its folder on the RAM disk forever. Second, the leak warning is missing the sizes and the full list of entry names that Revision 2a promised.

1. **Strengths**
   - `cmd/sdlc/internal/scratchdisk/scratchdisk.go`: one probe and one pure `Choose(Probe, floor)`, shared by `localcheck_run.go` and `reviewbox.Roots` (ARCH-DRY). The mount test compares the device of `/Volumes/TestScratch` with that of `/Volumes`. That means an unmounted path never sends writes to the SSD.
   - `Roots` lists the RAM disk's folder for sweeping whenever the disk is mounted and readable, even when it is too full to take new work. A checkout's folder is fixed when it is created, so nothing is ever moved halfway (ARCH-FUNERAL).
   - The fallback when the disk refuses the write is tested end to end. `TestMkdirTempFallsBack` covers it, and so does `TestReviewIsolationFallsBackFromScratch`, which redirects `HOME` so it never touches the real cache (ARCH-SECURE).
   - `TestRunUnderSlotUsesScratch` checks what the child process actually sees in its environment, that its checkout is in the right place, and that its run folder is removed afterwards. It covers both the scratch and the off modes.
   - The `dev-aliases.test.sh` fix explains the HFS+ one-second timestamp cause and ages the binary explicitly.

2. **Critical findings**
   - None.

3. **Important findings**
   - `cmd/sdlc/testscratch.go:53`: the pattern `runDirRE = ^[A-Za-z0-9_.]+-(\d+)-\d+$` doesn't allow `-` in the check name. But `localchecks.go:77` accepts names matching `^[a-z0-9][a-z0-9-]{0,40}$`. A run folder such as `go-lint-123-456` therefore never matches, so it isn't swept for a dead process or for age. The go-build rule doesn't catch it either. Once a crash or SIGKILL leaves one behind, it stays on the 16 GB RAM disk for good (ARCH-FUNERAL). Downstream repos that declare hyphenated checks hit this.
     - Fix: anchor the process id at the end, e.g. `^.+-(\d+)-\d+$`. The greedy match takes the last pid.
     - Test: add a hyphenated case to `TestSweepStaleTemp`.
   - `cmd/sdlc/testscratch.go:201-214` (`reportScratchLeaks`): Revision 2a says the warning "reports the count and size, and each entry's name". Plan step 2a also says "plus sizes". The code reports counts only, plus one example path per package, and never measures size. Either add size and the full list, or add a Revisions entry that narrows the contract to what was built.

4. **Minor findings**
   - `reviewisolation.go:42`: the fallback retries on any `Create` error, including a bad SHA or a git failure, and the warning blames the RAM disk. It would be cleaner to retry only on write or permission errors, or at least to word the warning neutrally.
   - `plannedReviewIsolation` (used by `--dry-run`) reports the RAM disk location even when the real run would fall back. That's acceptable for a dry run.
   - `declaredTests(wt)` walks and reads every `*_test.go` in the worktree after each scratched run. That's fine at today's scale, but worth noting.
   - The atlas paragraph in `sdlc-binary.md` has one very long reflowed line.

5. **Test coverage notes**
   - The pure parts are tested without IO: `Choose`, `scratchLeaks`, `attributeTempDir`, and `sweepStaleTemp` with an injected `alive` function and clock.
   - The run folders in the sweep test all use the non-hyphenated name `suite-…`, which is exactly why the regex gap got through.
   - Nothing tests the low-free-space path through `Roots`. `Choose` covers it as a pure function, so this is acceptable.

6. **Architecture notes**
   - ARCH-DRY: pass. The probe was moved into `scratchdisk` and is shared.
   - ARCH-PURE: pass. `Choose` and `scratchLeaks` are pure, and the IO is a thin layer around them.
   - ARCH-PURPOSE: pass. The issue lists three surfaces: suite temp files, the run's own checkout, and review checkouts. All three derive from the one rule, and the lint is widened.
   - ARCH-MOCK: pass. `WF_TEST_SCRATCH` is the seam, and the live check is recorded in the Log.
   - ARCH-CONSTRAINTS: pass. The 2 GiB floor is enforced, and the disk-write sample is documented.
   - ARCH-SECURE: pass. The environment override names only a directory, there are no credentials, and the tests isolate `HOME`.
   - ARCH-ORDER: pass. No state is held between events except folders on disk, and those are reconciled by the sweep. The pid check is backed by a 24-hour age limit in case a pid is reused.
   - ARCH-FUNERAL: **flag**. See the hyphenated-name sweep gap above. Everything else has a removal path: run folders are dropped by a defer, stale ones by the sweep, and checkouts are swept on both roots.
   - ARCH-EVOLVE: N/A. No persisted format changes.

7. **Plan revision recommendations**
   - If the leak warning stays count-only, add a Revisions entry saying 2a delivers counts by package plus one example path, with no sizes, replacing "count and size, and each entry's name".

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: sweep-pattern-narrower-than-writer
    title: |
      The run-folder sweep never matches folders made by hyphenated check names, so a crashed run's folder stays on the RAM disk for good
    detail: |
      testscratch.go:53 runDirRE requires the check name to match [A-Za-z0-9_.]+, but localchecks.go:77 allows hyphens ([a-z0-9-]). prepareScratch names a folder <check>-<pid>-<rand>, so a crashed run of a check such as go-lint leaves a folder that neither the dead-pid nor the 24 h age rule matches (ARCH-FUNERAL). Fix: ^.+-(\d+)-\d+$, plus a hyphenated case in TestSweepStaleTemp.
  - id: new
    severity: Important
    kind: defect
    family: revision-contract-drift
    title: |
      The leak warning omits the sizes and full entry names that Revision 2a promised
    detail: |
      Revision 2a says the warning reports the count and size, and each entry's name; plan step 2a says "plus sizes". reportScratchLeaks (testscratch.go:201) prints counts and one example per package, and never measures size. Either deliver size and the full list, or add a Revisions entry narrowing the contract.
  - id: new
    severity: Minor
    kind: defect
    family: fallback-on-unrelated-error
    title: |
      The review-checkout fallback retries on any Create error and blames the RAM disk
    detail: |
      reviewisolation.go:42 retries on the cache root for any error, such as a bad SHA or a git failure, and the warning says the RAM disk failed. Retry only on write or permission errors, or reword the warning neutrally.
```

A background command I started earlier was stopped when it hit its time limit, and this doesn't change the review. It was a regex check with a `go run -` mixed in by mistake, and the `go run` part hung. I reran the regex check on its own and it finished: `suite-rerun-123-456` and `go-lint-1-2` don't match the run-folder pattern, while `suite-123-456` does. That result is the evidence for the hyphenated-name sweep finding. I'm not restarting the stopped command, since nothing still depends on it.

The verdict is still **FIX-THEN-SHIP**, with the findings as reported.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"43236dff-a7c3-44e8-a00f-4975cfc789c5","repository":"github.com/xianxu/ariadne","issue_id":"000468","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/08489f19f6d85445","reviewed_head":"f0927a34254769c8179bd8aa501cc88587ff3398","slot":"ariadne:4","requested_at":"2026-10-19T17:57:22.340162Z"},"phase":"finished","pid":25536,"started_at":"2026-10-19T17:57:22.806844Z","finished_at":"2026-10-19T18:10:15.680014Z","session_id":"d5248414-85a1-4cb3-95dd-0f538e3d2a08","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T11:15:16-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 468 — sdlc: run suites with TMPDIR/GOTMPDIR on the TestScratch RAM disk when present |
| repo | ariadne |
| issue file | workshop/issues/000468-sdlc-run-suites-with-tmpdir-gotmpdir-on-the-testscratch-ram-disk-when-present.md |
| boundary | whole-issue close |
| milestone | — |
| window | 931a4ad6898a9883e1324701e92860bca401f2bc..a1145892568eb97974a4f73a4fd52e3bb3b65158 |
| command | sdlc close --issue 468 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T11:15:16-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Prior findings look addressed on read; now checking the localcheck seam and running the tests, including a revert check for BR-1.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All three round-1 fixes are in. I checked each against the code and the tests rather than the commit message. I reverted the BR-1 fix in this checkout and `TestSweepStaleTemp` failed: it left `go-lint-100-4` on disk. I then restored the file. BR-2 now delivers sizes and full names up to a cap, and the plan's new close-round-1 Revisions entry narrows the contract to match. BR-3's narrowing is correct, but no test fails without it, so I keep it open as Minor. The one new item is also Minor and does not block. The widened sweep pattern is now broader than the names `prepareScratch` writes. Run in this checkout, the tests for scratchdisk, reviewbox and golint passed, as did the targeted `cmd/sdlc` scratch, leak, review-fallback and local-worktree tests.

1. **Strengths**
   - `cmd/sdlc/internal/scratchdisk/scratchdisk.go:30` keeps the decision pure (`Choose`) and the probe (`Read`) separate. `Roots` gives both callers the same rule: create on the RAM disk only with room, sweep both roots whenever the disk is mounted (ARCH-DRY, ARCH-FUNERAL).
   - The mount check (`Dev` differs from the parent folder) means an unmounted `/Volumes/TestScratch` path is never written to the SSD by mistake.
   - `sweepStaleTemp` takes `now` and `alive` as parameters, so the dead-pid and age rules are tested without real processes.
   - `reviewisolation_scratch_test.go` exercises the real fallback end to end: a RAM root you can read but not write, with HOME isolated.
   - The `dev-aliases` fix for HFS+'s one-second file times targets the actual cause the log describes.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **The sweep pattern is now wider than its writer.** `cmd/sdlc/testscratch.go:56` `^.+-(\d+)-\d+$` matches any `x-<n>-<n>` folder in `/Volumes/TestScratch/tmp`. AGENTS.base.md now invites agents to use that disk, so a foreign folder such as `pr-468-2` would be removed if pid 468 is dead. See the finding below for the rule.
   - **BR-3 (still open):** no test shows that a non-permission `Create` error (bad head, git failure) skips the fallback. The existing test passes both with and without the `errors.Is` narrowing.
   - `treeSizes` has no direct test. It is simple, best-effort code.
   - `atlas/workflow/sdlc-binary.md:1931` has one over-long line (wrap nit).

5. **Test coverage notes**
   - BR-1 is pinned by the hyphenated cases, and the revert check went red.
   - BR-2 is pinned by the size, name and "… N more" assertions.
   - BR-3 lacks a negative case.
   - No test covers a foreign folder in the run-folder root, which is the gap behind the new finding.

6. **Architecture notes**
   - ARCH-DRY: pass. `scratchdisk` is shared by all three callers.
   - ARCH-PURE: pass. `Choose`, `scratchLeaks` and `attributeTempDir` are pure; IO sits at the edges.
   - ARCH-PURPOSE: pass. Suite temp, the suite's own checkout and review checkouts all move to the RAM disk; there was a live check and an A/B disk-write sample.
   - ARCH-MOCK: pass. The `WF_TEST_SCRATCH` seam stands in for the disk in fixtures, and the live run checked the real mount.
   - ARCH-CONSTRAINTS: pass. 2 GiB floor; sizes are measured only for attributed entries.
   - ARCH-SECURE: flag (Minor). The sweep acts on names it did not write in a shared folder; see the finding.
   - ARCH-ORDER: pass. Nothing is carried between events except the run folder's lifetime, keyed on pid plus age.
   - ARCH-FUNERAL: pass. Run folders, checkouts and review checkouts are swept on every root.
   - ARCH-EVOLVE: N/A. No persisted format changes.

7. **Plan revision recommendations:** none needed. If the run-folder prefix changes, add one line to the plan's "Per run" step.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      runDirRE now ^.+-(\d+)-\d+$; reverting it makes TestSweepStaleTemp fail on go-lint-100-4 (verified in this checkout).
  - id: BR-2
    disposition: addressed
    note: |
      reportScratchLeaks gives per-package and total size and every name up to leakListCap; TestScratchLeaksAttributesOnlyThisRun pins them; a Revisions entry narrows sizing to attributed entries.
  - id: BR-3
    disposition: not-addressed
    note: |
      The narrowing and the neutral wording are in, but no test shows a non-permission Create error skips the fallback; TestReviewIsolationFallsBackFromScratch passes with or without the errors.Is guard.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: sweep-pattern-narrower-than-writer
    title: |
      The run-folder sweep pattern is now wider than prepareScratch's names, so it can delete foreign folders on the shared RAM disk
    detail: |
      This is the 2nd finding in family sweep-pattern-narrower-than-writer. The rule covering both: a sweep's match must be built from the same constant as its writer, neither narrower nor wider. The local-check sweep already does this with localCheckTmpPrefix. testscratch.go:56 ^.+-(\d+)-\d+$ matches any x-N-N folder under /Volumes/TestScratch/tmp, which AGENTS.base.md now advertises for general scratch work, and removes it when the pid looks dead or it is a day old. Fix: one shared prefix constant (e.g. sdlc-run-) used by both the MkdirTemp at testscratch.go:44 and the sweep regex, plus a test case with a foreign pr-468-2 folder that must survive.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d2b0c5ce-34ec-4758-a9dd-c767d2e00569","repository":"github.com/xianxu/ariadne","issue_id":"000468","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/627dce757f9e23c8","reviewed_head":"a1145892568eb97974a4f73a4fd52e3bb3b65158","slot":"ariadne:4","requested_at":"2026-10-19T18:12:59.679767Z"},"phase":"finished","pid":77871,"started_at":"2026-10-19T18:12:59.994785Z","finished_at":"2026-10-19T18:15:16.049884Z","session_id":"05596eb7-026a-4065-afd1-f27e82c9e77b","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T11:18:52-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 468 — sdlc: run suites with TMPDIR/GOTMPDIR on the TestScratch RAM disk when present |
| repo | ariadne |
| issue file | workshop/issues/000468-sdlc-run-suites-with-tmpdir-gotmpdir-on-the-testscratch-ram-disk-when-present.md |
| boundary | whole-issue close |
| milestone | — |
| window | 38d94c4e606d5d13e928c2de146901fb8be8b4c7..9e4c4f89b69e712e4ad6bd608f3416530827c64a |
| command | sdlc close --issue 468 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T11:18:52-07:00 |
| verdict | SHIP |

## Review

Backup failed; restore via git.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and each fix has a regression test that I confirmed goes red when the fix is reverted. On BR-4, run folders and the sweep now share one constant: `scratchRunPrefix` (`cmd/sdlc/testscratch.go:57`) is used both when the folder is made (`:44`) and in the sweep pattern (`:62`). A new test case has another tool's folder, `other-tool-100-6`, and it must survive the sweep. On BR-3, the review-checkout fallback in `cmd/sdlc/reviewisolation.go:46` retries only on `fs.ErrPermission` or `EROFS`. A new test, `TestReviewIsolationKeepsOtherErrors`, opens a review with a missing head and checks that it fails without falling back. The change touches only three files, adds no new behavior, and makes no scope change to Spec, Revisions or Done-when. It introduces no defects.

1. **Strengths**
   - `cmd/sdlc/testscratch.go:57-62`: the run-folder name and the sweep pattern now come from the same constant, and `regexp.QuoteMeta` keeps the prefix literal. That is the rule the family `sweep-pattern-narrower-than-writer` asked for, applied to the whole class rather than one site.
   - `cmd/sdlc/testscratch_test.go:38`: the foreign folder is aged past a day, so it would be swept by either the dead-pid path or the age path. The test therefore pins the prefix itself, not one path.
   - `cmd/sdlc/reviewisolation_scratch_test.go:55-72`: the missing-head case runs through the real `openReviewIsolation` and real git, with no mocks.
   - `cmd/sdlc/testscratch_test.go:155`: the integration test also checks for the prefix, so the code that makes the folder is tied to the constant from the other end.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new.

5. **Test coverage notes**
   - The targeted tests pass at head: `go test -run 'TestReviewIsolation|TestSweepStaleTemp|TestRunUnderSlotUsesScratch|Scratch'` → ok.
   - **BR-4 revert check:** with the prefix taken out of both the folder name and the pattern, `TestSweepStaleTemp` fails. Its output is `left [... ] want [... other-tool-100-6 ...]`.
   - **BR-3 revert check:** with the condition loosened back to `root != fallback`, `TestReviewIsolationKeepsOtherErrors` fails with the message "fell back on a non-write error", and the output shows the old misleading "refused the review checkout" warning.
   - I restored both files after the revert checks and confirmed the tree is clean.

6. **Architectural notes**
   - **ARCH-DRY: pass.** The prefix is defined once and used by the code that makes folders, the sweep, and the test.
   - **ARCH-PURE: pass.** `sweepStaleTemp` still gets the clock and the liveness check injected, and the test stays free of IO beyond a temp folder.
   - **ARCH-PURPOSE: pass.** The fix applies the rule from the finding (the sweep is built from the same constant as the writer) to every place run folders are matched. `runDirRE` has only one consumer, `testscratch.go:84`.
   - **ARCH-EVOLVE: pass, nothing to check.** Run-folder names are temporary scratch, not a persisted format. Folders from before the prefix existed are never released state, because the issue has not shipped.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      reviewisolation.go:46 falls back only on fs.ErrPermission/EROFS; TestReviewIsolationKeepsOtherErrors goes red when reverted to the old any-error condition.
  - id: BR-4
    disposition: addressed
    note: |
      scratchRunPrefix (testscratch.go:57) is shared by MkdirTemp (:44) and runDirRE (:62); the foreign other-tool-100-6 case goes red when the prefix is removed.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2f8fbf2c-0a4e-43c1-a579-b710a1c9a521","repository":"github.com/xianxu/ariadne","issue_id":"000468","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c8a9f1459c47543c","reviewed_head":"9e4c4f89b69e712e4ad6bd608f3416530827c64a","slot":"ariadne:4","requested_at":"2026-10-19T18:17:20.578651Z"},"phase":"finished","pid":79644,"started_at":"2026-10-19T18:17:20.6951Z","finished_at":"2026-10-19T18:18:51.646812Z","session_id":"a726d180-d42d-412d-adbb-988a7efef389","outcome":"returned"}] -->
