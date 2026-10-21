---
id: 000521
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0b1244389a18b755fd7d32e23e8c6981622a827a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T01:01:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "1346236b", done: "1cb1011e"}
---

# landing: a car held from a train is never told; its runner shows the candidate banner and can cancel the train's suite

## Problem

## Spec

A car dropped from a landing train is invisible to its own runner, and that runner can cancel the train's suite.

Observed in pair, 10-20, train 213:
- Car #731 (PR 503) was held at 23:10:36 for a merge conflict. `events.jsonl` seq 213 records `result: held`, with
  the instruction "its merge conflicts with main or with the cars before it on the train; integrate main,
  re-close and run sdlc merge".
- #731's merge runner stream never said so. It printed `landing train 213: 2 cars: #731 (PR 503) #751 (PR 505)`,
  then "instance-conformance passed", then the suite on 919642ae1. That head, correctly, contained only #751.
- The worker took the banner as final membership, saw its code missing from the head, and cancelled "its" suite,
  which was #751's train suite. Train 213 was abandoned and #751 had to re-form (train 216). Nothing landed wrong.

Two defects:
1. **The banner announces candidate cars before composition, and a held car is never reported to its own runner.**
   The runner's stream must state the final membership once the head is composed. For a held car, it must print the
   held result and its next action, then exit (or wait) without following another car's suite.
2. **A runner whose car isn't on the composed head can cancel that train's suite.** Cancellation of a train suite
   should belong to the cars actually riding it, or to the conductor.

## Done when

- A runner whose car is held prints the held reason and next action, and doesn't report a suite or attach to it.
- The banner distinguishes candidates from final membership (or prints only the final membership).
- A held car's runner cannot cancel the train's suite (test).

### Approved correction

Stop a held conductor before publishing/testing surviving cars. Existing abort cleanup releases those cars to reform under the next head; discard the temporary composition. Continuing would require detached conductor lifetime/output/cancellation ownership, beyond a quick fix. Process the composition Hold results before this guard so every conflict gets its real reason. Distinguish the pre-build candidate banner; report composed membership from Built.Kept before a suite. Give held outcomes their reason and recovery instruction in the owning runner immediately.

Actions: integrate/re-close/retry the held branch; wait for automatic survivor reformation; cancel only a suite owned by the requesting checkout.

| Internal outcome | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Own conductor car held during build | Held reason + integrate/re-close/retry, runner exits before sibling suite | Integrate/re-close/retry | Temporary composed worktree discarded; branches/PRs retained |
| Another car held | That waiter prints held reason and exits; kept membership announced | Integrate/re-close/retry | No branch/history loss |
| Survivors of held conductor | Released to wait and reform under next head | Wait | Composition repeated, no committed work lost |
| Suite owned by another checkout after reformation | Cancel leaves other checkout’s run alive | Wait | None |
| Actual conductor’s suite | Existing managed cancellation | Cancel owned suite | Partial test output retained; no code loss |

ARCH-ORDER/PURPOSE: cancellation authority follows the existing lifetime-held run lease; do not allow excluded caller to create a sibling suite it still owns. ARCH-DRY/PURE: retain existing pure train transitions and abort/release machinery, no new outcome kind or sidecar. ARCH-MOCK: real Git and stateful GitHub train plus subprocess cancellation fixture. ARCH-CONSTRAINTS: added membership output visits N cars once, no extra Git/network per car; check N/2N rendered membership. ARCH-FUNERAL/EVOLVE: no new persisted fields or artifact family. This is semantic effect ordering, defended by forced-conflict and cancellation tests rather than a lexical lint.

## Plan

- [x] Get hub approval of abort/release tradeoff; reproduce held-conductor output/suite ownership with controlled real train fixtures.
- [x] Correct candidate/composed membership output, prompt held outcome and pre-publish self-held guard; verify survivor reformation and cancellation ownership.
- [ ] Run focused controls and recorded full suite, then prepare close verification.

## Log

- 10-20 23:50 (TL ops:0): filed from pair:6's report; the evidence is pair `.git/sdlc/landing/events.jsonl` (seq 213 car
  held 23:10:36; seq 213 train abandoned 23:39:38, cars [213, 216], PR 506 closed).

- 2026-10-21: Claimed/start-plan from main after #516 landed. Read-only diagnosis: conductTrain checks selfEnd only after effect queue drains; Built emits Hold then Publish/RunSuite, and settle(self) records without printing. Existing checkout-owned suite lease therefore remains cancellable from the excluded conductor’s checkout. Proposal sent to hub before implementation; no code changed yet.

- Hub approved abort-before-Publish and release/reformation (e4cc5d1b-712b-4e96-8d77-6b9e5d5ae408). No stage; next natural pair conflict is hub’s live follow-up. Tests must force a real conflict, survivor reformation and foreign-cancel refusal. #522 reassigned to ariadne:3; after #521 report and await next dispatch.

- TDD: old runtime published and merged the clean siblings after its own first car and a later third car conflicted; its waiter output omitted held/recovery and its human summary counted excluded cars. held-red.log retains the expected failures. The first green attempt correctly stopped publishing but exposed a fixture assertion omitting status history from a rewritten third branch; snapshotting the actual existing history corrected the measurement, without changing the runtime policy. That failed attempt is retained as held-green-attempt1-fixture-measurement.log.
- Corrected runtime guards Publish after all Hold effects, reuses one held-recovery formatter for own and waiting cars, and distinguishes candidate/composed banners with one membership formatter. Terminal human summaries count retained membership; persisted candidate event history remains unchanged. No lease fields or cancellation code changed: excluded conductor never creates a sibling suite.
- Real forced-conflict proof: first/third held, second/fourth released; no train PR/merge/suite from held conductor; both survivors reform under prepared second checkout, whose live suite lease names that new checkout. Actual cancelLocalChecks from held checkout stops zero runs and says the foreign checkout’s run remains running; release of the blocked suite lets both survivors merge while held PRs remain OPEN. All regression and existing archive/clock/multi-car/split/conflict/environment/malformed controls PASS141.821s; pure all-conflict/Aborted/random-state controls PASS0.156s. Membership tests at8/16 cars, both labels, render each car once with at mostN+2 writes. Raw evidence common test-stats/evidence/000521/{held-red.log,held-green.log,held-green-controls.log,held-state-controls.log}. Main’s independent native-cancel/lint checks and required recorded suite remain next.

- Main independently reran held-conductor/reformation/actual-foreign-cancel, held-waiter recovery and the native checkout-ownership cancellation control: PASS256.650s. Native control itself PASS240.56s, with its existing cleanup consuming each child completion channel twice and waiting HangCeiling twice after successful assertions; this independent test-cost discovery is being filed separately, not changed here. Full pkg/golint PASS0.914s; staged diff whitespace clean.

- PR #366 published at implementation43072e4e4f92cb5d8880bab3ab39cc6823357614. Required suite queued through local-check, attempt38951803c8c970479b302c1f7746a564, detached worktree sdlc-local-1891726477; pending, no pass claimed. Initial position5 for both slots. The completed native cancellation control exposed independent deterministic four-minute cleanup overhead, filed unclaimed as #523; hub assigned it here after this issue lands (913a7bee-50ed-42b3-8312-88723af72658). #522 stays assigned elsewhere. Finish #521’s suite, close and landing first; no stacked #523 implementation.

## Revisions

- 2026-10-21: Hub approved concrete abort/release policy and no-stage conformance; Done-when unchanged.
