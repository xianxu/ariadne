# Boundary Review — ariadne#551 (whole-issue close)

| field | value |
|-------|-------|
| issue | 551 — review box gives the reviewer no dependency checkouts; consumer reviews can't verify substrate claims (echo#9 BR-1 unverifiable) |
| repo | ariadne |
| issue file | workshop/issues/000551-review-box-gives-the-reviewer-no-dependency-checkouts-consumer-reviews-can-t-verify-substrate-claims-echo-9-br-1-unverifiable.md |
| boundary | whole-issue close |
| milestone | — |
| window | cce05bc3f48f61cb63d010060ad2cb79e5ebd1c2..0ae0016dbbe6634bfbf59132045e683bfa76862a |
| command | sdlc close --issue 551 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T06:10:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation delivers pinned disposable dependencies through the production boundary-review path, including command environments, retry reset, cancellation, and cleanup. Targeted tests pass; I found no correctness blocker. README coverage for the new user-facing environment behavior is missing.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-surface-documentation
    title: |
      README update appears missing for boundary-review dependency bindings
    detail: |
      cmd/sdlc/milestoneclose.go:899 and cmd/sdlc/internal/reviewbox/confine.go:50 introduce disposable SDLC_DEP_* bindings for reviewers and preparation commands, but README.md is unchanged. Its boundary-preparation instructions at README.md:146 omit these bindings, while its existing dependency environment documentation at README.md:709 describes local checks. Add the boundary-review contract: last-fetched origin/main pins, no automatic fetch, setup failure recovery, isolation modes, and retry/cleanup lifetime, linking the updated atlas. This is the required README docs gate and an ARCH-PURPOSE completion gap.
```

1. **Strengths**

   - Shared declaration resolution and `layergraph.DependencyEnvironment` avoid parallel naming/conflict rules.
   - Production-dispatch tests verify actual pinned fixture bytes through exported paths, including dirty caller/source separation.
   - Ownership tests cover retry pin retention, partial setup failure, cancellation, cleanup failures, and dead-owner sweeping.
   - The missing-promisor-object regression checks both unintended acquisition and source mutation.

2. **Critical findings**

   None.

3. **Important findings**

   The README gap above. Add a short paragraph near `README.md:146`; the detailed contract already exists in `atlas/workflow/sdlc-binary.md:1304`.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `reviewbox`, `judge`, and `layergraph` package tests.
   - Targeted production boundary, dependency lifecycle, freshness, and local-check environment tests.
   - An ephemeral overlay test confirming canonical protection paths under a symlinked checkout root.
   - `git diff --check`.

   Live OS-fence conformance could not execute here: `sandbox-exec` returned `sandbox_apply: Operation not permitted`. Kernel enforcement remains unverified in this environment. No tracked files were changed.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared resolver, environment projection, and checkout ownership.
   - **ARCH-PURE — pass:** descriptor rendering, budgets, and projection remain pure; Git effects stay at integration boundaries.
   - **ARCH-PURPOSE — flag:** functionality serves the issue; README completion remains outstanding.
   - **ARCH-MOCK — pass:** portable Git fixtures and existing injected runner exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** row, metadata, entry, blob, and preparation-time limits are enforced.
   - **ARCH-SECURE — pass:** transport restrictions, configuration suppression, strict metadata parsing, and canonical paths are present; live enforcement limitation noted above.
   - **ARCH-ORDER — pass:** ownership guards and cancellation-aware retry preserve pins and prevent another launch after cancellation.
   - **ARCH-FUNERAL — pass:** children participate in reset, close, and existing sweep cleanup.
   - **ARCH-EVOLVE — pass:** no persisted format change.

7. **Plan revision recommendations**

   Add a `## Revisions` entry recording the README update for boundary-review dependency bindings and preparation-command usage. No architectural redesign is needed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"be9481b7-d545-4949-9312-442f6cf56e9c","repository":"github.com/xianxu/ariadne","issue_id":"000551","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/8b88f57d222b86c6","reviewed_head":"0ae0016dbbe6634bfbf59132045e683bfa76862a","slot":"ariadne:1","requested_at":"2026-10-22T13:05:07.029365Z"},"phase":"finished","pid":91308,"started_at":"2026-10-22T13:05:07.208371Z","finished_at":"2026-10-22T13:10:10.698053Z","session_id":"01a163b7-1e4b-76a0-89a0-d52ba436dcc3","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T06:14:12-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 551 — review box gives the reviewer no dependency checkouts; consumer reviews can't verify substrate claims (echo#9 BR-1 unverifiable) |
