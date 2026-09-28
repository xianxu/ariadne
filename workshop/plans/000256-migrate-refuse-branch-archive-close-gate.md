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
---

# Gate ledger — ariadne#256 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-27T23:10:24-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `branch-deletion-assumed-own-archive` Post-cutover branch reconcile still treats any deleted details file as the branch's own archive move
  cmd/sdlc/issuemigrate.go:569-571 skips a removed details file ("its own archive move") without checking the imported card. A branch the dry run never saw (another clone, never pushed) that archived a still-active issue reconciles silently, and the card stays open. This is the #256 scenario after cutover. Family sweep: this is the only remaining instance (grep "own archive move"). Fix: refuse unless the card is done or the branch keeps other active details for the ID. Test both outcomes.

## Open findings

- **BR-1** [Important] `branch-deletion-assumed-own-archive` Post-cutover branch reconcile still treats any deleted details file as the branch's own archive move
