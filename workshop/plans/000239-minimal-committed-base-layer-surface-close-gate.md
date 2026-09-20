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

## Open findings

- **BR-1** [Critical] `source-resolution-provenance` Local links resolve relative origins against the wrong directory
- **BR-5** [Important] `durable-staging-reclamation` Process death leaves clone staging directories without cleanup