| repo | ariadne |
| issue file | workshop/issues/000551-review-box-gives-the-reviewer-no-dependency-checkouts-consumer-reviews-can-t-verify-substrate-claims-echo-9-br-1-unverifiable.md |
| boundary | whole-issue close |
| milestone | — |
| window | cce05bc3f48f61cb63d010060ad2cb79e5ebd1c2..80e095147c447abf1a3c41799029a050eca5817f |
| command | sdlc close --issue 551 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T06:14:12-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

BR-1 is addressed, and the inspected implementation matches the dependency snapshot contract. Package and targeted integration tests passed. No executable defect was found; the remaining blocker is verification of kernel write protection, which this environment prevented.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:146–160 now documents disposable bindings, last-fetched origin/main pins, no automatic fetch, repair/rerun recovery, isolation modes, and retry/cleanup lifetime, linking atlas/workflow/sdlc-binary.md. The documented behavior agrees with reviewdependencies.go, checkout.go, and confinement wiring; this correction changes prose only.
findings:
  - id: new
    severity: Important
    kind: unverifiable
    family: os-fence-conformance
    title: |
      Kernel enforcement of dependency write protection could not be verified
    detail: |
      At cmd/sdlc/reviewdependencies_live_test.go:22, WF_LIVE_REVIEW_DEPENDENCIES=1 go test ./cmd/sdlc -run '^TestReviewDependencyLiveFence$' -count=1 failed with exit status 71: sandbox-exec: sandbox_apply: Operation not permitted. ARCH-SECURE requires evidence that dependency clones and source Git directories reject writes while the primary review checkout remains writable. Provide captured passing output for this exact pinned head from a host permitting the OS fence, or rerun review in such an environment; the issue's prose report alone does not establish that measurement.
```

1. **Strengths**

   - Prompt and process bindings derive from the same dependency descriptors.
   - Dependency setup checks tree budgets before materialization and suppresses implicit fetching.
   - Ownership tests exercise reset, cancellation, dead-owner sweeping, and cleanup after partial failure.
   - README and atlas both explain the new surface and fallback guarantees.

2. **Critical findings**

   None.

3. **Important findings**

   The OS-fence verification gap above. Supply measured host evidence; no implementation change is established as necessary.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/reviewbox ./cmd/sdlc/internal/judge ./pkg/layergraph`
   - Targeted boundary, dependency lifecycle, freshness, promisor, and local-check environment tests.
   - `git diff --check`

   Pure descriptor, budget, and environment tests use values without IO. Integration tests inspect real disposable Git state. BR-1 is prose-only, so no behavior-changing regression proof is required. Tracked files remained unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared declaration resolution and environment projection.
   - **ARCH-PURE — pass:** descriptor rendering and budget policy remain separate from Git effects.
   - **ARCH-PURPOSE — pass:** both boundary paths expose pinned dependencies through prompts and commands.
   - **ARCH-MOCK — pass:** portable Git fixtures exercise the production process boundary.
   - **ARCH-CONSTRAINTS — pass:** explicit row, metadata, entry, byte, and time bounds.
   - **ARCH-SECURE — flag:** kernel enforcement awaits the measurement above.
   - **ARCH-ORDER — pass:** ownership and contextual retry paths have cancellation and partial-failure coverage.
   - **ARCH-FUNERAL — pass:** child clones reuse locked cleanup and sweep lifecycles.
   - **ARCH-EVOLVE — pass:** no new persisted format; descriptors are review-scoped.

   Core-concept locations and classifications match the inspected code.

7. **Plan revision recommendations**

   None. Attach the missing conformance evidence to the existing verification item.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"429bdc0f-ef80-458d-ace1-942344e84ca8","repository":"github.com/xianxu/ariadne","issue_id":"000551","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/ab6625d403cb6bb6","reviewed_head":"80e095147c447abf1a3c41799029a050eca5817f","slot":"ariadne:1","requested_at":"2026-10-22T13:12:26.269452Z"},"phase":"finished","pid":10699,"started_at":"2026-10-22T13:12:26.439918Z","finished_at":"2026-10-22T13:14:11.234551Z","session_id":"01a163bd-d22b-7150-91b7-3c87affaf86a","outcome":"returned"}] -->
