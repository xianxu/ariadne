# Boundary Review — ariadne#428 (whole-issue close)

| field | value |
|-------|-------|
| issue | 428 — Legacy-attempt actuals refusal should name --no-actual (exclude), not 'ask the TL to restore evidence' |
| repo | ariadne |
| issue file | workshop/issues/000428-legacy-attempt-actuals-refusal-should-name-no-actual-exclude-not-ask-the-tl-to-restore-evidence.md |
| boundary | whole-issue close |
| milestone | — |
| window | f0c955edb8014d073ebe7f007a4eee3d6f319dca..b0119e262de3e2a86679ddc24182bc970f1ace75 |
| command | sdlc close --issue 428 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T12:34:44-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The single-legacy refusal gives the correct boundary command, but classification uses only the first blocker. A mixed window can therefore lose recovery guidance, violating the Spec’s “every blocking attempt is legacy” condition.

Done-when: current — the revisions clarify command syntax and reason placement; the existing acceptance clauses still cover that behavior.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: complete-blocker-classification
    title: |
      Legacy exclusion advice is selected without checking every blocker
    detail: |
      cmd/sdlc/actual.go:381 classifies the window from one diagnostic, while cmd/sdlc/actualinputs.go:340 forwards only notes[0]. A temporary regression test using the production selector reproduced a legacy attempt followed by an exact attempt with unknown slot provenance: the output recommends exclusion and suppresses recovery guidance; reversing the order retains recovery guidance. This violates the Spec's all-legacy condition. The family includes legacy-first mixtures with either other selector-warning category (unknown queue provenance or unknown slot provenance), across actual previews with and without --milestone, issue close, and milestone close. Preserve and classify the complete blocker set, select exclusion-only guidance only when all blockers are legacy, and test both orders.
```

1. **Strengths**

   - Reuses `rerunCmd`, preserving issue and milestone command syntax.
   - Keeps numeric refusal intact; exclusion remains explicit.
   - Rendering tests cover all four boundary variants, and atlas/help explain the exclusion path.

2. **Critical findings**

   - `actual.go:381`: incomplete blocker classification, detailed above. Carry the complete diagnosis from the selector to the renderer.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused rendering, selector, and evidence-refusal tests passed. The temporary mixed-blocker regression failed as described above. `git diff --check` passed; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared command construction and rendering.
   - **ARCH-PURE: pass** — added classification and rendering require no IO-dependent tests.
   - **ARCH-PURPOSE: flag** — the all-legacy condition requires evaluating every blocker.

   Atlas coverage is present. No new command, flag, or configuration surface requires a README addition.

7. **Plan revision recommendations**

   Add a `## Revisions` entry requiring complete blocker classification and mixed-blocker regression tests in both orders across the shared rendering paths.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ec464232-8e97-449f-8bae-6102aef6414e","repository":"github.com/xianxu/ariadne","issue_id":"000428","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e14a2e514f1b61c9","reviewed_head":"b0119e262de3e2a86679ddc24182bc970f1ace75","slot":"ariadne:1","requested_at":"2026-10-17T19:33:07.400287Z"},"phase":"finished","pid":6070,"started_at":"2026-10-17T19:33:07.493936Z","finished_at":"2026-10-17T19:34:43.668604Z","session_id":"01a14b5a-8c08-7643-800a-c4e458535dc6","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T13:30:02-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 428 — Legacy-attempt actuals refusal should name --no-actual (exclude), not 'ask the TL to restore evidence' |
| repo | ariadne |
| issue file | workshop/issues/000428-legacy-attempt-actuals-refusal-should-name-no-actual-exclude-not-ask-the-tl-to-restore-evidence.md |
| boundary | whole-issue close |
| milestone | — |
| window | f0c955edb8014d073ebe7f007a4eee3d6f319dca..7534616855a6c0538ab48ad1cea189e77dbf5eac |
| command | sdlc close --issue 428 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T13:30:02-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation resolves BR-1: selector blockers survive the error boundary, and exclusion advice requires an entirely legacy blocker set. Mixed and malformed evidence retains recovery guidance. No new findings.

Done-when: current — the revisions clarify command syntax and complete-blocker classification within the existing acceptance behavior.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      actualinputs.go:340 preserves every selector blocker; actual.go:108 carries them through wrapped errors, and actual.go:357 classifies the entire set. Production-selector regressions cover both unknown-provenance categories in both orders across all four rendering paths. Temporary Go overlays restoring first-blocker behavior made the attachment and mixed-rendering regressions fail.
```

1. **Strengths**

   - Complete diagnostics reach classification without truncation (`cmd/sdlc/actualinputs.go:340`, `cmd/sdlc/actual.go:108`).
   - Shared rendering and `rerunCmd` preserve consistent issue and milestone guidance (`cmd/sdlc/actual.go:404`).
   - Regression coverage includes mixed orders, multiple legacy IDs, wrapped errors, and saved attempt records.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused regressions, the broader actuals tests, and malformed-record/endpoint tests passed. Both first-blocker mutations failed as expected. `git diff --check` passed. Temporary overlays were removed; tracked files remain unchanged. The full repository suite and build were not rerun.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared classifier, rendering, and command construction.
   - **ARCH-PURE: pass** — classification remains deterministic and directly tested; filesystem evidence stays in the existing resolver boundary.
   - **ARCH-PURPOSE: pass** — both mixed-blocker categories and both orders are covered across previews and close paths.

   Atlas and CLI help reflect the behavior. No new command, flag, or configuration surface requires a README addition.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"abb6bc6a-fbc7-45e4-be94-04b86962189f","repository":"github.com/xianxu/ariadne","issue_id":"000428","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/3641fd4ff1d5ca72","reviewed_head":"7534616855a6c0538ab48ad1cea189e77dbf5eac","slot":"ariadne:1","requested_at":"2026-10-17T20:28:34.408339Z"},"phase":"finished","pid":44064,"started_at":"2026-10-17T20:28:34.506922Z","finished_at":"2026-10-17T20:30:01.897675Z","session_id":"01a14b8d-503e-7ed3-895a-a95c3aebfcf5","outcome":"returned"}] -->
