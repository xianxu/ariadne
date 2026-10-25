---
id: 000628
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '92763149681d6129d50ccc21a04c2c4fc89c8505' # card fields mirrored from issue-cards; edit via sdlc
---

# 40-duplicate-issue-id.sh exits 141 (SIGPIPE) locally: early-exiting awk under pipefail

## Problem

## Spec

**`scripts/merge-checks.d/40-duplicate-issue-id.sh` exits 141 (SIGPIPE) locally (found by metis:1, metis#70, 10-25).**
It pipes `construct/dev-aliases.sh --list` into an `awk` that exits on its first match; under `set -o pipefail` the
producer's SIGPIPE fails the whole check. Local-only (CI has no dev-aliases.sh, and there it passes). Same class as
ariadne#534. Fix the pipeline so an early-exiting consumer can't fail it (read the full output, or don't let a
consumer exit early), and sweep the other merge-checks for the same shape.

## Done when

- The check passes locally with dev-aliases.sh present, with a test that reproduces the 141 on the old pipeline.
- The other merge-checks are swept for early-exit consumers under pipefail (or guarded by an existing lint, if
  ariadne#534 added one, extended to these scripts).

## Plan

- [ ]

## Log

### 2026-10-25
