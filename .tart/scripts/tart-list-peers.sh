#!/usr/bin/env bash
# tart-list-peers.sh — print the set of sibling repos to clone into
# the tart VM, derived from the current repo's go.mod (recursively).
#
# Usage:
#   tart-list-peers.sh [<repo-path>]
#
# Default <repo-path> is the current directory. Output is one absolute
# path per line, starting with the repo itself, followed by every
# transitive substrate upstream declared via `replace <module> => <local-path>`
# in any reached go.mod.
#
# For each repo in the walk we read BOTH its root go.mod AND its
# construct/go.mod. In ariadne-derivative repos the substrate dependency
# (the ariadne replace) is declared in construct/go.mod, not the root
# module — so a root-only walk would miss it and clone the repo alone.
# This matches bootstrap-peers.sh (reads $repo/construct/go.mod) and
# Makefile.workflow's `build:` (builds inside construct/). construct/ is
# never emitted as a peer itself — only its replace *targets* are — so the
# construct/ subdir stays nested inside its owning repo's clone.
#
# Same parser shape as construct/setup.sh's discover_ancestors so peers
# tracked by the VM clone match peers walked by setup.sh's manifest
# resolution. If no go.mod is reachable, output is just the repo itself —
# matches the pre-Go-modules tart behavior of single-repo clone.
set -euo pipefail

repo="$(cd "${1:-.}" && pwd -P)"

seen=()
peers=()
queue=("$repo")

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

    # Substrate replaces may live in the repo-root go.mod and/or construct/go.mod
    # (the derivative convention — see header). Walk both; construct/ itself is
    # never enqueued, only its replace targets.
    _enqueue_replaces "$current"
    _enqueue_replaces "$current/construct"
done

printf '%s\n' "${peers[@]}"
