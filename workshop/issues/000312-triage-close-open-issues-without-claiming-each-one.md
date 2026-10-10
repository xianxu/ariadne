---
id: 000312
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '448757db1ab3fd0765e65de30e5fe96a24c959a8' # card fields mirrored from issue-cards; edit via sdlc
---

# Triage-close open issues without claiming each one

## Problem

Ending an open issue as `wontfix`/`punt` with its details archived goes through `sdlc abandon`, which requires the caller to own the issue. A triage pass over unowned open issues therefore costs one `claim` plus one `abandon` per issue. In the ariadne-robustness-1 housekeeping (2026-10-09), closing 13 superseded issues took 13 claims and 13 abandons. Each claim is a tracker commit that only exists to satisfy the owner check, and it was released moments later by the terminal state.

Two related rough edges came up in the same pass:
- **A `working` card with no branch** (legacy #052, from before the tracker) can't be abandoned. `abandon` insists on running from the issue's branch, so a throwaway branch had to be created from main by hand just to run it.
- **A card that is already terminal but has live details** (#249, #293, set by `set-status wontfix` before abandon existed) has no archive path at all. That half is tracked in #305 (ariadne-robustness-1); this issue doesn't duplicate it.

Evidence: `workshop/projects/ariadne-robustness-1.md`, the housekeeping detail block.

## Spec

- `sdlc abandon --issue a,b,c --as wontfix|punt --reason …` accepts a set, and for **open, unowned** issues needs no prior claim. It uses one card compare-and-swap per issue (or one batched tracker commit) and one narrow main commit archiving all the details. The reason goes into each issue's Log; the claimant field records who triaged.
- An issue owned by *another* workspace still refuses (that would be a takeover; use `reclaim`).
- Started work with no branch, local or remote, is abandoned as if it were open: there is nothing to keep under `refs/ariadne/abandoned/`. Say so in the output instead of demanding a branch.

## Done when

- One `sdlc abandon --issue a,b,c --as wontfix --reason …` from a resting branch ends three unowned open issues, with their details archived on main, without any `claim`.
- A `working` card with no branch is abandoned without hand-creating a branch.
- Abandoning an issue owned by another workspace still refuses.

## Plan

## Log

### 2026-10-09

Filed from the ariadne-robustness-1 housekeeping, at the operator's request. Not in that project's MVP; a candidate for batch 2.
