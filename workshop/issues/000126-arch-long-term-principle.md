---
id: 000126
status: done
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-25
estimate_hours: 1.36
started: 2026-06-25T10:36:32-07:00
actual_hours: 0.54
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
the stated purpose.* The marker + statement must disambiguate — hence the chosen `ARCH-PURPOSE` (purpose, not
future-proofing).

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

## Estimate

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: atlas-docs           design=0.2 impl=0.5
item: milestone-review     design=0.0 impl=0.6
design-buffer: 0.30
total: 1.36
```

Build effort is doc-shaped: the registry entry + AGENTS.base.md bullet + atlas note are
prose (atlas-docs); the Go change is only extending 3 hardcoded marker fixtures in
judge_test.go (trivial — folded into impl). design≈0.2 because the Spec pre-resolves the
marker, lenses, and YAGNI-disambiguation constraint — the residual design is matching the
ARCH-DRY/ARCH-PURE house style. milestone-review covers the one close-time fresh review
(single-pass atomic, no Mx).

## Plan

- [x] Add the `ARCH-PURPOSE` entry to `cmd/sdlc/internal/judge/architecture.md` (3rd, after PURE: `principle:` / `at-plan:` / `at-review:`, YAGNI-disambiguated). Injection into start-plan / plan-quality / boundary-review prompts is **automatic** via the #75 mechanism (`ArchitectureBlock`, `{{ARCH_STAR}}`) — no prompt edits; verify it appears.
- [x] Companion narrative bullet in `AGENTS.base.md` "Core Design Principles" (the *source*; `AGENTS.md` is weave-generated — regenerate via `make weave`). The registry↔narrative drift test **already exists** (`TestArchitecture_NarrativeInSyncWithAgentsMd`, #75) and auto-derives from `ArchitectureMarkers()`, so it now *requires* AGENTS.md to mention ARCH-PURPOSE — no new test, it just extends. (ARCH-DRY: reuse the existing guard, don't add a parallel one.)
- [x] Extend the 3 hardcoded marker fixtures. Only ONE self-fails from a registry-only change: `TestArchitectureMarkers` (`len != want` → red) — plus the existing drift test `TestArchitecture_NarrativeInSyncWithAgentsMd` (red until AGENTS.md mentions ARCH-PURPOSE). The other two are deliberate **strengthening** edits (the old assertions stay green by substring/positive-Contains luck — dogfooding ARCH-PURPOSE: make each consumer test *prove it derives* the new marker, don't settle for the accidental green): `TestCodeReviewBody_Renders` (`"ARCH-DRY, ARCH-PURE"` → `"ARCH-DRY, ARCH-PURE, ARCH-PURPOSE"`), `TestArchitectureRegistry_Content` (add `"ARCH-PURPOSE"` to want).
- [x] Atlas note in `atlas/workflow/sdlc-binary.md` (parallels the existing "#71 adds ARCH-SHIM" line).
- [x] `make weave` (regenerate AGENTS.md) + `go test ./cmd/sdlc/...` green (drift test passes).
- [x] Dogfood: run the plan-quality judge against a purpose-deferring fixture plan; confirm it flags it and cites `ARCH-PURPOSE`.

## Log

### 2026-06-25
- 2026-06-25: closed — Full `go test ./cmd/sdlc/...` green incl. the auto-deriving drift test (now requires AGENTS.md mention of ARCH-PURPOSE) + the 3 strengthened marker fixtures. `sdlc start-plan` empirically delivers "3 entries" incl. ## ARCH-PURPOSE (injection verified). Live dogfood: plan-quality judge returns VERDICT: FAILURE citing ARCH-PURPOSE on a purpose-deferring fixture plan (the #122 failure caught at plan time). Atlas updated (sdlc-binary.md).; review verdict: SHIP

- Filed as the durable encoding of the #122 closing-discipline lesson (operator: "how do we
  avoid this?" → make it an `ARCH-*` principle, multi-stage, not a narrow shadow-sweep-at-close
  mechanic). Builds on #75 (the ARCH-* injection mechanism). Motivating case: #122 (settled
  for the easy win; deferred the consumers that were the point). See `workshop/lessons.md`
  "A single-source issue isn't DONE until every consumer DERIVES".
- **start-plan design (2026-06-25):** mapped the consumer set so this change *derives* rather
  than settles (dogfooding ARCH-PURPOSE on itself). Single source = `architecture.md`; the
  consumers that must derive — all automatic via #75 — are: the plan-quality prompt
  (`ArchitectureBlock("at-plan")`), `sdlc start-plan` delivery, the milestone-review/dry/pure
  prompts (`at-review`), `code-review.md`'s `{{ARCH_STAR}}` enumeration, and the AGENTS.md
  narrative (via the auto-deriving drift test). The narrative *source* is `AGENTS.base.md`
  (not `AGENTS.md`, which weave generates + gitignores); downstream entry files (brain
  `CLAUDE.md`, etc.) re-derive on their next weave compose — a consumer that derives, so no
  hand-edit there either. ARCH-DRY: the entire injection + drift-guard machinery is reused;
  the only net-new is one registry entry + one narrative bullet + 3 one-line fixture edits.
- **change-code judges consumed (2026-06-25):** plan-quality = INFO ("safe to start",
  "architecturally exemplary" ARCH-DRY). Its Finding 1 correctly caught that only 1 of the 3
  fixture edits self-fails — folded that into Plan step 3 (the other two are strengthening
  edits so each consumer test *asserts* the marker, not accidental-green). estimate-quality =
  advisory INFO only (test-work folded into atlas-docs, dogfood not itemized) — both judged
  "within tolerance / not material to the 1.36 total"; estimate left as-is (itemizing to
  placate would be the back-fitting that judge warns against). Re-ran `change-code --no-judge`
  having consumed both.
- **Implemented (2026-06-25):** TDD — extended the 3 marker fixtures red first (only
  `TestArchitectureMarkers` self-failed; the strengthening edits to `TestCodeReviewBody_Renders`
  + `TestArchitectureRegistry_Content` make each consumer test *assert* ARCH-PURPOSE), added
  the registry entry (which then turned the auto-deriving drift test red), added the
  `AGENTS.base.md` bullet, `make weave` regenerated `AGENTS.md` → all green. Verified injection
  empirically: `sdlc start-plan` now delivers **"3 entries"** including `## ARCH-PURPOSE`. Full
  `go test ./cmd/sdlc/...` green. Atlas note added (distinguishing the landed ARCH-PURPOSE from
  the still-planned ARCH-SHIM, since #71 — its owner — is open).
- **Dogfood PASSED (2026-06-25):** built the *real* plan-quality prompt (via the live
  `BuildPrompt(PlanQuality, …)`, throwaway test in scratchpad — not committed) around a
  purpose-deferring fixture (single-source a retry policy across 3 callers; Plan wires 1 +
  defers the other 2 as "follow-up"). The live judge returned **VERDICT: FAILURE** and flagged
  it as an explicit **ARCH-PURPOSE violation**: *"The 'follow-up' here is the deferred point of
  the issue, not an extension — flagged exactly per the at-plan lens"*, and ran the shadow-sweep
  ("consolidation isn't done until every consumer derives"). The #122 failure mode is now caught
  at plan time. This is an LLM-nondeterministic confirmation-of-wording (the deterministic
  injection + drift guarantees are test-covered), so it's evidence, not a CI gate.
