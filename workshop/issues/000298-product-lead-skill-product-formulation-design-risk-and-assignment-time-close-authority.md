---
id: 000298
status: open
deps: []
github_issue:
created: 2026-10-06
updated: 2026-10-06
estimate_hours:
card_mirror: '05a684a826c6ebb223ad45cee17a8605e6c96b1e' # card fields mirrored from issue-cards; edit via sdlc
---

# Product-lead skill: product formulation, design risk, and assignment-time close authority

## Problem

SDLC starts at an issue and ends at merge. The judgement before that is unowned:
what product to build, how the design should go, what could go wrong in bringing
it to life. ARCH-* principles give agents architectural taste at plan time. Nothing
gives them product taste, or makes the risk of a piece of work explicit when it is
handed to a worker.

The first concrete gap: whether a task may close autonomously or needs the
operator's smoke test (ariadne#297's `verify`) is a risk judgement. Today nobody
makes it deliberately.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

- **A product-lead skill**, distinct from a tech lead's mechanics, which covers:
  - product formulation: what problem, for whom, and what "works" means to the user;
  - design review from the product side;
  - the risks in delivering the product.
- **Its first concrete output** is the per-item risk reading that sets ariadne#297's
  `verify` when work is assigned. Examples of risk:
  - user-visible surface;
  - live-state migration;
  - behaviour tests cannot reach;
  - external services.
  The reasoning is recorded with the assignment.
- **Related.**
  - ariadne#236 (Product lens: PROD-* principles and a per-product user model) is
    the likely taste registry this skill consumes, as plan-time judges consume
    ARCH-*. Merge or sequence with it.
  - ariadne#15 (`product`/`roadmap` data types) is the charter the skill reads and
    writes.
- **Home: ariadne** (operator, 2026-10-06), as a module on top of the SDLC core,
  under two rules from `docs/vision/2026-10-06-01-pensive-before-and-after-the-sdlc.md`:
  - it costs nothing until used: a project that never invokes it still works, and
    `verify` defaults to operator without it;
  - it depends on the core, and the core never depends on it.
  The bar is the one that split metis out: most projects that evolve in a
  structured way need product formulation; most never need an ML workbench.

## Done when

- A product-lead skill exists in ariadne. A project that never invokes it is
  unaffected, and the SDLC core has no dependency on it.
- Given a work item, it produces a risk reading and a `verify` setting with
  reasons. It is exercised on real items from a live project, including one where
  `operator` is warranted and one where `auto` is safe.
- Its relation to ariadne#236 is resolved (merged, sequenced, or kept separate with
  reasons).

## Plan

- [ ]

## Log

### 2026-10-06

Filed from pair#362's brainstorm at the operator's request ("separate. that's pretty
a new issue"). The operator framed the role as more product-lead than tech-lead: the
design of a product and the risks in bringing it to life, not only mechanics. Details
left local for the operator to refine.
