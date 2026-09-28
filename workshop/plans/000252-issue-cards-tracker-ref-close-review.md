# Boundary Review — ariadne#252 (whole-issue close)

| field | value |
|-------|-------|
| issue | 252 — Issue cards: card fields on a tracker ref, details on the branch |
| repo | ariadne |
| issue file | workshop/issues/000252-issue-cards-tracker-ref.md |
| boundary | whole-issue close |
| milestone | — |
| window | e8fcf6116f82106b029519dc02009272a13adfcd..4a0fc9a84bd679bae20918eb8e92d73902744d84 |
| command | sdlc close --issue 252 |
| reviewer | claude |
| timestamp | 2026-09-27T18:51:52-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

**Home directory incident.** While building a reproduction, `mktemp` was blocked, `$T` came back empty, and my `cd $T` landed in your home directory. `~` is already a git repository: a scratch repo from July 5 with a single `init` commit by `t <t@t>`. Before the next command failed, my commands made these changes there:
- **Added and now removed:** the `remote.origin` config (URL `/nonexistent/origin.git`) and the `branch.main.remote`/`merge` config. I deleted both with `git remote remove origin`.
- **Created and now deleted:** `~/workshop/` and `~/workshop/issues/000001-x.md`.
- **Left in place:** `user.email=t@t`, `user.name=t` and `commit.gpgsign=false`. My commands wrote these, but the existing commit is authored `t <t@t>`, so they were probably already set. I didn't remove them because I can't prove they were absent. Run `git -C ~ config --local --unset …` if they shouldn't be there.
- **Unchanged:** nothing was staged or committed, and `~/.git/COMMIT_EDITMSG` now reads "seed".

I redid the reproduction in a guarded scratch directory and have deleted it.

**Summary.** Four of the five open findings are fixed:
- **BR-3:** `git diff --check` on the M1 review is clean.
- **BR-18:** `tracker.NewTransfer` now runs before the stale-mirror write (`issuemovedetail.go:200` vs `:204`).
- **BR-33:** a skipped pinned file now gets a `Close-Kept` trailer and a warning, and a test checks for it.
- **BR-39:** `TestShellFallbacksSpellTheTrackerVocabulary` reads the vocab and checks both shell files.

BR-40 is partly fixed. The migration now refuses a landed close whose branch still carries code main lacks, but the `MigrationAnchor` doc comment still says the check isn't applied to landed closes.

The main new problem is in the post-M4 legacy-mode work. The mode check (`repositoryTracked`) always runs `git ls-remote`. So a legacy repository can no longer run purely local verbs (`issue set-status`, `issue new`, `change-code`) when its remote is unreachable. That breaks the promise that legacy repositories "keep working exactly as before". I reproduced it: `sdlc issue set-status 1 --issue 1` in a legacy repo with an unreachable origin fails with `could not fetch from origin`. Before #252 that verb only edited a local file. The targeted tests and `go vet` pass.

### 1. Strengths
- `legacymode.go` restores the pre-252 verbs in one file with a clear end date ("Deleted once every fleet repository has cut over"), and every dispatch site uses the same helper.
- `gitx.BranchPoint`/`TrunkRef` (`window.go:92-130`) replace four hand-written `merge-base main HEAD` calls. Two real-git tests cover a stale local main and a local main that is ahead of the trunk.
- `ownerAuthored` (`transferguard.go:88`) decides the owner exemption from commit ancestry, not from the branch name. It checks both local and remote refs and fails closed.
- Reconcile no longer swallows a failed `merge --abort` or `reset` (4f7260ae); the error names the checkout state and the manual step.
- `TestShellFallbacksSpellTheTrackerVocabulary` enforces the "shell fallbacks copy the vocab" rule for every shell copy, not just one file.

### 2. Critical findings
None.

### 3. Important findings
- **`legacymode.go:32` (`repositoryTracked`), via `Repository.Initialized` → `TrunkFile.RemoteExists` (`ls-remote`).**
  - **Problem:** the mode check needs the network, so an unreachable remote aborts every dispatching verb before the legacy path runs. That drifts from the legacy-mode contract:
    - `set-status` was purely local.
    - `issue new` and `change-code` treated publication as best-effort, with a local-commit fallback.
    - `start-plan` hides the error, but `claim`, `set-status`, `issue new` and `change-code` refuse.
  - **Cost when online:** tracked repositories now pay one extra `ls-remote` per verb before `openTracker`.
  - **Fix:** decide from local evidence first, with no network:
    - The cutover marker in the main tree, or a fetched `refs/remotes/*/issue-tracker`, means tracked.
    - Otherwise run `ls-remote`. If that fails as offline and there's no local evidence, treat the repo as legacy, with a warning, for local-only verbs.
  - **Test:** a legacy repo whose origin is unreachable still runs `issue set-status`.
  - **Family:** new, `local-verb-network-dependency`.

