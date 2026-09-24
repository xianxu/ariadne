# Boundary Review — ariadne#248 (whole-issue close)

| field | value |
|-------|-------|
| issue | 248 — Agent procedure: move this branch to :N |
| repo | ariadne |
| issue file | workshop/issues/000248-move-branch-to-slot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 36991fc50b25377a0faa50397a498c38cda6dfc7..90abb11c90e299c01d26e33ab97084a60884afa4 |
| command | sdlc close --issue 248 |
| reviewer | codex |
| timestamp | 2026-09-23T23:55:14-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The diff improves slot-move discoverability and adds real Git fixture coverage, but it does not fully satisfy the Spec: upstream resting-history comparison is only described, not performed, and the required Pair post-move build declaration is absent from the reviewed boundary.

1. Strengths:

- Reuses the existing switching preflight and workspace identity procedure.
- Preserves noncolliding scratch files and protects collisions with `--no-overwrite-ignore`.
- Keeps resting refs unchanged and explicitly documents parked commits.
- Adds real Git-based tests for successful movement and untracked-file collision.
- Updates both atlas discoverability and `AGENTS.base.md`.

2. Critical findings:

- `atlas/workflow/workspace-branching.md:101-115` — The procedure compares `destination_rest` only against `issue_branch`; it never compares resting commits against the configured upstream as required by the Spec. “Can help identify” is not an executable comparison. Enumerated family instances: the move procedure’s history check at lines 101-115 and its corresponding fixture at `cmd/sdlc/workspace_procedure_test.go:191-232`, which tests only local-rest-vs-feature. Add an explicit upstream comparison and regression coverage.

3. Important findings:

- `workshop/issues/000248-move-branch-to-slot.md` Plan/Done when — The required Pair `AGENTS.local.md` declaration for `make build` in `:0` is not delivered in this ariadne range. The current Pair file has no post-move build declaration. Either include the coordinated Pair change in the reviewed work or revise the issue/plan to document the separate required boundary and evidence.

4. Minor findings:

- None.

5. Test coverage notes:

`go test ./cmd/sdlc -run 'TestWorkspaceProcedureMove|TestWorkspaceProcedureReadinessAndCollisions' -count=1` passed. Coverage does not exercise upstream-history comparison, post-move build verification, reverse movement, or ancestor-path collision handling.

6. Architectural notes:

- ARCH-DRY: Pass; the procedure links and reuses the existing preflight.
- ARCH-PURE: Pass; no runtime implementation was added, and tests use real Git fixtures.
- ARCH-PURPOSE: Flagged above; the committed procedure under-delivers the required ordering check and Pair build declaration.

```findings
findings:
  - id: new
    severity: Critical
    family: resting-history-ordering
    title: |
      Move procedure does not compare destination rest against its configured upstream
    detail: |
      atlas/workflow/workspace-branching.md:101-115 only runs "$issue_branch..$destination_rest"; the Spec requires comparison with both the feature branch and its configured upstream. The same family appears in the move fixture at cmd/sdlc/workspace_procedure_test.go:191-232, which validates only local rest versus feature. Add the upstream comparison and a regression test.
  - id: new
    severity: Important
    family: peer-post-move-declaration
    title: |
      Required Pair post-move build declaration is missing
    detail: |
      The issue's Plan and Done when require Pair's AGENTS.local.md to declare make build in :0, but this ariadne boundary contains no such declaration and the current Pair file has none. Include the coordinated Pair change or revise the plan to identify the separate required boundary and evidence.
```

---

## Re-review — 2026-09-24T00:01:09-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 248 — Agent procedure: move this branch to :N |
| repo | ariadne |
| issue file | workshop/issues/000248-move-branch-to-slot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 36991fc50b25377a0faa50397a498c38cda6dfc7..b87d89f71bf277674353b6c5f1bb2def407ae800 |
| command | sdlc close --issue 248 |
| reviewer | codex |
| timestamp | 2026-09-24T00:01:09-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The documentation and fixture changes are focused and the targeted tests pass, but BR-1 is not fully addressed: the test repeats the intended Git commands rather than proving the documented procedure contains and executes the upstream comparison. The boundary should be reworked with regression evidence tied to the procedure text.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The procedure now includes the upstream comparison at atlas/workflow/workspace-branching.md:105-107, but the test at cmd/sdlc/workspace_procedure_test.go:212-217 independently repeats that command; removing the documented upstream comparison would still leave the test green.
  - id: BR-2
    disposition: addressed
    note: |
      The issue revision identifies Pair's separate pair#318 boundary and records its commit and required close/merge evidence.
