#!/usr/bin/env bash
# parallel-checks.test.sh — the audit path of scripts/parallel-checks.sh, run with
# a stub pre-merge-checks.sh (the real checks launch LLM agents). Covers #351
# (temp dir from TMPDIR; the pids rebuild on an empty array under bash 3.2's
# set -u) and #353 (MAX_PARALLEL_CHECKS must be a positive integer).
# Run: bash scripts/test/parallel-checks.test.sh   (exit 0 = all pass)
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../.." && pwd)"
WORK="$(mktemp -d "${TMPDIR:-/tmp}/parallel-checks-test.XXXXXX")" || exit 1
trap 'rm -rf "${WORK:?}"' EXIT

# A scratch repository with one change against origin/main, the real script and
# its library, and a stub that answers every check instantly.
REPO="$WORK/repo"
mkdir -p "$REPO/scripts"
cp "$SOURCE/scripts/parallel-checks.sh" "$SOURCE/scripts/lib.sh" "$REPO/scripts/"
printf '#!/bin/sh\necho "stub $1"\n' > "$REPO/scripts/pre-merge-checks.sh"
chmod +x "$REPO/scripts/pre-merge-checks.sh"
git -C "$REPO" init -q -b main
git -C "$REPO" -c user.name=t -c user.email=t@t commit -q --allow-empty -m base
git -C "$REPO" update-ref refs/remotes/origin/main HEAD
echo change > "$REPO/file"
git -C "$REPO" add file
git -C "$REPO" -c user.name=t -c user.email=t@t commit -q -m change

pass=0 fail=0
# audit <want-exit> <want-text> [VAR=value ...]: run the audit with the given
# environment; the exit code must match and the output must contain the text.
audit() {
    local want="$1" text="$2" out rc=0
    shift 2
    # Bounded: a regression that spins (MAX_PARALLEL_CHECKS=0 once did) fails
    # here instead of hanging the suite. perl alarm: macOS has no timeout(1).
    out="$(cd "${REPO:?}" && perl -e 'alarm shift; exec @ARGV' 30 env "$@" bash scripts/parallel-checks.sh --audit 2>&1)" || rc=$?
    if [[ "$rc" -eq "$want" && "$out" == *"$text"* && "$out" != *"unbound variable"* ]]; then
        pass=$((pass + 1))
    else
        fail=$((fail + 1))
        printf 'FAIL: %s -> exit %d, want %d and %q\n%s\n' "$*" "$rc" "$want" "$text" "$out" >&2
    fi
}

audit 0 "Constitution checks" MAX_PARALLEL_CHECKS=3
# One slot: every launch waits and rebuilds pids from finished checks, the empty
# "${alive[@]}" that bash 3.2 rejected under set -u (#351).
audit 0 "Constitution checks" MAX_PARALLEL_CHECKS=1
audit 1 "must be a positive integer, got '0'" MAX_PARALLEL_CHECKS=0
audit 1 "must be a positive integer, got 'abc'" MAX_PARALLEL_CHECKS=abc
audit 1 "must be a positive integer, got '-2'" MAX_PARALLEL_CHECKS=-2
audit 1 "Cannot make a temp dir under $WORK/absent" TMPDIR="$WORK/absent"

printf 'parallel-checks: %d passed, %d failed\n' "$pass" "$fail"
[[ "$fail" -eq 0 ]]
