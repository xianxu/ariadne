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
    - "n": 2
      timestamp: "2026-10-12T14:31:20-07:00"
      dispatched: "2026-10-12T14:27:24-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Relative operands, --git-dir/--work-tree (both forms) and GIT_DIR/GIT_WORK_TREE are now checked; reverting the check turns TestGitShimRefusesRepositoriesOutside and TestGitShimChecksGitDirEnv red; a live run against real git refused ../out/rel and a relative GIT_DIR.
          round: 2
        - id: BR-2
          disposition: addressed
          note: claudeEnvFile re-exports every Env() var after sourcing the inherited file; TestClaudeEnvFileSourcesAnInheritedOne sets an inherited TMPDIR and asserts ours wins.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TakeSnapshot returns an error for an unknown or unlistable home, Confine and Check propagate it, and TestTakeSnapshotFailsWhenHomeCannotBeListed covers it.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The plan Cause section is marked superseded and points to Revisions; the incidentScript comment in judge/confine_unix_test.go now states the mktemp/var-folders cause.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Confine now calls the reviewbox gitOutput helper next to git(); it takes no ctx, which is acceptable for one rev-parse at setup.
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: Git shim guards only init/clone; worktree add and alias-expanded init still create repos outside
          detail: 'This is the 2nd finding in family confinement-shim-path-coverage. Verified live against real git: `git worktree add $OUT/wt` and `git -c alias.mk=init mk $OUT/al` both succeeded outside the root, and Check misses them unless they land as a new top-level entry in $HOME or an ancestor of the root or base. Rule: the shim guards the class "git commands that create a repository or working tree at a path" (init, clone, worktree add, and any alias that expands to one of them, by refusing -c alias.* and config-defined aliases of those), not a hand-kept list of two commands. Add test cases for each.'
          family: confinement-shim-path-coverage
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 717a39a265dd89af4772469f88326bafea25b1ec
      blocked: false
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

## Round 2 — 2026-10-12T14:31:20-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Relative operands, --git-dir/--work-tree (both forms) and GIT_DIR/GIT_WORK_TREE are now checked; reverting the check turns TestGitShimRefusesRepositoriesOutside and TestGitShimChecksGitDirEnv red; a live run against real git refused ../out/rel and a relative GIT_DIR.
- BR-2 — addressed — claudeEnvFile re-exports every Env() var after sourcing the inherited file; TestClaudeEnvFileSourcesAnInheritedOne sets an inherited TMPDIR and asserts ours wins.
- BR-3 — addressed — TakeSnapshot returns an error for an unknown or unlistable home, Confine and Check propagate it, and TestTakeSnapshotFailsWhenHomeCannotBeListed covers it.
- BR-4 — addressed — The plan Cause section is marked superseded and points to Revisions; the incidentScript comment in judge/confine_unix_test.go now states the mktemp/var-folders cause.
- BR-5 — addressed — Confine now calls the reviewbox gitOutput helper next to git(); it takes no ctx, which is acceptable for one rev-parse at setup.

### Raised

- **BR-6** [Minor] `confinement-shim-path-coverage` Git shim guards only init/clone; worktree add and alias-expanded init still create repos outside
  This is the 2nd finding in family confinement-shim-path-coverage. Verified live against real git: `git worktree add $OUT/wt` and `git -c alias.mk=init mk $OUT/al` both succeeded outside the root, and Check misses them unless they land as a new top-level entry in $HOME or an ancestor of the root or base. Rule: the shim guards the class "git commands that create a repository or working tree at a path" (init, clone, worktree add, and any alias that expands to one of them, by refusing -c alias.* and config-defined aliases of those), not a hand-kept list of two commands. Add test cases for each.

## Open findings

- **BR-6** [Minor] `confinement-shim-path-coverage` Git shim guards only init/clone; worktree add and alias-expanded init still create repos outside
