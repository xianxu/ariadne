---
id: 000026
status: done
deps: []
created: 2026-05-19
updated: 2026-05-19
estimate_hours: 0.5
actual_hours: 0.3
---

# tart: TART_SUFFIX knob for parallel VMs per repo

## Problem

`.tart/Makefile` currently hard-codes one VM per consuming repo:

```make
TART_VM ?= $(REPO_NAME)-test
```

Two VMs in the same repo (e.g., testing two recipient configs
simultaneously in `nous`, or "Emma's view" alongside "operator's
view" of a shared brain) require the operator to override `TART_VM`
fully — and the name is operator-invented, no convention. Easy to
forget cleanup, easy to collide with a teammate's mental model.

The 90% case is still "one VM per repo" — `make tart` should keep
working unchanged. We just want a friendly knob for the parallel
case that picks a name we can talk about.

## Done when

- `.tart/Makefile` derives `TART_VM` as
  `$(REPO_NAME)-$(or $(TART_SUFFIX),test)`. Empty/unset
  `TART_SUFFIX` → `<repo>-test` (current behavior, unchanged).
  `TART_SUFFIX=emma` in the `pair` repo → `pair-emma` (not
  `pair-test-emma`).
- Header block + `help-tart` mention `TART_SUFFIX` as the
  ergonomic knob, with `pair-emma` as the example.
- Direct `TART_VM=…` override still works (covers full-control
  cases the suffix shorthand doesn't model).
- `make tart` with no env vars in a consuming repo still
  produces `<repo>-test` — verified by reading the resolved
  variable, not by booting a VM (cheap to check).

## Spec

Single-line change to the `TART_VM ?=` derivation, plus
documentation. The `$(or X,Y)` make function returns `X` when
non-empty else `Y`, which is exactly the semantics we want:

```make
TART_VM ?= $(REPO_NAME)-$(or $(TART_SUFFIX),test)
```

`?=` is retained so direct `TART_VM=...` override on the
command line still bypasses the suffix logic entirely.

Cleanup notes for the operator (documented in help-tart):

- Per-suffix VMs accumulate independently. `make tart-stop
  TART_SUFFIX=emma` stops just that one; `make tart-clean
  TART_SUFFIX=emma` removes its 4–8 GB clone.
- `_tart_check_others` (the "stop other VMs?" prompt) already
  keys off `$TART_VM`, so each suffix is its own pipeline and
  the prompt offers to stop the others when you boot one.

## Plan

- [x] M1: Update `.tart/Makefile`:
      - Change `TART_VM ?= $(REPO_NAME)-test` to use `$(or)`.
      - Update header comment block (the `# Auto-derived per
        repo` section).
      - Update `help-tart` to mention `TART_SUFFIX`.
- [x] M2: Verify resolution via inline make probe in nous:
      default → `nous-test`, `TART_SUFFIX=emma` → `nous-emma`,
      `TART_VM=foo` → `foo`. `make -n -p` was unsuitable because
      `?=` prints the unresolved expression — switched to
      `include` + `show:` target for actual evaluation.
- [x] M3: `make refresh` in nous picked up the new Makefile.
      Other ariadne-styled consumers (charon, pair, parley.nvim)
      will refresh on their own next session — non-blocking.

## Out of scope

- Auto-discovery of "running VMs for this repo" (would need
  a tart-list grep helper). The `_tart_check_others` prompt
  already shows other VMs on every boot, which covers the
  visibility need.
- Renaming `-test` to something else by default. It's the
  established name; existing VMs would orphan on rename. The
  suffix knob is purely additive.

## Log

### 2026-05-19 — landed

Single-line change to `TART_VM ?=` plus two doc updates (header
block + help-tart). `$(or $(TART_SUFFIX),test)` evaluates to
`test` when the env var is empty/unset, `emma` (or whatever) when
set — exactly the "additive, no behavior change at default" shape
we wanted.

Verification (in nous after `make refresh`):

| invocation                  | resolved TART_VM |
|-----------------------------|------------------|
| `make tart`                 | `nous-test`      |
| `TART_SUFFIX=emma make tart`| `nous-emma`      |
| `TART_VM=foo make tart`     | `foo`            |

Probe used: `make --no-print-directory -f <(...) show` with an
inline `show:` recipe that echoes `$(TART_VM)`. `make -n -p`
prints the unresolved expression (because `?=` is recursive
expansion), so it doesn't actually answer "what name will tart
get?" — worth remembering for future make-variable verifications.

Pushed to nous via `make refresh`. Other consumers (charon, pair,
parley.nvim) will pick it up on their next refresh cycle; no
forced sync — the change is purely additive and breaks nothing.
