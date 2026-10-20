---
collaboration-mode: discussion
---

# Suite concurrency implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the suite's measured operating envelope and replace the CPU-count shard default with the supported cap, addressing the overload behind #480 and folding #484 into this issue.

**Architecture:** Measure one pinned checkout under the existing machine-wide two-slot gate, with external heavy work paused by the hub. Compare three runs each at 12, 8 and 6 shards; select the default from wall/load/kernel evidence. Preserve every existing landing lock, timeout and assertion; the identified cleanup window is a candidate follow-up.

**Tech stack:** Python standard library, Darwin Mach CPU counters, Go suite, existing Make/test-lock helpers.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Supported shard cap / CPU-count fallback | scripts/test-shard.py | modified |
| CPU tick deltas and run summaries | local test-stats/evidence/000480/benchmark/measure.py | new, experimental evidence only |

The shard cap bounds the default on larger hosts while respecting smaller CPU counts. Explicit --shards remains available for comparisons. One helper owns the default; tests in scripts/test-shard.test.py exercise it without starting a suite. The CPU-summary functions consume recorded samples, allowing synthetic counter/wrap/empty-data tests without IO.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Sharded suite runner | scripts/test-shard.py, Makefile.local | unchanged interface | real Go/Git test processes |
| Suite reservation | Makefile / Makefile.workflow / scripts/test-lock.py | existing | kernel flocks, both machine slots |
| Measurement sampler | local test-stats/evidence/000480/benchmark/measure.py | new, experimental evidence only | os.getloadavg, Darwin host_statistics, subprocess |

Use real suites for the comparison; a fake cannot establish CPU capacity. Validate CPU-delta arithmetic with synthetic samples and compare the live sampler with the operating system's CPU report before relying on it. Do not add a production telemetry service or configuration surface.

## Operating envelope and lifecycle

- Nine full suites, sequential, at one immutable code head. Interleave shard counts as 12/8/6, 8/6/12, 6/12/8 to balance order/cache/thermal effects. Each count appears once in each position.
- Ask the hub to coordinate the quiet interval and the two-slot reservation. Existing full suites finish first. No cancellation or alteration of another slot's runs. Keep the machine gate; do not use WF_TEST_LOCK=off.
- Proposed quiet precondition: at least five consecutive one-second samples with CPU idle >=80% and one-minute load <= half the logical CPUs. Record the actual baseline. This threshold excludes inherited load from the preceding suite; it is a measurement condition, not a product promise. Stop and report if quiet cannot be reached within ten minutes.
- Sample once per second with the standard library: one load read and one Mach CPU-stat read. O(T) work and output for T seconds; no subprocess per sample. Record raw ticks, sample timestamps, peak one-minute load, mean kernel percent of total CPU, and kernel share of busy CPU (both definitions remove ambiguity).
- Wall time starts after admission, immediately before launching each complete suite, and ends on its process exit. Preserve exit codes and complete test output; report failures by name. Never treat a failed or interrupted run as a passing performance result.
- Keep raw logs, samples, per-run manifests and checksums on persistent storage under the Git-common-dir's sdlc/test-stats/evidence/000480, never exclusively on TestScratch. A fixed nine-run experiment bounds the artifact family; do not copy compiled binaries or fixture trees. Test scratch/worktree uses TestScratch when available and is removed after the experiment.
- A unique directory identifies each attempt; an existing completed run is never overwritten. An interrupted run stays partial and is repeated in a new directory if needed. The owner joins the sampler and terminates its own subprocess group on cancellation; the existing suite lease releases with its owner.

| Internal state / failure | Visible evidence | Recovery action | Possible loss |
|---|---|---|---|
| Queued / waiting for quiet | No measured run; queue/baseline log | Wait for the reserved quiet interval | Time only |
| Running | Start manifest, samples and output | Wait, or cancel this experiment | In-flight run becomes non-comparable |
| Completed | Exit, duration, CPU/load summary, hashes | Compare complete runs | None |
| Interrupted / nonquiet / failed | Partial or failed manifest, retained logs | Classify, then repeat in a new directory | No performance conclusion for that run |

