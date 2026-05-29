#!/bin/bash
# Sandbox lifecycle management for OpenShell + mutagen.
# Called by Makefile targets: sandbox, sandbox-build, sandbox-stop
set -euo pipefail

# Make Ctrl+C quit on the first press. The build/connect path runs several
# `while … sleep 2` poll loops plus background `&` jobs; without a trap, SIGINT
# only kills the in-flight `sleep`/subprocess and the loop resumes on the next
# iteration (and `|| true` masks signal-kills), so a single Ctrl+C appeared to
# do nothing. This handler reaps any background jobs and exits immediately.
on_interrupt() {
    trap - INT TERM
    echo ""
    echo "Interrupted — stopping background jobs and exiting."
    local pids
    pids="$(jobs -p 2>/dev/null || true)"
    if [ -n "$pids" ]; then
        # shellcheck disable=SC2086 -- word-splitting is intentional (PID list)
        kill $pids 2>/dev/null || true
    fi
    exit 130
}
trap on_interrupt INT TERM

# Auto-detect Docker socket for Docker Desktop (macOS uses a non-standard path)
if [ -z "${DOCKER_HOST:-}" ] && [ ! -S /var/run/docker.sock ]; then
    for sock in "$HOME/.docker/run/docker.sock" "$HOME/.docker/desktop/docker.sock"; do
        if [ -S "$sock" ]; then
            export DOCKER_HOST="unix://$sock"
            break
        fi
    done
fi

ACTION="${1:-}"
SANDBOX_NAME="${2:-}"
SANDBOX_SSH_HOST="openshell-${SANDBOX_NAME}"

# Pre-flight checks: validate prerequisites before doing real work.
# Only runs for actions that need the full stack (build, connect, clean).
preflight() {
    local failed=0

    # 1. openshell CLI
    if ! command -v openshell >/dev/null 2>&1; then
        echo "ERROR: 'openshell' CLI not found in PATH."
        echo "  Install it per OpenShell docs, then retry."
        failed=1
    fi

    # 2. Docker daemon — auto-start on macOS
    if ! docker info >/dev/null 2>&1; then
        if [ "$(uname)" = "Darwin" ] && [ -d "/Applications/Docker.app" ]; then
            echo "  Docker not running — starting Docker Desktop..."
            open -a Docker
            local retries=0
            while ! docker info >/dev/null 2>&1; do
                retries=$((retries + 1))
                if [ "$retries" -ge 30 ]; then
                    echo "ERROR: Docker Desktop did not start within 60s."
                    failed=1
                    break
                fi
                sleep 2
            done
            if [ "$retries" -lt 30 ]; then
                echo "  Docker Desktop started."
            fi
        else
            echo "ERROR: Docker is not running or not accessible."
            echo "  Start Docker Desktop, then retry."
            failed=1
        fi
    fi

    # 3. mutagen CLI
    if ! command -v mutagen >/dev/null 2>&1; then
        echo "ERROR: 'mutagen' CLI not found in PATH."
        echo "  Install: brew install mutagen-io/mutagen/mutagen"
        failed=1
    fi

    # 4. OpenShell gateway — auto-restart if unreachable
    if command -v openshell >/dev/null 2>&1; then
        if ! openshell sandbox list >/dev/null 2>&1; then
            echo "  OpenShell gateway not reachable — restarting..."
            openshell gateway destroy --name openshell 2>/dev/null || true
            if openshell gateway start; then
                echo "  OpenShell gateway started."
            else
                echo "ERROR: Failed to start OpenShell gateway."
                failed=1
            fi
        fi
    fi

    if [ "$failed" -ne 0 ]; then
        echo ""
        echo "Pre-flight checks failed. Fix the above issues and retry."
        exit 1
    fi
}

PLENARY_HOST="${HOME}/.local/share/nvim/lazy/plenary.nvim"
# All paths derive from $0. .openshell/ is a real dir in every repo (even when
# its contents are symlinks), so dirname/$0/.. always gives the local repo.
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
REPO_NAME_LOCAL="$(basename "$REPO_DIR")"
WORKSPACE_DIR="$(cd "$REPO_DIR/.." && pwd)"
SCRIPT_DIR="$REPO_DIR/.openshell"

