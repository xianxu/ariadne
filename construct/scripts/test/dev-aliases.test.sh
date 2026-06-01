#!/usr/bin/env bash
# dev-aliases.test.sh — hermetic tests for construct/dev-aliases.sh (#57 M1).
# Builds a throwaway sibling-workspace under $TMPDIR (no network, no go build)
# and asserts which functions get emitted, the function-body shape, ownership
# (owner wins over re-export symlink + legacy copy), --list, and dup/--strict.
#
# Run:  bash construct/scripts/test/dev-aliases.test.sh
# Exit: 0 if all assertions pass, 1 otherwise.
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
GEN="$ARIADNE_ROOT/construct/dev-aliases.sh"

WS="$(mktemp -d "${TMPDIR:-/tmp}/devaliases57.XXXXXX")"
trap 'rm -rf "$WS"' EXIT

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }
assert_contains()     { case "$2" in *"$1"*) ok "$3" ;; *) ko "$3 — missing: $1";; esac; }
assert_not_contains() { case "$2" in *"$1"*) ko "$3 — unexpected: $1" ;; *) ok "$3";; esac; }

mkcmd() { mkdir -p "$1"; printf 'package main\nfunc main() {}\n' > "$1/main.go"; }

# ── fixture workspace ────────────────────────────────────────────────────────
# ariadne (owner of sdlc)
mkdir -p "$WS/ariadne/construct"; mkcmd "$WS/ariadne/cmd/sdlc"
# nous (owner of nous, gmail; a .private tool; an empty cmd dir)
mkdir -p "$WS/nous/construct"
mkcmd "$WS/nous/cmd/nous"; mkcmd "$WS/nous/cmd/gmail"
mkcmd "$WS/nous/cmd/secret"; : > "$WS/nous/cmd/secret/.private"
mkdir -p "$WS/nous/cmd/emptytool/bin"        # no .go → not buildable
# brain (derivative): re-export symlink to nous's gmail (must NOT own gmail)
mkdir -p "$WS/brain/construct/" "$WS/brain/cmd"
( cd "$WS/brain/cmd" && ln -s ../../nous/cmd/gmail gmail )
# brain.legacy (dead clone with a REAL gmail copy → must be skipped, no dup)
mkdir -p "$WS/brain.legacy/construct"; mkcmd "$WS/brain.legacy/cmd/gmail"
# a non-ariadne dir (no construct/) → ignored entirely
mkdir -p "$WS/random/cmd/tool"; mkcmd "$WS/random/cmd/tool"

out="$(bash "$GEN" --workspace "$WS" 2>/dev/null)"

# ── core exposure ────────────────────────────────────────────────────────────
assert_contains "sdlc() {"  "$out" "exposes ariadne-owned sdlc"
assert_contains "nous() {"  "$out" "exposes nous-owned nous"
assert_contains "gmail() {" "$out" "exposes nous-owned gmail"
assert_not_contains "secret()"    "$out" ".private tool is not exposed"
assert_not_contains "emptytool()" "$out" "non-buildable cmd dir is not exposed"
assert_not_contains "tool()"      "$out" "non-ariadne repo (no construct/) is ignored"

# ── ownership: gmail builds in nous, not brain (symlink) or brain.legacy ──────
gmail_line="$(printf '%s\n' "$out" | grep '^gmail()')"
assert_contains "cd $WS/nous " "$gmail_line" "gmail builds in the owner (nous)"
assert_not_contains "brain"    "$gmail_line" "gmail does not build in a derivative/legacy"
# exactly one gmail definition (symlink + legacy copy didn't add duplicates)
n_gmail="$(printf '%s\n' "$out" | grep -c '^gmail()')"
[ "$n_gmail" = "1" ] && ok "exactly one gmail() emitted" || ko "expected 1 gmail(), got $n_gmail"

# ── function-body shape (the defining property; fixture can't run go build) ───
assert_contains "cd $WS/ariadne && mkdir -p bin && rm -f bin/sdlc && go build -o bin/sdlc ./cmd/sdlc" "$out" "builds to owner bin/ (rm -f first), owner cmd path"
assert_contains "$WS/ariadne/bin/sdlc \"\$@\""                                                        "$out" "runs the owner-bin binary with caller args/cwd"

# ── --list ───────────────────────────────────────────────────────────────────
list_out="$(bash "$GEN" --workspace "$WS" --list 2>/dev/null)"
assert_contains "$(printf 'gmail\t%s/nous' "$WS")" "$list_out" "--list maps gmail → nous"
assert_not_contains "() {" "$list_out" "--list emits no function defs"

# ── duplicate detection + --strict ───────────────────────────────────────────
# two ACTIVE repos both owning a real cmd/dupe → collision
mkdir -p "$WS/repoa/construct" "$WS/repob/construct"
mkcmd "$WS/repoa/cmd/dupe"; mkcmd "$WS/repob/cmd/dupe"
dup_err="$(bash "$GEN" --workspace "$WS" 2>&1 >/dev/null)"
assert_contains "duplicate binary 'dupe'" "$dup_err" "warns on duplicate binary name"
# last-win: repob sorts after repoa
dup_line="$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^dupe()')"
assert_contains "cd $WS/repob " "$dup_line" "duplicate resolves last-win (repob)"
# --strict exits non-zero
if bash "$GEN" --workspace "$WS" --strict >/dev/null 2>&1; then
	ko "--strict should exit non-zero on duplicate"
else
	ok "--strict exits non-zero on duplicate"
fi
# without the dup, --strict is clean (remove repob's dupe)
rm -rf "$WS/repob" "$WS/repoa"
if bash "$GEN" --workspace "$WS" --strict >/dev/null 2>&1; then
	ok "--strict exits zero when no duplicates"
else
	ko "--strict should exit zero without duplicates"
fi

# ── workspace validation (M1 review minors) ─────────────────────────────────
if bash "$GEN" --workspace "$WS/does-not-exist" >/dev/null 2>&1; then
	ko "nonexistent --workspace should exit non-zero"
else
	ok "nonexistent --workspace exits non-zero"
fi
if bash "$GEN" --workspace >/dev/null 2>&1; then
	ko "--workspace with no value should exit non-zero (not fall back to default)"
else
	ok "--workspace with no value exits non-zero"
fi

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
