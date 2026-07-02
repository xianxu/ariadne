---
id: 000142
status: working
deps: []
github_issue:
created: 2026-06-29
updated: 2026-07-01
estimate_hours:
started: 2026-07-01T23:05:44-07:00
---

# sdlc pre-merge judges should run at the earliest useful gate

## Problem

`sdlc merge` runs pre-merge judges after the PR branch is pushed:

- `plan` — issue plan completeness;
- `specs` — atlas/README sync;
- `lessons` — whether new lessons should be captured.

In pair#84, the `specs` judge found an optional README keybinding gap only at
merge time, after the issue had already been closed with a `SHIP` boundary
review and a PR had been opened. The fix was small, but it forced another
commit/push/merge loop.

Some of these checks may be more useful earlier:

- plan completeness belongs near `sdlc close`, before status becomes done;
- atlas/README sync may belong near close because it is part of "is the issue
  complete?";
- lessons might still be appropriate at merge, because it considers the whole
  branch/session and is explicitly a pre-ship reflection.

The current all-at-merge placement maximizes late discovery.

## Spec

Audit the pre-merge judges and decide the earliest useful gate for each
category.

Candidate placement:

- `plan`: run at `sdlc close` or immediately before close metadata finalization.
- `specs`: run at `sdlc close` for user-facing docs/atlas sync, possibly still
  repeated at merge for final branch-level drift.
- `lessons`: likely remain at merge, or become a soft post-close prompt, because
  it is less tied to issue acceptance criteria.

The goal is not to remove pre-merge safety. The goal is to catch fixable
acceptance/documentation gaps before close has recorded a final verdict and
before PR publishing/merge cleanup.

This issue should also decide whether a close-time `specs` or `plan` judge
overlaps with the boundary review and whether any overlap should be folded into
the close prompt instead of running another full LLM pass.

## Done when

- [ ] The existing pre-merge judge categories and prompts are documented with
      their current gate.
- [ ] Each category has an explicit target gate: close, PR, merge, or multiple.
- [ ] Checks that move earlier do not create redundant slow LLM passes when the
      boundary review already covers the same requirement.
- [ ] Merge still protects against final drift after PR creation.
- [ ] Help text explains which judge runs where and why.
- [ ] Tests cover that moved judges run at the new gate and failures stop before
      later/published steps.

## Plan

- [ ] Inspect `cmd/sdlc/preflight.go`, `cmd/sdlc/merge.go`, and `cmd/sdlc/push.go`.
- [ ] Compare `plan`, `specs`, and `lessons` prompts with close boundary review
      coverage.
- [ ] Decide whether to move, duplicate, or fold each check.
- [ ] Implement the selected gate placement with flags preserving emergency
      bypass semantics.
- [ ] Update close/merge/push help text.
- [ ] Add regression tests for a specs failure caught before merge.

## Log

### 2026-06-29

- Created from pair#84 dogfooding: `sdlc merge`'s `specs` judge caught a README
  keybinding documentation gap after close/PR, causing an avoidable final
  commit/push/merge loop.
