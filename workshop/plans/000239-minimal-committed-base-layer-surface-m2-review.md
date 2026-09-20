# Boundary Review — ariadne#239 (milestone M2)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | a0939beeec126ec717ec94889107c93f0f3f76dd..176513e65dd6aad117a3354c43a1007c1170ebcf |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | codex |
| timestamp | 2026-09-20T14:28:53-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Startup, owner builds, bootstrap and CI are well covered, but generator writes bypass the new ownership protections. Scratch executions reproduced both overwritten authored edits and unreclaimable outputs after failed generation. The core-concepts table also misclassifies an integration component.

1. **Strengths**
   - Managed ignores derive exact paths and preserve local negations.
   - Inventory validation rejects malformed identities and symlink parents.
   - Apply failure tests cover durable preparation and final-save failure.
   - Bootstrap/Make/CI fixtures pass; source CI builds the candidate before consuming generated helpers. README and atlas updates cover the new surfaces.

2. **Critical findings**
   - Generator execution overwrites edited outputs before ownership validation.
   - Failed generation leaves outputs without recoverable ownership.
   - The updated PURE table points to filesystem/process orchestration rather than a pure entity.

3. **Important findings**
   - The new tools API requires concrete `ExecRunner`, preventing injection of the stateful external double required by the review contract.

4. **Minor findings:** None.

5. **Test coverage**
   - Passed: `go test ./cmd/weave/... ./pkg/layergraph/... -count=1`, the three bootstrap/portable Make/portable CI shell fixtures, and range `git diff --check`.
   - Scratch reproduction: compile → edit generated `SKILL.md` → compile returned success and erased the edit.
   - Scratch reproduction: generator writes then exits 1 → successful retry omitted that file from ownership → removing the generator left the file behind.
   - `TestGeneratedSnapshotDoesNotReclaimEditedOwnedFile` never executes a generator between snapshots, so it misses the overwrite.
   - Repository files remain unchanged.

6. **Architecture**
   - **ARCH-DRY — pass:** shared graph, scoped inventory and derived ignore entries.
   - **ARCH-PURE — flag:** incorrect core-concepts classification.
   - **ARCH-PURPOSE — flag:** preservation and interrupted-generation recovery remain incomplete.
   - **ARCH-MOCK — flag:** concrete tools runner prevents stateful fake injection.
   - **ARCH-CONSTRAINTS — pass:** sequential graph traversal; no new unbounded fan-out.
   - **ARCH-SECURE — flag:** inventory parsing is defensive, but generator writes precede ownership authorization.
   - **ARCH-ORDER — flag:** generator failure loses recoverable progress before inventory publication.
   - **ARCH-FUNERAL — flag:** failed-generation residue has no reliable retirement path.

7. **Plan revisions**
   - Append a dated `## Revisions` entry defining ownership checks before generator publication and recovery for partial generation, process interruption, and subsequent planning/Apply failure.
   - Correct the setup-discovery row to INTEGRATION; name `ToolEnvironment` separately as PURE.
   - Record the injectable tools-execution boundary and its stateful test fixture.