# Use Apple's system SSH/SCP throughout — Homebrew openssh lacks macOS-specific
# options (UseKeychain) and causes "Bad configuration option: usekeychain" errors.
# Write sandbox Host blocks to ~/.ssh/config so the mutagen daemon (a separate
# long-lived process) can reach all sandboxes without MUTAGEN_SSH_PATH.
SSH_CONFIG="$HOME/.ssh/config"
SSH="/usr/bin/ssh"
SCP="/usr/bin/scp"
BASE_IMAGE="ghcr.io/nvidia/openshell-community/sandboxes/base"
DIGEST_FILE="$SCRIPT_DIR/.base-image-digest"

# Fetch the latest digest of the base image from GHCR.
# Returns empty string on any failure (network, auth, etc).
fetch_remote_digest() {
    local token digest
    token=$(curl -sf --max-time 5 \
        "https://ghcr.io/token?service=ghcr.io&scope=repository:nvidia/openshell-community/sandboxes/base:pull" \
        | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))" 2>/dev/null) || true
    if [ -z "$token" ]; then return; fi
    digest=$(curl -sf --max-time 5 \
        -H "Authorization: Bearer $token" \
        -H "Accept: application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.v2+json" \
        -o /dev/null -w '' \
        --dump-header /dev/stdout \
        "https://ghcr.io/v2/nvidia/openshell-community/sandboxes/base/manifests/latest" \
        | grep -i 'docker-content-digest' | awk '{print $2}' | tr -d '\r') || true
    echo "$digest"
}

save_digest() {
    local digest
    digest=$(fetch_remote_digest)
    if [ -n "$digest" ]; then
        echo "$digest" > "$DIGEST_FILE"
    fi
}

# Check if the base image has been updated since the sandbox was created.
# Prints a warning and prompts user if an update is available.
# Returns 0 to continue, 1 if user wants to rebuild.
check_base_image_update() {
    local saved_digest remote_digest
    remote_digest=$(fetch_remote_digest)
    if [ -z "$remote_digest" ]; then return 0; fi  # can't reach registry, continue
    if [ ! -f "$DIGEST_FILE" ]; then
        echo "  Seeding base image digest for future update checks."
        echo "$remote_digest" > "$DIGEST_FILE"
        return 0
    fi
    saved_digest=$(cat "$DIGEST_FILE")
    if [ "$saved_digest" = "$remote_digest" ]; then
        echo "  Base image is up to date."
        return 0
    fi
    echo ""
    echo "  ** Base image update available **"
    echo "  Current: ${saved_digest:0:19}..."
    echo "  Latest:  ${remote_digest:0:19}..."
    echo ""
    printf "  Recreate sandbox with new base image? [y/N] "
    read -r answer </dev/tty
    if [[ "$answer" =~ ^[Yy] ]]; then
        echo "==> Rebuilding sandbox with updated base image..."
        cleanup
        return 1
    fi
    echo "  Continuing with current sandbox."
    return 0
}

get_phase() {
    openshell sandbox list 2>/dev/null \
        | sed 's/\x1b\[[0-9;]*m//g' \
        | grep "$SANDBOX_NAME" \
        | awk '{print $NF}' || true
}

ensure_ssh_config() {
    # Upsert this sandbox's Host block into ~/.ssh/config.
    # Each sandbox gets a BEGIN/END-delimited block; other sandboxes' blocks are preserved.
    local marker_begin="# BEGIN openshell-${SANDBOX_NAME}"
    local marker_end="# END openshell-${SANDBOX_NAME}"
    local new_block
    new_block=$(cat <<SSHEOF
${marker_begin}
$(openshell sandbox ssh-config "$SANDBOX_NAME")
    ServerAliveInterval 60
    ServerAliveCountMax 3
${marker_end}
SSHEOF
    )

    mkdir -p "$HOME/.ssh"
    touch "$SSH_CONFIG"
    chmod 600 "$SSH_CONFIG"

    # Remove old block if present, then append new one
    if grep -qF "$marker_begin" "$SSH_CONFIG" 2>/dev/null; then
        sed -i.bak "/${marker_begin}/,/${marker_end}/d" "$SSH_CONFIG"
        rm -f "${SSH_CONFIG}.bak"
    fi
    echo "$new_block" >> "$SSH_CONFIG"
}

ensure_bootstrap_sync() {
    local bootstrap_dir="$SCRIPT_DIR/.bootstrap"
    ensure_sync bootstrap "$bootstrap_dir" /tmp/bootstrap one-way-replica --ignore-vcs
}

