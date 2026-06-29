# Default SDLC Judges to Current Agent Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every `sdlc` judge dispatch default to the current pair-managed agent when no explicit agent override is supplied.

**Architecture:** Add one pure resolver in `cmd/sdlc/internal/judge` and route every judge-dispatching command through it. This keeps the precedence single-sourced (`ARCH-DRY`), pure and unit-tested (`ARCH-PURE`), and applied to all call sites named in the issue rather than only `sdlc judge` (`ARCH-PURPOSE`).

**Tech Stack:** Go, Cobra commands, existing `cmd/sdlc/internal/judge` dispatch adapters, existing `cmd/sdlc` command tests.

---

## Core Concepts

### Pure Entities

| Name | Lives in | Status |
|------|----------|--------|
| `AgentDefaultEnv` | `cmd/sdlc/internal/judge/agent_resolver.go` | new |
| `ResolveAgentCLI` | `cmd/sdlc/internal/judge/agent_resolver.go` | new |

- **AgentDefaultEnv** — normalized environment snapshot used for judge-agent default selection.
  - **Relationships:** One resolver input owns zero or one values for explicit CLI, `AGENT_CMD`, `PAIR_AGENT`, and known agent environment signals.
  - **DRY rationale:** Prevents each command from separately knowing how environment precedence maps to a judge agent.
  - **Future extensions:** Add configured strategies such as `same`, `other`, or explicit `codex:claude` mappings here without touching command call sites.

- **ResolveAgentCLI** — pure function that chooses `claude`, `codex`, or `gemini` from explicit and environment inputs.
  - **Relationships:** N command call sites call one resolver; one resolver returns one `AgentCLI`.
  - **DRY rationale:** Replaces duplicated `orStr(..., "claude")` fallback logic across judge, change-code, close, milestone-close, and preflight.
  - **Future extensions:** Widen the input with a strategy field while keeping the same call-site shape.

### Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `CurrentAgentDefaultEnv` | `cmd/sdlc/internal/judge/agent_resolver.go` | new | process environment |
| Judge command call sites | `cmd/sdlc/*.go` | modified | Cobra flags and judge dispatch |
| Close boundary review dry run | `cmd/sdlc/milestoneclose.go` and `cmd/sdlc/close.go` | modified | command-line rendering |

- **CurrentAgentDefaultEnv** — thin IO shell that reads process environment and constructs `AgentDefaultEnv`.
  - **Injected into:** Command call sites before building `judge.DispatchOptions`.
  - **Future extensions:** Add repo config or pair config reads here if issue follow-up chooses configurable strategies.

- **Judge command call sites** — existing dispatching commands pass explicit flag values plus `CurrentAgentDefaultEnv()` into the resolver.
  - **Injected into:** `runJudge`, `runPlanQualityJudge`, `runEstimateQualityJudge`, close boundary review, milestone close boundary review, and preflight judges.
  - **Future extensions:** If dispatching moves behind a shared command-runner, the runner can own this one resolver call.

- **Close boundary review dry run** — the close dry-run path prints the would-be review command line instead of only saying it would dispatch.
  - **Injected into:** The same `boundaryReviewParams` path as real close review dispatch, so dry-run command rendering cannot drift from real dispatch.
  - **Future extensions:** Milestone dry-run can reuse the same helper if the operator wants parity there too.

## Chunk 1: Resolver and Command Wiring

### Task 1: Add resolver tests

**Files:**
- Test: `cmd/sdlc/internal/judge/judge_test.go`

- [ ] **Step 1: Write failing tests for resolver precedence**

Add table tests covering:

```go
explicit claude beats PAIR_AGENT=codex
explicit bogus remains bogus so dispatch validation still fails
AGENT_CMD=gemini beats PAIR_AGENT=codex
AGENT_CMD=bogus remains bogus so dispatch validation still fails
PAIR_AGENT=codex selects codex when explicit and AGENT_CMD are empty
CODEX_CI or CODEX_THREAD_ID selects codex when AGENT_CMD and PAIR_AGENT are empty
empty environment falls back to claude
unknown PAIR_AGENT values fall through to known process signals or final claude fallback
```

- [ ] **Step 2: Run resolver tests and verify RED**

Run: `go test ./cmd/sdlc/internal/judge -run ResolveAgentCLI -count=1`

Expected: FAIL because `ResolveAgentCLI` and `AgentDefaultEnv` do not exist.

- [ ] **Step 3: Implement minimal resolver**

Create `cmd/sdlc/internal/judge/agent_resolver.go` with:

```go
type AgentDefaultEnv struct {
    AgentCmd string
    PairAgent string
    CodexCI string
    CodexThreadID string
    ClaudeCode string
}

func CurrentAgentDefaultEnv() AgentDefaultEnv
func ResolveAgentCLI(explicit string, explicitSet bool, env AgentDefaultEnv) AgentCLI
```

Recognize only supported `AgentCLI` values for auto-detected defaults. Precedence is explicit, `AGENT_CMD`, `PAIR_AGENT`, known process/session signals, then `AgentClaude`. Explicit means the Cobra `--agent` flag was actually set by the operator; use `cmd.Flags().Changed("agent")` at command boundaries or an equivalent boolean in tests. If `explicitSet` is true, return `AgentCLI(explicit)` even when the value is unknown so existing `BuildArgs` validation still reports `unknown agent`. Treat `AGENT_CMD` the same way because it is an operator/script override: if it is set, return `AgentCLI(env.AgentCmd)` even when unknown so existing dispatch validation still reports the bad override. Unknown `PAIR_AGENT` values may fall through because `PAIR_AGENT` is a current-agent detection input. The first implementation should read concrete process/session signals already visible in this repo/session: `CODEX_CI` and `CODEX_THREAD_ID` for Codex, and `CLAUDECODE` or `CLAUDE_CODE_ENTRYPOINT` for Claude if present. Do not use broad capability/configuration variables such as API keys for active-agent detection; omit Gemini process auto-detection until a conservative Gemini session signal is identified. Do not introduce wildcard environment scans.

