---
gate: boundary-review
issue: 239
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-20T13:56:48-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Local links resolve relative origins against the wrong directory
          detail: cmd/weave/link.go:38–47 passes the checkout-relative origin unchanged to acquire.Ensure, which resolves it against filepath.Dir(dir) at acquire.go:79. A supplemental local-link regression fails with a false source conflict; resolve origins at their owning checkout before validation and recording.
          family: source-resolution-provenance
          round: 1
        - id: BR-2
          severity: Critical
          title: Git origin-read errors silently become local-only links
          detail: cmd/weave/link.go:35–45 ignores every non-cancellation command failure, then records a source-less link. Distinguish confirmed missing origin from command/configuration failures and propagate the latter with regression coverage (ARCH-SECURE).
          family: failed-probe-treated-as-absence
          round: 1
        - id: BR-3
          severity: Important
          title: README omits the new dependency and source declaration surfaces
          detail: cmd/weave/dependencies.go:14 introduces dependencies and --dry-run, while link adds remote addresses and the parser adds strict source-bearing declarations. README.md is unchanged; document delivered behavior and compatibility limits at this boundary.
          family: user-facing-surface-documentation
          round: 1
        - id: BR-4
          severity: Important
          title: Acquisition cannot inject Git outcomes or publication ordering
          detail: acquire.go:14 and link.go:35 directly execute Git without a shared injectable boundary. Add stateful failure and ordering fixtures for production acquisition, retaining local-Git conformance tests (ARCH-MOCK, ARCH-ORDER).
          family: external-interaction-test-seam
          round: 1
        - id: BR-5
          severity: Important
          title: Process death leaves clone staging directories without cleanup
          detail: acquire.go:89–93 creates a fresh staging directory and relies solely on deferred removal. Add ownership-aware reclamation and interruption/retry tests so abandoned clones cannot accumulate indefinitely or cause active clones to be deleted (ARCH-FUNERAL).
          family: durable-staging-reclamation
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-20T14:02:47-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: The original checkout-origin resolution is fixed and covered by TestLinkRelativeOriginCanRestore. However, cmd/weave/link.go:87 compares an existing relative declaration using NormalizeSource rather than resolving it at the declaring root. A valid substrate ../base ../origins/base.git row causes weave link ../base to fail with a false source conflict. Complete the source-resolution-provenance family sweep across origin inspection, declaration comparison, and restoration.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Origin now distinguishes missing configuration values from inspection failures. TestLinkOriginInspectionErrorsPreserveDeps covers malformed configuration and missing Git through production linking; suppressing the error propagation via a scratch Go overlay makes both cases fail.
          round: 2
        - id: BR-3
          disposition: addressed
          note: README.md:15–55 documents address/local links, dependencies, --dry-run, declaration grammar and compatibility limits. These match link.go, dependencies.go, ParseRows, and startup.Dependencies; atlas/workflow/weave.md also describes the delivered surface.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Client.Git provides the shared acquisition/probe boundary. Fixtures inject clone failure and destination creation before publication, while local-Git tests retain conformance coverage. Disabling the publication conflict rejection via a scratch overlay makes TestEnsurePublicationConflictAfterClone fail.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: Ensure reclaims dead-owned stages, but Restore bypasses Ensure for existing destinations at acquire.go:237–238. After publication followed by process death before deferred cleanup, ordinary dependencies retries never reclaim the stage. A scratch CLI fixture with a matching checkout and dead-owned stage exits successfully while retaining that stage.
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-09-20T14:07:42-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Origins and recorded sources resolve at their owning directories. TestLinkRelativeOriginCanRestore passes; removing owner-aware recorded-source comparison in a temporary Go overlay makes it fail.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Cold and warm restoration reclaim dead-owned stages while preserving live, foreign and unrecognized stages. Both reclamation regressions fail when recovery is removed in a temporary overlay.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Origin distinguishes confirmed missing configuration from inspection failure; passing tests cover malformed configuration, missing Git and unchanged declarations on failure.
          round: 3
        - id: BR-3
          disposition: addressed
          note: README documents link, dependencies, strict declaration syntax and incomplete dry-run behavior, matching the delivered command implementations.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Client.Git drives production acquisition and injected failure/publication tests; real local-Git fixtures exercise the same acquisition flow.
          round: 3
      findings:
        - id: BR-6
          severity: Critical
          title: Source identity discards ports and accepts a different repository endpoint
          detail: 'cmd/weave/internal/acquire/source.go:45 uses URL.Hostname(), dropping the port before identity construction at lines 67–69. A supplemental regression through Ensure reused a checkout whose origin was ssh://git@example.com:2222/team/base.git for a declaration naming port 3333. This is the 2nd finding in family source-resolution-provenance. State and enforce the rule across supported source forms: preserve endpoint distinctions unless equivalence is explicitly established; enumerate URI authorities, SCP-style sources and local paths, then test both valid equivalences and required conflicts. ARCH-PURPOSE, ARCH-SECURE.'
          family: source-resolution-provenance
          round: 3
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-09-20T14:12:43-07:00"
      agent: codex
      dispose:
        - id: BR-6
          disposition: addressed
          note: source.go:48–65 preserves nonstandard URI authorities, schemes, paths, queries and SCP identities. Standard GitHub equivalence remains explicit. TestSourceIdentityPreservesEndpointDifferences and TestEnsureRejectsDifferentSourcePorts pass at HEAD and both fail with source.go from 71fa3ef substituted through a temporary Go overlay.
          round: 4
        - id: BR-1
          disposition: addressed
          note: Origins resolve against their owning checkout; recorded relative sources resolve against the declaring repository. TestLinkRelativeOriginCanRestore passes.
          round: 4
        - id: BR-2
          disposition: addressed
          note: Origin distinguishes confirmed missing configuration from inspection failures. Malformed-config and missing-Git regressions pass without modifying deps.
          round: 4
        - id: BR-3
          disposition: addressed
          note: README.md documents source-aware link, dependencies, dry-run, dependency row syntax and Brewfiles, matching the command handlers and shared parser.
          round: 4
        - id: BR-4
          disposition: addressed
          note: Client.Git injects outcomes into production acquisition. Tests exercise publication conflicts and clone failure stopping later acquisition.
          round: 4
        - id: BR-5
          disposition: addressed
          note: Staging recovery removes confirmed dead-owner stages while preserving live, foreign and unrecognized stages. Warm restoration and dry-run preservation regressions pass.
          round: 4
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 5
      timestamp: "2026-09-20T14:28:53-07:00"
      agent: codex
      findings:
        - id: BR-7
          severity: Critical
          title: Generators overwrite authored replacements before ownership checks
          detail: cmd/weave/main.go:522 executes generators directly against final output paths before ApplyManaged at line 540. A scratch compile, authored edit to generated SKILL.md, and second compile returned success while erasing the edit. SnapshotGenerated only observes; GeneratedActions then accepts changed bytes as generated. Enforce ownership before publication for every generator output, including files and links. Add a compile-level regression that actually runs the generator over an edited destination. ARCH-SECURE, ARCH-PURPOSE.
          family: generated-output-write-ownership
          round: 5
        - id: BR-8
          severity: Critical
          title: Failed generation leaves outputs outside durable ownership recovery
          detail: cmd/weave/main.go:522-535 returns on generator or subsequent planning failure before recording generated identities. In a scratch fixture, a generator wrote SKILL.md then exited 1; successful retry treated its unchanged bytes as unowned at ownership.go:393, and removing the generator left the file behind. This is the 2nd finding in family durable-staging-reclamation. State and enforce one recovery rule for every pre-publication writer, including generators, rather than patching one exit path. Preserve durable provenance across failure/process death and test retry followed by retirement. ARCH-ORDER, ARCH-FUNERAL.
          family: durable-staging-reclamation
          round: 5
        - id: BR-9
          severity: Critical
          title: The updated PURE setup entity has no corresponding pure implementation
          detail: workshop/plans/000239-minimal-committed-base-layer-surface-plan.md:113 points its PURE setup-input entity to startup/dependencies.go. That file exposes Dependencies, which performs fs.Stat at line 18 and runner.Run at line 32; it contains no corresponding pure setup-input model. Correct the table to classify discovery/execution as INTEGRATION and append the required revision. Keep genuinely pure PATH composition separately identified. Critical under the supplied core-concepts cross-check rule. ARCH-PURE.
          family: core-concept-purity-classification
          round: 5
        - id: BR-10
          severity: Important
          title: Tools execution requires the concrete process runner
          detail: cmd/weave/internal/startup/tools.go:15 accepts weavefs.ExecRunner and mutates its Stdin, unlike Dependencies' injectable Runner interface. Tests therefore require real Make and cannot substitute the required stateful external double. This is the 2nd finding in family external-interaction-test-seam. Apply the shared rule that production and fake execution consume the same invocation boundary, including stdin; retain real-Make conformance tests alongside the fake. ARCH-MOCK.
          family: external-interaction-test-seam
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 6
      timestamp: "2026-09-20T14:43:06-07:00"
      agent: codex
      dispose:
        - id: BR-7
          disposition: addressed
          note: Generators now stage before ownership-checked publication. Reverting staging in scratch makes real-generator authored file/link replacement regressions fail.
          round: 6
        - id: BR-8
          disposition: addressed
          note: The reported prepublication failure is covered by shared durable stages and failure/death/retry/retirement tests. Reverting staging makes the failed-generator publication regression fail; partial final publication is a separate finding below.
          round: 6
        - id: BR-9
          disposition: addressed
          note: The active table now lists only ToolEnvironment as pure, matching its implementation; the appended M2 review revision classifies discovery and execution as integration.
          round: 6
        - id: BR-10
          disposition: addressed
          note: Tools accepts weavefs.InputRunner and passes stdin through RunInput. The injected stateful failure/retry test and retained real-Make tests pass.
          round: 6
      findings:
        - id: BR-11
          severity: Critical
          title: Partial final-file writes escape ownership recovery
          detail: 'apply.go:263 uses truncating WriteFile, but ownership.go:255 recognizes only complete old/new hashes. Scratch injection writing half the output before returning an error makes retry refuse it as an authored replacement; retirement cannot identify it either. This is the 3rd finding in this family: enforce atomic or durably recoverable publication across all writers, with partial-write/retry/retirement regressions. ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE.'
          family: durable-staging-reclamation
          round: 6
        - id: BR-12
          severity: Important
          title: Staged publication discards executable permissions
          detail: staged.go:125 lowers regular outputs to content-only WriteFile actions, and weavefs/fs.go:72 creates them as 0644. A scratch regression publishing a generated 0755 run.sh fails because the destination is 0644. Preserve permissions and test cold publication and warm permission changes. ARCH-PURPOSE.
          family: generated-artifact-fidelity
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 7
      timestamp: "2026-09-20T14:58:01-07:00"
      agent: codex
      dispose:
        - id: BR-11
          disposition: addressed
          note: Shared atomic publication protects final files. Removing it in a scratch copy makes TestManagedPartialFileWritePreservesIdentityForRetryAndRetirement fail on both cold and warm partial writes.
          round: 7
        - id: BR-12
          disposition: addressed
          note: Staged modes propagate through publication and ownership identities. Removing mode propagation makes TestStagedFileModesAreOwnedAndPreserved fail with 0644 instead of 0755; production compile also tests executable permissions.
          round: 7
        - id: BR-7
          disposition: addressed
          note: Generators write into staging before managed ownership checks; compile regressions preserve edited files, symlink replacements, and cold authored destinations.
          round: 7
        - id: BR-8
          disposition: not-addressed
          note: Cancellation terminates only the immediate marker process. A scratch production-path regression confirms its child can recreate the removed generation stage without owner.json; a subsequent compile with the marker removed leaves that residue intact. See main.go:524 and internal/weavefs/runner.go:42. ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE; existing family durable-staging-reclamation.
          round: 7
        - id: BR-9
          disposition: addressed
          note: The revised active concept table identifies ToolEnvironment as pure and discovery/execution as integration, consistent with the implementation.
          round: 7
        - id: BR-10
          disposition: addressed
          note: Tools accepts InputRunner; production Make and the stateful failure/retry test double share that interface, with real Make conformance tests.
          round: 7
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 8
      timestamp: "2026-09-20T15:13:54-07:00"
      agent: codex
      dispose:
        - id: BR-8
          disposition: addressed
          note: Staging ownership precedes generator execution; inherited leases prevent cleanup while descendants can write. Passing production-path tests cover failed generation followed by retry/retirement, parent death, cancellation, and late writers. The assertions directly reject the former premature-cleanup behavior.
          round: 8
        - id: BR-7
          disposition: addressed
          note: Generators publish through StagedActions and ApplyManaged ownership checks. Tests cover edited files, replaced links, and cold authored destinations.
          round: 8
        - id: BR-9
          disposition: addressed
          note: The current plan names ToolEnvironment as the pure setup entity and classifies discovery, execution, and sequential setup as integrations; those implementations exist at the documented paths.
          round: 8
        - id: BR-10
          disposition: addressed
          note: Tools consumes InputRunner. Production Make and the stateful failure/retry double share that interface; real Make fixtures also pass.
          round: 8
        - id: BR-11
          disposition: addressed
          note: Publish writes into an owned stage before atomic rename. Partial-write and killed-writer tests verify preservation, recovery, and retirement.
          round: 8
        - id: BR-12
          disposition: addressed
          note: Staged permissions travel through WriteFile and ownership identities. Publication applies permissions before rename; executable-permission and legacy-inventory tests pass.
          round: 8
      boundary: M2
      recipe: milestone-review
      blocked: false
    - "n": 9
      timestamp: "2026-09-20T15:29:47-07:00"
      agent: codex
      findings:
        - id: BR-13
          severity: Important
          title: Interrupted release preparation leaves staging directories without reclamation
          detail: 'scripts/release-weave.sh:21 creates persistent sibling staging using TemporaryDirectory only. A controlled SIGTERM during the build left .weave-release-* behind, and retry preserved it. This is the 4th finding in family durable-staging-reclamation. Earlier rounds fixed instances: state and enforce the rule across clone, generator, publication, and release staging, including producer lifetime and retry reclamation, rather than adding only a signal trap here. Add interruption/retry regressions. ARCH-FUNERAL and ARCH-ORDER.'
          family: durable-staging-reclamation
          round: 9
        - id: BR-14
          severity: Important
          title: Scoped migration lacks stateful Git index failure coverage
          detail: cmd/sdlc/propagatebase.go:285 adds sequential index removals through direct exec calls, but propagation tests inject only weave failures and do not exercise partial Git index progress. This is the 3rd finding in family external-interaction-test-seam. State the shared-boundary rule and enumerate migration's status, ls-files, rm, add, and commit interactions; exercise them through a stateful Git fake with failures after earlier effects succeed, retaining real-Git conformance tests. Specify and test recovery or explicit operator remediation without losing partial-progress evidence. ARCH-MOCK and ARCH-ORDER.
          family: external-interaction-test-seam
          round: 9
      boundary: M3
      recipe: milestone-review
      blocked: true
    - "n": 10
      timestamp: "2026-09-20T15:42:34-07:00"
      agent: codex
      dispose:
        - id: BR-13
          disposition: addressed
          note: cmd/weave/internal/release/main.go:65 reclaims before existing-output refusal and uses shared staging ownership and RunOwned producer leases. Cancellation, owner-death/live-producer, and public-launcher tests pass. Removing reclamation in a scratch copy fails the recovery regression with “dead producer stage not reclaimed.”
          round: 10
        - id: BR-14
          disposition: addressed
          note: cmd/sdlc/propagatebase_failure_test.go exercises status, ls-files, rm, add, and commit through the production PATH boundary with a persistent real-index-backed faulting executable. Tests verify partial effects, commit-success-before-error, retained files, dirty public retry, and operator-resolved recovery. Removing the remediation wrapper fails all 11 failure-matrix cases.
          round: 10
      boundary: M3
      recipe: milestone-review
      blocked: false
    - "n": 11
      timestamp: "2026-09-20T15:48:36-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Relative-origin restoration and recorded-source comparisons resolve paths at their owning checkout; acquisition/link regressions cover these paths.
          round: 11
        - id: BR-2
          disposition: addressed
          note: Origin inspection distinguishes missing configuration from Git failures; malformed-config and dependency-preservation regressions exercise the distinction.
          round: 11
        - id: BR-3
          disposition: addressed
          note: README.md now documents link sources, construct/deps syntax, dependencies, and dry-run behavior, consistent with the inspected command implementations.
          round: 11
        - id: BR-4
          disposition: addressed
          note: The injected Git boundary supports clone failure and competing-publication tests through the production acquisition path.
          round: 11
        - id: BR-5
          disposition: addressed
          note: Acquisition tests cover dead-stage reclamation, live/foreign preservation, warm restoration after publication, and nonmutating dry-run.
          round: 11
        - id: BR-6
          disposition: addressed
          note: Source identity tests preserve endpoint distinctions; TestEnsureRejectsDifferentSourcePorts exercises actual checkout rejection.
          round: 11
        - id: BR-7
          disposition: addressed
          note: Compile regressions preserve edited and cold authored generator destinations through staged generation and managed publication.
          round: 11
        - id: BR-8
          disposition: addressed
          note: Failed-generation, process-death, cancellation, and live-descendant regressions exercise durable recovery and producer leases.
          round: 11
        - id: BR-9
          disposition: addressed
          note: The active plan identifies ToolEnvironment as pure and discovery/execution as integration; tools.go implements that separation.
          round: 11
        - id: BR-10
          disposition: addressed
          note: Tools accepts InputRunner; TestToolsInjectedProcessStateStopsOnFailure exercises failure and retry through that production boundary.
          round: 11
        - id: BR-11
          disposition: addressed
          note: Partial-write and killed-publication regressions exercise staged atomic publication, retry, and retirement.
          round: 11
        - id: BR-12
          disposition: addressed
          note: Staged-file and compile-level executable-permission regressions exercise permission preservation and ownership.
          round: 11
        - id: BR-13
          disposition: addressed
          note: Release preparation reuses owned staging; cancellation, killed-owner/live-producer, retry, and public launcher tests pass.
          round: 11
        - id: BR-14
          disposition: addressed
          note: Migration tests inject Git failures before and after persistent effects, verify retained index state, and exercise operator-resolved retry.
          round: 11
      findings:
        - id: BR-15
          severity: Critical
          title: Missing optional tools target permits unintended implicit Make builds
          detail: 'cmd/weave/internal/startup/tools.go:31 supplies tools: without suppressing implicit rules. Public compile reproduction creates an unmanaged executable from tools.sh, or invokes cc on tools.c and fails despite no declared tools target. Make the command target explicitly phony while preserving authored recipes/prerequisites, and add real-Make regressions for implicit candidates and existing target-named files. ARCH-PURPOSE and ARCH-FUNERAL.'
          family: optional-target-noop
          round: 11
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#239 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-20T13:56:48-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `source-resolution-provenance` Local links resolve relative origins against the wrong directory
  cmd/weave/link.go:38–47 passes the checkout-relative origin unchanged to acquire.Ensure, which resolves it against filepath.Dir(dir) at acquire.go:79. A supplemental local-link regression fails with a false source conflict; resolve origins at their owning checkout before validation and recording.
