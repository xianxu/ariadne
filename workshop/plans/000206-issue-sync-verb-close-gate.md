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
    - "n": 2
      timestamp: "2026-09-02T12:51:50-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Verified by revert: re-adding the f.Issue<=0 gate turns TestChangeCodeSyncIssue_SyncsWithoutTheIssueFlag red.'
          round: 2
        - id: BR-2
          disposition: addressed
          note: migrate.go:356/:359 and both --no-commit hints now carry the pathspec; the fix itself is unpinned — see the new class finding.
          round: 2
        - id: BR-3
          disposition: addressed
          note: archiveCommitArgs errors on empty moves and TestArchiveCommitArgs_RefusesEmptyMoves asserts it directly.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Confirmed by running the built binary; start-plan prints the checkpoint line, and AGENTS.base.md 2/14 carry it. No test pins it — folded into the class finding.
          round: 2
        - id: BR-5
          disposition: addressed
          note: The guard comment now names the only case the guard can catch (a --name branch at a non-NNNNNN- path), and TestChangeCodeSyncIssue_NonIssuePathIsANoop pins the behavior.
          round: 2
        - id: BR-6
          disposition: addressed
          note: 'Verified by revert: unthreading DryRun turns TestChangeCodeSyncIssue_DryRunCommitsNothing red.'
          round: 2
        - id: BR-7
          disposition: not-addressed
          note: Still no regression at merge.go:550 — reverting it leaves the suite green; now part of the measured class finding rather than a standalone recorded choice.
          round: 2
        - id: BR-8
          disposition: not-addressed
          note: The in-place arm's --dry-run is now covered at both the verb and change-code, but syncViaMainWorktree's dry-run early return (newly reachable via `issue sync --push --dry-run`) still has none.
          round: 2
        - id: BR-9
          disposition: addressed
          note: The atlas sdlc-binary entry now records the partial-commit-during-merge failure mode explicitly.
          round: 2
        - id: BR-10
          disposition: not-addressed
          note: No durable plan; the Revisions entry records a considered refusal (the design lives in Spec + Revisions, which is what the plan-quality gate read). Minor, does not block.
          round: 2
        - id: BR-11
          disposition: addressed
          note: issueFilesForID single-sources the glob for resolveBranchName and syncPathspec; issueIDFromPath reuses issueIDPrefix.
          round: 2
        - id: BR-12
          disposition: not-addressed
          note: Not touched in the rework round and not mentioned in Revisions; still no envelope for a verb designed to run frequently.
          round: 2
      findings:
        - id: BR-13
          severity: Important
          title: this round's call-site fixes revert green - 5 of 7 pathspec sites and the start-plan delivery have no pinning test
          detail: |-
            Measured by revert in a scratch clone at 0807e39, full package suite each time: reverting
            migrate.go:356/:359, push.go:204/:404, merge.go:550, or the syncPointer call at
            startplan.go:85 leaves the suite green. Cause is uniform - issuesync_test.go:593 hand-writes
            migrate's commit argv instead of calling runMigrate, and issuesync_test.go:540-556 hand-wires
            archiveCommitArgs instead of driving runPush/recoverInterruptedArchive. Both run real git and
            neither runs our code, so they pass in exactly the state the finding was about (ARCH-MOCK
            inverted). This is the 3rd finding in the family, so do NOT patch the instances - the rule is
            that a call-site fix is pinned only by a test driving the production entry point, and the
            enforcement that covers the whole class is a source-level guard test over cmd/sdlc asserting
            every git-commit argv built after a narrowed add carries a `--`, allowlisting the two
            deliberate whole-tree commits (push.go:116 `commit -a`, propagatebase.go:271 after `add -A`).
            Same shape as TestRepoLockCommandMetadata / TestForceAckMatchesGateCatalog. Also add
            "sdlc issue sync" to TestRunStartPlan_RendersAtPlanLens's want list - that test exists so a
            dropped pointer line cannot ship silently, and it just failed to do that.
          family: class-fix-without-class-test
          round: 2
        - id: BR-14
          severity: Important
          title: change-code's sync from a feature worktree publishes WIP to origin/main and leaves the branch copy dirty
          detail: |-
            Probed against a real two-worktree fixture: syncIssue (changecode.go:213) on a feature
            worktree with a tracked-and-edited issue file takes syncViaMainWorktree, pulls, copies,
            commits "#206: issue-sync: spec/plan at change-code" on main, pushes - and leaves
            "M workshop/issues/000206-issue-sync-verb.md" on the branch. The "branch starts from a
            tracked state" property asserted at changecode.go:167-176 is not achieved in that mode, main
            gains a divergent commit of a file the branch will later commit its own version of, and a
            milestone re-run (#156 made worktree creation idempotent for exactly this) gains two network
            round-trips. commitUntrackedIssueFile no-opped here. --worktree=yes is opt-in, hence Important
            not Critical. This is the 2nd finding in the family, so do NOT patch this cell - write the
            mode cross-product the old helper ran under (resolveBranchName's 3 name modes x {on main,
            in-place branch, feature worktree} x {untracked, tracked-and-edited}) as a test table and
            state the intended behavior for each. Round 1 fixed one cell; this round covers four; the
            cells whose behavior actually changed are uncovered.
          family: helper-swap-drops-a-mode
          round: 2
        - id: BR-15
          severity: Minor
          title: syncPathspec is documented as an argv builder but does filesystem IO via filepath.Glob
          detail: |-
            claim.go:235 reads as a pure pathspec constructor; issueFilesForID globs the working tree, so
            it is a thin IO helper. Worth saying so before someone table-tests it as pure (ARCH-PURE).
          family: comment-narrower-than-code
          round: 2
        - id: BR-16
          severity: Minor
          title: migrate.go restates the pathspec'd-commit shape four times inline and its hint assertion does not check the pathspec
          detail: |-
            migrate.go:352-359 builds the same `commit -m <msg> -- <path>` shape for two commits and two
            --no-commit hints; migrate_test.go:449 asserts only the "--no-commit: both sides staged"
            lead-in, so the hint could lose its `--` unnoticed. A two-line helper mirroring
            archiveCommitArgs would keep the printed command and the executed one from drifting.
          family: narrowed-add-bare-commit
          round: 2
        - id: BR-17
          severity: Minor
          title: pre-existing red test in the tree this boundary crosses - fleet_plan_test.go reads an archived plan path
          detail: |-
            TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (fleet_plan_test.go:14) opens
            workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, which dfeba9c archived to history.
            dfeba9c is an ancestor of the review base, so this is not from #206 - but it means the suite is
            red at the close boundary. Worth its own issue; the test should resolve the plan through the
            archive-inclusive lookup rather than a hardcoded workshop/plans path.
          family: stale-test-fixture-path
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-02T13:21:22-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: merge.go:550 is now caught by the source guard — verified by revert (guard fires, naming merge.go:runMerge).
          round: 3
        - id: BR-8
          disposition: addressed
          note: In-place arm covered by TestIssueSync_DryRunCommitsNothing + the change-code variant; syncViaMainWorktree's dry-run early return remains uncovered.
          round: 3
        - id: BR-10
          disposition: not-addressed
          note: Deliberately declined with a recorded rationale; the two new workshop/plans/ files are gate ledgers, not a durable plan. Minor, non-blocking; the rule belongs to ariadne#198.
          round: 3
        - id: BR-12
          disposition: not-addressed
          note: No envelope declared. The no-push default is the decision; it just is not stated as a budget. ariadne#205 owns the class.
          round: 3
        - id: BR-13
          disposition: not-addressed
          note: 6 of 7 now caught (revert-verified); push.go:204 still reverts the full suite green because the allowlist is keyed by function and runPush is exempt for its unrelated `commit -a`. Key the exemption to the exact argv literals, not the enclosing function.
          round: 3
        - id: BR-14
          disposition: addressed
          note: 'Revert-verified: reverting `NoPush: !onMain` reds 12 of the 18 matrix cells.'
          round: 3
        - id: BR-15
          disposition: addressed
          note: syncPathspec's doc now states it is a thin IO helper, not pure.
          round: 3
        - id: BR-16
          disposition: addressed
          note: migrateCommitArgs single-sources the shape; TestMigrateCommitArgs_HintAndCommandAreTheSameBuilder pins the hint's pathspec.
          round: 3
        - id: BR-17
          disposition: addressed
          note: Filed as ariadne#207. The suite is still red at this boundary (reproduced), so the close evidence should say so rather than claim a green suite.
          round: 3
      findings:
        - id: BR-18
          severity: Critical
          title: '`sdlc issue sync --push` cannot publish an already-committed body, and both new warnings name it as the recovery step'
          detail: |-
            claim.go:186-189 returns `cok("No issue changes to sync.")` before the add/commit/push
            whenever changedIssueFiles is empty. Since this issue split durability from publication,
            "committed locally, push failed" is a state the code deliberately creates — and in it
            --push short-circuits and never pushes, reporting success in green. Measured on a real
            repo + bare origin: local sync, then Push:true → `[ok] No issue changes to sync.`,
            origin/main unmoved at 3f20fb4 while local main is 4b2329f. changecode.go:250-254 and
            issue.go:331-332 both instruct the operator to run exactly that command to finish the
            failed publish; for `issue new` that is the #82 M1 reservation broadcast, left
            unrecoverable by the named route. Fix: when !NoPush, fall through to the push (or gate on
            `rev-list --count origin/main..main`) rather than returning on an empty working-tree diff.
          family: early-return-skips-second-concern
          round: 3
        - id: BR-19
          severity: Important
          title: runChangeCode's syncIssue call site enters no test — deleting the line leaves the suite green
          detail: |-
            Measured: replacing `syncIssue(stderr, f, issuePath)` at changecode.go:179 with a no-op
            leaves the full cmd/sdlc suite green apart from the known-unrelated fleet_plan failure.
            Every TestChangeCodeSyncIssue_* test calls syncIssue directly with a literal path, so they
            prove the helper and mock the wiring — BR-13's exact shape, on the Spec's third piece.
            Structural cause is tracked as ariadne#191 (runChangeCode is not in-process drivable).
            This is the 4th finding in family `class-fix-without-class-test`; do NOT pin this one site.
            The rule: a call-site fix is pinned only by a test entering the production entry point, and
            where that entry point is not drivable a source-level guard must assert the wiring. Enforce
            it once — a wiring table in commitpathspec_guard_test.go's shape covering
            runChangeCode→syncIssue, runPush/runMerge→archiveCommitArgs, runMigrate→migrateCommitArgs —
            or land ariadne#191. Measured prevalence: 5 sites in round 2, plus push.go:204 and this one
            still open in round 3.
          family: class-fix-without-class-test
          round: 3
        - id: BR-20
          severity: Minor
          title: atlas says "Six sites" over an enumeration of seven
          detail: |-
            atlas/workflow/sdlc-binary.md:111 — "both sync arms, both push.go archive commits,
            merge.go's, and migrate.go's two" is seven, and commitpathspec_guard_test.go's own comment
            says seven. Derive the count from the list or drop it.
          family: prose-count-restates-enumeration
          round: 3
        - id: BR-21
          severity: Minor
          title: syncPathspec's "unreachable" rationale does not hold for a deleted issue file
          detail: |-
            claim.go:237 argues the "pathspec matches nothing" error is unreachable because callers run
            changedIssueFiles first — but a deleted issue file makes changedIssueFiles non-empty while
            issueFilesForID's glob is empty, so the argument does not cover that state (the outcome is
            still practically unreachable). This is the 3rd finding in family `comment-narrower-than-code`;
            the rule is that a comment asserting unreachability must enumerate the states its guard
            quantifies over — if that cannot be enforced, drop the claim and let the error stand alone.
          family: comment-narrower-than-code
          round: 3
      blocked: true
    - "n": 4
      timestamp: "2026-09-02T13:39:24-07:00"
      agent: claude
      dispose:
        - id: BR-10
          disposition: not-addressed
          note: No workshop/plans/000206-*-plan.md; the two new plans/ files are the gate ledger and the review transcript. The recorded "back-filling would produce a copy" rationale stands as an operator call; Minor, non-blocking.
          round: 4
        - id: BR-12
          disposition: not-addressed
          note: Still no declared envelope, and round 3 added an unconditional network push on the no-changes path without pricing it.
          round: 4
        - id: BR-13
          disposition: addressed
          note: 'Verified by revert: four pathspec sites reverted at once produce four failures (incl. push.go:204 inside runPush); deleting syncPointer reds TestRunStartPlan_RendersAtPlanLens and the wiring guard.'
          round: 4
        - id: BR-18
          disposition: not-addressed
          note: 'syncInPlace is fixed and pinned (both tests red on revert), but claim.go:302-320 pushes origin main without carrying the body: from a feature worktree, `sdlc issue sync --issue N --push` after a local sync prints `[ok] Issues synced to main and pushed to origin.` with origin/main unmoved (measured, incl. via `issue new`''s advertised recovery, where origin/main kept only README.md). Same finding, second arm; that branch has no test.'
          round: 4
        - id: BR-19
          disposition: addressed
          note: 'Measured: replacing syncIssue(stderr, f, issuePath) at changecode.go:179 with a no-op fails TestVerbsWireTheirCommitHelpers.'
          round: 4
        - id: BR-20
          disposition: addressed
          note: atlas/workflow/sdlc-binary.md now says "Seven sites" over a seven-member enumeration.
          round: 4
        - id: BR-21
          disposition: addressed
          note: syncPathspec's doc drops the unreachability claim and names the deleted-issue-file state that broke it.
          round: 4
      findings:
        - id: BR-22
          severity: Minor
          title: recoverInterruptedArchive's dry-run hint hand-builds the commit argv instead of using archiveCommitArgs
          detail: |-
            push.go:394-397 prints `git commit -m %q -- <paths>` assembled inline, deriving only the
            path list from archiveAddArgs. This is the 4th finding in family `narrowed-add-bare-commit`,
            so do NOT patch the line: the rule round 2 already stated for migrate is that ONE builder
            feeds both the executed commit and the command printed for the operator, so the hint cannot
            lose what the executed form has. push.go is the unswept member of that enumeration
            (migrate.go:426 migrateCommitArgs is the swept one). Measured prevalence: 2 hint sites in
            the tree, 1 derived.
          family: narrowed-add-bare-commit
          round: 4
        - id: BR-23
          severity: Minor
          title: the pathspec class guard's seam enumeration misses gitx.RunGit, so a commit through it escapes
          detail: |-
            commitpathspec_guard_test.go:157-172 accepts a git argv only from .Git / .GitInDir /
            gitInDir / exec.Command("git", ...), while its own doc claims every `git commit` argv in
            cmd/sdlc and "the eighth site someone writes next year fails immediately". gitx.RunGit is a
            live seam at ~12 sites in this package (all reads today) and gitx.Capture at more; an inline
            gitx.RunGit("commit", "-m", msg) passes the guard. This is the 5th finding in family
            `class-fix-without-class-test`, so the fix is the rule, not the line: the guard's seam list
            must cover every git-invoking helper the package actually has — add RunGit/Capture and state
            that adding a new git seam requires adding it here, in the same shape as the stale-exemption
            check the guard already performs on its allowlist.
          family: class-fix-without-class-test
          round: 4
        - id: BR-24
          severity: Minor
          title: '`sdlc claim` with no issue changes now needs the network and dies offline, where it used to be a clean no-op'
          detail: |-
            Measured: on main, clean tree, unreachable origin, `sdlc claim --issue 206` now fails through
            pushMain; at base 92bd1ad the same command printed `[ok] No issue changes to sync.` and exited
            0. Deliberate consequence of "nothing to commit is not nothing to publish", but claim is the
            one caller that die()s on the error. This is the 2nd finding in family
            `undocumented-behavior-change`; the rule the first one (MERGE_HEAD partial commits) already
            established is that a change which creates a NEW way an existing verb can fail gets recorded
            in the atlas failure-mode list — apply it here, and decide explicitly whether an idempotent
            offline claim should be fatal or a warning.
          family: undocumented-behavior-change
          round: 4
        - id: BR-25
          severity: Minor
          title: change-code --dry-run promises "Would sync + push" from branches where it will not push
          detail: |-
            changecode.go:162-164 prints `Would sync + push issue #%d` unconditionally, while syncIssue
            sets NoPush: !onMain, so from an in-place branch or a feature worktree the real run commits
            locally and does not publish. This is the 4th finding in family `comment-narrower-than-code`;
            the rule is that a statement about behavior must branch on the same state the code branches
            on — here, derive the dry-run wording from the same `onMain` test syncIssue uses rather than
            restating the outcome.
          family: comment-narrower-than-code
          round: 4
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

