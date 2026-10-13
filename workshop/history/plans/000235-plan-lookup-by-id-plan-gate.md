---
gate: plan-quality
issue: 235
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-12T17:24:57-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Minor
          title: Plan doesn't say how planArtifactPath derives the id from the stem, or what happens with a stem that has no NNNNNN- prefix
          detail: planArtifactPath(plansDir, name) receives an issue stem (changecode.go:585). State the parse, and fall back to the exact canonical name when there is no numeric prefix.
          family: unstated-input-derivation
          round: 1
        - id: PQ-2
          severity: Minor
          title: Ambiguity refusal has no path through readers that return only a string or nothing (readOptionalPlanFile, recordChangeCodeFlow)
          detail: recordChangeCodeFlow returns nothing (changecode_flow.go:99), and closeflow.go:80 reads a bare string. Say whether these refuse earlier or change signature, so ambiguity never quietly becomes "no plan".
          family: refusal-path-unrouted
          round: 1
        - id: PQ-3
          severity: Minor
          title: Tests listed as prose cases; name pickPlanArtifact plus one adversarial strategy line instead
          detail: Generate arbitrary NNNNNN-* name sets (sidecars, same-prefix longer ids, multiple plans) and guard that the result is never a non -plan.md file and never a silent pick.
          family: test-strategy-not-function-named
          round: 1
        - id: PQ-4
          severity: Minor
          title: No non-goals stated (archive/history lookup, auto-rename); helptext change-code.md:94 still says the exact name
          family: no-stated-non-goals
          round: 1
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T17:26:02-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The id comes from the stem prefix via issueIDFromPath (issuefiles.go:155); a stem with no prefix keeps the exact canonical name.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: readOptionalPlanFile becomes (string, error); recordChangeCodeFlow and the closeflow.go:80 measure die on the error.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Names pickPlanArtifact with one adversarial strategy line and guard, plus two integration tests.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Non-goals stated (no archive lookup, no auto-rename); the helptext/change-code.md update is in the Plan.
          round: 2
      blocked: false
content_hash: d5c2318464ccb55cd06b761dc3609bf3928dfcbc11c6fe93d0613d936cac1fb6
---

# Gate ledger — ariadne#235 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T17:24:57-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `unstated-input-derivation` Plan doesn't say how planArtifactPath derives the id from the stem, or what happens with a stem that has no NNNNNN- prefix
  planArtifactPath(plansDir, name) receives an issue stem (changecode.go:585). State the parse, and fall back to the exact canonical name when there is no numeric prefix.
- **PQ-2** [Minor] `refusal-path-unrouted` Ambiguity refusal has no path through readers that return only a string or nothing (readOptionalPlanFile, recordChangeCodeFlow)
  recordChangeCodeFlow returns nothing (changecode_flow.go:99), and closeflow.go:80 reads a bare string. Say whether these refuse earlier or change signature, so ambiguity never quietly becomes "no plan".
- **PQ-3** [Minor] `test-strategy-not-function-named` Tests listed as prose cases; name pickPlanArtifact plus one adversarial strategy line instead
  Generate arbitrary NNNNNN-* name sets (sidecars, same-prefix longer ids, multiple plans) and guard that the result is never a non -plan.md file and never a silent pick.
- **PQ-4** [Minor] `no-stated-non-goals` No non-goals stated (archive/history lookup, auto-rename); helptext change-code.md:94 still says the exact name

## Round 2 — 2026-10-12T17:26:02-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — The id comes from the stem prefix via issueIDFromPath (issuefiles.go:155); a stem with no prefix keeps the exact canonical name.
- PQ-2 — addressed — readOptionalPlanFile becomes (string, error); recordChangeCodeFlow and the closeflow.go:80 measure die on the error.
- PQ-3 — addressed — Names pickPlanArtifact with one adversarial strategy line and guard, plus two integration tests.
- PQ-4 — addressed — Non-goals stated (no archive lookup, no auto-rename); the helptext/change-code.md update is in the Plan.

## Open findings

(none — every finding has been disposed)
