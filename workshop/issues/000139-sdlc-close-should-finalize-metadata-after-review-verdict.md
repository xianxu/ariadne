---
id: 000139
status: working
deps: []
github_issue:
created: 2026-06-29
updated: 2026-06-30
estimate_hours:
started: 2026-06-30T12:11:03-07:00
---

# sdlc close should finalize metadata after review verdict

## Problem

`sdlc close` currently mutates the issue file before the boundary review has
returned. In pair#84, the first close attempt flipped the issue to `status:
done`, wrote `actual_hours`, and appended a "closed" log line, then the boundary
review returned `REWORK`.

That leaves the repo in an awkward intermediate state:

- the issue says `done` while the boundary says "do not cross yet";
- rerunning close requires `--no-reclose-guard`;
- repeated close attempts append stale close log lines that the operator must
  clean up manually;
- the close bookkeeping itself becomes tangled with the code-review loop.

The status flip should be the last successful act of close, not something that
happens before the gate has established that the boundary can be crossed.

## Spec

Redesign `sdlc close` as a two-phase operation:

1. Validate close form and compute the proposed close metadata in memory:
   `actual_hours`, verification line, planned status/log changes, and any atlas
   or project requirements.
2. Run the boundary review against the implementation window while the issue is
   still logically in progress. The judge should still see enough issue content
   to check plan/done-when state, but the repo should not be committed to
   `status: done` before the verdict is known.
3. Only after an acceptable verdict (`SHIP`, or `FIX-THEN-SHIP` when policy
   allows) apply the close metadata and produce the final close bookkeeping
   commit/trailer.

The final close bookkeeping commit is allowed to bypass another close review:
it records the result of the gate rather than introducing new implementation
surface. This is the "commit on its own" shape: code commits are reviewed by the
boundary review; the final close commit records status/actual/verdict and should
not recursively trigger another review.

If the boundary review returns `REWORK`, close should leave the issue in
`working` (or restore the pre-close issue file), print the verdict and next
action, and avoid appending a misleading "closed" log entry.

## Done when

- [ ] A failed/REWORK boundary review does not leave the issue at `status: done`.
- [ ] Rerunning close after fixing review findings does not require
      `--no-reclose-guard`.
- [ ] Close appends exactly one final "closed" log line for the successful close.
- [ ] The final close bookkeeping commit can carry `Review-Verdict:` trailers
      without causing a recursive review requirement.
- [ ] Tests cover `SHIP`, `FIX-THEN-SHIP`, `REWORK`, judge failure, and
      `--no-judge` close flows.

## Plan

- [ ] Inspect the current close mutation order in `cmd/sdlc/close.go`.
- [ ] Define the acceptable verdict policy for finalizing close metadata.
- [ ] Introduce a close plan/result object that can be computed before writes.
- [ ] Move issue-file mutation until after boundary review success.
- [ ] Preserve or improve the judge prompt's access to issue plan/done-when
      context without requiring a pre-review status flip.
- [ ] Add regression tests for REWORK leaving the issue open/working and for a
      later successful rerun producing one clean close line.

## Log

### 2026-06-29

- Created from pair#84 dogfooding: `sdlc close` flipped pair#84 to `done`
  before the boundary review returned `REWORK`, forcing re-close guard bypasses
  and manual cleanup of stale close log lines.
