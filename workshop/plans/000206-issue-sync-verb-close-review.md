# Boundary Review — ariadne#206 (whole-issue close)

| field | value |
|-------|-------|
| issue | 206 — sdlc: commit planning output via issue sync |
| repo | ariadne |
| issue file | workshop/issues/000206-issue-sync-verb.md |
| boundary | whole-issue close |
| milestone | — |
| window | 92bd1ad1ebdc6bca1b218b318e4740b6fab3d787..f1deae381fb15029090eb5299b9db66c57a00605 |
| command | sdlc close --issue 206 |
| reviewer | claude |
| timestamp | 2026-09-02T12:23:35-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The core of #206 is well built and genuinely proven: the narrowed-add/bare-commit fix is pinned by real-git regressions that I confirmed go red when reverted (`syncInPlace`, `push`'s archive, and the `bookkeepingVerbs` entry all fail without their fix), the `NoPush`-first dispatch is the right discriminator and is documented honestly, and the atlas/helptext land with the change rather than after it. What blocks SHIP is one silent functional regression: `syncIssue` gates on `f.Issue > 0`, but `commitUntrackedIssueFile` — which it replaced — also fired in `change-code`'s **auto-detect** mode (neither `--issue` nor `--name`, `resolveBranchName` case 3). I ran the same probe against base `92bd1ad` and head `f1deae3`: at base the untracked issue file is committed; at head it is left `?? workshop/issues/…`, and with `--worktree=yes` the newly created worktree contains **no issue file at all**. That is exactly the "branch starts from a tracked state" property the Spec claims is preserved. Secondarily, the class sweep the Revisions section calls complete misses a sixth member in `migrate.go`.

### 1. Strengths

- **The revert test actually holds.** I reverted each claimed fix in a scratch copy: `claim.go:211` → `TestSyncInPlace_LeavesForeignStagedFileAlone` fails with `peer-work.go` swept into the commit; `push.go:264` → `TestArchiveCommit_LeavesForeignStagedFileAlone` fails the same way; `gitx/window.go:144` → both `TestIsShippedWorkSubject` rows and the sdlc-side coupling test fail. No fix here is a no-op passing its own mirror.
- **`syncPathspec` (`claim.go:236`) is the right shape** — one source feeding both the `add` and the `commit` argv, so the two lists structurally cannot drift. `archiveCommitArgs` (`push.go:263`) does the same by deriving from `archiveAddArgs(moves)[2:]` rather than restating it (ARCH-DRY).
- **The NoPush-before-branch-test dispatch** (`claim.go:117`) is the non-obvious correct call, and `TestIssueSync_FromFeatureBranchCommitsOnThatBranch` pins the property that motivates it.
- **ARCH-MOCK is handled with unusual care.** The fixture preamble in `issuesync_test.go:15-22` states why an argv fake can't answer the question, and `TestSyncViaMainWorktree_CommitCarriesPathspec` uses the stub *only* where argv genuinely is the question, with the split justified in the Revisions section.
- **The `#N`-anchoring hazard was found and closed at the source** — `issueSyncMessage`'s doc comment (`issue.go:404`) names the three downstream consumers that break if the prefix changes, and a test enforces the coupling from the sdlc side.

### 2. Critical findings

**C1 — `cmd/sdlc/changecode.go:196` (`syncIssue`): `change-code` with neither `--issue` nor `--name` no longer commits the issue file.**

`resolveBranchName` (`branchcreate.go:78-85`) supports a third mode: scan for exactly one untracked `NNNNNN-*.md` and use it. In that mode `f.Issue == 0`, so `syncIssue` returns at its first line and nothing commits. Verified by running `runChangeCode` in that mode against base and head:

```
base 92bd1ad:  ==> Committing workshop/issues/000300-brand-new.md before branch creation...
               status: (clean)
head f1deae3:  status: ?? workshop/issues/000300-brand-new.md
```

With `--worktree=yes` the consequence is worse — `git worktree add` does not carry untracked files, so the new worktree has no issue file:

