#!/usr/bin/env bash
# discover-ancestors.test.sh — hermetic tests for setup.sh's discover_ancestors
# ancestor walk (ariadne#50). The substrate-ancestor `replace` directive lives
# in a derivative's construct/go.mod, NOT its root go.mod (the root is the
# operator's app module). discover_ancestors must therefore walk BOTH go.mods
# per node — like list-peers.sh and bootstrap-peers.sh already do.
#
# The bug it guards: a root-only walk stops at depth 1. For a depth-2 chain
# (leaf → mid → base) where the mid→base hop is in mid/construct/go.mod, a
# root-only walk from leaf finds mid but never base, so base's base.manifest
# is never applied. This is exactly what broke brain → nous → ariadne: brain
# only ever got nous's layer, never ariadne's.
#
# Strategy: build throwaway layers under $TMPDIR, each shipping a
# construct/base.manifest (the filter discover_ancestors requires), and run
# `SETUP_DISCOVER_ONLY=1 setup.sh` from the leaf to print the discovered
# ancestor list without applying anything.
#
# Run:  bash construct/scripts/test/discover-ancestors.test.sh
# Exit: 0 if all assertions pass, 1 otherwise.
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
SETUP="$ARIADNE_ROOT/construct/setup.sh"

# pwd -P so ROOT matches discover_ancestors' canonicalized output (on macOS
# /tmp → /private/tmp; the walker emits physical paths).
ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/discover50.XXXXXX")" && pwd -P)"
trap 'rm -rf "$ROOT"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }

write_file() { mkdir -p "$(dirname "$1")"; cat > "$1"; }

# make_layer <abs-dir> <module-suffix>: a node with its own construct/base.manifest
# (so discover_ancestors' filter accepts it as an ancestor) and a real (empty)
# root go.mod (operator app module).
make_layer() {
    local dir="$1" mod="$2"
    write_file "$dir/go.mod" <<EOF
module example.com/$mod
go 1.24
EOF
    write_file "$dir/construct/base.manifest" <<EOF
# layer $mod
EOF
}

# substrate_replace <abs-dir> <upstream-rel> <upstream-mod>: declare the
# substrate ancestor in construct/go.mod (the derivative convention) — NOT root.
substrate_replace() {
    local dir="$1" rel="$2" mod="$3"
    write_file "$dir/construct/go.mod" <<EOF
module example.com/$(basename "$dir")-construct
go 1.24
require example.com/$mod v0.0.0
replace example.com/$mod => $rel
EOF
}

discover() { # <leaf-dir> → prints discovered ancestors (foundation-first)
    ( cd "$1" && SETUP_DISCOVER_ONLY=1 PATH="/usr/bin:/bin:/usr/local/bin:$PATH" \
        bash "$SETUP" 2>/dev/null )
}

echo "== discover_ancestors: both-go.mods walk =="

# ── Test 1: depth-2, ancestor declared in mid's construct/go.mod (THE BUG) ────
# leaf → mid → base. leaf/construct/go.mod replaces mid; mid/construct/go.mod
# replaces base. Root-only walk would find mid but never base.
t1() {
    local c="$ROOT/t1"
    make_layer "$c/base" base
    make_layer "$c/mid"  mid
    make_layer "$c/leaf" leaf
    # leaf consumes mid (in construct/go.mod); mid consumes base (in construct/go.mod)
    substrate_replace "$c/leaf" ../../mid  mid
    substrate_replace "$c/mid"  ../../base base
    local out; out="$(discover "$c/leaf")"
    grep -qx "$c/mid"  <<<"$out" && ok "depth-1 ancestor (mid) discovered"        || ko "mid missing: [$out]"
    grep -qx "$c/base" <<<"$out" && ok "depth-2 ancestor (base) discovered (#50)" || ko "base missing — the bug: [$out]"
}

