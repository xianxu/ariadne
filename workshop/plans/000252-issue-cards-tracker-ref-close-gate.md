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
    - "n": 7
      timestamp: "2026-09-25T22:19:01-07:00"
      agent: claude
      findings:
        - id: BR-19
          severity: Critical
          title: Squash/rebase-landed PR never flips its card to done, yet its details are archived
          detail: settleLandingCompletions accepts only evidence-ancestor-of-main; selectTrackedLandingIssues selects by PR head/base and archives. Under squash/rebase (supported by landingFakeGH) the card stays codecomplete forever. Derive done from the confirmed pr.MergeOID in the durable path; test merge/squash/rebase through runMerge.
          family: landing-completion-proof
          round: 7
        - id: BR-20
          severity: Critical
          title: Core concepts label IO entities PURE (loadWindowCommits, LoadRecords join) and Task 6 Files lists unchanged merge.go/reviewstate.go
          detail: '3rd in family. Rule: when a Revisions entry renames a planned entity, re-derive its kind from the shipped code and its test IO, and re-check every Task Files list against name-status. Relabel as INTEGRATION or extract a pure composeRecords.'
          family: core-concepts-inventory-drift
          round: 7
        - id: BR-21
          severity: Important
          title: Outstanding completion receipt is neither refused nor superseded by a re-close; resume can rebind the card to an older review
          detail: '2nd in family. FIX-THEN-SHIP saves receipt A; re-running close drives B; reconcile later resumes A and SetCardCompletion accepts any new token, or A wedges on nothing-to-record. Rule: one outstanding receipt per (issue, operation); a resume proves it is still the newest generation before effects.'
          family: resume-identity-check
          round: 7
        - id: BR-22
          severity: Important
          title: publishTrackerClose preconditions (branch, card, reviewed commit, evidence paths) run after applyClose wrote the Log, ledger and calibration and after the review
          detail: '3rd in family. Rule: every verdict-independent tracker-close precondition runs in computeClose before review dispatch; publishTrackerClose only performs effects.'
          family: refusal-after-local-effect
          round: 7
        - id: BR-23
          severity: Important
          title: Tracker reads and card CAS writes on merge/push/landing paths use context.Background()
          detail: '2nd in family. Sites: settleLandingCompletions, publishTrackerCompletions, selectTrackedLandingIssues, overlayCardStatus, historyFileIsTerminal, actualTrackerInputs and other readers. Rule: any path reaching loadIssueRecords/openTrackerAt/Repository takes the verb context; enforce with a source-guard test.'
          family: context-propagation
          round: 7
        - id: BR-24
          severity: Important
          title: Push/merge fetch the tracker many times per command (plus once per archived file) against the declared one-fetch budget
          detail: Load tracker.Records once per command and thread it to the publish gate, not-done scan, flip, archive and historyFileIsTerminal; pre-tracker repos also pay repeated ls-remote today.
          family: operating-envelope-enforcement
          round: 7
        - id: BR-25
          severity: Important
          title: A FIX-THEN-SHIP evidence commit rebuilds from worktree files that a fix commit (git commit -a) sweeps away, wedging the receipt
          detail: Prepare refuses when no evidence file differs from HEAD; the test passes only because the ledger stays untracked. Pin evidence blobs in the receipt or a recovery ref when it is saved.
          family: deferred-effect-input-drift
          round: 7
        - id: BR-26
          severity: Minor
          title: Tracker load errors become absence in collectGitHubIssueNumbers, guardIssueNotDone and actualTrackerInputs
          detail: '5th in family. Rule: a Records load error in a verb is surfaced (warn, or error in gates), never converted to no record; guardIssueNotDone should also use Fresh. Needs a lint over loadIssueRecords call sites.'
          family: silent-error-swallowing
          round: 7
        - id: BR-27
          severity: Minor
          title: fleet repoRecords hard-codes the main resting branch, duplicating recordsRepository
          detail: 2nd in family. Move the resting-branch-aware repository resolution into internal/tracker and use it from both.
          family: shared-helper-extraction
          round: 7
        - id: BR-28
          severity: Minor
          title: Module imports placed inside the stdlib import group in several changed files
          family: import-grouping
          round: 7
      boundary: M3
      recipe: milestone-review
      blocked: true
    - "n": 8
      timestamp: "2026-09-25T23:18:40-07:00"
      agent: claude
      dispose:
        - id: BR-19
          disposition: addressed
          note: completeLandingPR completes by HeadOID..BaseOID at MergeOID; runMerge test covers merge/squash/rebase plus idempotent retry, and the targeted run passed.
          round: 8
        - id: BR-20
          disposition: addressed
          note: Plan rows 30/52/53 relabelled (composeRecords PURE, loaders INTEGRATION); Task 5/6 Files lists revised against name-status.
          round: 8
        - id: BR-21
          disposition: addressed
          note: prepareTrackerClose supersedes discardable closes and refuses started ones; newestClose guards both stages; TestTrackerReCloseSupersedesAnUnstartedClose.
          round: 8
        - id: BR-22
          disposition: addressed
          note: Branch, card and pending-close checks moved into computeClose; TestTrackerCloseRefusesPreconditionsBeforeReview asserts zero judge calls.
          round: 8
        - id: BR-23
          disposition: addressed
          note: Contexts threaded through package main with the TestVerbContextsReachTrackerReads guard; the fleet leftover is raised as a new Minor.
          round: 8
        - id: BR-24
          disposition: addressed
          note: recordsScope gives one composed view per repo per command, invalidated on card writes; unit test TestRecordsScopeFetchesOncePerCommand.
          round: 8
        - id: BR-25
          disposition: addressed
          note: Evidence pinned as blobs held by the recovery ref; an empty commit is allowed; TestTrackerCloseFixThenShipSurvivesASweepingFixCommit. Pinned blobs now overwrite HEAD, raised as new.
          round: 8
        - id: BR-26
          disposition: addressed
          note: 'Load errors surfaced (warn, fail-closed guard, actual warning) and the regex guard extended. Residual: guardIssueNotDone still uses PreferFresh and its doc comment is stale.'
          round: 8
        - id: BR-27
          disposition: addressed
          note: tracker.RepositoryForCheckout is shared by fleet and recordsRepository.
          round: 8
        - id: BR-28
          disposition: addressed
          note: Mixed import groups in actual.go, pr.go and trackercompletion_test.go are fixed.
          round: 8
      findings:
        - id: BR-29
          severity: Important
          title: Non-durable merge completes owned cards to done even when the branch was abandoned unmerged
          detail: 'merge.go:575 runs completeOnCard whenever trackerEnvForMerge is non-nil. On the worktree path with no PR, where the operator answers remove-without-merging, the card goes done with LandedCommit set to a main HEAD that lacks the evidence. 2nd in family. Rule: a card goes done only on a confirmed landing observation of its evidence. Gate on merged and test the abandon path.'
          family: landing-completion-proof
          round: 8
        - id: BR-30
          severity: Important
          title: Deferred evidence commit overwrites HEAD's newer details, ledgers and sidecars with the blobs pinned at close
          detail: 'gitEvidence.Prepare reads HEAD''s tree and then writes pinned blobs over it. A post-FIX-THEN-SHIP Log edit committed with the fixes is reverted in the evidence commit, visible only as a dirty worktree. 2nd in family. Rule: replaying pinned inputs is a three-way merge against their base. Record the base blob, and refuse or keep HEAD when HEAD''s version differs from both.'
          family: deferred-effect-input-drift
          round: 8
        - id: BR-31
          severity: Minor
          title: fleet repoRecords still reads the tracker under context.Background(); the guard only parses package main
          detail: '3rd in family. Rule: every exported tracker entry point takes a context, and the source guard should scan internal packages too.'
          family: context-propagation
          round: 8
        - id: BR-32
          severity: Minor
          title: Ticked Task 5/6 plan rows name functions that do not exist
          detail: 'FinalizeTrackerClose, SelectCompletedIssues, SelectActivityEvents and ReadIssueRecord are absent from the code. 4th in family. Rule: every identifier in a ticked plan row greps to shipped code, or is renamed in a Revisions entry.'
          family: core-concepts-inventory-drift
          round: 8
      boundary: M3
      recipe: milestone-review
      blocked: true
    - "n": 9
      timestamp: "2026-09-25T23:59:34-07:00"
      agent: claude
      dispose:
        - id: BR-29
          disposition: addressed
          note: merge.go:577 gates completeOnCard on merged; TestAbandonedWorktreeMergeLeavesTheCloseCodecomplete exercises the no-PR abandon path and would go red without the gate.
          round: 9
        - id: BR-30
          disposition: addressed
          note: EvidenceEntry.Replays(base, head) three-way check keeps HEAD when changed since review; pure table test plus e2e TestTrackerCloseFixThenShipKeepsALaterEditOfAPinnedFile.
          round: 9
        - id: BR-31
          disposition: addressed
          note: LookupRepoIssues/repoRecords/CollectInventory take ctx; context_guard_test now parses every internal package with qualified allowlist.
          round: 9
        - id: BR-32
          disposition: addressed
          note: Plan rows 143-145, 207-217 now name shipped identifiers with planned names in parentheses, and Revisions line 359 records the rename; all shipped names grep to code.
          round: 9
      findings:
        - id: BR-33
          severity: Minor
          title: Evidence replay silently skips pinned files whose HEAD version diverged from both base and pin
          detail: '3rd in family. Rule: a deferred replay reports every pinned input it did not apply (path plus reason), so a later commit that also dropped close evidence such as a ledger row is visible rather than silently kept.'
          family: deferred-effect-input-drift
          round: 9
      boundary: M3
      recipe: milestone-review
      blocked: false
    - "n": 10
      timestamp: "2026-09-26T18:17:37-07:00"
      agent: claude
      findings:
        - id: BR-34
          severity: Critical
          title: 'Codecomplete migration: main conversion mirrors the unbound card; tracker holds the bound card (missing baseline, merge conflict)'
          detail: 'This is the 4th finding in family deferred-effect-input-drift. migration.go:133-157 derives `mirrored` from the pre-binding card, then replaces `card` with SetCardCompletion''s output, so main''s card_mirror names a blob never written, while --reconcile on the branch mirrors the bound blob. Result: refresh refuses on the missing baseline, and the PR merge conflicts on the card_mirror line. Rule for the class: any artifact that pins a source identity (OID/digest) must be computed from the source''s final bytes by one constructor called after the last mutation, and the pure plan must assert pin == identity(final). Fix by attaching the mirror after binding. Add a population invariant (each conversion''s MirrorBaselineOID == CardBlobOID(card)) and extend TestIssueMigrateImportsAProvableLegacyClose through merge and landing.'
          family: deferred-effect-input-drift
          round: 10
        - id: BR-35
          severity: Minor
          title: Moved-main apply error promises resumption that sameTrackerFiles refuses when issue files changed
          detail: issuemigrate.go:466 says the new plan resumes from the existing tracker; any change to active or archived details changes the plan's cards, so the bootstrap check refuses and only the atlas's manual abandonment works. The message should name that path.
          family: next-action-hint-correctness
          round: 10
        - id: BR-36
          severity: Minor
          title: Cutover guard adds a rev-list process per guarded read, outside the measured envelope
          detail: cutover.go:105 runs HasRoot on every read(), including each CAS retry; the plan's 2-process, 0.85 s measurement predates the guard.
          family: operating-envelope-enforcement
          round: 10
        - id: BR-37
          severity: Minor
          title: applyTrackerBootstrap re-implements TrunkFile.HasRoot's root listing
          detail: issuemigrate.go:407 runs rev-list --max-parents=0 through env.git; a TrunkFile Roots helper shared with HasRoot would keep one implementation.
          family: shared-helper-extraction
          round: 10
        - id: BR-38
          severity: Minor
          title: Consumer inventory's tracker ID validation for 40-duplicate-issue-id is neither delivered nor dispositioned
          detail: 'This is the 5th finding in family core-concepts-inventory-drift. Only an atlas note was added; a cardless details file merged through the UI can later collide with a card-allocated ID. Rule: every inventory row needs either a delivering diff or an explicit Revision disposition. The close gate could check this by diffing inventory rows against the window''s name-status.'
          family: core-concepts-inventory-drift
          round: 10
      boundary: M4
      recipe: milestone-review
      blocked: true
    - "n": 11
      timestamp: "2026-09-26T18:51:40-07:00"
      agent: claude
      dispose:
        - id: BR-34
          disposition: addressed
          note: MirrorDetails pins after binding (migration.go:157); population test asserts pin==CardBlobOID(final) incl. bound codecomplete; e2e merges, lands, settles.
          round: 11
        - id: BR-35
          disposition: addressed
          note: issuemigrate.go:471-474 now names both resume and abandon-unwritten-tracker outcomes.
          round: 11
        - id: BR-36
          disposition: addressed
          note: Guard caches verified (root, tip OID) in Repository.verified; benchmark measures 1 process per generation.
          round: 11
        - id: BR-37
          disposition: addressed
          note: TrunkFile.Roots is shared by HasRoot and applyTrackerBootstrap.
          round: 11
        - id: BR-38
          disposition: addressed
          note: lint-ids refuses cardless added details in tracker repos; pure unit test plus real-Git exit-code test.
          round: 11
      findings:
        - id: BR-39
          severity: Minor
          title: Makefile and close-issue.py hardcode the issue-tracker ref name that Go derives from vocab
          detail: This is the 4th finding in shared-helper-extraction. The shell fallbacks cannot call the binary, so the rule is that every non-Go copy of a vocab constant is pinned by a test that reads vocab. trackedlegacy_test also hardcodes the literal, so a rename would not turn it red.
          family: shared-helper-extraction
          round: 11
        - id: BR-40
          severity: Minor
          title: A landed branch close ignores CodeAfter, so unlanded post-close branch code settles the card to done
          detail: 'This is the 3rd finding in landing-completion-proof. Rule: done requires proof that every ref carrying the close has nothing beyond main after it. Apply CodeAfter to a landed anchor''s branch tip relative to main, not only to unlanded anchors.'
          family: landing-completion-proof
          round: 11
      boundary: M4
      recipe: milestone-review
      blocked: false
    - "n": 12
      timestamp: "2026-09-27T18:51:52-07:00"
      agent: claude
      dispose:
        - id: BR-3
          disposition: addressed
          note: git diff --check over the window reports nothing for the M1 review artifact.
          round: 12
        - id: BR-18
          disposition: addressed
          note: issuemovedetail.go:200 constructs tracker.NewTransfer before the staleMirror WriteFile and git add at :204-216.
          round: 12
        - id: BR-33
          disposition: addressed
          note: closetracker.go:65-68 names every superseded pinned file in a Close-Kept trailer plus a warning; closetracker_test.go:181 asserts the trailer.
          round: 12
        - id: BR-39
          disposition: addressed
          note: TestShellFallbacksSpellTheTrackerVocabulary derives the fetched-ref glob from vocab and the marker from tracker.CutoverMarkerPath, and checks Makefile.workflow plus scripts/close-issue.py.
          round: 12
        - id: BR-40
          disposition: addressed
          note: bindLegacyClose refuses a landed anchor whose CodeAfter (main...ref) is set; migration_test.go:197 covers it. The MigrationAnchor doc comment is stale (raised Minor below).
          round: 12
      findings:
        - id: BR-41
          severity: Important
          title: Legacy-mode detection requires ls-remote, so offline legacy repos can no longer run local-only verbs
          detail: repositoryTracked calls Repository.Initialized, whose TrunkFile.RemoteExists runs ls-remote; an unreachable remote returns an error, so claim, set-status, issue new and change-code refuse before the legacy path. Reproduced with issue set-status against an unreachable origin. Decide from local evidence first (cutover marker, fetched tracker ref), and treat an offline remote with no evidence as legacy for local verbs; add an unreachable-origin test.
          family: local-verb-network-dependency
          round: 12
        - id: BR-42
          severity: Minor
          title: MigrationAnchor doc still says CodeAfter is not asked of a landed close
          detail: This is the 4th finding in landing-completion-proof, and it is doc drift left behind by the BR-40 fix. The rule already applies in code; only migration.go:36-40 contradicts it. Update the comment to say a landed close checks the branch's code beyond main.
          family: landing-completion-proof
          round: 12
      recipe: milestone-review
      blocked: true
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

