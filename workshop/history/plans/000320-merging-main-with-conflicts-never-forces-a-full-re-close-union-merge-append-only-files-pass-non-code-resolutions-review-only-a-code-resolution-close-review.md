# Boundary Review — ariadne#320 (whole-issue close)

| field | value |
|-------|-------|
| issue | 320 — Merging main with conflicts never forces a full re-close: union-merge append-only files, pass non-code resolutions, review only a code resolution |
| repo | ariadne |
| issue file | workshop/issues/000320-merging-main-with-conflicts-never-forces-a-full-re-close-union-merge-append-only-files-pass-non-code-resolutions-review-only-a-code-resolution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8b9052040674db7b431ad3db3844837e0c561366..aca4448eae6b42f1c50b1d0beefe53c5b90cd734 |
| command | sdlc close --issue 320 |
| reviewer | claude |
| timestamp | 2026-10-10T17:23:58-07:00 |
| verdict | SHIP |

## Review

Still reviewing: checking that weave's real compile path (`ApplyManaged`) actually runs the new `.gitattributes` action.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: ship.** The change does what layers 1–2 of the revised Spec promise, and I found nothing that blocks it.

- **Weave:** it writes a managed `.gitattributes` block that gives `workshop/lessons.md` git's `merge=union`.
- **Replay:** the reviewed-patch replay now reads merge attributes from main's tree (`--attr-source=<B_now>`), not from the working tree. The replayed result therefore depends only on its inputs.
- **Publish gate:** a merge conflict confined to non-code files now passes, with a line naming the resolved files. A conflict in a code file still refuses with the plain re-close hint.

I ran the weave suite and the targeted sdlc tests (`TestRebasedReviewedBase*`, `TestClassifyPublishDelta`, `TestPublishDeltaRank_NonCodeConflict`, `TestRunPublishGate_BranchPatch`); all pass. Layer 3 moving to #183 is recorded in `## Revisions`, and the Done-when bullet is struck through to match. What's left is test coverage and documentation polish.

