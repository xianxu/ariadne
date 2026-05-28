---
id: 000041
status: working
deps: [000032]
created: 2026-05-28
updated: 2026-05-28
estimate_hours:
actual_hours:
---

# tart peer-walk reads the wrong go.mod (root, not construct/) — skips substrate peers

## Problem

`make tart` is supposed to APFS-clone the current repo **plus every transitive
peer declared via `replace … => ../path`** into the VM (issue #32 phase 2). In
ariadne-derivative repos it doesn't: the VM gets only the current repo, and the
ariadne substrate peer is missing. `setup.sh` resolutions then dangle in the VM
because the upstream it `replace`s isn't present.

### Root cause

The clone set comes from `.tart/scripts/tart-list-peers.sh "$(CURDIR)"`, where
`$(CURDIR)` is the **repo root**. The script seeds its BFS at `$repo` and reads
`$repo/go.mod` for replace directives.

But in a derivative repo the substrate dependency is **not** declared in the
root module. The root `go.mod` is a near-empty module of the repo's own code;
the ariadne `replace` lives one level down in **`construct/go.mod`**. Concrete
case (`you-decide`):

- `you-decide/go.mod` → `module github.com/xianxu/you-decide`, **zero replace directives**
- `you-decide/construct/go.mod` →
  ```
  module local.construct/you-decide
  require github.com/xianxu/ariadne …
  replace github.com/xianxu/ariadne => ../../ariadne
  tool github.com/xianxu/ariadne/cmd/sdlc
  ```

So the walk from the root finds nothing and clones `you-decide` alone. ariadne
never enters the VM.

| walk seeded at | peers found |
|---|---|
| `you-decide/` (what `make tart` does today) | `you-decide` only |
| `you-decide/construct/` | `construct` + `ariadne` |

### Why this is an inconsistency, not a config error

The other substrate resolvers in the base layer already know the replace lives
in `construct/`:

- **`construct/scripts/bootstrap-peers.sh`** (canonical peer-clone cascade) —
  `CONSTRUCT_GOMOD="$TARGET_DIR/construct/go.mod"`, resolves replace paths
  relative to `$TARGET_DIR/construct`.
- **`Makefile.workflow` `build:`** — `if [ -f construct/go.mod ]; then cd
  construct && go build … github.com/xianxu/ariadne/cmd/sdlc` (issue #32),
  precisely because Go resolves ariadne through `construct/go.mod`.

`tart-list-peers.sh` is the odd one out. Its header even claims parity ("Same
parser shape as construct/setup.sh's discover_ancestors so peers tracked by the
VM clone match peers walked by setup.sh's manifest resolution") — but it points
at `$repo/go.mod` while the real substrate replace is in `$repo/construct/go.mod`.
This is fallout from issue #32 phase 2 not being carried into the tart path.

## Spec

`tart-list-peers.sh` must discover the same peer set the rest of the base layer
does — root repo **plus** the peers declared in `construct/go.mod`.

### Why not "just seed the walk at construct/"

Re-seeding the BFS at `construct/` instead of the repo root is wrong: peers are
cloned by **basename** into `$TART_CLONE_DIR/<basename>` (see
`_tart_prepare_clone`), so seeding at construct yields a VM workspace of
`{construct/, ariadne/}` — it **drops the repo itself** (the thing being worked
on) and hoists the `construct/` subdir to a stray top-level peer, breaking the
layout the VM expects (`~/repo` → `<repo>/`, with `construct/` nested inside).

### Correct shape

Keep `$repo` as peer #1 (preserved as today), and **additionally** enqueue the
replace targets from `$repo/construct/go.mod` into the same recursive BFS —
mirroring `bootstrap-peers.sh`, which starts from `THIS_REPO` and reads
`$TARGET_DIR/construct/go.mod` for the extra peers. Desired result for
`you-decide`:

```
/Users/xianxu/workspace/you-decide      (root repo — peer #1)
/Users/xianxu/workspace/ariadne          (via construct/go.mod replace)
```

ariadne's own `go.mod` has no local-path replaces, so the recursive walk
terminates there — no transitive third peer in this case, but the recursion must
remain so deeper chains still resolve.

### Acceptance

- `tart-list-peers.sh <derivative-repo>` lists the root repo **and** every peer
  reachable from `construct/go.mod`'s replace directives, transitively.
- Pre-#32 / no-`construct/go.mod` consumers still degrade to single-repo clone
  (current behavior preserved).
- A repo whose root `go.mod` *does* carry sibling replaces still has those
  honored (don't regress the root-walk; add construct, don't replace it).
- After fix, `make tart` in `you-decide` clones both `you-decide/` and
  `ariadne/` into the VM; `~/repo`'s `construct/go.mod` replace resolves to a
  present sibling.

## Notes

- Base-layer file (`.tart/scripts/tart-list-peers.sh`) — change propagates to
  all derivative consumers; standard review gate applies. Confirm the file is in
  `construct/base.manifest`'s portable set and update if needed.
- Surfaced from a `you-decide` session: `make tart` did not mirror the ariadne
  dependency declared in `construct/go.mod`.

## Plan

- [x] Refactor the inline replace-parser in `tart-list-peers.sh` into a
  `_enqueue_replaces <dir>` helper (rhs resolves relative to `<dir>`, no-op if
  `<dir>/go.mod` absent).
- [x] In the BFS, call `_enqueue_replaces "$current"` **and**
  `_enqueue_replaces "$current/construct"` per node — union of root + construct
  replaces, deduped by the existing `_is_seen` guard.
- [x] Update the script header to document the dual-go.mod walk + why construct/
  is never emitted as a peer.
- [x] Verify acceptance cases.
- [x] Confirm propagation: `.tart/scripts` is a wholesale `symlink` in
  `construct/base.manifest` (line 98), so the fix reaches every consumer via the
  symlink — no manifest edit, already live in `you-decide`.

## Log

**2026-05-28 — implemented + verified.**

Changed `.tart/scripts/tart-list-peers.sh`: factored the replace-walk into
`_enqueue_replaces <dir>`, then walk both `$current/go.mod` and
`$current/construct/go.mod` for every BFS node. construct/ is never enqueued as a
peer (only its replace targets), so the construct/ subdir stays nested inside its
owning repo's clone — the layout `_tart_prepare_clone` (basename clone) expects.

Verification (`bash tart-list-peers.sh <path>`):

| Case | Expected | Result |
|---|---|---|
| `you-decide` (derivative) | you-decide + ariadne | ✅ both |
| `ariadne` (root + construct, no local replaces) | ariadne only | ✅ |
| dir with no go.mod | single repo (itself) | ✅ |
| root go.mod carries sibling replace | root walk still honored | ✅ |
| both root *and* construct replaces | union, deduped | ✅ |

`bash -n` syntax check clean. Root cause was issue #32 phase 2 (selective
go.mod-driven cloning) not being carried into the tart peer-walk — it read the
root module only, while the ariadne `replace` lives in `construct/go.mod` (same
location `bootstrap-peers.sh` and `Makefile.workflow build:` already read).
