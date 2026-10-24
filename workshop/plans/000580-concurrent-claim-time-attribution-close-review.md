# Boundary Review — ariadne#580 (whole-issue close)

| field | value |
|-------|-------|
| issue | 580 — Active-time attribution cross-counts concurrent same-slot claims instead of active branches |
| repo | ariadne |
| issue file | workshop/issues/000580-concurrent-claim-time-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84b68e2fc5c0773ff38f13771507d93aaa08efd4..8d3d5948cfdfaf4eefa19957d73be00e09442d4e |
| command | sdlc close --issue 580 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T20:51:25-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The main attribution repair works, and its incident regression fails when branch narrowing is disabled. One persisted-evidence validation gap remains: contradictory `ownership-only` evidence can restore foreign-branch minutes during replay. This needs a fix and regression test before closing the issue.

1. **Strengths**

   - Actual, close, and tracked `active-time` share the branch selector.
   - Native Git fixtures exercise worktree-local history, capture races, missing logs, and branch transitions.
   - Version-3 snapshots preserve dispatch admission separately from selected windows; old-reader refusal and nested unknown-field preservation are tested.
   - README, help, and atlas document the changed behavior and recovery path.

2. **Critical findings**

   None.

3. **Important findings**

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: retained-attribution-consistency
    title: |
      Ownership-only replay accepts a witnessed foreign issue branch
    detail: |
      cmd/sdlc/actualbranch.go:311–319 validates branch/foreign-branch identity but accepts ownership-only intervals regardless of their canonical branch ID; measurement_branch.go:85 also leaves this case unchecked. A temporary test overlay relabelled the incident fixture's foreign-branch interval as ownership-only and widened its saved selected window: resolveOwnedReplay and computeOwnedActualFor accepted it, crediting B with 29 minutes instead of 11. ARCH-SECURE requires contradictory persisted evidence to fail visibly. Reject ownership-only evidence naming any canonical issue branch, matching SelectBranchActivity's rule, and add a regression through saved-input replay.
```

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed complete `internal/activetime`, `internal/gatestate`, and `internal/tracker` package tests.
   - Passed targeted command, resolver, ownership-lifecycle, replay, and capture tests.
   - Independently disabling branch narrowing made the existing incident regression fail on A’s tail being credited to B.
   - The added temporary replay regression failed against unchanged production code: **29 minutes credited versus the original 11**.
   - Full repository suite was not rerun. Temporary overlays were removed; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** consumers share the classifier and activity engine.
   - **ARCH-PURE — pass:** interval selection and history decoding have direct pure tests; capture remains in the IO layer.
   - **ARCH-PURPOSE — pass:** the incident ordering and all three hour consumers are covered.
   - **ARCH-MOCK — pass:** injected readers and portable native Git fixtures exercise the production boundary.
   - **ARCH-CONSTRAINTS — pass:** capture and snapshot bounds are enforced; counted controls cover request caching and interval sweeps.
   - **ARCH-SECURE — flag:** contradictory retained attribution is accepted, as detailed above.
   - **ARCH-ORDER — pass:** half-open transitions, capture interleavings, dispatch boundaries, and reclaim behavior have controls.
   - **ARCH-FUNERAL — pass:** provenance uses the existing bounded snapshot and ledger lifecycle.
   - **ARCH-EVOLVE — pass:** nested unknown fields survive rewrites; the version boundary has a concrete replay-interpretation justification.

7. **Plan revision recommendations**

   Add a `## Revisions` entry recording that retained-proof validation must enforce every attribution kind’s branch constraints, including rejecting canonical issue branches labelled `ownership-only`, with a saved-replay regression.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4698faab-b6cf-401f-9716-a2124871edff","repository":"github.com/xianxu/ariadne","issue_id":"000580","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c5dff2b32ce509f2","reviewed_head":"8d3d5948cfdfaf4eefa19957d73be00e09442d4e","slot":"ariadne:3","requested_at":"2026-10-24T03:46:53.993784Z"},"phase":"finished","pid":71707,"started_at":"2026-10-24T03:46:54.106611Z","finished_at":"2026-10-24T03:51:24.903377Z","session_id":"01a16c04-c552-7e60-a70d-b66340c5d8f3","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T20:58:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 580 — Active-time attribution cross-counts concurrent same-slot claims instead of active branches |