ARCH-DRY/PURE: one default helper, pure tick arithmetic, existing runner/gate. ARCH-CONSTRAINTS: explicit concurrency, co-tenancy, sample cost and overload evidence. ARCH-ORDER/FUNERAL: immutable head, serial suites, process-owned leases, joined sampling and bounded experiment artifacts. ARCH-MOCK: synthetic arithmetic plus real OS conformance and real complete suites. ARCH-SECURE/EVOLVE: no production record or trust change; local manifests are evidence, not gate authority. ARCH-PURPOSE: change the measured suite envelope, not the landing protocol or its timeout.

## Task 1: preserve evidence and revise acceptance

- [x] Copy diagnostic logs/dumps off RAM disk and verify each hash. Persistent directory: Git-common-dir/sdlc/test-stats/evidence/000480/lock-diagnostic; manifest SHA256 209e0bf229bb07ef62275bacb70c90158509c850442c191c810ede44f419837f.
- [ ] Update #480 Spec/Done-when/Revisions with the hub's fold approval and this plan. Preserve the original timeout criterion and diagnostic uncertainty.
- [x] Triage #484's still-open card to wontfix as folded (no duplicate implementation). The hub independently published its fold note on main; preserve that richer body rather than overwrite it. Leave the project to the hub.
- [ ] Commit the design and send the concrete plan / quiet-window request to the hub before implementation.

## Task 2: measure the matrix

- [ ] Build the local evidence harness and its SKILL.md usage record. Test CPU delta calculations (normal, zero ticks, uint32 wrap), sample aggregation, and interrupted-run marking with deterministic inputs before measuring.
- [ ] Validate the live sampler against top's system/idle report; keep the raw comparison and units in the evidence manifest.
- [ ] Create a detached checkout pinned to the approved measurement head. Verify clean source and Git revision before and after the matrix; do not commit into that checkout.
- [ ] Execute the approved quiet matrix through the machine gate. Capture three attempts each at 12/8/6 shards, wall, peak one-minute load, kernel percentages and every failure. Preserve per-run output/JSON and environment (hardware, Python/Go versions, exact commands, scratch choice, sample interval).
- [ ] Summarize individual runs and medians in the issue. Apply #484's published criterion: wall time within about 10% of the best while keeping peak load below about twice the core count (about 24 on this 12-core host). Choose the cap and weight from this table, not from CPU count alone. If none meets both, report the conflict to the hub before expanding the matrix or silently relaxing the envelope. The solo matrix cannot justify letting two whole-module suites overlap, so retain weight 2 unless further evidence is explicitly approved.

## Task 3: enforce the selected default

**Files:** scripts/test-shard.py; scripts/test-shard.test.py; Makefile; atlas/workflow/sdlc-binary.md; README.md only if existing user-facing suite guidance changes.

- [ ] Write a failing unit test for the measured cap, small CPU counts and unavailable CPU count; include explicit --shards override coverage. Run python3 scripts/test-shard.test.py and confirm the old CPU-count default fails the new contract.
- [ ] Add the minimal pure default_shards helper, using min(measured_cap, detected_cpus) with a positive fallback bounded by the cap. Use it for argparse's default; preserve explicit overrides and the runner's full-test coverage check.
- [ ] Set or reaffirm WF_TEST_WEIGHT in Makefile with measured rationale and a pointer to the result. Add a focused policy assertion to the existing shard tests so the operating envelope cannot silently revert. This is the mechanical regression guard; do not add a lexical Go lint rule for a Python/Make policy.
- [ ] Run the shard unit tests and a real --list/selection smoke as appropriate. Update atlas and the issue's table/decision. Commit the result; no landingtrain.go or landing.go implementation change.

## Task 4: verify, review and land

- [ ] Push/open the PR; run the required recorded suite on the final pushed head. The original train test must pass all variants; preserve any known-flake attribution and classify unknown failures.
- [ ] Close #480 through the fresh SDLC review with matrix and regression evidence. Address Important/Critical findings; the binary owns review.
- [ ] Enter sdlc merge --yes promptly, allow its integration and CI/suite gates, then report the landed PR plus the chosen cap/weight and evidence to the hub. Remove the owned scratch checkout/diagnostic binary; retain persistent evidence.
