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

WS="$(mktemp -d "${TMPDIR:-/tmp}/devaliases57.XXXXXX")" || exit 1
WS="$(cd "${WS:?}" && pwd -P)"
trap 'rm -rf "${WS:?}"' EXIT
cd "$WS" || exit 1

# A Couch slot's COUCH_TREE skips every build (#405); the cases that need one
# set it themselves.
unset COUCH_TREE

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }
assert_contains()     { case "$2" in *"$1"*) ok "$3" ;; *) ko "$3 — missing: $1";; esac; }
assert_not_contains() { case "$2" in *"$1"*) ko "$3 — unexpected: $1" ;; *) ok "$3";; esac; }

# A diagnostic flag may add stderr, but must preserve machine-readable --list
# stdout and status, in both ordinary and strict modes.
assert_verbose_list_unchanged() {
	local label="$1" cwd="$2" strict="$3" expected="$4" strict_rc="$5" ordinary_rc verbose_rc
	# Strictness changes collision status, not the already checked list mapping.
	# Reuse that ordinary mapping while still checking both verbose modes.
	if [ -z "$strict" ]; then
		( cd "$cwd" && bash "$GEN" --workspace "$WS" --list > "$WS/ordinary.list" 2> "$WS/ordinary.err" ); ordinary_rc=$?
	else
		ordinary_rc=$strict_rc
	fi
	( cd "$cwd" && bash "$GEN" --workspace "$WS" --list ${strict:+--strict} --verbose > "$WS/verbose.list" 2> "$WS/verbose.err" ); verbose_rc=$?
	if [ "$ordinary_rc" -eq "$verbose_rc" ] && cmp -s "$WS/ordinary.list" "$WS/verbose.list"; then
		ok "$label preserves --list stdout and status"
	else
		ko "$label changed --list stdout or status"
	fi
	assert_contains "$expected" "$(cat "$WS/verbose.err")" "$label reports diagnostics on stderr"
}

mkcmd() { mkdir -p "$1"; printf 'package main\nfunc main() {}\n' > "$1/main.go"; }
# Identity cases consume Git config, not the remote-management CLI. Keep real
# initialized repositories and real Git reads; write their tiny fixture data.
set_origin() {
    printf '[core]\nrepositoryformatversion = 0\nbare = false\n' > "$1/.git/config"
    [ -z "$2" ] || printf '[remote "origin"]\nurl = %s\n' "$2" >> "$1/.git/config"
}

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
SKIP_SDLC='[ -x bin/sdlc ] && { [ -n "${COUCH_TREE:-}" ] || [ -z "$(find . \( -path ./.git -o -path ./bin \) -prune -o -newer bin/sdlc -print -quit 2>/dev/null || echo stale)" ]; }'
assert_contains 'cd '"$WS"'/ariadne && { '"$SKIP_SDLC"' || { mkdir -p bin' "$out" "skips the build in a Couch slot or when bin/ is fresh (#405)"
assert_contains '{ mkdir -p bin && d="$(mktemp -d bin/.sdlc.XXXXXX)" && : > "$d/.start" && { go build -trimpath -o "$d/sdlc" ./cmd/sdlc && touch -r "$d/.start" "$d/sdlc" && mv -f "$d/sdlc" bin/sdlc;' "$out" "builds beside owner bin/ and renames into place (#355), owner cmd path"
assert_contains "$WS/ariadne/bin/sdlc \"\$@\""                                                        "$out" "runs the owner-bin binary with caller args/cwd"

# Core exposure cases are complete; later scans need only selected owners.
mkdir "$WS/parked"
mv "$WS/ariadne" "$WS/brain" "$WS/brain.legacy" "$WS/random" "$WS/parked/"

