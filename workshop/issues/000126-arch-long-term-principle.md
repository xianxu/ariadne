---
id: 000126
status: open
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-25
estimate_hours:
---

# ARCH-PURPOSE: serve the issue's actual purpose — don't settle for the easy win; single-source ⇒ consumers derive (ARCH-* registry, checked at-plan + at-review)

## Problem

Agents are weakest at architecture / long-horizon judgment (the `ARCH-*` charter, #75).
#122 exposed a specific failure mode the existing principles (`ARCH-DRY`, `ARCH-PURE`,
Simplicity-First, Root-Cause) don't cover: at the close I **settled for the easy win** —
wired one consumer (sdlc Go) + enforcement and silently deferred the consumers that *were
the purpose* (parley Lua, the help-text prose) as "follow-up", leaving `issue.cue` as
just-documentation those surfaces don't derive from. I did this *despite repeatedly
articulating the risk* — proving "principle in my head" is not a mechanism. The durable
fix (per the operator) is to inject the posture as checkable taste at several stages, not
rely on remembering it at the close.

## Spec

Add an `ARCH-*` principle — marker **`ARCH-PURPOSE`** (decided: purpose, not
future-proofing — dodges the gold-plate misread `ARCH-LONG-TERM` risked) — to the
registry (`cmd/sdlc/internal/judge/architecture.md`), riding #75's existing injection into
`start-plan` (planning) + `change-code` (plan-quality judge) + `milestone-close`/`close`
(boundary review). The principle, roughly: *serve the issue's actual purpose / the
project's long-term goal; don't settle for the easy subset; a single-source / "compiled to
consumers" effort is not done until every consumer **derives** from the source (the source
is enforced, not documentation). "Follow-up" is for separable extensions, never for the
thing that is the point.*

Lenses (the multi-stage value):
- **at-plan** (highest leverage — catch it while cheap): flag a plan that wires the easy
  subset and defers the rest as "follow-up" when the rest is the purpose; does the scope
  match the issue's stated goal / Done-when?
- **at-review / at-close**: does the diff/close *fulfill* the purpose or settle? For a
  single-source change, the **shadow-sweep** is this principle's concrete lens — enumerate
  consumers, confirm each derives, no remaining hand-maintained restatements of the model.

**Wording constraint:** must NOT read as "build for the future / gold-plate" — that
collides with Simplicity-First/YAGNI. This is the opposite axis: *don't **under**-deliver
the stated purpose.* The marker + statement must disambiguate (this is why `ARCH-PURPOSE`
may be the clearer name).

Companion: the human-narrative half in AGENTS.md "Core Design Principles" + the
registry↔narrative **drift test** (the #75 / [[one-referenced-contract]] pattern).

## Done when

- The registry has the new `ARCH-*` entry (with `at-plan` + `at-review` lenses) and it is
  delivered into the `start-plan` / plan-quality / boundary-review prompts (verified it
  appears).
- AGENTS.md carries the companion narrative; a drift test keeps registry ↔ narrative in sync.
- The marker + statement are disambiguated from Simplicity-First (no "gold-plate" reading).
- Dogfood: a plan that wires the easy subset and defers the issue's *purpose* as "follow-up"
  is flagged by the principle at plan-quality (or at-plan).

## Plan

- [ ] Design at `start-plan`: final marker (`ARCH-LONG-TERM` vs `ARCH-PURPOSE`); the principle statement (disambiguated from YAGNI); the at-plan vs at-review lens text
- [ ] Add the entry to `architecture.md`; confirm injection at start-plan / change-code / boundary-review
- [ ] AGENTS.md companion narrative + the registry↔narrative drift test
- [ ] Dogfood against a purpose-deferring plan

## Log

### 2026-06-25

- Filed as the durable encoding of the #122 closing-discipline lesson (operator: "how do we
  avoid this?" → make it an `ARCH-*` principle, multi-stage, not a narrow shadow-sweep-at-close
  mechanic). Builds on #75 (the ARCH-* injection mechanism). Motivating case: #122 (settled
  for the easy win; deferred the consumers that were the point). See `workshop/lessons.md`
  "A single-source issue isn't DONE until every consumer DERIVES".
