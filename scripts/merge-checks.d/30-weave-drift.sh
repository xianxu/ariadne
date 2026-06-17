#!/usr/bin/env bash
# 30-weave-drift — dynamic-skill generated files must match `weave compile` (#111).
#
# A skill whose source is generated (committed codegen via a `.dynamic-skill`
# marker) goes stale the moment its inputs change without a regenerate+commit.
# This gate regenerates and fails if the committed generated SKILL.md drifts —
# delivering the "drift guard runs in CI" promise of #111's spec (the make target
# alone is human-only). Reuses `make weave-drift-check` so the drift logic has one
# source of truth.
#
# Contract: <check> <base_sha> <head_sha>, exit 0 = pass. The range args are
# unused — drift is a whole-tree property (the committed generated file vs a fresh
# regeneration), not scoped to the PR diff.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

# Only meaningful where the drift target exists (ariadne owns the datatype dynamic
# skill; derivatives have no dynamic skills of their own → no-op pass).
if ! make -n weave-drift-check >/dev/null 2>&1; then
    exit 0
fi

make weave-drift-check
