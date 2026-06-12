#!/usr/bin/env bash
# vm-log.sh — shared logging for the VM test targets (.tart + .colima): bold-cyan
# step headers + a dim/gray pass-through filter for underlying-process output.
# One source of the ANSI codes (ARCH-DRY); referenced directly by
# .colima/colima.sh and .tart/Makefile.
#
#   vm-log.sh step <msg>   bold-cyan   "==> <msg>"
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
