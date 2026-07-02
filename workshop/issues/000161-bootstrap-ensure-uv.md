---
id: 000161
status: done
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 0.27
started: 2026-07-01T22:28:46-07:00
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module   design=0.1 impl=0.15
total: 0.27
```

Single well-specced target that mirrors an existing one (`ensure-cue`): the only
design content is the base-vs-consumer placement call (pre-resolved in `## Spec`);
implementation is the mirrored recipe + one prereq wire + one comment reconcile.
Single-pass atomic — no milestone-review item.

## Plan

- [x] Confirm placement (base `bootstrap:` vs metis-local vs nous Brewfile);
      record the decision in `## Log`
- [x] Add `ensure-uv` target mirroring `ensure-cue` (idempotent / brew / fail-fast)
- [x] Wire it into the chosen `bootstrap:` prereq list
- [x] If base-placed, reconcile the `ensure-go` header comment
- [x] Verify: shadow `uv` off PATH → target installs → re-run is a no-op

## Log

### 2026-07-01
- 2026-07-01: closed — ensure-uv added via shared define ensure-tool macro (Makefile.workflow), wired into bootstrap prereq chain. All 3 legs verified: no-op when uv present (exit 0); uv-absent+fake-brew invokes `brew install uv` (exit 0); uv+brew absent fails fast with real install URL (exit 1). ensure-go/ensure-cue expansions behaviorally identical via make -n (behavior-preserving refactor); both still no-op (present).; review verdict: SHIP
- Filed from the metis#1 M3 work (kaggle-ml-base-layer project): M3's Python
  step-types run via `uv`, and the operator already installed `uv` locally, so
  the bootstrap should provision it for the next fresh env. Operator asked for
  it in the ariadne base-layer bootstrap; captured the base-vs-consumer
  placement tension in `## Spec` rather than deciding silently — `ensure-cue`
  (#122) is the target to mirror.
- **Placement decision → ariadne base `bootstrap:`** (ARCH-DRY / layer-placement).
  The strict layer-placement read says push `uv` down to metis (it's a
  downstream-*consumer* runtime dep, not an ariadne *build*-dep). Chose base
  anyway, honoring the operator's ask: `uv` is becoming the universal Python
  toolchain every derivative with a Python surface will want, so — like `go`/`cue`
  — provision it once at the base and every consumer inherits it via the symlinked
  `Makefile.workflow`. Asked the operator to confirm the base-vs-push-down fork
  before writing; no reply within the window, so proceeded on the twice-stated
  preference (Spec + Log). The recorded rationale lives in the `ensure-uv` header
  comment for the next reader.
- **ARCH-DRY: extracted `define ensure-tool` instead of adding a third copy.**
  The plan-quality judge flagged that `ensure-uv` would be the *third* near-
  identical block (after `ensure-go`/`ensure-cue`), differing only in tool /
  formula / reason / install-noun / url — a rule-of-three trigger. The Makefile
  already uses `define`/`$(call)` canned recipes (`check_undone_issues`), so the
  macro is idiomatic here, not a new arcane pattern. Collapsed all three targets
  through one parametrized `$(call ensure-tool,...)`; `make -n` confirms the
  `ensure-go`/`ensure-cue` expansions are behaviorally identical to the prior
  inline recipes (the multi-line `\`-continued recipe collapses to a single
  line under `$(call)`, but the shell receives an equivalent command — no
  behavioral change). Behavior-preserving refactor of shared base-layer infra;
  the pre-existing `ensure-go.test.sh` passes green against it (all 3 branches).
- **`ensure-go` comment reconciled.** It previously claimed the family provisions
  "only the one dep the base layer's own build needs" — now false with `uv` (a
  consumer dep) in the chain. Rewrote so `ensure-go`/`ensure-cue` are described as
  guarding the base layer's *own* build, and `ensure-uv` is explicitly called out
  as the one reaching past it.
- **Verified** (`Makefile.workflow`, three legs of `ensure-uv`):
  - *no-op*: `make ensure-uv` with `uv` present → silent, exit 0 (idempotent).
  - *install*: `PATH=<fakebin-with-echoing-brew>:/usr/bin:/bin make ensure-uv`
    (uv absent) → prints "==> uv not found — installing via Homebrew" and invokes
    `brew install uv`, exit 0.
  - *fail-fast*: `PATH=/usr/bin:/bin make ensure-uv` (uv + brew absent) → prints
    the reason + real install URL, exit 1.
  Used a fake `brew` for the install leg (per the Done-when note) so the "installs
  when absent" path is exercised deterministically without a real install or
  hitting "already installed" on this brew-provisioned machine. `ensure-go` /
  `ensure-cue` still no-op (present), confirming the refactor is non-breaking.
