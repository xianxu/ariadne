---
id: '000044'
status: done
created: 2026-05-29
updated: 2026-05-29
actual_hours: 4
---

# openshell sandbox: sync go.mod peers (not the whole workspace), mirror host layout, SYNC= flag

## Problem

The openshell sandbox (`.openshell/sandbox.sh`) is the laggard of the base
layer's peer-dependency mechanisms. `make tart` was brought onto the
`construct/go.mod` peer model in #32 phase 2 (selective clone of current repo +
transitive `replace` peers) and corrected in #41. The sandbox still uses the
**old "sync the whole parent workspace" model**:

- `ensure_mutagen_sync` one-way-syncs all of `WORKSPACE_DIR` (the entire
  `~/workspace/`) into `/sandbox/workspace`, then symlinks *every* sibling dir
  (`for dir in WORKSPACE_DIR/*/`) — pulling in repos that have nothing to do
  with the current repo's dependency graph.
- The main repo syncs to `/sandbox/repo`, while `overlay/setup.sh:182` makes
  `~/repo` (`/home/sandbox/repo`) — a different directory. The login layout and
  the sync target disagree, and neither matches the host's `~/workspace/<repo>`
  shape that tart already adopted.

This is the same inconsistency #41 fixed for tart, one level up: the sandbox
isn't go.mod-aware at all.
