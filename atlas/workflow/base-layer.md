# Ariadne Base Layer

Ariadne provides a portable base layer — constitution, workflow, sandbox, skills — that consuming repos adopt via `construct/setup.sh`.

## Adopting the Base Layer

### Prerequisites
- Clone ariadne as a sibling directory: `../ariadne` relative to your repo
- Or use `--vendor` mode for repos that can't depend on ariadne as a peer

### Setup

```bash
cd /path/to/your-repo
../ariadne/construct/setup.sh          # symlink mode (default)
../ariadne/construct/setup.sh --vendor # vendor mode (copies files)
```

Re-run to refresh after ariadne updates. Mode is recorded in `.ariadne-mode`.

### Modes

| Mode | How | When |
|---|---|---|
| **Symlink** | Files in your repo are symlinks into `../ariadne/` | Default. Requires ariadne as sibling clone. Updates automatically. |
| **Vendor** | Files are copied from ariadne into your repo | For public repos or CI without ariadne peer. Re-run setup.sh to refresh. |

## What Gets Installed

Defined in `construct/base.manifest` (in ariadne):

- **Constitution**: `AGENTS.md`, `CLAUDE.md` — shared development rules
- **Settings**: `.claude/settings.json` — merged from `.ariadne` and `.local` layers
- **Skills**: `.claude/skills/xx-*` — local-origin skills; `construct/adapted/` is symlinked so derivatives pick up ariadne's adapted superpowers verbatim (see [Construct: Adaptation is Ariadne-Only](construct-adaptation.md))
- **Makefile system**:
  - `Makefile` — generic root template (REPO_NAME, workflow + local include, help chain). Identical across consumers; per-repo concerns belong in `Makefile.local`.
  - `Makefile.workflow` — issue lifecycle targets + auto-includes of `.openshell/Makefile` and `.tart/Makefile`.
  - `scripts/` — issue-sync, pre-merge-checks, close-issue.py, lib.sh
