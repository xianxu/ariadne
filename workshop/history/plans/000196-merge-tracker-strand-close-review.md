# Boundary Review — ariadne#196 (whole-issue close)

| field | value |
|-------|-------|
| issue | 196 — merge: dirty tracker files are exempted from the block but still strand the post-merge switch |
| repo | ariadne |
| issue file | workshop/issues/000196-merge-tracker-strand.md |
| boundary | whole-issue close |
| milestone | — |
| window | 505a771abe91e2393c9033ceefc5809c9041ca59..e1703785b2e8cd2fc72afeb50f44a3478b46280f |
| command | sdlc close --issue 196 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T08:47:52-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Diff read; now running the merge tests and checking the recovery-message path.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The fix works. `mergeRefusal` (`cmd/sdlc/merge.go:213`) is now the one merge-block decision for both step 2 and the 9b re-check. It asks git's own question: a dirty tracker file is refused only if it differs between `main` and `HEAD`, or between `main` and `origin/main`. Step 9b fetches `origin/main` first, and a failed diff refuses rather than letting the merge through. Remaining tracker dirt still only warns, so #82 M2's intent survives. I ran the merge tests (`TestRunMerge*`, `TestMergeRefusal`, `TestAssessDirty`) at the pinned head and they pass. Each new e2e test pins a distinct path:
- the step-2 refusal (checked through its specific message);
- the 9b refusal, which only fails if the fetch is there;
- the stranded message for both the switch and the pull failure.

What stops a clean SHIP is the new recovery message. It tells the operator to *commit* the offending files and re-run, but in this flow a commit on the merged branch leads nowhere. There are also two smaller wording and dead-code issues.

### 1. Strengths
- Topology is now worked out before step 2 (`merge.go:311`), so both dirt guards share one `inPlace` value. The duplicate calculation at the old step 6 was removed rather than copied.
- `strandingTracker` maps directly onto the two git operations that can strand the checkout (switch is `main..HEAD`, pull is `main..origin/main`). It removes duplicate paths and includes rename sources (`assessDirty`, `merge.go:174-177`).
- An empty path list skips the diff, which avoids a bare `git diff --` that would list every changed file. A test pins that no git call is made in that case (`merge_test.go:420-423`).
- Fail-safe behaviour is preserved: a diff error refuses, and a fetch failure at 9b stops the merge.
- `strandedMessage` correctly tells the two failures apart: after a failed pull the checkout is already on `main`, so the message says `git switch <branch>` first.

### 2. Critical findings
None.

### 3. Important findings
- **The stranded-recovery message points to a dead end** (`merge.go:696-698`, and the same advice at 9b, `merge.go:522`).
  - The message says "Fix what git reports (commit, move aside or `git restore`…), then re-run `sdlc merge --yes`: it resumes."
  - If the operator commits on the already-merged feature branch, step 4 refuses because HEAD is ahead of `origin/<branch>`.
  - Pushing doesn't help: the resume path counts `origin/main..origin/<branch>` greater than 0 and lands in `actionResumeBlocked` (#148, "reused branch name"). The publish gate (#160) also refuses because HEAD changed since close.
  - The issue's own log admits this: "a commit made by merge itself refuses both".
  - **Fix:** recommend only the paths that actually resume: move the file aside or `git stash push -- <paths>` / `git restore`, re-run, then re-apply the edit on `main`. Change the 9b die text the same way, and have the test assert that "commit" is absent.

