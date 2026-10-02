# Fleet Inventory Tracker Reads Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** bring `sdlc fleet inventory` back under ~9s here (from ~17s) without
changing anything it reports.

**Architecture:** Three independent levers on the read-only path, measured with
`GIT_TRACE` before and after:
1. **Skip the fetch when nothing changed.** The presence check's `ls-remote`
   already returns the remote tracker tip; when the local tracking ref already
   points at it, the snapshot reads that commit without fetching.
2. **No automatic maintenance** on tracker fetches (`--no-auto-maintenance`).
3. **Concurrent, bounded, deadline-limited reads.** The fleet records cache
   becomes per-key once (today one mutex is held across every load, so loads
   cannot overlap); inventory warms it for every tracked repository with a
   bounded worker pool before the row walk, each read under a deadline.

**Tech Stack:** Go; `internal/gitx` (`TrunkFile`), `internal/tracker`
(`Repository`, `LoadRecords`), `internal/fleet` (`repoRecords`,
`CollectInventory`).

---

## Measurements (baseline, this machine)

`GIT_TRACE` of one `fleet inventory --json` run: 18.7s, 638 git processes.
`ls-remote` 5.35s (15), `fetch` 5.13s (15, each followed by `maintenance run`),
local git ~8s (~13 commands per worktree × 46). After: recorded in the issue
Log with the same trace script.

## Decisions

- **Fetch-skip is exact, not a cache.** `RemoteExists` (the `ls-remote --refs
  --exit-code` presence probe) keeps its absent/transport-error semantics and
  additionally returns the tip OID it read. `Snapshot` uses that tip only when
  the local tracking ref resolves to the same OID, and only once (the tip is
  consumed by the next snapshot of the same `TrunkFile`). Otherwise it fetches
  as today. The answer is "as of the remote tip ls-remote saw", the same
  guarantee a fetch gives. Write paths (`refreshTip`, CAS push/confirm) are
  untouched: they always fetch.
- **`--no-auto-maintenance` on every `TrunkFile.fetch`.** Maintenance is
  housekeeping the operator's own git runs; sdlc's tracker fetches should not
  pay for it on every read (git ≥ 2.29; this machine has 2.54).
- **Per-key once cache.** `repoRecords` keeps one load per repository, but
  waiting callers block on that key only. Errors are not cached (as today).
- **Warm-up pass.** `CollectInventory` first computes each candidate
  repository's canonical root, then loads records concurrently for the ones
  that use the tracker (`tracker.CutOver`, local), at most 8 at a time, before
  the sequential row walk. A spelling mismatch between warm-up and walk keys
  costs only a sequential load, never a wrong answer.
- **Deadline.** Each warm-up read runs under `context.WithTimeout` (15s,
  injectable for tests). A read that misses it is not cached, and the row walk
  degrades that repository exactly as an unreachable remote does today
  (stale from the last fetch, else unknown), with the timeout in the reason.
  #288's network-budget caveat ("a fetch has no deadline") is resolved.
- **Local facts stay sequential** in this issue. They are ~8s and the next
  lever if needed (independent per worktree); not required for the target.
- **ARCH-ORDER:** the cache is the only shared state; per-key once makes
  concurrent loads of one key a single load, and a test drives two callers
  into one key under a barrier. Output order is the existing sort.
- **ARCH-FUNERAL:** nothing durable; the cache lives for the process.
- **ARCH-CONSTRAINTS:** at most 8 concurrent git network reads, each ≤15s;
  network round trips per tracked repository drop from 2 to 1 when its tracker
  is unchanged.

## Core concepts

### Pure entities
| Name | Lives in | Status |
|------|----------|--------|
| `parseLsRemoteTip` | `cmd/sdlc/internal/gitx/candidate.go` | new |

- **parseLsRemoteTip** — the tip OID from `ls-remote --refs` output for exactly
  the requested ref (validated with `parseObjectID`); pure.

### Integration points
| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `TrunkFile.probedTip` + `RemoteExists` / `Snapshot` / `fetch` | `cmd/sdlc/internal/gitx/{trunkfile,candidate,snapshot}.go` | modified | `git ls-remote`, `git fetch` |
| `recordsCache` / `newRecordsCache` / `loadRepoRecords` / `recordsReadDeadline` | `cmd/sdlc/internal/fleet/issues.go` | new | tracker load |
| `warmRecords` | `cmd/sdlc/internal/fleet/issues.go` | new | concurrent tracker loads |
| `trackedRoots` / `recordsReadConcurrency` (warm-up call in `CollectInventory`) | `cmd/sdlc/internal/fleet/inventory.go` | new | `tracker.CutOver` (local) |

*(Table rewritten to the as-built names at close, BR-2.)*

## Tasks (single pass — one boundary at close)

- [ ] **Fetch-skip.** Failing tests (real git, bare origin): a second
  `Snapshot` after an unchanged `RemoteExists` runs no `fetch` (counted through
  `GIT_TRACE`); after the remote advances it fetches and reads the new tip; a
  tip read for a different `TrunkFile`/consumed tip is never reused; absent
  remote branch and transport error keep their meanings. Implement
  `parseLsRemoteTip`, the tip hand-off, `--no-auto-maintenance`.
- [ ] **Per-key once + warm-up.** Failing tests: two callers on one key under a
  barrier load once; different keys load concurrently (a gated loader proves
  overlap and the bound of 8); warm-up reads each tracked repository exactly
  once and the inventory equals a sequential run's JSON.
- [ ] **Deadline.** Failing test: a remote whose transport hangs (an `ssh`
  command script that sleeps) makes that repository's rows `stale`/`unknown`
  with a timeout reason within the injected deadline, while another repository
  in the same fleet reads `present`.
- [ ] Measure before/after with the trace script; record both in the Log.
- [ ] `make test`; atlas (`issue-tracker.md` read path, `sdlc-binary.md`
  fleet budget); help text network note; `sdlc close`.

## Revisions

- 2026-10-02 — close review (BR-1..4), plan brought to current reality:
  - **Deadline (replaces the Decisions bullet).** A timed-out read *is* kept in
    the cache (the walk must not wait twice); other errors are retried. A
    timed-out repository reports `unknown` naming the deadline, not stale:
    the local fallback read shares the expired context.
  - **Fetch-skip scope (amends "write paths untouched").** The skip applies to
    every `Snapshot` right after `RemoteExists`, including verbs that later
    write; that is exact (the view is the tip the probe saw), and the
    compare-and-swap push still fetches. Any fetch clears the probed tip.
  - **Core concepts, as built:** `parseLsRemoteTip` (`internal/gitx/candidate.go`,
    new); `TrunkFile.probedTip` + `RemoteExists`/`Snapshot`/`fetch`
    (`internal/gitx/{trunkfile,candidate,snapshot}.go`, modified);
    `recordsCache` / `newRecordsCache` / `loadRepoRecords` / `warmRecords` /
    `recordsReadDeadline` (`internal/fleet/issues.go`, new); `trackedRoots` /
    `recordsReadConcurrency` + the warm-up call in `CollectInventory`
    (`internal/fleet/inventory.go`). There is no `warmRepoRecords`.
  - **Tests for the two Done-when claims previously shown by hand:**
    `TestWarmedInventoryEqualsSequential` and
    `TestHangingRemoteDegradesOnlyItsRepository` (fleet package, real tracked
    repositories with bare origins).
