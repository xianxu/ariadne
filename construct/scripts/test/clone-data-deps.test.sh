#!/usr/bin/env bash
# clone-data-deps.test.sh — hermetic tests for clone-data-deps.sh reading BOTH
# the legacy two-column construct/data-deps AND the unified construct/deps `data`
# rows (#60), including dedup when a dep is declared in both carriers. No network:
# "origins" are local bare repos under $TMPDIR.
#
# Run:  bash construct/scripts/test/clone-data-deps.test.sh
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
CLONE_DD="$ARIADNE_ROOT/construct/scripts/clone-data-deps.sh"

ROOT="$(cd "$(mktemp -d "${TMPDIR:-/tmp}/clonedd.XXXXXX")" && pwd -P)"
trap 'rm -rf "$ROOT"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }
git_q() { git -c user.email=t@test -c user.name=test -c init.defaultBranch=main "$@"; }
wf() { mkdir -p "$(dirname "$1")"; cat > "$1"; }

# make_origin <name>: a bare git "origin" at $ROOT/origins/<name>.git
make_origin() {
    local name="$1" wd
    wd="$ROOT/build/$name"
    mkdir -p "$wd"; echo "$name" > "$wd/README"
    ( cd "$wd" && git_q init -q && git_q add -A && git_q commit -qm init ) >/dev/null 2>&1
    git_q clone -q --bare "$wd" "$ROOT/origins/$name.git" >/dev/null 2>&1
}

echo "== clone-data-deps: legacy data-deps + construct/deps data rows =="

# ── Test 1: legacy construct/data-deps clones + mounts ────────────────────────
t1() {
    local repo="$ROOT/t1/myrepo"
    make_origin alpha
    wf "$repo/construct/data-deps" <<EOF
$ROOT/origins/alpha.git   data/alpha
EOF
    ( cd "$repo" && "$CLONE_DD" ) >/dev/null 2>&1
    local rc=$?
    [[ $rc -eq 0 ]] && ok "legacy: exit 0" || ko "legacy: exit $rc"
    [[ -d "$ROOT/t1/alpha" ]]  && ok "legacy: sibling clone present" || ko "legacy: clone missing"
    [[ -L "$repo/data/alpha" ]] && ok "legacy: mount symlink present" || ko "legacy: mount missing"
}

# ── Test 2: construct/deps `data` row clones + mounts (#60) ───────────────────
t2() {
    local repo="$ROOT/t2/myrepo"
    make_origin beta
    wf "$repo/construct/deps" <<EOF
data $ROOT/origins/beta.git data/beta
EOF
    ( cd "$repo" && "$CLONE_DD" ) >/dev/null 2>&1
    local rc=$?
    [[ $rc -eq 0 ]] && ok "deps: exit 0" || ko "deps: exit $rc"
    [[ -d "$ROOT/t2/beta" ]]   && ok "deps: sibling clone present" || ko "deps: clone missing"
    [[ -L "$repo/data/beta" ]] && ok "deps: mount symlink present" || ko "deps: mount missing"
}

# ── Test 3: same dep in BOTH carriers mounts once (dedup by clone basename) ────
t3() {
    local repo="$ROOT/t3/myrepo"
    make_origin gamma
    wf "$repo/construct/data-deps" <<EOF
$ROOT/origins/gamma.git data/gamma
EOF
    wf "$repo/construct/deps" <<EOF
data $ROOT/origins/gamma.git data/gamma-dup
EOF
    local out; out=$( cd "$repo" && "$CLONE_DD" 2>&1 )
    [[ $? -eq 0 ]] && ok "dedup: exit 0" || ko "dedup: nonzero ($out)"
    # Legacy runs first → data/gamma mounted; the construct/deps dup is skipped,
    # so data/gamma-dup is never created.
    [[ -L "$repo/data/gamma" && ! -e "$repo/data/gamma-dup" ]] \
        && ok "dedup: dep mounted once, dup skipped" || ko "dedup: dup not skipped ($out)"
}

# ── Test 4: substrate rows in construct/deps are ignored by the data mounter ───
t4() {
    local repo="$ROOT/t4/myrepo"
    wf "$repo/construct/deps" <<EOF
substrate ../some-ancestor
EOF
    ( cd "$repo" && "$CLONE_DD" ) >/dev/null 2>&1
    [[ $? -eq 0 ]] && ok "substrate-only deps: exit 0 (no data action)" || ko "substrate-only: nonzero"
    [[ ! -e "$repo/data" ]] && ok "substrate-only deps: no spurious mount" || ko "substrate-only: spurious mount"
}

t1; t2; t3; t4

echo
echo "== $pass passed, $fail failed =="
[[ $fail -eq 0 ]]
