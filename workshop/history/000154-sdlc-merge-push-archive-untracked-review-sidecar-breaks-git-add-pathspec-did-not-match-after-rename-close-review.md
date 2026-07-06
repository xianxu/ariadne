# Boundary Review — ariadne#154 (whole-issue close)

| field | value |
|-------|-------|
| issue | 154 — sdlc merge/push archive: untracked review sidecar breaks git-add (pathspec did not match) after rename |
| repo | ariadne |
| issue file | workshop/issues/000154-sdlc-merge-push-archive-untracked-review-sidecar-breaks-git-add-pathspec-did-not-match-after-rename.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0179049ac93bbb63f7eec77d0cb78d58aefd0d8c..HEAD |
| command | sdlc close --issue 154 |
| reviewer | claude |
| timestamp | 2026-07-05T23:12:30-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All premises confirmed empirically and by code reading. I have enough to write the review.

## Verdict

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix is correct, minimal, and well-architected: it adds a `SourceUntracked` flag to `preparedArchiveMove`, omits a vanished untracked source from the `git add` list, and classifies trackedness through an injected `git ls-files` probe (`gitSrcUntracked`) consulted before the rename. I reproduced the exact exit-128 `pathspec did not match` symptom in a throwaway repo and confirmed the fix's staging (dest-only for untracked, both halves for tracked) succeeds, and confirmed the `git ls-files` probe semantics (tracked→echo, untracked→empty, both exit 0) match the code's assumptions. Nothing blocks SHIP. The one gap worth closing: the **merge path** — the exact path the bug was reproduced on 3× — has no end-to-end test of the untracked branch, so its `GitInDir(mainPath,…)` probe wiring is unexercised.

### 1. Strengths

- **Clean IO seam, shared by both callers (ARCH-DRY + ARCH-PURE).** `gitSrcUntracked` (`push.go:322-327`) is one builder consumed by push (`pushRunner.Git`, `push.go:626`) and merge (`GitInDir(mainPath,…)` closure, `merge.go:658-659`); the probe is injected into the otherwise-pure `archivePlanArtifacts` (`push.go:285`), and the pure tests run on plain temp dirs with a faked probe. Textbook pure-core/thin-shell.
- **Conservative failure direction, documented and tested.** On any git error the probe returns "tracked" (`push.go:325`), preserving pre-#154 behavior rather than risk dropping a real deletion — pinned by `TestGitSrcUntracked` (git-error→tracked, error-with-stale-output→tracked).
- **Real-repo regression on the push helper.** `TestArchiveDoneIssues_UntrackedSidecar_RealRepo` drives the real `git ls-files` → real `git add`/`commit` and asserts a clean worktree + all three files tracked in history — this is the exact exit-128 scenario. Log documents proving it catches the bug by reverting the guard.
- **Default-false field is genuinely behavior-preserving.** Every existing move site (issue files, recovery moves in `preparedArchiveMoves`) constructs the struct without the field → tracked → stages both halves, exactly as before. The reasoning that recovery sources are inherently tracked (formed only from a staged deletion) holds.
- **Probe-before-rename ordering** (`push.go:303-304`) is correct and clearly commented.
- **Atlas updated accurately** (`atlas/workflow/sdlc-binary.md`) — the prose matches the code.

### 2. Critical findings

None.

### 3. Important findings

- **`cmd/sdlc/merge.go:658` — the merge path's untracked-sidecar wiring is untested end-to-end, and merge is the path the bug was reproduced on.** `TestArchiveDoneIssuesInDir_SweepsPlanArtifacts` (`archiveartifacts_test.go:243`) runs on a bare temp dir, not a git repo, so its `git ls-files` probe *always errors → tracked*; the `SourceUntracked=true` branch through `gitSrcUntracked(GitInDir(mainPath,…))` is never exercised. The push side has a real-repo test; the merge side does not, even though the issue documents 3× reproduction at `sdlc merge`. **Fix sketch:** add a hermetic-repo merge test mirroring `TestArchiveDoneIssues_UntrackedSidecar_RealRepo` — create a main worktree (or in-place `hermeticRepo`), commit the issue + durable plan, leave the sidecar untracked, run `archiveDoneIssuesInDir(…, mainPath, …)` + the real `GitInDir(mainPath, archiveAddArgs(moves)...)` + commit, and assert a clean worktree. Cheap, and it pins the exact production path that failed. Non-blocking because the shared helpers (`archiveAddArgs`, `archivePlanArtifacts`, `gitSrcUntracked`) are covered and the merge closure is a thin analogue of the tested push one.

### 4. Minor findings

- `merge.go:216-220` — the `mergeAction` const-block whitespace realignment is unrelated to #154 (incidental `gofmt -w` of pre-existing drift, per the Log). Harmless, just scope-adjacent; fine to leave.

### 5. Test coverage notes

- Pure seams: `TestArchiveAddArgs` (untracked→dest-only, mixed) and `TestGitSrcUntracked` (all five classification cases including the `--`-guarded arg shape) are thorough and pin real logic, not mocks.
- `TestArchivePlanArtifacts_UntrackedSidecarStagesDestOnly` correctly uses a **faked** probe (not `os.Stat`), pinning that classification is index-based — a good guard against a working-tree-check regression.
- Gap: merge-path real-repo coverage (Important finding above). Everything else the diff could break is covered.

### 6. Architectural notes for upcoming work

- **Separate-worktree merge topology is out of scope here but worth a follow-up.** On a linked-worktree merge (mainPath ≠ wtPath), an untracked sidecar lives in `wtPath/workshop/plans/`, which `archiveDoneIssuesInDir` (globbing `mainPath/plans`) never sees — and `wtPath` is `worktree remove`d after merge. So in that topology the untracked sidecar is neither archived nor recovered. #154 correctly scopes to the *present-and-untracked* case (which matches the reproduced in-place topology where the sidecar was already moved to history), so this is not a #154 defect — but if linked-worktree merges are used, consider a separate issue to sweep the wtPath sidecar before removal.

### 7. Plan revision recommendations

None. The single Plan checklist item (`SourceUntracked` field + probe injection + `archiveAddArgs` omission + pure-seam table test + `archivePlanArtifacts` untracked regression) is fully delivered, and the Core-concepts reasoning in the Plan matches the code. No `## Revisions` entry needed.