```
REGRESSION: issue file MISSING in the new worktree .../worktree/002/000300-brand-new:
  stat .../workshop/issues/000300-brand-new.md: no such file or directory
```

Every later verb in that worktree that resolves the issue by file then fails. The in-place path degrades less loudly (the file rides along untracked) but still defers the failure to `sdlc push`'s untracked-file refusal.

*Fix sketch:* don't gate on the flag, gate on the resolved path. `resolveChangeCodeName` already returns `issuePath`; derive the id with the existing `issueIDPrefix(filepath.Base(issuePath))` + `strconv.Atoi` and pass that to `syncIssue`, so all three resolution modes sync. Pin it with a test in auto-detect mode asserting the issue file is committed (and, for `--worktree=yes`, present in the created worktree). `--name` mode is genuinely id-less and can keep the early return — but then the guard's comment needs to say so accurately (see M1).

### 3. Important findings

**I1 — `cmd/sdlc/migrate.go:343` and `:346`: an unswept member of the narrowed-add/bare-commit class.**

`migrate` stages precisely (`add -- destRel` at `:327`, `add -- relPath` at `:333`) and then commits bare in both repos. Source side has **no** whole-repo cleanliness guard at all — step (1) at `:241` checks `status --porcelain -- relPath`, i.e. only the migrated file — so a peer's staged work in the source repo is swept into `migrate: move X to Y`. Dest side has a guard at `:288-293`, but it is bypassable, and its bypass message reads *"proceeding into a dirty destination (staging is explicit-path)"* — promising exactly the property the bare commit removes. The success line at `:349` likewise says "both sides committed, scoped". This is the same rule as the five swept sites, and the `--no-commit` hint at `:340-341` hands the operator the defective form too. Fix: reuse the pathspec shape (`commit -m … -- <path>`) at both sites and in the printed hints; add a real-repo regression on the source side, where nothing guards it.

**I2 — `cmd/sdlc/push.go:263`: `archiveCommitArgs` silently reverts to a whole-index commit on empty `moves`.**

`archiveCommitArgs(msg, nil)` returns `["commit","-m",msg,"--"]`. I confirmed against real git that `git commit -m x --` with no paths commits the entire index. All three call sites currently guard `len(moves) > 0`, so nothing is broken today — but the helper whose whole reason for existing is "never commit the whole index" does exactly that on its degenerate input, and no test pins it. Its sibling `syncPathspec` returns an error rather than a silent widening; make this one consistent (return `nil` and have callers treat it as "nothing to commit", or return an error) and add a table row.

**I3 — the mid-planning trigger has no delivery point in the agent's attention path.**

The verb's entire justification is durability *during* planning, which is why `--push` defaults off — but the only place that says "run this whenever the Spec/Plan/Log has moved" is `cmd/sdlc/helptext/issue.md:67-72`, reachable only by typing `sdlc issue sync --help`. `sdlc start-plan` — the verb an agent runs *on entering planning*, whose output already points at where to author the plan and what to defer — says nothing (`helptext/start-plan.md`), and neither does AGENTS.md §2 (the `claim → start-plan → design → change-code` flow) or §14 (checkpoint before compaction, which is the canonical moment to sync). `change-code` covers the end-of-planning trigger automatically; the mid-planning trigger, which is the half the issue exists for, ships as documentation nobody is routed to (ARCH-PURPOSE). One line in `start-plan`'s output and one in AGENTS.md §2 closes it.

### 4. Minor findings