## Round 2 — 2026-09-02T12:51:50-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Verified by revert: re-adding the f.Issue<=0 gate turns TestChangeCodeSyncIssue_SyncsWithoutTheIssueFlag red.
- BR-2 — addressed — migrate.go:356/:359 and both --no-commit hints now carry the pathspec; the fix itself is unpinned — see the new class finding.
- BR-3 — addressed — archiveCommitArgs errors on empty moves and TestArchiveCommitArgs_RefusesEmptyMoves asserts it directly.
- BR-4 — addressed — Confirmed by running the built binary; start-plan prints the checkpoint line, and AGENTS.base.md 2/14 carry it. No test pins it — folded into the class finding.
- BR-5 — addressed — The guard comment now names the only case the guard can catch (a --name branch at a non-NNNNNN- path), and TestChangeCodeSyncIssue_NonIssuePathIsANoop pins the behavior.
- BR-6 — addressed — Verified by revert: unthreading DryRun turns TestChangeCodeSyncIssue_DryRunCommitsNothing red.
- BR-7 — not-addressed — Still no regression at merge.go:550 — reverting it leaves the suite green; now part of the measured class finding rather than a standalone recorded choice.
- BR-8 — not-addressed — The in-place arm's --dry-run is now covered at both the verb and change-code, but syncViaMainWorktree's dry-run early return (newly reachable via `issue sync --push --dry-run`) still has none.
- BR-9 — addressed — The atlas sdlc-binary entry now records the partial-commit-during-merge failure mode explicitly.
- BR-10 — not-addressed — No durable plan; the Revisions entry records a considered refusal (the design lives in Spec + Revisions, which is what the plan-quality gate read). Minor, does not block.
- BR-11 — addressed — issueFilesForID single-sources the glob for resolveBranchName and syncPathspec; issueIDFromPath reuses issueIDPrefix.
- BR-12 — not-addressed — Not touched in the rework round and not mentioned in Revisions; still no envelope for a verb designed to run frequently.

