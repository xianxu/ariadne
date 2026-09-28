# Boundary Review — ariadne#252 (milestone M2)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 547122d280ebbead8d4127e11d01eda5ff863be7..58c9c84afc5d5ae6579c60ebcef736abf5929e3a |
| command | sdlc milestone-close --issue 252 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-25T17:39:22-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

**M2 boundary review: #252, window `547122d2` → `58c9c84a`, verdict REWORK**

M2 does what its Plan claims: `issue new`, `claim`, `start-plan`, `change-code`, the card setters and `move-detail` now run on the tracker. One `tracker.Drive` loop owns every effect. Tests use real Git remotes, including two-clone races, interrupted pushes via pre-push hooks, fresh-clone guard replay, and the A→B→R vs A→D→E net-zero landing. `go vet`, the internal packages (tracker, gitx, issue) and the focused cmd/sdlc run (`Test(MoveDetail|Transfer|IssueRecovery|IssueNew|PreparePlanning|StartPlan|ChangeCode|Claim|SetStatus|PublishGate|Card|Tracker)`) all pass. `git diff --check` is clean.

One thing blocks SHIP. Recovery receipts are stored as refs in `.git`'s common directory, so every linked worktree (every slot) sees them. `recovery reconcile` never checks that it is running on the receipt's source branch. It can therefore remove the source, fast-forward, or write details in the wrong slot, then delete the receipt and print "finished". Four Important findings also need a cheap fix before the boundary.

### 1. Strengths
- **One dispatcher, persist first:** `tracker.Drive` (`internal/tracker/drive.go:57`) is the only place effects run. It saves the receipt before every protected effect and after any uncertain observation (ARCH-ORDER, ARCH-DRY).
- **Idempotent removal:** `moveDetailRemover` (`issuemovedetail.go:258`) refuses a source whose bytes changed and stays correct when re-run after a partial removal. `TestMoveDetailInterruptedRecordIsFinishedByReconcile` interrupts the record stage for real and proves main is not republished.
- **One-handoff rule:** `SetCardHandoff` (`internal/issue/handoff.go:80`) refuses a foreign handoff record while letting the same operation complete its own. That serialises concurrent `move-detail` runs across clones through card CAS.
- **Pure decision core:** `statusDecision` (`setstatus.go`) and `cardUpdate` separate the decision from IO cleanly (ARCH-PURE).
- **Planning branch:** `preparePlanningBranch` refuses a dirty rest, a rest that is ahead of or diverged from main, and unrelated branches. It never moves the resting ref, and its checkout-relation test covers those combinations.

### 2. Critical findings
- **`issuerecovery.go:118-127`, `tracker/store.go`, `gitx/recoveryref.go:42` — cross-worktree reconcile.**
  - The code comment says receipts are "checkout-local". In fact `git rev-parse --git-path refs/sdlc/recovery/x` resolves to the common dir (`/Users/xianxu/workspace/ariadne/.git/...`), so every slot shares them.
  - Reconcile builds `moveDetailRemover(env)` and `NewCreationOp(..., env.checkout(...))` from the *current* checkout and never compares `env.branch` with `spec.SourceBranch`.
  - Failure: slot1 is on a feature branch and its `move-detail` is interrupted after `transfer.record`. An agent in slot2, on `main-slot2`, runs `sdlc issue recovery reconcile --issue N` because `recovery list` shows it there too.
    - The remover finds no file, sees an untracked path on rest, and fast-forwards `main-slot2`.
    - It then reports Confirmed and deletes the receipt. Slot1's source is never removed and the recovery record is gone.
    - A resumed creation would likewise write details into the wrong worktree.
  - `move-detail`'s "unfinished operation here" refusal also fires from every slot.
  - Fix:
    - Local-effect stages (`transfer.remove`, `creation.detail`) must require the current worktree's branch to equal `spec.SourceBranch`. Otherwise refuse and name the owning branch or worktree.
    - `receiptsFor` / `recovery list` should mark receipts owned by another branch.
    - Add a two-worktree test.