- **BR-2** [Critical] `failed-probe-treated-as-absence` Git origin-read errors silently become local-only links
  cmd/weave/link.go:35–45 ignores every non-cancellation command failure, then records a source-less link. Distinguish confirmed missing origin from command/configuration failures and propagate the latter with regression coverage (ARCH-SECURE).
- **BR-3** [Important] `user-facing-surface-documentation` README omits the new dependency and source declaration surfaces
  cmd/weave/dependencies.go:14 introduces dependencies and --dry-run, while link adds remote addresses and the parser adds strict source-bearing declarations. README.md is unchanged; document delivered behavior and compatibility limits at this boundary.
- **BR-4** [Important] `external-interaction-test-seam` Acquisition cannot inject Git outcomes or publication ordering
  acquire.go:14 and link.go:35 directly execute Git without a shared injectable boundary. Add stateful failure and ordering fixtures for production acquisition, retaining local-Git conformance tests (ARCH-MOCK, ARCH-ORDER).
- **BR-5** [Important] `durable-staging-reclamation` Process death leaves clone staging directories without cleanup
  acquire.go:89–93 creates a fresh staging directory and relies solely on deferred removal. Add ownership-aware reclamation and interruption/retry tests so abandoned clones cannot accumulate indefinitely or cause active clones to be deleted (ARCH-FUNERAL).

