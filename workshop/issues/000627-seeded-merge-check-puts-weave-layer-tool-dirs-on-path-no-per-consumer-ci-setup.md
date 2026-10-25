---
id: 000627
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '708b28f11d4ecc081e6c5d8bf48bad87edb340b0' # card fields mirrored from issue-cards; edit via sdlc
---

# Seeded merge-check puts weave layer tool dirs on PATH (no per-consumer ci-setup hook)

## Problem

The seeded `merge-check.yml` runs `./bootstrap.sh` (weave compile), which builds each
layer's tools into `<layer>/bin` and then only prints "weave: add these layer tool
directories to your shell PATH: ...". Nothing in the workflow adds them, so a
consumer's merge-checks that call a layer tool by name (metis: `vocabulary
validate-instance`) get command-not-found in CI. Found by metis#70 (PR
xianxu/metis#52, run 40188715202): metis's experiment selftest hid stderr, so the
missing tool read as "valid-baseline should PASS but was rejected". metis added a
temporary `scripts/ci-setup.sh` that appends `../ariadne/bin` and `bin` to
`$GITHUB_PATH`.

## Spec

The seeded workflow puts the layer tool directories weave reports on PATH itself,
after compile and before `scripts/ci-setup.sh` and the merge-checks, derived from
weave (a machine-readable listing, not parsed prose), so no consumer needs a
per-repository hook.

## Done when

- A consumer merge-check calling a layer tool by name finds it in CI with no
  consumer hook.
- metis can delete its `scripts/ci-setup.sh` (a local-gates change the operator
  verifies).

## Plan

- [ ]

## Log

### 2026-10-25
