---
id: 000108
status: punt
deps: []
github_issue:
created: 2026-06-16
updated: 2026-06-16
estimate_hours:
---

# Continuation-doc garbage collection: define removal semantics

## Problem

Continuation docs (`workshop/continuation/*.md`) accumulate and go **stale**, but
there's no defined process for removing old ones — and removal isn't a sure bet
because a continuation is **multi-faceted**. The same artifact serves at least
three distinct purposes, with different lifetimes:

1. **Session compaction** — a within-session hand-off when context fills up.
   Ephemeral; safe to discard once the next session absorbs it.
2. **Agent-harness switch** — handing work from one agent/harness to another.
   Also short-lived; obsolete once the receiving harness picks up.
3. **Operator/machine hand-off** — a durable hand-off to *another operator on
   another machine* (the reason continuations live in the repo, not a local
   scratch dir). This one is genuinely durable and must NOT be GC'd prematurely.

Because (3) exists, a blunt "delete continuations older than N days" or
"delete on merge" rule would destroy legitimate cross-operator hand-offs. Yet
(1)/(2) pile up as noise and actively mislead (a stale continuation that
describes a since-changed plan is worse than none — observed concretely on
parley.nvim#128, where a 06-14 breadcrumb described an unmerged branch + a
design that was later re-scoped).

The cheap interim mitigation today is: once the work a continuation pointed at
lands (e.g. merged to main / the issue+plan become the authoritative surface),
the continuation is obsolete and can be deleted by hand. But that's ad hoc and
relies on the author remembering.

## Spec

Define the **removal semantics** for continuation docs — not necessarily an
automated GC yet, but a clear contract for *when a continuation is safe to
remove* and *who/what removes it*. Open questions to settle:

- **Can a continuation declare its own kind/lifetime?** e.g. a frontmatter field
  distinguishing `compaction` / `harness-switch` (ephemeral) from `handoff`
  (durable) — so a GC can treat them differently. (Weigh against
  [feedback: prefer single mechanism over parallel tracks] — don't bake in
  metadata if an in-context signal or a single mechanism with triggers suffices.)
- **What's the "consumed" signal?** A continuation is obsolete once its NEXT
  ACTION is done. Can that be detected — e.g. the referenced branch is merged,
  the referenced issue/milestone is closed, or a newer continuation supersedes
  it? Or does it require human judgment?
- **Where does removal happen in the SDLC?** Candidate hooks: `sdlc merge`
  (drop continuations whose branch just merged), a periodic sweep, or a
  `sdlc continuation gc` verb that lists stale candidates for human confirm
  (single-threaded-attention: a queue/confirm, not silent deletion).
- **Durable hand-offs:** how does a (3)-type continuation signal "keep me until
  the receiving operator acknowledges"? (Acknowledgement = the work starts /
  a reply continuation appears?)

Lean: **a `sdlc continuation` verb that proposes stale candidates** (by the
consumed-signal heuristics) for human confirmation, rather than any automatic
deletion — preserves the durable-handoff case and keeps a human in the loop.

## Done when

- The semantics of "a continuation is safe to remove" are documented (which
  kinds, what signal, who acts) — in the continuation datatype/convention doc.
- A decision recorded on automated vs. confirm-only GC (and any frontmatter the
  kinds require, or an explicit decision NOT to add any).

## Plan

- [ ] brainstorm the kinds + the consumed-signal heuristics; decide frontmatter-or-not
- [ ] document the removal contract in the continuation convention/datatype doc
- [ ] (if approved) implement the `sdlc continuation` GC-candidate lister/confirm verb

## Log

### 2026-06-16

Filed from a parley.nvim session. Trigger: a stale continuation breadcrumb on
parley.nvim `main` (`workshop/continuation/20260614T…-resume.md`) described an
unmerged branch + a design that was subsequently re-scoped, after the work it
pointed at had merged — i.e. it had become misleading noise. Operator notes
continuations are multi-faceted (compaction / harness-switch / operator-machine
hand-off), so there's no obvious sure-bet removal rule; worth defining the
semantics rather than a blunt time/merge-based purge.
