---
gate: boundary-review
issue: 285
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-08T13:58:45-07:00"
      agent: claude
      recipe: milestone-review
      blocked: false
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-10-08T14:02:49-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: issue restore also reverts verdictBehind entries, discarding an owner's own edit
          detail: issuerestore.go selects changedDetails by ID only; filter to verdictNotOwner or refuse in the owner's checkout with the merge-main action.
          family: verb-scope-wider-than-advertised
          round: 2
        - id: BR-2
          severity: Minor
          title: atlas issue-tracker.md transfer-guard paragraph runs into the unrelated recovery sentence
          family: atlas-prose-structure
          round: 2
        - id: BR-3
          severity: Minor
          title: a reopen Log edit to published details now needs a claim; note it in set-status reopen help
          family: behavior-change-undocumented-in-help
          round: 2
      recipe: milestone-review
      blocked: false
    - "n": 3
      timestamp: "2026-10-08T14:06:32-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: issuerestore.go:76-82 refuses on verdictBehind before any write; TestIssueRestoreRefusesTheOwnersOwnEdit fails with the loop removed (verified in scratch worktree) and passes with it.
          round: 3
        - id: BR-2
          disposition: addressed
          note: atlas/workflow/issue-tracker.md now ends the Transfer guard paragraph and starts the recovery sentence as its own paragraph.
          round: 3
        - id: BR-3
          disposition: addressed
          note: helptext/set-status.md reopen guard now says the Log entry lands only from the owner's checkout, or claim first; this matches detailsVerdict's NotOwner branch.
          round: 3
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#285 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-08T13:58:45-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-10-08T14:02:49-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `verb-scope-wider-than-advertised` issue restore also reverts verdictBehind entries, discarding an owner's own edit
  issuerestore.go selects changedDetails by ID only; filter to verdictNotOwner or refuse in the owner's checkout with the merge-main action.
- **BR-2** [Minor] `atlas-prose-structure` atlas issue-tracker.md transfer-guard paragraph runs into the unrelated recovery sentence
- **BR-3** [Minor] `behavior-change-undocumented-in-help` a reopen Log edit to published details now needs a claim; note it in set-status reopen help

## Round 3 — 2026-10-08T14:06:32-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — issuerestore.go:76-82 refuses on verdictBehind before any write; TestIssueRestoreRefusesTheOwnersOwnEdit fails with the loop removed (verified in scratch worktree) and passes with it.
- BR-2 — addressed — atlas/workflow/issue-tracker.md now ends the Transfer guard paragraph and starts the recovery sentence as its own paragraph.
- BR-3 — addressed — helptext/set-status.md reopen guard now says the Log entry lands only from the owner's checkout, or claim first; this matches detailsVerdict's NotOwner branch.

## Open findings

(none — every finding has been disposed)
