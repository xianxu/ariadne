---
id: 000290
status: done
deps: []
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: '762b03c26f7df0407f4f5383e294d35dbfdc503c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T15:58:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.36
---

# Fleet inventory tracker reads are slow

## Problem

`sdlc fleet inventory` went from ~5s to ~17s on this machine (46 worktrees, 27
repositories) after ariadne#288 added tracker claims. A `GIT_TRACE` of one run
(638 git processes, 18.7s) attributes 10.5s to 15 tracker reads, one per tracked
repository, run in sequence: each is an `ls-remote --refs` plus a `fetch
--quiet`, and each fetch triggers an automatic `git maintenance run`. Before
#288 only worktrees on an issue branch triggered a tracker read. Local git
(~13 commands per worktree, ~13ms each) is the other ~8s.

## Spec

Make the inventory's tracker reads cheap without changing what they report:

- run the per-repository tracker reads concurrently, with a bounded worker
  count, keeping one read per repository (the `repoRecords` cache) and
  deterministic output;
- drop network round trips the read does not need (the `ls-remote` before a
  fetch that already answers presence, if that holds for every state the
  presence check distinguishes);
- suppress automatic maintenance on this read-only path;
- bound each fetch in time (the read takes the command's context; a remote that
  does not answer degrades that repository's claims to stale/unknown instead of
  stalling the inventory).

Stale/unknown semantics, one-read-per-repository and the #288 contract are
unchanged. Prerequisite for ariadne#289, which adds dependency-clone rows.

## Done when

- `sdlc fleet inventory --json` on this machine runs in well under the pre-#288
  ~5s plus local work (target: under 9s here), measured before/after with
  `GIT_TRACE`, numbers in the Log.
- Concurrency is tested: N repositories are read with at most the bound in
  flight, each exactly once, output identical to a sequential run.
- An unresponsive remote yields that repository's rows `stale`/`unknown` within
  the per-fetch deadline, tested with a remote that hangs; other repositories
  are unaffected.
- No automatic maintenance runs during inventory (asserted by trace or by the
  command's git config).

## Plan

Durable plan: `workshop/plans/000290-fleet-inventory-tracker-reads-are-slow-plan.md`.

- [x] Fetch-skip when the remote tracker tip equals the local tracking ref; `--no-auto-maintenance` on tracker fetches
- [x] Per-key-once records cache + bounded concurrent warm-up of tracked repositories
- [x] Per-read deadline; a hanging remote degrades only its repository
- [x] Before/after trace in the Log; atlas + help; close

## Log

### 2026-10-02
- 2026-10-02: closed — fleet inventory 17.83/17.31s -> 6.02/5.81s here, JSON byte-identical; after-trace 15 concurrent ls-remote, 0 fetch, 0 maintenance. Tests: TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged (incl. a fetch consumes the probe; verified to fail without the clear), TestRecordsCacheIsPerKeyOnce, TestRecordsCacheKeepsTimeoutsRetriesOtherErrors, TestWarmRecordsIsBounded (-race), TestHangingRemoteDegradesOnlyItsRepository, TestWarmedInventoryEqualsSequential. make test green (processgroup fails only in sandbox).; review verdict: SHIP
- 2026-10-02: flow upgraded quick → full — 147 added lines in code files (limit 100); an earlier round of this close already ran the full review

Filed while designing ariadne#289 at the operator's direction ("it's already a
bit slow"); to land before #289 resumes. Trace: 5.35s `ls-remote`, 5.13s
`fetch`, rest local.

Implemented. Measured on this machine, alternating builds, same fleet (46 rows,
27 repositories): before 17.83s / 17.31s, after 6.02s / 5.81s; JSON output
byte-identical. After-trace: 611 git processes, 15 `ls-remote` (concurrent),
0 `fetch` (every tracker unchanged), 0 `maintenance`. Tests:
`TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged` (real git, fetch counted via
GIT_TRACE, moved tip fetched, probe tip used once, --no-auto-maintenance),
`TestRecordsCacheIsPerKeyOnce`, `TestRecordsCacheKeepsTimeoutsRetriesOtherErrors`,
`TestWarmRecordsIsBounded` (gated, -race), `TestHangingRemoteDegradesWithinTheDeadline`
(ssh transport that sleeps → unknown naming the deadline in ~0.6s). Fixture
note: the sandbox sets GIT_SSH_COMMAND, which overrides core.sshCommand. A
timed-out read reports unknown, not stale: the local fallback read shares the
expired context.

Close review round 1 (BR-1..4) addressed: plan revised to the as-built deadline
and fetch-skip scope and Core-concepts names; issue-tracker atlas describes the
probe-then-maybe-fetch read; `TestWarmedInventoryEqualsSequential` and
`TestHangingRemoteDegradesOnlyItsRepository` (real tracked repos, bare origins)
replace the by-hand checks; any fetch clears the probed tip. One `make test` run
saw `TestPlanningReviewCompetingResults` fail once (multi-process CLI test with
a 25s budget, under shard load); it passed 3/3 alone, 3/3 under -race, and in
the next full run — no shared TrunkFile is involved.