- [ ] **Step 4: Run resolver tests and verify GREEN**

Run: `go test ./cmd/sdlc/internal/judge -run ResolveAgentCLI -count=1`

Expected: PASS.

### Task 2: Route dispatch call sites through resolver

**Files:**
- Modify: `cmd/sdlc/judge.go`
- Modify: `cmd/sdlc/changecode.go`
- Modify: `cmd/sdlc/close.go`
- Modify: `cmd/sdlc/milestoneclose.go`
- Modify: `cmd/sdlc/preflight.go`
- Test: `cmd/sdlc/judge_test.go` or existing command tests
- Test: `cmd/sdlc/changecode_test.go`
- Test: `cmd/sdlc/preflight_test.go` or existing preflight tests
- Test: `cmd/sdlc/closereview_test.go`
- Test: `cmd/sdlc/milestoneclose_test.go`

- [ ] **Step 1: Write failing dispatch-path tests**

Add tests proving:

```go
PAIR_AGENT=codex with no AGENT_CMD makes sdlc judge --dry-run print codex exec
explicit --agent claude still prints claude even with PAIR_AGENT=codex
explicit --agent bogus still fails instead of falling through to PAIR_AGENT=codex
AGENT_CMD=gemini still prints gemini when PAIR_AGENT=codex
AGENT_CMD=bogus still fails instead of falling through to PAIR_AGENT=codex
change-code plan-quality uses codex when f.Agent is empty, f.AgentExplicit is false, and PAIR_AGENT=codex
change-code estimate-quality uses codex when f.Agent is empty, f.AgentExplicit is false, and PAIR_AGENT=codex
change-code explicit f.Agent=claude with f.AgentExplicit=true beats PAIR_AGENT=codex
change-code AGENT_CMD=gemini beats PAIR_AGENT=codex
preflight `runOnePreflight` uses codex when opts.Agent is empty and PAIR_AGENT=codex
preflight preserves explicit internal opts.Agent=claude when PAIR_AGENT=codex
full-issue close boundary review uses codex when Agent is empty and PAIR_AGENT=codex
full-issue close boundary review dry-run prints the would-be codex exec command
milestone boundary review uses codex when Agent is empty and PAIR_AGENT=codex
```

- [ ] **Step 2: Run focused command tests and verify RED**

Run: `go test ./cmd/sdlc -run 'Judge.*Agent|ChangeCode.*Agent|Milestone.*Agent|Close.*Agent|Preflight.*Agent' -count=1`

Expected: at least the new `PAIR_AGENT=codex` path fails because the current fallback is Claude.

- [ ] **Step 3: Wire call sites**

Make command flag defaults empty for judge-agent flags; do not pre-fill them from `AGENT_CMD`, because the resolver owns that environment input. Add explicit-source plumbing to the flag structs that cross into helper functions:

```go
AgentExplicit bool
```

Set it in Cobra `RunE` via `cmd.Flags().Changed("agent")` for `judge`, `change-code`, `close`, and `milestone-close`. Replace local fallbacks like `judge.AgentCLI(orStr(x, "claude"))` with `judge.ResolveAgentCLI(x, agentExplicit, judge.CurrentAgentDefaultEnv())`. For preflight, preserve `opts.Agent` as explicit internal input when it is non-empty (`agentExplicit := opts.Agent != ""`).

For full-issue close dry-run, build the same boundary review prompt and `judge.DispatchOptions` that real dispatch would use, then print `judge.FormatCommandLine` without invoking `judge.Dispatch`. This should make the issue's dry-run acceptance criterion literal: with `PAIR_AGENT=codex` and no `AGENT_CMD`, close boundary review dry-run output contains `codex exec`.

- [ ] **Step 4: Run focused command tests and verify GREEN**

Run: `go test ./cmd/sdlc -run 'Judge.*Agent|ChangeCode.*Agent|Milestone.*Agent|Close.*Agent|Preflight.*Agent' -count=1`

Expected: PASS.

### Task 3: Update help and verify the issue

**Files:**
- Modify: `cmd/sdlc/helptext/judge.md`
- Modify: `cmd/sdlc/helptext/change-code.md`
- Modify: `cmd/sdlc/helptext/close.md`
- Modify: `cmd/sdlc/helptext/milestone-close.md`
- Modify: `workshop/issues/000129-default-sdlc-judges-to-current-agent.md`

- [ ] **Step 1: Update help text**

Document the default as:

```text
explicit --agent, then AGENT_CMD, then PAIR_AGENT/current known agent signals, then claude
```

- [ ] **Step 2: Tick issue plan and log the implementation**

Update the three issue plan checkboxes and add a dated log entry naming the resolver and verification commands.

- [ ] **Step 3: Run full relevant verification**

Run:

```bash
sdlc issue validate --issue 129
go test ./cmd/sdlc/internal/judge ./cmd/sdlc -count=1
```

Expected: PASS.
