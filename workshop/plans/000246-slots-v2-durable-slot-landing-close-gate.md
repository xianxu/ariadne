---
gate: boundary-review
issue: 246
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-23T17:30:36-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Replace the undefined ARCH-STATE marker with ARCH-ORDER
          detail: |-
            The plan cites `ARCH-STATE` in the operating-envelope paragraph (plan:66), but the architecture registry defines this principle as ARCH-ORDER. Rename the marker so the state/event design is machine-recognizable and consistently reviewed.
            (carried from plan-quality PQ-1, deferred to the boundary review)
          family: architecture-marker-integrity
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-23T17:30:36-07:00"
      agent: codex
      findings:
        - id: BR-2
          severity: Critical
          title: Post-switch cleanup rejects preserved tracked dirt on the resting branch
          detail: cmd/sdlc/landing.go:246, :274, and :290 route post-switch validation through landingClean, which rejects tracked dirt. A non-colliding dirty resting branch can therefore be switched to successfully but cannot complete deletion or retry, stranding the already-integrated workspace. Separate pre-integration cleanliness from post-switch identity validation and add a regression test. ARCH-ORDER, ARCH-PURPOSE.
          family: resting-checkout-dirt-preservation
          round: 2
        - id: BR-3
          severity: Minor
          title: Issue specification retains undefined ARCH-STATE marker
          detail: workshop/issues/000246-slots-v2-durable-slot-landing.md:57 cites ARCH-STATE, while the registry and corrected plan use ARCH-ORDER. Update the issue specification for consistent architectural traceability.
          family: architecture-marker-integrity
          round: 2
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#246 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-23T17:30:36-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `architecture-marker-integrity` Replace the undefined ARCH-STATE marker with ARCH-ORDER
  The plan cites `ARCH-STATE` in the operating-envelope paragraph (plan:66), but the architecture registry defines this principle as ARCH-ORDER. Rename the marker so the state/event design is machine-recognizable and consistently reviewed.
  (carried from plan-quality PQ-1, deferred to the boundary review)

## Round 2 — 2026-09-23T17:30:36-07:00 (codex) — BLOCKED

### Raised

- **BR-2** [Critical] `resting-checkout-dirt-preservation` Post-switch cleanup rejects preserved tracked dirt on the resting branch
  cmd/sdlc/landing.go:246, :274, and :290 route post-switch validation through landingClean, which rejects tracked dirt. A non-colliding dirty resting branch can therefore be switched to successfully but cannot complete deletion or retry, stranding the already-integrated workspace. Separate pre-integration cleanliness from post-switch identity validation and add a regression test. ARCH-ORDER, ARCH-PURPOSE.
- **BR-3** [Minor] `architecture-marker-integrity` Issue specification retains undefined ARCH-STATE marker
  workshop/issues/000246-slots-v2-durable-slot-landing.md:57 cites ARCH-STATE, while the registry and corrected plan use ARCH-ORDER. Update the issue specification for consistent architectural traceability.

## Open findings

- **BR-1** [Minor] `architecture-marker-integrity` Replace the undefined ARCH-STATE marker with ARCH-ORDER
- **BR-2** [Critical] `resting-checkout-dirt-preservation` Post-switch cleanup rejects preserved tracked dirt on the resting branch
- **BR-3** [Minor] `architecture-marker-integrity` Issue specification retains undefined ARCH-STATE marker
