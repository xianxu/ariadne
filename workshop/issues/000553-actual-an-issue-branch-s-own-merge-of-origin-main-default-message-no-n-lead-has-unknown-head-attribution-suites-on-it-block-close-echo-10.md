---
id: 000553
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9e5e7bd525d6fe1838e7e2dfb31266fea7ae7e9e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T06:34:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "99b94420", done: "20a91033"}
actual_hours: 0.82
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
- Structural fallback applies to every otherwise unknown ordinary merge, including bookkeeping subjects
  and invalid issue leads; definite issue authority and malformed-train refusal retain precedence.
- Proof selection and classification share the same repository context, including qualified invalid
  refs and zero-padded zero refs in every lead-list position across repository names.
- echo#10's live 'sdlc actual --issue 10' measures (owner-run proof).

## Plan

- [x] Reproduce the ordinary integrate-main merge with real Git, retaining a foreign-parent control before changing production code.
- [x] Extend the existing immutable head reader with command-scoped pinned-main ancestry evidence; let the pure classifier/reducer inherit only first-parent membership for a structurally proven integration.
- [x] Verify repeated integrations, arbitrary messages, multiple side parents, unknown first parents, foreign parents, ref movement, replacement/graft controls, cancellation and bounded shared probes; retain the train grammar and existing budget controls.
- [x] Update the attribution atlas, obtain the owner-run echo#10 measurement, then record the pushed-head suite and close review before landing.

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
- 2026-10-22: closed — BR-2 root correction passes the resolver selfRepo through the typed reader call on every read; proof selection and membership classification use that same value, with no second identity derivation/default. New real-Git matrix RED all45 qualified-invalid-ref cases then GREEN across3 repositories/3 zero spellings/5 lead positions; definite local/foreign authority and foreign-parent controls retained. Full targeted queue-head/parser/reducer/owned-actual/Git-object checks PASS13.485s/0.495s. BR-1 prior red/green and recorded Echo fixture retained. Owner live actual10 passed twice on earlier clean source generations, exit0/1.07h/3 waits, receipts87d26f3f-2805-4d32-9414-a265c7361528 and c976681a-a5c8-45a9-b629-d6e4365b4daa; context correction separately covered by native matrix. Latest pushed fe9a4c9e4 suite is queued, not passing. Only unchecked Plan row is acceptance sequence; no-plan-check acknowledges tl-worker-authorized review overlap without concealing that pending obligation. Landing waits for recorded final-code suite, completed row, and canonical-checkout import after506. Atlas/Done-when/revisions/lessons updated; no format/new CLI.; review verdict: SHIP

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
- Fixture-final suite is queued on pushed 2dd67c231 as lease-1759503955, not yet a pass. The cancelled
  old lease retired through supported controls with no passing or failed attempt publication.
  Hub revoked #553's grant and approved finishing suite/close in this clone, then importing the pushed
  branch into the canonical checkout after #506 lands (receipt 1e45c411-1ba8-4ffc-b17d-ce7fb0d5b577).
  Start close review during that queue wait under tl-worker PERMISSIONS: the final code still needs its
  recorded suite pass before landing, and any review code change requires a new tested head.
- Close round 1 returned REWORK with BR-1 Critical, family message-independent-structural-attribution:
  bookkeeping-message early returns skipped proven structural integration. Independent real-Git RED
  cases confirmed both bookkeeping subjects; the complete early-return sweep also found invalid #0
  lead authority. Single-parent bookkeeping, definite issue leads, foreign parents and trains remain controls.
  Cancelled queued lease-1759503955 through its exact supported run ID before changing production code.
- BR-1 targeted assertions passed, but the command's hermeticity guard correctly refused my concurrent
  edit of workshop/lessons.md during its test window. This is an agent sequencing error, not a product
  failure or flake; rerun the same checks with the source checkout unchanged for their whole lifetime.
- BR-1 GREEN: the unchanged-checkout rerun passed cmd/sdlc in 21.575s and internal/gitx in 0.822s.
  The pure wrapper now applies structural fallback to every unknown authority result; recognized authority
  and malformed trains stay ahead of it. Real-Git cases cover both bookkeeping merges, invalid issue lead,
  their foreign-parent controls, retained single-parent bookkeeping, and definite other-issue authority.
- Close round 2 addressed BR-1 but returned REWORK for BR-2 Critical in the same family: the reader
  selected proofs with an empty self-repository name, while the reducer classified with the actual name.
  Qualified invalid local refs therefore received contradictory authority classifications. Cancelled only
  queued lease-801177825, stopped code changes, and replanned the seam with the hub.
- BR-2 RED: the real-Git context matrix failed for all 45 qualified invalid-ref cases (three repository
  names, three zero spellings, five list positions). Definite local/foreign authority and foreign-parent
  controls stayed strict. The reader now receives selfRepo from the resolver's typed call on every read;
  the proof-selection parse and membership parse consume that identical argument, with no IO-shell default.
