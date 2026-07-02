---
id: 000161
status: open
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours:
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

## Spec

Add an idempotent `ensure-uv` target that mirrors `ensure-cue` /`ensure-go`
(`Makefile.workflow`): no-op when `uv` is on PATH (don't fight asdf/manual/
pipx installs), auto-install via Homebrew on macOS (`brew install uv`), else fail
fast with guidance (`https://docs.astral.sh/uv/getting-started/installation/`).

**Design decision to resolve at implementation (the one real question):** where
does `uv` belong?

- **`ensure-go`/`ensure-cue` guard deps of *ariadne's own build*** — the header
  comment is explicit: "this guarantees only the one dep the base layer's own
  build needs." `uv` is **not** an ariadne build-dep; it's a *downstream
  consumer* dep (metis/kbench Python planes). By that rule it would live with
  the consumer (metis's own `bootstrap:` prereq, or nous's Brewfile) — see the
  layer-appropriate-placement principle.
- **Counter-argument (why base is defensible):** ariadne historically already
  assumed "shell + python" at the base (per the `ensure-go` comment), and `uv`
  is fast becoming the *universal* Python toolchain every derivative with any
  Python surface will want — like `go`/`cue`, provision it once at the base so
  every consumer inherits it. This is the operator's stated preference ("add to
  the base-layer bootstrap").

**Recommendation:** honor the operator's ask — add `ensure-uv` to ariadne's
base `bootstrap:` chain — but the implementer should confirm this over the
"push it down to metis" alternative before writing, and record the rationale in
`## Log` (ARCH-DRY / layer-placement). If base is chosen, the `ensure-go`
header comment ("only the base layer's own build needs") must be updated so it
no longer contradicts the new target.

Note: this is a Makefile toolchain-provisioning add, no new architectural
surface. `uv` installs its own managed Python, so it does not additionally
require a system `python3`.

## Done when

- `ensure-uv` exists in `Makefile.workflow`, idempotent (no-op when `uv`
  present), brew-installs on macOS, fails fast with a real URL otherwise.
- It is wired into the resolved bootstrap flow (either ariadne's `bootstrap:`
  prereq list per the recommendation, or the chosen consumer's) so a fresh
  `make bootstrap` leaves `uv` on PATH.
- The placement decision + rationale recorded in `## Log`; if base-placed, the
  `ensure-go` "only the base layer's own build needs" comment is reconciled.
- Verified: `command -v uv` after a bootstrap on a machine without it (or a
  faithful dry-run: temporarily shadow `uv` off PATH → target installs → no-op
  on re-run).

## Plan

- [ ] Confirm placement (base `bootstrap:` vs metis-local vs nous Brewfile);
      record the decision in `## Log`
- [ ] Add `ensure-uv` target mirroring `ensure-cue` (idempotent / brew / fail-fast)
- [ ] Wire it into the chosen `bootstrap:` prereq list
- [ ] If base-placed, reconcile the `ensure-go` header comment
- [ ] Verify: shadow `uv` off PATH → target installs → re-run is a no-op

## Log

### 2026-07-01
- Filed from the metis#1 M3 work (kaggle-ml-base-layer project): M3's Python
  step-types run via `uv`, and the operator already installed `uv` locally, so
  the bootstrap should provision it for the next fresh env. Operator asked for
  it in the ariadne base-layer bootstrap; captured the base-vs-consumer
  placement tension in `## Spec` rather than deciding silently — `ensure-cue`
  (#122) is the target to mirror.