## Round 2 — 2026-09-20T14:02:47-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — The original checkout-origin resolution is fixed and covered by TestLinkRelativeOriginCanRestore. However, cmd/weave/link.go:87 compares an existing relative declaration using NormalizeSource rather than resolving it at the declaring root. A valid substrate ../base ../origins/base.git row causes weave link ../base to fail with a false source conflict. Complete the source-resolution-provenance family sweep across origin inspection, declaration comparison, and restoration.
- BR-2 — addressed — Origin now distinguishes missing configuration values from inspection failures. TestLinkOriginInspectionErrorsPreserveDeps covers malformed configuration and missing Git through production linking; suppressing the error propagation via a scratch Go overlay makes both cases fail.
- BR-3 — addressed — README.md:15–55 documents address/local links, dependencies, --dry-run, declaration grammar and compatibility limits. These match link.go, dependencies.go, ParseRows, and startup.Dependencies; atlas/workflow/weave.md also describes the delivered surface.
- BR-4 — addressed — Client.Git provides the shared acquisition/probe boundary. Fixtures inject clone failure and destination creation before publication, while local-Git tests retain conformance coverage. Disabling the publication conflict rejection via a scratch overlay makes TestEnsurePublicationConflictAfterClone fail.
- BR-5 — not-addressed — Ensure reclaims dead-owned stages, but Restore bypasses Ensure for existing destinations at acquire.go:237–238. After publication followed by process death before deferred cleanup, ordinary dependencies retries never reclaim the stage. A scratch CLI fixture with a matching checkout and dead-owned stage exits successfully while retaining that stage.

