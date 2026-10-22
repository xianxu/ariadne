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

---

## Re-review — 2026-10-22T01:13:08-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 541 — weave refresh: 'target dependency declarations changed; reconcile that change separately' should name the recovery steps and the changed declaration |
| repo | ariadne |
| issue file | workshop/issues/000541-weave-refresh-target-dependency-declarations-changed-reconcile-that-change-separately-should-name-the-recovery-steps-and-the-changed-declaration.md |
| boundary | whole-issue close |
| milestone | — |
| window | f5aae36f628a4f9ea04439543f46995a45d61b37..47286711ea6fc2bca08fdaa78a7100fad5a6a086 |
| command | sdlc close --issue 541 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T01:13:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation satisfies the issue’s recovery-diagnostic contract, and BR-1’s README correction is complete. Targeted tests passed. No blocking findings remain.

Done-when: current — its clauses cover recovery commands, changed declarations and fields, source secrecy, divergence, and refusal before updates or compilation.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:373–380 now states the clean-behind prerequisite, fast-forward/setup/retry commands, ahead/divergent qualification, and atlas recovery link. These agree with atlas/workflow/setup-and-replication.md:66–85 and refresh.go:157–162. The pinned before/after confirms this prose-only correction; no wording test is required.
```

1. **Strengths**

   - `declarations.go:13–21` excludes source values by construction and suppresses URL/scp-like labels.
   - `declarations.go:26–55` preserves declaration ordinals, distinguishing duplicates and reporting additions, removals, reorderings, and changed fields.
   - `refresh.go:157–162` uses existing ancestry evidence to qualify recovery without additional Git calls.
   - The integration fixture checks diagnostic content, unchanged HEADs, and zero compilation.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None requiring action.

5. **Test coverage notes**

   Passed targeted refresh, CLI, and parser tests using:
   `go test ./cmd/weave/internal/refresh ./cmd/weave ./pkg/layergraph -run 'TestDeclaration|TestObserveRefusals|TestSubstrateOnly|TestRefresh|^TestParseRows' -count=1`.

   Coverage includes clean-behind, divergent, substrate-only, credential-bearing sources, ordered differences, duplicate declarations, and unchanged rows. Formatter tests execute without IO; integration tests use private real-Git fixtures and an injected transport seam.

6. **Architectural notes**

   - **ARCH-DRY: pass** — reuses parsed declarations, ancestry evidence, and existing fixture helpers.
   - **ARCH-PURE: pass** — deterministic formatting remains separate from injected Git IO.
   - **ARCH-PURPOSE: pass** — delivers declaration identification, safe recovery guidance, secrecy, and README/atlas documentation.
   - **ARCH-EVOLVE: pass** — no persisted format or read-modify-write behavior changes.

7. **Plan revision recommendations:** None. The plan matches the delivered implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"043b5c18-8d23-4a09-86f9-76ceaa789391","repository":"github.com/xianxu/ariadne","issue_id":"000541","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/925630ab084fa848","reviewed_head":"47286711ea6fc2bca08fdaa78a7100fad5a6a086","slot":"ariadne:5","requested_at":"2026-10-22T08:11:25.993397Z"},"phase":"finished","pid":99080,"started_at":"2026-10-22T08:11:26.141422Z","finished_at":"2026-10-22T08:13:08.015685Z","session_id":"01a162aa-3dca-72e1-84ca-e197c41c79b7","outcome":"returned"}] -->
