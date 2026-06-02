#!/usr/bin/env bash
# clone-data-deps.sh — clone declared DATA DEPENDENCIES as siblings and mount
# them via symlink. A data dependency is *content this repo consumes* from
# another repo — NOT a substrate layer it inherits. It is a looser git
# submodule: a sibling clone (not nested), floating-HEAD (not pinned), with an
# independent history, surfaced into this repo's tree through a relative symlink.
#
# This is deliberately language-agnostic — unlike the go.mod substrate walk
# (bootstrap-peers.sh / setup.sh), which both clones AND symlinks the peer's
# base-layer files in. Data deps are just "clone + symlink", nothing more, so
# any repo (Go, TypeScript, or plain markdown like a brain) can declare them.
#
# Manifest: construct/deps `data` rows (#60) — `data <git-url> <symlink-path>`,
# '#' comments and blank lines ignored:
#
#     # <kind> <git-url>                            <symlink-path-relative-to-repo-root>
#     data     git@github.com:xianxu/you-decide.git  data/life/politics/you-decide
#
# (The legacy two-column construct/data-deps file was retired in #60 M5.)
#
# The repo is cloned to a SIBLING of this repo root named after the URL
# basename (you-decide). The symlink at <symlink-path> points at that clone
# with a computed RELATIVE target, so it survives the repo living at different
# absolute paths on different machines.
#
# Called by Makefile.workflow's `data-deps` target (an additive prereq of
# `bootstrap`). Idempotent: present clones are skipped; the symlink is
# re-pointed each run so a moved dep self-heals.
#
# Env overrides:
#   DATADEP_URL_<name>=<url>   Override the clone URL for dep <name>
#                             (name with non-alnum → _, e.g. DATADEP_URL_you_decide).
set -euo pipefail

# Shared construct/deps parser (#60). This script is symlinked into derivatives;
# resolve through the symlink to source the real sibling lib-deps.sh.
_DEPS_LIB_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || realpath "${BASH_SOURCE[0]}")")" && pwd)"
# shellcheck source=/dev/null
. "$_DEPS_LIB_DIR/lib-deps.sh"

TARGET_DIR="${TARGET_DIR:-$PWD}"
DEPS="$TARGET_DIR/construct/deps"            # unified manifest (#60)

if [[ ! -f "$DEPS" ]]; then
    # No data deps. Silent success — most repos have none. (#60 M5 retired the
    # legacy construct/data-deps carrier; data deps are `data` rows here now.)
    exit 0
fi

PARENT_DIR="$(cd "$TARGET_DIR/.." && pwd -P)"

# mount_data <git-url> <symlink-path>: clone the dep as a sibling (origin
# basename), mount it via a computed relative symlink. Idempotent: present
# clones skipped, symlink re-pointed each run. Deduped by clone-dest basename.
_dd_seen=()
mount_data() {
    local url="$1" symlink_rel="$2" name clone_dest symlink_abs symlink_parent rel ovar s
    name="$(basename "$url")"; name="${name%.git}"
    for s in "${_dd_seen[@]+"${_dd_seen[@]}"}"; do [[ "$s" == "$name" ]] && return 0; done
    _dd_seen+=("$name")

    # Per-dep URL override (DATADEP_URL_<name> with non-alnum → _).
    ovar="DATADEP_URL_${name//[^a-zA-Z0-9]/_}"
    url="${!ovar:-$url}"
    clone_dest="$PARENT_DIR/$name"

    if [[ -d "$clone_dest" ]]; then
        echo "data-deps: '$name' present ($clone_dest)"
    else
        echo "data-deps: cloning '$name'"
        echo "    from $url"
        echo "    into $clone_dest"
        git clone "$url" "$clone_dest"
    fi

    # Mount: relative symlink at $symlink_rel → $clone_dest.
    symlink_abs="$TARGET_DIR/$symlink_rel"
    symlink_parent="$(dirname "$symlink_abs")"
    mkdir -p "$symlink_parent"
    rel="$(python3 -c 'import os,sys; print(os.path.relpath(sys.argv[1], sys.argv[2]))' \
        "$clone_dest" "$symlink_parent")"
    ln -sfn "$rel" "$symlink_abs"
    echo "data-deps: mounted $symlink_rel -> $rel"
}

# `data` rows in the unified construct/deps manifest (#60).
echo "data-deps: walking $DEPS (data rows)"
while IFS=$'\t' read -r url mount; do
    [[ -n "$url" && -n "$mount" ]] && mount_data "$url" "$mount"
done < <(deps_data_rows "$TARGET_DIR")

echo "data-deps: done"
