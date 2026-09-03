---
gate: boundary-review
issue: 211
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-02T18:59:55-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
          detail: |-
            2nd in family. Do not fix the instance: the rule is that the consumer set is
            enumerated in exactly one place, the grep-produced Spec table, and every other
            mention refers to it without restating a count. Measured prevalence 1 of 3
            references non-compliant (Plan row 3; Spec table and Estimate row 2 both comply).
            (carried from plan-quality PQ-5, deferred to the boundary review)
          family: consumer-enumeration-incomplete
          round: 1
        - id: BR-2
          severity: Minor
          title: Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
          detail: |-
            "Concatenating visited + skipped reproduces the input" is content-independent;
            "no file loses a real section" over 406 live workshop files needs an oracle, and
            a golden captured today reddens on the next issue that quotes a fence. State that
            the corpus supplies inputs while assertions stay invariants, and name how the test
            in cmd/sdlc/internal/issue resolves the repo-root path.
            (carried from plan-quality PQ-6, deferred to the boundary review)
          family: mutable-corpus-as-test-oracle
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-02T18:59:55-07:00"
      agent: claude
      boundary: M1
      blocked: false
      protocol_error: no valid findings block
---

# Gate ledger — ariadne#211 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T18:59:55-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `consumer-enumeration-incomplete` Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
  2nd in family. Do not fix the instance: the rule is that the consumer set is
  enumerated in exactly one place, the grep-produced Spec table, and every other
  mention refers to it without restating a count. Measured prevalence 1 of 3
  references non-compliant (Plan row 3; Spec table and Estimate row 2 both comply).
  (carried from plan-quality PQ-5, deferred to the boundary review)
- **BR-2** [Minor] `mutable-corpus-as-test-oracle` Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
  "Concatenating visited + skipped reproduces the input" is content-independent;
  "no file loses a real section" over 406 live workshop files needs an oracle, and
  a golden captured today reddens on the next issue that quotes a fence. State that
  the corpus supplies inputs while assertions stay invariants, and name how the test
  in cmd/sdlc/internal/issue resolves the repo-root path.
  (carried from plan-quality PQ-6, deferred to the boundary review)

## Round 2 — 2026-09-02T18:59:55-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

- **BR-1** [Minor] `consumer-enumeration-incomplete` Plan row 3 restates a stale consumer count ("four") that the Spec table has already superseded with six
- **BR-2** [Minor] `mutable-corpus-as-test-oracle` Corpus-seeded property test does not separate stable invariants from an oracle that goes stale as issues are authored
