# Boundary Review — ariadne#257 (whole-issue close)

| field | value |
|-------|-------|
| issue | 257 — lint-ids reads the cutover marker from the checked commit, not the checkout |
| repo | ariadne |
| issue file | workshop/issues/000257-lint-ids-marker-at-head.md |
| boundary | whole-issue close |
| milestone | — |
| window | 76c347e6829f88c629a3b20b7e999281dbe26257..3b08aeb8177cce72b7ede2a415816101752073ae |
| command | sdlc close --issue 257 |
| reviewer | claude |
| timestamp | 2026-09-27T23:39:56-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The fix does what the issue asks for. Before this change, lint-ids already made its tracked-or-legacy choice by looking at `--head`'s tree (`issuelintids.go:231`). But `repo.Snapshot()` then ran the cutover guard against the checkout's marker file. For the migration commit in a pre-push hook, those two answers disagree. The diff sends every marker read the guard makes through a single `Repository.readMarker` (`cutover.go:95`). That read uses a commit's tree when `markerAt` is set and the checkout otherwise, and lint-ids now guards at `headRef`. I checked the regression test with a revert: in a scratch worktree at 3b08aeb8, I swapped `GuardCutoverAt(dirs.Top, headRef).Snapshot()` back to `Snapshot()`. `TestIssueLintIDsJudgesTheMarkerAtHead` then failed with you-decide's exact "checkout has no marker … exiting 2" error. With the fix it passes, and `./internal/tracker` passes too. Both Done-when clauses are tested: the migration commit passes from a checkout without the marker, and a forged marker naming another root still refuses. The Revisions entry explains why the original negative case was changed. Only Minor findings remain.

**1. Strengths**
- All three guard sites read the marker through one reader: `checkCutover`, `checkAbsentTracker` and `Presence` (`candidates.go:205`). No site was left reading the checkout's file directly. `MarkerWithoutTracker` was split into a thin wrapper around the pure `markerWithoutTracker`, so the public API didn't change (ARCH-DRY pass).
- The test runs the real binary as a subprocess because lint-ids calls `os.Exit`. It also asserts the precondition that the checkout has no marker, so the test can't pass for a trivial reason.
- The Revisions entry is honest: it corrects a Done-when clause that was wrong about deliberate legacy handling instead of quietly editing it.
- The atlas has a precise one-sentence note in `issue-tracker-migration.md`.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- **`ReadCutoverMarkerAt` treats every `cat-file -e` failure as "absent"** (`cutover.go:104`). That includes a bad ref or a git failure. The checkout version (`ReadCutoverMarker`) is stricter: only `ErrNotExist` counts as absent there. It's safe in `checkCutover` today, because "absent" plus `markerAt` refuses. In `Presence`'s stale branch, though, a failed read would report "not marked". The same pattern exists at `issuelintids.go:231`. Fix: verify `commit^{commit}` once, then treat only a missing path as absent.
- **Duplicate presence probe (ARCH-DRY):** `issuelintids.go:231` probes `headRef:CutoverMarkerPath` right before `GuardCutoverAt` probes the same thing again. It could call `tracker.ReadCutoverMarkerAt` once and take `present` from that. The same family exists before this window: `issuemigrate.go:132-137` is a third reader of the marker from a commit's tree, via `mainView`.
- **Spec wording is slightly off.** The Spec says "without `--head` it keeps reading the checkout". But `--head` defaults to `HEAD` and lint-ids always calls `GuardCutoverAt`, so it now reads HEAD's committed tree, not the working-tree file. That matches line 231's existing behaviour, so it's fine, but the Spec should say so.
- **Untested new branch:** the `!present && r.markerAt != ""` refusal (`cutover.go:130`) has no test. Lint-ids can't reach it, because it checks presence first. It's a defensive branch whose only caller never enters it.

**5. Test coverage**
Both Done-when outcomes are covered end to end, and the fix is confirmed load-bearing by the revert. Missing: a unit test for `ReadCutoverMarkerAt`'s absent case and its parse-error case.

**6. Architecture**
- **ARCH-DRY:** pass, with the Minor note above about the parallel commit-tree marker readers.
- **ARCH-PURE:** pass. Parsing stays pure in `ParseCutoverMarker` and `markerWithoutTracker`; only the thin git read is IO.
- **ARCH-PURPOSE:** pass. I searched for other places that judge a commit that isn't checked out: `issuemigrate`'s reads already go through `mainView`, and no other checkout-marker readers remain that are called outside the guard.

**7. Plan revisions:** none needed. The plan matches the code. If you like, tighten the Spec sentence about the default `--head` in place.

```findings
findings:
  - id: new
    severity: Minor
    family: absent-vs-error-conflation
    title: |
      ReadCutoverMarkerAt treats any cat-file -e failure (bad ref, git error) as "marker absent"
    detail: |
      cutover.go:104 swallows the error; ReadCutoverMarker distinguishes ErrNotExist. Fails closed in checkCutover but Presence's stale branch would answer unmarked. Same pattern pre-exists at issuelintids.go:231. Verify commit^{commit} first, then treat only a missing path as absent.
  - id: new
    severity: Minor
    family: single-marker-reader
    title: |
      Marker presence probed twice for lint-ids, with a third commit-tree reader in issuemigrate
    detail: |
      issuelintids.go:231 cat-file -e duplicates ReadCutoverMarkerAt's probe; issuemigrate.go:132-137 is a parallel commit-tree marker reader via mainView. Route lint-ids' mode decision through tracker.ReadCutoverMarkerAt (ARCH-DRY).
  - id: new
    severity: Minor
    family: doc-claim-drift
    title: |
      Spec says lint-ids without --head reads the checkout; it reads HEAD's committed tree
    detail: |
      --head defaults to HEAD and cardlessAdditions always calls GuardCutoverAt, so a working-tree-only marker edit is not seen. Consistent with line 231, but the Spec sentence should say so.
  - id: new
    severity: Minor
    family: unexercised-branch
    title: |
      checkCutover's markerAt-absent refusal (cutover.go:130) has no test and no reaching caller
    detail: |
      lint-ids checks marker presence at head before guarding, so this branch is unreachable today; a tracker-package unit test would pin it for future GuardCutoverAt callers.
```