## Round 7 — 2026-09-25T22:19:01-07:00 (claude) — BLOCKED

### Raised

- **BR-19** [Critical] `landing-completion-proof` Squash/rebase-landed PR never flips its card to done, yet its details are archived
  settleLandingCompletions accepts only evidence-ancestor-of-main; selectTrackedLandingIssues selects by PR head/base and archives. Under squash/rebase (supported by landingFakeGH) the card stays codecomplete forever. Derive done from the confirmed pr.MergeOID in the durable path; test merge/squash/rebase through runMerge.
- **BR-20** [Critical] `core-concepts-inventory-drift` Core concepts label IO entities PURE (loadWindowCommits, LoadRecords join) and Task 6 Files lists unchanged merge.go/reviewstate.go
  3rd in family. Rule: when a Revisions entry renames a planned entity, re-derive its kind from the shipped code and its test IO, and re-check every Task Files list against name-status. Relabel as INTEGRATION or extract a pure composeRecords.
- **BR-21** [Important] `resume-identity-check` Outstanding completion receipt is neither refused nor superseded by a re-close; resume can rebind the card to an older review
  2nd in family. FIX-THEN-SHIP saves receipt A; re-running close drives B; reconcile later resumes A and SetCardCompletion accepts any new token, or A wedges on nothing-to-record. Rule: one outstanding receipt per (issue, operation); a resume proves it is still the newest generation before effects.
