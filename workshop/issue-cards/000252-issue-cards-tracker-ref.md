---
id: '000252'
status: done
started: 2026-09-25T14:26:14-07:00
created: 2026-09-25
updated: 2026-09-27
estimate_hours: 10.77
actual_hours: N/A
---

# Issue cards: card fields on a tracker ref, details on the branch

## Problem

An issue file mixes two kinds of content with different audiences, and one
file cannot live in two places:

- **Global, slow-changing:** id, status, claim, dates, hours, title, the
  original problem. Every agent, slot and peer needs the latest.
- **Branch-local, fast-changing:** Spec, Done when, Plan, Log. Only the agent
  doing the work needs them, and they belong next to the code (keeping the
  implementation plan with the code keeps agent context together).

Today both live in `workshop/issues/NNN-slug.md` on `main`. sdlc therefore
publishes *copies* of issue commits to `origin/main` (`issue new`, `claim`,
`change-code`'s `PublishCommit`) while the resting or code branch keeps its own
commits. Local and remote diverge, and every issue verb adds to it.

Evidence from pair on 2026-09-24/25 (on top of #251's list):

- `main-slot1` reached **3 ahead / 19 behind**. The 3 were #332 issue syncs
  (also on the #332 branch); the 19 were mostly peers' issue-sync commits.
  Reconciled by hand three times in one session (backup ref, rebase, verify
  tree, drop backup).
- `change-code` printed `Source commit 6f184dc2…: published (79982e7c…)`: the
  published copy has a different SHA, so the original stays "ahead" forever.
- Both landing PRs (pair#168, pair#169) were "not mergeable": `issue new` or
  `claim` had put a blank template or status-only edit on `origin/main`, the
  branch had the filled file with no shared ancestor, so add/add conflict,
  resolved by hand ("keep ours").
- `sdlc merge` ends with "workspace retained on unchanged main-slot1. Refresh is
  separate". Nothing refreshes it, so behind-counts only grow.
- Filing a spin-off issue mid-branch forces a choice between rebasing the code
  branch onto everyone's changes (disruptive) or making a local copy
  (divergence).
