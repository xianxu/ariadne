---
id: '000106'
status: done
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 4
actual_hours: 1.5
---

# sdlc propagate-base — first-class propagation of a base-layer change to all recursive dependents

## Problem

The SDLC handles an OWNER's base-layer change well — `claim → start-plan →
change-code (branch) → milestone/boundary review → pr → merge`. But propagating
that landed change DOWNSTREAM to all recursive dependents is INFORMAL: you `cd`
into each dependent, `make weave`, eyeball the diff, and commit the re-weave
output by hand. There is no first-class verb, so it gets done ad hoc — and
inconsistently with the owner's own discipline.

Surfaced concretely on **#104 M3** (the 10-repo skill migration) and earlier on
**#95 M5** (the weave cutover): the owner (ariadne) went through a real
branch+PR+review, but the 9 dependents got commits made **directly on their
`main`** (operator caught the asymmetry). Direct-on-main for a base-layer
consumption is:
- inconsistent with "if on the default branch, branch first";
- not reviewable or cleanly reversible per repo;
- error-prone at scale (9–10 repos, hand-driven, on tired context — exactly where
  the brain sandbox/gcrypt snarls bite);
- silently skippable (a dependent never re-wove drifts from the new base — the
  shared-binary hazard: `make weave` builds the OWNER's checkout, so a stale
  dependent regenerates differently).

Structurally, this is the **reverse of `substrateChain`**. We have owner→ancestor
resolution (`cmd/sdlc/startplan.go:substrateChain` walks `construct/deps` UP). We
have NO downstream walk — "who depends on this repo, recursively, across the
sibling repos" — which is exactly what propagation needs.
