---
gate: boundary-review
issue: 202
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T00:42:41-07:00"
      dispatched: "2026-10-11T00:37:14-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Dispute rationale/evidence is inlined unfenced into the next judge prompt
          detail: renderLiveDisputes (gatestate/prompt.go) emits implementer-authored free text as prompt prose; render it as quoted data so it cannot read as instructions to the judge (ARCH-SECURE).
          family: untrusted-text-in-judge-prompt
          round: 1
        - id: BR-2
          severity: Minor
          title: finding-dispute help says Important/Minor "stops blocking" but a REWORK verdict still blocks the boundary
          detail: The dispute settles the ledger side only; close.go's verdict-AND-ledger rule means a judge answering REWORK still holds the close. One sentence in helptext/finding-dispute.md fixes it.
          family: help-overstates-gate-effect
          round: 1
        - id: BR-3
          severity: Minor
          title: Plan gate reads its ledger before dispatch and persists after, so a dispute written mid-dispatch could be dropped
          detail: The boundary path re-reads at persist time; the plan gate does not. Likely serialized by the mutating-command lock, so this is noted rather than a defect.
          family: ledger-read-modify-write-window
          round: 1
      recipe: milestone-review
      reviewed: c79efc0373fffd5fd340d6929c7dc6116b46112d
      blocked: false
---

# Gate ledger — ariadne#202 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T00:42:41-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `untrusted-text-in-judge-prompt` Dispute rationale/evidence is inlined unfenced into the next judge prompt
  renderLiveDisputes (gatestate/prompt.go) emits implementer-authored free text as prompt prose; render it as quoted data so it cannot read as instructions to the judge (ARCH-SECURE).
- **BR-2** [Minor] `help-overstates-gate-effect` finding-dispute help says Important/Minor "stops blocking" but a REWORK verdict still blocks the boundary
  The dispute settles the ledger side only; close.go's verdict-AND-ledger rule means a judge answering REWORK still holds the close. One sentence in helptext/finding-dispute.md fixes it.
- **BR-3** [Minor] `ledger-read-modify-write-window` Plan gate reads its ledger before dispatch and persists after, so a dispute written mid-dispatch could be dropped
  The boundary path re-reads at persist time; the plan gate does not. Likely serialized by the mutating-command lock, so this is noted rather than a defect.

## Open findings

- **BR-1** [Minor] `untrusted-text-in-judge-prompt` Dispute rationale/evidence is inlined unfenced into the next judge prompt
- **BR-2** [Minor] `help-overstates-gate-effect` finding-dispute help says Important/Minor "stops blocking" but a REWORK verdict still blocks the boundary
- **BR-3** [Minor] `ledger-read-modify-write-window` Plan gate reads its ledger before dispatch and persists after, so a dispute written mid-dispatch could be dropped
