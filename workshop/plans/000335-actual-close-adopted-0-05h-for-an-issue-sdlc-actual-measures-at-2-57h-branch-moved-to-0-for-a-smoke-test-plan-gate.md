---
gate: plan-quality
issue: 335
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T02:01:01-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Define how pinned measurement inputs survive close and landing
          detail: 'ARCH-ORDER: The snapshot at plan line 52 has no stated persistence or completed-issue recovery contract beyond completion.reviewed_head and the ledger. Specify where the pre-close tracker ref, scope, effective endpoints and source evidence binding are retained, how later actual resolves them, and how legacy missing pins degrade; tracker close/correction commits must not silently change the measurement window.'
          family: durable-measurement-inputs
          round: 1
        - id: PQ-2
          severity: Important
          title: Specify correction states and idempotence across tracker and calibration effects
          detail: 'ARCH-ORDER: Plan line 66 delegates recovery to existing receipts, whose operation stages do not include calibration append. Name the pure correction transition component, effect order, durable operation identity and acknowledgment evidence, including restart after tracker success, uncertain append, generation change and competing corrections; state how retries distinguish their own confirmed writes from another generation.'
          family: multi-effect-recovery-contract
          round: 1
        - id: PQ-3
          severity: Important
          title: Define correction and exclusion records before sweeping calibration readers
          detail: 'ARCH-DRY and ARCH-PURPOSE: NewestPerIssue is pre-filtered last-wins selection, not an authoritative invalidation mechanism. Specify the correction/exclusion representation, generation binding and shared resolution rule applied before eligibility filtering, preserving ordinary re-close semantics; name the owning exclusion action promised when historical repair is impossible.'
          family: authoritative-calibration-supersession
          round: 1
        - id: PQ-4
          severity: Important
          title: Name the risky functions and compress test-case prose into strategies
          detail: The table and strategy paragraph name concepts rather than unit-tested functions. Provide proposed function identifiers for engagement derivation, eligibility, clipping, plausibility and correction transitions, each with its adversarial input class and mechanical guard; compress the enumerated correction cases into that strategy.
          family: named-function-test-strategies
          round: 1
        - id: PQ-5
          severity: Minor
          title: State correction-record retention and permanent growth cost
          detail: 'ARCH-FUNERAL: Existing receipt cleanup does not explain the lifetime of correction evidence in tracker history and appended calibration rows. State the cleanup owner for transient receipts and the bounded payload and permanent cost per correction for retained records.'
          family: durable-artifact-lifecycle
          round: 1
        - id: PQ-6
          severity: Minor
          title: Give historical conformance replay an execution cadence
          detail: 'ARCH-MOCK: Historical opt-in replay is named but no cadence is stated. Specify when it runs to detect Claude/Codex lineage metadata drift rather than relying solely on synthetic fixtures.'
          family: external-conformance-cadence
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-14T02:06:21-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Versioned boundary snapshots retain refs, scope, endpoints, split inputs and source bindings; completed replay and legacy degradation are specified.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: actualcorrection.Transition defines effect order, durable identity, exact acknowledgment evidence, recovery and competing-generation refusal.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Versioned generation-bound correction records resolve exact targets before eligibility filtering; an owning exclusion action is defined.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Named risky functions each have an adversarial input class and mechanical guard; executable tests own the case inventory.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Existing recovery owns transient receipt cleanup; retained snapshots and correction evidence have explicit bounded permanent costs.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Historical conformance runs before issue close, after lineage parser changes and when installed CLI metadata changes.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-14T02:08:22-07:00"
      agent: codex
      blocked: false
content_hash: c468938cbd43711306990b5167fa72cdfc8bbfbad6ca3da12eb38db8cb500e73
---

# Gate ledger — ariadne#335 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T02:01:01-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `durable-measurement-inputs` Define how pinned measurement inputs survive close and landing
  ARCH-ORDER: The snapshot at plan line 52 has no stated persistence or completed-issue recovery contract beyond completion.reviewed_head and the ledger. Specify where the pre-close tracker ref, scope, effective endpoints and source evidence binding are retained, how later actual resolves them, and how legacy missing pins degrade; tracker close/correction commits must not silently change the measurement window.
- **PQ-2** [Important] `multi-effect-recovery-contract` Specify correction states and idempotence across tracker and calibration effects
  ARCH-ORDER: Plan line 66 delegates recovery to existing receipts, whose operation stages do not include calibration append. Name the pure correction transition component, effect order, durable operation identity and acknowledgment evidence, including restart after tracker success, uncertain append, generation change and competing corrections; state how retries distinguish their own confirmed writes from another generation.
- **PQ-3** [Important] `authoritative-calibration-supersession` Define correction and exclusion records before sweeping calibration readers
  ARCH-DRY and ARCH-PURPOSE: NewestPerIssue is pre-filtered last-wins selection, not an authoritative invalidation mechanism. Specify the correction/exclusion representation, generation binding and shared resolution rule applied before eligibility filtering, preserving ordinary re-close semantics; name the owning exclusion action promised when historical repair is impossible.
- **PQ-4** [Important] `named-function-test-strategies` Name the risky functions and compress test-case prose into strategies
  The table and strategy paragraph name concepts rather than unit-tested functions. Provide proposed function identifiers for engagement derivation, eligibility, clipping, plausibility and correction transitions, each with its adversarial input class and mechanical guard; compress the enumerated correction cases into that strategy.
- **PQ-5** [Minor] `durable-artifact-lifecycle` State correction-record retention and permanent growth cost
  ARCH-FUNERAL: Existing receipt cleanup does not explain the lifetime of correction evidence in tracker history and appended calibration rows. State the cleanup owner for transient receipts and the bounded payload and permanent cost per correction for retained records.
- **PQ-6** [Minor] `external-conformance-cadence` Give historical conformance replay an execution cadence
  ARCH-MOCK: Historical opt-in replay is named but no cadence is stated. Specify when it runs to detect Claude/Codex lineage metadata drift rather than relying solely on synthetic fixtures.

## Round 2 — 2026-10-14T02:06:21-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Versioned boundary snapshots retain refs, scope, endpoints, split inputs and source bindings; completed replay and legacy degradation are specified.
- PQ-2 — addressed — actualcorrection.Transition defines effect order, durable identity, exact acknowledgment evidence, recovery and competing-generation refusal.
- PQ-3 — addressed — Versioned generation-bound correction records resolve exact targets before eligibility filtering; an owning exclusion action is defined.
- PQ-4 — addressed — Named risky functions each have an adversarial input class and mechanical guard; executable tests own the case inventory.
- PQ-5 — addressed — Existing recovery owns transient receipt cleanup; retained snapshots and correction evidence have explicit bounded permanent costs.
- PQ-6 — addressed — Historical conformance runs before issue close, after lineage parser changes and when installed CLI metadata changes.

## Round 3 — 2026-10-14T02:08:22-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
