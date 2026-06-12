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
#
# bash 3.2-safe (macOS system bash): no arrays expanded under `set -u`.
set -euo pipefail

# Resource knobs (override via env / `make colima COLIMA_MEMORY=8`).
COLIMA_CPUS=${COLIMA_CPUS:-4}
COLIMA_MEMORY=${COLIMA_MEMORY:-4}
COLIMA_DISK=${COLIMA_DISK:-60}
COLIMA_ARCH=${COLIMA_ARCH:-aarch64}     # x86_64 → amd64 guest (adds --vz-rosetta under vz)
COLIMA_VMTYPE=${COLIMA_VMTYPE:-vz}      # Apple Virtualization.framework, same as tart
COLIMA_VNC_GEOMETRY=${COLIMA_VNC_GEOMETRY:-1600x1000}
COLIMA_VNC_PASSWORD=${COLIMA_VNC_PASSWORD:-colima}   # VNC desktop password

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

# Shared VM logging (construct/scripts/vm-log.sh): colorized step headers +
# a dim filter for underlying-process output. Same helper .tart uses (ARCH-DRY).
VMLOG="$SCRIPT_DIR/../construct/scripts/vm-log.sh"
step() { bash "$VMLOG" step "$@"; }
dim()  { bash "$VMLOG" dim; }

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
        step "Profile '$profile' already running."
        return
    fi
    step "Starting Colima profile '$profile' (Linux VM; mounts $mountdir writable)..."
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
    colima start "$@" 2>&1 | dim
}

# Push the cheap rc on every boot (so host edits to vm-rc.sh propagate); run the
# heavier guest bootstrap (symlinks, nvim, oh-my-bash, hooks) only on a FRESH
# start ($4=fresh) — the per-repo profile's symlinks/marker don't change across
# re-entries. Caller makes this non-fatal so a setup hiccup never strands the
# operator out of a running VM.
_run_setup() {
    local profile=$1 repodir=$2 mountdir=$3 fresh=${4:-1}
    colima ssh -p "$profile" -- bash -c 'cat > ~/.colima-vm-rc.sh' < "$SCRIPT_DIR/vm-rc.sh"
    [ "$fresh" = 1 ] || return 0
    step "Provisioning guest (symlinks, nvim, oh-my-bash, dev-aliases; first run is slower)..."
    colima ssh -p "$profile" -- bash -s -- "$repodir" "$mountdir" \
        < "$SCRIPT_DIR/vm-setup.sh" 2>&1 | dim
}

# Interactive shell, landing in the repo dir (mounted at its host abs path
# inside the guest). `|| true`: the remote login shell's exit status on
# `exit` is a normal session-end, not a make failure.
_ssh_into() {
    local profile=$1 repodir=$2
    echo "==> colima ssh -p $profile  (landing in $repodir)"
    colima ssh -p "$profile" -- bash -lc "cd \"$repodir\" 2>/dev/null || true; exec bash -il" || true
}

_provision_vnc() {
    local profile=$1
    step "Provisioning VNC desktop in '$profile' (idempotent; first run apt-installs)..."
    # Geometry as a positional arg (bash -s -- ARG), NOT an env var: `colima ssh
    # -- cmd` execve's the command directly, so a `VAR=val cmd` prefix would be
    # parsed as the command name. `bash -s -- 1600x1000` lands in the guest's $1.
    colima ssh -p "$profile" -- bash -s -- "$COLIMA_VNC_GEOMETRY" "$COLIMA_VNC_PASSWORD" < "$SCRIPT_DIR/vnc-setup.sh" 2>&1 | dim
    step "VNC ready. Connect from the Mac with:"
    echo "        open vnc://localhost:5901          # macOS Screen Sharing"
    echo "    password: $COLIMA_VNC_PASSWORD  (override via COLIMA_VNC_PASSWORD)."
    echo "    Lima forwards guest 5901 → host 5901."
}

_usage() {
    echo "usage: colima.sh up|gui <profile> <repo-dir> <mount-dir> | stop|clean <profile>" >&2
    exit 2
}

cmd=${1:-}; shift || true
case "$cmd" in
  up)
    _need_colima; [ $# -ge 3 ] || _usage
    _running "$1" && fresh=0 || fresh=1
    _start_if_needed "$1" "$3"
    _run_setup "$1" "$2" "$3" "$fresh" || step "guest setup incomplete — continuing to shell"
    _ssh_into "$1" "$2" ;;
  gui)
    _need_colima; [ $# -ge 3 ] || _usage
    _running "$1" && fresh=0 || fresh=1
    _start_if_needed "$1" "$3"
    _run_setup "$1" "$2" "$3" "$fresh" || step "guest setup incomplete — continuing to shell"
    _provision_vnc "$1"; _ssh_into "$1" "$2" ;;
  stop)
    _need_colima; [ $# -ge 1 ] || _usage
    if _running "$1"; then step "Stopping '$1'..."; colima stop -p "$1" 2>&1 | dim;
    else step "'$1' not running."; fi ;;
  clean)
    _need_colima; [ $# -ge 1 ] || _usage
    if _running "$1"; then step "Stopping '$1'..."; colima stop -p "$1" 2>&1 | dim; fi
    if _exists "$1"; then step "Deleting profile '$1' (base image stays cached)..."; colima delete -p "$1" -f 2>&1 | dim;
    else step "'$1' does not exist."; fi ;;
  *)
    _usage ;;
esac
