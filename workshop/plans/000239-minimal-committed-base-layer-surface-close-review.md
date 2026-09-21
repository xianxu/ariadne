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

---

## Re-review — 2026-09-20T15:53:01-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..ce63180e6528676c380465384f18f1509b9afb3e |
| command | sdlc close --issue 239 |
| reviewer | codex |
| timestamp | 2026-09-20T15:53:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-15’s implicit-build fix is verified by a failing-without-fix regression. One related compatibility defect still blocks shipping: valid authored `tools::` rules fail through `weave compile`. The inspected ownership, publication, and recovery changes passed their tests.

```findings
dispose:
  - id: BR-15
    disposition: addressed
    note: |
      Removing only .PHONY through a temporary Go overlay makes the regression fail for shell/C implicit candidates and an existing tools file suppressing an authored recipe. The pinned implementation passes.
findings:
  - id: new
    severity: Critical
    family: optional-target-noop
    title: |
      Supplemental single-colon target rejects authored double-colon tools rules
    detail: |
      cmd/weave/internal/startup/tools.go:31 appends tools:, which conflicts with valid tools:: declarations. A scratch fixture succeeds with make tools but public weave compile exits before executing either authored recipe. This is the 2nd finding in family optional-target-noop: enforce the rule that optional-command augmentation suppresses implicit builds without imposing a rule flavor on authored targets. Use the phony declaration alone, which passed this reproduction, and sweep absent, single-colon, double-colon, prerequisite, existing-file, and failure cases with real Make regressions. ARCH-PURPOSE.
```

1. **Strengths**
   - BR-15 tests exercise real Make behavior, including unintended output creation.
   - Compilation and migration share ownership validation through `pkg/weaveownership`.
   - Publication and staging tests cover partial writes, permissions, interrupted producers, and recovery.
   - README and atlas document the new startup and generator contracts.

2. **Critical findings**
   - [tools.go:31](/Users/xianxu/workspace/ariadne/cmd/weave/internal/startup/tools.go:31): the injected `tools:` conflicts with `tools::`. Two authored double-colon recipes produced `firstsecond` with ordinary Make; public compile failed with “target file `tools` has both : and :: entries.” Remove the supplemental concrete rule and retain `.PHONY: tools`; add the regression.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Passed: complete weave, layergraph, and weaveownership Go suites; targeted SDLC propagation tests; bootstrap, portable Makefile, and CI shell fixtures.
   - BR-15 mutation failed as expected.
   - Missing coverage: authored double-colon command rules.
   - Native Linux and full release packaging were not rerun in this review. Repository files remained unchanged.

6. **Architecture**
   - **ARCH-DRY — pass:** shared ownership and staging helpers.
   - **ARCH-PURE — pass:** pure parsing, identity, and PATH logic separated from IO.
   - **ARCH-PURPOSE — flag:** valid owner-local Make commands are rejected.
   - **ARCH-MOCK — pass:** stateful process/index fixtures and real-binary checks.
   - **ARCH-CONSTRAINTS — pass:** sequential setup matches the documented operating scope.
   - **ARCH-SECURE — pass:** inventory validation and parent-path checks protect publication.
   - **ARCH-ORDER — pass:** inspected failure paths preserve partial-progress evidence.
   - **ARCH-FUNERAL — pass:** shared producer leases govern staging reclamation.

7. **Plan revision**
   - Append a `## Revisions` entry stating that optional-target augmentation must preserve authored rule flavor; specify phony-only augmentation and the complete real-Make case matrix.

---

## Re-review — 2026-09-20T15:56:45-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..7ba85cb2e2b88c1b826386f49f6d866757972fe2 |
| command | sdlc close --issue 239 |
| reviewer | codex |
| timestamp | 2026-09-20T15:56:45-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

BR-16 is addressed with meaningful regression evidence: production adds only `.PHONY: tools`, preserving authored rule forms. Restoring the old single-colon rule through a temporary Go overlay makes the double-colon recipe, prerequisite, and failure tests fail. The pinned implementation passes the affected suites. No new blocking findings emerged.

```findings
dispose:
  - id: BR-16
    disposition: addressed
    note: |
      cmd/weave/internal/startup/tools.go:31 supplies only .PHONY: tools. Real-Make tests at tools_test.go:103 and :148 cover omission, implicit candidates, existing files, single/double-colon recipes and prerequisites, and failures. Restoring tools: through a temporary Go overlay makes the double-colon regressions fail with the expected rule conflict.
```

