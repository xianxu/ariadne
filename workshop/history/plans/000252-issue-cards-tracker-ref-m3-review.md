# Boundary Review — ariadne#252 (milestone M3)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 6dbd8e24afafc52514f1d041f42e046e7ecd846d..0b7128cf9c58c37dded5bb2cac3c81c3cac2b0c6 |
| command | sdlc milestone-close --issue 252 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-09-25T22:19:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

Most of the M3 design is sound: completion is now an evidence commit plus a `codecomplete` card bound to it, readers compose through one `LoadRecords`, and the M1 landing stages that could never run were removed. The focused suites pass at `0b7128cf`: `internal/{tracker,activetime,fleet,issue,gitx}` and `go test ./cmd/sdlc -run 'Test(Tracker|PublishFlip|DurableLanding|ListIssues|ProjectIssueMeta|ActualTracker|Close|Landing|Publish|Push|Merge|Archive|State|Milestone)'`, the last taking 502s. The boundary is still blocked by one correctness bug. The durable landing marks a close `done` only if its evidence commit is an ancestor of `main`. When a PR lands by squash or rebase, the card never reaches `done`, but its details are archived anyway. The landing fixture explicitly supports squash and rebase, and the plan names squash as a required test. Several Important ordering and lifecycle gaps are also cheap to fix in this round.

**1. Strengths**
- `CodecompleteCard` (`internal/tracker/completeop.go`) and `doneCard` (`trackercompletion.go:74`) are pure card changes that refuse `done` or other statuses rather than overwriting them. `CodecompleteCard` is unit-tested without IO.
- `gitEvidence.Prepare` (`closetracker.go:28`) builds the evidence commit in a temporary index and moves the branch by compare-and-swap. `TestTrackerCloseCommitsEvidenceAndPublishesBoundCard` shows that staged unrelated work is neither committed nor disturbed.
- Card completion bindings are validated at the parse boundary (`issue/handoff.go:67`, full-OID checks), so the git commands in `ownedCompletions` can't receive option-shaped input (ARCH-SECURE).
- Trimming the M1 engine to two stages removes the checkout-local landing receipt that could not follow a close to another clone, and the receipt validation shrank to match. This is a good use of ARCH-FUNERAL and ARCH-ORDER.
- Tracked details are archived byte-for-byte, which keeps the archive's retry proof deterministic. The retry no-op is asserted in `TestDurableLandingArchivesTrackedCloseByBinding`.

**2. Critical findings**
- **Squash or rebase landing strands the card at `codecomplete` while archiving its details.**
  - `landing.go:436` → `settleLandingCompletions` → `ownedCompletions(env, rs, main, "", false)` only accepts evidence that is an ancestor of `main`.
  - After a squash or rebase merge, the evidence OID is never an ancestor of `main`, so the card never goes `done`.
  - Meanwhile `selectTrackedLandingIssues` (`landingarchive.go:541`) selects the same close from `pr.HeadOID --not pr.BaseOID` and archives its details.
  - No later settle can ever match it.
  - This contradicts the plan ("Confirm the merged PR identity/base/head/merge OID (or exact direct-main reachability), publish done"; "Test … squash landing").
  - **Fix sketch:** in the durable path, once `pr.State == MERGED` and `pr.MergeOID` is on `main`, complete the closes selected by the PR's binding with `landed = pr.MergeOID`, instead of relying on ancestry alone. Keep ancestry for direct push. Add merge, squash and rebase cases through `landingFakeGH` to the tracked-landing test (ARCH-PURPOSE).
- **Core concepts rows contradict the code** (3rd finding in family `core-concepts-inventory-drift`).
  - "Activity event selection | PURE | `activetime/commit.go`" shipped as `loadWindowCommits`, which runs `git log`. Its test (`trackerref_test.go`) needs a real repository.
  - "Composed issue records | PURE (+ thin loader)": the card/details join, duplicate handling and unknown-half logic all live inside the IO function `LoadRecords` (glob, `ReadFile`, `ls-remote`/fetch). No pure join function exists.
  - Task 6's Files list still names `merge.go` and `reviewstate.go`, which are unchanged in this window.
  - **The rule, rather than another instance fix:** when a Revisions entry says "planned X shipped as Y", re-derive Y's kind from the shipped code (does its test run without exec/fs/net?) and re-check every Task Files list against `git diff --name-status BASE HEAD`. Here that means relabelling the rows or extracting a pure `composeRecords(cards, details)`, then promoting the loader and `loadWindowCommits` to INTEGRATION.