- **BR-22** [Important] `refusal-after-local-effect` publishTrackerClose preconditions (branch, card, reviewed commit, evidence paths) run after applyClose wrote the Log, ledger and calibration and after the review
  3rd in family. Rule: every verdict-independent tracker-close precondition runs in computeClose before review dispatch; publishTrackerClose only performs effects.
- **BR-23** [Important] `context-propagation` Tracker reads and card CAS writes on merge/push/landing paths use context.Background()
  2nd in family. Sites: settleLandingCompletions, publishTrackerCompletions, selectTrackedLandingIssues, overlayCardStatus, historyFileIsTerminal, actualTrackerInputs and other readers. Rule: any path reaching loadIssueRecords/openTrackerAt/Repository takes the verb context; enforce with a source-guard test.
- **BR-24** [Important] `operating-envelope-enforcement` Push/merge fetch the tracker many times per command (plus once per archived file) against the declared one-fetch budget
  Load tracker.Records once per command and thread it to the publish gate, not-done scan, flip, archive and historyFileIsTerminal; pre-tracker repos also pay repeated ls-remote today.
- **BR-25** [Important] `deferred-effect-input-drift` A FIX-THEN-SHIP evidence commit rebuilds from worktree files that a fix commit (git commit -a) sweeps away, wedging the receipt
  Prepare refuses when no evidence file differs from HEAD; the test passes only because the ledger stays untracked. Pin evidence blobs in the receipt or a recovery ref when it is saved.