# ── owner-declared preparation step (#299) ───────────────────────────────────
# prepd declares `dev-prepare:` in Makefile.local (where the base layer's
# Makefile looks for local rules); nous declares nothing and keeps today's
# exact form.
mkdir -p "$WS/prepd/construct"; mkcmd "$WS/prepd/cmd/gen"
# Like the base layer's Makefile, the owner's Makefile includes Makefile.local.
printf -- '-include Makefile.local\n' > "$WS/prepd/Makefile"
printf '.PHONY: dev-prepare\ndev-prepare: assets/stamp\nassets/stamp:\n' > "$WS/prepd/Makefile.local"
prep_out="$(bash "$GEN" --workspace "$WS" 2>/dev/null)"
gen_line="$(printf '%s\n' "$prep_out" | grep '^gen()')"
assert_contains ' || { make -s dev-prepare >&2 && mkdir -p bin && d="$(mktemp -d bin/.gen.XXXXXX)" && : > "$d/.start" && { go build -trimpath -o "$d/gen" ./cmd/gen' "$gen_line" "a declaring owner runs dev-prepare before go build"
nous_line="$(printf '%s\n' "$prep_out" | grep '^nous()')"
[ "$nous_line" = 'nous() { ( cd '"$WS"'/nous && { [ -x bin/nous ] && { [ -n "${COUCH_TREE:-}" ] || [ -z "$(find . \( -path ./.git -o -path ./bin \) -prune -o -newer bin/nous -print -quit 2>/dev/null || echo stale)" ]; } || { mkdir -p bin && d="$(mktemp -d bin/.nous.XXXXXX)" && : > "$d/.start" && { go build -trimpath -o "$d/nous" ./cmd/nous && touch -r "$d/.start" "$d/nous" && mv -f "$d/nous" bin/nous; s=$?; rm -rf "${d:?}"; exit $s; }; }; } ) || return; '"$WS"'/nous/bin/nous "$@"; }' ] \
	&& ok "an owner without dev-prepare keeps the exact plain form" || ko "plain form changed: [$nous_line]"
# Keep unrelated owner trees out of repeated preparation scans (#465).
# They return for the ownership/list cases; the real generator is unchanged.
mv "$WS/nous" "$WS/parked/"

# Make decides what declares the step: a rule (here, or via an include) does;
# an assignment, a target-specific variable or a recipe line that merely
# mentions the name does not. A rule with a missing prerequisite is declared
# but broken, so it is emitted and the call fails loudly.
# Independent declaration fixtures share one real generator scan. Their unique
# binaries preserve each parser case without repeating process-wide setup.
MATRIX="$WS/prep-matrix"; mkdir "$MATRIX"
prep_case() {
    local name="$1" contents="$2" expected="$3" label="$4" dir="$MATRIX/$1"
    mkdir -p "$dir/construct"; mkcmd "$dir/cmd/$name"
    printf -- '-include Makefile.local\n' > "$dir/Makefile"
    printf '%b' "$contents" > "$dir/Makefile.local"
    printf '%s\t%s\t%s\n' "$name" "$expected" "$label" >> "$MATRIX/cases"
}
prep_case multi 'assets dev-prepare : gen.stamp\ngen.stamp:\n' declared 'a multi-target rule naming dev-prepare'
prep_case missing 'dev-prepare: missing\n' declared 'a rule whose prerequisite has no rule (broken, fails at call)'
prep_case assignment 'dev-prepare := x\n' plain 'a := assignment'
prep_case specific 'dev-prepare: V = 1\n' plain 'a target-specific variable'
prep_case recipe 'all:\n\t@echo run dev-prepare: now\n' plain 'a recipe line mentioning the name'
prep_case workflow '' declared 'a rule in Makefile.workflow, reached through the include'
printf -- '-include Makefile.workflow\n-include Makefile.local\n' > "$MATRIX/workflow/Makefile"
printf 'dev-prepare:\n\t@true\n' > "$MATRIX/workflow/Makefile.workflow"
prep_case direct '' declared 'a dev-prepare rule in Makefile also declares it'
printf 'dev-prepare:\n\t@true\n' > "$MATRIX/direct/Makefile"
matrix_out="$(bash "$GEN" --workspace "$MATRIX" 2>/dev/null)"
while IFS=$'\t' read -r name expected label; do
    line="$(printf '%s\n' "$matrix_out" | grep "^$name()")"
    case "$expected:$line" in
        declared:*'make -s dev-prepare'*) ok "declared: $label" ;;
        plain:*'make -s dev-prepare'*|*:) ko "$expected: $label" ;;
        plain:*) ok "plain: $label" ;;
        *) ko "$expected: $label" ;;
    esac