### 4. Minor findings
- **The tracker warning claims something unchecked** (`merge.go:331`). "they match main, so the switch carries them" is printed in the linked-worktree topology too, where nothing was compared with main. Either make it depend on `inPlace` (the worktree case doesn't switch at all) or reword it.
  - The 9b header "(likely a pre-merge judge/hook)" (`merge.go:520`) is a sibling: a refusal there can now come from a peer's push to main (the scenario in `TestRunMerge_TrackerFileChangedOnMainAfterCheck_RefusesPreMerge`), not a hook.
- **`Refuse()` no longer has any production caller** (`merge.go:158`; its only use is `merge_test.go:138`). It reads like the merge decision but isn't one — a second, weaker source of truth (ARCH-DRY). Remove it and have `TestAssessDirty` assert on `len(Blocking)`, or route the test through `mergeRefusal`.
- **Pathspecs depend on the working directory** (`merge.go:196`).
  - Porcelain paths are relative to the repo root, but `git diff -- <paths>` reads pathspecs relative to the current directory.
  - Run from a subdirectory, the diff matches nothing and the check silently lets the file through.
  - Glob characters in a path would also be treated as patterns.
  - **Fix:** add `--literal-pathspecs` and `:(top)`, or run the diff from the repo root.
- If no local `main` branch exists in an in-place clone, `git diff main HEAD` errors. The merge is then refused whenever any tracker file is dirty, even though `git switch main` would have created `main` from `origin/main`. It fails safe, but the refusal is confusing.

### 5. Test coverage notes
- Both states of the "Done when" clause are exercised:
  - refused up front: step 2 (`…DiffersFromMain_RefusesPreMerge`) and 9b (`…ChangedOnMainAfterCheck…`);
  - completes end to end: `TestRunMerge_DirtyTrackerFile_Proceeds`;
  - stranded message: both the `switch` and `pull` subtests.
- The topology split is covered at unit level (no git calls for a worktree). There is no e2e test that a linked worktree with a differing tracker file still proceeds.
- No test pins that the recovery advice actually resumes. That is the gap that let the Important finding through.

### 6. Architectural notes
- **ARCH-DRY:** pass, apart from the dead `Refuse()` noted above. Steps 2 and 9b share `mergeRefusal`, and the topology calculation was moved rather than duplicated.
- **ARCH-PURE:** pass. `mergeRefusal` and `strandingTracker` reach git only through the injected `gitRunner`, so the decision is tested with a fake runner. `strandedMessage` is pure.
- **ARCH-PURPOSE:** pass with the caveat above. The family "local post-merge recovery advice must resume" is only partly met. Both `strandedMessage` and the 9b die text suggest committing; sweep both.

### 7. Plan revision recommendations
- Add a "## Revisions" entry: the recovery advice excludes committing, because a commit on a merged branch trips step 4, #148's resume-blocked refusal and the #160 publish gate. Stash, restore or move aside is the supported path.

```findings
findings:
  - id: new
    severity: Important
    family: recovery-advice-must-resume
    title: |
      Stranded/9b messages advise committing, which makes the resume path refuse
    detail: |
      strandedMessage (merge.go:696) and the 9b die text (merge.go:522) suggest commit. A commit on the merged branch trips step 4 (ahead of origin); pushing it trips actionResumeBlocked (#148, origin/main..origin/branch > 0) and the #160 publish gate. Recommend only stash/restore/move-aside then re-run; sweep both sites and assert in TestRunMerge_StrandedAfterMerge_NamesRecovery.
  - id: new
    severity: Minor
    family: message-claim-unverified
    title: |
      Tracker warning says "they match main" in worktree topology where nothing was checked
    detail: |
      merge.go:331 prints unconditionally; in a linked worktree mergeRefusal skips the diff. Sibling: the 9b header at merge.go:520 blames a judge/hook, but a peer push to main now triggers it too.
  - id: new
    severity: Minor
    family: single-decision-source
    title: |
      Refuse() has no production callers and is no longer the merge decision (ARCH-DRY)
    detail: |
      Only merge_test.go:138 calls it; mergeRefusal is the real decision. Remove it and assert on Blocking (or mergeRefusal) in TestAssessDirty.
  - id: new
    severity: Minor
    family: pathspec-root-relative
    title: |
      strandingTracker passes root-relative porcelain paths as cwd-relative pathspecs
    detail: |
      Run from a subdirectory, the diff matches nothing and the check passes the file through; glob characters are also treated as patterns. Use --literal-pathspecs with :(top) or run git from the repo root.
```

---

## Re-review — 2026-10-12T09:16:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 196 — merge: dirty tracker files are exempted from the block but still strand the post-merge switch |
| repo | ariadne |
| issue file | workshop/issues/000196-merge-tracker-strand.md |
| boundary | whole-issue close |
| milestone | — |
| window | 505a771abe91e2393c9033ceefc5809c9041ca59..7c48c569cd650b3a698f495ac19caacad2a02757 |
| command | sdlc close --issue 196 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T09:16:25-07:00 |
| verdict | SHIP |

## Review

Prior findings checked; now running tests and checking the publish gate's tracker handling.

```verdict
verdict: SHIP
confidence: high
```

The four round-1 findings are all fixed, and the change does what #196 set out to do. `mergeRefusal` (`cmd/sdlc/merge.go:212`) is now the only merge-block decision for both step 2 and step 9b. On an in-place checkout it asks git's own question: does a dirty tracker file differ between `main` and `HEAD`, or between `main` and `origin/main`? Step 9b fetches `origin/main` before asking. A diff error refuses the merge rather than passing it. A linked worktree keeps the #82 M2 exemption, which matches the plan's topology note. `strandedMessage` says the PR is merged and gives a recovery that the resume path accepts. The merge test suite passes (`go test ./cmd/sdlc/ -run 'TestMergeRefusal|TestAssessDirty|TestRunMerge'`, ok in 18.3s). The atlas has the new paragraph, and README needs nothing because no new flags or commands were added. Nothing blocks shipping; two Minor notes follow.

### 1. Strengths
- **Asks git's question, not a semantic one.** `strandingTracker` (`merge.go:188`) runs path-limited `git diff --name-only --no-renames -z` over both ranges that matter, with `:(top,literal)` pathspecs. It deduplicates the results and includes rename sources in `TrackerPaths` (`merge.go:170-173`).
- **Guards against a bare `git diff --`.** With no tracker paths, the diff would list every file. `merge.go:214` skips it, and `TestMergeRefusal` checks that git is never called in that case.
- **The e2e tests cover every Done-when state:**
  - a tracker file that differs from main is refused before the merge (`PRMerge` is never called, the branch is unchanged);
  - a peer push to main after step 2 is caught at step 9b;
  - an identical tracker file still lands end to end (`TestRunMerge_DirtyTrackerFile_Proceeds`);
  - both post-merge failure points (`switch` and `pull`) print the stranded message, and the pull case adds `git switch <branch>`.
- **The recovery advice was checked against the resume path.** The `workshop/lessons.md` entry records why: a commit after the merge trips step 4 and #148.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- **No local `main` gives an opaque error.** On an in-place checkout with no local `main` branch and a dirty tracker file, `git diff main HEAD` fails. Merge then dies with a raw git error, even though `git switch main` would have created the branch from `origin/main`. This fails safe, as the plan intends, but the message doesn't say what to do. Not raised as a finding.
- **The 9b wording isn't tested.** The step 9b refusal text (`merge.go:520`) no longer advises a commit, but only `strandedMessage` has "do not commit" pinned by a test. Its parenthetical "a commit now would change the reviewed HEAD" is literally true. However, the publish gate lets doc-only commits through (#174), so commit plus push would also work before the merge. Not raised.

### 5. Test coverage notes
- `TestMergeRefusal` checks both topologies, deduplication across the two diffs, the no-paths guard, and that a diff error is surfaced as an error.
- The fake in that test asserts the `:(top,literal)` prefix. That is a check on how the code is written, not an end-to-end test of running from a subdirectory, but it would catch a regression of BR-4.
- The implementor reports that removing the in-place branch makes all three new refusal tests fail. That fits the test design; I didn't revert it myself.

### 6. Architectural notes
- **ARCH-DRY: pass.** `Refuse()` is gone and `mergeRefusal` is the single decision used at both guards. A grep for `Refuse()` finds no callers.
- **ARCH-PURE: pass.** `assessDirty` stays pure. The git question sits behind the injected `gitRunner`, and the unit test uses a fake.
- **ARCH-PURPOSE: pass.**
  - Every path that can strand on the legacy `runMerge` in-place topology is covered: the switch range, the pull range, and the late re-check.
  - The durable landing path already refuses any tracked dirt (per the plan's scope note).
  - Both post-merge failure sites use the recovery message, so this isn't the easy subset of the issue.

### 7. Plan revision recommendations
None. The plan and the Log describe the code as it is.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      strandedMessage (merge.go:688) says do not commit and re-run merge --yes; 9b text (merge.go:520) says restore or move aside; TestRunMerge_StrandedAfterMerge_NamesRecovery pins "do not commit" for both switch and pull.
  - id: BR-2
    disposition: addressed
    note: |
      Tracker warning now gives a per-topology reason (merge.go:328); the 9b header names main moving as a cause.
  - id: BR-3
    disposition: addressed
    note: |
      Refuse() removed (no callers remain per grep); TestAssessDirty asserts len(Blocking); mergeRefusal is the single decision.
  - id: BR-4
    disposition: addressed
    note: |
      Pathspecs are ":(top,literal)"+p (merge.go:194); the diffRunner fake in TestMergeRefusal rejects any unpinned pathspec.
```
