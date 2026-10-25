---
id: 000630
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours: 1.99
card_mirror: 'ad6efb64fb42ebaabd046947f82415d45f4e9531' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T09:51:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis

## Problem

## Spec

**An all-shard timeout under load is treated as an unattributable failure and triggers a full main-baseline run
(ariadne:5 on ariadne#616, 10-25).** #616's integrated merge suite (04:31–05:01) hit the 1800 s shard timeout on every
shard while the machine's load average was 30–50; 764 `cmd/sdlc` tests never ran (the same suite took 19–20 min
earlier). sdlc classified it as an unattributable failure and started a main-baseline diagnosis: another full ariadne
suite (both machine slots), queued behind everything, holding ariadne's landing line for an hour or more, to diagnose
a run that didn't fail any test.

A run where shards time out without a failing test isn't evidence against the code or main. Classify it as
"incomplete (timeout)", distinct from a failure: retry the same head's suite once (in its normal queue turn), and
diagnose against main only if a test actually fails. Report the load average at the timeout in the outcome.

## Done when

- A suite whose shards time out with no failing test records "incomplete (timeout)" with the load at the time,
  retries once, and runs no main-baseline diagnosis; a real test failure still diagnoses as now. Tests for both.
- **The retry runs only the packages that didn't finish** (operator, 10-25): Go reports results per package and sdlc records which passed in the attempt, so the retry selects the unfinished packages through the affected-run machinery instead of rerunning the whole suite; test-level granularity is a non-goal. Test: a timed-out attempt with some packages passed retries only the rest.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.05 impl=0.16
item: smaller-go-module      design=0.1  impl=0.16
item: smaller-go-module      design=0.1  impl=0.2
item: smaller-go-module      design=0.2  impl=0.2
item: smaller-go-module      design=0.1  impl=0.2
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0  impl=0.32
design-buffer: 0.15
total: 1.99
```

The rows are:
- the shard runner's alarm marker and its parse;
- the classifier and the kind;
- the v5 attempt with the load, the typed error, and the guard and flake-view treatment (M1);
- the resume composite's incomplete branch: the v5 suite link, the pass rule and the version gates;
- the child command, its envelope and the automatic retry (M2);
- the atlas;
- two boundary reviews.

Familiarity is 1.0: the resume code was read in full for the plan (both plan-quality rounds checked its line references).

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: [000630-shard-timeouts-incomplete-plan.md](../plans/000630-shard-timeouts-incomplete-plan.md) (TL decisions
2026-10-25: a passing retry posts green; the retry covers the unfinished packages plus every package with a named
failure).

- [x] M1 — classify, record, report, skip diagnosis: the shard runner's alarm marker, parsed process facts,
  `classifyIncomplete` and `flakes.IncompleteTimeout`, the v5 attempt with an additive `load_average`, the typed
  `suiteIncompleteError` (no main baseline, no train split), the guard and flake-view treatment.
- [x] M2 — the one retry: an incomplete parent is resumable with its own package selection; `runLocalCheck` retries
  once through `resumeSaved`; a pass posts green, incomplete again stops, a real failure diagnoses as today.

## Log

### 2026-10-25
- 2026-10-25: closed M2 — M2 built (6c8fd70a6, gofmt 69c20fc3d and 0dec5a6b5), origin/main integrated (bring in #616 etc.; clean merge, then vet, linux build, the incomplete tests, the #616 merge-diagnosis and train hooks, internal/affected, internal/flakes and pkg/golint all ok). Option (b): scripts/test-shard.py honors SDLC_AFFECTED_GO_PKGS (PackageSelectionTest; a real run of one listed package ran it alone, an unknown package exits 2). Retry through the resume composite with a v5 suite link and the check own envelope: TestIncompleteRetryRunsOnlyItsPackagesAndLands (exactly o/r/pkg; 30-minute envelope, not the targeted cap; posts success; reuse verifies), TestIncompleteRetryRealFailureDiagnoses, TestIncompleteBaselineNeverRetries, TestResolveIncompleteResumeProof, TestKnownFlakeSuiteLinkStaysVersion3, TestIncompleteRetryOutcome, TestIncompleteRetryCommands, TestLandingTrainIncompleteSuite (production conductor path, for #635). Mutation pass: all 15 M2 guards turn a test red alone. Regression slices ok before the merge: Resume/KnownFlake/LocalCheck/Suite/Saved/Republish/TypedSuite/Flake, LandingTrain (572s), Merge, Diagnose/PreClose/CIGate/Stats/Weekly/Recovery/Help/Queue/Decode.; review verdict: SHIP
- 2026-10-25: closed M1 — Round 2 after REWORK, all nine findings answered in d81ab0241: BR-2 plan tables corrected with a Revisions entry and the issue M1 bullet fixed; BR-3 hostload_other.go (GOOS=freebsd and linux builds ok); BR-4 savedAttemptRecord refuses incomplete-timeout (TestSavedAttemptRecord row, and republishSavedPass reuses nothing and posts nothing in TestSuiteIncompleteIsNotATestResult), with the outcome-consumer sweep listed in the plan; BR-5 incompleteEligible pins the typed-contract guard (TestIncompleteEligible) and the classifier typed-facts case is a table row; BR-6 scripts/test-shard.test.py runs the real Go alarm through go test -json and test2json and asserts the marker; Minors: Running in the report, the Baseline wording, v5 refuses typed identities and known_flakes, and TestSuiteMixedAlarmAndRealFailureDiagnoses. Plus the earlier M1 tests, the fuzz test, regression slices and pkg/golint.; review verdict: FIX-THEN-SHIP
- 2026-10-25: plan-quality: round 1 FAILURE (PQ-1 M2 resume reuse claimed unchanged; PQ-2 version gates and typed contract), answered by naming the M2 mechanics in the plan Revisions; rounds 2-3 INFO. PQ-5 (Minor, carried to close): in ariadne, which declares no .sdlc/affected.json, the retry reruns the whole suite; asked the TL whether to revise the Done when or make the suite honor a package list.
- 2026-10-25: estimate 1.99 (v3.1, Method A). estimate-quality INFO: impl looks thin on the two version-gated rows and expects 2.5-3h; kept, the measured actual will say.
- 2026-10-25: commit bf93b4f38 is titled plan: but carries the M1 code (a pathspec-less commit took the staged M1 files); cbf1db7e7 adds the fuzz test. A history rewrite to split them was refused by the auto-mode classifier (force-push), so the history stands as is; M1 is bf93b4f38 plus cbf1db7e7.
- 2026-10-25: M1 closed (d094578e9, FIX-THEN-SHIP, window e19cdb92..d81ab024). Its Minor (third in the plan-claimed-test-missing family) is fixed as the rule: every guard in classifyIncomplete, incompleteEligible, the v5 decode envelope, savedAttemptRecord and isFlakeEvidence was deleted alone and the tests rerun. All 14 go red: the typed case now uses well-formed unit events, the build failure sits in an alarmed process, there is a parse-warning case, the deadline and no-exit cases are split, and v5 refusing known_flakes is tested. Two guards no single deletion could expose were removed as redundant: the typed-identity check on failures (typed names only come from typed evidence) and !rec.passed() (implied by Exit > 0 before any outcome is set).
- 2026-10-25: M2 built. The TL chose option (b) for PQ-5, so test-shard.py honors SDLC_AFFECTED_GO_PKGS (tested by
  PackageSelectionTest; a real run of one package ran it alone in 2.7 s). The retry goes through the resume composite
  with a v5 suite link, the check's full envelope, and posts through the one existing path.
  - Tests: TestIncompleteRetryRunsOnlyItsPackagesAndLands (the operator's test: the retry got exactly o/r/pkg, under
    the check's 30 minutes rather than the targeted cap, and a later reuse verifies the pass),
    TestIncompleteRetryRealFailureDiagnoses, TestIncompleteBaselineNeverRetries, TestResolveIncompleteResumeProof
    (wider selection, group, v3 link, a parent no longer incomplete with and without a link, receipt version: all
    refused), TestKnownFlakeSuiteLinkStaysVersion3, TestIncompleteRetryOutcome, TestIncompleteRetryCommands,
    TestLandingTrainIncompleteSuite (for #635).
  - Mutation pass: all 15 M2 guards turn a test red when deleted alone. TestResumeEnvelopeVersionsByCheck's
    "v5 suite refused" row now says allowed, the contract #630 changes.
  - Regression slices pass: Resume/KnownFlake/LocalCheck/Suite/Saved/Republish/TypedSuite/Flake;
    LandingTrain (572 s); Merge; Diagnose/PreClose/CIGate/Stats/Weekly/Recovery/Help/Queue/Decode;
    pkg/golint; internal/flakes; scripts/test-shard.test.py.
- 2026-10-25: M2 closed (4781322d3, SHIP, window d81ab024..28d76cca, after integrating origin/main at 28d76ccae). Its Minor is fixed: resolvedSavedRecord returns the attempt whose log backs the record (the parent, or an incomplete parent's failed retry), so both callers set LogPath from it instead of inferring it from digest equality.
