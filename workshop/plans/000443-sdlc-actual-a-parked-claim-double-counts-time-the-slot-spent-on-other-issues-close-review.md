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

---

## Re-review — 2026-10-18T18:32:20-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 443 — sdlc actual: a parked claim double-counts time the slot spent on other issues |
| repo | ariadne |
| issue file | workshop/issues/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues.md |
| boundary | whole-issue close |
| milestone | — |
| window | f92193cf4afc9d2b302e132ec5a639bbe3e05098..aea775197d15df067e707abfd47e0f80308db0a9 |
| command | sdlc close --issue 443 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T18:32:20-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All four prior findings are addressed, with regressions that fail when their fixes are removed. One blocking gap remains: an adopted issue with an overlapping saved window and unknown slot identity is silently skipped, allowing its time to be counted again. The broader actual-time tests pass, but do not cover this case.

Done-when: current — the existing criteria cover exclusive attribution and interval/refusal diagnostics.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The diagnostic pass clears ProtectedIntervals. Removing that line makes TestOwnedActualEngineReplayAndExcludedBrainProvenance fail with the original ownership error.
  - id: BR-2
    disposition: addressed
    note: |
      Protection intersects the same counted phases used for adopted hours. Restoring full-segment protection fails both TestCreditedIntervalsProtectOnlyCountedPhases and TestActualQueueMeasurementReplay, covering wholly-Tail and boundary-crossing runs.
  - id: BR-3
    disposition: addressed
    note: |
      Proven other-slot ownership is checked before missing-window refusal. Disabling differentSlot makes the parked-claim regression fail on legacy issue #650 without saved inputs.
  - id: BR-4
    disposition: addressed
    note: |
      The parked-claim fixture now checks four and eight unrelated adoptions while retaining same-slot protection, plus a disjoint-window case after transcript loss. These pass; disabling unrelated-slot rejection makes the four-card assertion fail.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: overlap-relevance-before-evidence
    title: |
      Refuse overlapping adopted actuals whose slot relationship remains unknown
    detail: |
      cmd/sdlc/actualexclusion.go:232–243 treats absence of a proven same-slot engagement as proof of irrelevance. Legacy owners may have an empty Workspace: differentSlot correctly declines to prove them unrelated, but the subsequent loop finds no match and silently continues. In a disposable copy of the parked-claim fixture, clearing only issue #649's claimant workspace produced a measured 0.333333h for #636 with no exclusions, recounting #649's adopted interval. An explicit refusal regression fails on head and passes with a temporary unknown-slot guard. ARCH-PURPOSE, ARCH-SECURE: this is the 2nd finding in family overlap-relevance-before-evidence. Earlier rounds fixed instances; do NOT fix only this instance. State and enforce the rule that skipping requires affirmative evidence of disjoint time or slot, enumerate unknown/empty/mixed ownership cases, and sweep them with regressions.
