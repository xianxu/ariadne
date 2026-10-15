---
id: 000398
status: open
deps: [393]
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '1186727c83360bd4ce1cf7048519a01978d6f90d' # card fields mirrored from issue-cards; edit via sdlc
---

# Proven suite reuse after doc-only main drift

## Problem

A main movement during `sdlc merge` normally requires re-integration and a fresh full suite even when only non-test prose moved. #390 treats a doc-only reuse path as a later optimization, not a reason to weaken the exact-head landing guarantee without proof.

## Spec

After serial admission (#393) has pilot data, measure how often main drift changes no effective suite input and what suite time it costs. Only if material, design a fail-closed equivalence proof for the tested and proposed landing trees: code, tests, generated sources, test commands, gate configuration, dependencies and other effective inputs must be identical. Unknown inputs require the full suite. The proof must preserve exact-head landing evidence by binding a prior passing run to an equivalent effective test tree and must not use filename extensions as authority. Follow #390's approved later-work option; do not bypass review, CI or local-gates checks.

## Done when

- [ ] A measured pilot shows whether doc-only drift causes enough repeat full suites to justify a fast path; otherwise close the issue with evidence and no gate change.
- [ ] If justified, a complete effective-input comparison is specified and tested against generated files, changed make targets, dependencies, config and misleading doc-like paths; unknowns rerun the suite.
- [ ] Merge still refuses a moved PR head and records why suite evidence was reused; local-gates changes get operator verification.

## Plan

- [ ] Measure drift after #393; decide whether to implement.
- [ ] If warranted, test and implement exact effective-input equivalence with fail-closed behavior.

## Log

### 2026-10-15

Filed as a conditional later optimization from #390, dependent on serial-admission measurements.
