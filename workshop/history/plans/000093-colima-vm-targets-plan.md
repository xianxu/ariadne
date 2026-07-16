# `make colima` — Lima-VM Linux Testing Targets Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `.colima/` base-layer fragment that gives `make colima` / `colima-gui` / `colima-stop` / `colima-clean` (parallel profiles via `COLIMA_SUFFIX`) for clean **Linux** VM testing of the substrate, mirroring `.tart/`'s macOS-VM verbs.

**Architecture:** One `.colima/colima.sh` orchestrator (subcommands `up|gui|stop|clean`) wraps the Colima CLI; a `.colima/vnc-setup.sh` guest script provisions a VNC desktop for the GUI verb; `.colima/Makefile` is thin wrappers (mirrors `.openshell/sandbox.sh` + Makefile, not `.tart`'s inline-`define` style — keeps recipes a thin IO shell, ARCH-PURE). The per-repo unit is a **Colima profile** (a Lima Linux VM) named `<repo>-<suffix|test>` — the structural twin of `TART_VM`. The workspace parent dir is bind-mounted writable (live virtiofs, not a COW copy).

**Tech Stack:** GNU Make, **bash 3.2-compatible** shell (macOS system bash — no arrays-expanded-under-`set -u`), Colima 0.10.x (Lima + Apple `vz` framework), TigerVNC + fluxbox in-guest, Ubuntu guest userland.

---

## Why this diverges from `.tart` in two deliberate places

1. **Mount = whole workspace parent, not go.mod-peer subset.** Tart APFS-clones only the `list-peers.sh` set because `cp -cR` (clonefile) is O(1) per file but linear in *file count*, so a multi-GB workspace dominated boot (ariadne#32). Colima/Lima mounts are **live virtiofs shares** — nothing is copied, mounting is lazy — so that latency argument does not transfer. Mounting `$(WORKSPACE_DIR)` (`realpath $(CURDIR)/..`) writable is simpler, always-complete (no pinned peer-set that goes stale when `go.mod` changes), and keeps siblings adjacent so `replace ../peer` resolves. **ARCH-DRY note:** reuse `list-peers.sh` only where its justification holds; here it does not, so we don't. (Per-peer mounting is a documented future extension if workspace exposure ever matters.)

2. **`colima-gui` provisions its own desktop.** `tart-gui` inherits the macOS desktop for free via Screen Sharing. A Colima profile is a headless Ubuntu VM, so `colima-gui` must `apt-get` a minimal X + VNC server in the guest, start it bound to `0.0.0.0:5901` (so Lima auto-forwards it to the Mac's `127.0.0.1:5901`), and point the operator at `vnc://localhost:5901`. This is heavier than `tart-gui` and is why the brainstorm flagged it.

---

## Core concepts

This feature is overwhelmingly an **IO/glue layer** (Make + Colima orchestration); the pure surface is thin (two string derivations). Tests reflect that: a **process-level fake** of the `colima` binary (a stub on `PATH` that records argv) drives deterministic assertions on the command lines each verb assembles, and a **real boot→ssh→stop→clean cycle** on this Mac is the integration verification (captured in the issue Log). This is the constitution's "process-level fake, not function mocks" for the external `colima` dependency.

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `COLIMA_PROFILE` (profile-name derivation) | `.colima/Makefile` | new |
| `WORKSPACE_DIR` mount-flag derivation | `.colima/Makefile` | new |

- **`COLIMA_PROFILE`** — sanitized profile name `<repo>-<suffix|test>`, lowercased with `.`→`-` (Lima/Colima profile names, like the openshell k8s name, must be lowercase alphanumeric + dashes).
  - **Relationships:** 1:1 with a repo+suffix pair; the twin of `.tart`'s `TART_VM`.
  - **DRY rationale:** Reuses the exact sanitization idiom already in `.openshell/Makefile:7` (`tr '.' '-' | tr '[:upper:]' '[:lower:]'`) — same fact (a VM/instance name derived from `REPO_NAME`), same transform. **ARCH-DRY.**
  - **Future extensions:** none expected; suffix axis already covers parallel instances.

- **`WORKSPACE_DIR` mount-flag derivation** — `realpath $(CURDIR)/..` → passed to `colima.sh` which renders `--mount "<dir>:w"`. Trivially pure.
  - **Relationships:** 1:1 with the invoking repo.
  - **DRY rationale:** First occurrence; intentionally *not* the `list-peers.sh` machinery (see divergence #1).
  - **Future extensions:** per-peer mounting via `list-peers.sh` if whole-workspace exposure ever matters.

These two are one-liners inside the Makefile/script; they do not earn separate test files. They are exercised by the process-level-fake test (Task 5), which asserts the rendered profile name and `--mount` flag.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `colima.sh` (orchestrator) | `.colima/colima.sh` | new | Colima CLI (`start`/`ssh`/`stop`/`delete`/`status`/`list`) |
| `vnc-setup.sh` (guest provisioner) | `.colima/vnc-setup.sh` | new | `apt-get` + `vncserver` inside the guest |
| `.colima/Makefile` (target wiring + help) | `.colima/Makefile` | new | `make` → `colima.sh` |
| Workflow include | `Makefile.workflow:8` (after the `.tart` include) | modified | `-include` |
| Top-level help aggregation | `Makefile:20` | modified | `help` prereqs |
| Base-layer manifest | `construct/base.manifest` | modified | symlink propagation |
| Atlas doc | `atlas/workflow/colima-vm.md` + `atlas/index.md` | new/modified | documentation |

- **`colima.sh`** — single orchestrator with verbs `up|gui|stop|clean`. Holds all Colima interaction (one source of truth — ARCH-DRY) and all idempotency gating (`_running` / `_exists` helpers). Mirrors `.openshell/sandbox.sh`'s subcommand shape.
  - **Injected into:** invoked by `.colima/Makefile` recipes (thin wrappers). The `colima` binary is injected via `PATH`, which is what makes the process-level fake possible without touching the script.
  - **Future extensions:** a `status` verb; resource knobs already via env (`COLIMA_CPUS`/`MEMORY`/`DISK`/`ARCH`).

- **`vnc-setup.sh`** — idempotent guest-side script (streamed in via `colima ssh -- bash -s`): installs TigerVNC + fluxbox, writes `~/.vnc/xstartup`, (re)starts `vncserver :1` on `0.0.0.0:5901`.
  - **Injected into:** run by `colima.sh gui`.
  - **Future extensions:** password auth (currently `-SecurityTypes None`, safe because the port is only reachable from the host via Lima's user-mode net); geometry/depth env knobs.

---

## Implementation risks to nail down empirically (verification, not guesswork)

These are flagged so the executor verifies rather than assumes — each has a fallback:

- **R1 — Mount overlap with Colima's auto `$HOME` mount.** Colima auto-mounts `$HOME` read-only; our `--mount $(WORKSPACE_DIR):w` is nested under it. Expectation: the more-specific writable mount shadows for that subtree. **Verify:** `touch` a file under the mounted repo from inside the guest. **Fallback if it fails:** generate a profile config via `colima start --edit` (or a templated YAML) that marks the workspace mount writable / drops the auto-home overlap.
- **R2 — Landing in the repo dir over `colima ssh`.** Plan uses `colima ssh -p P -- bash -lc 'cd "<repo>"; exec bash -il'`; needs a PTY. **Verify:** run it, confirm interactive shell in the repo. **Fallback:** drop into `$HOME` and `echo "cd <repo>"`.
- **R3 — Lima auto-forwards guest `0.0.0.0:5901` to host `127.0.0.1:5901`.** **Verify:** after `colima-gui`, `nc -z localhost 5901` on the host. **Fallback:** add an explicit `portForwards` entry via the profile config, or document `colima ssh -- ... -L`.
- **R4 — `colima start` on an existing *stopped* profile reconciles `--mount` flags.** Determines whether a changed workspace requires `colima-clean` or just `colima-stop && colima`. **Verify:** start, stop, restart with a different `--mount`, inspect. Document whichever holds.

---

## Chunk 1: the `.colima` fragment, wiring, tests, docs

### Task 1: `.colima/colima.sh` orchestrator

**Files:**
- Create: `.colima/colima.sh`

- [ ] **Step 1: Write the orchestrator**

```bash
#!/usr/bin/env bash
# colima.sh — Colima-profile lifecycle for clean Linux VM testing, the
# Linux counterpart to .tart (macOS VMs). Invoked by .colima/Makefile.
#
#   colima.sh up    <profile> <repo-dir> <mount-dir>   start (idempotent) + ssh in
#   colima.sh gui   <profile> <repo-dir> <mount-dir>   up + provision VNC desktop
#   colima.sh stop  <profile>                          stop the profile (idempotent)
#   colima.sh clean <profile>                          stop + delete (base image cached)
#
# The per-repo unit is a Colima *profile* — a Lima Linux VM, the twin of
# .tart's TART_VM. <mount-dir> (the workspace parent) is bind-mounted
# writable: a LIVE virtiofs share (no copy), so unlike tart there's no
# per-peer clone — the whole workspace is exposed and siblings stay
# adjacent for go.mod `replace ../peer` resolution.
#
# All Colima interaction lives here (one source of truth); the Makefile
# recipes are thin wrappers. `colima` is resolved via PATH, which lets the
# test (.colima/test/colima.test.sh) inject a process-level fake.
set -euo pipefail

# Resource knobs (override via env / `make colima COLIMA_MEMORY=8`).
COLIMA_CPUS=${COLIMA_CPUS:-4}
COLIMA_MEMORY=${COLIMA_MEMORY:-4}
COLIMA_DISK=${COLIMA_DISK:-60}
COLIMA_ARCH=${COLIMA_ARCH:-aarch64}     # x86_64 → amd64 guest (adds --vz-rosetta under vz)
COLIMA_VMTYPE=${COLIMA_VMTYPE:-vz}      # Apple Virtualization.framework, same as tart
COLIMA_VNC_GEOMETRY=${COLIMA_VNC_GEOMETRY:-1600x1000}

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

_need_colima() {
    command -v colima >/dev/null 2>&1 || {
        echo "Install colima: brew install colima" >&2; exit 1; }
}

# Running iff `colima status -p P` exits 0 (it prints "colima is running"
# for a started profile, non-zero for stopped OR absent).
_running() { colima status -p "$1" >/dev/null 2>&1; }

# Exists (started or stopped) iff it appears in `colima list -j` (compact
# JSONL today; tolerate spacing in case a future colima pretty-prints).
_exists() { colima list -j 2>/dev/null | grep -q "\"name\"[[:space:]]*:[[:space:]]*\"$1\""; }

_start_if_needed() {
    local profile=$1 mountdir=$2
    if _running "$profile"; then
        echo "==> Profile '$profile' already running."
        return
    fi
    echo "==> Starting Colima profile '$profile' (Linux VM; mounts $mountdir writable)..."
    # Build argv in the positional params, NOT a bash array: macOS system bash
    # is 3.2, which aborts under `set -u` when expanding an empty array
    # ("${arr[@]}" → unbound variable). "$@" empty-expansion is always safe.
    set -- -p "$profile" \
        --cpu "$COLIMA_CPUS" --memory "$COLIMA_MEMORY" --disk "$COLIMA_DISK" \
        --arch "$COLIMA_ARCH" --vm-type "$COLIMA_VMTYPE" \
        --mount "$mountdir:w"
    if [ "$COLIMA_ARCH" = "x86_64" ] && [ "$COLIMA_VMTYPE" = "vz" ]; then
        set -- "$@" --vz-rosetta   # fast amd64 emulation under Apple vz
    fi
    colima start "$@"
}

# Interactive shell, landing in the repo dir (mounted at its host abs path
# inside the guest). `|| true`: the remote login shell's exit status on
# `exit` is a normal session-end, not a make failure. See R2.
_ssh_into() {
    local profile=$1 repodir=$2
    echo "==> colima ssh -p $profile  (landing in $repodir)"
    colima ssh -p "$profile" -- bash -lc "cd \"$repodir\" 2>/dev/null || true; exec bash -il" || true
}

_provision_vnc() {
    local profile=$1
    echo "==> Provisioning VNC desktop in '$profile' (idempotent; first run apt-installs)..."
    # Geometry as a positional arg (bash -s -- ARG), NOT an env var: `colima ssh
    # -- cmd` execve's the command directly, so a `VAR=val cmd` prefix would be
    # parsed as the command name. `bash -s -- 1600x1000` lands in the guest's $1.
    colima ssh -p "$profile" -- bash -s -- "$COLIMA_VNC_GEOMETRY" < "$SCRIPT_DIR/vnc-setup.sh"
    echo "==> VNC ready. Connect from the Mac with:"
    echo "        open vnc://localhost:5901          # macOS Screen Sharing"
    echo "    (or any VNC viewer at localhost:5901). Lima forwards guest 5901 → host 5901."
}

cmd=${1:-}; shift || true
case "$cmd" in
  up)
    _need_colima; _start_if_needed "$1" "$3"; _ssh_into "$1" "$2" ;;
  gui)
    _need_colima; _start_if_needed "$1" "$3"; _provision_vnc "$1"; _ssh_into "$1" "$2" ;;
  stop)
    _need_colima
    if _running "$1"; then echo "==> Stopping '$1'..."; colima stop -p "$1";
    else echo "==> '$1' not running."; fi ;;
  clean)
    _need_colima
    if _running "$1"; then echo "==> Stopping '$1'..."; colima stop -p "$1"; fi
    if _exists "$1"; then echo "==> Deleting profile '$1' (base image stays cached)..."; colima delete -p "$1" -f;
    else echo "==> '$1' does not exist."; fi ;;
  *)
    echo "usage: colima.sh up|gui <profile> <repo-dir> <mount-dir> | stop|clean <profile>" >&2
    exit 2 ;;