- **BR-26** [Minor] `silent-error-swallowing` Tracker load errors become absence in collectGitHubIssueNumbers, guardIssueNotDone and actualTrackerInputs
  5th in family. Rule: a Records load error in a verb is surfaced (warn, or error in gates), never converted to no record; guardIssueNotDone should also use Fresh. Needs a lint over loadIssueRecords call sites.
- **BR-27** [Minor] `shared-helper-extraction` fleet repoRecords hard-codes the main resting branch, duplicating recordsRepository
  2nd in family. Move the resting-branch-aware repository resolution into internal/tracker and use it from both.
- **BR-28** [Minor] `import-grouping` Module imports placed inside the stdlib import group in several changed files

## Round 8 — 2026-09-25T23:18:40-07:00 (claude) — BLOCKED

### Disposed

- BR-19 — addressed — completeLandingPR completes by HeadOID..BaseOID at MergeOID; runMerge test covers merge/squash/rebase plus idempotent retry, and the targeted run passed.
- BR-20 — addressed — Plan rows 30/52/53 relabelled (composeRecords PURE, loaders INTEGRATION); Task 5/6 Files lists revised against name-status.
- BR-21 — addressed — prepareTrackerClose supersedes discardable closes and refuses started ones; newestClose guards both stages; TestTrackerReCloseSupersedesAnUnstartedClose.
- BR-22 — addressed — Branch, card and pending-close checks moved into computeClose; TestTrackerCloseRefusesPreconditionsBeforeReview asserts zero judge calls.
- BR-23 — addressed — Contexts threaded through package main with the TestVerbContextsReachTrackerReads guard; the fleet leftover is raised as a new Minor.
- BR-24 — addressed — recordsScope gives one composed view per repo per command, invalidated on card writes; unit test TestRecordsScopeFetchesOncePerCommand.
- BR-25 — addressed — Evidence pinned as blobs held by the recovery ref; an empty commit is allowed; TestTrackerCloseFixThenShipSurvivesASweepingFixCommit. Pinned blobs now overwrite HEAD, raised as new.
- BR-26 — addressed — Load errors surfaced (warn, fail-closed guard, actual warning) and the regex guard extended. Residual: guardIssueNotDone still uses PreferFresh and its doc comment is stale.
- BR-27 — addressed — tracker.RepositoryForCheckout is shared by fleet and recordsRepository.
- BR-28 — addressed — Mixed import groups in actual.go, pr.go and trackercompletion_test.go are fixed.

