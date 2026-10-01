---
gate: boundary-review
issue: 277
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-01T14:01:43-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Claim race test rejects the foreign-owner refusal a late-reading loser now gets
          detail: A loser whose snapshot follows the winner's publish hits ownedBy, which returns "is working, claimed by ...". That message contains neither "not open" nor "changed while claiming", so the test is green only when both clones read before either publishes. Accept "claimed by" (or add a deterministic in-process loser test), and update the claim.md loser wording to match.
          family: test-oracle-single-interleaving
          round: 1
        - id: BR-2
          severity: Important
          title: Plan promises MachineFingerprint(raw, key) plus a key test; code has MachineFingerprint(raw)
          detail: The Core concepts table and the M1 step describe a key parameter and a "different key gives a different value" test. The code uses a fixed domain prefix and tests separation between different raw IDs. The behaviour is fine; add a plan Revisions entry so the plan matches the code.
          family: plan-code-contract-drift
          round: 1
        - id: BR-3
          severity: Minor
          title: An owner's repeat claim with --dry-run still refreshes and writes the local details mirror
          family: dry-run-writes
          round: 1
        - id: BR-4
          severity: Minor
          title: M1 already writes claimant cards to the shared tracker, but the flag-day rollout note is deferred to M2
          detail: Any claim made with this branch's binary breaks stale fleet binaries. Add the atlas rollout paragraph now, or avoid claiming with this binary until the work lands.
          family: rollout-doc-timing
          round: 1
        - id: BR-5
          severity: Minor
          title: The race-test skip on an unreadable host machine ID promised by the plan is not implemented
          family: plan-code-contract-drift
          round: 1
        - id: BR-6
          severity: Minor
          title: slotLabel re-derives the slot directory layout in cmd/sdlc instead of asking pkg/workspace
          family: layout-knowledge-outside-owner
          round: 1
        - id: BR-7
          severity: Minor
          title: No claimDecision table test with a non-nil claimant; the Unknown (legacy working card) refusal is untested in M1
          family: missing-branch-unit-test
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-01T14:06:41-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Oracle now accepts "claimed by" (claimremote_test.go:182); help text updated; TestClaimDecisionOwnership pins the foreign-owner refusal deterministically.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Plan Revisions entry 2026-10-01 records the fixed-key MachineFingerprint(raw); domain-separation test added.
          round: 2
        - id: BR-3
          disposition: addressed
          note: claim.go:138 dry-run guard; TestClaimDryRunOwnerRepeatWritesNothing goes red when the guard is reverted (verified in scratch worktree).
          round: 2
        - id: BR-4
          disposition: addressed
          note: Rollout paragraph now in atlas/workflow/issue-tracker.md within M1.
          round: 2
        - id: BR-5
          disposition: addressed
          note: TestClaimRaceHasExactlyOneWinner skips when machineID() errors.
          round: 2
        - id: BR-6
          disposition: addressed
          note: slotLabel delegates to pkg/workspace Identity.UsesSlotLayout, which reuses slotNumber.
          round: 2
        - id: BR-7
          disposition: addressed
          note: TestClaimDecisionOwnership covers stamp, Mine, Foreign, Unknown (--adopt), and non-working statuses.
          round: 2
      findings:
        - id: BR-8
          severity: Minor
          title: TestSlotLabelOnlyWhereSlotsExist's "no slots" check runs after the malformed slot dir is created, duplicating the prior assertion
          detail: Move the plain-primary assertion before writing worktree/ariadne-slotX so the no-slot-dirs state is still exercised.
          family: stale-test-precondition
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 3
      timestamp: "2026-10-01T14:34:56-07:00"
      agent: claude
      findings:
        - id: BR-9
          severity: Important
          title: Relocation reassigns a working card on any same-machine worktree whose owner merely is not on the issue branch
          detail: relocatable/RelocationAllowed (cmd/sdlc/claimant.go:162) accept "recorded worktree does not hold the branch" as evidence of a move; that is also true right after claim (before start-plan) or when the owner switched away, so `git switch -c 000NNN-slug && sdlc claim` in another slot takes the card. Require positive move evidence (e.g. a local relocation marker written by sdlc move before its CAS) and add the negative test.
          family: absence-as-authority
          round: 3
        - id: BR-10
          severity: Important
          title: Verb-contract cells for adopt/set-status untested and drifting (no adopt race/table; set-status working on unattributed working card silently adopts)
          detail: '3rd finding in this family. Rule: the plan''s Verb contract table is the source; one table-driven test enumerates every situation x verb cell. Lift adopt''s decision out of IO adoptClaim (claim.go:561) into a pure function, make statusDecision refuse working->working on an unattributed card toward --adopt (setstatus.go:144), add the planned two-clone adopt race.'
          family: plan-code-contract-drift
          round: 3
        - id: BR-11
          severity: Important
          title: moveRelocation success path and Unknown branch never exercised; planned real move test not delivered
          detail: '2nd finding in this family. Rule: every return path of a new effectful function gets a fixture. TestRelocationAfterMoveAndRepair covers only the still-held refusal and a stubbed failure; the move.go:140 relocateClaimant success and move.go:131 Unknown return are unreached. Drive runMove (or unstubbed moveRelocation) to a success and an Unknown case.'
          family: missing-branch-unit-test
          round: 3
        - id: BR-12
          severity: Important
          title: README Concurrent issue work omits claim ownership, --adopt and the rebuild flag day
          detail: README.md:15-27 describes claim/start-plan in tracker repos; add one sentence on the owning workspace, --adopt for legacy cards, and rebuilding sdlc after landing.
          family: docs-surface-gap
          round: 3
        - id: BR-13
          severity: Minor
          title: cmd/sdlc/move.go import block is not gofmt-sorted
          family: gofmt-clean
          round: 3
        - id: BR-14
          severity: Minor
          title: Identical OWNERSHIP paragraph copied into four helptexts; use a shared template placeholder
          detail: ARCH-DRY; also the change-code.md paragraph splits "in any checkout:" from its numbered list.
          family: duplicated-help-prose
          round: 3
        - id: BR-15
          severity: Minor
          title: requireIssueOwnership discards relocatable's error and reports a plain Foreign refusal
          family: swallowed-probe-error
          round: 3
        - id: BR-16
          severity: Minor
          title: Tracker-era close opens the tracker and snapshots twice (ownership gate, then prepareTrackerClose)
          family: repeated-tracker-snapshot
          round: 3
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-10-01T14:56:04-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: RelocationAllowed + relocatable require sdlc move's record naming from/to; negative integration test plus pure "no move record" case; red when the record check is removed.
          round: 4
        - id: BR-10
          disposition: addressed
          note: adoptDecision is pure; TestVerbContractTable covers every cell (red with the working->working refusal disabled); TestAdoptRaceHasExactlyOneWinner added.
          round: 4
        - id: BR-11
          disposition: addressed
          note: TestMoveRelocatesItsOwner drives real runMove to success; TestMoveLeavesAnUnattributedIssueUnknown covers the Unknown path; both pass.
          round: 4
        - id: BR-12
          disposition: addressed
          note: README.md now states owner recording, --adopt, move carrying ownership and the rebuild flag day.
          round: 4
        - id: BR-13
          disposition: addressed
          note: gofmt -l no longer lists move.go.
          round: 4
        - id: BR-14
          disposition: addressed
          note: '{{OWNERSHIP_GATE}} rendered from ownershipGateHelp in main.go renderLong; paragraph moved out of change-code''s numbered list (a stray double blank line remains).'
          round: 4
        - id: BR-15
          disposition: not-addressed
          note: Code wraps the probe error, but no test reaches it (no malformed relocation record fixture); the fix is unverified.
          round: 4
        - id: BR-16
          disposition: addressed
          note: Whole-issue close checks ownership inside prepareTrackerClose over the card it already reads; only milestone mode opens the tracker in computeClose.
          round: 4
      findings:
        - id: BR-17
          severity: Important
          title: Atlas and plan still state the pre-BR-9 relocation rule; relocation record family undocumented
          detail: '2nd in family. Rule: a fix round that changes a contract greps and updates every restatement (atlas, README, help, plan step/table, Revisions) in the same commit. atlas/workflow/issue-tracker.md:51-53 omits the required move record and its <git-common-dir>/sdlc/relocations/<id>.json lifecycle; the plan has no M2 Revision for the evidence rule or the set-status working->working refusal the Verb contract table still shows as "-".'
          family: docs-surface-gap
          round: 4
        - id: BR-18
          severity: Minor
          title: sdlc move leaves its relocation record when the first switch fails and nothing moved
          detail: move.go writes the record before switching; the "nothing was moved" return never removes it, leaving stale evidence that later authorizes a claim-relocation after hand switching. Remove it on that path (keep it on the second-switch failure, where it is the repair evidence).
          family: effect-record-outlives-effect
          round: 4
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 5
      timestamp: "2026-10-01T15:02:20-07:00"
      agent: claude
      dispose:
        - id: BR-15
          disposition: addressed
          note: TestGateSurfacesAnUnreadableMoveRecord drives a malformed record through the gate; reverting the wrapped error in a scratch copy turns it red.
          round: 5
        - id: BR-17
          disposition: addressed
          note: atlas/workflow/issue-tracker.md:51-67 states the record evidence rule and full lifecycle; plan has the M2 rounds 1-2 Revision, superseded note, and the updated Verb contract cell; lesson recorded.
          round: 5
        - id: BR-18
          disposition: addressed
          note: move.go removes the record on the first-switch failure; TestMoveFirstSwitchFailureLeavesNoRecord is red with the removal reverted; second-switch failure intentionally keeps it.
          round: 5
      findings:
        - id: BR-19
          severity: Important
          title: claim --help omits --adopt and still says any other workspace's repeat claim is refused, contradicting the relocation repair
          detail: '3rd in family. helptext/claim.md FLAGS lacks --adopt (registered at claim.go:78) and lines 34-36 contradict claim.go:143-165. Rule: every registered flag appears in its rendered help (enforce with a VisitAll-over-subcommands test beside helpflags_test.go), and the restatement sweep in lessons.md includes the help page of every verb whose behavior changed.'
          family: docs-surface-gap
          round: 5
        - id: BR-20
          severity: Minor
          title: Plan superseded-note cites Revision "M2 review round 1" but the entry is titled "M2 review rounds 1-2"
          family: plan-code-contract-drift
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#277 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-01T14:01:43-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `test-oracle-single-interleaving` Claim race test rejects the foreign-owner refusal a late-reading loser now gets
  A loser whose snapshot follows the winner's publish hits ownedBy, which returns "is working, claimed by ...". That message contains neither "not open" nor "changed while claiming", so the test is green only when both clones read before either publishes. Accept "claimed by" (or add a deterministic in-process loser test), and update the claim.md loser wording to match.