```

1. **Strengths**

   - Counted-phase protection derives from the timeline used to calculate adopted hours.
   - Protected intervals survive snapshot serialization and historical replay.
   - The parked-claim fixture exercises real transcript prefixes and isolated tracker history, including later transcript growth and missing evidence.

2. **Critical findings**

   The finding above blocks shipping. At `cmd/sdlc/actualexclusion.go:242`, distinguish **proven unrelated** from **unknown**. Use a total relevance classifier that returns skip, replay, or refusal; an unsuccessful same-slot lookup cannot establish permission to skip.

3. **Important findings**

   None beyond the blocking finding’s required regression and plan correction. The atlas maps the new flow. No new command, flag, or user configuration requires a README update.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `activetime` and `gatestate` package tests.
   - Focused regressions for the four prior findings.
   - `go test ./cmd/sdlc -run 'Test.*(Actual|Parked)' -count=1 -timeout 8m` — 223.311 seconds.
   - Pinned-range whitespace check.

   Fix-removal checks reproduced the prior failures. The additional unknown-slot refusal regression failed on head and passed with a temporary guard. Disposable copies were removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — historical replay and exclusion share snapshot selection.
   - **ARCH-PURE: pass** — interval subtraction and phase intersection are directly testable without IO.
   - **ARCH-PURPOSE: flag** — unknown adopted ownership can still permit duplicate attribution.
   - **ARCH-MOCK: pass for this delta** — existing seams and portable Git/tracker fixtures are reused; no new external dependency.
   - **ARCH-CONSTRAINTS: pass** — N/2N filtering coverage is present; snapshots retain their writer-enforced byte bound.
   - **ARCH-SECURE: flag** — legacy persisted identity uncertainty becomes fabricated evidence of non-overlap.
   - **ARCH-ORDER: pass** — no new asynchronous lifecycle; historical measurement generations remain pinned.
   - **ARCH-FUNERAL: pass** — exclusions use existing bounded snapshots and their established lifecycle.

7. **Plan revision recommendations**

   Add a `## Revisions` entry correcting the table at plan lines 73–78: **overlapping saved window + unknown slot relationship → refuse unless trustworthy evidence resolves relevance**. Enumerate empty, absent, and mixed ownership evidence. Follow the plan’s stated design-review requirement for a case outside its four rows.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bc02a72d-5886-421d-8761-dd310d146b97","repository":"github.com/xianxu/ariadne","issue_id":"000443","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/4ea3881561c3edfa","reviewed_head":"aea775197d15df067e707abfd47e0f80308db0a9","slot":"ariadne:1","requested_at":"2026-10-19T01:26:06.012622Z"},"phase":"finished","pid":11898,"started_at":"2026-10-19T01:26:06.102826Z","finished_at":"2026-10-19T01:32:12.373346Z","session_id":"01a151c4-1107-7010-b5ee-450d3e820140","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T19:30:02-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 443 — sdlc actual: a parked claim double-counts time the slot spent on other issues |
| repo | ariadne |
| issue file | workshop/issues/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues.md |
| boundary | whole-issue close |
| milestone | — |
| window | f92193cf4afc9d2b302e132ec5a639bbe3e05098..27b80215792fd4d22d3316830c11f88bbcabf1b5 |
| command | sdlc close --issue 443 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T19:30:02-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The exclusion core, pinned replay, and explicit decision table are sound improvements. BR-5’s empty-workspace fix is verified, but its broader rule remains incomplete: malformed, nonempty workspace identity still becomes affirmative “different slot” evidence and silently permits double counting.

Done-when: current — preventing duplicate attribution and naming exclusions/refusals still covers the revised design.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The excluded-brain pass clears ProtectedIntervals; its provenance regression passes.
  - id: BR-2
    disposition: addressed
    note: |
      creditedIntervals intersects positive allocations with counted Build/Converge phases; Tail and boundary-crossing regressions pass.
  - id: BR-3
    disposition: addressed
    note: |
      Proven different-slot ownership skips missing replay authority, including the no-completion legacy fixture; regressions pass.
  - id: BR-4
    disposition: addressed
    note: |
      The four/eight unrelated-adoption fixture passes without replaying unavailable sources while preserving same-slot protection.
  - id: BR-5
    disposition: not-addressed
    note: |
      Empty, absent, and mixed ownership are covered, but actualexclusion.go:154–177 treats malformed nonempty workspace identity as proven different. In a disposable pinned-head checkout, changing only #649's workspace to ariadne:06 makes #636 measure 0.333333h with no exclusions. A refusal regression fails on head and passes after requiring canonical parsed slot identity. This remains the same overlap-relevance-before-evidence rule failure.
