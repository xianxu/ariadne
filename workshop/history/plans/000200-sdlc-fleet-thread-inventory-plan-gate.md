---
gate: plan-quality
issue: 200
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-08-25T15:01:10-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Compress enumerated test cases into function-level risk strategies
          detail: The plan repeatedly lists individual test cases and procedural diff steps instead of naming each risky function with one adversarial input class and mechanical guard. Rewrite the test contract around functions such as LoadPolicy, ResolvePolicy, gitx.ParseWorktrees, NormalizeVantage, AssociateBranchIssue, CollectFacts, Inventory, and Pair's admission function; use fuzzing, properties, shared fixture corpora, or stateful conformance as the strategy rather than prose case inventories.
          family: executable-test-strategy
          round: 1
        - id: PQ-2
          severity: Important
          title: Resolve the contradictory inventory-to-policy resolver boundary
          detail: 'The plan says inventory cannot resolve an admission key without a prospective path, but also says inventory rows call ResolvePolicy; Task 4.2 says only the query resolves. Under ARCH-PURE and ARCH-DRY, state one executable boundary: inventory shares declaration loading and validation, while only fleet policy with a requested path invokes ResolvePolicy.'
          family: resolver-boundary-ownership
          round: 1
        - id: PQ-3
          severity: Important
          title: Specify a non-circular Pair and Ariadne gate sequence
          detail: pair#149 declares ariadne#200 as a dependency, yet Chunk 6 leaves its pre-#200 boundary as an ambiguous milestone/close. Identify the concrete Pair milestone that lands and reviews the consumer slice before ariadne#200 closes, and state that pair#149 issue-close follows its dependency, or revise the declared dependency and ownership.
          family: cross-issue-gate-ordering
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-08-25T15:12:29-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: not-addressed
          note: The function-level strategy table was added, but task sections still enumerate individual test scenarios and fixtures instead of referring compactly to those strategies.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Inventory now shares declaration loading and validation only; the prospective-path query alone invokes ResolvePolicy.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Pair M1 is the concrete reviewed consumer boundary before ariadne#200 closes, while pair#149 issue-close follows later and its frontmatter has no circular dependency.
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-08-25T15:14:23-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The authoritative function-level strategy table now covers each risky function with fuzzing, properties, shared corpora, or stateful conformance, and task sections refer back to it rather than maintaining independent prose test inventories.
          round: 3
      blocked: false
content_hash: 92f8d0a17a103340eb3187a381bd88faf8eeaedd2da173c9cda831f2b383980b
---

# Gate ledger — ariadne#200 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-25T15:01:10-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `executable-test-strategy` Compress enumerated test cases into function-level risk strategies
  The plan repeatedly lists individual test cases and procedural diff steps instead of naming each risky function with one adversarial input class and mechanical guard. Rewrite the test contract around functions such as LoadPolicy, ResolvePolicy, gitx.ParseWorktrees, NormalizeVantage, AssociateBranchIssue, CollectFacts, Inventory, and Pair's admission function; use fuzzing, properties, shared fixture corpora, or stateful conformance as the strategy rather than prose case inventories.
- **PQ-2** [Important] `resolver-boundary-ownership` Resolve the contradictory inventory-to-policy resolver boundary
  The plan says inventory cannot resolve an admission key without a prospective path, but also says inventory rows call ResolvePolicy; Task 4.2 says only the query resolves. Under ARCH-PURE and ARCH-DRY, state one executable boundary: inventory shares declaration loading and validation, while only fleet policy with a requested path invokes ResolvePolicy.
- **PQ-3** [Important] `cross-issue-gate-ordering` Specify a non-circular Pair and Ariadne gate sequence
  pair#149 declares ariadne#200 as a dependency, yet Chunk 6 leaves its pre-#200 boundary as an ambiguous milestone/close. Identify the concrete Pair milestone that lands and reviews the consumer slice before ariadne#200 closes, and state that pair#149 issue-close follows its dependency, or revise the declared dependency and ownership.

## Round 2 — 2026-08-25T15:12:29-07:00 (codex) — BLOCKED

### Disposed

- PQ-1 — not-addressed — The function-level strategy table was added, but task sections still enumerate individual test scenarios and fixtures instead of referring compactly to those strategies.
- PQ-2 — addressed — Inventory now shares declaration loading and validation only; the prospective-path query alone invokes ResolvePolicy.
- PQ-3 — addressed — Pair M1 is the concrete reviewed consumer boundary before ariadne#200 closes, while pair#149 issue-close follows later and its frontmatter has no circular dependency.

## Round 3 — 2026-08-25T15:14:23-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — The authoritative function-level strategy table now covers each risky function with fuzzing, properties, shared corpora, or stateful conformance, and task sections refer back to it rather than maintaining independent prose test inventories.

## Open findings

(none — every finding has been disposed)
