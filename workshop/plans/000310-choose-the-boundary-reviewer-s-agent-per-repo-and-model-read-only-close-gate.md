---
gate: boundary-review
issue: 310
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T09:52:15-07:00"
      dispatched: "2026-10-11T09:45:16-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Help/atlas say every judge runs read-only, but claude's unrestricted Bash (and codex workspace-write) still let shell writes through
          detail: AllowedTools is Read,Grep,Glob,Bash under dontAsk, so any Bash command runs. For ContainNone judges (plan-quality, sdlc judge) that write reaches the operator tree. Either narrow Bash to read-only patterns or reword close.md:376-384 and sdlc-binary.md:996-1000.
          family: read-only-claim-exceeds-enforcement
          round: 1
        - id: BR-2
          severity: Minor
          title: The codex cannot-nest warning fires on any unfenced review, including Linux without bwrap where codex's sandbox works
          family: warning-scope-too-broad
          round: 1
        - id: BR-3
          severity: Minor
          title: Plan checkbox claims an atlas update to openshell-sandbox; the range does not touch that file
          family: plan-claims-undelivered-item
          round: 1
        - id: BR-4
          severity: Minor
          title: Validate's "reviewer_pairs is read only with" wording can be misread as read-only
          family: error-message-clarity
          round: 1
        - id: BR-5
          severity: Minor
          title: 'cmd/doc-review still passes claude bypassPermissions to its reviewer (outside #310''s scope)'
          family: read-only-claim-exceeds-enforcement
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#310 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T09:52:15-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `read-only-claim-exceeds-enforcement` Help/atlas say every judge runs read-only, but claude's unrestricted Bash (and codex workspace-write) still let shell writes through
  AllowedTools is Read,Grep,Glob,Bash under dontAsk, so any Bash command runs. For ContainNone judges (plan-quality, sdlc judge) that write reaches the operator tree. Either narrow Bash to read-only patterns or reword close.md:376-384 and sdlc-binary.md:996-1000.
- **BR-2** [Minor] `warning-scope-too-broad` The codex cannot-nest warning fires on any unfenced review, including Linux without bwrap where codex's sandbox works
- **BR-3** [Minor] `plan-claims-undelivered-item` Plan checkbox claims an atlas update to openshell-sandbox; the range does not touch that file
- **BR-4** [Minor] `error-message-clarity` Validate's "reviewer_pairs is read only with" wording can be misread as read-only
- **BR-5** [Minor] `read-only-claim-exceeds-enforcement` cmd/doc-review still passes claude bypassPermissions to its reviewer (outside #310's scope)

## Open findings

- **BR-1** [Important] `read-only-claim-exceeds-enforcement` Help/atlas say every judge runs read-only, but claude's unrestricted Bash (and codex workspace-write) still let shell writes through
- **BR-2** [Minor] `warning-scope-too-broad` The codex cannot-nest warning fires on any unfenced review, including Linux without bwrap where codex's sandbox works
- **BR-3** [Minor] `plan-claims-undelivered-item` Plan checkbox claims an atlas update to openshell-sandbox; the range does not touch that file
- **BR-4** [Minor] `error-message-clarity` Validate's "reviewer_pairs is read only with" wording can be misread as read-only
- **BR-5** [Minor] `read-only-claim-exceeds-enforcement` cmd/doc-review still passes claude bypassPermissions to its reviewer (outside #310's scope)
