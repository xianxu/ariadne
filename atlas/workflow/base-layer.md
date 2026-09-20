# Ariadne Base Layer

Ariadne provides a portable base layer — constitution, workflow, sandbox,
skills — that consuming repos adopt through standalone `weave`.

## Adopting the Base Layer

Install Homebrew on macOS or Linux and put a compatible weave on PATH. The
`xianxu/ariadne/weave` formula is pending publication in #241; until then build
`go build -o bin/weave ./cmd/weave` in ariadne and use that candidate.

```sh
cd /path/to/your-repo
weave link github.com/xianxu/ariadne  # or an existing ../ariadne
weave compile
```

The link records a source-aware edge in `construct/deps`. Compile restores
missing sources, provisions root Brewfiles, builds owner-local tools, mounts
data, and generates the leaf's context. Existing matching checkouts are reused
without pull/reset. There is no vendor mode or go.mod substrate declaration.
See [Setup & Replication](setup-and-replication.md) for phase and ownership rules.

## What Gets Installed

Defined in `construct/base.manifest` (in ariadne):

- **Constitution**: `CLAUDE.md`, `AGENTS.md`, `GEMINI.md` — shared development rules (per-harness prose entry files, composed once + fanned; see [harness-integration.md](harness-integration.md))
- **Settings**: `.claude/settings.json` — weave folds manifest-declared layer
  settings fragments foundation-first, then applies `.claude/settings.local.json`
  last