done < "$MATRIX/cases"
rm -rf "$MATRIX"
printf 'dev-prepare:\n\t@true\n' > "$WS/prepd/Makefile"; rm -f "$WS/prepd/Makefile.local"

# Detection runs make once per owner, not once per binary: a counting make
# stub on PATH sees one call for an owner with two binaries.
mkcmd "$WS/prepd/cmd/gen2"
CNT="$WS/stubcount"; mkdir -p "$CNT"
printf '#!/bin/sh\necho x >> "%s/calls"\nexit 1\n' "$CNT" > "$CNT/make"; chmod +x "$CNT/make"
two_out="$(PATH="$CNT:$PATH" bash "$GEN" --workspace "$WS" 2>/dev/null)"
assert_contains " || { make -s dev-prepare" "$(printf '%s\n' "$two_out" | grep '^gen2()')" "the owner's second binary gets the cached answer"
[ "$(wc -l < "$CNT/calls" | tr -d ' ')" = "1" ] && ok "one make detection per owner" || ko "make detections: $(wc -l < "$CNT/calls")"
rm -rf "$CNT" "$WS/prepd/cmd/gen2"

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
echo "go $*" >> "$TRACE"
while [ "$#" -gt 0 ] && [ "$1" != -o ]; do shift; done
[ "$#" -ge 2 ] || exit 2
out="$2"
printf '#!/bin/sh\necho ran "$@"\n' > "$out"; chmod +x "$out"
STUBEOF
chmod +x "$STUB/make" "$STUB/go"
gen_def="$gen_line"
export TRACE="$WS/trace"; : > "$TRACE"
run_out="$(PATH="$STUB:$PATH" bash -c "$gen_def; gen a b" 2>/dev/null)"
[ "$run_out" = "ran a b" ] && ok "the binary's stdout carries only its own output" || ko "stdout polluted: [$run_out]"
case "$(cat "$TRACE")" in "make -s dev-prepare"$'\n'"go build -trimpath -o bin/.gen."??????"/gen ./cmd/gen") ok "dev-prepare runs before go build" ;; *) ko "order: [$(cat "$TRACE")]" ;; esac
: > "$TRACE"; rm -f "$WS/prepd/bin/gen"
fail_err="$(MAKE_FAIL=1 PATH="$STUB:$PATH" bash -c "$gen_def; gen a" 2>&1 >/dev/null)"; fail_rc=$?
[ "$fail_rc" -ne 0 ] && ok "a failed dev-prepare fails the call" || ko "failed dev-prepare returned 0"
assert_contains "prep failed" "$fail_err" "a failed dev-prepare is loud on stderr"
[ "$(cat "$TRACE")" = "make -s dev-prepare" ] && ok "a failed dev-prepare skips the build" || ko "built after failed prep: [$(cat "$TRACE")]"
unset TRACE
rm -rf "$WS/prepd" "$STUB" "$WS/trace"
mv "$WS/parked/ariadne" "$WS/parked/nous" "$WS/"

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

