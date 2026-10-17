---
id: 000420
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-17
estimate_hours: 4.92
severity: medium
card_mirror: 'be6cb820f9fa6906d6a742e48ffd00a4aacdaa69' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T00:53:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# Actual: queued suite status polling bridges idle time into focused-hours samples

## Problem

While closing #416, automatic actual adoption reported 2.40h for the window 2026-10-16 19:36:21–21:58:36 America/Los_Angeles, with a dominant 142.1m/99% attribution segment. Most of that window was waiting for machine test slots, with assistant status updates and write_stdin polls about every 45–60 seconds to obey the communication rule.

The recorded suite attempts retain exact queue evidence under the shared git directory's sdlc/test-stats/attempts/:
- d5ecb880e188a182590c4815cf749537.json: queue_seconds=3055.535157833; wall_seconds=532.971970042; suite SHA 005587765f51a3baf610fd3f70b1bed72f0b1a65.
- 0bbae3af1b6112bb7cb13acbc505a1ca.json: queue_seconds=3895.523716167; wall_seconds=635.788244625; suite SHA 1b31e6fc1f288e278b1f1c69fc921726c03cd142.

Together these are 115.85 minutes queued and 19.48 minutes executing suites inside a roughly 142-minute close window. Some real work (the staged live check) occurred concurrently with the first queue wait, so blindly subtracting whole queue windows would also be wrong. No replacement hours were typed; the tool's adopted value was retained pending a calibration-policy decision.

## Spec

Investigate how native task spans and frequent assistant/tool polling turn queue idle time into focused dev-hours. Make the metric's treatment of queue waits explicit, preserve genuine overlapping work, and keep raw elapsed/queue evidence separate from focused effort. Repair or exclude affected samples through the existing generation-bound correct-actual path with honest provenance, never guessed numeric replacements.

## Done when

- A deterministic native-transcript regression reproduces a long queued suite with periodic tool polls/status updates and meaningful work overlapping part of the wait.
- The measurement distinguishes idle queue time from genuine overlapping work according to the documented metric, with retained evidence that supports replay and correction.
- #416's adopted actual is evaluated by the TL/operator and corrected or explicitly excluded if the sample is invalid; no hand-typed hour estimate is used.

## Plan

