---
id: 000243
status: working
deps: [ariadne#242]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours: 10.8
started: 2026-09-22T23:30:57-07:00
flow: {kind: full, provenance: operator}
---

# Slots v2: dependency and tool bindings

## Problem

The flat ../worktree/repo-slotN layout changes the meaning of relative peer dependencies. Linking to a moving primary checkout may also change a slot indirectly.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Resolve and implement the minimum dependency setup needed by couch-slots-v2. Compare explicit source bindings, stable dependency checkouts, and existing Weave facilities against the actual parley.nvim trial and an ariadne tooling workspace. Specify which dependency state is intentionally shared, what is pinned, and what explicit action changes it. Do not assume primary-checkout symlinks or same-number peer slots are approved.

Separate source dependencies from the workspace supplying shared installed binaries. Ordinary slot provision/resume/build must not silently select a different machine-wide tool supplier. Existing explicit installation remains possible. Produce the provisioning contract Couch will call, including failure/retry behavior. Change Weave only if existing operations cannot satisfy the chosen contract. ARCH-PURPOSE: keep dependency ownership with the dependency tooling; Couch orchestrates setup. This task requires a binding-policy decision before implementation.

### Agreed scope — 2026-09-23

This section takes precedence over earlier conflicting layout or policy text.

The 2026-09-23 decision replaces the flat-layout/shared-baseline alternatives above. A numbered slot has an environment directory `/workspace/worktree/pair-slot1/` containing the main Git worktree `pair/` and ordinary dependency clones such as `ariadne/`. The primary stays `/workspace/pair`; its existing sibling environment remains `/workspace/`. Relative declarations such as `../ariadne` resolve inside the numbered environment without a shared shelf, primary symlink, or same-number peer mapping.

Extend the shared resolver delivered by #242 for the nested main-worktree path and distinguish the enclosing environment from the checkout root. Keep #242 closed as the delivered flat-layout baseline; this task owns its layout follow-up and the consumer audit needed to distinguish environment-local source dependencies from canonical repository/workflow identity. A dependency clone is an ordinary independent Git repository, not another numbered Couch slot. Do not infer canonical fleet, calibration, or project ownership merely from its containing directory.

Provision dependencies using ordinary clones from their recorded remote sources, initially selecting `origin/main` rather than assuming the remote default branch is main. Existing checkouts retain their chosen commits, branches, dirty files and local commits on compile/setup/resume. Operator and agent can explicitly select another revision using ordinary Git; no local-primary source selection, linked dependency worktrees, live dependency sharing, or automatic dependency refresh is part of this version. Source URLs must be available for missing dependencies; missing origin/main and partially completed setup need explicit diagnostics and retry behavior. Whether extra persisted revision metadata is necessary remains an engineering decision, not an approved new lockfile system.

Weave owns dependency acquisition/composition; Couch orchestrates it. Existing symlinks may target the private dependency clone: edits there intentionally affect this environment, while another slot and the primary remain isolated. Copied/merged generated artifacts follow normal explicit recompilation. Source bindings remain separate from machine-wide installed tool supply: provision/resume/build must not silently retarget installed tools, while explicit installation remains available.

Cross-repository work can be driven from any slot's thread, using each repository's normal issue, review and publication workflow. Publish required Ariadne changes before dependent Pair changes. Do not add automatic recursive merge, cross-repository transaction, dependency editing limits, or dependency Couch entries.

## Done when

- A recorded decision specifies source binding, freshness/update behavior, generated-link behavior, and the shared-tool supplier policy.
- A fresh parley.nvim slot resolves its required dependencies and can run its relevant build/tests; repeat setup is idempotent and does not create unexplained tracked changes.
- Tests/probes show what happens when a primary dependency switches branch, a dependency is missing, and setup is interrupted; no silent retargeting occurs.
- The agreed contract is usable by Couch, with precise commands/API and recovery steps; whether Weave code changes are necessary is evidenced.

- Shared identity and production consumers resolve `/workspace/worktree/<repo>-slotN/<repo>` correctly; sibling dependency clones are not misclassified as numbered slots.
- Two fresh numbered environments get separate ordinary dependency clones initially at origin/main; choosing or editing a dependency revision in one leaves the other and the primary unchanged.
- Repeated compile/setup/resume preserves existing dependency work and machine-wide tool selection; missing source/main and interrupted acquisition have explicit tested recovery.
- The Couch setup contract documents nested paths, recorded remote requirements, initial origin/main selection, explicit revision changes, generated links, and normal dependency-first SDLC publication.

## Plan

Task outline only; settle implementation design through start-plan before change-code.

Engineering proposal: [nested slots and dependency bindings plan](../plans/000243-slots-v2-dependency-bindings-plan.md). Product policy is agreed; fresh-context engineering review and operator plan approval precede change-code.

- [x] Inspect current dependency and installation paths; compare minimal binding policies and obtain the policy decision.
- [x] Design and implement the nested identity follow-up and required dependency/setup changes with fixture tests.
- [ ] Prove fresh and repeated setup and document explicit dependency/tool updates.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Calibration is stale/provisional. Thorough-plan design discount ×0.2, v3.1 implementation multiplier ×0.4, familiarity 1.0 and design buffer +15%. Reuse GitReader/layergraph/staging/runner and existing filesystem-lock facilities; no novel library stack. No vendor-mode propagation multiplier.

The first estimate grouped whole subsystems into single primitives. Estimate-quality correctly identified that each contains several independently implemented/tested units. The following decomposition applies the same calibrated upper-bound primitive values separately rather than changing their unit rates. Rows map, in order, to:

- nested path grammar/schema (smaller-go-module: design 0.06, implementation 0.20).
- classification and address resolution (smaller-go-module: design 0.06, implementation 0.20).
- fake/real topology contract (smaller-go-module: design 0.06, implementation 0.20).
- host proof and inherited environment (greenfield-go-module: design 0.20, implementation 0.32).
- private-clone validation (greenfield-go-module: design 0.20, implementation 0.32).
- workspace stateful interleaving model (greenfield-go-module: design 0.20, implementation 0.32).
- peer selector and artifact reads (cross-cutting-refactor: design 0.20, implementation 0.20).
- project scan roots and forecasts (cross-cutting-refactor: design 0.20, implementation 0.20).
- close peer writes (cross-cutting-refactor: design 0.20, implementation 0.20).
- migration inbound references (cross-cutting-refactor: design 0.20, implementation 0.20).
- propagation scope (cross-cutting-refactor: design 0.20, implementation 0.20).
- feature worktree placement (cross-cutting-refactor: design 0.20, implementation 0.20).
- source acquisition policy (greenfield-go-module: design 0.20, implementation 0.32).
- acquisition transition core (greenfield-go-module: design 0.20, implementation 0.32).
- origin-main staged verification (greenfield-go-module: design 0.20, implementation 0.32).
- stateful remote/ref/stage fake (greenfield-go-module: design 0.20, implementation 0.32).
- setup lock lifetime (greenfield-go-module: design 0.20, implementation 0.32).
- inherited producer cancellation (greenfield-go-module: design 0.20, implementation 0.32).
- startup adapter and dry-run (smaller-go-module: design 0.06, implementation 0.20).
- startup tool supplier fixture (smaller-go-module: design 0.06, implementation 0.20).
- two-environment integration harness (smaller-go-module: design 0.06, implementation 0.20).
- Parley repeated-compile acceptance (smaller-go-module: design 0.06, implementation 0.20).
- runtime/product and owner-build acceptance (smaller-go-module: design 0.06, implementation 0.20).
- Pair isolated metadata issue/PR (cross-repo-refactor-small: design 0.06, implementation 0.12).
- Parley isolated metadata issue/PR (cross-repo-refactor-small: design 0.06, implementation 0.12).
- CLI/schema and workspace atlas (atlas-docs: design 0.04, implementation 0.08).
- Weave setup/recovery atlas (atlas-docs: design 0.04, implementation 0.08).
- main close review (milestone-review: design 0.04, implementation 0.20).
- peer metadata close reviews (milestone-review: design 0.04, implementation 0.20).

Smaller modules derive from 0.3h design/0.5h implementation; greenfield modules from 1h/0.8h; cross-cutting units from 1h/0.5h; small peer changes from 0.3h/0.3h; docs from 0.2h/0.2h; review from 0.2h/0.5h, before the discounts above. Product acceptance includes harness construction separately from running and diagnosing real product checks. Scope includes all consumer groups, failure models, process lifetime, peer publication and final evidence, not merely production typing.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: cross-cutting-refactor design=0.20 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: cross-cutting-refactor design=0.20 impl=0.20
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-repo-refactor-small design=0.06 impl=0.12
item: cross-repo-refactor-small design=0.06 impl=0.12
item: atlas-docs design=0.04 impl=0.08
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.04 impl=0.20
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 10.8
```

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

### 2026-09-22 — claimed; binding audit

Claimed after #242 merged and ran start-plan. Parley's tracked `construct/deps` contains `substrate ../ariadne`; from a flat numbered slot this resolves to the shared `/fleet/worktree/ariadne` location. Its product/runtime tests do not need this optional maintainer overlay. `scripts/check-fresh-clone.sh --runtime --ref HEAD` passed against an isolated archive, including missing/corrupt/malformed runtime-data probes.

Existing Weave accepts explicit substrate paths and optional source URLs, but no revision pin or per-checkout override. It reuses existing origin-matching checkouts without fetch/pull/reset; missing source-bearing edges clone current remote HEAD, while source-less missing edges refuse. Thus reuse is not pin enforcement. Generated static skill links remain live: an explicit dependency-checkout update changes their contents immediately, while copied/merged artifacts require recompilation. `weave compile` builds owner tools into owner bin directories and scopes their PATH to children; explicit install/dev-alias surfaces need separate supplier analysis. No live provisioning or shell configuration changes were made.

Policy question sent to operator: shared stable dedicated dependency baseline (minimum initial-trial machinery, explicitly shared update blast radius) versus independent per-slot dependency pins (new resolution support). Primary symlinks and same-number peer mapping are not assumed approved. ARCH-DRY: any new mapping must reach all layergraph consumers, not only source acquisition. ARCH-PURPOSE: Couch orchestrates a dependency-tool-owned contract.

Read-only audit verification: existing acquire/staging/plan/startup fixture suites passed. Clone staging already owns interrupted-producer recovery; concurrent setup is not currently supported. Repeated compilation can initially seed/adopt Makefiles and generated-ignore metadata, so the acceptance probe must explain tracked changes against a prepared baseline rather than assume a universally clean diff. Implementation design and estimate await the binding-policy decision.

### 2026-09-23 — engineering proposal checkpoint

Ran start-plan for the revised design. Audits confirmed both Pair and Parley declarations lack acquisition URLs, existing Weave clones remote default HEAD, and flat workspace identity misclassifies nested main/dependency checkouts. The durable proposal covers the nested resolver follow-up, independent clone identity and environment-local content selection, canonical calibration, private origin/main acquisition, setup interruption/exclusion, and real isolated Parley/tooling verification. Source declarations need small peer updates; no runtime implementation has started. Fresh-context spec/plan review is in progress. Unrelated process-manual and #230/#240 edits remain untouched; estimates wait for plan-quality acceptance.

### 2026-09-23 — engineering review approved

Fresh-context plan review approved Chunk 1 after fixing two findings: acquisition now accepts only the direct-sibling topology that discovery recognizes, and dependency feature worktrees use clone-specific paths with context inherited through Git-verified primary identity. Unsupported relative composition paths fail with guidance, without generic acquisition fallback. Peer metadata publication uses isolated checkouts to avoid incidental commits. Plan commits: `0b393eb7`, `2966aba`. Scoped diff validation passed; implementation and estimate remain pending operator engineering-plan approval and change-code.

### 2026-09-23 — implementation checkpoint

Operator approved execution; change-code plan-quality CLEAN after clarifying named test strategies, exact provisioning CLI and output-only v2 compatibility. Estimate-quality accepted the calibrated 10.8h unit decomposition. Implemented nested identity/schema and Git proof (`92f6fdb`), environment-aware SDLC consumers (`bf30f4e`), and private remote-main acquisition with inherited setup leases (`43e53fe`). Full `go test ./pkg/workspace/... ./cmd/weave/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'` passed (SDLC package 367.829s), as did vet and scoped diff checks. The sole skip is existing #210. Additional focused tests verify lexical symlink refusal, whitespace-preserving Git reads, all accepted sibling names, local project writes and independent clone authority. An earlier parallel run passed assertions but failed the repo-status guard when another worker added a test file; the final broad run passed cleanly.

Peer metadata published via isolated clean clones, leaving operator checkouts unchanged: pair#310 PR154 (`d7d062dd`, merged) and parley.nvim#274 PR199 (`007817b4`, merged). Parley merge also auto-archived its pre-existing completed #220 bookkeeping under normal SDLC. Real two-environment compilation and runtime acceptance passed; full Parley suite and final exact script remain in progress. Source HEADs/dirty work and canonical fleet remained unchanged by repeat compilation. Initial owner binaries may differ on repeat because Go embeds VCS dirty-state metadata after generated .gitignore adoption; composition comparisons separate and record this build provenance.

## Revisions

### 2026-09-23 — Nested workspace and dependency policy agreed

Reason: operator agreed nested environments, ordinary remote dependency clones and existing per-repository publication. Delta: added the authoritative scope clarification and acceptance criteria above; original task context remains as provenance. No implementation or lifecycle-status change is claimed by this revision.

### 2026-09-23 — first engineering proposal

Reason: operator requested continuation after the policy/task update. Delta: linked the durable implementation plan and recorded live integration gaps. The plan supplies a concrete reviewable design; its additional engineering choices are proposed, not silently treated as prior product decisions.
