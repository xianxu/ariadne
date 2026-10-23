---
id: 000566
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0ae300b5d4c6bb0173685bbcf3e960712e89baa6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T03:16:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Issue filing builds over-long file names (331-byte basename broke parley checkouts); cap slugs, refuse oversized paths, add a rename

## Problem

## Spec

`sdlc flake register` (and anything else that files an issue from a generated title) builds the details/card file
name from the full title. On 10-22 it filed parley.nvim#421 from a title made of two long Lua test names; the
basename was 331 bytes, over macOS's 255-byte limit, and was pushed to `main` and `issue-tracker`, so every checkout
of parley main would fail with "File name too long". The TL renamed both files by hand with a plumbing commit
(main `26b27da7`). `sdlc issue set-title` keeps paths, and there's no rename verb.

Fix: cap the generated slug (e.g. to 80 bytes at a word boundary, keeping the ID prefix), in the one slug builder
every filer uses; refuse to publish any path whose basename exceeds the filesystem limit; and give `sdlc issue` a
supported rename for an existing card/details pair (both refs, content unchanged).

## Done when

- Filing from a 300-character title produces a basename within the cap (test), for both details and card.
- A path over the limit is refused before any push (test).
- `sdlc issue rename --issue N --slug S` renames both files on both refs (test).

## Plan

Durable plan: `workshop/plans/000566-issue-filing-builds-over-long-file-names-331-byte-basename-broke-parley-checkouts-cap-slugs-refuse-oversized-paths-add-a-rename-plan.md`.

- [x] FileSlug: one capped slug builder (80 bytes, word boundary) for issue new (tracker + legacy) and flake register
- [x] CheckPathLimits in gitx commitSet: refuse any component over 255 bytes before a candidate exists
- [x] NextRenameStep state machine + MovePath
- [x] `sdlc issue rename --issue N [--slug S]`, help + atlas
- [x] Verification: targeted tests, fresh-reader help check

Release gates (binary-owned, not plan items): the recorded suite on the pushed head precedes landing; `sdlc merge` enforces it.

## Log

### 2026-10-23
- 2026-10-23: closed — Round 3. BR-5: the card move pins the bytes its status/ownership checks read; a card changed in the gap refuses and the rerun re-checks (TestIssueRenameRefusesACardChangedAfterItsChecks, RED with the pin removed). Prior: FileSlug filer tests incl. legacy (RED with cap lifted), CheckPathLimits in commitSet and PublishCommitChecked (RED without, origin unchanged), NextRenameStep over 16 observations, rename move/derive/resume/no-op and seven refusals moving no ref (incl. a non-owner finishing a half-done rename). Recovery catalog proven; fresh-reader help check passed. internal/issue, gitx, tracker, recovery and cmd/sdlc rename/filer/legacy tests pass. Recorded suite lease on 1f64a0aa queued and must pass before landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 362 added lines in code files (limit 100); an earlier round of this close already ran the full review

Close round 2 FIX-THEN-SHIP: BR-1 to BR-4 disposed addressed. BR-5 (Important, family guard-observed-outside-cas): the permission and status checks read one snapshot while MoveCard moved whatever the fresh tip held, so a claim in the gap would be renamed unchecked. The card move now pins the checked bytes and refuses a change; test lands a status change in the gap through a renameBeforeStep seam, RED with the pin removed, GREEN now, and the rerun refuses the now-started issue. Class sweep: the details step is the only other write after a check; it stays unpinned because ownership is on the card, already renamed, and the move copies the details whole. Branch checks race start-plan only through the card status, which the pin covers.

Close round 1 FIX-THEN-SHIP: BR-1 (Important) rename skipped the permission check whenever the card already sat at the new name, so a non-owner could finish a half-done rename's details step. The check now runs before any moving step; new refusal case RED with the check limited to the card step, GREEN now. BR-2 (Important) the legacy filer had no cap test; added, RED with Slugify restored (file name too long), GREEN now. Minor: PublishCommitChecked merged trees bypassed commitSet; it now runs CheckPathLimits over the diff of base and merged tree (RED: the 300-byte name published; GREEN refuses with origin unchanged); the stateful git fake models that `diff`. Minor: plan Core concepts gains DetailNames, and Revisions record the dropped plans precondition. Cancelling suite lease-3247735814 (queued, never started) because the fixes supersede b672c598; requeueing on the corrected head.

Fresh-reader help check (no context, help text only) answered all five scenarios correctly (title-derived fix of a #421 name, rerun after a kill, refusal for a started issue, bad slug, unchanged checkout); its two ambiguities (refused vs corrected slug, which command fast-forwards) are now stated in the help. Targeted internal/issue, gitx, tracker and recovery pass; cmd/sdlc rename/filer/catalog tests pass individually (a broad -run pattern hit go test's 10m default timeout, not a failure). Recorded suite lease-3247735814 queued on Git-verified b672c59894837b0969700bc25daf9ea4b0a0df8f; close review overlaps its wait per tl-worker TESTS. Landing still requires its recorded pass.

TL approved the plan and all four scope choices (9e66382f); change-code inferred the quick flow, and close upgrades to the full review because the rename verb leaves the 100-line shell. Implemented FileSlug (RED/GREEN: the filer tests failed with the cap lifted only after their oracle became a literal 80, since a MaxSlugBytes-based bound was circular), CheckPathLimits in commitSet (RED: with the call removed, the 331-byte UpdateMany pushed), NextRenameStep/DetailNames/MovePath/MoveCard and `sdlc issue rename`. The resume test exposed a real defect: the rerun took the old details name from the card, which the card step had already moved. Details are now found on main by ID (`DetailNames`). A squatted new name is caught as an inconsistent observation before either ref moves. Deviation from plan: the "no plans on main" precondition was dropped; status open plus no local/remote branch covers the unstarted case. Recovery catalog entry, `issue` help and atlas row added.

Claimed and planned in ariadne:5. All filers reach one slug site per mode (`runIssueNew`, legacy `issue new`; `flake register` calls `runIssueNew`). Every tracker/main candidate tree is built by `gitx.TrunkFile.commitSet`, so the 255-byte refusal sits there (closetracker re-indexes existing paths; refbootstrap writes fixed names). `Slugify` stays uncapped because gate family slugs use it. Branch, plan and sidecar names derive from the stem, so rename is scoped to unstarted issues; four scope decisions sent to TL with the plan. No estimate yet (full flow: after plan-quality).

### 2026-10-22

## Revisions

2026-10-23: Split the recorded suite out of the Verification plan item into the binary-owned release gates; no scope change.
