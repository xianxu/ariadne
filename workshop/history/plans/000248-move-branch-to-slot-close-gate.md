---
gate: boundary-review
issue: 248
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-23T23:55:14-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Move procedure does not compare destination rest against its configured upstream
          detail: atlas/workflow/workspace-branching.md:101-115 only runs "$issue_branch..$destination_rest"; the Spec requires comparison with both the feature branch and its configured upstream. The same family appears in the move fixture at cmd/sdlc/workspace_procedure_test.go:191-232, which validates only local rest versus feature. Add the upstream comparison and a regression test.
          family: resting-history-ordering
          round: 1
        - id: BR-2
          severity: Important
          title: Required Pair post-move build declaration is missing
          detail: The issue's Plan and Done when require Pair's AGENTS.local.md to declare make build in :0, but this ariadne boundary contains no such declaration and the current Pair file has none. Include the coordinated Pair change or revise the plan to identify the separate required boundary and evidence.
          family: peer-post-move-declaration
          round: 1
      recipe: small-diff-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-24T00:01:09-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: The procedure now includes the upstream comparison at atlas/workflow/workspace-branching.md:105-107, but the test at cmd/sdlc/workspace_procedure_test.go:212-217 independently repeats that command; removing the documented upstream comparison would still leave the test green.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The issue revision identifies Pair's separate pair#318 boundary and records its commit and required close/merge evidence.
          round: 2
      recipe: small-diff-review
      blocked: true
    - "n": 3
      timestamp: "2026-09-24T00:04:26-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The guide now compares destination rest against both the feature branch and configured upstream at atlas/workflow/workspace-branching.md:101-107, and the real-Git fixture verifies the parked commit appears in both results at cmd/sdlc/workspace_procedure_test.go:214-233.
          round: 3
      findings:
        - id: BR-3
          severity: Important
          title: README does not discover the new branch-move procedure
          detail: The new user-runnable “move this branch to :N” surface is absent from README.md; its existing guidance at README.md:183-187 mentions only branching and refreshing. This is the only newly introduced user-runnable procedure in this review window. Add the move phrase and link to the procedure, updating the stale link title.
          family: docs-discoverability
          round: 3
      recipe: small-diff-review
      blocked: true
    - "n": 4
      timestamp: "2026-09-24T00:07:09-07:00"
      agent: codex
      dispose:
        - id: BR-3
          disposition: addressed
          note: README.md:183-186 now names “move this branch to :0” and links the renamed shared procedure.
          round: 4
        - id: BR-1
          disposition: addressed
          note: The procedure explicitly compares destination rest with both the feature branch and configured upstream; the real-Git test executes both commands.
          round: 4
        - id: BR-2
          disposition: addressed
          note: The issue records Pair's post-move build declaration under the separate Pair-owned pair#318 boundary; this Ariadne range introduces no contradictory behavior.
          round: 4
      recipe: small-diff-review
      blocked: false
    - "n": 5
      timestamp: "2026-09-24T00:19:23-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The guide compares destination rest with both the feature and configured upstream, and the fixture executes the authored command block and observes the parked commit in both outputs.
          round: 5
        - id: BR-2
          disposition: addressed
          note: The issue records the Pair-owned pair#318 boundary and its post-move build declaration; this Ariadne range introduces no contradictory behavior.
          round: 5
        - id: BR-3
          disposition: addressed
          note: README.md names “move this branch to :0” and links the renamed shared procedure.
          round: 5
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#248 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-23T23:55:14-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `resting-history-ordering` Move procedure does not compare destination rest against its configured upstream
  atlas/workflow/workspace-branching.md:101-115 only runs "$issue_branch..$destination_rest"; the Spec requires comparison with both the feature branch and its configured upstream. The same family appears in the move fixture at cmd/sdlc/workspace_procedure_test.go:191-232, which validates only local rest versus feature. Add the upstream comparison and a regression test.
- **BR-2** [Important] `peer-post-move-declaration` Required Pair post-move build declaration is missing
  The issue's Plan and Done when require Pair's AGENTS.local.md to declare make build in :0, but this ariadne boundary contains no such declaration and the current Pair file has none. Include the coordinated Pair change or revise the plan to identify the separate required boundary and evidence.

## Round 2 — 2026-09-24T00:01:09-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — The procedure now includes the upstream comparison at atlas/workflow/workspace-branching.md:105-107, but the test at cmd/sdlc/workspace_procedure_test.go:212-217 independently repeats that command; removing the documented upstream comparison would still leave the test green.
- BR-2 — addressed — The issue revision identifies Pair's separate pair#318 boundary and records its commit and required close/merge evidence.

## Round 3 — 2026-09-24T00:04:26-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The guide now compares destination rest against both the feature branch and configured upstream at atlas/workflow/workspace-branching.md:101-107, and the real-Git fixture verifies the parked commit appears in both results at cmd/sdlc/workspace_procedure_test.go:214-233.

### Raised

- **BR-3** [Important] `docs-discoverability` README does not discover the new branch-move procedure
  The new user-runnable “move this branch to :N” surface is absent from README.md; its existing guidance at README.md:183-187 mentions only branching and refreshing. This is the only newly introduced user-runnable procedure in this review window. Add the move phrase and link to the procedure, updating the stale link title.

## Round 4 — 2026-09-24T00:07:09-07:00 (codex) — passed

### Disposed

- BR-3 — addressed — README.md:183-186 now names “move this branch to :0” and links the renamed shared procedure.
- BR-1 — addressed — The procedure explicitly compares destination rest with both the feature branch and configured upstream; the real-Git test executes both commands.
- BR-2 — addressed — The issue records Pair's post-move build declaration under the separate Pair-owned pair#318 boundary; this Ariadne range introduces no contradictory behavior.

## Round 5 — 2026-09-24T00:19:23-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The guide compares destination rest with both the feature and configured upstream, and the fixture executes the authored command block and observes the parked commit in both outputs.
- BR-2 — addressed — The issue records the Pair-owned pair#318 boundary and its post-move build declaration; this Ariadne range introduces no contradictory behavior.
- BR-3 — addressed — README.md names “move this branch to :0” and links the renamed shared procedure.

## Open findings

(none — every finding has been disposed)