## Round 3 — 2026-09-20T14:07:42-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Origins and recorded sources resolve at their owning directories. TestLinkRelativeOriginCanRestore passes; removing owner-aware recorded-source comparison in a temporary Go overlay makes it fail.
- BR-5 — addressed — Cold and warm restoration reclaim dead-owned stages while preserving live, foreign and unrecognized stages. Both reclamation regressions fail when recovery is removed in a temporary overlay.
- BR-2 — addressed — Origin distinguishes confirmed missing configuration from inspection failure; passing tests cover malformed configuration, missing Git and unchanged declarations on failure.
- BR-3 — addressed — README documents link, dependencies, strict declaration syntax and incomplete dry-run behavior, matching the delivered command implementations.
- BR-4 — addressed — Client.Git drives production acquisition and injected failure/publication tests; real local-Git fixtures exercise the same acquisition flow.

### Raised

- **BR-6** [Critical] `source-resolution-provenance` Source identity discards ports and accepts a different repository endpoint
  cmd/weave/internal/acquire/source.go:45 uses URL.Hostname(), dropping the port before identity construction at lines 67–69. A supplemental regression through Ensure reused a checkout whose origin was ssh://git@example.com:2222/team/base.git for a declaration naming port 3333. This is the 2nd finding in family source-resolution-provenance. State and enforce the rule across supported source forms: preserve endpoint distinctions unless equivalence is explicitly established; enumerate URI authorities, SCP-style sources and local paths, then test both valid equivalences and required conflicts. ARCH-PURPOSE, ARCH-SECURE.

