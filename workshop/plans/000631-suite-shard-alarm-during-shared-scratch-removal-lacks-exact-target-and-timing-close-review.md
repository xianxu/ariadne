# Boundary Review — ariadne#631 (whole-issue close)

| field | value |
|-------|-------|
| issue | 631 — Suite shard alarm during shared scratch removal lacks exact target and timing evidence |
| repo | ariadne |
| issue file | workshop/issues/000631-suite-shard-alarm-during-shared-scratch-removal-lacks-exact-target-and-timing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 717667950a336c145042db18b8a563ef886b90ef..4dfc28fbab76faabf373045bc27da5be16a2b988 |
| command | sdlc close --issue 631 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T09:04:31-07:00 |
| verdict | SHIP |

## Review

Reviewing #631's diff; now running the targeted tests and a revert check.

Now the revert check: make `removeTree` behave like a bare `os.RemoveAll` and confirm the new tests go red.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** I found nothing that should block the close, only five minor notes. The diff does what the #631 Spec and Plan say. `removeTree` gives write permission to read-only directories inside a tree and then deletes it, so a dead run folder holding a Go module cache now goes in one pass. Both the sweep and a run's own `Cleanup` use it. The keep-or-take decision moved into a pure `staleTempReason`, and it gives the same answers as the old inline logic, including the rule that a live owner keeps its folder unless it is older than a day (the pid may have been reused). Each removal is logged before it starts and when it ends, with the full path, the reason and the time taken. The capability probe no longer sweeps. TestMain stops tests from sweeping the machine's shared temp roots. I ran the targeted tests at head and they passed. I then made `removeTree` a plain `os.RemoveAll`, and both new regression tests failed. The checkout was restored afterwards, and `git status` showed it clean.

1. **Strengths**
   - `cmd/sdlc/testscratch.go:163-186` — `makeTreeWritable` adds owner permissions as it walks down the tree, before reading each directory. `WalkDir` does not follow symlinks, and `TestSweepStaleTempRemovesReadOnlyTrees` checks that a read-only directory reached through a symlink keeps mode 0555.
   - `staleTempReason` (`testscratch.go:141`) is pure and has a table test (`TestStaleTempReason`) covering the cases from the old `TestSweepStaleTemp`, plus the reused-pid and non-run-folder cases.
   - The sweep checks exclusions against the resolved path (`sweepRootKey`), so `/tmp` and `/private/tmp` count as one root. The test checks both the skip and that TestMain registered every shared root.
   - Dropping the probe's sweep is safe:
     - `prepareScratch`'s roots (RAM disk, TMPDIR, user temp, `/tmp`) include the probe's (RAM disk and `/tmp`).
     - Probe folders are named `sdlc-run-probe-<pid>-<rand>`, which matches `runDirRE`.
     - The probe already deletes its own folder with a `defer`.
   - The localcheck artifacts test now also checks that a protected run folder is logged as kept.

2. **Critical**: none.

3. **Important**: none.

4. **Minor**
   - **Repeated "kept" lines:** a protected folder gets a `[scratch sweep] kept …` line on every run for as long as it stays protected. Consider logging only removals and failures, or only the first time a folder is kept.
   - **The "could not remove" path has no test:** for example, a tree owned by another user where `chmod` fails with a permission error.
   - **Small symlink race (ARCH-SECURE):** a folder could be swapped for a symlink between `WalkDir`'s check and `os.Chmod`, which follows symlinks. This only matters in a world-writable root where another user planted an `sdlc-run-*` folder. The damage is limited to adding owner permissions to our own files, so the risk is low.
   - **Spec wording drift:** Design #2 still names `staleTempCandidates(entries, …)`, but the code ships `staleTempReason` for one entry at a time. The Plan checklist already describes it correctly.
   - **Atlas formatting:** in `atlas/workflow/sdlc-binary.md:2549`, the added sentence leaves an overlong line.

5. **Test coverage**
   - Each regression test fails when the fix is reverted (confirmed here for `removeTree`).
   - The exclusion test fails if the skip guard is removed.
   - The decision table pins the pure logic without any file access.
   - The implementor logged a green pre-close suite (24m13s). I didn't re-run it.

6. **Architecture**
   - ARCH-DRY: pass. There is one sweep per run, and `removeTree` is shared by the sweep and `Cleanup`.
   - ARCH-PURE: pass. The decision is a pure function and `sweepStaleTemp` is a thin layer that does the file work.
   - ARCH-PURPOSE: pass.
     - Every removal now names its target and timing.
     - The root cause (read-only trees) is fixed.
     - Tests no longer sweep the shared roots.
     - The program that produces those trees is a different repo (pair), and the fix there was handed off (pair887).
   - ARCH-MOCK: pass. No external service is involved; the tests use real filesystem fixtures.
   - ARCH-CONSTRAINTS: pass. A dead folder is removed once instead of failing on every sweep, and the scan cost was measured at 20-60 ms.
   - ARCH-SECURE: pass, with the race noted above. `chmod` only touches a tree the sweep has already decided to delete.
   - ARCH-ORDER: pass. The sweep keeps no state between events. The exclusion map is written once in TestMain and only read after that.
   - ARCH-FUNERAL: pass. This change fixes exactly this kind of failure: the run folder's own cleanup could never succeed.
   - ARCH-EVOLVE: N/A. No persisted format changes.
   - **Watch:** the exclusion uses the roots computed when the test process starts. If the RAM disk is unmounted or unreadable at that moment and becomes mounted later in the process, a test could sweep it. That is an edge case, not a defect.