```

1. Strengths

- Clear slot-move procedure with identity, readiness, collision, and rollback safeguards.
- Real-Git fixtures cover parked commits, scratch preservation, and collision refusal.
- Base-layer discoverability and atlas index were updated.
- `go test ./cmd/sdlc -run '^TestWorkspaceProcedure' -count=1` passes.
- ARCH-DRY, ARCH-PURE, and ARCH-PURPOSE pass for this focused documentation-plus-fixture change.

2. Critical findings

- `cmd/sdlc/workspace_procedure_test.go:212-217`: tie the regression test to the documented procedure or otherwise make removal of the upstream comparison fail. The current test only proves a separately duplicated command works.

3. Important findings

None new.

4. Minor findings

None.

5. Test coverage notes

The targeted procedure suite passed in 53.462s. It does not exercise reverse movement or ancestor-path collisions, but those are not blocking findings for this boundary.

6. Architectural notes

- ARCH-DRY: Pass; the procedure reuses the existing preflight.
- ARCH-PURE: Pass; no production IO/business logic was introduced, and tests use real Git fixtures.
- ARCH-PURPOSE: Mostly pass; the Ariadne-side purpose is documented, while Pair’s build declaration is correctly delegated to its named Pair boundary.

7. Plan revision recommendations

Add a `## Revisions` entry recording that the upstream-history regression evidence must be connected to the procedure source, not merely duplicate its commands in the test.

---

## Re-review — 2026-09-24T00:04:26-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 248 — Agent procedure: move this branch to :N |
| repo | ariadne |
| issue file | workshop/issues/000248-move-branch-to-slot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 36991fc50b25377a0faa50397a498c38cda6dfc7..dab1c457f5eff1a3c3880aa898379e064eb8db73 |
| command | sdlc close --issue 248 |
| reviewer | codex |
| timestamp | 2026-09-24T00:04:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation correctly adds the move procedure, upstream comparison, real-Git regression coverage, and base-layer discoverability. One Important documentation gap remains: README does not identify the new move procedure.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The guide now compares destination rest against both the feature branch and configured upstream at atlas/workflow/workspace-branching.md:101-107, and the real-Git fixture verifies the parked commit appears in both results at cmd/sdlc/workspace_procedure_test.go:214-233.
findings:
  - id: new
    severity: Important
    family: docs-discoverability
    title: |
      README does not discover the new branch-move procedure
    detail: |
      The new user-runnable “move this branch to :N” surface is absent from README.md; its existing guidance at README.md:183-187 mentions only branching and refreshing. This is the only newly introduced user-runnable procedure in this review window. Add the move phrase and link to the procedure, updating the stale link title.
