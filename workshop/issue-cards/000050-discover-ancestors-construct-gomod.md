---
id: '000050'
status: done
created: 2026-05-30
updated: 2026-05-30
estimate_hours: 1
actual_hours: 1
---

# discover_ancestors walks only root go.mod — misses depth-2 substrate ancestors

## Problem

`construct/setup.sh`'s `discover_ancestors` (the function that decides which
upstream layers' `base.manifest` to apply into a target) walked **only each
node's root `go.mod`**. But the ariadne convention is that the substrate-ancestor
`replace` directive lives in **`construct/go.mod`**, not the root (root = the
operator's app module; `construct/go.mod` = substrate-tool deps —
see `atlas/workflow/setup-and-replication.md`).

`discover_ancestors` was the **lone walker** still reading the wrong file. The
other two substrate walkers already honor the convention:

- `construct/scripts/bootstrap-peers.sh` — reads `construct/go.mod` (only).
- `construct/scripts/list-peers.sh` — walks **both** per node
  (`_enqueue_replaces "$current"` + `_enqueue_replaces "$current/construct"`),
  with a header comment stating the convention explicitly.

### Symptom (how it surfaced)

brain → nous → ariadne. The nous→ariadne hop is declared in
`nous/construct/go.mod`. A root-only walk from **brain**:

- reads brain's root `go.mod` → finds `replace nous => ../nous` → enqueues nous
- reads **nous's root** `go.mod` → no ariadne replace (just app deps) → stops.

ariadne is never discovered, so **ariadne's entire `base.manifest` is never
applied to brain** — no `construct/setup.sh`, no `bootstrap-peers.sh`, no
`clone-data-deps.sh` (the #49 file that exposed this), etc. brain's construct
layer had been silently stale. The depth-1 case (nous itself) was masked by
Source-3's `ARIADNE_DIR` fallback, which only fires when zero ancestors are
found and only adds the direct script-parent — so it papered over the bug at
depth 1 and couldn't at depth 2.

The atlas (`setup-and-replication.md`) already *claimed* "each walker reads both
the root go.mod and construct/go.mod per node" — so this was a latent
divergence between documented intent and `discover_ancestors`' behavior, not a
design change.
