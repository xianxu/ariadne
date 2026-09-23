# Boundary Review — ariadne#245 (whole-issue close)

| field | value |
|-------|-------|
| issue | 245 — Slots v2: branch from a workspace and refresh |
| repo | ariadne |
| issue file | workshop/issues/000245-slots-v2-branch-and-refresh.md |
| boundary | whole-issue close |
| milestone | — |
| window | e3d7b4b8920f62d6075e7953e6fc724e7aacabeb..16a0d498ba6028f31e7b125c71f78cda47df9af0 |
| command | sdlc close --issue 245 |
| reviewer | codex |
| timestamp | 2026-09-23T15:13:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation and real-Git fixtures are strong and focused. No Critical issues were found. Two Important gaps remain: README discoverability for the new procedure, and incomplete regression coverage for preserving the source checkout state.

1. Strengths

- Real Git fixtures exercise branching, refresh, collisions, dirty states, ancestry, and sibling isolation.
- Procedures correctly pin full SHAs and avoid automatic reset/stash/commit behavior.
- Atlas and embedded `change-code` guidance are linked and detailed.
- Focused tests passed:
  - `go test ./cmd/sdlc -run 'TestWorkspaceProcedure' -count=1`
  - `go test ./pkg/workspace/... -count=1`
  - related workspace/change-code tests
- `git diff --check` passed.

2. Critical findings

None.

3. Important findings

- `README.md:134-156` — README update appears missing for the new user-runnable branch-from-slot and explicit-refresh procedures. The atlas and embedded help document them, but the top-level README only documents workspace identity. Add a concise link or usage summary. Family: `user-facing-surface-discoverability`.
- `cmd/sdlc/workspace_procedure_test.go:100-114` — the branch fixture does not verify that the source checkout’s branch, HEAD, resting ref, and upstream remain unchanged; it explicitly skips the shared `source-feature` ref and never snapshots source state. Add source-state assertions. This is the only source-preservation fixture in the window. Family: `source-state-preservation-coverage`.

4. Minor findings

None.

5. Test coverage notes

The focused suites pass and exercise the documented Git commands against real repositories. Capture movement is tested deterministically. Source-preservation assertions should be added before closing.

6. Architectural notes

- ARCH-DRY: pass — reuses the existing workspace resolver, issue workflow, and ordinary Git operations.
- ARCH-PURE: pass — no new runtime mechanism; tests use real Git at the integration boundary.
- ARCH-PURPOSE: pass — the diff covers branching, refresh, provenance, safety refusals, and sibling isolation.

7. Plan revision recommendations

Add a `## Revisions` entry recording:

- README discoverability for the new branch/refresh procedure.
- Explicit source checkout/ref/upstream preservation assertions in the regression fixture.
