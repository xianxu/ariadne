# Colima VM Post-Login Parity + Unified VM Logging — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the Colima Linux VM the same post-login customization Tart VMs get (aliases, nvim, dev-aliases, auto-cd, per-repo hooks) by mirroring the OpenShell *Linux* overlay; and give both `make tart` and `make colima` colorized `==>` step headers + dimmed underlying-process output via one shared `vm-log.sh`.

**Architecture:** New guest scripts `.colima/vm-setup.sh` (boot bootstrap) + `.colima/vm-rc.sh` (bash rc), run by `colima.sh` after start (mirrors how `.tart` pushes `tart-vm-setup.sh`). A shared `construct/scripts/vm-log.sh` is the single source of the ANSI codes (`step`/`warn`/`dim`), referenced directly by both `.colima/colima.sh` and `.tart/Makefile`.

**Tech Stack:** bash (guest is Ubuntu 24.04 bash 5 — no bash-3.2 limits *inside* the guest; host-side `colima.sh`/`vm-log.sh` stay bash-3.2-safe), GNU Make, Colima 0.10.x, apt (neovim), oh-my-bash.

---

## Scope decisions (brainstorm 2026-06-12)

- **Lean parity**, **skip agent aliases** (AskUserQuestion).
- Mirror the **OpenShell overlay** (`.openshell/overlay/setup.sh`, Linux/bash), NOT Tart's `tart-vm-rc.zsh` (macOS/zsh) — the Colima guest is Ubuntu/bash. **ARCH-DRY:** reuse the portable Linux source.
- **Deliberately not ported** (sandbox/macOS-specific, would be wrong or useless on Colima): the egress-proxy config (`https_proxy`, `git http.sslVerify false`, pip/npm proxy), `/tmp/bootstrap` credentials, macOS DNS-flush + Homebrew PATH, the `script(1)`+DEBUG-trap output-capture machinery, AI-agent auto-approve aliases. (Full table in the issue + the user-facing analysis.)

## Core concepts

