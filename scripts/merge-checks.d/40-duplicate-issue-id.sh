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

# Two kinds of "not running", and they are decidable apart (#213 BR-28):
#
#   exit 0  NOTHING TO CHECK — this repo has no issue tracker, or no published
#           id space to check against. The check is inapplicable, and clean is
#           the honest answer.
#   exit 2  COULD NOT CHECK — there IS something to check and the checker could
#           not be built or the trunk could not be read. Reporting 0 here
#           suppresses a real refusal, which is the whole failure mode this
#           issue exists to close: a green result from a check that never looked.
#
# Every gate is evaluated BEFORE any side effect, so a repo that cannot run this
# decides cleanly rather than dying in setup.

# ── nothing to check ────────────────────────────────────────────────────────
[ -d workshop/issues ] || exit 0
if ! git remote get-url origin >/dev/null 2>&1; then
    echo "40-duplicate-issue-id: no origin remote — no published id space to check against, skipping" >&2
    exit 0
fi

# ── could not check ─────────────────────────────────────────────────────────
cannot_check() {
    echo "40-duplicate-issue-id: COULD NOT RUN — $1" >&2
    echo "  exiting 2: this repo has issues to check, so clean would be a false pass (#213)." >&2
    exit 2
}

command -v go >/dev/null 2>&1 || cannot_check "go unavailable, cannot build the checker"

# Resolve sdlc the way the rest of the workflow does: build in its OWNER repo.
# A derivative has no ./cmd/sdlc of its own — it consumes the base layer's — so
# keying on a local ./cmd/sdlc made this check silently no-op in exactly the
# repos that carry collisions (#213 BR-6).
owner=""
if [ -x construct/dev-aliases.sh ]; then
    owner="$(construct/dev-aliases.sh --list 2>/dev/null | awk -F'\t' '$1=="sdlc"{print $2; exit}')"
fi
[ -n "$owner" ] || owner="$(pwd)"
[ -d "$owner/cmd/sdlc" ] || cannot_check "cannot locate the sdlc module (owner='$owner')"

# The published id space is the trunk TIP.
#
# Fetch INTO the remote-tracking ref explicitly (#213 BR-16). `git fetch origin
# main` updates FETCH_HEAD and only INCIDENTALLY refs/remotes/origin/main — in a
# CI checkout with no configured refspec it does not create it at all. The plain
# form therefore left $trunk empty on exactly the runners this check exists for,
# and it then fell back to the merge-base baseline that BR-1 proved structurally
# blind to this collision.
git fetch --quiet origin '+refs/heads/main:refs/remotes/origin/main' 2>/dev/null || true
trunk="$(git rev-parse --verify --quiet origin/main || true)"
[ -n "$trunk" ] || cannot_check "origin/main unreadable — the published id space cannot be seen"

# Explicit template: macOS `mktemp -d` with no template IGNORES $TMPDIR and uses
# a confstr path (/var/folders/...), so an environment that restricts that path
# made this check silently skip even with a perfectly writable TMPDIR set.
tmp="$(mktemp -d "${TMPDIR:-/tmp}/sdlc-idcheck.XXXXXX")" || cannot_check "no writable temp dir"
trap 'rm -rf "$tmp"' EXIT
( cd "$owner" && go build -o "$tmp/sdlc" ./cmd/sdlc ) 2>/dev/null || cannot_check "sdlc build failed in $owner"

# BOTH refs matter, and for different reasons:
#   --base  the merge-base the runner hands us — how a MOVE is recognised, so an
#           archive or renumber is not mistaken for a new claimant
#   --trunk the published tip — what this branch will actually merge INTO, and
#           the only place the colliding file exists
#
# lint-ids exits 0 clean, 1 collisions introduced, 2 could-not-run; all three
# propagate, so CI never sees green from a check that did not look.
if [ -n "$fallback_base" ]; then
    "$tmp/sdlc" issue lint-ids --base "$fallback_base" --trunk "$trunk" --head "$head"
else
    "$tmp/sdlc" issue lint-ids --base "$trunk" --head "$head"
fi
