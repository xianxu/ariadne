# Boundary Review — ariadne#385 (whole-issue close)

| field | value |
|-------|-------|
| issue | 385 — sdlc reads Qoder transcripts: actuals (active-time-v3), agent-id and usage |
| repo | ariadne |
| issue file | workshop/issues/000385-sdlc-reads-qoder-transcripts-actuals-active-time-v3-agent-id-and-usage.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9d03c5735e15a06f70e552dfcbcdd21f4d52e343..d4c9d29701d73e39a2718e22398b484d4c694460 |
| command | sdlc close --issue 385 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T11:45:12-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation delivers Qoder actuals/close adoption, invoking identity, and exact native-credit reporting through the existing attribution paths. Targeted tests pass. Two cheap fixes remain: surface unsupported counter semantics and update README for the new user-facing contract.

```findings
findings:
  - id: new
    severity: Important
    family: unsupported-format-visibility
    title: |
      Qoder nonzero token counters lack the promised format diagnostic
    detail: |
      cmd/sdlc/internal/usage/qoder.go:117 ignores the parsed token fields. A temporary overlay test with input_tokens=123 and output_tokens=456 returned null tokens but only sources/records coverage and spend_complete=true. The plan at lines 137–140 explicitly requires a visible diagnostic for unrecognized counter semantics. Validate all four placeholder fields, retain null tokens, and add regression coverage for nonzero and invalid values (ARCH-SECURE, ARCH-PURPOSE).
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README omits Qoder usage and links to a removed atlas anchor
    detail: |
      README.md:130–145 still describes Claude/Codex tokens only, omitting Qoder credit reporting, calibration support, and schema-version-2 nullable measurements. Its local-token-usage-365 link no longer matches the renamed atlas heading. Update this passage and link in the same boundary, as required by the README docs gate (ARCH-PURPOSE).
```

1. **Strengths**

   - Shared `qodertranscript.SessionState` centralizes sticky identity/workspace conflicts and timestamp-based configuration selection.
   - Exact decimal credits preserve measured precision without token conversion or double-counting `original_credits`.
   - `actual_qoder_test.go` exercises production ownership, reclaim, cross-checkout measurement, close adoption, retained replay, and truncation refusal.
   - Identity tests verify launch binding beats recency, competing sessions remain unknown, and contradictory middle metadata cannot escape detection.

2. **Critical findings**

   None.

3. **Important findings**

   Both are recorded above with locations and fix sketches.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed all four changed internal package suites, targeted root Qoder/output/cancellation/attribution tests, reviewer-resolution tests, and `git diff --check`. The temporary counter-format regression probe failed as described above. Checkout remained unchanged. I did not rerun the full repository suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared metadata interpretation, normalization, and ownership paths.
   - **ARCH-PURE — pass:** deterministic metadata/arithmetic cores with portable reader and filesystem seams.
   - **ARCH-PURPOSE — flag:** diagnostic and README gaps above; all three requested consumers are wired.
   - **ARCH-MOCK — pass:** portable transcript stores and temporary Git fixtures exercise production boundaries.
   - **ARCH-CONSTRAINTS — pass:** explicit candidate, line, numeric, identity-read, and cancellation bounds.
   - **ARCH-SECURE — flag:** unsupported token semantics silently accepted; otherwise typed parsing and content-free diagnostics are sound.
   - **ARCH-ORDER — pass:** sticky conflicts and replay/config ordering have sequence coverage.
   - **ARCH-FUNERAL — pass:** no new persistent cache family; existing meter replacement and actual-evidence lifecycles are reused.

7. **Plan revision recommendations**

   No scope revision needed. Implement the already-promised format diagnostic and complete the README documentation gate.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f19d47bd-d5af-41fd-8399-853500dce3c4","repository":"github.com/xianxu/ariadne","issue_id":"000385","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/1f99f510016bcc98","reviewed_head":"d4c9d29701d73e39a2718e22398b484d4c694460","slot":"ariadne:3","requested_at":"2026-10-15T18:42:18.468641Z"},"phase":"finished","pid":38476,"started_at":"2026-10-15T18:42:18.70588Z","finished_at":"2026-10-15T18:45:11.374353Z","session_id":"01a140df-4f76-7342-8669-a30a2b5886ed","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T12:25:59-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 385 — sdlc reads Qoder transcripts: actuals (active-time-v3), agent-id and usage |