### 3. Important findings
- **`claim.go:120`, `issue.go:334`, `issue.go:346`, `internal/tracker/createop.go:184` — zero-padded IDs in next-action commands.**
  - These print `--issue 000253`. pflag parses ints with `strconv.ParseInt(s, 0, 64)`, so a leading 0 means octal: `000253` becomes 171, and `000258` is a parse error.
  - An agent copying the hint runs `move-detail` or `reconcile` against a *different issue*.
  - The tests pin the broken strings (`claimremote_test.go:47`, `issuenew_test.go:47`).
  - Other sites already use `strings.TrimLeft(id, "0")`, so the formatting is scattered.
  - Fix: one helper for issue IDs in commands (ARCH-DRY), used at every site, plus a test that feeds a rendered hint back through the flag parser.
- **`transferguard.go:40` — the guard skips handoffs whose `main_commit` is not yet recorded.**
  - The handoff record is published *before* main publication. An interruption between `transfer.main` and `transfer.record` leaves D on main while the guard ignores the path.
  - That is exactly the window receipts exist for. If the source branch runs `sdlc pr` before reconcile and the owner has edited E, the result is an add/add conflict whose "keep ours" resolution loses E — the pair#168 pattern.
  - `TestTransferGuardInterruptedRemovalProtectsOwnerEdits` reconciles *before* guarding, so this state is never tested.
  - Fix: also include handoffs without `MainCommit` when the destination exists on main (or when a commit with the token trailer is reachable), and test the guard before reconcile.
- **`helptext/claim.md`, `issue.md`, `change-code.md:51`, `start-plan.md:35,55` — help for the changed verbs describes the old behaviour.**
  - `claim.md` still says it reserves "on origin/main" and recommends `issue sync` / `issue publish`.
  - `issue.md`'s subcommand list lacks `move-detail`, `recovery`, `set-title`, `set-estimate` and `set-github`.
  - `start-plan` and `change-code` help still point at `issue sync`, contradicting `syncPointer` and the new `change-code`.
  - The binary's help is the workflow contract agents read, and README only adds a pointer. M4's instruction sweep covers `AGENTS.base.md` and skills, not per-verb help for verbs M2 already changed.
- **Plan Core concepts table and Task 3 file list — second finding in family `core-concepts-inventory-drift`.**
  - Task 3 still says "add `issuerecord.go`, `issuemetadata.go`". Neither exists: the setters shipped as `cardsetters.go`.
  - These new M2 entities appear in no table row: `internal/issue/handoff.go` and `cardset.go` (both PURE), `planningbranch.go` (`PreparePlanningBranch` is named in the function contract), `trackerenv.go`, `issuerecovery.go`, `internal/tracker/candidates.go`.
  - Per the escalation, fix the class, not this instance. Rule: at each boundary, every non-test file added in the window (`git diff --name-status BASE HEAD | grep '^A' | grep -v _test`) appears in a Core concepts row or is explicitly classified as glue, and every Task "Files:" entry either exists or is revised. Run that enumeration before submitting each milestone.

### 4. Minor findings
- `transferguard.go`, `pr.go`, `publishgate.go`, `landinggate.go` call the guard with `context.Background()`, not the Cobra context. PQ-1 promised command-context propagation.
- `createop.go`'s `publishEvent` / `probeEvent` discard Push/Probe errors (`outcome, _ :=`). An auth or hook failure surfaces only as a bare "outcome uncertain", with no diagnostic.
- `cardsetters.go`: a details parse error silently becomes an empty body, which the set-status reopen guard then reads as "no Log". It should report the error.
- Handoff `destination` is validated only as a single-line value. It should be validated as `<issues dir>/<card basename>`, because the guard uses it both as an exemption key and as a path.
- One malformed card anywhere makes the guard fail every PR and push repo-wide. That is fail-closed, but the error should say it is repo-wide.
- `start-plan` now always needs the tracker and network, while `change-code` keeps a legacy path for files without a mirror. This inconsistency is acceptable under the freeze-and-cutover plan, but should be noted there.
- `issue.go:46`'s comment still references the deleted `applyStatus`.

### 5. Test coverage notes
- Strong real-Git coverage for creation races, lost acknowledgements, the transfer net-zero landing, rest fast-forward, card-derived details, and the fresh-clone guard.
- Missing:
  - reconcile run from a second worktree or another branch (the Critical finding);
  - the guard before reconcile in the interrupted-record state;
  - a test proving next-action hints parse to the same issue ID;
  - `move-detail` refusing (or refreshing) a hand-edited mirrored field before publishing to main. It currently only checks `HasMirror`, so a hand edit to `status` lands on main unrefused. Spec: "refused by operations consuming/updating details".