esac
```

- [ ] **Step 2: Make it executable**

Run: `chmod +x .colima/colima.sh`
Expected: no output; `test -x .colima/colima.sh` succeeds.

- [ ] **Step 3: Smoke the usage guard (no colima needed)**

Run: `.colima/colima.sh 2>&1; echo "exit=$?"`
Expected: prints the `usage:` line and `exit=2`.

- [ ] **Step 4: Commit**

```bash
git add .colima/colima.sh
git commit -m "#93: colima.sh — profile lifecycle orchestrator (up/gui/stop/clean)"
```

### Task 2: `.colima/vnc-setup.sh` guest provisioner

**Files:**
- Create: `.colima/vnc-setup.sh`

- [ ] **Step 1: Write the guest script**

```bash
#!/usr/bin/env bash
# vnc-setup.sh — runs INSIDE the Colima/Lima guest (streamed via
# `colima ssh -- bash -s`). Idempotently installs a minimal X + VNC
# desktop and (re)starts vncserver on :1 (port 5901), bound to 0.0.0.0
# so Lima forwards it to the Mac's 127.0.0.1:5901.
#
# Security: -SecurityTypes None is acceptable because the port is only
# reachable from the host through Lima's user-mode network — not the LAN.
# A password is a future extension.
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
GEOMETRY=${1:-1600x1000}   # passed via `bash -s -- <geometry>` from colima.sh