# ── Test 2: ordering is foundation-first (deepest ancestor first) ─────────────
t2() {
    local c="$ROOT/t1"   # reuse t1's tree
    local out; out="$(discover "$c/leaf")"
    local first last
    first="$(head -1 <<<"$out")"; last="$(tail -1 <<<"$out")"
    [[ "$first" == "$c/base" ]] && ok "ordering: base (deepest) first" || ko "ordering: first=$first"
    [[ "$last"  == "$c/mid"  ]] && ok "ordering: mid (shallowest) last" || ko "ordering: last=$last"
}

# ── Test 3: construct-only derivative (no root go.mod) still walks ────────────
# A pure-data consumer may have only construct/go.mod. The walk must still enter.
t3() {
    local c="$ROOT/t3"
    make_layer "$c/base" base3
    make_layer "$c/leaf" leaf3
    rm -f "$c/leaf/go.mod"                       # no root go.mod at all
    substrate_replace "$c/leaf" ../../base base3
    local out; out="$(discover "$c/leaf")"
    grep -qx "$c/base" <<<"$out" && ok "construct-only leaf discovers ancestor" || ko "construct-only: base missing: [$out]"
}

# ── Test 4: non-manifest replace target filtered; self never an ancestor ──────
# leaf replaces `plain` (which ships NO construct/base.manifest). plain must not
# appear as an ancestor; nor may leaf list itself. (With no manifest-bearing
# ancestor found, Source-3 may fall back to the script's own upstream — that's
# expected; we only assert plain and self are absent.)
t4() {
    local c="$ROOT/t4"
    make_layer "$c/leaf" leaf4
    mkdir -p "$c/plain"; write_file "$c/plain/go.mod" <<<'module example.com/plain'  # no base.manifest
    substrate_replace "$c/leaf" ../../plain plain
    local out; out="$(discover "$c/leaf")"
    grep -qx "$c/plain" <<<"$out" && ko "non-manifest dir 'plain' leaked in: [$out]" || ok "non-manifest replace target filtered"
    grep -qx "$c/leaf"  <<<"$out" && ko "leaf listed itself as ancestor: [$out]"    || ok "self never an ancestor"
}

# substrate_deps <abs-dir> <root-relative-path>: declare the substrate ancestor
# in construct/deps (#60) — repo-root-relative (../base), not construct-relative.
substrate_deps() {
    write_file "$1/construct/deps" <<EOF
substrate $2
EOF
}

# ── Test 5: depth-2 ancestor discovery via construct/deps (#60 dual-read) ──────
# leaf → mid → base, but each hop declared in construct/deps, not construct/go.mod.
t5() {
    local c="$ROOT/t5"
    make_layer "$c/base" base5
    make_layer "$c/mid"  mid5
    make_layer "$c/leaf" leaf5
    substrate_deps "$c/leaf" ../mid
    substrate_deps "$c/mid"  ../base
    local out; out="$(discover "$c/leaf")"
    grep -qx "$c/mid"  <<<"$out" && ok "deps: depth-1 ancestor (mid) discovered"        || ko "deps mid missing: [$out]"
    grep -qx "$c/base" <<<"$out" && ok "deps: depth-2 ancestor (base) discovered (#60)" || ko "deps base missing: [$out]"
}

# ── Test 6: deps-only leaf (no root go.mod, no construct/go.mod) still walks ───
t6() {
    local c="$ROOT/t6"
    make_layer "$c/base" base6
    make_layer "$c/leaf" leaf6
    rm -f "$c/leaf/go.mod"                 # no root go.mod
    substrate_deps "$c/leaf" ../base       # only construct/deps declares the ancestor
    local out; out="$(discover "$c/leaf")"
    grep -qx "$c/base" <<<"$out" && ok "deps-only leaf discovers ancestor" || ko "deps-only: base missing: [$out]"
}

t1; t2; t3; t4; t5; t6

echo
echo "== $pass passed, $fail failed =="
[[ $fail -eq 0 ]]