### Raised

- **BR-13** [Important] `class-fix-without-class-test` this round's call-site fixes revert green - 5 of 7 pathspec sites and the start-plan delivery have no pinning test
  Measured by revert in a scratch clone at 0807e39, full package suite each time: reverting
  migrate.go:356/:359, push.go:204/:404, merge.go:550, or the syncPointer call at
  startplan.go:85 leaves the suite green. Cause is uniform - issuesync_test.go:593 hand-writes
  migrate's commit argv instead of calling runMigrate, and issuesync_test.go:540-556 hand-wires
  archiveCommitArgs instead of driving runPush/recoverInterruptedArchive. Both run real git and
  neither runs our code, so they pass in exactly the state the finding was about (ARCH-MOCK
  inverted). This is the 3rd finding in the family, so do NOT patch the instances - the rule is
  that a call-site fix is pinned only by a test driving the production entry point, and the
  enforcement that covers the whole class is a source-level guard test over cmd/sdlc asserting
  every git-commit argv built after a narrowed add carries a `--`, allowlisting the two
  deliberate whole-tree commits (push.go:116 `commit -a`, propagatebase.go:271 after `add -A`).
  Same shape as TestRepoLockCommandMetadata / TestForceAckMatchesGateCatalog. Also add
  "sdlc issue sync" to TestRunStartPlan_RendersAtPlanLens's want list - that test exists so a
  dropped pointer line cannot ship silently, and it just failed to do that.