```

**1. Strengths**

- Pure interval subtraction unions overlapping credit and preserves other-slot activity.
- Completion/correction snapshot selection is shared between historical replay and exclusion discovery.
- Tests exercise the nine decision-table cells, queue ordering, saved prefixes, counted phases, and unrelated-card filtering.

**2. Critical findings**

**BR-5 remains open — validate identity before proving slot disjointness.** At [actualexclusion.go:154](/Users/xianxu/Library/Caches/sdlc/review-checkouts/16caa765eb6cea21/cmd/sdlc/actualexclusion.go:154), only an empty workspace produces `slotUnknown`. Any nonmatching, nonempty string eventually produces `slotDifferent`, which the loader skips.

Both claimant parsing and engagement validation accept `ariadne:06`, although `workspace.ParseAddress` rejects it. The resulting numeric measurement silently recounts #649’s adopted interval.

This continues the repeated family `overlap-relevance-before-evidence` (**ARCH-PURPOSE, ARCH-SECURE**). Complete the proof rule rather than adding another string exception: require valid canonical addresses before comparing slots; otherwise preserve unknown identity and refuse potentially overlapping adoption. Sweep malformed, noncanonical, incomplete, and mixed valid/invalid ownership with regressions.

**3. Important findings**

None.

**4. Minor findings**

None.

**5. Test coverage notes**

- Broader actual-time selection passed across `cmd/sdlc`, `internal/activetime`, and `internal/gatestate`; the command package completed in 190.048 seconds.
- Removing the empty-workspace guard makes the existing BR-5 regression fail.
- The additional malformed-workspace refusal regression fails on pinned head and passes with canonical slot validation.
- Disposable checkouts were removed; the reviewed checkout remains unchanged. `git diff --check` passed.

**6. Architectural notes**

- **ARCH-DRY: pass** — shared snapshot authority selection.
- **ARCH-PURE: pass** — subtraction and classification have direct pure tests.
- **ARCH-PURPOSE: flag** — unknown identity can still bypass exclusivity.
- **ARCH-MOCK: pass** — existing IO boundaries and isolated portable tracker fixtures are reused.
- **ARCH-CONSTRAINTS: pass** — unrelated-card filtering and snapshot bounds are exercised.
- **ARCH-SECURE: flag** — persisted workspace strings become disjointness evidence without address validation.
- **ARCH-ORDER: pass** — production relevance decisions use the explicit state classifier.
- **ARCH-FUNERAL: pass** — exclusions reuse bounded completion snapshots and their existing lifecycle.

**7. Plan revision recommendations**

Add a `## Revisions` entry defining “proven different” as comparison of validated canonical slot identities. Enumerate invalid/noncanonical identity as unknown, including mixed ownership, and require the malformed-workspace refusal regression before the next review.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8966d7ed-836f-417b-ace2-6fac9be315af","repository":"github.com/xianxu/ariadne","issue_id":"000443","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/16caa765eb6cea21","reviewed_head":"27b80215792fd4d22d3316830c11f88bbcabf1b5","slot":"ariadne:1","requested_at":"2026-10-19T02:24:00.304988Z"},"phase":"finished","pid":67593,"started_at":"2026-10-19T02:24:00.45813Z","finished_at":"2026-10-19T02:29:57.291668Z","session_id":"01a151f9-14d7-77e2-8c9a-02399a906eea","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T20:37:06-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 443 — sdlc actual: a parked claim double-counts time the slot spent on other issues |
| repo | ariadne |
| issue file | workshop/issues/000443-sdlc-actual-a-parked-claim-double-counts-time-the-slot-spent-on-other-issues.md |
| boundary | whole-issue close |
| milestone | — |
| window | f92193cf4afc9d2b302e132ec5a639bbe3e05098..7502fb9e0bd1e8254cd6be369bf64b82215a6b96 |
| command | sdlc close --issue 443 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-18T20:37:06-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The exclusion core, shared measurement wiring, and claimant-slot correction are validated. BR-5 remains blocking: malformed slot identity in a retained measurement source still silently bypasses exclusion. A disposable-checkout regression reproduced the double count and passed with a temporary canonical-slot guard.

