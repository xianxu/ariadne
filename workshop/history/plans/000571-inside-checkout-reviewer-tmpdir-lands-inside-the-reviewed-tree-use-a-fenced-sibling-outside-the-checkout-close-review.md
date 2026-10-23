# Boundary Review — ariadne#571 (whole-issue close)

| field | value |
|-------|-------|
| issue | 571 — Inside-checkout reviewer TMPDIR lands inside the reviewed tree; use a fenced sibling outside the checkout |
| repo | ariadne |
| issue file | workshop/issues/000571-inside-checkout-reviewer-tmpdir-lands-inside-the-reviewed-tree-use-a-fenced-sibling-outside-the-checkout.md |
| boundary | whole-issue close |
| milestone | — |
| window | e9e960c70a0513ad3235edd6ba7e1f56de188545..c13bf1c5d8f513a4b3f40c5db9d7e33525d79257 |
| command | sdlc close --issue 571 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T02:35:18-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation satisfies the scratch-placement contract and carries the sibling through reset, close, and dead-owner cleanup. Tests passed, documentation covers the new layout, and the mixed-binary rollout limitation is explicitly recorded. No blocking findings.

Done-when: current — the clauses cover outside-checkout scratch placement and a clean Parley review clone; the revision documents a deployment limitation without changing those outcomes.

```findings
{}
```

1. **Strengths**
   - One companion-path helper serves confinement and cleanup (`cmd/sdlc/internal/reviewbox/confine.go:152`).
   - Sweep preserves live companions through the existing checkout lock and reclaims orphaned scratch (`cmd/sdlc/internal/reviewbox/checkout.go:327`).
   - Regression tests exercise marker ancestry, invalid sibling types, lifetime ownership, and fresh scratch across retries.
   - README, atlas, and close help consistently describe the changed layout.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed both complete packages: `internal/reviewbox` and `internal/judge`.
   - Passed selected boundary-review isolation, retry, cancellation, and dependency lifecycle controls.
   - Inspected retained native Parley conformance evidence; production/probe blob IDs and raw-output hashes match. The native probe and full repository suite were not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY: pass** — shared path derivation and removal; existing lock reused.
   - **ARCH-PURE: pass** — deterministic derivation stays separate from filesystem operations.
   - **ARCH-PURPOSE: pass** — placement, marker discovery, retry, and cleanup are addressed together.
   - **ARCH-EVOLVE: pass** — no persisted-format changes. Mixed-binary cleanup compatibility is documented.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cad4b7e2-89e4-4f20-947f-0f1bdf4bd626","repository":"github.com/xianxu/ariadne","issue_id":"000571","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/41b704d36eb51fab","reviewed_head":"c13bf1c5d8f513a4b3f40c5db9d7e33525d79257","slot":"ariadne:4","requested_at":"2026-10-23T09:33:22.29513Z"},"phase":"finished","pid":14190,"started_at":"2026-10-23T09:33:22.459637Z","finished_at":"2026-10-23T09:35:17.661806Z","session_id":"01a1681b-9e5e-76a1-87a8-6b9a6c3a4a78","outcome":"returned"}] -->
