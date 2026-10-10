---
id: 000304
status: codecomplete
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-09
estimate_hours: 6.27
card_mirror: '21281d1277d02e7c9e3314395804b41b249ad511' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T20:01:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 2.23
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

- [x] M1 — review windows on the branch patch: `Round.Reviewed` stamped on finalize, rebased-reviewed-base primitive, milestone interdiff, printed window, milestone-close commits its own evidence + reviewed-head pin (#197, A2, A3)
- [x] M2 — publish gate compares the rebased reviewed patch (A4); close binding survives a rebase and unowned codecomplete details refuse (B3); pin lifecycle end

## Log

### 2026-10-08

### 2026-10-09
- 2026-10-09: closed — Re-close after fixing the three advisory close-review Minors (0f4e1bf5): unusable ledger facts become a named fallback (TestPlanReviewWindow, TestMilestoneWindow_RejectsANonSHAReviewedValue); legacy archives unpin exactly the moved issues, and another slot pin survives (TestLegacyArchivesEndPins); the done and settle sites warn on pin failures. Full sharded suite green: 1014 cmd/sdlc tests + 20 packages, 0 failed. Done-when unchanged and satisfied (see earlier close line).; review verdict: SHIP
- 2026-10-09: closed — Done-when: TestMilestoneWindow_ExcludesIntegratedMain passes (merge/rebase/conflict; milestone window = issue changes only, whole-issue window = all issue changes, none of main). #269 spec note published by the TL (1b2f56b1). M1 closed SHIP (binary evidence commit 1b01de5b). M2: TestRunPublishGate_BranchPatch (A4: merge/rebase pass, own code refuses naming only it, conflict names paths), TestRebasedCloseIsStillOwnedAndGated / LandsToDone (merge, squash, settle) / SquashedCloseRefusesLoudly / OlderCloseTokenNeverClaimsANewerClose (B3), pin end-site tests incl. TestCompleteOnCardEndsPins; each mutation-checked. Full sharded suite (1014 cmd/sdlc tests, 20 packages): only 2 failures, both fixed in c953b962 (stale message assertion, CLIRef hint) and rerun green. M2 is a trailing milestone covered by this review (#175).; review verdict: SHIP
- 2026-10-09: closed M1 — Round-2 fixes for BR-2/BR-3 + Minors (1f37e2ea). Full sharded suite: 1004 cmd/sdlc tests + 20 packages green except TestPlanningReviewConcurrencySchedules, a load-timing flake that passes alone (planning review untouched by #304). New tests: NotRunEvidenceIsNotABoundary, CrissCrossIsANamedFallback, FIX-THEN-SHIP milestone evidence, legacy stamp without commit; clean mutations fail each. Actual = measured 1.01h (M1 is all work so far). --no-project: the project tracks issues; its file is TL-owned.; review verdict: SHIP
- Claimed + start-plan from the TL dispatch (ariadne:1). Read the evidence pensive (Part 1 #304/#197, Part 2 §A) and the source: `boundaryWindowBase`/`previousReviewBoundary`, `judge/reviewwindow.go`, `publishgate.go` `validatePublishAnchors`, `trackercompletion.go` `ownedCompletions`, `closetracker.go` evidence commit, `gatestate.Round`.
- Found B3 fails *open*: after a rebase, `ownedCompletions` drops the card (evidence commit not an ancestor), so the publish gate reports "nothing to verify". Plan D8 makes it fail closed.
- Design: the identity is the reviewed head only, with the base derived as merge-base(main, H_r) (ARCH-DRY, no card schema change); `merge-tree --merge-base` replays the patch; a deterministic synthetic base commit, collected by gc; pin ref `refs/sdlc/reviewed/<id>` with its removal at done/abandon/archive (ARCH-FUNERAL).
- Plan file is named after the issue stem so the boundary reviewer finds it (`reviewPlanPaths`).
- TL review (ariadne:1): APPROVE WITH CHANGES (`ariadne-slot1/tl-reviews/304-plan-review.md`); all three flagged decisions approved. Plan revised in place, with a `## Revisions` entry: a Close-Token trailer replaces the patch-id (I1); one ownership rule across publish, landing and settle, with rebase→land→done e2e tests (I2); conflicts review the resolution interdiff (I3); one fresh main `reviewMainRef` (I4); per-boundary pins plus a sweep (I5); Minors folded. No TL re-review needed (same shape).
- change-code passed (plan-quality: 3 rounds; round 1 lost to a sandbox ERR_PROXY_TUNNEL, which is #300 D2 evidence, reported to the TL). Estimate 6.27h (estimate-quality INFO).
- M1 progress: Tasks 1–3 committed (`Round.Reviewed`, finalize-only stamp, `gitx.RebasedReviewedBase` + `CommitsWithCloseToken`; the subagent added mutation case (j)). Tasks 4–5 (window planner, fresh main, printed line, trailer) and Task 6 (pins, milestone evidence) in progress.
- **TL instruction (10-09):** after #304 lands, do NOT start #183. Report to ariadne:1 and wait; the TL restarts the slot before the next batch.
- M1 Tasks 4–6 committed (`reviewwindowplan.go`, `reviewpin.go`, `commitMilestoneEvidence`; help text and atlas updated). Mutations checked: interdiff base, fresh main, ledger block, evidence wiring, CAS.
- Proposed #269 spec line (Done-when; asked TL ariadne:1 how to land it, since #269 is unclaimed): "Interaction with #304: milestone review windows are the interdiff since the last finalized review, replayed onto today's main, so integrating main (merge or rebase) no longer widens a milestone window, and a conflict shows only its resolution. The freshness rule here can therefore integrate main daily without resetting review cost."
- Done-when "#269's spec notes the milestone-window interaction": satisfied. The TL published the line to main (`1b2f56b1`, option A); #269 is unclaimed again.
- Done-when "fixture passes for milestone and whole-issue closes": `TestMilestoneWindow_ExcludesIntegratedMain` (merge, rebase and conflict sub-tests) passes.
- M1 boundary review round 1: FIX-THEN-SHIP, with BR-2 and BR-3 (Important) blocking.
  - BR-2 fixed: the legacy trailer grep counts only finalizing verdicts. Test `TestMilestoneWindow_NotRunEvidenceIsNotABoundary`; a clean mutation fails it.
  - BR-3 fixed: `SoleMergeBase` is shared, and criss-cross gives a named fallback. Test `TestReviewWindow_CrissCrossIsANamedFallback`; mutation-checked.
  - Minors fixed in the same round: stale-main note, `reviewed:` SHA validation, help text.
  - Tests added: FIX-THEN-SHIP milestone evidence (Task 6c); legacy no-commit stamp + pin (6d).
  - Lesson added: re-read readers when a verb starts producing an artifact.
- Process note: milestone-close needed `--actual` (measured 1.01h) and `--no-project` (the project tracks issues; TL-owned file).
- **M1 closed** (round 2 SHIP; the binary's own evidence commit `1b01de5b`, pushed). Advisory BR-4/BR-5 now have tests (`NamesAStaleMainWhenTheFetchFails`, `RejectsANonSHAReviewedValue`), both mutation-checked. Starting M2.
- M2 Tasks 8–12 committed (`e2ebcb09`, `1ca3c2b3`). The publish gate replays the reviewed patch; the Close-Token binding works across publish, landing and settle; pins end with the issue; atlas and lessons updated. Mutation-checked: replay, token rule, unowned refusal, and each of the five pin end sites. The done-site unpin was masked by the settle sweep until `TestCompleteOnCardEndsPins` drove it alone. Found and fixed: the main-ref cache went stale after an in-process fetch.
- **Closed** (SHIP; evidence `792a583b`; actual 2.08h vs estimate 6.27, ratio 3.0×). Its three advisory Minors are fixed in the same round:
  - a rejected or unreadable ledger fact is now a named fallback, not a pre-#304 label (fixing the class);
  - the legacy archives unpin exactly what they moved, never by local liveness of shared refs;
  - the done and settle sites warn on pin failures.
  These are post-close code changes, so a re-close follows.
- **Re-closed** (SHIP; evidence `5890aae2`; actual 2.23h). Round-4 advisory Minor: legacy repositories have no removal path for a hand-archived issue's pins. It is recorded as a named known limit in `atlas/workflow/pre-merge-checks.md` rather than fixed, because another fix means another full re-close. Each round has surfaced a new Minor, and that loop is the cost #183's interdiff re-review exists to cut. Flagged to the TL as a candidate for #183 or a follow-up.