- **Skills**: per-harness skill dirs — `.claude/skills/xx-*` (claude) + `.agents/skills/xx-*` (codex/gemini), each carrying the local (`xx-*`) + adapted (`superpowers-*`) skills — weave lowers these per layer (#107 Option B; see [harness-integration.md](harness-integration.md)); derivatives pick up ariadne's local + adapted skills through the weave LAYER WALK, each `<skill-dir>/<name>` pointing straight at ariadne's source dir (NO whole-dir `construct/adapted` symlink — #104 M3 dropped those; see [Construct: Adaptation is Ariadne-Only](construct-adaptation.md))
- **Makefile system**:
  - `Makefile` — consumer-authored, with an optional `-include Makefile.workflow`. Product targets remain available in a standalone checkout; ariadne does not seed or refresh this root. Each tool-owning layer declares its own `tools` target here.
  - `Makefile.workflow` — issue lifecycle targets + auto-includes of `.openshell/Makefile`, `.tart/Makefile`, and `.colima/Makefile`.
  - `scripts/` — issue-sync, pre-merge-checks, close-issue.py, lib.sh
- **Construct system**: `construct/scripts/` — skill tooling; `construct/datatype/` — datatype prototypes, **per-layer-owned (NOT symlinked)**: each layer owns its own dir and the `datatype` binary reads the DAG-merged union across the layer graph (#115 retired the `symlink construct/datatype` manifest row). (`construct/local/` + `construct/adapted/` are ariadne's OWN skill dirs, read by derivatives through the weave layer walk — NOT installed by symlink since #104 M3.)
- **Sandbox** (`.openshell/`) — Linux container dev environment (see below)
- **Tart VMs** (`.tart/`) — `make tart` (headless) and `make tart-gui` (display via macOS Screen Sharing.app via `--vnc`; tart's built-in UI is broken on Tahoe as of 2026-05) for macOS VM testing (Apple Silicon only); helpers under `.tart/scripts/`. The mount is an APFS clone of `$(CURDIR)` at `~/.tart/clones/$(TART_VM)` (writable, O(1) prepare via `clonefile(2)`; replaced the per-boot rsync in #29), exposed inside the VM at `/Volumes/My Shared Files/$(REPO_NAME)` and symlinked from `~/repo`. `tart-stop` / `tart-clean` remove the clone; an orphan-GC step at every boot reaps clones older than 7 days. Override `RUN_FLAGS=` for a no-mount boot (setup still runs), or `VANILLA=1 make tart` to additionally skip `tart-vm-setup.sh` and boot the pristine base image with only the ssh-pubkey install (ariadne#89; `make tart-clean` first for a guaranteed from-scratch base). `make help-tart` for the full surface.
  - **VM hooks (`.tart/vm-hooks.d/`)** — per-repo VM customization without patching the base-layer setup (ariadne#59). After standard setup, `tart-vm-setup.sh` runs every `*.sh` in the **booted repo's** `.tart/vm-hooks.d/` in lexical `LC_ALL=C` order (zero-pad with `NN-` prefixes to sequence). Each runs as `bash <hook> <repo>`. Hooks run on **every cold-boot** ⇒ must be idempotent; a failing hook prints a `[warn]` and the loop continues (never strands you out of the shell). No dir → no-op. First consumer: nous's `00-gpg-setup.sh` (headless brain testing, nous#36).
- **Colima VMs** (`.colima/`) — `make colima` family for clean **Linux** VM testing, the tart counterpart (ariadne#93/#94); shares the colorized-step/dimmed-log helper `construct/scripts/vm-log.sh` with `.tart`. See [colima-vm.md](colima-vm.md).
- **Directory scaffolds**: `workshop/`, `atlas/` — standard repo layout

Dynamic skills generate into isolated staging directories before ownership-checked
publication in the leaf. Their executable markers declare
`# weave-output: argv1`, accept the absolute output directory as their first
argument, and produce a regular nonempty `SKILL.md` there. See the
[generator contract](weave.md#dynamic-skill-output-contract) for marker migration
and stage recovery; ordinary bootstrap and compile commands are unchanged.

## Repo-Specific Extensions

These local files own everything
that doesn't generalize across consumers:

- `AGENTS.local.md` — repo-specific rules (merged with `AGENTS.md`)
- `Makefile.local` — repo-specific make targets and overrides:
  - `-include Makefile.nous` chain for repos that consume the nous layer (brain, brain.legacy*)
  - Any genuinely one-of-a-kind target the repo needs
- `.claude/settings.local.json` — repo-specific Claude Code settings (merged into `settings.json`)
- `.openshell/.bootstrap/`, `.openshell/.base-image-digest` — runtime artifacts (gitignored)

If you find yourself wanting to edit a shared file directly, the
right move is almost always to (a) generalize the change and push it
into ariadne, or (b) override it in the `.local` layer. Direct edits
get clobbered on the next `make weave`.

## Dev binaries — ownership = location (`dev-aliases.sh`)

**A Go binary is owned by the repo whose `cmd/X` source physically lives there.**
Derivatives run that owner's binaries; they do not copy its Go source through
the manifest. `weave compile` runs each owner's root `make tools` before
composition. The target builds explicit binaries into its own `bin/` from
tracked sources and must not depend on weave or generated helpers. Ariadne
builds sdlc, datatype, vocabulary, and doc-review; weave itself is distributed
separately. Child commands receive owner bins on PATH, while the human shell
requires an explicit PATH update (compile prints the directories).

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
needs one of these binaries non-interactively must provide the owner bin on
PATH or use its absolute path. Filters: skips re-export
symlinks and non-buildable dirs (so a derivative never shadows the owner), and
`cmd/X/.private` opts a binary out. `--list` shows `binary → owner`; `--strict`
fails on a duplicate name. The script lives at `construct/dev-aliases.sh`
with its hermetic test under
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
`make weave` in each one that has a `Makefile.workflow` (the universal
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
3. **`start-plan` reads dependency-path contention (#82 M3 / #83)** — a
   non-blocking heads-up, one line per repo on the dependency chain
   (`construct/deps`, walked transitively): branch / dirty-code / other in-flight
   issues for each upstream you read live. A derivative surfaces its upstream(s)
   (`base (ariadne): …`); the root reports its own concurrent work. Never refuses.
   See [sdlc-binary.md](sdlc-binary.md). (#83 fixed M3's original cwd==base
   heuristic, which mis-fired in every derivative.)

Scoped OUT (a separate, larger concern): a layout-preserving worktree-set for the
rare case that needs ariadne *isolated* while other base work continues.

## Sandbox (.openshell/)

The sandbox is an OpenShell containerized dev environment. Base layer provides the full infrastructure.

### Path Resolution Convention

**Critical design rule**: all scripts in `.openshell/` resolve runtime paths to the **local repo**, not to ariadne.

- `.openshell/` is a real directory materialized by weave
- Its shared contents (sandbox.sh, overlay/, dotfiles/, etc.) are symlinks to ariadne
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

## Standalone product checkouts

`./bootstrap.sh` ensures a compatible gateway via PATH or Homebrew and execs
`weave compile` from its own root. It does not parse layers or hand off to Make.
`make weave` and the shared `make bootstrap` prerequisite delegate to the same
CLI; consumer bootstrap extensions remain additive. Generic startup never
runs shell configuration installation.

The root Makefile remains authored; the workflow overlay and generated helpers
appear during compilation. CI also compiles before accessing these helpers,
then runs the optional repository setup hook and merge checks. The executable
fixtures in `construct/scripts/test/{bootstrap-transitive,portable-makefile}.test.sh`
and `scripts/test/portable-ci.test.sh` cover gateway reuse/failure, owner builds,
and the workflow's actual run-block order without host package installation.
