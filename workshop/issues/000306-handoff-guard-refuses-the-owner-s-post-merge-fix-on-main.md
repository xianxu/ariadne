---
id: 000306
status: open
deps: [ariadne#274]
github_issue:
created: 2026-10-08
updated: 2026-10-08
estimate_hours:
card_mirror: '1eaefb542a1bc8697c473af50faee27854d76c55' # card fields mirrored from issue-cards; edit via sdlc
---

# handoff guard refuses the owner's post-merge fix on main

## Problem

A third case for ariadne#274's ownership-by-branch-name rule
(`ownerAuthored`, `cmd/sdlc/transferguard.go`), observed in tools#83:

1. tools#83's issue branch merged as PR #57 (`sdlc merge`); `sdlc merge` returned the
   workspace to main and the local issue branch was removed.
2. The tag/tap release followed on main; `sdlc close --issue 83` then raised a
   review finding (BR-12), fixed on main, and recorded the close there (evidence
   commit `468300e`, card → codecomplete bound to it).
3. `sdlc push` refused: "landing would change handed-off issue details:
   workshop/issues/000083-….md would be overwritten by this branch". The editor was
   the card's own claimant (`relation: this-workspace`), editing its own issue.

The card was now bound to a commit only the local checkout had. Recovery used
(operator-authorized): fast-forward the four main commits onto the still-present
`origin/000083-…` so they count as owner-authored, then `sdlc push` — which
landed them, flipped the card to done and archived the issue.

The workaround depends on the remote issue branch surviving the merge (GitHub
auto-delete would remove it), and it reuses a merged branch name (#148 territory).

## Spec

Fold into ariadne#274's fix: when the card's claimant is this workspace (or the
commits are `#N`-tagged and touch only #N's details), the owner's edits on main
are owner-authored, whether or not an issue branch still exists. Also: the
issue-close flow that legitimately runs on main after a release (tag → tap →
close → push) should not need a branch-ref trick to publish.

## Done when

- `sdlc close` + `sdlc push` on main after the issue's PR merged publishes the close
  for the claimant, with the issue branch deleted locally and remotely; a test pins it.
- A non-claimant's rewrite of the same details from main is still refused.

## Plan

## Log

### 2026-10-08

Filed from tools#83 at the operator's request; see ariadne#274 for the shared rule.