- **BR-14** [Important] `helper-swap-drops-a-mode` change-code's sync from a feature worktree publishes WIP to origin/main and leaves the branch copy dirty
  Probed against a real two-worktree fixture: syncIssue (changecode.go:213) on a feature
  worktree with a tracked-and-edited issue file takes syncViaMainWorktree, pulls, copies,
  commits "#206: issue-sync: spec/plan at change-code" on main, pushes - and leaves
  "M workshop/issues/000206-issue-sync-verb.md" on the branch. The "branch starts from a
  tracked state" property asserted at changecode.go:167-176 is not achieved in that mode, main
  gains a divergent commit of a file the branch will later commit its own version of, and a
  milestone re-run (#156 made worktree creation idempotent for exactly this) gains two network
  round-trips. commitUntrackedIssueFile no-opped here. --worktree=yes is opt-in, hence Important
  not Critical. This is the 2nd finding in the family, so do NOT patch this cell - write the
  mode cross-product the old helper ran under (resolveBranchName's 3 name modes x {on main,
  in-place branch, feature worktree} x {untracked, tracked-and-edited}) as a test table and
  state the intended behavior for each. Round 1 fixed one cell; this round covers four; the
  cells whose behavior actually changed are uncovered.
- **BR-15** [Minor] `comment-narrower-than-code` syncPathspec is documented as an argv builder but does filesystem IO via filepath.Glob
  claim.go:235 reads as a pure pathspec constructor; issueFilesForID globs the working tree, so
  it is a thin IO helper. Worth saying so before someone table-tests it as pure (ARCH-PURE).
