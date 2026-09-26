---
gate: boundary-review
issue: 252
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-25T15:23:52-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Core concepts table claims future entities are delivered
          detail: The plan claims tracker/migration.go is new and activetime/commit.go is modified, but the former is absent and the latter is unchanged in the pinned M1 range; several later-milestone entities are also absent. Scope or revise the inventory before crossing the boundary. ARCH-PURPOSE
          family: core-concepts-inventory-drift
          round: 1
        - id: BR-2
          severity: Important
          title: Final M1 plan checklist item remains unchecked
          detail: The M1 row covering repeated tests, benchmarks, atlas documentation, and milestone-close evidence remains unchecked while the review is being submitted. Close the row with evidence or defer the boundary.
          family: boundary-checklist-not-closed
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-25T15:30:42-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The plan now distinguishes delivered M1 entities from absent planned M2–M4 entities and records the correction in Revisions.
          round: 2
        - id: BR-2
          disposition: addressed
          note: M1 implementation checklist items are checked; acceptance is explicitly separated and remains pending until milestone-close.
          round: 2
      findings:
        - id: BR-3
          severity: Minor
          title: Committed M1 review artifact contains trailing whitespace
          detail: git diff --check reports trailing whitespace in workshop/plans/000252-issue-cards-tracker-ref-m1-review.md:34 and :39; remove it before final cleanup.
          family: review-artifact-hygiene
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 3
      timestamp: "2026-09-25T17:39:23-07:00"
      agent: claude
      findings:
        - id: BR-4
          severity: Critical
          title: recovery reconcile drives local-effect stages from any worktree; receipts are shared across slots
          detail: refs/sdlc/recovery lives in the git common dir, so every slot lists every receipt. runRecoveryReconcile builds moveDetailRemover/CreationOp from the current checkout and never compares env.branch with spec.SourceBranch. Another slot can fast-forward its own rest, report Confirmed, and delete the receipt while the real source is never removed. Require the source branch for transfer.remove and creation.detail, mark receipts owned elsewhere in list, and add a two-worktree test.
          family: resume-identity-check
          round: 3
        - id: BR-5
          severity: Important
          title: next-action hints print zero-padded --issue 000NNN, which pflag parses as octal (wrong issue)
          detail: 'claim.go:120, issue.go:334, issue.go:346, createop.go:184. Example: 000253 parses as 171, and 000258 fails to parse. The tests pin the broken strings. Use one ID-formatting helper at every site and test that rendered hints parse back to the same ID.'
          family: next-action-hint-correctness
          round: 3
        - id: BR-6
          severity: Important
          title: transfer guard skips handoff records without main_commit, missing the interrupted main-published window
          detail: The handoff record is published before main. An interruption between transfer.main and transfer.record leaves D on main unguarded, and a PR from the source branch before reconcile reaches the add/add keep-ours loss. Include records whose destination exists on main, and test the guard before reconcile.
          family: transfer-guard-coverage
          round: 3
        - id: BR-7
          severity: Important
          title: help for claim/issue/start-plan/change-code still describes origin/main publication and issue sync
          detail: claim.md says it reserves on origin/main and recommends issue sync/publish. issue.md lacks move-detail, recovery and the set-* setters. start-plan.md and change-code.md point at issue sync, contradicting the new behaviour. The help is the agent workflow contract.
          family: changed-verb-docs-drift
          round: 3
        - id: BR-8
          severity: Important
          title: Task 3 files and Core concepts table omit or misname M2 entities (2nd in family)
          detail: 'Task 3 still lists issuerecord.go and issuemetadata.go; neither exists (the setters shipped as cardsetters.go). No table row covers handoff.go and cardset.go (PURE), planningbranch.go, trackerenv.go, issuerecovery.go or candidates.go. Rule: every non-test file added in the boundary window appears in a Core concepts row or is classified as glue, and every Task Files entry exists or is revised. Enumerate with git diff --name-status BASE HEAD before each milestone submission.'
          family: core-concepts-inventory-drift
          round: 3
        - id: BR-9
          severity: Minor
          title: move-detail publishes local details to main after only a HasMirror check (no hand-edit refusal)
          detail: A hand-edited mirrored field (for example status) in local details lands on main unrefused. Run refreshMirror before publication.
          family: mirror-validation-coverage
          round: 3
        - id: BR-10
          severity: Minor
          title: publishEvent/probeEvent discard Push/Probe errors; cardsetters swallows details parse errors
          detail: Uncertain outcomes lose their diagnostic (auth or hook failures), and a malformed details body becomes an empty Log for the reopen guard.
          family: silent-error-swallowing
          round: 3
        - id: BR-11
          severity: Minor
          title: transfer guard is invoked with context.Background() from pr/publish/landing gates
          family: context-propagation
          round: 3
        - id: BR-12
          severity: Minor
          title: handoff destination is not validated as issues-dir plus card basename
          detail: The guard uses it as both an exemption key and a path. One malformed card also fails every PR repo-wide without saying so.
          family: untrusted-input-validation
          round: 3
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-09-25T18:16:11-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: requireSourceCheckout in materialize/removeSource plus reconcile preflight and list labelling; TestRecoveryReconcileRefusesAnotherWorktree fails without it.
          round: 4
        - id: BR-5
          disposition: addressed
          note: issue.CLIRef at every hint; issuehint_guard_test.go source guard plus base-0 round-trip.
          round: 4
        - id: BR-6
          disposition: addressed
          note: records without main_commit are guarded once main holds the destination; TestTransferGuardProtectsAnUnrecordedPublication goes red without it.
          round: 4
        - id: BR-7
          disposition: not-addressed
          note: 'issue.md:119 and atlas/workflow/issue-sync.md:51 still say change-code publishes; atlas/workflow/issue-lifecycle.md:34 and process-manual.md:111 still describe claim via origin/main. Rule: rg every changed verb name across cmd/sdlc/helptext and atlas, and classify each hit.'
          round: 4
        - id: BR-8
          disposition: addressed
          note: All 17 added non-test files have Core concepts rows; Task 3 Files revised; the rule is recorded in Revisions.
          round: 4
        - id: BR-9
          disposition: addressed
          note: refreshMirror runs before publication; TestMoveDetailRefusesHandEditedCardFieldBeforePublishing. It introduced the write-before-check regression raised below.
          round: 4
        - id: BR-10
          disposition: addressed
          note: diagnostics mixin plus TestUncertainStopReportsTheGitError; the cardsetters malformed refusal has no test.
          round: 4
        - id: BR-11
          disposition: addressed
          round: 4
        - id: BR-12
          disposition: addressed
          note: basename, clean-path and no-escape checks plus a repo-wide message; the issues directory itself is not pinned (low impact).
          round: 4
      findings:
        - id: BR-13
          severity: Important
          title: move-detail and change-code write the refreshed mirror before dry-run and precondition checks
          detail: 'Reproduced: after a set-title from another worktree, move-detail --dry-run rewrites the committed details file, then both dry and real runs refuse with "nothing was changed" (issuemovedetail.go:166 runs before checkMoveSource and DryRun). change-code writes at changecode.go:307 before its DryRun check at :196. Rule: checks run before effects; a refusing or dry-run path leaves files unchanged. Fix: refresh in memory, check the original, then write and git add.'
          family: refusal-after-local-effect
          round: 4
        - id: BR-14
          severity: Minor
          title: transfer guard treats any git cat-file failure as the destination being absent from main
          detail: 'This is the 2nd finding in family silent-error-swallowing. Rule: an observation error is never evidence of absence; distinguish "path missing" from a Git failure at every probe (transferguard.go:62, and sweep the other cat-file -e probes in trackerenv/issuemovedetail).'
          family: silent-error-swallowing
          round: 4
        - id: BR-15
          severity: Minor
          title: '"refs/heads/"+env.branch is built by hand at four sites alongside env.checkout'
          detail: issuerecovery.go:62,102,137 and issuemovedetail.go:143,203; add trackerEnv.branchRef() (ARCH-DRY).
          family: shared-helper-extraction
          round: 4
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 5
      timestamp: "2026-09-25T18:46:09-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: claim.md/start-plan.md/change-code.md/issue.md now describe tracker reservation, local-only design, move-detail/recovery/set-* verbs; remaining AGENTS.md issue-sync prose is M4 (instructions) scope per plan Task 7/8.
          round: 5
        - id: BR-13
          disposition: not-addressed
          note: move-detail fixed and tested; change-code only fixed for dry-run — changecode.go:317 still writes the refreshed mirror before gates that exitWithCode(1) at :185, so a refused run leaves the details rewritten, contradicting its own step-2b comment.
          round: 5
        - id: BR-14
          disposition: addressed
          note: transferguard.go:61 and the cat-file/merge-base/diff probes now use trackerEnv.gitTest/has (trackerenv.go:85-104), pinned by TestGitTestSeparatesFalseFromFailure; residual ls-files/rev-parse drops raised separately.
          round: 5
        - id: BR-15
          disposition: addressed
          note: trackerEnv.branchRef() (trackerenv.go:106) replaces all env.branch sites in issuerecovery.go and issuemovedetail.go; remaining refs/heads/+name in planningbranch.go names a different branch.
          round: 5
      findings:
        - id: BR-16
          severity: Minor
          title: move-detail and the transfer guard still discard env.git errors on ls-files/rev-parse probes
          detail: '3rd in family. Rule: no env.git result may drop its error; probes go through gitTest/has or propagate. Instances: issuemovedetail.go:202, :301, :322 (ls-files failure skips git rm --cached), transferguard.go:106-107 (fails safe). Enforce with a grep guard test rather than per-site fixes.'
          family: silent-error-swallowing
          round: 5
        - id: BR-17
          severity: Minor
          title: gitTest doc claims a bad name is an error, but rev-parse -q --verify exits 1 for a bad commit too
          detail: has(badCommit, p) returns (false, nil). Callers pass resolved OIDs today; correct the comment or verify the commit separately.
          family: silent-error-swallowing
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 6
      timestamp: "2026-09-25T19:04:14-07:00"
      agent: claude
      dispose:
        - id: BR-13
          disposition: addressed
          note: change-code writes the mirror only after the gates and review validate (changecode.go:222); reverting the fix turns TestChangeCodeGateRefusalLeavesStaleMirrorOnDisk red (verified); the move-detail dry-run and refusal paths are pinned by TestMoveDetailStaleMirrorIsWrittenOnlyWhenProceeding.
          round: 6
        - id: BR-16
          disposition: addressed
          note: All named sites now propagate errors; the whole package is enforced by TestTrackerGitResultsKeepTheirErrors in gitresult_guard_test.go, a rule rather than per-site fixes.
          round: 6
        - id: BR-17
          disposition: addressed
          note: The gitTest doc in trackerenv.go now says rev-parse -q --verify exits 1 for an unknown commit and that callers pass resolved commits; the has() callers pass mainTip, the merge-tree result, HEAD, or a pinned base.
          round: 6
      findings:
        - id: BR-18
          severity: Minor
          title: move-detail rewrites the stale mirror before tracker.NewTransfer validates the receipt spec
          detail: '2nd finding in this family. Rule: every fallible validation, including receipt/op construction, completes before the first file or index write. In issuemovedetail.go the staleMirror WriteFile and git add run before NewTransfer(spec), whose newReceipt validation can refuse; move NewTransfer above the write.'
          family: refusal-after-local-effect
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#252 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-25T15:23:52-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `core-concepts-inventory-drift` Core concepts table claims future entities are delivered
  The plan claims tracker/migration.go is new and activetime/commit.go is modified, but the former is absent and the latter is unchanged in the pinned M1 range; several later-milestone entities are also absent. Scope or revise the inventory before crossing the boundary. ARCH-PURPOSE
