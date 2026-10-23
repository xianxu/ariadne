---
id: 000558
status: open
deps: [ariadne#556]
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6a00d43284025c98ffccaee3406d47431cb25b2d' # card fields mirrored from issue-cards; edit via sdlc
---

# Distribute the typed flake Plenary producer through the base manifest with operator verification

## Problem

ariadne#556 adds the typed script/Lua flake implementation and the Plenary producer source `scripts/sdlc-plenary-events.lua`. Peer adoption needs that producer distributed through the base layer. The required `construct/base.manifest` row is protected by `local-gates-guard`, so ops:0 explicitly split it from #556 into this operator-verified issue (Couch delivery `8556cb2f-b354-48e4-ad56-4a99963da64e`).

## Spec

After ariadne#556 lands, add only `symlink   scripts/sdlc-plenary-events.lua` to Ariadne's `construct/base.manifest`, beside the shared SDLC scripts. Do not alter live local-check declarations, seeded workflows, test policy or producer behavior. The operator holds local-gates verification and applies any required `local-gates-verified` label; agents never apply it.

## Done when

- The one distribution row is landed through its own normal issue/PR with operator-held local-gates verification.
- A private consuming-repo weave fixture resolves the symlink to the owning Ariadne producer source; existing consumer-owned files retain the existing collision/refusal behavior.
- ops:0 receives the outcome and can coordinate Pair/Parley runner/declaration adoption after binary rollout.

## Plan

- Wait for #556, add the single manifest row, verify private base-layer installation, and obtain operator verification before landing.

## Log

### 2026-10-22

Filed unclaimed by ariadne:3 at ops:0's direction. #556 deliberately leaves the distribution row out. This issue owns distribution and its local-gates approval; peer runner/declaration adoption follows it.
