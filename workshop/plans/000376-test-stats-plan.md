# ariadne#376 — recorded suite statistics

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Tasks below are checkable and the SDLC close gate supplies the fresh-eyes review.

**Goal:** Make `sdlc test-stats` show where suite time and failures accumulate across the local fleet, with measured coverage rather than invented historical values.

**Architecture:** The local-check runner writes one compact, immutable result per suite attempt; the command reads those results and surviving legacy logs from sibling repositories. Parsing and aggregation are pure functions. Filesystem, Git discovery, clock, and terminal output stay at the edges. This design extends the original “from what already exists” wording because the current recorder discards retry and queue history; the TL approved that scope on 2026-10-14.

**Tech stack:** Go/Cobra in `cmd/sdlc`, existing local-check logs and fleet-root helpers, standard-library JSON/filesystem APIs. No changes to `.sdlc/`, CI, or suite commands.

## Core concepts

| Pure entity | Lives in | Status |
| --- | --- | --- |
| `SuiteAttempt` — one run's immutable measurements and named failures | `cmd/sdlc/teststats_parse.go` | new |
| `SuiteLogFacts` — package/shard durations and failure names parsed from one log | `cmd/sdlc/teststats_parse.go` | new |
| `RepoStats` — quantiles, coverage, slow work and flake classifications | `cmd/sdlc/teststats.go` | new |

Integration points: `runLocalCheck` measures queue acquisition and execution, parses its completed log, and atomically saves an attempt under the Git common directory before posting GitHub status. `sdlc test-stats` discovers eligible sibling repositories through the existing fleet-root helpers, reads each common-dir store and surviving old logs, then renders text or JSON. The command never runs a suite or calls GitHub.

An attempt records schema version, repository, exact suite SHA, check name, slot, agent/model/effort when known, UTC start/end time, baseline-on-main flag, queue seconds, execution seconds, exit/timeout, log digest and path, failed package/test identities, and package/shard durations. Its exact head, exit and log hash are sufficient for #380 to reconstruct the status payload after a failed POST; the record is committed to disk before that POST, and a failed POST does not erase it. Agent identity uses #367's `Agent: agent/model/effort` trailer at the exact head when available, with nullable fields for older commits; slot comes from the verified workspace identity. A successful later suite on the same repo/SHA confirms each named failure in an earlier attempt as a flake; a failed diagnostic baseline on main is listed separately as a suspected flake. Confirmed flake rate is failing attempts later followed by a success on the same head divided by all fully recorded suite attempts. An unknown field remains null/“n/a”, not zero. Old per-SHA logs cannot recover retry history or queue waits; pair's old logs cannot recover total wall time. Legacy rows are labelled partial and excluded from flake-rate and missing-metric denominators.

The writer retains compact attempt files and every failed attempt's immutable log for 30 days, pruning both on each write. Successful attempts retain their compact record; their existing named log remains subject to 20-head pruning. A retry uses a unique log while it runs, so it cannot overwrite another attempt's failure evidence; the legacy named log can then be updated for compatibility. Save the completed record with atomic rename only after its referenced failed log is durable, and clean crash residue on the next write. Reader memory is bounded by the retained 30-day store and line-streamed legacy logs (ARCH-CONSTRAINTS, ARCH-FUNERAL). An interrupted suite before completion has no completed-attempt record; a completed suite whose GitHub status post fails still has its local measurement. No background process is introduced (ARCH-ORDER).

The parser is shared by recorder and legacy reader (ARCH-DRY). It accepts ariadne's `packages`/`shardN` summaries and pair's Go `ok package seconds` lines, plus the failure forms they already emit. These are distinct measurement kinds: a shard is never mislabeled as a package. Parse errors and unavailable metrics are surfaced as coverage notes (ARCH-PURPOSE). Percentiles use nearest-rank over measured samples; top five slow work units rank by p90 with sample counts.

Text shape (illustrative values):

```text
repo      runs  wall p50/p90  queue p50/p90  confirmed flakes  coverage
ariadne     12  10m/13m       5m/48m                 1/12     wall 12/12 queue 6/12
pair        19  11m/15m       2m/30m                 2/19     wall 8/19 queue 8/19
  ariadne slow: shard7 640s (n=12), packages 300s (n=12)
  pair slow: github.com/xianxu/pair/cmd/internal/wrapcmd 57s (n=19)
  flakes: TestExample (failed then passed on SHA abc...), TestBaseline (main baseline failure)
```

