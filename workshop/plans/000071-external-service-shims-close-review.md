# Boundary Review — ariadne#71 (whole-issue close)

| field | value |
|-------|-------|
| issue | 71 — construct a testable shim for every external service (gh/github first): the deterministic-shell mock pattern |
| repo | ariadne |
| issue file | workshop/issues/000071-external-service-shims.md |
| boundary | whole-issue close |
| milestone | — |
| window | 041f3293ca7e27f4843c21d9f27d4e6fd513021e..HEAD |
| command | sdlc close --issue 71 |
| reviewer | codex |
| timestamp | 2026-07-26T23:05:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Summary

`ARCH-MOCK` satisfies the narrowed ariadne promotion scope: the principle is in
the single-source architecture registry, rendered through `sdlc arch-principles`,
propagated into prompt marker extraction/goldens, and reflected in active agent
guidance.

## Findings

- **Important:** The generated close-review sidecar originally persisted the full
  raw reviewer session, including harness metadata and the entire prompt/diff.
  This violates the repo lesson that generated review sidecars must be bounded
  and normalized. Fix before shipping by replacing the sidecar with this compact
  record.

No Critical findings.

## Verification

- `go test ./cmd/sdlc/internal/judge ./cmd/sdlc -count=1`
- `go run ./cmd/sdlc arch-principles --lens at-review`
- `git diff --check 041f3293ca7e27f4843c21d9f27d4e6fd513021e..HEAD`

## Resolution

Condensed this sidecar to the durable review facts: verdict, window, finding,
verification commands, and resolution. The implementation itself needed no code
changes after review.
