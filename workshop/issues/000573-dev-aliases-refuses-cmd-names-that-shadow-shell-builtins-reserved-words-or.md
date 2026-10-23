---
id: 000573
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '86ce5b83effc6755c7992a6b0b37727ab43e4863' # card fields mirrored from issue-cards; edit via sdlc
---

# dev-aliases refuses cmd names that shadow shell builtins, reserved words or system commands

## Problem

## Spec

**dev-aliases shadows shell builtins and system commands (found by echo:0, verified by ops-echo:0, 10-23).**
`construct/dev-aliases.sh` emits a shell function for every owned `cmd/X` with no check on the name. echo has
`cmd/echo`, so every operator shell that sources dev-aliases (from the operator's zshrc) got an `echo()` function
overriding the builtin. The prompt calls `echo`; the function rebuilt `~/workspace/echo/bin/echo` when missing and
ran echo's room CLI, which broke the operator's shell, and deleting the binary didn't stick. echo#19 renames that CLI
to `echoctl`, but nothing stops the next repo doing the same.

Fix: dev-aliases refuses, with one warning naming the repo and command, any `cmd/X` whose name is a shell builtin or
reserved word (in the sourcing shell: `builtin`/`whence -w` in zsh, `type -t` in bash), or that already resolves to a
system command outside the workspace (at least the POSIX utilities). No function is emitted for a refused name; the
binary stays reachable by its path.

## Done when

- Sourcing dev-aliases with a repo that has `cmd/echo` (and `cmd/test`, `cmd/ls`) emits no function for them and
  prints one warning each; ordinary cmd names still get their functions (test in bash and zsh).
- The check runs at source time without measurable shell-start delay (measured).

## Plan

- [ ]

## Log

### 2026-10-23