- **BR-2** [Important] `boundary-checklist-not-closed` Final M1 plan checklist item remains unchecked
  The M1 row covering repeated tests, benchmarks, atlas documentation, and milestone-close evidence remains unchecked while the review is being submitted. Close the row with evidence or defer the boundary.

## Round 2 — 2026-09-25T15:30:42-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The plan now distinguishes delivered M1 entities from absent planned M2–M4 entities and records the correction in Revisions.
- BR-2 — addressed — M1 implementation checklist items are checked; acceptance is explicitly separated and remains pending until milestone-close.

### Raised

- **BR-3** [Minor] `review-artifact-hygiene` Committed M1 review artifact contains trailing whitespace
  git diff --check reports trailing whitespace in workshop/plans/000252-issue-cards-tracker-ref-m1-review.md:34 and :39; remove it before final cleanup.

## Round 3 — 2026-09-25T17:39:23-07:00 (claude) — BLOCKED

### Raised

- **BR-4** [Critical] `resume-identity-check` recovery reconcile drives local-effect stages from any worktree; receipts are shared across slots
  refs/sdlc/recovery lives in the git common dir, so every slot lists every receipt. runRecoveryReconcile builds moveDetailRemover/CreationOp from the current checkout and never compares env.branch with spec.SourceBranch. Another slot can fast-forward its own rest, report Confirmed, and delete the receipt while the real source is never removed. Require the source branch for transfer.remove and creation.detail, mark receipts owned elsewhere in list, and add a two-worktree test.