1. **Strengths**

   - Optional owner commands preserve Make semantics without masking build failures.
   - Compiler retirement and migration share ownership evidence through `pkg/weaveownership`, protecting authored replacements and unrelated tracked files.
   - Publication and staging tests exercise partial writes, process death, surviving producers, and retry.
   - README and atlas document the new startup, generator, release, and migration contracts, including publication deferred to #241.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None raised.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/weave/... ./pkg/layergraph/... ./pkg/weaveownership/... -count=1`
   - Focused SDLC migration and propagation tests.
   - Bootstrap, portable Makefile, and portable CI shell fixtures.
   - BR-16 mutation check: the previous implementation fails the new regressions.

   The prescribed range’s whitespace check passes. Including the issue file reports one trailing blank line at EOF. Native Linux and full release packaging were not rerun in this review. Repository files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** shared dependency parsing, ownership matching, and staging machinery.
   - **ARCH-PURE — Pass:** deterministic parsing, identity, ignore, and PATH transformations remain separate from filesystem/process integration.
   - **ARCH-PURPOSE — Pass:** startup ordering and derived-output ownership fulfill the active contract; BR-16 covers the requested Make rule matrix.
   - **ARCH-MOCK — Pass:** process seams support stateful fixtures, with real Git and Make conformance tests.
   - **ARCH-CONSTRAINTS — Pass:** sequential layer execution avoids new concurrency growth; producer cleanup has bounded waits.
   - **ARCH-SECURE — Pass:** source, mount, inventory, and publication validation protect the inspected boundaries.
   - **ARCH-ORDER — Pass:** tests cover failure ordering, retained partial Git effects, and controlled producer lifetime.
   - **ARCH-FUNERAL — Pass:** owned stages have normal cleanup and interrupted-run reclamation; unchanged retired outputs are removed through identity proof.

7. **Plan revision recommendations:** None required. The existing BR-16 revision accurately describes the implemented correction.

---

## Re-review — 2026-09-20T16:02:38-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..d1cbd6e04840f7e25947d270be094bc09b648464 |
| command | sdlc close --issue 239 |
| reviewer | codex |
| timestamp | 2026-09-20T16:02:38-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation substantially delivers the approved contract, and the reviewed suites pass. One new Important finding blocks approval: malformed repository URLs can expose credentials in CLI diagnostics. No previously disposed finding needs reopening. Repository files remain unchanged.

1. **Strengths**
   - Compiler retirement and Git migration share the same ownership proof.
   - Atomic publication preserves prior content and executable permissions.
   - Producer leases and interruption tests cover surviving children and safe reclamation.
   - README and atlas document startup, generator contracts, and the separate #241 rollout.

2. **Critical findings:** None.

3. **Important findings**

   **Credential-bearing parse errors reach stderr** — [source.go:37](/Users/xianxu/workspace/ariadne/cmd/weave/internal/acquire/source.go:37), **ARCH-SECURE**.

   Reproduced without mutation:
   ```sh
   go run ./cmd/weave link 'https://user:review-secret@example.com:bad/repo.git'
   ```
   The fictitious password appears twice: directly through `%q`, and inside the wrapped `url.Parse` error. Parsing fails before credential rejection executes.

   Return sanitized diagnostics without raw input or the nested URL-bearing error. Add regressions covering malformed ports, hosts, and escapes, asserting credentials never appear in returned errors or CLI output.

4. **Minor findings:** None.

