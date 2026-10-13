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
    - "n": 2
      timestamp: "2026-10-12T18:44:05-07:00"
      dispatched: "2026-10-12T18:40:35-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'citesIssue now calls exported gitx.IsShippedWorkSubject (window.go:190), the anchor plus bookkeeping filter; tests pin the "#330: close" refusal and the "only bookkeeping" range refusal.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: A 2026-10-12 Revisions entry moves item 1 to after landing, with the exact repair command; item 2 is delivered in the window.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The status switch now runs before CardCompletion; TestDoneUnboundCardRefusesABoundCard asserts ErrNoChange for a bound done card, which would fail under the old order.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: The code fix (gitTest separates exit 1 from other errors) is correct, but no regression test drives a git failure other than exit 1 through repairLanded. Minor, non-blocking.
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: e71649e74406111e46462d76af9be657173b7cba
      blocked: false
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

## Round 2 — 2026-10-12T18:44:05-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — citesIssue now calls exported gitx.IsShippedWorkSubject (window.go:190), the anchor plus bookkeeping filter; tests pin the "#330: close" refusal and the "only bookkeeping" range refusal.
- BR-2 — addressed — A 2026-10-12 Revisions entry moves item 1 to after landing, with the exact repair command; item 2 is delivered in the window.
- BR-3 — addressed — The status switch now runs before CardCompletion; TestDoneUnboundCardRefusesABoundCard asserts ErrNoChange for a bound done card, which would fail under the old order.
- BR-4 — not-addressed — The code fix (gitTest separates exit 1 from other errors) is correct, but no regression test drives a git failure other than exit 1 through repairLanded. Minor, non-blocking.

## Open findings

- **BR-4** [Minor] `error-cause-masked` repairLanded maps any rev-parse error to "no such commit here", hiding git/IO failures
