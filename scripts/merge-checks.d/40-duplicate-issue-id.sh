#!/usr/bin/env bash
# 40-duplicate-issue-id — refuse a PR that reuses an issue id (#213).
#
# Why on top of the gate inside `sdlc merge`: that one is operator feedback, not
# enforcement. It is skipped by merging in the GitHub UI, by a bare `gh pr
# merge`, by `--no-validate`, and by any actor who has not pulled the fix. An id
# space shared across machines and repos needs a server-side check.
#
# The bug: `sdlc issue new` on a branch cut before an issue landed on main used
# to allocate that published id. The two files then get different slugs —
# different PATHS — so git merges both cleanly and nothing else objects.
#
# TRUNK TIP, NOT MERGE-BASE. The runner contract hands us merge-base(base,head),
# and comparing against it CANNOT see this collision: the branch was cut before
# the colliding id was published, so the merge-base predates that file too and
# the id looks new on both sides. Measured — merge-base said "no reused ids" on
# the exact reproduction the issue documents. So this resolves the trunk tip
# itself and uses the runner's base only as a fallback.
#
# Logic lives in `sdlc issue lint-ids`, not here (ARCH-DRY).
#
# Contract: <check> <base_sha> <head_sha>, exit 0 = pass, findings to stderr.
set -euo pipefail

fallback_base="${1:-}"
head="${2:-HEAD}"

cd "$(git rev-parse --show-toplevel)"

# Every skip condition BEFORE any side effect, so a repo that cannot run this
# exits cleanly rather than dying in setup.
[ -d workshop/issues ] || exit 0
command -v go >/dev/null 2>&1 || {
    echo "40-duplicate-issue-id: go unavailable — SKIPPING the id-collision check" >&2
    exit 0
}

# Resolve sdlc the way the rest of the workflow does: build in its OWNER repo.
# A derivative has no ./cmd/sdlc of its own — it consumes the base layer's —
# so keying on a local ./cmd/sdlc made this check silently no-op in exactly the
# repos that carry collisions (#213 close review BR-6).
owner=""
if [ -x construct/dev-aliases.sh ]; then
    owner="$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$1=="sdlc"{print $2; exit}')"
fi
[ -n "$owner" ] || owner="$(pwd)"
if [ ! -d "$owner/cmd/sdlc" ]; then
    echo "40-duplicate-issue-id: cannot locate the sdlc module (owner='$owner') — SKIPPING" >&2
    exit 0
fi

# Explicit template: macOS `mktemp -d` with no template IGNORES $TMPDIR and uses
# a confstr path (/var/folders/...), so an environment that restricts that path
# made this check silently skip even with a perfectly writable TMPDIR set.
tmp="$(mktemp -d "${TMPDIR:-/tmp}/sdlc-idcheck.XXXXXX")" || {
    echo "40-duplicate-issue-id: no writable temp dir — skipping" >&2; exit 0; }
trap 'rm -rf "$tmp"' EXIT
if ! ( cd "$owner" && go build -o "$tmp/sdlc" ./cmd/sdlc ) 2>/dev/null; then
    echo "40-duplicate-issue-id: sdlc build failed in $owner — SKIPPING" >&2
    exit 0
fi

# The published id space is the trunk TIP. Fetch so a CI checkout has the ref at
# all; fall back to the runner's base only if it cannot be resolved.
git fetch --quiet origin main 2>/dev/null || true
trunk="$(git rev-parse --verify --quiet origin/main || true)"
if [ -z "$trunk" ]; then
    echo "40-duplicate-issue-id: no origin/main — comparing against the range base only" >&2
fi

# BOTH refs matter, and for different reasons:
#   --base  the merge-base the runner hands us — how a MOVE is recognised, so an
#           archive or renumber is not mistaken for a new claimant
#   --trunk the published tip — what this branch will actually merge INTO, and
#           the only place the colliding file exists
if [ -n "$fallback_base" ] && [ -n "$trunk" ]; then
    "$tmp/sdlc" issue lint-ids --base "$fallback_base" --trunk "$trunk" --head "$head"
elif [ -n "$trunk" ]; then
    "$tmp/sdlc" issue lint-ids --base "$trunk" --head "$head"
elif [ -n "$fallback_base" ]; then
    "$tmp/sdlc" issue lint-ids --base "$fallback_base" --head "$head"
else
    "$tmp/sdlc" issue lint-ids --head "$head"
fi
