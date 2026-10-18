---
id: 000409
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-18
estimate_hours:
card_mirror: 'eb01620b3331f757da9b9cddc960b3d8fd014c0b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T05:16:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] M1 — family, change kind (design-heavy/bounded), family-aware `other` + one default constant (still `same`), pure `ReviewPlan`, `Unavailable` (no-review-began evidence), pre-launch skips
- [ ] M2 — dispatch loop with visible fallback/refusal at every review site; ledger, sidecar, `Review-Reviewer:` trailer; help and atlas

## Log

### 2026-10-16

### 2026-10-18

- Claimed by ariadne:2 (card was unowned). Current-state map in the durable plan; plan sent to the hub before code.
- Hub approved; plan-quality round 1 raised PQ-1..5 (I had missed the issue's Spec: kind = design-heavy/bounded, #310 as floor, the over-budget fallback). All folded in as plan Revisions; the default stays `same` behind one constant until the operator decides the flip.
