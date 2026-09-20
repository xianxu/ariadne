#!/usr/bin/env bash
# Run ariadne's base-layer conformance tests as a merge check.
#
# WHY: these tests had NO automated runner at all. `scripts/parallel-checks.sh`
# is an LLM constitution-check runner and executes no bash tests, and
# portable-makefile.test.sh was referenced nowhere in the tree — it had only ever
# been run by hand. That is a large part of why ariadne#239's defect 2 survived
# #225: the test that would have caught it existed but nothing ran it.
#
# ARIADNE-LOCAL. Deliberately NOT a `symlink` row in base.manifest: these tests
# exercise ariadne's own sources (they `go build ./cmd/weave`), so running them
# in a derivative's CI would build the wrong tree. `scripts/merge-checks.d` is a
# `scaffold` row, so a derivative gets an empty dir and never sees this file.
#
# Usage: 50-base-layer-tests.sh <base_sha> <head_sha>   (merge-check contract)
set -uo pipefail

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

# Skip cleanly where the inputs do not exist — a derivative that somehow
# acquires this file, or a checkout without Go, exits 0 rather than dying.
if [ ! -d cmd/weave ]; then
    echo "✓ base-layer-tests: no cmd/weave here — not the base layer, skipping"
    exit 0
fi
if ! command -v go >/dev/null 2>&1; then
    echo "✓ base-layer-tests: no go toolchain — skipping"
    exit 0
fi

TESTS="
construct/scripts/test/portable-makefile.test.sh
construct/scripts/test/gitignore-surface.test.sh
construct/scripts/test/merge-checks.test.sh
"

failed=0
for t in $TESTS; do
    [ -f "$t" ] || { echo "✗ base-layer-tests: $t is missing"; failed=1; continue; }
    if out=$(bash "$t" 2>&1); then
        printf '✓ base-layer-test passed: %s\n' "$t"
    else
        printf '✗ base-layer-test FAILED: %s\n' "$t"
        printf '%s\n' "$out" | tail -20
        failed=1
    fi
done

[ "$failed" -eq 0 ] || exit 1
echo "✓ base-layer-tests: all passed"
