---
collaboration-mode: discussion
---

# Weekly test-economy snapshot implementation plan

> **For agentic workers:** Consult AGENTS.md §3 for execution approach. Check steps in place; commit design decisions before long waits.

**Goal:** Produce one immutable seven-day brain snapshot that measures local full-suite cost and its exact attribution to landed issues across the SDLC fleet.

**Architecture:** Reuse the validated per-attempt reader and quantile/flake calculations from `sdlc test-stats`. Join attempt IDs to a landed issue only when its tracker completion, landing-line interval and Git commit graph agree. Render one versioned JSON document for both machine comparisons and human review, then publish it by no-replace atomic file creation in brain; the ops TL adds a stable project link.

**Tech stack:** Go/Cobra, existing `SuiteAttempt`, `RepoStats`, `tracker`, `landingline`, Git commit graph, brain's encrypted Git capture repo.

## Measurement contract

The command is `sdlc test-stats snapshot --end <UTC RFC3339> --brain-root <path>`; `--end` defaults to the latest completed weekly UTC boundary. It measures the half-open interval `[end-7d, end)` by attempt `finished_at` and landing event time, and refuses an end in the future. A fixed end makes reruns comparable. The first pilot may use an explicit completed UTC-day end before a full post-#376 week exists; the document labels its observed coverage rather than implying seven days of records.

One versioned JSON file under `brain/memory/work/devinfra/test-economy/weekly/<end-date>.json` contains: interval (with no wall-clock generation field that would destabilize reruns); source ref/record counts and warnings; each repo's completed/fully-recorded attempts, wall and queue p50/p90 with sample counts, five slowest packages with counts, named failed tests classified as confirmed same-head fail→pass, suspected main-baseline, or unknown; landed-issue count; exact-attributed full-suite attempt count (pass/fail/repeat counts), unassignable count and nullable attempts-per-landed-issue ratio. Failed attempts keep their own IDs; status retries do not become new attempts. Source records expire locally after 30 days; the brain snapshot does not. Its brain Git/gcrypt backup responsibility follows the brain charter; the command neither deletes source records nor performs a second retention sweep (ARCH-FUNERAL).

For the headline ratio, use #390's landing window: the first recorded enqueue for an issue through its confirmed landing. A landed issue requires a done tracker card with completion `landed_commit`, a matching landing-line `landed` event (issue, PR and time), and a Git merge commit whose issue branch parent is present. An attempt is assigned only if its full SHA lies in that issue branch's ancestry excluding the merge commit's main parent **and** its completed time lies in the window. A SHA shared with main, a squash/rebase landing without provable branch ancestry, a missing enqueue/landed event, malformed card/log, or a commit not available locally is unassignable/unknown, never guessed from slot, branch name or nearest time. The summary shows the physical full-suite count in the window separately, so an attribution gap cannot improve the ratio. Pre-enqueue branch suites are a separate diagnostic count, not headline integration-tax attempts. Zero confirmed landings produces a null ratio, not zero. Legacy logs without attempt ID or exact completion remain partial coverage and never count as a confirmed retry.

`sdlc test-stats snapshot` discovers every SDLC repo in the fleet as `test-stats` does; the first live check requires ariadne and pair. It reads each repo's Git common dir, attempts, landing events, tracker and local main history. It may fetch tracker/main for fresh read-only evidence, with bounded calls and a declared source ref; offline or unreadable sources yield per-repo unknown coverage rather than a fabricated zero. It writes no issue/process artifact in brain. The TL owns the ops project file and adds the exact brain link after the worker reports it (ARCH-DRY).

## Core concepts

| Name | Kind | Lives in | Status |
|---|---|---|---|
| `WeeklyWindow` | PURE | `cmd/sdlc/weeklystats.go` | new |
| `WeeklyRepoStats` | PURE | `cmd/sdlc/weeklystats.go` | new |
| `AttemptAttribution` | PURE | `cmd/sdlc/weeklyattribution.go` | new |
| `WeeklySnapshot` | PURE | `cmd/sdlc/weeklystats.go` | new |
| `WeeklyEvidenceReader` | INTEGRATION | `cmd/sdlc/weeklystats_inputs.go` | new |
| `BrainSnapshotWriter` | INTEGRATION | `cmd/sdlc/weeklystats_write.go` | new |
| `SnapshotCommand` | INTEGRATION | `cmd/sdlc/teststats_cmd.go` | modified |