```findings
findings:
  - id: new
    severity: Critical
    family: generated-output-write-ownership
    title: |
      Generators overwrite authored replacements before ownership checks
    detail: |
      cmd/weave/main.go:522 executes generators directly against final output paths before ApplyManaged at line 540. A scratch compile, authored edit to generated SKILL.md, and second compile returned success while erasing the edit. SnapshotGenerated only observes; GeneratedActions then accepts changed bytes as generated. Enforce ownership before publication for every generator output, including files and links. Add a compile-level regression that actually runs the generator over an edited destination. ARCH-SECURE, ARCH-PURPOSE.
  - id: new
    severity: Critical
    family: durable-staging-reclamation
    title: |
      Failed generation leaves outputs outside durable ownership recovery
    detail: |
      cmd/weave/main.go:522-535 returns on generator or subsequent planning failure before recording generated identities. In a scratch fixture, a generator wrote SKILL.md then exited 1; successful retry treated its unchanged bytes as unowned at ownership.go:393, and removing the generator left the file behind. This is the 2nd finding in family durable-staging-reclamation. State and enforce one recovery rule for every pre-publication writer, including generators, rather than patching one exit path. Preserve durable provenance across failure/process death and test retry followed by retirement. ARCH-ORDER, ARCH-FUNERAL.
  - id: new
    severity: Critical
    family: core-concept-purity-classification
    title: |
      The updated PURE setup entity has no corresponding pure implementation
    detail: |
      workshop/plans/000239-minimal-committed-base-layer-surface-plan.md:113 points its PURE setup-input entity to startup/dependencies.go. That file exposes Dependencies, which performs fs.Stat at line 18 and runner.Run at line 32; it contains no corresponding pure setup-input model. Correct the table to classify discovery/execution as INTEGRATION and append the required revision. Keep genuinely pure PATH composition separately identified. Critical under the supplied core-concepts cross-check rule. ARCH-PURE.
  - id: new
    severity: Important
    family: external-interaction-test-seam
    title: |
      Tools execution requires the concrete process runner
    detail: |
      cmd/weave/internal/startup/tools.go:15 accepts weavefs.ExecRunner and mutates its Stdin, unlike Dependencies' injectable Runner interface. Tests therefore require real Make and cannot substitute the required stateful external double. This is the 2nd finding in family external-interaction-test-seam. Apply the shared rule that production and fake execution consume the same invocation boundary, including stdin; retain real-Make conformance tests alongside the fake. ARCH-MOCK.
```

---

## Re-review — 2026-09-20T14:43:06-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | a0939beeec126ec717ec94889107c93f0f3f76dd..24fe59e4cab3354f3696df9fc459f1cd1c003b2f |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | codex |
| timestamp | 2026-09-20T14:43:06-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The four prior findings are addressed, with regression evidence for the generator fixes. Two additional publication defects block M2: interrupted file writes defeat ownership recovery, and staged executables lose their permissions. The committed Go and shell suites pass; review-only regressions expose both gaps.

1. **Strengths**
   - Generators publish through ownership checks; reverting staging in scratch makes the authored-replacement and failed-generation tests fail.
   - Tools execution shares an injectable stdin boundary with a stateful fake while retaining real-Make tests.
   - README and atlas document startup ordering, marker migration, ownership, and explicit PATH setup.

2. **Critical findings**
   - **Partial publication breaks recovery** — [apply.go:263](/Users/xianxu/workspace/ariadne/cmd/weave/internal/plan/apply.go:263), [ownership.go:255](/Users/xianxu/workspace/ariadne/cmd/weave/internal/plan/ownership.go:255). Final files use truncating writes, while the inventory records only complete old/new hashes. A scratch filesystem that writes half the bytes before returning an error causes retry to fail with `preserving authored replacement`. Retirement likewise cannot recognize those partial bytes.

     **This is the 3rd finding in family `durable-staging-reclamation`.** State and enforce the rule across every publication writer: interruption must leave either a recognized complete output or durably owned recoverable staging. Use atomic publication with recoverable temporary files; test partial writes, retry, and retirement. **ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE.**

3. **Important findings**
   - **Generated executables lose permissions** — [staged.go:125](/Users/xianxu/workspace/ariadne/cmd/weave/internal/plan/staged.go:125). Staged regular files become content-only `WriteFile` actions; new destinations are created as `0644`. A scratch regression publishing a generated `0755 run.sh` observes `0644`. Carry permissions through publication and test cold publication plus warm permission changes. **ARCH-PURPOSE.**

4. **Minor findings**
   - None.

