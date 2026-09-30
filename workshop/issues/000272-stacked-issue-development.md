---
id: 000272
status: working
deps: []
github_issue:
created: 2026-09-29
updated: 2026-09-29
estimate_hours:
card_mirror: 'b89d35a5c4a39113ba675f60a1491b2306426d0c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-29T16:14:23-07:00
flow: {kind: quick, provenance: inferred, spec: "b067f456", done: "3946662b"}
---

# One issue per branch: no work on an unlanded base

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
7. At final landing, `sdlc merge` accepted #303's latest full-review evidence
   with documentation-only drift, then refused because #300 was quick-flow
   and "its diff grew to 385 added lines in code files" (limit 200), directing
   another `sdlc close --issue 300` to upgrade it to full review. #300's own
   code was unchanged; descendant code was charged against the ancestor's
   quick-flow budget despite #303's full review covering the combined HEAD.

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

### Direction (2026-09-29): ban stacked development

Brainstorm outcome with the operator; supersedes the exploratory bullets above
(kept for the record). Stacks were an accidental buffer in front of the one
serial resource in AI coding — the operator's smoke test, which doubles as
reading the spec back and grounding it. Slots, not stacks, are the buffer.

- **One issue, one branch.** The branch is named with the full six-digit issue
  ID prefix (`000272-…`) and carries only that issue's work.
- **Never start work on an unlanded base.** An issue branch starts from main;
  commits reachable only from another unlanded issue branch are refused with a
  diagnostic, not supported. Merging main in stays legal.
- **Discoveries mid-issue:** a change to the current item's contour folds into
  the current issue (a scope revision); independent work goes to another slot
  from main; work that depends on unlanded code is filed and waits for its
  parent to land.
