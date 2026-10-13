---
id: 000349
status: done
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: 'a4e1920ffb60d2d57f478b9d8c9c0bf4cc2cbf29' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T14:18:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.15
---

# Pin agent-specific review fixtures instead of inheriting session defaults

## Problem

During #337's full suite, three unchanged baseline tests failed under the
session's PAIR_AGENT=codex. TestCloseRecordsTheVerdictBeforeAPostscript and
TestCloseWithoutAVerdictKeepsBothRuns supply Claude JSONL but leave the review
agent implicit. TestBoundaryReviewFencedArgvAndStamp expects Claude argv while
also inheriting the session default. Production selection correctly chose Codex.

## Spec

Pin the intended agent in agent-specific fixtures; audit analogous assertions.
Keep production agent resolution and its explicit precedence tests unchanged.

## Done when

- The three named tests pass under PAIR_AGENT=codex and PAIR_AGENT=claude, including conflicting AGENT_CMD, without requiring a suite-wide override.
- Agent-specific fixtures explicitly declare their agent; production resolution precedence remains covered.

## Plan

- [x] `pinReviewAgent(t, agent)` (`cmd/sdlc/agentpin_test.go`): sets `AGENT_CMD`, the
  operator override that outranks `PAIR_AGENT` and the reviewer policy, so it pins
  against any ambient `PAIR_AGENT` or `AGENT_CMD`; only `--agent` is above it. Pin the
  three named fixtures to `judge.AgentClaude`.
- [x] Audit: the whole suite under `PAIR_AGENT=codex`, and again under
  `AGENT_CMD=codex` (they catch different gaps: a fixture relying on the Claude default,
  and one that sets `PAIR_AGENT` but not `AGENT_CMD`); pin every further fixture either
  run exposes.
- [x] Production resolution unchanged: its precedence tests (`agent_resolver_test.go`,
  and the `t.Setenv("AGENT_CMD", "")` + `PAIR_AGENT` tests) stay as they are.

## Log

### 2026-10-13
- 2026-10-13: closed — Test-only change (no atlas surface): pinReviewAgent(t, agent) sets AGENT_CMD, which outranks PAIR_AGENT and the reviewer policy; the three named fixtures pin judge.AgentClaude. Reproduced first: all three failed under PAIR_AGENT=codex. After: they pass under PAIR_AGENT=codex, PAIR_AGENT=claude, PAIR_AGENT=claude AGENT_CMD=codex, PAIR_AGENT=codex AGENT_CMD=gemini. Audit: full make test under PAIR_AGENT=codex (16:19) and under AGENT_CMD=codex (16:29), 1120 cmd/sdlc tests + 47 packages each, both green; test-shard.py passes env through (no env= override). Production resolution and its precedence tests unchanged (diff is test files only).; review verdict: SHIP
- 2026-10-13: unclaimed: Independent baseline test-fixture issue discovered during #337; left unassigned for another slot.

Filed independently from #337. The source and fixtures match origin/main at
19dafa47ec755f12dd5a22d3c83a769b76f65a81. All three pass with PAIR_AGENT=claude:
closestream pair 21.947s; fenced-argv test 4.558s. No fixture changes made here.

### 2026-10-13 — ariadne:3

Reproduced under `PAIR_AGENT=codex`: the two closestream tests reported verdict
"unknown" / a sidecar without both runs (Claude JSONL fed to a Codex parse), and the
fenced-argv test got a `codex exec` argv. Precedence (`judge.ResolveReviewerCLI`):
`--agent`, then `AGENT_CMD`, then the reviewer policy over `MainAgent` (`PAIR_AGENT`,
then `CODEX_CI`/`CODEX_THREAD_ID`, else Claude). After pinning, the three pass under
`PAIR_AGENT=codex`, `PAIR_AGENT=claude`, `PAIR_AGENT=claude AGENT_CMD=codex` and
`PAIR_AGENT=codex AGENT_CMD=gemini`. This slot's ambient env is `PAIR_AGENT=claude`,
which is why Claude slots never see it.

Audit: full `make test` under `PAIR_AGENT=codex` (16:19) and under `AGENT_CMD=codex`
(16:29), 12 shards over 1120 cmd/sdlc tests + 47 packages each: both green, so no further
fixture depends on the ambient agent. `scripts/test-shard.py` spawns `go test` without an
`env=`, so both variables reached every test; the original failure was found the same way
(#337's full suite under `PAIR_AGENT=codex`). The diff touches only test files and this
issue: production resolution and its precedence tests are unchanged.
