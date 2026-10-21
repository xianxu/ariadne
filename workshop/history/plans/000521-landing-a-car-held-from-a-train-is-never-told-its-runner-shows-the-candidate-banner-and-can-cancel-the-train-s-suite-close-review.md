# Boundary Review — ariadne#521 (whole-issue close)

| field | value |
|-------|-------|
| issue | 521 — landing: a car held from a train is never told; its runner shows the candidate banner and can cancel the train's suite |
| repo | ariadne |
| issue file | workshop/issues/000521-landing-a-car-held-from-a-train-is-never-told-its-runner-shows-the-candidate-banner-and-can-cancel-the-train-s-suite.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1f28e5a6c8a5ab3ffb5af140ec1b8e969f3a662d..441ac00b8654e9404eaa9914ac2eed1a73edca42 |
| command | sdlc close --issue 521 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T04:00:44-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy #521’s Spec and approved correction. Held runners receive recovery instructions, candidate and composed membership are distinct, and an excluded conductor stops before publishing or running survivors’ suites. No blocking findings.

```findings
{}
```

1. **Strengths**
   - The pre-publication guard preserves delivery of every composition hold before aborting (`cmd/sdlc/landingtrain.go:176`).
   - Both conductor and waiter reuse the held-recovery formatter (`cmd/sdlc/landingqueue.go:180`).
   - The integration regression verifies survivor reformation, live lease ownership, foreign-cancel refusal, and final PR states (`cmd/sdlc/landingtrain_held_test.go:57`).
   - README and atlas document the changed behavior and recovery policy.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Focused integration tests passed in 45.213s; pure state-machine controls passed in 0.140s. Temporary Go overlays independently removed the publication guard and waiter reporting; each corresponding regression failed for the expected reason. Diff whitespace checks passed. The full suite was not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Shared formatters prevent divergent membership and recovery output.
   - **ARCH-PURE — pass:** Existing pure transitions remain intact; effect execution stays in the IO shell.
   - **ARCH-PURPOSE — pass:** Both held-runner paths and cancellation ownership are covered.
   - **ARCH-EVOLVE — pass:** No persisted formats or field meanings change.

7. **Plan revision recommendations:** None. The implementation matches the approved abort/release policy.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"35d8c256-7ad9-42e1-a9ff-5a854a494206","repository":"github.com/xianxu/ariadne","issue_id":"000521","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c22d00fd09961638","reviewed_head":"441ac00b8654e9404eaa9914ac2eed1a73edca42","slot":"ariadne:1","requested_at":"2026-10-21T10:58:35.570215Z"},"phase":"finished","pid":31897,"started_at":"2026-10-21T10:58:35.659996Z","finished_at":"2026-10-21T11:00:44.113116Z","session_id":"01a15e1c-eb09-7653-979f-301af0f83526","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T05:34:53-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 521 — landing: a car held from a train is never told; its runner shows the candidate banner and can cancel the train's suite |
| repo | ariadne |
| issue file | workshop/issues/000521-landing-a-car-held-from-a-train-is-never-told-its-runner-shows-the-candidate-banner-and-can-cancel-the-train-s-suite.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2aec903b4ed1f4abf297930180777dc0d57ff169..f88059b7d170162b1f74b803073e2387411d3e16 |
| command | sdlc close --issue 521 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T05:34:53-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff correctly reconciles the regression test with the new cancellation API and adds exact run-ID ownership coverage. Runtime behavior is unchanged. The revision records an integration repair without expanding scope. No prior findings require disposition, and no new defects were found.

```findings
{}
```

Done-when: current — the repair strengthens the existing requirement that a held runner cannot cancel the train’s suite.

1. **Strengths**
   - Bulk cancellation supplies the new empty run-ID operand correctly (`cmd/sdlc/landingtrain_held_test.go:198`).
   - Exact-ID cancellation uses the actual reformed suite’s lease filename and requires foreign-checkout refusal (`cmd/sdlc/landingtrain_held_test.go:188`, `:202`).
   - Existing completion checks still verify that survivors merge and held PRs remain open.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Focused conductor, waiter, and membership tests passed in 21.294s. Package compilation confirms signature compatibility. Diff whitespace checks passed. The full suite was not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY — pass:** Reuses the existing fixture and cancellation API.
   - **ARCH-PURE — pass:** Integration assertions remain in the integration test.
   - **ARCH-PURPOSE — pass:** Covers both bulk and exact-ID cancellation against the reformed suite.
   - **ARCH-EVOLVE — pass:** No persisted-format changes.
   
   No new architectural or user-facing surface requires atlas or README updates.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2bb6784e-4b87-41fd-829e-996677b688dc","repository":"github.com/xianxu/ariadne","issue_id":"000521","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/3055ba0017211f83","reviewed_head":"f88059b7d170162b1f74b803073e2387411d3e16","slot":"ariadne:1","requested_at":"2026-10-21T12:33:46.758012Z"},"phase":"finished","pid":42019,"started_at":"2026-10-21T12:33:46.943806Z","finished_at":"2026-10-21T12:34:52.552886Z","session_id":"01a15e74-11d0-7ed1-bd5f-8aee8e1fe564","outcome":"returned"}] -->