**3. Important findings**
- **Stale completion receipts can rebind the card to an older review or wedge recovery** (2nd finding in family `resume-identity-check`).
  - On FIX-THEN-SHIP, `publishTrackerClose` saves receipt A. The close help still says "re-run close" for fixes that land after the close commit.
  - A re-run close drives receipt B and never checks for A.
  - A later `recovery reconcile` resumes A. `SetCardCompletion` accepts any new token, so A replaces B's binding with an older reviewed head. If nothing changed, A's Prepare fails "nothing to record" on every reconcile, forever.
  - **Rule:** a new operation of kind K on issue N must refuse or supersede any outstanding K receipt for N, and a resumed receipt must prove it is still the newest generation (for example, the card binding is absent or carries its own token) before any effect (ARCH-ORDER, ARCH-FUNERAL).
- **Tracker-close refusals happen after close's local effects and after the LLM review** (3rd finding in family `refusal-after-local-effect`).
  - `publishTrackerClose` checks for a checked-out branch, card presence, a resolvable reviewed commit and evidence paths inside the checkout.
  - These checks run only after `applyClose` has already written the Log line, the ledger round and the calibration row.
  - **Rule:** every tracker-close precondition that doesn't depend on the verdict runs in `computeClose` before the review is dispatched, and `publishTrackerClose` performs only effects. Enumerate: `env.branch`, the card, `closeEvidencePaths` resolution, and the `MaxEvidencePaths` bound.
- **Tracker reads and card writes on merge/push paths run under `context.Background()`** (2nd finding in family `context-propagation`).
  - The sites are `settleLandingCompletions`, `publishTrackerCompletions`, `selectTrackedLandingIssues`, `overlayCardStatus`, `historyFileIsTerminal`, `actualTrackerInputs`, `collectGitHubIssueNumbers`, `guardIssueNotDone`, `runIssueShow`, `listIssueStates` and fleet `repoRecords`.
  - The plan requires tracker verbs to pass the Cobra context through repository and lock boundaries.
  - **Rule:** any call path reaching `loadIssueRecords`, `openTrackerAt` or `Repository.*` takes the verb's context. Enforce it with a source-guard test (in the style of the `CLIRef` guard) that forbids `context.Background()` in `cmd/sdlc` files calling those entry points.
- **The declared operating envelope ("one tracker fetch per command") is not enforced** (ARCH-CONSTRAINTS).
  - `sdlc push` does an `ls-remote` plus fetch in each of: the publish gate, the not-done scan, the publish flip (twice), the archive scan, and once per archived file in `historyFileIsTerminal`.
  - Pre-tracker downstream repos now pay an `ls-remote` in every one of those scans too.
  - **Fix:** load `tracker.Records` once per command and thread it to the consumers.
- **A FIX-THEN-SHIP deferred evidence commit rebuilds from worktree state that the normal fix-commit habit sweeps away.**
  - Committing the fix with `git commit -a` includes the close Log line, and the ledger too if it is already tracked (for example, from milestone closes).
  - `gitEvidence.Prepare` then refuses "the close changed none of its evidence files", and the receipt can never complete.
  - The test `TestTrackerCloseFixThenShipLandsEvidenceAfterTheFixes` uses `commit -qam` and passes only because the ledger file is still untracked.
  - **Fix:** pin the evidence blobs in the receipt, or snapshot them in a recovery ref when FIX-THEN-SHIP saves the receipt, and build the commit from those blobs. Add a test where the ledger is already tracked (ARCH-ORDER).

**4. Minor findings**
- **Tracker load errors become absence or a legacy fallback** (5th finding in family `silent-error-swallowing`).
  - `collectGitHubIssueNumbers` silently drops `Closes #N` links.
  - `guardIssueNotDone` passes silently, and it uses `PreferFresh` despite the rule that gates authorizing a write use `Fresh`.
  - `actualTrackerInputs` quietly measures without the tracker's claim/close events, which feeds velocity calibration.
  - **Rule:** a `Records` load error in a verb is surfaced (a warning at least, an error in gates), never converted to "no record". That needs a lint over `loadIssueRecords` call sites.
