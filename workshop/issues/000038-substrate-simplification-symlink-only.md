---
id: 000038
status: done
deps: [000037]
created: 2026-05-27
updated: 2026-05-27
estimate_hours: 4
actual_hours: 3.0
---

# substrate simplification — symlink-only, drop vendor/copy, bootstrap-clones-peers

## Problem

#37 landed the `construct/go.mod` split — surgical Go vendoring (substrate-tool source in `construct/vendor/`, app deps separate at root). The split solved the immediate "pair's vendor/ ballooned" complaint but kept the dual-mode (`--symlink` vs `--vendor`) machinery around in setup.sh.

After further design work, the dual-mode + vendor concept turns out to be unnecessary complexity:

1. **Privacy was overstated.** Ariadne's artifacts aren't really sensitive — what gets shipped as substrate (AGENTS.md, manifests, skill docs) is hard to keep private once any consumer ships derived content. Making ariadne public (or accepting that its content effectively leaks through consumers) removes vendor's main justification.

2. **Vendor is just symlink-with-ergonomic-veneer (modulo privacy).** Functionally, vendoring a snapshot is equivalent to "symlink to a frozen branch in upstream + don't pull updates." The differences are operational ergonomics (where the divergent copy lives, how clone DX works) — not semantic.

3. **Bootstrap-clones-peer addresses the clone DX argument for vendor.** If the only reason for vendor was "operator can clone derivative without needing upstream sibling," then `make bootstrap` that auto-clones missing peers achieves the same DX without bundling source.

4. **Per-artifact `copy` for divergence has its own headache.** Mixed copy/symlink within a logical command (cmd/gmail with some files copied, some symlinked) makes state hard to reason about. Per-operator branches in source repos (one branch per operator's customizations, shared across all that operator's brains) is a cleaner divergence mechanism than per-artifact copy.

The simplification: drop vendor mode entirely, drop `copy` action entirely, everything is symlink. Divergence happens via per-operator branches in source repos. Clone DX happens via bootstrap-cascades-and-clones.

## Spec

### Substrate after simplification

**Five actions** in `base.manifest` (down from six):

| Action | Semantics |
|---|---|
| `symlink` | Live reference from upstream. Uniform with upstream's checked-out branch. |
| `tool` | Go module tool dependency. Adds require + replace (`../<peer>`) + tool to target's `construct/go.mod`. Built via sibling-checkout at `make tools` time. |
| `scaffold` | Empty directory creation (target-local; no upstream involvement). |
| `touch` | Empty file creation if missing (target-local). |
| `merge` | Settings merge — reads `.X.<layer>.json`, writes `.X.json`. Implicit source-rename. |

**Dropped:** `copy` (single-file template OR directory). Use `symlink` instead; if operator wants divergence, branch in source.

### Bootstrap cascade

```
make bootstrap (in any repo)
  → parse construct/go.mod for replace directives matching `../<name>` pattern
  → for each such peer:
       if peer not cloned → git clone <derived-url> ../<name>
       cd ../<name> && make bootstrap   (recurse; depth limit 5)
  → make refresh   (peers exist now; symlinks materialize)
  → make tools     (build this repo's tools; sources resolved via sibling-replaces)
  → local-env setup (per-repo; e.g., nous's GPG setup in Makefile.nous)
```

**Cycle detection:** visited-set during recursion; defense against malformed graphs. Hard depth limit of 5 as backstop (current chain depth is 3: brain → nous → ariadne).

**Peer URL derivation:** by convention from current repo's `git remote get-url origin`, substituting `<this-repo-name>` → `<peer-name>` in the path. Override via `Makefile.local` variable for edge cases.

### Substrate vs transitive Go deps

The bootstrap-clones-peer rule applies only to `replace` directives matching the **sibling pattern** (`=> ../<name>` or `=> ../../<name>`). Other replaces (absolute paths, version pins) are ignored by bootstrap — those resolve via Go's standard module mechanism (proxy / cache).

`construct/go.mod` lists only **substrate peers** (ariadne, nous, future co-developed substrate repos). Transitive Go library deps (charmbracelet, cobra, etc.) come from Go's module proxy when tools are built.

### Per-operator branch for divergence