- **M1** — `changecode.go:197`: the comment `// --name mode: no issue ID to name in the commit message` describes one of the two modes the guard catches; the auto-detect mode is the one that regressed (C1).
- **M2** — `syncIssue` doesn't thread `f.DryRun` into `claimFlags`; correct only because the dry-run branch returns at `changecode.go:159`. If the sync ever moves earlier in the sequence this silently starts committing under `--dry-run`.
- **M3** — `merge.go:550` has no direct regression test; it rides on `archiveCommitArgs` plus the two-worktree fixture. The Log states this rationale explicitly, so it's a recorded choice, not an oversight — noting it because `merge` is the site the Revisions section called "the live one".
- **M4** — no coverage for `sdlc issue sync --dry-run` on either arm.
- **M5** — a pathspec'd commit is a *partial* commit, which git refuses outright while `MERGE_HEAD` is set (`fatal: cannot do a partial commit during a merge`, confirmed). `sdlc issue sync` run mid-merge now dies where the bare commit would have created a merge commit. Almost certainly the better behavior, but it is a new failure mode and isn't mentioned in the atlas entry.
- **M6** — AGENTS.md §1 asks for a durable plan in `workshop/plans/NNNNNN-slug-plan.md` for work over 3 files / 100 lines; this is 9 source files and ~900 lines, and `workshop/plans/` holds only the gate ledger. The issue's `## Plan` is thorough, so this is a filing convention gap, not a design gap.
- **M7** — "the files for issue N" is globbed in `syncPathspec` and `resolveBranchName`, and prefix-matched a third way in `changedIssueFiles`. Same count as before the change (no new duplication), but a shared `issueFilesForID` would be the natural next consolidation.

### 5. Test coverage notes

Coverage is strong where it matters and honest about what it can't reach. All 14 new tests pass; the only failure in `./cmd/sdlc` is `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, pre-existing — its plan file was archived in `dfeba9c`, which is an ancestor of the review base. Confirmed by revert that C-level claims are pinned rather than mirrored. No `t.Parallel()` anywhere in package `main`, so the process-global `chdir` in `syncRepo` is safe. Gaps: the C1 mode has no test at all (which is why the regression shipped — a `change-code` auto-detect test would have caught it), plus I2, M3 and M4 above.

### 6. Architectural notes

- **ARCH-DRY — pass.** One sync dispatch with four callers; `archiveCommitArgs` derived from `archiveAddArgs`; `syncPathspec` feeding both argv lists. `commitUntrackedIssueFile` deleted rather than left beside the new path.
- **ARCH-PURE — pass.** `syncPathspec`, `syncMessage`, `archiveCommitArgs`, `issueSyncMessage` and the `bookkeepingVerbs` classification are pure and unit-tested without IO; the git calls stay behind the `gitRunner` seam. `syncViaMainWorktree` remains a long IO-and-logic function, but that's inherited, not added.
- **ARCH-PURPOSE — flag (I1, I3, and C1).** The class enumeration stopped one site short of the class (`migrate.go`), the mid-planning trigger ships without a delivery point, and the replacement helper covers a subset of the modes the replaced one did. Each is the "instance, not the class" pattern the principle names.
- **ARCH-MOCK — pass, and a model for the repo.** Real `testfix.Repo` for semantics, the argv stub confined to the one argv question, with the split reasoned out in the Revisions section rather than assumed.
- **ARCH-CONSTRAINTS — pass with a note.** The envelope is stated in prose and enforced by design (default arm: no network, ~4 git subprocesses, one file), which is what makes "run it often" safe. No budgeted envelope block in the issue, though — worth adding now that #205 has made that the house style, if only to record "local git only, no network on the default path" as a constraint rather than a description.

### 7. Plan revision recommendations

Two `## Revisions` entries; the plan currently claims things the code doesn't deliver.

1. **`2026-09-02 — correction: the class has a sixth member`** — the enumeration table asserts every member of "narrow the `git add`, then commit the whole index" is covered by the five listed rows. `migrate.go:343,346` is a sixth, unguarded on the source side (step (1) checks only the migrated path), with a `--no-clean-check` message that promises scoped staging the commit doesn't provide. Either sweep it in this round or record the verified reason it's excluded — the same standard PQ-5 established.
2. **`2026-09-02 — correction: the tracked-start property is not preserved in every case that has it`** — the Spec asserts the property "is preserved in the case that has it". `commitUntrackedIssueFile` also fired in `change-code`'s auto-detect mode, and `syncIssue`'s `f.Issue <= 0` early return drops it there; `--worktree=yes` in that mode leaves the new worktree with no issue file. Add the corresponding `## Plan` row and a `## Done when` line ("`change-code` commits the issue file in all three name-resolution modes").

