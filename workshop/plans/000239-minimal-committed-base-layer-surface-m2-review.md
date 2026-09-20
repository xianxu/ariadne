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