- Fleet `repoRecords` calls `tracker.RepositoryFor(ctx, repoRoot, "main")` with a hard-coded resting branch, duplicating `recordsRepository`'s workspace resolution (2nd finding in family `shared-helper-extraction`). The shared helper should live in `internal/tracker`.
- `runRecoveryReconcile`'s deferred settle mutates every landed card, not only `--issue N`, and still runs after an earlier error. It resolves `WF_ISSUES_DIR` against the working directory rather than `env.root`.
- Several files place the module import inside the stdlib import group (`close.go`, `publishgate.go`, `push.go`, `pr.go`, `repoguard.go`, `actual.go`, `landingarchive.go`, `issuerecovery.go`, `trackercompletion_test.go`); goimports would regroup them.

**5. Test coverage notes**
- `doneCard` has no direct unit test. Only the stale-token case is covered, through integration.
- There is no test that `ownedCompletions` refuses a binding from another repository, or that a malformed binding fails a publish — both are named in the function-level contract.
- The tracked landing test calls `settleLandingCompletions` directly on a fast-forward. It doesn't go through `runMerge` with `landingFakeGH`, which is why the squash/rebase bug slipped through (ARCH-MOCK: the fake exists but the path bypasses it).
- There is no sequence test for FIX-THEN-SHIP followed by a re-close, then reconcile.

**6. Architectural notes for upcoming work**
- M4's migration must reconstruct bindings that the landing path can complete under every merge strategy. Fix the landing proof first, or migrated squash-merged PRs inherit the same stuck state.
- A per-command `Records` value, passed down with its context, fixes the fetch budget and the context-propagation problem together. It also gives M4's cutover-marker check a single place to live.

ARCH walk:
| Principle | Result |
|---|---|
| ARCH-DRY | Flag (fleet helper; per-call record loads) |
| ARCH-PURE | Flag (join logic inside the IO loader); card changes pass |
| ARCH-PURPOSE | Flag (squash landing) |
| ARCH-MOCK | Pass, with a note (the fake is bypassed for the tracked landing) |
| ARCH-CONSTRAINTS | Flag |
| ARCH-SECURE | Pass |
| ARCH-ORDER | Flag (stale receipts; deferred evidence inputs) |
| ARCH-FUNERAL | Pass, except that unreconciled FIX-THEN-SHIP receipts persist with no superseding rule (covered by the stale-receipt finding) |

**7. Plan revision recommendations**
- Add a Revisions entry that relabels "Activity event selection" and "Composed issue records" as INTEGRATION (or records the extracted pure join), and revises Task 6's Files list to what changed.
- Add a Revisions entry for the landing proof: `done` is derived from the confirmed PR merge OID for durable landings and from ancestry for direct push, with tests for all three merge strategies.
- State the one-outstanding-receipt-per-(issue, operation) rule in "Completion and recovery".