| repo | ariadne |
| issue file | workshop/issues/000580-concurrent-claim-time-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84b68e2fc5c0773ff38f13771507d93aaa08efd4..dd39e615a6561921075b0066ba0cd09a8e0ca22d |
| command | sdlc close --issue 580 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T20:58:04-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the attribution repair: branch evidence narrows ownership-admitted activity, foreign tails stay excluded, ambiguous concurrent reservations refuse measurement, and saved snapshots replay independently of live history. BR-1 is verified addressed. No new blocking findings were identified.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      actualbranch.go:311 rejects ownership-only intervals naming any canonical issue branch. Saved-replay and direct-execution regressions cover both own and foreign branches. Removing only this guard with a temporary Go overlay made both regressions fail; foreign replay again credited 29 minutes instead of 11. The pinned-head tests pass.
```

1. **Strengths**

   - The shared selector preserves ownership prerequisites and excludes witnessed foreign branches, including unadopted tails (`cmd/sdlc/internal/activetime/branch.go:137–154`).
   - Saved-proof validation runs before replay and execution (`cmd/sdlc/actualbranch.go:298`).
   - Native worktree tests exercise real checkout, rename, detach, merge and rebase history, alongside controlled capture races.
   - README, atlas and command help document the tracked-command semantic change and v3 compatibility boundary.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Passed focused command, incident, capture, selector and branch-codec tests.
   - Passed complete `activetime`, `gatestate` and `tracker` package suites.
   - Independently confirmed BR-1’s regression fails without its fix.
   - `git diff --check` passed; checkout remains clean.
   - The repository-wide suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** actual, close and tracked active-time share branch selection.
   - **ARCH-PURE — pass:** interval selection, history decoding and evidence codecs remain separate from IO.
   - **ARCH-PURPOSE — pass:** the incident fixture covers foreign post-close tails and cross-issue partitioning.
   - **ARCH-MOCK — pass:** existing dependency seams, private native Git fixtures and injected capture ordering exercise production boundaries.
   - **ARCH-CONSTRAINTS — pass:** capture and snapshot bounds are enforced; counted tests cover request-scoped IO reuse.
   - **ARCH-SECURE — pass:** malformed, incomplete and contradictory evidence fails visibly; BR-1 closes the retained-kind inconsistency.
   - **ARCH-ORDER — pass:** explicit half-open interval transitions preserve dispatch origins; deterministic tests cover boundary ordering and capture races.
   - **ARCH-FUNERAL — pass:** evidence uses the existing bounded snapshot and ledger lifecycle; no independent durable artifact family is introduced.
   - **ARCH-EVOLVE — pass:** nested unknown fields survive rewrites, known fields remain strict, and old readers refuse the unsafe-to-ignore v3 interpretation.

7. **Plan revision recommendations:** None. Existing Revisions explain the implemented entities, standalone-command change and approved version boundary.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b39d6168-bb8e-4e0c-9df9-40cc0e7692f7","repository":"github.com/xianxu/ariadne","issue_id":"000580","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2c5f6f4d6f7ba6df","reviewed_head":"dd39e615a6561921075b0066ba0cd09a8e0ca22d","slot":"ariadne:3","requested_at":"2026-10-24T03:55:28.370329Z"},"phase":"finished","pid":57059,"started_at":"2026-10-24T03:55:28.469924Z","finished_at":"2026-10-24T03:58:03.231204Z","session_id":"01a16c0c-9e7c-7b51-95cd-059a51c55db1","outcome":"returned"}] -->
