---
id: 000075
status: working
deps: []
github_issue:
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 3
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
   The workflow has no "entering planning" transition today (`claim` = start
   work; `change-code` = the plan-quality *review* gate, which is too late — the
   design is already made). Add a new verb **`sdlc start-plan`** as the *forward*
   planning-entry injection: it renders `architecture.md`'s `at-plan` lens to the
   agent's main thread so the design accounts for the principles from the start.
   The agent runs it on entering plan mode (AGENTS.md §2 workflow instructs it);
   it delivers each time a new design begins (no warmup-suppress — re-delivery
   per planning session is desired, since agents don't reread a static doc).
   This pairs with `change-code`'s plan-quality judge (the *backward* check
   against the same `at-plan` lens): forward inject + backward review, one file.
   The exact verb name (`start-plan` vs `plan`) is a change-code bikeshed.
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
- A new **`sdlc start-plan`** verb delivers the `at-plan` lens to the main thread
  at planning entry; AGENTS.md §2 (workflow) adds the `claim → start-plan →
  change-code` step and §6 points at the registry.
- `sdlc judge dry`/`pure` are re-expressed as registry entries (or clearly
  deprecated in favor of the folded-in review), so a principle is authored once.
- Adding a new `ARCH-<NAME>` entry flows into all three consumers with no other
  edits.

## Plan

Decision: **embed** (`//go:embed`, ariadne-owned, simplest); derivative-override
via a runtime `construct/` doc is a deferred refinement.

- [x] M1 — registry + judge wiring: author `architecture.md` (ARCH-DRY, ARCH-PURE
  with principle/at-plan/at-review) + `//go:embed` into the judge package; render
  the `at-plan` lens into the plan-quality prompt and the `at-review` lens into
  the boundary/milestone-review prompt; fold `judge dry/pure` in; tests (embed
  presence in both prompts; markers present; à la #70's pattern).
- [x] M2 — `sdlc start-plan` verb + workflow: new verb rendering the `at-plan`
  lens to the main thread at plan-entry; wire AGENTS.md §2 (`claim → start-plan →
  change-code`) + §6 pointer; verb-registration + render tests.

## Relationships

- **#69** (unify the boundary review, binary-owned) consumes this registry's
  at-review lens — soft dependency (do this first, #69 builds on it).
- **#71** (external-service shims) adds `ARCH-SHIM` as a registry entry — this is
  the mechanism that makes #71 enforceable at plan + review time.

## Log

### 2026-06-03 — M2

- New verb **`sdlc start-plan`** (`startplan.go` + `helptext/start-plan.md`,
  registered in main.go after `claim`): delivers `judge.ArchitectureBlock("at-plan")`
  + a planning framing to the main thread — the *forward* injection at design
  time. Exported `architectureBlock` → `ArchitectureBlock` for cross-package use.
- AGENTS.md: §2 workflow adds the `claim → start-plan → change-code` step; the
  Core Design Principles section now declares it the human narrative + points at
  the `ARCH-*` registry, with `(ARCH-DRY)`/`(ARCH-PURE)` markers on the entries.
- **Drift guard** (deferred from M1, per the fresh-eyes review):
  `TestArchitecture_NarrativeInSyncWithAgentsMd` parses the `ARCH-<NAME>` markers
  from the registry and asserts each `<NAME>` appears in AGENTS.md + that AGENTS.md
  points at `architecture.md` — so adding `ARCH-SHIM` later forces the narrative to
  keep up (mirrors #70's doc-sync test).
- Tests: `startplan_test.go` (registration + renders the at-plan lens, labeled +
  generic). Live: `sdlc start-plan --issue 75` prints the at-plan principles.
  `go test ./cmd/sdlc/...` + `go vet` + gofmt green.

### 2026-06-03 — M1
- 2026-06-03: closed M1 — architecture.md registry (ARCH-DRY/ARCH-PURE) //go:embed'd + rendered into plan-quality (at-plan), milestone-review (at-review), dry/pure (from registry); TestArchitectureRegistry_* + embed-routing tests + updated DRY test pass; go test+vet+gofmt green; atlas noted. One file → all consumers. actual=judgment; review verdict: FIX-THEN-SHIP

- **Decisions (from plan-quality INFO):** (a) DRY/PURE stay as `Category`
  constants (no `AllCategories` churn → no enumeration-test churn) but their
  prompts now **render from the registry** (embed `ArchitectureRegistry`, focus
  the relevant marker) — so the principle is authored once. (b) The at-review
  lens targets the **existing `MilestoneReview`** prompt now (#69 inherits it
  later). (c) This is intentional **base-layer churn** (judge package + prompts
  propagate to derivatives via the shared sdlc binary).
- `cmd/sdlc/internal/judge/architecture.md` — `ARCH-DRY`, `ARCH-PURE`, each with
  `principle` / `at-plan` / `at-review`. `architecture.go` `//go:embed`s it as
  `ArchitectureRegistry` + an `architectureBlock(lens)` renderer.
- Wired into 4 prompts: **PlanQuality** (at-plan — "highest-leverage, design still
  changeable"), **MilestoneReview** (at-review backstop), **DRY**/**PURE** (render
  ARCH-DRY/ARCH-PURE from the registry instead of hand-written prose).
- Tests: `TestArchitectureRegistry_Content` (markers + both lenses present),
  `TestArchitectureRegistry_EmbeddedInPrompts` (all 4 prompts embed the one file;
  lens labels reach the right consumers), updated `TestBuildPrompt_DRY`. Also
  tidied a pre-existing orphaned doc-comment (#70 M2 leftover). `go test
  ./cmd/sdlc/...` + `go vet` + gofmt green.

### 2026-06-03
