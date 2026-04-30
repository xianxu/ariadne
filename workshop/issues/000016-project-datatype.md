---
id: 000016
status: open
deps: [000012, 000015]
created: 2026-04-29
updated: 2026-04-29
references: [/Users/xianxu/workspace/brain/memory/life/42shots/ideas/2026-04-28-02-pensive-ariadne-arc.md, workshop/issues/000015-product-and-roadmap-data-types.md]
---

# project datatype

Add a `project` typed-document prototype to ariadne's datatype system. A project is the *execution container* — "what we've decided to do for a purpose, with an MVP scope, sequenced and tracked day-to-day." Distinct from `product` (durable charter) and `roadmap` (month-level aggregate). Operator-POV, time-bounded, cuts across multiple products and repos.

This is the missing piece between issues (units of work) and roadmaps (month-level targets) — and it's where the velocity calibration loop closes (each completed entry records `actual_hours`).

## Done when

- `construct/datatype/project.md` exists, conforming to the meta-prototype contract.
- Prototype ships with `Search recipes` and authoring instructions sufficient for a fresh agent to author good instances unaided.
- Dogfooded by authoring `data/project/charon-release-push.md` in brain — the immediate active project covering charon issues #13, #14, #15 (with #16 explicitly out of scope).
- `atlas/data-artifacts.md` updated with the project type.
- Decision recorded on whether/how the existing `roadmap` prototype needs adjustment to reference projects rather than product components directly. (Decision can be: "no change needed, defer to v2 of roadmap" — that's a valid outcome.)

## Spec

### Why this is needed

The product/roadmap pair from #000015 covers two of the three concerns: *what is built* (product) and *what we want true by month T* (roadmap). It misses *what we're working on right now and in what order* — the day-to-day execution view.

This isn't an issue tracker. Issues live in product repos and describe units of work that exist regardless of when they're worked on. A project is the operator's view: which units are *currently in flight*, *in what sequence*, with *what blockers*, *toward what defined MVP*.

The Facebook precedent: tasks (issues) shipped day 1; project management came years later. The two are genuinely separate concerns. Project management is downstream of issues and operator-centric; brain (the operator's substrate) is the right home.

### Project datatype — design intent

A project is a *flexible container of execution toward external value*, with a defined MVP. It declares:

- **What it's accomplishing** — a `goal` and `done_when` criterion (the MVP boundary).
- **Which issues are in scope** — referenced by `[<repo>#<id>]` syntax.
- **Which are explicitly out** — the conversation about what *isn't* included is more useful than a vague in-scope list.
- **Sequence + state** — body sections function as kanban columns (`doing`, `next`, `blocked`, `done`).

Key design points to nail down during brainstorming:

- **Frontmatter likely carries:** `type: project`, `name`, `goal`, `done_when`, `status` (active | paused | done | dropped), `mvp_scope: [refs]`, `explicitly_out: [refs]`, `created`, `updated`, lineage.
- **Body skeleton (working assumption):** title, lede line, `## scope` (MVP in/out + reasoning), `## doing` (1–3 items), `## next` (priority-ordered queue), `## blocked` (with reason), `## done` (with `actual_hours` and close-date).
- **Entry format:** `- [ ] <task description> [<repo>#<id>] (est: ~Nh, started YYYY-MM-DD)` for open items; `- [x] ... (actual: Mh, closed YYYY-MM-DD)` for done.
- **Issue references** use the same `[<repo>#<id>]` shape across the system — composes with cross-product references in roadmap and product files.
- **Per-issue milestones** can be referenced as `[charon#13 M2]` for finer granularity.
- **Single-active discipline** — schema doesn't enforce "only one active project at a time," but operator convention should. Multiple `active` projects fragment focus; project authoring instructions should flag this.
- **Default instance path:** `data/project/<slug>.md`.

### KTLO handling

KTLO (Keep The Lights On — bug fixes, oncall response, ad-hoc small work) is modeled as an **issue-priority flag**, not a separate datatype. Issues carrying `priority: ktlo` (or similar) are picked off when the active project has slack. This avoids spawning a perpetual "ktlo" pseudo-project that never closes — and keeps the project datatype clean ("a project always has a definable MVP and an end").

The exact frontmatter spelling for the KTLO flag is out-of-scope for this issue (lives in the issue conventions, not the datatype system). Just record the decision so future authoring uses it.

### Velocity calibration loop

A project is where the velocity skill's calibration discipline manifests:

- Each `## doing` / `## next` entry carries `(est: ~Nh)` mirroring the issue's `estimate_hours` frontmatter.
- On completion, the entry moves to `## done` with `(actual: Mh, closed YYYY-MM-DD)`.
- The `actual_hours` value is also written back to the issue's frontmatter and appended to `brain/data/life/42shots/velocity/estimate-logic-v1.md`'s validation table.
- This closes the loop: estimate → execute → record actuals → recalibrate.

The project datatype's authoring instructions should explicitly call out this discipline so it's not left to memory.

### Authoring flow

When the dispatcher applies the project prototype:

1. Distill from conversation: name, goal, candidate issues to include.
2. Force the MVP conversation upfront: "what's NOT in this project? what's the smallest version that delivers external value?"
3. Resolve `mvp_scope` as `[<repo>#<id>, ...]` — confirm each issue exists.
4. Set initial `## doing` (1 item, the user's first action) and `## next` (priority-ordered).
5. Confirm + write.

### Roadmap implications

The current `roadmap` prototype (from #000015) references product components directly. New framing suggests roadmap should aggregate *projects + KTLO* per month, not components.

**Decision: defer the roadmap rework to discovery during dogfood.** Land project first; author a real roadmap that references projects; if it doesn't fit, restructure roadmap as a v2. Don't pre-commit to a roadmap rework based on theory.

### Why now

The immediate trigger: tomorrow's work session needs a project to organize the charon-release-push (issues #13/#14/#15 in charon, with #16 explicitly out for scope discipline). Without a project datatype, that work either uses ad-hoc markdown or fragments across files. With one, the velocity calibration loop closes immediately on the first issue completed.

## Plan

- [ ] **Brainstorm `project` prototype** via `superpowers-brainstorming`. Settle frontmatter (especially `mvp_scope`, `explicitly_out`, `done_when`), body skeleton (kanban-style sections), entry format ([<repo>#<id>] refs, est/actual hours), and the authoring flow's MVP-discipline question.
- [ ] Decide where instances live by default — `data/project/<slug>.md` is the working assumption.
- [ ] Write `construct/datatype/project.md`.
- [ ] **Dogfood — author `data/project/charon-release-push.md`.** Real project covering charon #13 + #14 + #15 (#16 explicitly out). Use immediately to organize tomorrow's work.
- [ ] Verify the velocity calibration loop works in practice: as items in `## done` accumulate, `actual_hours` flows to issue frontmatter and `estimate-logic-v1.md`'s validation table.
- [ ] Update `atlas/data-artifacts.md` to include `project`.
- [ ] Decide on roadmap rework: does the existing `roadmap` prototype need adjustment to reference projects instead of (or in addition to) components? Record decision in `## Log` either way. If yes, file as a separate issue.
- [ ] Update `construct/datatype/roadmap.md` if the decision says yes.

## Log

### 2026-04-29

- Issue created from a design conversation reframing product vs project. Original issue 000015 collapsed product+project+infra into a single "product" datatype with no discriminator; this reframe restores `project` as a distinct execution-container datatype while keeping `product` as the durable charter.
- KTLO modeled as an issue-priority flag rather than a separate long-running pseudo-project. Cleaner.
- Trigger: needing to organize the charon-release-push (issues #13/#14/#15) for tomorrow's work session, with explicit MVP scope (charon #16 deferred).