mv "$WS/nous" "$WS/parked/"

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
# A slot's substrate and the workspace sibling are two checkouts of one origin.
# The substrate is last in sorted path order and remains the selected owner.
SLOT="$WS/worktree/slot/deriv"
SUBSTRATE="$WS/worktree/slot/ariadne"
mkdir -p "$SLOT/construct" "$SUBSTRATE/construct"
mkcmd "$SUBSTRATE/cmd/sdlc"
printf 'substrate %s\n' "$SUBSTRATE" > "$SLOT/construct/deps"
git -C "$WS/ariadne" init -q
git -C "$SUBSTRATE" init -q
set_origin "$WS/ariadne" https://github.com/example/ariadne.git
set_origin "$SUBSTRATE" git@github.com:Example/Ariadne.git
slot_out="$(cd "$SLOT" && bash "$GEN" --workspace "$WS" --list --strict 2>"$WS/slot.err")"; slot_rc=$?
[ "$slot_rc" -eq 0 ] && ok "same-origin checkouts pass --strict" || ko "same-origin checkouts failed --strict"
[ ! -s "$WS/slot.err" ] && ok "same-origin checkouts produce no warning" || ko "same-origin warning: [$(cat "$WS/slot.err")]"
assert_contains "$(printf 'sdlc\t%s' "$SUBSTRATE")" "$slot_out" "slot substrate remains selected owner"
assert_verbose_list_unchanged "same-origin verbose" "$SLOT" "" "same-origin" 0
assert_verbose_list_unchanged "same-origin strict verbose" "$SLOT" "strict" "same-origin" 0
set_origin "$SUBSTRATE" ssh://git@github.com/example/ariadne.git
if (cd "$SLOT" && bash "$GEN" --workspace "$WS" --strict --list >"$WS/ssh-list" 2>"$WS/ssh.err"); then
	[ ! -s "$WS/ssh.err" ] && ok "GitHub ssh:// origin is equivalent and quiet under --strict" || ko "GitHub ssh:// warned: [$(cat "$WS/ssh.err")]"
else
	ko "GitHub ssh:// origin should pass --strict"
fi
rm -rf "$WS/worktree"
# The identity matrix needs only its two real Git owners.
mv "$WS/ariadne" "$WS/parked/"

# two ACTIVE repos both owning a real cmd/dupe → collision
mkdir -p "$WS/repoa/construct" "$WS/repob/construct"
mkcmd "$WS/repoa/cmd/dupe"; mkcmd "$WS/repob/cmd/dupe"
git -C "$WS/repoa" init -q; git -C "$WS/repob" init -q
set_origin "$WS/repoa" https://github.com/example/repoa.git
set_origin "$WS/repob" https://github.com/example/repob.git
bash "$GEN" --workspace "$WS" > "$WS/dupe.defs" 2> "$WS/dupe.err"
dup_err="$(cat "$WS/dupe.err")"
assert_contains "duplicate binary 'dupe'" "$dup_err" "warns on duplicate binary name"
assert_verbose_list_unchanged "distinct-origin verbose" "$WS" "" "duplicate binary 'dupe'" 1
assert_verbose_list_unchanged "distinct-origin strict verbose" "$WS" "strict" "duplicate binary 'dupe'" 1
# last-win: repob sorts after repoa
dup_line="$(grep '^dupe()' "$WS/dupe.defs")"
assert_contains "cd $WS/repob " "$dup_line" "duplicate resolves last-win (repob)"
# --strict exits non-zero
if bash "$GEN" --workspace "$WS" --strict >/dev/null 2>&1; then
	ko "--strict should exit non-zero on duplicate"
else
	ok "--strict exits non-zero on duplicate"
fi
# Missing origin cannot prove sameness, so it must not hide a collision.
set_origin "$WS/repoa" ""; set_origin "$WS/repob" ""
if bash "$GEN" --workspace "$WS" --strict >/dev/null 2>&1; then
	ko "unknown origins should not suppress a duplicate"
else
	ok "unknown origins keep --strict collision"
