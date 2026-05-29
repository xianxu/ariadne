---
id: 000044
status: open
deps: [000032, 000029, 000041, 000042]
created: 2026-05-29
updated: 2026-05-29
estimate_hours:
actual_hours:
---

# openshell sandbox: sync go.mod peers (not the whole workspace), mirror host layout, SYNC= flag

## Problem

The openshell sandbox (`.openshell/sandbox.sh`) is the laggard of the base
layer's peer-dependency mechanisms. `make tart` was brought onto the
`construct/go.mod` peer model in #32 phase 2 (selective clone of current repo +
transitive `replace` peers) and corrected in #41. The sandbox still uses the
**old "sync the whole parent workspace" model**:

- `ensure_mutagen_sync` one-way-syncs all of `WORKSPACE_DIR` (the entire
  `~/workspace/`) into `/sandbox/workspace`, then symlinks *every* sibling dir
  (`for dir in WORKSPACE_DIR/*/`) — pulling in repos that have nothing to do
  with the current repo's dependency graph.
- The main repo syncs to `/sandbox/repo`, while `overlay/setup.sh:182` makes
  `~/repo` (`/home/sandbox/repo`) — a different directory. The login layout and
  the sync target disagree, and neither matches the host's `~/workspace/<repo>`
  shape that tart already adopted.

This is the same inconsistency #41 fixed for tart, one level up: the sandbox
isn't go.mod-aware at all.

## Spec

Bring the openshell sandbox onto the same `construct/go.mod` peer model tart
uses, harmonize the in-sandbox path layout with the host, and add a `SYNC=`
flag for opting peers into writable two-way sync.

### Decisions (resolved with operator, 2026-05-29)

1. **One shared peer-walk for substrate-present consumers.** Promote the
   transitive `go.mod`-replace BFS (currently `.tart/scripts/tart-list-peers.sh`)
   to a neutral home, `construct/scripts/list-peers.sh`, and have tart, the
   openshell sandbox, and `setup.sh:discover_ancestors` all consume it.
   `.tart/scripts/tart-list-peers.sh` becomes a symlink to it.
   - **`bootstrap.sh` is the documented exception** — `seed`-delivered (real,
     write-once) so it runs on a peerless clone where a symlinked shared script
     would dangle. It keeps its inline, direct-only replace parser. A regression
     test asserts bootstrap's parser and `list-peers.sh` extract the **same
     replace targets** from a fixture `go.mod`, so they cannot silently drift.
   - Rejected: seed-delivering `list-peers.sh` to make it available pre-substrate
     — that reintroduces copy-divergence #38 retired.

2. **Sync only go.mod peers, not the whole parent.** Replace the
   `WORKSPACE_DIR` one-way sync and the "symlink every sibling" loop with one
   mutagen sync per peer returned by `list-peers.sh "$REPO_DIR"`.

3. **Mirror host layout: `~/workspace/<repo>` canonical, auto-cd on login.**
   - Peers sync into `~/workspace/<basename>` (real sibling layout — same as
     host `~/workspace/`), replacing the flat `/sandbox/<peer>` symlink hack.
   - Current repo is `~/workspace/<current-repo>`; shell auto-`cd`s there on
     interactive start via a marker file (`~/.sandbox-current-repo`, written by
     `apply_config`) — the bash analogue of tart's `~/.tart-current-repo` /
     `tart-vm-rc.zsh` (openshell base image is bash + oh-my-bash).
   - Keep `~/repo` as a back-compat symlink → `~/workspace/<current-repo>`
     (tart does the same; `~/repo/bin` is on PATH, `overlay/setup.sh` references
     `~/repo`).

