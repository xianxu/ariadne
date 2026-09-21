#!/usr/bin/env bash
# Prepare local artifacts only. Publication and tap updates belong to #241.
set -euo pipefail
repo_root="$(cd "$(dirname "$0")/.." && pwd -P)"
if [ "$#" -ne 2 ]; then
    echo 'usage: scripts/release-weave.sh weave-vVERSION NEW_OUTPUT_DIRECTORY' >&2
    exit 2
fi
# Resolve against the caller before Go switches to the source module.
case "$2" in
    /*) output="$2" ;;
    *) output="$PWD/$2" ;;
esac
# go run prepends GOROOT/bin; preserve the caller tool selection for builds.
export WEAVE_RELEASE_CALLER_PATH="$PATH"
exec go -C "$repo_root" run ./cmd/weave/internal/release "$1" "$output"
