#!/usr/bin/env bash
# Generic merge-gate runner (ariadne base layer) — issue #52.
#
#   run-merge-checks.sh <base_sha> <head_sha>
#
# Runs every executable in <repo>/scripts/merge-checks.d/ over the range
# base..head. Each check is invoked as:
#
#     <check> <base_sha> <head_sha>      exit 0 = pass, non-zero = fail
#
# Aggregates: exit 0 iff all checks pass — OR no checks are defined (a repo
# with no publish gate is a no-op pass). Per-check findings go to stderr.
# Checks run in filename order, so `10-`, `20-` prefixes give ordering.
#
# One runner, two call sites (kept DRY):
#   - .github/workflows/merge-check.yml  — CI, over the PR's merge-base..head
#   - a repo's local pre-push hook        — fast local feedback before publish
#
# This script is symlinked into derivatives, so $ROOT resolves to the
# CONSUMING repo and it runs that repo's own merge-checks.d/.
set -euo pipefail

BASE="${1:?usage: run-merge-checks.sh <base_sha> <head_sha>}"
HEAD="${2:?usage: run-merge-checks.sh <base_sha> <head_sha>}"

ROOT="$(git rev-parse --show-toplevel)"
DIR="$ROOT/scripts/merge-checks.d"

# Collect executable, non-README checks (bash 3.2-safe: build the array, then
# guard every expansion behind a count check so `set -u` never sees an empty
# "${arr[@]}").
checks=()
if [ -d "$DIR" ]; then
    for c in "$DIR"/*; do
        [ -e "$c" ] || continue                 # nullglob-safe (literal '*' when empty)
        case "$(basename "$c")" in README*|*.md) continue ;; esac
        [ -f "$c" ] && [ -x "$c" ] && checks+=("$c")
    done
fi

if [ "${#checks[@]}" -eq 0 ]; then
    echo "✓ merge-checks: none defined in scripts/merge-checks.d/ — pass (no-op)"
    exit 0
fi

failed=0
for c in "${checks[@]}"; do
    name="$(basename "$c")"
    if "$c" "$BASE" "$HEAD"; then
        echo "✓ merge-check passed: $name"
    else
        echo "✖ merge-check FAILED: $name" >&2
        failed=1
    fi
done

if [ "$failed" -ne 0 ]; then
    echo "✖ merge-checks: one or more checks failed (range ${BASE:0:7}..${HEAD:0:7})" >&2
    exit 1
fi
echo "✓ merge-checks: all ${#checks[@]} check(s) passed (range ${BASE:0:7}..${HEAD:0:7})"
exit 0
