---
id: 000160
status: working
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-02
estimate_hours:
started: 2026-07-02T11:15:34-07:00
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

## Spec

### The two-gate model

- **`sdlc close` — the LOCAL acceptance gate (all LLM review lives here).** Runs
  the fresh-context boundary review (code quality, requirements traceability,
  docs/atlas + **README** sync, architecture). On a finalizing verdict it flips
  `working → codecomplete` (NOT `done`). This is the *only* place LLM review runs.
- **`sdlc merge` — the REMOTE publish gate (deterministic, no LLM).** Runs the
  server-side CI merge-check (`scripts/merge-checks.d/*`, `atlas/workflow/ci-merge-check.md`),
  then mechanically flips `codecomplete → done`, merges, and pushes. **No LLM judge.**

### The load-bearing invariant

> `codecomplete ⟹ the close boundary review covered HEAD.`

`sdlc merge` verifies this deterministically before flipping to `done`: HEAD must
equal the commit carrying the `codecomplete`-establishing `Review-Verdict:` trailer.
If commits landed after close, merge **refuses**: *"commits landed after `sdlc
close`; re-run it to re-review."* This deterministic check **replaces** #142's
proposed post-close LLM delta-review — same guarantee (no un-reviewed drift reaches
main), but deterministic, and it forces a real re-review instead of silently
re-judging a delta. It reuses the existing `Review-Verdict:` trailer machinery
(`previousReviewBoundary`, `ParseVerdictTrailer`, `vocab.Verdict().IsFinalizing`).

### What folds in from #142

1. **Strengthen the close boundary review to own README docs sync** (not just
   `atlas/`) — the pair#84 root cause. Mandatory now that close is the sole LLM
   gate (merge won't catch it). *No-regret; already planned as #142 Task 4.*
2. **Remove the `plan` + `specs` LLM judges from `sdlc merge`/`push`.** `plan` is
   already covered by close's Requirements-traceability + the deterministic #124
   conformance gate; `specs` by close's strengthened docs gate.
3. `lessons` (a no-LLM reminder ping): keep at merge as a pre-ship reflection, or
   drop. **[open — Q4]**

### Status-model changes (`construct/vocabulary/issue.cue`)

- Add `codecomplete` to `categories.active` (it is active, not terminal — work
  isn't finished until merged).
- `when.codecomplete: "code complete; passed local acceptance review, awaiting merge"`.
- Lifecycle transitions:
  - `working → codecomplete` (event `close`, guards: actual-recorded, verified,
    atlas-updated, boundary-review) — relocated off today's `working → done`.
  - `codecomplete → done` (event `merge`, guard: reviewed-HEAD-unchanged + merged).
  - `codecomplete → working` (event `reopen`/`rework` — new commits after close, or
    reviewer feedback).
  - `codecomplete → wontfix | punt` (abandon/defer late).
- Extend the compiled guard `if status == "done" { actual_hours! }` to also cover
  `codecomplete` (actuals are measured at close, which now yields codecomplete).
- This is a **base-layer vocabulary change** — it propagates to every downstream
  repo via the manifest. Weigh downstream impact (`atlas/workflow/base-layer.md`).

### Open design questions (decide before planning)

- **Q1 — verb semantics.** Keep `sdlc close` = "finish locally → codecomplete", and
  give `sdlc merge` the `→ done` flip? (Proposed.) Or introduce a distinct verb?
- **Q2 — actual_hours timing.** Actuals computed at close → codecomplete carries
  them; the done-guard extends to codecomplete. Confirm.
- **Q3 — `sdlc push` (direct-to-main, no PR).** It has no separate merge step. Does
  push do close(→codecomplete) + flip(→done) atomically, or does direct-push keep
  the legacy working→done? (Proposed: push = the merge-equivalent flip for the
  no-PR path — deterministic HEAD check + `→ done`.)
- **Q4 — `lessons` placement** (keep at merge / drop).
- **Q5 — the boundary review at close still emits a `Review-Verdict:` trailer the
  operator pastes into the codecomplete commit** (the invariant's anchor). Confirm
  this stays the mechanism, or make close auto-commit the trailer.

## Done when

- [ ] `codecomplete` is in `construct/vocabulary/issue.cue` with `when` + lifecycle
      transitions; `pkg/vocab` consumers (set-status help, gates) derive it — no
      hardcoded enum.
- [ ] `sdlc close` flips `working → codecomplete` (not done) on a finalizing
      boundary-review verdict, carrying the actual/verified/atlas guards.
- [ ] `sdlc merge` flips `codecomplete → done` after the deterministic
      reviewed-HEAD-unchanged check + CI, running no LLM judge.
- [ ] `sdlc merge` refuses when commits landed after close (invariant enforced),
      with a re-run-`sdlc close` next-action message.
- [ ] The close boundary review owns README docs sync (folded #142 pair#84 fix).
- [ ] `plan`/`specs` LLM judges removed from merge/push; close/merge/push help
      text explains the two-gate model.
- [ ] Tests: the `working → codecomplete → done` path; merge refuses on post-close
      drift; a README gap caught at close (not merge).

## Plan

- [ ] (design first — brainstorm the open questions, then author the durable plan)

## Log

### 2026-07-01

- Created.

### 2026-07-02

- Folded #142 in (operator: "fold #142 into #160"). Captured the two-gate model,
  the `codecomplete` invariant (deterministic reviewed-HEAD-unchanged check at
  merge, replacing #142's LLM delta-review), the base-layer vocabulary change, and
  the no-regret README docs-gate slice inherited from #142 Task 4.
- #142 set to `wontfix` (subsumed); its plan + audit remain the reference for the
  merge-side changes.
- Open questions Q1–Q5 above need operator decisions before the durable plan.
