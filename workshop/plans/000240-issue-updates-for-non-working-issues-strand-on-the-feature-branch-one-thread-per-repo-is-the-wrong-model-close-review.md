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

---

## Re-review — 2026-10-14T10:23:27-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 240 — Issue updates for non-working issues strand on the feature branch — one thread per repo is the wrong model |
| repo | ariadne |
| issue file | workshop/issues/000240-issue-updates-for-non-working-issues-strand-on-the-feature-branch-one-thread-per-repo-is-the-wrong-model.md |
| boundary | whole-issue close |
| milestone | — |
| window | 407e48a39057503fab7cdecaf6c69ce338ce7631..48255461dfd57311a8fd24a74d818826f97320d3 |
| command | sdlc close --issue 240 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T10:23:27-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The publication path, conflict checks, literal filenames, and documentation are substantially covered. Two existing findings still block shipping: interruption after cleanup’s rename cannot recover through the documented retry, and writes through an open descriptor can still be silently discarded. The reviewed checkout remains unchanged; the disposable checkout was removed.

Done-when: current — it covers the narrowed publication surface, retry behavior, local-work preservation, tests, and atlas documentation.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      ARCH-ORDER / ARCH-FUNERAL: cmd/sdlc/trunkpublish.go:276 creates an intermediate renamed-backup state that retry cannot recover. A deterministic interruption at trunkProseAfterMove reproduced both failures: tracked retry refuses at line 183 as a local deletion; untracked retry reports already clean at line 185 while retaining the backup. Define authoritative cleanup states and safe recovery for every intermediate effect, including process death, with production-transition sequence tests.
  - id: BR-2
    disposition: addressed
    note: |
      Plan Tasks 1–2 now name adversarial inputs and mechanical guards for the path predicate, publication decision, command, and cleanup. Lexical eligibility is separated from filesystem inspection; the pure tests invoke the predicates directly without IO. Cleanup strategy completeness remains covered by BR-1 and BR-6.
  - id: BR-3
    disposition: addressed
    note: |
      The issue Plan uses plain checkboxes; the implementation plan contains no M1 boundary or milestone-close. This matches AGENTS.base.md:51's single-pass close rule.
  - id: BR-4
    disposition: addressed
    note: |
      The plan specifies a 1 MiB prose envelope, two-minute shared remote deadline, cancellation, and three CAS attempts. cmd/sdlc/trunkpublish.go enforces the file-size limit and invokes BoundRemote; the existing Git seam supplies cancellation and the CAS budget.
  - id: BR-5
    disposition: addressed
    note: |
      Selected-path staging checks use literal pathspecs and cleanup reads the exact pinned HEAD blob instead of Git restore patterns. TestTrunkPublishTreatsSelectedWildcardFilenameLiterally passes at pinned head and fails when the pre-fix cleanup is restored in a disposable checkout: unrelated other.md loses its dirty edit.
  - id: BR-6
    disposition: not-addressed
    note: |
      ARCH-ORDER: cmd/sdlc/trunkpublish.go:313–317 and 332–336 still separate the final backup-byte check from unlink. A disposable overlay added only an ordering hook immediately before unlink; writing through a descriptor opened before rename made both tracked and untracked cleanup return nil with no backup, while the later edit disappeared. Exclusive creation also does not exclude subsequent writers during replacement.Write at line 327. Preserve recoverable data throughout cleanup and test all destructive-effect boundaries, rather than only the after-move event.
  - id: BR-7
    disposition: addressed
    note: |
      README.md:28–37 adds command usage, eligibility, moved-main refusal, uncertain-push retry guidance, and an atlas link. These passages match the allowlist and publication checks; atlas/workflow/sdlc-binary.md also documents the new surface.
```

1. **Strengths**

   - Reuses `gitx.TrunkFile.UpdateMany` for CAS publication rather than introducing another remote transaction mechanism.
   - Separates lexical eligibility from filesystem checks and tests the pure decisions directly.
   - Real-Git tests verify branch/index isolation, selected-path conflicts, and preservation of unrelated dirty work.
   - README and atlas both document the new command.

2. **Critical findings**

   - **BR-6 remains open:** `cmd/sdlc/trunkpublish.go:317,336` can unlink a concurrently modified inode after validation. Preserve recoverable bytes across the entire destructive sequence and add controllable late-write tests.

3. **Important findings**

   - **BR-1 remains open:** `cmd/sdlc/trunkpublish.go:181–186,276` cannot reconcile interruption after rename. Add a recoverable cleanup state model that distinguishes incomplete cleanup from genuine local deletion or completed cleanup.

4. **Minor findings**

   - None newly raised.

5. **Test coverage notes**

   The targeted recovery, eligibility, decision, publication, and cleanup tests passed. The wildcard regression went red with pre-fix cleanup restored. Two additional deterministic overlays exposed the outstanding interruption and late-write failures. The full suite was not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing CAS and Git boundaries are reused.
   - **ARCH-PURE — pass:** eligibility and publication decisions are independently tested without IO.
   - **ARCH-PURPOSE — flag:** BR-1/BR-6 leave the promised recovery and preservation behavior incomplete.
   - **ARCH-MOCK — pass:** existing shared Git seam and stateful fixtures are reused.
   - **ARCH-CONSTRAINTS — pass:** explicit input and remote budgets are present.
   - **ARCH-SECURE — flag:** BR-6 treats a byte observation as authority for a later destructive effect.
   - **ARCH-ORDER — flag:** intermediate cleanup states and late writes remain insufficiently modeled.
   - **ARCH-FUNERAL — flag:** interrupted cleanup leaves backups without an automatic recovery/removal path.

7. **Plan revision recommendations**

   Add a `## Revisions` entry enumerating cleanup states, interruption points, permitted recovery actions, and backup lifecycle. Correct the claim that one selected path makes cleanup a single transition, and replace the after-move-only test strategy with coverage of every destructive-effect boundary.
