# ariadne#376 — recorded suite statistics

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. Tasks below are checkable and the SDLC close gate supplies the fresh-eyes review.

**Goal:** Make `sdlc test-stats` show where suite time and failures accumulate across the local fleet, with measured coverage rather than invented historical values.

**Architecture:** The local-check runner writes one compact, immutable result per suite attempt; the command reads those results and surviving legacy logs from sibling repositories. Parsing and aggregation are pure functions. Filesystem, Git discovery, clock, and terminal output stay at the edges. This design extends the original “from what already exists” wording because the current recorder discards retry and queue history; the TL approved that scope on 2026-10-14.

**Tech stack:** Go/Cobra in `cmd/sdlc`, existing local-check logs and fleet-root helpers, standard-library JSON/filesystem APIs. No changes to `.sdlc/`, CI, or suite commands.

## Core concepts

| Concept | Lives in | Role |
| --- | --- | --- |
| `SuiteAttempt` — one run's immutable measurements and named failures | `cmd/sdlc/teststats.go` | PURE data entity; `teststats_store.go` owns INTEGRATION persistence |
| `SuiteLogFacts` — package/shard durations and failure names parsed from one log | `cmd/sdlc/teststats_parse.go` | PURE parser output |
| `RepoStats` — quantiles, coverage, slow work and flake classifications | `cmd/sdlc/teststats.go` | PURE aggregate output |

Integration points: `runLocalCheck` measures queue acquisition and execution, parses its completed log, and atomically saves an attempt under the Git common directory before posting GitHub status. `sdlc test-stats` discovers eligible sibling repositories through the existing fleet-root helpers, reads each common-dir store and surviving old logs, then renders text or JSON. The command never runs a suite or calls GitHub.

An attempt records schema version, repository, exact suite SHA, check name, slot, agent/model/effort when known, UTC start/end time, baseline-on-main flag, queue seconds, execution seconds, exit/timeout, log digest and path, failed package/test identities, and package/shard durations. Its exact head, exit and log hash are sufficient for #380 to reconstruct the status payload after a failed POST; the record is committed to disk before that POST, and a failed POST does not erase it. Agent identity uses #367's `Agent: agent/model/effort` trailer at the exact head when available, with nullable fields for older commits; slot comes from the verified workspace identity. A successful later suite on the same repo/SHA confirms each named failure in an earlier attempt as a flake; a failed diagnostic baseline on main is listed separately as a suspected flake. Confirmed flake rate is failing attempts later followed by a success on the same head divided by all fully recorded suite attempts. An unknown field remains null/“n/a”, not zero. Old per-SHA logs cannot recover retry history or queue waits; pair's old logs cannot recover total wall time. Legacy rows are labelled partial and excluded from flake-rate and missing-metric denominators.

The writer retains compact attempt files and every failed attempt's immutable log for 30 days, pruning both on each write. Successful attempts retain their compact record; their existing named log remains subject to 20-head pruning. A retry uses a unique log while it runs, so it cannot overwrite another attempt's failure evidence; the legacy named log can then be updated for compatibility. Save the completed record with atomic rename only after its referenced failed log is durable; reclaim stale crash residue only with dead-owner proof as specified below. Reader memory is bounded by the retained 30-day store and line-streamed legacy logs (ARCH-CONSTRAINTS, ARCH-FUNERAL). An interrupted suite before completion has no completed-attempt record; a completed suite whose GitHub status post fails still has its local measurement. No background process is introduced (ARCH-ORDER).

The parser is shared by recorder and legacy reader (ARCH-DRY). It accepts ariadne's `packages`/`shardN` summaries, pair's Go `ok package seconds` lines, and `go test -json` package/test events, plus the failure forms they emit. These are distinct measurement kinds: a shard is never mislabeled as a package. Parse errors and unavailable metrics are surfaced as coverage notes (ARCH-PURPOSE). Percentiles use nearest-rank over measured samples; top five slow work units rank by p90 with sample counts.

The existing `failingTests` in `cmd/sdlc/suitefailures.go:11` remains the single package-qualified failure extractor for both baseline diagnosis and test-stats. `parseSuiteLogFacts` handles timings and calls it for failures; one bounded log read feeds both rather than introducing another name grammar (ARCH-DRY).

### Publication, validation and operating envelope

Each run owns a cryptographically random attempt ID and one private in-flight directory under `<git-common-dir>/sdlc/test-stats/inflight/<id>/`. No global SDLC transaction lock is assumed: two `sdlc local-check run` processes can overlap. The state sequence is `inflight log → completed log and parsed facts → immutable published record → status POST`. On failure, move the closed log to `test-stats/failed/<id>.log` first; publish `test-stats/attempts/<id>.json` by an owner-named temporary write, file sync, and atomic rename only after that log is durable. On success, copy the closed log to the legacy `<sha>/suite.log` path for compatibility, with the copy temp under the same private in-flight directory, and publish the record with that path as a non-authoritative convenience; the exact SHA/exit/log hash in the immutable JSON are the recovery authority for #380. A status POST failure changes neither record nor failed log. The published JSON filename is the only commit marker readers enumerate. Identical-ID collision refuses rather than overwrites. Other writers can publish independently.