5. **Test coverage notes**
   - Passed: `go test ./cmd/weave/... ./pkg/layergraph/... -count=1`, bootstrap-transitive, portable-makefile, portable-ci, and pinned-range `git diff --check`.
   - Existing partial-Apply tests fail *before* writing a file; they miss truncation followed by failure.
   - Scratch mutations verified BR-7/BR-8 regression sensitivity. Repository files remain unchanged.

6. **Architectural notes**
   - **ARCH-DRY — pass:** clone and generator staging share recovery machinery.
   - **ARCH-PURE — pass:** the corrected setup table separates pure PATH composition from integration.
   - **ARCH-PURPOSE — flag:** publication must preserve executable behavior and fulfill interruption recovery.
   - **ARCH-MOCK — pass:** tools now support the shared production/fake invocation boundary.
   - **ARCH-CONSTRAINTS — pass:** sequential setup introduces no unbounded fan-out.
   - **ARCH-SECURE — pass:** staged publication checks destination ownership and rejects unsafe output paths.
   - **ARCH-ORDER — flag:** partial final writes fall outside the recoverable identity states.
   - **ARCH-FUNERAL — flag:** partial output can survive retirement without recognized ownership.

7. **Plan revision recommendations**
   - Append a `## Revisions` entry defining atomic final publication and recovery of interrupted publication staging.
   - Define generated-file permission preservation, including how permission changes affect ownership identity.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      Generators now stage before ownership-checked publication. Reverting staging in scratch makes real-generator authored file/link replacement regressions fail.
  - id: BR-8
    disposition: addressed
    note: |
      The reported prepublication failure is covered by shared durable stages and failure/death/retry/retirement tests. Reverting staging makes the failed-generator publication regression fail; partial final publication is a separate finding below.
  - id: BR-9
    disposition: addressed
    note: |
      The active table now lists only ToolEnvironment as pure, matching its implementation; the appended M2 review revision classifies discovery and execution as integration.
  - id: BR-10
    disposition: addressed
    note: |
      Tools accepts weavefs.InputRunner and passes stdin through RunInput. The injected stateful failure/retry test and retained real-Make tests pass.
findings:
  - id: new
    severity: Critical
    family: durable-staging-reclamation
    title: |
      Partial final-file writes escape ownership recovery
    detail: |
      apply.go:263 uses truncating WriteFile, but ownership.go:255 recognizes only complete old/new hashes. Scratch injection writing half the output before returning an error makes retry refuse it as an authored replacement; retirement cannot identify it either. This is the 3rd finding in this family: enforce atomic or durably recoverable publication across all writers, with partial-write/retry/retirement regressions. ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE.
  - id: new
    severity: Important
    family: generated-artifact-fidelity
    title: |
      Staged publication discards executable permissions
    detail: |
      staged.go:125 lowers regular outputs to content-only WriteFile actions, and weavefs/fs.go:72 creates them as 0644. A scratch regression publishing a generated 0755 run.sh fails because the destination is 0644. Preserve permissions and test cold publication and warm permission changes. ARCH-PURPOSE.
