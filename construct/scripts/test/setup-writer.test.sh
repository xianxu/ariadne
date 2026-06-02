#!/usr/bin/env bash
# setup-writer.test.sh — hermetic tests for setup.sh's `tool` action writer
# (ensure_go_tool_dependency) after #60 M3. A cross-target consumer now gets a
# `substrate <rel>` row in construct/deps (repo-root-relative), NOT a
# construct/go.mod stub; an existing construct/go.mod is left untouched
# (dual-read). Sources setup.sh with SETUP_LIB_ONLY=1 and calls the function
# directly — no Go toolchain needed for the cross-target path.
#
# Run:  bash construct/scripts/test/setup-writer.test.sh
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
SETUP="$ARIADNE_ROOT/construct/setup.sh"

ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/setupwriter.XXXXXX")" && pwd -P)"
trap 'rm -rf "$ROOT"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }
wf() { mkdir -p "$(dirname "$1")"; cat > "$1"; }

mkdir -p "$ROOT/up" "$ROOT/up2" "$ROOT/leaf/construct" "$ROOT/leaf2/construct"

# Source setup.sh for its functions only (SETUP_LIB_ONLY skips discovery/apply).
SETUP_LIB_ONLY=1 source "$SETUP" >/dev/null 2>&1
set +e   # setup.sh enables `set -e`; we want to run all assertions

echo "== setup.sh tool-action writer (#60 M3) =="

# ── Test 1: cross-target writes a substrate row to construct/deps ─────────────
TARGET_DIR="$ROOT/leaf"
ensure_go_tool_dependency "$ROOT/up" "cmd/sdlc" >/dev/null 2>&1
got="$(cat "$ROOT/leaf/construct/deps" 2>/dev/null)"
[ "$got" = "substrate ../up" ] && ok "writes 'substrate ../up' to construct/deps" || ko "deps content: [$got]"
[ ! -f "$ROOT/leaf/construct/go.mod" ] && ok "does NOT stub construct/go.mod" || ko "construct/go.mod was created"

# ── Test 2: idempotent — a second call adds no duplicate row ──────────────────
ensure_go_tool_dependency "$ROOT/up" "cmd/sdlc" >/dev/null 2>&1
n="$(grep -c '^substrate \.\./up$' "$ROOT/leaf/construct/deps")"
[ "$n" = "1" ] && ok "idempotent: exactly one substrate row" || ko "duplicate rows: $n"

# ── Test 3: an existing construct/go.mod is left untouched (dual-read) ────────
wf "$ROOT/leaf2/construct/go.mod" <<EOF
module local.construct/leaf2
go 1.24
require example.com/up v0.0.0
replace example.com/up => ../../up
EOF
before="$(cat "$ROOT/leaf2/construct/go.mod")"
TARGET_DIR="$ROOT/leaf2"
ensure_go_tool_dependency "$ROOT/up" "cmd/sdlc" >/dev/null 2>&1
after="$(cat "$ROOT/leaf2/construct/go.mod")"
[ "$before" = "$after" ] && ok "legacy construct/go.mod left untouched" || ko "construct/go.mod was modified"
[ "$(cat "$ROOT/leaf2/construct/deps")" = "substrate ../up" ] && ok "substrate row written alongside legacy go.mod" || ko "deps row missing"

# ── Test 4: a second distinct owner gets its own row ──────────────────────────
TARGET_DIR="$ROOT/leaf"
ensure_go_tool_dependency "$ROOT/up2" "cmd/other" >/dev/null 2>&1
grep -q '^substrate \.\./up$'  "$ROOT/leaf/construct/deps" && grep -q '^substrate \.\./up2$' "$ROOT/leaf/construct/deps" \
    && ok "distinct owners → distinct substrate rows" || ko "second owner row missing"

# ── Test 5: integration — walk_manifest routes a real `tool` line to the writer ─
# Proves the parse→route→write path (the same path the real ariadne base.manifest
# `tool cmd/sdlc` takes), so we don't need to mutate a real peer to verify it.
wf "$ROOT/up3/construct/base.manifest" <<EOF
# minimal owner manifest
tool cmd/sdlc
EOF
mkdir -p "$ROOT/leaf3/construct"
TARGET_DIR="$ROOT/leaf3"
walk_manifest "$ROOT/up3" >/dev/null 2>&1
[ "$(cat "$ROOT/leaf3/construct/deps" 2>/dev/null)" = "substrate ../up3" ] \
    && ok "walk_manifest routes 'tool' action → substrate row in construct/deps" || ko "manifest-walk integration: [$(cat "$ROOT/leaf3/construct/deps" 2>/dev/null)]"

echo
echo "== $pass passed, $fail failed =="
[ "$fail" -eq 0 ]
