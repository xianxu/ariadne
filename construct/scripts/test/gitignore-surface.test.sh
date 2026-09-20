#!/usr/bin/env bash
# #239: the committed-surface invariant, end to end, against real git.
#
# A derivative commits only its bootstrap core plus its own source; everything
# `make weave` RE-DERIVES is ignored, per-path.
#
# Two things the obvious version of this test gets wrong, both measured:
#
#   1. `git check-ignore` is INDEX-AWARE — for a TRACKED file it exits 1
#      whatever the patterns say. So `! git check-ignore -q f` passes even when
#      a blanket pattern does match f, and the assertion that matters most would
#      be vacuous. Every "is not ignored" check below uses --no-index, and the
#      decisive one mirrors the real sweep with `git ls-files -i -c`, which is
#      literally what sdlc's commitConsumption runs.
#   2. The fixture needs a `prose` row, or plan.Plan emits no entry-file
#      WriteFile and there is no /CLAUDE.md entry to assert.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/gitignore-surface.XXXXXX")"
trap 'rm -rf "$SCRATCH"' EXIT
(cd "$SOURCE" && go build -o "$SCRATCH/weave" ./cmd/weave)

# Upstream: a miniature ariadne carrying one row of each ownership class.
mkdir -p "$SCRATCH/up/construct/scripts" "$SCRATCH/up/scripts/merge-checks.d"
printf 'UP\n'       > "$SCRATCH/up/scripts/lib.sh"
printf 'CHECK\n'    > "$SCRATCH/up/scripts/merge-checks.d/40-dup.sh"
printf 'BOOT\n'     > "$SCRATCH/up/bootstrap.sh"
printf 'TEMPLATE\n' > "$SCRATCH/up/construct/Makefile.seed"
printf '# Base constitution\n' > "$SCRATCH/up/AGENTS.base.md"
cat > "$SCRATCH/up/construct/base.manifest" <<'EOF'
export    prose AGENTS.base.md
symlink   scripts/lib.sh
symlink   scripts/merge-checks.d/40-dup.sh
scaffold  scripts/merge-checks.d
scaffold  workshop/issues
touch     workshop/lessons.md
seed      bootstrap.sh
seed-once construct/Makefile.seed Makefile
EOF

# Leaf: a derivative with its OWN files in the mixed directories, plus its own
# .gitignore entries and negations (pair's bin/* + !bin/*.sh shape).
mkdir -p "$SCRATCH/leaf/construct" "$SCRATCH/leaf/scripts/merge-checks.d" "$SCRATCH/leaf/bin"
printf 'substrate ../up\n' > "$SCRATCH/leaf/construct/deps"
: > "$SCRATCH/leaf/construct/base.manifest"
printf 'MINE\n' > "$SCRATCH/leaf/scripts/merge-checks.d/20-vocabulary.sh"
printf 'MINE\n' > "$SCRATCH/leaf/scripts/ci-setup.sh"
printf 'MINE\n' > "$SCRATCH/leaf/bin/helper.sh"
printf 'bin/*\n!bin/*.sh\ncache/\n' > "$SCRATCH/leaf/.gitignore"
cp "$SCRATCH/leaf/.gitignore" "$SCRATCH/repo-entries-before"
(cd "$SCRATCH/leaf" && git init -q . \
  && git -c user.email=t@t -c user.name=t add -A \
  && git -c user.email=t@t -c user.name=t commit -qm init)

cd "$SCRATCH/leaf"
"$SCRATCH/weave" compile >/dev/null

fail() { echo "FAIL: $*" >&2; exit 1; }

# 1. THE decisive assertion: nothing repo-owned would be swept. `ls-files -i -c`
#    is exactly what sdlc's commitConsumption runs to untrack now-ignored files,
#    so an empty intersection with the repo's own files IS the guarantee.
git ls-files -i -c --exclude-standard > "$SCRATCH/would-untrack"
for f in scripts/merge-checks.d/20-vocabulary.sh scripts/ci-setup.sh bin/helper.sh \
         workshop/lessons.md bootstrap.sh Makefile construct/deps; do
  ! grep -qxF "$f" "$SCRATCH/would-untrack" || fail "the sweep would untrack repo-owned $f"
  ! git check-ignore -q --no-index "$f" || fail "repo-owned $f is ignored"
done
! git check-ignore -q --no-index workshop/issues || fail "scaffold dir workshop/issues is ignored"

# 2. Re-derived paths ARE ignored. (weave compile defaults to TargetAll, so all
#    three entry files exist.)
for f in scripts/lib.sh scripts/merge-checks.d/40-dup.sh CLAUDE.md AGENTS.md GEMINI.md; do
  git check-ignore -q --no-index "$f" || fail "weave-generated $f is not ignored"
done

# 3. The repo's own entries and negations round-trip, ahead of the block.
sed -n "1,/^# >>> weave-generated/p" .gitignore | sed '$d' | cmp - "$SCRATCH/repo-entries-before"

# 4. A second weave writes nothing.
cp .gitignore "$SCRATCH/after-first"
"$SCRATCH/weave" compile >/dev/null
cmp "$SCRATCH/after-first" .gitignore

# 5. RETIRING a manifest row removes its ignore line (append-only could not).
grep -v '40-dup' "$SCRATCH/up/construct/base.manifest" > "$SCRATCH/m"
mv "$SCRATCH/m" "$SCRATCH/up/construct/base.manifest"
"$SCRATCH/weave" compile >/dev/null
! grep -q '40-dup' .gitignore || fail "retired row left a stale ignore line"
grep -q 'bin/\*' .gitignore   || fail "repo entries lost on retire"

echo 'PASS gitignore committed-surface invariant'