`WeeklyWindow` validates the exact UTC bounds. `WeeklyRepoStats` reuses the existing `SuiteAttempt` and `suiteQuantiles` semantics rather than developing a second test-stats calculator (ARCH-DRY). `AttemptAttribution` accepts already-validated events, cards, attempt records and precomputed Git ancestry sets; it has no IO, and emits assigned IDs plus named reasons for unknowns. `WeeklySnapshot` is the stable versioned JSON shape; sorted repos, issues, tests and IDs make equal evidence render to equal bytes.

`WeeklyEvidenceReader` wraps filesystem, tracker fetch and bounded Git queries. Tests use real temp Git repositories, tracker cards and landing event files as a stateful fixture, then call the same reader as the CLI (ARCH-MOCK). It gathers each repo once per run, preserving warnings and source ref pins. Git work is bounded by the week's landed commits and candidate attempt SHAs, not every commit in every branch (ARCH-CONSTRAINTS). `BrainSnapshotWriter` takes final bytes only; tests inject a temp brain root and exercise interrupted temporary files and two concurrent writers. Its no-replace publish compares an existing final file byte-for-byte: identical rerun is a no-op, different bytes refuse with the existing path. It fsyncs a temp file and publishes atomically without overwriting the final file.

## Visible states and recovery

Actions: rerun the same fixed window, repair a missing input before first publication, or run the next complete week. Published history is immutable; a correction requires a separately reviewed issue and versioned replacement, not an in-place rewrite.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete sources, first publication | published with coverage | run next week | none |
| Source unreadable or evidence unassignable | published with explicit partial/unknown fields | repair source before first publication, otherwise run next week | numeric comparison for this week |
| Temp write interrupted before final path | no published snapshot | rerun same window | none |
| Final path exists with identical bytes | published, stable rerun | run next week | none |
| Final path exists with different bytes | conflict, unchanged published snapshot | review a correction issue | late evidence absent from this week |

Covered: failed attempts, retries, malformed records, missing cards/events, stale Git objects, partial old records and zero landings. Not covered: reconstructing attempts after local records expire or retroactively changing an immutable snapshot; those data can be unknown (ARCH-STATE).

## Tasks

### 1. Pin the pure window, schema and aggregation

- [ ] Add tests in `cmd/sdlc/weeklystats_test.go` for half-open UTC windows, stable JSON ordering, p50/p90 with counts, failed/retried attempts, slow packages, confirmed/suspected/unknown named failures, missing old fields and zero landings. Run the targeted test and observe red.
- [ ] Add minimal pure types/aggregation in `cmd/sdlc/weeklystats.go`, reusing `suiteQuantiles` and `aggregateSuiteAttempts` where semantics match. Run targeted tests to green; commit.

### 2. Prove exact landing attribution

- [ ] Add tests in `cmd/sdlc/weeklyattribution_test.go` for one failed then passing attempt on the same issue head, a main-drift integration rerun, another issue's suite in the same time window, missing enqueue, missing done binding, shared main SHA, squash/unknown ancestry, duplicate attempt ID and zero denominator. Assert numerator, denominator, physical count and each unknown reason. Observe red.
- [ ] Implement the pure join in `cmd/sdlc/weeklyattribution.go`. Count an attempt ID once; keep failed and repeated attempts as separate IDs. Run targeted tests to green; commit.
- [ ] Add temp-Git fixture tests for `WeeklyEvidenceReader`: a tracker done card, merge parents, events and attempt files; include malformed data and a missing object. Implement the bounded reader and run targeted tests; commit.

### 3. Publish one immutable brain artifact

- [ ] Add tests for first write, equal rerun, changed rerun refusal, interrupted temp cleanup, concurrent no-replace writers and a read-only source tree. Observe red.
- [ ] Implement `BrainSnapshotWriter` and the `sdlc test-stats snapshot` command with fixed-window flags. Use one JSON file, no second scheduler or writer. Run focused tests; commit.
- [ ] Update `cmd/sdlc/helptext/test-stats.md` and `atlas/workflow/sdlc-binary.md` (plus index if needed) with the command, brain path, coverage meaning, weekly default, project-link ownership and retention/backup. Run `make -k build`, focused tests, `git diff --check`; commit and push.

### 4. Verify, capture and land

- [ ] Run `sdlc local-check run --name suite` on the pushed head. In the live run, inspect ariadne and pair source counts, warnings, attempt IDs, landing cards and ratio provenance; save a fixed-window snapshot under brain, commit it there via the normal capture path, and report its link to the hub for the TL-owned ops project.
- [ ] Record command, brain path, source coverage, tests and limitations in #392's Log. Close with the mandatory review, then PR and merge through the landing line. Report the verified merge commit and snapshot link to the hub.
