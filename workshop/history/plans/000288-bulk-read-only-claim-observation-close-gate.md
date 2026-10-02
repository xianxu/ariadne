---
gate: boundary-review
issue: 288
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-02T14:04:49-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Task 2/4 test bullets enumerate cases in prose; compress to one strategy line per risky function
          detail: (carried from plan-quality PQ-3, deferred to the boundary review)
          family: test-prose-enumeration
          round: 1
        - id: BR-2
          severity: Minor
          title: Task 2 site table is a line-numbered call-site inventory; four lines do not match a Card == nil check
          detail: |-
            close.go:515, actual.go:173, trackercompletion.go:39 and projectstatus.go:304 were confirmed; push.go:561, issuefiles.go:85, issue.go:606 and observe.go:50 did not match a grep for Card == nil. Find the sites by rec.Card use rather than trusting the line numbers.
            (carried from plan-quality PQ-4, deferred to the boundary review)
          family: test-prose-enumeration
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-10-02T14:04:49-07:00"
      agent: claude
      findings:
        - id: BR-3
          severity: Important
          title: ownedCompletions silently skips unreadable cards on landing/recovery paths that transferguard does not guard
          detail: trackercompletion.go:39 skips rec.Card == nil (now including CardErr). completeLandingPR (landing.go:456, post-merge) and settleLandedCompletions (issuerecovery.go:105, publishgate.go:349) do not run transferguard, so a codecomplete card that became unreadable is never completed and no one is told; the plan's "unreachable" claim is false. Refuse or warn naming the card, add a test, and revise the plan row.
          family: unreadable-card-read-as-absent
          round: 2
        - id: BR-4
          severity: Important
          title: Per-reader CardErr branches (close, actual, push, projectstatus, issuefiles, fleet LookupRepoIssues, issue show text, set-status) have no test
          detail: Only claim, observe, listIssueStates/drift, transferguard and issue new are exercised by TestOneMalformedCardDoesNotBlockOthers; the plan's e2e also promised an issue set-status refusal. Extend the same fixture with direct calls to these sites.
          family: per-site-branch-untested
          round: 2
        - id: BR-5
          severity: Minor
          title: overlayCardStatus fails the whole scan rather than showing the card unreadable as the plan table states
          detail: Fail-closed is defensible since every caller is a publish path transferguard refuses; record it as a plan revision.
          family: plan-table-drift
          round: 2
        - id: BR-6
          severity: Minor
          title: transferguard names only the first unreadable card
          detail: With several malformed cards the operator repairs them one per round trip; list every ID.
          family: error-names-one-of-many
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-02T14:11:18-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Task 4 now uses a one-line "strategy:" test bullet; Task 2's enumerated bullet describes the delivered fixture and is moot.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Reader table rewritten keyed by function name (plan lines 170-182) with a Revisions entry; matches the grep of CardErr sites.
          round: 3
        - id: BR-3
          disposition: addressed
          note: trackercompletion.go:39 refuses on CardErr; asserted by ownedCompletions call in malformedcard_test.go, which fails under the old skip; recovery surfaces it via cwarn.
          round: 3
        - id: BR-4
          disposition: not-addressed
          note: All reader sites now exercised except close.go:515 milestone-mode CardErr die; add it to the fixture via the die-catching helper.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Plan row now records overlayCardStatus as fail-closed, plus a Revisions entry.
          round: 3
        - id: BR-6
          disposition: addressed
          note: transferguard.go joins every Unreadable() entry via errors.Join.
          round: 3
      findings:
        - id: BR-7
          severity: Minor
          title: Landing completion is all-or-nothing on any unreadable card; document the deferral
          detail: '2nd in family; the rule (every CardErr reader refuses or reports, never skips) now holds at all 11 grep-enumerated sites. Remaining nuance: an unrelated malformed card defers healthy post-merge completions until repair; settle re-derivation recovers. Note it in the atlas quarantine section.'
          family: unreadable-card-read-as-absent
          round: 3
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-10-02T14:14:45-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: malformedcard_test.go now calls set-status, historyFileIsTerminal, lookupIssueMeta, overlayCardStatus, fleet.LookupRepoIssues, ownedCompletions, actualTrackerInputs, computeClose and runIssueShow text directly. With the projectstatus and fleet fixes removed in a scratch copy, the test fails at exactly those two checks.
          round: 4
        - id: BR-7
          disposition: addressed
          note: atlas/workflow/issue-tracker.md quarantine section now states that landing and recovery refuse while any card is unreadable, deferring healthy completions until repair, with the next settle re-deriving them.
          round: 4
      findings:
        - id: BR-8
          severity: Minor
          title: guardIssueNotDone reads an unreadable card's status as not-done, contrary to its fail-closed contract
          detail: '3rd in family. Rule: every card-owned read on an IssueRecord (Card, Status(), card-owned Field) checks CardErr first. The prior sweep grepped .Card == nil and snap.Card, so it missed rs.Get(...).Status()/Field() reads. Of 10 rs.Get sites, 3 do not check CardErr: repoguard.go:104 (fails open, masked because start-plan and change-code call snap.Require next), pr.go:175 (safe, both callers run after transferguard), observe.go:95 (harmless, card already Unknown). Structural fix: add Records.Require(id) mirroring Snapshot.Require and route card-owned reads through it, starting with guardIssueNotDone.'
          family: unreadable-card-read-as-absent
          round: 4
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 5
      timestamp: "2026-10-02T14:29:20-07:00"
      agent: claude
      findings:
        - id: BR-9
          severity: Important
          title: dangling_claims keyed by git checkout identity; separate clones of one repo report each other's live claims as dangling
          detail: 'PlaceClaims keys placed/dangling by RepoIdentity (the git dir) while each repo read returns every tracker card; a claim placed on clone B''s row is also emitted as dangling under clone A. Fix: dangling only when the claimant worktree matches no row inventory-wide, or key the join by tracker repository; add a two-identity PlaceClaims case.'
          family: tracker-entity-keyed-by-checkout
          round: 5
        - id: BR-10
          severity: Important
          title: repoClaimsFrom's partial/absent/unknown/duplicate branches are untested; Done-when promises partial is tested
          detail: '2nd finding in this family. Rule: every pure function mapping an input state space to a quality value is table-tested over that state space, not only through its consumer. TestPlaceClaims feeds hand-built RepoClaims; add a repoClaimsFrom table test and assert claims_state partial naming 42 in TestOneMalformedCardDoesNotBlockOthers.'
          family: per-site-branch-untested
          round: 5
        - id: BR-11
          severity: Important
          title: inventory now probes every repository's tracker remote one after another with no latency bound (ARCH-CONSTRAINTS)
          detail: Previously only repos with issue-prefixed branches triggered repoRecords; now every repo with rows runs LoadRecords(PreferFresh) and Presence(), which checks the remote, including gcrypt brain remotes. Decide absent locally first, or declare a budget and bound each probe via ctx.
          family: undeclared-io-fanout
          round: 5
        - id: BR-12
          severity: Minor
          title: observe.go:95 reads rec.Status() via Get; an unreadable card's empty status selects the non-done evidence path
          detail: 4th instance in this family. The rule exists (Records.Require); make it structural, e.g. IssueRecord.Status reports unknown, or a lint forbidding Get(...).Status().
          family: unreadable-card-read-as-absent
          round: 5
        - id: BR-13
          severity: Minor
          title: Inventory.MarshalJSON writes defaults into the caller's Rows backing array
          family: marshal-mutates-input
          round: 5
        - id: BR-14
          severity: Minor
          title: fleet fingerprintPattern duplicates issue.fingerprintRE; export one validator from issue
          family: duplicated-validator
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 6
      timestamp: "2026-10-02T14:38:08-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: PlaceClaims marks a claim dangling only when its worktree is no row anywhere and dedupes by trackerIssueKey; TestPlaceClaimsAcrossClones fails on the old per-identity key.
          round: 6
        - id: BR-10
          disposition: addressed
          note: 'TestRepoClaimsFrom table covers absent/unknown x2/present/stale/partial/stale+partial plus duplicate; malformedcard_test asserts partial naming #000042.'
          round: 6
        - id: BR-11
          disposition: not-addressed
          note: The CutOver early return in LookupRepoClaims (issues.go:147) has no regression test; removing it leaves the suite green. Add an untracked checkout with an unreachable origin and assert absent with no error. Tracked repos still fetch one after another with no declared budget; record that in the plan.
          round: 6
        - id: BR-12
          disposition: addressed
          note: Records.Get is now unexported (records.go:112), so outside callers must use Require; observe.go:97 uses Require and skips Status on cerr. All() callers each check CardErr.
          round: 6
        - id: BR-13
          disposition: addressed
          note: MarshalJSON copies rows (types.go:420); TestInventoryMarshalDoesNotMutate fails against the old in-place write.
          round: 6
        - id: BR-14
          disposition: addressed
          note: issue.ValidFingerprint is exported and used by fleet; the local fingerprintPattern is deleted.
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 7
      timestamp: "2026-10-02T14:40:10-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: addressed
          note: LookupRepoClaims checks the local CutOver signal (marker or fetched ref) before any remote probe, so untracked repos cost no network; TestFleetClaimsSkipUntrackedRemotes would read unknown without the check. Budget declared in the plan.
          round: 7
      findings:
        - id: BR-15
          severity: Minor
          title: Plan budget says tracker fetches are bounded by Git's transport timeout, but Presence takes no ctx and git fetch has no default deadline
          detail: 'This is the 2nd finding in family undeclared-io-fanout. The rule: a network call on a read-only view path either takes the caller''s ctx deadline or its budget says it is unbounded. Thread ctx into Repository.Presence/Initialized, or correct the plan''s wording.'
          family: undeclared-io-fanout
          round: 7
        - id: BR-16
          severity: Minor
          title: renderMachineClaims and TestFleetInventoryPlacesClaims compare Machine.State against ClaimsPresent instead of MachinePresent
          family: cross-enum-constant
          round: 7
      boundary: M2
      recipe: milestone-review
      blocked: false
    - "n": 8
      timestamp: "2026-10-02T14:42:23-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: Records.Require added; all rs.Get card reads (repoguard, pr, observe, close, actual, push, issue, projectstatus, issuefiles) routed through it; guardIssueNotDone fail-closed asserted in malformedcard_test.go.
          round: 8
        - id: BR-15
          disposition: not-addressed
          note: Plan now states unbounded (good), but asserts the command ctx cancels the fetch; Presence/RemoteExists (nil ctx) and Snapshot take no ctx, ctx.Err() is only checked after. Delete that clause.
          round: 8
        - id: BR-16
          disposition: addressed
          note: render.go:137 and fleetclaims_test.go:99 now use MachinePresent; same string value so no behavioral test can fail, verified by inspection.
          round: 8
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#288 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T14:04:49-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `test-prose-enumeration` Task 2/4 test bullets enumerate cases in prose; compress to one strategy line per risky function
  (carried from plan-quality PQ-3, deferred to the boundary review)
