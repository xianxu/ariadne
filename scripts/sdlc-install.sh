#!/usr/bin/env bash
# scripts/sdlc-install.sh — put the in-tree sdlc binary on the developer's PATH.
#
# Idempotent: verifies Go toolchain, builds bin/sdlc via `make sdlc-build`,
# appends $REPO_DIR/bin to the user's shell rc if not already there, and
# prints the export line so the user can paste it manually as backup
# (in case rc-detection picked the wrong file).
#
# Wired into `make bootstrap` as a final step (via Makefile.workflow).
# Standalone invocation: `make sdlc-install`.
#
# Renamed from sdlc-bootstrap.sh in #41: the script no longer symlinks
# into ~/bin — it puts the in-tree bin/ on PATH instead, mirroring the
# nous-bootstrap convention. The `make sdlc-bootstrap` target stays as
# a backward-compat alias for pre-rename muscle memory.

set -euo pipefail

RED=$'\033[0;31m'; GREEN=$'\033[0;32m'; YELLOW=$'\033[1;33m'; CYAN=$'\033[0;36m'; RESET=$'\033[0m'
info() { printf "%s==>%s %s\n" "$CYAN" "$RESET" "$*" >&2; }
ok()   { printf "%s  [ok]%s %s\n" "$GREEN" "$RESET" "$*" >&2; }
warn() { printf "%s  [!]%s %s\n" "$YELLOW" "$RESET" "$*" >&2; }
die()  { printf "%serror:%s %s\n" "$RED" "$RESET" "$*" >&2; exit 1; }

REPO_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$REPO_DIR"

# ── 1. Toolchain ────────────────────────────────────────────────────────────
# go is provisioned by sdlc-build's `ensure-go` prereq (#61) — don't die here;
# just report the version when it's already present. `make sdlc-build` below
# auto-installs (brew) or fails fast with guidance if it's missing.
if command -v go >/dev/null 2>&1; then
    ok "Go found: $(go version | awk '{print $3}')"
fi

# ── 2. Build ────────────────────────────────────────────────────────────────
# sdlc-build produces $REPO_DIR/bin/sdlc in both ariadne (source) and
# downstream repos (via the construct/go.mod replace => ../ariadne path).
info "building bin/sdlc"
make --no-print-directory sdlc-build

if [ ! -x "$REPO_DIR/bin/sdlc" ]; then
    die "expected $REPO_DIR/bin/sdlc to exist after sdlc-build; aborting"
fi

# ── 3. PATH wiring ──────────────────────────────────────────────────────────
# Append $REPO_DIR/bin to the user's shell rc if not already there.
# Mirrors nous-bootstrap.sh's PATH-write convention so multiple
# `*-install` targets compose idempotently.
SHELL_RC=""
case "${SHELL:-}" in
    */zsh)  SHELL_RC="$HOME/.zshrc" ;;
    */bash) [ -f "$HOME/.bash_profile" ] && SHELL_RC="$HOME/.bash_profile" || SHELL_RC="$HOME/.bashrc" ;;
esac

EXPORT_LINE="export PATH=\"$REPO_DIR/bin:\$PATH\""

if [ -n "$SHELL_RC" ] && ! grep -q "$REPO_DIR/bin" "$SHELL_RC" 2>/dev/null; then
    info "Adding $REPO_DIR/bin to PATH in $SHELL_RC..."
    printf '\n# Added by sdlc-install: in-tree bin/ for sdlc + other cmd/* binaries\n%s\n' "$EXPORT_LINE" >> "$SHELL_RC"
    ok "PATH updated. Open a new shell (or run: source $SHELL_RC) to pick it up."
elif [ -z "$SHELL_RC" ]; then
    warn "couldn't auto-detect shell rc from SHELL=$SHELL — see manual step below"
fi

# ── 4. Manual-paste reminder (belt-and-suspenders) ──────────────────────────
# Print the line even when auto-write succeeded so the user can verify,
# or paste into a different rc file if their setup differs from what
# shell detection assumes.
echo
ok "If sdlc isn't on PATH in a new shell, add this line to your ~/.zshrc or ~/.bashrc:"
echo "    $EXPORT_LINE"
echo

info "done. open a new shell, then: sdlc --help"
