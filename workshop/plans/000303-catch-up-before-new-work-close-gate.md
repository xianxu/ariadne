---
gate: boundary-review
issue: 303
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T09:08:32-07:00"
      dispatched: "2026-10-11T09:03:48-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: A failed weave refresh that already advanced checkouts is treated as "nothing moved" and may plan with the stale binary
          detail: 'weave refresh keeps completed Git updates on error, including when every repo moved and only the compile failed (refresh.go:83,228). catchUpBeforeWork then runs freshness on the now-current HEADs, finds nothing Stale, prints a false "landed after cutoff or docs only" warning, and plans with the pre-refresh sdlc. Fix: snapshot HEADs before weave; if any moved, refuse and ask for a re-run. Add a fake-weave test that fast-forwards and then errors.'
          family: uncertain-outcome-collapsed
          round: 1
        - id: BR-2
          severity: Important
          title: A repo with no origin, or a missing/non-git dependency path, is silently treated as current, contrary to the spec
          detail: catchup.go:96-98 returns (0, "") on any `remote get-url origin` failure. The spec requires a no-origin repo to be reported as unchecked (a warning). substrateChain also returns dependency paths that don't exist, which pass silently. Return a why so cwarn fires (or revise the spec), and add a test.
          family: unchecked-reported-as-current
          round: 1
        - id: BR-3
          severity: Minor
          title: The probe fetch relies on the configured refspec rather than weave's explicit +refs/heads/main:refs/remotes/origin/main
          family: refspec-consistency
          round: 1
        - id: BR-4
          severity: Minor
          title: No tests for fetch-failure warnings, weave missing from PATH, or the real weaveRefresh argv/dir
          family: test-coverage-gap
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-11T09:28:40-07:00"
      dispatched: "2026-10-11T09:14:42-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: HEADs snapshotted pre-weave; any move stops for re-run (catchup.go:73-89). Reverting the guard turns TestStartPlanStopsWhenAFailedRefreshAdvancedACheckout red.
          round: 2
        - id: BR-2
          disposition: addressed
          note: no-origin returns "no origin remote"; substrateWalk surfaces absent deps as unchecked. Reverting both turns TestCatchUpWarnsUncheckedCheckouts red.
          round: 2
        - id: BR-3
          disposition: addressed
          note: fetch now uses explicit +refs/heads/main:refs/remotes/origin/main (catchup.go:128).
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestCatchUpWarnsUncheckedCheckouts covers fetch failure; TestRunInstalledWeaveRunsRefreshInRoot covers argv/dir and weave missing from PATH.
          round: 2
      findings:
        - id: BR-5
          severity: Minor
          title: behindOrigin and file-header comments contradict the round-1 behavior
          detail: catchup.go:116-118 still says a no-origin checkout has nothing to trail (it now warns as unchecked), and the header's "or nothing moves" implies full rollback, which BR-1 showed is false after a compile failure.
          family: stale-doc-comment
          round: 2
        - id: BR-6
          severity: Minor
          title: start-plan help and atlas omit the unchecked-checkout warning and the stop after a partial refresh
          family: docs-surface-gap
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 34545b375f2727108d7b89aea95ba1c8c158fe1d
      blocked: false
    - "n": 3
      timestamp: "2026-10-11T09:33:51-07:00"
      dispatched: "2026-10-11T09:33:17-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: catchup.go:4-5 header now says refresh keeps completed updates on a later failure; catchup.go:118-120 says no-origin is reported unchecked, matching line 124.
          round: 3
        - id: BR-6
          disposition: addressed
          note: helptext/start-plan.md CATCH-UP section and atlas/workflow/sdlc-binary.md:1107-1110 now describe the stop after a partial refresh and the absent/no-origin/failed-fetch unchecked warnings, matching catchup.go:51,124,129 and 76-88.
          round: 3
      findings:
        - id: BR-7
          severity: Minor
          title: New comment and help-text lines are not wrapped to the surrounding width
          detail: catchup.go:5 and one added line in helptext/start-plan.md (CATCH-UP section) run to about 130-150 characters, while neighbouring lines wrap at about 80. Cosmetic only.
          family: prose-line-wrap
          round: 3
      recipe: fix-review
      isolation: checkout
      reviewed: 1ec8628e5997a7db47ca4040c903a7ca3dff6423
      blocked: false
