---
gate: boundary-review
issue: 200
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-27T14:33:33-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Core concepts misclassify IO-dependent entities as PURE and name the wrong type location
          detail: ARCH-PURE requires promoting AssociateBranchIssue and writer-based renderers to INTEGRATION; the policy envelope types also live in types.go, not policy.go. Append a plan revision correcting both claims before close.
          family: core-concept-classification
          round: 1
        - id: BR-2
          severity: Important
          title: Policy diagnostic envelopes accept codes outside their promised closed variants
          detail: validatePolicyDiagnostic checks only non-empty code/message, so capability and prospective-result diagnostics accept arbitrary or wrong-surface codes. Add surface-specific allowed sets and regression tests.
          family: closed-result-algebra
          round: 1
        - id: BR-3
          severity: Important
          title: README omits the new sdlc fleet command surface
          detail: Add runnable inventory and policy examples, including --path, --json, and typed refusal behavior.
          family: user-facing-doc-sync
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-08-27T14:50:51-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The appended 2026-08-27 plan revision corrects the entity locations and PURE/INTEGRATION classifications without overwriting history.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Surface-specific validators now reject unknown and wrong-surface codes, with direct marshal and unmarshal regression tests.
          round: 2
        - id: BR-3
          disposition: addressed
          note: README now documents both commands, --path, --json, and typed nonzero refusals; TestREADME_DocumentsFleetQueries pins the section.
          round: 2
      findings:
        - id: BR-4
          severity: Critical
          title: The path-outside-repo result variant is unreachable through the production command
          detail: 'This is the 2nd finding in family closed-result-algebra. cmd/sdlc/fleet.go:148-168 canonicalizes the requested path and normalizes repository identity from that path''s containing directory, so an outside path either selects another repository or fails before ResolvePolicy. A live sdlc fleet policy --path /tmp --json invocation produced empty stdout and a raw not-a-git-repository error, while README.md:26-29 promises a typed diagnostic. Do not patch only this case: state the rule that every public PolicyResult diagnostic must be reachable through the production CLI, enumerate all four variants, and add real CLI coverage for each. Either provide repository context independently so path-outside-repo reaches ResolvePolicy, or remove that impossible variant and all associated claims.'
          family: closed-result-algebra
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-08-27T15:08:38-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: The appended revision describes the correction, but the Core concepts table remains incorrect and no test pins the corrected classifications.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Surface-specific validators and negative marshal/unmarshal tests enforce the closed capability and result diagnostic sets.
          round: 3
        - id: BR-3
          disposition: addressed
          note: README.md documents both fleet commands, and TestREADME_DocumentsFleetQueries pins the command and refusal surface.
          round: 3
        - id: BR-4
          disposition: addressed
          note: All three remaining variants are reached by built-process E2E tests; reverting the production fix makes the resolver and closed-envelope tests fail.
          round: 3
      blocked: true
    - "n": 4
      timestamp: "2026-08-27T15:22:15-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The authoritative plan revision corrects both classifications and locations, and TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory fails if those corrected rows are absent or changed.
          round: 4
      blocked: false
---

# Gate ledger — ariadne#200 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-27T14:33:33-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `core-concept-classification` Core concepts misclassify IO-dependent entities as PURE and name the wrong type location
  ARCH-PURE requires promoting AssociateBranchIssue and writer-based renderers to INTEGRATION; the policy envelope types also live in types.go, not policy.go. Append a plan revision correcting both claims before close.
- **BR-2** [Important] `closed-result-algebra` Policy diagnostic envelopes accept codes outside their promised closed variants
  validatePolicyDiagnostic checks only non-empty code/message, so capability and prospective-result diagnostics accept arbitrary or wrong-surface codes. Add surface-specific allowed sets and regression tests.
- **BR-3** [Important] `user-facing-doc-sync` README omits the new sdlc fleet command surface
  Add runnable inventory and policy examples, including --path, --json, and typed refusal behavior.

## Round 2 — 2026-08-27T14:50:51-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The appended 2026-08-27 plan revision corrects the entity locations and PURE/INTEGRATION classifications without overwriting history.
- BR-2 — addressed — Surface-specific validators now reject unknown and wrong-surface codes, with direct marshal and unmarshal regression tests.
- BR-3 — addressed — README now documents both commands, --path, --json, and typed nonzero refusals; TestREADME_DocumentsFleetQueries pins the section.

### Raised

- **BR-4** [Critical] `closed-result-algebra` The path-outside-repo result variant is unreachable through the production command
  This is the 2nd finding in family closed-result-algebra. cmd/sdlc/fleet.go:148-168 canonicalizes the requested path and normalizes repository identity from that path's containing directory, so an outside path either selects another repository or fails before ResolvePolicy. A live sdlc fleet policy --path /tmp --json invocation produced empty stdout and a raw not-a-git-repository error, while README.md:26-29 promises a typed diagnostic. Do not patch only this case: state the rule that every public PolicyResult diagnostic must be reachable through the production CLI, enumerate all four variants, and add real CLI coverage for each. Either provide repository context independently so path-outside-repo reaches ResolvePolicy, or remove that impossible variant and all associated claims.

## Round 3 — 2026-08-27T15:08:38-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — The appended revision describes the correction, but the Core concepts table remains incorrect and no test pins the corrected classifications.
- BR-2 — addressed — Surface-specific validators and negative marshal/unmarshal tests enforce the closed capability and result diagnostic sets.
- BR-3 — addressed — README.md documents both fleet commands, and TestREADME_DocumentsFleetQueries pins the command and refusal surface.
- BR-4 — addressed — All three remaining variants are reached by built-process E2E tests; reverting the production fix makes the resolver and closed-envelope tests fail.

## Round 4 — 2026-08-27T15:22:15-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The authoritative plan revision corrects both classifications and locations, and TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory fails if those corrected rows are absent or changed.

## Open findings

(none — every finding has been disposed)
