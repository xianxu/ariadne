---
gate: boundary-review
issue: 260
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T12:32:29-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Post-move verify does not check the source's resting HEAD is unchanged
          detail: move.go:63 passes head "" for the source row, so only the branch name is checked; the Spec requires "source is on its unchanged resting HEAD". Record rev-parse of from.Resting in moveFacts during observeMove and verify it after the switch.
          family: postcondition-verify-incomplete
          round: 1
        - id: BR-2
          severity: Minor
          title: helptext/move.md omits the no-configured-upstream refusal
          detail: observeMove errors when the target resting branch has no upstream; the help refusal list does not mention it.
          family: help-refusal-list-incomplete
          round: 1
        - id: BR-3
          severity: Minor
          title: Post-move verify failure does not say both switches already ran
          detail: move.go:66 error reads like a precondition failure; say the move ran and what state each slot is in.
          family: error-states-partial-progress
          round: 1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-28T12:38:34-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: Code fix correct (RestHead observed, in DeepEqual, verified at move.go:68-75) but no test fails without it; add a slot1-only post-checkout hook that advances main-slot1 and assert the "both switches ran" error.
          round: 2
        - id: BR-2
          disposition: addressed
          note: helptext/move.md now lists the no-configured-upstream refusal; TestMoveRefusals "destination without upstream" pins the refusal and that nothing changed.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: Wording at move.go:71-74 is right, but no test asserts it; the BR-1 hook test covers both.
          round: 2
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-09-28T12:42:16-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: RestHead is recorded in observeMoveSide (move.go:144-150) and verified after the move (move.go:68-75). TestMoveVerifiesSourceRestingHead fails once the old empty-head skip is restored in a scratch copy, confirmed.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Both failure messages after the switches now start "both switches ran, but ..." (move.go:71,74), and the new test asserts it.
          round: 3
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#260 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T12:32:29-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `postcondition-verify-incomplete` Post-move verify does not check the source's resting HEAD is unchanged
  move.go:63 passes head "" for the source row, so only the branch name is checked; the Spec requires "source is on its unchanged resting HEAD". Record rev-parse of from.Resting in moveFacts during observeMove and verify it after the switch.
- **BR-2** [Minor] `help-refusal-list-incomplete` helptext/move.md omits the no-configured-upstream refusal
  observeMove errors when the target resting branch has no upstream; the help refusal list does not mention it.
- **BR-3** [Minor] `error-states-partial-progress` Post-move verify failure does not say both switches already ran
  move.go:66 error reads like a precondition failure; say the move ran and what state each slot is in.

## Round 2 — 2026-09-28T12:38:34-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — not-addressed — Code fix correct (RestHead observed, in DeepEqual, verified at move.go:68-75) but no test fails without it; add a slot1-only post-checkout hook that advances main-slot1 and assert the "both switches ran" error.
- BR-2 — addressed — helptext/move.md now lists the no-configured-upstream refusal; TestMoveRefusals "destination without upstream" pins the refusal and that nothing changed.
- BR-3 — not-addressed — Wording at move.go:71-74 is right, but no test asserts it; the BR-1 hook test covers both.

## Round 3 — 2026-09-28T12:42:16-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — RestHead is recorded in observeMoveSide (move.go:144-150) and verified after the move (move.go:68-75). TestMoveVerifiesSourceRestingHead fails once the old empty-head skip is restored in a scratch copy, confirmed.
- BR-3 — addressed — Both failure messages after the switches now start "both switches ran, but ..." (move.go:71,74), and the new test asserts it.

## Open findings

(none — every finding has been disposed)
