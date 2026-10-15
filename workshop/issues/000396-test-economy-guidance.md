---
id: 000396
status: open
deps: [392, 393, 394, 395]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'd67eba88b37132a04f42212667f47efe511981d7' # card fields mirrored from issue-cards; edit via sdlc
---

# Test-economy guidance as the SDLC default

## Problem

The #390 test-economy design only becomes the fleet's default if agents in every SDLC repository see one consistent rule at the point of work. The approved evidence tiers, landing admission, flake response, measurement and minimization policy currently live in a design plan and side docs.

## Spec

After their implementing gates land, compile the approved #390 pattern into ariadne's base-layer guidance and atlas, and propagate it to downstream SDLC repositories. State targeted development/close evidence, exact-head full-suite landing, issue-backed known-flake reruns, daily flake ownership, weekly brain snapshots, per-repo serial admission and feature-scoped minimization. Keep deferred package budgets identified as track-first, with no 600/240 s policy. Document which rule each binary gate enforces versus which requires TL review. Treat ariadne/pair as pilots but the resulting mechanism as the default for every SDLC repo, without per-repo opt-in.

## Done when

- [ ] Shared guidance and atlas describe the approved behavior and link to its owning commands/issues; no stale conflicting rule remains in ariadne or pair.
- [ ] Base-layer propagation and at least one cold downstream compile show the default reaches consumers; changes to local-gates files receive the operator's verification path.
- [ ] A worker following only the shared guidance can find the weekly snapshot, the daily flake list and the exact landing requirement; budget values remain deferred.

## Plan

- [ ] Inventory guidance consumers after the implementation issues land.
- [ ] Update one source of truth, propagate and verify ariadne/pair and one cold consumer.

## Log

### 2026-10-15

Filed from #390's approved default-scope decision. Coordinate with #392-#395; do not document unlanded behavior as current.
