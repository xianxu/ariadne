---
id: 000311
status: working
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-12
estimate_hours:
card_mirror: 'dc7035ffcc096d96dcdeb629733c78852195ad44' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T18:12:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "acfe02c3", done: "f97daeb7"}
---

# Repair a codecomplete card whose branch merged outside the detector (#239)

## Problem

ariadne#239's PR #126 merged, and its details are archived in `workshop/history/`. Its card on `issue-tracker` is stuck at `codecomplete`, and `sdlc state` shows "#239 codecomplete card only".

#287's external-merge detector derives the branch from the card path (`cmd/sdlc/externalmerge.go:84`, `issue.BranchName`). But #239's real branch was `000239-standalone-weave-restart`, so the detector never sees the merge. `set-status → done` is refused by design. No verb can finish it.

## Spec

- Let the external-merge reconcile accept an explicit merged branch or PR for an issue (`sdlc issue recovery reconcile --issue N --merged-branch B` or `--pr N`). Verify ancestry on main, then finish: card → `done`, archive if needed.
- Alternatively, record the branch name on the card at start-plan so derivation never guesses. Decide at design.
- Repair #239 with it.

**Design (2026-10-12).** #239's card is a migrated legacy card: `codecomplete` with
no completion binding (the only such card on the tracker), and its real branch
is deleted from the remote. So neither the evidence settle nor the branch
detector can ever see it, and recording the branch at start-plan would neither
fix #239 nor cover the #148 rename (a branch renamed after start-plan).
Chosen: `sdlc issue recovery reconcile --issue N --landed <commit>`, which is git-only, so it works
with a deleted branch and with a squash merge:
- `<commit>` must be on main. Its first-parent range (`C^1..C`: a merge's
  brought-in work, or the commit itself) must carry a commit whose subject
  references the issue (`#N:` per the commit convention). That binds the
  commit to the issue without guessing a branch.
- Only a codecomplete card with **no** completion binding is repaired this way.
  A bound card's close carries its own evidence, and plain reconcile settles it.
  A done card is a no-op; any other status refuses.
- Card → `done` (+ `updated`); reconcile's existing sweep archives live details.
- Creates nothing durable beyond the one card write (ARCH lifecycle: N/A because
  no new file family).

## Done when

- #239's card reads `done`, and `sdlc state` no longer reports it.
- A test covers a card whose branch name differs from its details slug.

## Plan

- [x] `--landed` on reconcile + pure card transition + range check
- [x] tests: renamed branch merged then deleted → done, archived, state clean; refusals (not on main, no issue commit, bound card)
- [x] help/atlas mention
- [ ] repair #239 after landing

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1).

### 2026-10-12
- Only one unbound codecomplete card on the tracker: #239 (migrated 2026-09-27 by `issue migrate`). Its branch `000239-standalone-weave-restart` is deleted on origin; PR #126 merge commit is `b066449c`, whose `^1..` range carries `#239:` subjects.
- Implemented `repairLanded` / `doneUnboundCard` / `citesIssue` in `externalmerge.go`, flag in `issuerecovery.go`. Tests: `landedrepair_test.go` (renamed + deleted branch, refusals, bound card, subject matching). Targeted `Recovery|Help|Reconcile|ExternalMerge|Contract` green.