---

# Gate ledger — ariadne#303 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T09:08:32-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `uncertain-outcome-collapsed` A failed weave refresh that already advanced checkouts is treated as "nothing moved" and may plan with the stale binary
  weave refresh keeps completed Git updates on error, including when every repo moved and only the compile failed (refresh.go:83,228). catchUpBeforeWork then runs freshness on the now-current HEADs, finds nothing Stale, prints a false "landed after cutoff or docs only" warning, and plans with the pre-refresh sdlc. Fix: snapshot HEADs before weave; if any moved, refuse and ask for a re-run. Add a fake-weave test that fast-forwards and then errors.
- **BR-2** [Important] `unchecked-reported-as-current` A repo with no origin, or a missing/non-git dependency path, is silently treated as current, contrary to the spec
  catchup.go:96-98 returns (0, "") on any `remote get-url origin` failure. The spec requires a no-origin repo to be reported as unchecked (a warning). substrateChain also returns dependency paths that don't exist, which pass silently. Return a why so cwarn fires (or revise the spec), and add a test.
- **BR-3** [Minor] `refspec-consistency` The probe fetch relies on the configured refspec rather than weave's explicit +refs/heads/main:refs/remotes/origin/main
- **BR-4** [Minor] `test-coverage-gap` No tests for fetch-failure warnings, weave missing from PATH, or the real weaveRefresh argv/dir

## Round 2 — 2026-10-11T09:28:40-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — HEADs snapshotted pre-weave; any move stops for re-run (catchup.go:73-89). Reverting the guard turns TestStartPlanStopsWhenAFailedRefreshAdvancedACheckout red.
- BR-2 — addressed — no-origin returns "no origin remote"; substrateWalk surfaces absent deps as unchecked. Reverting both turns TestCatchUpWarnsUncheckedCheckouts red.
- BR-3 — addressed — fetch now uses explicit +refs/heads/main:refs/remotes/origin/main (catchup.go:128).
- BR-4 — addressed — TestCatchUpWarnsUncheckedCheckouts covers fetch failure; TestRunInstalledWeaveRunsRefreshInRoot covers argv/dir and weave missing from PATH.

### Raised

- **BR-5** [Minor] `stale-doc-comment` behindOrigin and file-header comments contradict the round-1 behavior
  catchup.go:116-118 still says a no-origin checkout has nothing to trail (it now warns as unchecked), and the header's "or nothing moves" implies full rollback, which BR-1 showed is false after a compile failure.
- **BR-6** [Minor] `docs-surface-gap` start-plan help and atlas omit the unchecked-checkout warning and the stop after a partial refresh

## Round 3 — 2026-10-11T09:33:51-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — catchup.go:4-5 header now says refresh keeps completed updates on a later failure; catchup.go:118-120 says no-origin is reported unchecked, matching line 124.
- BR-6 — addressed — helptext/start-plan.md CATCH-UP section and atlas/workflow/sdlc-binary.md:1107-1110 now describe the stop after a partial refresh and the absent/no-origin/failed-fetch unchecked warnings, matching catchup.go:51,124,129 and 76-88.

### Raised

- **BR-7** [Minor] `prose-line-wrap` New comment and help-text lines are not wrapped to the surrounding width
  catchup.go:5 and one added line in helptext/start-plan.md (CATCH-UP section) run to about 130-150 characters, while neighbouring lines wrap at about 80. Cosmetic only.

## Open findings

- **BR-7** [Minor] `prose-line-wrap` New comment and help-text lines are not wrapped to the surrounding width
