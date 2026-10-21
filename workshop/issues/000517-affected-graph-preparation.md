---
id: 000517
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1fd7c7d910470bc51c0304a0e00a06ff83331f7c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T23:25:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "54c1cfe0", done: "d9749df0"}
---

# Prepare generated inputs before affected graph discovery

## Problem

Pair's fresh affected graph checkout has no generated runtime bundle. Native go list -test ./... fails at cmd/internal/runtimebundle/embed.go:8:41 with pattern assets/runtime/files: no matching files found. Running the existing runtimebundle generator in that same checkout makes the graph load (408 entries measured on 2026-10-20). Every affected run currently falls back to a full suite, including docs-only changes. Pair's command runs after graph discovery, too late to establish the prerequisite.

## Spec

Add an optional bounded graph-preparation argv declared by the repository, executed once in the same detached HEAD checkout before LoadGoGraph. No shell interpolation, borrowed generated outputs, dummy embed files, or partial graph acceptance. Any preparation failure, timeout or cancellation preserves conservative full-suite fallback and names the preparation cause; cancellation must not record a passing affected check. Reuse existing subprocess lifetime/timeout/environment boundaries where applicable; define the bounded contract explicitly. No declaration retains today's behavior. Pair#738 will declare make runtimebundle-generate only after this provider lands; that declaration is operator-verified local-gates configuration.

## Done when

- An unprepared disposable repo with a generated embed prerequisite computes the real native graph after its declared argv prepares the same HEAD checkout.
- A stateful fixture proves prepare precedes graph, runs exactly once, and never uses the caller checkout's generated output.
- Failure, timeout and cancellation controls do not produce a passing affected result; genuine graph failure still falls back with its cause.
- Config validation covers missing optional declaration, malformed argv and bounded limits, with compatibility documented.

## Plan

Quick-flow proposal: optional `graph_prepare` object in main's affected.json,
with argv `command` and `timeout_minutes` (default 2, range 1..30). Absent/null
object means no preparation; empty command or invalid limits refuse config.
Run once in the existing detached graph checkout, before LoadGoGraph. Inherit
caller cancellation; kill/reap the process group and bound pipe draining to one
second before graph or cleanup. Stream command output to the existing stderr;
return a named preparation error into the existing full-suite fallback. No new
status, cache, log family, queue policy or caller-tree writes.

ARCH-PURE: validation stays pure, execution stays at the graph IO boundary.
ARCH-DRY: existing processgroup, detached-checkout cleanup and fallback own their
current responsibilities. ARCH-CONSTRAINTS: one preparation per discovery;
two-minute default follows the existing make probe, 30-minute maximum follows
review preparation. No per-file subprocess growth. ARCH-SECURE: main owns the
argv, the requested head supplies its inputs; no implicit shell or new secrets.
ARCH-EVOLVE: optional additive field; existing older strict readers refuse it,
never accept a partial graph. ARCH-FUNERAL: generated files die with the existing
checkout scope. ARCH-ORDER: no state survives a call; completion or cancellation
precedes graph access and cleanup. ARCH-MOCK: real Git/module/argv fixtures and
stateful fake GitHub use production seams. ARCH-PURPOSE: pair#738 adopts the
provider later; hardcoding a consumer generator or reusing review preparation
would conflate separate responsibilities.

Actions: run affected; rerun after repairing preparation.
| Internal outcome | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| absent/success, graph valid | ordinary affected selection | run affected | none |
| command/timeout/graph failure | full suite with named cause | rerun after repair | disposable generated output |
| caller cancellation | interrupted, no passing affected record | run affected | partial disposable output |
Existing stale checkout cleanup covers process death; no new durable state.

- [x] Red fresh-module missing-embed regression, same-head/once/cleanup assertions.
- [x] Bounded preparation and config validation; failure, timeout, cancellation,
  omitted declaration and genuinely broken graph controls.
- [x] README/atlas contract and verification evidence prepared.

## Log

- Filed from pair:7 for pair#738 under ops:0 approval. Source: cmd/sdlc/localcheck_affected.go productionAffectedDeps creates its private checkout and immediately calls affected.LoadGoGraph; affected/config.go currently has no preparation field. Pair cold-checkout reproduction log: .git/manual-verification/pair738-cold-graph.log. Unclaimed provider work, not permission to alter another slot or bypass local-gates verification.

- 2026-10-20: read actual issue body after discovering issue show omits section
  bodies. Provider scope already approved. Concrete schema/bounds above are
  checkpointed for TL approval before implementation; expected under 100 added
  production lines. No pair configuration will be edited here.

## Revisions

- 2026-10-20 23:29 PDT — TL approved schema/bounds, requiring immutable main
  declaration authority. Resolve main to a commit before reading affected.json;
  regression edits candidate graph_prepare and proves only pinned main's argv
  runs. Pair#738 adoption remains separate and requires the operator label when
  local-gates-guard covers its affected.json. No readiness policy change.

- 2026-10-20: red first on unknown graph_prepare, then after schema support the
  real fresh-checkout fixture fell back with native go list's missing
  assets/data.txt embed error (suite-only success). Added execution afterward.
  Full command fixture now proves once-only generation, exact head inputs,
  cleanup, caller poison isolation and candidate declaration rejection, including
  origin/main advancing to the candidate after its immutable pin is captured.
  Focused affected/CLI/config controls pass (17.329s for cmd/sdlc). Preparation
  nonzero, launch failure, deadline and still-invalid graph all preserve suite;
  cancellation and exited-parent inherited-pipe controls stop descendants and
  never publish affected success. No new lint: the missing declared graph-input
  prerequisite is a semantic contract, defended through real production fixtures.

- 2026-10-20: broader affected/CLI/config/CI-readiness controls passed (33.365s);
  five race-enabled repetitions of the new preparation fixtures passed (25.333s).
  All affected package tests and the full Go lint catalog passed; diff check clean.
  Only the deadline control shrinks the timeout unit; healthy/failure controls
  retain ample startup budget. About 81 added production lines, within quick flow.
  Next: pushed-head suite and close review; no live :0 stage held.

- 2026-10-20 23:39 PDT — planning bookkeeping: last checkbox now describes the
  completed docs/evidence deliverable. The recorded suite, close review and
  merge remain required workflow gates, tracked by their own results, rather
  than marking future gate outcomes as already complete.

- 2026-10-20: PR #361 opened; pushed-head suite requested. Close review may
  overlap it under tl-worker; a final recorded pass remains required for landing.
  Asked TL for smoke decision; real native graph/argv/Git fixtures cover this
  provider path, no :0 stage held. Pair adoption remains pair#738 and operator
  verification of its local-gates declaration.
