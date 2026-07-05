---
id: 000145
status: working
deps: []
github_issue:
created: 2026-06-30
updated: 2026-07-05
estimate_hours: 1.08
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.2 impl=0.2
item: smaller-go-module   design=0.2 impl=0.2
item: atlas-docs          design=0.1 impl=0.1
design-buffer: 0.15
total: 1.08
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*
Decomposition: (1) `pkg/vocab` extension (`Section`/`Scaffold` types, `Sections()`,
`InitialStatus()`, regen embedded JSON) + `scaffold.go:Render` rewrite + byte-stable /
model-driven tests; (2) `structural.go` const single-sourcing + the three guard/drift
tests (`Problem`/`Log` coupling, gated ⊆ model, help ⊇ model); (3) `scaffold.sections`
cue data + regen the `construct/generated` face + atlas. `impl=` at v3.1's 40% of the
v2 table; +15% design buffer (thorough plan doc). familiarity 1.0 (warm codebase).

## Plan

Durable design: [`workshop/plans/000145-derive-issue-template-from-cue-plan.md`](../plans/000145-derive-issue-template-from-cue-plan.md).
Route: concrete-data + embed-JSON (`scaffold.sections` in `issue.cue`). Single-pass
(no `Mx` tags — one `sdlc close`). Beyond the literal Done-when, an **invariant
chain** enforces the shadow-sweep by test: `structural.go` gated ⊆ `scaffold.sections`
⊆ `helptext/issue.md` documented.

- [ ] Model `scaffold.sections` in `issue.cue`; vet + confirm it exports.
- [ ] `pkg/vocab`: `Sections()` + `InitialStatus()` (status = `categories.open[0]`, no new data); regen embedded `issue.json`.
- [ ] `issue.Render` derives sections + status from `vocab.Issue()`; byte-stable (golden test) + model-driven test.
- [ ] Guard the `Problem`/`Log` name-coupling with a test.
- [ ] Enforce `structural.go` gated ⊆ model (`gatedSections` consts + drift test).
- [ ] Enforce `helptext/issue.md` documents ⊇ model (drift test).
- [ ] Regenerate `construct/generated/vocabulary/issue.json` face (also clears the latent `codecomplete` drift).
- [ ] Build + `go test ./...` + manual `sdlc issue new --dry-run` parity + propagation e2e; atlas update; close.

## Log

### 2026-06-30

Filed from the parley.nvim#116 design conversation. Relationship: parley#116
delegates issue **creation** to `sdlc issue new` (revision (i) — so #116 is NOT
blocked on this ticket) and sources issue **status/location** from the emitted
`issue.json`. This ticket is the deeper unification (ii): once creation derives
from cue, the issue noun has a single model in cue that both sdlc and derivatives
consume. No hard cross-repo dep either way.

### 2026-07-05

Claimed → start-plan → durable plan authored (`workshop/plans/000145-…-plan.md`) →
change-code gate run. **Plan-quality judge: INFO (pass)** — "architecturally
exemplary; safe to start." Its concrete findings folded into the plan: (a) `sdlc
issue new` takes the title **positionally**, not via `--title` (Task 8 e2e fixed);
(b) added a `--from-github` byte-golden (`TestRender_ByteStable_FromGitHub`); (c)
Task 7 regen names the full generated-face drift — `codecomplete` status + its
lifecycle edges (#160) + `discovery.plans`/`archive` (#144), not just `codecomplete`.

**Estimate-quality judge: findings (adjudicated, no numeric change).** The block
follows `estimate-logic-v3.1` faithfully — `smaller-go-module design=0.2 impl=0.2`
matches the canonical example in `helptext/estimate.md` verbatim; v3.1 keeps design
hours and scales impl to 40%, so design==impl is expected, not inflation. The judge
read the `+15%` design-buffer as a double-count against "thorough plan," but +15% is
the **reduced** buffer v3.1 prescribes *because* a plan exists (vs +30% without) —
applied correctly. The judge also noted it lacked `baseline-v3.1.md` in-tree to
table-verify, and concluded the **total (1.08h) reads well-calibrated for the scope**.
Re-crossed with `--no-judge` (both judges already ran + adjudicated), not `--force`.

**Frontmatter fields (ARCH-PURPOSE, plan-quality finding #4):** only the one
duplicated *fact* — `status: open` — is derived (→ `InitialStatus()`).
`id/deps/github_issue/target/created/updated/estimate_hours` stay Go rendering logic:
they're per-instance *values* or live only in the non-exporting `#Issue` definition,
so there is nothing concrete to derive them from. Not an under-delivery — recorded
here so the close-boundary reviewer doesn't re-flag "frontmatter still hardcoded."