### 6. Architectural notes (ARCH-*)
- **ARCH-DRY:** flag. Issue-ID formatting for commands is duplicated inconsistently (Important above). Otherwise publication reuse through `publisher` is good.
- **ARCH-PURE:** pass. Transitions are pure; the adapters are thin.
- **ARCH-PURPOSE:** mostly pass, with the M2 help surface unfinished (Important above).
- **ARCH-MOCK:** pass. Real Git remotes plus the M1 stateful fake.
- **ARCH-CONSTRAINTS:** pass with a note. `sdlc pr`/push now fetch the tracker and main on every call, and `start-plan` gained a network dependency.
- **ARCH-SECURE:** minor flag on handoff `destination` validation.
- **ARCH-ORDER:** flag. The Critical is a missing identity/ownership check before a local effect resumed from a shared receipt. The guard also ignores the in-flight state.
- **ARCH-FUNERAL:** pass. Receipts are released on finalization or discard, handoffs are bounded at one per card, and `recovery list` surfaces what remains.
- **For M3:** the composed reader should own "which checkout and branch holds these details", so the ownership check does not stay scattered.

### 7. Plan revision recommendations
- `## Revisions` entry for M2 inventory: replace Task 3's `issuerecord.go`/`issuemetadata.go` with `cardsetters.go` (the composed reader stays deferred), and add rows for `handoff.go`, `cardset.go`, `planningbranch.go`, `trackerenv.go`, `issuerecovery.go` and `candidates.go`. Record the added-file enumeration rule.
- Correct the plan's durable-residue text: recovery refs are shared across linked worktrees, not checkout-local, and resuming requires the matching source branch.

```findings
findings:
  - id: new
    severity: Critical
    family: resume-identity-check
    title: |
      recovery reconcile drives local-effect stages from any worktree; receipts are shared across slots
    detail: |
      refs/sdlc/recovery lives in the git common dir, so every slot lists every receipt. runRecoveryReconcile builds moveDetailRemover/CreationOp from the current checkout and never compares env.branch with spec.SourceBranch. Another slot can fast-forward its own rest, report Confirmed, and delete the receipt while the real source is never removed. Require the source branch for transfer.remove and creation.detail, mark receipts owned elsewhere in list, and add a two-worktree test.
  - id: new
    severity: Important
    family: next-action-hint-correctness
    title: |
      next-action hints print zero-padded --issue 000NNN, which pflag parses as octal (wrong issue)
    detail: |
      claim.go:120, issue.go:334, issue.go:346, createop.go:184. Example: 000253 parses as 171, and 000258 fails to parse. The tests pin the broken strings. Use one ID-formatting helper at every site and test that rendered hints parse back to the same ID.
  - id: new
    severity: Important
    family: transfer-guard-coverage
    title: |
      transfer guard skips handoff records without main_commit, missing the interrupted main-published window
    detail: |
      The handoff record is published before main. An interruption between transfer.main and transfer.record leaves D on main unguarded, and a PR from the source branch before reconcile reaches the add/add keep-ours loss. Include records whose destination exists on main, and test the guard before reconcile.
  - id: new
    severity: Important
    family: changed-verb-docs-drift
    title: |
      help for claim/issue/start-plan/change-code still describes origin/main publication and issue sync
    detail: |
      claim.md says it reserves on origin/main and recommends issue sync/publish. issue.md lacks move-detail, recovery and the set-* setters. start-plan.md and change-code.md point at issue sync, contradicting the new behaviour. The help is the agent workflow contract.
  - id: new
    severity: Important
    family: core-concepts-inventory-drift
    title: |
      Task 3 files and Core concepts table omit or misname M2 entities (2nd in family)
    detail: |
      Task 3 still lists issuerecord.go and issuemetadata.go; neither exists (the setters shipped as cardsetters.go). No table row covers handoff.go and cardset.go (PURE), planningbranch.go, trackerenv.go, issuerecovery.go or candidates.go. Rule: every non-test file added in the boundary window appears in a Core concepts row or is classified as glue, and every Task Files entry exists or is revised. Enumerate with git diff --name-status BASE HEAD before each milestone submission.
  - id: new
    severity: Minor
    family: mirror-validation-coverage
    title: |
      move-detail publishes local details to main after only a HasMirror check (no hand-edit refusal)
    detail: |
      A hand-edited mirrored field (for example status) in local details lands on main unrefused. Run refreshMirror before publication.
  - id: new
    severity: Minor
    family: silent-error-swallowing
    title: |
      publishEvent/probeEvent discard Push/Probe errors; cardsetters swallows details parse errors
    detail: |
      Uncertain outcomes lose their diagnostic (auth or hook failures), and a malformed details body becomes an empty Log for the reopen guard.
  - id: new
    severity: Minor
    family: context-propagation
    title: |
      transfer guard is invoked with context.Background() from pr/publish/landing gates
  - id: new
    severity: Minor
    family: untrusted-input-validation
    title: |
      handoff destination is not validated as issues-dir plus card basename
    detail: |
      The guard uses it as both an exemption key and a path. One malformed card also fails every PR repo-wide without saying so.
```