## Round 4 — 2026-09-20T14:12:43-07:00 (codex) — passed

### Disposed

- BR-6 — addressed — source.go:48–65 preserves nonstandard URI authorities, schemes, paths, queries and SCP identities. Standard GitHub equivalence remains explicit. TestSourceIdentityPreservesEndpointDifferences and TestEnsureRejectsDifferentSourcePorts pass at HEAD and both fail with source.go from 71fa3ef substituted through a temporary Go overlay.
- BR-1 — addressed — Origins resolve against their owning checkout; recorded relative sources resolve against the declaring repository. TestLinkRelativeOriginCanRestore passes.
- BR-2 — addressed — Origin distinguishes confirmed missing configuration from inspection failures. Malformed-config and missing-Git regressions pass without modifying deps.
- BR-3 — addressed — README.md documents source-aware link, dependencies, dry-run, dependency row syntax and Brewfiles, matching the command handlers and shared parser.
- BR-4 — addressed — Client.Git injects outcomes into production acquisition. Tests exercise publication conflicts and clone failure stopping later acquisition.
- BR-5 — addressed — Staging recovery removes confirmed dead-owner stages while preserving live, foreign and unrecognized stages. Warm restoration and dry-run preservation regressions pass.

## Round 5 — 2026-09-20T14:28:53-07:00 (codex) — BLOCKED

