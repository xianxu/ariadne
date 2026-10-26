---
id: 000642
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '5d00a8c1cc621ae786fae409bee168e4a42c0dd4' # card fields mirrored from issue-cards; edit via sdlc
---

# Registered Lua flakes aren't attributed: a parley suite failing only on one is judged unattributable

## Problem

## Spec

**A registered Lua (busted) flake isn't attributed, so its failure is judged "unattributable" (parley.nvim:1 on
parley.nvim#446, 10-25).** The pre-close suite on 684a12e5 failed only on parley.nvim#426's registered flake
(`agentic_document_owned_move_spec`), yet sdlc judged the attempt (7757e7bc64ce4e990fee730e48b7b49c)
`unattributable-failure` and refused the resume, the same outcome as an unexplained failure. Go flakes are matched by
test name; the Lua suite's failure output evidently isn't parsed into the same named-failure form, so the registry
never matches it. ariadne#457 added non-Go suites to the registry; this looks like a gap in parsing busted/parley
output into named failures.

## Done when

- A Lua suite failure whose only failing case is a registered flake is attributed to it (known-flake handling), and
  an unregistered Lua failure stays unattributable; tests with captured parley suite output for both.

## Plan

- [ ]

## Log

### 2026-10-25