---

## Re-review — 2026-09-25T18:16:10-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 547122d280ebbead8d4127e11d01eda5ff863be7..32ce246ce40307d3b81aff2fe5e1141834736416 |
| command | sdlc milestone-close --issue 252 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-25T18:16:10-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

Round 3 fixed the blocking problem. Recovery now refuses to run another checkout's local effects. The guard sits in both adapters and in the `reconcile` preflight, and a real two-worktree test covers it. The octal `--issue` hints, the unguarded window between main publication and card record, the plan inventory and the Minors are also fixed, and each behaviour change has a test that fails without it. Two cheap items remain:

- **The docs fix is incomplete.** One false sentence is left in `issue.md`, and three atlas pages still describe how `claim` and `change-code` worked before M2.
- **The BR-9 fix introduced a regression.** I reproduced it. `move-detail` rewrites a stale mirror on disk before its dry-run check and its preconditions. That dirties the committed file, and then the command refuses with "nothing was changed". `change-code --dry-run` has the same write-before-check ordering.

**Strengths**
- Local effects are bound to the receipt's source checkout. `requireSourceCheckout` (`internal/tracker/store.go:133`) is called from both `CreationOp.materialize` and `TransferOp.removeSource`, and `reconcile` refuses early (`issuerecovery.go:104`). `TestRecoveryReconcileRefusesAnotherWorktree` checks that the other worktree's HEAD doesn't move, the source isn't touched, and a later reconcile from the right checkout succeeds.
- `issue.CLIRef` is the single place IDs are rendered for hints. `issuehint_guard_test.go` enforces that across the source tree and checks the round-trip with base-0 parsing, so the rule is covered as a class, not one site.
- `TestTransferGuardProtectsAnUnrecordedPublication` uses the real interrupted-handoff fixture. It checks both sides: an identical unfinished copy passes, and an owner edit on main is refused.
- The diagnostics mixin (`createop.go:219`) keeps the last Git error without changing any event's meaning, and `TestUncertainStopReportsTheGitError` pins it.
- The inventory rule is written into the plan's Revisions, and I confirmed it holds: all 17 non-test files added in the window appear in the Core concepts tables.

**Critical:** none.

**Important**
1. **Mirror refresh writes before the dry-run and precondition checks** (`issuemovedetail.go:159-171`; `changecode.go:130` → `:307`).
   - **Reproduced:** in a scratch copy, `spinOffOnCodeBranch`, then `set-title` from another worktree, then `move-detail --dry-run`. The dry run rewrote the committed details file. Both the dry run and the real run then refused with "has unstaged changes over a staged version … nothing was changed".
   - **Impact:** any card change made from another checkout (`set-title`, `set-estimate`, `set-github`) blocks `move-detail` with a false message, and `--dry-run` breaks its "change nothing" contract. `change-code --dry-run` also writes the refreshed mirror before its `DryRun` check at `:196`.
   - **Fix:** compute the refreshed bytes in memory, run `checkMoveSource` on the original file, return early for dry run, and only then write and `git add` the refreshed file so index and worktree agree. In `change-code`, skip the write under `DryRun`.
   - **Rule:** checks run before side effects, and any refusing or dry-run path leaves the files unchanged.

2. **BR-7 is only partly fixed** (disposed `not-addressed` below):
   - `helptext/issue.md:119` still says "`change-code` similarly publishes its new narrow issue checkpoint". Since M2, `change-code` never publishes, including for pre-tracker details.
   - `atlas/workflow/issue-sync.md:51` says the same.
   - `atlas/workflow/issue-lifecycle.md:34` and `atlas/workflow/process-manual.md:111` still describe `claim` as reading and broadcasting to `origin/main`.