if ! command -v vncserver >/dev/null 2>&1; then
    echo "  installing tigervnc + fluxbox (one-time)..."
    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        tigervnc-standalone-server tigervnc-common fluxbox xterm dbus-x11 >/dev/null
fi

mkdir -p "$HOME/.vnc"
cat > "$HOME/.vnc/xstartup" <<'EOF'
#!/bin/sh
unset SESSION_MANAGER DBUS_SESSION_BUS_ADDRESS
[ -x /usr/bin/dbus-launch ] && eval "$(dbus-launch --sh-syntax)"
exec fluxbox
EOF
chmod +x "$HOME/.vnc/xstartup"

vncserver -kill :1 >/dev/null 2>&1 || true
vncserver :1 -geometry "$GEOMETRY" -depth 24 -localhost no -SecurityTypes None >/dev/null 2>&1
echo "  vncserver up on :1 (guest 0.0.0.0:5901), geometry $GEOMETRY"
```

- [ ] **Step 2: chmod + shellcheck-clean**

Run: `chmod +x .colima/vnc-setup.sh && bash -n .colima/vnc-setup.sh && echo OK`
Expected: `OK` (syntax valid).

- [ ] **Step 3: Commit**

```bash
git add .colima/vnc-setup.sh
git commit -m "#93: vnc-setup.sh — guest-side TigerVNC desktop provisioner"
```

### Task 3: `.colima/Makefile` thin wrappers + help

**Files:**
- Create: `.colima/Makefile`

- [ ] **Step 1: Write the Makefile fragment**

```makefile
# Colima-VM management — included from Makefile.workflow:
#   -include .colima/Makefile
#
# The Linux counterpart to .tart (macOS VMs). The per-repo unit is a
# Colima *profile* (a Lima Linux VM), named like TART_VM:
#   COLIMA_PROFILE = <repo>-<COLIMA_SUFFIX|test>
#
# Provides:
#   make colima        Start profile $(COLIMA_PROFILE) (idempotent), bind-mount
#                      the workspace ($(WORKSPACE_DIR)) writable, ssh in landing
#                      in the current repo. Mounts are LIVE virtiofs (no copy),
#                      so host edits — incl. uncommitted — are reflected live.
#   make colima-gui    Same, plus provision a TigerVNC desktop in the guest and
#                      print how to connect (vnc://localhost:5901). Colima/Linux
#                      is headless, so this apt-installs X+VNC — heavier than
#                      tart-gui (which inherits the macOS desktop for free).
#   make colima-stop   Stop $(COLIMA_PROFILE).
#   make colima-clean  Stop + delete $(COLIMA_PROFILE) (Lima base image cached).
#
# Parallel profiles per repo: COLIMA_SUFFIX=ying make colima → <repo>-ying.
# Resource knobs (env): COLIMA_CPUS/MEMORY/DISK/ARCH/VMTYPE — see colima.sh.
# amd64 testing: COLIMA_ARCH=x86_64 make colima (adds --vz-rosetta under vz).
#
# All Colima interaction lives in .colima/colima.sh; these recipes are thin
# wrappers (mirrors .openshell/sandbox.sh). Requires `brew install colima`.

