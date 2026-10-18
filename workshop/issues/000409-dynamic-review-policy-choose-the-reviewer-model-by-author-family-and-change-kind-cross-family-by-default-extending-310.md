---
id: 000409
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-18
estimate_hours: 1.17
card_mirror: '2cee3e4ed193aa2c6dc68ce88d986d033dd17790' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T05:16:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Dynamic review policy: choose the reviewer model by author family and change kind (cross-family by default), extending #310

## Problem

ariadne#310 made the reviewer agent a static per-repo policy. The operator (2026-10-16): having different models
review is useful robustness; the policy should be dynamic, choosing the reviewer by which model family designed and
coded the change, and by condition. Evidence from 10-15/16: Codex-authored designs under safety requirements
over-engineered (pair#556/#584, #587, #571), while Claude's reviews and redesigns caught it; Codex reviews caught real
bugs in Claude-authored code too.

## Spec

- A review policy with multiple configurations, per repo and per stage (plan-quality, boundary/close review), picking
  the reviewer from: the author's model family (recorded from the session that wrote the commits, as sdlc already does
  for actuals), the change kind (design-heavy: a new mechanism, state or invariant; or bounded), and repo defaults.
- Default rule: cross-family review (Codex-authored → Claude reviewer; Claude-authored → Codex reviewer), with a
  per-repo override and a quota-aware fallback (if the preferred family is unavailable or over budget, use the other,
  stated in the verdict).
- The verdict trailer records author family and reviewer family, so test-economy and retros can measure what each
  catches.
- Static #310 policy remains the floor; the dynamic layer only chooses within it.

## Done when

- Tests: the policy picks per author family, kind and override; the fallback is stated; trailers record both families.
- One week's verdict trailers can be tallied by author and reviewer family.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module   design=0.1 impl=0.24
item: smaller-go-module      design=0.05 impl=0.2
item: smaller-go-module      design=0.05 impl=0.16
item: atlas-docs             design=0.0 impl=0.06
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 1.17
```

Derivation: design mostly settled by two plan-quality rounds (small design= values). impl= at 40% of the v2
midpoints (v3.1): the pure family/ReviewPlan/Unavailable module (greenfield, 0.6 → 0.24); the dispatch loop and
seven-site wiring (0.5 → 0.2); ledger, sidecar, trailer and tally (0.4 → 0.16); atlas and help (0.15 → 0.06); two
milestone reviews (0.35 → 0.14 each). Calibration doc flagged stale; numbers provisional.

## Plan

Durable plan: `workshop/plans/000409-dynamic-review-policy-choose-the-reviewer-model-by-author-family-and-change-kind-cross-family-by-default-extending-310-plan.md`.

- [x] M1 — family, change kind (design-heavy/bounded), family-aware `other` + one default constant (still `same`), pure `ReviewPlan`, `Unavailable` (no-review-began evidence), pre-launch skips
- [ ] M2 — dispatch loop with visible fallback/refusal at every review site; ledger, sidecar, `Review-Reviewer:` trailer; help and atlas

## Log

### 2026-10-16

### 2026-10-18

- Claimed by ariadne:2 (card was unowned). Current-state map in the durable plan; plan sent to the hub before code.
- Hub approved; plan-quality round 1 raised PQ-1..5 (I had missed the issue's Spec: kind = design-heavy/bounded, #310 as floor, the over-budget fallback). All folded in as plan Revisions; the default stays `same` behind one constant until the operator decides the flip.
- M1 implemented: `judge.Family`/`FamilyOf`, `ChangeKind` (design-heavy/bounded), `DefaultReviewerMode = ReviewerSame`
  (the one constant the operator's answer changes), `ReviewerPolicy.Orders`/`Models` with validation and loader
  fields (`reviewer_order`, `reviewer_models`), the pure `ReviewPlan`, `Unavailable` (launch failure, or a stream that
  failed before any assistant/tool event with an unreachable/quota/sign-in cause; `AgentRun.Began` added), and
  `LocalAvailability` (CLI on PATH; codex under checkout containment on darwin). Tests: TestFamilyOf, TestReviewPlan
  (13 rows: floor modes, kinds, mixed/unknown authors, skips, repo order), TestReviewPlanCarriesModels,
  TestUnavailable (8 rows), TestValidateOrdersAndModels, TestLocalAvailability, and the loader test.
- Deviation from the hub-approved order, kept to honour #310 as the floor: the bounded default is codex, claude,
  gemini (not codex, gemini, claude). #310's `other` promised that claude and codex review each other, and with
  gemini second a codex-authored bounded change would have gone to gemini. Gemini is the fallback in both kinds. The
  one behaviour change under `other` is a gemini-driven session, now reviewed by codex rather than claude.
- M1 boundary review round 1 (FIX-THEN-SHIP): BR-1..5 and both Minors fixed (see the plan's second Revisions entry):
  order validation and a never-same-family `Reviewer()`; per-CLI unavailability evidence with measured codex and
  gemini samples (gemini refuses entirely on this machine: `IneligibleTierError`); the model flag wired through
  `DispatchOptions.Model`; help and atlas updated; the plan table corrected; dead `trimReason` removed; flag-like
  models and duplicate order entries refused.
- M1 boundary review round 2: BR-1..3, 5 and both Minors disposed addressed. BR-4 stayed open, because nothing tested
  that the boundary review passes the configured model: `TestBoundaryReviewPassesTheConfiguredModel` (a configured
  model reaches the reviewer's argv; an unconfigured one adds no flag) fails with the wiring removed, checked by
  revert. The new Minor (help and atlas described M2 behaviour) is fixed by marking the kind order and the fallback
  as landing in M2.