- **BR-5** [Important] `next-action-hint-correctness` next-action hints print zero-padded --issue 000NNN, which pflag parses as octal (wrong issue)
  claim.go:120, issue.go:334, issue.go:346, createop.go:184. Example: 000253 parses as 171, and 000258 fails to parse. The tests pin the broken strings. Use one ID-formatting helper at every site and test that rendered hints parse back to the same ID.
- **BR-6** [Important] `transfer-guard-coverage` transfer guard skips handoff records without main_commit, missing the interrupted main-published window
  The handoff record is published before main. An interruption between transfer.main and transfer.record leaves D on main unguarded, and a PR from the source branch before reconcile reaches the add/add keep-ours loss. Include records whose destination exists on main, and test the guard before reconcile.
- **BR-7** [Important] `changed-verb-docs-drift` help for claim/issue/start-plan/change-code still describes origin/main publication and issue sync
  claim.md says it reserves on origin/main and recommends issue sync/publish. issue.md lacks move-detail, recovery and the set-* setters. start-plan.md and change-code.md point at issue sync, contradicting the new behaviour. The help is the agent workflow contract.
- **BR-8** [Important] `core-concepts-inventory-drift` Task 3 files and Core concepts table omit or misname M2 entities (2nd in family)
  Task 3 still lists issuerecord.go and issuemetadata.go; neither exists (the setters shipped as cardsetters.go). No table row covers handoff.go and cardset.go (PURE), planningbranch.go, trackerenv.go, issuerecovery.go or candidates.go. Rule: every non-test file added in the boundary window appears in a Core concepts row or is classified as glue, and every Task Files entry exists or is revised. Enumerate with git diff --name-status BASE HEAD before each milestone submission.