# Profile name: sanitize like .openshell/Makefile (lowercase, '.'→'-').
COLIMA_PROFILE = $(shell echo "$(REPO_NAME)-$(or $(COLIMA_SUFFIX),test)" | tr '.' '-' | tr '[:upper:]' '[:lower:]')

# Workspace parent dir, bind-mounted writable into the VM (siblings adjacent
# so go.mod `replace ../peer` resolves). recursive (=) so realpath only runs
# when a colima target actually expands it, not on every `make`.
WORKSPACE_DIR = $(realpath $(CURDIR)/..)

.PHONY: colima colima-gui colima-stop colima-clean help-colima

help-colima:
	@printf '%s\n' \
	"Colima (Linux VMs):" \
	"  make colima            Start Linux profile $(COLIMA_PROFILE) (idempotent)," \
	"                         bind-mount the workspace writable, ssh in (lands in repo)." \
	"  make colima-gui        Same + provision a TigerVNC desktop (vnc://localhost:5901)." \
	"  make colima-stop       Stop $(COLIMA_PROFILE)." \
	"  make colima-clean      Stop + delete $(COLIMA_PROFILE) (base image stays cached)." \
	"" \
	"  Parallel profiles: COLIMA_SUFFIX=ying make colima → <repo>-ying." \
	"  amd64 guest: COLIMA_ARCH=x86_64 make colima. Knobs: COLIMA_CPUS/MEMORY/DISK." \
	""

