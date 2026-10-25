---
id: 000631
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '06ae828085c627a72f8559230cccb82ccabe1784' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T08:18:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "90e3e2cf", done: "79bd4978"}
---

# Suite shard alarm during shared scratch removal lacks exact target and timing evidence

## Problem

A recorded #624 pre-close suite reached the 30-minute shard alarm while TestSysVObservationPreservesCallerValuesWithIndependentBound/true was inside prepareScratch -> sweepStaleTemp -> os.RemoveAll. This is an incomplete suite, not a registered flake or valid pass. Diagnose synchronous stale-root cleanup and filesystem progress before retry policy changes.

The alarm captures the stack at the deadline; it does NOT prove RemoveAll itself had run for all 30 minutes. Its entry time and exact target basename were not logged.

## Spec

Record and investigate which stale cleanup target can hold a managed run's setup, whether filesystem removal stalls or normal shared-root work consumes the package deadline, and how setup cleanup can remain safe and bounded. Preserve live-owner and retained-artifact protection. Do not delete live or unknown peer roots or raise test deadlines as the diagnosis.

**Root cause (measured 2026-10-25 08:1x, load 28-46; production scratch only read).** `/Volumes/TestScratch/tmp`
held 41 `sdlc-run-*` folders: 3 live, 38 with a dead owner. 21 of the dead ones are suite run folders of about 6,000
entries and 77 MB each, from 10-24 05:39 on, about one more per pair suite. Each holds `pair-launcher-*/go/pkg/mod`,
a Go module cache whose 566 directories are read-only (pair's `testcouch/*_build_test.go` and
`testdisk/cache_conformance_test.go` set GOMODCACHE under TMPDIR). `os.RemoveAll` cannot unlink inside a read-only
directory, so the run's own Cleanup failed and so does every later sweep. On a `cp -Rp` copy: one attempt 130-160 ms,
0 of 6,063 entries removed, the same on every retry; across the set about 3 s per sweep, growing. The scan itself is
cheap (ReadDir plus Info of /tmp's 5,922 entries: 20-60 ms).

Production sweeps once per recorded run (`prepareScratch`), plus a second, redundant sweep in the host capability
probe (`suiteHostCapabilities`, #572). The cost lands in tests: about 270 `cmd/sdlc` test call sites reach
`runUnderResourceBatch` -> `prepareScratch`, and every one sweeps the machine's shared roots (`scratchdisk.Roots`:
the RAM disk, TMPDIR, the user temp folder, /tmp), so all 8 shards rescan and refail the same trees concurrently.
That is the stack #624 and #616 caught at the shard alarm (ARCH-FUNERAL: the run folder's end was a removal that
could not succeed, and nothing noticed).

**Design.**
1. `removeTree` (testscratch.go): RemoveAll; on failure, make each directory in the tree owner-writable (WalkDir, so
   a symlink is never followed or chmodded) and RemoveAll again. The sweep and a run's own Cleanup both use it, so a
   dead folder goes once.
2. The decision is pure: `staleTempCandidates(entries, runDirs, now, alive)` returns each candidate with its reason
   (`go-build older than 24h`, `owner pid N exited`, `older than 24h`); `sweepStaleTemp` is the IO shell (read,
   protection check, remove, log) (ARCH-PURE). Live-owner and `runnerArtifactProtected` rules are unchanged.
3. Evidence: before each removal the sweep writes the root, the candidate and its reason, then the duration and
   result, to the run's stderr (the suite log), so a stall's stack has its target named in the log. A protected
   candidate is logged as kept.
4. One sweep per run: the capability probe stops sweeping; `prepareScratch`'s roots are a superset of the probe's
   (RAM disk and /tmp) and its run-folder pattern matches `sdlc-run-probe-*` (ARCH-DRY).
