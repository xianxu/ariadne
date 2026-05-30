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
# Manifest: construct/data-deps — one dep per line, two whitespace-separated
# columns; '#' comments and blank lines ignored:
#
#     # <git-url>                              <symlink-path-relative-to-repo-root>
#     git@github.com:xianxu/you-decide.git     data/life/politics/you-decide
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

TARGET_DIR="${TARGET_DIR:-$PWD}"
MANIFEST="$TARGET_DIR/construct/data-deps"

if [[ ! -f "$MANIFEST" ]]; then
    # No manifest = no data deps. Silent success — most repos have none.
    exit 0
fi

PARENT_DIR="$(cd "$TARGET_DIR/.." && pwd -P)"
echo "data-deps: walking $MANIFEST"

while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%%#*}"                       # strip comments
    # shellcheck disable=SC2086
    set -- $line                             # word-split on whitespace
    [[ $# -eq 0 ]] && continue               # blank / comment-only line
    if [[ $# -ne 2 ]]; then
        echo "data-deps: malformed line (need '<git-url> <symlink-path>'): $line" >&2
        exit 1
    fi
    url="$1"
    symlink_rel="$2"

    name="$(basename "$url")"
    name="${name%.git}"

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
done < "$MANIFEST"

echo "data-deps: done"
