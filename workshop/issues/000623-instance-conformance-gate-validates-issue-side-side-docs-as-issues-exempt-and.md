---
id: 000623
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '1d591f6810da0979668e9f82d2b23c95b318cfdf' # card fields mirrored from issue-cards; edit via sdlc
---

# Instance-conformance gate validates issue .side/ side docs as issues; exempt and validate by type

## Problem

## Spec

**The instance-conformance gate validates side docs under an issue's `.side/` folder as issues (parley.nvim:1, 10-25).**
parley.nvim#440's merge (PR #354) was refused: the gate checks every `.md` under `workshop/issues/`, including
`workshop/issues/000440-….side/upstream-neovim-draft.md`, a `type: side-doc` file, and demands an issue's status,
Spec, Plan and Done when. Merge's own archive step already moves an issue's `.side/` folder with it, so side docs
there are an intended layout; the gate is the inconsistency.

Exempt files under `<issue>.side/` from Issue validation, and validate them by their own `type:` (side-doc
frontmatter: type, parent) instead. Use the same notion of "this file belongs to issue N's side folder" that the
archive step uses (one shared helper, not a second path rule).

## Done when

- A repo with an issue `.side/` side doc passes the gate; a malformed side doc (missing type or parent) fails with a
  side-doc message; a malformed issue still fails as before. Tests for all three.
- The archive step and the gate share the side-folder rule.
- **Plan side, same gap:** merge's archive step refuses a `workshop/plans/<issue>-…-plan.side/` folder whose plan document doesn't exist ('active or conflicting archive artifact'), leaving it on main after the issue archives (parley.nvim#440: its upstream-neovim-draft.md stayed in workshop/plans while the archived issue's link points under history/plans). A side folder archives with its issue whether or not a plan document exists; test both. Then archive parley.nvim's leftover file.

## Plan

- [ ]

## Log

### 2026-10-25
