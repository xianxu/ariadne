# Boundary Review — ariadne#247 (whole-issue close)

| field | value |
|-------|-------|
| issue | 247 — Explicitly refresh repositories and dependencies with Weave |
| repo | ariadne |
| issue file | workshop/issues/000247-slot-claim-dependency-refresh.md |
| boundary | whole-issue close |
| milestone | — |
| window | 76c4e4c84539b14c18428c604947a352bf5b8acf..f3ba40b881829ebfc4317fcae34ab1032c1453b5 |
| command | sdlc close --issue 247 |
| reviewer | codex |
| timestamp | 2026-09-23T23:23:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The pinned range fulfills the issue spec and plan. Refresh behavior, captured-target safety, revalidation, compile retry, lease reuse, documentation, and real-Git tests are present. No blocking findings.

1. Strengths

- All-repository preflight and captured SHA application are clearly implemented.
- Real Git fixtures cover races, partial progress, rebase conflicts, declaration drift, and compile retry.
- Refresh shares the existing compile/setup lease correctly.
- README and atlas document the new `weave refresh` surface.
- Error handling avoids exposing credentials and distinguishes uncertain Git outcomes.

2. Critical findings

None.

3. Important findings

None.

4. Minor findings

None.

5. Test coverage notes

Focused Weave, acquisition, refresh, and CLI suites passed. The landing effective-remote regression also passed. The full `cmd/sdlc` suite did not complete in this environment because an unrelated `test calibrate` subprocess failed outside a Git repository.

6. Architectural notes

- ARCH-DRY: Pass — shared acquisition, Git, discovery, and compile seams are reused.
- ARCH-PURE: Pass — refresh state/eligibility logic is separated from Git and CLI IO.
- ARCH-PURPOSE: Pass — host, substrate dependencies, captured targets, retries, and docs are covered.
- ARCH-MOCK: Pass — stateful real-Git fixtures exercise the shared Git seam.
- ARCH-CONSTRAINTS: Pass — layer, document, output, and timeout bounds are enforced.
- ARCH-SECURE: Pass — external Git data is framed, bounded, validated, and credential-safe.
- ARCH-ORDER: Pass — explicit phases, revalidation, uncertainty handling, and partial progress are modeled.
- ARCH-FUNERAL: Pass — no new durable refresh artifacts are created; standard Git rebase state remains operator-managed.

7. Plan revision recommendations

None.

```findings
findings: []
```
