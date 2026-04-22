# Sandbox Environments

Ariadne uses two distinct sandboxes. Understanding which you're in matters for what you can do.

## 1. Claude Code Sandbox

Claude Code runs commands in a **permission-controlled sandbox** managed by the CLI itself. This restricts:
- **Filesystem**: read/write limited to the working directory and select config paths
- **Network**: allowlisted hosts only (GitHub, package registries, AI APIs, localhost)

This sandbox is always active during Claude Code sessions. It cannot be disabled by the agent — the user controls it via `/sandbox` or by approving individual commands.

**Key distinction**: this is a *policy layer* on the host machine. You're still on the host OS, just with restricted permissions.

## 2. OpenShell Sandbox (`.openshell/`)

A full **containerized development environment** running in Docker via OpenShell. This is a separate Linux machine synced to the repo via mutagen.

### Architecture

```
Host (macOS)
  ├── .openshell/sandbox.sh    — lifecycle management (start/stop/status)
  ├── .openshell/policy.yaml   — network allowlist for the container
  ├── .openshell/overlay/      — bootstrap scripts run inside the container
  │   ├── bootstrap.sh         — downloads dependencies (host-side)
  │   ├── post-install.sh      — installs from bootstrap cache (sandbox-side)
  │   └── setup.sh             — applies dotfiles, credentials, git config
  ├── .openshell/dotfiles/     — config files copied into the sandbox
  │   └── zellij/              — zellij terminal multiplexer config
  └── .openshell/.bootstrap/   — cached downloads (nvim, zellij, oh-my-bash, lua, etc.)
```

### Lifecycle

```bash
make sandbox-start    # preflight checks, docker, mutagen sync, bootstrap
make sandbox-stop     # tear down
make sandbox-status   # check sync and container state
make sandbox-ssh      # SSH into the sandbox
```

The sandbox syncs several directories via mutagen: the repo itself, git state, worktrees, bootstrap cache, and nvim state.

### Network Policy

Defined in `.openshell/policy.yaml`. Allows: GitHub, package managers (npm, cargo, pip, go, rubygems, hex, luarocks), AI APIs (Anthropic, OpenAI, Google, Amazon Q), and Ubuntu apt.

## Zellij (Terminal Multiplexer)

Zellij is installed inside the OpenShell sandbox and configured via `.openshell/dotfiles/zellij/`. It replaces tmux with a gruvbox-dark themed multiplexer.

### Quick Reference

**Leader key**: `Ctrl+q` (replaces default `Ctrl+b`)

#### Normal Mode Keybindings (always active)

| Action | Binding |
|---|---|
| Enter leader/tmux mode | `Ctrl+q` |
| New tab | `Alt+t` |
| Previous / next tab | `Alt+Left` / `Alt+Right` |
| Move tab left / right | `Alt+i` / `Alt+o` |
| Focus left / right / down / up | `Alt+h` / `Alt+l` / `Alt+j` / `Alt+k` |
| New pane above / below | `Alt+Up` / `Alt+Down` |
| Close focused pane | `Alt+x` |
| Toggle floating panes | `Alt+w` |
| Resize increase / decrease | `Alt+=` / `Alt+-` |
| Cycle swap layout | `Alt+[` / `Alt+]` |
| Enter scroll/search mode | `Alt+/` |

#### Leader Mode (`Ctrl+q` → key)

| Action | Key |
|---|---|
| Tab mode | `t` |
| Pane mode | `p` |
| Session mode | `o` |
| Resize mode | `n` |
| Move mode | `h` |
| Scroll mode | `s` |
| Lock mode | `g` |
| Quit zellij | `q` |
| Back to normal | `Esc` |

### Common Operations

**Search text in scrollback**: `Alt+/` → `s` → type query → `Enter` → `n`/`p` to jump between matches → `Esc` to exit.

**Rename a tab**: `Ctrl+q` → `t` (Tab mode) → `r` → type name → `Enter`.

**Rename a session**: `Ctrl+q` → `o` (Session mode) → `r` → type name → `Enter`.

**Split pane horizontally** (top/bottom): `Alt+Down` or `Alt+Up`.

**Split pane vertically** (left/right): `Ctrl+q` → `p` (Pane mode) → `r` (right) or `d` (down) or `n` (auto).

**Attach to existing session**: `zellij attach <name>` or `zellij a` to pick interactively.

**List sessions**: `zellij ls`.

### Layout

The default layout (`.openshell/dotfiles/zellij/layouts/default.kdl`) provides:
- Tab bar at the top
- Main pane in the middle
- Status bar + clock widget at the bottom

### Configuration Note

Keybindings use `clear-defaults=true` on the `normal` block. This means `shared_except` blocks may not apply — always put bindings directly in the `normal` block.