- **BR-2** [Minor] `test-prose-enumeration` Task 2 site table is a line-numbered call-site inventory; four lines do not match a Card == nil check
  close.go:515, actual.go:173, trackercompletion.go:39 and projectstatus.go:304 were confirmed; push.go:561, issuefiles.go:85, issue.go:606 and observe.go:50 did not match a grep for Card == nil. Find the sites by rec.Card use rather than trusting the line numbers.
  (carried from plan-quality PQ-4, deferred to the boundary review)

## Round 2 — 2026-10-02T14:04:49-07:00 (claude) — BLOCKED

### Raised

- **BR-3** [Important] `unreadable-card-read-as-absent` ownedCompletions silently skips unreadable cards on landing/recovery paths that transferguard does not guard
  trackercompletion.go:39 skips rec.Card == nil (now including CardErr). completeLandingPR (landing.go:456, post-merge) and settleLandedCompletions (issuerecovery.go:105, publishgate.go:349) do not run transferguard, so a codecomplete card that became unreadable is never completed and no one is told; the plan's "unreachable" claim is false. Refuse or warn naming the card, add a test, and revise the plan row.
- **BR-4** [Important] `per-site-branch-untested` Per-reader CardErr branches (close, actual, push, projectstatus, issuefiles, fleet LookupRepoIssues, issue show text, set-status) have no test
  Only claim, observe, listIssueStates/drift, transferguard and issue new are exercised by TestOneMalformedCardDoesNotBlockOthers; the plan's e2e also promised an issue set-status refusal. Extend the same fixture with direct calls to these sites.
