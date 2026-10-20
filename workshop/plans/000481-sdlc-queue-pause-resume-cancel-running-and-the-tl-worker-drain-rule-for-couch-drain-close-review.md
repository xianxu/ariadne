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