- **BR-16** [Minor] `narrowed-add-bare-commit` migrate.go restates the pathspec'd-commit shape four times inline and its hint assertion does not check the pathspec
  migrate.go:352-359 builds the same `commit -m <msg> -- <path>` shape for two commits and two
  --no-commit hints; migrate_test.go:449 asserts only the "--no-commit: both sides staged"
  lead-in, so the hint could lose its `--` unnoticed. A two-line helper mirroring
  archiveCommitArgs would keep the printed command and the executed one from drifting.
- **BR-17** [Minor] `stale-test-fixture-path` pre-existing red test in the tree this boundary crosses - fleet_plan_test.go reads an archived plan path
  TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (fleet_plan_test.go:14) opens
  workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, which dfeba9c archived to history.
  dfeba9c is an ancestor of the review base, so this is not from #206 - but it means the suite is
  red at the close boundary. Worth its own issue; the test should resolve the plan through the
  archive-inclusive lookup rather than a hardcoded workshop/plans path.

## Round 3 — 2026-09-02T13:21:22-07:00 (claude) — BLOCKED

### Disposed

- BR-7 — addressed — merge.go:550 is now caught by the source guard — verified by revert (guard fires, naming merge.go:runMerge).
- BR-8 — addressed — In-place arm covered by TestIssueSync_DryRunCommitsNothing + the change-code variant; syncViaMainWorktree's dry-run early return remains uncovered.
- BR-10 — not-addressed — Deliberately declined with a recorded rationale; the two new workshop/plans/ files are gate ledgers, not a durable plan. Minor, non-blocking; the rule belongs to ariadne#198.
- BR-12 — not-addressed — No envelope declared. The no-push default is the decision; it just is not stated as a budget. ariadne#205 owns the class.
- BR-13 — not-addressed — 6 of 7 now caught (revert-verified); push.go:204 still reverts the full suite green because the allowlist is keyed by function and runPush is exempt for its unrelated `commit -a`. Key the exemption to the exact argv literals, not the enclosing function.
- BR-14 — addressed — Revert-verified: reverting `NoPush: !onMain` reds 12 of the 18 matrix cells.
- BR-15 — addressed — syncPathspec's doc now states it is a thin IO helper, not pure.
- BR-16 — addressed — migrateCommitArgs single-sources the shape; TestMigrateCommitArgs_HintAndCommandAreTheSameBuilder pins the hint's pathspec.
- BR-17 — addressed — Filed as ariadne#207. The suite is still red at this boundary (reproduced), so the close evidence should say so rather than claim a green suite.