```findings
findings:
  - id: new
    severity: Critical
    family: landing-completion-proof
    title: |
      Squash/rebase-landed PR never flips its card to done, yet its details are archived
    detail: |
      settleLandingCompletions accepts only evidence-ancestor-of-main; selectTrackedLandingIssues selects by PR head/base and archives. Under squash/rebase (supported by landingFakeGH) the card stays codecomplete forever. Derive done from the confirmed pr.MergeOID in the durable path; test merge/squash/rebase through runMerge.
  - id: new
    severity: Critical
    family: core-concepts-inventory-drift
    title: |
      Core concepts label IO entities PURE (loadWindowCommits, LoadRecords join) and Task 6 Files lists unchanged merge.go/reviewstate.go
    detail: |
      3rd in family. Rule: when a Revisions entry renames a planned entity, re-derive its kind from the shipped code and its test IO, and re-check every Task Files list against name-status. Relabel as INTEGRATION or extract a pure composeRecords.
  - id: new
    severity: Important
    family: resume-identity-check
    title: |
      Outstanding completion receipt is neither refused nor superseded by a re-close; resume can rebind the card to an older review
    detail: |
      2nd in family. FIX-THEN-SHIP saves receipt A; re-running close drives B; reconcile later resumes A and SetCardCompletion accepts any new token, or A wedges on nothing-to-record. Rule: one outstanding receipt per (issue, operation); a resume proves it is still the newest generation before effects.
  - id: new
    severity: Important
    family: refusal-after-local-effect
    title: |
      publishTrackerClose preconditions (branch, card, reviewed commit, evidence paths) run after applyClose wrote the Log, ledger and calibration and after the review
    detail: |
      3rd in family. Rule: every verdict-independent tracker-close precondition runs in computeClose before review dispatch; publishTrackerClose only performs effects.
  - id: new
    severity: Important
    family: context-propagation
    title: |
      Tracker reads and card CAS writes on merge/push/landing paths use context.Background()
    detail: |
      2nd in family. Sites: settleLandingCompletions, publishTrackerCompletions, selectTrackedLandingIssues, overlayCardStatus, historyFileIsTerminal, actualTrackerInputs and other readers. Rule: any path reaching loadIssueRecords/openTrackerAt/Repository takes the verb context; enforce with a source-guard test.
  - id: new
    severity: Important
    family: operating-envelope-enforcement
    title: |
      Push/merge fetch the tracker many times per command (plus once per archived file) against the declared one-fetch budget
    detail: |
      Load tracker.Records once per command and thread it to the publish gate, not-done scan, flip, archive and historyFileIsTerminal; pre-tracker repos also pay repeated ls-remote today.
  - id: new
    severity: Important
    family: deferred-effect-input-drift
    title: |
      A FIX-THEN-SHIP evidence commit rebuilds from worktree files that a fix commit (git commit -a) sweeps away, wedging the receipt
    detail: |
      Prepare refuses when no evidence file differs from HEAD; the test passes only because the ledger stays untracked. Pin evidence blobs in the receipt or a recovery ref when it is saved.
  - id: new
    severity: Minor
    family: silent-error-swallowing
    title: |
      Tracker load errors become absence in collectGitHubIssueNumbers, guardIssueNotDone and actualTrackerInputs
    detail: |
      5th in family. Rule: a Records load error in a verb is surfaced (warn, or error in gates), never converted to no record; guardIssueNotDone should also use Fresh. Needs a lint over loadIssueRecords call sites.
  - id: new
    severity: Minor
    family: shared-helper-extraction
    title: |
      fleet repoRecords hard-codes the main resting branch, duplicating recordsRepository
    detail: |
      2nd in family. Move the resting-branch-aware repository resolution into internal/tracker and use it from both.
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      Module imports placed inside the stdlib import group in several changed files
```

---

## Re-review — 2026-09-25T23:18:40-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 6dbd8e24afafc52514f1d041f42e046e7ecd846d..98ac562f875ef16ea85ec915b9bba9c6c6a5aa2c |
| command | sdlc milestone-close --issue 252 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-09-25T23:18:40-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The prior round's ten findings are handled; nine are fully fixed. BR-26's main rule holds, but two parts are still open (details below). Slot landings now complete a PR's closes at its `MergeOID`, whatever the merge strategy. Evidence blobs are pinned in the receipt and kept alive by the recovery ref. A re-close replaces an earlier close that never started and refuses one that is half done, and every resume first checks that no newer close owns the card. The branch and card checks now run in `computeClose`, before the review. Verb contexts reach tracker reads, with a source-guard test, and each command shares one records scope. Two new Important problems remain, both introduced by this round's fixes:

1. **Abandoned branch marked done.** The non-durable `sdlc merge` completes cards even when nothing merged.
2. **Deferred evidence reverts later edits.** A FIX-THEN-SHIP evidence commit overwrites HEAD's newer copies of the details and plan files with the blobs pinned at close.

Test evidence:
- **Targeted tests pass** (97s): the new BR-19/21/22/23/24/25/26 tests plus `TestDurableLandingArchives*`.
- **Other packages:** all pass except `internal/processgroup`. Its one failure is the sandbox blocking `/bin/ps`, not this change.
- **Full `cmd/sdlc` package:** did not finish. It hit Go's default 600s limit, and a rerun with a 40m timeout was still going when this was written.

## Strengths
- `completeLandingPR` (`trackercompletion.go:154`) picks owned closes by `HeadOID --not BaseOID` and records `MergeOID`. `TestDurableRunMergeCompletesTrackedCloseForEveryStrategy` runs `runMerge` under merge, squash and rebase, and checks that a retry leaves the card unchanged.
- `CompletionOp.newestClose` (`completeop.go:47`) guards both the evidence stage and the card stage. Its test covers older, newer and same-token cards.
- `prepareTrackerClose` (`closetracker.go:190`) runs before the review, and `TestTrackerCloseRefusesPreconditionsBeforeReview` asserts zero judge calls and a clean tree.
- The records scope (`issuerecord.go:67`) never lets a stale cached read answer a Fresh request, and card writes invalidate it.
- `composeRecords` is a pure join, and the plan's concepts table now labels the loaders INTEGRATION.

