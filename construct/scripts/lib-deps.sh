#!/usr/bin/env bash
# lib-deps.sh — shared parser for construct/deps, the unified substrate + data
# dependency manifest (ariadne#60). Sourceable: defines functions, no side
# effects on source. The DRY core that the present-peers walkers, the clone
# drivers, and the data mounter all share.
#
# Format: one dep per line, whitespace-split positional columns; '#' comments
# and blank lines ignored. Same grammar class as the legacy two-column
# construct/data-deps (so the parser is a trivial generalization):
#
#     # <kind>      <target>                            [<mount>]
#     substrate    ../ariadne
#     data         git@github.com:xianxu/you-decide.git data/life/politics/you-decide
#
# kind=substrate — <target> is a sibling path relative to the REPO ROOT (e.g.
#   `../ariadne`, NOT the construct-relative `../../ariadne` the old go.mod
#   `replace` used). Resolves to the substrate ancestor; no mount column.
# kind=data — <target> is a git URL, <mount> is the symlink path (repo-root-
#   relative). Subsumes legacy construct/data-deps verbatim.
#
# DUAL-READ (transition): during the #60 rollout the substrate graph lives in
# BOTH construct/go.mod (legacy) and construct/deps (new). Walkers read both;
# this file only parses construct/deps.
#
# bootstrap.sh CANNOT source this — it runs on a bare clone where the symlinked
# construct/ dangles — so it keeps an INLINE copy of the substrate parse, locked
# to deps_substrate_targets() by construct/scripts/test/bootstrap-transitive.test.sh.
# Keep the two in lockstep.

# deps_substrate_targets <repo_root>
#   Print one absolute peer path per `substrate` row, resolved SYNTACTICALLY
#   relative to <repo_root> (via the always-present parent dir, so an absent peer
#   still resolves — each caller decides present-skip vs clone-absent). No-op if
#   <repo_root>/construct/deps is absent.
deps_substrate_targets() {
    local repo_root="$1" deps="$1/construct/deps"
    local line kind target raw parent
    [[ -f "$deps" ]] || return 0
    while IFS= read -r line || [[ -n "$line" ]]; do
        line="${line%%#*}"
        # shellcheck disable=SC2086
        set -- $line                              # word-split on whitespace
        [[ $# -ge 2 ]] || continue                # blank / comment / malformed
        kind="$1"; target="$2"
        [[ "$kind" == "substrate" ]] || continue
        if [[ "$target" == /* ]]; then raw="$target"; else raw="$repo_root/$target"; fi
        parent="$(cd "$(dirname "$raw")" 2>/dev/null && pwd -P || true)"
        # Present-peers semantics: an unresolvable parent is skipped silently.
        # bootstrap.sh's inline walk_deps DELIBERATELY differs (it aborts loudly)
        # — a clone driver should reject a broken manifest, a present-walker shouldn't.
        [[ -n "$parent" ]] || continue
        printf '%s\n' "$parent/$(basename "$raw")"
    done < "$deps"
}

# deps_data_rows <repo_root>
#   Print `url<TAB>mount` per `data` row. No-op if construct/deps is absent.
deps_data_rows() {
    local repo_root="$1" deps="$1/construct/deps"
    local line kind url mount
    [[ -f "$deps" ]] || return 0
    while IFS= read -r line || [[ -n "$line" ]]; do
        line="${line%%#*}"
        # shellcheck disable=SC2086
        set -- $line
        [[ $# -ge 3 ]] || continue                # data needs url + mount
        kind="$1"; url="$2"; mount="$3"
        [[ "$kind" == "data" ]] || continue
        printf '%s\t%s\n' "$url" "$mount"
    done < "$deps"
}
