---
id: 000230
status: working
deps: []
github_issue:
created: 2026-09-15
updated: 2026-10-12
estimate_hours:
card_mirror: 'f699a4b4604203f5e37fdde3a525b377c023ed21' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T18:45:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9eb8f023", done: "49129e11"}
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

Durable plan: `workshop/plans/000230-worktree-dependency-resolution-plan.md`
(single pass, one close boundary).

- [x] Pure `ChooseSubstrate` chooser with table tests (plan Task 1).
- [x] `workspace.DependencyAnchor` Git boundary: linked worktree, primary, subdirectory,
  non-Git, bare, numbered-environment host, spaces/symlink (Task 2).
- [x] `layergraph.Resolver` + walk on it; real-worktree tests incl. precedence, invalid
  local, transitive, absolute, spaces, no-anchor present-skip, vocabulary regression (Task 3).
- [x] `acquire.Restore` on the resolver: no duplicate clone, fallback notice, anchor
  reason in missing errors, compile output preserved on failure (Task 4).
- [x] sdlc `substrateWalk` on the resolver (Task 5).
- [ ] Atlas/comments/help sweep; full suite; close (Task 6).

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

### 2026-10-12 — design (ariadne:3)

Measured before designing: compile now runs through `acquire.Restore` (#296), so in a
linked worktree it no longer prunes silently — it fails ("record its source") for a
sourceless row and would clone a duplicate beside the worktree for a sourced one.
`layergraph.Walk` consumers (vocabulary, datatype) still present-skip. Plan resolves all
three resolvers (layergraph walk, Restore, sdlc substrateWalk) through one chooser
(ARCH-DRY) with the Git lookup at one boundary, run lazily (ARCH-PURE).

Out of scope, by decision: `data` rows (weave provisions them as clones; a worktree-local
clone is redundant, not wrong) and the bash bootstrap parser `lib-deps.sh` (pre-weave).

### 2026-10-12 — paused on quota hold (ariadne:3)

Tasks 1–3 done and committed (layergraph tests mutation-checked: disabling the anchor
reds six walk tests). Task 4 in progress, committed as WIP: `Restore` resolves relative
substrate rows through `layergraph.Resolver`, records `Result.Fallbacks`, and refuses a
row absent beside both the worktree and the primary (`FallbackAbsentError`) instead of
cloning a duplicate beside the worktree — a refinement of the plan's Task 4 (spec 6).
Remaining for Task 4: `restoreSetup` fallback notice, acquire + weave compile tests;
then Task 5 (sdlc `substrateWalk`) and Task 6 (atlas sweep, full suite, close).

### 2026-10-13 — resumed (ariadne:3)

Merged origin/main (2707b66a, #265 landed; no conflicts). Task 4 finished: compile
prints one notice per fallback; acquire tests (sourceless and sourced fallback, local
wins, absent-everywhere refuses in real and dry runs without creating either path) and
weave compile tests (fallback composes base prose and reports it; a refused compile
leaves the worktree byte-identical). Task 5: sdlc `substrateWalk` resolves through
`layergraph.Resolver`. Every new test file was mutation-checked by disabling the anchor:
all fallback tests went red. Fixture note: `cmd/weave`'s `setupGit` does not isolate the
global excludes file, so the worktree fixture uses `git add -f`. Atlas: weave.md key
decision and a *Linked worktrees* section in setup-and-replication.md.

## Revisions

- 2026-10-12 — the 2026-09-21 entry says a slot falls back to the primary checkout when
  no slot-local peer exists. Since then numbered environments restrict dependencies to
  the environment and recover sources slot-locally (#296). So: a vantage in a numbered
  environment whose primary lies outside it gets no fallback (slot-local provisioning
  stays); a feature worktree of an environment's own dependency clone still falls back,
  inside the environment.
- 2026-10-13 — plan Task 4 refinement: a relative row absent beside both the linked
  worktree and its primary refuses (`layergraph.FallbackAbsentError`) in real and dry
  runs, instead of cloning a duplicate beside the worktree (spec 6: an actionable
  error before compile writes).
