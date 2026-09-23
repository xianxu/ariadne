---
id: 000230
status: open
deps: []
github_issue:
created: 2026-09-15
updated: 2026-09-15
estimate_hours:
---

# Make Weave dependency resolution worktree-aware

## Problem

During publication of parley.nvim#254, `weave compile` in a linked worktree
could not resolve `substrate ../ariadne` from `construct/deps`. The dependency
exists next to the primary checkout, but not next to the linked worktree:

```text
~/workspace/parley.nvim/../ariadne                         -> exists
~/workspace/worktree/parley.nvim/000254-.../../ariadne     -> absent
```

The first compile silently skipped the dependency, applied only four actions,
and pruned generated vocabulary. SDLC then refused the merge because the issue
schema was unavailable. Creating a sibling symlink to the real Ariadne checkout
allowed the next compile to apply99 actions and validation to pass. Worktrees
should not require this manual filesystem setup.

## Spec

Preserve relative dependency semantics with a deterministic, worktree-aware
fallback in the shared layer resolver (`pkg/layergraph/walk.go`):

1. Resolve a relative substrate path against the current layer checkout first.
2. If that target is absent and this layer is a linked Git worktree, discover
   its primary checkout using Git metadata, then resolve the same relative path
   against that checkout.
3. Use a valid fallback transparently to consumers, while reporting which
   dependency used it and the resolved location. Do not rewrite `construct/deps`
   or create sibling symlinks as part of resolution.
4. Do not search arbitrary worktrees and select the first match: multiple
   candidate checkouts must not make resolution order-dependent.
5. Absolute paths retain their existing meaning. An existing but invalid local
   target remains an error rather than being silently replaced by a fallback.
6. If the declared dependency cannot be resolved, report an actionable missing
   dependency error before compile mutates or prunes generated outputs. Handle
   unavailable primary checkouts, bare repositories and non-Git layers explicitly.

Apply the same rule when traversing transitive dependencies. Weave, vocabulary
validation and other layergraph consumers must share this behavior (ARCH-DRY).
Keep Git/filesystem discovery in the integration boundary and candidate selection
independently testable (ARCH-PURE). Primary checkout means Git's primary working
tree, not whichever worktree currently has the branch named main checked out.

## Done when

- A real linked-worktree fixture resolves `../ariadne` beside its primary
  checkout without changing the dependency declaration or adding a symlink.
- A valid worktree-local dependency takes precedence over the fallback; an
  invalid existing target is diagnosed instead of being bypassed.
- Absolute paths, symlinked paths, transitive dependencies and paths containing
  spaces behave consistently; fallback use identifies the chosen path.
- Missing dependencies or unavailable fallback anchors produce useful errors
  without deleting previously generated output.
- Weave compilation and vocabulary validation resolve the same layer graph;
  regression coverage includes the downstream merge-validation failure above.

## Plan

- [ ] Design and test primary-checkout discovery and deterministic relative-path
  resolution, including failure behavior and existing present-skip compatibility.
- [ ] Implement the shared resolver and diagnostics; add native Git-worktree
  conformance tests and compile-output preservation regressions.
- [ ] Update atlas/documentation and consumer mappings; verify and review.

## Log

### 2026-09-15

Filed at the user's request after parley.nvim#254's merge exposed missing
worktree-relative substrate resolution. The agreed direction is current-checkout
resolution first, then the same relative path from the primary checkout; no
arbitrary search across all worktrees. This task records follow-up work only.

### 2026-09-21 — dependency of the rigid couch-slot contract

The proposed couch surface is repo-first: `../pair :1` may become
`../worktree/pair-1` (or the explicit `pair-slot1`) without provisioning a
same-slot `ariadne` sibling. That makes this issue part of the slot boundary,
not incidental cleanup. Resolution must remain deterministic: prefer a valid
slot-local peer when one exists, otherwise use the primary checkout as the
fallback, and report which one was selected. Do not search arbitrary numbered
slots or silently choose a peer by recency. This is the cross-repo case the
couch pensive leaves open.

The companion provisioning path is now explicit: after entering a fresh slot,
couch should run `weave link` from the slot root to establish the main-slot
dependency targets, then `weave compile` there. This issue should provide the
resolver/conformance seam that proves the resulting graph is deterministic and
that a repeated compile is clean; couch should not grow a second dependency
linker.
