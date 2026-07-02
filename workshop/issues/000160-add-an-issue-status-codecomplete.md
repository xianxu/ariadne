---
id: 000160
status: working
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-02
estimate_hours: 4
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

### Resolved design decisions

- **Q1 — verb semantics → RESOLVED (operator).** `sdlc close` = "finish locally →
  `codecomplete`" (the LLM acceptance gate); `sdlc merge` = "→ `done`" (deterministic
  publish). Operator: *"final merge would do a simple flip to done, and merge, and
  push."*
- **Q2 — actual_hours timing → RESOLVED.** Actuals are measured at close, so a
  `codecomplete` issue carries them; the compiled `actual_hours!` guard extends from
  `done` to `{done, codecomplete}`.
- **Q3 — `sdlc push` (direct-to-main) → RESOLVED (agent judgment; operator away).**
  Push = the merge-equivalent for the no-PR path: it runs the deterministic
  reviewed-HEAD-unchanged check + flips `codecomplete → done` + pushes. One lifecycle
  everywhere; push is "merge without a PR." *(Revisit if the operator prefers
  legacy working→done for direct push.)*
- **Q4 — `lessons` → RESOLVED (operator: move to close).** The no-LLM reminder ping
  moves from the publish gate to `sdlc close` — it fires while the agent is engaged
  and boundary-review findings are fresh (lessons usually come straight out of the
  review). The publish gate (merge/push) runs no LLM and no reminder — purely
  deterministic.
- **Q5 — invariant mechanism → RESOLVED (operator: Option B).** The `codecomplete`
  boundary = **the commit that set `status: codecomplete`** in the issue file
  (derived from the issue file's git history — `git log` for the issue path). `sdlc
  merge`/`push` refuse unless `HEAD` is that commit, i.e. nothing landed after close.
  No dependency on a hand-pasted `Review-Verdict:` git trailer. **Sub-decision: the
  AGENT commits the flip** (keeps today's "close mutates, agent commits/bundles"
  convention); merge derives the anchor from git history — merge's existing
  clean-tree + branch-pushed guards already prevent an uncommitted-close footgun.
  The `Review-Verdict:` trailer still records the *verdict* for the audit trail and
  still anchors the *milestone* window (`previousReviewBoundary`); it just isn't
  this invariant's anchor.

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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.2 impl=0.25
item: smaller-go-module      design=0.25 impl=0.4
item: greenfield-go-module   design=0.3 impl=0.5
item: smaller-go-module      design=0.2 impl=0.4
item: atlas-docs             design=0.15 impl=0.25
item: milestone-review       design=0.0 impl=0.8
design-buffer: 0.30
total: 4.03
```

Derivation: M1 vocabulary (smaller-go-module), M2 close/set-status (smaller-go-module),
M3 publishgate.go greenfield (greenfield-go-module) + merge/push wiring
(smaller-go-module), help+atlas+README-gate (atlas-docs), and 4 review passes (3
milestone-closes + the whole-issue integration close). Much of the design is
already spent (brainstorm/Spec), which the design items + buffer capture.

## Plan

Detailed TDD plan: `workshop/plans/000160-codecomplete-status-plan.md`. Each `Mx` is
its own review boundary (`sdlc milestone-close`).

- [x] M1 — Vocabulary: add `codecomplete` to `issue.cue` + regenerate `pkg/vocab` + tests + atlas (+ set-status enforcement, pulled in — see Log)
- [ ] M2 — Close → codecomplete: flip target + `set-status` refusal + lessons-at-close + README docs gate (folded #142)
- [ ] M3 — Publish gate: `merge`/`push` flip `codecomplete → done` + the reviewed-HEAD-unchanged invariant + remove pre-merge LLM judges

## Log

### 2026-07-01

- Created.

### 2026-07-02
- 2026-07-02: closed M1 — codecomplete added to issue.cue (active status; working|blocked→codecomplete, codecomplete→done/working/wontfix/punt); conformance laws (reachable/escapable/documented-value) hold; pkg/vocab regenerated; set-status refuses →codecomplete and →done; go test ./cmd/sdlc/ ./pkg/vocab/ green; review verdict: FIX-THEN-SHIP

- Folded #142 in (operator: "fold #142 into #160"). Captured the two-gate model,
  the `codecomplete` invariant (deterministic reviewed-HEAD-unchanged check at
  merge, replacing #142's LLM delta-review), the base-layer vocabulary change, and
  the no-regret README docs-gate slice inherited from #142 Task 4.
- #142 set to `wontfix` (subsumed); its plan + audit remain the reference for the
  merge-side changes.
- Q1–Q5 resolved (operator confirmed Q1–Q4; Q5 = Option B). Durable plan authored,
  fresh-eyes reviewed twice, hardened. `sdlc change-code` gates passed (plan-quality
  + estimate-quality both INFO); in-place branch created.

#### M1 — Vocabulary (+ set-status enforcement)

- Added `codecomplete` to `issue.cue`: `categories.active`, a `when` line, and the
  lifecycle edges. Relocated **both** close edges (`working→` and `blocked→`) from
  `done` to `codecomplete` (close writes status unconditionally, so the model must
  match — plan-quality FAILURE fix); added `codecomplete→done` (merge, guard
  `reviewed-head-unchanged`), `codecomplete→working` (reopen), `codecomplete→
  wontfix|punt`. Extended the compiled `actual_hours!` guard to `{done, codecomplete}`.
- Regenerated `pkg/vocab/issue.json` (`make vocab-embed`); conformance laws
  (reachable/escapable/documented-value) all hold. Updated `vocab_test.go`
  (predicates + `AllStatuses` ordering).
- **Pulled set-status enforcement into M1** (was M2 Step 3): relocating `working→done`
  made it illegal, which broke `TestCheckTransitionGuards_RefusesDone` — the
  enforcement is the model's direct counterpart (ARCH-PURPOSE), so it belongs with
  the model. Added Guard 1b (`→ codecomplete` refused → route to `sdlc close`, the
  Q5 anchor-trust enforcement); updated Guard 1 (`→ done` routes through the
  close→merge publish flow); updated/added guard tests.
- Shadow-swept status consumers: `state.go` handles codecomplete gracefully (no
  change); `merge.go:568` archiving + `push.go` archiving/GH-close and
  `touchedIssuesNotDone` are M3 (flip-before-archive + carve-out). `close.go:486`
  status write is M2.
- Atlas: `issue-lifecycle.md` (states table + two-gate model + flow + closing),
  `vocabulary.md` (the codecomplete split + dual set-status refusals). Tests green:
  `go test ./cmd/sdlc/ ./pkg/vocab/`.
