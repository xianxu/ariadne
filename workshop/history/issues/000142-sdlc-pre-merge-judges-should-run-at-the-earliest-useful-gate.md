---
id: 000142
status: wontfix
deps: []
github_issue:
created: 2026-06-29
updated: 2026-07-02
estimate_hours:
started: 2026-07-01T23:05:44-07:00
---

# sdlc pre-merge judges should run at the earliest useful gate

## Problem

`sdlc merge` runs pre-merge judges after the PR branch is pushed:

- `plan` — issue plan completeness;
- `specs` — atlas/README sync;
- `lessons` — whether new lessons should be captured.

In pair#84, the `specs` judge found an optional README keybinding gap only at
merge time, after the issue had already been closed with a `SHIP` boundary
review and a PR had been opened. The fix was small, but it forced another
commit/push/merge loop.

Some of these checks may be more useful earlier:

- plan completeness belongs near `sdlc close`, before status becomes done;
- atlas/README sync may belong near close because it is part of "is the issue
  complete?";
- lessons might still be appropriate at merge, because it considers the whole
  branch/session and is explicitly a pre-ship reflection.

The current all-at-merge placement maximizes late discovery.

## Spec

Audit the pre-merge judges and decide the earliest useful gate for each
category.

Candidate placement:

- `plan`: run at `sdlc close` or immediately before close metadata finalization.
- `specs`: run at `sdlc close` for user-facing docs/atlas sync, possibly still
  repeated at merge for final branch-level drift.
- `lessons`: likely remain at merge, or become a soft post-close prompt, because
  it is less tied to issue acceptance criteria.

The goal is not to remove pre-merge safety. The goal is to catch fixable
acceptance/documentation gaps before close has recorded a final verdict and
before PR publishing/merge cleanup.

This issue should also decide whether a close-time `specs` or `plan` judge
overlaps with the boundary review and whether any overlap should be folded into
the close prompt instead of running another full LLM pass.

## Done when

- [ ] The existing pre-merge judge categories and prompts are documented with
      their current gate.
- [ ] Each category has an explicit target gate: close, PR, merge, or multiple.
- [ ] Checks that move earlier do not create redundant slow LLM passes when the
      boundary review already covers the same requirement.
- [ ] Merge still protects against final drift after PR creation.
- [ ] Help text explains which judge runs where and why.
- [ ] Tests cover that moved judges run at the new gate and failures stop before
      later/published steps.

## Plan

- [x] Inspect `cmd/sdlc/preflight.go`, `cmd/sdlc/merge.go`, and `cmd/sdlc/push.go`.
- [x] Compare `plan`, `specs`, and `lessons` prompts with close boundary review
      coverage. (Boundary review already covers plan + specs; see Log 2026-07-01.)
- [x] Decide whether to move, duplicate, or fold each check. (Fold plan+specs into
      close boundary review; post-close delta re-review at merge; lessons stays.)
- [ ] Implement the selected gate placement with flags preserving emergency
      bypass semantics. (Detailed in `workshop/plans/000142-earliest-useful-judge-gate-plan.md`.)
- [ ] Update close/merge/push help text.
- [ ] Add regression tests for a specs failure caught before merge.

## Log

### 2026-06-29

- Created from pair#84 dogfooding: `sdlc merge`'s `specs` judge caught a README
  keybinding documentation gap after close/PR, causing an avoidable final
  commit/push/merge loop.

### 2026-07-01

- Claimed; ran `sdlc start-plan`. Audited the judge gates.
- **Audit finding:** the `sdlc close`/`milestone-close` boundary review
  (`code-review.md`) already covers `plan` (via "Requirements traceability") and
  `specs` (via "Atlas update gate" + "Production readiness"). So the merge-time
  `plan`/`specs` judges are a **second full LLM pass** over what close already
  judged (ARCH-DRY), firing late (after close's verdict + PR). The pair#84 root
  cause: the boundary review's atlas gate under-covers **README** specifically.
- **Decision:** (1) strengthen the close boundary review's gate to own README
  sync (not just atlas) — earliest gate; (2) replace push/merge's blanket
  `plan+specs+lessons` pass with a **post-close delta re-review** — re-judge only
  the window since the last `Review-Verdict:` commit, **skipping the LLM passes
  when close already covered HEAD**; keep `lessons` + the #124 conformance gate at
  merge. Satisfies Done-when #3 (no redundant passes) AND #4 (post-PR drift still
  caught). Reuses `previousReviewBoundary` (generalized to `latestVerdictCommit`).
- Plan authored: `workshop/plans/000142-earliest-useful-judge-gate-plan.md`.
  Fresh-eyes plan review dispatched; hardened plan against 5 findings (finalizing-
  verdict-only skip; the `Atlas update gate` ×3 coupling; `-update-golden` flag).

### 2026-07-02 — reframing (operator steer)

- **Two-tier clarification.** Merge has two independent gates, only one server-side:
  - LLM judges (`plan`/`specs`/`lessons`) = **LLM, local, generic** — run *client-
    side* inside `sdlc merge`, dispatching a local agent. This is what #142 targets.
  - CI merge-check (`scripts/merge-checks.d/*`, `atlas/workflow/ci-merge-check.md`)
    = **deterministic, server-side, repo-specific** — GitHub Actions + pre-push
    hook. NOT an LLM. This is the genuinely server-side merge protection.
- **`plan` judge** = TPM reviewer over changed issue files (Plan checklist filled,
  done-but-unchecked items, Log entries, `status:` correct). **`specs` judge** =
  read-only docs reviewer comparing the diff to `atlas/` + `README.md` (the pair#84
  catcher).
- **Synthesis with #160 (codecomplete).** Operator's two-gate model: `sdlc close`
  (local) = LLM acceptance review → flips `working → codecomplete`; `sdlc merge`
  (remote) = deterministic CI + mechanical `codecomplete → done` flip + merge +
  push. Under it, **merge should run NO LLM judge** — all LLM review is close-time.
  #160's status invariant (*codecomplete ⟹ close reviewed HEAD*; any new commit
  drops back to `working`, forcing re-close) **structurally replaces** this issue's
  post-close delta-review — so #142 collapses to: (1) strengthen close's boundary
  review to own README sync [no-regret, needed in every option]; (2) delete
  `plan`+`specs` LLM judges from merge/push; (3) drop the delta-review machinery,
  delegating post-close-drift protection to #160's status machine.
- **OPEN — sequencing #142 × #160** (asked operator; awaiting answer): (a) #160
  first, then trivial #142; (b) fold #142 into #160 as one re-architecture; (c)
  ship the current delta-review plan now as a throwaway bridge. Recommendation:
  **(a) or (b)** — the delta-review is throwaway under the codecomplete model; the
  only no-regret slice to do now is the README-gate strengthening (Task 4).
- **Held implementation.** Not running `sdlc change-code` — the design is mid-
  reshape and unapproved; building the delta-review now risks throwaway core-SDLC
  code. Waiting on the sequencing decision.
- **FOLDED into #160 (operator chose "fold").** The two-gate model + `codecomplete`
  status is the coherent home for this work: the post-close delta-review is
  replaced by #160's deterministic *reviewed-HEAD-unchanged* invariant at merge
  (simpler + deterministic); the README docs-gate fix (Task 4) and the removal of
  the merge-time `plan`/`specs` LLM judges are inherited by #160. This issue →
  `wontfix` (subsumed). The audit here + `workshop/plans/000142-earliest-useful-
  judge-gate-plan.md` remain the reference for #160's merge-side changes.