5. **Test coverage**

   Passed:
   - Full `cmd/weave/...`, `pkg/layergraph/...`, and `pkg/weaveownership/...` suites.
   - Focused SDLC propagation and migration tests.
   - Bootstrap, portable Makefile, and CI run-block fixtures.
   - Pinned-range `git diff --check`.

   Existing credential tests assert rejection, but do not check diagnostic confidentiality. The corrected Homebrew action resolves at its [upstream source](https://raw.githubusercontent.com/Homebrew/actions/main/setup-homebrew/action.yml); hosted CI and four-platform release packaging were not rerun in this review.

6. **Architecture**
   - **ARCH-DRY — pass:** shared dependency parsing, ownership proof, and staging utilities.
   - **ARCH-PURE — pass:** pure transformations remain separate from filesystem/process integration.
   - **ARCH-PURPOSE — pass:** startup, ownership, and scoped migration fulfill the active contract.
   - **ARCH-MOCK — pass:** stateful process fixtures and real-Git/Make conformance exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** serial setup avoids added concurrency or fan-out.
   - **ARCH-SECURE — flag:** credential disclosure described above.
   - **ARCH-ORDER — pass:** tests exercise publication conflicts, partial effects, cancellation, and surviving producers.
   - **ARCH-FUNERAL — pass:** owned stages have cleanup/reclamation paths; inventories support retirement.

7. **Plan revision recommendation**

   Append a `## Revisions` entry specifying credential-safe diagnostics for all source-validation failures, including failures before successful URL parsing, and the regression matrix above.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Owner-relative source resolution and TestLinkRelativeOriginCanRestore cover restoration.
  - id: BR-2
    disposition: addressed
    note: |
      Origin distinguishes confirmed absence from failed inspection; preservation regressions pass.
  - id: BR-3
    disposition: addressed
    note: |
      README documents link, dependencies, source declarations, and dry-run limitations consistently with the implementation.
  - id: BR-4
    disposition: addressed
    note: |
      GitRunner supports injected clone failures and publication conflicts through the production acquisition path.
  - id: BR-5
    disposition: addressed
    note: |
      Dead-stage reclamation tests cover both missing destinations and already-published checkouts.
  - id: BR-6
    disposition: addressed
    note: |
      Endpoint identity tests preserve ports and transport distinctions; wrong-port checkout rejection passes.
  - id: BR-7
    disposition: addressed
    note: |
      Staged generator publication passes through ownership checks that preserve authored replacements.
  - id: BR-8
    disposition: addressed
    note: |
      Durable generation stages and producer leases have cancellation, parent-death, late-writer, and retry coverage.
  - id: BR-9
    disposition: addressed
    note: |
      The active concept table identifies ToolEnvironment as pure and discovery/execution as integration.
  - id: BR-10
    disposition: addressed
    note: |
      Tools accepts InputRunner; injected process-state and real-Make tests exercise the shared boundary.
  - id: BR-11
    disposition: addressed
    note: |
      Atomic publication has partial-write and killed-writer regressions covering retry and retirement.
  - id: BR-12
    disposition: addressed
    note: |
      Staged publication carries permissions; executable-mode ownership and preservation tests pass.
  - id: BR-13
    disposition: addressed
    note: |
      Release preparation shares owned staging and leases; cancellation and killed-owner recovery tests pass.
  - id: BR-14
    disposition: addressed
    note: |
      Real-index-backed Git fault injection covers retained effects and operator-resolved retries.
  - id: BR-15
    disposition: addressed
    note: |
      Phony augmentation suppresses implicit builds and prevents target-named files from skipping authored commands.
  - id: BR-16
    disposition: addressed
    note: |
      Augmentation contains no concrete tools rule; double-colon execution and failure regressions pass.
findings:
  - id: new
    severity: Important
    family: credential-safe-diagnostics
    title: |
      Malformed repository URLs expose credentials in CLI errors
    detail: |
      cmd/weave/internal/acquire/source.go:35–37 echoes raw input and wraps a URL-bearing parse error before credential rejection. A link input containing fictitious userinfo and an invalid port prints its password twice. ARCH-SECURE: sanitize all source-validation diagnostics and add malformed-port, host, and escape regressions asserting credentials are absent from errors and CLI output.
```

---

## Re-review — 2026-09-20T16:06:48-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..68482fd8319db2787a0ad6cba215622d598d77a2 |
| command | sdlc close --issue 239 |
| reviewer | codex |
| timestamp | 2026-09-20T16:06:48-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-17 is addressed, and all executed tests pass. One active plan classification contradicts the implementation; the requested core-concepts gate explicitly makes this blocking. No new runtime defect was established.

```findings
dispose:
  - id: BR-17
    disposition: addressed
    note: |
      Source-validation errors omit raw input and nested URL errors. Unit and CLI regressions pass at HEAD and fail with HEAD^ source.go substituted through a temporary Go overlay.
findings:
  - id: new
    severity: Critical
    family: core-concept-purity-classification
    title: |
      Active concept table incorrectly classifies filesystem identity matching as PURE
    detail: |
      workshop/plans/000239-minimal-committed-base-layer-surface-plan.md:249–253 classifies generated identity matching as pure and promises pure tests. However, plan/ownership.go:34 delegates to pkg/weaveownership/ownership.go:79 Matches, which calls Lstat, Readlink and ReadFile; ownership_test.go:34 exercises matching using temporary files and OSReader. ARCH-PURE: classify observation-based matching as INTEGRATION and distinguish pure digest/identity construction. This is the second finding in family core-concept-purity-classification, distinct from BR-9's corrected setup entity. State and apply the rule across every active concept row: filesystem observation is integration even behind an injected reader. Append a Revisions entry correcting classifications, locations and test descriptions.
```

1. **Strengths**
   - Credential regression evidence is independently confirmed red/green.
   - Publication and retirement share ownership proof; tests cover partial writes, authored replacements and permissions.
   - Startup, release packaging and scoped migration have meaningful failure coverage.
   - README and atlas describe the new commands, generator contract and migration boundary.

2. **Critical findings:** The concept-table mismatch above. Correcting the plan is sufficient; no wording-presence test is needed.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Passed weave, layergraph and ownership suites; targeted SDLC propagation/migration tests; bootstrap, Make and CI fixtures; four-platform release packaging and native formula composition. Range whitespace checks pass; repository remains unchanged.

6. **Architecture**
   - **ARCH-DRY — pass:** shared ownership and staging implementations.
   - **ARCH-PURE — flag:** matching classification contradicts filesystem observation.
   - **ARCH-PURPOSE — pass:** derived outputs and scoped migration serve the minimal committed-surface contract.
   - **ARCH-MOCK — pass:** injected failures and real-tool conformance exercise production boundaries.
   - **ARCH-CONSTRAINTS — pass:** serial setup avoids added concurrent fan-out.
   - **ARCH-SECURE — pass:** BR-17 verified; persisted ownership validation rejects malformed state.
   - **ARCH-ORDER — pass:** partial-progress and producer-lifetime tests cover interrupted operations.
   - **ARCH-FUNERAL — pass:** shared stage reclamation and identity-based retirement provide cleanup paths.

7. **Plan revision:** Append a classification correction covering all active concept rows. Separate pure identity construction/digest from filesystem matching, name `pkg/weaveownership` as the shared implementation, and retain the existing integration tests as evidence.
