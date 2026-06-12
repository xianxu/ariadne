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

## Files

| Path | Purpose |
|---|---|
| `.colima/colima.sh` | Lifecycle orchestrator (`up`/`gui`/`stop`/`clean`); owns all Colima interaction. |
| `.colima/vnc-setup.sh` | Guest-side TigerVNC + fluxbox provisioner (streamed via `colima ssh`). |
| `.colima/Makefile` | Thin make wrappers + `help-colima` + profile-name / mount derivation. |
| `.colima/test/colima.test.sh` | Process-level fake (`colima` stub on `PATH`) asserting command assembly + idempotency gates. |

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
