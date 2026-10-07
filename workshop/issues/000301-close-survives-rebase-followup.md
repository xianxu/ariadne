---
id: 000301
status: working
deps: [ariadne#283]
github_issue:
created: 2026-10-07
updated: 2026-10-07
estimate_hours:
card_mirror: '253e628ffce9d125515e0ca41e850164f502f11f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-07T13:10:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "0910e614", done: "44b4b072"}
---

# Pin the rebase-aware close rule: e2e test and atlas

## Problem

#283 made a close survive a rebase. `tracker/completeop.go` `newestClose` now treats the card's earlier close as replaced when the rebase rewrote its reviewed commit off the branch and the new close's reviewed commit is on it. The close command (`closetracker.go`) and reconcile (`issuerecovery.go`) pass `trackerEnv.closeAncestorOf`, which reads a commit this clone lacks as "not an ancestor". #283's final close review left two advisory Minors:

1. The rule and the wrapper are unit-tested separately (`TestNewestCloseAfterARebase`, `TestCloseAncestorOfTreatsAnUnknownCommitAsNoAncestor`). No test would fail if either caller went back to `env.ancestorOf`.
2. Only the recovery catalog's close `Ends` text states the rule; the atlas doesn't.

## Spec

- An end-to-end test over real git, using the existing tracker fixtures (`newTrackerRepo`, the close-tracker test helpers):
  - close an issue, then rewrite the branch (rebase onto an advanced main, or amend the reviewed commit);
  - reopen with `set-status working`, then close again: the new completion lands, bound to an evidence commit on the branch;
  - a FIX-THEN-SHIP resume through `sdlc issue recovery reconcile` lands the same way;
  - the old record's reviewed commit is absent from the clone (for example the test drops the backup ref and prunes), so the unknown-commit path is also exercised.
- One sentence in `atlas/workflow/issue-tracker.md`, near the close/receipt section, stating the supersession rule: newer by ancestry, or a close on the branch over one a rebase rewrote off it; a stale pre-rebase receipt still loses.

## Done when

- Reverting either `closetracker.go` or `issuerecovery.go` to `env.ancestorOf` makes the new end-to-end test fail (checked by doing it once).
- The atlas sentence is present and linked from the close section.
- `make test` is green.

## Plan

- [x] Write the close → rebase → reopen → close end-to-end test, including the reconcile path; check that it fails with either caller reverted.
- [x] Add the atlas sentence.

## Log

### 2026-10-07
- 2026-10-07: closed — TestTrackerCloseSurvivesARebase: three real-git variants of close → rebase onto advanced main → reopen → close (SHIP with old review present; SHIP with it pruned; FIX-THEN-SHIP via reconcile with it pruned) all land; reverting closetracker.go or issuerecovery.go to env.ancestorOf fails exactly its own variant (checked once each, restored); atlas issue-tracker.md states the rule; make test green (911 cmd/sdlc + 20 pkgs; processgroup /bin/ps sandbox-only); review verdict: SHIP

Filed from #283's last close review (round 6, two advisory Minors), at the operator's request. Quick flow expected.
Added #301 to project `claimant-ownership` (scope event in the project Log).

`TestTrackerCloseSurvivesARebase` (`closetracker_test.go`) runs three variants of close → rebase onto an advanced main → reopen → close again:
- (a) SHIP, with the old reviewed commit still in the clone;
- (b) SHIP, with it pruned;
- (c) FIX-THEN-SHIP landed by `recovery reconcile`, with it pruned.

The revert check:
- My first draft had only (a) and (c), and reverting `closetracker.go` to `env.ancestorOf` went undetected. With the old commit present, plain ancestry works. FIX-THEN-SHIP defers the generation check to reconcile.
- With (b) added, each revert fails exactly its own variant: closetracker → (b), issuerecovery → (c). Both pass restored.

Atlas: the supersession rule is stated in `atlas/workflow/issue-tracker.md` § Readers and completion, beside the binding. The close recovery contract cites the new test.

