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
    - "n": 2
      timestamp: "2026-10-11T10:39:39-07:00"
      dispatched: "2026-10-11T10:33:32-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: issuerestore.go:26-34 help now names foreign/unclaimed-behind cases and "hand it to its owner"; no claim advice; catalog.go:295 updated too.
          round: 2
        - id: BR-2
          disposition: addressed
          note: abandon.go:250 now reads "Another workspace's card is refused (requireOwnedToPublish)", matching mayNotEditDetails.
          round: 2
        - id: BR-3
          disposition: addressed
          note: mayRestore (transferguard.go:180) allows unclaimed; TestIssueRestoreMakesAStaleFilingBranchLand/unclaimed fails when reverted (verified by mutation).
          round: 2
      findings:
        - id: BR-4
          severity: Important
          title: publish summaries still say only the owner publishes later edits
          detail: '3rd in family. Rule: every statement of who may publish/land published-details edits must state the mayNotEditDetails policy (refused only for another workspace''s issue); derive it from one shared help constant (like ownershipGateHelp/{{OWNERSHIP_GATE}}) plus a test rejecting owner-only wording. Remaining instances: helptext/issue.md:11, atlas/workflow/issue-sync.md:9, atlas/workflow/sdlc-binary.md:49 and :72, republish.go:56 (republishOwned "publishes owned issues''").'
          family: doc-claims-track-policy
          round: 2
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

## Round 2 — 2026-10-11T10:39:39-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — issuerestore.go:26-34 help now names foreign/unclaimed-behind cases and "hand it to its owner"; no claim advice; catalog.go:295 updated too.
- BR-2 — addressed — abandon.go:250 now reads "Another workspace's card is refused (requireOwnedToPublish)", matching mayNotEditDetails.
- BR-3 — addressed — mayRestore (transferguard.go:180) allows unclaimed; TestIssueRestoreMakesAStaleFilingBranchLand/unclaimed fails when reverted (verified by mutation).

### Raised

- **BR-4** [Important] `doc-claims-track-policy` publish summaries still say only the owner publishes later edits
  3rd in family. Rule: every statement of who may publish/land published-details edits must state the mayNotEditDetails policy (refused only for another workspace's issue); derive it from one shared help constant (like ownershipGateHelp/{{OWNERSHIP_GATE}}) plus a test rejecting owner-only wording. Remaining instances: helptext/issue.md:11, atlas/workflow/issue-sync.md:9, atlas/workflow/sdlc-binary.md:49 and :72, republish.go:56 (republishOwned "publishes owned issues'").

## Open findings

- **BR-4** [Important] `doc-claims-track-policy` publish summaries still say only the owner publishes later edits
