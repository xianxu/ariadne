---
gate: boundary-review
issue: 339
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T14:18:34-07:00"
      dispatched: "2026-10-12T14:13:23-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: git shim lets relative-path and --git-dir init create repositories outside, unseen by Check
          detail: 'confine.go gitShimTemplate skips relative operands after init/clone and leading --git-dir/--work-tree values; verified live: `git init ../../../../../../relesc` from $TMPDIR/a/b created ~/Library/Caches/sdlc/relesc/.git and `git --git-dir=<review-checkouts>/escgd init` created a repo there, neither in an ancestor or $HOME so Check passes. Resolve relative args against $dir, check --git-dir/--work-tree (both forms) and GIT_DIR, add shim test cases.'
          family: confinement-shim-path-coverage
          round: 1
        - id: BR-2
          severity: Minor
          title: env.sh sources the inherited CLAUDE_ENV_FILE but re-exports only PATH, so an inherited TMPDIR wins
          detail: claudeEnvFile (confine.go:133) should re-export TMPDIR/TMP/TEMP/GIT_CEILING_DIRECTORIES after sourcing the inherited file.
          family: confinement-env-reassert
          round: 1
        - id: BR-3
          severity: Minor
          title: TakeSnapshot ignores ReadDir(home) errors, making the home check a silent no-op
          family: silent-degraded-check
          round: 1
        - id: BR-4
          severity: Minor
          title: Plan "Cause (for the Log)" section and incidentScript comment still state the superseded CLAUDE_CODE_TMPDIR cause; plan task boxes unchecked
          family: plan-stale-after-revision
          round: 1
        - id: BR-5
          severity: Minor
          title: Confine runs exec.Command("git","rev-parse") directly, without ctx and outside reviewbox's git helper
          family: external-call-outside-seam
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#339 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T14:18:34-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `confinement-shim-path-coverage` git shim lets relative-path and --git-dir init create repositories outside, unseen by Check
  confine.go gitShimTemplate skips relative operands after init/clone and leading --git-dir/--work-tree values; verified live: `git init ../../../../../../relesc` from $TMPDIR/a/b created ~/Library/Caches/sdlc/relesc/.git and `git --git-dir=<review-checkouts>/escgd init` created a repo there, neither in an ancestor or $HOME so Check passes. Resolve relative args against $dir, check --git-dir/--work-tree (both forms) and GIT_DIR, add shim test cases.
- **BR-2** [Minor] `confinement-env-reassert` env.sh sources the inherited CLAUDE_ENV_FILE but re-exports only PATH, so an inherited TMPDIR wins
  claudeEnvFile (confine.go:133) should re-export TMPDIR/TMP/TEMP/GIT_CEILING_DIRECTORIES after sourcing the inherited file.
- **BR-3** [Minor] `silent-degraded-check` TakeSnapshot ignores ReadDir(home) errors, making the home check a silent no-op
- **BR-4** [Minor] `plan-stale-after-revision` Plan "Cause (for the Log)" section and incidentScript comment still state the superseded CLAUDE_CODE_TMPDIR cause; plan task boxes unchecked
- **BR-5** [Minor] `external-call-outside-seam` Confine runs exec.Command("git","rev-parse") directly, without ctx and outside reviewbox's git helper

## Open findings

- **BR-1** [Important] `confinement-shim-path-coverage` git shim lets relative-path and --git-dir init create repositories outside, unseen by Check
- **BR-2** [Minor] `confinement-env-reassert` env.sh sources the inherited CLAUDE_ENV_FILE but re-exports only PATH, so an inherited TMPDIR wins
- **BR-3** [Minor] `silent-degraded-check` TakeSnapshot ignores ReadDir(home) errors, making the home check a silent no-op
- **BR-4** [Minor] `plan-stale-after-revision` Plan "Cause (for the Log)" section and incidentScript comment still state the superseded CLAUDE_CODE_TMPDIR cause; plan task boxes unchecked
- **BR-5** [Minor] `external-call-outside-seam` Confine runs exec.Command("git","rev-parse") directly, without ctx and outside reviewbox's git helper