### Raised

- **BR-18** [Critical] `early-return-skips-second-concern` `sdlc issue sync --push` cannot publish an already-committed body, and both new warnings name it as the recovery step
  claim.go:186-189 returns `cok("No issue changes to sync.")` before the add/commit/push
  whenever changedIssueFiles is empty. Since this issue split durability from publication,
  "committed locally, push failed" is a state the code deliberately creates — and in it
  --push short-circuits and never pushes, reporting success in green. Measured on a real
  repo + bare origin: local sync, then Push:true → `[ok] No issue changes to sync.`,
  origin/main unmoved at 3f20fb4 while local main is 4b2329f. changecode.go:250-254 and
  issue.go:331-332 both instruct the operator to run exactly that command to finish the
  failed publish; for `issue new` that is the #82 M1 reservation broadcast, left
  unrecoverable by the named route. Fix: when !NoPush, fall through to the push (or gate on
  `rev-list --count origin/main..main`) rather than returning on an empty working-tree diff.
- **BR-19** [Important] `class-fix-without-class-test` runChangeCode's syncIssue call site enters no test — deleting the line leaves the suite green
  Measured: replacing `syncIssue(stderr, f, issuePath)` at changecode.go:179 with a no-op
  leaves the full cmd/sdlc suite green apart from the known-unrelated fleet_plan failure.
  Every TestChangeCodeSyncIssue_* test calls syncIssue directly with a literal path, so they
  prove the helper and mock the wiring — BR-13's exact shape, on the Spec's third piece.
  Structural cause is tracked as ariadne#191 (runChangeCode is not in-process drivable).
  This is the 4th finding in family `class-fix-without-class-test`; do NOT pin this one site.
  The rule: a call-site fix is pinned only by a test entering the production entry point, and
  where that entry point is not drivable a source-level guard must assert the wiring. Enforce
  it once — a wiring table in commitpathspec_guard_test.go's shape covering
  runChangeCode→syncIssue, runPush/runMerge→archiveCommitArgs, runMigrate→migrateCommitArgs —
  or land ariadne#191. Measured prevalence: 5 sites in round 2, plus push.go:204 and this one
  still open in round 3.
