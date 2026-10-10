---
id: 000304
status: working
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-09
estimate_hours: 6.27
card_mirror: 'c0bf5b24040365dafcc9fba2a720a906c40edf0e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T20:01:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
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

**Scope (2026-10-09, TL dispatch from ariadne:1, project ariadne-robustness-1 requirement 5):**
umbrella for the branch patch `diff(merge-base(main,HEAD),HEAD)`. Absorbs #197
(milestone-close commits its own evidence; the ledger stores the reviewed head),
A4 (the publish gate counts merged-in main commits), A2 (stale Review-Window
trailers) and B3's close binding (a rebase orphans the card's evidence commit, and
the publish then skips it silently). Later reviews are the interdiff since the last
reviewed patch: the reviewed head rebased onto today's main with `git merge-tree`.
Design: the durable plan (D1–D10).

Related: #269 (integrate before close), #270 (active-time after integration;
measured on pair#410: ~60 min of #410 work credited to #411, which was filed
inside the segment), #197 (review boundary from ledger).

## Done when

- The fixture passes for both milestone and whole-issue closes.
- #269's spec notes the milestone-window interaction.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec              design=1.0 impl=0.04
item: smaller-go-module       design=0.0 impl=0.08
item: smaller-go-module       design=0.1 impl=0.12
item: greenfield-go-module    design=0.5 impl=0.32
item: greenfield-go-module    design=0.3 impl=0.32
item: smaller-go-module       design=0.2 impl=0.2
item: greenfield-go-module    design=0.3 impl=0.24
item: greenfield-go-module    design=0.5 impl=0.32
item: smaller-go-module       design=0.1 impl=0.16
item: atlas-docs              design=0.1 impl=0.08
item: atlas-docs              design=0.1 impl=0.08
item: milestone-review        design=0.1 impl=0.2
item: milestone-review        design=0.1 impl=0.2
design-buffer: 0.15
total: 6.27
```

Items, in plan order:
1. design/brainstorm + TL loop;
2. gatestate `Reviewed`;
3. finalize-only stamping;
4. gitx rebased base + Close-Token lookup (git semantics: conflicts, criss-cross);
5. window planner + fresh main + printed line;
6. milestone evidence commit + pins;
7. publish gate on the rebased patch;
8. B3 ownership across publish/landing/settle + supersession;
9. pin lifecycle + sweep;
10. M1 atlas;
11. M2 atlas;
12. M1 review;
13. M2 review.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* (Calibration is flagged stale, so treat it as provisional.)

## Plan

Durable plan: `workshop/plans/000304-whole-issue-close-review-window-collapses-after-integrating-main-plan.md`
(tasks 1–7 → M1, tasks 8–12 → M2).

- [ ] M1 — review windows on the branch patch: `Round.Reviewed` stamped on finalize, rebased-reviewed-base primitive, milestone interdiff, printed window, milestone-close commits its own evidence + reviewed-head pin (#197, A2, A3)
- [ ] M2 — publish gate compares the rebased reviewed patch (A4); close binding survives a rebase and unowned codecomplete details refuse (B3); pin lifecycle end

## Log

### 2026-10-08

### 2026-10-09
- Claimed + start-plan from the TL dispatch (ariadne:1). Read the evidence pensive (Part 1 #304/#197, Part 2 §A) and the source: `boundaryWindowBase`/`previousReviewBoundary`, `judge/reviewwindow.go`, `publishgate.go` `validatePublishAnchors`, `trackercompletion.go` `ownedCompletions`, `closetracker.go` evidence commit, `gatestate.Round`.
- Found B3 fails *open*: after a rebase, `ownedCompletions` drops the card (evidence commit not an ancestor), so the publish gate reports "nothing to verify". Plan D8 makes it fail closed.
- Design: the identity is the reviewed head only, with the base derived as merge-base(main, H_r) (ARCH-DRY, no card schema change); `merge-tree --merge-base` replays the patch; a deterministic synthetic base commit, collected by gc; pin ref `refs/sdlc/reviewed/<id>` with its removal at done/abandon/archive (ARCH-FUNERAL).
- Plan file is named after the issue stem so the boundary reviewer finds it (`reviewPlanPaths`).
- TL review (ariadne:1): APPROVE WITH CHANGES (`ariadne-slot1/tl-reviews/304-plan-review.md`); all three flagged decisions approved. Plan revised in place, with a `## Revisions` entry: a Close-Token trailer replaces the patch-id (I1); one ownership rule across publish, landing and settle, with rebase→land→done e2e tests (I2); conflicts review the resolution interdiff (I3); one fresh main `reviewMainRef` (I4); per-boundary pins plus a sweep (I5); Minors folded. No TL re-review needed (same shape).
