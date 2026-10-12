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

# ── owner-declared preparation step (#299) ───────────────────────────────────
# prepd declares `dev-prepare:` in Makefile.local (where the base layer's Makefile
# looks for local rules); nous declares nothing and keeps today's exact form.
mkdir -p "$WS/prepd/construct"; mkcmd "$WS/prepd/cmd/gen"
printf '.PHONY: dev-prepare\ndev-prepare: assets/stamp\n' > "$WS/prepd/Makefile.local"
prep_out="$(bash "$GEN" --workspace "$WS" 2>/dev/null)"
gen_line="$(printf '%s\n' "$prep_out" | grep '^gen()')"
assert_contains "cd $WS/prepd && make -s dev-prepare >&2 && mkdir -p bin && rm -f bin/gen && go build -o bin/gen ./cmd/gen" "$gen_line" "a declaring owner runs dev-prepare before go build"
nous_line="$(printf '%s\n' "$prep_out" | grep '^nous()')"
[ "$nous_line" = "nous() { ( cd $WS/nous && mkdir -p bin && rm -f bin/nous && go build -o bin/nous ./cmd/nous ) || return; $WS/nous/bin/nous \"\$@\"; }" ] \
	&& ok "an owner without dev-prepare keeps the exact plain form" || ko "plain form changed: [$nous_line]"
# A .PHONY mention alone is not a declaration.
printf '.PHONY: dev-prepare\n' > "$WS/prepd/Makefile.local"
assert_not_contains "make -s dev-prepare" "$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^gen()')" "a .PHONY mention alone does not declare the step"
printf 'dev-prepare:\n\t@true\n' > "$WS/prepd/Makefile"; rm -f "$WS/prepd/Makefile.local"
assert_contains "make -s dev-prepare" "$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^gen()')" "a dev-prepare rule in Makefile also declares it"
printf 'dev-prepare := x\n' > "$WS/prepd/Makefile.local"; rm -f "$WS/prepd/Makefile"
assert_not_contains "make -s dev-prepare" "$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^gen()')" "a dev-prepare variable assignment is not a declaration"
printf 'assets dev-prepare : gen.stamp\n' > "$WS/prepd/Makefile.local"
assert_contains "make -s dev-prepare" "$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^gen()')" "a multi-target rule naming dev-prepare declares it"

# Run the emitted function against stub make/go: preparation runs first, its
# stdout stays off the binary's stdout, and a failed step fails the call loudly
# without building or running.
STUB="$WS/stubbin"; mkdir -p "$STUB"
cat > "$STUB/make" <<'STUBEOF'
#!/bin/sh
echo "make $*" >> "$TRACE"; echo "prep-chatter"
[ -z "${MAKE_FAIL:-}" ] || { echo "prep failed" >&2; exit 2; }
STUBEOF
cat > "$STUB/go" <<'STUBEOF'
#!/bin/sh
echo "go $*" >> "$TRACE"; printf '#!/bin/sh\necho ran "$@"\n' > "$3"; chmod +x "$3"
STUBEOF
chmod +x "$STUB/make" "$STUB/go"
gen_def="$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^gen()')"
export TRACE="$WS/trace"; : > "$TRACE"
run_out="$(PATH="$STUB:$PATH" bash -c "$gen_def; gen a b" 2>/dev/null)"
[ "$run_out" = "ran a b" ] && ok "the binary's stdout carries only its own output" || ko "stdout polluted: [$run_out]"
[ "$(cat "$TRACE")" = "$(printf 'make -s dev-prepare\ngo build -o bin/gen ./cmd/gen')" ] && ok "dev-prepare runs before go build" || ko "order: [$(cat "$TRACE")]"
: > "$TRACE"; rm -f "$WS/prepd/bin/gen"
fail_err="$(MAKE_FAIL=1 PATH="$STUB:$PATH" bash -c "$gen_def; gen a" 2>&1 >/dev/null)"; fail_rc=$?
[ "$fail_rc" -ne 0 ] && ok "a failed dev-prepare fails the call" || ko "failed dev-prepare returned 0"
assert_contains "prep failed" "$fail_err" "a failed dev-prepare is loud on stderr"
[ "$(cat "$TRACE")" = "make -s dev-prepare" ] && ok "a failed dev-prepare skips the build" || ko "built after failed prep: [$(cat "$TRACE")]"
unset TRACE
rm -rf "$WS/prepd" "$STUB" "$WS/trace"

# ── --list ───────────────────────────────────────────────────────────────────
list_out="$(bash "$GEN" --workspace "$WS" --list 2>/dev/null)"
assert_contains "$(printf 'gmail\t%s/nous' "$WS")" "$list_out" "--list maps gmail → nous"
assert_not_contains "() {" "$list_out" "--list emits no function defs"

