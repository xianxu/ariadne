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