| repo | ariadne |
| issue file | workshop/issues/000385-sdlc-reads-qoder-transcripts-actuals-active-time-v3-agent-id-and-usage.md |
| boundary | whole-issue close |
| milestone | — |
| window | ea2cee461c058080a25b03e689e9142c3ea21a63..0ffe2ebd220d7aca981c81bbad63697e45dc9b2f |
| command | sdlc close --issue 385 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T12:25:59-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Both prior findings are addressed and the targeted suites pass. One correctness blocker remains: malformed input before workspace metadata lets actual discovery omit an owned Qoder transcript, then close adopts the remaining partial measurement.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      qoder.go checks all four placeholders, including streaming rows. Regression tests cover nonzero and invalid values; removing the check through a temporary overlay makes the placeholder and streaming tests fail.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:130–149 now documents Qoder credits, calibration and schema-version-2 nullable measurements, and links to the matching Local usage heading in atlas/workflow/sdlc-binary.md. The collector and calibration implementation support the documented behavior.
findings:
  - id: new
    severity: Critical
    family: incomplete-evidence-adoption
    title: |
      Malformed prefixes allow close to adopt incomplete Qoder actuals
    detail: |
      cmd/sdlc/internal/transcripts/qoder.go:119–121 returns on a malformed row before later workspace metadata can establish relevance. discovery.go:54–59 then treats the failed source as skippable. A temporary extension of TestQoderActualDiscoveryRefusesPartialOwnedEvidence with "not-json\n" before valid owned rows returned the good source with no error. A production actual/close overlay confirmed actualMeasured, 0.1 hours and adopted=true despite that omitted owned source. Preserve uncertainty until exclusion is proven: continue bounded relevance discovery after recoverable parse failures, and refuse automatic adoption when scan failures or limits leave ownership unresolved. Sweep parse, read and bound failures before ownership resolution, with discovery and close regressions (ARCH-SECURE, ARCH-ORDER, ARCH-PURPOSE).
```

1. **Strengths**

   - Shared typed metadata and sticky session/config conflicts avoid divergent provider interpretations.
   - Exact credit arithmetic and nullable measurements preserve native amounts without fabricating token usage.
   - Identity tests validate launch binding, competing sessions, contradictory copies and cancellation.
   - README, help and atlas describe the new reporting surface and schema change.

2. **Critical findings**

   - The malformed-prefix adoption defect above must be fixed before crossing the boundary.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   All four changed internal package suites passed. Targeted command tests covering Qoder, output, cancellation, attribution and commit stamping passed. Removing BR-1’s check made its regressions fail. Two temporary overlays reproduced the new defect through discovery and production close adoption. Tracked files remain unchanged; the full suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared adapter, normalization and measurement aggregation.
   - **ARCH-PURE — pass:** deterministic metadata/arithmetic core with filesystem boundaries.
   - **ARCH-PURPOSE — flag:** incomplete evidence can enter actuals and velocity calibration.
   - **ARCH-MOCK — pass:** portable transcript stores and temporary Git fixtures exercise production seams.
   - **ARCH-CONSTRAINTS — pass:** explicit scan, numeric and cancellation bounds.
   - **ARCH-SECURE — flag:** unknown ownership after parsing failure becomes safe exclusion without evidence.
   - **ARCH-ORDER — flag:** failure before metadata resolution permits successful partial adoption.
   - **ARCH-FUNERAL — pass:** existing meter replacement and retained-evidence lifecycles are reused.

7. **Plan revision recommendations**

   Add a `## Revisions` entry defining unresolved, proven-unrelated and relevant failed-source outcomes. Require automatic actuals refusal while ownership remains unresolved, and extend the failure matrix to errors before workspace metadata.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c6262c8f-db3d-4050-afb1-07d45375f688","repository":"github.com/xianxu/ariadne","issue_id":"000385","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a9751bcd78f9e3a6","reviewed_head":"0ffe2ebd220d7aca981c81bbad63697e45dc9b2f","slot":"ariadne:3","requested_at":"2026-10-15T19:23:40.725396Z"},"phase":"finished","pid":42027,"started_at":"2026-10-15T19:23:40.827993Z","finished_at":"2026-10-15T19:25:58.415464Z","session_id":"01a14105-2e87-7582-99cf-85f33824b443","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T13:07:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 385 — sdlc reads Qoder transcripts: actuals (active-time-v3), agent-id and usage |