- **BR-2** [Important] `plan-code-contract-drift` Plan promises MachineFingerprint(raw, key) plus a key test; code has MachineFingerprint(raw)
  The Core concepts table and the M1 step describe a key parameter and a "different key gives a different value" test. The code uses a fixed domain prefix and tests separation between different raw IDs. The behaviour is fine; add a plan Revisions entry so the plan matches the code.
- **BR-3** [Minor] `dry-run-writes` An owner's repeat claim with --dry-run still refreshes and writes the local details mirror
- **BR-4** [Minor] `rollout-doc-timing` M1 already writes claimant cards to the shared tracker, but the flag-day rollout note is deferred to M2
  Any claim made with this branch's binary breaks stale fleet binaries. Add the atlas rollout paragraph now, or avoid claiming with this binary until the work lands.
- **BR-5** [Minor] `plan-code-contract-drift` The race-test skip on an unreadable host machine ID promised by the plan is not implemented
- **BR-6** [Minor] `layout-knowledge-outside-owner` slotLabel re-derives the slot directory layout in cmd/sdlc instead of asking pkg/workspace
- **BR-7** [Minor] `missing-branch-unit-test` No claimDecision table test with a non-nil claimant; the Unknown (legacy working card) refusal is untested in M1

## Round 2 — 2026-10-01T14:06:41-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Oracle now accepts "claimed by" (claimremote_test.go:182); help text updated; TestClaimDecisionOwnership pins the foreign-owner refusal deterministically.
- BR-2 — addressed — Plan Revisions entry 2026-10-01 records the fixed-key MachineFingerprint(raw); domain-separation test added.
- BR-3 — addressed — claim.go:138 dry-run guard; TestClaimDryRunOwnerRepeatWritesNothing goes red when the guard is reverted (verified in scratch worktree).
- BR-4 — addressed — Rollout paragraph now in atlas/workflow/issue-tracker.md within M1.
- BR-5 — addressed — TestClaimRaceHasExactlyOneWinner skips when machineID() errors.
- BR-6 — addressed — slotLabel delegates to pkg/workspace Identity.UsesSlotLayout, which reuses slotNumber.
- BR-7 — addressed — TestClaimDecisionOwnership covers stamp, Mine, Foreign, Unknown (--adopt), and non-working statuses.

