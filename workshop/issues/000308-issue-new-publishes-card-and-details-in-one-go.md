---
id: 000308
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '04acfdcb4f6e8be18cfd87711170f5c010933926' # card fields mirrored from issue-cards; edit via sdlc
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

## Plan

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). Absorbs evidence items C1, C2–C5, F1, and #273's remainder. Depends on #307 (slot claimant) and #274 (content-based Behind).
