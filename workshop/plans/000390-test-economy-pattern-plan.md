# Test Economy Pattern Design Plan

> **For agentic workers:** This is the operator's decision record, not authorization to implement. The implementing issues below each enter the normal claim → plan → change-code flow; use AGENTS.md Section 3 for their execution strategy.

**Goal:** Keep ariadne and pair test evidence trustworthy while reducing suite queue delay and preventing test-runtime growth as agents add tests.

**Architecture:** Preserve a full-suite landing guarantee on the exact integration head. Use targeted evidence during development, serial admission of ready PRs, measured soft package budgets, daily flake triage, and periodic evidence-led minimization. Roll out one independently testable concern per issue.

**Tech stack:** Existing `sdlc` local-check attempt records and merge gate, Git/GitHub PRs, Make, Go and Lua tests, ops project reporting.

---

## Recommendation for operator review

Adopt the five-part pattern below. Pilot **serial landing admission** before batching several PRs in one suite run. Keep one passing full suite on the exact integrated head that can land. A doc-only drift fast path is a separate, later optimization requiring an effective-input equivalence proof; the current ancestor allowance for close bookkeeping cannot establish that an arbitrary moved-main merge is harmless. Targeted-only or calendar-only landing would weaken main's health guarantee.

| Approach | Benefit | Cost / failure mode | Recommendation |
|---|---|---|---|
| Independent landing | Simple, every head tested | Main drift restarts contenders; #376 paid three integrations | Baseline only |
| Doc-only drift reuse | Cheap when main changed prose only | Must prove no change to code, tests, gates, generated inputs or effective suite | Separate, later issue |
| Serial merge admission | One integration owner avoids contenders testing against the same main; allows later batching | New state and recovery; a failed candidate delays those behind it | Pilot first |

### Evidence and baseline

Captured 2026-10-15 with `sdlc test-stats --since 7d` (nearest-rank p50/p90):

| Repo | Suite wall | Queue wait | Slow unit | Coverage limit |
|---|---:|---:|---:|---|
| ariadne | 531/621 s | 602/1769 s | shard2 p90 614 s | wall 19/20; queue 9/20; shards do not identify a package bottleneck |
| pair | 556/786 s | 0.1/1842 s | `cmd/internal/couchcore` p90 481 s, n=21 | wall and queue 24/24; long queue tail despite near-zero median |

`test-stats` confirms same-head fail→pass flakes only: ariadne 0/9 complete attempts, pair 0/24. Its 30-day local retention and older partial logs limit historical inference. Ops separately recorded 17/200 failing `couch-zellij-conformance` GitHub runs since 10-11 (8.5%, five on main); those include branch regressions and infrastructure/fixture faults, so they are **not** a measured flake rate. Do not combine those denominators. The project goal's “halve p50 queue wait” is meaningful for ariadne's measured subset, but pair needs p90 and coverage reported alongside p50. Rebaseline after seven days of complete records; retain the 10-14 anecdotal ~50-minute wait as context, not a statistical baseline.

At ariadne's observed median, three full integrations cost about 27 minutes of suite runtime before queueing. This illustrates the tax; it does not reconstruct #376's actual runs. Today `sdlc merge` re-fetches main, integrates it, requires `sdlc-local/suite` on the landing head, and retries moved main up to three times. It already releases the repo transaction lock while the suite runs, so the bottleneck is weighted suite capacity and repeated integration, not that lock.

## Core concepts for the implementing issues

These are proposed entities, not code in #390. Their precise source paths are settled by each implementing issue's plan.

| Pure entity | Intended home | Responsibility |
|---|---|---|
| `EvidenceTier` | ariadne `cmd/sdlc/` policy | Classify a checkpoint as targeted development, close review or exact-head landing; declare evidence without running tests. |
| `LandingCandidate` | ariadne merge-queue model | Bind repository, PR, issue head, observed main OID and integration OID; decide whether evidence is current. |
| `BudgetDecision` | ariadne test-budget model | Compare complete package durations with a soft p90 budget and coverage floor; return an explanation request, never green a failed test. |
| `FlakeCase` | ops triage record | Bind test/run/head to reproduction evidence, owner, issue, first-seen time and disposition. |
| `MinimizationCandidate` | repo-GC report | Record proposed removal, detection evidence and before/after runtime; absence of observed mutants alone is not proof. |

