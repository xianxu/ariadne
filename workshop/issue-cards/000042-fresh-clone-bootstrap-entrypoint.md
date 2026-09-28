---
id: '000042'
status: done
created: 2026-05-28
updated: 2026-05-28
actual_hours: 1.0
---

# Fresh-clone first-run bootstrap entrypoint (./bootstrap.sh)

## Problem

A standalone `git clone` of an ariadne-style derivative (e.g. `you-decide`)
cannot run *any* `make` target — including `make bootstrap`:

```
$ make bootstrap
make: Makefile: No such file or directory
make: *** No rule to make target `Makefile'.  Stop.
```

Root cause (surfaced in #41's session): nearly the entire workflow substrate is
committed as **sibling-relative symlinks** into `../<upstream>` — `Makefile`,
`Makefile.workflow` (where `bootstrap:` is defined), `construct/setup.sh`,
`construct/scripts/bootstrap-peers.sh`, `AGENTS.md`, `.tart/`, `.claude/skills/`,
… On a bare clone with no upstream checked out beside it, all of these dangle.
`make` can't even read its own `Makefile`, so no target exists.

This is the chicken-and-egg the design already acknowledges (`Makefile.workflow`
header: "run `../ariadne/construct/setup.sh` manually once … then `make
bootstrap`") — but that escape hatch *also* presupposes `../ariadne` is already
cloned. The genuine first step on a clean machine is always "clone the upstream
as a sibling," and nothing in the derivative tells you that or does it for you.
The failure message points at `Makefile`, giving zero hint that the real fix is
"clone ariadne next to me."