```findings
findings:
  - id: new
    severity: Critical
    family: helper-swap-drops-a-mode
    title: |
      change-code's auto-detect mode (no --issue, no --name) no longer commits the untracked issue file
    detail: |
      syncIssue (changecode.go:196) returns early on f.Issue <= 0, but the helper it
      replaced also fired in resolveBranchName's case-3 auto-detect mode, where Issue is 0.
      Verified against base 92bd1ad (file committed, tree clean) vs head f1deae3 (left
      "?? workshop/issues/..."); with --worktree=yes the created worktree contains no issue
      file at all. Derive the id from the already-resolved issuePath via issueIDPrefix
      instead of gating on the flag, and pin it with an auto-detect-mode test.
  - id: new
    severity: Important
    family: narrowed-add-bare-commit
    title: |
      migrate.go is an unswept member of the narrowed-add/bare-commit class
    detail: |
      migrate stages precisely (migrate.go:327, :333) then commits bare in both repos
      (:343, :346). The source side has no whole-repo cleanliness guard — step (1) at :241
      checks only the migrated path — so a peer's staged work is swept into
      "migrate: move X to Y". The dest-side guard is bypassable and its --no-clean-check
      message claims "staging is explicit-path", the very property the bare commit removes.
      The --no-commit hints at :340-341 print the defective form too.
  - id: new
    severity: Important
    family: narrowed-add-bare-commit
    title: |
      archiveCommitArgs silently emits a whole-index commit when moves is empty
    detail: |
      archiveCommitArgs(msg, nil) returns ["commit","-m",msg,"--"], and git treats a bare
      "--" with no paths as no pathspec at all — confirmed against real git. All three call
      sites guard len(moves) > 0, so nothing breaks today, but the helper whose purpose is
      "never commit the whole index" does exactly that on empty input, untested. Its sibling
      syncPathspec errors instead of widening; make this consistent and add a table row.
  - id: new
    severity: Important
    family: verb-without-a-trigger
    title: |
      the mid-planning trigger for `sdlc issue sync` has no delivery point in the agent's path
    detail: |
      The only place telling an agent to run the verb mid-planning is helptext/issue.md:67-72,
      reachable only via `sdlc issue sync --help`. start-plan — the verb that opens planning and
      already points at where to author the plan — says nothing, and neither does AGENTS.md §2 or
      §14. change-code covers the end-of-planning trigger automatically; the mid-planning half,
      which is why --push defaults off, ships as documentation nobody is routed to (ARCH-PURPOSE).
  - id: new
    severity: Minor
    family: comment-narrower-than-code
    title: |
      syncIssue's guard comment names only --name mode, not the auto-detect mode it also catches
  - id: new
    severity: Minor
    family: latent-flag-not-threaded
    title: |
      syncIssue does not thread f.DryRun into claimFlags; safe only because the dry-run branch returns earlier
  - id: new
    severity: Minor
    family: class-fix-without-class-test
    title: |
      merge.go:550's archive commit has no direct regression test (recorded choice, not an oversight)
  - id: new
    severity: Minor
    family: class-fix-without-class-test
    title: |
      no coverage for `sdlc issue sync --dry-run` on either arm
  - id: new
    severity: Minor
    family: undocumented-behavior-change
    title: |
      a pathspec'd commit is a partial commit, which git refuses while MERGE_HEAD is set
    detail: |
      Confirmed: "fatal: cannot do a partial commit during a merge". `sdlc issue sync` run
      mid-merge now dies where the bare commit would have created a merge commit. Almost
      certainly the better behavior, but it is a new failure mode and the atlas entry
      doesn't mention it.
  - id: new
    severity: Minor
    family: durable-plan-artifact
    title: |
      no durable plan in workshop/plans/ for a 9-file, ~900-line change (AGENTS.md §1)
  - id: new
    severity: Minor
    family: shared-issue-file-lookup
    title: |
      "the files for issue N" is computed three ways (syncPathspec, resolveBranchName, changedIssueFiles)
  - id: new
    severity: Minor
    family: operating-envelope-undeclared
    title: |
      no declared operating envelope for a verb designed to be run frequently (ARCH-CONSTRAINTS)
```