This is an IO/glue feature: guest provisioning + terminal formatting. The pure surface is thin (the ANSI-wrapping decision in `vm-log.sh`). Verification is a real Colima boot (exercises `vm-setup`/`vm-rc`/logging end-to-end) plus deterministic unit tests for `vm-log.sh` and the `colima.sh` command assembly (process-level fake from #93).

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `vm-log.sh` | `construct/scripts/vm-log.sh` | new | tty / ANSI formatting |
| `vm-setup.sh` | `.colima/vm-setup.sh` | new | guest apt / oh-my-bash / symlinks |
| `vm-rc.sh` | `.colima/vm-rc.sh` | new | guest bash rc |
| `colima.sh` | `.colima/colima.sh` | modified | + `vm-log` step/dim, + run setup |
| `.tart/Makefile` (`_tart_boot_and_ssh`, recipes) | `.tart/Makefile` | modified | + `vm-log` step, + dimmed boot-log tail |
| base manifest | `construct/base.manifest` | modified | symlink propagation |
| `colima.test.sh` | `.colima/test/colima.test.sh` | modified | + setup-call assertion |

- **`vm-log.sh`** — one source of the ANSI codes for both VM fragments. `step <msg>` → bold-cyan `==> msg`; `warn <msg>` → yellow `[!] msg`; `dim` → filter stdin to dimmed (gray) lines. Color emitted only when `[ -t 1 ]` (or `CLICOLOR_FORCE`) and `NO_COLOR` unset — so logfiles/pipes stay clean and CI is unaffected.
  - **Injected into:** `colima.sh` (calls `bash vm-log.sh …`) and `.tart/Makefile` (recipe lines `$(VM_LOG) step …`).
  - **DRY rationale:** Without it, the cyan/dim codes would be duplicated across the Makefile recipes and the shell script. **ARCH-DRY.**
  - **Future extensions:** `substep`/`ok` verbs; spinner.

- **`vm-setup.sh`** — guest bootstrap, run on every `make colima` via `colima ssh -- bash -s -- <repo-dir> <mount-dir>`, idempotent. Linux counterpart to `tart-vm-setup.sh`: `~/workspace`/`~/repo` symlinks to the host-abs mount, `~/.colima-current-repo` marker, `apt install neovim`, oh-my-bash, append the `vm-rc.sh` source-line to `~/.bashrc`, run `.colima/vm-hooks.d/*.sh`.
  - **Injected into:** invoked by `colima.sh`'s `_run_setup`.
  - **Future extensions:** `COLIMA_VANILLA` skip; zellij/lua install.

- **`vm-rc.sh`** — bash rc pushed to `~/.colima-vm-rc.sh`, sourced from `~/.bashrc`. Linux/bash port of `tart-vm-rc.zsh`'s portable bits: git aliases, `v`, vi-mode + Ctrl+R/S, PATH, `EDITOR=nvim`, GPG_TTY guard, `dev-aliases.sh` wrappers, auto-cd.
  - **Injected into:** sourced by the guest login shell.

### Pure-ish core

- **`vm-log.sh` color decision** (`_color()`): `NO_COLOR` → off; `CLICOLOR_FORCE` → on; else `[ -t 1 ]`. The one testable branch (Task 1 test).

## Implementation risks (verify, don't assume)

- **R1 — oh-my-bash installer rewrites `~/.bashrc`.** Must run *before* appending the `vm-rc.sh` source-line (else the append is clobbered) — same ordering Tart uses. **Verify:** after boot, `grep colima-vm-rc.sh ~/.bashrc` present and a login shell has the aliases.
- **R2 — `ln -sfn` nesting trap.** If `~/workspace` is ever a real dir, `ln -sfn` links *inside* it. Guard: rm a non-symlink first (mirror the OpenShell-overlay note). Fresh Lima guest has neither, so low risk. **Verify:** `readlink ~/workspace` == mount dir.
- **R3 — Tart dimmed boot-log tail.** Streaming `/tmp/tart-<vm>.log` dimmed during the SSH-wait, then killing the tailer, is the riskiest edit to working Tart code. Tail must start only when we actually booted, and a `printf '\033[0m'` must reset color after kill. **Verify:** `make -n tart` expands cleanly; standalone `printf 'a\nb\n' | vm-log.sh dim` dims; (full visual is an operator eyeball on a real `make tart` — booting a macOS guest is heavy/capped, so this task verifies mechanics, not a full boot).

---

## Chunk 1: shared logging, guest setup, wiring, verification

### Task 1: `construct/scripts/vm-log.sh` + test

**Files:**
- Create: `construct/scripts/vm-log.sh`
- Create: `construct/scripts/test/vm-log.test.sh`

- [ ] **Step 1: Write `vm-log.sh`**

```bash
#!/usr/bin/env bash
# vm-log.sh — shared logging for the VM test targets (.tart + .colima): bold-cyan
# step headers + a dim/gray pass-through filter for underlying-process output.
# One source of the ANSI codes (ARCH-DRY); referenced directly by
# .colima/colima.sh and .tart/Makefile.
#
#   vm-log.sh step <msg>   bold-cyan "==> <msg>"
#   vm-log.sh warn <msg>   bold-yellow "  [!] <msg>"
#   vm-log.sh dim          filter stdin → dimmed (gray) lines
#
# Color only to a tty (or CLICOLOR_FORCE) and when NO_COLOR is unset — so
# logfiles/pipes stay clean and CI is unaffected. bash-3.2 safe.
set -euo pipefail

_color() {
    [ -n "${NO_COLOR:-}" ] && return 1
    [ -n "${CLICOLOR_FORCE:-}" ] && return 0
    [ -t 1 ]
}

cmd=${1:-}; shift || true
case "$cmd" in
  step)
    if _color; then printf '\033[1;36m==>\033[0m %s\n' "$*"; else printf '==> %s\n' "$*"; fi ;;
  warn)
    if _color; then printf '\033[1;33m  [!]\033[0m %s\n' "$*"; else printf '  [!] %s\n' "$*"; fi ;;
  dim)
    if _color; then
        # `|| [ -n "$line" ]` flushes a final line that lacks a trailing newline.
        while IFS= read -r line || [ -n "$line" ]; do
            printf '\033[2m%s\033[0m\n' "$line"
        done
    else
        cat
    fi ;;
  *)
    echo "usage: vm-log.sh step|warn <msg> | dim" >&2; exit 2 ;;
esac
```

- [ ] **Step 2: Write the test**

```bash
#!/usr/bin/env bash
# vm-log.test.sh — deterministic checks of color gating + dim filtering.
set -euo pipefail
LOG="$(cd "$(dirname "$0")/.." && pwd)/vm-log.sh"

# NO_COLOR → plain (no ESC), even with CLICOLOR_FORCE also set (NO_COLOR wins)
out=$(NO_COLOR=1 CLICOLOR_FORCE=1 bash "$LOG" step hello)
[ "$out" = "==> hello" ] || { echo "FAIL: NO_COLOR step: '$out'"; exit 1; }

# CLICOLOR_FORCE → colored (contains ESC[1;36m)
out=$(CLICOLOR_FORCE=1 bash "$LOG" step hello | cat -v)
case "$out" in *'^[[1;36m==>^[[0m hello'*) ;; *) echo "FAIL: forced color step: '$out'"; exit 1 ;; esac

# dim, plain mode (no tty, no force) → passthrough unchanged
out=$(printf 'a\nb\n' | bash "$LOG" dim)
[ "$out" = "$(printf 'a\nb')" ] || { echo "FAIL: dim passthrough: '$out'"; exit 1; }

# dim, forced color → each line wrapped in ESC[2m…ESC[0m
out=$(printf 'a\n' | CLICOLOR_FORCE=1 bash "$LOG" dim | cat -v)
[ "$out" = '^[[2ma^[[0m' ] || { echo "FAIL: dim color: '$out'"; exit 1; }

# dim flushes a final newline-less line
out=$(printf 'x' | bash "$LOG" dim)
[ "$out" = "x" ] || { echo "FAIL: dim final line: '$out'"; exit 1; }

echo "PASS: vm-log.sh"
```

- [ ] **Step 3: chmod + run**

Run: `chmod +x construct/scripts/vm-log.sh construct/scripts/test/vm-log.test.sh && bash construct/scripts/test/vm-log.test.sh`
Expected: `PASS: vm-log.sh`

- [ ] **Step 4: Commit**

```bash
git add construct/scripts/vm-log.sh construct/scripts/test/vm-log.test.sh
git commit -m "#94: vm-log.sh — shared colorized step / dimmed-passthrough VM logging"
```

### Task 2: `.colima/vm-rc.sh` (guest bash rc)

**Files:**
- Create: `.colima/vm-rc.sh`

- [ ] **Step 1: Write it**

```bash
# vm-rc.sh — bash rc for the Colima Linux guest. Pushed to ~/.colima-vm-rc.sh;
# sourced from ~/.bashrc by vm-setup.sh. The Linux/bash counterpart to
# .tart/scripts/tart-vm-rc.zsh (zsh, for the macOS guest); mirrors the portable
# pieces of .openshell/overlay/setup.sh. Edit on the host, re-push next make colima.

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
```

- [ ] **Step 2: Syntax check** — `bash -n .colima/vm-rc.sh && echo OK` → `OK`
- [ ] **Step 3: Commit** — `git add .colima/vm-rc.sh && git commit -m "#94: vm-rc.sh — Colima guest bash rc (aliases, vi-mode, dev-aliases, auto-cd)"`

### Task 3: `.colima/vm-setup.sh` (guest bootstrap)

**Files:**
- Create: `.colima/vm-setup.sh`

- [ ] **Step 1: Write it**

```bash
#!/usr/bin/env bash
# vm-setup.sh — guest bootstrap for `make colima`, run on every boot via
# `colima ssh -- bash -s -- <repo-dir> <mount-dir>`. Idempotent. The Linux
# counterpart to .tart/scripts/tart-vm-setup.sh; mirrors the portable pieces of
# the OpenShell overlay, omitting the sandbox proxy/creds + macOS bits.
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

# ── neovim (apt; idempotent) ─────────────────────────────────────────────────
if ! command -v nvim >/dev/null 2>&1; then
    echo "installing neovim..."
    sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq
    sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq neovim >/dev/null
fi

# ── oh-my-bash (network; non-fatal; BEFORE the rc-append, see R1) ─────────────
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
```

- [ ] **Step 2: Syntax check** — `bash -n .colima/vm-setup.sh && echo OK` → `OK`
- [ ] **Step 3: Commit** — `git add .colima/vm-setup.sh && git commit -m "#94: vm-setup.sh — Colima guest bootstrap (symlinks, nvim, oh-my-bash, hooks)"`

### Task 4: wire `colima.sh` (vm-log step/dim + run setup)

**Files:**
- Modify: `.colima/colima.sh`

- [ ] **Step 1: Add the vm-log wrappers** (after `SCRIPT_DIR=…`):

```bash
VMLOG="$SCRIPT_DIR/../construct/scripts/vm-log.sh"
step() { bash "$VMLOG" step "$@"; }
dim()  { bash "$VMLOG" dim; }
```

- [ ] **Step 2: Replace the plain `echo "==> …"` step lines with `step "…"`** in `_start_if_needed`, `_provision_vnc`, `stop`, `clean` (the `==>` lines only — keep substep `echo "  …"` plain or convert to `vm-log warn`/indented as preferred). Leave non-step echoes (the VNC connect instructions) as plain `echo`.

- [ ] **Step 3: Dim the boot + setup sub-output.** In `_start_if_needed`, change the final `colima start "$@"` to `colima start "$@" 2>&1 | dim`. Add `_run_setup` — it **always re-pushes the cheap rc** (so host edits to `vm-rc.sh` propagate) but only runs the **heavier `vm-setup.sh`** (symlinks, apt nvim, oh-my-bash, hooks) on a *fresh* start (`$4=fresh`), since the profile is per-repo so the marker/symlinks never change across re-entries:

```bash
_run_setup() {
    local profile=$1 repodir=$2 mountdir=$3 fresh=${4:-1}
    # Always refresh the rc (cheap; lets host edits to vm-rc.sh take effect).
    colima ssh -p "$profile" -- bash -c 'cat > ~/.colima-vm-rc.sh' < "$SCRIPT_DIR/vm-rc.sh"
    [ "$fresh" = 1 ] || return 0   # skip the heavy pass on re-entry to a running VM
    step "Provisioning guest (symlinks, nvim, oh-my-bash, dev-aliases; first run is slower)..."
    colima ssh -p "$profile" -- bash -s -- "$repodir" "$mountdir" \
        < "$SCRIPT_DIR/vm-setup.sh" 2>&1 | dim
}
```

- [ ] **Step 4: Call `_run_setup` in `up`/`gui`** — detect fresh-vs-running BEFORE starting, and make setup **non-fatal to the session** (`|| step …`): a guest-setup hiccup must never strand you out of a VM that's actually up (mirrors Tart's "wiring can't strand" principle). The final `_ssh_into` is NOT piped through `dim` — it's the interactive session.

