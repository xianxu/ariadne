---
id: 000320
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'beba17b44afd910ef04b6cc85a2884d2cb75364b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T16:53:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6cf465e4", done: "c6e95d30"}
---

# Merging main with conflicts never forces a full re-close: union-merge append-only files, pass non-code resolutions, review only a code resolution

## Problem

Project requirement 5 of ariadne-robustness-1 says merging or rebasing main never forces a re-close. #304 made that hold for a clean merge, and #174 lets doc-only files beyond the reviewed patch pass. A merge that needs **conflict resolution** still forces a full re-close, even when the only conflict is an append-only workshop file.

Evidence (10-10, pair#426, ops learnings kink 50): pair#426 sat codecomplete while pair#425/#427/#429 landed. PR #228 came up CONFLICTING on `workshop/lessons.md` only: two lists of appended lessons, resolved by keeping both. `sdlc merge` then refused (`cmd/sdlc/publishgate.go` `classifyPublishDelta`, the `Conflicted` case): "main's changes conflict with the reviewed patch in workshop/lessons.md — the resolution is unreviewed." The TL had to re-run `sdlc close` (a full boundary review) before the merge went through. Shared append-only files conflict on nearly every parallel landing, so this recurs.

## Spec

Three layers, cheapest first:

1. **Stop the conflict happening.** Ship a `.gitattributes` `merge=union` entry for append-only workshop list files (at least `workshop/lessons.md`), delivered through weave so every woven repo gets it. Union keeps both sides, which is the correct resolution for these files, so no conflict and no resolution exist. Choose the file set deliberately (append-only lists only; never a file whose lines are edited in place).
2. **Don't re-review non-code conflicts.** In `classifyPublishDelta`, a `Conflicted` path with no code surface (`publishGateHasCodeSurface` false, the same test as #174's doc-only pass) passes with a loud line naming the files. Classify by file kind, not by the `workshop/` path. If any conflicted path has code surface, layer 3 applies.
3. **(Moved to #183, see Revisions.) When code conflicts, review only the resolution.** Instead of a full re-close, review the resolution: the diff between the reviewed patch replayed onto today's main and HEAD, restricted to the conflicted code paths. The full test suite still runs (operator decision 10-10); only the review narrows. The verdict binds the new branch-patch identity, like #304's ledger. Coordinate with #183 (`--fixed-then-ship`, ariadne:4 in the same batch), which also narrows a post-close review; share the mechanism if one emerges, but don't wait on it.

## Done when

- A merge of main that conflicts only in a union-merge file produces no conflict, and `sdlc merge` passes (test: two branches appending to `workshop/lessons.md`).
- A conflict confined to non-code files passes the publish gate with a loud line naming them; a conflict in a code file does not (unit tests on `classifyPublishDelta`).
- ~~A code-file conflict is admitted after a review of the resolution only, plus a full test run, without a full `sdlc close` re-review; the refusal text names that path as the next action.~~ Moved to #183 (see Revisions). A code-file conflict keeps the refusal with the plain re-close hint.
- The pair#426 shape (codecomplete branch, main moved, lessons-only conflict) lands with no re-close, verified on a real branch.

## Plan

Quick flow (layers 1–2 only after the revision).

- [ ] Layer 1, weave: a new `EnsureGitattributes` action writes a `# BEGIN weave generated` … `# END weave generated` block into `.gitattributes` (block-replace like `.gitignore`, authored lines kept). Entries are a fixed list single-sourced in weave: `workshop/lessons.md merge=union`. Wired in `planActions`, `Apply`, the dry-run formatter and the golden classifier. Tests: block create, idempotent, authored lines preserved.
- [ ] Layer 1, sdlc: `gitx.RebasedReviewedBase` runs `merge-tree` with `--attr-source=<B_now>` so the replay honors main's merge attributes, and S stays a pure function of its inputs instead of depending on the worktree's `.gitattributes`. Git floor 2.40 → 2.42 (`--attr-source`). Test: a union-attributed file appended on both sides replays with no conflict.
- [ ] Layer 2, publish gate: `classifyPublishDelta` refuses a conflict only when a conflicted path has code surface (`publishGateHasCodeSurface`). A non-code conflict falls through to the normal path classification, and the pass line names the resolved files. `publishDeltaRank` ranks only code conflicts last. Unit tests on both.
- [ ] Commit `.gitattributes` in ariadne (run weave), and check lessons.md isn't edited in place by any tool (union duplicates a line edited on both sides).
- [ ] Real-branch check: two branches append to `workshop/lessons.md`; `git merge` is clean and the publish gate passes.

ARCH-FUNERAL: creates nothing durable beyond one managed block in `.gitattributes`, which weave rewrites every compile, so it never grows.

## Log

### 2026-10-10
- Claimed and started from ariadne:1 (TL batch 2).
- `git merge-tree --write-tree` honors `merge=union` from the worktree `.gitattributes`, or from `--attr-source=<tree>` when given. Without either it conflicts. Verified in a scratch repo, git 2.54.
- Close re-review always reviews the whole branch patch (no path filter exists); layer 3 would have needed a new window, which #183 now builds.

## Revisions

### 2026-10-10 — layer 3 moves to #183
Reason: TL scope change, agreed between ariadne:1 (#320) and ariadne:4 (#183). #183's `--fixed-then-ship` re-close reviews the interdiff since the last reviewed head replayed onto main. That window covers a code conflict resolution, and that path runs the full suite.
Delta: #320 does layers 1–2 only, union merge for append-only workshop files via weave and non-code conflict resolutions passing the publish gate. Spec layer 3 and the code-conflict Done-when bullet move to #183. A code conflict keeps the refusal with the plain re-close hint. Whichever of #183/#320 lands second points that hint at `sdlc close --fixed-then-ship`.
