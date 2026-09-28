---
id: '000041'
status: done
created: 2026-05-28
updated: 2026-05-28
actual_hours: 0.5
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
