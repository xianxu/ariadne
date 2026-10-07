---
id: 000301
status: open
deps: [ariadne#283]
github_issue:
created: 2026-10-07
updated: 2026-10-07
estimate_hours:
card_mirror: '40b99a77782c38aabbd1f22246f8f158d49cda23' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Write the close → rebase → reopen → close end-to-end test, including the reconcile path; check that it fails with either caller reverted.
- [ ] Add the atlas sentence.

## Log

### 2026-10-07

Filed from #283's last close review (round 6, two advisory Minors), at the operator's request. Quick flow expected.
