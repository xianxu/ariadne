---
id: 000631
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '83e7da7d39d9e99742ccc2370c1f33625685bca0' # card fields mirrored from issue-cards; edit via sdlc
---

# Suite shard alarm during shared scratch removal lacks exact target and timing evidence

## Problem

A recorded #624 pre-close suite reached the 30-minute shard alarm while TestSysVObservationPreservesCallerValuesWithIndependentBound/true was inside prepareScratch -> sweepStaleTemp -> os.RemoveAll. This is an incomplete suite, not a registered flake or valid pass. Diagnose synchronous stale-root cleanup and filesystem progress before retry policy changes.

The alarm captures the stack at the deadline; it does NOT prove RemoveAll itself had run for all 30 minutes. Its entry time and exact target basename were not logged.

## Spec

Record and investigate which stale cleanup target can hold a managed run's setup, whether filesystem removal stalls or normal shared-root work consumes the package deadline, and how setup cleanup can remain safe and bounded. Preserve live-owner and retained-artifact protection. Do not delete live or unknown peer roots or raise test deadlines as the diagnosis.

## Done when

- Capture the cleanup root, exact candidate, time spent and liveness/protection decision before removal, so a timeout names its source.
- Reproduce with an isolated writable fixture (including slow or interrupted removal as appropriate), establish the root cause and test the correction without sweeping production scratch with fake liveness.
- Verify live/retained roots remain protected; run the recorded suite through the existing queue.

## Plan

- [ ]

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
