#!/usr/bin/env bash
# vm-setup.sh — guest bootstrap for `make colima`, run via
# `colima ssh -- bash -s -- <repo-dir> <mount-dir>`. Idempotent. The Linux
# counterpart to .tart/scripts/tart-vm-setup.sh; mirrors the portable pieces of
# the OpenShell overlay, omitting the sandbox proxy/creds + macOS bits.
#
# colima.sh runs this only on a FRESH start (the per-repo profile's symlinks/
# marker don't change across re-entries); the cheap vm-rc.sh is re-pushed every
# boot. colima.sh also makes this non-fatal — a hiccup here must never strand
# the operator out of a VM that's actually up.
set -euo pipefail
REPO_DIR=${1:?repo-dir required}
MOUNT_DIR=${2:?mount-dir required}
REPO_NAME=$(basename "$REPO_DIR")

# ── Workspace/repo symlinks (network-free; first so a net hiccup can't strand) ─
# Colima mounts the host workspace at its literal host path inside the guest;
# ~/workspace → that mount and ~/repo → the current repo so paths and the
# dev-aliases (which look under ~/workspace/ariadne) resolve. rm a real dir
# first so `ln -sfn` can't nest the link inside it.
for pair in "workspace:$MOUNT_DIR" "repo:$REPO_DIR"; do
    link="$HOME/${pair%%:*}"; target="${pair#*:}"
    [ -e "$link" ] && [ ! -L "$link" ] && rm -rf "$link"
    ln -sfn "$target" "$link"
done
printf '%s\n' "$REPO_NAME" > "$HOME/.colima-current-repo"

# ── neovim (apt; idempotent; NON-FATAL) ──────────────────────────────────────
# Must not abort under `set -e` on a network hiccup: the network-free
# rc-append below is the load-bearing step (aliases/auto-cd), and because the
# profile is per-repo, a fresh boot that died here would be skipped on the next
# `make colima` (fresh=0) — silently stranding the VM without aliases until a
# colima-clean. So warn + continue, matching the oh-my-bash block.
if ! command -v nvim >/dev/null 2>&1; then
    echo "installing neovim..."
    if ! { sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq \
        && sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq neovim >/dev/null; }; then
        echo "[warn] neovim install failed — continuing."
    fi
fi

# ── oh-my-bash (network; non-fatal; BEFORE the rc-append so it survives) ──────
if [ ! -d "$HOME/.oh-my-bash" ]; then
    echo "installing oh-my-bash..."
    bash -c "$(curl -fsSL https://raw.githubusercontent.com/ohmybash/oh-my-bash/master/tools/install.sh)" \
        --unattended || echo "[warn] oh-my-bash install failed — continuing."
fi

# ── Source the extension rc from ~/.bashrc (after oh-my-bash rewrites it) ─────
if ! grep -q '.colima-vm-rc.sh' "$HOME/.bashrc" 2>/dev/null; then
    cat >> "$HOME/.bashrc" <<'EOF'

# Extension rc managed by host Makefile (.colima/vm-rc.sh)
[ -f ~/.colima-vm-rc.sh ] && source ~/.colima-vm-rc.sh
EOF
fi

# ── Per-repo VM hooks (run-parts; idempotent; continue-on-error) ─────────────
HOOKS_DIR="$HOME/workspace/$REPO_NAME/.colima/vm-hooks.d"
if [ -d "$HOOKS_DIR" ]; then
    for hook in "$HOOKS_DIR"/*.sh; do
        [ -e "$hook" ] || continue
        echo "vm-hook: $(basename "$hook")"
        rc=0; bash "$hook" "$REPO_NAME" || rc=$?
        [ "$rc" -eq 0 ] || echo "[warn] vm-hook $(basename "$hook") failed (rc=$rc)"
    done
fi

echo "VM setup complete."
