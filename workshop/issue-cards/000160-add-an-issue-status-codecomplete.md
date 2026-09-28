---
id: '000160'
status: done
started: 2026-07-02T11:15:34-07:00
created: 2026-07-01
updated: 2026-07-02
estimate_hours: 4
actual_hours: 4.72
---

# add an issue status: codecomplete

## Problem

Agents routinely mark an issue `done` and run `sdlc close`/`merge`, only to hit
closing-gate feedback (boundary-review findings, docs gaps) — then loop back to
fix, re-commit, re-push. `done` is being used as "I think I'm finished" when it
should mean "verified, merged, nothing left but history." We're missing the
intermediate state: **the agent believes the code is complete and it has passed
the local acceptance gate, but it isn't published yet.**

This issue also **subsumes #142** (*pre-merge judges should run at the earliest
useful gate*). #142's audit found that `sdlc merge`'s LLM judges (`plan`/`specs`)
are a *local, client-side* second pass duplicating the `sdlc close` boundary
review (which already covers plan completeness + docs sync), firing late — after
close's verdict and the PR (the pair#84 loop). The clean fix is the two-gate model
below, which only becomes coherent once `codecomplete` exists — so #142 folds here.
See `workshop/plans/000142-earliest-useful-judge-gate-plan.md` for the audit + the
no-regret README-gate slice this issue inherits.
