---
gate: plan-quality
issue: 377
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T20:44:02-07:00"
      agent: claude
      findings:
        - kind: defect
          id: PQ-1
          severity: Minor
          title: Done-when "help contains every rule" has no test; add G1-G8, G10, G11 phrases to TestLeadHelpCoversPlaybook
          detail: The only planned help test covers verb names. TestLeadHelpCoversPlaybook (cmd/tl-lead/main_test.go:29) is the existing per-topic phrase guard and should get one phrase per ported rule, so a later edit can't drop a rule silently. TestLeadHelpNamesSubcommands (main_test.go:118) is a hand-written list, so it must be extended with the instructions verbs and their command-tree paths.
          family: rule-port-unguarded-by-test
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: link's no-retry, hash-pinned write doesn't fit writeSection as it stands; say how the write path is parameterised
          detail: writeSection (cmd/tl-lead/status.go:33-71) always retries on ErrStale, and its compute callback never sees the read hash. The plan should name the small option (for example, an expected-hash, no-retry mode) instead of leaving room for a copied loop or a silent re-target.
          family: write-path-variant-unstated
          round: 1
      blocked: false
    - "n": 2
      timestamp: "2026-10-14T20:45:36-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Revisions add one phrase guard per rule and make the subcommand test walk the command tree.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: link is a single-attempt linkInstruction that shares writeBody and refusal, with stale mapped to exit 6 (status.go:71,75,91).
          round: 2
      blocked: false
content_hash: 6c979dbf0331ca5883d81993f75a88da885fcabe281022245b5913e843284218
---

# Gate ledger — ariadne#377 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T20:44:02-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `rule-port-unguarded-by-test` (kind: defect) Done-when "help contains every rule" has no test; add G1-G8, G10, G11 phrases to TestLeadHelpCoversPlaybook
  The only planned help test covers verb names. TestLeadHelpCoversPlaybook (cmd/tl-lead/main_test.go:29) is the existing per-topic phrase guard and should get one phrase per ported rule, so a later edit can't drop a rule silently. TestLeadHelpNamesSubcommands (main_test.go:118) is a hand-written list, so it must be extended with the instructions verbs and their command-tree paths.
- **PQ-2** [Minor] `write-path-variant-unstated` (kind: defect) link's no-retry, hash-pinned write doesn't fit writeSection as it stands; say how the write path is parameterised
  writeSection (cmd/tl-lead/status.go:33-71) always retries on ErrStale, and its compute callback never sees the read hash. The plan should name the small option (for example, an expected-hash, no-retry mode) instead of leaving room for a copied loop or a silent re-target.

## Round 2 — 2026-10-14T20:45:36-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Revisions add one phrase guard per rule and make the subcommand test walk the command tree.
- PQ-2 — addressed — link is a single-attempt linkInstruction that shares writeBody and refusal, with stale mapped to exit 6 (status.go:71,75,91).

## Open findings

(none — every finding has been disposed)
