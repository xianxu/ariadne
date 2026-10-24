---
id: 000605
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'd106c8cc154e75e940985adf336ae539ec940f1e' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

Classify each post-hoc `ctx.Err()` read after a command (`Run`, `Wait`,
`Output`, `CombinedOutput`) returned:

- **Safe fallback:** a late context error only degrades to a safe outcome.
  Record why in a comment at the site.
- **Misreport:** it can turn a success into a timeout or a cancellation. Fix
  it, reusing `stoppedBy` where the shape fits (moved to a shared package if
  more than `cmd/sdlc` needs it).

Because this is a mechanical class, add a `pkg/golint` rule: flag a
`ctx.Err()` used as a verdict after a `Run`/`Wait`/`Output` returned, when the
run's own error was not checked first. Exempt a site with a reasoned
annotation, the way the existing rules take one.

## Done when

- Every listed site, plus any the new lint finds, is classified, and each
  misreporting site is fixed with a test that fails without the fix.
- The golint rule exists, has positive and negative cases, and the tree passes
  it.

## Plan

- [ ]

## Log

### 2026-10-24
