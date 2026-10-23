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
flow: {kind: quick, provenance: inferred, spec: "3cf180a1", done: "f6144709"}
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

- [ ] FileSlug: one capped slug builder (80 bytes, word boundary) for issue new (tracker + legacy) and flake register
- [ ] CheckPathLimits in gitx commitSet: refuse any component over 255 bytes before a candidate exists
- [ ] NextRenameStep state machine + MovePath
- [ ] `sdlc issue rename --issue N [--slug S]`, help + atlas
- [ ] Verification: targeted tests, fresh-reader help check, recorded suite

## Log

### 2026-10-23

Claimed and planned in ariadne:5. All filers reach one slug site per mode (`runIssueNew`, legacy `issue new`; `flake register` calls `runIssueNew`). Every tracker/main candidate tree is built by `gitx.TrunkFile.commitSet`, so the 255-byte refusal sits there (closetracker re-indexes existing paths; refbootstrap writes fixed names). `Slugify` stays uncapped because gate family slugs use it. Branch, plan and sidecar names derive from the stem, so rename is scoped to unstarted issues; four scope decisions sent to TL with the plan. No estimate yet (full flow: after plan-quality).

### 2026-10-22
