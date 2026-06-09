#!/usr/bin/env bash
# tart-vm-setup.sh — VM-side bootstrap for the tart-vm-* targets.
# Pushed to the VM by the Makefile (alongside tart-vm-rc.zsh); runs
# after every `make tart`. Idempotent.
#
# Argument: $1 is the current repo name (the repo whose Makefile.workflow
# the operator invoked). Used to write the current-repo marker so the rc
# can cd into ~/workspace/<repo>/ on shell start.
#
# Ordering rationale (ariadne: VM-setup network race): the network-FREE,
# essential work — wiring ~/workspace / ~/repo to the host mount — runs
# FIRST, so a DNS hiccup during the network-DEPENDENT steps (oh-my-zsh
# pulls its installer from github) can never strand the operator out of
# their workspace. The network steps then wait for DNS to come up and
# are non-fatal: they warn + continue rather than aborting under `set -e`.
set -euo pipefail

CURRENT_REPO="${1:-}"

# ── Wire ~/workspace to the host's workspace mount (network-free) ─
# Post-ariadne#32 the host mounts a workspace-shaped APFS clone holding
# the current repo + its go.mod-declared peers (recursive replace walk)
# into the VM. The contents look like a sibling-checkout: each peer is
# a top-level dir, so Go's replace directives pointing at sibling paths
# resolve correctly inside the VM.
#
# Mount appears at /Volumes/My Shared Files/workspace; symlink
# ~/workspace there so paths like ~/workspace/parley.nvim feel
# natural to operators. The set of peers under ~/workspace is
# determined by the current repo's go.mod — not the host's full
# sibling directory. This block needs no network, so it runs first.
MOUNT="/Volumes/My Shared Files/workspace"
if [ -d "$MOUNT" ]; then
    # Symlink ~/workspace → the writable share.
    if [ -d "$HOME/workspace" ] && [ ! -L "$HOME/workspace" ]; then
        echo "==> Removing old ~/workspace directory (pre-#32 artifact)..."
        rm -rf "$HOME/workspace"
    fi
    if [ ! -L "$HOME/workspace" ] || [ "$(readlink "$HOME/workspace")" != "$MOUNT" ]; then
        rm -f "$HOME/workspace"
        ln -s "$MOUNT" "$HOME/workspace"
        echo "==> Symlinked ~/workspace → $MOUNT"
    fi

    # Backward-compat: keep ~/repo pointing at the current repo inside
    # the workspace. Anything that hardcoded ~/repo (PATH entries,
    # legacy scripts) keeps working through the symlink chain.
    if [ -n "$CURRENT_REPO" ] && [ -d "$MOUNT/$CURRENT_REPO" ]; then
        if [ -d "$HOME/repo" ] && [ ! -L "$HOME/repo" ]; then
            echo "==> Removing old ~/repo directory (pre-#32 artifact)..."
            rm -rf "$HOME/repo"
        fi
        EXPECTED="$HOME/workspace/$CURRENT_REPO"
        if [ ! -L "$HOME/repo" ] || [ "$(readlink "$HOME/repo")" != "$EXPECTED" ]; then
            rm -f "$HOME/repo"
            ln -s "$EXPECTED" "$HOME/repo"
            echo "==> Symlinked ~/repo → $EXPECTED"
        fi
    fi

    # Marker file for the rc: which repo to cd into on shell start.
    if [ -n "$CURRENT_REPO" ]; then
        printf '%s\n' "$CURRENT_REPO" > "$HOME/.tart-current-repo"
    fi
elif [ -L "$HOME/workspace" ]; then
    # No mount but a leftover symlink — clean up.
    echo "==> No workspace mount found; removing stale ~/workspace symlink."
    rm "$HOME/workspace"
fi

# ── Wait for outbound DNS before the network-dependent steps ─────
# Right after boot, configd/mDNSResponder may not have applied the
# DHCP-provided resolver yet, so the first curl fails with "Could not
# resolve host: github.com". Kick the resolver once and poll briefly so
# that transient case self-heals instead of aborting setup. A genuine
# no-network VM falls through after the timeout with NET_OK=0 — the
# steps below then warn + skip (the workspace wiring above already
# succeeded, so the operator still lands in a working shell).
# Probe the ACTUAL dependency URL (returns 200) — a bare host hits a
# redirect/4xx that `-f` would misread as "not ready" on a healthy net.
net_ready() { curl -fsS --max-time 5 -o /dev/null \
    https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh 2>/dev/null; }