## Critical
None.

## Important
1. **`merge.go:575` — cards go done on a branch that never merged.** Step 10.5 runs whenever `trackerEnvForMerge != nil`. In the worktree topology with no PR and unmerged commits, the operator can answer "n" to creating a PR and "y" to "Remove worktree without merging". Control then falls through, and `completeOnCard` writes `done` with `LandedCommit` = main HEAD, which does not contain the work. Before #252 this was harmless, because the old flip only touched files already on main.
   - This is the 2nd finding in family `landing-completion-proof`. The rule: a card goes done only on a confirmed observation that its evidence landed.
   - Fix: gate completion on `merged`. Also add a test for the abandon path.
2. **`closetracker.go:50-63` — deferred evidence overwrites HEAD.** `Prepare` reads HEAD's tree, then writes each pinned blob over its path. After FIX-THEN-SHIP, the agent logs the outcome in the details `## Log`, per the constitution, and commits it with the fixes. Reconcile's evidence commit then puts back the details (and any ledgers) as they were at close. Because `Apply` only resets the index, the worktree keeps the newer bytes, so the revert shows up only as a dirty file.
   - This is the 2nd finding in family `deferred-effect-input-drift`. The rule: an effect that replays pinned inputs is a three-way merge against the base they were pinned from.
   - Fix: record the base blob for each path. At `Prepare`, if HEAD's blob equals the pinned blob or the base, proceed. Otherwise refuse and say which path, or keep HEAD's version. Test with a fix commit that edits the details' Log.

## Minor
- **`internal/fleet/issues.go:71` — fleet ignores the verb context.** `repoRecords` still uses `context.Background()` with `PreferFresh`, and the context guard only parses package `main`. This is the 3rd finding in family `context-propagation`. The rule: every exported tracker entry point takes a context, and the guard should cover `internal/...` too.
- **`repoguard.go:91-93` — stale doc comment.** It says an unreadable record "is left to the verb's own error path", but the code now refuses when the record can't be read (fails closed).
- **`repoguard.go:99` — stale read in the done guard.** It still reads with `PreferFresh`, which BR-26 asked to change to `Fresh`.
- **`merge.go:576` — landed commit may be wrong.** On the non-durable path, the landed commit is main HEAD after the pull, not the PR's merge commit.
- **Plan checklist names functions that don't exist.** Ticked rows in Task 5 and Task 6 name `FinalizeTrackerClose`, `SelectCompletedIssues`, `SelectActivityEvents` and `ReadIssueRecord`. Only `StepCompletion` and `LookupRepoIssues` exist. This is the 4th finding in family `core-concepts-inventory-drift`. The rule: every identifier in a ticked plan row must grep to shipped code, or be renamed in a Revisions entry.

## Test coverage notes
- The abandon path and a fix commit that edits the evidence files have no tests. Both Important bugs sit in these gaps.
- The fetch budget is tested only at the unit level (`TestRecordsScopeFetchesOncePerCommand`). No command-level test counts fetches.

## Architectural notes
- **ARCH-DRY:** pass. `RepositoryForCheckout` is shared, and records use `issue.FilenamePattern`.
- **ARCH-PURE:** pass. `composeRecords`, `doneCard` and `CodecompleteCard` are pure.
- **ARCH-PURPOSE:** pass.
- **ARCH-MOCK:** pass. The landing fake models all three merge strategies.
- **ARCH-CONSTRAINTS:** pass. Fetches are bounded to one per command plus one per card write.
- **ARCH-SECURE:** pass. Evidence entries are checked (path and OID) when the receipt is read.
- **ARCH-ORDER:** flagged. Important 1 lets the abandon path complete cards without a landing, and Important 2 lets the deferred evidence commit replay stale input.
- **ARCH-FUNERAL:** pass. Pinned blobs live and die with the recovery ref, and the temporary index directory is removed.

## Plan revision recommendations
- Add a Revisions entry that renames the Task 5 and Task 6 checklist identifiers to the functions that shipped: `CompletionStepper`/`CompletionOp`, `ownedCompletions`/`completeLandingPR`, `loadWindowCommits`, `loadIssueRecords`.