- **Construct system**: `construct/scripts/`, `construct/local/`, `construct/datatype/` — skill + datatype management
- **Sandbox** (`.openshell/`) — Linux container dev environment (see below)
- **Tart VMs** (`.tart/`) — `make tart` (headless) and `make tart-gui` (display via macOS Screen Sharing.app via `--vnc`; tart's built-in UI is broken on Tahoe as of 2026-05) for macOS VM testing (Apple Silicon only); helpers under `.tart/scripts/`. The mount is an APFS clone of `$(CURDIR)` at `~/.tart/clones/$(TART_VM)` (writable, O(1) prepare via `clonefile(2)`; replaced the per-boot rsync in #29), exposed inside the VM at `/Volumes/My Shared Files/$(REPO_NAME)` and symlinked from `~/repo`. `tart-stop` / `tart-clean` remove the clone; an orphan-GC step at every boot reaps clones older than 7 days. Override `RUN_FLAGS=` for a no-mount boot. `make help-tart` for the full surface.
  - **VM hooks (`.tart/vm-hooks.d/`)** — per-repo VM customization without patching the base-layer setup (ariadne#59). After standard setup, `tart-vm-setup.sh` runs every `*.sh` in the **booted repo's** `.tart/vm-hooks.d/` in lexical `LC_ALL=C` order (zero-pad with `NN-` prefixes to sequence). Each runs as `bash <hook> <repo>`. Hooks run on **every cold-boot** ⇒ must be idempotent; a failing hook prints a `[warn]` and the loop continues (never strands you out of the shell). No dir → no-op. First consumer: nous's `00-gpg-setup.sh` (headless brain testing, nous#36).
- **Directory scaffolds**: `workshop/`, `atlas/` — standard repo layout

## Repo-Specific Extensions

These files are **not** overwritten by setup.sh and own everything
that doesn't generalize across consumers:

- `AGENTS.local.md` — repo-specific rules (merged with `AGENTS.md`)
- `Makefile.local` — repo-specific make targets and overrides:
  - `UPSTREAM_NAME` / `UPSTREAM_REFRESH` for re-export layers (nous has its own `setup.sh` that re-vendors ariadne, so its `Makefile.local` points refresh through that path)
  - `-include Makefile.nous` chain for repos that consume the nous layer (brain, brain.legacy*)
  - Any genuinely one-of-a-kind target the repo needs
- `.claude/settings.local.json` — repo-specific Claude Code settings (merged into `settings.json`)
- `.openshell/.bootstrap/`, `.openshell/.base-image-digest` — runtime artifacts (gitignored)

If you find yourself wanting to edit a vendored file directly, the
right move is almost always to (a) generalize the change and push it
into ariadne, or (b) override it in the `.local` layer. Direct edits
get clobbered on the next `make refresh`.

## Dev binaries — ownership = location (`dev-aliases.sh`)

**A Go binary is owned by the repo whose `cmd/X` source physically lives there.**
Derivatives never copy or symlink the source; they run the built binary or
compile in the owner (build-in-owner since #60 — `make sdlc-build` resolves the
owner and builds its `cmd/X`, no per-derivative `construct/go.mod`). Source
distributed through the file-symlink *substrate* channel (the old `symlink
cmd/X` directive) is the deprecated anti-pattern — code flows through Go
modules, not the symlink channel reserved for docs/config (#56, #57).
nous's `symlink lib/gmail` / `cmd/gmail` / `cmd/oneshot` directives (and the
9 resulting brain* symlinks) were retired under #57 — derivatives now obtain
gmail/oneshot via the dev-alias (build-in-owner), not symlinked source.

For a smooth dev loop, `construct/dev-aliases.sh` walks the active
ariadne-styled siblings and emits a shell function per owned `cmd/X`:

```
source <(~/workspace/ariadne/construct/dev-aliases.sh)
```

Each function builds to the **owner's** `bin/X` (the official, gitignored path
— not a temp dir, so it's safe for a service binary like `nous`) and runs it in
the **caller's** cwd, so it's always fresh and works for both repo-bound tools
(`sdlc`, operates on the repo you're in) and run-anywhere tools (`nous`). The
emitted form is `X() { ( cd OWNER && mkdir -p bin && rm -f bin/X && go build -o
bin/X ./cmd/X ) || return; OWNER/bin/X "$@"; }` (the `rm -f` mirrors the owner
Makefiles' code-signing-inode safety). The function only **builds + runs** — it
does **not** manage services (no `launchctl bootout`); use the owner's `make
<name>-dev` target for the stop-prod-then-serve flow. It's also a *shell
function* — not on PATH and not reachable from cron/launchd; a derivative that
needs one of these binaries non-interactively must add the `replace` + `tool`
consume-wiring (the module channel), not rely on the alias. Filters: skips re-export
symlinks and non-buildable dirs (so a derivative never shadows the owner), and
`cmd/X/.private` opts a binary out. `--list` shows `binary → owner`; `--strict`
fails on a duplicate name. The script lives at `construct/dev-aliases.sh`
(alongside `setup.sh`/`rollback.sh`), with its hermetic test under
`construct/scripts/test/dev-aliases.test.sh`. Like the other substrate scripts
it's documented by header comment + test, not a `SKILL.md` (those are for agent
skills, not dev-env helpers).

## Pushing Updates to All Consumers

Ariadne maintainers can propagate base-layer changes in one shot:

```bash
cd /path/to/ariadne
make refresh-recursive
```

This iterates every peer repo in the parent directory and runs
`make refresh` in each one that has a `Makefile.workflow` (the universal
"uses the ariadne base layer" signal — catches direct consumers via
`.ariadne-mode`, indirect ones via `.nous-mode`, and re-export layers
like nous itself). Failures are collected into a final summary; partial
progress is better than aborting on the first hiccup.

Defined in `ariadne/Makefile.local` — ariadne-only, not vendored
(consumers don't push to their own peers).

## Base-as-trunk: three layers, different physics (#82)

Because derivatives symlink ariadne's working tree, a base change is *live* in
every derivative the moment it's saved — high churn is fine, but *long-lived,
concurrent, invisible* base branches break reasoning. The fix reframes three
things that look alike but behave differently:

- **Tracker state** (issues, claims, status) — append-only, instantly shared,
  committed to main *out-of-band*; should never be working-tree residue.
- **Base-layer code** (`construct/`, `cmd/`) — shared *live* via symlinks; the
  real contention surface.
- **Leaf code** (derivative-specific) — naturally isolated per session.

Three `sdlc` mechanisms keep the common path smooth without adding a gate:

1. **`issue new` auto-syncs to main (#82 M1)** — filing an issue broadcasts it to
   origin/main via claim's shared `syncIssuesToMain` (best-effort: the file is
   still created if the push can't land). Tracker state lands on main, not as
   residue. See [issue-sync.md](issue-sync.md).
2. **Dirty-tree guards ignore tracker files (#82 M2)** — `assessDirty` buckets
   `workshop/issues|history/*.md` as non-blocking (tracked-modified or
   untracked); only dirty *code* blocks a merge. See [sdlc-binary.md](sdlc-binary.md).
3. **`start-plan` reads base contention (#82 M3)** — a one-line, non-blocking
   heads-up (branch / dirty-code / other in-flight base issues), emitted only in
   the base repo. Surfaces a "moving base" at the commitment point; never refuses.

Scoped OUT (a separate, larger concern): a layout-preserving worktree-set for the
rare case that needs ariadne *isolated* while other base work continues.

## Sandbox (.openshell/)

The sandbox is an OpenShell containerized dev environment. Base layer provides the full infrastructure.

### Path Resolution Convention

**Critical design rule**: all scripts in `.openshell/` resolve runtime paths to the **local repo**, not to ariadne.

- `.openshell/` is a real directory in every repo (created by setup.sh)
- Its contents (sandbox.sh, overlay/, dotfiles/, etc.) are symlinks to ariadne (symlink mode) or copies (vendor mode)
- `sandbox.sh` derives paths from `$0` (how it was invoked), not from where the script physically lives
- `REPO_DIR` = consuming repo root (from `dirname "$0"/..`)
- `SCRIPT_DIR` = `$REPO_DIR/.openshell` (always local)

### Runtime Artifacts (local per-repo, gitignored)

| Path | Created by | Purpose |
|---|---|---|
| `.openshell/.bootstrap/` | `make sandbox` (bootstrap.sh) | Pre-downloaded dependencies (nvim, zellij, lua, etc.) |
| `.openshell/.bootstrap/.done` | bootstrap.sh | Marker to skip re-downloading |
| `.openshell/.base-image-digest` | sandbox.sh | Tracks container base image version |

These are **not** in `base.manifest` — they're created at runtime by `make sandbox` and are local to each repo.

### Bootstrap Trampoline

The `.bootstrap/` cache is a small pre-download trampoline to avoid slow package manager installs inside the sandbox. `bootstrap.sh` downloads on the host (fast, no proxy), mutagen syncs to `/tmp/bootstrap/` in the sandbox, then `post-install.sh` installs from there.

### Sandbox Commands

```bash
make sandbox        # build (if needed) + connect
make sandbox-clean  # re-sync config, reconnect with fresh shell
make sandbox-nuke   # destroy everything including bootstrap cache
```
