---
id: 000555
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b06bf6fd6c6339474650069c3bf84be0ba3c9cdb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T11:30:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc merge holds the repo lock across its suite wait, blocking stage --clear/--build (and operator checks) for the whole suite

## Problem

## Spec

10-22 ~11:30: pair:9's `sdlc stage --clear` on pair:0 (to free it for the operator's next staged check) waited on
pair's common sdlc repo lock, held for 30+ minutes by pid 17156, pair:6's `sdlc merge --yes` for pair#803, which
was sitting in its landing turn waiting for its integrated suite. So a landing waiting for a suite blocks every
stage operation in the repo, and with it the operator's live checks, for the whole suite duration.

Fix: a merge runner holds the common repo lock only around the short critical sections it needs (integrate, push,
merge, archive), not across the suite wait; or stage operations use a separate lock (the stage target's lock) that
doesn't conflict with a landing's suite wait. Keep the existing safety: a stage must still never race the landing's
integration or push.

## Done when

- With a merge runner waiting on its integrated suite, `sdlc stage --clear` and `--build` on the stage target proceed
  without waiting for the suite (test with a held fake suite).
- Stage operations still refuse or wait during the landing's actual integrate, push or merge sections (test).

## Plan

- [ ] Pin the corrected incident timeline and approved test-only design in the durable plan.
- [ ] Exercise actual CLI stage clear/build during separately blocked suite admission and execution; prove the regression fails under whole-command wrapping.
- [ ] Pin short-transaction serialization and post-suite head-drift refusal through the command path.
- [ ] Run focused checks and the recorded suite; close as not reproduced with regressions pinned, then land.

## Log

### 2026-10-22

### 2026-10-22 — investigation evidence and approval

Claimed and entered planning from main; no production or test implementation changed. Current NewMergeCmd uses markManualLockCommand; durable and legacy suite/CI paths run outside f.withLock. Existing TestLandingReleasesRepoLockDuringSuite, TestRunMerge_LegacyReleasesRepoLockDuringLocalCheck and TestWithRepoTransactionLockIsContextReentrantOnly passed together in30.731s. #370 originally introduced the scoped lock (implementation230f61e2f43550c0405bb9c5b7bbf4dc7666f975; merged via ef420af80d2b81cae2d5a9008c4151ddca6b6a34 on2026-10-14).

Pair:9 retained session96254's exact line:

```text
waiting for sdlc repo lock held by pid 17156: sdlc merge: on Xians-MacBook-Pro.local (long-running review/ship transaction)
```

Its clear started after TL authorization11:07:33.679, and completion report was accepted11:11:04.226 after restoring pair:0 main fb79d27bd66bb42f52601055915b3798fe5a6dc8 and verifying four binaries. At about11:09, ps showed pid17156 PPID46243 elapsed29:32. That is process age, not lock-held duration.

Pair:6 positively bound pid17156 to #803 using landing seq274, operation op-6a1bb7bf0f02b3eaa656cb5dcdaffd05, pid_started1792690767.793984 (~10:39:27 PDT), admitted10:39:29.800737. Suite attempt f5e320381ff102cd8180f600780f0ccc queued10:42:17.469, acquired10:51:21.388, finished11:03:30.681 exit0. Its retained output by11:07:09 still waited for refreshed-head CI; by11:08:12 it showed CI4green and a repo-lock wait on issue set-status pid39100; by11:09:02 it waited on stage pid54025. Merge commit61b60a8f has committer time11:08:20, landing event11:10:12.643 and final successful tool snapshot11:10:45. Buffered tool output receipt times are upper bounds, not lock-transition timestamps. These establish post-suite contention, not continuous merge ownership. Exact executable launch revision and individual lock acquisition/release timestamps were not retained; do not invent either. Prior observed clean binary1c0595d at10:27 does not prove the10:39 launch revision.

Sources supplied read-only by pair:6: pair .git/sdlc/test-stats/attempts/f5e320381ff102cd8180f600780f0ccc.json and .git/sdlc/landing/events.jsonl lines852–856; Codex rollout2026-10-17T11-42-56-01a14b2c-9890-74f3-a738-58879659bef5.jsonl tool receipts55267/55280/55299/55319/55331. No peer processes, worktrees or data changed.

A read-only subagent audit confirms no context leak: the repo-lock wrapper skips manual commands and restores its context after each scope. Stage's target flock is nonblocking, and stage/move destinations alone honor it. Removing stage's broad lock would expose build/switch races with landing and move sources. Reuse existing exclusion (ARCH-DRY/ARCH-ORDER), add coverage through buildRoot/NewMergeCmd rather than synthetic manual flags (ARCH-PURPOSE). TL approved this test-only direction in delivery68c0cfbe-1299-4a61-99ec-cbd24b029ef3.

## Revisions

### 2026-10-22 — correct the reported duration and retain the existing lock design

The original Spec's thirty-minute whole-suite hold is not supported by recovered evidence. Pair:9 observed stage --clear wait about11:07:34–11:11:04; elapsed29:32 was process age. Pair:6's recorded suite finished11:03:30, before that wait. The wait belongs to a short, legitimate post-suite transaction rather than suite admission/execution. Keep #370's existing scoped locking. TL approved command-level clear/build regressions for separately blocked admission/execution, short-transaction serialization and head-drift controls; no runtime change unless these fail. If they pass, close as “not reproduced; regressions pinned”. Durable plan: [implementation plan](../plans/000555-merge-stage-lock-regressions-plan.md).