ensure_setup() {
    # Re-run post-install if any expected tool is missing
    if ! $SSH "$SANDBOX_SSH_HOST" "test -x \$HOME/.local/bin/nvim && test -x \$HOME/.local/bin/zellij" 2>/dev/null; then
        echo "==> Running post-install on sandbox..."
        mutagen sync flush "${SANDBOX_NAME}-bootstrap" 2>/dev/null || true
        $SCP -q "$SCRIPT_DIR/overlay/post-install.sh" "$SANDBOX_SSH_HOST:/tmp/post-install.sh"
        $SSH "$SANDBOX_SSH_HOST" "bash /tmp/post-install.sh"
    fi
    apply_config
}

# Gather host credentials into bootstrap cache so mutagen syncs them to sandbox.
gather_credentials() {
    local creds_dir="$SCRIPT_DIR/.bootstrap/credentials"
    mkdir -p "$creds_dir"
    if [ -f "$HOME/.codex/auth.json" ]; then
        cp "$HOME/.codex/auth.json" "$creds_dir/codex-auth.json"
    fi
}

# Apply setup.sh, dotfiles, and credentials to sandbox. Idempotent — safe to re-run.
apply_config() {
    echo "==> Applying config to sandbox..."
    gather_credentials
    mutagen sync flush "${SANDBOX_NAME}-bootstrap" 2>/dev/null || true
    local host_tz
    host_tz=$(readlink /etc/localtime 2>/dev/null | sed 's|.*/zoneinfo/||' || echo "UTC")
    $SCP -q "$SCRIPT_DIR/overlay/setup.sh" "$SANDBOX_SSH_HOST:/tmp/setup.sh"
    $SSH "$SANDBOX_SSH_HOST" "HOST_TZ='$host_tz' bash /tmp/setup.sh"

    local git_name git_email
    git_name=$(git config user.name 2>/dev/null || true)
    git_email=$(git config user.email 2>/dev/null || true)
    if [ -n "$git_name" ]; then
        $SSH "$SANDBOX_SSH_HOST" "git config --global user.name '$git_name'" 2>/dev/null || true
    fi
    if [ -n "$git_email" ]; then
        $SSH "$SANDBOX_SSH_HOST" "git config --global user.email '$git_email'" 2>/dev/null || true
    fi

    # Copy dotfiles to sandbox
    echo "  Copying dotfiles..."
    $SSH "$SANDBOX_SSH_HOST" "mkdir -p ~/.config/zellij/layouts"
    $SCP -q "$SCRIPT_DIR/dotfiles/zellij/config.kdl" "$SANDBOX_SSH_HOST:~/.config/zellij/config.kdl"
    $SCP -q "$SCRIPT_DIR/dotfiles/zellij/layouts/default.kdl" "$SANDBOX_SSH_HOST:~/.config/zellij/layouts/default.kdl"
    $SCP -q "$SCRIPT_DIR/dotfiles/zellij/clock.sh" "$SANDBOX_SSH_HOST:~/.config/zellij/clock.sh"
    $SSH "$SANDBOX_SSH_HOST" "chmod +x ~/.config/zellij/clock.sh"

    # Forward GitHub CLI auth from host to sandbox (write config directly — fast)
    local gh_token
    gh_token=$(gh auth token 2>/dev/null || true)
    if [ -n "$gh_token" ]; then
        echo "  Forwarding gh auth to sandbox..."
        $SSH "$SANDBOX_SSH_HOST" "mkdir -p ~/.config/gh && cat > ~/.config/gh/hosts.yml" <<EOF
github.com:
    oauth_token: ${gh_token}
    user: $(gh api user --jq .login 2>/dev/null || echo "")
    git_protocol: https
EOF
    fi
}

# Create a mutagen sync if it doesn't already exist.
# Usage: ensure_sync <label> <local_path> <remote_path> <mode> [extra_args...]
# Flushes after creation for one-way-replica and two-way-resolved modes.
ensure_sync() {
    local label="$1" local_path="$2" remote_path="$3" mode="$4"
    shift 4
    local sync_name="${SANDBOX_NAME}-${label}"

    if mutagen sync list "$sync_name" >/dev/null 2>&1; then
        return
    fi
    echo "  Starting ${label} sync..."
    mutagen sync create \
        --name "$sync_name" \
        --mode "$mode" \
        "$@" \
        "$local_path" "${SANDBOX_SSH_HOST}:${remote_path}" || true
    mutagen sync flush "$sync_name" 2>/dev/null || true
}

