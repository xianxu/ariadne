# Boundary Review — ariadne#512 (whole-issue close)

| field | value |
|-------|-------|
| issue | 512 — Train-car cleanup refuses confirmed archived issue after landing |
| repo | ariadne |
| issue file | workshop/issues/000512-train-archive-cleanup.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5d8be1a3c990d87e2d709b9d2a944e7ee0f31df7..6dc7f788bd172f4d4c21d6251aba929cbd90f1f0 |
| command | sdlc close --issue 512 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T00:38:45-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The change delivers the documented cleanup recovery through the shared archive checker while retaining exact completion and archive-transition verification. No blocking defects found. The focused regression matrix passed, and disabling the fallback made the publisher-interleaving regression fail with the reported refusal.

Done-when: current — the criteria cover sibling-first cleanup, confirmed archive recovery, and refusal without matching durable proof; the revisions do not expand that scope.

```findings
{}
```

1. **Strengths**

   - Shared proof selection covers initial cleanup, fresh retries, publisher rechecks, and final confirmation ([landingarchive.go](/Volumes/TestScratch/review-checkouts/fc172637230846a2/cmd/sdlc/landingarchive.go:663)).
   - Fresh-resume tests verify unchanged completion cards, no repeated server merge, an unchanged resting branch, and deletion of the landed issue branch ([landingtrain_archive_test.go](/Volumes/TestScratch/review-checkouts/fc172637230846a2/cmd/sdlc/landingtrain_archive_test.go:205)).
   - Negative controls exercise missing, altered, malformed, conflicting, ambiguous, wrong-issue, wrong-completion, and untracked archive evidence.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   The focused archive/train/proof matrix passed in 110.875s. Using a temporary Go overlay to disable the fallback, `TestLandingArchiveSiblingPublishesBetweenChecks` failed with `active issue missing without confirmed archive`, confirming regression sensitivity. `git diff --check` passed; the checkout remained clean. The opt-in live conformance harness was inspected but not executed.

6. **Architectural notes**

   - **ARCH-DRY: pass** — callers share proof selection and reuse the existing transition verifier.
   - **ARCH-PURE: pass** — Git and tracker work remains in the existing integration layer, exercised through real-Git fixtures and injected service seams.
   - **ARCH-PURPOSE: pass** — the fix covers fresh recovery and publication interleaving, beyond the transient waiter path.
   - **ARCH-EVOLVE: pass** — no persisted-format changes; legacy untracked archives retain their provenance requirement.

   Atlas documents the revised proof flow. No new command, flag, or configuration surface requires a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4e09254a-c796-423a-b793-36bfcacadbf0","repository":"github.com/xianxu/ariadne","issue_id":"000512","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/fc172637230846a2","reviewed_head":"6dc7f788bd172f4d4c21d6251aba929cbd90f1f0","slot":"ariadne:2","requested_at":"2026-10-21T07:35:40.400229Z"},"phase":"finished","pid":88583,"started_at":"2026-10-21T07:35:40.496084Z","finished_at":"2026-10-21T07:38:44.598378Z","session_id":"01a15d63-23d3-7f41-b7ee-d90aa390d8ee","outcome":"returned"}] -->
