---
id: 000332
status: open
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-11
estimate_hours:
card_mirror: '5366064c82408cae8f62f4875221b209c75d91f0' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc merge integrates a moved main itself, then re-runs the suite before landing

## Problem

A branch that passed its close review can sit unmerged for hours, because main keeps moving under it. ariadne#330 passed its close review at 00:53 on 10-11 and was still open at 13:30: each `sdlc merge --yes` (run by the operator, since the auto-mode classifier refuses it for workers until ariadne#318) found main had moved and stopped with conflicts or a stale base, and the worker re-merged main by hand, waited for the operator again, and lost the race again. The operator's question (10-11): shouldn't `sdlc merge` just do that?

## Spec

`sdlc merge` integrates origin/main itself when main moved since the branch's last integration:
- **Clean merge** (git resolves it): commit the merge on the issue branch, re-run the repo's full suite (the same command `close` names as its suite), and land if it passes. No new review: the reviewed diff is unchanged, and #269's freshness rule already defines what main content a review must have seen; reuse it to decide if a re-review is owed (main code that touches the branch's files since the review → re-review; otherwise not).
- **Conflicts:** stop, leave the branch as it was (merge aborted), and print the conflicting paths and the next action for the agent (resolve, then close's `--fixed-then-ship` path, which already demands `--full-suite`).
- **Suite failure after a clean merge:** stop and name the failing tests; with a baseline run of main's suite, distinguish failures the merge introduced from ones already on main (and say so).
- Retry loop: if main moves again during the suite run, integrate again (bounded, e.g. 3 times), then stop with a clear message.
- This is also where the suite becomes enforced at landing, not agent-typed evidence: in repos without an in-binary test gate, `close` records `--verified` text, and failures have reached pair main that way (pair#450).

## Done when

- A branch whose main moved without conflicts lands with one `sdlc merge --yes`: merge commit, suite run, land (test with a fake remote).
- A conflicting branch stops with the paths and the next action, and the branch is unchanged (test).
- A suite failure introduced by the merge is reported as such, and one already on main is labeled pre-existing (test).

## Plan

- [ ] Design: integration step, suite command source, re-review rule via #269's freshness API, retry bound.
- [ ] Implement with fakes; tests for the Done-when rows; atlas.

## Log

### 2026-10-11
