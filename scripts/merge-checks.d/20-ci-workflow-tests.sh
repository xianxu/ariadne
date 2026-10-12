#!/usr/bin/env bash
# 20-ci-workflow-tests — run the tests that execute the seeded CI workflows'
# real run blocks (#309): the local-gates guard and the merge-check shim.
# Ariadne-local: the workflows and their tests live here; consumers receive
# the workflows through weave, not these tests.
#
# Contract: <check> <base_sha> <head_sha>, exit 0 = pass, findings to stderr.
# The range is unused: the tests are cheap and cover the workflows whole.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
for t in scripts/test/local-gates-guard.test.sh scripts/test/portable-ci.test.sh; do
    bash "$t" >&2
done