- BR-2 GREEN: targeted queue-head, parser/reducer, owned-actual consumer and Git-object checks passed
  cmd/sdlc in 13.485s and internal/gitx in 0.495s. The source sweep found one repository derivation at
  the actual resolver call; both production parser calls now receive its unchanged selfRepo parameter.
  No second context default or constructor derivation remains. The corrected 954dea6 private binary's
  earlier owner rerun also exited 0 with the same 1.07h/three-wait result (receipt c976681a-a5c8-45a9-b629-d6e4365b4daa);
  the new context correction is independently covered by the 45-case real-Git matrix.
- Close round 3 returned SHIP/high, BR-2 addressed, zero new/open findings; evidence
  fe83377b1a4c4204680ae5b170331012c443a29d, measured actual 0.82h. The independent reviewer passed
  the targeted suite and restored the old empty reader context through an overlay: all 45 context
  regressions failed. No additional production correction was required.
- Current recorded suite remains queued on fe9a4c9e4baff739a319d5969a7adeb8ca683212 as lease-2843917753
  (terminal18446). Earlier replacement leases retired without execution through their exact supported
  cancel controls. #506's retained runner remains in the canonical checkout (terminal18055); after it
  lands and this suite passes, import this pushed branch there and ask the hub to restore landing priority.
  Do not merge through this clone's separate line. The acceptance checkbox stays open until the suite result.
- Final owner-run live proof on the accepted clean binary 4948bbf56ac908ffb7670b8def517a41afff4901
  (modified=false) passed again, bound by receipt 7d02fd88-21c1-4a6f-8cbd-5bdf92a245bc. Echo#10 head
  83daa1f7d86fa5eb8d9b574bfa83e3b027a2798a measured exit 0, actual 1.07h, with three recorded waits
  (78.50 raw/70.86 excluded minutes), identically to both earlier runs. This binary's code is exactly
  reviewed fe9a4c9e; Git confirmed only issue/review records changed afterward. Keep that binary available
  for its consumer until the canonical landing and fleet rebuild, then remove the private clone.

- #506 landed as PR390; its retained runner exited successfully and returned the canonical checkout
  cleanly to unchanged main-slot5. Imported this pushed branch into canonical ariadne:5 at
  bd18639007961226a81456be93c15b32c330ce05 and opened PR397. The branch suite remains in its original
  clone, with no cancellation or duplicate runner. Grants and landing lines are per-checkout: the
  clone cannot see the canonical grant. The hub applied #553's grant separately in both checkouts
  (receipts 51de1d34-a1e0-4b31-975e-9b6b9adb3756 and a8af588d-9401-49b2-a26c-a9b6743fb66c), moving
  the retained suite to next-to-run. This issue still lands only from the canonical checkout.
- The independent clone's active suite also has an empty workspace identity in queue records:
  op-f2ed5b568305275eb3782c1a51006ef3 appears as `unknown ()`, making queue status report
  `complete=false`. The hub identified this display limitation (receipt
  65d87200-2e03-4876-b7f0-fc40dc639689); the retained terminal/lease still identify the live owned run.
  Preserve that run rather than interpreting the incomplete queue display as a dead runner.

- The retained branch suite on fe9a4c9e4baff739a319d5969a7adeb8ca683212 completed PASS:
  all eight shards, 2242 SDLC tests and 76 other packages, zero failures, wall 832.2s (13m52s).
  Log SHA256 prefix bd2dc5987527, retained at the independent clone's
  .git/sdlc/local-checks/fe9a4c9e4baff739a319d5969a7adeb8ca683212/suite.log.
  The lease and terminal retired normally; no retry, flake allowance or status promotion was needed.
  Subsequent canonical commits change only issue/review records, leaving the reviewed executable code
  unchanged. Acceptance is now complete; proceed through canonical sdlc merge for current-main integration.

## Revisions

### 2026-10-22 — structural integration plan

- Retain the original acceptance contract and add the implementation/proof plan above. The hub requested
  priority work while #506 waits; its recorded runner remains independent and no queue ordering is changed here.

### 2026-10-22 — BR-1: message-independent fallback

- Put the structural fallback after the existing pure authority parser's result, rather than after only
  some message branches. Every unknown ordinary merge can reach it; definite authority and malformed
  train refusal still win. Extend the real-Git matrix across bookkeeping subjects and invalid issue leads,
  preserving single-parent bookkeeping and foreign-parent/definite-authority controls (ARCH-PURPOSE).

### 2026-10-22 — BR-2: one authority context through the reader

- Hub approved (receipt fe3a1a38-cd7f-4aa4-9687-86636c52a7ae): thread selfRepo from the existing classifier
  parameter into the typed head-reader call on every read, rather than deriving/defaulting repository
  identity again in the IO shell. Both authority parses consume that one value (ARCH-DRY/PURPOSE).
  Cover qualified zero and zero-padded zero leads alone and in every lead-list position, across repository
  names, plus definite local/foreign authority and foreign-parent controls. Existing train precedence stays.
