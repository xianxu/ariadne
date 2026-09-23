---
id: 000245
status: open
deps: [ariadne#242]
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours:
---

# Slots v2: branch from a workspace and refresh

## Problem

Agents need precise instructions for branching from :0/:1/:2 and refreshing a resting branch without implicit synchronization or hidden work loss.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Support the instruction “In :2, create an issue branch from :1’s current commit.” Resolve addresses within the repo, require a clean locally committed source and a destination ready to switch, pin the source commit, and record source address plus SHA in the issue workflow evidence. Branch the whole snapshot; do not infer issue-specific files or transfer claims. Neither resting branch nor its upstream moves. Subsequent source changes do not propagate.

Starting independent work on a slot branches from its current baseline without fetching or refreshing implicitly. Refresh is separately requested: fetch the configured remote main and fast-forward the resting branch when safe. Local planning commits/divergence require a separate explicit reconciliation choice, never an automatic reset/stash/commit. Refuse refresh of an active issue workspace unless the chosen explicit procedure first safely returns to rest. Define untracked/ignored-file readiness and source-change-during-capture behavior in the design.

Audit change-code and existing Git/procedure surfaces; prefer documented Git operations where adequate. A new adopt command or duplicate Git framework is not required. ARCH-DRY: one address resolver and existing branch/gate mechanisms.

## Done when

- Branching :2 from clean :1 starts at the captured SHA, preserves source and both resting refs/upstreams, and records provenance.
- Dirty source, unsafe destination, branch-name conflict, and source movement during capture have tested safe outcomes; no auto-stash/commit occurs.
- Issue start leaves an intentionally old baseline unchanged; explicit refresh fast-forwards a clean behind resting branch.
- Dirty/divergent/planning-commit cases preserve all work and expose explicit recovery; configured remotes other than origin are covered.
- Agent guidance uses branch-from wording and :0/:N shorthand, with ordinary primary workflow regression coverage.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

- [ ] Specify readiness and provenance recording using current Git/SDLC surfaces.
- [ ] Add fixture tests and adapt issue-start/address resolution where needed.
- [ ] Implement/document explicit refresh and verify refusal/recovery cases.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.
