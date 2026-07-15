---
id: 000180
status: working
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours:
started: 2026-07-15T15:12:34-07:00
---

# project vocabulary model: schematize project like issue (cue + lifecycle + processes)

## Problem

`project` is a datatype in prose only (`construct/datatype/project.md`:
frontmatter table, `status: active|paused|done|dropped`, done_when, MVP
scope, single-operator discipline) — it is NOT a vocabulary model.
`construct/vocabulary/` holds issue.cue / pensive.cue / verdict.cue; issue's
status enum + lifecycle are formally modeled and ENFORCED (sdlc reads them
via `pkg/vocab`: set-status transition guards, the compiled done-guard,
discovery dirs feeding resolve). Project has none of that:

- `sdlc close`'s project gate parses project files by convention (task tick
  + detail-block upsert against an unchecked shape);
- no gate validates a project instance's conformance (issues get the
  instance-conformance gate at merge; projects get nothing);
- the status enum exists only as a markdown table nothing reads — a
  hand-maintained restatement, not an enforced source (the exact
  ARCH-PURPOSE gap the issue vocabulary work closed for issues).

This matters more now: #171 lifts project management into the sdlc spine
(project files move into coding repos; the close gate resolves them across
peers; project verbs join sdlc). A lift onto an unschematized noun bakes the
prose-only shape into code.

## Spec

Candidates (settle at design time):

- **`construct/vocabulary/project.cue`** — the formal model: status enum +
  lifecycle transitions (active → paused/done/dropped; what reopens?),
  required frontmatter fields, and a `discovery:` block (interesting twist
  vs issue: project discovery is CROSS-REPO by #171 — the model may encode
  "glob data/project/*.md across peer repos" or defer location entirely to
  resolution).
- **`pkg/vocab` accessor** (`vocab.Project()`) mirroring `Issue()` — the
  single source sdlc consumers derive from; no hardcoded enums.
- **Processes around it** (the #171 lift consumes this model): project
  verbs on the spine (new/list/show/set-status at minimum; tick semantics
  at close), transition guards matching the lifecycle, and instance
  conformance (validate frontmatter + status against the cue — the same
  gate class issues get at merge).
- **Prose doc derives, not duplicates:** `construct/datatype/project.md`
  keeps the human narrative but cites the cue as the schema authority
  (procedure refers, registry defines — lessons.md #69 Rule 2), with a
  drift test binding the two (the #70 one-referenced-contract pattern).

Out of scope (own tickets later): the sibling datatypes `product` and
`roadmap` deserve the same lift; do project first — it is the one the sdlc
spine touches.

Related: #171 (the residency/navigation/close-gate half — its design should
consume this model; soft ordering: cue model first or together, since the
gate lift wants typed parsing, not more convention).

## Done when

- `construct/vocabulary/project.cue` models the project noun (fields,
  status enum, lifecycle) and `pkg/vocab` exposes it; no consumer hardcodes
  the enum.
- `sdlc close`'s project-file update parses/validates against the model
  (typed records, not substring convention — lessons.md #167).
- A project instance failing conformance is caught by a gate (which gate —
  merge instance-conformance vs close — is a design decision).
- `construct/datatype/project.md` cites the cue as schema authority; a
  drift test binds prose table ↔ model.
- xx-vocabulary skill's claim ("the system's nouns are formally modeled in
  construct/vocabulary/*.cue") becomes true for project.

## Plan

- [ ] brainstorm/design at start-plan: cue shape (esp. cross-repo
      discovery), lifecycle transitions, which gate owns conformance,
      ordering vs #171's lift

## Log

### 2026-07-15

Filed from the #171 thread (operator): "is project a datatype? we should
lift it to be properly schematized just like issue and think about
processes around it." Current state verified: datatype prose exists,
vocabulary model does not; sdlc's project gate parses by convention.
