---
id: 000328
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '0f8cf659bdde12eda78377566dadf7f99f5d4ea6' # card fields mirrored from issue-cards; edit via sdlc
---

# A language-neutral traceability map: atlas feature to code to tests, for targeted runs and test minimization

## Problem

parley.nvim keeps `atlas/traceability.yaml`: about 1,300 lines mapping each atlas feature to its code and its
tests, kept current (updated by parley.nvim#313 on 10-10). It powers `make test-spec SPEC=<feature>` and
`make test-changed` (run the tests for changed atlas entries; fail when a changed entry has no mapping). ariadne,
pair and the other repos have nothing like it: "requirements traceability" there is only a close-review checklist
item. Since the atlas is already an agent-maintained feature map, feature-to-code-to-tests is one step further.
Raised at ops' batch-2 recap (10-10).

## Spec

Design question first: what is the generic, language-neutral version?

- **What's hand-curated vs derived.** Feature-to-code records intent and stays curated, as part of the atlas
  update gate at each close. Code-to-tests can often be derived from the language's dependency graph (Go: `go list
  -deps`; Lua has none, hence parley's hand map) or from per-test coverage. Each derived half is checked against
  the curated map, and disagreements are flagged.
- **Uses:**
  - targeted test runs from a diff (changed code maps to affected tests), complementing ariadne#324's gate;
  - the close review's requirements check (which tests protect the touched features);
  - test minimization at the feature level (ariadne#327).
- **Shape:** one schema (keyed by atlas path, as parley's), language adapters for the derived half, and an `sdlc`
  verb (e.g. `sdlc trace tests --changed`). parley.nvim's file migrates to the schema rather than being replaced.
- Decide whether the map lives in `atlas/` or is generated beside it.

## Done when

- [ ] A design (durable plan) settles the schema, the curated/derived split, and the Go adapter.
- [ ] ariadne has a map for its atlas and a verb that lists the tests for a diff; `make test-changed`'s parley behavior is reproducible through it.
- [ ] The close gate's atlas update keeps the map current (mapping missing for a changed feature means refuse, as parley does).

## Plan

- [ ]

## Log

### 2026-10-10