| Integration point | Intended home | Wraps / proof seam |
|---|---|---|
| Existing local-check recorder and `test-stats` | ariadne `cmd/sdlc/` | Reuse #376's bounded per-attempt records and 30-day retention; malformed/partial input stays unknown (ARCH-DRY, ARCH-SECURE, ARCH-FUNERAL). |
| Merge-queue adapter | ariadne `cmd/sdlc/` | Git, GitHub PR and status operations; reuse existing seams and stateful fakes, plus opt-in live read-only conformance (ARCH-MOCK). |
| Repository test adapter | ariadne/pair Makefiles and atlas traceability | Resolve paths to tests through #328; unknown mapping widens the run. |
| Ops daily triage | ops test-economy project | Human-owned issue/CI observations and metric updates; no new daemon. |

## 1. Evidence tiers

1. **Implementation:** run focused unit/integration tests for changed behavior immediately. #328's traceability map supplies reproducible `test-changed` selection; unclassified changed code falls back to the affected package and its dependents or the full suite. Preserve command, head, result and selection rationale. Targeted tests give feedback, not a coverage guarantee.
2. **Close/review:** require behavioral tests named by Done when and a targeted run for touched surfaces. Keep current SDLC review and its special `--fixed-then-ship` full-suite requirement for code conflict resolution until an explicit gate change is approved. Closing may precede the final full suite; later code changes invalidate review evidence under current gate rules.
3. **Landing:** require one passing full suite recorded for the **exact integrated head** eligible to merge, plus PR CI, review and local-gates checks. Main or PR-head movement makes a candidate stale. Do not relabel a prior run on a different code tree. Batch sharing is allowed only if all PRs land atomically from the tested tree with each review and CI verified; defer that protocol until serial admission is measured.

Targeted runs shorten iteration while exact-head suite evidence still protects main (ARCH-PURPOSE). #328 must make the map language-neutral for ariadne/pair and preserve parley's current behavior.

## 2. Integration tax: serial admission first

Use **one active integration candidate per repository**. Ready PRs keep independent reviews and CI, then queue FIFO. Admission snapshots main and PR head, creates the integration head, runs the declared suite under the existing weighted suite lock, verifies exact-head status and CI, re-fetches main, and lands only when pinned identities still match. Main movement invalidates the candidate; rebuild against new main while later candidates wait. A failed or timed-out suite blocks only its candidate and produces a diagnosis; it never silently admits it. Preserve its original branch and review evidence on refusal. A failed candidate may be removed from the queue without blocking later ready PRs, but no failed candidate lands.

