---
id: 000342
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '3d70e7faaf37670ed043c2f7d3e739bb791cbf83' # card fields mirrored from issue-cards; edit via sdlc
---

# agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR)

## Problem

On macOS, `mktemp` / `mktemp -d` / `mktemp -t x` without a path template ignore `TMPDIR` and create under the per-user `/var/folders/.../T`, which the Claude Code sandbox denies (measured by pair:7, 2026-10-12; the same mkdtemp failure appears in transcripts since 2026-09-11 across ariadne, tools, pair, brain, parley.nvim and slots). A probe script that then `cd "$d"` with an empty `$d` lands in HOME: on 10-12 that wrote `~/Makefile` and `~/inc.mk` (ariadne#299's probe), the second home write that day (cf. #339).

## Spec

Fix it where agents and scripts learn and run:
1. **AGENTS rule** (in the shared base, so every repo's agents get it): make temp dirs with `mktemp -d "${TMPDIR:-/tmp}/name.XXXXXX"`, and never `cd` (or write) into a path built from a variable that may be empty: use `"${d:?}"` or `|| exit 1`.
2. **Sweep this repo's scripts** for template-less `mktemp` and fix them the same way; a test (or lint check in an existing guard) that refuses a template-less `mktemp` in tracked shell scripts.
3. No settings or machine change: allowing `/var/folders/.../T` in the sandbox would put agent churn back on the journaled Data volume (pair#496).

## Done when

- The AGENTS rule is in the base and reaches peers through weave.
- No tracked script in ariadne uses a template-less `mktemp`, and the check fails on one.

## Plan

- [ ]

## Log

### 2026-10-12