- **BR-9** [Minor] `mirror-validation-coverage` move-detail publishes local details to main after only a HasMirror check (no hand-edit refusal)
  A hand-edited mirrored field (for example status) in local details lands on main unrefused. Run refreshMirror before publication.
- **BR-10** [Minor] `silent-error-swallowing` publishEvent/probeEvent discard Push/Probe errors; cardsetters swallows details parse errors
  Uncertain outcomes lose their diagnostic (auth or hook failures), and a malformed details body becomes an empty Log for the reopen guard.
- **BR-11** [Minor] `context-propagation` transfer guard is invoked with context.Background() from pr/publish/landing gates
- **BR-12** [Minor] `untrusted-input-validation` handoff destination is not validated as issues-dir plus card basename
  The guard uses it as both an exemption key and a path. One malformed card also fails every PR repo-wide without saying so.

## Round 4 — 2026-09-25T18:16:11-07:00 (claude) — BLOCKED

### Disposed

- BR-4 — addressed — requireSourceCheckout in materialize/removeSource plus reconcile preflight and list labelling; TestRecoveryReconcileRefusesAnotherWorktree fails without it.
- BR-5 — addressed — issue.CLIRef at every hint; issuehint_guard_test.go source guard plus base-0 round-trip.
- BR-6 — addressed — records without main_commit are guarded once main holds the destination; TestTransferGuardProtectsAnUnrecordedPublication goes red without it.
- BR-7 — not-addressed — issue.md:119 and atlas/workflow/issue-sync.md:51 still say change-code publishes; atlas/workflow/issue-lifecycle.md:34 and process-manual.md:111 still describe claim via origin/main. Rule: rg every changed verb name across cmd/sdlc/helptext and atlas, and classify each hit.
- BR-8 — addressed — All 17 added non-test files have Core concepts rows; Task 3 Files revised; the rule is recorded in Revisions.
- BR-9 — addressed — refreshMirror runs before publication; TestMoveDetailRefusesHandEditedCardFieldBeforePublishing. It introduced the write-before-check regression raised below.
- BR-10 — addressed — diagnostics mixin plus TestUncertainStopReportsTheGitError; the cardsetters malformed refusal has no test.
- BR-11 — addressed
- BR-12 — addressed — basename, clean-path and no-escape checks plus a repo-wide message; the issues directory itself is not pinned (low impact).

### Raised

