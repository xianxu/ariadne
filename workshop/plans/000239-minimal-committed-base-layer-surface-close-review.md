# Boundary Review — ariadne#239 (whole-issue close)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..3c6852d554a6a13c334655c9c0f7a6a33cc61d78 |
| command | sdlc close --issue 239 |
| reviewer | codex |
| timestamp | 2026-09-20T15:48:36-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation substantially delivers the approved contract, and the tested acquisition, ownership, recovery, packaging, and migration paths pass. One reproduced correctness bug blocks shipping: a layer without a `tools` target can trigger Make’s implicit build rules, creating unintended files or failing compilation.

1. **Strengths**

   - Acquisition preserves source identity and distinguishes failed Git probes from absent origins.
   - Generated-output publication checks ownership, preserves permissions, and recovers interrupted writes.
   - Migration uses shared ownership evidence and tests Git failures before and after effects.
   - README and atlas document the new startup and generator contracts.

2. **Critical findings**

   **Absent optional `tools` target triggers unintended builds** — [tools.go:31](/Users/xianxu/workspace/ariadne/cmd/weave/internal/startup/tools.go:31).

   The supplemental `tools:\n` declaration has no recipe and permits implicit-rule search. Reproduced through the pinned public `weave compile` binary in disposable layers with no declared `tools` target:

   - With `tools.sh`, compilation executes `cat tools.sh >tools` and creates an executable outside managed ownership.
   - With `tools.c`, compilation invokes `cc`; invalid C makes the entire compile fail.

   Make the optional command target explicitly phony, preserving authored recipes and prerequisites. Add real-Make regressions covering implicit-rule candidates and an existing file named `tools`. **ARCH-PURPOSE:** omission must mean no owner build. **ARCH-FUNERAL:** the fallback currently creates unowned residue.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - Full weave, layergraph, and weaveownership suites.
   - Targeted SDLC propagation, migration, and consumption tests.
   - Bootstrap, portable Makefile, and CI shell fixtures.
   - Release tests covering four real archives, metadata, checksums, native execution, formula composition, and failures.

   Existing optional-target tests omit implicit-rule candidates. The public-command reproduction exposes that gap. Repository files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared identity and staging implementations serve multiple consumers.
   - **ARCH-PURE — pass:** pure parsing, matching, and PATH composition remain separable from IO.
   - **ARCH-PURPOSE — flag:** optional owner builds violate their omission contract.
   - **ARCH-MOCK — pass:** stateful process fixtures cover acquisition, builds, and Git index transitions.
   - **ARCH-CONSTRAINTS — pass:** serial setup and bounded producer cleanup match the inspected operating assumptions.
   - **ARCH-SECURE — pass:** inspected source and inventory boundaries validate input and preserve authored replacements.
   - **ARCH-ORDER — pass:** failure and interruption tests cover partial publication and retained external effects.
   - **ARCH-FUNERAL — flag:** implicit Make outputs escape the managed artifact lifecycle; otherwise staging reclamation is covered.

7. **Plan revision recommendations**

   Append a `## Revisions` entry correcting the claim that bare `tools:` supplies a harmless fallback. Specify an explicit command target that suppresses implicit rules, with regression coverage for absent targets, implicit candidates, existing target-named files, and authored build failures.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Relative-origin restoration and recorded-source comparisons resolve paths at their owning checkout; acquisition/link regressions cover these paths.
  - id: BR-2
    disposition: addressed
    note: |
      Origin inspection distinguishes missing configuration from Git failures; malformed-config and dependency-preservation regressions exercise the distinction.
  - id: BR-3
    disposition: addressed
    note: |
      README.md now documents link sources, construct/deps syntax, dependencies, and dry-run behavior, consistent with the inspected command implementations.
  - id: BR-4
    disposition: addressed
    note: |
      The injected Git boundary supports clone failure and competing-publication tests through the production acquisition path.
  - id: BR-5
    disposition: addressed
    note: |
      Acquisition tests cover dead-stage reclamation, live/foreign preservation, warm restoration after publication, and nonmutating dry-run.
  - id: BR-6
    disposition: addressed
    note: |
      Source identity tests preserve endpoint distinctions; TestEnsureRejectsDifferentSourcePorts exercises actual checkout rejection.
  - id: BR-7
    disposition: addressed
    note: |
      Compile regressions preserve edited and cold authored generator destinations through staged generation and managed publication.
  - id: BR-8
    disposition: addressed
    note: |
      Failed-generation, process-death, cancellation, and live-descendant regressions exercise durable recovery and producer leases.
  - id: BR-9
    disposition: addressed
    note: |
      The active plan identifies ToolEnvironment as pure and discovery/execution as integration; tools.go implements that separation.
  - id: BR-10
    disposition: addressed
    note: |
      Tools accepts InputRunner; TestToolsInjectedProcessStateStopsOnFailure exercises failure and retry through that production boundary.
  - id: BR-11
    disposition: addressed
    note: |
      Partial-write and killed-publication regressions exercise staged atomic publication, retry, and retirement.
  - id: BR-12
    disposition: addressed
    note: |
      Staged-file and compile-level executable-permission regressions exercise permission preservation and ownership.
  - id: BR-13
    disposition: addressed
    note: |
      Release preparation reuses owned staging; cancellation, killed-owner/live-producer, retry, and public launcher tests pass.
  - id: BR-14
    disposition: addressed
    note: |
      Migration tests inject Git failures before and after persistent effects, verify retained index state, and exercise operator-resolved retry.
findings:
  - id: new
    severity: Critical
    family: optional-target-noop
    title: |
      Missing optional tools target permits unintended implicit Make builds
    detail: |
      cmd/weave/internal/startup/tools.go:31 supplies tools: without suppressing implicit rules. Public compile reproduction creates an unmanaged executable from tools.sh, or invokes cc on tools.c and fails despite no declared tools target. Make the command target explicitly phony while preserving authored recipes/prerequisites, and add real-Make regressions for implicit candidates and existing target-named files. ARCH-PURPOSE and ARCH-FUNERAL.
```