# ── go.mod peer sync (ariadne#44) ─────────────────────────────────────────────
# The sandbox syncs the current repo + its transitive construct/go.mod peers
# (not the whole parent workspace), into ~/workspace/<name> — mirroring the host
# layout, same model as `make tart`. Peers are read-only by default; SYNC=
# opts a peer into two-way writable.

# Resolve SYNC=../a,../b (relative to REPO_DIR) into absolute, physical paths.
sync_extra_paths() {
    local IFS=',' raw abs
    for raw in ${SYNC:-}; do
        [ -z "$raw" ] && continue
        abs="$(cd "$REPO_DIR/$raw" 2>/dev/null && pwd -P || true)"
        [ -z "$abs" ] && abs="$(cd "$raw" 2>/dev/null && pwd -P || true)"
        [ -n "$abs" ] && printf '%s\n' "$abs"
    done
}

# Emit the sync set as "mode<TAB>abspath<TAB>name" lines. The current repo and
# every SYNC= entry are rw (two-way); all other go.mod peers are ro (one-way).
# list-peers.sh provides membership (current repo + transitive peers + SYNC=
# extras as union seeds); writability classification is decided here.
compute_sync_set() {
    local repo_canon extras list_peers list rwset p name mode
    repo_canon="$(cd "$REPO_DIR" && pwd -P)"
    extras="$(sync_extra_paths)"
    list_peers="$REPO_DIR/construct/scripts/list-peers.sh"
    if [ -x "$list_peers" ]; then
        # shellcheck disable=SC2086 -- extras is a newline list; word-split as args
        list="$("$list_peers" "$repo_canon" $extras 2>/dev/null)" || list="$repo_canon"
    else
        # Pre-#44 derivative not yet refreshed (no list-peers symlink): degrade
        # to current-repo-only, like tart's no-go.mod single-repo fallback.
        echo "  (list-peers.sh not found — syncing current repo only; run 'make refresh')" >&2
        list="$repo_canon"
    fi
    rwset=" $repo_canon "
    while IFS= read -r p; do [ -n "$p" ] && rwset="$rwset$p "; done <<<"$extras"
    local seen_names=" "
    while IFS= read -r p; do
        [ -z "$p" ] && continue
        name="$(basename "$p")"
        # Peers map to ~/workspace/<basename> and to sync names keyed on
        # <basename>. Two repos with the same basename in different parents
        # (e.g. workspace/foo + vendor/foo via a replace/SYNC=) would collide on
        # both — and ensure_sync's idempotency would silently drop the second
        # (or worse, share a rw remote). Fail loudly instead of merging.
        case "$seen_names" in
            *" $name "*)
                echo "ERROR: peer basename collision '$name' — two repos in the peer set share a basename and would both map to ~/workspace/$name. Rename or exclude one." >&2
                return 1 ;;
        esac
        seen_names="$seen_names$name "
        if [[ "$rwset" == *" $p "* ]]; then mode=rw; else mode=ro; fi
        printf '%s\t%s\t%s\n' "$mode" "$p" "$name"
    done <<<"$list"
}

# All mutagen sync names this sandbox currently owns (prefix-scoped). Excludes
# the bootstrap sync, which the cleanup paths manage explicitly.
# NOTE: the prefix is `${SANDBOX_NAME}-`, so sandbox names must not be prefixes
# of one another (a sandbox `foo` would match `foo-bar-*` and could reconcile-
# terminate sandbox `foo-bar`'s syncs). SANDBOX_NAME = "<repo>-sandbox", and repo
# names aren't prefix-nested in practice; documented so it stays that way.
list_owned_syncs() {
    mutagen sync list --template '{{range .}}{{.Name}}{{"\n"}}{{end}}' 2>/dev/null \
        | grep "^${SANDBOX_NAME}-" \
        | grep -v "^${SANDBOX_NAME}-bootstrap$" || true
}

# Terminate any owned sync whose name is not in the desired set ($@). Drives
# RO→RW upgrades, RW→RO downgrades, dropped peers, and migration off the old
# fixed names (repo/git/workspace) — all as a pure name-set diff.
reconcile_syncs() {
    local n
    for n in $(list_owned_syncs); do
        case " $* " in
            *" $n "*) ;;  # still desired — keep
            *) echo "  reconcile: terminating stale sync $n"; mutagen sync terminate "$n" 2>/dev/null || true ;;
        esac
    done
}

