---
id: '000049'
status: done
created: 2026-05-29
updated: 2026-06-02
actual_hours: 2.5
---

# Data dependencies: a looser git-submodule for content peers

## Problem

An ariadne repo sometimes wants to **consume the content of another repo**
without that repo being part of its substrate layer graph. The motivating
case: `brain` consumes `you-decide` (candidate reads, controversies, the
voter-advisor skills) but does **not** derive its base layer from it —
`you-decide` is a *sibling* derivative of ariadne, not an ancestor of `brain`.

Today there is no first-class way to declare this. The two things you can reach
for are both wrong:

- **The substrate peer mechanism** (root/`construct` `go.mod` `replace` →
  `bootstrap-peers.sh` clone + `setup.sh discover_ancestors` apply). Declaring a
  content repo here makes the walker apply its `construct/base.manifest`,
  symlinking the peer's base-layer files into the consumer. Wrong — it couples
  *clone* with *substrate-apply* (see "the enabling seam" below).
- **A hand-made symlink + a stray go.mod in the data tree** (what the
  brain↔you-decide experiment did first). The data-tree `go.mod` is read by
  nothing — no walker recurses into `data/` — so it does not clone on
  bootstrap. It's inert decoration; only the manual symlink actually mounts.

So the operator is left hand-cloning the sibling and hand-making the symlink,
with nothing to drive it on a fresh brain bootstrap.
