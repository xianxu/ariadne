---
gate: boundary-review
issue: 267
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T22:35:50-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Upstream warning guesses the cause; git's stderr and non-exit-1 config errors are discarded
          detail: The Plan says "warn with the cause instead of dropping git's stderr", but runGitCmd still drops stderr on success, and `upstream, _ :=` at landing.go:562 discards a real config error, which then shows up as the sandbox guess. Only instance in this window.
          family: success-path-stderr-dropped
          round: 1
        - id: BR-2
          severity: Minor
          title: pr --dry-run still announces gh pr create when an open PR would be updated
          detail: landing.go:525 returns before the PR query, so the dry run no longer matches the real path's update branch. Only instance in this window.
          family: dry-run-mirrors-real-path
          round: 1
        - id: BR-3
          severity: Minor
          title: Upstream warning is tested only in its fires state
          detail: No test asserts that stderr has no warning when branch.<b>.merge was recorded (TestLandingPRUpdatesOpenPR discards stderr), so a warning that fires every time would go unnoticed.
          family: two-state-clause-coverage
          round: 1
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#267 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T22:35:50-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `success-path-stderr-dropped` Upstream warning guesses the cause; git's stderr and non-exit-1 config errors are discarded
  The Plan says "warn with the cause instead of dropping git's stderr", but runGitCmd still drops stderr on success, and `upstream, _ :=` at landing.go:562 discards a real config error, which then shows up as the sandbox guess. Only instance in this window.
- **BR-2** [Minor] `dry-run-mirrors-real-path` pr --dry-run still announces gh pr create when an open PR would be updated
  landing.go:525 returns before the PR query, so the dry run no longer matches the real path's update branch. Only instance in this window.
- **BR-3** [Minor] `two-state-clause-coverage` Upstream warning is tested only in its fires state
  No test asserts that stderr has no warning when branch.<b>.merge was recorded (TestLandingPRUpdatesOpenPR discards stderr), so a warning that fires every time would go unnoticed.

## Open findings

- **BR-1** [Minor] `success-path-stderr-dropped` Upstream warning guesses the cause; git's stderr and non-exit-1 config errors are discarded
- **BR-2** [Minor] `dry-run-mirrors-real-path` pr --dry-run still announces gh pr create when an open PR would be updated
- **BR-3** [Minor] `two-state-clause-coverage` Upstream warning is tested only in its fires state