`pruneSuiteAttempts` removes only published records older than 30 days and their matching failed logs, never another process's in-flight directory. In-flight cleanup requires both age over 24 hours and positive proof that the owner PID is dead on this host; if liveness is uncertain, leave it and warn. A crashed writer after moving a failed log but before publishing JSON leaves an orphan; prune orphans and owner-named attempt publication temps older than 24 hours only when no live owner is recorded. Legacy copy temps live in the in-flight directory and leave with it after dead-owner proof. All cleanup uses generated ID filenames, never a path from JSON. The old 20-head log pruning touches only `local-checks/`, outside `test-stats/` (ARCH-ORDER, ARCH-FUNERAL). A controlled-order test pauses one writer before publication while another publishes and prunes, then proves both surviving records and failed logs; a crash test leaves an in-flight/orphan fixture and verifies safe cleanup after owner death.

`decodeSuiteAttempt` is the trust boundary for persisted JSON: cap reads at 1 MiB, require schema v1 and exactly one JSON object with every outcome/provenance field present and no duplicate keys, full hex SHA/log hash, recognized check name, UTC timestamps in order, nonnegative finite duration values, legal exit/timeout combination, and bounded names/counts. Unknown fields or versions and malformed/truncated/oversized records are skipped with a visible coverage diagnostic; they never enter quantiles, successful-run evidence or flake classification. Relative log references must be generated IDs or confined legacy `<sha>/<check>.log` paths after cleaning; a stats reader does not need to open them. Any future #380 reader must resolve under the store root and verify the saved hash before use. No record-derived path drives deletion (ARCH-SECURE).

The observed current retained log stores are 76 KiB in ariadne and 756 KiB in pair (measured 2026-10-14 with `du -sh`). Expected use is 2 repositories, tens of runs per week, and logs well under 1 MiB each; 30-day failed-log retention should ordinarily be tens of MiB. Parse one log at a time with a 64 MiB read limit and 1 MiB line limit; over-limit logs keep exact hash and raw failed evidence but timing/failure facts are marked unavailable with a warning. Cap reader intake at 10,000 record files and 64 MiB total JSON bytes per repository, stream directory batches, and report omitted records rather than a silently partial rate. Recorder overhead beyond suite execution is one log pass plus a small JSON write; warn visibly if persistence fails and preserve the existing check outcome. No new background worker or hard suite timeout is introduced (ARCH-CONSTRAINTS).

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

- [x] Test `parseSuiteLogFacts`: feed captured ariadne/pair snippets with malformed/partial lines, duplicate failures and package-vs-shard ambiguity; assert exact named units and that `failingTests` supplies the same failure names baseline diagnosis uses. Confirm red before implementation.
- [x] Test `aggregateSuiteAttempts`: permute fail→pass ordering on one exact head, main-baseline failures, missing fields and out-of-window records; mechanically assert nearest-rank quantiles, coverage denominators, flake classification, slow-unit ordering and repo filter against fixed expected values. Confirm red, then implement the pure functions.
- [x] Run `go test ./cmd/sdlc -run 'TestSuiteStats|TestSuiteLog' -count=1` and commit the pure core.

### 2. Preserve each suite attempt

**Files:** modify `cmd/sdlc/localcheck_run.go`, `cmd/sdlc/localchecks.go`; create `cmd/sdlc/teststats_store.go` and focused tests in `cmd/sdlc/localcheck_run_test.go`/`cmd/sdlc/teststats_store_test.go`.

- [x] Test `decodeSuiteAttempt` against truncated/oversized/unknown-version JSON, traversal paths, illegal outcomes and valid nullable identity; assert rejected records never reach aggregation. Confirm red, then implement validation.
- [x] Test `publishSuiteAttempt`/`pruneSuiteAttempts` with two controlled concurrent writers, one paused after failed-log move while the other publishes and prunes; assert no active attempt is removed, crash residue is reclaimed only with dead-owner proof, 30-day metadata/log pruning is paired, and IDs cannot overwrite. Confirm red, then implement the store.
- [x] Test `runLocalCheck` failing then successful on the same SHA: two immutable records and the failed log survive even when the legacy named log changes. Use the existing stateful lock helper/CI fake to verify measured queue wait, baseline flag, exact SHA/exit/hash on status-post failure, and slot plus nullable #367 identity.
- [x] Measure acquisition wait around `acquireSuiteSlot`; extend the local run record; run into a unique attempt log, preserve failed logs, parse the completed log and save the compact attempt before the GitHub status call. A stats-write failure emits an explicit warning but does not turn a passing suite into a failed check.
- [x] Run targeted local-check/store tests and commit.

