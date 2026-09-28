---
id: '000045'
status: done
created: 2026-05-29
updated: 2026-05-29
actual_hours: 1.5
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
