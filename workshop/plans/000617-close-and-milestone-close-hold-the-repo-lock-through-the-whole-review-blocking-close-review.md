# Boundary Review — ariadne#617 (whole-issue close)

| field | value |
|-------|-------|
| issue | 617 — close and milestone-close hold the repo lock through the whole review, blocking other slots' bookkeeping |
| repo | ariadne |
| issue file | workshop/issues/000617-close-and-milestone-close-hold-the-repo-lock-through-the-whole-review-blocking.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7e510e37241401e9877273dbea49a2d462968f5c..0c8a0644ed9f7efdebf1bf960fede37263e27eb3 |
| command | sdlc close --issue 617 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T21:18:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation moves expensive preparation and optional publication outside the repository lock while retaining fresh validation before authoritative writes. Targeted tests passed. No correctness defect was found, but the new adoption-invalidation safety path needs a command-level regression test before shipping.

Done-when: current — the existing criteria cover independent bookkeeping during slow work and refusal of stale finalization.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: authority-invalidation-command-coverage
    title: |
      Exercise adoption invalidation through both close commands
    detail: |
      cmd/sdlc/actualexclusion_readset_test.go:139 tests capture and comparison through a test helper, while cmd/sdlc/closeprepare_cli_test.go:84 disables measurement with --no-actual. No regression exercises the transfer from actualResult.Adoptions through closeFlags into the preparation guard, so removing guard.adopted = f.Adoptions at closeprepare.go:123 would leave these safety tests green. Add private-tracker schedules for both commands that introduce or change a competing adoption during preparation and review; assert refusal before sidecar, ledger, issue or card writes, and confirm the tests fail when propagation or final validation is removed. This supplies the command-level verification promised by the plan (ARCH-PURPOSE, ARCH-ORDER).
```

1. **Strengths**

   - [closeprepare.go:114](/Volumes/TestScratch/review-checkouts/8b1117313472c8df/cmd/sdlc/closeprepare.go:114) runs measurement and churn unlocked, then validates captured authority under the lock.
   - [actualexclusion_readset.go:25](/Volumes/TestScratch/review-checkouts/8b1117313472c8df/cmd/sdlc/actualexclusion_readset.go:25) shares the candidate filter between measurement and validation, avoiding parallel attribution rules.
   - [boundarypush.go:102](/Volumes/TestScratch/review-checkouts/8b1117313472c8df/cmd/sdlc/boundarypush.go:102) publishes the captured commit with its captured lease and preserves newer tracking metadata through CAS.

2. **Critical findings:** None.

3. **Important findings:** The command-level adoption regression described above. Existing helper tests validate the comparison policy but do not protect its production wiring.

4. **Minor findings:** None.

5. **Test coverage notes**

   The targeted run covering preparation, publication, adoption read-sets, reviewer lock release and concurrency schedules passed: `ok .../cmd/sdlc 228.812s`. `git diff --check` passed. The full suite was not run during this review. Native barrier tests successfully cover sibling bookkeeping during preparation and review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing snapshot, filter and lock seams are reused.
   - **ARCH-PURE — pass:** comparison remains separate from fresh authority collection; the plan revision correctly classifies the guard as integration.
   - **ARCH-PURPOSE — flag:** promised command-level adoption-race verification is incomplete.
   - **ARCH-MOCK — pass:** private Git/tracker fixtures and executable reviewer barriers exercise shared production boundaries.
   - **ARCH-CONSTRAINTS — pass:** long scans leave the lock; validation growth has bounded-read coverage.
   - **ARCH-SECURE — pass:** unavailable or changed authority refuses finalization.
   - **ARCH-ORDER — flag:** adoption changes lack sequence coverage through the command coordinator.
   - **ARCH-FUNERAL — pass:** read-sets are invocation-local; deferred effects finish synchronously.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

   Atlas documents the revised flow. No new command, flag or configuration surface requires a README update.

7. **Plan revision recommendations:** None required; the plan already calls for the missing regression. Complete that verification before marking it delivered.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fe3a607b-4044-4227-b9a7-284ae57e8f56","repository":"github.com/xianxu/ariadne","issue_id":"000617","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8b1117313472c8df","reviewed_head":"0c8a0644ed9f7efdebf1bf960fede37263e27eb3","slot":"ariadne:3","requested_at":"2026-10-25T04:13:41.557839Z"},"phase":"finished","pid":77455,"started_at":"2026-10-25T04:13:41.660021Z","finished_at":"2026-10-25T04:18:42.929612Z","session_id":"01a17143-a8ed-7090-af19-5621e0b02257","outcome":"returned"}] -->
