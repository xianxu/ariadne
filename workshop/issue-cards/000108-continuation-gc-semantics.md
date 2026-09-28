---
id: '000108'
status: punt
created: 2026-06-16
updated: 2026-06-16
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
