---
id: 000042
status: done
deps: [000041]
created: 2026-05-28
updated: 2026-05-28
estimate_hours:
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

## Spec

Add a real (non-symlink) committed `bootstrap.sh` at the derivative root — the
one entrypoint that runs with **zero substrate present**.

### Why it must be a real file, not a symlink

If `bootstrap.sh` symlinked into `../<upstream>` it would dangle exactly like
`Makefile` does today and solve nothing. It must be a standalone file in the
derivative's git so it exists the instant you `git clone`. This is the defining
constraint.

### What it does

It depends only on what survives a peerless clone: `git` + the **real**
`construct/go.mod` (confirmed real, not symlinked) + the `origin` remote.

```
./bootstrap.sh
  1. read construct/go.mod        → direct peers (replace … => ../../<name>)
  2. derive each peer URL from `git remote get-url origin` (swap repo-name,
     same convention as bootstrap-peers.sh)
  3. git clone each missing peer as a sibling                 (clones ariadne)
  4. exec `make bootstrap`        → symlinks now resolve → full cascade
                                     (bootstrap-peers → refresh → tools → sdlc)
```

It clones only the **direct** peers (enough to make the symlinks resolve); the
transitive cascade is left to the now-present `bootstrap-peers.sh`. Idempotent —
skips peers already present; safe to re-run.

### Materialization: the `seed` manifest action

`bootstrap.sh`'s canonical copy lives in ariadne. It can't be delivered by the
existing manifest actions: `symlink` (would dangle — the whole point),
`scaffold` (empty dirs only), `touch` (empty files only); `copy` was retired in
#38. So add a narrow new action:

- **`seed <path>`** — write-once copy of a real file from upstream into the
  target as a standalone file (preserving mode via `cp -p`). Never overwrites an
  existing target. For first-run entrypoints that **cannot** be symlinks because
  they must work before the upstream is present.

This is *not* a revival of the general `copy` action #38 retired (which existed
so operators could diverge substrate files — discouraged, now branch-based).
`seed` is the opposite: a one-time delivery of an intentionally-generic,
not-meant-to-be-edited file. Write-once + trivial-by-design ⇒ no drift to
manage (decision from the #41-session design discussion: "write-once + keep it
trivial" over "refresh-synced," since the real logic stays in the symlinked
`bootstrap-peers.sh`).

Manifest line: `seed  bootstrap.sh`. The existing self-reference filter in
setup.sh skips seeding ariadne's `bootstrap.sh` onto itself (source path ==
target path on self-walk).

### Acceptance

- A clone of a derivative with **no upstream sibling** can run `./bootstrap.sh`
  and end with all peers cloned + `make bootstrap` completed.
- `./bootstrap.sh` is idempotent (peers present → skip-and-handoff).
- `seed` action: real file in target, mode preserved, write-once (re-run is a
  no-op), skipped on self-walk in ariadne.
- The immediate repo that hit this (`you-decide`) gets a committed
  `bootstrap.sh` so a fresh clone works today.
- Document the first-run command in the bootstrap docs (`Makefile.workflow`
  header + atlas) so it's discoverable.

## Plan

- [x] Add canonical `bootstrap.sh` at ariadne root (real, generic, idempotent).
- [x] Add `seed` action: `create_seed` fn + dispatch case in `construct/setup.sh`;
      document in `construct/base.manifest` action list; add `seed bootstrap.sh`.
- [x] Update `Makefile.workflow` bootstrap header + atlas
      (`setup-and-replication.md`) with the fresh-clone first-run command.
- [x] Verify: simulate a peerless clone; `bootstrap.sh` clones the upstream and
      hands off; re-run is a no-op.
- [x] Seed + commit `bootstrap.sh` into `you-decide` (+ README install block).

## Log


- 2026-05-28: closed — offline e2e: peerless clone → make bootstrap fails → ./bootstrap.sh clones upstream sibling + hands off → BOOTSTRAP_OK; idempotent re-run; seed write-once in you-decide (real exec file, not symlink); seed self-walk skipped in ariadne; bash -n clean
**2026-05-28 — implemented + verified.**

Added `bootstrap.sh` (ariadne root): reads real `construct/go.mod`, derives peer
URLs from `origin` (swap repo-name), clones direct peers as siblings, `exec make
bootstrap`. Idempotent. Added `seed` manifest action (`create_seed` + dispatch in
`construct/setup.sh`; documented in `base.manifest`; line `seed bootstrap.sh`) —
write-once real-file copy, mode-preserving, for entrypoints that can't be
symlinks. Docs: `Makefile.workflow` bootstrap header + `atlas/workflow/
setup-and-replication.md` (six actions, fresh-clone section, fourth go.mod
consumer).

Verification:

- **End-to-end (offline, local bare remotes):** fresh peerless clone of a fake
  derivative → `make bootstrap` fails (`Makefile: No such file or directory`) →
  `./bootstrap.sh` derives upstream URL, clones it as sibling, hands off →
  `BOOTSTRAP_OK`. Re-run: peer present → skip + handoff (idempotent). ✅
- **`seed` in derivative (`you-decide`):** `construct/setup.sh` printed
  `seeded bootstrap.sh`; result is a real executable file (not symlink),
  byte-identical to canonical; re-run is a no-op (write-once). ✅
- **`seed` self-walk (ariadne):** printed `skipped bootstrap.sh (self-reference
  at canonical location)`; canonical file md5-unchanged, still a real file. ✅
- `bash -n` clean on `bootstrap.sh`.

Seeded + committing `bootstrap.sh` into `you-decide` so a fresh clone works
today; README install block updated `make bootstrap` → `./bootstrap.sh`.
