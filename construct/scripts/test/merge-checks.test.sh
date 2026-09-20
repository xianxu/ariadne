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

run46() { ( cd "$1" && bash "$C46" base HEAD >/dev/null 2>&1 ); }

# 1. The plain shape: a removed func still named in a comment.
r=$(build46 plain no '// removedFunc did the thing.')
run46 "$r" && bad "46/plain-removed-func" "exited 0, want non-zero" || ok "46/plain-removed-func"

# 2. BORN AND BURIED inside the range. A two-point `git diff BASE HEAD` collapses
#    this and sees nothing — the bug that made the check blind to one of the two
#    sites it was built for.
r=$(build46 bornburied yes '// removedFunc did the thing.')
run46 "$r" && bad "46/born-and-buried" "exited 0, want non-zero" || ok "46/born-and-buried"

# 3. A GROUPED declaration. const/var/type inside a `( … )` block is not at
#    column 0, so a column-0-only extractor cannot see it renamed — and the file
#    that motivated this check declares its markers exactly that way.
r="$SCRATCH/46-grouped"; mkdir -p "$r/pkg" && git init -q "$r"
git -C "$r" config user.email t@t; git -C "$r" config user.name t
printf 'package pkg\n\nconst (\n\tgroupedMarker = "x"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm base; git -C "$r" tag base
printf 'package pkg\n\n// groupedMarker delimits the region.\nconst (\n\trenamedMarker = "x"\n)\n' > "$r/pkg/a.go"
git -C "$r" add -A >/dev/null; git -C "$r" commit -qm rename
run46 "$r" && bad "46/grouped-const-decl" "exited 0, want non-zero" || ok "46/grouped-const-decl"

# 4. A stale restatement that merely CONTAINS a retirement substring. The
#    allowlist must key on the mention being ABOUT the removal, not on the
#    string "remov" appearing anywhere in the sentence.
r=$(build46 substring no '// removedFunc appends absent entries and never removes.')
run46 "$r" && bad "46/retirement-substring-not-a-pass" "exited 0, want non-zero" || ok "46/retirement-substring-not-a-pass"

# 5. A genuinely HISTORICAL mention must be allowed — it explains the current
#    shape and is the opposite of a stale restatement.
r=$(build46 historical no '// kept replaced removedFunc, which could not handle X.')
run46 "$r" && ok "46/historical-mention-allowed" || bad "46/historical-mention-allowed" "exited non-zero, want 0"

# 6. A clean repo must pass.
r=$(build46 clean no '// kept does the thing.')
run46 "$r" && ok "46/clean-tree-passes" || bad "46/clean-tree-passes" "exited non-zero, want 0"

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
    if ( cd "$r" && bash "$C45" >/dev/null 2>&1 ); then
        [ "$want" = green ] && ok "45/$label" || bad "45/$label" "exited 0, want non-zero"
    else
        [ "$want" = red ] && ok "45/$label" || bad "45/$label" "exited non-zero, want 0"
    fi
}

probe45 lowercase-verb-list '// Handles symlink, seed, seed-once, scaffold, touch and merge rows.' red
# The Action sum type restates the SAME set in CamelCase — invisible to a
# lowercase-only match, and action.go was one of the measured sites.
probe45 camelcase-action-list '// The set: Symlink, Seed, SeedOnce, Scaffold, Touch, Merge.' red
# Ordinary prose naming a verb or two must NOT trip it — a noisy check is routed
# around, which is worse than no check.
probe45 ordinary-prose '// applySeed copies the file; a symlink in the slot is replaced.' green
probe45 names-the-source '// The live set is intent.kindByVerb — symlink, seed, scaffold, touch, merge.' green

printf '\n== %d passed, %d failed ==\n' "$pass" "$fail"
[ "$fail" -eq 0 ] || exit 1
echo 'PASS merge-check falsification fixtures'
