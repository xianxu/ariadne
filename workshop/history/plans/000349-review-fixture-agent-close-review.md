# Boundary Review — ariadne#349 (whole-issue close)

| field | value |
|-------|-------|
| issue | 349 — Pin agent-specific review fixtures instead of inheriting session defaults |
| repo | ariadne |
| issue file | workshop/issues/000349-review-fixture-agent.md |
| boundary | whole-issue close |
| milestone | — |
| window | ce848b4736f3df5ab8339936b96946859915997e..a89beca702b4571da671aa6eb62f5592511d848b |
| command | sdlc close --issue 349 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T16:50:53-07:00 |
| verdict | SHIP |

## Review

I've read the diff and resolver; now running the three pinned tests under conflicting agent environments.

Now reverting the pins to confirm the tests go red without them.

The pins work: all three tests pass under conflicting environments and fail with the pins removed. Spot-checking the audit claim next, on other tests that use the stubbed judge.

```verdict
verdict: SHIP
confidence: high
```

This is a small change to test fixtures only. It adds a `pinReviewAgent` helper that sets `AGENT_CMD`, which in `judge.ResolveReviewerCLI` (`cmd/sdlc/internal/judge/agent_resolver.go:44-52`) is outranked only by an explicit `--agent`. The three named fixtures are pinned to `judge.AgentClaude`. I checked the Done-when items directly:
- **With the pins:** all three tests pass under `PAIR_AGENT=codex` and under `PAIR_AGENT=claude AGENT_CMD=codex`.
- **Without the pins:** I removed the three pin lines in this checkout and ran again under `PAIR_AGENT=codex AGENT_CMD=` (unset). All three failed (`--- FAIL` for each), then I restored the files.
- **Production code:** nothing changed. Resolution code and its precedence tests (`judge_command_test.go`, `changecode_test.go`, `closereview_test.go`, `judge_test.go:995`) are untouched.

Nothing blocks SHIP.

1. **Strengths**
   - `cmd/sdlc/agentpin_test.go:16-19`: one named helper whose doc comment states where it sits in the precedence order and why that is enough. It reads better than scattering raw `Setenv` calls.
   - Each call site says why the pin is there (`// the stream is Claude JSONL`, `// asserts Claude's argv`), so the fixture's dependency on one agent is explicit.
   - The fix is the right size: production resolution stays as it is and the precedence tests stay authoritative, as the Spec asks.
   - The audit used two different variables, `PAIR_AGENT=codex` and `AGENT_CMD=codex`, which catch different kinds of hidden dependency. That was a sensible design.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The helper takes any `judge.AgentCLI`, but only Claude is used. That's fine and needs no change.

5. **Test coverage notes**
   - The fix is proven to matter: the three tests fail without the pins and pass with them, under conflicting `PAIR_AGENT` and `AGENT_CMD` values.
   - **Audit claim (full suite green under each variable):** I could not confirm it independently. My own run of the 87 tests that use `stubJudge` under `PAIR_AGENT=codex` hit `go test`'s default 10-minute timeout without finishing. It printed no `--- FAIL` lines, so the result is inconclusive rather than a failure. That is consistent with the suite being split into shards by `scripts/test-shard.py`. I'm relying on the implementor's logged sharded runs for the wider audit.
   - There is no test that the helper itself keeps working, i.e. that `AGENT_CMD` still outranks `PAIR_AGENT`. That ordering is already pinned by `judge_test.go:995` and `judge_command_test.go:83` (both set `AGENT_CMD=gemini` with `PAIR_AGENT=codex`), so the helper's assumption is covered.

6. **Architectural notes**
   - **ARCH-DRY:** pass. There is one helper for the pin. The existing `AGENT_CMD="" + PAIR_AGENT` pairs in the precedence tests do the opposite job (clearing the override so the policy is tested), so they shouldn't be merged into it.
   - **ARCH-PURE:** pass. Test-only change; no logic was added.
   - **ARCH-PURPOSE:** pass. The three named sites are fixed, and the wider class (fixtures relying on whatever agent the environment happens to set) was swept under both variables.
   - **ARCH-MOCK:** pass. The existing judge stubs are reused; no new external call is added.
   - **ARCH-CONSTRAINTS:** N/A. The change only touches test environment setup and has no runtime effect.
   - **ARCH-SECURE:** pass. `t.Setenv` is undone when the test ends and touches no credentials or real user state.
   - **ARCH-ORDER:** N/A. Nothing holds state between events: `t.Setenv` is scoped to one test and also rules out `t.Parallel`.
   - **ARCH-FUNERAL:** N/A. Nothing durable is created; `t.Setenv` restores the variable through `t.Cleanup`.
   - **Docs gate:** no new user-facing surface, so no atlas or README update is needed.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: test-audit-evidence
    title: |
      Full-suite audit under PAIR_AGENT=codex not independently reproduced by review
    detail: |
      An unsharded run of the 87 stubJudge tests hit go test's 10-minute default timeout without finishing and printed no --- FAIL lines, so the result is inconclusive. The sharded runs recorded in the Log remain the evidence for the wider audit.
```
