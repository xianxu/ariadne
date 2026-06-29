---
id: 000141
status: open
deps: []
github_issue:
created: 2026-06-29
updated: 2026-06-29
estimate_hours:
---

# sdlc merge confirmation should support agent-safe defaults

## Problem

`sdlc merge` asks for final confirmation before irreversible actions:

- server-side GitHub PR merge;
- remote branch deletion;
- switching/pulling the local checkout;
- archiving completed issues to history;
- deleting the local feature branch or worktree.

That confirmation is sensible for humans in an interactive terminal. In an
agent/non-interactive run, however, the prompt defaults to "no". In pair#84,
`sdlc merge` ran the expensive pre-merge judges successfully and then aborted at
the final prompt because no interactive answer could be supplied. The operator
had to rerun with `--yes`.

The problem is not the confirmation itself. The problem is that non-interactive
contexts discover the need for `--yes` only after spending time on slow judges.

## Spec

Make `sdlc merge` agent-safe without weakening the irreversible-action guard for
humans.

Desired behavior:

- In interactive terminals, keep the final confirmation prompt by default.
- In non-interactive contexts, fail fast before pre-merge judges with a clear
  message: rerun `sdlc merge --yes` after confirming irreversible actions.
- `--yes` remains the explicit opt-in for scripted/agent flows.
- Dry runs remain non-mutating and should not require final confirmation.

The confirmation exists to protect irreversible merge/cleanup actions, not to
guard the read-only pre-merge judges. Therefore the prompt/precondition should
be checked early enough that an agent does not wait through all judges only to
abort at the end.

## Done when

- [ ] `sdlc merge` detects non-interactive stdin/stdout before running
      pre-merge judges.
- [ ] Non-interactive merge without `--yes` fails fast with an actionable error.
- [ ] Interactive merge still prompts before irreversible actions.
- [ ] `sdlc merge --yes` keeps the current scripted flow.
- [ ] Tests cover interactive, non-interactive, `--yes`, and `--dry-run`
      combinations.

## Plan

- [ ] Inspect `cmd/sdlc/merge.go` prompt and prompter abstraction.
- [ ] Define the TTY/non-interactive detection rule.
- [ ] Move or add a preflight confirmation-capability check before slow judges.
- [ ] Preserve the final confirmation in interactive runs.
- [ ] Update help text to explain when agents should use `--yes`.
- [ ] Add tests for early refusal before judge invocation.

## Log

### 2026-06-29

- Created from pair#84 dogfooding: the merge judges passed, then `sdlc merge`
  aborted at the final confirmation prompt in a non-interactive agent run. A
  rerun with `--yes` succeeded.
