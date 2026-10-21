#!/bin/sh
# Opt-in compatibility control. Runs only one targeted test on immutable old
# source in a private clone; does not replace any live binary or store.
set -eu
repo=$(git rev-parse --show-toplevel)
base=63ebb3385af267d64084921c6e8d1150a188c734
scratch=$(mktemp -d "${TMPDIR:-/tmp}/ariadne-pre524-reader.XXXXXX") || exit 1
git clone --quiet --shared --no-checkout "$repo" "${scratch:?}/repo"
git -C "$scratch/repo" checkout --quiet --detach "$base"
cp "$repo/cmd/sdlc/testdata/evidence-pre524-control_test.go.txt" "$scratch/repo/cmd/sdlc/compat_pre524_test.go"
git -C "$scratch/repo" add cmd/sdlc/compat_pre524_test.go
printf 'Pre-524 reader source: %s\nEvidence checkout: %s\n' "$base" "$scratch/repo"
cd "${scratch:?}/repo"
go test ./cmd/sdlc -run '^TestPre524FutureRecordPoisonsUnrelatedProof$' -count=1 -timeout 5m
