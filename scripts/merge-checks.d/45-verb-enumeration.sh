#!/usr/bin/env bash
# Fail a comment that RESTATES the manifest-verb set instead of naming its source.
#
# Why this exists: ariadne#239's thesis is that a hand-maintained restatement of
# a model is a deferred consumer, not a finished one (ARCH-PURPOSE). Adding one
# verb (`seed-once`) turned up FOURTEEN prose copies of the verb set across Go
# doc comments, atlas pages and a target — and the same finding family recurred
# for three review rounds, because each round fixed the instances it remembered
# and the class was never enumerated. One round even edited a list into a FRESH
# error while fixing it.
#
# The rule this enforces: prose names the source, it does not restate the set.
# The source is `intent.kindByVerb` (cmd/weave/internal/intent/manifest.go) and
# the `Action` sum type (cmd/weave/internal/plan/action.go).
#
# The check: a COMMENT line naming >= MIN_VERBS DISTINCT verbs (word-boundary
# matched) must also name a source. Only comments are scanned — the switches
# themselves legitimately list every verb, and they are what prose should point
# at. Matching is CASE-INSENSITIVE, because the Action sum type restates the
# same set in CamelCase (Symlink, Seed, SeedOnce…) and a lowercase-only match
# could not see it. The threshold is deliberately high: prose that mentions two or three verbs
# is usually explaining a behaviour, while four or more is enumerating the SET.
# A check that fires on legitimate prose gets routed around, so it errs toward
# silence and catches the shape that actually recurred.
#
# Scope note: base.manifest is scanned too. It is the DECLARATION SURFACE every
# agent reads, so a stale verb list there is the costliest of all — and the first
# version of this check could not see it, because it globbed only *.go and *.md
# (#239 M1 BR-16).
#
# Usage: 45-verb-enumeration.sh <base_sha> <head_sha>   (merge-check contract)
set -euo pipefail

BASE="${1:-}"
HEAD="${2:-}"
MIN_VERBS=4
ROOT="$(git rev-parse --show-toplevel)"

# The verb vocabulary, read from the SOURCE rather than hardcoded here — this
# check must not become the fifteenth restatement.
MANIFEST_GO="$ROOT/cmd/weave/internal/intent/manifest.go"
if [ ! -f "$MANIFEST_GO" ]; then
    echo "✓ verb-enumeration: no intent/manifest.go here — not the weave owner, skipping"
    exit 0
fi
VERBS=$(sed -n '/^var kindByVerb = map\[string\]Kind{/,/^}/p' "$MANIFEST_GO" \
        | grep -oE '"[a-z-]+"' | tr -d '"' | sort -u)
[ -n "$VERBS" ] || { echo "verb-enumeration: could not read kindByVerb from $MANIFEST_GO" >&2; exit 1; }

# Scope: changed files when given a range, else the whole tree.
if [ -n "$BASE" ] && [ -n "$HEAD" ]; then
    FILES=$(git diff --name-only "$BASE" "$HEAD" -- '*.go' '*.md' '*.manifest' | grep -v '^workshop/history/' || true)
else
    FILES=$(git ls-files '*.go' '*.md' '*.manifest' | grep -v '^workshop/history/' || true)
fi
[ -n "$FILES" ] || { echo "✓ verb-enumeration: no Go/Markdown files in range"; exit 0; }

# ONE awk pass over all candidate files. A per-verb subprocess per line is
# O(lines x verbs) processes and takes minutes on this tree — a merge check that
# slow never gets run.
report=$(printf '%s\n' $FILES \
  | grep -vE '_test\.go$|^workshop/(plans|issues)/|^workshop/lessons\.md$|^cmd/weave/internal/intent/(manifest|intent)\.go$|^scripts/merge-checks\.d/45-verb-enumeration\.sh$' \
  | while read -r f; do [ -f "$ROOT/$f" ] && echo "$f"; done \
  | xargs -r awk -v verbs="$(printf '%s ' $VERBS)" -v min="$MIN_VERBS" '
    BEGIN { n = split(verbs, V, /[ \n]+/) }
    # comment lines only: the switches themselves may list every verb
    /^[[:space:]]*(\/\/|#|\*|-|>)/ {
      c = 0
      for (i = 1; i <= n; i++) {
        if (V[i] == "") continue
        if (tolower($0) ~ "(^|[^a-zA-Z-])" V[i] "([^a-zA-Z-]|$)") c++
      }
      if (c < min) next
      if ($0 ~ /kindByVerb|Action sum type|the source|switch below/) next
      line = $0; sub(/^[[:space:]]+/, "", line)
      printf "  %s:%d\n    %s\n", FILENAME, FNR, line
    }' )

violations=$(printf '%s' "$report" | grep -c ':[0-9]*$' || true)
[ -n "$report" ] && printf '%s\n' "$report"

if [ "$violations" -gt 0 ]; then
    cat >&2 <<EOF

✗ verb-enumeration: $violations comment(s) restate the manifest-verb set.

  Prose must NAME THE SOURCE, not restate the set. A hand-written list derives
  from nothing, so it goes stale the next time a verb is added — which is
  ariadne#239's own thesis, one level down.

  Source of truth: intent.kindByVerb (cmd/weave/internal/intent/manifest.go)
                   the Action sum type (cmd/weave/internal/plan/action.go)

  Rewrite each line above to point at one of those instead of listing verbs.
EOF
    exit 1
fi

echo "✓ verb-enumeration: no comment restates the verb set"