ensure_mutagen_sync() {
    local desired=() mode path name remote

    # Capture the set first (not a process substitution) so a collision/error
    # exit from compute_sync_set aborts the build instead of being swallowed by
    # a subshell. Then feed the loop on fd 3.
    local sync_set
    sync_set="$(compute_sync_set)" || { echo "==> Aborting: peer set error (see above)." >&2; exit 1; }

    echo "==> Sync plan (current repo + construct/go.mod peers):"
    # Read on fd 3, not stdin: ssh/mutagen inside the loop would otherwise
    # consume the here-string from stdin and truncate the loop after the first
    # peer (the classic "ssh eats the while-read pipe" bug).
    while IFS=$'\t' read -r mode path name <&3; do
        [ -z "$name" ] && continue
        remote="/sandbox/workspace/$name"
        # mutagen creates the sync-root leaf but NOT missing parents: the beta
        # root /sandbox/workspace/<name> needs /sandbox/workspace to exist, and
        # the rw .git sync's parent is /sandbox/workspace/<name>. Pre-create the
        # peer dir so neither hits a beta "transition problem" (0 files synced).
        # -n: don't read stdin (belt-and-suspenders with the fd-3 read).
        $SSH -n "$SANDBOX_SSH_HOST" "mkdir -p '$remote'" 2>/dev/null || true
        if [ "$mode" = rw ]; then
            printf '    %-24s ~/workspace/%-20s writable\n' "$name" "$name"
            ensure_sync "peer-${name}-rw" "$path" "$remote" two-way-resolved \
                --ignore-vcs \
                --ignore node_modules \
                --ignore .test-home --ignore .test-xdg --ignore .test-tmp
            desired+=("${SANDBOX_NAME}-peer-${name}-rw")
            # Writable repos get .git (one-way host→sandbox) so in-sandbox git
            # ops have history; share commits by pushing, not via mutagen
            # (two-way .git is conflict-prone — see issue #44 decision 5).
            ensure_sync "peergit-${name}" "$path/.git" "$remote/.git" one-way-replica \
                --ignore "index.lock"
            desired+=("${SANDBOX_NAME}-peergit-${name}")
        else
            printf '    %-24s ~/workspace/%-20s read-only\n' "$name" "$name"
            ensure_sync "peer-${name}-ro" "$path" "$remote" one-way-replica \
                --ignore-vcs \
                --ignore node_modules \
                --ignore .test-home --ignore .test-xdg --ignore .test-tmp
            desired+=("${SANDBOX_NAME}-peer-${name}-ro")
        fi
    done 3<<<"$sync_set"

    # Host-mirroring layout: ~/workspace is the peer tree; ~/repo → current repo;
    # marker drives the shell's auto-cd on login (see overlay/setup.sh). In this
    # base image $HOME *is* /sandbox, so ~/workspace already equals the sync
    # parent /sandbox/workspace — symlinking it would create ~/workspace → itself
    # (a name == target loop), so skip on PATH equality (a string test, since the
    # dir may not be mutagen-populated yet — an -ef inode test would falsely
    # fall through). Only alias when $HOME differs (e.g. /home/sandbox). ~/repo
    # always links to the current repo under the workspace tree.
    $SSH "$SANDBOX_SSH_HOST" \
        "[ \"\$HOME/workspace\" = /sandbox/workspace ] || ln -sfn /sandbox/workspace \"\$HOME/workspace\"; \
         { [ -d \"\$HOME/repo\" ] && [ ! -L \"\$HOME/repo\" ] && rm -rf \"\$HOME/repo\"; }; \
         ln -sfn /sandbox/workspace/$REPO_NAME_LOCAL \"\$HOME/repo\"; \
         printf '%s\n' '$REPO_NAME_LOCAL' > \"\$HOME/.sandbox-current-repo\"" 2>/dev/null || true

    # ── Orthogonal syncs (not peers; kept as-is) ─────────────────────────────
    mkdir -p "$REPO_DIR/../worktree"
    ensure_sync worktree "$REPO_DIR/../worktree" /sandbox/worktree two-way-resolved \
        --ignore-vcs
    desired+=("${SANDBOX_NAME}-worktree")

    local nvim_state="${HOME}/.local/state/nvim"
    if [ -d "$nvim_state" ]; then
        ensure_sync nvim-state "$nvim_state" /sandbox/.local/state/nvim one-way-replica \
            --ignore-vcs
        desired+=("${SANDBOX_NAME}-nvim-state")
    fi

    if [ -d "$PLENARY_HOST" ]; then
        ensure_sync plenary "$PLENARY_HOST" /sandbox/.local/share/nvim/lazy/plenary.nvim one-way-replica \
            --ignore-vcs
        desired+=("${SANDBOX_NAME}-plenary")
    fi

    # Claude Code sessions: bi-directional so sessions can be resumed across
    # host and sandbox (use `claude --resume <session-id>`).
    local claude_projects="${HOME}/.claude/projects"
    if [ -d "$claude_projects" ]; then
        $SSH "$SANDBOX_SSH_HOST" "mkdir -p /sandbox/.claude/projects" 2>/dev/null || true
        ensure_sync claude-sessions "$claude_projects" /sandbox/.claude/projects two-way-resolved \
            --ignore-vcs
        desired+=("${SANDBOX_NAME}-claude-sessions")
    fi

    # Drive the declarative reconcile: drop anything we own that's no longer
    # desired (stale peers, mode flips, old fixed-name syncs from pre-#44).
    # The +-guard keeps `set -u` happy if `desired` is ever empty (it isn't
    # today — current repo is always rw + worktree unconditional — but don't
    # let a future refactor turn that into an unbound-variable abort).
    reconcile_syncs "${desired[@]+"${desired[@]}"}"
}

