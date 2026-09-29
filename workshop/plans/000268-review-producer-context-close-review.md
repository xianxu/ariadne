# Boundary Review — ariadne#268 (whole-issue close)

| field | value |
|-------|-------|
| issue | 268 — Bind hosted review producer effects to activation context |
| repo | ariadne |
| issue file | workshop/issues/000268-review-producer-context.md |
| boundary | whole-issue close |
| milestone | — |
| window | d6ace419b5dd9a645a34af1a8a3ae4d5df7548dd..a46f2540fcf5e053ad4e4bd204dc84324b88d002 |
| command | sdlc close --issue 268 |
| reviewer | codex |
| timestamp | 2026-09-28T21:02:55-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The scoped producer instructions cover the intended envelope, immediate context validation, artifact preservation, and unchanged commit bodies. The recorded probes support the principal scenarios, but leave the legacy compatibility contract untested despite potentially conflicting instructions.

1. **Strengths**
   - The authoritative shared skill owns validation; the atlas points to it.
   - Context checks cover human rounds, agent rounds, and shipping.
   - Landed commit bodies remain verbatim, and HEAD advancement does not invalidate activation.

2. **Critical findings:** None.

3. **Important findings**
   - **Legacy compatibility lacks behavioral coverage** — [SKILL.md:242](/Users/xianxu/workspace/ariadne/construct/local/fix/SKILL.md:242). The unconditional “missing/malformed context … stop” rule precedes permission for uninterrupted legacy traffic at lines 249–251. Neither recorded probe exercises that exception. Test the entire family: confirmed uninterrupted legacy activation, restored activation, retargeted activation, and absent evidence of legacy status. Confirm the first remains usable and the others refuse unscoped traffic; clarify rule precedence if necessary.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Inspected both recorded probe outputs: envelope echo, branch mismatch, human/ship activation mismatch, and initial human-round authorization are represented.
   - `git diff --check` passes.
   - No corresponding legacy producer scenarios were found. Pair transport tests cannot establish how this producer prompt interprets its exception.

6. **Architecture**
   - **ARCH-DRY: pass** — shared authoritative instructions, referenced by the atlas.
   - **ARCH-PURE: pass** — no new implementation mixes business logic with IO.
   - **ARCH-PURPOSE: pass in scope** — all intended Git-effect paths are addressed; compatibility verification remains incomplete.
   - Atlas updated. No new user-facing CLI/configuration surface requires a README change.

7. **Plan revision recommendation**
   - Append a `## Revisions` entry adding the four legacy scenarios and record their inputs, expected decisions, and observed results.

```findings
findings:
  - id: new
    severity: Important
    family: legacy-compatibility-coverage
    title: |
      Exercise both sides of the legacy producer exception
    detail: |
      construct/local/fix/SKILL.md:242–251 requires stopping on missing context while allowing confirmed uninterrupted legacy traffic. The recorded probes exercise neither legacy acceptance nor rejection. Cover every instance of this family: uninterrupted legacy, restored activation, retargeted activation, and unknown legacy status. Verify acceptance only for the first and clarify the exception's precedence if the producer blocks it.
```
