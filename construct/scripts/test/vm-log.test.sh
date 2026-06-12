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