Done-when: current — preventing duplicate attribution and naming exclusions/refusals still covers the revised implementation.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Diagnostic computation clears protected intervals; the excluded-brain regression passes.
  - id: BR-2
    disposition: addressed
    note: |
      Credit intersects counted Build/Converge phases; tail and boundary-crossing regressions pass.
  - id: BR-3
    disposition: addressed
    note: |
      Proven different slots skip missing adoption evidence; the parked-claim legacy cases pass.
  - id: BR-4
    disposition: addressed
    note: |
      The passing integration fixture exercises four and eight unrelated adoptions with unavailable replay evidence while retaining same-slot protection.
  - id: BR-5
    disposition: not-addressed
    note: |
      ARCH-PURPOSE, ARCH-SECURE: canonical validation covers claimant workspaces but not retained source slots. actualexclusion.go:42–46 accepts any nonempty source slot, then lines 359–367 silently discard unmatched credit. MeasurementInputs.Validate does not validate Sources[].Slot. In a disposable pinned-head checkout, changing only #649's saved source slot to ariadne:06 made #636 report 0.333333h, equal to baseline, with no exclusions. A refusal regression failed on head and passed after canonical validation in creditedIntervals. This continues BR-5 in repeated family overlap-relevance-before-evidence: enforce the proof rule across every slot-identity consumer, rather than fixing only claimant identity.
```

1. **Strengths**

   - Pure interval subtraction unions overlapping protection and runs before attribution.
   - Completion/correction snapshot selection is shared with historical replay.
   - The nine-cell classifier explicitly preserves unknown relevance.
   - Protected intervals survive snapshot serialization and historical replay.

2. **Critical findings**

   - **BR-5 remains open:** `cmd/sdlc/actualexclusion.go:42` and `:360`. Validate retained source identity before using string mismatch to discard credited intervals. Sweep candidate, claimant, replay-source, and protected-interval identities; preserve legacy missing-pin compatibility through explicit unknown/refusal handling.

3. **Important findings**

   None additional. Atlas documents the changed flow; no new command, flag, or configuration surface requires a README update.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed full `activetime` and `gatestate` package tests, focused adopted/credited/parked tests, and diagnostic/queue replay regressions. Removing claimant canonical validation made its regressions fail. The additional retained-source reproduction failed without its guard and passed with it. Disposable checkout removed; reviewed checkout unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared snapshot authority and subtraction helpers.
   - **ARCH-PURE — pass:** interval logic remains pure; tracker/replay IO stays in the loader.
   - **ARCH-PURPOSE — flag:** BR-5 still permits duplicate attribution.
   - **ARCH-MOCK — pass:** existing dependency seams and portable tracker/Git fixtures reused.
   - **ARCH-CONSTRAINTS — pass:** unrelated-card filtering exercised at N and 2N.
   - **ARCH-SECURE — flag:** persisted source-slot identity remains unvalidated.
   - **ARCH-ORDER — pass:** explicit relevance decisions and pinned historical replay.
   - **ARCH-FUNERAL — pass:** exclusions extend the existing size-bounded snapshot.

7. **Plan revision recommendations**

   Add a `## Revisions` entry enumerating every slot-identity provenance and requiring canonical proof before any mismatch can justify skipping credit. Include retained-source malformed, incomplete, and missing-identity regressions alongside the claimant matrix.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"17154650-041e-42ca-84de-e054357d0e8b","repository":"github.com/xianxu/ariadne","issue_id":"000443","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0734785b0a2e452f","reviewed_head":"7502fb9e0bd1e8254cd6be369bf64b82215a6b96","slot":"ariadne:1","requested_at":"2026-10-19T03:33:35.727853Z"},"phase":"finished","pid":72385,"started_at":"2026-10-19T03:33:35.995503Z","finished_at":"2026-10-19T03:37:05.560272Z","session_id":"01a15238-cc50-7d51-93aa-56720fe31b8b","outcome":"returned"}] -->