fi
same_origin_check() {
	local label="$1" err rc
	err="$(bash "$GEN" --workspace "$WS" --strict --list 2>&1 >/dev/null)"; rc=$?
	[ "$rc" -eq 0 ] && [ -z "$err" ] && ok "$label" || ko "$label: rc=$rc stderr=[$err]"
}
different_origin_check() {
	local label="$1" err rc
	err="$(bash "$GEN" --workspace "$WS" --strict --list 2>&1 >/dev/null)"; rc=$?
	if [ "$rc" -ne 0 ]; then
		assert_contains "duplicate binary 'dupe'" "$err" "$label"
	else
		ko "$label: rc=$rc stderr=[$err]"
	fi
}
set_origin "$WS/repoa" https://git.example.com/group/shared.git
set_origin "$WS/repob" https://git.example.com/group/shared.git
same_origin_check "identical non-GitHub origins are quiet under --strict"
set_origin "$WS/repob" git@git.example.com:group/shared.git
different_origin_check "non-GitHub transport difference remains a collision"
set_origin "$WS/repob" https://git.example.com/group/Shared.git
different_origin_check "non-GitHub path case difference remains a collision"
set_origin "$WS/repoa" ../shared.git
set_origin "$WS/repob" ../shared.git
different_origin_check "unrecognized relative origins remain distinct by checkout"
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

# #355: pause two real alias rebuilds after writing executable partial temps.
# Both must leave the old binary executable until their final rename (#465).
# Readiness FIFOs select that ordering without sleeps or a date/exec busy loop.
mv "$WS/parked/nous" "$WS/"
SLOW="$WS/slowbin"; mkdir -p "$SLOW"
cat > "$SLOW/go" <<'STUBEOF'
#!/bin/sh
while [ "$#" -gt 0 ] && [ "$1" != -o ]; do shift; done
[ "$#" -ge 2 ] || exit 2
out="$2"
printf '#!/bin/sh\n' > "$out"; chmod +x "$out"
if [ -n "${BUILD_READY:-}" ]; then
    mkfifo "$out.release" || exit 2
    printf '%s/%s\n' "$PWD" "$out" > "$BUILD_READY"
    read -r release < "$out.release" || exit 2
    rm "$out.release"
fi
printf 'echo built\n' >> "$out"
STUBEOF
chmod +x "$SLOW/go"
nous_def="$nous_line"
PATH="$SLOW:$PATH" bash -c "$nous_def; nous >/dev/null" || ko "the first build failed"
touch -t 203001010000 "$WS/nous/cmd/nous/main.go"
mkfifo "$WS/ready"
exec 3<> "$WS/ready"
BUILD_READY="$WS/ready" PATH="$SLOW:$PATH" bash -c "$nous_def; nous >/dev/null" & b1=$!
BUILD_READY="$WS/ready" PATH="$SLOW:$PATH" bash -c "$nous_def; nous >/dev/null" & b2=$!
if ! read -r -t 20 partial1 <&3 || ! read -r -t 20 partial2 <&3; then
    ko "concurrent builds did not reach partial-output barriers"
    kill "$b1" "$b2" 2>/dev/null || true
    wait "$b1" "$b2" 2>/dev/null || true
    exit 1
fi
MISS="$WS/misses"; : > "$MISS"
# Check while both partial executables exist, before allowing either rename.
got="$("$WS/nous/bin/nous" 2>>"$MISS")" || echo "exec failed" >> "$MISS"
[ "$got" = built ] || echo "partial binary: [$got]" >> "$MISS"
printf 'release\n' > "$partial1.release"
printf 'release\n' > "$partial2.release"
wait $b1; r1=$?; wait $b2; r2=$?
exec 3>&-
[ $r1 -eq 0 ] && [ $r2 -eq 0 ] && ok "two concurrent rebuilds both succeed" || ko "concurrent rebuilds failed: $r1 $r2"
[ ! -s "$MISS" ] && ok "the binary is never missing during a rebuild" || ko "exec saw a missing or partial binary: $(sort "$MISS" | uniq -c | head -3)"
left="$(ls -A "$WS/nous/bin" | grep -v '^nous$')"
[ -z "$left" ] && ok "no build directory is left behind" || ko "left in bin/: $left"
[ "$("$WS/nous/bin/nous")" = "built" ] && ok "the last rename installed a whole binary" || ko "final binary is not whole"

