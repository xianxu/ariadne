---
gate: plan-quality
issue: 280
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-02T10:54:15-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: The required verb set is never defined, so the coverage test is circular; push, issue sync and move-detail are missing
          detail: '"Every MVP verb has exactly one entry" has no list independent of the catalog, so a missing verb never fails the test. push (a spine verb, helptext/push.md), issue sync and issue move-detail are left out although the Spec names publication. Build the required set from the command tree (the 21 markMutatingCommand sites plus explicit read-only exemptions).'
          family: coverage-set-derives-from-source
          round: 1
        - id: PQ-2
          severity: Important
          title: '`sdlc recovery` clashes with the existing `sdlc issue recovery reconcile` and differs from the recorded `sdlc help recovery` decision'
          detail: issuerecovery.go:22 already defines a mutating `issue recovery` group. The operator's Log decision said `sdlc help recovery`; the plan changes it without a Revisions entry. Pick a name, record it, and cross-link the two.
          family: cli-surface-naming-collision
          round: 1
        - id: PQ-3
          severity: Important
          title: No proof is named for the reopened/reclaimed-generation case or the close-after-SHIP re-run that Done-when requires
          detail: New tests cover only claim and set-status lost-ack. Name an existing or new test for each class Done-when lists (race, duplicate, lost-ack, reopened/reclaimed generation); otherwise those claims render as unproven and the Done-when line is not met.
          family: done-when-case-class-unproven
          round: 1
        - id: PQ-4
          severity: Minor
          title: The uncertain-publication helper should also replace reclaim's inline message at reclaim.go:178
          family: single-source-helper-adoption
          round: 1
        - id: PQ-5
          severity: Minor
          title: Use one injection seam around UpdateCard/trackerEnv, or the existing stateful tracker fake, instead of a package var per verb
          detail: Add a caller-guard test like TestReclaimIsOnlyOperatorInvoked for any new seam.
          family: shared-test-seam
          round: 1
        - id: PQ-6
          severity: Minor
          title: The plan does not say how {{RECOVERY}} renders several verb contracts on one page (issue.md hosts show/sync/move-detail/recovery)
          family: help-page-granularity
          round: 1
      blocked: true
---

# Gate ledger — ariadne#280 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T10:54:15-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `coverage-set-derives-from-source` The required verb set is never defined, so the coverage test is circular; push, issue sync and move-detail are missing
  "Every MVP verb has exactly one entry" has no list independent of the catalog, so a missing verb never fails the test. push (a spine verb, helptext/push.md), issue sync and issue move-detail are left out although the Spec names publication. Build the required set from the command tree (the 21 markMutatingCommand sites plus explicit read-only exemptions).
- **PQ-2** [Important] `cli-surface-naming-collision` `sdlc recovery` clashes with the existing `sdlc issue recovery reconcile` and differs from the recorded `sdlc help recovery` decision
  issuerecovery.go:22 already defines a mutating `issue recovery` group. The operator's Log decision said `sdlc help recovery`; the plan changes it without a Revisions entry. Pick a name, record it, and cross-link the two.
- **PQ-3** [Important] `done-when-case-class-unproven` No proof is named for the reopened/reclaimed-generation case or the close-after-SHIP re-run that Done-when requires
  New tests cover only claim and set-status lost-ack. Name an existing or new test for each class Done-when lists (race, duplicate, lost-ack, reopened/reclaimed generation); otherwise those claims render as unproven and the Done-when line is not met.
- **PQ-4** [Minor] `single-source-helper-adoption` The uncertain-publication helper should also replace reclaim's inline message at reclaim.go:178
- **PQ-5** [Minor] `shared-test-seam` Use one injection seam around UpdateCard/trackerEnv, or the existing stateful tracker fake, instead of a package var per verb
  Add a caller-guard test like TestReclaimIsOnlyOperatorInvoked for any new seam.
- **PQ-6** [Minor] `help-page-granularity` The plan does not say how {{RECOVERY}} renders several verb contracts on one page (issue.md hosts show/sync/move-detail/recovery)

## Open findings

- **PQ-1** [Important] `coverage-set-derives-from-source` The required verb set is never defined, so the coverage test is circular; push, issue sync and move-detail are missing
- **PQ-2** [Important] `cli-surface-naming-collision` `sdlc recovery` clashes with the existing `sdlc issue recovery reconcile` and differs from the recorded `sdlc help recovery` decision
- **PQ-3** [Important] `done-when-case-class-unproven` No proof is named for the reopened/reclaimed-generation case or the close-after-SHIP re-run that Done-when requires
- **PQ-4** [Minor] `single-source-helper-adoption` The uncertain-publication helper should also replace reclaim's inline message at reclaim.go:178
- **PQ-5** [Minor] `shared-test-seam` Use one injection seam around UpdateCard/trackerEnv, or the existing stateful tracker fake, instead of a package var per verb
- **PQ-6** [Minor] `help-page-granularity` The plan does not say how {{RECOVERY}} renders several verb contracts on one page (issue.md hosts show/sync/move-detail/recovery)
