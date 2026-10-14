# Boundary Review — ariadne#240 (whole-issue close)

| field | value |
|-------|-------|
| issue | 240 — Issue updates for non-working issues strand on the feature branch — one thread per repo is the wrong model |
| repo | ariadne |
| issue file | workshop/issues/000240-issue-updates-for-non-working-issues-strand-on-the-feature-branch-one-thread-per-repo-is-the-wrong-model.md |
| boundary | whole-issue close |
| milestone | — |
| window | 407e48a39057503fab7cdecaf6c69ce338ce7631..c66f2889426d7d9024853f455ab9eeac0e206efc |
| command | sdlc close --issue 240 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T09:36:09-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The publication path correctly reuses the existing CAS machinery, and the targeted tests pass. Two confirmed cleanup defects can discard unpublished edits, blocking shipment. The new command also needs README coverage.

Done-when: current — it covers the narrowed prose-publication scope, conflict refusal, retry, cleanup preservation, tests, and atlas documentation.

1. **Strengths**

   - The presence-aware decision distinguishes absent files from empty files and refuses conflicting main edits (`cmd/sdlc/trunkpublish.go:56`).
   - Publication reuses `TrunkFile.UpdateMany`, including its bounded CAS retry and uncertain-outcome handling (`cmd/sdlc/trunkpublish.go:196`).
   - Real-Git tests verify feature-branch isolation, unrelated staged-work preservation, and retry without another main commit.
   - Atlas documents both publication paths and the reviewed home for code, settings, and atlas changes (`atlas/workflow/sdlc-binary.md:83`).

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: literal-selected-path
    title: |
      Cleanup interprets the selected filename as a Git pattern and discards unrelated edits
    detail: |
      cmd/sdlc/trunkpublish.go:250 passes p to git restore without literal-pathspec handling, although trunkProsePath accepts wildcard characters. A temporary regression overlay invoking runTrunkPublish with a tracked workshop/parley/*.md filename published successfully but restored an unrelated dirty other.md to its base bytes. Use literal pathspecs consistently for selected-path Git operations and add regression coverage for wildcard filenames.
  - id: new
    severity: Critical
    kind: defect
    family: cleanup-preserves-concurrent-work
    title: |
      Cleanup can overwrite edits made after its validation
    detail: |
      cmd/sdlc/trunkpublish.go:232-253 reads and compares the file, then separately restores or removes it; these destructive effects are not conditional on the validated file still being current. A deterministic temporary overlay injected an edit when the restore subprocess started: finishTrunkPublish returned nil and replaced that edit with base bytes. ARCH-ORDER: redesign cleanup to preserve intervening writes, with explicit cleanup transitions and controllable ordering tests for tracked restore, untracked removal, and HEAD/index changes.
  - id: new
    severity: Important
    kind: defect
    family: readme-user-surface
    title: |
      README update appears missing for trunk publish --path
    detail: |
      cmd/sdlc/main.go:151 introduces a user-facing command and cmd/sdlc/trunkpublish.go:34 introduces its --path flag, but README.md has no change in the pinned range. Add usage, eligibility, conflict, and retry guidance alongside README.md:15's concurrent-work documentation, linking to the updated atlas.
```

2. **Critical findings**

   Both findings above violate the promised preservation of local work. The existing “later local edit” test changes bytes before cleanup begins; it does not exercise an edit between validation and destruction.

3. **Important findings**

   README documentation is missing for the new command. Atlas coverage is present.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Targeted command, recovery-contract, cleanup, and pure-function tests passed.
   - Targeted `gitx` publication, retry, and cancellation tests passed.
   - Both reviewer-added regression overlays failed against the pinned head, confirming the defects through production functions.
   - Temporary fixtures and overlays were removed; the checkout remains unchanged.
   - The full-suite result recorded in the issue was not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** publication shares the existing CAS machinery.
   - **ARCH-PURE — pass:** eligibility and publication decisions are pure and tested without IO; Git/filesystem effects reside in integration functions.
   - **ARCH-PURPOSE — flag:** destructive cleanup defeats the issue’s goal of preserving unrelated and concurrent work.
   - **ARCH-MOCK — pass:** portable real-Git fixtures and existing stateful publication tests exercise the shared Git boundary.
   - **ARCH-CONSTRAINTS — pass:** the implementation declares a 1 MiB file limit, remote deadline, and bounded CAS attempts.
   - **ARCH-SECURE — flag:** an accepted filename crosses into Git’s pattern language during cleanup; literal-path handling is required.
   - **ARCH-ORDER — flag:** validation and cleanup effects admit destructive interleavings; current tests omit those schedules.
   - **ARCH-FUNERAL — pass:** no new sidecar or receipt family is introduced; publication uses existing Git history retention.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining literal selected-path handling and the revised cleanup state/effect contract. Enumerate tracked restoration, untracked removal, intervening file edits, and HEAD/index changes, with deterministic regression tests for each applicable ordering. Add README documentation to Task 3 before marking the boundary ready.
