# vm-rc.sh — bash rc for the Colima Linux guest. Pushed to ~/.colima-vm-rc.sh;
# sourced from ~/.bashrc by vm-setup.sh. The Linux/bash counterpart to
# .tart/scripts/tart-vm-rc.zsh (zsh, for the macOS guest); mirrors the portable
# pieces of .openshell/overlay/setup.sh. Edit on the host, re-push next make colima.
#
# DRY note (deliberate copy-not-extract): this portable subset also lives in
# .openshell/overlay/setup.sh (bash) and tart-vm-rc.zsh (zsh). Extracting a
# shared sourced fragment would force editing the working OpenShell sandbox path
# (real scope creep), and tart's is zsh anyway — so per-target rc is the
# accepted call here. A *third* bash consumer should tip the balance toward
# extracting a shared construct/ portable-rc fragment. (The high-churn surface —
# the ANSI codes — IS shared, in construct/scripts/vm-log.sh.)

export PATH="$HOME/repo/bin:$HOME/.local/bin:$PATH"
export EDITOR="${EDITOR:-nvim}"
export VISUAL="${VISUAL:-nvim}"

# Vi mode + history search (interactive only; guard so a non-interactive source
# doesn't error on `bind`).
if [[ $- == *i* ]]; then
    set -o vi
    bind '"\C-r": reverse-search-history' 2>/dev/null
    bind '"\C-s": forward-search-history' 2>/dev/null
fi

# Git workflow aliases — parity with .openshell/overlay/setup.sh.
alias s='git status'
alias ss='git diff --stat'
alias a='git add'
alias d='git diff'
alias p='git commit -a && git push'
alias v='${EDITOR}'

{ [ -d "$HOME/repo" ] || [ -L "$HOME/repo" ]; } && alias repo='cd $HOME/repo'
{ [ -d "$HOME/workspace" ] || [ -L "$HOME/workspace" ]; } && alias workspace='cd $HOME/workspace'

# GPG pinentry self-heal (guarded no-op if gpg absent).
export GPG_TTY=$(tty 2>/dev/null || true)
command -v gpg-connect-agent >/dev/null 2>&1 && \
    gpg-connect-agent updatestartuptty /bye >/dev/null 2>&1

# Dev-binary build-on-demand wrappers (ariadne#57) — guarded no-op unless ariadne
# is in the workspace clone. `eval "$(bash …)"` not `source <(bash …)`: the
# latter (process substitution) is unreliable for defining functions on older
# bash; eval is portable and equivalent. (tart-vm-rc.zsh uses the <() form, but
# that's zsh.)
_dev_aliases="$HOME/workspace/ariadne/construct/dev-aliases.sh"
[ -r "$_dev_aliases" ] && eval "$(bash "$_dev_aliases")"
unset _dev_aliases

# Auto-cd into the current repo on interactive login. Marker written by
# vm-setup.sh; falls back to $HOME if anything's missing.
if [[ $- == *i* ]] && [ -f "$HOME/.colima-current-repo" ]; then
    _repo="$(cat "$HOME/.colima-current-repo" 2>/dev/null)"
    if [ -n "$_repo" ] && [ -d "$HOME/workspace/$_repo" ]; then cd "$HOME/workspace/$_repo"; fi
    unset _repo
fi
