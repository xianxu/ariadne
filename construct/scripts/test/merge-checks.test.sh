#!/usr/bin/env bash
# Falsification fixtures for the restatement merge checks (45-, 46-).
#
# WHY THIS FILE EXISTS: a check's falsification must live in the repo as a
# fixture the runner re-executes — not as a one-time manual sweep recorded in
# prose. ariadne#239 proved the difference the hard way: a hand-run sweep
# recorded "✓ site enforced" for a site the check never covered, and two of the
# checks' own bugs (a regex that ate the symbol name; `git grep -E` having no
# `\b`) made them print green for a whole round. A green run is not evidence
# until something has been observed to make it red, repeatably (#239 M2 BR-26).
#
# Each case below is a SHAPE that motivated one of the checks. Every case must
# make its check exit non-zero; the clean tree must exit zero.
#
# Run: bash construct/scripts/test/merge-checks.test.sh
set -uo pipefail
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/merge-checks.XXXXXX")"
trap 'rm -rf "$SCRATCH"' EXIT
C45="$SOURCE/scripts/merge-checks.d/45-verb-enumeration.sh"
C46="$SOURCE/scripts/merge-checks.d/46-removed-symbol-references.sh"
pass=0 fail=0

ok()   { printf '  ok   %s\n' "$1"; pass=$((pass+1)); }
bad()  { printf '  FAIL %s — %s\n' "$1" "$2"; fail=$((fail+1)); }

# expect_red <label> <repo> <expected-substring> — the check must exit NON-ZERO
# *and* say why, naming the site. Exit status alone is not enough: the first
# version of this harness did `run && bad || ok`, so ANY non-zero exit counted as
# "went red — enforced", including a non-zero exit because the fixture failed to
# build, the check path was wrong, or git errored. Every red case would have
# reported ok while testing nothing — the cannot-fail shape, inside the harness
# written to prevent it (#239 M2 BR-31).
expect_red() {
    local label="$1" repo="$2" want="$3" out status
    out=$( cd "$repo" && bash "$CHECK" $ARGS 2>&1 ); status=$?
    if [ "$status" -eq 0 ]; then
        bad "$label" "exited 0, want non-zero"
    elif ! printf '%s' "$out" | grep -qF "$want"; then
        bad "$label" "exited $status but never mentioned '$want' — red for the WRONG REASON: $(printf '%s' "$out" | head -1)"
    else
        ok "$label"
    fi
}

# expect_green <label> <repo> — must exit 0 AND not have died before checking.
expect_green() {
    local label="$1" repo="$2" out status
    out=$( cd "$repo" && bash "$CHECK" $ARGS 2>&1 ); status=$?
    if [ "$status" -ne 0 ]; then
        bad "$label" "exited $status, want 0: $(printf '%s' "$out" | head -1)"
    elif ! printf '%s' "$out" | grep -q '^✓'; then
        bad "$label" "exited 0 without a ✓ line — did it actually run? $(printf '%s' "$out" | head -1)"
    else
        ok "$label"
    fi
}

# --- 46: a repo where a symbol is removed and a comment still names it -------
# build46 <label> <born-in-range?> <trailing comment line> ; echoes the repo path
build46() {
    local label="$1" born="$2" comment="$3"
    local r="$SCRATCH/46-$label"
    mkdir -p "$r/pkg" && git init -q "$r"
    git -C "$r" config user.email t@t; git -C "$r" config user.name t
    if [ "$born" = no ]; then
        printf 'package pkg\n\nconst removedConst = 1\n\nfunc removedFunc() {}\n' > "$r/pkg/a.go"
    else
        printf 'package pkg\n' > "$r/pkg/a.go"
    fi
    git -C "$r" add -A >/dev/null; git -C "$r" commit -qm base
    git -C "$r" tag base
    if [ "$born" = yes ]; then
        # BORN inside the range, then buried by the next commit.
        printf 'package pkg\n\nconst removedConst = 1\n\nfunc removedFunc() {}\n' > "$r/pkg/a.go"
        git -C "$r" add -A >/dev/null; git -C "$r" commit -qm born
    fi
    printf 'package pkg\n\n%s\nfunc kept() {}\n' "$comment" > "$r/pkg/a.go"
    git -C "$r" add -A >/dev/null; git -C "$r" commit -qm buried
    echo "$r"
}

CHECK="$C46"; ARGS="base HEAD"

# 1. The plain shape: a removed func still named in a comment.
r=$(build46 plain no '// removedFunc did the thing.')
expect_red "46/plain-removed-func" "$r" "removedFunc"

# 2. BORN AND BURIED inside the range. A two-point `git diff BASE HEAD` collapses
#    this and sees nothing — the bug that made the check blind to one of the two
#    sites it was built for.
r=$(build46 bornburied yes '// removedFunc did the thing.')
expect_red "46/born-and-buried" "$r" "removedFunc"

# 3. A GROUPED declaration. const/var/type inside a `( … )` block is not at
#    column 0, so a column-0-only extractor cannot see it renamed — and the file
#    that motivated this check declares its markers exactly that way.
r="$SCRATCH/46-grouped"; mkdir -p "$r/pkg" && git init -q "$r"
git -C "$r" config user.email t@t; git -C "$r" config user.name t
printf 'package pkg\n\nconst (\n\tgroupedMarker = "x"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm base; git -C "$r" tag base
printf 'package pkg\n\n// groupedMarker delimits the region.\nconst (\n\trenamedMarker = "x"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm rename
expect_red "46/grouped-const-decl" "$r" "groupedMarker"

