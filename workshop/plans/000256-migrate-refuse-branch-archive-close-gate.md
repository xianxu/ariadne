---
gate: boundary-review
issue: 256
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-27T23:10:24-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Post-cutover branch reconcile still treats any deleted details file as the branch's own archive move
          detail: 'cmd/sdlc/issuemigrate.go:569-571 skips a removed details file ("its own archive move") without checking the imported card. A branch the dry run never saw (another clone, never pushed) that archived a still-active issue reconciles silently, and the card stays open. This is the #256 scenario after cutover. Family sweep: this is the only remaining instance (grep "own archive move"). Fix: refuse unless the card is done or the branch keeps other active details for the ID. Test both outcomes.'
          family: branch-deletion-assumed-own-archive
          round: 1
      recipe: small-diff-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-27T23:13:29-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: issuemigrate.go:579-585 now refuses via shared tracker.RemovalArchivesActive (kept + terminal card); TestIssueMigrateReconcileRefusesAnUnlandedArchive pins the refusal; grep "own archive move" finds no other code site.
          round: 2
      findings:
        - id: BR-2
          severity: Minor
          title: Reconcile's closed-card silent path (cardClosed true) has no test
          detail: 'The reconcile test''s only silent case is a rename (kept). Stubbing cardClosed to false would stay green. Add a case where #1''s card is done and assert that reconcile does not refuse with ArchivesActiveReason.'
          family: test-both-outcomes
          round: 2
        - id: BR-3
          severity: Minor
          title: cardClosed maps an unparseable card to open, so the refusal wrongly says the issue is still open
          detail: issuemigrate.go:680-686 ignores the ParseCard error. This fails safe, but the message is inaccurate. Consider surfacing the parse error as its own refusal.
          family: parse-error-coerced-to-state
          round: 2
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#256 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-27T23:10:24-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `branch-deletion-assumed-own-archive` Post-cutover branch reconcile still treats any deleted details file as the branch's own archive move
  cmd/sdlc/issuemigrate.go:569-571 skips a removed details file ("its own archive move") without checking the imported card. A branch the dry run never saw (another clone, never pushed) that archived a still-active issue reconciles silently, and the card stays open. This is the #256 scenario after cutover. Family sweep: this is the only remaining instance (grep "own archive move"). Fix: refuse unless the card is done or the branch keeps other active details for the ID. Test both outcomes.

## Round 2 — 2026-09-27T23:13:29-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — issuemigrate.go:579-585 now refuses via shared tracker.RemovalArchivesActive (kept + terminal card); TestIssueMigrateReconcileRefusesAnUnlandedArchive pins the refusal; grep "own archive move" finds no other code site.

### Raised

- **BR-2** [Minor] `test-both-outcomes` Reconcile's closed-card silent path (cardClosed true) has no test
  The reconcile test's only silent case is a rename (kept). Stubbing cardClosed to false would stay green. Add a case where #1's card is done and assert that reconcile does not refuse with ArchivesActiveReason.
- **BR-3** [Minor] `parse-error-coerced-to-state` cardClosed maps an unparseable card to open, so the refusal wrongly says the issue is still open
  issuemigrate.go:680-686 ignores the ParseCard error. This fails safe, but the message is inaccurate. Consider surfacing the parse error as its own refusal.

## Open findings

- **BR-2** [Minor] `test-both-outcomes` Reconcile's closed-card silent path (cardClosed true) has no test
- **BR-3** [Minor] `parse-error-coerced-to-state` cardClosed maps an unparseable card to open, so the refusal wrongly says the issue is still open