```bash
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
```

(`_run_setup … || step …` also disables `set -e` inside `_run_setup`, so an internal failure returns cleanly instead of aborting colima.sh before the shell.)

- [ ] **Step 5: Run the fake test** — `bash .colima/test/colima.test.sh` → `PASS` (existing assertions still hold; `step`/`dim` go through `vm-log.sh`, the fake `colima start` still records argv before the dim pipe).
- [ ] **Step 6: Commit** — `git add .colima/colima.sh && git commit -m "#94: colima.sh — colorized steps, dimmed sub-output, run guest setup on boot"`

### Task 5: wire `.tart/Makefile` (vm-log step + dimmed boot-log tail)

**Files:**
- Modify: `.tart/Makefile`

- [ ] **Step 1: Define `VM_LOG`** near the top (after `TART_VM` derivation):

```makefile
# Shared VM logging (construct/scripts/vm-log.sh): colorized steps + dimmed
# underlying-process output. Same helper .colima uses (ARCH-DRY).
VM_LOG := bash $(CURDIR)/construct/scripts/vm-log.sh
```

- [ ] **Step 2: Replace `@echo "==> …"` with `@$(VM_LOG) step "…"`** across the recipes (`tart`, `tart-gui`, the `define` blocks, `_tart_boot_and_ssh`). Keep the non-`==>` echoes (multi-line instructions) plain. Convert the `[warn]`/`[!]` lines to `$(VM_LOG) warn "…"` where natural.

