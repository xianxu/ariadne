#!/usr/bin/env bash
# 40-duplicate-issue-id — refuse a PR that reuses an issue id (#213).
#
# Why this exists ON TOP OF the gate inside `sdlc merge`: that one is operator
# feedback, not enforcement. It is skipped by merging in the GitHub UI, by a bare
# `gh pr merge`, by `--no-validate`, and by any actor on a machine that has not
# pulled the fix. An id space shared across machines and repos needs a
# server-side check, and this one can be a required status check.
#
# The bug it guards: `sdlc issue new` on a branch cut before an issue landed on
# main used to allocate that published id, because the branch's workshop/issues/
# never contained it. The two files then get different slugs — different PATHS —
# so git merges both cleanly and nothing else in the lifecycle objects. Eight
# such collisions existed across two repos when this was written.
#
# The logic lives in `sdlc issue lint-ids`, not here (ARCH-DRY): filename
# parsing, the three id-bearing directories, and the introduced-vs-pre-existing
# distinction are decided once, in Go, with tests. This script is the CI adapter.
#
# Contract: <check> <base_sha> <head_sha>, exit 0 = pass, findings to stderr.
set -euo pipefail

base="${1:-}"
head="${2:-HEAD}"

cd "$(git rev-parse --show-toplevel)"

# No tracker, nothing to check — a derivative without workshop/issues/ is a
# no-op pass, matching the runner's "no checks defined" posture.
[ -d workshop/issues ] || exit 0

# Every reason to skip is checked BEFORE any side effect, so a repo that cannot
# run this check exits cleanly rather than dying inside setup. (First cut created
# the temp dir first and failed hard where mktemp was restricted — the skip paths
# have to come first to be reachable at all.)
#
# A derivative repo has no ./cmd/sdlc of its own; it consumes the base layer's
# binary. Nothing to build here, nothing to check.
if [ ! -d ./cmd/sdlc ]; then
    echo "40-duplicate-issue-id: no ./cmd/sdlc in this repo — skipping" >&2
    exit 0
fi
if ! command -v go >/dev/null 2>&1; then
    echo "40-duplicate-issue-id: go unavailable — SKIPPING the id-collision check" >&2
    exit 0
fi

# Build from THIS checkout rather than trusting a PATH binary: CI has none
# installed, and a stale one would check with the old logic.
tmp="$(mktemp -d)" || { echo "40-duplicate-issue-id: no writable temp dir — skipping" >&2; exit 0; }
trap 'rm -rf "$tmp"' EXIT
if ! go build -o "$tmp/sdlc" ./cmd/sdlc 2>/dev/null; then
    echo "40-duplicate-issue-id: sdlc build failed — SKIPPING the id-collision check" >&2
    exit 0
fi

if [ -n "$base" ]; then
    "$tmp/sdlc" issue lint-ids --base "$base" --head "$head"
else
    "$tmp/sdlc" issue lint-ids --head "$head"
fi