# 4. A stale restatement that merely CONTAINS a retirement substring. The
#    allowlist must key on the mention being ABOUT the removal, not on the
#    string "remov" appearing anywhere in the sentence.
r=$(build46 substring no '// removedFunc appends absent entries and never removes.')
expect_red "46/retirement-substring-not-a-pass" "$r" "removedFunc"

# 5. A genuinely HISTORICAL mention must be allowed — it explains the current
#    shape and is the opposite of a stale restatement.
r=$(build46 historical no '// kept replaced removedFunc, which could not handle X.')
expect_green "46/historical-mention-allowed" "$r"

# 6. A clean repo must pass.
r=$(build46 clean no '// kept does the thing.')
expect_green "46/clean-tree-passes" "$r"

# 7. A grouped member REORDERED, not removed. The diff shows it as `-name`, so
#    an extractor that sees grouped members must be matched by a still-declared
#    test that ALSO sees them — otherwise a reorder reads as a removal and the
#    check false-positives. The two halves of the declaration grammar were
#    written separately and disagreed (#239 M2 BR-30).
r="$SCRATCH/46-reorder"; mkdir -p "$r/pkg" && git init -q "$r"
git -C "$r" config user.email t@t; git -C "$r" config user.name t
printf 'package pkg\n\nconst (\n\talpha = "a"\n\tbeta  = "b"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm base; git -C "$r" tag base
printf 'package pkg\n\n// alpha and beta delimit the region.\nconst (\n\tbeta  = "b"\n\talpha = "a"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm reorder
expect_green "46/grouped-reorder-is-not-a-removal" "$r"

# 8. The REAL shape that motivated the check: a grouped marker pair renamed,
#    with a comment still naming the old one. gitignore.go declares its markers
#    exactly this way.
r="$SCRATCH/46-realmarkers"; mkdir -p "$r/pkg" && git init -q "$r"
git -C "$r" config user.email t@t; git -C "$r" config user.name t
printf 'package pkg\n\nconst (\n\tmanagedBlockOpen  = ">>>"\n\tmanagedBlockClose = "<<<"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm base; git -C "$r" tag base
printf 'package pkg\n\n// The region runs from managedBlockOpen to its close.\nconst (\n\tregionOpen  = ">>>"\n\tregionClose = "<<<"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm rename
expect_red "46/grouped-marker-rename" "$r" "managedBlockOpen"

# --- 45: a comment restating the manifest verb set ---------------------------
# 45 reads the WORKING TREE, so each case is a file dropped into a copy of the
# real repo's check inputs.
probe45() {
    local label="$1" line="$2" want="$3" # want: red|green
    local r="$SCRATCH/45-$label"
    rm -rf "$r"; mkdir -p "$r/cmd/weave/internal/intent" "$r/pkg"
    git init -q "$r"; git -C "$r" config user.email t@t; git -C "$r" config user.name t
    # The check reads its vocabulary FROM kindByVerb, so the fixture ships one.
    cat > "$r/cmd/weave/internal/intent/manifest.go" <<'EOF'
package intent

var kindByVerb = map[string]Kind{
	"symlink":   Symlink,
	"seed":      Seed,
	"seed-once": SeedOnce,
	"scaffold":  Scaffold,
	"touch":     Touch,
	"merge":     Merge,
	"prose":     Prose,
	"skill":     Skill,
}
EOF
    printf 'package pkg\n\n%s\nfunc x() {}\n' "$line" > "$r/pkg/a.go"
    git -C "$r" add -A >/dev/null; git -C "$r" commit -qm base
    CHECK="$C45"; ARGS=""
    if [ "$want" = red ]; then expect_red "45/$label" "$r" "restate"; else expect_green "45/$label" "$r"; fi
}

probe45 lowercase-verb-list '// Handles symlink, seed, seed-once, scaffold, touch and merge rows.' red
# The Action sum type restates the SAME set in CamelCase — invisible to a
# lowercase-only match, and action.go was one of the measured sites.
probe45 camelcase-action-list '// The set: Symlink, Seed, SeedOnce, Scaffold, Touch, Merge.' red
# Ordinary prose naming a verb or two must NOT trip it — a noisy check is routed
# around, which is worse than no check.
probe45 ordinary-prose '// applySeed copies the file; a symlink in the slot is replaced.' green
probe45 names-the-source '// The live set is intent.kindByVerb — symlink, seed, scaffold, touch, merge.' green

# SELF-TEST: a deliberately broken invocation must be reported as a FAILURE, not
# as ok. This is the guard on the guard — without it, every red assertion above
# could be passing for the wrong reason and nothing would say so.
CHECK="$SCRATCH/does-not-exist.sh"; ARGS="base HEAD"
selftest_before=$fail
expect_red "selftest/broken-invocation-must-not-report-ok" "$SCRATCH" "removedFunc" >/dev/null 2>&1
if [ "$fail" -gt "$selftest_before" ]; then
    ok "selftest/broken-invocation-is-caught"
    fail=$selftest_before   # that failure was the expected outcome
else
    bad "selftest/broken-invocation-is-caught" "a missing check script reported ok — red assertions are unfalsified"
fi

printf '\n== %d passed, %d failed ==\n' "$pass" "$fail"
[ "$fail" -eq 0 ] || exit 1
echo 'PASS merge-check falsification fixtures'