```findings
dispose:
  - id: BR-19
    disposition: addressed
    note: |
      completeLandingPR completes by HeadOID..BaseOID at MergeOID; runMerge test covers merge/squash/rebase plus idempotent retry, and the targeted run passed.
  - id: BR-20
    disposition: addressed
    note: |
      Plan rows 30/52/53 relabelled (composeRecords PURE, loaders INTEGRATION); Task 5/6 Files lists revised against name-status.
  - id: BR-21
    disposition: addressed
    note: |
      prepareTrackerClose supersedes discardable closes and refuses started ones; newestClose guards both stages; TestTrackerReCloseSupersedesAnUnstartedClose.
  - id: BR-22
    disposition: addressed
    note: |
      Branch, card and pending-close checks moved into computeClose; TestTrackerCloseRefusesPreconditionsBeforeReview asserts zero judge calls.
  - id: BR-23
    disposition: addressed
    note: |
      Contexts threaded through package main with the TestVerbContextsReachTrackerReads guard; the fleet leftover is raised as a new Minor.
  - id: BR-24
    disposition: addressed
    note: |
      recordsScope gives one composed view per repo per command, invalidated on card writes; unit test TestRecordsScopeFetchesOncePerCommand.
  - id: BR-25
    disposition: addressed
    note: |
      Evidence pinned as blobs held by the recovery ref; an empty commit is allowed; TestTrackerCloseFixThenShipSurvivesASweepingFixCommit. Pinned blobs now overwrite HEAD, raised as new.
  - id: BR-26
    disposition: addressed
    note: |
      Load errors surfaced (warn, fail-closed guard, actual warning) and the regex guard extended. Residual: guardIssueNotDone still uses PreferFresh and its doc comment is stale.
  - id: BR-27
    disposition: addressed
    note: |
      tracker.RepositoryForCheckout is shared by fleet and recordsRepository.
  - id: BR-28
    disposition: addressed
    note: |
      Mixed import groups in actual.go, pr.go and trackercompletion_test.go are fixed.
findings:
  - id: new
    severity: Important
    family: landing-completion-proof
    title: |
      Non-durable merge completes owned cards to done even when the branch was abandoned unmerged
    detail: |
      merge.go:575 runs completeOnCard whenever trackerEnvForMerge is non-nil. On the worktree path with no PR, where the operator answers remove-without-merging, the card goes done with LandedCommit set to a main HEAD that lacks the evidence. 2nd in family. Rule: a card goes done only on a confirmed landing observation of its evidence. Gate on merged and test the abandon path.
  - id: new
    severity: Important
    family: deferred-effect-input-drift
    title: |
      Deferred evidence commit overwrites HEAD's newer details, ledgers and sidecars with the blobs pinned at close
    detail: |
      gitEvidence.Prepare reads HEAD's tree and then writes pinned blobs over it. A post-FIX-THEN-SHIP Log edit committed with the fixes is reverted in the evidence commit, visible only as a dirty worktree. 2nd in family. Rule: replaying pinned inputs is a three-way merge against their base. Record the base blob, and refuse or keep HEAD when HEAD's version differs from both.
  - id: new
    severity: Minor
    family: context-propagation
    title: |
      fleet repoRecords still reads the tracker under context.Background(); the guard only parses package main
    detail: |
      3rd in family. Rule: every exported tracker entry point takes a context, and the source guard should scan internal packages too.
  - id: new
    severity: Minor
    family: core-concepts-inventory-drift
    title: |
      Ticked Task 5/6 plan rows name functions that do not exist
    detail: |
      FinalizeTrackerClose, SelectCompletedIssues, SelectActivityEvents and ReadIssueRecord are absent from the code. 4th in family. Rule: every identifier in a ticked plan row greps to shipped code, or is renamed in a Revisions entry.
```

---

## Re-review — 2026-09-25T23:59:34-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 6dbd8e24afafc52514f1d041f42e046e7ecd846d..bf7d3e4a337cd4c656d0e12edf150a7ce6794693 |
| command | sdlc milestone-close --issue 252 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-09-25T23:59:34-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