### Raised

- **BR-29** [Important] `landing-completion-proof` Non-durable merge completes owned cards to done even when the branch was abandoned unmerged
  merge.go:575 runs completeOnCard whenever trackerEnvForMerge is non-nil. On the worktree path with no PR, where the operator answers remove-without-merging, the card goes done with LandedCommit set to a main HEAD that lacks the evidence. 2nd in family. Rule: a card goes done only on a confirmed landing observation of its evidence. Gate on merged and test the abandon path.
- **BR-30** [Important] `deferred-effect-input-drift` Deferred evidence commit overwrites HEAD's newer details, ledgers and sidecars with the blobs pinned at close
  gitEvidence.Prepare reads HEAD's tree and then writes pinned blobs over it. A post-FIX-THEN-SHIP Log edit committed with the fixes is reverted in the evidence commit, visible only as a dirty worktree. 2nd in family. Rule: replaying pinned inputs is a three-way merge against their base. Record the base blob, and refuse or keep HEAD when HEAD's version differs from both.
- **BR-31** [Minor] `context-propagation` fleet repoRecords still reads the tracker under context.Background(); the guard only parses package main
  3rd in family. Rule: every exported tracker entry point takes a context, and the source guard should scan internal packages too.
- **BR-32** [Minor] `core-concepts-inventory-drift` Ticked Task 5/6 plan rows name functions that do not exist
  FinalizeTrackerClose, SelectCompletedIssues, SelectActivityEvents and ReadIssueRecord are absent from the code. 4th in family. Rule: every identifier in a ticked plan row greps to shipped code, or is renamed in a Revisions entry.

