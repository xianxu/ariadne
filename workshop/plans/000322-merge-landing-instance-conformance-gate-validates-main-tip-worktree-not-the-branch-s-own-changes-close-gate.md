---
gate: boundary-review
issue: 322
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-10T18:15:39-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Duplicate-id gate measures merge-base from "HEAD" while the instance gate uses the pinned head
          detail: landinggate.go:30 calls landingMergeBase(r, mainOID, "HEAD"); landing.go:406 now passes the pinned head to the instance gate. They are equal after landingClean, but threading head into runLandingDuplicateGate would make both gates measure the same pinned window by construction. Only these two landing gates are in this family.
          family: landing-gates-pinned-window
          round: 1
      recipe: small-diff-review
      reviewed: dfa7d125af64faf13feb320e0cd95e1328337862
      blocked: false
---

# Gate ledger — ariadne#322 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-10T18:15:39-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `landing-gates-pinned-window` Duplicate-id gate measures merge-base from "HEAD" while the instance gate uses the pinned head
  landinggate.go:30 calls landingMergeBase(r, mainOID, "HEAD"); landing.go:406 now passes the pinned head to the instance gate. They are equal after landingClean, but threading head into runLandingDuplicateGate would make both gates measure the same pinned window by construction. Only these two landing gates are in this family.

## Open findings

- **BR-1** [Minor] `landing-gates-pinned-window` Duplicate-id gate measures merge-base from "HEAD" while the instance gate uses the pinned head
