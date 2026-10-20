# Boundary Review — ariadne#481 (whole-issue close)

| field | value |
|-------|-------|
| issue | 481 — sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain |
| repo | ariadne |
| issue file | workshop/issues/000481-sdlc-queue-pause-resume-cancel-running-and-the-tl-worker-drain-rule-for-couch-drain.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8a44224ff0f867837283368f62c0344bcd322b25..887279de149e943bdf664d9a1ad35d34ec2d91e2 |
| command | sdlc close --issue 481 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T03:27:35-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range delivers owner-scoped pause/resume, shared suite and landing admission, conservative status, and interruption evidence separated from test failures. Targeted tests and race checks passed. No new blocking findings remain; this review did not rerun the full recorded suite or the peer repository’s conformance adapter.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Exact-ID intents are consumed by owners and observation; recovery retains interruption evidence alongside descendant uncertainty. Recovery tests pass. Disabling owner intent consumption made TestQueueCrashIntentBeforePhaseIsAuthoritative/owner fail.
  - id: BR-2
    disposition: addressed
    note: |
      The plan's Core concepts and BR-2 revision now describe Control, string phases, InterruptionRecord, and observeLegacyQueue, matching their implementation and actual transition signatures.
  - id: BR-3
    disposition: addressed
    note: |
      Transaction loops check cancellation and preserve applied owner state. Deadline regressions pass; disabling Observe's internal checks made TestQueueSnapshotDeadlineInsideTransaction fail with 16 reads instead of one.
  - id: BR-4
    disposition: addressed
    note: |
      queueLegacyFields delegates structural validation to queuegate.StrictObject. Shared adversarial codec tests and legacy-reader tests pass.
  - id: BR-5
    disposition: addressed
    note: |
      Go and Python admission attempt secondary locks without blocking under the global gate. Contention regressions pass; restoring blocking Go acquisition made TestQueueContendedLineNeverPinsGlobalGate fail because pause expired.
```

1. **Strengths**

   - Admission and active-lease publication share the pause transaction in both scheduling paths.
   - Cancellation/completion ordering is tested, including retained logs, no failure-store poisoning, and same-head retry.
   - Legacy and unmanaged activity remains visibly uncertain instead of becoming false quiescence.
   - README, atlas, command help, and worker instructions cover the new behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed queuegate and landingline tests normally and under race; `TestQueue` normally and under race; all 15 Python pause controls; all 16 priority controls through the supplied-provider harness; saved-pass reuse and descendant cleanup; worker/helptext/lint tests; and diff whitespace checks. Temporary mutation copies were removed. The initial standalone priority invocation lacked `bin/sdlc`; the supported provider harness resolved that prerequisite and passed.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared JSON validation and existing schedulers are reused.
   - **ARCH-PURE — pass:** transition, snapshot, and codec logic is directly testable without IO.
   - **ARCH-PURPOSE — pass:** suite admission, ordinary merges, landing, and cleanup paths are covered.
   - **ARCH-MOCK — pass:** portable stores, real flocks, and injected process/topology dependencies exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** bounded registries/output, cancellation checks, and counted scan tests enforce the declared envelope.
   - **ARCH-SECURE — pass:** strict boundary parsing and exact process identity checks preserve uncertainty safely.
   - **ARCH-ORDER — pass:** publication, cancellation, completion, and recovery have reproducible ordering tests.
   - **ARCH-FUNERAL — pass:** leases/intents have retirement paths; diagnostic retention and unresolved-state bounds are explicit.
   - **ARCH-EVOLVE — pass:** opaque additive fields survive rewrites, including nested interruption operations.

7. **Plan revision recommendations:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"955e47e7-ee2c-46e8-84eb-28860da281c6","repository":"github.com/xianxu/ariadne","issue_id":"000481","recipe":"milestone-review","round":5,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/88d5dae1d776a430","reviewed_head":"887279de149e943bdf664d9a1ad35d34ec2d91e2","slot":"ariadne:3","requested_at":"2026-10-20T10:24:00.959685Z"},"phase":"finished","pid":51431,"started_at":"2026-10-20T10:24:01.056243Z","finished_at":"2026-10-20T10:27:35.122245Z","session_id":"01a158d6-e721-7952-a184-7a1a9b4f928a","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T04:09:38-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 481 — sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain |
| repo | ariadne |
| issue file | workshop/issues/000481-sdlc-queue-pause-resume-cancel-running-and-the-tl-worker-drain-rule-for-couch-drain.md |
| boundary | whole-issue close |
| milestone | — |
| window | 613a99f52fb8fdfeb4d52cb34d2aa2ef9bd0c700..0d27dcb7fb64d3e1e6f3e367455fcc423296a256 |
| command | sdlc close --issue 481 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T04:09:38-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned interdiff correctly threads caller contexts through queue retirement while preserving cancellation-independent cleanup and its existing timeout. All production callers are updated. No scope change or new blocking defect was found; previously disposed findings remain undisturbed.

```findings
{}
```

1. **Strengths**
   - Retirement preserves caller values through `context.WithoutCancel` and retains the three-second cleanup budget (`cmd/sdlc/queue_runtime.go:109`).
   - The new real-store regression verifies that cancellation still permits retirement and leaves no active operation (`cmd/sdlc/queue_runtime_test.go:70`).
   - Suite, ordinary merge, and durable landing callers consistently pass their contexts.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Focused runtime, helper, and context-guard tests passed. Race-enabled runtime, managed interruption, same-head retry, and merge admission tests passed in 22.247s. Diff whitespace checks passed. The full recorded suite was not rerun during this review.

6. **Architecture:** ARCH-DRY passes: retirement remains centralized. ARCH-PURE passes: the change stays within the existing IO boundary. ARCH-PURPOSE passes: every production retirement caller is updated. ARCH-EVOLVE passes: persisted formats are unchanged. No new surface requires atlas or README updates.

7. **Plan revision recommendations:** None; Spec, Done-when, and Revisions remain unchanged in this window.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d7f8b2ac-9de5-42bd-a643-f87a7aef91c5","repository":"github.com/xianxu/ariadne","issue_id":"000481","recipe":"fix-review","round":6,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/aba61e58e6c29533","reviewed_head":"0d27dcb7fb64d3e1e6f3e367455fcc423296a256","slot":"ariadne:3","requested_at":"2026-10-20T11:08:23.452694Z"},"phase":"finished","pid":98968,"started_at":"2026-10-20T11:08:23.652374Z","finished_at":"2026-10-20T11:09:37.925405Z","session_id":"01a158ff-8897-79d0-b270-e05dfafc41d8","outcome":"returned"}] -->