if net_ready; then
    NET_OK=1
else
    echo "==> Waiting for VM network/DNS..."
    sudo -n dscacheutil -flushcache         2>/dev/null || true
    sudo -n killall -HUP mDNSResponder      2>/dev/null || true
    NET_OK=0
    for _ in $(seq 1 15); do
        sleep 2
        if net_ready; then NET_OK=1; break; fi
    done
    [ "$NET_OK" = 1 ] && echo "==> Network up." \
                      || echo "  [warn] no outbound network after wait — skipping network steps."
fi

# ── oh-my-zsh (network; non-fatal) ───────────────────────────────
# The macOS-native counterpart to openshell's oh-my-bash. Installer
# replaces ~/.zshrc with a template that activates the framework
# (default theme robbyrussell, default plugins=(git)). Non-fatal: a
# network blip here must NOT abort setup, so a failure warns + continues
# under `set -e`. Runs BEFORE the rc-append below so its template
# overwrite can't clobber the appended source line.
if [ ! -d "$HOME/.oh-my-zsh" ]; then
    if [ "$NET_OK" = 1 ]; then
        echo "==> Installing oh-my-zsh..."
        CHSH=no RUNZSH=no sh -c \
            "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" \
            "" --unattended \
            || echo "  [warn] oh-my-zsh install failed — continuing without it."
    else
        echo "  [warn] skipping oh-my-zsh install (no network)."
    fi
fi

# ── Source the extension rc from ~/.zshrc ────────────────────────
# After oh-my-zsh (whose installer rewrites ~/.zshrc), so this append
# survives. Network-free; runs whether or not oh-my-zsh succeeded.
if ! grep -q tart-vm-rc.zsh "$HOME/.zshrc" 2>/dev/null; then
    cat >> "$HOME/.zshrc" <<'EOF'

# Extension rc managed by host Makefile (scripts/tart-vm-rc.zsh)
[ -f ~/.tart-vm-rc.zsh ] && source ~/.tart-vm-rc.zsh
EOF
fi

# ── Per-repo VM hooks (run-parts) ────────────────────────────────
# After standard setup, run the BOOTED repo's hook directory so a
# consumer can customize its VM without patching this base-layer
# script. Convention (ariadne#59):
#   - Location: ~/workspace/<repo>/.tart/vm-hooks.d/  (the repo whose
#     `make tart` was invoked, i.e. $CURRENT_REPO).
#   - Every *.sh, lexical LC_ALL=C order — use zero-padded NN- prefixes
#     (00-, 10-) to sequence deterministically. Run as `bash <hook> <repo>`.
#   - Runs on EVERY cold-boot (no run-once marker) → hooks MUST be idempotent.
#   - Continue-on-error: a failing hook prints a [warn] with its rc and the
#     loop continues, so a broken hook never strands the operator out of the
#     VM shell. The `|| echo …` form is also what keeps `set -e` (line 9)
#     from aborting the whole setup when a hook exits non-zero.
# Opt-in + additive: no hook dir → no-op, so existing consumers are unaffected.
HOOKS_DIR="$HOME/workspace/$CURRENT_REPO/.tart/vm-hooks.d"
if [ -n "$CURRENT_REPO" ] && [ -d "$HOOKS_DIR" ]; then
    # Iterate via a glob, not `ls` in $(): a command substitution word-splits
    # filenames, so a hook named with a space would silently never run (only
    # warn). The process-sub subshell pins LC_ALL=C for the documented lexical
    # order without leaking the locale into the hooks; `nullglob` makes an
    # empty/absent dir a clean no-op (and dodges `set -u` empty-array traps in
    # macOS bash 3.2); NUL-delimited so any filename survives intact.
    while IFS= read -r -d '' hook; do
        base="$(basename "$hook")"
        echo "==> vm-hook: $base"
        # Capture rc explicitly: a command substitution in the warn string
        # would otherwise reset $? before we read it. `|| rc=$?` also keeps
        # `set -e` from aborting the whole setup on a non-zero hook.
        rc=0
        bash "$hook" "$CURRENT_REPO" || rc=$?
        [ "$rc" -eq 0 ] || echo "  [warn] vm-hook $base failed (rc=$rc) — continuing"
    done < <(LC_ALL=C; shopt -s nullglob; for f in "$HOOKS_DIR"/*.sh; do printf '%s\0' "$f"; done)
fi

echo "==> VM setup complete."
