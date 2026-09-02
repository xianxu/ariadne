---
id: 000210
status: open
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# fleet plan test reads a hardcoded workshop/plans path that archiving breaks

## Problem

`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`
(`cmd/sdlc/fleet_plan_test.go:14`) opens
`workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md` by literal path.
`dfeba9c` archived that plan to `workshop/history/`, which is the normal end of
a plan's life — so the test has been red ever since, for the intended behavior
of a different verb.

The suite is therefore red at every close boundary. That is corrosive in a way
the individual failure is not: a permanently-failing test trains every reader
(and every `--verified` claim) to explain a failure away instead of reading it,
and a real regression landing next to it is indistinguishable from the noise.
The #206 close review flagged it from outside for exactly that reason.

The bug is the lookup, not the archiving. `sdlc resolve` already answers "where
is issue N's plan, wherever it currently lives" — the archive-inclusive
resolution built for #144 — and this test predates using it.

## Spec

Resolve the plan through the archive-inclusive lookup rather than a hardcoded
`workshop/plans/` path, so archiving a plan never reds a test that only wants to
read it. Prefer the existing resolver over a second glob (`ARCH-DRY`); if the
test genuinely wants "the plan as of when it was authoritative", say so and pin
it to a ref instead.

Then sweep the class: any other test reading a `workshop/plans/` or
`workshop/issues/` path by literal string has the same latent failure, armed by
the ordinary act of closing the issue it belongs to.

## Done when

- The fleet-plan test passes on a tree where `000200`'s plan is archived, and
  keeps passing if it is archived again from a different location.
- No test in `cmd/` reads a plan or issue artifact by hardcoded live-directory
  path without going through the archive-inclusive resolver — enforced, not
  just fixed at the one site.
- `go test ./cmd/... ./pkg/...` is fully green, so a red suite is once again
  information.

## Plan

- [ ] Point the fleet-plan test at the archive-inclusive resolver.
- [ ] Enumerate the sibling literal-path reads in `cmd/*_test.go` and sweep them.
- [ ] Guard the class so a new one can't be added silently.

## Log

### 2026-09-02

Filed from the #206 close review, which found it while measuring whether that
issue's own fixes went red on revert — the pre-existing failure was noise it had
to work around to get a clean signal. Not caused by #206; its base already had
it.
