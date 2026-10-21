---
id: 000517
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8699886f4a2a520bb9bd17a8247becef136a07c9' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

- Filed from pair:7 for pair#738 under ops:0 approval. Source: cmd/sdlc/localcheck_affected.go productionAffectedDeps creates its private checkout and immediately calls affected.LoadGoGraph; affected/config.go currently has no preparation field. Pair cold-checkout reproduction log: .git/manual-verification/pair738-cold-graph.log. Unclaimed provider work, not permission to alter another slot or bypass local-gates verification.
