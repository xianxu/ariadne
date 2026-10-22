---
id: 000551
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b693f91ea7e19c0a951e20b4037b2b69fb6ffbcf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T05:22:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# review box gives the reviewer no dependency checkouts; consumer reviews can't verify substrate claims (echo#9 BR-1 unverifiable)

## Problem

## Spec

From ops-echo:0 (10-22 ~05:30): echo#9's boundary review returned REWORK with one Critical, BR-1 "unverifiable",
because the reviewer couldn't check echo against ariadne's fixtures and templates: it looked for
`/Volumes/TestScratch/review-checkouts/ariadne`, which doesn't exist (`review-checkouts` is the review box's root for
the disposable clone of the repo **under review**, `cmd/sdlc/internal/reviewbox/checkout.go`). No product defect was
found. Every echo close (echo#9, #10, #3, #5) touches the ariadne-seeded workflow, so each would hit the same wall.

The review box gives the reviewer only the repo under review, never its declared dependencies, so a consumer repo's
review can't verify claims that depend on the substrate. Same class as ariadne#545 (suites couldn't find
dependencies).

Fix: the review box also provides read-only, disposable checkouts of the repo's `construct/deps` dependencies (the
substrate at the exact pinned revision or main, as the deps row says), placed beside the review clone, with their
paths told to the reviewer in the review prompt (mirroring #545's SDLC_DEP_* names). Same lifecycle as the review
clone (ARCH-FUNERAL), never writable, bounded size and time.

## Done when

- A review of a consumer repo whose deps name ariadne sees a read-only ariadne checkout at the documented path, and
  the prompt names it (test).
- Removed with the review clone; repos with no deps are unchanged (test).

## Plan

Implementation plan: [review dependency checkouts](../plans/000551-review-dependency-checkouts-plan.md).

- [ ] Share bounded declaration resolution and dependency descriptors.
- [ ] Create pinned disposable dependency clones with retry/cleanup ownership.
- [ ] Expose paths, SHA, freshness and isolation mode in boundary prompts; verify and land.

## Log

### 2026-10-22


## Revisions

### 2026-10-22 — TL-approved dependency and isolation semantics

The current construct/deps grammar has no revision-pin field. Reuse it without extension, resolve each direct local dependency, and pin refs/remotes/origin/main as last fetched once per review. Do not fetch in setup or retry. The prompt names the exact SHA and evidenced fetch age (unknown when not provable).

The original never-writable condition is revised by TL instruction: with an available OS fence, dependency clones are read-only; without it, give the reviewer disposable dependency clones, permit writes only under the existing checkout isolation guarantee, label that mode, and discard their mutations on retry/cleanup. Do not refuse an otherwise supported review solely because the fence is unavailable. No live dependency checkout is substituted for its disposable copy.

### 2026-10-22 — design checkpoint

Claimed and started planning after #548 landed. #531 is parked by TL. Read #545 environment resolution, layergraph grammar/resolver, reviewbox checkout/fence/confinement and the shared boundary dispatch. Read-only delegated lifecycle audit identified that an unowned sibling directory would be swept during a live review; the plan therefore reuses individually locked Create clones retained by the primary owner. No source implementation started. Full plan includes direct substrate/data bindings, exact last-fetched pins, freshness uncertainty, bounded materialization, retry/death cleanup and real-Git/fence proof. ARCH-DRY reuses existing declaration and checkout owners; ARCH-FUNERAL keeps all artifacts within their current cleanup lifecycle.
