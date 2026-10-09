---
id: 000304
status: open
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-08
estimate_hours:
card_mirror: '6cdf36050950af29b84a4df3fb3587fd723cbe83' # card fields mirrored from issue-cards; edit via sdlc
---

# Milestone review window absorbs integrated main; close's printed window misleads

## Problem

Integrating main into an issue branch by merge (the #395 lesson, #269's
direction for pushed branches) interacts with review windows in two ways. One
is correct but misleading, the other over-covers.

Measured on pair#410 (2026-10-08). M1 closed at `24112709`; M2's work landed;
`git merge origin/main` (`a37b9bee`, 93 main commits); then more #410 commits.

- **Whole-issue close: correct, but misleading output.** `boundaryWindowBase`
  returns `merge-base(main, HEAD)` = main's tip after the merge, and `close`
  prints `commit window: f904c117 → ad2fb30c`, which reads like "only the
  post-merge commits". The reviewer is handed a tree diff, though
  (`judge/reviewwindow.go`: `git diff BASE HEAD`), and since HEAD contains main's
  tip, that diff is exactly the branch's net change against main: all of M1 and
  M2 (86 files), nothing of main's. The window is right; the printed range and
  any commit-count reading of it are not.
- **Milestone close after a merge: over-covers.** A milestone's base is the prior
  milestone's `Review-Verdict` commit (`previousReviewBoundary`), which predates
  the merge. `git diff <prior boundary> HEAD` therefore includes every file main
  changed in the merge (here, 93 commits and roughly 11.8k lines of #395/#409
  work), and the reviewer reviews foreign code as this milestone's.

## Spec

- Milestone windows exclude integrated main: when a merge of main lies inside
  the window, diff against the merge's first parent, or compute the base as the
  tree "prior boundary + main's changes since" (for example the merge commit's
  second parent combined with the boundary), so only the issue's own changes
  are shown. Alternatively, refuse the milestone close with a next action
  ("close the whole issue instead, or re-base the boundary").
- `close` prints the window as what it is: "net change vs <main tip>", with the
  issue-commit count, not a bare SHA range that looks truncated.
- Fixture: issue commits, milestone close, merge of main with foreign changes,
  more issue commits. The milestone diff shows only issue changes; the
  whole-issue diff shows all issue changes and no foreign ones.

Related: #269 (integrate before close), #270 (active-time after integration;
measured on pair#410: ~60 min of #410 work credited to #411, which was filed
inside the segment), #197 (review boundary from ledger).

## Done when

- The fixture passes for both milestone and whole-issue closes.
- #269's spec notes the milestone-window interaction.

## Plan

- [ ]

## Log

### 2026-10-08
