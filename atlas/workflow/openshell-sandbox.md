# OpenShell Sandbox in the Workflow

The OpenShell sandbox is a containerized Linux dev environment that runs alongside the host. It's part of ariadne's base layer — any repo that adopts ariadne gets sandbox support via `.openshell/`.

## Why a Sandbox

- **Consistent environment**: Linux container regardless of host OS
- **Network isolation**: allowlisted egress only (GitHub, package registries, AI APIs)
- **Safe for AI agents**: agents get `--permission-mode bypassPermissions` inside the sandbox
- **Parallel to host**: mutagen keeps repo, git state, worktrees, and nvim state in sync

## Setup (One-Time)

```bash
make bootstrap      # install prerequisites (openshell CLI, gh, mutagen)
make sandbox        # create container, sync repo, install tools, connect
```

First run downloads a small bootstrap cache (~50MB: nvim, zellij, oh-my-bash, lua, luacheck) to `.openshell/.bootstrap/` on the host, syncs it into the container, then installs. Subsequent runs reuse the cache.

## Daily Use

```bash
make sandbox        # connect (builds if needed)
make sandbox-clean  # re-apply config + reconnect with fresh shell
make sandbox-nuke   # destroy everything, re-download bootstrap cache
```

Inside the sandbox, zellij is the terminal multiplexer (leader: `Ctrl+q`, new tab: `Alt+t`, panes: `Alt+arrows`, search scrollback: `Alt+/`). Config lives in `.openshell/dotfiles/zellij/`.

## What's Inside the Container

- **Shell**: bash with vi mode, oh-my-bash
- **Editor**: neovim (synced nvim state from host)
- **Multiplexer**: zellij (gruvbox-dark theme)
- **Languages**: lua 5.4 + luacheck (for neovim plugin development)
- **AI agents**: claude (bypass permissions), codex (full-auto), gemini (auto-approve)
- **Git**: configured from host (name, email, gh auth forwarded); transport is
  HTTPS, not SSH — see below.

## Git transport: HTTPS, not SSH (#152)

Both sandboxes proxy **HTTP(S) only** — raw SSH can't traverse the proxy (it
fails the handshake: `nc: authentication method negotiation failed`). So every
sandboxed `git push`/`pull` — and thus every `sdlc` verb (`claim`, `pr`, `merge`,
`push`, `issue new`) — must use HTTPS transport, not the `git@github.com:` SSH
remotes the repos are cloned with. Auth is already wired: `gh`'s `git_protocol` is
`https` and gh is the HTTPS credential helper, so no tokens-in-URLs.

The switch is a runtime **`insteadOf` rewrite** (keeps the SSH URL in each repo's
config, rewrites it to HTTPS per-operation), applied in two places:

- **OpenShell container** — `.openshell/overlay/setup.sh` sets it in the
  container's `~/.gitconfig` (plus `http.sslVerify=false`, because the OpenShell
  proxy terminates TLS without a CA cert in the container).
- **Host `~/.gitconfig`** — used by the **Claude Code** sandbox (sandbox-exec on
  the host) and the host itself. A one-time operator step (NOT `sslVerify=false`
  — host TLS to github.com is real):
  ```bash
  git config --global --unset-all url."https://github.com/".insteadOf 2>/dev/null || true
  git config --global --add url."https://github.com/".insteadOf "git@github.com:"
  git config --global --add url."https://github.com/".insteadOf "ssh://git@github.com/"
  ```
  `insteadOf` is **multi-valued** — use `--add`, or the second line replaces the
  first and only `ssh://` survives (so `git@github.com:`, the form real origins
  use, silently stays SSH); the leading `--unset-all` keeps it idempotent on
  re-run. It's kept a **manual** step rather than baked into a base-layer script
  for **blast radius**: auto-applying a *global* transport rewrite from e.g.
  `bootstrap-peers.sh` would silently flip git transport for every downstream
  user, sandbox or not. If a host post-clone setup phase ever materializes, this
  is the line to migrate into it.

**Caveat — encrypted brain remotes.** `brain` / `brain-family` use
`gcrypt::ssh://git@github.com/…`; the `git@github.com:` prefix doesn't match, so
the rewrite leaves them on SSH (they still won't push from a sandbox). Switching
an encrypted gcrypt remote's transport is a separate, sensitive change (needs a
tested `gcrypt::https://` + GPG-in-sandbox) — not covered here.