### 4. Minor findings
- `internal/tracker/migration.go:36-40`: the `MigrationAnchor` doc says "CodeAfter is not asked of" a landed close, but `bindLegacyClose` now refuses on it and `issuemigrate.go` computes it as `main...ref`. Update the comment so it doesn't contradict the BR-40 fix.
- `pr.go:83` / `gitCommitsSince`: when local main is ahead of the trunk, the PR body lists `BranchPoint..HEAD`. The GitHub PR diffs against the remote trunk, so unpushed local-main commits appear in the PR but not in its body. This is an edge case of the legacy fork.

### 5. Test coverage notes
- No test covers the offline/unreachable-remote case for legacy mode. `TestLegacyGatesNeedNoUpstream` only removes the upstream, and a missing upstream is not the same as a remote that can't be reached.
- The `BranchPoint` tests don't cover a branch fully merged into an unpushed local main, where `BranchPoint` returns `HEAD`. It behaves as before #252 (`MergeBaseWithMain` returns `""`), but nothing locks that in.
- Targeted runs pass: `gitx`, `tracker`, and `cmd/sdlc -run 'Legacy|TransferGuard|ShellFallbacks|Migrat|Reconcile|Leftover|StartPlan'`.

### 6. Architectural notes
- **ARCH-DRY: pass.** `BranchPoint` consolidates the callers. `ownerAuthored` builds `refs/heads/<owner>` by hand, but for a different branch than `env.branchRef()`, so that's acceptable.
- **ARCH-PURE: pass.** `bindLegacyClose` stays pure and the git IO stays in `issuemigrate.go`. `syncPointer` is pure with its mode injected.
- **ARCH-PURPOSE: flag.** Legacy mode's purpose is exact pre-#252 behaviour, and needing the network breaks that (Important finding above).
- **ARCH-MOCK: pass.** The new tests use real git plus a bare origin, as the seam already does.
- **ARCH-CONSTRAINTS: flag.** Every dispatching verb gets an unbudgeted `ls-remote`, and there's no fallback when the remote can't be reached (same finding).
- **ARCH-SECURE: pass.** The owner exemption fails closed when a ref is missing, and a malformed card still refuses repo-wide.
- **ARCH-ORDER: pass.** Reconcile checks run before effects, and both undo paths now report failures.
- **ARCH-FUNERAL: pass.** `legacymode.go` names when it is deleted (fleet cutover, handed to #255). No new persisted families.

### 7. Plan revision recommendations
- Add a Revisions entry noting that legacy-mode detection must degrade offline to local evidence, and name the regression test.
- Once fixed, record in the #255 cutover checklist that `legacymode.go` is deleted when the last repository cuts over. It's already implied; make it an explicit checklist row.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      git diff --check over the window reports nothing for the M1 review artifact.
  - id: BR-18
    disposition: addressed
    note: |
      issuemovedetail.go:200 constructs tracker.NewTransfer before the staleMirror WriteFile and git add at :204-216.
  - id: BR-33
    disposition: addressed
    note: |
      closetracker.go:65-68 names every superseded pinned file in a Close-Kept trailer plus a warning; closetracker_test.go:181 asserts the trailer.
  - id: BR-39
    disposition: addressed
    note: |
      TestShellFallbacksSpellTheTrackerVocabulary derives the fetched-ref glob from vocab and the marker from tracker.CutoverMarkerPath, and checks Makefile.workflow plus scripts/close-issue.py.
  - id: BR-40
    disposition: addressed
    note: |
      bindLegacyClose refuses a landed anchor whose CodeAfter (main...ref) is set; migration_test.go:197 covers it. The MigrationAnchor doc comment is stale (raised Minor below).
findings:
  - id: new
    severity: Important
    family: local-verb-network-dependency
    title: |
      Legacy-mode detection requires ls-remote, so offline legacy repos can no longer run local-only verbs
    detail: |
      repositoryTracked calls Repository.Initialized, whose TrunkFile.RemoteExists runs ls-remote; an unreachable remote returns an error, so claim, set-status, issue new and change-code refuse before the legacy path. Reproduced with issue set-status against an unreachable origin. Decide from local evidence first (cutover marker, fetched tracker ref), and treat an offline remote with no evidence as legacy for local verbs; add an unreachable-origin test.
  - id: new
    severity: Minor
    family: landing-completion-proof
    title: |
      MigrationAnchor doc still says CodeAfter is not asked of a landed close
    detail: |
      This is the 4th finding in landing-completion-proof, and it is doc drift left behind by the BR-40 fix. The rule already applies in code; only migration.go:36-40 contradicts it. Update the comment to say a landed close checks the branch's code beyond main.
```