7. **Plan revisions**
   - Add a `## Revisions` entry noting that Design #2 shipped as the per-entry `staleTempReason`, with the loop left in the IO layer, instead of `staleTempCandidates`.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: log-signal-per-event-not-per-run
    title: |
      Protected run folders log a "kept" line on every sweep for as long as they stay protected
    detail: |
      cmd/sdlc/testscratch.go:126 writes a kept line on every prepareScratch call while a runner artifact owner holds the folder; consider logging only removals and failures, or only the first time a folder is kept.
  - id: new
    severity: Minor
    kind: defect
    family: error-path-untested
    title: |
      The sweep's "could not remove" log path has no test
    detail: |
      No fixture covers a removeTree failure (for example a tree owned by another user where chmod fails with a permission error), so that log line and the continue that follows it are unpinned.
  - id: new
    severity: Minor
    kind: defect
    family: symlink-toctou-chmod
    title: |
      makeTreeWritable can follow a symlink swapped in between WalkDir's check and os.Chmod
    detail: |
      This only matters in a world-writable root where another user planted an sdlc-run-* folder; the damage is limited to adding owner permissions to the caller's own files, so the risk is low. Using fchmodat with AT_SYMLINK_NOFOLLOW, or an open plus fchmod, would close it.
  - id: new
    severity: Minor
    kind: defect
    family: plan-spec-drift
    title: |
      Spec Design #2 names staleTempCandidates, but the code ships a per-entry staleTempReason
    detail: |
      The Plan checklist already reflects the change; add a Revisions entry so the Spec stops naming the batch function.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"27259333-3e48-4f1d-9b1a-2a2116bbda4a","repository":"github.com/xianxu/ariadne","issue_id":"000631","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/c98eb8f24de31318","reviewed_head":"4dfc28fbab76faabf373045bc27da5be16a2b988","slot":"ariadne:5","requested_at":"2026-10-25T16:02:45.687909Z"},"phase":"finished","pid":2984,"started_at":"2026-10-25T16:02:45.788802Z","finished_at":"2026-10-25T16:04:29.344739Z","session_id":"dc1ed31a-cb69-478c-9c04-3bdcb43fe2ae","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T09:11:48-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 631 — Suite shard alarm during shared scratch removal lacks exact target and timing evidence |
| repo | ariadne |
| issue file | workshop/issues/000631-suite-shard-alarm-during-shared-scratch-removal-lacks-exact-target-and-timing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 931efc51cdebab13c6ff8ff4c4f37c6ea59bf90e..6fc21e0fc04e6f8650062b16a00c5649d297154d |
| command | sdlc close --issue 631 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T09:11:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Now test the no-read-bit directory case empirically, plus run the targeted tests and the regression revert for BR-2.

Confirmed: a directory with no read bit now can't be removed. Now checking that the base version handled that case and that BR-2's test actually fails without the `continue`.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** Of the four prior Minors, three are fixed and one was reasonably declined. BR-2 now has a regression test, and it does fail when the fix is undone. BR-4 is recorded as a Revisions entry, and that entry only records the fix; it does not change scope. BR-1 was declined with a sound reason. The BR-3 fix, though, breaks something that worked before. `makeTreeWritable` now opens each directory with `O_RDONLY` before changing its permissions, and that open needs the read bit on the directory itself. So `removeTree` can no longer take a tree containing a directory that lacks owner read, such as mode 0o311 or 0o000. The base version, which used path-based `os.Chmod`, removed those fine. I ran both versions on a probe test in this checkout to confirm it. Module-cache trees (0o555) still work, so the root cause the issue is about stays fixed. The new failure is a narrower case that no test covers. It is cheap to fix but should be fixed before shipping.

### 1. Strengths
- `cmd/sdlc/testscratch_darwin_test.go:16`: `chflags uchg` makes a removal fail for real, with no mocks. It pins both the "could not remove" log line and the `continue` that follows. When I changed `continue` to `return`, the test failed as expected.
- `cmd/sdlc/testscratch.go:193-201`: switching to the early-return form makes the permission check easy to read, and `errors.Join(f.Chmod…, f.Close())` keeps both errors instead of dropping one.
- The BR-1 reply (issue Log, 2026-10-25) holds up against the issue text. Spec Design #3 (line 58-59) says "A protected candidate is logged as kept", and the Plan row at line 84 lists "kept: protected". The sweep runs once per run, and the shared roots excluded via `testmain_test.go:193` are never swept from tests. Removing the line would break the Spec.

