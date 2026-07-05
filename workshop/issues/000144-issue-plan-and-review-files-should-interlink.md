---
id: 000144
status: working
deps: []
created: 2026-06-29
updated: 2026-07-05
started: 2026-07-05T10:18:05-07:00
---

# sdlc resolve — read-only artifact-reference resolver

## Problem

ariadne artifacts (issues, plans, review sidecars, targets) refer to each other —
and across peer repos — with **symbolic** refs (`ariadne#11`, `#15 M4`, `pair#84`),
but there's no mechanism to turn a ref into the file it names. Consumers (a human in
parley, an agent, the CLI) each have to glob/guess the path.

This issue **was** framed as "files should carry stored cross-links" — rejected:
the id is stable but the path is not (slug renames; `issues/ → history/` on
close/merge, which ariadne#160 made happen on every merge), so stored links rot on
archive. The fix is **read-time resolution** — keep the symbolic ref canonical and
resolve the path on demand. This issue is the **ariadne slice** of that: a
read-only `sdlc resolve`. The editor UX is **parley#160** (`navigate ariadne
artifact references`), which shells to this.

## Spec

Add a **read-only** `sdlc resolve <ref>` that maps a symbolic ref → the current
file path(s).

- **Read-only ⟹ no git transaction lock.** Unlike mutating verbs (`issue new`,
  `close`), it never takes `.git/sdlc.lock`, so it's not subject to lock-contention
  slowness — cost is just process spawn + a glob. This is the property that makes it
  acceptable to route parley navigation through the binary (single source of truth)
  rather than re-implementing a resolver in Lua.
- **Derive locations from the models, don't hardcode.** Use the `discovery:` blocks
  in the vocab/datatype models (parley already sources the issue home from
  `issue.cue`, ariadne#116) so resolution tracks ariadne's structure — incl. the
  `issues/ → history/` mirror and `plans/` — and can't drift.
- **Single-source the ref grammar** here so parley + agents can't diverge:
  - `repo#id` → that repo's workshop issue (`<parent>/<repo>/workshop/{issues,history}/<id>-*.md`).
  - `#id` → the current repo.
  - `repo#id Mx` → the issue + milestone context (the `Mx` row / review sidecar).
  - By 6-digit id, resolve the whole **family**: issue + `<id>-*-plan.md` +
    `<id>-*-m*-review.md` (this is the original "interlink" ask — surfaced as a
    resolvable set, not stored links).
  - Disambiguate the GitHub inbox from the workshop tracker (sdlc already splits
    `--issue` vs `--github-issue`) — pick a form for GitHub refs (e.g. `repo gh#id`).
- Output: path(s), machine-readable (for parley/agents) — consider `--json` and a
  human `sdlc open <ref>` sugar that opens `$EDITOR`.

## Done when

- [ ] `sdlc resolve ariadne#11` prints the current path, correct after the file
      archived (`issues/ → history/`) and across sibling repos.
- [ ] It's read-only (no lock; provable by running under a held lock).
- [ ] Locations + grammar derive from the models (no hardcoded artifact paths).
- [ ] Resolving by id returns the family (issue + plan + reviews).
- [ ] Grammar is documented/single-sourced so parley#160 + agents consume one spec.

## Plan

- [ ] Define the ref grammar (single source — a small doc or vocab entry).
- [ ] `sdlc resolve` reading `discovery:` from the models; family resolution by id.
- [ ] Read-only guarantee (no lock) + a test proving it resolves under a held lock.
- [ ] `--json` output + optional `sdlc open` sugar.
- [ ] Tests: archived-file resolution, cross-repo, milestone/family refs.

## Log

### 2026-06-29

### 2026-07-03 — reframed (NOT wontfix): this is the ariadne `sdlc resolve` slice

Design discussion with the operator: the original "files carry stored cross-links"
premise is rejected (link-rot on archive). The replacement — **read-time resolution**
— keeps the symbolic ref canonical and resolves on demand. The editor UX lives in
parley (parley#160), but the **resolver is ariadne work** (`sdlc resolve`, base-layer
Go). Operator caught that I'd wrongly wontfix'd this: since sdlc owns the resolver,
this issue stays as the ariadne slice. Reframed off the stored-link premise onto
`sdlc resolve`; parley#160 `deps: [ariadne#144]`.