- **BR-20** [Minor] `prose-count-restates-enumeration` atlas says "Six sites" over an enumeration of seven
  atlas/workflow/sdlc-binary.md:111 — "both sync arms, both push.go archive commits,
  merge.go's, and migrate.go's two" is seven, and commitpathspec_guard_test.go's own comment
  says seven. Derive the count from the list or drop it.
- **BR-21** [Minor] `comment-narrower-than-code` syncPathspec's "unreachable" rationale does not hold for a deleted issue file
  claim.go:237 argues the "pathspec matches nothing" error is unreachable because callers run
  changedIssueFiles first — but a deleted issue file makes changedIssueFiles non-empty while
  issueFilesForID's glob is empty, so the argument does not cover that state (the outcome is
  still practically unreachable). This is the 3rd finding in family `comment-narrower-than-code`;
  the rule is that a comment asserting unreachability must enumerate the states its guard
  quantifies over — if that cannot be enforced, drop the claim and let the error stand alone.

## Round 4 — 2026-09-02T13:39:24-07:00 (claude) — BLOCKED

### Disposed

- BR-10 — not-addressed — No workshop/plans/000206-*-plan.md; the two new plans/ files are the gate ledger and the review transcript. The recorded "back-filling would produce a copy" rationale stands as an operator call; Minor, non-blocking.
- BR-12 — not-addressed — Still no declared envelope, and round 3 added an unconditional network push on the no-changes path without pricing it.
- BR-13 — addressed — Verified by revert: four pathspec sites reverted at once produce four failures (incl. push.go:204 inside runPush); deleting syncPointer reds TestRunStartPlan_RendersAtPlanLens and the wiring guard.
- BR-18 — not-addressed — syncInPlace is fixed and pinned (both tests red on revert), but claim.go:302-320 pushes origin main without carrying the body: from a feature worktree, `sdlc issue sync --issue N --push` after a local sync prints `[ok] Issues synced to main and pushed to origin.` with origin/main unmoved (measured, incl. via `issue new`'s advertised recovery, where origin/main kept only README.md). Same finding, second arm; that branch has no test.
- BR-19 — addressed — Measured: replacing syncIssue(stderr, f, issuePath) at changecode.go:179 with a no-op fails TestVerbsWireTheirCommitHelpers.
- BR-20 — addressed — atlas/workflow/sdlc-binary.md now says "Seven sites" over a seven-member enumeration.
- BR-21 — addressed — syncPathspec's doc drops the unreachability claim and names the deleted-issue-file state that broke it.

### Raised

- **BR-22** [Minor] `narrowed-add-bare-commit` recoverInterruptedArchive's dry-run hint hand-builds the commit argv instead of using archiveCommitArgs
  push.go:394-397 prints `git commit -m %q -- <paths>` assembled inline, deriving only the
  path list from archiveAddArgs. This is the 4th finding in family `narrowed-add-bare-commit`,
  so do NOT patch the line: the rule round 2 already stated for migrate is that ONE builder
  feeds both the executed commit and the command printed for the operator, so the hint cannot
  lose what the executed form has. push.go is the unswept member of that enumeration
  (migrate.go:426 migrateCommitArgs is the swept one). Measured prevalence: 2 hint sites in
  the tree, 1 derived.
- **BR-23** [Minor] `class-fix-without-class-test` the pathspec class guard's seam enumeration misses gitx.RunGit, so a commit through it escapes
  commitpathspec_guard_test.go:157-172 accepts a git argv only from .Git / .GitInDir /
  gitInDir / exec.Command("git", ...), while its own doc claims every `git commit` argv in
  cmd/sdlc and "the eighth site someone writes next year fails immediately". gitx.RunGit is a
  live seam at ~12 sites in this package (all reads today) and gitx.Capture at more; an inline
  gitx.RunGit("commit", "-m", msg) passes the guard. This is the 5th finding in family
  `class-fix-without-class-test`, so the fix is the rule, not the line: the guard's seam list
  must cover every git-invoking helper the package actually has — add RunGit/Capture and state
  that adding a new git seam requires adding it here, in the same shape as the stale-exemption
  check the guard already performs on its allowlist.
- **BR-24** [Minor] `undocumented-behavior-change` `sdlc claim` with no issue changes now needs the network and dies offline, where it used to be a clean no-op
  Measured: on main, clean tree, unreachable origin, `sdlc claim --issue 206` now fails through
  pushMain; at base 92bd1ad the same command printed `[ok] No issue changes to sync.` and exited
  0. Deliberate consequence of "nothing to commit is not nothing to publish", but claim is the
  one caller that die()s on the error. This is the 2nd finding in family
  `undocumented-behavior-change`; the rule the first one (MERGE_HEAD partial commits) already
  established is that a change which creates a NEW way an existing verb can fail gets recorded
  in the atlas failure-mode list — apply it here, and decide explicitly whether an idempotent
  offline claim should be fatal or a warning.
- **BR-25** [Minor] `comment-narrower-than-code` change-code --dry-run promises "Would sync + push" from branches where it will not push
  changecode.go:162-164 prints `Would sync + push issue #%d` unconditionally, while syncIssue
  sets NoPush: !onMain, so from an in-place branch or a feature worktree the real run commits
  locally and does not publish. This is the 4th finding in family `comment-narrower-than-code`;
  the rule is that a statement about behavior must branch on the same state the code branches
  on — here, derive the dry-run wording from the same `onMain` test syncIssue uses rather than
  restating the outcome.

## Open findings

- **BR-10** [Minor] `durable-plan-artifact` no durable plan in workshop/plans/ for a 9-file, ~900-line change (AGENTS.md §1)
- **BR-12** [Minor] `operating-envelope-undeclared` no declared operating envelope for a verb designed to be run frequently (ARCH-CONSTRAINTS)
- **BR-18** [Critical] `early-return-skips-second-concern` `sdlc issue sync --push` cannot publish an already-committed body, and both new warnings name it as the recovery step
- **BR-22** [Minor] `narrowed-add-bare-commit` recoverInterruptedArchive's dry-run hint hand-builds the commit argv instead of using archiveCommitArgs
- **BR-23** [Minor] `class-fix-without-class-test` the pathspec class guard's seam enumeration misses gitx.RunGit, so a commit through it escapes
- **BR-24** [Minor] `undocumented-behavior-change` `sdlc claim` with no issue changes now needs the network and dies offline, where it used to be a clean no-op
- **BR-25** [Minor] `comment-narrower-than-code` change-code --dry-run promises "Would sync + push" from branches where it will not push