### 3. Read the fleet and render the command

**Files:** create `cmd/sdlc/teststats_cmd.go`, `cmd/sdlc/helptext/test-stats.md`, command tests; modify `cmd/sdlc/main.go`, `README.md`, `atlas/workflow/sdlc-binary.md` and `atlas/index.md` if needed for a new map entry.

- [x] Test `collectSuiteStats`/`renderSuiteStats` from a numbered slot in a synthetic sibling fleet: vary record age, malformed input, legacy/new overlap and repository filter; assert `--since 7d`, JSON/text, warnings and unknown coverage, plus no GitHub call or repository mutation.
- [x] Register `sdlc test-stats`; reuse fleet normalization and common-dir discovery. Stream legacy logs only when no attempt record covers that log; render explicit coverage and flake classifications.
- [x] Run focused command tests, `make -k tools`, and the real read-only `sdlc test-stats --since 7d` against ariadne and pair. Capture the actual output as evidence, not the illustrative table above.

### 4. Boundary and landing

- [ ] Run `git diff --check`, the recorded `sdlc local-check run --name suite` on the pushed head, and `sdlc close --issue 376 --verified ...` for the fresh review. Address all blocking findings.
- [ ] For the operator-loaded CLI, stage the built head in :0 for a live check only with the hub's approval; clear the stage or let landing clear it. Open the PR, let CI and `sdlc merge --yes` gate it, and report the PR, merge commit, measured output and any baseline gaps to the hub.

## Revisions

### 2026-10-14 — TL approval and recovery/identity requirements

The TL approved per-attempt instrumentation and 30-day retention. The record now carries verified slot and nullable agent/model/effort compatible with #367, and exact-head/exit/log-hash evidence for #380's later status re-upload. Failed attempts retain immutable logs across same-head retries; the writer's crash and retention tests cover both JSON and logs. This supersedes the earlier pending-scope and overwritten-failure-log wording.

### 2026-10-14 — plan-quality round 1

PQ-1: defined private in-flight ownership, published JSON as commit marker, failed-log ordering, concurrent writers and safe orphan cleanup. PQ-2: defined bounded schema/path validation and skip-with-coverage behavior. PQ-3: named risky functions and adversarial test strategies. Also addressed PQ-4 by reusing `failingTests`, and PQ-5 with measured log-store size plus explicit reader/recorder budgets. The implementation tasks above supersede their earlier case-only wording.

### 2026-10-14 — aggregate read budget

During the first queued suite wait, noticed that 10,000 individually bounded 1 MiB records could still require 10 GiB. Added a 64 MiB total JSON intake cap and batched directory reading, with visible coverage loss, before running the suite on the final head (ARCH-CONSTRAINTS).

### 2026-10-14 — required outcome fields

During the next queued suite wait, found that absent or duplicate JSON `exit` could silently become a successful run. The decoder now requires all provenance/outcome fields, rejects duplicates and nulls for required scalars, and leaves only measured queue/wall and optional identity nullable. This closes the false-flake-evidence class in PQ-2 (ARCH-SECURE).

### 2026-10-14 — close review round 1

BR-1 added explicit `go test -json` pass/fail/malformed fixtures and bounded parser warnings to the log-format contract. BR-2 added a mixed-check fixture: only a recorded suite attempt suppresses a legacy suite log at the same SHA. BR-3 corrected the `SuiteAttempt` location and marked PURE entities versus INTEGRATION persistence. BR-4 named attempt publication temps by owner ID and covered old dead-owner residue, while legacy copy temps use the private in-flight directory. These changes supersede the earlier text-only parsing and same-directory publication-temp descriptions (ARCH-PURPOSE, ARCH-FUNERAL).

### 2026-10-14 — close review round 2

BR-5 tightened the persisted-input boundary: every top-level attempt field and nested duration field must use its exact canonical JSON name, before Go's case-insensitive struct decoder runs; aliases, including conflicting outcome keys, are rejected. BR-6 completed the format-by-outcome matrix: ariadne shard summaries, plain Go package `ok` and `FAIL` summaries, and `go test -json` package `pass` and `fail` events all retain available duration and failure evidence. BR-7 enforces the planned 1 MiB per-line limit before either failure extraction or timing parsing; an oversized line makes all parsed facts unavailable with a warning, while the raw log and hash remain available. The new boundary-size and outcome fixtures are the regression contract (ARCH-SECURE, ARCH-PURPOSE, ARCH-CONSTRAINTS).

### 2026-10-14 — close review round 3

BR-8 found that Go's scalar decoder also maps nested JSON nulls to zero values. Required scalar fields reject null at every depth before decoding; only queue/wall measurements and agent/model/effort identity are nullable by schema. A record with a null duration kind, name or seconds now produces a coverage diagnostic and contributes no samples, and duration kind is validated against the emitted measurement kinds (ARCH-SECURE, ARCH-PURPOSE).
