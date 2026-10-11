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
