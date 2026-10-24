---
id: 000605
status: open
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
---

# Audit post-hoc ctx.Err() verdicts after a command returns, with a golint rule

## Problem

ariadne#599 found that a local check read its context's error after the
command returned, and after bookkeeping, as the run's verdict. A command that
succeeded could then be recorded TimedOut. #599 fixed the local-check path
(the run verdict, `prepareAffectedGraph`, the make probe) through one pure
helper, `stoppedBy(runErr, ctxErr) (cause, stopped)` in
`cmd/sdlc/localcheck_run.go`. Its close review listed the same shape
elsewhere, outside #599's scope:

- `cmd/sdlc/internal/affected/golist.go:153` (the reviewer judged it a safe
  fallback: it only falls back to the full suite)
- `cmd/sdlc/weeklystats_inputs.go:163`
- `cmd/sdlc/weeklyresource_machine.go:41`
- `cmd/sdlc/internal/netcheck/probe.go:64`
- `cmd/sdlc/internal/judge/dispatch.go:186`
- `cmd/sdlc/cigate.go:341`
- `cmd/sdlc/localcheck_run.go`: `interrupted := ownedCtx.Err()`, read after
  slot release and lease release, so a signal arriving during bookkeeping can
  turn a passing run into `queueInterruptedError`

Line numbers are as of main 6afd91697.
