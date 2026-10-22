# Boundary Review — ariadne#541 (whole-issue close)

| field | value |
|-------|-------|
| issue | 541 — weave refresh: 'target dependency declarations changed; reconcile that change separately' should name the recovery steps and the changed declaration |
| repo | ariadne |
| issue file | workshop/issues/000541-weave-refresh-target-dependency-declarations-changed-reconcile-that-change-separately-should-name-the-recovery-steps-and-the-changed-declaration.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5aae36f628a4f9ea04439543f46995a45d61b37..048fc292ccbef0f535748ed1b50fc6e3df5f358c |
| command | sdlc close --issue 541 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T01:09:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation delivers actionable recovery guidance while preserving preflight refusal and omitting source values. Targeted tests passed. One Important documentation gap remains: README retains the vague recovery instruction this issue replaces.

Done-when: current — the clauses cover recovery commands, declaration details, secrecy, divergence, and refusal before updates or compilation.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-facing-recovery-docs
    title: |
      README update appears missing for declaration-drift recovery
    detail: |
      README.md:372–376 still instructs readers to “reconcile that declaration change separately,” while cmd/weave/internal/refresh/refresh.go:157–161 introduces concrete recovery commands. The complete family in this window is this declaration-drift recovery flow: its atlas documentation was updated at atlas/workflow/setup-and-replication.md:66–85, but its README entry was not. Update the README entry with the clean-behind prerequisite, fast-forward/setup/retry steps, and ahead/divergent qualification, linking to the atlas recovery section.
```

1. **Strengths**

   - `declarations.go:13–21` omits Source values by construction and suppresses suspicious path/mount labels.
   - `declarations.go:26–55` reports ordered declaration differences, including duplicate rows and additions/removals.
   - `refresh.go:157–161` reuses observed ancestry to qualify recovery guidance without adding Git calls.
   - The integration fixture checks unchanged HEADs and zero compilation alongside the diagnostic.

2. **Critical findings**

   None.

3. **Important findings**

   README recovery documentation, detailed above. This is a small documentation correction.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:
   
   - Refresh and CLI tests selected by `TestDeclaration|TestObserveRefusals|TestSubstrateOnly|TestRefresh`.
   - `pkg/layergraph` tests selected by `^TestParseRows`.
   - Pinned-range `git diff --check`.

   Coverage exercises clean-behind, divergent, substrate-only, and credential-bearing source cases. Pure formatter tests cover changed fields, ordering, duplicates, unchanged rows, and suspicious labels. No prior findings required regression-disposition verification.

6. **Architectural notes**

   - **ARCH-DRY: pass** — reuses parsed dependency rows, ancestry evidence, and existing fixtures.
   - **ARCH-PURE: pass** — formatting is deterministic and tested without IO; Git remains behind the existing injected runner.
   - **ARCH-PURPOSE: pass** — identifies changed declarations and supplies conditional recovery across the shared refresh path.
   - **ARCH-EVOLVE: pass** — no persisted format or rewrite path changes.

7. **Plan revision recommendations**

   None. The implementation matches the plan; the README correction satisfies the boundary’s additional documentation gate.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"cafd23b9-bfa7-4b60-b085-71ebb1ae3888","repository":"github.com/xianxu/ariadne","issue_id":"000541","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/fdac5adef85ff8b7","reviewed_head":"048fc292ccbef0f535748ed1b50fc6e3df5f358c","slot":"ariadne:5","requested_at":"2026-10-22T08:06:50.634705Z"},"phase":"finished","pid":91467,"started_at":"2026-10-22T08:06:50.843227Z","finished_at":"2026-10-22T08:09:43.775316Z","session_id":"01a162a6-0aae-7583-a510-f89b7ea16ff9","outcome":"returned"}] -->
