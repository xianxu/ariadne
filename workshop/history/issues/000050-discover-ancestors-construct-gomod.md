---
id: 000050
status: done
deps: []
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

## Spec

`discover_ancestors` must walk **both** `$node/go.mod` and
`$node/construct/go.mod` for every node in the BFS, matching `list-peers.sh`.
Entry guard widened so a construct-only derivative (no root go.mod) still walks.

## Plan

- [x] BFS call site parses both go.mods per node
- [x] Widen entry condition to `-f root/go.mod || -f construct/go.mod`
- [x] Add `SETUP_DISCOVER_ONLY=1` test seam (print ancestors, apply nothing)
- [x] Regression test: `construct/scripts/test/discover-ancestors.test.sh`
- [x] Verify the existing bootstrap-transitive suite still passes
- [x] Verify against the real brain → nous → ariadne chain

## Log

### 2026-05-30
Found while pushing #49's `clone-data-deps.sh` down to brain: it never arrived.
Traced to `discover_ancestors` reading only root go.mod (verified by reading the
function + `grep` showing setup.sh never references `construct/go.mod` for the
walk, while bootstrap-peers.sh and list-peers.sh do).

Fix (one call site + entry guard):
```
done < <(_parse_replace_paths "$current"; _parse_replace_paths "$current/construct")
```
plus `[[ -f "$TARGET_DIR/go.mod" || -f "$TARGET_DIR/construct/go.mod" ]]`.

Added a `SETUP_DISCOVER_ONLY=1` seam so the walk is testable without applying
(apply mutates the target). New hermetic test builds a 3-layer chain where the
deepest hop lives in the mid layer's `construct/go.mod` and asserts the leaf
discovers it — the exact shape of the brain bug.

**Verified:**
- `discover-ancestors.test.sh`: 7/7 pass (incl. depth-2 discovery, ordering,
  construct-only leaf, non-manifest filtering).
- `bootstrap-transitive.test.sh`: 11/11 still pass.
- Real chain: `../nous/construct/setup.sh` from brain now prints
  `[ariadne]` + `[nous]` (was `[nous]` only) and applies ariadne's full
  manifest — `clone-data-deps.sh`, `setup.sh`, `construct/go.mod`, etc. now
  land in brain. (brain re-provisioning committed separately in the brain repo.)
