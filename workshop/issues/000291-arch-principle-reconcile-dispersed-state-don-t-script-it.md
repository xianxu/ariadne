---
id: 000291
status: open
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: '2e19e8eea249e5cd63e6e841e9a0091e0dac1a75' # card fields mirrored from issue-cards; edit via sdlc
---

# ARCH principle: reconcile dispersed state, don't script it

## Problem

When a feature's state is spread across several systems, agents keep designing
it as scripted step sequences with compensating undo (a saga / phase journal).
Each review round then finds another partial-failure path the script does not
cover, because the design encodes paths through a global state space that grows
exponentially with the number of pieces. Observed on pair#367 (2026-10-05): a
`couch --rebuild` design for a Couch slot, whose state spans the git worktree
registration, the checkout, weave's setup marker, dependency clones and generated
files, Couch's slot store and thread record, the agent session, and sdlc claims,
went through four review rounds of uncovered paths and was dropped in favor of a
reconciler (pair#387). The operator's stated preference: dispersed state is
modeled and reconciled, not patched.

The ARCH-* registry (`cmd/sdlc/internal/judge/architecture.md`) has ARCH-ORDER
(make ordering and state transitions explicit) but nothing that steers designs
over state in several systems toward reconciliation, so the gates neither push it
at plan time nor flag the saga shape at review.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

Add a registry entry (working name `ARCH-RECONCILE`, "Reconcile dispersed state;
don't script it") with the registry's required fields (quick-flow, principle,
at-plan, at-review):

- **Principle.** When state lives in several systems or stores, model each piece
  as a resource with a desired spec and a small, idempotent observe → compare →
  converge step; declare dependencies between resources as a graph; converge in
  topological order (teardown in reverse). Recovery from any partial failure is
  the same operation as creation from scratch: run the reconciler again from
  whatever is observed (level-triggered, not event-replayed). User data is a
  resource to preserve, never converge over; an unknown observation stops the
  walk at that node instead of guessing. Plan/apply: an observable diff before
  acting. Prior art: Kubernetes controllers, Terraform plan/apply,
  CFEngine/promise theory.
- **at-plan.** For state across ≥2 systems: require the resource table (owner,
  kind: internal / derived / user data / external, desired spec, converge step,
  preservation rule), the dependency graph, and how a re-run converges from each
  partial state. Flag a design that is a step sequence with undo over such
  state unless it argues the state truly is a sequence.
- **at-review.** Flag phase journals or sagas over multi-system state;
  non-idempotent converge steps; global-state enumerations that grow with the
  product of components; recovery paths that differ from the create path; user
  data that can be overwritten or deleted without a preservation step; unknown
  observations collapsed into a converge decision.
- **Boundary with ARCH-ORDER** stated in both entries: ARCH-ORDER governs
  ordering within one owner's transitions; ARCH-RECONCILE governs state owned by
  several systems. Decide quick-flow (likely `no`; the bug class needs multi-file
  designs).
- Mirror the human narrative in AGENTS.md "Core Design Principles" (and the atlas
  page `atlas/workflow/architecture-principles.md`) as the registry's own docs
  require.

## Done when

- The registry has the new entry with every required field, and the build's
  field checks pass.
- `sdlc arch-principles`, `start-plan` and the plan-quality and boundary-review
  prompts deliver it (existing embedding tests cover the new entry).
- ARCH-ORDER and the new entry each state their boundary with the other.
- AGENTS.md "Core Design Principles" and the atlas page reflect it.

## Plan

Implementation plan to be designed after issue claim and start-plan.

## Log

### 2026-10-05

Filed at the operator's request from pair#367: reconciliation is the operator's
preferred model for state dispersed across many systems; pair#387 (the slot
reconciler) is its first application.
