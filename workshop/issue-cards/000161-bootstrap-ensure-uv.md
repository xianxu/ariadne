---
id: '000161'
status: done
started: 2026-07-01T22:28:46-07:00
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 0.27
actual_hours: 1.59
---

# bootstrap: ensure-uv — install uv (Python package manager) in the bootstrap chain

## Problem

The stack is growing a Python data plane. metis#1 M3 ships pure-Python step-types
(`Dataset`/`Schema`/`cv-split`/`train`/`predict`) run hermetically via **uv**
(`uv run --project <root> python -m metis.steps.<type>`); kbench and future
competition workspaces inherit that contract. Today nothing in the bootstrap
cascade guarantees `uv` is present — a fresh derivative clone can `make bootstrap`
green and still fail the first `metis run` at the Python boundary. We already
solved this shape for `go` (#61) and `cue` (#122) with idempotent `ensure-*`
targets; `uv` is the Python equivalent and wants the same treatment.
