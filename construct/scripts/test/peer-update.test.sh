#!/usr/bin/env bash
# peer-update.test.sh — hermetic tests for bootstrap-peers.sh's update_peer():
# on re-run, an already-present peer is fast-forwarded to its upstream, but only
# when that's safe. Builds throwaway local git repos (bare "origins" + clones)
# under $TMPDIR; no network. The function under test is extracted verbatim from
# bootstrap-peers.sh so this guards the real code, not a copy.
#
# Run:  bash construct/scripts/test/peer-update.test.sh
# Exit: 0 if all assertions pass, 1 otherwise.
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
SRC="$ARIADNE_ROOT/construct/scripts/bootstrap-peers.sh"

ROOT="$(mktemp -d "${TMPDIR:-/tmp}/peerupdate.XXXXXX")"
trap 'rm -rf "$ROOT"' EXIT
cd "$ROOT"

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }

git_q() { git -c user.email=t@test -c user.name=test -c init.defaultBranch=main "$@"; }

# Load the real update_peer() from the script (def spans `update_peer() {` … `}`).
awk '/^update_peer\(\) \{/,/^\}/' "$SRC" > "$ROOT/fn.sh"
grep -q 'git -C "\$dir" pull --ff-only' "$ROOT/fn.sh" \
    && ok "extracted update_peer() from bootstrap-peers.sh" \
    || ko "could not extract update_peer() — function shape changed?"
# shellcheck disable=SC1091
source "$ROOT/fn.sh"

# Bare origin with one commit, plus a "pusher" clone to advance it later.
git_q init -q --bare origin.git
git_q clone -q origin.git pusher >/dev/null 2>&1
( cd pusher && echo v1 > f && git_q add f && git_q commit -qm v1 && git_q push -q origin main ) >/dev/null 2>&1
advance() { ( cd pusher && echo "$1" > f && git_q commit -qam "$1" && git_q push -q origin main ) >/dev/null 2>&1; }
at() { ( cd "$1" && head -1 f ); }

echo "== update_peer safety =="

# 1. CLEAN + tracking, remote ahead → fast-forwards.
git_q clone -q origin.git peerA >/dev/null 2>&1
advance v2
update_peer "$ROOT/peerA" peerA >/dev/null 2>&1
[[ "$(at peerA)" == v2 ]] && ok "clean+tracking, remote ahead → ff to v2" || ko "expected v2, got $(at peerA)"

# 2. Re-run when already current → idempotent no-op, exit 0.
update_peer "$ROOT/peerA" peerA >/dev/null 2>&1; rc=$?
[[ $rc -eq 0 && "$(at peerA)" == v2 ]] && ok "already current → idempotent no-op" || ko "rc=$rc, at=$(at peerA)"

# 3. DIRTY tree → skipped, local edit preserved (not pulled to remote).
git_q clone -q origin.git peerB >/dev/null 2>&1   # peerB at v2
advance v3
( cd peerB && echo wip >> f )                      # dirty
out=$(update_peer "$ROOT/peerB" peerB 2>&1)
[[ "$out" == *"working tree not clean"* && "$(at peerB)" == v2 ]] \
    && ok "dirty tree → skipped, WIP preserved (still v2, not v3)" \
    || ko "out=[$out] at=$(at peerB)"

# 4. Detached HEAD / no upstream → skipped.
git_q clone -q origin.git peerC >/dev/null 2>&1
( cd peerC && git_q checkout -q --detach HEAD )
out=$(update_peer "$ROOT/peerC" peerC 2>&1)
[[ "$out" == *"no upstream tracking branch"* ]] && ok "detached HEAD → skipped" || ko "out=[$out]"

# 5. Diverged branch → warn, non-fatal (rc 0), left as-is.
git_q clone -q origin.git peerD >/dev/null 2>&1
( cd peerD && echo local > g && git_q add g && git_q commit -qm local-divergent ) >/dev/null 2>&1
advance v4
out=$(update_peer "$ROOT/peerD" peerD 2>&1); rc=$?
[[ $rc -eq 0 && "$out" == *"leaving as-is"* ]] \
    && ok "diverged branch → warn, non-fatal (rc 0)" \
    || ko "rc=$rc out=[$out]"

echo
echo "== $pass passed, $fail failed =="
[[ $fail -eq 0 ]]
