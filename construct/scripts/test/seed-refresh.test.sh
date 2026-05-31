#!/usr/bin/env bash
# seed-refresh.test.sh — hermetic tests for setup.sh's create_seed(): a `seed`
# is a content-tracking real-file copy (a flattened symlink), so it is created
# on first run, refreshed when it drifts from upstream, and a no-op when already
# identical. Guards the regression where a derivative was stranded on a stale
# entrypoint (nous's pre-#45 bootstrap.sh) because seed was write-once.
# create_seed + ensure_parent are extracted verbatim from setup.sh so this tests
# the real code, not a copy.
#
# Run:  bash construct/scripts/test/seed-refresh.test.sh
# Exit: 0 if all assertions pass, 1 otherwise.
set -uo pipefail

ARIADNE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
SRC="$ARIADNE_ROOT/construct/setup.sh"

ROOT="$(mktemp -d "${TMPDIR:-/tmp}/seedrefresh.XXXXXX")"
trap 'rm -rf "$ROOT"' EXIT
cd "$ROOT"

pass=0; fail=0
ok() { printf '  \033[32mPASS\033[0m %s\n' "$1"; pass=$((pass+1)); }
ko() { printf '  \033[31mFAIL\033[0m %s\n' "$1"; fail=$((fail+1)); }

# Globals create_seed references, plus the two functions under test, lifted
# straight from setup.sh (def spans `name() {` … first line that is just `}`).
RED=''; GREEN=''; YELLOW=''; RESET=''
upstream="$ROOT/up"; TARGET_DIR="$ROOT/tgt"; mkdir -p "$upstream" "$TARGET_DIR"
awk '/^ensure_parent\(\) \{/,/^\}/' "$SRC" > "$ROOT/fns.sh"
awk '/^create_seed\(\) \{/,/^\}/'   "$SRC" >> "$ROOT/fns.sh"
grep -q 'cmp -s "\$src" "\$dst"' "$ROOT/fns.sh" \
    && ok "extracted create_seed() with content-compare from setup.sh" \
    || ko "create_seed() shape changed — no cmp guard found"
# shellcheck disable=SC1091
source "$ROOT/fns.sh"

src="$upstream/bootstrap.sh"
dst="$TARGET_DIR/bootstrap.sh"
printf '#!/usr/bin/env bash\necho v1\n' > "$src"; chmod +x "$src"

echo "== create_seed tracks upstream =="

# 1. First run on absent target → seeded.
out=$(create_seed "$src" "$dst")
[[ "$out" == *seeded* && "$(cat "$dst")" == "$(cat "$src")" ]] \
    && ok "absent target → seeded, content matches" || ko "out=[$out]"

# 2. Mode preserved (cp -p): executable source → executable target.
[[ -x "$dst" ]] && ok "executable bit preserved" || ko "target not executable"

# 3. Re-run, identical → silent no-op (no output, content unchanged).
out=$(create_seed "$src" "$dst")
[[ -z "$out" && "$(cat "$dst")" == "$(cat "$src")" ]] \
    && ok "identical → silent no-op (idempotent, no churn)" || ko "out=[$out]"

# 4. Upstream drifts → updated, target converges (THE regression this fixes).
printf '#!/usr/bin/env bash\necho v2-transitive\n' > "$src"
out=$(create_seed "$src" "$dst")
[[ "$out" == *updated* && "$(cat "$dst")" == "$(cat "$src")" ]] \
    && ok "upstream drifted → updated, target converges to v2" || ko "out=[$out] dst=[$(cat "$dst")]"

# 5. Re-run after update, identical again → no-op.
out=$(create_seed "$src" "$dst")
[[ -z "$out" ]] && ok "post-update re-run → no-op" || ko "out=[$out]"

# 6. Missing source → warn, non-fatal, target left intact.
out=$(create_seed "$upstream/nonexistent" "$dst"); rc=$?
[[ $rc -eq 0 && "$out" == *"seed source missing"* && "$(cat "$dst")" == "$(cat "$src")" ]] \
    && ok "missing source → warn, non-fatal, target untouched" || ko "rc=$rc out=[$out]"

echo
echo "== $pass passed, $fail failed =="
[[ $fail -eq 0 ]]