## How It's Provided by the Base Layer

The `.openshell/` directory is listed in `construct/base.manifest`. In symlink mode, contents point back to ariadne. In vendor mode, files are copied.

**Key convention**: all runtime state is local to each repo. Scripts derive paths from how they're invoked (`$0`), not from where they physically live. See [Base Layer](base-layer.md) for the path resolution rules.

### Files (git-controlled, from base layer)

| Path | Purpose |
|---|---|
| `.openshell/sandbox.sh` | Lifecycle orchestrator (build, connect, clean, stop, nuke) |
| `.openshell/Makefile` | Make targets included by repo Makefile |
| `.openshell/overlay/bootstrap.sh` | Host-side dependency downloader |
| `.openshell/overlay/post-install.sh` | Sandbox-side installer (from bootstrap cache) |
| `.openshell/overlay/setup.sh` | Shell config, aliases, output capture (^Y) |
| `.openshell/policy.yaml` | Network egress allowlist |
| `.openshell/dotfiles/` | Zellij config, layouts |
| `.openshell/ssh_wrapper.sh`, `ssh-bin/` | SSH connectivity (`~/.ssh/config` block managed at runtime by `sandbox.sh:ensure_ssh_config`) |

### Runtime artifacts (local per-repo, gitignored)

| Path | Purpose |
|---|---|
| `.openshell/.bootstrap/` | Cached downloads (created by `make sandbox`) |
| `.openshell/.base-image-digest` | Tracks base container image version |

## File Sync — substrate peers (ariadne#44)

The sandbox syncs the **current repo plus its transitive substrate peers** (not
the whole parent workspace), via the shared `construct/scripts/list-peers.sh` —
the same walker `make tart` uses (#32/#41). The walker reads `construct/deps`
`substrate` rows (#60) plus the root `go.mod` for real Go app-dep siblings (the
legacy `construct/go.mod` carrier is no longer read since #60 M4).
Layout mirrors the host: each peer mutagen-syncs to `/sandbox/workspace/<name>`,
and since `$HOME` is `/sandbox` in the base image, that *is* `~/workspace/<name>`.
`~/repo` symlinks the current repo; a `~/.sandbox-current-repo` marker drives the
bashrc auto-cd so a login lands in `~/workspace/<repo>`.

- **Read-only by default.** Peers (e.g. ariadne under pair) sync one-way
  host→sandbox — you don't accidentally mutate the base layer from a
  derivative's sandbox. The current repo is two-way.
- **`SYNC=../a,../b make sandbox`** opts those peers into two-way writable; a
  writable peer also gets its `.git` (one-way host→sandbox — two-way `.git` over
  mutagen is conflict-prone, so share in-sandbox commits by pushing).
- **Mode-encoded sync names + declarative reconcile.** Sessions are named
  `${SANDBOX_NAME}-peer-<name>-{ro,rw}` (+ `-peergit-<name>`). Each `make
  sandbox` recomputes the desired set (`compute_sync_set`) and `reconcile_syncs`
  terminates any owned session not in it — so a re-run with `SYNC=` upgrades a
  peer ro→rw (and migrates off old session names) without a rebuild.
- **Teardown by prefix.** `sandbox-stop`/`-nuke` terminate every
  `${SANDBOX_NAME}-*` session (no static name list to drift). Startup prints the
  per-repo sync plan with each repo's mode.

`make tart` shares the `SYNC=` flag (decision 9) to widen the VM clone set
(no rw/ro axis there — everything is a COW clone). Orthogonal syncs (worktree,
nvim-state, plenary, claude-sessions) are unchanged.

## Output Capture (^Y)

The sandbox bash shell wraps the session in `script(1)`, providing a real pty. `preexec`/`precmd` hooks record byte offsets in the script log. Ctrl+Y extracts the last command's output and copies to clipboard via OSC 52 (works through SSH, zellij, tmux). No TUI exclusion list needed — programs see a real TTY.

The host zsh has the same mechanism (`~/.zshrc`), using `pbcopy` instead of OSC 52.
