---
id: 000320
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'c333f6862ed1b103471b03079a644bbc687e1683' # card fields mirrored from issue-cards; edit via sdlc
---

# Merging main with conflicts never forces a full re-close: union-merge append-only files, pass non-code resolutions, review only a code resolution

## Problem

Project requirement 5 of ariadne-robustness-1 says merging or rebasing main never forces a re-close. #304 made that hold for a clean merge, and #174 lets doc-only files beyond the reviewed patch pass. A merge that needs **conflict resolution** still forces a full re-close, even when the only conflict is an append-only workshop file.

Evidence (10-10, pair#426, ops learnings kink 50): pair#426 sat codecomplete while pair#425/#427/#429 landed. PR #228 came up CONFLICTING on `workshop/lessons.md` only: two lists of appended lessons, resolved by keeping both. `sdlc merge` then refused (`cmd/sdlc/publishgate.go` `classifyPublishDelta`, the `Conflicted` case): "main's changes conflict with the reviewed patch in workshop/lessons.md — the resolution is unreviewed." The TL had to re-run `sdlc close` (a full boundary review) before the merge went through. Shared append-only files conflict on nearly every parallel landing, so this recurs.

## Spec

Three layers, cheapest first:

1. **Stop the conflict happening.** Ship a `.gitattributes` `merge=union` entry for append-only workshop list files (at least `workshop/lessons.md`), delivered through weave so every woven repo gets it. Union keeps both sides, which is the correct resolution for these files, so no conflict and no resolution exist. Choose the file set deliberately (append-only lists only; never a file whose lines are edited in place).
2. **Don't re-review non-code conflicts.** In `classifyPublishDelta`, a `Conflicted` path with no code surface (`publishGateHasCodeSurface` false, the same test as #174's doc-only pass) passes with a loud line naming the files. Classify by file kind, not by the `workshop/` path. If any conflicted path has code surface, layer 3 applies.
3. **When code conflicts, review only the resolution.** Instead of a full re-close, review the resolution: the diff between the reviewed patch replayed onto today's main and HEAD, restricted to the conflicted code paths. The full test suite still runs (operator decision 10-10); only the review narrows. The verdict binds the new branch-patch identity, like #304's ledger. Coordinate with #183 (`--fixed-then-ship`, ariadne:4 in the same batch), which also narrows a post-close review; share the mechanism if one emerges, but don't wait on it.

## Done when

- A merge of main that conflicts only in a union-merge file produces no conflict, and `sdlc merge` passes (test: two branches appending to `workshop/lessons.md`).
- A conflict confined to non-code files passes the publish gate with a loud line naming them; a conflict in a code file does not (unit tests on `classifyPublishDelta`).
- A code-file conflict is admitted after a review of the resolution only, plus a full test run, without a full `sdlc close` re-review; the refusal text names that path as the next action.
- The pair#426 shape (codecomplete branch, main moved, lessons-only conflict) lands with no re-close, verified on a real branch.

## Plan

- [ ]

## Log

### 2026-10-10
