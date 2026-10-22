# Boundary Review — ariadne#554 (whole-issue close)

| field | value |
|-------|-------|
| issue | 554 — actual: a legacy-claimed issue with post-cutover commits still reports 'no post-cutover owned interval' (echo#5) |
| repo | ariadne |
| issue file | workshop/issues/000554-actual-a-legacy-claimed-issue-with-post-cutover-commits-still-reports-no-post-cutover-owned-interval-echo-5.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5f961bcf88d0196ccc33109543d4cb6a5b49361c..bf6e38e1ebc285ef74315d90849b2abe2a485db6 |
| command | sdlc close --issue 554 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T11:39:14-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation matches the revised scope: proven missing ownership produces actionable claim guidance, and hybrid measurements expose the uncounted gap without changing attribution arithmetic or persisted formats. No blocking defects found. Live Echo proof remains outstanding under the plan’s explicit post-review, pre-landing hold.

Done-when: current — the superseding criteria in `## Revisions` cover recovery guidance, gap exclusion, and retained replay.

```findings
{}
```

1. **Strengths**
   - Recovery requires validated pinned history, preserving later-claim and released-owner distinctions (`cmd/sdlc/actualinputs.go:507`).
   - Preview and close consume the same typed recovery metadata.
   - Real Claude/Codex lifecycle fixtures verify migration, reconcile, claim, excluded-gap credit, and immutable replay.
   - README, help, and atlas document the changed behavior.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** Independently passed ownership tests, ownership race tests, tracker engagement/history controls, selected legacy-cutover/replay controls, and diff whitespace checks. The full suite and live Echo measurement were not independently verified.

6. **Architectural notes**
   - **ARCH-DRY: pass** — shared error formatter and preview/close rendering.
   - **ARCH-PURE: pass** — diagnostic formatting remains pure; history IO stays in the existing resolver.
   - **ARCH-PURPOSE: pass** — revised recovery and gap-visibility requirements are implemented across preview, close, and replay.
   - **ARCH-EVOLVE: pass** — transient metadata and derived warnings introduce no persisted-format change.

7. **Plan revision recommendations:** None. Retain the documented suite/live-proof hold before landing.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"52dbebc9-6cf2-40fd-b08d-a3def93c9340","repository":"github.com/xianxu/ariadne","issue_id":"000554","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/cd3daa1a45a0f36e","reviewed_head":"bf6e38e1ebc285ef74315d90849b2abe2a485db6","slot":"ariadne:4","requested_at":"2026-10-22T18:36:35.094779Z"},"phase":"finished","pid":87965,"started_at":"2026-10-22T18:36:35.340666Z","finished_at":"2026-10-22T18:39:13.288218Z","session_id":"01a164e6-971a-7a12-804f-9c1433a5709e","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T12:44:48-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 554 — actual: a legacy-claimed issue with post-cutover commits still reports 'no post-cutover owned interval' (echo#5) |
| repo | ariadne |
| issue file | workshop/issues/000554-actual-a-legacy-claimed-issue-with-post-cutover-commits-still-reports-no-post-cutover-owned-interval-echo-5.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9cc064c0ad2a06c8746853c1a87803523130d5a5..c7778d78873aaa2a2e61e48ef295c6fa181df6bd |
| command | sdlc close --issue 554 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T12:44:48-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The queue-boundary implementation passed focused tests and includes matching documentation. The blocking issue is review scope: this window adds behavior authorized after the previous close, with no prior finding requesting it. The supplied fix-review contract requires a full `sdlc close` for that change.

Done-when: current — the superseding criteria cover measurement of the evidenced owned suffix, excluded-gap visibility, immutable replay, and owner-run live proof.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: post-close-scope-review
    title: |
      scope changed after close — run a full `sdlc close`
    detail: |
      workshop/plans/000554-cutover-ownership-recovery-plan.md:79-88 and the issue's new Revisions passage authorize first-claim queue clipping, implemented at cmd/sdlc/actualinputs.go:413-414. No prior finding requested this behavior; it changes which queue evidence hybrid measurement accepts. The fix-only review contract explicitly requires a full sdlc close for added behavior or expanded scope, regardless of the recorded approval.
```

1. **Strengths**

   - The attachment change reuses pinned engagements and the existing strict selector.
   - The new Claude/Codex fixture exercises captured queue evidence, exact owned-window clipping, preclaim exclusion, and immutable replay.
   - README, help, and atlas consistently document the new behavior.

2. **Critical findings**

   - Scope expansion at `workshop/plans/000554-cutover-ownership-recovery-plan.md:79`. Run a full `sdlc close` against the expanded contract.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed the new lifecycle/replay regression, queue-boundary tests, selector controls, and tested-issue ownership controls: two runs completed in 27.435s and 30.186s. Diff whitespace checks passed. I did not independently reproduce the reported pre-fix failure, run the full suite, or verify live Echo output.

6. **Architectural notes**

   - **ARCH-DRY: pass** — existing engagement derivation and queue selection are reused.
   - **ARCH-PURE: pass** — the IO attachment layer supplies the bound; selection remains pure.
   - **ARCH-PURPOSE: pass for the interdiff** — tests exercise the production attachment seam and retained replay.
   - **ARCH-EVOLVE: pass** — no persisted-format change or rewrite path is introduced.

7. **Plan revision recommendations**

   Record that the expanded first-claim queue contract requires full close review; replace the revision plan’s fix-only re-close instruction accordingly. Retain the recorded-suite and owner-run live-proof holds.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8a68198c-8773-49e3-9fe9-f3629478aab3","repository":"github.com/xianxu/ariadne","issue_id":"000554","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/bcd26a9fac12e5e3","reviewed_head":"c7778d78873aaa2a2e61e48ef295c6fa181df6bd","slot":"ariadne:4","requested_at":"2026-10-22T19:43:14.663186Z"},"phase":"finished","pid":44074,"started_at":"2026-10-22T19:43:15.016283Z","finished_at":"2026-10-22T19:44:47.383643Z","session_id":"01a16523-a053-7272-ba1e-225b4b43c553","outcome":"returned"}] -->
