---
gate: plan-quality
issue: 344
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-13T14:05:43-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: A staged :0 shows as needs-recovery (detached) in the fleet view; the plan never mentions JudgeCheckout
          detail: cmd/sdlc/internal/fleet/slots.go:84 adds a detached recovery reason for every detached checkout, including the :0 host (slots.go:217,225). Either have the inventory read the staged record and report a distinct reason that does not need recovery, or state the exclusion and why.
          family: new-state-unknown-to-consumers
          round: 1
        - id: PQ-2
          severity: Important
          title: Spec says any later normal use returns :0 to rest; the plan clears the record only in stage --clear and move
          detail: A manual git switch, or another verb run in :0, leaves a stale record. State this as a non-goal, and say Leavable trusts the record only when it equals HEAD exactly, so a stale or hand-edited record cannot make another detached commit look safe to leave.
          family: spec-narrowing-unstated
          round: 1
        - id: PQ-3
          severity: Minor
          title: The plan does not say the staged record is written only after a verified switch
          detail: With the write after the switch, a crash in between leaves :0 at a commit the issue branch contains, which the ref rule already treats as leavable. Write that order and the reason into the plan.
          family: effect-ordering-unstated
          round: 1
        - id: PQ-4
          severity: Minor
          title: Task 2 lists test cases one by one instead of a strategy line per risky function
          detail: 'Replace the list with: a table test over target states x verbs for checkStage, checkUnstage and Leavable, asserting the refusal or acceptance and that a refused run changed nothing.'
          family: test-prose-enumeration
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T14:07:06-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: TreeRow.Staged (exact HEAD match) gives the hold staged in JudgeCheckout; any other detached checkout stays recovery (verified slots.go:85).
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Non-goal says the record can go stale and is trusted only on an exact HEAD match.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: 'Effect order section: the record is written only after a verified switch, with the crash reasoning.'
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Task 2 uses one table strategy (states x verbs, a refused run changes nothing).
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-13T14:08:16-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: TreeRow.Staged plus JudgeCheckout hold `staged`; other detached checkouts stay needs-recovery.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Explicit non-goal; the record is trusted only on an exact HEAD match.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: 'The Effect order paragraph: the record is written after the verified switch.'
          round: 3
        - id: PQ-4
          disposition: addressed
          note: One table-test strategy line plus one real-git flow.
          round: 3
      blocked: false
content_hash: 778ffe0432a7f287e4aab78af78c83c13f16374654795c66a70b4b06235c0f37
---

# Gate ledger — ariadne#344 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T14:05:43-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `new-state-unknown-to-consumers` A staged :0 shows as needs-recovery (detached) in the fleet view; the plan never mentions JudgeCheckout
  cmd/sdlc/internal/fleet/slots.go:84 adds a detached recovery reason for every detached checkout, including the :0 host (slots.go:217,225). Either have the inventory read the staged record and report a distinct reason that does not need recovery, or state the exclusion and why.
- **PQ-2** [Important] `spec-narrowing-unstated` Spec says any later normal use returns :0 to rest; the plan clears the record only in stage --clear and move
  A manual git switch, or another verb run in :0, leaves a stale record. State this as a non-goal, and say Leavable trusts the record only when it equals HEAD exactly, so a stale or hand-edited record cannot make another detached commit look safe to leave.
- **PQ-3** [Minor] `effect-ordering-unstated` The plan does not say the staged record is written only after a verified switch
  With the write after the switch, a crash in between leaves :0 at a commit the issue branch contains, which the ref rule already treats as leavable. Write that order and the reason into the plan.
- **PQ-4** [Minor] `test-prose-enumeration` Task 2 lists test cases one by one instead of a strategy line per risky function
  Replace the list with: a table test over target states x verbs for checkStage, checkUnstage and Leavable, asserting the refusal or acceptance and that a refused run changed nothing.

## Round 2 — 2026-10-13T14:07:06-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — TreeRow.Staged (exact HEAD match) gives the hold staged in JudgeCheckout; any other detached checkout stays recovery (verified slots.go:85).
- PQ-2 — addressed — Non-goal says the record can go stale and is trusted only on an exact HEAD match.
- PQ-3 — addressed — Effect order section: the record is written only after a verified switch, with the crash reasoning.
- PQ-4 — addressed — Task 2 uses one table strategy (states x verbs, a refused run changes nothing).

## Round 3 — 2026-10-13T14:08:16-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — TreeRow.Staged plus JudgeCheckout hold `staged`; other detached checkouts stay needs-recovery.
- PQ-2 — addressed — Explicit non-goal; the record is trusted only on an exact HEAD match.
- PQ-3 — addressed — The Effect order paragraph: the record is written after the verified switch.
- PQ-4 — addressed — One table-test strategy line plus one real-git flow.

## Open findings

(none — every finding has been disposed)
