---
id: 000428
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '863c8f54b86818c4b81755e51e8c2a8a229ec870' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T11:19:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "868d4ca2", done: "91934a0e"}
---

# Legacy-attempt actuals refusal should name --no-actual (exclude), not 'ask the TL to restore evidence'

## Problem

## Spec

Since ariadne#420, `sdlc close`, `milestone-close` and `actual` refuse a numeric actual when a suite attempt in the window is legacy, meaning it has no exact `queued_at` and `acquired_at`. The refusal tells the worker to "ask the TL/hub to restore named measurement evidence". For a legacy attempt that evidence never existed and can't be restored. The approved disposition is to exclude the sample and close with `--no-actual` and a reason. On 2026-10-17 at least three slots stopped and asked the hub (parley.nvim:1 on #373 M1, pair:5 on #622 M2, ariadne:5 on #427), costing a round trip each.

Wanted: when every blocking attempt is legacy, the refusal says so and names the exact command, for example `sdlc milestone-close --issue N --milestone M --no-actual 'legacy suite attempt <id> lacks exact queue endpoints; excluded from calibration per ariadne#420'`. The "restore evidence" wording stays only for cases where evidence could exist but is missing or malformed. Keep the refusal; don't auto-apply `--no-actual`.

## Done when

- A test: a window whose only blocker is a legacy attempt gets a refusal naming `--no-actual` with the attempt ID and the reason text. A malformed new-format record still gets the restore wording.
- `close`, `milestone-close` and `actual --issue` give the same wording.

## Plan

- [x] Add red rendering tests for legacy `actual --issue`, `close`, and `milestone-close --milestone` refusals, plus malformed-new-format and unrelated-error controls.
- [x] Recognize only the controlled legacy-attempt warning and validated attempt ID; reuse the existing boundary-aware `rerunCmd` to print the matching verb and `--no-actual` with an explicit reason. Keep restore advice for other errors (ARCH-DRY, ARCH-PURE, ARCH-PURPOSE).
- [x] Run focused tests, `make -k build`, `git diff --check`, and the recorded pushed-head suite.

## Log

### 2026-10-17

- Claimed in ariadne:1. `printActual` currently gives every `actualError` the same restore message; `explainActual` already knows the refused boundary and has `rerunCmd`. The TL approved the quick-flow plan with explicit milestone and preview coverage.
- The red rendering test reproduced the misleading restore advice for preview, close, and milestone-close. The focused suite now passes, including new-format and unrelated-error controls. `make -k build` and the sentinel-owned `make sdlc-build` pass. The atlas and CLI help now describe the exclusion path.
- PR #286 has a recorded passing suite on d600913a2cf0f9cb4ce86db827b281d662754ffa: all 12 sdlc shards and 67 other packages passed, zero failed tests. On ariadne:0, `stage --build` verified all seven binaries at that head; staged `sdlc actual --issue 428` ran, and `stage --clear` rebuilt all seven at main 0c268805c1c547b42a45e0792c5493e5d0a79c4d. The first close attempt stopped before review solely because the final checklist row included future close/land work.
- The exact-head suite on b0119e262de3e2a86679ddc24182bc970f1ace75 also passed all shards after a 28-minute queue wait. Close review round 1 returned REWORK (BR-1): only the first selector warning reached the renderer, so a legacy-first mixed window could incorrectly suggest exclusion. This is a complete-blocker classification defect; the selector emits all warnings, but the resolver dropped all except the first.

## Revisions

- 2026-10-17 — Clarify the Spec's example command: `--no-actual` is a boolean switch; the human-readable reason belongs beside the command and in `--verified`, which also needs the issue's verification evidence. An `actual --issue N` preview points to `sdlc close --issue N`; an issue close points to `sdlc close`, and a milestone refusal points to `sdlc milestone-close --issue N --milestone M`. The ID and non-restorable legacy reason accompany each. Other actuals errors retain restore guidance; no gate semantics change (TL decision, ARCH-PURPOSE).
- 2026-10-17 — Separate verification from the future close and landing actions in the final checklist row. The close gate requires every Plan checkbox to be complete before it can run its review; verification is complete, while close and landing follow the gate.
- 2026-10-17 — BR-1: carry the complete queue blocker set from the resolver to the actual result as structured data, preserving every diagnostic in the error text. Offer `--no-actual` only if every blocker is a validated legacy attempt; any malformed or unknown-provenance blocker retains restoration guidance. Add mixed-blocker regression cases in both orders for unknown queue and unknown slot provenance, then exercise preview, close, and milestone rendering (ARCH-PURPOSE).