# Terminate every sync this sandbox owns (prefix-scoped). Replaces the old
# static SYNC_NAMES list — correct regardless of how many peers / SYNC= extras
# existed (issue #44 decision 7).
terminate_all_syncs() {
    local n
    for n in $(list_owned_syncs); do
        mutagen sync terminate "$n" 2>/dev/null || true
    done
}

# Light cleanup: terminate syncs and wipe sandbox working dirs, but keep the
# sandbox container and installed tools. Fast way to get a fresh repo state.
soft_cleanup() {
    echo "==> Soft cleanup (keeping sandbox + tools)..."
    terminate_all_syncs
    # Wipe the synced working trees. Post-#44 repos live under
    # /sandbox/workspace/<name> (not the retired /sandbox/repo); ~/repo is just a
    # symlink into it. mutagen recreates the peer dirs on re-sync below.
    $SSH "$SANDBOX_SSH_HOST" "rm -rf /sandbox/workspace /sandbox/worktree && mkdir -p /sandbox/workspace /sandbox/worktree" 2>/dev/null || true
    echo "==> Re-syncing files..."
    ensure_mutagen_sync
    echo "==> Re-applying config..."
    apply_config
    echo "==> Clean done."
}

# Full cleanup: destroy everything including the sandbox container.
cleanup() {
    mutagen sync terminate "${SANDBOX_NAME}-bootstrap" 2>/dev/null || true
    terminate_all_syncs
    openshell sandbox delete "$SANDBOX_NAME" 2>/dev/null || true
    # Remove this sandbox's block from the shared SSH config
    local marker_begin="# BEGIN openshell-${SANDBOX_NAME}"
    local marker_end="# END openshell-${SANDBOX_NAME}"
    if [ -f "$SSH_CONFIG" ] && grep -qF "$marker_begin" "$SSH_CONFIG" 2>/dev/null; then
        sed -i.bak "/${marker_begin}/,/${marker_end}/d" "$SSH_CONFIG"
        rm -f "${SSH_CONFIG}.bak"
    fi
}

# Nuclear cleanup: full cleanup + wipe bootstrap cache so deps are re-downloaded.
cleanup_nuke() {
    cleanup
    local bootstrap_dir="$SCRIPT_DIR/.bootstrap"
    if [ -d "$bootstrap_dir" ]; then
        echo "==> Removing bootstrap cache..."
        rm -rf "$bootstrap_dir"
    fi
}

timer_start() { date +%s; }
timer_show() { echo "    (${1}: $(( $(date +%s) - $2 ))s)"; }

