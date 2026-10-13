---
gate: boundary-review
issue: 305
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T13:12:01-07:00"
      dispatched: "2026-10-13T13:04:04-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Done reopen never converges when origin/<branch> tracking ref is stale for the landed branch
          detail: 'reopendone.go:300 pushes via leasedBranchPush, which leases on the local remote-tracking ref; if origin deleted the landed branch after this checkout fetched, the push fails with stale info every time, and a plain fetch (no prune) does not clear it, though the error says rerun. Verified by a probe (update-ref refs/remotes/origin/000330-e2e to the landed tip, then reopen twice with a fetch between). Fix: when the local branch is judged landed, also drop the tracking ref if main contains it (or prune-fetch that ref) before pushing; add the probe as a regression test.'
          family: rerun-must-converge
          round: 1
        - id: BR-2
          severity: Minor
          title: A lost finish (mirror commit/push) after a done reopen's card write is not resumed by rerun
          detail: restoreForReopen sees status working and calls unfinishedReopen, which needs an archive ref the done path never has, so finishReopen never runs again; the catalog's "a lost response is settled by rerunning" covers only the card here.
          family: rerun-must-converge
          round: 1
        - id: BR-3
          severity: Minor
          title: restoreDone switches to an unlanded local branch before refusing it as predating main's archive
          detail: reopendone.go:279 switches, then :293 refuses, leaving the operator on the old branch; check archiveCommit against the tip before switching.
          family: refuse-before-mutate
          round: 1
        - id: BR-4
          severity: Minor
          title: Foreign-owner, wrong-branch, dirty-tree and --force paths of restoreDone have no tests
          family: refusal-paths-untested
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T13:43:25-07:00"
      dispatched: "2026-10-13T13:28:58-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: leaseOnRemoteBranch refreshes/drops the tracking ref; TestReopenDoneWithAStaleTrackingRef goes red with the call removed (stale info).
          round: 2
        - id: BR-2
          disposition: addressed
          note: unfinishedDoneReopen wired after unfinishedReopen; TestReopenDoneRerunFinishesALostFinish goes red without it.
          round: 2
        - id: BR-3
          disposition: addressed
          note: predates-archive check now precedes the switch (reopendone.go:94-101); refusal case red when the check is disabled.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestReopenDoneRefusals covers foreign owner, other branch, dirty tree, stale local branch, and --force.
          round: 2
      findings:
        - id: BR-5
          severity: Critical
          title: Reopen run while on the landed issue branch never cuts from main, un-archives nothing, yet pushes and flips the card
          detail: reopendone.go:87-112 resets to main only when env.branch != name; on the landed branch HEAD lacks the archive, unarchiveIssue no-ops, the old tip is pushed and the ok line lies. Probe confirmed (not based on main, no un-archive commit, card working). Refuse or switch to main first, before any mutation, and verify HEAD contains archiveCommit with live details before reporting success; add the probe as a test.
          family: success-without-postcondition
          round: 2
        - id: BR-6
          severity: Minor
          title: Log-guard refusal says "Nothing was changed" after the resting branch was fast-forwarded
          detail: '2nd in family. Rule: decide every refusal before the first mutation except a mutation the refusal''s remedy needs, and then name it in the refusal. fastForwardRest (reopendone.go:47) precedes the guard (:75) by design.'
          family: refuse-before-mutate
          round: 2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 3
      timestamp: "2026-10-13T14:12:30-07:00"
      dispatched: "2026-10-13T14:06:49-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: The landed-branch refusal at reopendone.go:74-75 is decided before any mutation; reverting it turns TestReopenDoneRefusals/the_landed_branch_itself red. Its postcondition half is hollow and is raised as a new finding.
          round: 3
        - id: BR-6
          disposition: not-addressed
          note: 'Only the Log guard names the move. The dirty-tree refusal (:104) and the malformed refusal (:114) still fire after fastForwardRest (:86) without naming it (probe: rest at HEAD~1 with cmd/a.go dirty, HEAD moved, refusal silent). Move the dirty check before :81; route post-ff refusals through one helper that appends moved.'
          round: 3
      findings:
        - id: BR-7
          severity: Important
          title: The BR-5 postcondition (reopendone.go:136-143) passes on the landed branch, so it guards nothing
          detail: '2nd in family. Rule: a postcondition must test the step''s effect, which the starting state could not already satisfy. A landed branch already has the live details and no archived copy; with the :74 refusal removed the old tip was pushed and "un-archived … and pushed" printed. Require env.ancestorOf(archiveCommit, "HEAD") as well, in a helper tested directly.'
          family: success-without-postcondition
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 4
      timestamp: "2026-10-13T14:47:02-07:00"
      dispatched: "2026-10-13T14:39:07-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: Dirty check moved above fastForwardRest; later refusals go through after(); TestReopenDoneDirtyTreeRefusesBeforeTheFastForward fails when the block is moved back below (verified).
          round: 4
        - id: BR-7
          disposition: addressed
          note: checkUnarchived requires ancestorOf(archiveCommit, HEAD); TestCheckUnarchived fails on the landed branch with that clause disabled (verified).
          round: 4
      findings:
        - id: BR-8
          severity: Minor
          title: Lease refusal for foreign work on origin's branch fires after the switch and un-archive commit, unnamed and untested
          detail: '3rd in family. Rule (lessons.md): decide every refusal before the first change, or name the change in it. Enumerate the returns after reopendone.go:126 and fix them all; the lease can be decided before the switch against mainTip and the local tip of the branch. Add a refusal test for this path.'
          family: refuse-before-mutate
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 225c7da3dc431a9162516d69ca96c27abc413d68
      blocked: false