All four open findings from the earlier rounds are fixed in this window. Each behavior change has a regression test that would fail without the fix, and the targeted tests pass: `go test` on `internal/tracker` and `internal/fleet`, plus the root-package run filtered to `TestVerbContexts|TestAbandonedWorktree|TestTrackerCloseFixThenShip|TestDurableRunMerge|TestTrackerReClose`. The full `./...` suite did not run: the first attempt was cut off by a shell glob error, and I only re-ran the targeted packages. Nothing blocks the M3 boundary. I found one new Minor issue: the three-way evidence replay skips files silently, which has no effect on correctness.

1. **Strengths**
   - `cmd/sdlc/merge.go:577`: completion now requires `merged`, which is true only for a PR merged now or one found merged on resume. `TestAbandonedWorktreeMergeLeavesTheCloseCodecomplete` runs the real no-PR, remove-without-merging path and checks that the card is unchanged.
   - `internal/tracker/receipt.go` `EvidenceEntry.Replays` keeps the three-way decision as a pure function with a nine-case table test. The git plumbing stays in `gitEvidence.treeEntry`, which is the right split under ARCH-PURE.
   - `context_guard_test.go` now scans `package main` and every `internal/*` package, with package-qualified allowlist entries. The rule "every exported entry point takes the caller's context" is now checked by a test, not just reviewers.
   - The plan's Revisions entry and its ticked rows now name the functions that actually shipped. `LoadRecords`, `loadWindowCommits`, `publishTrackerClose` and `ownedCompletions` all exist in the code.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - `closetracker.go` `Prepare`: when HEAD's copy of a pinned file differs from both the reviewed copy and the pin, the file is skipped and nothing is logged. This is the 3rd finding in family `deferred-effect-input-drift`. The rule: every deferred replay reports what it did not apply, so a later edit that also dropped the close's evidence (for example a ledger row) can be seen, not just survive quietly.
   - `repoguard.go` `guardIssueNotDone` now reads `Fresh`, so running start-plan or change-code offline on a repo with a tracker is refused. That is intended (it fails closed), but the refusal message could say that a network read is required.

5. **Test coverage**
   - BR-29: the abandon-path end-to-end test is the regression test.
   - BR-30: `TestTrackerCloseFixThenShipKeepsALaterEditOfAPinnedFile` covers the end-to-end case and `TestEvidenceReplaysOnlyOverUntouchedFiles` covers the pure logic.
   - BR-31: the widened source guard is the regression test.

6. **Architecture**
   - **ARCH-DRY:** pass.
   - **ARCH-PURE:** pass. `Replays` is pure.
   - **ARCH-PURPOSE:** pass.
   - **ARCH-MOCK:** pass. The tests use real git fixtures and the fake `e2eGH`.
   - **ARCH-CONSTRAINTS:** pass.
   - **ARCH-SECURE:** pass. Reading `ls-tree` output checks the entry is a file (`blob`) and refuses anything else.
   - **ARCH-ORDER:** pass. Completion is gated on a confirmed landing observation.
   - **ARCH-FUNERAL:** pass. No new durable artifact families.
   - For later work: a worktree branch that already landed in main without a PR stays `codecomplete` rather than `done`. That errs on the safe side, and reconcile can finish it.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-29
    disposition: addressed
    note: |
      merge.go:577 gates completeOnCard on merged; TestAbandonedWorktreeMergeLeavesTheCloseCodecomplete exercises the no-PR abandon path and would go red without the gate.
  - id: BR-30
    disposition: addressed
    note: |
      EvidenceEntry.Replays(base, head) three-way check keeps HEAD when changed since review; pure table test plus e2e TestTrackerCloseFixThenShipKeepsALaterEditOfAPinnedFile.
  - id: BR-31
    disposition: addressed
    note: |
      LookupRepoIssues/repoRecords/CollectInventory take ctx; context_guard_test now parses every internal package with qualified allowlist.
  - id: BR-32
    disposition: addressed
    note: |
      Plan rows 143-145, 207-217 now name shipped identifiers with planned names in parentheses, and Revisions line 359 records the rename; all shipped names grep to code.
findings:
  - id: new
    severity: Minor
    family: deferred-effect-input-drift
    title: |
      Evidence replay silently skips pinned files whose HEAD version diverged from both base and pin
    detail: |
      3rd in family. Rule: a deferred replay reports every pinned input it did not apply (path plus reason), so a later commit that also dropped close evidence such as a ledger row is visible rather than silently kept.
```
