---
type: pensive
date: 2026-07-01
topic: sdlc multi-milestone ship-flow gotchas
mode: thoughts
description: "Transferable close/merge lessons from shipping a two-milestone issue (#153) reviewed-together — verdict gate, milestone-close vs close --milestone, conformance + specs gates, issue-sync on a branch"
references: [workshop/history/000153-sdlc-retro-process-manual.md, workshop/issues/000156-change-code-idempotent-branch-for-milestone-re-runs.md]
---

# sdlc multi-milestone ship-flow gotchas

Shipping #153 (M1+M2, reviewed *together*) surfaced a cluster of close/merge-flow
frictions worth remembering — none are in an issue as a unit.

- **`milestone-close` is the REVIEWED path; `close --milestone Mx` is the no-judge
  escape.** `sdlc close --milestone` does the mechanical close but does NOT dispatch the
  fresh-context review. Worse trap: running `sdlc milestone-close` *without* `--actual`
  stops at the actual gate and its re-run hint points you at `sdlc close --milestone …`
  (the non-reviewing path). To get the mandatory review, re-run
  `sdlc milestone-close --actual <n> --verified '<ev>'`, not `close`. (I fell into this;
  had to reset the no-judge finalize and re-run the reviewed path.)

- **"Review-together" milestones need `--no-verdict` at issue close.** When M1 is never
  separately milestone-closed (M1+M2 reviewed as one boundary at M2's close), the
  whole-issue close's verdict gate flags "M1 lacks a Review-Verdict trailer." Legit
  `--no-verdict` with the reason in `--verified` (M1 covered by the combined window).

- **Boundaries re-review.** M2's milestone-close reviewed branch-point→HEAD (all M1+M2);
  the whole-issue close re-reviewed the same window + the post-review fix commit. Somewhat
  redundant for a combined-milestone issue, but it does catch the delta a milestone review
  didn't see (fixes committed *after* that review ran).

- **`sdlc issue new` can't auto-sync to main from a feature branch** (main not checked out
  in a worktree) → the new issue file is untracked on the branch. `sdlc merge` recognizes
  these as "tracker files sync out-of-band (#82)" and doesn't block, but they must land on
  main separately (I committed #156/#157 directly to main post-merge).

- **The merge instance-conformance gate validates the whole branch diff's issue files** —
  including pre-existing unrelated ones. Two stale bug issues (#154/#155) with placeholder
  Plans blocked #153's merge. Fix (add a Plan item) + push, or `--no-validate`.

- **The merge specs judge catches stale atlas.** Renaming M2 (session-recon → prompts-md)
  + splitting M3→#157 left "M2 adds session reconstruction" stale in `atlas/index.md`; the
  pre-merge specs judge FAILED on it. Atlas milestone labels must track renames/splits.

- **Bare `git push` can silently no-op** (coincided with a `failed to store: 100001`
  keychain warning). Use `git push origin HEAD` and verify the `<old>..<new>` ref line.

- **`change-code` isn't idempotent for a 2nd-milestone re-run** (branch already exists) →
  filed #156. Worked around it (gates run first; only the branch-create step errors).

Meta: an issue split into "review-together" milestones trades one clean boundary review for
several gate-flag detours at close. If milestones will genuinely close separately, tag them
and close each; if they'll ship as one unit, consider *not* tagging Mx and closing once.
