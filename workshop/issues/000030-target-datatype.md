---
id: 000030
status: done
deps: []
created: 2026-05-22
updated: 2026-05-27
estimate_hours: 4
actual_hours: 2.0
---

# target datatype (narrative intent above projects + issues)

Add a `target` datatype that holds the *durable narrative intent* of a piece of work, sitting above projects and issues in the dependency graph. Captures "what we want and what for," under-specified by design; let the substrate (projects, issues, agent reading) derive specifics.

Pensive: [`docs/vision/2026-05-22-01-pensive-durable-target-pattern.md`](../../docs/vision/2026-05-22-01-pensive-durable-target-pattern.md).

## Done when

- `construct/datatype/target.md` exists, defining frontmatter + body skeleton + authoring instructions. Models on `project.md` and `pensive.md`.
- `workshop/targets/` directory exists in this repo as the canonical home; folder is referenced from `AGENTS.md` §1.
- `AGENTS.md` §1 artifact hierarchy mentions targets, the dependency relationship (target → projects/issues), and the promotion-from-`## Problem` pattern.
- The `🤖{}` agent-contribution convention is documented (either as a section in `target.md` or as a separate skill/convention doc — decide during M1).
- The `xx-datatype` skill recognizes `target` as a valid type and can author a new instance.
- A first real target gets written for the shared-brain work (`nous/workshop/targets/shared-brain-infrastructure-and-ui.md` or similar) so the convention is dogfooded immediately.

## Spec

### Why this datatype now

We've been doing target-shaped work without a name for it. The `## Problem` section of any non-trivial issue is doing it implicitly. The pensive lays out the asymmetry: human accumulates the full narrative; agent reads the delta. Separating intent (target) from execution (project) from work-unit (issue) lets each evolve at its native cadence.

