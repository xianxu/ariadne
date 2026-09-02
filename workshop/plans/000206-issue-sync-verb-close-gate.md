---
gate: boundary-review
issue: 206
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-02T12:23:35-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: change-code's auto-detect mode (no --issue, no --name) no longer commits the untracked issue file
          detail: |-
            syncIssue (changecode.go:196) returns early on f.Issue <= 0, but the helper it
            replaced also fired in resolveBranchName's case-3 auto-detect mode, where Issue is 0.
            Verified against base 92bd1ad (file committed, tree clean) vs head f1deae3 (left
            "?? workshop/issues/..."); with --worktree=yes the created worktree contains no issue
            file at all. Derive the id from the already-resolved issuePath via issueIDPrefix
            instead of gating on the flag, and pin it with an auto-detect-mode test.
          family: helper-swap-drops-a-mode
          round: 1
        - id: BR-2
          severity: Important
          title: migrate.go is an unswept member of the narrowed-add/bare-commit class
          detail: |-
            migrate stages precisely (migrate.go:327, :333) then commits bare in both repos
            (:343, :346). The source side has no whole-repo cleanliness guard — step (1) at :241
            checks only the migrated path — so a peer's staged work is swept into
            "migrate: move X to Y". The dest-side guard is bypassable and its --no-clean-check
            message claims "staging is explicit-path", the very property the bare commit removes.
            The --no-commit hints at :340-341 print the defective form too.
          family: narrowed-add-bare-commit
          round: 1
        - id: BR-3
          severity: Important
          title: archiveCommitArgs silently emits a whole-index commit when moves is empty
          detail: |-
            archiveCommitArgs(msg, nil) returns ["commit","-m",msg,"--"], and git treats a bare
            "--" with no paths as no pathspec at all — confirmed against real git. All three call
            sites guard len(moves) > 0, so nothing breaks today, but the helper whose purpose is
            "never commit the whole index" does exactly that on empty input, untested. Its sibling
            syncPathspec errors instead of widening; make this consistent and add a table row.
          family: narrowed-add-bare-commit
          round: 1
        - id: BR-4
          severity: Important
          title: the mid-planning trigger for `sdlc issue sync` has no delivery point in the agent's path
          detail: |-
            The only place telling an agent to run the verb mid-planning is helptext/issue.md:67-72,
            reachable only via `sdlc issue sync --help`. start-plan — the verb that opens planning and
            already points at where to author the plan — says nothing, and neither does AGENTS.md §2 or
            §14. change-code covers the end-of-planning trigger automatically; the mid-planning half,
            which is why --push defaults off, ships as documentation nobody is routed to (ARCH-PURPOSE).
          family: verb-without-a-trigger
          round: 1
        - id: BR-5
          severity: Minor
          title: syncIssue's guard comment names only --name mode, not the auto-detect mode it also catches
          family: comment-narrower-than-code
          round: 1
        - id: BR-6
          severity: Minor
          title: syncIssue does not thread f.DryRun into claimFlags; safe only because the dry-run branch returns earlier
          family: latent-flag-not-threaded
          round: 1
        - id: BR-7
          severity: Minor
          title: merge.go:550's archive commit has no direct regression test (recorded choice, not an oversight)
          family: class-fix-without-class-test
          round: 1
        - id: BR-8
          severity: Minor
          title: no coverage for `sdlc issue sync --dry-run` on either arm
          family: class-fix-without-class-test
          round: 1
        - id: BR-9
          severity: Minor
          title: a pathspec'd commit is a partial commit, which git refuses while MERGE_HEAD is set
          detail: |-
            Confirmed: "fatal: cannot do a partial commit during a merge". `sdlc issue sync` run
            mid-merge now dies where the bare commit would have created a merge commit. Almost
            certainly the better behavior, but it is a new failure mode and the atlas entry
            doesn't mention it.
          family: undocumented-behavior-change
          round: 1
        - id: BR-10
          severity: Minor
          title: no durable plan in workshop/plans/ for a 9-file, ~900-line change (AGENTS.md §1)
          family: durable-plan-artifact
          round: 1
        - id: BR-11
          severity: Minor
          title: '"the files for issue N" is computed three ways (syncPathspec, resolveBranchName, changedIssueFiles)'
          family: shared-issue-file-lookup
          round: 1
        - id: BR-12
          severity: Minor
          title: no declared operating envelope for a verb designed to be run frequently (ARCH-CONSTRAINTS)
          family: operating-envelope-undeclared
          round: 1
      blocked: true
---

# Gate ledger — ariadne#206 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T12:23:35-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `helper-swap-drops-a-mode` change-code's auto-detect mode (no --issue, no --name) no longer commits the untracked issue file
  syncIssue (changecode.go:196) returns early on f.Issue <= 0, but the helper it
  replaced also fired in resolveBranchName's case-3 auto-detect mode, where Issue is 0.
  Verified against base 92bd1ad (file committed, tree clean) vs head f1deae3 (left
  "?? workshop/issues/..."); with --worktree=yes the created worktree contains no issue
  file at all. Derive the id from the already-resolved issuePath via issueIDPrefix
  instead of gating on the flag, and pin it with an auto-detect-mode test.
