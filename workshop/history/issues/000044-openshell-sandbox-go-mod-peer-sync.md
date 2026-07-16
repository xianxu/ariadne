---
id: 000044
status: done
deps: [000032, 000029, 000041, 000042, 000045]
created: 2026-05-29
updated: 2026-05-29
estimate_hours:
actual_hours: 4
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
- `SYNC=../ariadne make sandbox` makes `~/workspace/ariadne` two-way writable —
  working-tree edits in the sandbox appear on the host — and syncs its `.git`
  one-way (host→sandbox) so in-sandbox git ops have history. (Revised from
  "commit appears on host": two-way `.git` over mutagen is conflict-prone, so
  in-sandbox commits are shared by `git push`, not by mutagen — decision 5.)
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

- [x] **M1 — Promote the peer walker.** `git mv .tart/scripts/tart-list-peers.sh
  construct/scripts/list-peers.sh`; `.tart/scripts/tart-list-peers.sh` is now a
  back-compat symlink → it (resolves in derivatives via the wholesale
  `.tart/scripts` symlink chaining to it). Added optional extra-repo args
  (decision 9: union seeds). Added `symlink construct/scripts/list-peers.sh` to
  `construct/base.manifest`. `discover_ancestors` stays inline (has extra
  sources: `go list -m all` + no-go.mod fallback — not a clean swap) — shares
  grammar, documented in the script header + atlas. Verified `make tart` peer
  set byte-identical before/after (pair: {pair, ariadne}); extras union works
  (pair + you-decide → {pair, you-decide, ariadne}).
- [x] **M2 — Drift test for bootstrap.** Landed in #45
  (`construct/scripts/test/bootstrap-transitive.test.sh`, t8); reference updated
  to the canonical `construct/scripts/list-peers.sh` path; green.
- [x] **M3 — Sandbox layout + go.mod sync.** `ensure_mutagen_sync` rewritten to
  drive off `list-peers.sh` (via `compute_sync_set`); each peer syncs to
  `/sandbox/workspace/<name>`; in-sandbox `ln -sfn` makes `~/workspace` → the
  peer tree and `~/repo` → current repo, plus a `~/.sandbox-current-repo` marker.
  Dropped the `WORKSPACE_DIR` sync + the "symlink every sibling" loop.
  `overlay/setup.sh`: added the bashrc auto-cd block; **removed `mkdir ~/repo`**
  (now a symlink — a pre-made real dir would make `ln -sfn` nest inside it).
- [x] **M4 — Mode reconcile + SYNC= flag.** Mode-encoded sync names
  (`…-peer-<name>-{ro,rw}`, `…-peergit-<name>`); `compute_sync_set` classifies
  current-repo + `SYNC=` entries as rw, other go.mod peers ro; `reconcile_syncs`
  diffs desired-vs-owned and terminates stale (drives ro→rw upgrade, rw→ro
  downgrade, migration off old fixed names). `SYNC=` read from env. Writable
  peers get `.git` (one-way host→sandbox; two-way `.git` is conflict-prone —
  decision 5 amended below). `ensure_sync` keys on the mode-encoded name.
- [x] **M5 — Teardown by prefix + startup print.** `terminate_all_syncs` +
  `list_owned_syncs` enumerate `mutagen sync list --template` filtered by
  `${SANDBOX_NAME}-` (minus bootstrap); static `SYNC_NAMES` retired. Startup
  prints the per-repo sync plan (name → path, read-only/writable).
- [x] **M6 — `SYNC=` for tart.** `.tart/Makefile` passes `SYNC` (comma→space) as
  extra seeds to the shared walker. Verified: pair clone set widens from
  {pair, ariadne} to {pair, you-decide, ariadne} with `SYNC=../you-decide`.
- [x] **M7 — Live E2E verification (integration).** Verified against a real
  `pair` sandbox: build + sync-plan print; both peers sync into
  `~/workspace/{pair,ariadne}`; rw `pair` gets `.git`, ro `ariadne` doesn't;
  `~/repo`→pair + marker; **ro→rw upgrade** via `SYNC=../ariadne` re-run
  (reconcile drops `-ro`, adds `-rw`+`peergit`, no rebuild); `sandbox-stop` →
  zero sessions; migration off old fixed-name syncs. Caught + fixed 3 real
  integration bugs (self-loop symlink, ssh-eats-stdin, mutagen missing-parent).
  auto-cd: snippet present, cd target verified, fires on interactive connect
  (same mechanism existing Ctrl+Y binds use) — final pty confirmation is the
  operator's `make sandbox`.
- [x] Post-milestone code review (AGENTS.md §3): fresh-eyes over the full diff,
  no Critical; I1 (basename collision) + M1 (soft_cleanup) + I2/M3 addressed.
  Atlas updated (`workflow/openshell-sandbox.md` file-sync section +
  `setup-and-replication.md`). Lessons logged.

## Notes

- Base-layer change across `.openshell/`, `.tart/`, `construct/scripts/`,
  `construct/base.manifest` — propagates to every derivative; standard review
  gate applies.
- Strong precedent to mirror: #41's dual-go.mod walk and #32's selective-clone
  rationale. This issue is "extend that model to the openshell sandbox."

## Log


- 2026-05-29: closed — sandbox syncs go.mod peers into ~/workspace (rw current+SYNC=, ro peers); live E2E on pair: layout, ro→rw upgrade, teardown-to-zero all verified; hermetic 6/6 + 11/11; review SHIP. FORCE: M1-M7 single cohesive change reviewed in one fresh-eyes pass over the full diff (Review-Verdict SHIP, window 0d7bf3d..HEAD); no per-milestone close commits by design.
- 2026-05-29: opened. Surfaced while first using the openshell sandbox in
  `pair`; design converged with operator in-session (decisions 1–9 above).
  Sibling fixes that unblocked sandbox usage landed first: `/dev/ptmx` policy
  fix (`70e434e`) and Ctrl+C trap / fail-fast (`f05d9d2`).
- 2026-05-29: added dep on #45 (bootstrap transitivity gap), split out of this
  design discussion. The shared `replace`-line parser (M1/M2) is co-owned with
  #45; the "one parser, two walk drivers (clone-absent vs list-present)" split
  is documented there.
- 2026-05-29: implemented M1–M6 + live-verified M7. Hermetic tests:
  `sandbox-sync-set.test.sh` 6/6 (rw/ro classification, SYNC ro→rw, non-peer
  union, multi-SYNC, leaf fallback, basename-collision guard); list-peers
  regression byte-identical + extras union; bootstrap drift green. Live E2E on
  `pair`: full layout, ro→rw upgrade, teardown-to-zero all confirmed; 3
  integration bugs (self-loop symlink under $HOME=/sandbox; ssh consuming the
  while-read loop's stdin → fixed with fd-3 + `ssh -n`; mutagen beta
  missing-parent → pre-`mkdir`). **Code review** (fresh-eyes, full diff): no
  Critical; I1 basename-collision (now fails loudly + capture-aborts the build)
  and M1 soft_cleanup (wipe /sandbox/workspace) addressed; I2/M3 hardened.
  Verdict SHIP, window 0d7bf3d..HEAD.
- 2026-05-29 — **operator follow-up needed (rollout):** `make refresh` was run
  in `pair` to materialize `construct/scripts/list-peers.sh`; other derivatives
  need the same to pick up the manifest symlink before their sandbox uses the
  go.mod peer set (else it degrades to current-repo-only with a warning).
  `bootstrap.sh` (#45) is `seed`-delivered (write-once) — already-seeded
  derivatives keep their old copy until re-seeded.
