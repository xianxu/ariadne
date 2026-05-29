---
id: 000045
status: working
deps: [000042]
created: 2026-05-29
updated: 2026-05-29
estimate_hours:
actual_hours:
---

# bootstrap.sh is direct-only — 3-deep go.mod chains fail at the `make` handoff

## Problem

`bootstrap.sh` (the fresh-clone entrypoint from #42) clones **only the direct
peers** declared in the current repo's `construct/go.mod` — a single pass over
its own go.mod, no recursion — then `exec make bootstrap` to delegate the
transitive cascade to `bootstrap-peers.sh`.

That works for a **2-deep** chain (the common case: `pair → ariadne`,
`nous → ariadne`), because the direct peer *is* the target of the repo's
Makefile symlink. It **breaks for a 3-deep chain**, because `make` can't even
read its own Makefile until the *entire* substrate chain is present.

### Trace: `foo → nous → ariadne`

`foo/construct/go.mod` declares only its direct upstream (`replace nous =>
../../nous`) — matching the established convention (nous declares only ariadne;
pair declares only ariadne; no repo carries a flattened transitive replace set).

| step | state |
|---|---|
| `./bootstrap.sh` in `foo` | reads `foo/construct/go.mod`, clones **nous** only. ariadne not cloned. |
| `exec make bootstrap` in `foo` | `make` reads `foo/Makefile`, a symlink → `../nous/Makefile` → `../ariadne/Makefile`. **ariadne absent** → chain dangles → `make: Makefile: No such file or directory`. |
| `bootstrap-peers.sh` (the recursive cloner that *would* fetch ariadne) | **never runs** — the handoff died before `make` could start. |

The transitive cloner is unreachable because reaching it requires the substrate
it was supposed to fetch. Chicken-and-egg.

### Why it's latent, not live

No real 3-deep **go.mod** chain exists today. `brain` is `brain → nous →
ariadne` at the *Makefile-symlink* level (`brain/Makefile → ../nous/Makefile →
../ariadne/Makefile`), but `brain` has **no `construct/go.mod` and no
`bootstrap.sh`** (it's a brain repo — `.brain/config.md` — bootstrapped by a
different path). Every repo that uses the go.mod bootstrap (nous, pair,
parley.nvim, you-decide) is exactly 2-deep. The gap will surface the first time
someone fresh-clones a derivative-of-a-derivative.

## Spec

`bootstrap.sh` must guarantee the **entire transitive substrate tree** is on
disk before handing off, so the top repo's Makefile symlink chain resolves.

### Mechanism: transitive clone-walk, single handoff

Two phases, not interleaved (the naïve "cd into each cloned peer and re-run its
`bootstrap.sh`" is wrong — `bootstrap.sh` ends in `exec make bootstrap`, so the
deepest peer's `exec` replaces the process and control never returns to the top
repo; it would bootstrap the substrate root and orphan the leaf):

1. **Clone phase — in-process BFS (clone-absent).** Clone each direct peer,
   then read *that peer's* `construct/go.mod` and continue the BFS, cloning each
   transitive peer. Use an in-process visited-set (like `list-peers.sh`'s
   `_is_seen`) for cycle safety; cap depth (mirror `bootstrap-peers.sh`'s
   `MAX_DEPTH=5`). Reuse the existing origin-URL derivation
   (`${origin_url//$this_repo/$name}`) and `PEER_URL_<name>` override hook.
2. **Single handoff.** One `exec make bootstrap` in the top repo. With the tree
   present, `bootstrap-peers.sh` cascades refresh/tools per peer (its clone
   steps become idempotent no-ops).

Prefer in-process BFS over recursing into peers' `bootstrap.sh`: the BFS depends
only on each peer's `construct/go.mod` (which survives a bare clone by
definition), not on each peer shipping a committed `bootstrap.sh` (a recent
`seed`; older peers may lack it).

### Shared parser (coordinates with #44)

The `replace <module> => ../<path>` line parser is now duplicated across
`bootstrap.sh`, `bootstrap-peers.sh`, `tart-list-peers.sh`, and
`setup.sh:discover_ancestors`. #44 M1 extracts the present-peers BFS into
`construct/scripts/list-peers.sh`. This issue shares the **parser primitive**
with that work, but keeps a distinct **walk driver**: `list-peers.sh` *skips*
absent dirs; bootstrap *resolves-syntactically-and-clones* them. One parser, two
drivers (clone-absent vs list-present). A drift test asserts both extract the
same replace targets from a fixture `go.mod`.

### Acceptance

- A synthetic 3-deep fixture (`foo → mid → ariadne`, each declaring only its
  direct upstream) cold-starts from a bare `foo` clone: `./bootstrap.sh` clones
  **both** `mid` and `ariadne`, then `make bootstrap` runs cleanly.
- 2-deep behavior unchanged (pair/nous/you-decide still bootstrap as today).
- Cycle safety: a fixture with a replace cycle terminates (visited-set), and a
  chain deeper than the cap errors with a clear message (not infinite clone).
- No-`construct/go.mod` consumers still degrade to "hand off to make" (current
  behavior).
- `origin`-less repo still errors with the manual-clone instruction.
- Drift test: `bootstrap.sh`'s parser and `list-peers.sh` agree on a fixture.
- `bash -n` clean.

## Plan

> Draft — awaiting operator approval before implementation.

- [ ] **M1 — shared parser primitive.** Coordinate with #44 M1: factor the
  `replace => ../path` regex/extractor so `bootstrap.sh` and `list-peers.sh`
  share it (or, if sharing a *file* is blocked by bootstrap's zero-substrate
  constraint, lock them with the drift test and document why bootstrap keeps a
  copy). Decide and record which.
- [ ] **M2 — transitive clone-walk.** Rewrite `bootstrap.sh`'s clone phase as an
  in-process BFS with visited-set + depth cap; keep the single trailing
  `exec make bootstrap`. Preserve URL derivation + `PEER_URL_<name>` override.
- [ ] **M3 — fixtures + tests.** 3-deep cold-start fixture; cycle fixture; depth
  fixture; 2-deep regression; drift test.
- [ ] Post-milestone review per AGENTS.md §3; update `atlas/` (bootstrap/peer
  model) noting bootstrap is now depth-independent; log outcome.

## Notes

- Base-layer + `seed`-delivered file (`bootstrap.sh`) — changes reach
  derivatives only on their next `seed` refresh (write-once; already-seeded
  derivatives keep their copy unless re-seeded). Call out the propagation story
  in the close note; this is the one base-layer file that does NOT auto-propagate
  via symlink, so the rollout has a manual edge.
- Relationship to `bootstrap-peers.sh`: that script remains the transitive
  refresh/tools cascade and the "add a peer later" path; its clone steps become
  idempotent once `bootstrap.sh` pre-clones the tree. No removal — overlap is
  intentional and idempotent.
- Discovered while designing #44 (openshell sandbox go.mod sync) — the
  "is bootstrap's walk transitive?" question surfaced the gap. #44 depends on
  this issue's parser extraction.

## Log

- 2026-05-29: opened. Split from #44 design discussion per operator. Gap
  confirmed by trace against the real `brain → nous → ariadne` symlink chain
  (brain itself is non-go.mod, so the gap is latent until a 3-deep go.mod
  derivative is fresh-cloned).