- **BR-2** [Important] `narrowed-add-bare-commit` migrate.go is an unswept member of the narrowed-add/bare-commit class
  migrate stages precisely (migrate.go:327, :333) then commits bare in both repos
  (:343, :346). The source side has no whole-repo cleanliness guard — step (1) at :241
  checks only the migrated path — so a peer's staged work is swept into
  "migrate: move X to Y". The dest-side guard is bypassable and its --no-clean-check
  message claims "staging is explicit-path", the very property the bare commit removes.
  The --no-commit hints at :340-341 print the defective form too.
- **BR-3** [Important] `narrowed-add-bare-commit` archiveCommitArgs silently emits a whole-index commit when moves is empty
  archiveCommitArgs(msg, nil) returns ["commit","-m",msg,"--"], and git treats a bare
  "--" with no paths as no pathspec at all — confirmed against real git. All three call
  sites guard len(moves) > 0, so nothing breaks today, but the helper whose purpose is
  "never commit the whole index" does exactly that on empty input, untested. Its sibling
  syncPathspec errors instead of widening; make this consistent and add a table row.
- **BR-4** [Important] `verb-without-a-trigger` the mid-planning trigger for `sdlc issue sync` has no delivery point in the agent's path
  The only place telling an agent to run the verb mid-planning is helptext/issue.md:67-72,
  reachable only via `sdlc issue sync --help`. start-plan — the verb that opens planning and
  already points at where to author the plan — says nothing, and neither does AGENTS.md §2 or
  §14. change-code covers the end-of-planning trigger automatically; the mid-planning half,
  which is why --push defaults off, ships as documentation nobody is routed to (ARCH-PURPOSE).
- **BR-5** [Minor] `comment-narrower-than-code` syncIssue's guard comment names only --name mode, not the auto-detect mode it also catches
- **BR-6** [Minor] `latent-flag-not-threaded` syncIssue does not thread f.DryRun into claimFlags; safe only because the dry-run branch returns earlier
- **BR-7** [Minor] `class-fix-without-class-test` merge.go:550's archive commit has no direct regression test (recorded choice, not an oversight)
- **BR-8** [Minor] `class-fix-without-class-test` no coverage for `sdlc issue sync --dry-run` on either arm
- **BR-9** [Minor] `undocumented-behavior-change` a pathspec'd commit is a partial commit, which git refuses while MERGE_HEAD is set
  Confirmed: "fatal: cannot do a partial commit during a merge". `sdlc issue sync` run
  mid-merge now dies where the bare commit would have created a merge commit. Almost
  certainly the better behavior, but it is a new failure mode and the atlas entry
  doesn't mention it.
- **BR-10** [Minor] `durable-plan-artifact` no durable plan in workshop/plans/ for a 9-file, ~900-line change (AGENTS.md §1)
- **BR-11** [Minor] `shared-issue-file-lookup` "the files for issue N" is computed three ways (syncPathspec, resolveBranchName, changedIssueFiles)
- **BR-12** [Minor] `operating-envelope-undeclared` no declared operating envelope for a verb designed to be run frequently (ARCH-CONSTRAINTS)

## Open findings

- **BR-1** [Critical] `helper-swap-drops-a-mode` change-code's auto-detect mode (no --issue, no --name) no longer commits the untracked issue file
- **BR-2** [Important] `narrowed-add-bare-commit` migrate.go is an unswept member of the narrowed-add/bare-commit class
- **BR-3** [Important] `narrowed-add-bare-commit` archiveCommitArgs silently emits a whole-index commit when moves is empty
- **BR-4** [Important] `verb-without-a-trigger` the mid-planning trigger for `sdlc issue sync` has no delivery point in the agent's path
- **BR-5** [Minor] `comment-narrower-than-code` syncIssue's guard comment names only --name mode, not the auto-detect mode it also catches
- **BR-6** [Minor] `latent-flag-not-threaded` syncIssue does not thread f.DryRun into claimFlags; safe only because the dry-run branch returns earlier
- **BR-7** [Minor] `class-fix-without-class-test` merge.go:550's archive commit has no direct regression test (recorded choice, not an oversight)
- **BR-8** [Minor] `class-fix-without-class-test` no coverage for `sdlc issue sync --dry-run` on either arm
- **BR-9** [Minor] `undocumented-behavior-change` a pathspec'd commit is a partial commit, which git refuses while MERGE_HEAD is set
- **BR-10** [Minor] `durable-plan-artifact` no durable plan in workshop/plans/ for a 9-file, ~900-line change (AGENTS.md §1)
- **BR-11** [Minor] `shared-issue-file-lookup` "the files for issue N" is computed three ways (syncPathspec, resolveBranchName, changedIssueFiles)
- **BR-12** [Minor] `operating-envelope-undeclared` no declared operating envelope for a verb designed to be run frequently (ARCH-CONSTRAINTS)