### 2. Critical findings
None.

### 3. Important findings
- **`cmd/sdlc/testscratch.go:197`: the `O_RDONLY` open fails with EACCES on a directory with no owner read bit, so `removeTree` now leaves trees it used to remove.** I checked this with a probe in the disposable checkout. With a `sub` directory at mode 0o311 or 0o000, the head returns `open …/sub: permission denied` and the tree survives; the base returns nil and the tree is gone. WalkDir calls the callback on a directory before reading it, so the old path-based chmod fixed the bits first. The new handle-based chmod cannot open the directory to begin with.
  - **Fix sketch:** on darwin, use `unix.Fchmodat(AT_FDCWD, p, perm|0o700, AT_SYMLINK_NOFOLLOW)`. Alternatively, open with `O_SEARCH`/`O_PATH` where the platform supports `fchmod` on such a handle. Or, if the open fails with EACCES, fall back to Lstat, confirm it is still a directory and not a symlink, then `os.Chmod`.
  - **Test:** add a regression case to `TestSweepStaleTempRemovesReadOnlyTrees` with a 0o311 or 0o000 directory.
  - **ARCH:** this is a correctness issue, not an architecture one.

### 4. Minor findings
- With `O_NOFOLLOW`, only the last part of the path is protected. A parent directory swapped for a symlink mid-walk would still be followed. Closing that fully would need an `openat` walk. Given BR-3's threat model, that residue is acceptable, and I've noted it under BR-3's disposition rather than raising it again.

### 5. Test coverage notes
- These targeted tests pass at the head: TestSweepStaleTemp*, TestPrepareScratch*, TestStaleTempReason, TestExcludedScratchRootIsNeverSwept and TestSweepStaleTempLogsAFailedRemovalAndContinues.
- The new failure-path test is darwin-only because of `chflags`. That's acceptable since the package is already Unix-only. `GOOS=linux go vet` is clean.
- No test covers directories without a read bit. That gap is how the Important finding got through.

### 6. Architectural notes
- **ARCH-DRY: pass.** Nothing is duplicated.
- **ARCH-PURE: pass.** The change stays inside the IO shell (`makeTreeWritable`). The pure `staleTempReason` is untouched.
- **ARCH-PURPOSE: pass**, with the caveat that the BR-3 fix made the tree-repair function work in fewer cases than before; it is now partial. The class BR-3 names (a symlink followed during chmod) is handled at the one chmod site, and there are no other chmod sites in the diff.
- **ARCH-EVOLVE: pass.** No persisted format changed.

### 7. Plan revision recommendations
None beyond the Revisions entry the fix added. It matches the code: `staleTempReason` is at `testscratch.go:140`.

```findings
dispose:
  - id: BR-1
    disposition: withdrawn
    note: |
      The Spec requires the kept line (Design 3, lines 58-59; Plan line 84), and production sweeps once per run, so it is one line per run, not per prepareScratch call.
  - id: BR-2
    disposition: addressed
    note: |
      TestSweepStaleTempLogsAFailedRemovalAndContinues pins the log line and the continue; changing continue to return in this checkout makes it FAIL.
  - id: BR-3
    disposition: addressed
    note: |
      The chmod now goes through an O_NOFOLLOW|O_DIRECTORY handle (testscratch.go:197), so a swapped final path component fails. A swapped parent is still followed, which is acceptable at this threat level. The read-bit regression this fix introduced is raised separately.
  - id: BR-4
    disposition: addressed
    note: |
      The Revisions entry names staleTempReason(name, mtime, runDirs, now, alive), which matches testscratch.go:140; it records the fix and does not change scope.
findings:
  - id: new
    severity: Important
    kind: defect
    family: tree-repair-coverage-regression
    title: |
      makeTreeWritable's O_RDONLY open fails on directories without owner read, so removeTree now leaves trees the base removed
    detail: |
      testscratch.go:197 opens each directory O_RDONLY before fchmod, and the open needs the read bit. Probe in this checkout: with a subdirectory at mode 0o311 or 0o000, the head returns "open .../sub: permission denied" and the tree survives; the base, using path-based os.Chmod, removes it. Fix: Fchmodat with AT_SYMLINK_NOFOLLOW on darwin, or fall back to an Lstat-checked os.Chmod on EACCES. Add a 0o311 or 0o000 case to TestSweepStaleTempRemovesReadOnlyTrees.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"176a21e4-2f82-4d37-a319-fd3f23e398c7","repository":"github.com/xianxu/ariadne","issue_id":"000631","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/a59b87d2df75d3f7","reviewed_head":"6fc21e0fc04e6f8650062b16a00c5649d297154d","slot":"ariadne:5","requested_at":"2026-10-25T16:09:25.553747Z"},"phase":"finished","pid":15129,"started_at":"2026-10-25T16:09:25.727472Z","finished_at":"2026-10-25T16:11:45.628097Z","session_id":"b8c8c6ac-13a7-4dc8-9dfa-8722326ebc6d","outcome":"returned"}] -->
