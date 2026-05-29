#!/usr/bin/env bash
# sandbox-sync-set.test.sh — hermetic tests for .openshell/sandbox.sh's
# compute_sync_set (ariadne#44): the pure mapping from (current repo, go.mod
# peers, SYNC= extras) → per-repo rw/ro classification. No docker/mutagen/ssh:
# sandbox.sh is sourced with SANDBOX_LIB_ONLY=1 and compute_sync_set is exercised
# against throwaway go.mod fixture trees under $TMPDIR.
#
# Run:  bash construct/scripts/test/sandbox-sync-set.test.sh
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
SANDBOX="$ARIADNE_ROOT/.openshell/sandbox.sh"
LIST_PEERS="$ARIADNE_ROOT/construct/scripts/list-peers.sh"

ROOT="$(mktemp -d "${TMPDIR:-/tmp}/sandbox44.XXXXXX")"
trap 'rm -rf "$ROOT"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }
wf() { mkdir -p "$(dirname "$1")"; cat > "$1"; }

# Build a fixture workspace: zztop → zzmid → zzbase (go.mod peers), plus a
# standalone zznew (not a peer). Every repo is present on disk.
WS="$ROOT/ws"
wf "$WS/zztop/go.mod" <<<'module example.com/zztop'
wf "$WS/zztop/construct/go.mod" <<EOF
module local.construct/zztop
replace example.com/zzmid => ../../zzmid
EOF
# list-peers.sh must resolve from REPO_DIR/construct/scripts/ — symlink the real one.
mkdir -p "$WS/zztop/construct/scripts"
ln -s "$LIST_PEERS" "$WS/zztop/construct/scripts/list-peers.sh"
wf "$WS/zzmid/go.mod" <<<'module example.com/zzmid'
wf "$WS/zzmid/construct/go.mod" <<EOF
module local.construct/zzmid
replace example.com/zzbase => ../../zzbase
EOF
wf "$WS/zzbase/go.mod" <<<'module example.com/zzbase'
wf "$WS/zznew/go.mod" <<<'module example.com/zznew'

# Collision fixture: zzcollide declares two peers with the SAME basename in
# different parents — both would map to ~/workspace/shared.
wf "$WS/zzcollide/go.mod" <<<'module example.com/zzcollide'
wf "$WS/zzcollide/construct/go.mod" <<EOF
module local.construct/zzcollide
replace example.com/dup1 => ../../dupa/shared
replace example.com/dup2 => ../../dupb/shared
EOF
mkdir -p "$WS/zzcollide/construct/scripts"
ln -s "$LIST_PEERS" "$WS/zzcollide/construct/scripts/list-peers.sh"
wf "$WS/dupa/shared/go.mod" <<<'module example.com/dupa-shared'
wf "$WS/dupb/shared/go.mod" <<<'module example.com/dupb-shared'

# Source sandbox.sh for its functions only (no dispatch, no IO).
SANDBOX_LIB_ONLY=1 source "$SANDBOX" >/dev/null 2>&1
set +e  # sandbox.sh enables `set -e`; we want to run all assertions

# run_set <repo> <SYNC value>  → prints "mode name" lines, sorted
run_set() {
    REPO_DIR="$1" SYNC="$2" compute_sync_set | awk -F'\t' '{print $1, $3}' | sort
}

echo "== compute_sync_set rw/ro classification =="

# ── Test 1: no SYNC — current repo rw, all go.mod peers ro ────────────────────
got="$(run_set "$WS/zztop" "")"
exp="$(printf 'ro zzbase\nro zzmid\nrw zztop\n' | sort)"
[[ "$got" == "$exp" ]] && ok "no SYNC: zztop rw; zzmid, zzbase ro" || { ko "no SYNC"; printf '   got:\n%s\n   exp:\n%s\n' "$got" "$exp"; }

# ── Test 2: SYNC=../zzmid upgrades that peer to rw ────────────────────────────
got="$(run_set "$WS/zztop" "../zzmid")"
exp="$(printf 'ro zzbase\nrw zzmid\nrw zztop\n' | sort)"
[[ "$got" == "$exp" ]] && ok "SYNC=../zzmid: zzmid flips ro→rw, zzbase stays ro" || { ko "SYNC=../zzmid"; printf '   got:\n%s\n   exp:\n%s\n' "$got" "$exp"; }

# ── Test 3: SYNC of a NON-peer repo includes it as rw ─────────────────────────
got="$(run_set "$WS/zztop" "../zznew")"
exp="$(printf 'ro zzbase\nro zzmid\nrw zznew\nrw zztop\n' | sort)"
[[ "$got" == "$exp" ]] && ok "SYNC=../zznew: non-peer joins set as rw (union seed)" || { ko "SYNC=../zznew"; printf '   got:\n%s\n   exp:\n%s\n' "$got" "$exp"; }

# ── Test 4: multiple SYNC entries (comma-separated) ───────────────────────────
got="$(run_set "$WS/zztop" "../zzmid,../zzbase")"
exp="$(printf 'rw zzbase\nrw zzmid\nrw zztop\n' | sort)"
[[ "$got" == "$exp" ]] && ok "SYNC=../zzmid,../zzbase: both flip rw" || { ko "SYNC multi"; printf '   got:\n%s\n   exp:\n%s\n' "$got" "$exp"; }

# ── Test 5: current repo is always rw, even with empty peer graph ─────────────
got="$(run_set "$WS/zzbase" "")"   # zzbase has no construct/go.mod replace
exp="rw zzbase"
[[ "$got" == "$exp" ]] && ok "leaf repo (no peers): just itself, rw" || { ko "leaf"; printf '   got:\n%s\n' "$got"; }

# ── Test 6: basename collision fails loudly (not a silent peer drop) ──────────
REPO_DIR="$WS/zzcollide" SYNC="" compute_sync_set >/dev/null 2>&1
[[ $? -ne 0 ]] && ok "collision: same-basename peers error (exit ≠ 0)" || ko "collision: did not error"

echo
echo "== $pass passed, $fail failed =="
[[ $fail -eq 0 ]]