- [x] Pin each eligible suite queue record to this issue's repository, owning slot, engagement and measurement window; retain the exact evidence used by a close or correction.
- [x] In the native active-time engine, remove queued idle intervals from the owning session's activity while retaining timestamped work inside them; preserve phase splits and uncounted review tails.
- [x] Cover a polled queue, overlapping live work, duplicate/foreign/missing records and replay with deterministic native-transcript tests; run targeted tests and the recorded suite before close.
- [x] Re-measure #416 from the corrected implementation and give the TL/operator the measured result and evidence for generation-bound correct-actual or explicit exclusion.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=1.0 impl=0.3
item: greenfield-go-module design=1.0 impl=0.3
item: smaller-go-module design=0.2 impl=0.16
item: smaller-go-module design=0.2 impl=0.16
item: smaller-go-module design=0.2 impl=0.16
item: issue-spec design=0.5 impl=0.1
item: atlas-docs design=0.1 impl=0.06
design-buffer: 0.15
total: 4.92
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Two novel concerns are queue attribution and structural native-tool evidence; the three smaller extensions are attempt recording, snapshot/correction wiring (including #416's generation-bound dry-run, evidence assessment and correction/exclusion handoff), and phase/report integration. This uses the already available Go and SDLC libraries, so no novel-stack library discount applies. The calibration is marked stale, hence provisional.

## Log

- Found while closing #416; independent of its queue-helper lifecycle fix. The close prints active-time-v3-engagement-v1 and 2 admitted sessions, with the complete attribution segment covering the wait-heavy window.

- TL decision, 2026-10-16 (receipt b6986b85-dc38-4def-a646-346c65ee6726): queue waiting is not active work; actuals should exclude recorded queue_seconds by subtracting those measured intervals, not guessing replacement hours. The issue stays unassigned. Implementation must handle overlap with real work honestly as described above.

- 2026-10-17: Claimed in ariadne:1. Current `sdlc actual --issue 416` replays 5.07h (2.41h build, 2.66h converge); the native engine admits assistant status/poll events into one 304.2-minute attribution segment. Four validated suite records on ariadne:5 lie in that window, including the staged-check overlap. The issue is full-flow work because measurement inputs must retain queue evidence for later correction replay.
- 2026-10-17: TL approved the full-flow plan. Legacy attempt windows reconstructed from `started_at` carry a documented post-lock setup shift; new attempts will retain the actual queue start and acquisition timestamps. The plan includes a regression for both formats.
- 2026-10-17: Plan-quality round 1 returned REWORK on four executable-design gaps: an enclosing suite call could restore its own queue wait, historical replay needed a separate correction purpose, exact timestamps lacked a recorder/schema path, and risky functions needed named tests. Revised the durable plan to classify wait/work spans structurally, pin a generation-bound correction candidate, extend v1 attempt records compatibly, and test those seams adversarially (ARCH-PURPOSE, ARCH-ORDER, ARCH-SECURE).
- 2026-10-17: Plan-quality round 2 cleared three findings. For the remaining test-strategy finding, the durable plan now names a unit target, adversarial input and mechanical guard for every risky function, with the native transcript and correction replay reserved for end-to-end acceptance.
- 2026-10-17: Full-flow `change-code` passed plan-quality and estimate-quality (info only). Endpoint TDD exposed that `time.Sub` on live values can use monotonic readings that differ from serialized UTC endpoints; the recorder now derives `queue_seconds` from those persisted endpoints. New v1 records contain exact queue start/acquisition times; legacy records retain the documented late-shift window.
- 2026-10-17: Queue-aware measurement now pins validated attempts to an owning slot and engagement, retains selected waits with the snapshot, and subtracts only queued idle from the bound native source. Paired stage/edit/task work survives; polls and the enclosing suite wait do not. Native, source-binding, snapshot and replay regressions pass. A completed issue's ordinary historical read still uses its old snapshot; a correction candidate requires fresh complete owned queue evidence. The parser also follows numeric JSON session IDs from yielded tool results.
- 2026-10-17: #416's read-only correction preview found no uniquely bound owning suite invocation for recorded attempt d5ecb880. The TL confirmed its adopted 1.12h was a hand subtraction of the whole queue wait, not a verified interval result, and approved exclusion. `sdlc issue correct-actual --exclude` completed operation `actual-correction-c6c9f528ea66`, changing the effective sample from 1.12h to N/A with the original and reason retained in correction history. No numeric replacement was typed.
- 2026-10-17: Queue replay now tests a wait crossing build, converge and tail cuts: the corrected 25-minute attribution partitions 10/5/10 minutes. Targeted sdlc and internal-engine tests, `make -k build`, and `git diff --check` pass; the recorded full suite remains before close.
- 2026-10-17: Pushed integrated head b21211102c6611f499527d26f1623aa487f75af0 to PR #279. The recorded `sdlc-local/suite` on that exact head passed: `make test` ran 12 sdlc shards (1,544 tests) and 66 other packages, all exit 0, wall 511.1 seconds after 1,380 seconds in the machine queue; saved log SHA-256 begins 427a981e0cf2. The standalone `make -k test` was not duplicated: its earlier queued invocation was cancelled before taking slots, and this recorded full suite is the required test evidence.

## Revisions

- 2026-10-17 — TL clarified that "subtract exactly" means use recorded queue intervals, remove only idle portions, and retain genuinely overlapping work evidenced by transcripts or another task; no estimated replacement hours. This resolves the Spec's overlap question and makes retained queue provenance part of the implementation contract.