- **BR-5** [Minor] `plan-table-drift` overlayCardStatus fails the whole scan rather than showing the card unreadable as the plan table states
  Fail-closed is defensible since every caller is a publish path transferguard refuses; record it as a plan revision.
- **BR-6** [Minor] `error-names-one-of-many` transferguard names only the first unreadable card
  With several malformed cards the operator repairs them one per round trip; list every ID.

## Round 3 — 2026-10-02T14:11:18-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Task 4 now uses a one-line "strategy:" test bullet; Task 2's enumerated bullet describes the delivered fixture and is moot.
- BR-2 — addressed — Reader table rewritten keyed by function name (plan lines 170-182) with a Revisions entry; matches the grep of CardErr sites.
- BR-3 — addressed — trackercompletion.go:39 refuses on CardErr; asserted by ownedCompletions call in malformedcard_test.go, which fails under the old skip; recovery surfaces it via cwarn.
- BR-4 — not-addressed — All reader sites now exercised except close.go:515 milestone-mode CardErr die; add it to the fixture via the die-catching helper.
- BR-5 — addressed — Plan row now records overlayCardStatus as fail-closed, plus a Revisions entry.
- BR-6 — addressed — transferguard.go joins every Unreadable() entry via errors.Join.

### Raised

- **BR-7** [Minor] `unreadable-card-read-as-absent` Landing completion is all-or-nothing on any unreadable card; document the deferral
  2nd in family; the rule (every CardErr reader refuses or reports, never skips) now holds at all 11 grep-enumerated sites. Remaining nuance: an unrelated malformed card defers healthy post-merge completions until repair; settle re-derivation recovers. Note it in the atlas quarantine section.