5. Tests never sweep the machine's shared roots: TestMain registers the roots `scratchdisk.Roots` resolves at
   process start (resolved through symlinks) with `excludeScratchSweep`, and the sweep skips them. A fixture root a
   test makes (t.TempDir, WF_TEST_SCRATCH, SDLC_USER_TEMP_FIXTURE) is still swept, so the existing sweep tests stand.

ARCH-CONSTRAINTS: batch setup on a loaded host; budget a sweep at under 100 ms when nothing is stale (measured
20-60 ms scan), one removal per dead folder ever. ARCH-SECURE: the chmod is confined to a tree the sweep already
decided to delete, under the prefix rule; symlinks are skipped. ARCH-ORDER: holds no state between events; a
concurrent sweep of the same candidate races only RemoveAll, which tolerates a vanished path. ARCH-EVOLVE: no
persisted format. ARCH-MOCK: no external service; filesystem fixtures.

## Done when

- Capture the cleanup root, exact candidate, time spent and liveness/protection decision before removal, so a timeout names its source.
- Reproduce with an isolated writable fixture (including slow or interrupted removal as appropriate), establish the root cause and test the correction without sweeping production scratch with fake liveness.
- Verify live/retained roots remain protected; run the recorded suite through the existing queue.

## Plan

- [x] Failing test first: a fixture run folder with a read-only module-cache tree survives `sweepStaleTemp` and
  `scratchPreparation.Cleanup` on main.
- [x] `removeTree`; the sweep and Cleanup use it; symlink and protected-folder cases.
- [x] Pure `staleTempReason` with reasons (per entry; the shell keeps the loop); table test covering today's `TestSweepStaleTemp` cases.
- [x] Sweep log lines (before: root, candidate, reason; after: duration, result; kept: protected).
- [x] Drop the capability probe's sweep.
- [x] `excludeScratchSweep` plus TestMain registration; a test that a registered root is never swept.
- [ ] Targeted tests (testscratch, suitecapability, localcheck artifacts), then the recorded suite through the queue.

## Log

2026-10-25: filed unclaimed on ops:0 instruction d2ab6974-89ff-4d6c-9a9d-1783efd0025b. Candidate Git head 5724aaba9338a40537db4841a9e11e762fc73bd8, attempt d27b2c86f3f1abbc6864ec1451e9738c, exact saved log SHA256 bdb8d31e93054c9ec57539e5917a86f1479f71b5441aaa203ff09c7190873462. Admitted 05:01:34 PDT; finished 05:31:40, wall1803.71s. Shard3 exit1 at1800.4s; 22 listed tests did not run. Other failures were the candidate's context guard and the independently owned #622/#625 train-barrier family.

At05:30 host uptime reported load15.49/20.83/29.62 (1/5/15 minute averages). TestScratch was the admitted run's scratch facility. The stack shows sweep root length0x18 (24 bytes), matching /Volumes/TestScratch/tmp; neither this SysV test nor localRunRepo overrides WF_TEST_SCRATCH. This is the inferred shared TestScratch root, not an observed private fixture. Gap: the removed target's full path is NOT recoverable from the saved stack: it prints pointer+length0x38 (56 bytes), not string contents. Do not invent its basename. Temporary raw per-shard JSON was removed by scripts/test-shard.py's TemporaryDirectory; retained failed log contains this stack and attempt metadata. No cleanup, signals or retries against that root were performed for diagnosis.

Retained local log: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/d27b2c86f3f1abbc6864ec1451e9738c.log. Stack excerpt:

