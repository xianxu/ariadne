#!/bin/sh
# Opt-in control: feed actual #530 encoder bytes to the landed #524 reader.
# Every repository and status effect is private; no live binary is replaced.
set -eu
repo=$(git rev-parse --show-toplevel)
base=0c1fecc1b93d8f07bfcd138bf88a9bfd69b716fe
scratch=$(mktemp -d "${TMPDIR:-/tmp}/ariadne-524-reader.XXXXXX") || exit 1
printf 'Retained for inspection; invoker removes this private directory afterwards: %s\n' "$scratch"
cd "${repo:?}"
SDLC_RECLASSIFY_EXPORT="${scratch:?}/reclassified-v3.json" go test ./cmd/sdlc -run '^TestSuiteReclassificationPublication$' -count=1 -timeout 5m
git clone --quiet --shared --no-checkout "$repo" "${scratch:?}/repo"
git -C "$scratch/repo" checkout --quiet --detach "$base"
cp "$repo/cmd/sdlc/testdata/evidence-524-control_test.go.txt" "$scratch/repo/cmd/sdlc/compat_530_test.go"
cp "$scratch/reclassified-v3.json" "$scratch/repo/cmd/sdlc/testdata/reclassified-v3.json"
git -C "$scratch/repo" add cmd/sdlc/compat_530_test.go cmd/sdlc/testdata/reclassified-v3.json
printf '524 reader source: %s\nEvidence checkout: %s\n' "$base" "$scratch/repo"
cd "${scratch:?}/repo"
go test ./cmd/sdlc -run '^Test524ReaderReclassifiedAttemptCompatibility$' -count=1 -timeout 5m
