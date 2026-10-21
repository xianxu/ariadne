# Boundary Review — ariadne#463 (whole-issue close)

| field | value |
|-------|-------|
| issue | 463 — local-check: resume recorded known-flake failures with targeted tests |
| repo | ariadne |
| issue file | workshop/issues/000463-resume-recorded-known-flakes.md |
| boundary | whole-issue close |
| milestone | — |
| window | 82e5783a0defcf1745849cb7d0086fbffa5d3396..31d5e5b22a8629a9aa774c2e3ff924b9f334b24b |
| command | sdlc close --issue 463 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T13:29:15-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised spec, including the operator-selected one-pass cleanup policy. Completed-child publication follows confirmed cleanup; uncertain resumes retain surviving inputs, and reconciliation records interruption without destructive effects. No new blocking findings.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Superseded by operator option A, decision 390faec8-dc49-498e-9ad5-2c1a718c2c37. Cleanup now precedes child publication; uncertain inputs remain pinned and reconciliation performs zero cleanup. Private crash-window tests pass. Removing retirement in a disposable checkout makes TestResumeCrashCommittedChildStatusPost fail with a retained locked checkout.
  - id: BR-1
    disposition: addressed
    note: |
      Receipt publication validates supported link authority explicitly. TestResumeStorePublishesTerminalReceipt passes, including unchanged unknown JSON bytes and duplicate-publication refusal.
  - id: BR-2
    disposition: addressed
    note: |
      Directory-mutex ownership covers staging lifetime and crash reclamation. The staging-lifetime and killed-publisher regressions pass.
  - id: BR-4
    disposition: addressed
    note: |
      Scratch and environment cleanup share resumeCleanupEffect. Cancellation-boundary tests pass; bypassing that helper in a disposable checkout makes both regressions fail on subsequent effects after cancellation.
```

1. **Strengths**

   - Cleanup precedes child publication at [localcheck_resume.go:193](/Volumes/TestScratch/review-checkouts/888890407b1ee200/cmd/sdlc/localcheck_resume.go:193), closing the committed-child crash window.
   - [resolvedResumeRecords](/Volumes/TestScratch/review-checkouts/888890407b1ee200/cmd/sdlc/localcheck_resume_saved.go:136) validates every relevant parent, preventing a later green attempt from bypassing unresolved proof.
   - README, embedded help and atlas describe the delivered commands, evidence-only reconciliation and indefinite retention cost.
   - Private subprocess barriers and independent retained-byte assertions test actual execution and crash behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - Resume, known-flake, saved-proof, guard and lint-rule checks: **232.397s**.
   - Private CLI conformance, crash windows and first-pass cleanup failure: **161.844s**.
   - Resume tests under the race detector: **31.523s**.
   - Pinned-range `git diff --check`.

   Both disposable negative controls failed as expected. The disposable checkout was removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Shared classifier, proof resolver and cleanup-effect boundary.
   - **ARCH-PURE — pass:** Decoder, proof and transition tests require no IO; integration effects have explicit seams.
   - **ARCH-PURPOSE — pass:** Targeted-only execution and shared success-reader enforcement are delivered.
   - **ARCH-MOCK — pass:** Private stateful GH/tracker fixtures exercise production paths alongside real private Git/process behavior.
   - **ARCH-CONSTRAINTS — pass:** Bounded evidence reads, linear-work tests, one suite lease and aggregate timeout remain enforced.
   - **ARCH-SECURE — pass:** Strict decoding, verified digests, exact repository/head bindings and owner-incarnation checks guard authority.
   - **ARCH-ORDER — pass:** Explicit transitions, lifetime parent locking and deterministic crash barriers cover uncertainty and repeated requests.
   - **ARCH-FUNERAL — pass:** Successful cleanup precedes publication; indefinite uncertain-input retention is an explicit documented operator tradeoff.
   - **ARCH-EVOLVE — pass:** Immutable publication preserves unknown bytes; required fields, types and duplicate-key validation remain strict.

7. **Plan revision recommendations:** None. The M3 revision explicitly supersedes M2 and matches the delivered policy.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"57147204-fc56-4424-a84e-f8df6a37479e","repository":"github.com/xianxu/ariadne","issue_id":"000463","recipe":"milestone-review","round":8,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/888890407b1ee200","reviewed_head":"31d5e5b22a8629a9aa774c2e3ff924b9f334b24b","slot":"ariadne:3","requested_at":"2026-10-21T20:23:33.087861Z"},"phase":"finished","pid":81627,"started_at":"2026-10-21T20:23:33.253249Z","finished_at":"2026-10-21T20:29:14.588165Z","session_id":"01a16022-2841-7253-bdf9-5fdc6c035692","outcome":"returned"}] -->
