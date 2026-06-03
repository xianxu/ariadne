---
id: 000075
status: open
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours:
---

# architecture.md — a referenced architectural-principles registry delivered to planning + reviews

## Problem

Agents are strong at **tactics** (clean function, handled edge case) and weak at
**architecture** (whole-board shape — will this design still be sound months
out?). The likely reason: architectural payoff is so far downstream that there's
little training signal for it, so the model can't have learned good taste here.
So architecture is the one place we should *inject human judgment* as an explicit,
persistent, prompt-level scaffold rather than hope the model supplies it.

Today the architectural principles are **scattered and under-delivered**:
- AGENTS.md "## Core Design Principles" (DRY, PURE, Simplicity, Root Cause) —
  human prose, never structured for machine consumption.
- `sdlc judge dry` / `sdlc judge pure` — two standalone judge prompts, each
  re-stating a principle, run as separate passes (not folded into the boundary
  review).
- The plan-quality judge checks executability but **not** architecture — yet
  planning is exactly where architecture is *decided*, so it's the highest-
  leverage place to inject these and the one currently missing them.

There's no single place to author an architectural principle once and have it
reach planning, plan-review, and code-review. (#71's "shim every external
service" is a future principle that needs exactly this to become enforceable
rather than aspirational.)

## Spec

A single **markered registry** — `cmd/sdlc/internal/judge/architecture.md`,
`//go:embed`'d — that is the machine-readable companion to AGENTS.md's prose
Core Design Principles. Each entry:

```
ARCH-<NAME>
  principle:  <one-line rule>
  at-plan:    <what to check when DESIGNING — flag a plan that ignores it>
  at-review:  <what to check in a DIFF — flag a violation>
```

- **Stable semantic markers** (`ARCH-DRY`, `ARCH-PURE`, `ARCH-SHIM`) — no
  ordinals (numbers shift on insert/reorder; the handle is cited in plans, Logs,
  and review findings months apart, so it must be stable).
- First entries: **DRY** and **PURE** (fold the standalone `sdlc judge dry/pure`
  prose in — they become registry entries, checked within the one review +
  planning, not as separate passes). `ARCH-SHIM` lands later via #71.

**Delivery to its three contexts** (one *file*, delivered into each; markers for
in-context selection/emphasis + cross-referencing — NOT bandwidth savings, since
each fresh context needs the full definitions or the marker is a dangling pointer):

1. **Main-thread planning** (highest leverage — architecture is decided here).
   Agents don't reread a file already read in-session, so deliver it
   **binary-controlled, warmup-style**: reuse the `printSemanticWarmup` /
   per-session-count pattern (close.go) so a planning-phase verb (e.g. `claim` /
   `change-code`) re-injects `architecture.md` once per session, self-suppressing.
   AGENTS.md §6 points at it.
2. **plan-quality judge** (fresh subprocess) — embed the registry; render the
   `at-plan` lens. "Does this plan respect our architecture? (cite ARCH-…)"
3. **boundary review judge** (fresh subprocess) — embed the registry; render the
   `at-review` lens. (This is what #69's unified review consumes.)

Within each context: deliver the full block once, then `check ARCH-DRY,
ARCH-SHIM` selects/emphasizes — the model resolves the marker against the
co-present definition (pairwise attention). Across contexts, each embeds its own
copy (no shared memory).

**Open question for planning:** embed (compile-time, ariadne-owned, simplest) vs
runtime-read a `construct/` doc (lets derivatives override with their own
architecture). Start embedded; add override later if wanted.

## Done when

- `cmd/sdlc/internal/judge/architecture.md` exists with `ARCH-DRY` + `ARCH-PURE`
  (principle + at-plan + at-review), `//go:embed`'d as the one source.
- The **plan-quality** prompt renders the at-plan lens; the **boundary/milestone
  review** prompt renders the at-review lens — both from the one file (no
  re-authored prose; a test pins each prompt embeds the registry, à la #70).
- The **main thread** gets `architecture.md` delivered once per session
  (warmup-style binary delivery) at the planning entry point; AGENTS.md §6 points
  at it.
- `sdlc judge dry`/`pure` are re-expressed as registry entries (or clearly
  deprecated in favor of the folded-in review), so a principle is authored once.
- Adding a new `ARCH-<NAME>` entry flows into all three consumers with no other
  edits.

## Plan

*(draft — to refine at change-code)*

- [ ] Author `architecture.md` (ARCH-DRY, ARCH-PURE; the at-plan/at-review
  template) + `//go:embed` it into the judge package.
- [ ] Render it into the plan-quality + boundary-review prompts (at-plan /
  at-review lenses); embed-presence tests.
- [ ] Warmup-style main-thread delivery at the planning entry point (reuse the
  close.go warmup mechanism); AGENTS.md §6 pointer.
- [ ] Fold `judge dry/pure` into the registry; decide standalone-verb fate.

## Relationships

- **#69** (unify the boundary review, binary-owned) consumes this registry's
  at-review lens — soft dependency (do this first, #69 builds on it).
- **#71** (external-service shims) adds `ARCH-SHIM` as a registry entry — this is
  the mechanism that makes #71 enforceable at plan + review time.

## Log

### 2026-06-03