```text
os.removeAll({0x33016e38d800, 0x38})
	os/removeall_at.go:56 +0x1c4
os.RemoveAll(...)
	os/path.go:74
github.com/xianxu/ariadne/cmd/sdlc.sweepStaleTemp({0x33016dda83c0, 0x18}, 0x1, {0x38?, 0x33016e192ff8?, 0x1043343a0?}, 0x1042abf80)
	github.com/xianxu/ariadne/cmd/sdlc/testscratch.go:120 +0x224
github.com/xianxu/ariadne/cmd/sdlc.prepareScratch({0x10429b118, 0x33016e1ec270}, {0x1036166ce, 0x6}, 0x0)
	github.com/xianxu/ariadne/cmd/sdlc/testscratch.go:39 +0xc0
github.com/xianxu/ariadne/cmd/sdlc.runUnderResourceBatch({_, _}, {0x0, {0x10429b2b8, 0x33016de742a0}, 0x0, 0x0, 0x0, 0x0, 0x0, ...}, ...)
	github.com/xianxu/ariadne/cmd/sdlc/localcheck_run.go:641 +0x1300
github.com/xianxu/ariadne/cmd/sdlc.runUnderResource({_, _}, {0x0, {0x10429b2b8, 0x33016de742a0}, 0x0, 0x0, 0x0, 0x0, 0x0, ...}, ...)
	github.com/xianxu/ariadne/cmd/sdlc/localcheck_run.go:490 +0x120
github.com/xianxu/ariadne/cmd/sdlc.runLocalCheck.func3({0x0?, 0x33016ddc8418?})
	github.com/xianxu/ariadne/cmd/sdlc/localcheck_run.go:317 +0x15c
github.com/xianxu/ariadne/cmd/sdlc.runSuiteLanes(0x0, {0x10429b118, 0x33016e1ec270}, 0x33016e1956e0)
	github.com/xianxu/ariadne/cmd/sdlc/suitelane.go:128 +0x24c
github.com/xianxu/ariadne/cmd/sdlc.runLocalCheck({_, _}, {0x0, {0x10429b2b8, 0x33016de742a0}, 0x0, 0x0, 0x0, 0x0, 0x0, ...}, ...)
	github.com/xianxu/ariadne/cmd/sdlc/localcheck_run.go:314 +0x10d4
github.com/xianxu/ariadne/cmd/sdlc.TestSysVObservationPreservesCallerValuesWithIndependentBound.func1(0x33016de2f688)
	github.com/xianxu/ariadne/cmd/sdlc/sysvwarning_test.go:305 +0x2e0
testing.tRunner(0x33016de2f688, 0x33016db96ac0)
	testing/testing.go:2195 +0xc4
created by testing.(*T).Run in goroutine 187763
	testing/testing.go:2260 +0x3b8

goroutine 187860 [select]:
os/signal.NotifyContext.func1()
	os/signal/signal.go:304 +0x74
created by os/signal.NotifyContext in goroutine 187846
	os/signal/signal.go:303 +0x144

goroutine 187763 [chan receive]:
testing.(*T).Run(0x33016dfde908, {0x33016db96a90?, 0x1?}, 0x33016db96ac0)
	testing/testing.go:2268 +0x3cc
github.com/xianxu/ariadne/cmd/sdlc.TestSysVObservationPreservesCallerValuesWithIndependentBound(0x33016dfde908)
	github.com/xianxu/ariadne/cmd/sdlc/sysvwarning_test.go:283 +0xac
testing.tRunner(0x33016dfde908, 0x1042a9f60)
	testing/testing.go:2195 +0xc4
created by testing.(*T).Run in goroutine 1
	testing/testing.go:2260 +0x3b8
```
- 2026-10-25: built. New tests fail on main (a dead run folder holding a read-only module cache survives both the sweep and the run's own Cleanup) and pass with removeTree. Targeted: TestSweepStaleTemp*, TestPrepareScratch*, TestSuiteCapability*, TestRunnerArtifacts*, TestRunUnderSlotUsesScratch, TestSweepStaleLocalWorktrees*, TestExcludedScratchRootIsNeverSwept, TestStaleTempReason, TestLocalCheck*/TestPreClose*/TestKnownFlakeGate, pkg/golint (incl. TestTrackedGoFilesPassTheCatalog), internal/scratchdisk: all ok.
