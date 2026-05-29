#!/usr/bin/env bash
# bootstrap-transitive.test.sh — hermetic tests for bootstrap.sh's transitive
# clone walk (ariadne#45). Builds throwaway local git repos (bare "origins" +
# cold clones) under $TMPDIR and exercises bootstrap.sh end to end, with no
# network and no real `make` handoff (BOOTSTRAP_CLONE_ONLY / BOOTSTRAP_DRY_RUN).
#
# Run:  bash construct/scripts/test/bootstrap-transitive.test.sh
# Exit: 0 if all assertions pass, 1 otherwise.
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
BOOTSTRAP="$ARIADNE_ROOT/bootstrap.sh"
LIST_PEERS="$ARIADNE_ROOT/construct/scripts/list-peers.sh"

ROOT="$(mktemp -d "${TMPDIR:-/tmp}/bootstrap45.XXXXXX")"
trap 'rm -rf "$ROOT"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }

git_q() { git -c user.email=t@test -c user.name=test -c init.defaultBranch=main "$@"; }

# write_file <path> <<<content via heredoc on stdin>
write_file() { mkdir -p "$(dirname "$1")"; cat > "$1"; }

# commit_and_bare <case-dir> <name>: turn build/<name> into origins/<name>.git
commit_and_bare() {
    local cs="$1" name="$2"
    local wd="$ROOT/$cs/build/$name"
    ( cd "$wd" && git_q init -q && git_q add -A && git_q commit -qm init ) >/dev/null 2>&1
    git_q clone -q --bare "$wd" "$ROOT/$cs/origins/$name.git" >/dev/null 2>&1
}

# cold_clone <case-dir> <name>: clone origins/<name>.git into ws/<name> (peerless)
cold_clone() {
    local cs="$1" name="$2"
    git_q clone -q "$ROOT/$cs/origins/$name.git" "$ROOT/$cs/ws/$name" >/dev/null 2>&1
}

# inject_bootstrap <build-repo-dir>: drop the bootstrap.sh under test in (sim seed)
inject_bootstrap() { cp "$BOOTSTRAP" "$1/bootstrap.sh"; chmod +x "$1/bootstrap.sh"; }

# Helper to write a construct/go.mod with a single sibling replace.
construct_replace() { # <build-repo> <peer-name>
    write_file "$1/construct/go.mod" <<EOF
module local.construct/$(basename "$1")
require example.com/$2 v0.0.0
replace example.com/$2 => ../../$2
EOF
}

echo "== bootstrap.sh transitive clone walk =="

# ── Test 1: 3-deep cold-start clones the WHOLE chain ──────────────────────────
# zztop → zzmid → zzbase. Cold-clone only zztop; expect zzmid AND zzbase appear.
t1() {
    local c=t1 b="$ROOT/t1/build"
    mkdir -p "$b/zztop" "$b/zzmid" "$b/zzbase"
    write_file "$b/zztop/go.mod"  <<<'module example.com/zztop'
    construct_replace "$b/zztop" zzmid
    inject_bootstrap "$b/zztop"
    write_file "$b/zzmid/go.mod"  <<<'module example.com/zzmid'
    construct_replace "$b/zzmid" zzbase
    write_file "$b/zzbase/go.mod" <<<'module example.com/zzbase'   # terminal: no replace
    commit_and_bare "$c" zztop; commit_and_bare "$c" zzmid; commit_and_bare "$c" zzbase
    cold_clone "$c" zztop

    ( cd "$ROOT/$c/ws/zztop" && BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh ) >/dev/null 2>&1
    local rc=$?
    [[ $rc -eq 0 ]] && ok "3-deep: exit 0" || ko "3-deep: exit $rc"
    [[ -d "$ROOT/$c/ws/zzmid"  ]] && ok "3-deep: direct peer zzmid cloned"      || ko "3-deep: zzmid missing"
    [[ -d "$ROOT/$c/ws/zzbase" ]] && ok "3-deep: TRANSITIVE peer zzbase cloned" || ko "3-deep: zzbase missing (the bug)"
}