The motivating example, written into the pensive: the shared-brain work in nous (May 2026) was effectively one rough commitment at the top, spawning sub-issues (#30 autosave, #31 TUI list async, #32 leave, tart-VM fixes, doc sweeps). Treating that as a top-level `target` with proper discipline would have made the whole arc legible at a glance.

### Frontmatter shape (sketch — iterate during M1)

```yaml
---
type: target
slug: <kebab-case>         # filename without .md
status: active             # active | achieved | split | deferred | abandoned
created: <ISO>
updated: <ISO>
sources: [...]             # parley chats, prior pensives that crystallized into this
---
```

Notably *absent* from frontmatter (relative to project): `done_when`, `operator`, `mvp_scope`, `explicitly_out`. Targets are narrative; if you need an MVP boundary or operator assignment, you're writing a project, not a target.

### Body skeleton (sketch)

```markdown
# Target: <title>

<lede — 1-3 paragraphs of "what we want and what for". The *what for* gets most of the words.>

## Why now
<motivation — what's making this matter at this moment.>

## What this is NOT
<rule-outs — under-specifying by negation. Narrows the space without enumerating sub-features.>

## Open questions
<things explicitly not committed on yet. Distinguished from project's task list — these are intent-level uncertainties, not execution todos.>
```

### Distinct from sibling datatypes

- **`product`** — durable charter ("what is being built"). A target is more narrative-driven and may not yet have a defined product. Targets can crystallize into products as the shape firms up.
- **`project`** — execution container ("what we've decided to do, by when"). A project advances one or more targets via tracked tasks.
- **`pensive`** — captures thinking at a moment ("the moment when the thought happened"). A target captures durable commitment ("what we still want, refined over time"). Pensives can promote to targets when the operator's intent crystallizes.
- **`issue`** — leaf work unit. Issues reference targets via `target: <slug>` frontmatter when not part of a project, or live under a project (which itself references the target).

### Under-specification posture

The discipline: **only get more specific when the agent fills the gap wrong.** A target that says "create a shared brain infrastructure and user interface" doesn't need to enumerate features unless the agent's natural decomposition produces the wrong ones. Trust the substrate; let the target stay durable.

Most edits to a target should be *insertions* (refined narrative, new sub-paragraph), not *modifications* (changed wording). Modifications mean intent shifted; insertions mean intent crystallized further.

### The `🤖{}` agent-contribution convention

Inline markers for agent-proposed additions or refinements to a human-centric document:

- **Accept**: delete `🤖{` and `}`, leaving the text in place.
- **Reject**: delete the whole `🤖{...}` block.
- **Edit**: change the content inside the braces, then accept.

Each is a one-edit-class operation; the resulting diff IS the contract change, atomically reviewable. The convention generalizes to any human-centric document where the agent should contribute proposals rather than overwrite — atlas drafts, product files, etc. Decide during M1 whether to document this in `target.md` itself or extract to a separate convention doc.

### Lifecycle / status semantics

Targets transition on *intent* boundaries, not execution states:

- `active` — current commitment.
- `achieved` — the world now looks the way the target wanted.
- `split` — single target broke into two because they turned out to be separable. Both halves get their own files; original gets pointer to the splits and goes `split`.
- `deferred` — still want it, eventually.
- `abandoned` — no longer want it. Kept for context; doesn't get deleted.

Distinct from project's `active | paused | done | dropped` execution state machine.

## Plan

- [x] **M1 — Datatype spec.** Done (fdbb1ce). `construct/datatype/target.md` written modeling on `project.md` + `pensive.md`. 🤖{} convention documented in target.md (decided during M1). `workshop/targets/` created.
- [x] **M2 — Wire into hierarchy.** Done (fdbb1ce). AGENTS.md §1 updated with target in artifact hierarchy + 🤖{} convention as a §1 rule. Note: §1 framing was further refined to backward-looking ("what we defend and why") during the SDLC walk-through — see commit 11cf885 + ariadne#38.
- [x] **M3 — Skill integration.** No-op — `xx-datatype` auto-discovers types by globbing `construct/datatype/*.md` and matching on `description:` field. Confirmed during M1 implementation.
- [x] **M4 — Dogfood + close.** Done (e94a78a). First real target written at `nous/workshop/targets/shared-brain-infrastructure-and-ui.md`. Three in-flight nous issues (#30 autosave, #31 TUI list async, #32 leave) backfilled with `target: shared-brain-infrastructure-and-ui`. Verified via `rg "^target: shared-brain-infrastructure-and-ui"`.

## Log


- 2026-05-27: closed — construct/datatype/target.md written (fdbb1ce); workshop/targets/ created; AGENTS.md §1 updated with target hierarchy + 🤖{} convention; first target dogfooded at nous/workshop/targets/shared-brain-infrastructure-and-ui.md; three in-flight nous issues backfilled with target: frontmatter (e94a78a). Work landed 2026-05-22; formal close delayed.
- 2026-05-22: opened. Comes from a conversation about target file pattern; the durable target pensive was written first, then this issue spawned from it. The pensive captures the *why*; this issue is the execution plan to land it.
- 2026-05-22: M1-M3 complete in one commit (`fdbb1ce`): `construct/datatype/target.md` written modeling on `project.md` + `pensive.md`; `workshop/targets/` created; `AGENTS.md` §1 updated with target in artifact hierarchy + 🤖{} convention as a §1 rule. M3 (xx-datatype recognition) is a no-op — the skill auto-discovers types by globbing `construct/datatype/*.md` and matching on `description:` field. No skill code change needed.
- 2026-05-22: M4 dogfood — first real target written at `nous/workshop/targets/shared-brain-infrastructure-and-ui.md`. Three in-flight nous issues (#30 autosave, #31 TUI list async, #32 leave) backfilled with `target: shared-brain-infrastructure-and-ui` frontmatter. Dependency graph verified via `rg "^target: shared-brain-infrastructure-and-ui" workshop/issues/` — all three issues resolve cleanly.
- 2026-05-22: ready to close. The pattern is dogfooded and the convention is documented.