# Ensure sandbox exists and is fully set up. Idempotent.
cmd_build() {
    preflight
    local phase t0 t_total
    t_total=$(timer_start)
    phase=$(get_phase)

    # Clean up broken state
    if [ -n "$phase" ] && [ "$phase" != "Running" ] && [ "$phase" != "Ready" ]; then
        echo "==> Sandbox in bad state ($phase), cleaning up..."
        cleanup
        phase=""
    fi

    # Check for base image updates on existing sandbox
    if [ -n "$phase" ]; then
        if ! check_base_image_update; then
            phase=""  # user chose to rebuild
        fi
    fi

    # Create sandbox and download deps in parallel
    if [ -z "$phase" ]; then
        echo "==> Creating sandbox + downloading deps (parallel)..."
        t0=$(timer_start)
        openshell sandbox create \
            --name "$SANDBOX_NAME" \
            --from base \
            --policy .openshell/policy.yaml \
            --auto-providers \
            -- true &
        local sandbox_pid=$!
        BOOTSTRAP_DIR="$SCRIPT_DIR/.bootstrap" bash "$SCRIPT_DIR/overlay/bootstrap.sh" &
        local bootstrap_pid=$!
        wait "$sandbox_pid" || true
        wait "$bootstrap_pid"
        save_digest  # record the base image digest for future update checks
        timer_show "create+bootstrap" "$t0"
    else
        echo "==> Bootstrapping dependencies on host..."
        t0=$(timer_start)
        BOOTSTRAP_DIR="$SCRIPT_DIR/.bootstrap" bash "$SCRIPT_DIR/overlay/bootstrap.sh"
        timer_show "bootstrap" "$t0"
    fi

    # Wait for sandbox to be Running before proceeding to SSH/mutagen
    echo "==> Waiting for sandbox to be Running..."
    t0=$(timer_start)
    local retries=0
    while true; do
        phase=$(get_phase)
        if [ "$phase" = "Running" ] || [ "$phase" = "Ready" ]; then
            break
        fi
        # Bail immediately on a terminal-failure phase instead of polling the
        # full 60s. A failed create (e.g. a rejected policy) lands in "Error".
        case "$phase" in
            Error|Failed|Terminated|Unknown)
                echo "ERROR: Sandbox entered '$phase' phase — startup failed."
                echo "  Details: openshell sandbox get $SANDBOX_NAME"
                echo "  Reset:   make sandbox-stop"
                exit 1
                ;;
        esac
        retries=$((retries + 1))
        if [ "$retries" -ge 30 ]; then
            echo "ERROR: Sandbox did not reach Running state within 60s (current: ${phase:-unknown})."
            exit 1
        fi
        sleep 2
    done
    timer_show "sandbox ready" "$t0"

    echo "==> Ensuring SSH config..."
    t0=$(timer_start)
    ensure_ssh_config
    timer_show "ssh config" "$t0"

    echo "==> Ensuring bootstrap sync..."
    t0=$(timer_start)
    ensure_bootstrap_sync
    timer_show "bootstrap sync" "$t0"

    t0=$(timer_start)
    ensure_setup
    timer_show "setup+post-install" "$t0"

    echo "==> Ensuring file sync..."
    t0=$(timer_start)
    ensure_mutagen_sync
    timer_show "file sync" "$t0"

    echo "==> Sandbox ready. (total: $(( $(date +%s) - t_total ))s)"
}

# Connect to sandbox. Builds first if needed.
cmd_connect() {
    local phase
    phase=$(get_phase)

    if [ "$phase" != "Running" ] && [ "$phase" != "Ready" ]; then
        cmd_build
    fi

    caffeinate -i env PATH="/usr/bin:$PATH" openshell sandbox connect "$SANDBOX_NAME" || true
    stty sane 2>/dev/null  # restore terminal after abnormal disconnect (e.g. sleep/wake)
}

cmd_clean() {
    soft_cleanup
}

cmd_stop() {
    echo "==> Stopping sandbox..."
    cleanup
    echo "Sandbox stopped."
}

cmd_nuke() {
    echo "==> Nuking sandbox (including bootstrap cache)..."
    cleanup_nuke
    echo "Sandbox nuked."
}

# SANDBOX_LIB_ONLY=1 sources this file for its functions without dispatching
# (used by construct/scripts/test/ to exercise pure helpers like
# compute_sync_set in isolation — no docker/mutagen/ssh required).
if [ "${SANDBOX_LIB_ONLY:-}" != 1 ]; then
    case "$ACTION" in
        build)   cmd_build ;;
        connect) cmd_connect ;;
        clean)   cmd_clean ;;
        stop)    cmd_stop ;;
        nuke)    cmd_nuke ;;
        *)       echo "Usage: $0 {build|connect|stop|clean|nuke} <sandbox-name>"; exit 1 ;;
    esac
fi
