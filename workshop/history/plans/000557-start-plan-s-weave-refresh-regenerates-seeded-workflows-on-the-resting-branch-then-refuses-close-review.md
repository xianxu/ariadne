# Boundary Review — ariadne#557 (whole-issue close)

| field | value |
|-------|-------|
| issue | 557 — start-plan's weave refresh regenerates seeded workflows on the resting branch, then refuses |
| repo | ariadne |
| issue file | workshop/issues/000557-start-plan-s-weave-refresh-regenerates-seeded-workflows-on-the-resting-branch-then-refuses.md |
| boundary | whole-issue close |
| milestone | — |
| window | 10a6b3101bdf7a80ae0b9fd99ff4d60e8c891c68..bb5ae3e39ab6c08ca320f0727501f39f3943fee3 |
| command | sdlc close --issue 557 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T18:00:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy the approved seeded-output scope of #557. Planning refresh suppresses Seed and SeedOnce publication before managed preflight, preserves protected settings, and reports pending regeneration. Native tests confirm clean resting-branch catch-up, subsequent issue-branch creation, and explicit regeneration on a separate consumer branch. No blocking findings.

```findings
{}
```

1. **Strengths**

   - `cmd/weave/deferred_seeds.go:31` provides a pure classifier covering both seed variants without Git-tracking races.
   - `cmd/weave/main.go:579` applies deferral before managed preflight and every publication pass, avoiding fragile restoration after writes.
   - `cmd/weave/planning_refresh_test.go:18` exercises the captured workflow change, preserves legitimate fast-forwards, reproduces the legacy failure, and verifies separate regeneration.
   - README, CLI help, and atlas explain the new flag, policy limits, and operator-verified consumer PR route.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed targeted runs covering the new classifier, destination/evidence preservation, observation bounds, errors, interrupted publication, native refresh, installed-weave/start-plan integration, protected settings, ordinary refresh, and managed deferral. Both Claude and Codex native fixtures passed. Range whitespace checks passed; checkout remained clean.

   The Core concepts entries match their implementation locations. The pure classifier tests require no IO. No prior findings required disposition. The full module suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Reuses OutputCurrent, DeferredOutput, and one compilation/publication path.
   - **ARCH-PURE — pass:** Classification is pure; observation and diagnostics use injected filesystem/writer boundaries.
   - **ARCH-PURPOSE — pass:** Covers both seed variants and the separate regeneration route.
   - **ARCH-MOCK — pass:** Retains existing dependency seams and stateful Git/conformance infrastructure; adds isolated native integration proof.
   - **ARCH-CONSTRAINTS — pass:** Adds sequential classification/currentness work without per-seed subprocesses; counted observations cover scaling.
   - **ARCH-SECURE — pass:** No restoration over caller content, new credential handling, or production-state test dependency.
   - **ARCH-ORDER — pass:** Suppression precedes publication; failure/cancellation tests preserve deferred destinations.
   - **ARCH-FUNERAL — pass:** Creates no new durable artifact family; existing staging cleanup remains intact.
   - **ARCH-EVOLVE — pass:** Changes no persisted format; ordinary publication and settings-only compatibility remain covered.

7. **Plan revision recommendations:** None. The documented seed-only scope matches the implementation; final suite, close, and landing remain explicitly outstanding operational steps.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"670295ed-7b69-45c8-9fed-f8d40a541a9b","repository":"github.com/xianxu/ariadne","issue_id":"000557","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/6f3de822ab873a93","reviewed_head":"bb5ae3e39ab6c08ca320f0727501f39f3943fee3","slot":"ariadne:2","requested_at":"2026-10-23T00:58:59.875995Z"},"phase":"finished","pid":40575,"started_at":"2026-10-23T00:59:00.002906Z","finished_at":"2026-10-23T01:00:25.925361Z","session_id":"01a16644-b189-7b82-9d8d-4303f83cc3f5","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T19:29:59-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 557 — start-plan's weave refresh regenerates seeded workflows on the resting branch, then refuses |
| repo | ariadne |
| issue file | workshop/issues/000557-start-plan-s-weave-refresh-regenerates-seeded-workflows-on-the-resting-branch-then-refuses.md |
| boundary | whole-issue close |
| milestone | — |
| window | da4fcfbec8134f7e116dfe0d3dd86f7608428ecd..63a9053009e44fb1b1fcffb2dfebcfd9ab88e7b4 |
| command | sdlc close --issue 557 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T19:29:59-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff correctly updates the dependency-policy fixture to match the existing `refresh --planning` adapter contract. It preserves the original policy assertions and introduces no production behavior or scope changes. No prior findings require disposition, and no new defects were found.

```findings
{}
```

1. **Strengths**

   - Fixture parser and expected transcript agree with `cmd/sdlc/catchup.go:30`.
   - `cmd/sdlc/dependencycontext_test.go:69` retains checks for child environment, command order, working directories, output streams, and unchanged parent environment.
   - Tracker changes record the fixture failure and correction without changing the Spec, acceptance criteria, or implementation plan.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed dependency-policy tests, native seeded-workflow start-plan tests for Claude and Codex, and installed-weave argv coverage. Range whitespace checks passed; checkout remained clean. The full suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** No duplicated production logic introduced.
   - **ARCH-PURE — pass:** Process-fixture changes remain at the existing integration seam.
   - **ARCH-PURPOSE — pass:** Corrects the stale fixture contract while preserving its policy oracle; no stale flag references remain under `cmd/sdlc`.
   - **ARCH-EVOLVE — pass:** No persisted-format or reader/writer changes.

   No new architectural or user-facing surface requires atlas or README updates.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"fa5105bf-8849-401f-8369-8761ee007b80","repository":"github.com/xianxu/ariadne","issue_id":"000557","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b81c8fbaeb6c56f1","reviewed_head":"63a9053009e44fb1b1fcffb2dfebcfd9ab88e7b4","slot":"ariadne:2","requested_at":"2026-10-23T02:28:50.386452Z"},"phase":"finished","pid":11526,"started_at":"2026-10-23T02:28:50.479169Z","finished_at":"2026-10-23T02:29:58.656141Z","session_id":"01a16696-f1c7-7602-9ebc-aa97e355a050","outcome":"returned"}] -->