# A failed build fails the call without running anything, removes its build
# directory, and leaves the previous binary in place (rm -f used to destroy it).
FAILGO="$WS/failbin"; mkdir -p "$FAILGO"
cat > "$FAILGO/go" <<'STUBEOF'
#!/bin/sh
while [ "$#" -gt 0 ] && [ "$1" != -o ]; do shift; done
[ "$#" -ge 2 ] || exit 2
out="$2"
printf '#!/bin/sh\necho partial\n' > "$out"; chmod +x "$out"; echo "compile error" >&2; exit 1
STUBEOF
chmod +x "$FAILGO/go"
fail_out="$(PATH="$FAILGO:$PATH" bash -c "$nous_def; nous" 2>/dev/null)"; fail_rc=$?
[ "$fail_rc" -ne 0 ] && ok "a failed build fails the call" || ko "a failed build returned 0"
[ -z "$fail_out" ] && ok "a failed build runs nothing" || ko "a failed build ran: [$fail_out]"
left="$(ls -A "$WS/nous/bin" | grep -v '^nous$')"
[ -z "$left" ] && ok "a failed build removes its build directory" || ko "left in bin/ after a failed build: $left"
[ "$("$WS/nous/bin/nous")" = "built" ] && ok "a failed build keeps the previous binary" || ko "the previous binary was lost"

# The freshness matrix owns one repository; earlier cases are complete.
rm -rf "$WS/nous"

