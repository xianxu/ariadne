---
id: 000272
status: open
deps: []
github_issue:
created: 2026-09-29
updated: 2026-09-29
estimate_hours:
card_mirror: '820594ec5fb58fdda250e1dfc16a0666d042d3a3' # card fields mirrored from issue-cards; edit via sdlc
---

# Support stacked issue development

## Problem

SDLC can carry several completed issues in one branch and PR, but a natural
stack of small dependent changes still requires manual Git coordination and
repeated review of already accepted ancestors. Support stacked development as
an explicit workflow while retaining exact evidence and publication guards.

The motivating session was parley.nvim#300 → parley.nvim#301 →
parley.nvim#302 → parley.nvim#303, proposed together in
[parley.nvim PR #215](https://github.com/xianxu/parley.nvim/pull/215):

| Issue / branch | Scope |
| --- | --- |
| parley.nvim#300 / `000300-app-labels-buffer-completion` | Outline question icons and Blink buffer source |
| parley.nvim#301 / `000301-local-outline-tags` | Whole-line `@@tag@@` excluded from context |
| parley.nvim#302 / `000302-pair-at-markers` | Pair markers when typing `@@` |
| parley.nvim#303 / `000303-app-completion-keys` | Tab/Down/Up/Enter/Esc completion keys |

New requests arrived while earlier work was still being tested. Stacking fit
the user's interaction: #302 used an isolated worktree while #301 fixes were
ongoing, then absorbed #301 review fixes and evidence; #303 inherited both.
The observations below describe that session, not a claim that every SDLC
configuration behaves identically.

1. Boundary reviews used the common main base (`4539950f..HEAD`) and repeatedly
   reviewed accumulated ancestor changes. More than 100 added production lines
   in the stack turned #303's five-line keymap change into a full review. Review
   time and scope grew with the stack instead of the new issue's change.
2. Initial issue details were published to main with `issue move-detail`, then
   restored on the working stack. A later `git merge origin/main` produced
   add/add conflicts in the #301/#302/#303 details files. Resolution retained
   completed details over blank main templates. No data was lost, but the
   initial-detail lineage required manual repair.
3. Parent/base relationships were implicit. Branch creation, worktree isolation,
   and merging ancestor fixes and evidence were manual. Architecture checks also
   attributed inherited helper exports to the descendant issue, requiring
   inherited Core concepts rows in its documentation.
4. Completion bindings were per issue while publication was branch-oriented.
   One combined PR could carry all four bindings, but cleanup of safe ancestor
   branches #300/#301/#302 after landing #303 was expected to be manual.
5. Active-time measurement emitted mention-fallback/unattributed diagnostics
   across #301/#302/#303. These are observations to reproduce and explain;
   stack-related misattribution has not been established.
6. An unrelated user demo-viewer commit arrived on the tip after #303 closed.
   Publication correctly rejected post-close code and required another review.
   That review covered the full stack, and viewer corrections became scoped
   into #303. The gap is identifying and reviewing the extra commit's scope
   without discarding earlier acceptance, not bypassing the drift guard.

Preserve what worked: cheap claims, independent issue records, binary-owned
fresh-context close reviews, multiple completion bindings in one PR, and a
deterministic publish gate that caught real post-close drift.

## Spec

Explore first-class support for a stack of issue branches with interleaved
ancestor corrections and a combined landing. This ticket records requirements
and candidate directions; it does not choose a data model or command syntax.

- Consider explicit parent/base and issue ownership metadata shared by review,
  architecture accounting, evidence validation, publication, and cleanup
  (ARCH-DRY). Distinguish inherited changes from a descendant's own changes.
- Review an issue's incremental change against a known parent state. Preserve
  accepted ancestor evidence only while its relevant identities remain valid;
  ancestor corrections must invalidate affected acceptance. Keep final
  integration validation across the selected landing.
- Explore stack-aware refresh/rebase/merge that propagates ancestor corrections
  and preserves issue bodies, review evidence, and initial-detail ancestry.
  Conflicts in genuine competing edits must remain visible.
- Make combined versus per-branch publication selection explicit. Establish
  safe cleanup of landed ancestors without deleting unlanded or independently
  advanced work (ARCH-FUNERAL).
- Define post-close extra-commit ownership and review policy. An unrelated
  addition must acquire its own explicit scope/evidence or reopen the relevant
  scope; unrelated code must not silently attach to the tip issue.
- Name legal transitions for ancestor drift, refresh, close, publication,
  interruption/retry, and cleanup (ARCH-ORDER). Keep exact evidence checks and
  actionable refusals; a stale ancestor verdict is not reusable by assertion.

## Done when

- A reproducible regression scenario exercises four small dependent issues,
  an isolated descendant worktree, ancestor review corrections arriving during
  descendant work, and one combined landing with all four completion bindings.
- Review diagnostics identify each selected base/head, owned delta, inherited
  acceptance, and invalidation reason. Unchanged accepted ancestors do not
  trigger redundant full-stack reviews at every close; changed ancestor code
  invalidates affected evidence, and final integration is validated.
- Refreshing the scenario from main retains completed issue details and their
  evidence without add/add conflicts caused solely by initial-detail
  publication; genuine conflicting edits are reported without data loss.
- Issue and architecture-artifact attribution separates inherited helper
  exports from new exports. Active-time diagnostics in the scenario are
  explained and tested, with any reproducible attribution defect corrected.
- Post-close unrelated code is detected and assigned explicit review scope;
  prior valid acceptance survives, and no unreviewed code can publish through
  evidence reuse. Regression tests cover both permitted reuse and refusal.
- A combined landing completes the selected issues and safely cleans up
  eligible ancestor branches/worktrees, preserving independently advanced,
  unlanded, dirty, or otherwise unproven work. Interrupted operations can be
  inspected and retried without losing artifacts or recording false completion.
- CLI help and workflow documentation explain supported stacking, refresh,
  publication choices, drift handling, and cleanup; any unsupported topology
  has an explicit diagnostic rather than requiring guessed Git commands.

## Plan

Design pending. Reproduce the session topology before choosing metadata,
commands, or implementation boundaries.

## Log

### 2026-09-29

Filed at the user's request from the parley.nvim stacked-development session.
This is a follow-up ticket only; no implementation or claim is requested.
PR #215 was still in publication work when this record was authored, so its
eventual landing and cleanup outcome are not asserted here.
