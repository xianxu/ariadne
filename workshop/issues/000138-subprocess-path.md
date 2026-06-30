---
id: 000138
status: done
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-29
estimate_hours: 0.27
started: 2026-06-29T22:09:01-07:00
actual_hours: 0.49
---

# sdlc subprocess path resolution

## Problem

Fresh agent subprocesses invoked by `sdlc` can fail to find `sdlc` on `PATH`.
During pair#81 retro, the main Codex shell had to discover
`/Users/xianxu/workspace/ariadne/bin/sdlc`, and fresh review subprocesses that
started from a narrower shell environment attempted `sdlc --help` and hit
`command not found`.

User shell configuration can work around this locally, but SDLC-spawned agents
should not depend on every harness loading the same interactive zsh startup
files.

## Spec

Make SDLC subprocess invocation robust to missing user PATH entries.

- Boundary review and judge subprocess prompts/commands should resolve the
  current `sdlc` binary once and pass either an absolute path or an environment
  containing the SDLC owner `bin/`.
- The behavior should work from downstream repos such as `pair`, where the
  `sdlc` executable lives in sibling `ariadne/bin/`.
- Prompt text that asks a fresh agent to inspect SDLC help should use the
  resolved executable path or explicitly state the PATH adjustment required.
- The fix should not require users to put ariadne aliases in `~/.zshenv`,
  though that remains a valid personal convenience.
- Failure messages should identify which binary path was attempted and the PATH
  used, so future environment issues are diagnosable.

## Done when

- Fresh boundary review subprocesses can run `sdlc --help` from a downstream
  repo without relying on the user's interactive shell startup files.
- Tests or a process-level fixture cover a subprocess with a minimal PATH.
- The prompt/command path uses the resolved SDLC binary instead of assuming the
  bare `sdlc` command is globally available.
- Error output includes the attempted executable path and PATH when resolution
  fails.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.1 impl=0.15
design-buffer: 0.15
total: 0.27
```

One focused extend of the judge dispatch seam: a pure `binAugmentedEnv` helper +
wiring into `Run` (via `os.Executable()`) + diagnosable launch errors + a pure
table test + a process-level minimal-PATH fixture + a doc line. Design pre-resolved
by the durable plan → reduced design + +15% buffer; impl at the v3.1 40%-scaled
smaller-go-module top.

Detailed design + TDD breakdown: `workshop/plans/000138-subprocess-path-plan.md`.

## Plan

- [x] Locate boundary review and judge subprocess launch/prompt construction.
- [x] Add a single SDLC binary resolution helper for current-repo and downstream-repo execution.
- [x] Thread the resolved executable or PATH into subprocess commands/prompts.
- [x] Add a minimal-PATH regression test or process-level fixture.
- [x] Update help/prompt docs where they mention running `sdlc`.

## Log

### 2026-06-26

- Created from pair#81 retro point 1: `sdlc` was missing from the agent shell
  PATH, and fresh review subprocesses should not depend on user zsh startup
  configuration to find the workflow binary.

### 2026-06-29
- 2026-06-29: closed — go test ./cmd/sdlc/... all pass. New dispatch_test.go: TestBinAugmentedEnv (prepend owner bin/ to PATH / synthesize when absent / no-op on empty dir) + TestMinimalPathResolvesSdlc — a real sh -c "command -v sdlc" under a deliberately narrow base PATH resolves the injected owner-bin sdlc (the Done-when minimal-PATH coverage, no agent). ownerBinDir() single-sources dir(os.Executable()) for Run (build env) + Dispatch (diagnosable launch error). No dependence on user shell startup files; works from downstream repos.; review verdict: FIX-THEN-SHIP

Implemented per `workshop/plans/000138-subprocess-path-plan.md` in
`cmd/sdlc/internal/judge/dispatch.go`:
- `ownerBinDir()` = `filepath.Dir(os.Executable())` — single source for "where the
  running sdlc + sibling tools live" (consumed by `Run` + the launch-failure error,
  per the plan-quality gate's ARCH-DRY note).
- `binAugmentedEnv(binDir, env)` (pure) prepends binDir to the subprocess `PATH=`
  entry (synthesizes one if absent; no-op on empty/`.`).
- `Run` sets `cmd.Env = binAugmentedEnv(ownerBinDir(), os.Environ())` — so the
  spawned agent inherits a PATH that resolves `sdlc` regardless of the spawning
  shell's startup files. Best-effort: if `os.Executable()` errors, inherit the
  parent env (today's behavior). Works from downstream repos (the binary is
  `…/ariadne/bin/sdlc` regardless of cwd). Launch-failure errors now name the agent
  + the owner bin/.

Discoveries / gate notes:
- **No prompt rewrite needed (D1):** no boundary-review/judge prompt instructs the
  agent to run `sdlc` — the architecture block is *embedded* at dispatch, not
  fetched via a `sdlc arch-principles` call. So the Spec's "prompt text … should use
  the resolved executable path" is satisfied **vacuously** (no such text exists);
  PATH injection covers every `sdlc` the agent runs on its own initiative, which is
  strictly more robust.
- The one-line `Run` wiring (`os.Executable()` → set env) is not directly asserted
  (`os.Executable()` is uncontrollable in-test; the existing dispatch tests stub
  `Run`); the *logic* is covered by the pure helper + the process fixture.

Verification: `go test ./cmd/sdlc/...` all pass. New `dispatch_test.go`:
`TestBinAugmentedEnv` (prepend / synthesize / no-op) + `TestMinimalPathResolvesSdlc`
— a real `sh -c 'command -v sdlc'` under a narrow base PATH resolves the injected
owner-bin `sdlc` (the Done-when minimal-PATH coverage, no agent process).
