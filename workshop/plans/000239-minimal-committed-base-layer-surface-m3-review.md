# Boundary Review — ariadne#239 (milestone M3)

| field | value |
|-------|-------|
| issue | 239 — Minimal committed base-layer surface |
| repo | ariadne |
| issue file | workshop/issues/000239-minimal-committed-base-layer-surface.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 0af48033a0a25a4d2b60d4f77c86c63905f60843..fc8b84ebd6bcb7e39fdcaf236295482472820f2d |
| command | sdlc milestone-close --issue 239 --milestone M3 |
| reviewer | codex |
| timestamp | 2026-09-20T15:29:47-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The ownership extraction preserves the compiler’s existing checks, and migration correctly intersects ownership proof with Git’s ignored set. Release packaging and affected tests pass. Two Important gaps remain: interrupted release staging has no reclamation path, and the new Git index mutation sequence lacks failure-injection coverage.

1. **Strengths**

   - `pkg/weaveownership` centralizes schema validation, identity matching, and parent-symlink protection without introducing a second ownership model.
   - Migration tests cover authored replacements, changed permissions, local ignore negations, literal pathspecs, and unrelated force-tracked files.
   - Release tests build all four platform archives and verify targets, CGO settings, executable permissions, checksums, version output, and the formula’s composition fixture.
   - README and atlas document candidate preparation, migration, and the explicit #241 publication boundary.

2. **Critical findings**

   None.

3. **Important findings**

   - [scripts/release-weave.sh:21](/Users/xianxu/workspace/ariadne/scripts/release-weave.sh:21): interrupted staging is never reclaimed. I terminated the release process group with SIGTERM while its build was waiting; `.weave-release-*` remained after termination and another invocation. `TemporaryDirectory` handles ordinary exceptions, but this path has no durable ownership or retry recovery. Define reclamation across all staging producers, including release builds, and test cancellation/death followed by retry. **ARCH-FUNERAL; ARCH-ORDER.**
   - [cmd/sdlc/propagatebase.go:285](/Users/xianxu/workspace/ariadne/cmd/sdlc/propagatebase.go:285): the new per-path index removals have no stateful Git failure tests. Current tests exercise real Git success and pre-mutation inventory rejection; their failure injection targets only weave. They cannot verify partial index progress when a later removal, status, add, or commit fails. Introduce a shared injectable Git boundary with a stateful index fake and test the resulting recovery or explicit operator-remediation contract. Retain real-Git conformance tests. **ARCH-MOCK; ARCH-ORDER.**

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `cmd/weave/...`, `pkg/layergraph/...`, and `pkg/weaveownership` suites.
   - Focused SDLC ownership, propagation, discovery, and ordering tests.
   - `scripts/test/release-weave.test.sh`.
   - Pinned-range `git diff --check`.

   The additional cancellation probe reproduced the staging leak. The full SDLC suite and previously reported consumer pilots were not independently rerun. Repository files remained unchanged.

6. **Architectural notes**

   | Marker | Assessment |
   |---|---|
   | ARCH-DRY | Pass: compiler and migration share ownership validation and proof. |
   | ARCH-PURE | Pass for the extraction: filesystem observation remains behind `Reader`; deterministic predicates remain separate. |
   | ARCH-PURPOSE | Pass: packaging and scoped migration are delivered; public rollout remains explicitly assigned to #241. |
   | ARCH-MOCK | Flag: Git index transitions lack a stateful failure-injection fixture. |
   | ARCH-CONSTRAINTS | Pass: release builds are sequential and limited to four targets; no new unbounded fan-out. |
   | ARCH-SECURE | Pass: validated release tags, argv-based builds, literal Git pathspecs, and inventory validation preserve relevant boundaries. |
   | ARCH-ORDER | Flag: interruption/retry and partial index progress need explicit tested outcomes. |
   | ARCH-FUNERAL | Flag: abandoned release stages have no automatic reclamation mechanism. |

7. **Plan revision recommendations**

   Append `## Revisions` entries specifying:

   - The common staging lifecycle rule, including release interruption, producer termination evidence, and retry reclamation.
   - The migration Git seam, partial-progress outcomes, and failure/retry test matrix.

```findings
findings:
  - id: new
    severity: Important
    family: durable-staging-reclamation
    title: |
      Interrupted release preparation leaves staging directories without reclamation
    detail: |
      scripts/release-weave.sh:21 creates persistent sibling staging using TemporaryDirectory only. A controlled SIGTERM during the build left .weave-release-* behind, and retry preserved it. This is the 4th finding in family durable-staging-reclamation. Earlier rounds fixed instances: state and enforce the rule across clone, generator, publication, and release staging, including producer lifetime and retry reclamation, rather than adding only a signal trap here. Add interruption/retry regressions. ARCH-FUNERAL and ARCH-ORDER.
  - id: new
    severity: Important
    family: external-interaction-test-seam
    title: |
      Scoped migration lacks stateful Git index failure coverage
    detail: |
      cmd/sdlc/propagatebase.go:285 adds sequential index removals through direct exec calls, but propagation tests inject only weave failures and do not exercise partial Git index progress. This is the 3rd finding in family external-interaction-test-seam. State the shared-boundary rule and enumerate migration's status, ls-files, rm, add, and commit interactions; exercise them through a stateful Git fake with failures after earlier effects succeed, retaining real-Git conformance tests. Specify and test recovery or explicit operator remediation without losing partial-progress evidence. ARCH-MOCK and ARCH-ORDER.
```