- **BR-13** [Important] `refusal-after-local-effect` move-detail and change-code write the refreshed mirror before dry-run and precondition checks
  Reproduced: after a set-title from another worktree, move-detail --dry-run rewrites the committed details file, then both dry and real runs refuse with "nothing was changed" (issuemovedetail.go:166 runs before checkMoveSource and DryRun). change-code writes at changecode.go:307 before its DryRun check at :196. Rule: checks run before effects; a refusing or dry-run path leaves files unchanged. Fix: refresh in memory, check the original, then write and git add.
- **BR-14** [Minor] `silent-error-swallowing` transfer guard treats any git cat-file failure as the destination being absent from main
  This is the 2nd finding in family silent-error-swallowing. Rule: an observation error is never evidence of absence; distinguish "path missing" from a Git failure at every probe (transferguard.go:62, and sweep the other cat-file -e probes in trackerenv/issuemovedetail).
- **BR-15** [Minor] `shared-helper-extraction` "refs/heads/"+env.branch is built by hand at four sites alongside env.checkout
  issuerecovery.go:62,102,137 and issuemovedetail.go:143,203; add trackerEnv.branchRef() (ARCH-DRY).

## Round 5 — 2026-09-25T18:46:09-07:00 (claude) — BLOCKED

### Disposed

- BR-7 — addressed — claim.md/start-plan.md/change-code.md/issue.md now describe tracker reservation, local-only design, move-detail/recovery/set-* verbs; remaining AGENTS.md issue-sync prose is M4 (instructions) scope per plan Task 7/8.
- BR-13 — not-addressed — move-detail fixed and tested; change-code only fixed for dry-run — changecode.go:317 still writes the refreshed mirror before gates that exitWithCode(1) at :185, so a refused run leaves the details rewritten, contradicting its own step-2b comment.
- BR-14 — addressed — transferguard.go:61 and the cat-file/merge-base/diff probes now use trackerEnv.gitTest/has (trackerenv.go:85-104), pinned by TestGitTestSeparatesFalseFromFailure; residual ls-files/rev-parse drops raised separately.
- BR-15 — addressed — trackerEnv.branchRef() (trackerenv.go:106) replaces all env.branch sites in issuerecovery.go and issuemovedetail.go; remaining refs/heads/+name in planningbranch.go names a different branch.

### Raised

- **BR-16** [Minor] `silent-error-swallowing` move-detail and the transfer guard still discard env.git errors on ls-files/rev-parse probes
  3rd in family. Rule: no env.git result may drop its error; probes go through gitTest/has or propagate. Instances: issuemovedetail.go:202, :301, :322 (ls-files failure skips git rm --cached), transferguard.go:106-107 (fails safe). Enforce with a grep guard test rather than per-site fixes.
- **BR-17** [Minor] `silent-error-swallowing` gitTest doc claims a bad name is an error, but rev-parse -q --verify exits 1 for a bad commit too
  has(badCommit, p) returns (false, nil). Callers pass resolved OIDs today; correct the comment or verify the commit separately.

## Round 6 — 2026-09-25T19:04:14-07:00 (claude) — passed

### Disposed

- BR-13 — addressed — change-code writes the mirror only after the gates and review validate (changecode.go:222); reverting the fix turns TestChangeCodeGateRefusalLeavesStaleMirrorOnDisk red (verified); the move-detail dry-run and refusal paths are pinned by TestMoveDetailStaleMirrorIsWrittenOnlyWhenProceeding.
- BR-16 — addressed — All named sites now propagate errors; the whole package is enforced by TestTrackerGitResultsKeepTheirErrors in gitresult_guard_test.go, a rule rather than per-site fixes.
- BR-17 — addressed — The gitTest doc in trackerenv.go now says rev-parse -q --verify exits 1 for an unknown commit and that callers pass resolved commits; the has() callers pass mainTip, the merge-tree result, HEAD, or a pinned base.

### Raised

- **BR-18** [Minor] `refusal-after-local-effect` move-detail rewrites the stale mirror before tracker.NewTransfer validates the receipt spec
  2nd finding in this family. Rule: every fallible validation, including receipt/op construction, completes before the first file or index write. In issuemovedetail.go the staleMirror WriteFile and git add run before NewTransfer(spec), whose newReceipt validation can refuse; move NewTransfer above the write.

## Open findings

- **BR-3** [Minor] `review-artifact-hygiene` Committed M1 review artifact contains trailing whitespace
- **BR-18** [Minor] `refusal-after-local-effect` move-detail rewrites the stale mirror before tracker.NewTransfer validates the receipt spec