```

---

## Re-review — 2026-09-20T14:58:01-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | a0939beeec126ec717ec94889107c93f0f3f76dd..f94d52c10d6d475a3c1feb4f4d2443f5f8e3853a |
| command | sdlc milestone-close --issue 239 --milestone M2 |
| reviewer | codex |
| timestamp | 2026-09-20T14:58:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-11 and BR-12 are addressed, with regression tests that fail when their fixes are removed. The existing suites pass, and README/atlas updates cover the new startup and generator contracts. However, BR-8 remains incomplete: a generator child can survive cancellation, recreate an already-cleaned stage without ownership metadata, and escape subsequent reclamation.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      Shared atomic publication protects final files. Removing it in a scratch copy makes TestManagedPartialFileWritePreservesIdentityForRetryAndRetirement fail on both cold and warm partial writes.
  - id: BR-12
    disposition: addressed
    note: |
      Staged modes propagate through publication and ownership identities. Removing mode propagation makes TestStagedFileModesAreOwnedAndPreserved fail with 0644 instead of 0755; production compile also tests executable permissions.
  - id: BR-7
    disposition: addressed
    note: |
      Generators write into staging before managed ownership checks; compile regressions preserve edited files, symlink replacements, and cold authored destinations.
  - id: BR-8
    disposition: not-addressed
    note: |
      Cancellation terminates only the immediate marker process. A scratch production-path regression confirms its child can recreate the removed generation stage without owner.json; a subsequent compile with the marker removed leaves that residue intact. See main.go:524 and internal/weavefs/runner.go:42. ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE; existing family durable-staging-reclamation.
  - id: BR-9
    disposition: addressed
    note: |
      The revised active concept table identifies ToolEnvironment as pure and discovery/execution as integration, consistent with the implementation.
  - id: BR-10
    disposition: addressed
    note: |
      Tools accepts InputRunner; production Make and the stateful failure/retry test double share that interface, with real Make conformance tests.
```

1. **Strengths**
   - File publication is centralized and atomic across generated files, seeds, inventory, and ignores.
   - Permission ownership preserves generated executables and protects subsequent authored permission edits.
   - Startup tests exercise real Make ordering, build failures, bootstrap paths with spaces, and actual CI shell blocks.
   - Output retirement uses recorded identities and preserves unrelated files.

2. **Critical findings**
   - **BR-8 — generator lifetime exceeds stage lifetime.** At [runner.go:42](/Users/xianxu/workspace/ariadne/cmd/weave/internal/weavefs/runner.go:42), `exec.CommandContext` cancels the immediate shell, without ensuring its descendants stop. [main.go:524](/Users/xianxu/workspace/ariadne/cmd/weave/main.go:524) then removes their stage. A surviving child can recreate it after cleanup; reclamation skips it because `owner.json` is gone.

     I reproduced this with a marker invoking a child blocked on a fixture signal: cancel compile, wait for compile to return, release the child, then compile again without the marker. The child’s staged `SKILL.md` survives.

     This repeats family `durable-staging-reclamation`, already carrying three findings. Apply the rule across producer lifecycles: **ownership metadata must outlive every process capable of writing the payload**. Bound and terminate descendant execution before cleanup; parent death alone must not establish that a producer has stopped. Cover cancellation, abrupt parent death, late writes, retry, and retirement. **ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE.**

3. **Important findings:** None additional.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed: `go test ./cmd/weave/... ./pkg/layergraph/... -count=1`, bootstrap, portable Makefile, portable CI fixtures, and diff whitespace checks.
   - Scratch mutation tests confirmed both open fixes have effective regression coverage.
   - Existing process-death coverage immediately exits the marker after killing its parent; it does not cover a surviving producer.
   - Repository remains unchanged.

6. **Architectural notes**
   - **ARCH-DRY — pass:** shared publication, staging, and inventory-derived ignores.
   - **ARCH-PURE — pass:** pure transformations remain separate from filesystem/process integration.
   - **ARCH-PURPOSE — flag:** interrupted-producer recovery remains incomplete under BR-8.
   - **ARCH-MOCK — pass:** injected process doubles and real Make/Git fixtures exercise shared boundaries.
   - **ARCH-CONSTRAINTS — pass:** sequential startup introduces no unbounded fan-out.
   - **ARCH-SECURE — pass:** inspected inventory validation and parent-path checks reject malformed or escaping state.
   - **ARCH-ORDER — flag:** compile can return while a producer remains active.
   - **ARCH-FUNERAL — flag:** late writes recreate stages without recoverable ownership.

7. **Plan revision recommendation**

   Append a `## Revisions` entry specifying producer-tree lifetime, cancellation/death recovery, and the invariant that stage cleanup requires evidence that all writers have stopped. Add deterministic late-writer regressions to M2’s recovery contract.