### Raised

- **BR-8** [Minor] `stale-test-precondition` TestSlotLabelOnlyWhereSlotsExist's "no slots" check runs after the malformed slot dir is created, duplicating the prior assertion
  Move the plain-primary assertion before writing worktree/ariadne-slotX so the no-slot-dirs state is still exercised.

## Round 3 — 2026-10-01T14:34:56-07:00 (claude) — BLOCKED

### Raised

- **BR-9** [Important] `absence-as-authority` Relocation reassigns a working card on any same-machine worktree whose owner merely is not on the issue branch
  relocatable/RelocationAllowed (cmd/sdlc/claimant.go:162) accept "recorded worktree does not hold the branch" as evidence of a move; that is also true right after claim (before start-plan) or when the owner switched away, so `git switch -c 000NNN-slug && sdlc claim` in another slot takes the card. Require positive move evidence (e.g. a local relocation marker written by sdlc move before its CAS) and add the negative test.
- **BR-10** [Important] `plan-code-contract-drift` Verb-contract cells for adopt/set-status untested and drifting (no adopt race/table; set-status working on unattributed working card silently adopts)
  3rd finding in this family. Rule: the plan's Verb contract table is the source; one table-driven test enumerates every situation x verb cell. Lift adopt's decision out of IO adoptClaim (claim.go:561) into a pure function, make statusDecision refuse working->working on an unattributed card toward --adopt (setstatus.go:144), add the planned two-clone adopt race.
