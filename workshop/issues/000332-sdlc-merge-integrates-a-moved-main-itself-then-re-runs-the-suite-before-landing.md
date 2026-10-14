---
id: 000332
status: working
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-13
estimate_hours: 2.35
card_mirror: '89f3d390ea3abfa5a267585072506bda46e83a45' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T22:09:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# sdlc merge integrates a moved main itself, then re-runs the suite before landing

## Problem

A branch that passed its close review can sit unmerged for hours, because main keeps moving under it. ariadne#330 passed its close review at 00:53 on 10-11 and was still open at 13:30: each `sdlc merge --yes` (run by the operator, since the auto-mode classifier refuses it for workers until ariadne#318) found main had moved and stopped with conflicts or a stale base, and the worker re-merged main by hand, waited for the operator again, and lost the race again. The operator's question (10-11): shouldn't `sdlc merge` just do that?

## Spec

`sdlc merge` integrates origin/main itself when main moved since the branch's last integration:
- **Clean merge** (git resolves it): commit the merge on the issue branch, re-run the repo's full suite (the same command `close` names as its suite), and land if it passes. No new review: the reviewed diff is unchanged, and #269's freshness rule already defines what main content a review must have seen; reuse it to decide if a re-review is owed (main code that touches the branch's files since the review → re-review; otherwise not).
- **Conflicts:** stop, leave the branch as it was (merge aborted), and print the conflicting paths and the next action for the agent (resolve, then close's `--fixed-then-ship` path, which already demands `--full-suite`).
- **Suite failure after a clean merge:** stop and name the failing tests; with a baseline run of main's suite, distinguish failures the merge introduced from ones already on main (and say so).
- Retry loop: if main moves again during the suite run, integrate again (bounded, e.g. 3 times), then stop with a clear message.
- This is also where the suite becomes enforced at landing, not agent-typed evidence: in repos without an in-binary test gate, `close` records `--verified` text, and failures have reached pair main that way (pair#450).

## Done when

- A branch whose main moved without conflicts lands with one `sdlc merge --yes`: merge commit, suite run, land (test with a fake remote).
- A conflicting branch stops with the paths and the next action, and the branch is unchanged (test).
- A suite failure introduced by the merge is reported as such, and one already on main is labeled pre-existing (test).

## Plan

Durable plan: `workshop/plans/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing-plan.md`.

- [x] M1 — the suite as a recorded, enforced check (`sdlc-local/suite`: declared or default `make test` with the Makefile's weight; required on every landing; a docs-only-distance ancestor record counts; lock weight fix; `sdlc local-check run --name suite` documented as the close suite)
- [ ] M2 — integration loop in the durable path (merge main, conflict abort, publish-replay pre-check, lease push, bounded CI wait, re-integrate up to 3, resume an interrupted integration); legacy path refuses a moved main
- [ ] M3 — failure labelling: a baseline run on main classifies introduced vs pre-existing

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.05 impl=0.2
item: smaller-go-module      design=0.05 impl=0.2
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0 impl=0.16
item: greenfield-go-module   design=0.25 impl=0.3
item: smaller-go-module      design=0.05 impl=0.16
item: smaller-go-module      design=0.05 impl=0.12
item: milestone-review       design=0.0 impl=0.16
item: smaller-go-module      design=0.05 impl=0.2
item: milestone-review       design=0.0 impl=0.16
design-buffer: 0.15
total: 2.35
```

- The items, in order:
  - M1: suite check (weight, default suite, effective declaration, main-side probe); ancestor-record rule with GitHub statuses; docs; M1 review.
  - M2: the integration state machine and its shell (merge, abort, replay pre-check, lease push, PR-head sync, resume); CI wait; legacy refusal plus recovery contract; M2 review.
  - M3: parser, classifier and baseline; the close review.
- Design is at ×0.2, because the plan and R1 settle the decisions. Impl is at 40% of the v2 table midpoints.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-11

### 2026-10-13

- Mapped the landing path (ariadne:1). `runDurableMerge` has no branch-contains-main check: GitHub merges a stale head server-side, so the landed tree is never tested. The re-review rule after integrating main is already the publish gate's (#304: the reviewed patch is replayed onto the new base; a code conflict owes `--fixed-then-ship`), not #269 freshness (stale/fresh only). sdlc never runs the suite; #345's `runLocalCheck` is the engine (exact sha, detached worktree, machine lock, recorded status). A bug was found there: `acquireSuiteSlot` takes weight 1, so `make test` goes reentrant and ariadne's weight-2 sharded suite would hold one slot.
- Draft design sent to the hub:
  - integrate in `runDurableMerge`; on conflict abort and name the paths;
  - pre-check the publish collision on the merge commit before running the suite;
  - push with a lease;
  - the suite as a REQUIRED recorded local check on the exact head (enforced at every landing), with a `weight` field;
  - on failure, a baseline run on main labels failures introduced or pre-existing;
  - bounded CI wait;
  - re-integrate when main moves again, at most 3 times;
  - an interrupted push resumes via an `Integrated-Main:` trailer.
- Hub approved (2026-10-13): (a) declaration, else default `make test` (Makefile weight); (b) every landing needs `sdlc-local/suite` on its head, with `sdlc local-check run --name suite` as THE close suite; (c) the legacy path refuses a moved main; (d) a 20-minute CI wait. Keep the baseline on failure. Close's own commits change the head, so a record on an ancestor with a docs-only distance counts (it reuses the publish gate's docs/code line).
- (superseded) Open questions to the hub: the suite source (declaration and/or a `make test` default), enforcement on every landing, the legacy path, the CI wait bound.

- Codex resumed at `5a4e3cdf09b31f8591625586130295767131e7c9`; M1 round 2 returned REWORK. R3 records the complete correction of optional-suite enforcement, failed-parent lookup, descendant-held probe pipes and README guidance. New tests reproduced the bypass/error/lifetime failures before correction. Initial post-fix tests printed PASS but concurrent documentation edits tripped the hermeticity guard; verification must run with all writers quiet.
- R3 verification: the combined declaration, CI-gate, ancestor, make-probe, local-check command/runner, weight and status-parser regressions passed (`go test ./cmd/sdlc -run '^(TestDeclaredSuite|TestRunCIGate|TestLocalRecordsOnAncestors|TestEffectiveLocalChecks|TestMakeProbe|TestProbeMakeTestFacts|TestDefaultSuiteFactsComeFromMain|TestParseMakeTestFacts|TestParseLocalChecks|TestAcquireSuiteSlot|TestLocalCheck|TestSelectLocalChecks|TestRunLocalCheck|TestParseCommitStatuses|TestDecideCIGate)' -count=1`, 28.346s). The same subprocess lifetime fix covers the local-check runner; both descendant tests failed before it and passed in this run.
- M1 round 3 accepted R3 implementation but returned REWORK for a probe-cleanup test false negative and missing live status conformance. R4 tightens the exit bound to one second (removing group cleanup now fails both cases in 4.867s) and adds `TestLiveCommitStatusesConformance`, read-only against four independently pinned status records on ariadne PR #229. Live conformance plus all four probe/runner process cases passed in 8.298s with `SDLC_LIVE_GH=1 SDLC_LIVE_COMMIT_STATUSES_GH="$(command -v gh)" go test ./cmd/sdlc -run '^(TestLiveCommitStatusesConformance|TestMakeProbeBoundsDescendantPipes|TestRunLocalCheckTerminatesDescendants)$' -count=1 -v`.