1. **Strengths**
   - `cmd/sdlc/internal/gitx/rebasedpatch.go:65`: main's attributes decide the replay. The test deletes the worktree `.gitattributes` before replaying, which proves the result doesn't depend on the working tree.
   - `cmd/sdlc/publishgate.go:310`: `codeSurfacePaths` pulls out the per-path loop that was duplicated in `publishDeltaRank` and `classifyPublishDelta` (ARCH-DRY). Conflicts are classified by file kind through the same `publishGateHasCodeSurface` test as #174, not by a `workshop/` path prefix. The embedded `helptext/*.md` case pins this.
   - `publishgate_test.go:396-426`: every pass line is checked against the gate catalog's refusal pattern, so a pass line can't be misread as a refusal (#172).
   - The end-to-end test covers the pair#426 shape both ways: resolved by hand without the attribute, and merged cleanly with union.
   - `writeManagedFile` was extracted from `writeManagedIgnore` rather than copied.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **No compile-level test for the new action.** `EnsureGitattributes` is only tested through `plan.Apply`. Reading the code, the real compile path (`ApplyManaged` → `materializeManaged`'s default branch → `Apply`) does write the file. But no test runs `planActions`/`ApplyManaged` and asserts `.gitattributes` appears. The Log also says `weave compile` itself never ran in the sandbox.
   - **Error text says "gitignore" for `.gitattributes` problems.** The reused `splitIgnore` reports "malformed weave gitignore block" even when the bad block is in `.gitattributes`. The `managedIgnoreText` name is also misleading now that two files share it.
   - **The compaction caveat lives only in the issue Log.** Under union merge, a lessons.md compaction that lands at the same time as someone's append keeps both versions without any conflict. "Land compactions alone" should be stated in the `UnionMergeAttributes` doc comment or in atlas `weave.md`.
   - **Conflict markers can pass unreviewed.** A non-code resolution that still contains `<<<<<<<` passes the gate. This is acceptable for docs, but a cheap marker check would close it.

5. **Test coverage notes**
   - The tests cover:
     - the four cases of non-code vs code conflicts, including a mixed set;
     - the rank ordering;
     - replay under union attributes, independent of the worktree;
     - block create, idempotence, authored-line precedence and non-regular-file refusal.
   - The only gap is the compile-level wiring above.

6. **Architecture**
   - **ARCH-DRY:** pass (`codeSurfacePaths`, `writeManagedFile`).
   - **ARCH-PURE:** pass. `classifyPublishDelta` and `publishDeltaRank` stay pure, and the replay result is now a pure function of its git inputs.
   - **ARCH-PURPOSE:** pass for the revised scope. Layer 3 went to #183 by an explicit scope revision, not a quiet deferral.
   - **ARCH-MOCK:** pass. git tests run against real temp repos through the existing `testfix` setup.
   - **ARCH-CONSTRAINTS:** pass. Raising the git floor to 2.42 is loud and documented; no hot-path cost.
   - **ARCH-SECURE:** pass, apart from the conflict-marker note above.
   - **ARCH-ORDER:** pass. The gate keeps no state between events because it re-derives the replay on every publish.
   - **ARCH-FUNERAL:** pass. The one managed block is rewritten on every compile and doesn't grow.

7. **Plan revisions:** none needed; the plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: wiring-untested-at-entrypoint
    title: |
      EnsureGitattributes is tested via plan.Apply only; no planActions/ApplyManaged test asserts .gitattributes is written
    detail: |
      By reading, the ApplyManaged default branch passes the action through to Apply, but weave compile never ran (sandbox) and no compile-level test pins the planActions wiring.
  - id: new
    severity: Minor
    family: reused-helper-misleading-name
    title: |
      splitIgnore/managedIgnoreText errors say "gitignore block" when the bad block is in .gitattributes
  - id: new
    severity: Minor
    family: union-merge-invariant-undocumented
    title: |
      The "land lessons.md compactions alone" caveat for union merge is only in the issue Log, not in UnionMergeAttributes' doc comment or atlas
  - id: new
    severity: Minor
    family: unreviewed-resolution-content
    title: |
      A non-code conflict resolution that still contains conflict markers passes the publish gate
```

---

## Re-review — 2026-10-10T17:50:12-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 320 — Merging main with conflicts never forces a full re-close: union-merge append-only files, pass non-code resolutions, review only a code resolution |
| repo | ariadne |
| issue file | workshop/issues/000320-merging-main-with-conflicts-never-forces-a-full-re-close-union-merge-append-only-files-pass-non-code-resolutions-review-only-a-code-resolution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8b9052040674db7b431ad3db3844837e0c561366..94dbc3d8a9f7450d44e1a04240df8ca1e55aff22 |
| command | sdlc close --issue 320 |
| reviewer | claude |
| timestamp | 2026-10-10T17:50:12-07:00 |
| verdict | SHIP |

## Review

Verifying the four prior findings against the diff, then running the targeted tests.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** I found no new issues, and all four findings from the last round are fixed and backed by tests.

**Layer 1 (stop the conflict).** Weave now writes a managed block into `.gitattributes` that marks `workshop/lessons.md` as `merge=union`, so appends from parallel branches merge without conflict. It reuses the same block code as `.gitignore` and is committed in ariadne. The review replay (`RebasedReviewedBase`) now takes merge attributes from main's tree rather than the working tree. That keeps the replay a pure function of its inputs, but it raises the minimum git version to 2.42. The code, the doc comment and the atlas all say 2.42 now, and no stale 2.40 floor remains.

**Layer 2 (don't re-review non-code conflicts).** The publish gate now refuses a conflict only when a conflicted path is code. A conflict confined to non-code files passes with a line naming those files. A non-code file that still has conflict markers at `HEAD` is refused with its own message.

**Tests.** The targeted tests all pass on git 2.54: gitx `RebasedReviewedBase*`, sdlc `ClassifyPublishDelta|HasConflictMarkers|PublishDeltaRank|RunPublishGate_BranchPatch`, and weave compile/golden/gitattributes.

1. **Strengths**
   - `publishgate.go:318`: `codeSurfacePaths` replaces two copies of the same "which paths are code" loop, and both the rank and the classify function use it.
   - `publishgate.go:223`: the marker check reuses the existing `hasConflictMarkers` (`republish.go:404`) instead of adding a second detector.
   - `TestClassifyPublishDelta` checks every pass line against the gate catalog's refusal pattern, so the new "merge resolution … has no code surface" line can't be misread as a refusal.
   - `TestRunPublishGate_BranchPatch` covers three real-branch cases: a conflict resolved by hand, a resolution committed with its markers, and a union merge.
   - `TestRebasedReviewedBase_UnionAttributeFromMain` deletes the working-tree `.gitattributes` before checking the replay. That proves the attributes really come from main's tree.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. One known limit, not raised as a finding: if a non-code file legitimately contains real conflict-marker lines, such as a quoted example in a lessons entry, the marker check would refuse it. That is unlikely, and the refusal is visible.
5. **Test coverage:** each piece of new behavior has a test that would fail without it:
   - marker refusal → the "committed with its markers" subtest;
   - non-code pass → the hand-resolved subtest;
   - compile wiring → `TestCompileEnsuresGitattributes`;
   - attributes taken from main → the gitx test.
6. **Architecture:**
   - **ARCH-DRY** pass: `codeSurfacePaths`, `writeManagedFile` and `hasConflictMarkers` are all reused.
   - **ARCH-PURE** pass: `classifyPublishDelta` and `publishDeltaRank` stay pure. Git input is collected in the thin loop that calls them.
   - **ARCH-PURPOSE** pass: Done-when layers 1–2 are delivered. Layer 3 was moved to #183 by a recorded revision.
   - **ARCH-MOCK** pass: git is called through the existing `gitx` seam, and tests run against real temp repos, as elsewhere in the repo.
   - **ARCH-CONSTRAINTS** pass: the change adds one `git show` per non-code conflicted path, which is bounded.
   - **ARCH-SECURE** pass: the `HEAD` content it reads is only used as a refuse/pass signal.
   - **ARCH-ORDER** pass: no state is kept between events.
   - **ARCH-FUNERAL** pass: the `.gitattributes` block is rewritten on every compile, so it never grows. The temporary replay commit is unreferenced and garbage-collected.
7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestCompileEnsuresGitattributes (cmd/weave/main_test.go) runs a real compile twice and pins the exact block; passes.
  - id: BR-2
    disposition: addressed
    note: |
      splitIgnore errors now say "weave generated block" (managed_ignore.go:35,41,55), neutral across both files.
  - id: BR-3
    disposition: addressed
    note: |
      Caveat now in the UnionMergeAttributes doc comment (gitattributes.go:14-16) and atlas/workflow/weave.md.
  - id: BR-4
    disposition: addressed
    note: |
      d.Markers populated at publishgate.go:223 and refused in classifyPublishDelta; the integration subtest "committed with its markers refuses" goes red without it.
```
