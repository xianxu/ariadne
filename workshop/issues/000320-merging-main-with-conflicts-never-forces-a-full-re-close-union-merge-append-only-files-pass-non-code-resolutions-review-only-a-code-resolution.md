---
id: 000320
status: codecomplete
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'ef4157321b3a1d9bc8b7a8c555e1e3676f43954b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T16:53:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.90
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

- [x] Layer 1, weave: a new `EnsureGitattributes` action writes a `# BEGIN weave generated` … `# END weave generated` block into `.gitattributes` (block-replace like `.gitignore`, authored lines kept). Entries are a fixed list single-sourced in weave: `workshop/lessons.md merge=union`. Wired in `planActions`, `Apply`, the dry-run formatter and the golden classifier. Tests: block create, idempotent, authored lines preserved.
- [x] Layer 1, sdlc: `gitx.RebasedReviewedBase` runs `merge-tree` with `--attr-source=<B_now>` so the replay honors main's merge attributes, and S stays a pure function of its inputs instead of depending on the worktree's `.gitattributes`. Git floor 2.40 → 2.42 (`--attr-source`). Test: a union-attributed file appended on both sides replays with no conflict.
- [x] Layer 2, publish gate: `classifyPublishDelta` refuses a conflict only when a conflicted path has code surface (`publishGateHasCodeSurface`). A non-code conflict falls through to the normal path classification, and the pass line names the resolved files. `publishDeltaRank` ranks only code conflicts last. Unit tests on both.
- [x] Commit `.gitattributes` in ariadne (run weave), and check lessons.md isn't edited in place by any tool (union duplicates a line edited on both sides).
- [x] Real-branch check: two branches append to `workshop/lessons.md`; `git merge` is clean and the publish gate passes.

ARCH-FUNERAL: creates nothing durable beyond one managed block in `.gitattributes`, which weave rewrites every compile, so it never grows.

## Log

### 2026-10-10
- 2026-10-10: closed — Round 3: docs-only change on top of round 2 (atlas: the git 2.42 floor). The round-2 evidence stands: targeted sdlc publish/merge/landing/gate tests ok, weave suite green (generator lifetime timing flake passes -count=5), the earlier full make test green apart from the sandbox-only processgroup ps denial.; review verdict: SHIP
- 2026-10-10: closed — make test: 12 sdlc shards 0 failed (1021 tests); packages green except processgroup TestCancellationKillsDescendants, which fails only because the sandbox denies fork/exec /bin/ps, and passes outside the sandbox (package untouched). New tests: TestClassifyPublishDelta non-code/mixed/helptext conflict rows plus a pass-line-vs-gatesig check, TestPublishDeltaRank_NonCodeConflict, TestRunPublishGate_BranchPatch pair#426 shape (hand-resolved lessons-only conflict passes; union-attributed lessons merges with no conflict and passes; code conflict still refuses), TestRebasedReviewedBase_UnionAttributeFromMain (attributes read from B_now, not the worktree), weave gitattributes block tests; go test ./cmd/weave/... green. weave compile --dry-run lists gitattr .gitattributes (1 entries); git check-attr shows lessons.md merge=union.; review verdict: SHIP
- 2026-10-10: flow upgraded quick → full — 103 added lines in code files (limit 100)
- Claimed and started from ariadne:1 (TL batch 2).
- `git merge-tree --write-tree` honors `merge=union` from the worktree `.gitattributes`, or from `--attr-source=<tree>` when given. Without either it conflicts. Verified in a scratch repo, git 2.54.
- Layer 2 done: `classifyPublishDelta` refuses only code-surface conflicts, names the non-code resolved files in the pass line, and the pass line is tested against gatesig's refusal pattern so it can't be misread as a refusal (#172).
- Layer 1 done: `RebasedReviewedBase` passes `--attr-source=<B_now>`. The replay unions lessons.md cleanly, but its line order can differ from the branch's own merge (each side puts its own lines first), so the file shows up as a doc-only delta, which passes. Git floor raised to 2.42.
- weave: `EnsureGitattributes` reuses the managed-block text and `writeManagedFile` (extracted from `writeManagedIgnore`, ARCH-DRY). `weave compile` could not finish in the sandbox (`.claude/settings.json` is write-protected), so ariadne's `.gitattributes` was written by hand, byte-identical to the block the test pins. The dry-run lists `gitattr .gitattributes (1 entries)`.
- lessons.md is sectioned and occasionally compacted, which is an in-place rewrite. Under union, a compaction that lands while someone else appends keeps both versions without a conflict, so compactions should land alone. No tool writes lessons.md.
- e2e (`TestRunPublishGate_BranchPatch`): the pair#426 shape passes the gate both hand-resolved (no attribute) and with union (no conflict at all). A code conflict still refuses with the plain re-close hint.
- Close re-review always reviews the whole branch patch (no path filter exists); layer 3 would have needed a new window, which #183 now builds.
- Close review round 1: SHIP with 4 Minors, all fixed in the same round:
  - `TestCompileEnsuresGitattributes` pins the planActions → ApplyManaged wiring.
  - The shared block errors now say "weave generated block" instead of "gitignore block".
  - The compaction caveat is in `UnionMergeAttributes`' doc comment and in atlas weave.md.
  - A non-code conflicted path whose HEAD still carries conflict markers refuses ("conflict markers remain in …"). This reuses republish.go's `hasConflictMarkers` (ARCH-DRY) and has an e2e row.

- Close review round 2: SHIP, with one new Minor fixed: the git 2.42 floor is now documented in atlas sdlc-binary.md. Finalization was refused once as stale because the ops project file changed while the review ran.

## Revisions

### 2026-10-10 — layer 3 moves to #183
Reason: TL scope change, agreed between ariadne:1 (#320) and ariadne:4 (#183). #183's `--fixed-then-ship` re-close reviews the interdiff since the last reviewed head replayed onto main. That window covers a code conflict resolution, and that path runs the full suite.
Delta: #320 does layers 1–2 only, union merge for append-only workshop files via weave and non-code conflict resolutions passing the publish gate. Spec layer 3 and the code-conflict Done-when bullet move to #183. A code conflict keeps the refusal with the plain re-close hint. Whichever of #183/#320 lands second points that hint at `sdlc close --fixed-then-ship`.
