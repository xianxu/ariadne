#!/usr/bin/env bash
# Real Make entry points over a stateful scratch tree; no user tools/install.
set -euo pipefail
REAL_PATH="$PATH"
REAL_GO="$(command -v go)"
export GOMODCACHE="$(go env GOMODCACHE)"
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/portable-make.XXXXXX")"
trap 'rm -rf "$SCRATCH"' EXIT
mkdir -p "$SCRATCH/leaf/construct" "$SCRATCH/home"
export HOME="$SCRATCH/home"
cp "$SOURCE/Makefile" "$SCRATCH/leaf/Makefile"
cp "$SOURCE/bootstrap.sh" "$SCRATCH/leaf/bootstrap.sh"
printf 'product:\n\t@echo product-ok\nhelp: help-product\nhelp-product:\n\t@echo product-help\n' > "$SCRATCH/leaf/Makefile.local"
make -s -C "$SCRATCH/leaf" product help > "$SCRATCH/out"
grep -q product-ok "$SCRATCH/out"
grep -q product-help "$SCRATCH/out"
# Now bootstrap has supplied the upstream; all leaf helper links remain absent.
mkdir -p "$SCRATCH/ariadne/construct/scripts" "$SCRATCH/ariadne/cmd/weave" "$SCRATCH/bin"
cp "$SOURCE/Makefile.workflow" "$SCRATCH/ariadne/Makefile.workflow"
printf 'substrate ../ariadne\n' > "$SCRATCH/leaf/construct/deps"
cat > "$SCRATCH/ariadne/construct/dev-aliases.sh" <<'SCRIPT'
#!/bin/sh
printf 'weave\t%s/ariadne\ndatatype\t%s/ariadne\nvocabulary\t%s/ariadne\n' "$FIXTURE" "$FIXTURE" "$FIXTURE"
SCRIPT
cat > "$SCRATCH/ariadne/construct/scripts/bootstrap-peers.sh" <<'SCRIPT'
#!/bin/sh
printf 'peers\n' >> "$FIXTURE/events"
touch "$FIXTURE/peers-ready"
SCRIPT
cat > "$SCRATCH/bin/go" <<'SCRIPT'
#!/bin/sh
[ -f "$FIXTURE/peers-ready" ] || { echo 'build raced peers' >&2; exit 1; }
while [ "$#" -gt 0 ]; do
 if [ "$1" = -o ]; then
  shift
  mkdir -p "$(dirname "$1")"
  cp "$FIXTURE/weave-fixture" "$1"
  chmod +x "$1"
  break
 fi
 shift