4. **Default peers read-only; `SYNC=` overrides to writable.**
   - go.mod peers → **one-way** host→sandbox (read-only mirror; you don't
     accidentally mutate the base layer from a derivative's sandbox).
   - Current repo → **two-way**.
   - `SYNC=../repo1,../repo2 make sandbox` → those peers become **two-way
     writable**. Paths resolve relative to `$REPO_DIR`.

5. **Writable repos get `.git` synced; read-only peers do not.** A writable
   peer needs its history to commit/push from inside the sandbox; a read-only
   mirror doesn't (skip `.git` → faster, less disk).

6. **Re-run reconciles mode (the upgrade path).** Starting a sandbox with
   ariadne read-only, exiting, then `SYNC=../ariadne make sandbox` must
   **upgrade** ariadne's sync to writable. Implement declaratively:
   - Each run recomputes the desired sync set: `{peer → mode}` where
     `mode ∈ {ro, rw}`, current-repo = rw, go.mod-peer = ro, `SYNC=`-entry = rw.
   - **Encode mode in the sync name** (`${SANDBOX_NAME}-peer-<name>-ro` vs
     `-rw`, plus `${SANDBOX_NAME}-peergit-<name>` for writable peers' `.git`).
   - Reconcile = name-set diff: terminate existing `${SANDBOX_NAME}-*` syncs not
     in the desired set, create the missing ones. RO→RW upgrade, RW→RO
     downgrade, and stale-peer removal all fall out of this. (mutagen can't
     mutate a session's mode in place, so upgrade == terminate + recreate; the
     name encoding makes the diff drive that for free.)
   - `ensure_sync`'s current "exists → skip" early-return becomes mode-aware
     (keyed on the mode-encoded name).

7. **Teardown by prefix, not a static list.** `make sandbox-stop` /
   `sandbox-nuke` terminate **every** mutagen session named `${SANDBOX_NAME}-*`
   (query `mutagen sync list`, filter prefix). This satisfies "tear down
   whatever sync we started" regardless of peer count or `SYNC=` extras, and
   retires the brittle hardcoded `SYNC_NAMES="repo git workspace …"` (a latent
   leak today).

8. **`make sandbox` prints the sync plan on startup.** One `log` line per
   synced repo: name, path, and `read-only` / `writable`. No silent peer set.

9. **`SYNC=` for tart too, with specialized semantics.** tart COW-clones
   everything writable-divergent, so there `SYNC=` has no RO/RW axis — it just
   adds repos to the clone set (union with go.mod peers). `list-peers.sh`
   optionally accepts the extras as args so both consumers compute the same
   union in one place.

### Non-goals / preserved

- `claude-sessions`, `nvim-state`, `plenary`, `worktree` syncs are orthogonal
  (not peers) — keep as-is.
- Pre-#32 / no-`construct/go.mod` consumers degrade to current-repo-only sync
  (mirrors tart's single-repo fallback).
- The Ctrl+C trap + fail-fast-on-Error changes (just landed, `f05d9d2`) stay.

### Acceptance

- `make sandbox` in `pair` syncs `pair` (rw) + `ariadne` (ro) into
  `~/workspace/{pair,ariadne}`; login lands in `~/workspace/pair`; `~/repo`
  resolves there. Unrelated siblings (e.g. `brain-*`) are NOT synced.
- `SYNC=../ariadne make sandbox` makes `~/workspace/ariadne` two-way writable
  and syncs its `.git`; a commit made there appears on the host.
- Start ro → exit → re-run with `SYNC=../ariadne`: ariadne flips ro→rw without a
  full rebuild; the stale `-ro` session is gone, `-rw` + `.git` present.
- `make sandbox-stop` leaves zero `${SANDBOX_NAME}-*` mutagen sessions.
- Startup prints each synced repo with its mode.
- `list-peers.sh` and `bootstrap.sh`'s inline parser extract identical replace
  targets from the fixture `go.mod` (drift test green).
- `bash -n` clean on all touched scripts; `make tart` peer set unchanged
  (regression: the shared-script promotion must not alter tart's behavior).

## Plan

> Draft — awaiting operator approval before implementation (non-trivial,
> base-layer, multi-file per AGENTS.md §2).

- [ ] **M1 — Promote the peer walker.** Move `tart-list-peers.sh` →
  `construct/scripts/list-peers.sh` (add optional extra-repo args for the
  union); symlink the old tart path to it; point `setup.sh:discover_ancestors`
  at it (or document why it stays inline). Update `construct/base.manifest`.
  Verify `make tart` peer set is byte-identical before/after.
- [ ] **M2 — Drift test for bootstrap.** Fixture `go.mod` + test asserting
  `bootstrap.sh`'s inline parser and `list-peers.sh` agree on replace targets.
- [ ] **M3 — Sandbox layout + go.mod sync.** Rewrite `ensure_mutagen_sync` to
  drive off `list-peers.sh`; sync into `~/workspace/<basename>`; `~/repo` +
  auto-cd marker; drop the `WORKSPACE_DIR` sync and sibling-symlink loop.
  Reconcile `overlay/setup.sh` (`~/repo` mkdir, login cwd).
- [ ] **M4 — Mode reconcile + SYNC= flag.** Mode-encoded sync names; declarative
  reconcile; `SYNC=` plumbed Makefile→sandbox.sh; `.git` for writable peers;
  mode-aware `ensure_sync`.
- [ ] **M5 — Teardown by prefix + startup print.** Prefix-based
  `terminate_all_syncs`; retire `SYNC_NAMES`; per-repo sync-plan log line.
- [ ] **M6 — `SYNC=` for tart.** Union semantics in the tart Makefile via the
  shared walker.
- [ ] Post-milestone code review per AGENTS.md §3; update `atlas/` (workflow
  map) for the converged peer-sync model; log outcome.

## Notes

- Base-layer change across `.openshell/`, `.tart/`, `construct/scripts/`,
  `construct/base.manifest` — propagates to every derivative; standard review
  gate applies.
- Strong precedent to mirror: #41's dual-go.mod walk and #32's selective-clone
  rationale. This issue is "extend that model to the openshell sandbox."

## Log

- 2026-05-29: opened. Surfaced while first using the openshell sandbox in
  `pair`; design converged with operator in-session (decisions 1–9 above).
  Sibling fixes that unblocked sandbox usage landed first: `/dev/ptmx` policy
  fix (`70e434e`) and Ctrl+C trap / fail-fast (`f05d9d2`).
