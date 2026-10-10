---
id: 000316
status: codecomplete
deps: []
github_issue:
target: review-convention
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '92ce69dc38f0a05e98c4088872024bf266ced8e6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-09T21:20:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    worktree: /Users/xianxu/workspace/worktree/parley.nvim-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "8ebc71b7", done: "b55b6c6a"}
actual_hours: N/A
---

# review-convention: single-line markers + <br> newline escape

## Problem

parley.nvim#312 renders 🤖 markers compactly (chain concealed, anchor
highlighted) and opens a chain in a thread float. That rendering is line-local:
a marker that spans lines can't be concealed cleanly and the incremental
re-render depends on markers being line-local. Agents (xx-fix, Claude Code,
pair) also write markers, so the rule has to live in the shared grammar or
agent-written markers keep spanning lines.

## Spec

Revise `construct/local/fix/review-convention.md`:
- §3: a marker is a single line; a newline inside a `[]`/`{}` turn is written
  `<br>` (a literal `<br>` is `\<br>`); anchors (`<X>`, `~D~`) quote one line
  verbatim and are never encoded. (Supersedes the #125-era multi-line allowance
  for writers; readers may still tolerate legacy multi-line markers.)
- §5: resolution decodes `<br>` (and `\<br>`) in turn text only; anchor text
  is restored verbatim.
- xx-fix `SKILL.md`: the one-line writing rule and the decode scope on both the
  per-marker (step 4) and bulk resolution paths — xx-fix writes and resolves
  markers too.
- `## Revisions` entry naming parley.nvim#312.

## Done when

- The target carries the single-line rule, `<br>` for line breaks inside
  turns and `\<br>` for a literal `<br>`, anchors verbatim, turn-only decoding
  on resolve, and a Revisions entry.
- xx-fix `SKILL.md` writes single-line markers and decodes turn text on both
  its per-marker and bulk resolution paths.
- parley's weaved copy picks both up on the next weave.

## Plan

- [x] Edit §3 / §5 + Revisions in `construct/local/fix/review-convention.md`
- [x] Carry the one-line rule + decode scope into `construct/local/fix/SKILL.md`

## Log

### 2026-10-09
- 2026-10-09: closed — review-convention.md §3 "One line" (single-line markers; <br> line breaks inside []/{} turns, \<br> for a literal <br>; anchors verbatim; legacy multi-line rendered broken; Alt+q refuses multi-line), §5 decodes turn text only, updated + Revisions (parley.nvim#312, whose codec implements and tests the escape). xx-fix SKILL.md: one-line rule + decode scope on step 4 and bulk paths. Issue Spec/Done-when/Log/Revisions synced. Docs-only. No actual: authored inside the parley.nvim#312 session.; review verdict: FIX-THEN-SHIP
- Close review round 1: BR-1 (decoding every resolved result would corrupt an
  anchor's verbatim prose, e.g. table `<br>`) → decode turn text only; BR-2
  (xx-fix writes markers too) → SKILL.md carries the rule.
- Round 2: step 4 of xx-fix also decodes; grammar paragraph made tool-neutral;
  `updated:` bumped.
- Round 3: BR-6 — no literal `<br>` in a turn (a table-row edit would split on
  accept) → `\<br>` escape; parley.nvim#312's codec implements it.
- Round 4 (post-close fixes): the decode rule is defined once in xx-fix's
  "One line" paragraph and both resolution sites point to it (the restated
  copies had drifted and missed `\<br>`); the Alt+q multi-line refusal moved
  to §4; the editor-specific "render it as broken" wording dropped.


## Revisions

- **2026-10-09** — scope grew from the target file alone to the target + xx-fix
  `SKILL.md` (close review BR-2), and the grammar gained the `\<br>` escape and
  the anchors-verbatim rule (BR-1, BR-6).