- **BR-11** [Important] `missing-branch-unit-test` moveRelocation success path and Unknown branch never exercised; planned real move test not delivered
  2nd finding in this family. Rule: every return path of a new effectful function gets a fixture. TestRelocationAfterMoveAndRepair covers only the still-held refusal and a stubbed failure; the move.go:140 relocateClaimant success and move.go:131 Unknown return are unreached. Drive runMove (or unstubbed moveRelocation) to a success and an Unknown case.
- **BR-12** [Important] `docs-surface-gap` README Concurrent issue work omits claim ownership, --adopt and the rebuild flag day
  README.md:15-27 describes claim/start-plan in tracker repos; add one sentence on the owning workspace, --adopt for legacy cards, and rebuilding sdlc after landing.
- **BR-13** [Minor] `gofmt-clean` cmd/sdlc/move.go import block is not gofmt-sorted
- **BR-14** [Minor] `duplicated-help-prose` Identical OWNERSHIP paragraph copied into four helptexts; use a shared template placeholder
  ARCH-DRY; also the change-code.md paragraph splits "in any checkout:" from its numbered list.
- **BR-15** [Minor] `swallowed-probe-error` requireIssueOwnership discards relocatable's error and reports a plain Foreign refusal
- **BR-16** [Minor] `repeated-tracker-snapshot` Tracker-era close opens the tracker and snapshots twice (ownership gate, then prepareTrackerClose)