- [ ] **Step 3: Combine boot + SSH-wait into one recipe line in `_tart_boot_and_ssh` with a dimmed log-tail** (replaces the current first two `@` lines). Exact replacement:

```makefile
	@state=$$(tart list 2>/dev/null | awk -v vm=$(TART_VM) '$$2==vm {print $$NF}'); \
	booted=0; \
	if [ "$$state" != "running" ]; then \
	    $(VM_LOG) step "Booting VM..."; \
	    nohup tart run $(GRAPHICS_FLAG) $(RUN_FLAGS) $(TART_VM) >/tmp/tart-$(TART_VM).log 2>&1 & \
	    booted=1; \
	fi; \
	$(VM_LOG) step "Waiting for SSH..."; \
	tail_pid=; \
	if [ "$$booted" = 1 ] && [ -t 1 ]; then \
	    ( sleep 0.3; tail -n +1 -f /tmp/tart-$(TART_VM).log 2>/dev/null | $(VM_LOG) dim ) & tail_pid=$$!; \
	fi; \
	for i in $$(seq 1 60); do \
	    ip=$$(tart ip $(TART_VM) 2>/dev/null); \
	    if [ -n "$$ip" ] && nc -z -G 2 "$$ip" 22 2>/dev/null; then break; fi; \
	    sleep 2; \
	done; \
	if [ -n "$$tail_pid" ]; then kill "$$tail_pid" 2>/dev/null || true; printf '\033[0m'; fi
```

