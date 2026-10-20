---
id: 000479
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '40868def9f8ef79ada001b708b546efc2f2c674d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T20:21:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c6b0c4ca", done: "490ac2c8"}
---

# fixedIssueCovers matches flake_tests exactly; route it through flakes.Covers (subtest-aware, from #466 review)

## Problem

From ariadne#466's close review (Minor, fails safe): `fixedIssueCovers` (cmd/sdlc/flake_fixed.go:104, from #458)
matches a fixed issue's `flake_tests` entries against a failure row by exact string. Since #466 lets `flake_tests`
name exact subtests, a fixed issue registered on a subtest never hides or labels a later failure reported on its
top-level test (or the reverse), so a fixed flake can resurface as an unlabeled failure.

## Spec

Route `fixedIssueCovers` through `flakes.Covers`, the same subtest-aware matcher #466 introduced for known flakes, so
fixed and known flakes use one coverage rule (ARCH-DRY).

## Done when

- A fixed issue registered on `TestX/sub` covers a failure row naming `TestX/sub` and is handled by the same rule as
  a known flake for `TestX` rows (table test mirroring #466's Covers tests).
- No other caller keeps an exact-string match for flake coverage (grep check in the test or a lint).

## Plan

Quick flow. The class (ARCH-PURPOSE): every flake-coverage decision derives from `flakes.Covers` applied to a
failure's leaves (`explain`). `flake list`'s fixed-issue path is the one place still matching the top-level failure
name exactly.

- [ ] `flakes`: export the gate's leaf split — `Unknown(failures, subtests, reg) []string`, the failing leaves no
  registered name covers; `Explained` becomes `len(Unknown)==0` (ARCH-DRY: one coverage rule).
- [ ] `classifyFixedFlake` takes a leaf and picks the fixed issues whose `flake_tests` cover it via `flakes.Covers`,
  nearest name first (as `Registry.Cover`); two issues at the same nearest name stay ambiguous → unknown. Delete
  `fixedIssueCovers`.
- [ ] `runFlakeListMode`: for a confirmed attempt, each uncovered leaf (`flakes.Unknown`) is classified and listed by
  its own name — the name to register — instead of the top-level test.
- [ ] Tests: table for leaf coverage (fixed on `TestX/sub` covers `TestX/sub` and `TestX/sub/deeper`, not
  `TestX/sib` nor `TestX`; fixed on `TestX` covers `TestX/sub`; nearest wins; same-name tie unknown); a `flake list`
  test with a logged subtest fixed by a terminal card; an AST check that no `cmd/sdlc` function besides
  `flakes.Covers` is a `*Covers` predicate.

## Log

### 2026-10-19
