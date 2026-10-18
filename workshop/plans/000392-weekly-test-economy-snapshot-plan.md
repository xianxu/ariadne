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

One versioned JSON file under `brain/memory/work/devinfra/test-economy/weekly/<end-date>.json` contains: interval (with no wall-clock generation field that would destabilize reruns); source ref/record counts and warnings; each repo's completed/fully-recorded attempts, wall and queue p50/p90 with sample counts, five slowest packages with counts, named failed tests classified as confirmed same-head fail→pass, suspected main-baseline, or unknown; landed-issue count; exact-attributed full-suite attempt count (pass/fail/repeat counts), unassignable count and nullable attempts-per-landed-issue ratio. Failed attempts keep their own IDs; status retries do not become new attempts. Source records expire locally after 30 days; the brain snapshot does not. Its brain Git/gcrypt backup responsibility follows the brain charter; the command neither deletes source records nor performs a second retention sweep (ARCH-FUNERAL). The schema starts at version 1 so #402 can extend it deliberately.

The headline cohort is issues **landed in the seven-day interval**, selected by their confirmed landing event time. Its numerator includes every exact attempt completed from that issue's first enqueue through landing, even when enqueue and an attempt were before the seven-day boundary. The reader therefore loads retained attempts and events before the boundary; if the first enqueue or any earlier attempt evidence might have expired, cohort coverage is unknown and the ratio is null, not a partial numerator. A physical full-suite count selected strictly by attempt completion in the seven-day interval is a separate metric, never presented as the cohort numerator. The document shows the reconciliation buckets: attempts outside the week attributed to this week's landings, attempts inside the week belonging to pending/future landings, pre-enqueue attempts, and unassignable attempts. Thus the two populations need not have equal totals (ARCH-PURPOSE, ARCH-ORDER).

For each landed issue, a done tracker card's completion `landed_commit`, one `landed` event, and a reachable Git merge commit must agree. `validateLandingEvidence` checks exactly one enqueue/landed pair by ticket sequence and issue/branch, a nonzero PR on the landed event, strict enqueue <= landed time, no duplicate or conflicting sequence/issue/branch/PR, card completion bound to that merge SHA, and a merge with exactly two available parents. The enqueue has no PR field; its ticket sequence and issue/branch bind it to the landed event, whose PR identifies the landing. The completion's reviewed/evidence head must be reachable from the merge's issue parent, not its main parent. A malformed line, skipped JSON entry, missing card, conflicting event, missing Git object, unreachable merge or unknown first enqueue marks the affected cohort unknown before attribution (ARCH-SECURE). An attempt is assigned only if its full SHA lies in the issue parent's ancestry excluding the main parent **and** its completed time lies in that issue's enqueue-to-landed interval. A SHA shared with main or a squash/rebase landing without provable branch ancestry remains unassignable, never guessed from slot, branch name or nearest time. Pre-enqueue branch suites are diagnostic only. Zero confirmed landings produces a null ratio. Legacy logs without attempt ID or exact completion remain partial coverage and never count as a confirmed retry.

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

`WeeklyEvidenceReader` wraps filesystem, tracker fetch and bounded Git queries. Tests use real temp Git repositories, tracker cards and landing event files as a stateful fixture, then call the same reader as the CLI (ARCH-MOCK). It gathers each repo once per run, preserving warnings and source ref pins. Reuse the existing 10,000-attempt and 64 MiB legacy-log limits; cap the event input at 64 MiB and each Git command at 8 MiB and 30 seconds. An exceeded limit, offline fetch or cancellation yields named unknown coverage, never zero. A fixture remote exercises fetch/retry/cancellation through the production reader; a live ariadne/pair run checks conformance. Git work is bounded by the week's landed commits and candidate attempt SHAs, not every commit in every branch (ARCH-CONSTRAINTS). `BrainSnapshotWriter` takes final bytes only; tests inject a temp brain root and exercise interrupted temporary files and two concurrent writers. Its no-replace publish compares an existing final file byte-for-byte: identical rerun is a no-op, different bytes refuse with the existing path. It fsyncs a temp file and publishes atomically without overwriting the final file. A normal run removes its own temp; a later run removes stale temp files older than one day in this command-owned weekly directory. Budget under 50 KiB per repo-week; #402's weekly resource gauge tracks observed growth (ARCH-FUNERAL).

## Visible states and recovery

Actions: rerun the same fixed window, repair a missing input before first publication, or run the next complete week. Published history is immutable; a correction requires a separately reviewed issue and versioned replacement, not an in-place rewrite.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete sources, first publication | published with coverage | run next week | none |
| Source unreadable or evidence unassignable | published with explicit partial/unknown fields | repair source before first publication, otherwise run next week | numeric comparison for this week |
| Temp write interrupted before final path | no published snapshot | rerun same window | none |
| Final path exists with identical bytes | published, stable rerun | run next week | none |
| Final path exists with different bytes | conflict, unchanged published snapshot | review a correction issue | late evidence absent from this week |

