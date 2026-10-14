---
gate: plan-quality
issue: 348
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-13T16:32:01-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: New sequence-valued card kind is not handled by sameCardValue or SetField (mirror drift and setter rewrite)
          detail: sameCardValue (cmd/sdlc/internal/issue/card.go:323) compares non-mapping nodes by .Value, which is "" for every SequenceNode, so mirror.go:79,99 and migrate.go:68 treat any two area lists as equal. SetField/dropBlockContinuation (frontmatter.go:73-99) do not drop column-one block-sequence items, so a block-form area card cannot be re-set. Task 3 must cover sequences in both helpers, with mirror/migrate round-trip tests.
          family: card-kind-assumes-scalar
          round: 1
        - id: PQ-2
          severity: Minor
          title: Tasks 1-6 enumerate test cases in prose instead of naming functions plus one strategy line per risky function
          detail: Compress to function names plus a strategy line each, e.g. fuzz repoenum.Parse over malformed JSON with a parse/render round-trip, and fuzz the repo-enum card shape over nested or mixed-kind YAML nodes.
          family: plan-enumerates-test-cases
          round: 1
        - id: PQ-3
          severity: Minor
          title: The Revisions entry asks for a check at issue publish; the plan neither adds one nor says why the gates suffice
          detail: 'In tracker repos, setters change only the card, so the main-diff #124 gate never sees those values. Only issue validate sweeps cards. State whether that is enough or add the publish check.'
          family: write-path-validation-coverage
          round: 1
        - id: PQ-4
          severity: Minor
          title: In a legacy repo without origin/main, issue new reads the declaration from HEAD (the branch), contradicting the stated trust rule
          family: declaration-trust-rev
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T16:34:01-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: R1 enumerates the whole scalar-assumption class (sameCardValue, SetField, IssueRecord.Field, the migrate switch, ParseCard) and fixes it with named tests; verified at card.go:323-335.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: R1 names test functions plus one strategy line per risky function (FuzzParse, FuzzRepoEnumCardShape, property tests).
          round: 2
        - id: PQ-3
          disposition: addressed
          note: A publish-time check is added in issuepublish.go with TestIssuePublish_RefusesUndeclaredArea.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: The declaration is read from origin/main, then local main, else refused; never HEAD (TestLoadRepoEnums_NeverReadsHead).
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-13T16:35:42-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: 7b9e6a632a7ee6060a453e6f489d0d13ab4edbbe846cc8012a85a0e33341ea2b
---

# Gate ledger — ariadne#348 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T16:32:01-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `card-kind-assumes-scalar` New sequence-valued card kind is not handled by sameCardValue or SetField (mirror drift and setter rewrite)
  sameCardValue (cmd/sdlc/internal/issue/card.go:323) compares non-mapping nodes by .Value, which is "" for every SequenceNode, so mirror.go:79,99 and migrate.go:68 treat any two area lists as equal. SetField/dropBlockContinuation (frontmatter.go:73-99) do not drop column-one block-sequence items, so a block-form area card cannot be re-set. Task 3 must cover sequences in both helpers, with mirror/migrate round-trip tests.
- **PQ-2** [Minor] `plan-enumerates-test-cases` Tasks 1-6 enumerate test cases in prose instead of naming functions plus one strategy line per risky function
  Compress to function names plus a strategy line each, e.g. fuzz repoenum.Parse over malformed JSON with a parse/render round-trip, and fuzz the repo-enum card shape over nested or mixed-kind YAML nodes.
- **PQ-3** [Minor] `write-path-validation-coverage` The Revisions entry asks for a check at issue publish; the plan neither adds one nor says why the gates suffice
  In tracker repos, setters change only the card, so the main-diff #124 gate never sees those values. Only issue validate sweeps cards. State whether that is enough or add the publish check.
- **PQ-4** [Minor] `declaration-trust-rev` In a legacy repo without origin/main, issue new reads the declaration from HEAD (the branch), contradicting the stated trust rule

## Round 2 — 2026-10-13T16:34:01-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — R1 enumerates the whole scalar-assumption class (sameCardValue, SetField, IssueRecord.Field, the migrate switch, ParseCard) and fixes it with named tests; verified at card.go:323-335.
- PQ-2 — addressed — R1 names test functions plus one strategy line per risky function (FuzzParse, FuzzRepoEnumCardShape, property tests).
- PQ-3 — addressed — A publish-time check is added in issuepublish.go with TestIssuePublish_RefusesUndeclaredArea.
- PQ-4 — addressed — The declaration is read from origin/main, then local main, else refused; never HEAD (TestLoadRepoEnums_NeverReadsHead).

## Round 3 — 2026-10-13T16:35:42-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
