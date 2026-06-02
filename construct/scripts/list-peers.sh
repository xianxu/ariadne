#!/usr/bin/env bash
# list-peers.sh — print the set of sibling repos that make up a repo's
# substrate, derived from go.mod (recursively). The canonical PRESENT-PEERS
# walker for the base layer (ariadne#44).
#
# Usage:
#   list-peers.sh [<repo-path>] [<extra-repo> ...]
#
# Default <repo-path> is the current directory. Output is one absolute path per
# line, starting with the repo itself, followed by every transitive substrate
# upstream declared via `replace <module> => <local-path>` in any reached
# go.mod, followed by the transitive set of each <extra-repo>. Deduped.
#
# Consumers (this is the shared walker; keep them on this one file):
#   - .tart/Makefile via the back-compat symlink .tart/scripts/tart-list-peers.sh
#     → which repos to APFS-clone into the tart VM (ariadne#32 phase 2).
#   - .openshell/sandbox.sh → which peers to mutagen-sync into the sandbox.
# `setup.sh:discover_ancestors` shares the SAME grammar but keeps its own walk
# (it has extra sources: `go list -m all` + a no-go.mod fallback) — not a clean
# swap, so it stays inline; just keep the replace regex in sync.
#
# PRESENT-PEERS vs CLONE-ABSENT: this walker SKIPS replace targets that don't
# exist on disk (peers are expected present). bootstrap.sh runs the same grammar
# in the opposite mode — it resolves absent targets syntactically and CLONES
# them (it runs pre-substrate). bootstrap.sh keeps an inline copy of the parser
# (it can't source this symlinked script on a bare clone) locked to this one by
# construct/scripts/test/bootstrap-transitive.test.sh. Keep the regex identical.
#
# <extra-repo> args (ariadne#44 decision 9): additional BFS seeds, unioned with
# the go.mod peer set. The openshell sandbox passes SYNC=../x entries here so
# explicitly-requested repos join the synced set; tart passes them to widen the
# VM clone. Writability/RO classification is the CALLER's job — this script only
# decides membership.
#
# For each repo in the walk we read its construct/deps `substrate` rows (the
# substrate ancestor, #60) AND its root go.mod `replace` directives (real Go
# app-dep siblings, e.g. brain's `replace nous`). The legacy construct/go.mod
# substrate carrier is no longer read (#60 M4). construct/ is never emitted as a
# peer itself — only its targets are.
#
# If neither construct/deps nor a go.mod is reachable, output is just the repo
# itself — matches the pre-Go-modules single-repo behavior.
set -euo pipefail

# Shared construct/deps parser (#60). This script is symlinked into derivatives,
# so resolve through the symlink to source the real sibling lib-deps.sh.
_DEPS_LIB_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || realpath "${BASH_SOURCE[0]}")")" && pwd)"
# shellcheck source=/dev/null
. "$_DEPS_LIB_DIR/lib-deps.sh"

repo="$(cd "${1:-.}" && pwd -P)"
shift || true   # remaining args ($@) are extra-repo seeds

seen=()
peers=()
queue=("$repo")

# Seed extra repos (ariadne#44): resolve each to an absolute path and enqueue
# alongside the root. Non-directories are skipped silently (caller's concern).
for extra in "$@"; do
    abs="$(cd "$extra" 2>/dev/null && pwd -P || true)"
    [[ -n "$abs" && -d "$abs" ]] && queue+=("$abs")
done

_is_seen() {
    local p="$1" s
    for s in "${seen[@]+"${seen[@]}"}"; do
        [[ "$s" == "$p" ]] && return 0
    done
    return 1
}

# _enqueue_replaces <dir>: parse <dir>/go.mod for `replace <module> => <local-path>`
# directives and append each resolved absolute peer path to the BFS queue. rhs
# paths resolve relative to <dir> (the go.mod's own directory), per Go's replace
# semantics. No-op if <dir>/go.mod is absent. Appends to the global `queue`.
_enqueue_replaces() {
    local dir="$1" line rhs abs
    [[ -f "$dir/go.mod" ]] || return 0
    while IFS= read -r line; do
        line="${line%%//*}"
        if [[ "$line" =~ ^[[:space:]]*replace[[:space:]]+[^[:space:]]+([[:space:]]+[^[:space:]]+)?[[:space:]]+=\>[[:space:]]+([^[:space:]]+) ]]; then
            rhs="${BASH_REMATCH[2]}"
            if [[ "$rhs" == /* || "$rhs" == ./* || "$rhs" == ../* ]]; then
                if [[ "$rhs" == /* ]]; then
                    abs="$rhs"
                else
                    abs="$(cd "$dir" && cd "$rhs" 2>/dev/null && pwd -P || true)"
                fi
                if [[ -n "$abs" && -d "$abs" ]]; then
                    queue+=("$abs")
                fi
            fi
        fi
    done < "$dir/go.mod"
}

while [[ ${#queue[@]} -gt 0 ]]; do
    current="${queue[0]}"
    queue=("${queue[@]:1}")

    _is_seen "$current" && continue
    seen+=("$current")
    peers+=("$current")

    # Substrate is declared in construct/deps (#60); real Go app-dep siblings
    # (e.g. brain's `replace nous => ../nous`) still live in the repo-root go.mod
    # and are walked too. The legacy construct/go.mod substrate carrier is no
    # longer read (#60 M4 dropped the dual-read fallback — all derivatives carry
    # construct/deps). construct/ itself is never a peer; only its targets are.
    _enqueue_replaces "$current"
    while IFS= read -r dep_abs; do
        [[ -n "$dep_abs" && -d "$dep_abs" ]] && queue+=("$dep_abs")
    done < <(deps_substrate_targets "$current")
done

printf '%s\n' "${peers[@]}"