## Round 9 — 2026-09-25T23:59:34-07:00 (claude) — passed

### Disposed

- BR-29 — addressed — merge.go:577 gates completeOnCard on merged; TestAbandonedWorktreeMergeLeavesTheCloseCodecomplete exercises the no-PR abandon path and would go red without the gate.
- BR-30 — addressed — EvidenceEntry.Replays(base, head) three-way check keeps HEAD when changed since review; pure table test plus e2e TestTrackerCloseFixThenShipKeepsALaterEditOfAPinnedFile.
- BR-31 — addressed — LookupRepoIssues/repoRecords/CollectInventory take ctx; context_guard_test now parses every internal package with qualified allowlist.
- BR-32 — addressed — Plan rows 143-145, 207-217 now name shipped identifiers with planned names in parentheses, and Revisions line 359 records the rename; all shipped names grep to code.

### Raised

- **BR-33** [Minor] `deferred-effect-input-drift` Evidence replay silently skips pinned files whose HEAD version diverged from both base and pin
  3rd in family. Rule: a deferred replay reports every pinned input it did not apply (path plus reason), so a later commit that also dropped close evidence such as a ledger row is visible rather than silently kept.

## Round 10 — 2026-09-26T18:17:37-07:00 (claude) — BLOCKED

### Raised

- **BR-34** [Critical] `deferred-effect-input-drift` Codecomplete migration: main conversion mirrors the unbound card; tracker holds the bound card (missing baseline, merge conflict)
  This is the 4th finding in family deferred-effect-input-drift. migration.go:133-157 derives `mirrored` from the pre-binding card, then replaces `card` with SetCardCompletion's output, so main's card_mirror names a blob never written, while --reconcile on the branch mirrors the bound blob. Result: refresh refuses on the missing baseline, and the PR merge conflicts on the card_mirror line. Rule for the class: any artifact that pins a source identity (OID/digest) must be computed from the source's final bytes by one constructor called after the last mutation, and the pure plan must assert pin == identity(final). Fix by attaching the mirror after binding. Add a population invariant (each conversion's MirrorBaselineOID == CardBlobOID(card)) and extend TestIssueMigrateImportsAProvableLegacyClose through merge and landing.
- **BR-35** [Minor] `next-action-hint-correctness` Moved-main apply error promises resumption that sameTrackerFiles refuses when issue files changed
  issuemigrate.go:466 says the new plan resumes from the existing tracker; any change to active or archived details changes the plan's cards, so the bootstrap check refuses and only the atlas's manual abandonment works. The message should name that path.
- **BR-36** [Minor] `operating-envelope-enforcement` Cutover guard adds a rev-list process per guarded read, outside the measured envelope
  cutover.go:105 runs HasRoot on every read(), including each CAS retry; the plan's 2-process, 0.85 s measurement predates the guard.