Non-goals: reconstructing attempts after local records expire, correcting a published snapshot in place, introducing a daily scheduler, changing local-check/merge gates, or setting package budgets. Missing history remains unknown (ARCH-STATE).

## Risk-based verification

| Function | Adversarial input and mechanical guard |
|---|---|
| `newWeeklyWindow` | Random UTC offsets/boundary instants; exact half-open membership and future-end refusal. |
| `aggregateWeeklyRepo` | Partial/legacy/repeated records and reordered input; compare each metric's sample IDs and stable serialization, including unknown failed-test class. Extend `aggregateSuiteAttempts` where it currently drops that class (`cmd/sdlc/teststats.go:148`). |
| `validateLandingEvidence` | Malformed, duplicate, reordered and contradictory event/card/parent tuples; require one consistent ticket and reachable bound merge before emitting an issue interval. |
| `attributeWeeklyAttempts` | Mixed-cohort records, shared-main SHA and missing ancestry; set-based oracle checks each ID at most once and makes unknown cohort coverage null. |
| `readWeeklyEvidence` | Stateful temp Git remote/tracker plus truncated logs and missing objects; bounded production reader reports source ref and unknown instead of zero. |
| `publishBrainSnapshot` | Concurrent, interrupted and different-byte writers; atomic no-replace and fsync preserve exactly one immutable final file. |

The live ariadne/pair snapshot verifies the Git/tracker/landing-line model against production sources; a second run with the same pinned inputs must yield identical bytes. Fleet rollout is proven by running from both repositories' checkouts after propagation, not merely by compiling ariadne. The hub confirms the exact link in the ops project after the path handoff.

## Tasks

### 1. Pin the pure window, schema and aggregation

- [ ] Test `newWeeklyWindow` and `aggregateWeeklyRepo` by the strategies above in `cmd/sdlc/weeklystats_test.go`; observe red.
- [ ] Add minimal pure types/aggregation in `cmd/sdlc/weeklystats.go`, reusing `suiteQuantiles` and `aggregateSuiteAttempts` where semantics match. Run targeted tests to green; commit.

### 2. Prove exact landing attribution

- [ ] Test `validateLandingEvidence` and `attributeWeeklyAttempts` by the strategies above in `cmd/sdlc/weeklyattribution_test.go`; observe red.
- [ ] Implement the pure join in `cmd/sdlc/weeklyattribution.go`. Count an attempt ID once; keep failed and repeated attempts as separate IDs. Run targeted tests to green; commit.
- [ ] Test `readWeeklyEvidence` through a temp-Git remote/tracker/landing fixture, including malformed inputs and unavailable objects; implement the bounded reader and run targeted tests; commit.

### 3. Publish one immutable brain artifact

- [ ] Test `publishBrainSnapshot` by the adversarial strategy above, with a read-only source tree; observe red.
- [ ] Implement `BrainSnapshotWriter` and the `sdlc test-stats snapshot` command with fixed-window flags. Use one JSON file, no second scheduler or writer. Run focused tests; commit.
- [ ] Update `cmd/sdlc/helptext/test-stats.md` and `atlas/workflow/sdlc-binary.md` (plus index if needed) with the command, brain path, coverage meaning, weekly default, project-link ownership and retention/backup. Run `make -k build`, focused tests, `git diff --check`; commit and push.

### 4. Verify, capture and land

- [ ] Run `sdlc local-check run --name suite` on the pushed head. In the live run, inspect ariadne and pair source counts, warnings, attempt IDs, landing cards and ratio provenance; save a fixed-window snapshot under brain, commit it there via the normal capture path, and report its link to the hub for the TL-owned ops project.
- [ ] Record command, brain path, source coverage, tests and limitations in #392's Log. Close with the mandatory review, then PR and merge through the landing line. Report the verified merge commit and snapshot link to the hub.

## Revisions

2026-10-17 — Plan-quality PQ-1/PQ-2/PQ-3 and minor findings: fixed the landed-in-week cohort and cross-week reconciliation, added a semantic evidence validator before attribution, replaced case inventories with function-level adversarial strategies, and specified rollout, limits and artifact cleanup. No implementation has begun.

2026-10-17 — The live ariadne pilot showed shard durations can hide package rows in the generic slow-unit top five. The snapshot now carries a distinct package-only top five and package-duration coverage. It also inventories main merge timestamps to detect a done card whose weekly landing event is missing, rather than silently shrinking the denominator. The dated path requires midnight UTC to avoid two windows writing the same date.

2026-10-17 — The full-fleet pilot exposed non-SDLC sibling repositories in generic fleet discovery and a weak event-to-PR binding. The snapshot filters for the vocabulary-declared issue home and verifies that a GitHub merge subject's PR and branch match the landing event. Repositories without source support remain out of the weekly artifact; pre-landing-line merges inside the first week's cohort are explicit unknown coverage, not a guessed baseline.

2026-10-17 — A second fleet pilot showed brain capture repos can retain an issue-home directory. The discovery predicate now applies brain's `.brain/config.md` spine exclusion before the issue-home test, so the output never treats a capture store as an SDLC repo. This is a coverage correction discovered before the queued suite began; the old queued run was cancelled and will be replaced on the corrected head.