## Round 4 — 2026-10-01T14:56:04-07:00 (claude) — BLOCKED

### Disposed

- BR-9 — addressed — RelocationAllowed + relocatable require sdlc move's record naming from/to; negative integration test plus pure "no move record" case; red when the record check is removed.
- BR-10 — addressed — adoptDecision is pure; TestVerbContractTable covers every cell (red with the working->working refusal disabled); TestAdoptRaceHasExactlyOneWinner added.
- BR-11 — addressed — TestMoveRelocatesItsOwner drives real runMove to success; TestMoveLeavesAnUnattributedIssueUnknown covers the Unknown path; both pass.
- BR-12 — addressed — README.md now states owner recording, --adopt, move carrying ownership and the rebuild flag day.
- BR-13 — addressed — gofmt -l no longer lists move.go.
- BR-14 — addressed — {{OWNERSHIP_GATE}} rendered from ownershipGateHelp in main.go renderLong; paragraph moved out of change-code's numbered list (a stray double blank line remains).
- BR-15 — not-addressed — Code wraps the probe error, but no test reaches it (no malformed relocation record fixture); the fix is unverified.
- BR-16 — addressed — Whole-issue close checks ownership inside prepareTrackerClose over the card it already reads; only milestone mode opens the tracker in computeClose.

### Raised

- **BR-17** [Important] `docs-surface-gap` Atlas and plan still state the pre-BR-9 relocation rule; relocation record family undocumented
  2nd in family. Rule: a fix round that changes a contract greps and updates every restatement (atlas, README, help, plan step/table, Revisions) in the same commit. atlas/workflow/issue-tracker.md:51-53 omits the required move record and its <git-common-dir>/sdlc/relocations/<id>.json lifecycle; the plan has no M2 Revision for the evidence rule or the set-status working->working refusal the Verb contract table still shows as "-".
- **BR-18** [Minor] `effect-record-outlives-effect` sdlc move leaves its relocation record when the first switch fails and nothing moved
  move.go writes the record before switching; the "nothing was moved" return never removes it, leaving stale evidence that later authorizes a claim-relocation after hand switching. Remove it on that path (keep it on the second-switch failure, where it is the repair evidence).

## Round 5 — 2026-10-01T15:02:20-07:00 (claude) — BLOCKED

### Disposed

- BR-15 — addressed — TestGateSurfacesAnUnreadableMoveRecord drives a malformed record through the gate; reverting the wrapped error in a scratch copy turns it red.
- BR-17 — addressed — atlas/workflow/issue-tracker.md:51-67 states the record evidence rule and full lifecycle; plan has the M2 rounds 1-2 Revision, superseded note, and the updated Verb contract cell; lesson recorded.
- BR-18 — addressed — move.go removes the record on the first-switch failure; TestMoveFirstSwitchFailureLeavesNoRecord is red with the removal reverted; second-switch failure intentionally keeps it.

### Raised

- **BR-19** [Important] `docs-surface-gap` claim --help omits --adopt and still says any other workspace's repeat claim is refused, contradicting the relocation repair
  3rd in family. helptext/claim.md FLAGS lacks --adopt (registered at claim.go:78) and lines 34-36 contradict claim.go:143-165. Rule: every registered flag appears in its rendered help (enforce with a VisitAll-over-subcommands test beside helpflags_test.go), and the restatement sweep in lessons.md includes the help page of every verb whose behavior changed.
- **BR-20** [Minor] `plan-code-contract-drift` Plan superseded-note cites Revision "M2 review round 1" but the entry is titled "M2 review rounds 1-2"

## Open findings

- **BR-8** [Minor] `stale-test-precondition` TestSlotLabelOnlyWhereSlotsExist's "no slots" check runs after the malformed slot dir is created, duplicating the prior assertion
- **BR-19** [Important] `docs-surface-gap` claim --help omits --adopt and still says any other workspace's repeat claim is refused, contradicting the relocation repair
- **BR-20** [Minor] `plan-code-contract-drift` Plan superseded-note cites Revision "M2 review round 1" but the entry is titled "M2 review rounds 1-2"