```

1. Strengths

- The guide explicitly performs both required history comparisons.
- The fixture consumes the authored guide commands rather than duplicating them.
- Real-Git tests cover parked commits, preserved scratch files, and collision refusal.
- `AGENTS.base.md` and `atlas/index.md` provide discoverability.

2. Critical findings

None.

3. Important findings

See `docs-discoverability` above.

4. Minor findings

None.

5. Test coverage notes

`go test ./cmd/sdlc -run '^TestWorkspaceProcedure' -count=1` passed. `git diff --check` passed.

6. Architectural notes

- ARCH-DRY: Pass — the procedure reuses the existing switching preflight.
- ARCH-PURE: Pass — this is documentation plus integration-style real-Git tests; no business logic is mixed into a new runtime path.
- ARCH-PURPOSE: Pass — the guide and fixture cover the stated move behavior, including both history comparisons.

7. Plan revision recommendations

None.

---

## Re-review — 2026-09-24T00:07:09-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 248 — Agent procedure: move this branch to :N |
| repo | ariadne |
| issue file | workshop/issues/000248-move-branch-to-slot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 36991fc50b25377a0faa50397a498c38cda6dfc7..82ca44da4f7b6464fbcc7d67658cbf22741f1647 |
| command | sdlc close --issue 248 |
| reviewer | codex |
| timestamp | 2026-09-24T00:07:09-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range is valid and fulfills the issue’s Spec/Plan. The move procedure, README/atlas discoverability, agent guidance, and real-Git regression fixtures are present; targeted and package tests pass. No new blocking findings.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      README.md:183-186 now names “move this branch to :0” and links the renamed shared procedure.
  - id: BR-1
    disposition: addressed
    note: |
      The procedure explicitly compares destination rest with both the feature branch and configured upstream; the real-Git test executes both commands.
  - id: BR-2
    disposition: addressed
    note: |
      The issue records Pair's post-move build declaration under the separate Pair-owned pair#318 boundary; this Ariadne range introduces no contradictory behavior.
```

### Strengths

1. The procedure resolves both slots through `sdlc workspace` and preserves identity/HEAD safety checks.
2. Destination scratch-file preservation and collision refusal are documented and exercised with real Git.
3. Resting-history comparisons are sourced directly from the authored guide, preventing test/procedure drift.
4. README and atlas discoverability were updated alongside the new operator surface.

### Critical findings

None.

### Important findings

None.

### Minor findings

None.

### Test coverage notes

`go test ./cmd/sdlc -run '^TestWorkspaceProcedure' -count=1` and `go test ./cmd/sdlc` pass. `git diff --check` reports no whitespace errors.

### Architectural notes

- ARCH-DRY: Pass. The procedure reuses the existing switching preflight.
- ARCH-PURE: Pass. No runtime behavior or IO abstraction was added; tests use real Git at the integration boundary.
- ARCH-PURPOSE: Pass. The guide, discoverability surfaces, history checks, collision behavior, and regression fixtures cover the stated purpose.

### Plan revision recommendations

None.

---

## Re-review — 2026-09-24T00:19:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 248 — Agent procedure: move this branch to :N |
| repo | ariadne |
| issue file | workshop/issues/000248-move-branch-to-slot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 36991fc50b25377a0faa50397a498c38cda6dfc7..347d8c201781f30fb4335686f0e222fe5ab00533 |
| command | sdlc close --issue 248 |
| reviewer | codex |
| timestamp | 2026-09-24T00:19:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range fulfills the Spec and Plan. The procedure, discoverability updates, real-Git fixtures, and prior finding fixes are evidenced; no new blocking findings.

1. Strengths

- Explicit source/destination identity and HEAD validation.
- Correct untracked-file preservation and collision refusal.
- History checks are sourced directly from the authored guide and test both comparisons.
- README, AGENTS, and atlas discoverability are updated.
- Focused real-Git tests pass; `git diff --check` passes.

2. Critical findings

None.

3. Important findings

None.

4. Minor findings

None.

5. Test coverage notes

`go test ./cmd/sdlc -run '^TestWorkspaceProcedure' -count=1` passed. The full package suite was stopped after several minutes without output; no conclusion was drawn from it.

6. Architectural notes

- ARCH-DRY: Pass — reuses existing switching preflight.
- ARCH-PURE: Pass — no new runtime logic; tests use real Git fixtures.
- ARCH-PURPOSE: Pass — documented procedure, consumers, history checks, and fixtures cover the stated goal.

7. Plan revision recommendations

None.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The guide compares destination rest with both the feature and configured upstream, and the fixture executes the authored command block and observes the parked commit in both outputs.
  - id: BR-2
    disposition: addressed
    note: |
      The issue records the Pair-owned pair#318 boundary and its post-move build declaration; this Ariadne range introduces no contradictory behavior.
  - id: BR-3
    disposition: addressed
    note: |
      README.md names “move this branch to :0” and links the renamed shared procedure.
```
