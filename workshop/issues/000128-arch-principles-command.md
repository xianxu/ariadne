---
id: 000128
status: working
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-25
estimate_hours: 1.66
started: 2026-06-25T11:25:20-07:00
---

# sdlc arch-principles command — single-source the ARCH-* narrative; replace AGENTS.base.md abstract with a command, keep the start-plan push

## Problem

`AGENTS.base.md` "Core Design Principles" hand-maintains a one-line abstract of each
`ARCH-*` marker (DRY/PURE/PURPOSE) plus a file-path pointer to the registry. That abstract
is a **parallel restatement** of `architecture.md`'s `principle:` lines, bound only by a
*presence-check* drift test (`TestArchitecture_NarrativeInSyncWithAgentsMd` asserts each
marker string appears) — the prose itself can drift unguarded. This is an `ARCH-PURPOSE`
smell (#126) on #126's own narrative: a consumer that *restates* the model instead of
*deriving* from it.

Two delivery facts frame the fix:
- The registry is already **pushed** at the gates: `sdlc start-plan` prints the `at-plan`
  lens to the main thread; the plan-quality + boundary-review judges embed it inline. The
  judge prompts are **fresh-context subagents** — they MUST keep the inline embed
  (`architecture.go`: "a marker alone would be a dangling pointer in a fresh-context
  subagent — the definitions must be co-present"). That embed is non-negotiable.
- The gap pure-injection leaves: **non-gate work** (§7 autonomous bug-fixing, trivial
  fixes, Q&A, in-session iteration) never runs `start-plan`, so it never sees ARCH-* —
  today its only source is the static abstract.

## Spec

Add a thin `sdlc arch-principles` command that prints the registry (the shared render
primitive `judge.ArchitectureBlock`, which `start-plan` already calls). Then collapse the
`AGENTS.base.md` abstract to a single instruction to run it.

**Deliberately NOT doing** (decided with the operator): do **not** remove the push at
`start-plan`. A printed-at-gate block is a deterministic **push** (lands in context whether
or not the agent asks); "AGENTS.base.md says run the command" is a model-dependent **pull**.
Replacing the push with a pull re-introduces the exact "principle in my head is not a
mechanism" failure #126 was filed against, and trades the cheap *at-plan* forward catch for
an expensive late plan-quality rejection. So: keep the push, ADD the command as (a) the
standalone non-gate pull and (b) a tested CLI consumer of the one registry.

Changes:
1. `cmd/sdlc/archprinciples.go` — `NewArchPrinciplesCmd()`: prints `judge.ArchitectureBlock(lens)`;
   `--lens at-plan|at-review` flag, default `at-plan`. (start-plan is untouched — it already
   calls the same `ArchitectureBlock` function; the real DRY is that shared primitive, the
   command is just a second entry point to it.)
2. `cmd/sdlc/helptext/arch-principles.md` — Long help (`MustGet` panics without it).
3. `cmd/sdlc/main.go` — register via `add(NewArchPrinciplesCmd(), "arch-principles", "<short>")`.
4. `AGENTS.base.md` "Core Design Principles" — drop the per-marker abstracts + the file-path
   pointer; replace with: *architectural taste (ARCH-*) is single-sourced and delivered by
   `sdlc arch-principles` (also pushed at start-plan + the plan-quality/review judges); run it
   when designing/reviewing or for non-gate work; cite the marker.* Keep **Simplicity First**
   + **Root Cause** in full (no registry home). Regenerate `AGENTS.md` via `make weave`.
5. `cmd/sdlc/internal/judge/judge_test.go` — repurpose `TestArchitecture_NarrativeInSyncWithAgentsMd`:
   the narrative no longer enumerates markers, so the invariant becomes "AGENTS.md routes to
   `sdlc arch-principles`" (assert that string present) instead of per-marker presence. The
   marker set stays guarded as a derived thing by the command's own test + the existing
   `TestArchitectureMarkers`/`TestArchitectureRegistry_Content`.
6. New `cmd/sdlc` test: `sdlc arch-principles` output contains ARCH-DRY/PURE/PURPOSE + the
   at-plan lens header; `--lens at-review` renders the at-review framing. (The command is now a
   tested consumer that derives from the registry — ARCH-PURPOSE.)
7. `atlas/workflow/sdlc-binary.md` — document the new command + the push/pull rationale.

## Done when

- `sdlc arch-principles` prints the full registry (markers + lens); `--lens at-review` switches
  the framing; verified in a test.
- `AGENTS.base.md` carries no per-marker abstract — only the command instruction + the two
  non-registry principles; `AGENTS.md` regenerated and the drift test asserts the route-to-command
  invariant (green).
- `start-plan` still pushes the at-plan block (unchanged); the judge prompts still embed inline
  (unchanged) — verified the existing tests still pass.
- Full `go test ./cmd/sdlc/...` green.

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: smaller-go-module    design=0.1 impl=0.4
item: atlas-docs           design=0.1 impl=0.4
item: milestone-review     design=0.0 impl=0.6
design-buffer: 0.30
total: 1.66
```

smaller-go-module = the new command + helptext (mirrors start-plan's cobra shape). atlas-docs =
the AGENTS.base.md narrative rewrite + atlas note + the test reconciliation (folded into impl).
milestone-review = the one close-time review (single-pass atomic, no Mx). NOTE: estimate-logic-v2
is producing ~2.3× high vs ship-wall-clock actuals (#127) — this estimate is the honest v2 build-
effort number, deliberately NOT hand-compensated, so the ledger keeps a clean calibration point
for #127.

## Plan

- [ ] `cmd/sdlc/archprinciples.go` + `helptext/arch-principles.md`; register in `main.go`. Test: command output renders the registry (markers + at-plan lens) and `--lens at-review` switches framing.
- [ ] Rewrite `AGENTS.base.md` Core Design Principles: drop the abstracts + path, add the `sdlc arch-principles` instruction, keep Simplicity-First/Root-Cause. `make weave` → regenerate `AGENTS.md`.
- [ ] Repurpose `TestArchitecture_NarrativeInSyncWithAgentsMd` to the route-to-command invariant; confirm `start-plan` + judge-embed tests still pass unchanged.
- [ ] Atlas note in `atlas/workflow/sdlc-binary.md` (new command + push/pull rationale).
- [ ] Full `go test ./cmd/sdlc/...` green; spot-check `sdlc arch-principles` output by hand.

## Log

### 2026-06-25

- Filed from the #126 follow-up design conversation (operator). Supersedes the earlier
  "derive the narrative via weave" framing: a command is the better mechanism — it serves the
  non-gate pull AND is a tested CLI consumer, where weave-deriving the prose only solved the
  always-on copy. Key decision recorded in Spec: **keep the start-plan push** (deterministic)
  and only ADD the command (pull + standalone); do NOT move ARCH-* out of start-plan, and the
  fresh-context judge prompts keep their inline embed. Builds on #75 (the ARCH-* mechanism) +
  #126 (ARCH-PURPOSE, which this dogfoods on its own narrative).