## Round 4 — 2026-10-02T14:14:45-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — malformedcard_test.go now calls set-status, historyFileIsTerminal, lookupIssueMeta, overlayCardStatus, fleet.LookupRepoIssues, ownedCompletions, actualTrackerInputs, computeClose and runIssueShow text directly. With the projectstatus and fleet fixes removed in a scratch copy, the test fails at exactly those two checks.
- BR-7 — addressed — atlas/workflow/issue-tracker.md quarantine section now states that landing and recovery refuse while any card is unreadable, deferring healthy completions until repair, with the next settle re-deriving them.

### Raised

- **BR-8** [Minor] `unreadable-card-read-as-absent` guardIssueNotDone reads an unreadable card's status as not-done, contrary to its fail-closed contract
  3rd in family. Rule: every card-owned read on an IssueRecord (Card, Status(), card-owned Field) checks CardErr first. The prior sweep grepped .Card == nil and snap.Card, so it missed rs.Get(...).Status()/Field() reads. Of 10 rs.Get sites, 3 do not check CardErr: repoguard.go:104 (fails open, masked because start-plan and change-code call snap.Require next), pr.go:175 (safe, both callers run after transferguard), observe.go:95 (harmless, card already Unknown). Structural fix: add Records.Require(id) mirroring Snapshot.Require and route card-owned reads through it, starting with guardIssueNotDone.

## Round 5 — 2026-10-02T14:29:20-07:00 (claude) — BLOCKED

### Raised

- **BR-9** [Important] `tracker-entity-keyed-by-checkout` dangling_claims keyed by git checkout identity; separate clones of one repo report each other's live claims as dangling
  PlaceClaims keys placed/dangling by RepoIdentity (the git dir) while each repo read returns every tracker card; a claim placed on clone B's row is also emitted as dangling under clone A. Fix: dangling only when the claimant worktree matches no row inventory-wide, or key the join by tracker repository; add a two-identity PlaceClaims case.
- **BR-10** [Important] `per-site-branch-untested` repoClaimsFrom's partial/absent/unknown/duplicate branches are untested; Done-when promises partial is tested
  2nd finding in this family. Rule: every pure function mapping an input state space to a quality value is table-tested over that state space, not only through its consumer. TestPlaceClaims feeds hand-built RepoClaims; add a repoClaimsFrom table test and assert claims_state partial naming 42 in TestOneMalformedCardDoesNotBlockOthers.