# #405: the build runs only when needed. It is skipped when bin/X exists and
# either COUCH_TREE is set (an agent shell in a Couch slot, which runs what the
# operator's :0 checkout last built) or no file in the owner or its outside
# local replace targets, outside .git and bin, is newer than bin/X. A skipped build skips dev-prepare too. A counting stub go
# and make record what ran; file times are set explicitly, so no case depends
# on timestamp resolution.
OWN="$WS/skipd"; mkdir -p "$OWN/construct" "$OWN/.git" "$OWN/assets"; mkcmd "$OWN/cmd/skipd"
printf '.PHONY: dev-prepare\ndev-prepare:\n' > "$OWN/Makefile"
: > "$OWN/assets/page.md"; : > "$OWN/.git/index"
CG="$WS/countbin"; mkdir -p "$CG"
cat > "$CG/go" <<'STUBEOF'
#!/bin/sh
echo go >> "$TRACE"
while [ "$#" -gt 0 ] && [ "$1" != -o ]; do shift; done
[ "$#" -ge 2 ] || exit 2
printf '#!/bin/sh\necho ran "$@"\n' > "$2"; chmod +x "$2"
STUBEOF
printf '#!/bin/sh\necho make >> "$TRACE"\n' > "$CG/make"
chmod +x "$CG/go" "$CG/make"
skipd_def="$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^skipd()')"
export TRACE="$WS/skiptrace"
# call [ENV=VAL…]: run the alias once; leaves what ran in $TRACE and the
# alias's stdout in $got.
call() {
    : > "$TRACE"
    got="$(export PATH="$CG:$PATH"; [ "$#" -eq 0 ] || export "$@"; eval "$skipd_def"; skipd x)" 2>/dev/null
}
age_all() { find "$OWN" -path "$OWN/bin" -prune -o -exec touch -t 200001010000 {} +; }
expect() { # expect TRACE-CONTENT LABEL
	local trace="" event
    while IFS= read -r event; do trace="$trace$event "; done < "$TRACE"
    if [ "$trace" = "$1" ] && [ "$got" = "ran x" ]; then ok "$2"; else ko "$2 — ran: [$trace] stdout: [$got]"; fi
}
call; expect "make go " "a missing binary builds (dev-prepare, then go build)"
age_all
call; expect "" "a fresh binary skips dev-prepare and the build"
touch -t 203001010000 "$OWN/.git/index" "$OWN/bin/.stray"
call; expect "" "files under .git and bin never make it stale"
touch -t 203001010000 "$OWN/assets/page.md"
call; expect "make go " "any newer owner file, not only Go source, makes it stale"
age_all; touch -t 203001010000 "$OWN/cmd/skipd/main.go"
call COUCH_TREE=/some/slot; expect "" "in a Couch slot a stale binary is run, not rebuilt"
rm -f "$OWN/bin/skipd"
call COUCH_TREE=/some/slot; expect "make go " "in a Couch slot a missing binary still builds"
# Deleting or adding a file changes no file's time but its directory's, and
# the scan sees that.
age_all; : > "$OWN/cmd/skipd/extra.go"; touch -t 200001010000 "$OWN/cmd/skipd/extra.go"
call; expect "make go " "an added file makes it stale"
age_all; rm -f "$OWN/cmd/skipd/extra.go"
call; expect "make go " "a deleted file makes it stale"
# A find error (an unreadable directory) counts as stale, not fresh.
age_all; mkdir -p "$OWN/locked"; chmod 000 "$OWN/locked"
call; expect "make go " "a scan error counts as stale"
chmod 755 "$OWN/locked"; rm -rf "$OWN/locked"
# An edit made while the build runs is newer than the binary it produced, so
# the next call rebuilds: the binary takes the time its build started.
age_all
printf '#!/bin/sh\necho go >> "$TRACE"\ntouch -t 203001010000 "%s/cmd/skipd/main.go"\nwhile [ "$#" -gt 0 ] && [ "$1" != -o ]; do shift; done\nprintf "#!/bin/sh\\necho ran \\"\\$@\\"\\n" > "$2"; chmod +x "$2"\n' "$OWN" > "$CG/go"
touch -t 203001010000 "$OWN/assets/page.md"
call; expect "make go " "a stale tree rebuilds (the stub edits a source mid-build)"
printf '#!/bin/sh\necho go >> "$TRACE"\nwhile [ "$#" -gt 0 ] && [ "$1" != -o ]; do shift; done\nprintf "#!/bin/sh\\necho ran \\"\\$@\\"\\n" > "$2"; chmod +x "$2"\n' > "$CG/go"
touch -t 200001010000 "$OWN/assets/page.md"
call; expect "make go " "an edit made during the build makes the next call rebuild"
# A local replace target outside the owner is scanned too; its .git and bin are
# not, and a replace inside the owner adds no second root.
DEP="$WS/skipdep"; mkdir -p "$DEP/.git" "$DEP/bin" "$OWN/third_party/vt"
: > "$DEP/lib.go"; : > "$DEP/.git/index"; : > "$DEP/bin/tool"
printf 'module example.com/skipd\n\nreplace example.com/dep => ../skipdep // local\nreplace (\n\texample.com/vt => ./third_party/vt\n\texample.com/remote => example.com/fork v1.0.0\n)\n' > "$OWN/go.mod"
skipd_def="$(bash "$GEN" --workspace "$WS" 2>/dev/null | grep '^skipd()')"
assert_contains "find . ../skipdep \\( -path ./.git -o -path ./bin -o -path ../skipdep/.git -o -path ../skipdep/bin \\)" "$skipd_def" "the scan roots are the owner and its outside local replace target"
assert_not_contains "third_party" "$skipd_def" "a replace inside the owner adds no root"
call; age_all; find "$DEP" -exec touch -t 200001010000 {} +
call; expect "" "an unchanged replace target keeps the binary fresh"
touch -t 203001010000 "$DEP/.git/index" "$DEP/bin/tool"
call; expect "" "a replace target's .git and bin never make it stale"
touch -t 203001010000 "$DEP/lib.go"
call; expect "make go " "a newer file in a replace target makes it stale"
rm -rf "$DEP"
unset TRACE
rm -rf "$OWN" "$CG" "$WS/skiptrace"

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