(The rest of `_tart_boot_and_ssh` — the `@ip=…` pubkey/setup/ssh line — stays unchanged except `echo "==> …"` → `$(VM_LOG) step "…"`.)

- [ ] **Step 4: Verify the recipe expands cleanly** — `make -n tart 2>&1 | head -40` parses without error and shows the `vm-log.sh step`/`tail … | … dim` structure. (A full boot needs a real macOS guest — heavy/2-VM-capped — so mechanics are verified here + via Task 1's standalone dim test; the full visual is an operator eyeball, noted in the Log.)
- [ ] **Step 5: Commit** — `git add .tart/Makefile && git commit -m "#94: tart — colorized steps + dimmed boot-log tail during SSH wait"`

### Task 6: base manifest + colima test assertion

**Files:**
- Modify: `construct/base.manifest`
- Modify: `.colima/test/colima.test.sh`

- [ ] **Step 1: Add manifest entries.** Near the `list-peers.sh` line add `symlink construct/scripts/vm-log.sh` (shared walker neighborhood); in the `# ── Colima` block add `symlink .colima/vm-setup.sh` and `symlink .colima/vm-rc.sh`.
- [ ] **Step 2: Add a setup-call assertion** to `colima.test.sh` after the `up` block:

```bash
# up also pushes vm-rc.sh + runs vm-setup.sh (fresh start) before the interactive ssh
FAKE_RUNNING=0 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has "colima-vm-rc.sh"                              # vm-rc push
assert_has "bash -s -- /Users/me/ws/repo /Users/me/ws"   # vm-setup invocation

# up on an ALREADY-RUNNING profile skips the heavy vm-setup (still pushes rc)
FAKE_RUNNING=1 run up p-test /Users/me/ws/repo /Users/me/ws >/dev/null
assert_has  "colima-vm-rc.sh"
assert_none "bash -s -- /Users/me/ws/repo /Users/me/ws"
```

- [ ] **Step 3: Run** — `bash .colima/test/colima.test.sh` → `PASS`
- [ ] **Step 4: Commit** — `git add construct/base.manifest .colima/test/colima.test.sh && git commit -m "#94: base.manifest entries + colima test setup-call assertion"`

### Task 7: real verification (Colima boot) + Log

- [ ] **Step 1: Boot + setup (non-interactive drive).** `colima start` the probe profile, push `vm-rc.sh`, run `vm-setup.sh` via `colima ssh -- bash -s -- <repo> <mount>`. Then verify in the guest:
  - `command -v nvim` → present
  - `readlink ~/workspace` == mount dir; `readlink ~/repo` == repo dir; `cat ~/.colima-current-repo` == repo name
  - `grep colima-vm-rc.sh ~/.bashrc` present (R1)
  - a login shell has the aliases + lands in the repo: `colima ssh -- bash -lic 'alias s; type repo; pwd'`
  - dev-aliases: with ariadne in the workspace, `colima ssh -- bash -lic 'type sdlc' ` shows the function (or a clean no-op if absent)
- [ ] **Step 2: Logging.** Confirm `vm-log.sh` test PASS; confirm `colima start … | vm-log.sh dim` dims (force color) and the interactive shell is not dimmed; `make -n tart` clean.
- [ ] **Step 3: Teardown** — `colima delete -p <probe> -f`; docker context back to default.
- [ ] **Step 4: Record evidence** in the issue `## Log` (commands + key output; note the tart full-boot eyeball residual).

---

## Close-out

Single review boundary (cohesive VM-polish, verified in one boot) → one `sdlc close` (no `Mx`). Mandatory fresh-context review auto-dispatches.

```
sdlc close --issue 94 --actual <measured> --verified '<colima boot: nvim+aliases+symlinks+auto-cd+dev-aliases; vm-log test PASS + dim/step verified; make -n tart clean>'
```
Then publish #93+#94 together via `sdlc pr` → `sdlc merge`.