**Minor**
- `transferguard.go:62`: any `git cat-file -e` failure, not just a missing path, is treated as "not on main", so the guard fails open on a Git error. This belongs to the `silent-error-swallowing` family; see the findings block for the rule.
- `"refs/heads/"+env.branch` is built by hand at four sites (`issuerecovery.go:62,102,137`, `issuemovedetail.go:143,203`) as well as in `env.checkout`. Add one `trackerEnv.branchRef()` (ARCH-DRY).
- `refreshLocalMirror` and `refreshChangeCodeMirror` write with a fixed `0o644`, dropping the file's mode. `move-detail` keeps `info.Mode()`.
- `validHandoffDestination` checks the basename and blocks path escapes but doesn't pin the issues directory (BR-12 residue; low impact).
- The new malformed-details refusal in `cardsetters.go:54` has no regression test.

**Test coverage**
- The foreign-worktree test goes through the `reconcile` preflight. The adapter-level `requireSourceCheckout` in `materialize` and `removeSource` is only covered indirectly. A direct `TransferOp.Probe` test with a mismatched branch would pin it.
- Test runs in a non-git scratch copy (`git archive`):
  - `internal/tracker`, `internal/issue` and `internal/gitx`: pass.
  - Main package `cmd/sdlc`: completed with 12 failures, all consistent with the scratch copy not being a git checkout. The three I inspected fail on `git rev-parse --show-toplevel` or "could not resolve repository root"; the other nine have the same kind of names (branch creation, close, issue-sync Makefile). I didn't see any failure in the tracker-era tests.
  - Two more failures outside the main package also look environmental: `internal/judge` couldn't find `AGENTS.md` in the copy, and `internal/processgroup` wasn't allowed to run `ps`.

**Architecture**
- ARCH-DRY: minor flag (the branch-ref construction).
- ARCH-PURE: pass. The destination check and `NeedsSourceCheckout` are pure, and the IO stays in the adapters.
- ARCH-PURPOSE: flag. The BR-7 docs fix covered help text but not the atlas pages for the same changed verbs.
- ARCH-MOCK: pass. Tests use real Git remotes, hooks and worktrees.
- ARCH-CONSTRAINTS: pass. The guard adds one `cat-file` per unrecorded handoff.
- ARCH-SECURE: pass, with the destination-directory Minor.
- ARCH-ORDER: flag. Important #1 is an effect placed before its guard. The ownership check itself is correctly pushed down into the effect sites.
- ARCH-FUNERAL: pass. Receipts are discarded when finished, and a foreign one is labelled "finish from <branch>" rather than left unexplained.

**For M3:** the composed reader should own mirror refresh as a pure projection plus one writer that respects dry-run, so the refresh isn't repeated in `claim`, `change-code`, `move-detail` and the setters.

**Plan revision:** add a note that `atlas/workflow/issue-lifecycle.md`, `issue-sync.md` and `process-manual.md` were brought forward into M2 because `claim` and `change-code` semantics changed in M2, not M3.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      requireSourceCheckout in materialize/removeSource plus reconcile preflight and list labelling; TestRecoveryReconcileRefusesAnotherWorktree fails without it.
  - id: BR-5
    disposition: addressed
    note: |
      issue.CLIRef at every hint; issuehint_guard_test.go source guard plus base-0 round-trip.
  - id: BR-6
    disposition: addressed
    note: |
      records without main_commit are guarded once main holds the destination; TestTransferGuardProtectsAnUnrecordedPublication goes red without it.
  - id: BR-7
    disposition: not-addressed
    note: |
      issue.md:119 and atlas/workflow/issue-sync.md:51 still say change-code publishes; atlas/workflow/issue-lifecycle.md:34 and process-manual.md:111 still describe claim via origin/main. Rule: rg every changed verb name across cmd/sdlc/helptext and atlas, and classify each hit.
  - id: BR-8
    disposition: addressed
    note: |
      All 17 added non-test files have Core concepts rows; Task 3 Files revised; the rule is recorded in Revisions.
  - id: BR-9
    disposition: addressed
    note: |
      refreshMirror runs before publication; TestMoveDetailRefusesHandEditedCardFieldBeforePublishing. It introduced the write-before-check regression raised below.
  - id: BR-10
    disposition: addressed
    note: |
      diagnostics mixin plus TestUncertainStopReportsTheGitError; the cardsetters malformed refusal has no test.
  - id: BR-11
    disposition: addressed
  - id: BR-12
    disposition: addressed
    note: |
      basename, clean-path and no-escape checks plus a repo-wide message; the issues directory itself is not pinned (low impact).
