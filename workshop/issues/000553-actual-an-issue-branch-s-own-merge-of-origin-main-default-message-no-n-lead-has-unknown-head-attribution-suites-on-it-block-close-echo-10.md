---
id: 000553
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '707a528314e9042450920b8a46d84d144c7c95ae' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T06:34:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "99b94420", done: "20a91033"}
---

# actual: an issue branch's own merge of origin/main (default message, no #N lead) has unknown head attribution; suites on it block close (echo#10)

## Problem

## Spec

From ops-echo:0 (10-22 ~06:4x): echo#10's close refuses at measurement on both the fleet sdlc (with #548) and the
#551 binary 26b4be0, before any review:

    Measured actual unavailable for issue close: can't auto-compute (queue evidence incomplete; ...: suite attempt
    dffca4b701d5264e244d6f019055b21d SHA 885dc171dcfde7aaaae15d79f2b7c3863f207331 has unknown issue-head
    attribution: commit has no definite issue lead or recognized neutral authority; ...)

885dc171 is echo:3's ordinary integrate-main merge on the echo#10 branch: git's default message "Merge
remote-tracking branch 'origin/main' into 000010-...", with no `#10:` lead, and its recorded suite is green. Before that
suite ran, the same close measured 1.63h. #548 ("suite evidence must belong to the measured issue head") follows
first-parent neutral integration only when it's recognized, and a plain `git merge origin/main` on the issue branch
isn't. Every landing integrates main, so echo#3 and #5 (and any issue that merges main before its final suite) hit
this.

Fix (structural, no commit-message heuristics): a merge commit on the issue branch whose **first parent** is
attributable to the issue (issue-led, or itself attributable through this same rule) and whose other parents are all
reachable from the pinned origin/main is the issue's own integration, so it's attributed to the issue. Unknown or
foreign parents keep strict refusal.

## Done when

- echo#10's exact shape (issue commits, then a default-message merge of origin/main, then a suite on the merge)
  measures, with the suite attributed to the issue (real-git fixture).
- A merge bringing in a foreign issue branch (a second parent not on main) still refuses (test).
- echo#10's live 'sdlc actual --issue 10' measures (owner-run proof).

## Plan

- [ ] Reproduce the ordinary integrate-main merge with real Git, retaining a foreign-parent control before changing production code.
- [ ] Extend the existing immutable head reader with command-scoped pinned-main ancestry evidence; let the pure classifier/reducer inherit only first-parent membership for a structurally proven integration.
- [ ] Verify repeated integrations, arbitrary messages, multiple side parents, unknown first parents, foreign parents, ref movement, replacement/graft controls, cancellation and bounded shared probes; retain the train grammar and existing budget controls.
- [ ] Update the attribution atlas, obtain the owner-run echo#10 measurement, then record the pushed-head suite and close review before landing.

### Design

Reuse #548's raw-object reader and pure parse/reduce seam (ARCH-DRY/PURE).
For an otherwise unattributed merge, read its immutable first-parent membership;
inherit it only if every remaining parent is reachable from one pinned publication-main OID.
The Git shell supplies the ancestry proof, not the subject or current branch name.
Main contributes no issue members. Explicit train parsing retains precedence,
so malformed train authority cannot be rescued by the new path (ARCH-PURPOSE/ORDER).

Pass the publication remote already resolved by actual's caller into the reader;
resolve its main ref once per pass, before ancestry checks, and use only that OID afterward.
Cache each distinct parent/main proof within that reader, under the existing node ceiling;
retain the raw reader's node/byte/depth/batch bounds and bounded cancellable Git process seam.
Disable replace refs and graft/shallow overrides for structural ancestry; missing objects,
invalid main pins, failed probes and exhausted bounds remain unavailable evidence (ARCH-SECURE/CONSTRAINTS).

The existing pure classifier/reducer gets table controls; the IO reader gets real private Git
graphs and a counting Git boundary for shared-probe/ref-movement controls (ARCH-MOCK).
No queue/status records or persisted schemas change (ARCH-EVOLVE); command-local proof maps
die with the pass and private fixtures/clones are removed after their last user (ARCH-FUNERAL).
This is a single atomic change; if implementation leaves the quick-flow shell, use its full gate.

## Log

### 2026-10-22

- Claimed and started in ariadne:5. #506 retains its managed merge runner in the original checkout;
  #553 uses the independent ordinary clone at ariadne-slot5/ariadne-553, based on main e4bd907ae47d.
  Read #548 commit ebf48d5e4 and its immutable parser, worklist/reducer and queue-evidence consumer.
  Root cause: the ordinary Git merge has no issue lead and fails the recognized neutral-message grammar,
  so traversal stops before its attributable first parent. No production edits made during diagnosis.
- Planning alternative: adding Git's default merge subject to the grammar is smaller but repeats the
  message-authority bug; walking and unioning all parents incorrectly imports main's issue membership.
  Selected structural side-parent reachability plus first-parent inheritance, matching the dispatch.
- Hub approved the committed plan (receipt b9d91ac8-ff32-4fab-ae53-f490f2a84f56).
  #548's author confirmed train precedence and the absence of an existing bounded ancestry helper
  (receipt 74ed2f43-8cfa-461c-83c2-8f6e57ef22f4). change-code accepted the quick flow.
- RED: TestQueueHeadOrdinaryMainIntegrationRealGit failed with unknown issue-head attribution on Git's
  actual default merge message; the expanded matrix also failed arbitrary, nested, octopus, older-main
  and next-pass-pin cases. Foreign, unknown-parent and malformed-train controls remained strict.
  GREEN: the structural reader/classifier inherited only first-parent members.
- The native counting boundary separately caught missing lazy-fetch suppression in raw commit reads
  (exit 73), then passed with it enabled. Targeted cmd/sdlc and internal/gitx checks passed (11.493s/0.560s):
  real merge controls, upstream remote selection, one pin and one shared-parent probe across repeated
  batches, cache ceiling, cancellation, replacement/graft/shallow controls, missing main/object failures,
  existing train/budget controls, and the recorded suite-to-owned-actual consumer. Diff check clean.
- Owner-run live proof (receipt 87d26f3f-2805-4d32-9414-a265c7361528): clean binary 43799bb36a7f
  measured echo#10 on pushed head 83daa1f7d86f with exit 0, actual 1.07h, provenance
  active-time-v3-engagement-queue-v1; three recorded waits contributed 78.50 raw queue minutes and
  70.86 excluded activity minutes. The owner retained full output locally. Exact merge metadata,
  default message with retained conflict comments, and verbatim suite attempt JSON now live in
  cmd/sdlc/testdata/actualqueue/echo-10-default-main-merge.json and feed the pure classifier/filter regression.
  The first queued suite lease-484717765 was cancelled through its exact owner-scoped run ID before
  execution so the executable recorded fixture joins the suite's final tested head. No production delta.
- Recorded-fixture mutation check: a temporary Go overlay disabling only the structural fallback reproduced
  the exact dffca4b7/885dc171 unknown-attribution error; the tracked implementation passed the fixture.
  The independent clone has a separate landing line, so its eventual landing must be coordinated through
  the canonical checkout rather than bypassing that queue. Reported this to the hub before merge.

## Revisions

### 2026-10-22 — structural integration plan

- Retain the original acceptance contract and add the implementation/proof plan above. The hub requested
  priority work while #506 waits; its recorded runner remains independent and no queue ordering is changed here.
