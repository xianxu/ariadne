---
id: 000145
status: working
deps: []
github_issue:
created: 2026-06-30
updated: 2026-07-05
estimate_hours:
started: 2026-07-05T15:01:33-07:00
---

# sdlc issue new: derive on-disk template from the issue.cue model

## Problem

`sdlc issue new`'s on-disk template is a **hardcoded Go renderer**
(`cmd/sdlc/internal/issue/scaffold.go:Render`), independent of the
`construct/vocabulary/issue.cue` model. Its own doc comment calls it "the single
source of truth for the on-disk template" — including literal `status: open` and
the `## Spec` / `## Done when` / `## Plan` / `## Log` section list, none of it
derived from cue.

So the issue noun has **two representations kept consistent by hand**:

| Concern | Source |
|---|---|
| Creation (template written to disk) | hardcoded Go — `scaffold.go:Render` |
| Validation (`sdlc issue validate` vs `#Issue`) | cue `#Issue` definition |
| Lifecycle / status (`CanTransition`, `AllStatuses`, `IsOpen`) | cue → `issue.json` → `pkg/vocab` |

The validate gate (`CheckStructural` at `sdlc change-code`) catches drift, but
creation and the cue model don't *share* a source. This blocks true
single-sourcing of the issue shape, and means downstream consumers that want the
issue structure from cue can't get the *creation* template that way (they must
either delegate to `sdlc issue new` or re-derive). Surfaced while designing
parley.nvim#116 (parley's ariadne-support discovery subsystem): parley sources
issue **status** from `construct/generated/vocabulary/issue.json` already, but
for **creation** it can only delegate to `sdlc issue new` because the template
isn't in the cue model.

## Spec

Make `issue.Render` derive its structure from the issue.cue model rather than
hardcoding it, so a change in cue propagates to created issues.

Key design constraint (verified): cue `#Issue` is a **definition**, and
`cue export` drops `#`-definitions (it emits only concrete data — that's why
`issue.json` carries `categories`/`when`/`lifecycle` but not the `#Issue` field
shape). So sourcing the template from cue requires one of:

- model the creation **template** (section list + frontmatter scaffold) as
  *concrete* cue data (e.g. a `scaffold:` / `sections:` block) that exports to
  `issue.json` and that `scaffold.go` consumes via `pkg/vocab` (consistent with
  the existing embed-JSON pattern — no new cuelang Go dep); **or**
- evaluate the `#Issue` definition via the cue Go API / `cue def` (adds a real
  cue dependency to the runtime path — heavier).

Lean toward the concrete-data + embed-JSON route (matches how `pkg/vocab`
already works). Decide at design time.

Optionally (and relevant to parley.nvim#116): also model `discovery` /
**location** (home folder, filename convention) as concrete cue data so it flows
to `issue.json`, unifying the issue noun's *structure* AND *location* in cue for
derivative consumers.

## Done when

- `issue.Render` sources its sections + frontmatter from the cue-derived model —
  no hardcoded `## Spec` / `status: open` literals that duplicate cue.
- A single change in `issue.cue` (e.g. add/rename a section, change default
  status) propagates to `sdlc issue new` output with no Go edit.
- Created issues still pass `#Issue` validation; `sdlc issue new` output is byte-
  stable (or intentionally evolved, with the change tested).
- (Optional) location/discovery exposed in `issue.json` for derivative consumers.

## Plan

- [ ]

## Log

### 2026-06-30

Filed from the parley.nvim#116 design conversation. Relationship: parley#116
delegates issue **creation** to `sdlc issue new` (revision (i) — so #116 is NOT
blocked on this ticket) and sources issue **status/location** from the emitted
`issue.json`. This ticket is the deeper unification (ii): once creation derives
from cue, the issue noun has a single model in cue that both sdlc and derivatives
consume. No hard cross-repo dep either way.