Operator-owned customizations to substrate live as **per-operator branches in source repos**, not per-derivative copies in derivative trees:

```
nous/                          # main branch = canonical / community-shared
  ├─ main                      # community version
  └─ xianxu-customizations     # xianxu's edits, shared across all xianxu's brains
```

Each operator's brains symlink to nous's checked-out branch. Switch branch in nous → all brains see the new customizations. Merge from main periodically to pull upstream improvements.

This replaces the `copy` action's per-derivative divergence mechanism. Branch space stays small (one per operator, not per brain).

### `make refresh` purity

`make refresh` is now purely substrate-state sync:
- Verifies peers are present (errors with clear "run `make bootstrap`" hint if not)
- Runs `construct/setup.sh` to update symlinks
- Does NOT clone peers, does NOT build tools

Operator separation of concerns:
- `make bootstrap` = first-time setup OR full state recovery
- `make refresh` = "I think substrate is stale; re-sync symlinks"
- `make tools` = "rebuild binaries from current source"

## Plan

- [x] **M1 — `setup.sh` cleanup.** Done (commit bdbdd46). setup.sh shrunk 678→516 lines. Dropped MODE/vendor/copy machinery; symlink-only.
- [x] **M2 — Manifest action cleanup.** Done (commits 4462af4 ariadne + c102a07 nous). All `copy` entries switched to `symlink`; action list docstring updated; setup.sh warns on stale `copy` entries.
- [x] **M3 — Makefile.workflow refresh + bootstrap.** Done (commit 7905bd6). New `bootstrap-peers.sh` does recursive peer-cloning via origin-URL convention. `make refresh` pure sync; `make bootstrap` prereqs-only composition (bootstrap-peers + refresh + tools); `make tools` composes per-repo build prereqs.
- [x] **M4 — Per-derivative cleanup.** Done (nous 1b0c47a + parley.nvim + pair 30842e7). `rm -rf construct/vendor/` in all three; `make sdlc-build` works via sibling-checkout + Go replace. Substrate refreshed against #38.
- [x] **M5 — Atlas update.** Done (commit a11767d). `setup-and-replication.md` rewrites the modes section as symlink-only + bootstrap-cascade; updates the `construct/go.mod` section to reflect no-vendor model.

## Out of scope

- **Making ariadne public.** A separate operational decision (gh repo edit, audit content). Substrate simplification works either way; public ariadne removes a justification for vendor but isn't a prerequisite.
- **The `make refresh` peer-missing UX.** Currently errors with hint "run `make bootstrap`." Could auto-bootstrap on refresh, but that conflates verbs. Keep separation for now.
- **`merge` action evolution.** Stays as-is — `.X.<layer>.json` overlay convention works.

## Log


- 2026-05-27: closed — setup.sh 678→516 lines; symlink-only model; bootstrap cascade via bootstrap-peers.sh; vendor/ removed from nous/parley.nvim/pair; make sdlc-build works via sibling-replace in all 3 derivatives; atlas updated. FORCE since M1-M5 committed direct (no milestone-close → no verdict trailers; same bootstrap pattern as #31/#37)
### 2026-05-27 — issue created

Issue extracted from extended design discussion immediately following #37. The operator's reasoning chain:

1. Privacy was overstated as a substrate concern — public ariadne is fine, and even if private, vendored substrate text leaks anyway through any public consumer that ships derived content.
2. Symlink-with-frozen-branch in source is functionally equivalent to vendor (operator-divergent copy), modulo where the copy physically lives. So vendor is just ergonomics over symlink+branch.
3. Clone DX (vendor's main remaining argument) is solvable via bootstrap-cascades-and-clones — operator types `make bootstrap` once, peer chain materializes.
4. Per-artifact `copy` adds state-reasoning complexity — mixed copy/symlink within a logical command (e.g., cmd/gmail with some files copied) gets murky. Per-operator branches in source are cleaner.
5. Per-operator (not per-brain) branches scale well: one branch per operator's preferences, shared across all that operator's derivatives.

Net direction: symlink-only substrate, bootstrap-clones-peers (recursive), divergence-via-source-branches.

deps: [000037] — #37 landed the construct/go.mod split that this builds on; #38 simplifies the layers above.
