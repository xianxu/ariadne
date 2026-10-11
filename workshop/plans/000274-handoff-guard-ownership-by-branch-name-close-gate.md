---
gate: boundary-review
issue: 274
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T10:24:01-07:00"
      dispatched: "2026-10-11T10:03:28-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: '`sdlc issue restore --help` still recommends `sdlc claim` and describes the old non-owner policy'
          detail: 'cmd/sdlc/issuerestore.go:25-33 says restore is offered from a checkout that does not own the issue, says to claim to keep the edit (claim cannot take a foreign-held issue, which is now the only case restore is offered), and says it refuses only in the owner''s own checkout, but it now also refuses unclaimed+behind (issuerestore.go:77). Family instances in window: this help text, plus the abandon.go:250 comment (Minor). Rewrite to match the detailsRefusal wording.'
          family: doc-claims-track-policy
          round: 1
        - id: BR-2
          severity: Minor
          title: archiveAbandoned doc comment still says "Only its owner publishes."
          detail: cmd/sdlc/abandon.go:250; its Allow (requireOwnedToPublish) now also accepts an unclaimed card.
          family: doc-claims-track-policy
          round: 1
        - id: BR-3
          severity: Minor
          title: restore refuses to drop an unclaimed issue's stale copy (it used to be offered); there is no test for it
          detail: issuerestore.go:77 treats unclaimed+behind like the owner's own work. This matches the plan, but a user who explicitly asks to drop an unclaimed stale copy is refused toward merging main; consider allowing it, or pin the refusal with a test.
          family: restore-scope
          round: 1
      recipe: small-diff-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#274 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T10:24:01-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `doc-claims-track-policy` `sdlc issue restore --help` still recommends `sdlc claim` and describes the old non-owner policy
  cmd/sdlc/issuerestore.go:25-33 says restore is offered from a checkout that does not own the issue, says to claim to keep the edit (claim cannot take a foreign-held issue, which is now the only case restore is offered), and says it refuses only in the owner's own checkout, but it now also refuses unclaimed+behind (issuerestore.go:77). Family instances in window: this help text, plus the abandon.go:250 comment (Minor). Rewrite to match the detailsRefusal wording.
- **BR-2** [Minor] `doc-claims-track-policy` archiveAbandoned doc comment still says "Only its owner publishes."
  cmd/sdlc/abandon.go:250; its Allow (requireOwnedToPublish) now also accepts an unclaimed card.
- **BR-3** [Minor] `restore-scope` restore refuses to drop an unclaimed issue's stale copy (it used to be offered); there is no test for it
  issuerestore.go:77 treats unclaimed+behind like the owner's own work. This matches the plan, but a user who explicitly asks to drop an unclaimed stale copy is refused toward merging main; consider allowing it, or pin the refusal with a test.

## Open findings

- **BR-1** [Important] `doc-claims-track-policy` `sdlc issue restore --help` still recommends `sdlc claim` and describes the old non-owner policy
- **BR-2** [Minor] `doc-claims-track-policy` archiveAbandoned doc comment still says "Only its owner publishes."
- **BR-3** [Minor] `restore-scope` restore refuses to drop an unclaimed issue's stale copy (it used to be offered); there is no test for it
