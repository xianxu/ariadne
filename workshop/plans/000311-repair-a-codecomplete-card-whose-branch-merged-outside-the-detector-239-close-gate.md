---
gate: boundary-review
issue: 311
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T18:38:56-07:00"
      dispatched: "2026-10-12T18:34:36-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: citesIssue re-implements gitx.subjectAnchorRE, the declared single source for the subject-opens-with-#N anchor
          detail: 'externalmerge.go:366 restates the anchor regex (adding zero-padding) instead of reusing internal/gitx/window.go:264, and skips the bookkeeping-subject filter ShippedWorkOnMain applies, so a range of only `#N: plan:` commits passes. Export a gitx anchor helper and call it.'
          family: single-source-helper-bypassed
          round: 1
        - id: BR-2
          severity: Important
          title: Done-when item 1 (#239's card reads done, state clean) is not met in the window; moved post-landing only in the Log
          detail: Either run `reconcile --issue 239 --landed b066449c` before close or add a Revisions entry amending Done-when, so the close does not attest an outcome that has not happened.
          family: done-when-deferred-unrecorded
          round: 1
        - id: BR-3
          severity: Minor
          title: doneUnboundCard checks the binding before status, so a bound done card errors instead of being a no-op
          family: noop-precedence
          round: 1
        - id: BR-4
          severity: Minor
          title: repairLanded maps any rev-parse error to "no such commit here", hiding git/IO failures
          family: error-cause-masked
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#311 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T18:38:56-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `single-source-helper-bypassed` citesIssue re-implements gitx.subjectAnchorRE, the declared single source for the subject-opens-with-#N anchor
  externalmerge.go:366 restates the anchor regex (adding zero-padding) instead of reusing internal/gitx/window.go:264, and skips the bookkeeping-subject filter ShippedWorkOnMain applies, so a range of only `#N: plan:` commits passes. Export a gitx anchor helper and call it.
- **BR-2** [Important] `done-when-deferred-unrecorded` Done-when item 1 (#239's card reads done, state clean) is not met in the window; moved post-landing only in the Log
  Either run `reconcile --issue 239 --landed b066449c` before close or add a Revisions entry amending Done-when, so the close does not attest an outcome that has not happened.
- **BR-3** [Minor] `noop-precedence` doneUnboundCard checks the binding before status, so a bound done card errors instead of being a no-op
- **BR-4** [Minor] `error-cause-masked` repairLanded maps any rev-parse error to "no such commit here", hiding git/IO failures

## Open findings

- **BR-1** [Important] `single-source-helper-bypassed` citesIssue re-implements gitx.subjectAnchorRE, the declared single source for the subject-opens-with-#N anchor
- **BR-2** [Important] `done-when-deferred-unrecorded` Done-when item 1 (#239's card reads done, state clean) is not met in the window; moved post-landing only in the Log
- **BR-3** [Minor] `noop-precedence` doneUnboundCard checks the binding before status, so a bound done card errors instead of being a no-op
- **BR-4** [Minor] `error-cause-masked` repairLanded maps any rev-parse error to "no such commit here", hiding git/IO failures