- **Out of scope here, follow-ups:** grounding tests (shrink the set of changes
  that need the operator's smoke test; make the rest cheap) and live cross-slot
  dispatch in Couch (pair repo) so kicking off a slot costs no attention.
- **Grounding tests are per repo**, at different maturity; the long-run vision
  (pensive) is a testing agent that reads the user manual and simulates a
  typical user. Filed per product, not here.

## Done when

- `sdlc start-plan` refuses to plan an issue on a branch that carries another
  issue's unlanded work: when the checkout is already on, or switches to, an
  existing issue branch whose commits beyond fresh main include commits of an
  unlanded issue branch it does not lead, it names that branch and the next
  action (land it, or restart the design from main). Creating a new issue
  branch keeps branching at freshly pinned main, named with the six-digit ID.
- Regression tests: the parley topology (a descendant branch started on an
  unlanded ancestor) is refused; a branch that merged main in, a branch that
  an unlanded descendant was built on, and a branch sharing only landed
  commits all still pass.
- Soft instruction, not gates: the constitution and the SDLC atlas state one
  issue per branch and never starting on an unlanded base, and how to handle
  mid-issue discoveries (fold, dispatch to another slot, or file and wait).
  No `change-code`/`close`/publish guard is added (operator decision).
- The handoff-guard ownership defect is filed as its own issue.

## Plan

- [x] Retitle and rewrite Done when against the ban.
- [x] File the handoff-guard ownership bug.
- [x] start-plan: foreign-unlanded-base check on existing issue branches
  (`preparePlanningBranch`), with tests.
- [x] Docs: constitution line, atlas workflow page, `start-plan --help`.

## Log

### 2026-09-29
- 2026-09-29: closed — TestPreparePlanningBranchRefusesAnUnlandedBase: 10 cases (3 refusals incl. switch path and remote-only parent; merged main on both paths, child, child-then-parent-advanced, parent commit mentioning child, #31-vs-#3 boundary, landed parent pass); refusal cases fail with guard stubbed; cmd/sdlc planning/start-plan/migrate tests green; review verdict: FIX-THEN-SHIP
- 2026-09-29: closed — TestPreparePlanningBranchRefusesAnUnlandedBase: 8 cases (3 refusals incl. switch path and remote-only parent; merged main, child, child-then-parent-advanced (BR-1), #31-vs-#3 tag boundary, landed parent pass); refusal cases fail with guard stubbed; make test cmd/sdlc green earlier (processgroup fails only under sandbox /bin/ps; TestClose_MilestoneRefusesWithRedirect flaky, passes alone); review verdict: SHIP

Filed at the user's request from the parley.nvim stacked-development session.
This is a follow-up ticket only; no implementation or claim is requested.
PR #215 was still in publication work when this record was authored, so its
eventual landing and cleanup outcome are not asserted here.

Ticket-authoring follow-up: initial details were published from ariadne's main
with `sdlc issue move-detail --issue 272` at `0f57ded303b1`. A subsequent
schema correction (make the deferred Plan step a checklist) was committed on
main, but `sdlc push --yes` refused: "landing would change handed-off issue
details ... Its details were handed off to main ... and may have a new owner."
The prescribed `sdlc issue recovery reconcile --issue 272` reported no
unfinished operations. The valid correction and the later #300 quick-budget
finding are committed locally; their publication is blocked by that guard.
This is additional observed workflow friction during ticket creation, distinct
from the stack's acceptance scope. No guard was bypassed or implementation
attempted.

Final stacked-landing follow-up: parley.nvim#300 received a full re-review
with SHIP verdict at `fa255b82` on the combined #303 branch, and the latest
branch was pushed to PR #215. `sdlc merge` passed conformance, then refused:
"landing would change handed-off issue details:
workshop/issues/000300-app-labels-buffer-completion.md would be overwritten by
this branch. Its details were handed off to main (sdlc issue move-detail) and
may have a new owner." The suggested `sdlc issue recovery reconcile --issue
300` returned "no unfinished operations". Main held the initial template;
the branch held the reviewed completed record, so restoring main's version
would discard the issue body and history. No guard was bypassed. The branch
remained pushed, PR #215 open, and main unmerged at this checkpoint. The same
handoff guard therefore affects both the stack's landing and this ticket's
on-main follow-up publication.

Recovery outcome: the user explicitly authorized manual merging.
`gh pr merge 215 --merge --match-head-commit fa255b82` landed the reviewed
stack at `aedd9d66`. Local main was fast-forwarded to the fetched merge and
switching to main preserved tracked demo edits. Running
`sdlc merge --branch 000303-app-completion-keys --yes` from resting main then
recognized the already-merged PR, archived all four issues and their reviews
remotely at `7d2ffa67`, set all four cards to done with
`landed_commit: aedd9d66`, and removed tip branch #303. After fast-forwarding
the archive commit, safe `git branch -d` deletion of ancestor branches
#300/#301/#302 succeeded. No changes were lost; the untracked recording was
also preserved. This narrows the observed gap: the pre-merge handoff guard
blocked the stack, while post-merge recovery accepted the same lineage and
completion bindings. Preserve this successful recovery behavior in regression
coverage; manual forge merging required explicit user authorization and is
not the proposed normal stack workflow.

Claimed; brainstorm pivoted to banning stacked development (Spec → Direction).
Follow-ups filed: pair#352 (`couch --notify`, stateful operator notifications
as the verification inbox), pair#353 (live cross-slot dispatch; the sender
files in the target repo, pair injects peer messages only at a safe insertion
point), and ariadne#273 (rename `move-detail` → `publish-detail`). No
grounding-test issue yet: per repo, and still visionary.

Operator clarification on `move-detail` (arrived after #273/pair#353 were
filed, so their Specs need it folded in when claimed): from a **feature
branch** it stays an escape hatch — details found while testing are kept
private until the branch lands, unless another slot must start sooner. From a
**resting branch** (e.g. a cross-repo issue filed in a free slot for
dispatch) publishing is the normal path and should be smooth once dispatch
exists. #273's "not an escape hatch" framing is therefore only half right.

Implemented the start-plan guard: `refuseUnlandedBase` (planningbranch.go)
runs when the checkout is already on, or switches to, an existing issue
branch. It refuses when the branch shares commits beyond main with another
unlanded `NNNNNN-*` branch (local or on the publication remote) that is not
built on it — ownership read from refs, not subject tags, so it needs no
commit convention. Merged-in main and landed parents never count
(ARCH-DRY: reuses trackerEnv's pinned main). Six regression cases in
`TestPreparePlanningBranchRefusesAnUnlandedBase`; the three refusal cases
fail with the guard stubbed out. Soft instruction added to AGENTS.base.md
§2; start-plan help and the issue-tracker atlas table updated. `make test`:
cmd/sdlc green; `internal/processgroup` fails only under the sandbox
(`/bin/ps` not permitted); `TestClose_MilestoneRefusesWithRedirect` flaked in
2 of 3 sharded runs (different shards each time) and passes alone —
unrelated to this change. Composed AGENTS.md/CLAUDE.md are gitignored;
`weave compile` could not rewrite `.claude/settings.json` in the sandbox, so
the operator recomposes locally.

Close review round 1 (FIX-THEN-SHIP): BR-1 Important — the ancestry-based
exemption refused a parent once it advanced past an unlanded child's fork.
Fixed at the class (lesson added): ownership of a shared commit is read from
its `#N` subject tag, not from topology; untagged shared commits are left to
the soft instruction. The same rewrite removes the merge-base error swallow
(Minor) and reuses the pinned main on the switch path instead of fetching it
twice (Minor; the already-on-branch path still fetches once, by design).
Regression cases added: a child built on this branch that then advanced, and
a shared commit tagged `#31` not matching issue #3. An earlier close attempt
recorded `unknown` because the sandbox blocked the reviewer's network.

Close advisory Minors fixed in the same round: a commit's owner is its
subject's first `#N` (`commitIssue`), so "#9: prep hook for #10" stays #9's;
the branch-prefix regex is shared with migrate.go's `issueFamilyRE`
(ARCH-DRY); a passing switch-path case was added. 10 cases now.

## Revisions

- 2026-09-29 — Added the final stacked-landing refusal after #300's full
  re-review cleared the earlier budget gate. Extended Done when to require
  evidence-based legitimate-owner admission and stale-owner refusal across
  initial-detail handoff, without discarding completed records.
- 2026-09-29 — Brainstorm pivot: ban stacked development instead of supporting
  it (see Spec → Direction). Done when still describes stack support and gets
  rewritten against the ban when the plan lands.
- 2026-09-29 — Retitled to "One issue per branch: no work on an unlanded
  base". Done when rewritten: operator chose soft agent instruction over
  gate-time base/multi-issue guards; only `start-plan` enforces the base. The
  quick-budget-to-HEAD defect is moot under one issue per branch.
