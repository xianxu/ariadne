#!/usr/bin/env bash
# scripts/sdlc-install.sh — put the sdlc binary on the developer's PATH.
#
# Idempotent: verifies Go toolchain, builds sdlc via `make sdlc-build`,
# appends the OWNER's bin/ to the user's shell rc if not already there, and
# prints the export line so the user can paste it manually as backup
# (in case rc-detection picked the wrong file).
#
# Build-in-owner (#60, #95 M5): sdlc-build builds into its OWNER's bin/, not
# this repo's bin/ — so there is exactly one sdlc on disk (../ariadne/bin/sdlc).
# This script therefore resolves the owner and puts the OWNER's bin/ on PATH.
# In the owner's own install the owner IS this repo, so PATH gets $REPO_DIR/bin
# exactly as before; in a consumer it gets ../ariadne/bin (where sdlc lives) and
# the consumer's empty bin/ never needs to be on PATH for sdlc to resolve.
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
# sdlc-build is build-in-owner (#60, #95 M5): it builds sdlc into its OWNER's
# bin/ (resolved by location via construct/dev-aliases.sh --list), NOT this
# repo's bin/. So the binary lives at exactly one place — $OWNER/bin/sdlc — and
# a consumer never gets a duplicate $REPO_DIR/bin/sdlc. When this repo IS the
# owner (ariadne), $OWNER == $REPO_DIR, so it lands in $REPO_DIR/bin as before.
info "building sdlc (build-in-owner)"
make --no-print-directory sdlc-build

# Resolve where sdlc actually landed — the owner's bin/, which is the dir we put
# on PATH. In the owner's own install this is $REPO_DIR/bin (unchanged); in a
# consumer it's ../ariadne/bin (where the one true sdlc lives).
OWNER="$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$1=="sdlc"{print $2}')"
if [ -z "$OWNER" ]; then
    die "sdlc owner not found beside this repo; run 'make bootstrap-peers' + 'make weave' first"
fi
SDLC_BIN_DIR="$OWNER/bin"

if [ ! -x "$SDLC_BIN_DIR/sdlc" ]; then
    die "expected $SDLC_BIN_DIR/sdlc to exist after sdlc-build; aborting"
fi

# ── 3. PATH wiring ──────────────────────────────────────────────────────────
# Put the OWNER's bin/ on PATH (where the single sdlc binary lives), not this
# repo's empty bin/. In the owner's case these coincide; in a consumer this
# points PATH at ../ariadne/bin so `sdlc` resolves to the one true binary —
# the consumer's own bin/ stays empty and never needs to be on PATH for sdlc.
# Mirrors nous-bootstrap.sh's PATH-write convention so multiple `*-install`
# targets compose idempotently.
SHELL_RC=""
case "${SHELL:-}" in
    */zsh)  SHELL_RC="$HOME/.zshrc" ;;
    */bash) [ -f "$HOME/.bash_profile" ] && SHELL_RC="$HOME/.bash_profile" || SHELL_RC="$HOME/.bashrc" ;;
esac

EXPORT_LINE="export PATH=\"$SDLC_BIN_DIR:\$PATH\""

if [ -n "$SHELL_RC" ] && ! grep -q "$SDLC_BIN_DIR" "$SHELL_RC" 2>/dev/null; then
    info "Adding $SDLC_BIN_DIR to PATH in $SHELL_RC..."
    printf '\n# Added by sdlc-install: owner bin/ for sdlc + other cmd/* binaries\n%s\n' "$EXPORT_LINE" >> "$SHELL_RC"
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
