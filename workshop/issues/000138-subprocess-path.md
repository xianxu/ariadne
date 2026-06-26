---
id: 000138
status: open
deps: []
github_issue:
created: 2026-06-26
updated: 2026-06-26
estimate_hours:
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

## Plan

- [ ] Locate boundary review and judge subprocess launch/prompt construction.
- [ ] Add a single SDLC binary resolution helper for current-repo and downstream-repo execution.
- [ ] Thread the resolved executable or PATH into subprocess commands/prompts.
- [ ] Add a minimal-PATH regression test or process-level fixture.
- [ ] Update help/prompt docs where they mention running `sdlc`.

## Log

### 2026-06-26

- Created from pair#81 retro point 1: `sdlc` was missing from the agent shell
  PATH, and fresh review subprocesses should not depend on user zsh startup
  configuration to find the workflow binary.