# ── Test 2: 2-deep regression (the common case) ───────────────────────────────
t2() {
    local c=t2 b="$ROOT/t2/build"
    mkdir -p "$b/zztop" "$b/zzbase"
    write_file "$b/zztop/go.mod" <<<'module example.com/zztop'
    construct_replace "$b/zztop" zzbase
    inject_bootstrap "$b/zztop"
    write_file "$b/zzbase/go.mod" <<<'module example.com/zzbase'
    commit_and_bare "$c" zztop; commit_and_bare "$c" zzbase
    cold_clone "$c" zztop
    ( cd "$ROOT/$c/ws/zztop" && BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh ) >/dev/null 2>&1
    local rc=$?
    [[ $rc -eq 0 && -d "$ROOT/$c/ws/zzbase" ]] && ok "2-deep: direct peer cloned, exit 0" || ko "2-deep: rc=$rc, zzbase $( [[ -d $ROOT/$c/ws/zzbase ]] && echo present || echo missing)"
}

# ── Test 3: replace cycle terminates (visited-set) ────────────────────────────
t3() {
    local c=t3 b="$ROOT/t3/build"
    mkdir -p "$b/zzc1" "$b/zzc2"
    write_file "$b/zzc1/go.mod" <<<'module example.com/zzc1'
    construct_replace "$b/zzc1" zzc2
    inject_bootstrap "$b/zzc1"
    write_file "$b/zzc2/go.mod" <<<'module example.com/zzc2'
    construct_replace "$b/zzc2" zzc1   # cycle back
    commit_and_bare "$c" zzc1; commit_and_bare "$c" zzc2
    cold_clone "$c" zzc1
    ( cd "$ROOT/$c/ws/zzc1" && BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh ) >/dev/null 2>&1
    local rc=$?
    [[ $rc -eq 0 && -d "$ROOT/$c/ws/zzc2" ]] && ok "cycle: terminates, exit 0, peer cloned" || ko "cycle: rc=$rc"
}

# ── Test 4: chain deeper than MAX_DEPTH errors clearly ────────────────────────
t4() {
    local c=t4 b="$ROOT/t4/build"
    mkdir -p "$b"/zzd{0,1,2,3}
    for i in 0 1 2 3; do write_file "$b/zzd$i/go.mod" <<<"module example.com/zzd$i"; done
    construct_replace "$b/zzd0" zzd1
    construct_replace "$b/zzd1" zzd2
    construct_replace "$b/zzd2" zzd3   # depth 3 > cap 2
    inject_bootstrap "$b/zzd0"
    for i in 0 1 2 3; do commit_and_bare "$c" "zzd$i"; done
    cold_clone "$c" zzd0
    local out
    out=$( cd "$ROOT/$c/ws/zzd0" && BOOTSTRAP_CLONE_ONLY=1 BOOTSTRAP_MAX_DEPTH=2 ./bootstrap.sh 2>&1 )
    local rc=$?
    [[ $rc -ne 0 ]] && ok "depth: errors (exit $rc) past cap" || ko "depth: did not error (exit 0)"
    grep -q "MAX_DEPTH" <<<"$out" && ok "depth: message names MAX_DEPTH" || ko "depth: no MAX_DEPTH message"
}

# ── Test 5: no go.mod → hand off (clone-only: no-op, exit 0) ───────────────────
t5() {
    local c=t5 b="$ROOT/t5/build"
    mkdir -p "$b/zznogo"
    write_file "$b/zznogo/README" <<<'no go.mod here'
    inject_bootstrap "$b/zznogo"
    commit_and_bare "$c" zznogo
    cold_clone "$c" zznogo
    local out
    out=$( cd "$ROOT/$c/ws/zznogo" && BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh 2>&1 )
    local rc=$?
    [[ $rc -eq 0 ]] && grep -q "no go.mod" <<<"$out" && ok "no-go.mod: handoff path, exit 0" || ko "no-go.mod: rc=$rc out=[$out]"
}

