---
id: 000409
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'a27256d687d7cf74594fb3c2b14e9d4e84e6e545' # card fields mirrored from issue-cards; edit via sdlc
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

## Plan

- [ ]

## Log

### 2026-10-16
