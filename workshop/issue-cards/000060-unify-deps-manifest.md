---
id: '000060'
status: done
created: 2026-06-01
updated: 2026-06-02
estimate_hours: 5
actual_hours: 2.5
---

# Unify substrate + data dependencies into one manifest; retire construct/go.mod as the peer graph

## Problem

Substrate dependencies (the ariadne-styled ancestor a repo inherits its base
layer + Go tools from) are declared today inside `construct/go.mod`, via a
`require` + `replace => ../../ariadne` + `tool …/cmd/sdlc` stub module
(`<name>-construct`). This was an artifact of an imperfect grasp of the go.mod
system — go.mod is being used as a substrate-peer graph, which is **not what it
is intended for**. Consequences:

- A whole second Go module (`<name>-construct`) exists per derivative that
  builds nothing of the repo's own — it carries only the substrate declaration.
- Non-Go substrate consumers (a markdown-only brain) are forced to carry a fake
  go.mod just to declare lineage.
- The peer graph is split across two carriers: substrate lives in
  `construct/go.mod`, content lives in `construct/data-deps` (a clean, flat,
  language-agnostic, two-column manifest). Two mechanisms for the same
  primitive — "a sibling clone, floating-HEAD, surfaced via symlink."

Meanwhile `construct/go.mod`'s **build** role (compile `cmd/sdlc` through the
`replace`) is already redundant: both the deploy path and the dev path can —
and the dev path (`dev-aliases.sh`) already does — **build the tool from the
owner's checkout** (`cd ../ariadne && go build ./cmd/sdlc`), using ariadne's
own go.mod. See [[000055-sdlc-binary-single-owner-path-freshness]] and
[[000049-data-dependencies]].
