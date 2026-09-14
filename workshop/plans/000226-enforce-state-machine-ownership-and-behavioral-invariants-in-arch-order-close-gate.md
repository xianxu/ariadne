---
gate: boundary-review
issue: 226
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-14T15:40:40-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Name the existing functions covered by the verification step.
          detail: |-
            Identify BuildPrompt, runArchPrinciples, and runStartPlan, with the strategy: missing or partial registry delivery is caught by existing full-registry assertions; compare freshly built CLI delivery against the canonical source. No new production function or enumerated test-case inventory is needed.
            (carried from plan-quality PQ-1, deferred to the boundary review)
          family: named-test-surface
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-14T15:40:40-07:00"
      agent: codex
      findings:
        - id: BR-2
          severity: Important
          title: Update all four prompt snapshots for the intentional policy change.
          detail: cmd/sdlc/internal/judge/architecture.md:162 changes rendered prompts, but dry.prompt, pure.prompt, plan-quality.prompt, and milestone-review.prompt still contain the base registry. TestBuildPrompt_Golden fails; regenerate and inspect these snapshots, then rerun the judge tests. ARCH-PURPOSE.
          family: executable-prompt-snapshot-consistency
          round: 2
        - id: BR-3
          severity: Important
          title: Correct the atlas explanation of ARCH-ORDER review ordering.
          detail: atlas/workflow/architecture-principles.md:80-84 describes oracle-first ordering as deliberate, while architecture.md:192 now leads with production routing and bypass prevention. Update the map-level explanation to match the approved policy without copying its clauses. ARCH-DRY.
          family: architecture-map-consistency
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-14T15:43:57-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The issue's verification revision names BuildPrompt, runArchPrinciples, and runStartPlan and describes full-registry assertions plus CLI comparison; inspected tests support this description.
          round: 3
        - id: BR-2
          disposition: addressed
          note: All four snapshots exactly replace the base registry with the pinned head registry. The complete judge suite, including TestBuildPrompt_Golden, passes; reversing the replacement breaks each snapshot equality.
          round: 3
        - id: BR-3
          disposition: addressed
          note: atlas/workflow/architecture-principles.md:80 now describes executable-model checks preceding the retained oracle critique, matching cmd/sdlc/internal/judge/architecture.md:192.
          round: 3
      blocked: false
---

# Gate ledger — ariadne#226 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-14T15:40:40-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `named-test-surface` Name the existing functions covered by the verification step.
  Identify BuildPrompt, runArchPrinciples, and runStartPlan, with the strategy: missing or partial registry delivery is caught by existing full-registry assertions; compare freshly built CLI delivery against the canonical source. No new production function or enumerated test-case inventory is needed.
  (carried from plan-quality PQ-1, deferred to the boundary review)

## Round 2 — 2026-09-14T15:40:40-07:00 (codex) — BLOCKED

### Raised

- **BR-2** [Important] `executable-prompt-snapshot-consistency` Update all four prompt snapshots for the intentional policy change.
  cmd/sdlc/internal/judge/architecture.md:162 changes rendered prompts, but dry.prompt, pure.prompt, plan-quality.prompt, and milestone-review.prompt still contain the base registry. TestBuildPrompt_Golden fails; regenerate and inspect these snapshots, then rerun the judge tests. ARCH-PURPOSE.
- **BR-3** [Important] `architecture-map-consistency` Correct the atlas explanation of ARCH-ORDER review ordering.
  atlas/workflow/architecture-principles.md:80-84 describes oracle-first ordering as deliberate, while architecture.md:192 now leads with production routing and bypass prevention. Update the map-level explanation to match the approved policy without copying its clauses. ARCH-DRY.

## Round 3 — 2026-09-14T15:43:57-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The issue's verification revision names BuildPrompt, runArchPrinciples, and runStartPlan and describes full-registry assertions plus CLI comparison; inspected tests support this description.
- BR-2 — addressed — All four snapshots exactly replace the base registry with the pinned head registry. The complete judge suite, including TestBuildPrompt_Golden, passes; reversing the replacement breaks each snapshot equality.
- BR-3 — addressed — atlas/workflow/architecture-principles.md:80 now describes executable-model checks preceding the retained oracle critique, matching cmd/sdlc/internal/judge/architecture.md:192.

## Open findings

(none — every finding has been disposed)
