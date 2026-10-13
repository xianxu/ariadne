---
id: 000348
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours: 2.11
card_mirror: '19d09c0066837e047265ef8bfe6782c7c1ceddf3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T16:23:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# issues: a hierarchical area path (couch/broadcast/web) and typed facets (severity), queryable by prefix

## Problem

Issues carry no classification beyond their repo and title, so the TL can't record or query which product feature an issue belongs to or how severe it is; the 10-12/13 couch-broadcast-web hardening triaged ~30 issues by memory and prose.

## Spec

The operator (2026-10-13) leans to a semi-hierarchical scheme over JIRA-style flat tags; the analysis that settled it: *where an issue belongs* is genuinely hierarchical (product > feature > sub-area, wanting roll-ups), while *facets* (severity) are orthogonal and single-valued, so mixing them into one path (couch:broadcast:high) breaks cross-product queries and churns.
1. **`area:`** — one or more hierarchical paths with `/` (`:` already means a slot), e.g. `couch/broadcast/web`; `sdlc issue list --area couch/broadcast` matches by prefix.
2. **Typed facets** — start with `severity` (a small enum from the vocabulary model); add others only on need.
3. **Before-MVP vs follow-up is NOT a tag:** it's membership in the project's `mvp_scope` (#329), a fact about the project.
4. Modelled in `construct/vocabulary/issue.cue`; set via `sdlc issue set-*` setters; shown by `sdlc issue show/list` with filters; unknown areas allowed (free paths), severity validated.

## Done when

- An issue carries areas and a severity through the vocabulary model and setters; list filters by area prefix and severity; tests cover the model, setters and prefix matching.

## Plan

Durable plan: `workshop/plans/000348-issues-a-hierarchical-area-path-couch-broadcast-web-and-typed-facets-severity-queryable-by-prefix-plan.md` (single pass, one close).

- [ ] Model: `enums`/`repoEnums` sections, area + severity card rows, `#Issue` shape (vocab-embed)
- [ ] `internal/repoenum`: strict declaration, ancestor rule, segment prefix
- [ ] Card kinds `enum` / `repo-enum` (shape only in ParseCard)
- [ ] Loader (`.sdlc/vocabulary.json` at main) + writers: `issue new --area/--severity`, `set-area`, `set-severity`
- [ ] Gates: #124 instance gate + `issue validate` check membership
- [ ] `issue list --area <prefix> --severity`; state JSON fields
- [ ] Help + atlas
- [ ] Verify, close, merge

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: typed-data-prototype   design=0.2 impl=0.16
item: greenfield-go-module   design=0.2 impl=0.24
item: cross-cutting-refactor design=0.1 impl=0.2
item: smaller-go-module      design=0.05 impl=0.2
item: smaller-go-module      design=0.05 impl=0.16
item: smaller-go-module      design=0.05 impl=0.12
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0 impl=0.16
design-buffer: 0.15
total: 2.11
```

- Items, in order: model sections + card rows + `#Issue` shape; the pure `repoenum` core; card kinds plus the scalar-assumption class (R1); loader + writers; gates; list query; help/atlas; the close review. Design is at ×0.2 (the plan and R1 pre-resolve the decisions); impl is at 40% of the v2 table midpoints.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-13

- Mapped the plumbing. Card fields come from `issue.cue` `card.fields` (kind + setter) and are validated by `validateCardScalar` (scalar-only; `claimant` is the one structured kind). Details carry `target`/`deps`. sdlc reads the model from the embedded `pkg/vocab/issue.json`, so a repo's own vocabulary layer is invisible to it, and the `vocabulary` binary's layer merge is whole-file shadowing. Precedent for shared-shape/repo-values: `fleet-policy.cue` + `.sdlc/fleet.json` + `internal/fleet/load.go`. `issue list` filters only `--status`; legacy repos have no setters.
- Draft design sent to the hub: a `repoEnums` model section (area: path syntax, multi) + `.sdlc/vocabulary.json` declaration (declared or ancestor-of-declared allowed); area/severity as CARD fields (new kinds path-list, enum; ParseCard checks shape only); membership validated at `issue new` / `set-area` / `set-severity` against main's declaration and in `issue validate` + the #124 gate; `issue list --area` (segment prefix) / `--severity`. Open questions: declaration location (guarded `.sdlc/` vs `workshop/`), severity values, card vs details.
- Hub approved (2026-10-13): `.sdlc/vocabulary.json` under the local-gates guard (operator governs the taxonomy), severity critical/high/medium/low, card fields, no starter set. An absent file is the empty declaration, so this change adds no guarded file.

## Revisions

- 2026-10-13 — the operator: areas are not free paths. **Each repo defines a small set of allowed values**, and a
  validation gate checks them. Generalize it as a new vocabulary facility: a **repo-defined enum** — a frontmatter field
  whose type is declared in the shared model (`construct/vocabulary`) but whose allowed values come from the repo's
  own layer (e.g. `construct/local/vocabulary/<field>.cue` or a small declared file), so ariadne ships the contract
  and each repo fills it in. Values may still be `/`-paths (`couch/broadcast/web`) for prefix queries. Validation runs
  where frontmatter is written (`issue new`, setters, `issue publish`) and in the existing validation gates, refusing an
  undeclared value and naming the declaration to edit. `area` is the first user; `severity` stays a shared enum.