colima:
	@.colima/colima.sh up   $(COLIMA_PROFILE) $(CURDIR) $(WORKSPACE_DIR)

colima-gui:
	@.colima/colima.sh gui  $(COLIMA_PROFILE) $(CURDIR) $(WORKSPACE_DIR)

colima-stop:
	@.colima/colima.sh stop $(COLIMA_PROFILE)

colima-clean:
	@.colima/colima.sh clean $(COLIMA_PROFILE)
```

- [ ] **Step 2: Verify variable derivation parses (no colima run)**

Run: `make -f .colima/Makefile help-colima REPO_NAME=Foo.Bar 2>&1 | head -3`
Expected: the help text shows profile `foo-bar-test` (sanitization works).

- [ ] **Step 3: Commit**

```bash
git add .colima/Makefile
git commit -m "#93: .colima/Makefile — thin make wrappers + help-colima"
```

### Task 4: Wire into the workflow includes, help, and base manifest

**Files:**
- Modify: `Makefile.workflow:8` (add `-include .colima/Makefile` after the `.tart` include)
- Modify: `Makefile:14-20` (add `help-colima` to the `help` aggregation + the comment)
- Modify: `construct/base.manifest` (add `.colima` symlinks + extend the walker comment)

- [ ] **Step 1: Add the workflow include**

In `Makefile.workflow`, after line 8 (`-include .tart/Makefile`), add:

```makefile