### Raised

- **BR-7** [Critical] `generated-output-write-ownership` Generators overwrite authored replacements before ownership checks
  cmd/weave/main.go:522 executes generators directly against final output paths before ApplyManaged at line 540. A scratch compile, authored edit to generated SKILL.md, and second compile returned success while erasing the edit. SnapshotGenerated only observes; GeneratedActions then accepts changed bytes as generated. Enforce ownership before publication for every generator output, including files and links. Add a compile-level regression that actually runs the generator over an edited destination. ARCH-SECURE, ARCH-PURPOSE.
- **BR-8** [Critical] `durable-staging-reclamation` Failed generation leaves outputs outside durable ownership recovery
  cmd/weave/main.go:522-535 returns on generator or subsequent planning failure before recording generated identities. In a scratch fixture, a generator wrote SKILL.md then exited 1; successful retry treated its unchanged bytes as unowned at ownership.go:393, and removing the generator left the file behind. This is the 2nd finding in family durable-staging-reclamation. State and enforce one recovery rule for every pre-publication writer, including generators, rather than patching one exit path. Preserve durable provenance across failure/process death and test retry followed by retirement. ARCH-ORDER, ARCH-FUNERAL.
- **BR-9** [Critical] `core-concept-purity-classification` The updated PURE setup entity has no corresponding pure implementation
  workshop/plans/000239-minimal-committed-base-layer-surface-plan.md:113 points its PURE setup-input entity to startup/dependencies.go. That file exposes Dependencies, which performs fs.Stat at line 18 and runner.Run at line 32; it contains no corresponding pure setup-input model. Correct the table to classify discovery/execution as INTEGRATION and append the required revision. Keep genuinely pure PATH composition separately identified. Critical under the supplied core-concepts cross-check rule. ARCH-PURE.