# ── Test 6: origin-less + absent peer → clear error ───────────────────────────
t6() {
    local c=t6 b="$ROOT/t6/build"
    mkdir -p "$b/zztop"
    write_file "$b/zztop/go.mod" <<<'module example.com/zztop'
    construct_replace "$b/zztop" zzmid   # zzmid never created
    inject_bootstrap "$b/zztop"
    commit_and_bare "$c" zztop
    cold_clone "$c" zztop
    ( cd "$ROOT/$c/ws/zztop" && git remote remove origin )   # strip origin
    local out
    out=$( cd "$ROOT/$c/ws/zztop" && BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh 2>&1 )
    local rc=$?
    [[ $rc -ne 0 ]] && grep -q "no 'origin' remote" <<<"$out" && ok "origin-less: errors with guidance" || ko "origin-less: rc=$rc out=[$out]"
}

# ── Test 7: PEER_URL_<name> override clones without origin ─────────────────────
t7() {
    local c=t7 b="$ROOT/t7/build"
    mkdir -p "$b/zztop" "$b/zzmid"
    write_file "$b/zztop/go.mod" <<<'module example.com/zztop'
    construct_replace "$b/zztop" zzmid
    inject_bootstrap "$b/zztop"
    write_file "$b/zzmid/go.mod" <<<'module example.com/zzmid'
    commit_and_bare "$c" zztop; commit_and_bare "$c" zzmid
    cold_clone "$c" zztop
    ( cd "$ROOT/$c/ws/zztop" && git remote remove origin )
    ( cd "$ROOT/$c/ws/zztop" && PEER_URL_zzmid="$ROOT/$c/origins/zzmid.git" BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh ) >/dev/null 2>&1
    local rc=$?
    [[ $rc -eq 0 && -d "$ROOT/$c/ws/zzmid" ]] && ok "PEER_URL override: clones absent peer" || ko "PEER_URL override: rc=$rc"
}

# ── Test 8: DRIFT — DRY_RUN peer set == tart-list-peers.sh (minus root) ────────
t8() {
    local c=t8 b="$ROOT/t8/build"
    mkdir -p "$b/zztop" "$b/zzmid" "$b/zzbase"
    write_file "$b/zztop/go.mod"  <<<'module example.com/zztop'
    construct_replace "$b/zztop" zzmid
    inject_bootstrap "$b/zztop"
    write_file "$b/zzmid/go.mod"  <<<'module example.com/zzmid'
    construct_replace "$b/zzmid" zzbase
    write_file "$b/zzbase/go.mod" <<<'module example.com/zzbase'
    commit_and_bare "$c" zztop; commit_and_bare "$c" zzmid; commit_and_bare "$c" zzbase
    # all present
    cold_clone "$c" zztop; cold_clone "$c" zzmid; cold_clone "$c" zzbase
    local top="$ROOT/$c/ws/zztop"
    local dr lp
    dr=$( cd "$top" && BOOTSTRAP_DRY_RUN=1 ./bootstrap.sh 2>/dev/null | sort -u )
    lp=$( "$LIST_PEERS" "$top" 2>/dev/null | grep -v "/zztop$" | sort -u )
    if [[ "$dr" == "$lp" && -n "$dr" ]]; then
        ok "drift: DRY_RUN peer set matches tart-list-peers.sh"
    else
        ko "drift: mismatch"
        printf '    bootstrap dry-run:\n%s\n    list-peers (no root):\n%s\n' "$dr" "$lp"
    fi
}

t1; t2; t3; t4; t5; t6; t7; t8

echo
echo "== $pass passed, $fail failed =="
[[ $fail -eq 0 ]]
