---
id: 000245
status: working
deps: [ariadne#242, ariadne#243]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours:
started: 2026-09-23T14:55:42-07:00
---

# Slots v2: branch from a workspace and refresh

## Problem

Agents need precise instructions for branching from :0/:1/:2 and refreshing a resting branch without implicit synchronization or hidden work loss.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Support the instruction “In :2, create an issue branch from :1’s current commit.” Resolve addresses within the repo, require a clean locally committed source and a destination ready to switch, pin the source commit, and record source address plus SHA in the issue workflow evidence. Branch the whole snapshot; do not infer issue-specific files or transfer claims. Neither resting branch nor its upstream moves. Subsequent source changes do not propagate.

Starting independent work on a slot branches from its current baseline without fetching or refreshing implicitly. Refresh is separately requested: fetch the configured remote main and fast-forward the resting branch when safe. Local planning commits/divergence require a separate explicit reconciliation choice, never an automatic reset/stash/commit. Refuse refresh of an active issue workspace unless the chosen explicit procedure first safely returns to rest. Define untracked/ignored-file readiness and source-change-during-capture behavior in the design.

Audit change-code and existing Git/procedure surfaces; prefer documented Git operations where adequate. A new adopt command or duplicate Git framework is not required. ARCH-DRY: one address resolver and existing branch/gate mechanisms.

### Agreed scope — 2026-09-23

This section takes precedence over earlier conflicting layout or policy text.

Resolve numbered main checkouts at `/workspace/worktree/<repo>-slotN/<repo>` through the shared contract updated in #243. Branch-from captures the selected repository's committed snapshot, not a multi-repository snapshot of its enclosing environment. Refresh changes only the explicitly targeted repository/resting branch; neither operation clones, switches, refreshes or resets sibling dependency repositories. Dependency revisions are controlled explicitly by operator/agent using ordinary Git. Cross-repository development is available from every slot and does not grant :0 special branch privileges.

## Done when

- Branching :2 from clean :1 starts at the captured SHA, preserves source and both resting refs/upstreams, and records provenance.
- Dirty source, unsafe destination, branch-name conflict, and source movement during capture have tested safe outcomes; no auto-stash/commit occurs.
- Issue start leaves an intentionally old baseline unchanged; explicit refresh fast-forwards a clean behind resting branch.
- Dirty/divergent/planning-commit cases preserve all work and expose explicit recovery; configured remotes other than origin are covered.
- Agent guidance uses branch-from wording and :0/:N shorthand, with ordinary primary workflow regression coverage.

- Branch-from and explicit main-worktree refresh preserve sibling dependency checkout branches, SHAs and dirty/unpublished work in nested-environment fixtures.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

- [ ] Specify readiness and provenance recording using current Git/SDLC surfaces.
- [ ] Add fixture tests and adapt issue-start/address resolution where needed.
- [ ] Implement/document explicit refresh and verify refusal/recovery cases.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

## Revisions

### 2026-09-23 — Nested slots retain per-repository branch and refresh scope

Reason: operator agreed nested environments, ordinary remote dependency clones and existing per-repository publication. Delta: added the authoritative scope clarification and acceptance criteria above; original task context remains as provenance. Added #243 as a prerequisite for the nested identity contract. No implementation or lifecycle-status change is claimed by this revision.