done
SCRIPT
cat > "$SCRATCH/weave-fixture" <<'SCRIPT'
#!/bin/sh
printf 'weave\n' >> "$FIXTURE/events"
[ "${FAIL_WEAVE:-0}" != 1 ] || exit 7
mkdir -p construct/scripts scripts
printf '#!/bin/sh\ntest -f woven\necho install >> "$FIXTURE/events"\n' > scripts/sdlc-install.sh
printf '#!/bin/sh\ntest -f woven\necho data >> "$FIXTURE/events"\n' > construct/scripts/clone-data-deps.sh
chmod +x scripts/sdlc-install.sh construct/scripts/clone-data-deps.sh
touch woven
SCRIPT
printf '#!/bin/sh\nexit 0\n' > "$SCRATCH/bin/cue"
cp "$SCRATCH/bin/cue" "$SCRATCH/bin/uv"
chmod +x "$SCRATCH/bin/"* "$SCRATCH/ariadne/construct/dev-aliases.sh" "$SCRATCH/ariadne/construct/scripts/bootstrap-peers.sh"
export FIXTURE="$SCRATCH" PATH="$SCRATCH/bin:$PATH"
# Override only post-weave builds; retain real pre-weave owner resolution/builds.
printf 'tools: fixture-tools\nfixture-tools:\n\t@test -f woven\n\t@echo tools >> "$(FIXTURE)/events"\nsdlc-build build:\n\t@:\n' >> "$SCRATCH/leaf/Makefile.local"
make -s -C "$SCRATCH/leaf" help > "$SCRATCH/out"
grep -q 'AI Workflow' "$SCRATCH/out"
make -s -j4 -C "$SCRATCH/leaf" bootstrap > "$SCRATCH/out" 2>&1 || { cat "$SCRATCH/out"; exit 1; }
[ ! -e "$SCRATCH/leaf/bootstrap" ] || { echo "implicit make rule created bootstrap"; exit 1; }
python3 - "$SCRATCH/events" <<'PY'
import sys
rows=open(sys.argv[1]).read().splitlines()
assert rows.index('peers') < rows.index('weave') < rows.index('tools'), rows
assert rows.index('weave') < rows.index('install'), rows
assert rows.index('weave') < rows.index('data'), rows
PY
: > "$SCRATCH/events"
if FAIL_WEAVE=1 make -s -C "$SCRATCH/leaf" bootstrap > "$SCRATCH/out" 2>&1; then echo 'failed weave passed'; exit 1; fi
! grep -Eq 'tools|install|data' "$SCRATCH/events"
# Explicit local overlay wins, and does not require workflow-defined help.
printf 'local-overlay:\n\t@echo local-overlay-ok\n' > "$SCRATCH/leaf/Makefile.workflow"
make -s -C "$SCRATCH/leaf" local-overlay help > "$SCRATCH/out"
grep -q local-overlay-ok "$SCRATCH/out"
# Live conformance: compile the production manifest's changed delivery surface
# twice with real weave. The fixture owns its upstream; no live peer is touched.
PATH="$REAL_PATH" "$REAL_GO" build -o "$SCRATCH/real-weave" "$SOURCE/cmd/weave"
cp "$SOURCE/Makefile" "$SCRATCH/ariadne/Makefile"
cp "$SOURCE/construct/Makefile.seed" "$SCRATCH/ariadne/construct/Makefile.seed"
cp "$SOURCE/bootstrap.sh" "$SCRATCH/ariadne/bootstrap.sh"
mkdir -p "$SCRATCH/ariadne/.github/workflows"
cp "$SOURCE/.github/workflows/merge-check.yml" "$SCRATCH/ariadne/.github/workflows/merge-check.yml"
awk '$2 == "construct/Makefile.seed" || $2 == "Makefile.workflow" || $2 == "bootstrap.sh" || $2 == ".github/workflows/merge-check.yml"' "$SOURCE/construct/base.manifest" > "$SCRATCH/ariadne/construct/base.manifest"
: > "$SCRATCH/leaf/construct/base.manifest"
printf '#!/bin/sh\necho consumer-setup\n' > "$SCRATCH/leaf/scripts/ci-setup.sh"
chmod +x "$SCRATCH/leaf/scripts/ci-setup.sh"
cp "$SCRATCH/leaf/scripts/ci-setup.sh" "$SCRATCH/hook-before"
rm "$SCRATCH/leaf/Makefile" "$SCRATCH/leaf/Makefile.workflow"
ln -s ../ariadne/Makefile "$SCRATCH/leaf/Makefile"
cp "$SCRATCH/ariadne/Makefile" "$SCRATCH/ancestor-before"
(cd "$SCRATCH/leaf" && "$SCRATCH/real-weave" compile)
[ ! -L "$SCRATCH/leaf/Makefile" ]
cmp "$SCRATCH/ancestor-before" "$SCRATCH/ariadne/Makefile"
cmp "$SCRATCH/ariadne/construct/Makefile.seed" "$SCRATCH/leaf/Makefile"
cp "$SCRATCH/leaf/Makefile" "$SCRATCH/first-weave"
(cd "$SCRATCH/leaf" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/first-weave" "$SCRATCH/leaf/Makefile"
cmp "$SCRATCH/hook-before" "$SCRATCH/leaf/scripts/ci-setup.sh"
[ -x "$SCRATCH/leaf/scripts/ci-setup.sh" ]
cmp "$SOURCE/.github/workflows/merge-check.yml" "$SCRATCH/leaf/.github/workflows/merge-check.yml"
make -s -C "$SCRATCH/leaf" product help > "$SCRATCH/out"
grep -q product-ok "$SCRATCH/out"
# #239: a repo-owned root Makefile survives weave byte-for-byte, FOREVER.
# Pre-#239 `seed Makefile` was content-tracking and silently destroyed it on the
# first weave, and any later edit to it on EVERY subsequent weave. No test
# covered either half — which is how the defect survived #225.
mkdir -p "$SCRATCH/adopter/construct"
printf 'substrate ../ariadne\n' > "$SCRATCH/adopter/construct/deps"
: > "$SCRATCH/adopter/construct/base.manifest"
printf 'MY OWN BUILD SYSTEM\ninclude Makefile.workflow\n' > "$SCRATCH/adopter/Makefile"
cp "$SCRATCH/adopter/Makefile" "$SCRATCH/adopter-before"
(cd "$SCRATCH/adopter" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/adopter-before" "$SCRATCH/adopter/Makefile"
printf 'AND A LATER LOCAL EDIT\n' >> "$SCRATCH/adopter/Makefile"
cp "$SCRATCH/adopter/Makefile" "$SCRATCH/adopter-edited"
(cd "$SCRATCH/adopter" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/adopter-edited" "$SCRATCH/adopter/Makefile"
echo 'PASS portable Make product/overlay/bootstrap ordering and real weave convergence'
