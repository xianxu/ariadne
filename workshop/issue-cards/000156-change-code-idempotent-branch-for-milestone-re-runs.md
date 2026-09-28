---
id: '000156'
status: done
started: 2026-07-06T15:08:58-07:00
created: 2026-07-01
updated: 2026-07-06
estimate_hours: 0.68
actual_hours: 0.59
---

# change-code idempotent branch for milestone re-runs

## Problem

`sdlc change-code --issue N` errors at its branching step when re-run for a
**subsequent milestone** on an issue whose feature branch already exists and is
already checked out:

```
==> branching: in-place (default; --worktree=yes for an isolated worktree)
Error: git checkout -b 000153-sdlc-retro-process-manual: exit status 128
fatal: a branch named '000153-sdlc-retro-process-manual' already exists
```

Multi-milestone issues legitimately re-run `change-code` per milestone to get that
milestone's **plan-quality + estimate-quality gates** on the (re-)designed plan. In
the failing run the gates ran fine (both INFO) — only the final branch-creation step
failed, because it unconditionally does `git checkout -b <branch>` instead of
detecting that the branch already exists / we're already on it.

Encountered on ariadne#153 M2 (re-running `change-code` for the M2 milestone on the
existing `000153-*` branch).
