---
type: pensive
date: 2026-10-06
topic: before and after the SDLC
mode: thoughts
description: SDLC covers issue → merge. Driving a project from one TL slot exposes the unowned stages on either side — product formulation and risk before, verification, release and SRE after — and decides they belong in ariadne, as modules on top of the SDLC core.
references: [workshop/issues/000297-sdlc-verification-authority-and-an-operator-smoke-test-state.md, workshop/issues/000298-product-lead-skill-product-formulation-design-risk-and-assignment-time-close-authority.md, workshop/issues/000236-product-lens-user-model.md, workshop/issues/000015-product-and-roadmap-data-types.md, ../pair/workshop/issues/000396-tl-driven-project-execution-from-a-driver-repo.md, ../pair/workshop/projects/cross-slot-work-scheduling.md]
---

# Pensive: before and after the SDLC

This came out of pair#362's brainstorm. The operator wants to work in larger
pieces: drive a whole project from one TL slot instead of shepherding each issue.
Asking what that TL needs showed that we are no longer designing the SDLC. We are
designing the stages on either side of it.

## What SDLC covers today

`claim → start-plan → change-code → milestone-close → close → merge`. Its unit is
the issue. It assumes:
- someone already decided the issue is worth doing and roughly how risky it is;
- "done" is decided by a review gate.

Both assumptions are the operator's head today.

## The stage before: product formulation

- **What to build and for whom**, and what "works" means to the user.
- **Design from the product side**, not only architecture. ARCH-* gives agents
  architectural taste at plan time. Nothing gives them product taste.
  ariadne#236's PROD-* lens is the seed.
- **Risk in bringing the product to life.** This judgement has a concrete
  downstream effect already: it decides whether a piece of work may close on its
  own or needs the operator's smoke test (ariadne#298 sets ariadne#297's `verify`).

The role is closer to a product lead than a tech lead: judgement, not mechanics.

## The stage after: verification, release, SRE

- **Verification by the operator is a distinct fact from "the review passed".**
  pair#387 showed this: the review passed, and the live smoke test still found a
  gap. ariadne#297 makes it a field and a gate.
- **Where a smoke test runs depends on the product.** Most products can be tested
  in place. `:0` as a test bench is an artifact of building developer tools that
  build the apps. A bench queue waits until it hurts.
- **Release and SRE are absent.** Release means things like a tap bump or a
  version tag. SRE means things like whether the binary on `PATH` is the one we
  think it is: pair#387's smoke test ran an old Homebrew weave. Today these are
  runbooks in memory notes.

## Where they live: in ariadne (decided 2026-10-06)

The bar is the one that split metis out: does a project need this to evolve in a
structured way? Metis fails it, because most projects never run an ML workbench.
These pass:
- **Product formulation passes.** Every project that evolves past one chat has a
  "what, for whom, and what could go wrong", even a hobby one.
- **Release, broadly construed, passes.** Every real project ships something: a
  Homebrew formula, a tag, a deploy, a CI/CD pipeline.
- **Narrow SRE (on-call, SLOs, incident response for cloud services) may not
  pass.** It stays inside the release/SRE area for now and is the one possible
  later split, if it grows heavy enough that hobby projects pay for it.

So product-lead and release/SRE are part of base ariadne, as modules on top of the
SDLC core. Two rules keep them from making ariadne heavier for a new user:
1. **Cost nothing until used.** The quick flow is the precedent. A project that
   never invokes the product-lead still works (`verify` defaults to operator
   without any judgement), and a project with nothing to release never meets the
   release stage.
2. **Depend on the core, never the reverse.** The modules consume SDLC's contracts
   (cards, gates, observations); the SDLC core never imports them. This keeps a
   later split cheap if one turns out wrong.

Constraints that hold regardless:
- state lives in the repository (messages are hints, not records);
- the gates stay binary-owned;
- brain stays capture-only.
