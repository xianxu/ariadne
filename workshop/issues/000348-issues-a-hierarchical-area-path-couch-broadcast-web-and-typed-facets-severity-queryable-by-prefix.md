---
id: 000348
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '436afbbefce8cf2ca12dfefb245165ade972239d' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-13

## Revisions

- 2026-10-13 — the operator: areas are not free paths. **Each repo defines a small set of allowed values**, and a
  validation gate checks them. Generalize it as a new vocabulary facility: a **repo-defined enum** — a frontmatter field
  whose type is declared in the shared model (`construct/vocabulary`) but whose allowed values come from the repo's
  own layer (e.g. `construct/local/vocabulary/<field>.cue` or a small declared file), so ariadne ships the contract
  and each repo fills it in. Values may still be `/`-paths (`couch/broadcast/web`) for prefix queries. Validation runs
  where frontmatter is written (`issue new`, setters, `issue publish`) and in the existing validation gates, refusing an
  undeclared value and naming the declaration to edit. `area` is the first user; `severity` stays a shared enum.