findings:
  - id: new
    severity: Important
    family: refusal-after-local-effect
    title: |
      move-detail and change-code write the refreshed mirror before dry-run and precondition checks
    detail: |
      Reproduced: after a set-title from another worktree, move-detail --dry-run rewrites the committed details file, then both dry and real runs refuse with "nothing was changed" (issuemovedetail.go:166 runs before checkMoveSource and DryRun). change-code writes at changecode.go:307 before its DryRun check at :196. Rule: checks run before effects; a refusing or dry-run path leaves files unchanged. Fix: refresh in memory, check the original, then write and git add.
  - id: new
    severity: Minor
    family: silent-error-swallowing
    title: |
      transfer guard treats any git cat-file failure as the destination being absent from main
    detail: |
      This is the 2nd finding in family silent-error-swallowing. Rule: an observation error is never evidence of absence; distinguish "path missing" from a Git failure at every probe (transferguard.go:62, and sweep the other cat-file -e probes in trackerenv/issuemovedetail).
  - id: new
    severity: Minor
    family: shared-helper-extraction
    title: |
      "refs/heads/"+env.branch is built by hand at four sites alongside env.checkout
    detail: |
      issuerecovery.go:62,102,137 and issuemovedetail.go:143,203; add trackerEnv.branchRef() (ARCH-DRY).
