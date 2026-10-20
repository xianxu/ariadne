# Boundary Review — ariadne#492 (whole-issue close)

| field | value |
|-------|-------|
| issue | 492 — Landing-train suite attempts omit slot provenance, so overlapping issues' sdlc actual refuses |
| repo | ariadne |
| issue file | workshop/issues/000492-landing-train-suite-attempts-omit-slot-provenance-so-overlapping-issues-sdlc-actual-refuses.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7939055521971b04141867c9773551d21acb01e0..12aff78f6515173afc4aabd5f28ee7f47b4306e4 |
| command | sdlc close --issue 492 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T09:09:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The change fulfills #492: train attempts retain conductor provenance, and overlapping issues measure successfully while missing-slot evidence still refuses. No blocking findings. The implementation and Plan match; atlas documents the ownership boundary, and no new README-facing surface is introduced.

Done-when: current — the criteria cover the persisted slot and successful overlapping-issue measurement; the revision completes the same caller/content boundary.

```findings
{}
```

1. Strengths

   - One pure selector centralizes caller identity in `cmd/sdlc/localcheck_run.go:62`.
   - The train regression verifies persisted provenance plus live queue, lease, and capacity-helper ownership.
   - The actual-command regression exercises both valid foreign-slot evidence and missing-slot refusal.
   - Split/baseline assertions and the relative common-directory regression cover adjacent paths.

2. Critical findings: None.

3. Important findings: None.

4. Minor findings: None.

5. Test coverage notes

   Both targeted runs passed: three boundary regressions (39.035s), then seven surrounding local-check and train controls (55.316s). `git diff --check` passed. The checkout remained clean. Full-suite and mutation testing were not performed during this read-only review.

6. Architectural notes

   - **ARCH-DRY: pass** — all five changed sites reuse the same selector.
   - **ARCH-PURE: pass** — selection is pure; existing IO shells resolve and persist identity.
   - **ARCH-PURPOSE: pass** — ownership consumers are swept, while helper/content and priority roots retain their intended semantics. Known-flake reruns inherit the parent attempt’s slot.
   - **ARCH-EVOLVE: pass** — the existing optional `slot` field is populated without schema changes, historical rewrites, or changes to decoding strictness.

7. Plan revision recommendations: None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6eb27442-2ec7-4de4-8afc-a3801a1d4e4f","repository":"github.com/xianxu/ariadne","issue_id":"000492","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/bbcab8e3715ce1ac","reviewed_head":"12aff78f6515173afc4aabd5f28ee7f47b4306e4","slot":"ariadne:3","requested_at":"2026-10-20T16:07:23.453844Z"},"phase":"finished","pid":23478,"started_at":"2026-10-20T16:07:23.722935Z","finished_at":"2026-10-20T16:09:09.697828Z","session_id":"01a15a11-47fe-7df3-86e8-c92e6ab28efb","outcome":"returned"}] -->