`--json` emits the same fields with numeric seconds, sample counts and explicit nulls. `--since 7d` is the default; `--repo` filters by repository name. The default discovers local sibling repositories with retained suite attempts/logs, including ariadne and pair from any ariadne slot. The output states that old records lack some fields rather than implying a complete 10-14 baseline.

## Tasks

### 1. Pin parsing and aggregation before IO

**Files:** create `cmd/sdlc/teststats_parse.go`, `cmd/sdlc/teststats.go`, and matching `_test.go` files with compact fixture logs from both repositories.

- [ ] Test ariadne shard summaries, pair Go package lines, named failures, malformed/partial lines, and duplicate failure suppression. Confirm the parser fails those assertions before implementation.
- [ ] Test same-head fail→pass and main-baseline classifications, sample coverage, nearest-rank p50/p90, slow-unit ordering, window cutoff, and repository filter with fixed clock and records. Confirm red, then implement the pure functions.
- [ ] Run `go test ./cmd/sdlc -run 'TestSuiteStats|TestSuiteLog' -count=1` and commit the pure core.

### 2. Preserve each suite attempt

**Files:** modify `cmd/sdlc/localcheck_run.go`, `cmd/sdlc/localchecks.go`; create `cmd/sdlc/teststats_store.go` and focused tests in `cmd/sdlc/localcheck_run_test.go`/`cmd/sdlc/teststats_store_test.go`.

- [ ] Test a failing suite followed by a successful same-head retry: two immutable records and the failed log survive even if the old named log changes. Test real measured queue wait with a controlled lock helper, baseline flag, exact SHA/exit/hash on status-post failure, slot and nullable #367 identity, atomic rename/crash residue, and 30-day pruning of metadata and failed logs.
- [ ] Measure acquisition wait around `acquireSuiteSlot`; extend the local run record; run into a unique attempt log, preserve failed logs, parse the completed log and save the compact attempt before the GitHub status call. A stats-write failure emits an explicit warning but does not turn a passing suite into a failed check.
- [ ] Run targeted local-check/store tests and commit.

### 3. Read the fleet and render the command

**Files:** create `cmd/sdlc/teststats_cmd.go`, `cmd/sdlc/helptext/test-stats.md`, command tests; modify `cmd/sdlc/main.go`, `README.md`, `atlas/workflow/sdlc-binary.md` and `atlas/index.md` if needed for a new map entry.

- [ ] Test discovery from a numbered slot via a synthetic sibling fleet, `--since 7d`, `--repo`, JSON and text shapes, and an honest mixed legacy/new history. Ensure no GitHub call or repository mutation.
- [ ] Register `sdlc test-stats`; reuse fleet normalization and common-dir discovery. Stream legacy logs only when no attempt record covers that log; render explicit coverage and flake classifications.
- [ ] Run focused command tests, `make -k tools`, and the real read-only `sdlc test-stats --since 7d` against ariadne and pair. Capture the actual output as evidence, not the illustrative table above.

### 4. Boundary and landing

- [ ] Run `git diff --check`, the recorded `sdlc local-check run --name suite` on the pushed head, and `sdlc close --issue 376 --verified ...` for the fresh review. Address all blocking findings.
- [ ] For the operator-loaded CLI, stage the built head in :0 for a live check only with the hub's approval; clear the stage or let landing clear it. Open the PR, let CI and `sdlc merge --yes` gate it, and report the PR, merge commit, measured output and any baseline gaps to the hub.

## Revisions

### 2026-10-14 — TL approval and recovery/identity requirements

The TL approved per-attempt instrumentation and 30-day retention. The record now carries verified slot and nullable agent/model/effort compatible with #367, and exact-head/exit/log-hash evidence for #380's later status re-upload. Failed attempts retain immutable logs across same-head retries; the writer's crash and retention tests cover both JSON and logs. This supersedes the earlier pending-scope and overwritten-failure-log wording.