Use a durable state machine (ARCH-ORDER): `waiting → integrating → testing → verified → landing → landed`, with explicit `stale`, `failed` and `uncertain` outcomes. A crash, timeout or unknown GitHub response resumes by fetching PR/main and checking exact OIDs; probe uncertain publication before retry. A terminal receipt is kept for the 30-day analysis window, then swept (ARCH-FUNERAL). Do not hold the repo transaction lock across suite wait/run (#370); lock only short Git/tracker transitions. The machine-wide weighted suite lock still limits load (ARCH-CONSTRAINTS). Test main drift, head drift, process death at each effect, duplicate requests and uncertain GitHub replies through the production transition function and stateful fakes.

**Pilot stop rule:** if serial admission does not reduce repeat full suites per merged PR, or raises p90 ready-to-land delay over seven complete days, pause expansion and inspect queue scheduling versus suite runtime. Batching is separate because it changes merge/recovery identity. A later doc-only reuse issue must compare effective test inputs, including generated files, test commands, configuration and dependency lockfiles, on both trees; unknowns run the suite. A file-extension-only exemption is insufficient.

## 3. Flake policy and owner

The TL owns a **daily standing list** in the ops project. Every intermittent failure observed in local attempts or CI gets a named issue within 24 hours, with repo, test, exact head, log/run link, first-seen time, reproduction status and owner. Within that day the owner must identify the root cause or record the failed hypotheses, next diagnostic and an explicit blocker; an unexplained retry is not resolution. A retry may diagnose but cannot convert an unexplained failure into passing evidence. Fix the scheduling, fixture or root cause; never lengthen a timeout solely to green the suite. Classify reproducible branch regressions apart from confirmed same-head fail→pass and suspected environmental faults. Retain failed-attempt logs (#376), and extend measurement to CI before reporting a combined health trend.

Quarantine needs operator approval for a named test, an owning issue, replacement coverage (or written risk acceptance), a seven-day expiry and an explicit main-health signal. Expiry restores the test or blocks landing until renewed. Do not quarantine a whole package. The target is **zero unowned flakes**, not a claim that a short sample proves zero flake probability. The daily review checks each new failure and the filing clock.

## 4. Package budgets, with couchcore first

Keep versioned soft p90 budgets in each repo's `.sdlc/test-budgets.json` (or equivalent `.sdlc/` declaration chosen in the implementing plan). The budget belongs to one package and test profile, not a global timeout. Compare at least ten complete comparable attempts in the last seven days; show count and wall/queue separately; mark incomplete values unknown. An overrun does not fail a correct suite. It requires an issue or documented reason and a measured split/parallelization/minimization plan. Only an operator-approved baseline revision raises a budget; one slow run cannot ratchet it upward. Keep revisions in Git.

Pair's `cmd/internal/couchcore` p90 is **481.3 s (n=21)**. Propose initial soft ceiling **600 s** (about 25% headroom, rounded; operator choice) and separate improvement target **≤240 s p90** after profiling. The ceiling catches large regressions without accepting 481 s as healthy. Profile cost and test semantics before splitting; record cache state and host. Do not derive ariadne package budgets from shard p90; first expose package durations. Track suite and queue p50/p90 separately, and improve the bottleneck before raising suite capacity (ARCH-CONSTRAINTS).

## 5. Periodic minimization

#327's monthly repo-GC issue owns a report, starting with couchcore. Rank package time by p90, then nominate tests using overlapping line/branch coverage or duplicate scenarios. **Coverage overlap nominates; it does not prove equal fault detection.** For each candidate compare unique mutants killed on a bounded code slice where practical, requirements/traceability coverage, and before/after package wall time. Keep a test if it uniquely kills a mutant, guards an unmodeled contract, or covers a distinct integration path. Remove or merge only with a reviewable evidence table and a full-suite landing run. Report “not measurable” rather than fabricated redundancy.

The monthly report names owner, sampled package and run cost. #327's schedule/report implementation must bound retention and archive old reports (ARCH-FUNERAL). The cadence complements immediate flake fixes.

## Implementing issues and rollout

These are existing or proposed **separate issues to file/revise after operator review**; #390 authorizes no implementation. Every new issue names acceptance tests, exact paths and gate impact in its own plan.

| Order | Issue | Deliverable and acceptance evidence |
|---:|---|---|
| 1 | **ariadne#328** — traceability and targeted selection | Changed paths map to tests or widen to a documented run; ariadne/pair adapters and parley compatibility demonstrated. |
| 2 | **ariadne#380** — saved local-check status re-upload | Exact-head, exit and log hash survive failed POST; retries do not rerun a passed suite. This is a dependency of reliable queue admission. |
| 3 | **New: “sdlc merge queue: serial exact-head landing admission”** | Crash-safe single candidate, fair queue, exact OIDs, stateful Git/GitHub fake, drift/uncertain-outcome tests; each candidate still gets a suite. |
| 4 | **New: “test economy: CI failures and daily flake ledger”** | CI outcomes separate from local same-head flakes; TL's daily pass files/links each observed intermittent failure within 24h; historical 17/200 labeled mixed. |
| 5 | **New: “soft package test budgets in .sdlc”** | Versioned declaration, coverage-aware p90 comparison, explanation workflow and couchcore 600 s pilot; no timeout or green retry. |
| 6 | **New: “profile and reduce pair couchcore test runtime”** | Reproducible profile, cause-level changes, unchanged behavioral coverage, p90 trend toward 240 s over ≥10 comparable runs. |
| 7 | **ariadne#327** — periodic GC/minimization | Monthly issue and first couchcore report with unique-detection evidence and before/after wall time. |
| 8 | **New: “test-economy guidance for ariadne and pair”** | Evidence tiers, flake response, budget behavior and minimization cadence are visible in both repos' agent guidance and atlas; rules agree with the gates. |
| Later | **New: “reuse suite after proven doc-only main drift”** | Complete effective-input equivalence proof, fail-closed classifier, adversarial gate tests; pilot only if drift data warrants it. |

The queue and budgets are local-gates changes if they alter `.sdlc/` declarations or merge policy; each must follow the local-gates-guard and operator-label workflow. Do not combine gate changes with unrelated optimization. The guidance issue is part of the project, not an optional appendix.

## Measurement and review

After a seven-day pilot, report by repo: count/coverage; suite and queue p50/p90; full suites and integration rebuilds per merged PR; ready-to-land p50/p90; slow-package p90/count; confirmed same-head flakes; CI failure classes; and filing lag for intermittent failures. Keep local queue and GitHub CI waits separate. Compare ariadne queue p50 with the 602 s measured subset and seek ≤301 s, but do not declare success if coverage or workload differs materially. Pair's 0.1 s median is a poor halving target; seek p90 below 1842 s and no suite-wall growth. The stronger invariant is no unowned flake older than one day and no unexplained package-budget regression. Attribute improvement to the specific car; a queue must show fewer repeat integrations, not merely a quiet fleet week.

## Operator choices

1. Approve serial exact-head admission before batch landing? **Recommended yes.** Batching changes the unit of merge and failure recovery and deserves a separate design.
2. Approve soft budgets and couchcore's **600 s ceiling / 240 s target** as pilot values? Both are choices, not dictated by the data alone.
3. Confirm TL daily flake ownership and seven-day operator-approved quarantine exception.

After approval, record changed decisions in dated `## Revisions` sections here and in #390, commit the artifacts, file the new issues, update the ops project, and close #390 as design-only. Implementation starts on separate issue branches from main.

## Revisions

### 2026-10-15 13:39 PDT — Fleet-blocking priority lane

**Reason:** ariadne#391 blocks `sdlc close` across the fleet and waited fourth for a suite today. A strict FIFO serial admission queue would preserve that failure mode.

**Delta:** The serial-admission recommendation's “queue FIFO” means FIFO **within a priority class**, not across all candidates. Add a `fleet-blocking` class above `normal`, granted only by the operator or TL for a named issue that blocks a shared fleet workflow. The grant records grantor, issue, reason, time and expiry (at most 24 hours unless renewed); the queue's admission record also logs when the candidate jumps ahead of normal candidates. At each **next admission**, choose the oldest ready fleet-blocking candidate, then the oldest ready normal candidate. Never interrupt a running suite, revoke a candidate already landing, or treat the priority grant as a waiver of review, CI, exact-head suite, or local-gates checks. A failed/stale prioritized candidate leaves admission and re-enters only when ready; priority expires on issue landing or its timestamp. Report priority and normal wait distributions plus displacement count in the pilot, so chronic emergency use is visible (ARCH-ORDER, ARCH-CONSTRAINTS).

The new merge-queue implementation issue must test: a granted fleet blocker overtakes waiting normal candidates; same-class FIFO remains stable; a running suite is not preempted; expired/unlogged/unauthorized grants have no effect; and retry after crash preserves the logged grant without duplicating admission. This revision supersedes the unqualified FIFO wording in section 2 and the rollout table.

### 2026-10-15 13:49 PDT — Track before budgets; feature-scoped minimization

**Reason:** Operator review found that the proposed 600 s ceiling and 240 s target for couchcore outrun the available trend. The first test-economy change needs a measurable baseline for how many full suites a landing consumes, and minimization needs a feature-level protection oracle.

**Delta — measurement first:** Supersede section 4's proposed couchcore values and the rollout row that would implement them now. `481.3 s (n=21)` remains an observed package p90, **not a budget or target**. Add a first implementing issue, **“Weekly test-economy snapshot from local-check and landing records,”** before changes to evidence tiers or budgets. Once a week it writes a dated, immutable summary to brain under its measurement charter, with a stable link in the ops test-economy project. For each repo the snapshot includes the exact seven-day window and coverage, suite and queue p50/p90, slowest packages with sample counts, named/confirmed/suspected flakes separately, and **full suites per landed issue** with numerator, denominator and unknown attribution. Read #376's per-attempt records and authoritative landed-issue evidence; never count an overwritten retry as a new issue or divide by an inferred landing. Keep missing fields explicitly unknown. The ops project points to each brain snapshot rather than duplicating its data. Do not set `.sdlc/` package ceilings until several weekly snapshots show a trend and the operator chooses the budget values; keep the future soft-budget implementation issue in the backlog. Profiling couchcore can proceed as investigation, but it must not silently install a 600/240 s policy.

**Delta — headline comparison:** Make **full suites per landed issue** the headline before/after metric for the evidence-tier change. Count all full-suite attempts attributable to an issue from first ready-for-landing through confirmed landing, including main-drift reruns and failed attempts; separately show successful runs, queue waits and unassignable attempts. A shared future batch suite must state each issue's attribution rule and report the physical suite count as well, so a denominator change cannot manufacture savings. Preserve the exact-head full-suite landing gate while testing the tier policy. Report the metric by repo and week, with counts and coverage, before interpreting a change.

**Delta — minimization:** Supersede section 5's package-first removal sequence. Start only after #328's feature → code → tests traceability map is operating. Work **one mapped feature at a time**: coverage overlap nominates possible duplicate tests; feature-scoped mutation testing supplies deterministic fault-detection evidence; a reviewer judges removal against behavioral contracts and integration paths. Record which mutants each candidate uniquely kills, any unmodeled requirement, the decision, and before/after runtime. A missing mutation tool or incomplete map means “not measured,” never automatic deletion. #327's first GC run follows this order; keep performance profiling separate from deletion decisions.

**Rollout amendment:** The weekly brain snapshot issue precedes #328; #328 precedes #327's minimization. The budget declaration issue waits for several weekly snapshots and a further operator decision. The plan's earlier 600/240 s pilot and immediate budget row are withdrawn. Add the snapshot linkage and headline ratio to the seven-day review, and compare them only when both weeks have enough complete issue attribution.