2026-10-17 — A final adversarial review tightened the same evidence-boundary rule: every earlier enqueue used to widen the cohort interval must have one matching refusal, and a legacy/partial attempt whose SHA is on the cohort branch makes the ratio unknown. Both prevent a seemingly precise but undercounted numeric baseline. The queued suite on the preceding head was cancelled before acquisition and will run on this commit.

2026-10-17 — Immutable weekly publication must represent the fleet-wide denominator. Removed the snapshot subcommand's optional repo filter: a partial run could otherwise claim the same dated path and block the full artifact forever. Per-repo inspection remains available in the read-only `sdlc test-stats --repo` command; snapshot publication is all-or-nothing for the discovered SDLC fleet.

2026-10-17 — Hub review approved the earlier evidence-boundary revisions and required glanceable headline coverage. Each repository now renders the ratio beside covered and observed landed-issue counts, unknown count and top three coverage reasons. The observed count is a lower bound when source reads are incomplete; a null ratio gives zero fully covered issues. The queued suite on the prior head had not started and was cancelled before this TDD change.

2026-10-17 — The recorded suite on the headline head found the context-guard test's only failure: `weeklyGit` minted `context.Background()` for a nil caller. All production callers already thread the command context, so the nil case now refuses before launching Git; a targeted test proves the contract. The failed suite record and log are retained, and a new exact-head suite will verify the fix.

2026-10-17 — Whole-issue review round 1 returned REWORK (BR-1..BR-6). The hub approved a consolidated re-plan. For ratio certification, compute a conservative source-history floor at collection from the later of the attempt and landing-event 30-day retention cutoffs and the first retained landing event's UTC day. A landed issue's tracker `created` day must be strictly later than that floor day; missing or unparsable creation, empty event history, or an older issue makes coverage unknown. This proves an issue could not have had an expired earlier admission or branch attempt; it intentionally excludes old long-running issues. Checking only the earliest retained enqueue would fail because older admissions and suites can already have been swept. Validate every prior ticket against the *original* successful enqueue timestamp, independent of input order, then take the minimum validated timestamp. Use `suiteAttemptCompletedAt` for both physical and reconciliation week selection; legacy records stay unassignable, never exact-attributed. Pass the requested window into evidence collection, inventory only its main merges, and fetch graphs only for that cohort's landed events/cards, so unrelated historical failures cannot poison coverage. Add permanent red/green regressions for all four classes plus README invocation, bounds and coverage semantics.

2026-10-17 — The hub retained post-review immutable publication despite BR-5. A fresh full-fleet temporary pilot on the fixed head will be logged with its output hash, source coverage and stable rerun; the same binary will be exercised from ariadne and pair checkouts. The real brain file is written by the same command only after a SHIP close review, then committed and linked by the TL before merge. If BR-5 repeats solely because this ordered final capture is still pending, dispute it with the hub's immutability rationale and the reproducible pilot evidence rather than publishing an unreviewed immutable artifact.

2026-10-17 — The hub tightened the history floor: use the later of the retention cutoff, the first retained landing event, and the first retained structured suite attempt, comparing UTC dates. If no structured attempt remains, the recorder cannot prove whole-issue coverage; report unknown rather than a numeric zero. The first retained attempt matters because a repo can adopt the structured recorder after its landing log began.

2026-10-18 — Close round 2 disposed BR-1/2/3/4/6 but found BR-7: separate done cards and tickets can claim one landing identity and double the denominator. The hub approved one cohort-wide identity check over in-week landed candidates: issue, ticket sequence, branch, PR and card completion merge must each be unique. Any shared identity makes every claimant unknown before attribution; do not choose a winner without authority. Test each collision in both event orders, including a shared merge, while preserving the existing per-record join. A simpler rule that merely picks the first valid record fails because event order is not evidence of ownership. BR-5 remains a publication-order dispute: a hashed 15-repo temporary pilot and identical ariadne/pair rerun are recorded; the immutable brain artifact is published only after SHIP, as the hub directed. If close round 3 finds another attribution hole, stop patching and re-plan the metric itself rather than extending this join again.

2026-10-18 — A pre-round-3 adversarial test exposed another attribution hole, so the queued suite was cancelled before acquisition and the hub approved a metric re-plan. One classifier now decides numerator authority for *every retained attempt* against the fully validated landed cohort: zero exact SHA-and-admission-interval matches is outside the numerator; one counts once; more than one always makes coverage unknown and the ratio null. Partial/duplicate records on a cohort SHA likewise make coverage unknown. The seven-day physical-window flag is only for reconciliation buckets and cannot decide numerator completeness. A property-style invariant tests one-to-many branch ancestry across attempt times and candidate orders: no attempt can count toward two landings or leave a numeric ratio when its ownership is ambiguous. Dropping issue/ticket joins and using the merge PR alone is simpler superficially but cannot work: `SuiteAttempt` records SHA, not PR, and the enqueue interval would be unproved. This re-plan preserves the strict baseline rather than replacing it with a guessed physical-suites/landings division.