| repo | ariadne |
| issue file | workshop/issues/000385-sdlc-reads-qoder-transcripts-actuals-active-time-v3-agent-id-and-usage.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8e637c6e2dd431b4c82c42c2dea680f93918c785..c907a4d0ba566153263b2334826a9eeb5779f1dd |
| command | sdlc close --issue 385 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T13:07:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation satisfies the issue’s revised scope across actuals, close adoption, identity, and native-credit usage reporting. BR-3 is addressed across tracked and legacy selection paths, with regression tests that fail when the former behavior is restored. No new blocking findings.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Failed scans retain unresolved ownership unless complete evidence proves exclusion. Production tracked and legacy actual/close regressions pass. Restoring the former early parse return and error-selection rule in a disposable copy makes the discovery failure matrix and unreadable-candidate regression fail; the copy was removed.
```

1. **Strengths**

   - Shared typed parsing and encapsulated session transitions keep metadata interpretation consistent across all three consumers.
   - Exact credit arithmetic preserves native precision; nullable measurements and labelled subtotals avoid fabricated token totals.
   - Ownership uncertainty survives parse, read, bound, and pathname-replacement failures in `cmd/sdlc/internal/transcripts/qoder.go:78`.
   - Legacy actuals propagate selection errors before adoption and retain exclusion warnings in `cmd/sdlc/actual.go:156`.
   - README and atlas document Qoder support, schema version 2, identity binding, and the changed selection API.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - All four changed internal package suites passed.
   - Targeted production actual/close, Qoder usage, native-output, and caller-context tests passed in 25.752 seconds.
   - Removing the central BR-3 correction reproduced partial-source admission in the new regression matrix.
   - `git diff --check` passed; the reviewed checkout remains unchanged.
   - The complete repository suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared adapter and strict discovery path serve the consumers.
   - **ARCH-PURE — pass:** parsing, arithmetic, normalization, and state transitions have direct tests; filesystem work stays at injected boundaries.
   - **ARCH-PURPOSE — pass:** all requested consumers are delivered; the BR-3 sweep includes legacy actuals.
   - **ARCH-MOCK — pass:** portable transcript stores and temporary Git fixtures exercise production seams without provider calls.
   - **ARCH-CONSTRAINTS — pass:** scan limits, numeric bounds, and caller cancellation have regression coverage.
   - **ARCH-SECURE — pass:** malformed evidence remains visible and cannot silently justify exclusion.
   - **ARCH-ORDER — pass:** sticky conflicts, replay, retained prefixes, and deterministic replacement tests cover relevant sequencing.
   - **ARCH-FUNERAL — pass:** existing bounded meter and snapshot lifecycles are reused; no new persistent artifact family.

   The core-concept inventory matches the delivered implementation when read with its explicit reconciliation revisions.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"deb475f4-837f-428e-b764-9569698eb952","repository":"github.com/xianxu/ariadne","issue_id":"000385","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/201d08d1f0de78ef","reviewed_head":"c907a4d0ba566153263b2334826a9eeb5779f1dd","slot":"ariadne:3","requested_at":"2026-10-15T20:04:59.847852Z"},"phase":"finished","pid":9588,"started_at":"2026-10-15T20:04:59.956079Z","finished_at":"2026-10-15T20:07:23.068597Z","session_id":"01a1412b-029d-7fc0-aa89-d0362e8e17f2","outcome":"returned"}] -->