- **BR-10** [Important] `external-interaction-test-seam` Tools execution requires the concrete process runner
  cmd/weave/internal/startup/tools.go:15 accepts weavefs.ExecRunner and mutates its Stdin, unlike Dependencies' injectable Runner interface. Tests therefore require real Make and cannot substitute the required stateful external double. This is the 2nd finding in family external-interaction-test-seam. Apply the shared rule that production and fake execution consume the same invocation boundary, including stdin; retain real-Make conformance tests alongside the fake. ARCH-MOCK.

## Round 6 — 2026-09-20T14:43:06-07:00 (codex) — BLOCKED

### Disposed

- BR-7 — addressed — Generators now stage before ownership-checked publication. Reverting staging in scratch makes real-generator authored file/link replacement regressions fail.
- BR-8 — addressed — The reported prepublication failure is covered by shared durable stages and failure/death/retry/retirement tests. Reverting staging makes the failed-generator publication regression fail; partial final publication is a separate finding below.
- BR-9 — addressed — The active table now lists only ToolEnvironment as pure, matching its implementation; the appended M2 review revision classifies discovery and execution as integration.
- BR-10 — addressed — Tools accepts weavefs.InputRunner and passes stdin through RunInput. The injected stateful failure/retry test and retained real-Make tests pass.

### Raised

- **BR-11** [Critical] `durable-staging-reclamation` Partial final-file writes escape ownership recovery
  apply.go:263 uses truncating WriteFile, but ownership.go:255 recognizes only complete old/new hashes. Scratch injection writing half the output before returning an error makes retry refuse it as an authored replacement; retirement cannot identify it either. This is the 3rd finding in this family: enforce atomic or durably recoverable publication across all writers, with partial-write/retry/retirement regressions. ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE.
- **BR-12** [Important] `generated-artifact-fidelity` Staged publication discards executable permissions
  staged.go:125 lowers regular outputs to content-only WriteFile actions, and weavefs/fs.go:72 creates them as 0644. A scratch regression publishing a generated 0755 run.sh fails because the destination is 0644. Preserve permissions and test cold publication and warm permission changes. ARCH-PURPOSE.

## Round 7 — 2026-09-20T14:58:01-07:00 (codex) — BLOCKED

### Disposed

- BR-11 — addressed — Shared atomic publication protects final files. Removing it in a scratch copy makes TestManagedPartialFileWritePreservesIdentityForRetryAndRetirement fail on both cold and warm partial writes.
- BR-12 — addressed — Staged modes propagate through publication and ownership identities. Removing mode propagation makes TestStagedFileModesAreOwnedAndPreserved fail with 0644 instead of 0755; production compile also tests executable permissions.
- BR-7 — addressed — Generators write into staging before managed ownership checks; compile regressions preserve edited files, symlink replacements, and cold authored destinations.
- BR-8 — not-addressed — Cancellation terminates only the immediate marker process. A scratch production-path regression confirms its child can recreate the removed generation stage without owner.json; a subsequent compile with the marker removed leaves that residue intact. See main.go:524 and internal/weavefs/runner.go:42. ARCH-ORDER, ARCH-FUNERAL, ARCH-PURPOSE; existing family durable-staging-reclamation.
- BR-9 — addressed — The revised active concept table identifies ToolEnvironment as pure and discovery/execution as integration, consistent with the implementation.
- BR-10 — addressed — Tools accepts InputRunner; production Make and the stateful failure/retry test double share that interface, with real Make conformance tests.

## Round 8 — 2026-09-20T15:13:54-07:00 (codex) — passed

### Disposed

- BR-8 — addressed — Staging ownership precedes generator execution; inherited leases prevent cleanup while descendants can write. Passing production-path tests cover failed generation followed by retry/retirement, parent death, cancellation, and late writers. The assertions directly reject the former premature-cleanup behavior.
- BR-7 — addressed — Generators publish through StagedActions and ApplyManaged ownership checks. Tests cover edited files, replaced links, and cold authored destinations.
- BR-9 — addressed — The current plan names ToolEnvironment as the pure setup entity and classifies discovery, execution, and sequential setup as integrations; those implementations exist at the documented paths.
- BR-10 — addressed — Tools consumes InputRunner. Production Make and the stateful failure/retry double share that interface; real Make fixtures also pass.
- BR-11 — addressed — Publish writes into an owned stage before atomic rename. Partial-write and killed-writer tests verify preservation, recovery, and retirement.
- BR-12 — addressed — Staged permissions travel through WriteFile and ownership identities. Publication applies permissions before rename; executable-permission and legacy-inventory tests pass.

## Round 9 — 2026-09-20T15:29:47-07:00 (codex) — BLOCKED