- **BR-11** [Important] `undeclared-io-fanout` inventory now probes every repository's tracker remote one after another with no latency bound (ARCH-CONSTRAINTS)
  Previously only repos with issue-prefixed branches triggered repoRecords; now every repo with rows runs LoadRecords(PreferFresh) and Presence(), which checks the remote, including gcrypt brain remotes. Decide absent locally first, or declare a budget and bound each probe via ctx.
- **BR-12** [Minor] `unreadable-card-read-as-absent` observe.go:95 reads rec.Status() via Get; an unreadable card's empty status selects the non-done evidence path
  4th instance in this family. The rule exists (Records.Require); make it structural, e.g. IssueRecord.Status reports unknown, or a lint forbidding Get(...).Status().
- **BR-13** [Minor] `marshal-mutates-input` Inventory.MarshalJSON writes defaults into the caller's Rows backing array
- **BR-14** [Minor] `duplicated-validator` fleet fingerprintPattern duplicates issue.fingerprintRE; export one validator from issue

## Round 6 — 2026-10-02T14:38:08-07:00 (claude) — BLOCKED

### Disposed

- BR-9 — addressed — PlaceClaims marks a claim dangling only when its worktree is no row anywhere and dedupes by trackerIssueKey; TestPlaceClaimsAcrossClones fails on the old per-identity key.
- BR-10 — addressed — TestRepoClaimsFrom table covers absent/unknown x2/present/stale/partial/stale+partial plus duplicate; malformedcard_test asserts partial naming #000042.
- BR-11 — not-addressed — The CutOver early return in LookupRepoClaims (issues.go:147) has no regression test; removing it leaves the suite green. Add an untracked checkout with an unreachable origin and assert absent with no error. Tracked repos still fetch one after another with no declared budget; record that in the plan.
- BR-12 — addressed — Records.Get is now unexported (records.go:112), so outside callers must use Require; observe.go:97 uses Require and skips Status on cerr. All() callers each check CardErr.
- BR-13 — addressed — MarshalJSON copies rows (types.go:420); TestInventoryMarshalDoesNotMutate fails against the old in-place write.
- BR-14 — addressed — issue.ValidFingerprint is exported and used by fleet; the local fingerprintPattern is deleted.

## Round 7 — 2026-10-02T14:40:10-07:00 (claude) — passed

### Disposed

- BR-11 — addressed — LookupRepoClaims checks the local CutOver signal (marker or fetched ref) before any remote probe, so untracked repos cost no network; TestFleetClaimsSkipUntrackedRemotes would read unknown without the check. Budget declared in the plan.

### Raised

- **BR-15** [Minor] `undeclared-io-fanout` Plan budget says tracker fetches are bounded by Git's transport timeout, but Presence takes no ctx and git fetch has no default deadline
  This is the 2nd finding in family undeclared-io-fanout. The rule: a network call on a read-only view path either takes the caller's ctx deadline or its budget says it is unbounded. Thread ctx into Repository.Presence/Initialized, or correct the plan's wording.
- **BR-16** [Minor] `cross-enum-constant` renderMachineClaims and TestFleetInventoryPlacesClaims compare Machine.State against ClaimsPresent instead of MachinePresent

## Round 8 — 2026-10-02T14:42:23-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — Records.Require added; all rs.Get card reads (repoguard, pr, observe, close, actual, push, issue, projectstatus, issuefiles) routed through it; guardIssueNotDone fail-closed asserted in malformedcard_test.go.
- BR-15 — not-addressed — Plan now states unbounded (good), but asserts the command ctx cancels the fetch; Presence/RemoteExists (nil ctx) and Snapshot take no ctx, ctx.Err() is only checked after. Delete that clause.
- BR-16 — addressed — render.go:137 and fleetclaims_test.go:99 now use MachinePresent; same string value so no behavioral test can fail, verified by inspection.

## Open findings

- **BR-15** [Minor] `undeclared-io-fanout` Plan budget says tracker fetches are bounded by Git's transport timeout, but Presence takes no ctx and git fetch has no default deadline
