---
id: 000308
status: working
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-12
estimate_hours: 2.88
card_mirror: '756dd4b5587c77837b2594a2e673a522f315eacc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T07:19:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# issue new publishes card and details in one go

## Problem

`sdlc issue new` reserves the card on `issue-tracker`, but writes the details only locally. On a resting branch (usually :0) they stay uncommitted until someone runs `issue move-detail`/`issue publish`. That unpublished state causes repeated friction (evidence `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`):
- **C2/C3:** agents committed details straight onto resting main ("that's the normal path"), and `move-detail --help` calls the routine step an "escape hatch, consult the operator".
- **C4:** stray untracked details sat in ariadne:0 for days, and the operator assumed other slots dumped them.
- **C5:** in emma-private, `issue new` committed into the home-directory git repo.
- **B5:** cross-repo filing into a stale peer checkout hit "cutover mismatch".

Separately, **C1:** the close gate's project tick commits onto a peer's checked-out main without fetching or pushing. pair:0 diverged repeatedly, with about 8 reconciles. It's the same root cause: writing into someone's checkout instead of to origin.

## Spec

Operator design D-1 (revised) with round-3/4 decisions:
- `issue new` publishes the card **and** the details to origin in one operation, with no local working-tree write. Use the checkout-free compare-and-swap write `cmd/sdlc/internal/gitx/trunkfile.go` (#209).
- From an issue slot `:N`, the creating slot is the claimant (slot identity, #307). From a resting-branch brainstorm (e.g. :0), the issue is published **unclaimed**.
- When run on an issue branch, the identical bytes are also committed on that branch, so later edits there are ordinary 3-way changes (pairs with #274's content-based Behind).
- `move-detail` and the first-publication half of `publish` retire; `publish` keeps only later edits. This supersedes #273's remaining scope.
- **Skeleton placeholders.** Published skeletons must pass the instance-conformance gate. Either the scaffold stops seeding a bare `- [ ]` Plan (`construct/vocabulary/issue.cue:106`; the operator suggests a ticked example row), or the validator exempts not-started issues.
- **F1:** the instance-conformance gate at merge validates the branch's changed files, not main's tip (3 `--no-validate` bypasses: `p1-1004:1769`, `p3-7:19409`, `p4-6:8887`). Skeletons on main would otherwise trip every unrelated merge.
- **C1:** the close-time peer project tick writes to the peer's origin via the same trunk write, never into the peer's checked-out branch.
- Refuse when there is no origin, or when the git toplevel isn't the repo being filed into (C5).

## Done when

- `sdlc issue new` from pair:1, targeting ariadne, leaves no file in any ariadne working tree. The card and details are on origin, claimed by pair:1.
- From ariadne:0 on the resting branch, it leaves :0 clean, with the issue published unclaimed.
- `move-detail` is gone or a deprecated alias, and help text no longer mentions an escape hatch.
- Closing an issue tracked by a peer repo's project leaves the peer's checkout untouched; the tick lands on the peer's origin.
- A merge never refuses on an issue file the branch didn't change.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec               design=0.5  impl=0.05
item: smaller-go-module        design=0.15 impl=0.2
item: smaller-go-module        design=0.1  impl=0.16
item: smaller-go-module        design=0.05 impl=0.1
item: cross-cutting-refactor   design=0.1  impl=0.14
item: smaller-go-module        design=0.05 impl=0.14
item: greenfield-go-module     design=0.2  impl=0.24
item: smaller-go-module        design=0.05 impl=0.08
item: atlas-docs               design=0.05 impl=0.05
item: milestone-review         design=0.0  impl=0.14
item: milestone-review         design=0.0  impl=0.14
design-buffer: 0.15
total: 2.88
```

Items, in order:
- issue-spec: exploration, the TL question, the plan, and a plan-quality round.
- smaller-go-module ×3:
  1. `DetailPublication` + the `creation.main` stage + legacy receipts;
  2. `runIssueNew` claim from a slot + the C5 guard;
  3. status-aware section presence.
- cross-cutting-refactor: retire `move-detail` and sweep its wording.
- smaller-go-module: extract `projectCloseEdit`.
- greenfield-go-module: `publishPeerProject` + the apply path, and delete `peerwrite`.
- smaller-go-module: `--no-project`.
- atlas-docs, then two milestone reviews.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* The calibration source is flagged stale (#127).

## Plan

Durable plan: `workshop/plans/000308-issue-new-publishes-card-and-details-in-one-go-plan.md`.

- [ ] M1 — `issue new` publishes card + details to origin (creation.main stage), claims from a slot `:N`, writes no working tree; status-aware section presence for skeletons; C5 root guard; `move-detail` hidden deprecated alias; F1 regression row
- [ ] M2 — close ticks peer projects through a CAS trunk write to the peer's origin main (symlinks resolved, tl: re-checked); peer projects leave the review snapshot; `--no-project` skips the whole project update

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Absorbs evidence items C1, C2–C5, F1, and #273's remainder. Depends on #307 (slot claimant) and #274 (content-based Behind).

### 2026-10-12
Claimed at ariadne:4 (TL dispatch). Design questions went to the TL (ops:0), who chose option (a) for point 3. F1 is already fixed by #322 (`runLandingInstanceGate` diffs merge-base..head), and the tl: skip by #326; both are verified here, not rebuilt.

M1 implemented (commits 3ec07b8b..HEAD):
- **Tracker:** `creation.main` replaces `creation.detail` and shares `DetailPublication` with the transfer. A pre-#308 receipt in the retired stage is released, with `issue publish` as the next action.
- **`runIssueNew`:** writes no checkout. A numbered slot claims through `claimSetDecision` and an `OpClaim` token, and a claim failure is a warning. The detached-HEAD refusal stays, because the receipt names a branch (plan-quality Minor 4, justified).
- **C5 guard:** the plan's cwd==toplevel rule contradicted the pinned `TestRunIssueNew_FromASubdirectory…`, where a subdirectory files into its repository. It now refuses a cwd holding its own issues tree under another repository, and, in a legacy repository, a toplevel without an issues dir.
- **Section presence:** status-aware.
- **`move-detail`:** a hidden alias that keeps the legacy refusal.
- **Docs:** wording swept, and `TestNoTextSaysIssueNewWritesDetailsLocally` pins it.

Evidence: `make test` ran 12 shards. 5 tests failed: 4 were this change's fixtures (e2e/fetch/legacy alias), fixed; `TestCLISignalCancelsOwnedReviewer` was load-flaky. A targeted rerun of all 5 is green. The live check runs the real binary in the e2e slot fleet instead of filing a scratch issue on ariadne's tracker: `TestTrackerFullSlotCycle` (`issue new` from slot1 leaves the checkout untouched, slot1 holds the claim, and a competing slot2 claim refuses).

## Revisions

### 2026-10-12: scope revision at start-plan (TL decision, ops:0)
- **Point 3 dropped.** It said to commit identical bytes on an issue branch, assuming #274's content-based Behind. #274 shipped ownership only: `transferverdict.go` still judges Behind by ancestry. A separate identical-bytes commit has no common base with main's copy, so a later branch edit is refused as Behind, and git merges it as an add/add conflict (checked in a scratch repo). Now `issue new` never touches the branch or the working tree. Details reach a branch with its next main merge, and later edits go through `issue publish`.
- **F1, the tl: skip:** verification only (landed in #322 and #326).
- **C1:** peer ticks become a CAS write to the peer's origin main that re-derives from fetched bytes and refuses when the fetched copy is now `tl:`-owned. `--no-project` skips the whole project update.
- **Skeletons:** structure only while `open`; Plan and Done-when content are required from `working` on.
- **move-detail:** a hidden deprecated alias of `issue publish`. Publish keeps first publication only to finish an incomplete creation (a pre-#308 card, or a failed creation's main stage).


### 2026-10-12: M1 review round 1 (FIX-THEN-SHIP)
- **BR-1 (Important), fixed:** `NeedsSourceCheckout` still said creation needs its source checkout. My scripted edit targeted `receipt.go`, but the function lives in `store.go`, and the replace silently matched nothing. That is the lessons.md rule about asserting a scripted edit's anchor, broken here; the Python edits assert their anchors since. `TestIssueNewStoppedAtMainFinishesFromAnotherCheckout` (main refuses the details' push → reconcile from another branch finishes it) fails with the old line restored.
- **BR-2 (Important), fixed:** `TestRecoveryLegacyCreationReceiptPointsAtPublish` rewrites a stuck receipt into the pre-#308 shape. Reconcile releases it, names `issue publish`, and leaves main unchanged.
- **Minors:**
  - The uncertain-main-push adapter case is covered by the stuck-filing fixture above.
  - The legacy missing-issues-dir refusal now says what to do in a fresh repository.
  - The alias calls `guardSpineRepo` as publish does.
- **Plan deltas:**
  - Task 2's legacy test is a predicate test (`TestLegacyCreationDetailReceipt`) plus the reconcile test above.
  - Task 3's `TestIssueNewRefusesWhenCwdIsNotTheRepoRoot` became `TestIssueNewRefusesNestedIssuesTreeUnderAnotherRepo` and `TestLegacyIssueNewRefusesRepoWithoutIssuesDir`, because a cwd==toplevel rule contradicts subdirectory filing.
  - `NeedsSourceCheckout(creation)` is false.