### Raised

- **BR-13** [Important] `durable-staging-reclamation` Interrupted release preparation leaves staging directories without reclamation
  scripts/release-weave.sh:21 creates persistent sibling staging using TemporaryDirectory only. A controlled SIGTERM during the build left .weave-release-* behind, and retry preserved it. This is the 4th finding in family durable-staging-reclamation. Earlier rounds fixed instances: state and enforce the rule across clone, generator, publication, and release staging, including producer lifetime and retry reclamation, rather than adding only a signal trap here. Add interruption/retry regressions. ARCH-FUNERAL and ARCH-ORDER.
- **BR-14** [Important] `external-interaction-test-seam` Scoped migration lacks stateful Git index failure coverage
  cmd/sdlc/propagatebase.go:285 adds sequential index removals through direct exec calls, but propagation tests inject only weave failures and do not exercise partial Git index progress. This is the 3rd finding in family external-interaction-test-seam. State the shared-boundary rule and enumerate migration's status, ls-files, rm, add, and commit interactions; exercise them through a stateful Git fake with failures after earlier effects succeed, retaining real-Git conformance tests. Specify and test recovery or explicit operator remediation without losing partial-progress evidence. ARCH-MOCK and ARCH-ORDER.

## Round 10 — 2026-09-20T15:42:34-07:00 (codex) — passed

### Disposed

- BR-13 — addressed — cmd/weave/internal/release/main.go:65 reclaims before existing-output refusal and uses shared staging ownership and RunOwned producer leases. Cancellation, owner-death/live-producer, and public-launcher tests pass. Removing reclamation in a scratch copy fails the recovery regression with “dead producer stage not reclaimed.”
- BR-14 — addressed — cmd/sdlc/propagatebase_failure_test.go exercises status, ls-files, rm, add, and commit through the production PATH boundary with a persistent real-index-backed faulting executable. Tests verify partial effects, commit-success-before-error, retained files, dirty public retry, and operator-resolved recovery. Removing the remediation wrapper fails all 11 failure-matrix cases.

## Round 11 — 2026-09-20T15:48:36-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Relative-origin restoration and recorded-source comparisons resolve paths at their owning checkout; acquisition/link regressions cover these paths.
- BR-2 — addressed — Origin inspection distinguishes missing configuration from Git failures; malformed-config and dependency-preservation regressions exercise the distinction.
- BR-3 — addressed — README.md now documents link sources, construct/deps syntax, dependencies, and dry-run behavior, consistent with the inspected command implementations.
- BR-4 — addressed — The injected Git boundary supports clone failure and competing-publication tests through the production acquisition path.
- BR-5 — addressed — Acquisition tests cover dead-stage reclamation, live/foreign preservation, warm restoration after publication, and nonmutating dry-run.
- BR-6 — addressed — Source identity tests preserve endpoint distinctions; TestEnsureRejectsDifferentSourcePorts exercises actual checkout rejection.
- BR-7 — addressed — Compile regressions preserve edited and cold authored generator destinations through staged generation and managed publication.
- BR-8 — addressed — Failed-generation, process-death, cancellation, and live-descendant regressions exercise durable recovery and producer leases.
- BR-9 — addressed — The active plan identifies ToolEnvironment as pure and discovery/execution as integration; tools.go implements that separation.
- BR-10 — addressed — Tools accepts InputRunner; TestToolsInjectedProcessStateStopsOnFailure exercises failure and retry through that production boundary.
- BR-11 — addressed — Partial-write and killed-publication regressions exercise staged atomic publication, retry, and retirement.
- BR-12 — addressed — Staged-file and compile-level executable-permission regressions exercise permission preservation and ownership.
- BR-13 — addressed — Release preparation reuses owned staging; cancellation, killed-owner/live-producer, retry, and public launcher tests pass.
- BR-14 — addressed — Migration tests inject Git failures before and after persistent effects, verify retained index state, and exercise operator-resolved retry.

### Raised

- **BR-15** [Critical] `optional-target-noop` Missing optional tools target permits unintended implicit Make builds
  cmd/weave/internal/startup/tools.go:31 supplies tools: without suppressing implicit rules. Public compile reproduction creates an unmanaged executable from tools.sh, or invokes cc on tools.c and fails despite no declared tools target. Make the command target explicitly phony while preserving authored recipes/prerequisites, and add real-Make regressions for implicit candidates and existing target-named files. ARCH-PURPOSE and ARCH-FUNERAL.

## Open findings

- **BR-15** [Critical] `optional-target-noop` Missing optional tools target permits unintended implicit Make builds