```

---

## Re-review — 2026-09-25T18:46:09-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 547122d280ebbead8d4127e11d01eda5ff863be7..d5877eabc7da5b32fe4130f66512d4bfca48fc14 |
| command | sdlc milestone-close --issue 252 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-25T18:46:09-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The round-4 fixes mostly hold up. `gitTest`/`has` separate a false predicate (exit 1) from a Git failure, and `TestGitTestSeparatesFalseFromFailure` pins that. `branchRef()` replaced the hand-built refs, and the help sweep for claim, issue, start-plan and change-code now describes local-only design and publication through `move-detail`. move-detail is now correctly ordered: checks run first, the refresh happens in memory, and it writes only when it proceeds. A regression test covers the dry-run, split-refusal and proceeding paths. One problem is left. BR-13 was fixed for move-detail, but change-code only for its dry-run path. A real change-code run still writes the refreshed mirror at step 1b, before gates that can `exitWithCode(1)`. That contradicts the rule BR-13 stated, and the function's own comment that "a refused run leaves the issue untouched". The fix is cheap. A few swallowed `env.git` probes also remain in move-detail (Minor). Targeted suites pass: `internal/{gitx,issue,tracker}` and the `cmd/sdlc` tests for MoveDetail, ChangeCode, GitTest, TransferGuard, Recovery, Planning, Claim, IssueNew, SetStatus and Card. The one failure is `internal/processgroup`, where the sandbox refuses `fork/exec /bin/ps: operation not permitted`. That is environmental and outside this diff.

**1. Strengths**
- `cmd/sdlc/trackerenv.go:85-115`: `gitTest` puts the exit-1-versus-failure distinction in one place and routes both `has` and `branchRef` through it. It fixes the class, not just the one site (ARCH-DRY, BR-14/BR-15).
- `cmd/sdlc/issuemovedetail.go:160-206`: checks come before effects. `checkMoveSource` → in-memory refresh → blob → dry-run exit → first write. The split predicate (index ≠ HEAD **and** worktree ≠ index) is now precise.
- `TestMoveDetailStaleMirrorIsWrittenOnlyWhenProceeding` covers dry-run, refusal and proceeding in one sequence. It also asserts that main receives the refreshed title.
- The transfer guard (`transferguard.go:98-117`) now propagates probe errors instead of treating them as absence.

**2. Critical:** none.

**3. Important**
- BR-13 (still open): `cmd/sdlc/changecode.go:130` → `refreshChangeCodeMirror` writes at `:317` whenever `!f.DryRun`. Any later refusal leaves the rewritten details on disk: the gate loop's `exitWithCode(1)` at `:185` or `resolveBranchingStrategy`'s `die`. Fix: always return the refreshed bytes without writing. Write them right before `recordChangeCodeFlow` at step 7, which is already the first effect. Add a test where a gate refuses on a stale mirror and the file bytes stay unchanged.

**4. Minor**
- `silent-error-swallowing` (3rd in family). The rule: no `env.git` result may drop its error; probes go through `gitTest`/`has` or propagate. Remaining instances: `issuemovedetail.go:202` and `:322` (`ls-files` error dropped, so `git rm --cached` is skipped), `:301` (`rev-parse HEAD`), and `transferguard.go:106-107` (these fail safe into a refusal). A grep guard test like `issuehint_guard_test.go` would enforce the rule.
- The `gitTest` doc claims "a bad name … is an error". For `rev-parse -q --verify`, a bad commit also exits 1, so `has(badCommit, p)` reads as absent. The callers pass resolved OIDs today, so the fix is to correct the comment.
- `issuemovedetail.go:202-203` runs `git add` on any tracked source. That also stages the user's unrelated unstaged edit, which matters if the transfer later fails before removal.

**5. Test coverage**
- There are new regression tests for BR-13 (move-detail, and change-code dry-run), BR-14 (`has` on a non-repository) and the split predicate.
- Missing: a change-code refusal with a stale mirror (see Important above).

**6. Architecture, principle by principle**
- **ARCH-DRY:** pass. `branchRef` and `gitTest` are consolidated.
- **ARCH-PURE:** pass. The card and handoff logic in `internal/issue` stays pure; the git glue lives in `trackerEnv`.
- **ARCH-PURPOSE:** pass for M2 scope. The AGENTS.md instruction rewrite is M4 in the plan.
- **ARCH-MOCK:** pass. Tests run against real temporary Git repos and origins.
- **ARCH-CONSTRAINTS:** pass.
- **ARCH-SECURE:** pass. The handoff destination is validated and `Lstat` checks for a regular file.
- **ARCH-ORDER:** flag, covered by BR-13. The effect ordering in change-code is still implicit.
- **ARCH-FUNERAL:** pass. Receipts are discarded or resumed, and recovery refs are swept by reconcile.

**7. Plan revisions:** none needed. The plan's M2 status and Core concepts table match the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      claim.md/start-plan.md/change-code.md/issue.md now describe tracker reservation, local-only design, move-detail/recovery/set-* verbs; remaining AGENTS.md issue-sync prose is M4 (instructions) scope per plan Task 7/8.
  - id: BR-13
    disposition: not-addressed
    note: |
      move-detail fixed and tested; change-code only fixed for dry-run — changecode.go:317 still writes the refreshed mirror before gates that exitWithCode(1) at :185, so a refused run leaves the details rewritten, contradicting its own step-2b comment.
  - id: BR-14
    disposition: addressed
    note: |
      transferguard.go:61 and the cat-file/merge-base/diff probes now use trackerEnv.gitTest/has (trackerenv.go:85-104), pinned by TestGitTestSeparatesFalseFromFailure; residual ls-files/rev-parse drops raised separately.
  - id: BR-15
    disposition: addressed
    note: |
      trackerEnv.branchRef() (trackerenv.go:106) replaces all env.branch sites in issuerecovery.go and issuemovedetail.go; remaining refs/heads/+name in planningbranch.go names a different branch.
findings:
  - id: new
    severity: Minor
    family: silent-error-swallowing
    title: |
      move-detail and the transfer guard still discard env.git errors on ls-files/rev-parse probes
    detail: |
      3rd in family. Rule: no env.git result may drop its error; probes go through gitTest/has or propagate. Instances: issuemovedetail.go:202, :301, :322 (ls-files failure skips git rm --cached), transferguard.go:106-107 (fails safe). Enforce with a grep guard test rather than per-site fixes.
  - id: new
    severity: Minor
    family: silent-error-swallowing
    title: |
      gitTest doc claims a bad name is an error, but rev-parse -q --verify exits 1 for a bad commit too
    detail: |
      has(badCommit, p) returns (false, nil). Callers pass resolved OIDs today; correct the comment or verify the commit separately.
```

---

## Re-review — 2026-09-25T19:04:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 547122d280ebbead8d4127e11d01eda5ff863be7..b23b98178075b448082d8c28724a1de975f48a7d |
| command | sdlc milestone-close --issue 252 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-25T19:04:14-07:00 |
| verdict | SHIP |

## Review

The prior findings are done: all three are addressed, with regression evidence where the behaviour changed. One new Minor finding.

```verdict
verdict: SHIP
confidence: high
```

