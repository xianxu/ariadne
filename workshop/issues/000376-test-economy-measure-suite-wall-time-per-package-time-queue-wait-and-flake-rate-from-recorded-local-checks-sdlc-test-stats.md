---
id: 000376
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours: 3.48
card_mirror: '59a5dbff06fab52bc6a7a0daf717f88f901fb6b6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T20:40:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# Test economy: measure suite wall time, per-package time, queue wait and flake rate from recorded local checks (sdlc test-stats)

## Problem

The operator (2026-10-14) made test health and performance a train ("test economy" in the ops TL status file): find a sustainable pattern to keep pair's and ariadne's tests fast and trustworthy. On 10-14 the suite queue set the fleet's pace: ariadne:3's recorded suite waited ~50 min for its turn; pair#528 spent most of the day waiting for or rerunning ~10-minute suites; flakes (#369's two, pair#423, pair#554, the #514 class) cost whole extra runs. Nobody can say where the time goes, because nothing aggregates it.

## Spec

`sdlc test-stats [--since 7d] [--repo …] [--json]`, read-only, from what already exists: recorded local checks (#345/#332 `.git/sdlc/local-checks/<sha>/`), suite logs (`go test -json` / shard summaries), and the suite lock's ticket log (#357) for queue waits. Report per repo: runs, wall time (p50/p90), time per package or shard (top N), queue wait (p50/p90), and **flakes**: tests that failed and then passed on the same head, or failed on main in a baseline run, with counts. It's the first car of the test-economy train: it says where to aim targeted runs (#328) and minimization sweeps (#327).

## Done when

- `sdlc test-stats --since 7d` on this machine prints the per-repo table for ariadne and pair, with the top slow packages and the flake list (tests on recorded fixtures for the parsers).

## Plan

- [x] Parse recorded suite logs and aggregate per-repository wall, queue, slow-unit and flake metrics; test pure calculations with ariadne and pair fixtures.
- [x] Preserve one compact record per future suite attempt so retries, queue waits and baseline runs are measurable; test retention and lost status posts.
- [x] Expose `sdlc test-stats --since 7d [--repo …] [--json]`, verify the real ariadne/pair output, update CLI documentation, run the suite and close review.

Detailed design: [test-stats plan](../plans/000376-test-stats-plan.md). The TL approved recording future attempts because current logs cannot reconstruct all requested metrics.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.4 impl=0.32
item: greenfield-go-module design=0.4 impl=0.32
item: greenfield-go-module design=0.4 impl=0.32
item: smaller-go-module design=0.1 impl=0.2
item: atlas-docs design=0.1 impl=0.08
item: milestone-review design=0.1 impl=0.2
item: issue-spec design=0.2 impl=0.08
design-buffer: 0.15
total: 3.48
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Three greenfield Go concerns are log parsing, aggregation, and durable attempt storage (design 2.0 hours each discounted ×0.2 for the detailed plan, implementation 0.8 hours each scaled ×0.4); the smaller module is CLI wiring on existing Cobra/fleet helpers. The calibration source reports stale relative to its ledger, so this estimate is provisional.

## Log

### 2026-10-14

- Claimed in ariadne:1. Source inspection found ariadne's retained suite logs contain shard summaries and pair's contain Go package timings, but the single log per SHA is overwritten by a retry; queue tickets are ephemeral; pair logs lack total wall time. A per-attempt compact record is needed to measure the full Spec going forward. Historical fields that cannot be recovered will be marked unknown (ARCH-PURPOSE), not estimated from file times.
- Plan-quality round 1 raised PQ-1–PQ-3 (concurrent publication, persisted-input validation, named function tests) and two Minor findings. The durable plan now specifies a private in-flight directory and atomic record marker, safe cleanup, bounded validation, controlled-order tests, reuse of `failingTests`, and an operating envelope. Re-running the gate on that revision.
- Plan-quality passed on round 2; change-code entered full-flow implementation after a reconciled 3.48-hour v3.1 estimate. Estimate-quality accepted it with Info that storage and fleet integration may consume more than their primitive allocations; the measured actual at close will calibrate that.
- Focused tests prove parser/aggregation, malformed-record rejection, failure-log collision refusal, 30-day pruning, crash-orphan cleanup, same-head retry after a failed status POST, exact-head Agent trailer, measured lock wait, and long-line log hashing. `make -k tools` and `git diff --check` pass.
- The real read-only CLI from ariadne:1 reports ariadne 20 historical suite runs (wall coverage 19/20, queue 0/20) and pair 20 (wall and queue 0/20), with slow-unit lists; historical unknowns remain `n/a`. A third sibling, parley.nvim, has 8 partial runs. No confirmed flake rate denominator exists until new per-attempt records accumulate.
- The first recorded suite was queued behind two other runs; I cancelled it before it acquired the suite lock after spotting a total-reader-memory gap. The reader now streams directory entries and caps aggregate JSON intake at 64 MiB with an explicit warning. The suite will run on the corrected pushed head.
- During the next queue wait, validation review found an omitted JSON `exit` defaulted to zero and could fabricate success. Added required-field and duplicate-key checks, with regression fixtures; cancelled before acquiring the suite lock again so final verification applies to the corrected head.
- The recorded `sdlc local-check run --name suite` passed on pushed head `0554b5cd5c50dfbf274db6948f2107cb28ee7094`: 12 shards and 59 other packages, 0 failures, wall 538.8s, queue wait 2557s; check `sdlc-local/suite` recorded with log SHA-256 prefix `0eb8013093b0`. `git diff --check` passed. Submitting the branch to the close boundary review now.
- Close review round 1 returned REWORK: BR-1 JSON event format was unparsed, BR-2 a smoke attempt hid same-SHA legacy suite evidence, BR-3 the plan mislocated `SuiteAttempt`, and BR-4 crash-left publication temps had no owner-based cleanup. Reproduced both behavior defects with red tests; fixes and the corresponding plan/lessons revisions are in progress (ARCH-PURPOSE, ARCH-FUNERAL).
- Round 1 fixes pass focused parser, command, publication, pruning and local-check tests. `make -k tools`, `git diff --check`, and the real read-only CLI pass; the rerun shows ariadne 20 historical runs, pair 21, and explicit unknown queue coverage. A recorded suite on the fixed head is required before re-close.
- The fixed-head recorded suite passed on `d65cc623a420bbd019909153783917f102d2ba38`: 12 shards and 59 other packages, 0 failures, wall 547.7s, queue wait 1831s; `sdlc-local/suite` recorded with log SHA-256 prefix `4df30d4c4715`. Re-submitting the four review dispositions at the close boundary.
- Close review round 2 disposed BR-1–BR-4 but returned REWORK on BR-5 (case-insensitive JSON aliases could override an outcome), BR-6 (plain `FAIL` package timing absent), and BR-7 (planned 1 MiB line limit unenforced). Reproduced the alias failure with a red test; the parser tests now cover every text/JSON/shard outcome and exact/over-limit lines. Persisted keys are checked for canonical spelling before struct decode, including nested durations (ARCH-SECURE, ARCH-PURPOSE, ARCH-CONSTRAINTS).
- The round-2 fix head `d4b102b5d41c7cc4bceecde4b7f74c54d598db52` passed `sdlc local-check run --name suite`: 12 shards and 59 other packages, 0 failures, wall 617.0s, queue wait 1345s, log SHA-256 prefix `eba87fa23c08`, status `sdlc-local/suite` recorded. The broader targeted tests, `make -k tools`, and `git diff --check` also passed; submitting BR-5–BR-7 dispositions for re-review.
- Close review round 3 disposed BR-5–BR-7, then found BR-8: a nested duration's JSON null could decode to zero and bias slow-unit statistics. A red read-and-aggregate test proved null `kind` and `seconds` were admitted. The decoder now rejects null for every required nested scalar, emits coverage warnings instead of rows, and validates duration kinds (ARCH-SECURE, ARCH-PURPOSE).

### 2026-10-15

- The BR-8 fix head `e1f4826083f90bcf6540d108394926cc072391d7` passed `sdlc local-check run --name suite`: 12 shards and 59 other packages, 0 failures, wall 482.1s, queue wait 2603s, log SHA-256 prefix `8107e1d22518`, status `sdlc-local/suite` recorded. The broader targeted tests, `make -k tools`, and `git diff --check` passed; submitting BR-8 for re-review.

## Revisions

### 2026-10-14 — TL-approved recorder scope

Record verified slot and nullable agent/model/effort compatible with #367. Save exact head, exit and log hash before status posting so #380 can retry a failed upload from the saved result. Preserve each failed attempt's log across same-head retries for 30 days, together with its metadata. Suite commands and protected gates remain unchanged; historical unknowns stay unknown.
