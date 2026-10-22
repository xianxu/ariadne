# Boundary Review — ariadne#535 (whole-issue close)

| field | value |
|-------|-------|
| issue | 535 — landing: a priority grant on a car doesn't reach its train's queued integration suite |
| repo | ariadne |
| issue file | workshop/issues/000535-landing-a-priority-grant-on-a-car-doesn-t-reach-its-train-s-queued-integration-suite.md |
| boundary | whole-issue close |
| milestone | — |
| window | dfa61bf96531e253b2ac2b4300caa295abccf8fb..0daa65594127d8000b33b3d44e87d2f6dea6976c |
| command | sdlc close --issue 535 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T20:02:02-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills the issue’s contract: train suites inherit active grants from their exact composed prefix, observe later grants, and return to FIFO after revoke or expiry. Targeted tests passed, including the production provider/process/flock harness. No blocking defects or documentation gaps found. The recorded full-suite result remains a separate landing prerequisite.

```findings
{}
```

1. **Strengths**

   - Membership is appended only after successful composition, excluding rejected cars (`cmd/sdlc/landingtrain.go:322–325`).
   - Exact-SHA projection limits grants to the tested prefix and excludes stale revisions (`cmd/sdlc/landingtrain.go:400`).
   - Non-test checks and unrecorded diagnostics explicitly clear priority identities (`cmd/sdlc/localcheck_run.go:118`).
   - README, atlas, and shipped help document the new behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Targeted Go selection passed in **61.023s**, covering train membership, split/reintegration, diagnostic exclusions, grant lifecycle, and the real helper.
   - All **7 Python unit tests** passed, including malformed metadata and counted N/2N projection.
   - An in-memory negative control restored singular-issue projection and made the new train regression fail.
   - An additional actual ticket → queue → slot check confirmed preservation of unknown top-level and nested metadata.
   - `git diff --check` passed. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Train identities expand into the existing singular grant-provider requests; grant authority and refresh logic remain shared.
   - **ARCH-PURE — pass:** Membership projection and queue ordering remain deterministic; IO stays in the conductor, local-check, and provider seams.
   - **ARCH-PURPOSE — pass:** Both admission and displayed classification use train-aware deadlines. Tests cover late grant, revoke, expiry, prefix isolation, and repository isolation.
   - **ARCH-EVOLVE — pass:** Optional metadata is additive, legacy issue-zero behavior remains FIFO, malformed lists confer no priority, and record rewrites preserve unknown members.

7. **Plan revision recommendations:** None. The implementation matches the bounded design; the remaining suite/close/landing checklist appropriately tracks completion work.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b6be0809-e4dd-442d-a8d2-48494ef1e0e1","repository":"github.com/xianxu/ariadne","issue_id":"000535","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/946508aa0ed60fb4","reviewed_head":"0daa65594127d8000b33b3d44e87d2f6dea6976c","slot":"ariadne:3","requested_at":"2026-10-22T02:59:30.89132Z"},"phase":"finished","pid":81744,"started_at":"2026-10-22T02:59:31.075482Z","finished_at":"2026-10-22T03:02:02.012351Z","session_id":"01a1618c-ac1f-71e0-88b8-be165ae3e58a","outcome":"returned"}] -->
