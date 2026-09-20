#!/usr/bin/env bash
# Fail a comment that names a top-level Go symbol this range DELETED.
#
# Why this exists: ariadne#239 M2 renamed `ensureGitignoreText` to
# `mergeManagedBlock` and left three comments still describing the old
# append-only behaviour — one of them naming the deleted function directly above
# its replacement. The same round's BR-12 left `golden.go` naming
# `SeedOnceSlotIsRepoOwned` after it was replaced by `ClassifySlot`. Both are the
# same shape: a comment is a restatement that derives from nothing, so it does
# not move when the code does, and nothing fails.
#
# Sibling of 45-verb-enumeration.sh, which enforces the same rule for the verb
# SET. This one catches what that check cannot see: a comment naming an
# identifier the package no longer defines.
#
# HISTORICAL MENTIONS ARE ALLOWED. A comment explaining why the current code is
# shaped as it is ("the boolean it replaced…", "the retired X guarded this…") is
# valuable and must not be flagged — it is the opposite of a stale restatement.
# Such a line is recognised by a retirement word near the name.
#
# Usage: 46-removed-symbol-references.sh <base_sha> <head_sha>  (merge-check contract)
set -euo pipefail

BASE="${1:-}"
HEAD="${2:-}"
ROOT="$(git rev-parse --show-toplevel)"

# With no explicit range, fall back to the range CI uses rather than passing
# vacuously — a bare local run that "checks nothing" and prints ✓ is the same
# cannot-fail shape this check exists to catch (#239 M2 BR-28). Its sibling
# 45- scans the whole tree in the same situation.
if [ -z "$BASE" ] || [ -z "$HEAD" ]; then
    HEAD=HEAD
    BASE=$(git merge-base origin/main HEAD 2>/dev/null || git merge-base main HEAD 2>/dev/null || true)
    if [ -z "$BASE" ]; then
        echo "removed-symbol-references: no range given and no main/origin-main to derive one from" >&2
        exit 1
    fi
fi

# Top-level identifiers DELETED anywhere in the range.
#
# Computed PER COMMIT and unioned, never from a single `git diff BASE HEAD`.
# A two-point diff collapses a symbol that was both introduced and removed
# INSIDE the range — it never appears to have existed, so the check prints green
# over exactly the range CI uses. That is not hypothetical: SeedOnceSlotIsRepoOwned
# was added and removed on this branch, and it was one of the two sites this
# check was built for (#239 M2 BR-25). A check blind to its own motivating site
# is not evidence.
#
# Extraction is awk, not a chain of seds: the sed version's receiver-stripping
# group ate the function NAME, so the check silently reported "nothing removed"
# for the very rename that motivated it.
extract_removed() {
    awk '
        /^-func \(/      { sub(/^-func \([^)]*\) */, ""); sub(/[(<].*/, ""); print; next }
        /^-func /        { sub(/^-func */, "");            sub(/[(<].*/, ""); print; next }
        /^-(var|const|type) / { sub(/^-(var|const|type) */, ""); sub(/[ (<=].*/, ""); print; next }
        # GROUPED declarations: a member of `const ( … )` / `var ( … )` is not at
        # column 0, so a column-0-only extractor cannot see it renamed. The file
        # that motivated this check declares its markers exactly that way
        # (managedBlockOpen/Close), so the check was blind to its own subject
        # (#239 M2 BR-26).
        #
        # STATEFUL, because a naive "indented NAME =" pattern also matches every
        # assignment inside a function body — the first version extracted the
        # local `block` and then flagged every comment containing that English
        # word. Only lines between a group opener and its `)` count.
        /^[-+ ]?(const|var|type)[[:space:]]*\($/ { ingroup = 1; next }
        /^[-+ ]?\)[[:space:]]*$/               { ingroup = 0; next }
        /^diff --git|^@@/                      { ingroup = 0 }
        ingroup && /^-[[:space:]]+[A-Za-z_][A-Za-z0-9_]*/ {
            sub(/^-[[:space:]]+/, ""); sub(/[^A-Za-z0-9_].*/, ""); print; next }
      ' | grep -E '^[A-Za-z_][A-Za-z0-9_]*$' || true
}

removed=$(
    for c in $(git rev-list "$BASE".."$HEAD"); do
        git diff "$c^" "$c" -- '*.go' 2>/dev/null | extract_removed
    done | sort -u
)
[ -n "$removed" ] || { echo "✓ removed-symbol-references: no top-level symbols removed in range"; exit 0; }

# NOTE on the pattern: git grep -E is POSIX ERE, which has NO \b. Using it
# matched nothing and the check passed silently — the same "cannot fail" shape
# this check exists to catch. Word boundaries are spelled out explicitly.
WB_L='(^|[^A-Za-z0-9_])'
WB_R='([^A-Za-z0-9_]|$)'

still_declared() {
    git grep -qE "^(func|var|const|type) (\([^)]*\) )?$1$WB_R" "$HEAD" -- '*.go' 2>/dev/null
}

violations=0
for name in $removed; do
    still_declared "$name" && continue   # renamed-in-place or re-added: not removed
    # Comment lines naming it, outside tests and the process trees.
    # git grep with a REV prefixes each line "REV:path:lineno:line" — four
    # fields, not three. Reading three silently shifted every field.
    while IFS=: read -r _rev f lineno line; do
        [ -n "$f" ] || continue
        case "$f" in *_test.go|workshop/*|scripts/merge-checks.d/46-*) continue ;; esac
        # A historical mention explains the CURRENT shape — keep it. Two rules
        # learned the hard way (#239 M2 BR-26):
        #   1. Strip the SYMBOL NAME first. A bare substring test exempted every
        #      mention of a symbol whose own name contains "remov" (removeCache,
        #      removedFunc…) — the check could never flag those at all.
        #   2. Match PHRASES that are about the retirement, not fragments that
        #      appear in ordinary description. "…and never removes." contains
        #      "remov" but is a stale restatement, not a history note.
        stripped=$(printf '%s' "$line" | sed "s/${name}//g")
        case "$stripped" in
            *replac*|*Replac*|*retire*|*Retire*|*supersed*|*Supersed*|\
            *former*|*Former*|*"used to"*|*"no longer"*|*"removed in"*|*"deleted in"*) continue ;;
        esac
        printf '  %s:%s  names removed symbol %s\n    %s\n' \
            "$f" "$lineno" "$name" "$(printf '%s' "$line" | sed 's/^[[:space:]]*//')"
        violations=$((violations + 1))
    done < <(git grep -nE "^[[:space:]]*(//|\*).*${WB_L}${name}${WB_R}" "$HEAD" -- '*.go' '*.md' 2>/dev/null || true)
done

if [ "$violations" -gt 0 ]; then
    cat >&2 <<EOF

✗ removed-symbol-references: $violations comment(s) name a symbol this range deleted.

  A comment describing code that no longer exists is a restatement that derives
  from nothing — it does not move when the code moves, and nothing fails. Update
  each line to describe the replacement.

  If the mention is DELIBERATELY historical (explaining why the current shape is
  what it is), say so in the line — "replaced by", "retired", "superseded" — and
  this check will keep it.
EOF
    exit 1
fi

echo "✓ removed-symbol-references: no comment names a removed symbol"
