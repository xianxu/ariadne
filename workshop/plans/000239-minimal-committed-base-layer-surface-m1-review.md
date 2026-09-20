# Boundary Review — ariadne#239 (milestone M1)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..ab6fcfdd6d4c79b075e2811402ca5130b148d664 |
| command | sdlc milestone-close --issue 239 --milestone M1 |
| reviewer | codex |
| timestamp | 2026-09-20T13:56:48-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M1 delivers transitive restoration, staged clones, and ordered Bundle execution, with useful fixture coverage. A reproduced local-link failure and silent origin-read errors block shipping. Documentation, acquisition test seams, and interrupted-clone cleanup also need attention.

1. **Strengths**

   - `ParseDeps` derives graph edges from `ParseRows`; source metadata does not create a second topology.
   - Acquisition tests exercise diamonds, cycles, dirty/offline checkout reuse, source conflicts, and failed-clone cleanup.
   - Bundle tests use persisted package state and verify ordering, repeat execution, and stopping after failure.
   - Atlas clearly distinguishes delivered M1 behavior from pending compile/bootstrap integration.

2. **Critical findings**

   - **Relative origins resolve against the wrong directory** — `cmd/weave/link.go:38–47`, `cmd/weave/internal/acquire/acquire.go:79`. A local checkout’s `../origins/base.git` is relative to that checkout, but `Ensure` resolves it against the checkout’s parent. The review probe failed with a false origin conflict. Resolve origin paths at their owning checkout before validation and recording; test subsequent restoration too.
   - **Origin-read failures become successful local-only links** — `cmd/weave/link.go:35–45`. Every non-cancellation Git error—including missing Git or malformed configuration—is treated as absent origin. Distinguish confirmed absence from failed inspection and propagate other errors with diagnostics. **ARCH-SECURE**.

3. **Important findings**

   - **README update missing** — `cmd/weave/dependencies.go:14`, `README.md:7`. This range introduces `dependencies --dry-run`, repository-address links, source declarations, and Brewfiles without README changes. Document the delivered commands, platform limits, incomplete-preview behavior, and newly strict declaration grammar.
   - **Acquisition lacks an injectable Git boundary** — `cmd/weave/internal/acquire/acquire.go:14`, `cmd/weave/link.go:35`. Production directly executes Git; real-origin fixtures cannot deterministically inject probe failures or publication races. Introduce a shared injectable boundary and stateful failure/ordering fixtures, retaining real-Git conformance tests. **ARCH-MOCK, ARCH-ORDER**.
   - **Interrupted staging directories have no reclamation path** — `cmd/weave/internal/acquire/acquire.go:89–93`. Deferred removal handles returned errors, but process death leaves `.name-weave-*` directories; retries create additional directories without collecting them. Add ownership-aware recovery that preserves active concurrent clones, with interruption/retry coverage. **ARCH-FUNERAL**.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Committed weave/layergraph tests passed using an overlay for the uncommitted acquisition test file and a temporary build cache. The initial cache location was sandbox-blocked. The supplemental `TestReviewRelativeOriginLocalLink` failed on the pinned implementation. Repository files were unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared row parser and existing graph resolver reused.
   - **ARCH-PURE — pass:** source parsing and topology remain directly testable without IO.
   - **ARCH-PURPOSE — pass for M1 scope:** acquisition/dependencies are delivered; compile integration is explicitly M2.
   - **ARCH-MOCK — flag:** acquisition boundary lacks controllable stateful doubles.
   - **ARCH-CONSTRAINTS — pass:** serial execution introduces no unbounded fan-out; no performance claim was validated.
   - **ARCH-SECURE — flag:** failed origin inspection is interpreted as absence.
   - **ARCH-ORDER — flag:** acquisition races and interrupted publication lack deterministic sequence coverage.
   - **ARCH-FUNERAL — flag:** abandoned staging directories accumulate after process death.

7. **Plan revision recommendations**

   Append a `## Revisions` entry specifying origin-path resolution ownership, Git probe outcomes, the injectable acquisition boundary, and interrupted-stage reclamation. Update concept locations to reflect the delivered `startup/dependencies.go`, distinguishing remaining M2 entities.

```findings
findings:
  - id: new
    severity: Critical
    family: source-resolution-provenance
    title: |
      Local links resolve relative origins against the wrong directory
    detail: |
      cmd/weave/link.go:38–47 passes the checkout-relative origin unchanged to acquire.Ensure, which resolves it against filepath.Dir(dir) at acquire.go:79. A supplemental local-link regression fails with a false source conflict; resolve origins at their owning checkout before validation and recording.
  - id: new
    severity: Critical
    family: failed-probe-treated-as-absence
    title: |
      Git origin-read errors silently become local-only links
    detail: |
      cmd/weave/link.go:35–45 ignores every non-cancellation command failure, then records a source-less link. Distinguish confirmed missing origin from command/configuration failures and propagate the latter with regression coverage (ARCH-SECURE).
  - id: new
    severity: Important
    family: user-facing-surface-documentation
    title: |
      README omits the new dependency and source declaration surfaces
    detail: |
      cmd/weave/dependencies.go:14 introduces dependencies and --dry-run, while link adds remote addresses and the parser adds strict source-bearing declarations. README.md is unchanged; document delivered behavior and compatibility limits at this boundary.
  - id: new
    severity: Important
    family: external-interaction-test-seam
    title: |
      Acquisition cannot inject Git outcomes or publication ordering
    detail: |
      acquire.go:14 and link.go:35 directly execute Git without a shared injectable boundary. Add stateful failure and ordering fixtures for production acquisition, retaining local-Git conformance tests (ARCH-MOCK, ARCH-ORDER).
  - id: new
    severity: Important
    family: durable-staging-reclamation
    title: |
      Process death leaves clone staging directories without cleanup
    detail: |
      acquire.go:89–93 creates a fresh staging directory and relies solely on deferred removal. Add ownership-aware reclamation and interruption/retry tests so abandoned clones cannot accumulate indefinitely or cause active clones to be deleted (ARCH-FUNERAL).
```

