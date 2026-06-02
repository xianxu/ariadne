#!/usr/bin/env bash
# ensure-go.test.sh — hermetic tests for Makefile.workflow's `ensure-go` (#61).
# Runs the real `make ensure-go` recipe under a controlled PATH with stub
# go/brew, asserting the three branches: no-op (go present), brew-install
# (absent + brew), fail-fast (absent + no brew).
#
# Run:  bash construct/scripts/test/ensure-go.test.sh
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
MAKE="$(command -v make)"
ROOT="$(mktemp -d "${TMPDIR:-/tmp}/ensurego.XXXXXX")"
trap 'rm -rf "$ROOT"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }

# Stub dirs: a fake `go` and a fake `brew` (records its args, never installs).
mkdir -p "$ROOT/withgo" "$ROOT/withbrew" "$ROOT/bare"
printf '#!/bin/sh\nexit 0\n' > "$ROOT/withgo/go";    chmod +x "$ROOT/withgo/go"
printf '#!/bin/sh\necho "brew $*" >> "%s/brew.log"\n' "$ROOT" > "$ROOT/withbrew/brew"
chmod +x "$ROOT/withbrew/brew"

# Base PATH gives /bin/sh + coreutils but NOT go/brew (so the stubs control presence).
BASE="/usr/bin:/bin"

run() { PATH="$1" "$MAKE" -C "$ARIADNE_ROOT" ensure-go 2>&1; }

echo "== ensure-go (#61) =="

# ── 1: go present → no-op, exit 0, no install attempt ─────────────────────────
out="$(run "$ROOT/withgo:$BASE")"; rc=$?
{ [ $rc -eq 0 ] && ! grep -qi "installing\|Error" <<<"$out"; } \
    && ok "go present → no-op (exit 0, no install)" || ko "go-present: rc=$rc out=[$out]"

# ── 2: go absent + brew present → takes brew-install branch ───────────────────
: > "$ROOT/brew.log"
out="$(run "$ROOT/withbrew:$BASE")"; rc=$?
{ [ $rc -eq 0 ] && grep -qi "installing via Homebrew" <<<"$out" && grep -q "install go" "$ROOT/brew.log"; } \
    && ok "go absent + brew → 'brew install go' invoked" || ko "brew-branch: rc=$rc out=[$out] log=[$(cat "$ROOT/brew.log")]"

# ── 3: go absent + no brew → fail fast with guidance ──────────────────────────
out="$(run "$BASE")"; rc=$?
{ [ $rc -ne 0 ] && grep -qi "go.dev/dl" <<<"$out"; } \
    && ok "go absent + no brew → fail-fast with go.dev/dl guidance" || ko "fail-fast: rc=$rc out=[$out]"

echo
echo "== $pass passed, $fail failed =="
[ "$fail" -eq 0 ]
