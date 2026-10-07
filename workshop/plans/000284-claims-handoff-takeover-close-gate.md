---
gate: boundary-review
issue: 284
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-07T14:35:04-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
          detail: 'cmd/sdlc/claim.go refreshAfterClaim: surface the actual git error in the warning instead of returning silently or naming the wrong cause.'
          family: silent-error-in-io-glue
          round: 1
        - id: BR-2
          severity: Minor
          title: claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
          detail: 'ARCH-DRY: one ref-join helper parameterized by prefix/separator, and a cumulative validateReplacements in tracker/reader.go.'
          family: duplicated-id-rendering
          round: 1
        - id: BR-3
          severity: Minor
          title: tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
          family: api-trusts-caller-normalization
          round: 1
        - id: BR-4
          severity: Minor
          title: a started unowned card in a set is refused toward --adopt, which itself refuses sets
          detail: Message should say to claim that issue alone with --adopt.
          family: refusal-points-at-unusable-action
          round: 1
        - id: BR-5
          severity: Minor
          title: 'project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line'
          family: hand-recorded-actuals
          round: 1
        - id: BR-6
          severity: Minor
          title: no test drives a claim set through a lost publication response and a rerun
          family: test-gap-lost-response-set
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#284 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-07T14:35:04-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `silent-error-in-io-glue` refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
  cmd/sdlc/claim.go refreshAfterClaim: surface the actual git error in the warning instead of returning silently or naming the wrong cause.
- **BR-2** [Minor] `duplicated-id-rendering` claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
  ARCH-DRY: one ref-join helper parameterized by prefix/separator, and a cumulative validateReplacements in tracker/reader.go.
- **BR-3** [Minor] `api-trusts-caller-normalization` tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
- **BR-4** [Minor] `refusal-points-at-unusable-action` a started unowned card in a set is refused toward --adopt, which itself refuses sets
  Message should say to claim that issue alone with --adopt.
- **BR-5** [Minor] `hand-recorded-actuals` project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line
- **BR-6** [Minor] `test-gap-lost-response-set` no test drives a claim set through a lost publication response and a rerun

## Open findings

- **BR-1** [Minor] `silent-error-in-io-glue` refreshAfterClaim swallows rev-parse errors and misreports merge-base failures as "commits main lacks"
- **BR-2** [Minor] `duplicated-id-rendering` claimRefs, claimArg and cardsMessage each re-implement the CLIRef join; ChangeCards inlines validateReplacement's budget math
- **BR-3** [Minor] `api-trusts-caller-normalization` tracker.ChangeCards does not dedupe ids; relies on callers (claimIssues) to do so
- **BR-4** [Minor] `refusal-points-at-unusable-action` a started unowned card in a set is refused toward --adopt, which itself refuses sets
- **BR-5** [Minor] `hand-recorded-actuals` project file ticks M1 with actual/closed before milestone-close measures it; M lines not nested under the #284 line
- **BR-6** [Minor] `test-gap-lost-response-set` no test drives a claim set through a lost publication response and a rerun