---

## Re-review — 2026-09-20T14:02:47-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | ca9ae7f6d71d48e02ae6c23895c05b0d47f35c17..c8b5879bbed2a0ab307701303937597e8ddf9d0b |
| command | sdlc milestone-close --issue 239 --milestone M1 |
| reviewer | codex |
| timestamp | 2026-09-20T14:02:47-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The pinned range delivers source acquisition, strict declarations, dependency installation, and appropriate README/atlas updates. The full targeted suite passes. However, scratch CLI reproductions show that BR-1 remains incomplete for existing relative-source declarations, and BR-5 misses cleanup when retrying an already-published checkout.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The original checkout-origin resolution is fixed and covered by TestLinkRelativeOriginCanRestore. However, cmd/weave/link.go:87 compares an existing relative declaration using NormalizeSource rather than resolving it at the declaring root. A valid substrate ../base ../origins/base.git row causes weave link ../base to fail with a false source conflict. Complete the source-resolution-provenance family sweep across origin inspection, declaration comparison, and restoration.
  - id: BR-2
    disposition: addressed
    note: |
      Origin now distinguishes missing configuration values from inspection failures. TestLinkOriginInspectionErrorsPreserveDeps covers malformed configuration and missing Git through production linking; suppressing the error propagation via a scratch Go overlay makes both cases fail.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:15–55 documents address/local links, dependencies, --dry-run, declaration grammar and compatibility limits. These match link.go, dependencies.go, ParseRows, and startup.Dependencies; atlas/workflow/weave.md also describes the delivered surface.
  - id: BR-4
    disposition: addressed
    note: |
      Client.Git provides the shared acquisition/probe boundary. Fixtures inject clone failure and destination creation before publication, while local-Git tests retain conformance coverage. Disabling the publication conflict rejection via a scratch overlay makes TestEnsurePublicationConflictAfterClone fail.
  - id: BR-5
    disposition: not-addressed
    note: |
      Ensure reclaims dead-owned stages, but Restore bypasses Ensure for existing destinations at acquire.go:237–238. After publication followed by process death before deferred cleanup, ordinary dependencies retries never reclaim the stage. A scratch CLI fixture with a matching checkout and dead-owned stage exits successfully while retaining that stage.
```

1. **Strengths**
   - `ParseDeps` derives from `ParseRows`, keeping declaration parsing shared.
   - Existing dirty checkouts are preserved; local-Git tests verify reuse without contacting unavailable origins.
   - Git inspection failures propagate visibly, with regression tests confirmed sensitive to the fix.
   - README and atlas accurately distinguish delivered M1 behavior from later startup integration.

2. **Critical findings**
   - **BR-1 — incomplete source comparison**, `cmd/weave/link.go:87`. Resolve the existing declaration against `root` before comparing identities. Add a production-link regression with an existing relative-source row. **ARCH-PURPOSE, ARCH-DRY:** apply the provenance rule to every source comparison, rather than correcting only origin inspection.

3. **Important findings**
   - **BR-5 — warm retries skip reclamation**, `cmd/weave/internal/acquire/acquire.go:237`. Run ownership-aware reclamation during mutating restoration for existing as well as missing destinations; preserve dry-run behavior. Test interruption after publication and retry through `Restore` or the CLI. **ARCH-FUNERAL, ARCH-ORDER:** cleanup must cover every interruption point, including successful publication before deferred removal.

4. **Minor findings**
   - None.

5. **Test coverage notes**
   - Passed: `go test ./cmd/weave/... ./pkg/layergraph/... -count=1`.
   - Scratch CLI reproductions confirmed both remaining defects.
   - Scratch overlay mutations made BR-2 and publication-conflict regressions fail.
   - Repository files remain unchanged. Real Homebrew installation was not exercised.

6. **Architectural notes**
   - **ARCH-DRY — flag:** existing declaration comparison bypasses owner-aware source resolution.
   - **ARCH-PURE — pass:** parsing and normalization remain pure; acquisition is classified as integration.
   - **ARCH-PURPOSE — flag:** the source-provenance correction needs the complete comparison-site sweep.
   - **ARCH-MOCK — pass:** injected Git failure/ordering fixtures supplement real local-Git coverage.
   - **ARCH-CONSTRAINTS — pass:** serial acquisition introduces no concurrent fan-out; concurrent setup is explicitly unsupported.
   - **ARCH-SECURE — pass:** configuration failures remain errors; credential-bearing URL rejection and mount containment have coverage.
   - **ARCH-ORDER — flag:** recovery misses the published-checkout/stale-stage state.
   - **ARCH-FUNERAL — flag:** ordinary warm retries leave provably abandoned stages behind.

7. **Plan revision recommendations**
   - Append a `## Revisions` entry enumerating all source-identity comparison sites and their owning directories.
   - Append the post-publication interruption case and require cleanup verification through normal restoration, alongside existing missing-destination recovery tests.