# Include colima targets if available (clean Linux VM testing — Apple Silicon)
-include .colima/Makefile
```

- [ ] **Step 2: Add `help-colima` to the top-level help**

In `Makefile`, change line 20 from:

```makefile
help: help-workflow help-sandbox help-tart
```
to:
```makefile
help: help-workflow help-sandbox help-tart help-colima
```

And extend the comment block above it (lines 14-19) so it names `.colima/Makefile` alongside `.openshell`/`.tart` as a help-fragment source. Add a line noting the transient window: a consumer that pulls the updated `Makefile` (with the `help-colima` prereq) *before* running `setup.sh` to materialize the new `.colima/Makefile` symlink will get a `No rule to make target 'help-colima'` until setup runs. This matches the existing accepted fragility for `help-tart`/`help-sandbox`; refreshed consumers are fine because Task 4 Step 3 adds `.colima` to `base.manifest`.

- [ ] **Step 3: Add base-manifest entries**

In `construct/base.manifest`, after the `# ── Tart …` block (the `symlink .tart/Makefile` / `symlink .tart/scripts` lines), add:

```
# ── Colima (Linux VM testing — Apple Silicon) ───────────────────────────────
symlink   .colima/Makefile
symlink   .colima/colima.sh
symlink   .colima/vnc-setup.sh
```