- **BR-37** [Minor] `shared-helper-extraction` applyTrackerBootstrap re-implements TrunkFile.HasRoot's root listing
  issuemigrate.go:407 runs rev-list --max-parents=0 through env.git; a TrunkFile Roots helper shared with HasRoot would keep one implementation.
- **BR-38** [Minor] `core-concepts-inventory-drift` Consumer inventory's tracker ID validation for 40-duplicate-issue-id is neither delivered nor dispositioned
  This is the 5th finding in family core-concepts-inventory-drift. Only an atlas note was added; a cardless details file merged through the UI can later collide with a card-allocated ID. Rule: every inventory row needs either a delivering diff or an explicit Revision disposition. The close gate could check this by diffing inventory rows against the window's name-status.

## Round 11 — 2026-09-26T18:51:40-07:00 (claude) — passed

### Disposed

- BR-34 — addressed — MirrorDetails pins after binding (migration.go:157); population test asserts pin==CardBlobOID(final) incl. bound codecomplete; e2e merges, lands, settles.
- BR-35 — addressed — issuemigrate.go:471-474 now names both resume and abandon-unwritten-tracker outcomes.
- BR-36 — addressed — Guard caches verified (root, tip OID) in Repository.verified; benchmark measures 1 process per generation.
- BR-37 — addressed — TrunkFile.Roots is shared by HasRoot and applyTrackerBootstrap.
- BR-38 — addressed — lint-ids refuses cardless added details in tracker repos; pure unit test plus real-Git exit-code test.

### Raised

- **BR-39** [Minor] `shared-helper-extraction` Makefile and close-issue.py hardcode the issue-tracker ref name that Go derives from vocab
  This is the 4th finding in shared-helper-extraction. The shell fallbacks cannot call the binary, so the rule is that every non-Go copy of a vocab constant is pinned by a test that reads vocab. trackedlegacy_test also hardcodes the literal, so a rename would not turn it red.
- **BR-40** [Minor] `landing-completion-proof` A landed branch close ignores CodeAfter, so unlanded post-close branch code settles the card to done
  This is the 3rd finding in landing-completion-proof. Rule: done requires proof that every ref carrying the close has nothing beyond main after it. Apply CodeAfter to a landed anchor's branch tip relative to main, not only to unlanded anchors.

## Round 12 — 2026-09-27T18:51:52-07:00 (claude) — BLOCKED

### Disposed

- BR-3 — addressed — git diff --check over the window reports nothing for the M1 review artifact.
- BR-18 — addressed — issuemovedetail.go:200 constructs tracker.NewTransfer before the staleMirror WriteFile and git add at :204-216.
- BR-33 — addressed — closetracker.go:65-68 names every superseded pinned file in a Close-Kept trailer plus a warning; closetracker_test.go:181 asserts the trailer.
- BR-39 — addressed — TestShellFallbacksSpellTheTrackerVocabulary derives the fetched-ref glob from vocab and the marker from tracker.CutoverMarkerPath, and checks Makefile.workflow plus scripts/close-issue.py.
- BR-40 — addressed — bindLegacyClose refuses a landed anchor whose CodeAfter (main...ref) is set; migration_test.go:197 covers it. The MigrationAnchor doc comment is stale (raised Minor below).

### Raised

- **BR-41** [Important] `local-verb-network-dependency` Legacy-mode detection requires ls-remote, so offline legacy repos can no longer run local-only verbs
  repositoryTracked calls Repository.Initialized, whose TrunkFile.RemoteExists runs ls-remote; an unreachable remote returns an error, so claim, set-status, issue new and change-code refuse before the legacy path. Reproduced with issue set-status against an unreachable origin. Decide from local evidence first (cutover marker, fetched tracker ref), and treat an offline remote with no evidence as legacy for local verbs; add an unreachable-origin test.
- **BR-42** [Minor] `landing-completion-proof` MigrationAnchor doc still says CodeAfter is not asked of a landed close
  This is the 4th finding in landing-completion-proof, and it is doc drift left behind by the BR-40 fix. The rule already applies in code; only migration.go:36-40 contradicts it. Update the comment to say a landed close checks the branch's code beyond main.

## Open findings

- **BR-41** [Important] `local-verb-network-dependency` Legacy-mode detection requires ls-remote, so offline legacy repos can no longer run local-only verbs
- **BR-42** [Minor] `landing-completion-proof` MigrationAnchor doc still says CodeAfter is not asked of a landed close
