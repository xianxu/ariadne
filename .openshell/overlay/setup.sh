#!/bin/bash
# One-shot setup for OpenShell sandbox — git config, shell config, workspace dirs.
# Dependency installation is handled by post-install.sh (from bootstrap cache).
# Idempotent: safe to re-run (e.g. via `make sandbox-clean`).
set -euo pipefail

mkdir -p "$HOME/.local/bin"

# ── Git config ───────────────────────────────────────────────────────────────
echo "==> Configuring git..."
git config --global url."https://github.com/".insteadOf "git@github.com:"
git config --global url."https://github.com/".insteadOf "ssh://git@github.com/"
# OpenShell proxy terminates TLS — sandbox doesn't have its CA cert
git config --global http.sslVerify false

# ── Shell config ─────────────────────────────────────────────────────────────
echo "==> Configuring shell..."
# Remove old block if present (idempotent)
sed -i '/^# BEGIN openshell-overlay/,/^# END openshell-overlay/d' "$HOME/.bashrc"
cat >> "$HOME/.bashrc" << 'BASHEOF'
# BEGIN openshell-overlay
export PATH="$HOME/.luarocks/bin:$HOME/.local/bin:$PATH"
export EDITOR="nvim"
export VISUAL="nvim"
unset LC_ALL

# Vi mode
set -o vi
bind '"\C-r": reverse-search-history'
bind '"\C-s": forward-search-history'

# Aliases
alias v=nvim
alias s="git status"
alias ss="git diff --stat"
alias a="git add"
alias d="git diff"
alias p="git commit -a; git push"
alias todo="nvim tasks/todo.md"
alias issue="nvim tasks/issue.md"
alias lesson="nvim tasks/lessons.md"
alias zl="zellij list-sessions"
alias ze="zellij"
alias za="zellij a"

# AI agent sandbox permissions — agents get full auto-approve
alias claude="claude --permission-mode bypassPermissions"
alias codex="codex --full-auto"
export GEMINI_CLI_AUTO_APPROVE=true

# ── Output capture (Ctrl+Y to copy last cmd+output) ─────────────────────────
# Strategy: use script(1) to synchronously capture output. No async tee races.
set +o noclobber
_bash_last_out=$(mktemp)
_bash_collect_out=$(mktemp)
_bash_collecting=false
_bash_last_cmd=""

_bash_strip_escapes() {
    perl -pe '
        s/\x1b\[[0-9;]*[A-Za-z]//g;
        s/\x1b\].*?(\x07|\x1b\\)//gs;
        s/\x1b[^\[\]]//g;
        s/\r//g;
    '
}

# Strip script(1) header/footer lines
_bash_strip_script() {
    sed '1{/^Script started on /d}; ${/^Script done on /d}'
}

# Clipboard via OSC 52 (works through SSH, zellij, tmux)
_bash_clip_copy() {
    local data
    data=$(base64 -w0 2>/dev/null || base64)
    printf '\033]52;c;%s\a' "$data" > /dev/tty
}

# Save last user command each prompt (skip our own commands)
_bash_precmd() {
    local _hist
    _hist=$(HISTTIMEFORMAT='' history 1)
    _hist="${_hist##*([[:space:]])+([0-9])*([[:space:]])}"
    case "$_hist" in
        clast*|clast_append*|ycollect|yy) ;;
        *) _bash_last_cmd="$_hist" ;;
    esac
}
PROMPT_COMMAND="_bash_precmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}"

clast() {
    # Re-run last command under script(1) to capture output synchronously
    local cmd="$_bash_last_cmd"
    [[ -z "$cmd" ]] && echo "[nothing to copy]" && return
    script -qc "bash -ic $(printf '%q' "$cmd")" "$_bash_last_out" > /dev/null 2>&1
    { printf '$ %s\n' "$cmd"; _bash_strip_script < "$_bash_last_out" | _bash_strip_escapes; } \
        | _bash_clip_copy
    echo "[copied]"
}

clast_append() {
    local cmd="$_bash_last_cmd"
    [[ -z "$cmd" ]] && echo "[nothing to copy]" && return
    script -qc "bash -ic $(printf '%q' "$cmd")" "$_bash_last_out" > /dev/null 2>&1
    local prev
    prev=$({ printf '$ %s\n' "$cmd"; _bash_strip_script < "$_bash_last_out" | _bash_strip_escapes; })
    if [[ -f /tmp/_bash_clip_buf ]]; then
        printf '%s\n%s' "$(cat /tmp/_bash_clip_buf)" "$prev" > /tmp/_bash_clip_buf
    else
        printf '%s' "$prev" > /tmp/_bash_clip_buf
    fi
    cat /tmp/_bash_clip_buf | _bash_clip_copy
    echo "[appended]"
}

ycollect() {
    _bash_collecting=true
    : > "$_bash_collect_out"
    echo "[collecting...]"
}

yy() {
    _bash_collecting=false
    cat "$_bash_collect_out" | _bash_strip_escapes | _bash_clip_copy
    echo "[copied]"
}

# Bind Ctrl+Y / Alt+Y: clear line, run clast silently, redraw prompt
bind -m vi-insert -x '"\C-y": clast'
bind -m vi-insert -x '"\ey": clast_append'
# END openshell-overlay
BASHEOF

# ── Workspace dirs ───────────────────────────────────────────────────────────
echo "==> Creating workspace dirs..."
mkdir -p "$HOME/repo" "$HOME/worktree"
mkdir -p "$HOME/.local/share/nvim/lazy"

# ── Credentials (from bootstrap cache) ──────────────────────────────────────
CREDS="/tmp/bootstrap/credentials"
if [ -d "$CREDS" ]; then
    echo "==> Distributing credentials..."
    if [ -f "$CREDS/codex-auth.json" ]; then
        mkdir -p "$HOME/.codex"
        cp "$CREDS/codex-auth.json" "$HOME/.codex/auth.json"
        echo "  [ok] codex auth"
    fi
fi

echo "==> Done."
