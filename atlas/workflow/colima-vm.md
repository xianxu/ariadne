# Colima VMs (`make colima`) — clean Linux VM testing

`make colima` is the **Linux** counterpart to [`make tart`](base-layer.md#what-gets-installed)
(macOS VMs). It boots a [Colima](https://github.com/abiosoft/colima) profile —
a Lima Linux VM on Apple's Virtualization framework (`vz`, the same framework
tart uses) — bind-mounts the workspace, and drops you into a shell. Use it to
test the substrate (the `sdlc` binary, agent CLIs, the install/bootstrap flow)
on clean Linux.

Vendored under `.colima/` as part of ariadne's base layer (see
[Base Layer](base-layer.md)); any repo that adopts ariadne inherits the verbs.

## The four verbs

| Verb | Behavior |
|---|---|
| `make colima` | Start profile `<repo>-test` (idempotent), bind-mount the workspace **writable**, `colima ssh` in landing in the current repo. |
| `make colima-gui` | Same + provision a TigerVNC + fluxbox desktop in the guest; connect from the Mac at `vnc://localhost:5901`. |
| `make colima-stop` | Stop the profile. |
| `make colima-clean` | Stop + delete the profile (the Lima base image stays cached). |

Parallel profiles per repo: `COLIMA_SUFFIX=ying make colima` → `<repo>-ying`
(the twin of tart's `TART_SUFFIX`). Resource knobs via env:
`COLIMA_CPUS` / `COLIMA_MEMORY` / `COLIMA_DISK` / `COLIMA_ARCH` / `COLIMA_VMTYPE`.
amd64 guest for cross-arch testing: `COLIMA_ARCH=x86_64 make colima` (adds
`--vz-rosetta` under `vz`).

Requires `brew install colima` (Apple Silicon).

## The per-repo unit is a Colima *profile*

A profile is a full Lima Linux VM, named `<repo>-<COLIMA_SUFFIX|test>` — the
structural twin of tart's `TART_VM`. The name is sanitized (lowercase, `.`→`-`)
the same way `.openshell/Makefile` derives `SANDBOX_NAME`. `colima.sh`'s
idempotency gates (`_running` via `colima status -p`, `_exists` via
`colima list -j`) make every verb safe to re-run.

## Live mount, not a COW clone (the key difference from tart)

Tart APFS-**clones** only the `go.mod` peer set into the VM, because its
`cp -cR` (clonefile) cost is linear in file count, so cloning a multi-GB
workspace dominated boot (ariadne#32). Colima/Lima mounts are **live virtiofs
shares** — nothing is copied, mounting is lazy — so that rationale doesn't
transfer. `make colima` therefore bind-mounts the **whole workspace parent**
(`realpath ../`) writable:

- Always complete — no pinned peer-set that goes stale when `go.mod` changes.
- Siblings stay adjacent at their host paths inside the guest, so Go `replace
  ../peer` directives resolve.
- Host edits (including uncommitted) are reflected live — no cold-reboot-for-
  freshness dance (tart#28's VirtIO-FS staleness doesn't bite here).

Consequence: writes inside the VM hit the host worktree directly (unlike tart,
where VM writes diverge into a throwaway APFS clone). Keep that in mind when a
test mutates files.

## The GUI path provisions its own desktop

`tart-gui` inherits the macOS desktop for free via Screen Sharing. A Colima
profile is a headless Ubuntu VM, so `make colima-gui` runs
`.colima/vnc-setup.sh` in the guest: it `apt`-installs TigerVNC + fluxbox
(one-time), writes `~/.vnc/xstartup`, and starts `vncserver :1` bound to
`0.0.0.0:5901`. Lima auto-forwards that to the Mac's `127.0.0.1:5901`, so
`open vnc://localhost:5901` (macOS Screen Sharing) connects. Auth is a VNC
password (default `colima`, override `COLIMA_VNC_PASSWORD`) — TigerVNC refuses
an auth-less non-localhost bind, and a password is cheap defense-in-depth even
though the port is only reachable from the host via Lima's user-mode NAT.

## Post-login customization (parity with Tart, #94)

On boot `colima.sh` provisions the guest — the Linux counterpart to Tart's
`tart-vm-setup.sh`, mirroring the **portable** pieces of the OpenShell *Linux*
overlay (not Tart's macOS zsh rc):

- `.colima/vm-setup.sh` (guest bootstrap): `~/workspace`/`~/repo` symlinks to the
  host-abs mount, a `~/.colima-current-repo` marker, `apt install neovim`
  (non-fatal), oh-my-bash, appends the `vm-rc.sh` source-line to `~/.bashrc`
  (after oh-my-bash so it survives), and runs `.colima/vm-hooks.d/*.sh`.
- `.colima/vm-rc.sh` (guest bash rc, pushed to `~/.colima-vm-rc.sh`): git aliases
  (`s ss a d p`), `v`, vi-mode + Ctrl+R/S, `EDITOR=nvim`, GPG_TTY, the
  `dev-aliases.sh` Go build-on-demand wrappers, and auto-cd into the repo.

`colima.sh` re-pushes the cheap `vm-rc.sh` every boot (so host edits propagate)
but runs the heavier `vm-setup.sh` only on a **fresh** start, and makes it
**non-fatal** — a setup hiccup never strands you out of a running VM.
**Deliberately NOT ported** (sandbox/macOS-specific): the egress-proxy config,
`/tmp/bootstrap` credentials, macOS DNS-flush/Homebrew PATH, the `script(1)`
output-capture machinery, and AI-agent aliases. Per-repo customization goes in
`.colima/vm-hooks.d/*.sh` (idempotent, run-parts order, continue-on-error — the
twin of `.tart/vm-hooks.d`).

## Logging (shared with Tart, #94)

Both `make colima` and `make tart` route step headers + underlying-process
output through one shared helper, `construct/scripts/vm-log.sh` (single source of
the ANSI codes — ARCH-DRY): `step`/`warn` print bold-cyan/bold-yellow headers;
`dim` grays a piped stream (tty + `NO_COLOR`/`CLICOLOR_FORCE` gated). Colima dims
its boot/setup/vnc sub-output; Tart streams its previously-hidden boot log dimmed
during the SSH-wait. The final interactive shell is never dimmed.

## Files

| Path | Purpose |
|---|---|
| `.colima/colima.sh` | Lifecycle orchestrator (`up`/`gui`/`stop`/`clean`); owns all Colima interaction. |
| `.colima/vm-setup.sh` | Guest bootstrap on boot (symlinks, marker, nvim, oh-my-bash, rc-append, hooks). |
| `.colima/vm-rc.sh` | Guest bash rc (aliases, vi-mode, dev-aliases, auto-cd); pushed to `~/.colima-vm-rc.sh`. |
| `.colima/vnc-setup.sh` | Guest-side TigerVNC + fluxbox provisioner (streamed via `colima ssh`). |
| `.colima/Makefile` | Thin make wrappers + `help-colima` + profile-name / mount derivation. |
| `.colima/test/colima.test.sh` | Process-level fake (`colima` stub on `PATH`) asserting command assembly + idempotency gates. |
| `construct/scripts/vm-log.sh` | Shared colorized-step / dimmed-passthrough logger, used by both `.colima` and `.tart`. |

Wired via `Makefile.workflow` (`-include .colima/Makefile`), the top-level
`help` aggregation (`help-colima`), and `construct/base.manifest` (symlinks).
The recipes are thin wrappers over `colima.sh` (mirrors
[`.openshell/sandbox.sh`](openshell-sandbox.md), not tart's inline-`define`
style) — logic lives in the shell script, testable via a `PATH`-injected fake
rather than mocks.

## Related

- [Base Layer](base-layer.md) — how `.colima/` propagates; the tart family lives here too.
- [OpenShell Sandbox](openshell-sandbox.md) — the other Linux-container option (container, not VM; mutagen-synced dev env).
- [Sandbox](sandbox.md) — Claude Code sandbox vs container sandbox.
