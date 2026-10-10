---
id: 000311
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '58e51eaff81e61807c448ddfe11397558ecd9e31' # card fields mirrored from issue-cards; edit via sdlc
---

# Repair a codecomplete card whose branch merged outside the detector (#239)

## Problem

ariadne#239's PR #126 merged, and its details are archived in `workshop/history/`. Its card on `issue-tracker` is stuck at `codecomplete`, and `sdlc state` shows "#239 codecomplete card only".

#287's external-merge detector derives the branch from the card path (`cmd/sdlc/externalmerge.go:84`, `issue.BranchName`). But #239's real branch was `000239-standalone-weave-restart`, so the detector never sees the merge. `set-status → done` is refused by design. No verb can finish it.

## Spec

- Let the external-merge reconcile accept an explicit merged branch or PR for an issue (`sdlc issue recovery reconcile --issue N --merged-branch B` or `--pr N`). Verify ancestry on main, then finish: card → `done`, archive if needed.
- Alternatively, record the branch name on the card at start-plan so derivation never guesses. Decide at design.
- Repair #239 with it.

## Done when

- #239's card reads `done`, and `sdlc state` no longer reports it.
- A test covers a card whose branch name differs from its details slug.

## Plan

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1).
