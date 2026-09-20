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

if [ -z "$BASE" ] || [ -z "$HEAD" ]; then
    echo "✓ removed-symbol-references: no commit range given — nothing to compare"
    exit 0
fi

# Top-level identifiers the range DELETED: a removed `func|var|const|type NAME`
# line whose name no longer appears as a declaration at HEAD.
# Extraction is awk, not a chain of seds: the sed version's receiver-stripping
# group ate the function NAME, so the check silently reported "nothing removed"
# for the very rename that motivated it. A check that cannot see its own case is
# the failure mode this whole family is about.
removed=$(git diff "$BASE" "$HEAD" -- '*.go' \
    | awk '
        /^-func \(/      { sub(/^-func \([^)]*\) */, ""); sub(/[(<].*/, ""); print; next }
        /^-func /        { sub(/^-func */, "");            sub(/[(<].*/, ""); print; next }
        /^-(var|const|type) / { sub(/^-(var|const|type) */, ""); sub(/[ (<=].*/, ""); print; next }
      ' \
    | grep -E '^[A-Za-z_][A-Za-z0-9_]*$' | sort -u || true)
[ -n "$removed" ] || { echo "✓ removed-symbol-references: no top-level symbols removed in range"; exit 0; }

# NOTE on the pattern: git grep -E is POSIX ERE, which has NO \b. Using it
# matched nothing and the check passed silently — the same "cannot fail" shape it
# exists to catch. Word boundaries are spelled out explicitly instead.
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
        # A historical mention explains the CURRENT shape — keep it.
        case "$line" in
            *retire*|*Retire*|*RETIRE*|*replac*|*Replac*|*remov*|*Remov*|\
            *delet*|*Delet*|*supersed*|*Supersed*|*former*|*Former*|*"used to"*) continue ;;
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