Round 5 closes the three open findings, and I checked each against evidence. For change-code, the refreshed mirror is now held in memory. The gates read it through `gateContent`, the review snapshot covers the untouched disk bytes, and the file is written only after every gate passes and `f.review.validate()` succeeds (`cmd/sdlc/changecode.go:222`). I reverted `changecode.go` in a scratch worktree and `TestChangeCodeGateRefusalLeavesStaleMirrorOnDisk` failed ("refused change-code rewrote the details"). With the fix in place, the targeted suite passes, including MoveDetail, TransferGuard and the new guard test. The move-detail half of BR-13 was already correct, and `TestMoveDetailStaleMirrorIsWrittenOnlyWhenProceeding` covers both the dry run and the staged/unstaged split refusal. For the silent-error family, round 5 added a whole-package source guard test (`gitresult_guard_test.go`) instead of another per-site patch, which is the right response to a third repeat. The only new finding is a Minor ordering gap in move-detail; nothing blocks SHIP.

1. **Strengths**
   - `changecode.go:130-225`: refreshes the mirror in memory, runs the gates on the refreshed fields, and writes last. This is the "checks before effects" rule applied properly, and a built-binary test pins it.
   - `gitresult_guard_test.go`: turns the `silent-error-swallowing` rule into a package-wide check (ARCH-PURPOSE: the fix covers the whole class, not one instance).
   - `transferguard.go:98-117`: `want` and `got` are resolved only when `has` confirms the path exists, and failures propagate. A missing path can no longer be mistaken for a changed one.
   - `issuemovedetail.go:302-309`: the pointless `rev-parse HEAD` probe is gone, which simplifies the rest-branch fast-forward check.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `issuemovedetail.go:195-215`: the stale-mirror write (the first effect) runs before `tracker.NewTransfer(spec)`, which validates the receipt spec and can still refuse. This is the 2nd finding in family `refusal-after-local-effect`. The rule that covers both: every fallible validation, including building the receipt or operation, completes before the first file or index write. Fix: move `NewTransfer` above the `staleMirror` block.
   - The guard test's regex catches `_ =` and `, _ :=` discards of `env.git`/`e.git`. It does not catch a bare `env.git(...)` statement; none exist today.

5. **Test coverage:** the BR-13 regression test goes red when the fix is reverted (verified). BR-16 is enforced by the source guard, which would have flagged all three former sites and the `transferguard.go` pair. BR-17 was a comment-only fix.

6. **Architecture**
   - ARCH-DRY pass.
   - ARCH-PURE pass: `refreshChangeCodeMirror` no longer writes, which narrows the side-effecting part of the code.
   - ARCH-PURPOSE pass: the silent-error family is fixed as a rule, not per instance.
   - ARCH-MOCK pass: tests run against local Git fixtures and fake origins.
   - ARCH-CONSTRAINTS pass.
   - ARCH-SECURE pass.
   - ARCH-ORDER flag (Minor above): the only effect that still precedes a validation is in move-detail.
   - ARCH-FUNERAL pass: `WriteBlob` before a dry run leaves an unreferenced loose object, which `git gc` removes.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-13
    disposition: addressed
    note: |
      change-code writes the mirror only after the gates and review validate (changecode.go:222); reverting the fix turns TestChangeCodeGateRefusalLeavesStaleMirrorOnDisk red (verified); the move-detail dry-run and refusal paths are pinned by TestMoveDetailStaleMirrorIsWrittenOnlyWhenProceeding.
  - id: BR-16
    disposition: addressed
    note: |
      All named sites now propagate errors; the whole package is enforced by TestTrackerGitResultsKeepTheirErrors in gitresult_guard_test.go, a rule rather than per-site fixes.
  - id: BR-17
    disposition: addressed
    note: |
      The gitTest doc in trackerenv.go now says rev-parse -q --verify exits 1 for an unknown commit and that callers pass resolved commits; the has() callers pass mainTip, the merge-tree result, HEAD, or a pinned base.
findings:
  - id: new
    severity: Minor
    family: refusal-after-local-effect
    title: |
      move-detail rewrites the stale mirror before tracker.NewTransfer validates the receipt spec
    detail: |
      2nd finding in this family. Rule: every fallible validation, including receipt/op construction, completes before the first file or index write. In issuemovedetail.go the staleMirror WriteFile and git add run before NewTransfer(spec), whose newReceipt validation can refuse; move NewTransfer above the write.
```
