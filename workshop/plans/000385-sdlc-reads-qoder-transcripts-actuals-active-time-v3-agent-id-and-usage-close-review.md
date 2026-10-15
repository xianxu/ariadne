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
