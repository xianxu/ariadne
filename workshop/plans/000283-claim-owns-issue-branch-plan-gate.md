---
gate: plan-quality
issue: 283
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-02T12:09:57-07:00"
      agent: claude
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-10-02T12:11:36-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Minor
          title: adoptDecision still hard-codes working/blocked; no task routes it through CanHoldOwner
          detail: The core-concepts DRY rationale says the ownership block replaces the status list in adoptDecision (claimdecision.go:99), but no Task touches it. Sweep it with claimDecision/reclaimDecision.
          family: consumer-shadow-sweep
          round: 2
        - id: PQ-2
          severity: Minor
          title: Model and claim help will state claim never changes status, yet legacy repos still flip open→working
          detail: legacymode.go:399 keeps claim-as-flip; the shared issue.cue `start` edge (guard owned) and the Task 6 claim.md rewrite would misdescribe downstream legacy repos. State how the model/help describes the legacy path.
          family: single-source-scope-divergence
          round: 2
        - id: PQ-3
          severity: Minor
          title: Owned-open claims invisible to contention and state reads that count only working
          detail: startplan.go:476 and state.go:322 treat in-flight as status working; a shaping claim held by another slot is not surfaced. Decide intended or include.
          family: consumer-shadow-sweep
          round: 2
        - id: PQ-4
          severity: Minor
          title: Plan enumerates test cases in prose instead of one strategy line per risky function
          detail: Tasks 2–4 list cases bullet by bullet. Compress; name the adversarial class for planningDirtyAllowed (rename/copy entries in -z porcelain with two paths, one being the details file; quoted paths) and use a status×owner product for the decisions.
          family: test-prose-enumeration
          round: 2
        - id: PQ-5
          severity: Minor
          title: No explicit FUNERAL/SECURE/CONSTRAINTS statements; leftover branch on a lost start CAS unnamed
          detail: An unstarted claim persists with no expiry (cite the reclaim + claim-age decision); a concurrent reclaim between startDecision and UpdateCard leaves a created branch behind; say so.
          family: arch-envelope-unstated
          round: 2
        - id: PQ-6
          severity: Minor
          title: TransitionFor/FirstTransitionForEvent exist only on ProjectModel, not IssueModel
          detail: conformance_test.go:121 does not establish an IssueModel method; pkg/vocab/project.go:96,109 are the only definitions. Expect to add both (the plan's hedge covers it). Also completeop.go lives at cmd/sdlc/internal/tracker/.
          family: unbacked-code-claim
          round: 2
      blocked: false
content_hash: 71ea9e52db00475c2efca3f40fc827bffbaf3e0caa8f6b5c35646f18c7aef4df
---

# Gate ledger — ariadne#283 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T12:09:57-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-10-02T12:11:36-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `consumer-shadow-sweep` adoptDecision still hard-codes working/blocked; no task routes it through CanHoldOwner
  The core-concepts DRY rationale says the ownership block replaces the status list in adoptDecision (claimdecision.go:99), but no Task touches it. Sweep it with claimDecision/reclaimDecision.
- **PQ-2** [Minor] `single-source-scope-divergence` Model and claim help will state claim never changes status, yet legacy repos still flip open→working
  legacymode.go:399 keeps claim-as-flip; the shared issue.cue `start` edge (guard owned) and the Task 6 claim.md rewrite would misdescribe downstream legacy repos. State how the model/help describes the legacy path.
- **PQ-3** [Minor] `consumer-shadow-sweep` Owned-open claims invisible to contention and state reads that count only working
  startplan.go:476 and state.go:322 treat in-flight as status working; a shaping claim held by another slot is not surfaced. Decide intended or include.
- **PQ-4** [Minor] `test-prose-enumeration` Plan enumerates test cases in prose instead of one strategy line per risky function
  Tasks 2–4 list cases bullet by bullet. Compress; name the adversarial class for planningDirtyAllowed (rename/copy entries in -z porcelain with two paths, one being the details file; quoted paths) and use a status×owner product for the decisions.
- **PQ-5** [Minor] `arch-envelope-unstated` No explicit FUNERAL/SECURE/CONSTRAINTS statements; leftover branch on a lost start CAS unnamed
  An unstarted claim persists with no expiry (cite the reclaim + claim-age decision); a concurrent reclaim between startDecision and UpdateCard leaves a created branch behind; say so.
- **PQ-6** [Minor] `unbacked-code-claim` TransitionFor/FirstTransitionForEvent exist only on ProjectModel, not IssueModel
  conformance_test.go:121 does not establish an IssueModel method; pkg/vocab/project.go:96,109 are the only definitions. Expect to add both (the plan's hedge covers it). Also completeop.go lives at cmd/sdlc/internal/tracker/.

## Open findings

- **PQ-1** [Minor] `consumer-shadow-sweep` adoptDecision still hard-codes working/blocked; no task routes it through CanHoldOwner
- **PQ-2** [Minor] `single-source-scope-divergence` Model and claim help will state claim never changes status, yet legacy repos still flip open→working
- **PQ-3** [Minor] `consumer-shadow-sweep` Owned-open claims invisible to contention and state reads that count only working
- **PQ-4** [Minor] `test-prose-enumeration` Plan enumerates test cases in prose instead of one strategy line per risky function
- **PQ-5** [Minor] `arch-envelope-unstated` No explicit FUNERAL/SECURE/CONSTRAINTS statements; leftover branch on a lost start CAS unnamed
- **PQ-6** [Minor] `unbacked-code-claim` TransitionFor/FirstTransitionForEvent exist only on ProjectModel, not IssueModel
