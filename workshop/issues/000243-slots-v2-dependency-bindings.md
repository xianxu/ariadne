---
id: 000243
status: working
deps: [ariadne#242]
github_issue:
created: 2026-09-22
updated: 2026-09-22
estimate_hours:
started: 2026-09-22T23:30:57-07:00
---

# Slots v2: dependency and tool bindings

## Problem

The flat ../worktree/repo-slotN layout changes the meaning of relative peer dependencies. Linking to a moving primary checkout may also change a slot indirectly.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Resolve and implement the minimum dependency setup needed by couch-slots-v2. Compare explicit source bindings, stable dependency checkouts, and existing Weave facilities against the actual parley.nvim trial and an ariadne tooling workspace. Specify which dependency state is intentionally shared, what is pinned, and what explicit action changes it. Do not assume primary-checkout symlinks or same-number peer slots are approved.

Separate source dependencies from the workspace supplying shared installed binaries. Ordinary slot provision/resume/build must not silently select a different machine-wide tool supplier. Existing explicit installation remains possible. Produce the provisioning contract Couch will call, including failure/retry behavior. Change Weave only if existing operations cannot satisfy the chosen contract. ARCH-PURPOSE: keep dependency ownership with the dependency tooling; Couch orchestrates setup. This task requires a binding-policy decision before implementation.

## Done when

- A recorded decision specifies source binding, freshness/update behavior, generated-link behavior, and the shared-tool supplier policy.
- A fresh parley.nvim slot resolves its required dependencies and can run its relevant build/tests; repeat setup is idempotent and does not create unexplained tracked changes.
- Tests/probes show what happens when a primary dependency switches branch, a dependency is missing, and setup is interrupted; no silent retargeting occurs.
- The agreed contract is usable by Couch, with precise commands/API and recovery steps; whether Weave code changes are necessary is evidenced.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

- [ ] Inspect current dependency and installation paths; compare minimal binding policies and obtain the policy decision.
- [ ] Implement only the required resolution/setup changes with fixture tests.
- [ ] Prove fresh and repeated setup and document explicit dependency/tool updates.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

### 2026-09-22 — claimed; binding audit

Claimed after #242 merged and ran start-plan. Parley's tracked `construct/deps` contains `substrate ../ariadne`; from a flat numbered slot this resolves to the shared `/fleet/worktree/ariadne` location. Its product/runtime tests do not need this optional maintainer overlay. `scripts/check-fresh-clone.sh --runtime --ref HEAD` passed against an isolated archive, including missing/corrupt/malformed runtime-data probes.

Existing Weave accepts explicit substrate paths and optional source URLs, but no revision pin or per-checkout override. It reuses existing origin-matching checkouts without fetch/pull/reset; missing source-bearing edges clone current remote HEAD, while source-less missing edges refuse. Thus reuse is not pin enforcement. Generated static skill links remain live: an explicit dependency-checkout update changes their contents immediately, while copied/merged artifacts require recompilation. `weave compile` builds owner tools into owner bin directories and scopes their PATH to children; explicit install/dev-alias surfaces need separate supplier analysis. No live provisioning or shell configuration changes were made.

Policy question sent to operator: shared stable dedicated dependency baseline (minimum initial-trial machinery, explicitly shared update blast radius) versus independent per-slot dependency pins (new resolution support). Primary symlinks and same-number peer mapping are not assumed approved. ARCH-DRY: any new mapping must reach all layergraph consumers, not only source acquisition. ARCH-PURPOSE: Couch orchestrates a dependency-tool-owned contract.

Read-only audit verification: existing acquire/staging/plan/startup fixture suites passed. Clone staging already owns interrupted-producer recovery; concurrent setup is not currently supported. Repeated compilation can initially seed/adopt Makefiles and generated-ignore metadata, so the acceptance probe must explain tracked changes against a prepared baseline rather than assume a universally clean diff. Implementation design and estimate await the binding-policy decision.