(Individual-file symlinks, mirroring the `.openshell` entries. The `list-peers.sh` walker comment needs no edit — colima deliberately doesn't use it; see divergence #1.)

- [ ] **Step 4: Verify the aggregated help resolves end-to-end**

Run: `make help 2>&1 | grep -A1 'Colima (Linux VMs)'`
Expected: the `make colima` help line appears (proves the include chain + help wiring).

- [ ] **Step 5: Commit**

```bash
git add Makefile Makefile.workflow construct/base.manifest
git commit -m "#93: wire .colima into workflow includes, help, and base.manifest"
```

### Task 5: Process-level fake test (deterministic, no real VM)

**Files:**
- Create: `.colima/test/colima.test.sh`

This is the constitution's process-level fake for the external `colima` dependency: a stub `colima` on `PATH` records argv; we assert each verb assembles the right command lines and honors idempotency gating. Mirrors the repo's existing `*.test.sh` shell-test style (e.g. `construct/scripts/test/bootstrap-transitive.test.sh`).

- [ ] **Step 1: Write the test**

```bash
#!/usr/bin/env bash
# colima.test.sh — drives .colima/colima.sh against a FAKE `colima` binary
# (records argv; canned status/list via env), asserting the command lines
# each verb builds. No real VM is started.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
LOG="$WORK/colima.log"

# Fake colima: append argv to LOG; emulate status/list from env.
mkdir -p "$WORK/bin"
cat > "$WORK/bin/colima" <<EOF
#!/usr/bin/env bash
echo "\$@" >> "$LOG"
case "\$1" in
  status) [ "\${FAKE_RUNNING:-0}" = 1 ] && exit 0 || exit 1 ;;
  list)   [ "\${FAKE_EXISTS:-0}" = 1 ] && echo '{"name":"p-test","status":"Stopped"}'; exit 0 ;;
  ssh)    exit 0 ;;   # non-interactive in test
  *)      exit 0 ;;
esac
EOF
chmod +x "$WORK/bin/colima"
export PATH="$WORK/bin:$PATH"

run() { : > "$LOG"; "$ROOT/.colima/colima.sh" "$@"; cat "$LOG"; }
assert_has()  { grep -q -- "$1" "$LOG" || { echo "FAIL: expected '$1' in:"; cat "$LOG"; exit 1; }; }
assert_none() { grep -q -- "$1" "$LOG" && { echo "FAIL: did not expect '$1' in:"; cat "$LOG"; exit 1; } || true; }

# up, profile stopped → starts with mount + arch + ssh
FAKE_RUNNING=0 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "start -p p-test"
assert_has "--mount /Users/me/ws:w"
assert_has "ssh -p p-test"

# up, profile already running → no start, just ssh (idempotent)
FAKE_RUNNING=1 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_none "start -p p-test"
assert_has  "ssh -p p-test"

# stop, running → stops
FAKE_RUNNING=1 run stop p-test >/dev/null
assert_has "stop -p p-test"

# stop, not running → no stop (idempotent)
FAKE_RUNNING=0 run stop p-test >/dev/null
assert_none "stop -p p-test"

# clean, exists → delete -f
FAKE_RUNNING=0 FAKE_EXISTS=1 run clean p-test >/dev/null
assert_has "delete -p p-test -f"

# clean, ABSENT → no delete (proves the _exists gate isn't a tautology: an
# empty `list -j` must suppress the delete)
FAKE_RUNNING=0 FAKE_EXISTS=0 run clean p-test >/dev/null
assert_none "delete -p p-test -f"

# amd64 → rosetta flag
FAKE_RUNNING=0 COLIMA_ARCH=x86_64 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "--vz-rosetta"

echo "PASS: .colima/colima.sh"
```

- [ ] **Step 2: Run it red-then-green**

Run: `bash .colima/test/colima.test.sh`
Expected: `PASS: .colima/colima.sh`. (If any assertion fails, fix `colima.sh`, not the test.)

- [ ] **Step 3: Commit**

```bash
git add .colima/test/colima.test.sh
git commit -m "#93: process-level fake test for .colima/colima.sh command assembly"
```

### Task 6: Real boot→ssh→stop→clean verification on this Mac

This is the integration verification the Done-when requires — captured in the issue Log. **Run these yourself and record output.**

- [ ] **Step 1: Boot + mount writability (R1, R2)**

Run: `make colima`
Inside the guest, run: `pwd` (expect the repo's host abs path), then `touch _colima_probe && rm _colima_probe && echo WRITABLE`.
Expected: lands in repo dir; `WRITABLE` prints (confirms the writable mount overlap resolved). Then `exit`.
- If R1 fails (read-only), apply the `colima start --edit` fallback and re-verify; note the resolution in the Log.

- [ ] **Step 2: Idempotent re-entry**

Run: `make colima` again.
Expected: `Profile 'p-test' already running.` then a shell — no second boot.

- [ ] **Step 3: Parallel profile via suffix**

Run: `COLIMA_SUFFIX=ying make colima` → boots a *second* profile `<repo>-ying`; `exit`.
Run: `colima list` → both profiles present.

- [ ] **Step 4: GUI / VNC (R3)**

Also capture the real `colima list -j` output here (`colima list -j`) and confirm it's compact JSONL matching `_exists`'s grep (R: pins the format assumption with evidence).

Run: `make colima-gui`; from another host shell: `nc -z localhost 5901 && echo FORWARDED`.
Run: `open vnc://localhost:5901` → Screen Sharing shows the fluxbox desktop.
- If R3 fails, apply the `portForwards` fallback; note it.

- [ ] **Step 5: Stop + clean idempotency**

Run: `make colima-stop` (the suffix one too: `COLIMA_SUFFIX=ying make colima-stop`), then `make colima-stop` again → second is a no-op.
Run: `make colima-clean` and `COLIMA_SUFFIX=ying make colima-clean` → profiles deleted; `colima list` empty (of these).

- [ ] **Step 6: Record evidence in the issue Log**

Append a `### 2026-06-12` block to `workshop/issues/000093-colima-vm-targets.md` `## Log` with the commands + key output (mount WRITABLE, idempotent skip, VNC FORWARDED, clean removal) and the resolution of R1–R4.

### Task 7: Atlas doc

**Files:**
- Create: `atlas/workflow/colima-vm.md`
- Modify: `atlas/index.md` (add a link, next to the tart/openshell entries)

- [ ] **Step 1: Write `atlas/workflow/colima-vm.md`** — mirror the structure of `atlas/workflow/openshell-sandbox.md`: what it is (Linux VM testbed, twin of tart), the four verbs, the profile-as-unit model, the live-mount vs tart-COW-clone distinction, the VNC path, and `COLIMA_SUFFIX`. Cross-link `base-layer.md` and the tart doc.

- [ ] **Step 2: Link it from `atlas/index.md`.**

- [ ] **Step 3: Commit**

```bash
git add atlas/workflow/colima-vm.md atlas/index.md
git commit -m "#93: atlas — document make colima (Linux VM testbed)"
```

---

## Close-out (after all tasks + verification)

Single review boundary (atomic feature, no `Mx` split per AGENTS.md §3): one `sdlc close`. The mandatory fresh-context review auto-dispatches at close.

```
sdlc close --issue 93 --verified '<boot→ssh→stop→clean cycle evidence + test PASS; R1–R4 resolved in Log>'
```

- `--actual` omitted → close computes it (active-time-v3); do not hand-type.
- Atlas updated (Task 7) → `--no-atlas` NOT needed.
- Then publish via `sdlc pr` → `sdlc merge`.

---

## Revisions

### 2026-06-12 — VNC auth: password (VncAuth), not `-SecurityTypes None`

- **Delta:** Task 2's `vnc-setup.sh` listing (and the integration-points note)
  specify `-SecurityTypes None` with "a password is a future extension." The
  shipped script instead writes `~/.vnc/passwd` via `vncpasswd -f` and starts
  `vncserver` with default VncAuth — default password `colima`, override
  `COLIMA_VNC_PASSWORD` (passed as a 2nd positional: `bash -s -- <geom> <pw>`).
- **Reason:** TigerVNC refuses an auth-less server on a non-localhost bind
  (`-localhost no`, needed so Lima forwards the port) without
  `--I-KNOW-THIS-IS-INSECURE`. A VNC password is cheaper defense-in-depth than
  shipping that flag in a propagating base-layer artifact, even though the port
  is only host-reachable via Lima's user-mode NAT. Verified in the real
  boot→gui cycle (guest `0.0.0.0:5901` → host `localhost:5901` forwarded).

### 2026-06-12 — `colima.sh` usage guard + test hardening (boundary-review minors)

- **Delta:** `up`/`gui`/`stop`/`clean` arms now guard arg count (`[ $# -ge N ]
  || _usage`) so direct-CLI misuse prints `usage:` instead of an opaque
  `unbound variable` under `set -u`. Fake test gained a `gui` command-assembly
  assertion and a missing-args exit-2 assertion. `vnc-setup.sh` notes the
  8-char VNC password truncation.
- **Reason:** FIX-THEN-SHIP boundary review (no Critical/Important) recommended
  these cheap hardening items.