---

# Gate ledger — ariadne#305 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T13:12:01-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `rerun-must-converge` Done reopen never converges when origin/<branch> tracking ref is stale for the landed branch
  reopendone.go:300 pushes via leasedBranchPush, which leases on the local remote-tracking ref; if origin deleted the landed branch after this checkout fetched, the push fails with stale info every time, and a plain fetch (no prune) does not clear it, though the error says rerun. Verified by a probe (update-ref refs/remotes/origin/000330-e2e to the landed tip, then reopen twice with a fetch between). Fix: when the local branch is judged landed, also drop the tracking ref if main contains it (or prune-fetch that ref) before pushing; add the probe as a regression test.
- **BR-2** [Minor] `rerun-must-converge` A lost finish (mirror commit/push) after a done reopen's card write is not resumed by rerun
  restoreForReopen sees status working and calls unfinishedReopen, which needs an archive ref the done path never has, so finishReopen never runs again; the catalog's "a lost response is settled by rerunning" covers only the card here.
- **BR-3** [Minor] `refuse-before-mutate` restoreDone switches to an unlanded local branch before refusing it as predating main's archive
  reopendone.go:279 switches, then :293 refuses, leaving the operator on the old branch; check archiveCommit against the tip before switching.
- **BR-4** [Minor] `refusal-paths-untested` Foreign-owner, wrong-branch, dirty-tree and --force paths of restoreDone have no tests

## Round 2 — 2026-10-13T13:43:25-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — leaseOnRemoteBranch refreshes/drops the tracking ref; TestReopenDoneWithAStaleTrackingRef goes red with the call removed (stale info).
- BR-2 — addressed — unfinishedDoneReopen wired after unfinishedReopen; TestReopenDoneRerunFinishesALostFinish goes red without it.
- BR-3 — addressed — predates-archive check now precedes the switch (reopendone.go:94-101); refusal case red when the check is disabled.
- BR-4 — addressed — TestReopenDoneRefusals covers foreign owner, other branch, dirty tree, stale local branch, and --force.

### Raised

- **BR-5** [Critical] `success-without-postcondition` Reopen run while on the landed issue branch never cuts from main, un-archives nothing, yet pushes and flips the card
  reopendone.go:87-112 resets to main only when env.branch != name; on the landed branch HEAD lacks the archive, unarchiveIssue no-ops, the old tip is pushed and the ok line lies. Probe confirmed (not based on main, no un-archive commit, card working). Refuse or switch to main first, before any mutation, and verify HEAD contains archiveCommit with live details before reporting success; add the probe as a test.
- **BR-6** [Minor] `refuse-before-mutate` Log-guard refusal says "Nothing was changed" after the resting branch was fast-forwarded
  2nd in family. Rule: decide every refusal before the first mutation except a mutation the refusal's remedy needs, and then name it in the refusal. fastForwardRest (reopendone.go:47) precedes the guard (:75) by design.

## Round 3 — 2026-10-13T14:12:30-07:00 (claude) — BLOCKED

### Disposed

- BR-5 — addressed — The landed-branch refusal at reopendone.go:74-75 is decided before any mutation; reverting it turns TestReopenDoneRefusals/the_landed_branch_itself red. Its postcondition half is hollow and is raised as a new finding.
- BR-6 — not-addressed — Only the Log guard names the move. The dirty-tree refusal (:104) and the malformed refusal (:114) still fire after fastForwardRest (:86) without naming it (probe: rest at HEAD~1 with cmd/a.go dirty, HEAD moved, refusal silent). Move the dirty check before :81; route post-ff refusals through one helper that appends moved.

### Raised

- **BR-7** [Important] `success-without-postcondition` The BR-5 postcondition (reopendone.go:136-143) passes on the landed branch, so it guards nothing
  2nd in family. Rule: a postcondition must test the step's effect, which the starting state could not already satisfy. A landed branch already has the live details and no archived copy; with the :74 refusal removed the old tip was pushed and "un-archived … and pushed" printed. Require env.ancestorOf(archiveCommit, "HEAD") as well, in a helper tested directly.

## Round 4 — 2026-10-13T14:47:02-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — Dirty check moved above fastForwardRest; later refusals go through after(); TestReopenDoneDirtyTreeRefusesBeforeTheFastForward fails when the block is moved back below (verified).
- BR-7 — addressed — checkUnarchived requires ancestorOf(archiveCommit, HEAD); TestCheckUnarchived fails on the landed branch with that clause disabled (verified).

### Raised

- **BR-8** [Minor] `refuse-before-mutate` Lease refusal for foreign work on origin's branch fires after the switch and un-archive commit, unnamed and untested
  3rd in family. Rule (lessons.md): decide every refusal before the first change, or name the change in it. Enumerate the returns after reopendone.go:126 and fix them all; the lease can be decided before the switch against mainTip and the local tip of the branch. Add a refusal test for this path.

## Open findings

- **BR-8** [Minor] `refuse-before-mutate` Lease refusal for foreign work on origin's branch fires after the switch and un-archive commit, unnamed and untested
