# Boundary Review — ariadne#569 (whole-issue close)

| field | value |
|-------|-------|
| issue | 569 — Diagnose legacy actual-window refusal blocking pair#162 |
| repo | ariadne |
| issue file | workshop/issues/000569-legacy-actual-window.md |
| boundary | whole-issue close |
| milestone | — |
| window | b10a950e2693b79fe69b1d4279f1b752c9d48162..96cd72ed248907d4cdee3a026054be87cbd47a69 |
| command | sdlc close --issue 569 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T02:07:20-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the revised scope: exact old unmarked imports can establish temporal disjointness without certifying their historical completion or actual. No blocking defects found. README, help, and atlas describe the behavior and its limits.

Done-when: current — the clauses cover labelled disjointness, strict refusal controls, hybrid regression coverage, and the landing/readiness sequence.

```findings
{}
```

1. **Strengths**
   - Shared source and generation validation avoids a second history walker (`cmd/sdlc/legacycompletion.go:36`).
   - Bound-only mode returns before completion authority is populated (`cmd/sdlc/legacycompletion.go:197`).
   - The exclusion path requires strict temporal separation and emits an issue/end warning (`cmd/sdlc/actualexclusion.go:323`).
   - Real-Git tests cover restored generations and verify that disjointness leaves refs and checkout unchanged (`cmd/sdlc/legacydisjoint_test.go:59`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - New disjointness and hybrid measurement regressions passed normally and under race detection.
   - Strict damaged-evidence, unmarked-import, history-limit, history-read-failure, and pure generation-selection controls passed.
   - `git diff --check` passed; checkout remained clean.
   - The full recorded suite was not rerun during this review.

6. **Architectural notes**
   - **ARCH-DRY: pass** — immutable history validation remains shared.
   - **ARCH-PURE: pass** — temporal comparison and generation selection remain pure; Git reads stay in the integration layer.
   - **ARCH-PURPOSE: pass** — covers the full earliest prefix/engagement boundary and preserves overlap refusal.
   - **ARCH-EVOLVE: pass** — no persisted format changes or new rewrite paths.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"76de1c9f-4d21-4ad9-8008-975baebf257e","repository":"github.com/xianxu/ariadne","issue_id":"000569","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/38cb31914bca19b2","reviewed_head":"96cd72ed248907d4cdee3a026054be87cbd47a69","slot":"ariadne:3","requested_at":"2026-10-23T09:05:06.531615Z"},"phase":"finished","pid":52738,"started_at":"2026-10-23T09:05:06.625098Z","finished_at":"2026-10-23T09:07:20.249297Z","session_id":"01a16801-bd55-7ad2-9952-0cd55678e92a","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T02:23:01-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 569 — Diagnose legacy actual-window refusal blocking pair#162 |
| repo | ariadne |
| issue file | workshop/issues/000569-legacy-actual-window.md |
| boundary | whole-issue close |
| milestone | — |
| window | b10a950e2693b79fe69b1d4279f1b752c9d48162..662b28ccbfc4c6e2fc16a3d3f07cee1446196c23 |
| command | sdlc close --issue 569 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T02:23:01-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared generation reader, strict temporal comparison, and documentation updates are sound approaches. However, the new archive shortcut can bypass modern or unreadable evidence saved after the selected close commit. Reproduced failures block SHIP.

Done-when: current — its clauses cover labelled disjointness, earliest-prefix bounds, and modern-corruption refusal, including the revised source-authority requirement.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: legacy-bound-must-check-source-authority
    title: |
      Modern evidence saved after the selected close escapes the archive-bound guard
    detail: |
      cmd/sdlc/legacycompletion.go:187-192 checks only selected.CloseCommit when present. A modern or unreadable ledger saved at archival, or in a later ledger-only commit before cutover, is therefore ignored and loadAdoptedExclusions permits the disjoint skip. Both timings were reproduced with independent damaged-input, retained-input, reviewed-head-only, and unreadable-ledger cases: all returned nil error and a skip warning. This is the sole production guard in this window, consumed by cmd/sdlc/actualexclusion.go:323. ARCH-PURPOSE: validate authority at the terminal and pinned source snapshots as well as the selected close; add regressions covering both timings and all four evidence forms.
  - id: new
    severity: Minor
    kind: defect
    family: comments-must-match-covered-branches
    title: |
      The old-import fixture comment contradicts the new partial-ledger branch
    detail: |
      cmd/sdlc/actualexclusion_test.go:625-626 says this generation never had a review-bearing close or archive transition, but mode.damaged immediately creates both through legacyCapturedClose and f.terminal. This is the only contradictory fixture-description instance found in the window. Move the comment into the born-archived branch or describe both modes.
```

1. **Strengths**

   - Source and tracker generation checks remain shared rather than introducing another history walker.
   - Bound-only mode returns without granting completion authority.
   - Real-Git tests exercise equality, overlap, restored generations, and unchanged refs/cards.
   - README, embedded help, and both relevant atlas pages document the new behavior.

2. **Critical findings**

   - `legacycompletion.go:187–192`: inspect later immutable ledger snapshots before granting archive-bound authority. The existing fixture always writes its ledger before the close, masking this gap.

3. **Important findings**

   None.

4. **Minor findings**

   - `actualexclusion_test.go:625`: narrow the born-archived comment to its actual branch.

5. **Test coverage notes**

   - Unmodified disjointness and hybrid measurement tests passed: **40.046s**.
   - Strict damaged-evidence, unmarked-recovery, and temporal-order controls passed: **7.540s**.
   - Pure close-selection tests passed: **2.154s**.
   - Temporary test overlays reproduced all four authority failures at both later-write timings, without changing production code. Overlays were removed; checkout remains clean.
   - `git diff --check` passed. Full suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared history and generation readers.
   - **ARCH-PURE: pass** — temporal comparison and selection remain pure; Git access stays in integration functions.
   - **ARCH-PURPOSE: flag** — the modern-source-authority refusal is incomplete.
   - **ARCH-EVOLVE: pass** — no persisted format or rewrite changes.

7. **Plan revision recommendations**

   Append a `## Revisions` entry extending the source-authority acceptance matrix to evidence saved at archival and after archival before cutover. Require independent retained-input, damaged-input, reviewed-head-only, and unreadable-ledger regressions for both timings.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f1dc5fac-41f0-49b3-adad-7e763e482c2b","repository":"github.com/xianxu/ariadne","issue_id":"000569","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/92da1d8b41348ec6","reviewed_head":"662b28ccbfc4c6e2fc16a3d3f07cee1446196c23","slot":"ariadne:3","requested_at":"2026-10-23T09:18:45.317596Z"},"phase":"finished","pid":10671,"started_at":"2026-10-23T09:18:45.420206Z","finished_at":"2026-10-23T09:23:00.772908Z","session_id":"01a1680e-3bf2-7263-9dc6-3d743cd31bfd","outcome":"returned"}] -->