# ── build-in-owner resolver (#60): the exact extraction `sdlc-build` relies on ─
# sdlc's owner must resolve to ariadne by LOCATION regardless of the caller (a
# derivative is a transitive consumer; --list scans workspace siblings, so depth
# is irrelevant). This is the contract Makefile.workflow:sdlc-build builds on.
assert_contains "$(printf 'sdlc\t%s/ariadne' "$WS")" "$list_out" "--list maps sdlc → ariadne (owner)"
owner="$(printf '%s\n' "$list_out" | awk -F'\t' '$1=="sdlc"{print $2}')"
[ "$owner" = "$WS/ariadne" ] && ok "sdlc-build owner-resolution: awk extracts the owner path" || ko "owner-resolution: got [$owner]"

# The PRODUCTION path sdlc-build actually takes: invoked with NO --workspace,
# as a relative symlink, FROM a derivative. The default workspace must resolve
# from BASH_SOURCE so sdlc still maps to ariadne (this is the path #60 M4 makes
# load-bearing once construct/go.mod's build branch is gone).
cp "$GEN" "$WS/ariadne/construct/dev-aliases.sh"
mkdir -p "$WS/deriv/construct"
( cd "$WS/deriv/construct" && ln -s ../../ariadne/construct/dev-aliases.sh dev-aliases.sh )
def_out="$( cd "$WS/deriv" && bash construct/dev-aliases.sh --list 2>/dev/null )"
def_owner="$(printf '%s\n' "$def_out" | awk -F'\t' '$1=="sdlc"{print $2}')"
# Default workspace is derived via `pwd -P` (physical), so compare against the
# physical $WS (macOS /tmp → /private/tmp); same inode, different spelling.
WS_P="$(cd "$WS" && pwd -P)"
[ "$def_owner" = "$WS_P/ariadne" ] && ok "default-workspace via derivative symlink: sdlc → ariadne" || ko "default-workspace symlink: got [$def_owner] want [$WS_P/ariadne]"
rm -rf "$WS/deriv" "$WS/ariadne/construct/dev-aliases.sh"

# ── non-peer substrate-following (#95) ───────────────────────────────────────
# A repo created OUTSIDE ariadne's workspace (NOT a sibling) links ariadne via
# construct/deps `substrate <abs ariadne>`. The sibling scan alone can't see
# ariadne (it's not under the non-peer repo's parent); dev-aliases must ALSO
# follow the current repo's substrate graph (via the real lib-deps.sh, reached
# through the dev-aliases.sh symlink) to resolve sdlc → ariadne. This is the bug
# the non-peer bootstrap test uncovered.
NP="$(mktemp -d "${TMPDIR:-/tmp}/devaliases-np.XXXXXX")"
# Stage a self-contained ariadne owner with the real script + lib-deps parser.
mkdir -p "$NP/ariadne/construct/scripts"; mkcmd "$NP/ariadne/cmd/sdlc"
cp "$GEN" "$NP/ariadne/construct/dev-aliases.sh"
cp "$ARIADNE_ROOT/construct/scripts/lib-deps.sh" "$NP/ariadne/construct/scripts/lib-deps.sh"
# The non-peer repo lives in a SEPARATE parent (so ariadne is NOT a sibling),
# with an absolute substrate edge to the staged ariadne + the symlinked script.
mkdir -p "$NP/elsewhere/another/construct"
ARI_P="$(cd "$NP/ariadne" && pwd -P)"
printf 'substrate %s\n' "$ARI_P" > "$NP/elsewhere/another/construct/deps"
( cd "$NP/elsewhere/another/construct" && ln -s "$ARI_P/construct/dev-aliases.sh" dev-aliases.sh )
np_out="$( cd "$NP/elsewhere/another" && bash construct/dev-aliases.sh --list 2>/dev/null )"
np_owner="$(printf '%s\n' "$np_out" | awk -F'\t' '$1=="sdlc"{print $2}')"
[ "$np_owner" = "$ARI_P" ] && ok "non-peer substrate: sdlc → ariadne (via construct/deps)" || ko "non-peer substrate: got [$np_owner] want [$ARI_P]"
# The non-peer repo's OWN parent has no ariadne sibling — proves resolution came
# from the substrate graph, not an accidental sibling match.
assert_contains "$(printf 'sdlc\t%s' "$ARI_P")" "$np_out" "non-peer --list maps sdlc to the substrate owner"
rm -rf "$NP"

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
