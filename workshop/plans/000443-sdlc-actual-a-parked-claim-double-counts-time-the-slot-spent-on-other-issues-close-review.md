# Boundary Review — ariadne#443 (whole-issue close)

| field | value |
|-------|-------|
| issue | 443 — sdlc actual: a parked claim double-counts time the slot spent on other issues |
| repo | ariadne |
| issue file | workshop/issues/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues.md |
| boundary | whole-issue close |
| milestone | — |
| window | f92193cf4afc9d2b302e132ec5a639bbe3e05098..9d55411b82bca7659de5a1cd0d34aa9c911edc32 |
| command | sdlc close --issue 443 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T17:42:03-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The pinned range implements same-slot interval subtraction and retained replay evidence, but three reproduced correctness defects block shipping: diagnostic-only sources can abort measurement, unrelated slots can cause refusals, and unadopted verification time becomes exclusion authority.

Done-when: current — preventing duplicate attribution and explaining exclusions still covers the revised implementation.

1. **Strengths**

   - Pure interval subtraction handles overlapping protection, slot isolation, and queue subtraction order.
   - Protected intervals survive snapshot serialization and historical replay.
   - Completion-window filtering covers the retained-claimant case; atlas documentation explains the changed flow.

2. **Critical findings**

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: measurement-population-isolation
    title: |
      Clear protected intervals from the excluded-brain diagnostic pass
    detail: |
      cmd/sdlc/actualmeasurement.go:38 adds ProtectedIntervals to options, but the diagnostic copy at line 98 retains them while clearing SourceSlots. An overlapping unowned brain session therefore aborts a valid candidate with "excluded brain provenance: ... without unique slot ownership". A temporary overlay of TestOwnedActualEngineReplayAndExcludedBrainProvenance reproduced this with five protected minutes. Clear ProtectedIntervals alongside QueueIntervals in the diagnostic population and add a combined regression.
  - id: new
    severity: Critical
    kind: defect
    family: adopted-credit-authority
    title: |
      Do not protect verification-tail time omitted from adopted actuals
    detail: |
      cmd/sdlc/actualexclusion.go:206 converts all replayed segments into protection, although actualmeasurement.go:69 adopts only Build+Converge and excludes Tail. An overlay of TestActualQueueMeasurementReplay demonstrated protection of 00:50–01:00 while the result explicitly reported those ten minutes as uncounted tail. This can silently remove another issue's valid activity. ARCH-PURPOSE: derive protected intervals from the same counted phases that determine adopted hours; cover wholly-tail runs and runs crossing phase boundaries.
  - id: new
    severity: Critical
    kind: defect
    family: overlap-relevance-before-evidence
    title: |
      Rule out unrelated slots before refusing missing adoption evidence
    detail: |
      cmd/sdlc/actualexclusion.go:153 requires a saved measurement window before checking same-slot engagement at line 182. A temporally plausible legacy adoption without ActualInputs blocks candidates even when its engagement belongs exclusively to another slot. An overlay of the parked-claim fixture, using candidate slot ariadne:7 versus adopted slot ariadne:6 and omitting the saved inputs, reproduced "issue #649 adopted actual has no trustworthy saved window". Use trustworthy ownership history to rule out unrelated slots before refusing missing pins; retain refusal when same-slot overlap remains possible. ARCH-PURPOSE.
  - id: new
    severity: Important
    kind: defect
    family: declared-workload-verification
    title: |
      Deliver the planned unrelated-card scaling regression
    detail: |
      workshop/plans/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues-plan.md:37 claims tests count replays for N and 2N unrelated cards, but no such test exists in the reviewed changes. ARCH-CONSTRAINTS: add representative loader coverage proving unrelated cards do not trigger transcript replay, including other-slot and disjoint-window cases, or revise the unsupported verification claim and supply equivalent measured evidence.
```

3. **Important findings**

   The missing workload regression is recorded above.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed both complete `activetime` and `gatestate` package tests.
   - Passed the targeted actual-time command suite in 161.19 seconds.
   - Passed seven focused baseline integration/helper tests.
   - Three temporary Go test overlays reproduced the Critical findings. Temporary files were removed; tracked files remained unchanged.
   - No prior findings required disposition.

6. **Architectural notes**

   - **ARCH-DRY: pass** — completion/correction snapshot selection is shared.
   - **ARCH-PURE: pass** — subtraction remains pure and directly tested; evidence loading stays in the IO layer.
   - **ARCH-PURPOSE: flag** — raw allocations overstate adopted credit, and unrelated slots can block measurement.
   - **ARCH-MOCK: pass** — integration coverage uses existing portable, stateful Git/tracker fixtures.
   - **ARCH-CONSTRAINTS: flag** — the declared scaling regression is absent.
   - **ARCH-SECURE: pass** — protected snapshot fields receive identity and interval validation.
   - **ARCH-ORDER: pass** — historical replay retains its selected exclusion generation.
   - **ARCH-FUNERAL: pass** — evidence extends the existing snapshot family under its existing byte limit.

   Atlas coverage is present. No new command, flag, or configuration surface requires a README update.

7. **Plan revision recommendations**

   Append `## Revisions` entries defining protection as counted Build/Converge intervals, documenting other-slot rejection before missing-evidence refusal, and adding the diagnostic-population interaction and workload regressions. Keep conservative fractional-allocation protection distinct from verification-tail exclusion.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5d67fbf6-8be0-41bc-a333-1cfe4b5dc604","repository":"github.com/xianxu/ariadne","issue_id":"000443","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/99d57ab5275b223c","reviewed_head":"9d55411b82bca7659de5a1cd0d34aa9c911edc32","slot":"ariadne:1","requested_at":"2026-10-19T00:38:10.879495Z"},"phase":"finished","pid":62582,"started_at":"2026-10-19T00:38:10.989691Z","finished_at":"2026-10-19T00:42:02.8808Z","session_id":"01a15198-32f2-7b20-83fa-e328e26688cb","outcome":"returned"}] -->
