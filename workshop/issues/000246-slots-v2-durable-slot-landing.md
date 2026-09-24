---
id: 000246
status: working
deps: [ariadne#242, ariadne#243, ariadne#244, ariadne#245]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours: 3.83
started: 2026-09-23T15:24:19-07:00
flow: {kind: full, provenance: operator}
---

# Slots v2: land while retaining the workspace

## Problem

Current SDLC merge cleanup removes feature worktrees and returns in-place work to updated main. A durable slot must survive publication and return to its unchanged resting branch.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Adapt existing publication/review/merge flow for durable slots. Confirm integration into the configured remote main, safely return to the existing main-slotN, then safely remove the completed local issue branch. Do not refresh the resting branch. Preserve directory, unpublished commits, dirty files, and other workspaces. Do not operate on the primary checkout merely to complete a slot’s cleanup or archive its artifacts.

Account for squash/rebase merge evidence when deciding that branch deletion is safe; an upstream relationship or branch name is insufficient proof. Detect commits added after the integrated head. Define recoverable phases for integration completed but switch/deletion/archive cleanup interrupted. Re-running must recognize already-confirmed integration and finish safely without duplicating publication. Keep ordinary feature-worktree and :0 behavior explicit and tested. ARCH-FUNERAL: retain reusable workspaces intentionally, while completed issue branches have an evidence-backed removal path.

### Agreed scope — 2026-09-23

This section takes precedence over earlier conflicting layout or policy text.

The durable main worktree is `/workspace/worktree/<repo>-slotN/<repo>`; its parent contains ordinary dependency clones. Landing the main repository must preserve that enclosing directory and every dependency checkout, including dirty files and unpublished commits. Dependency repositories continue to use their normal per-repository SDLC publication/cleanup behavior; do not classify them as numbered worktrees or recursively delete them. Coordinated work is driven from the parent thread and ships each repository separately in dependency order when required. This version adds no automatic compare-and-merge of all dependencies and no multi-repository publication transaction.

### Proposed engineering design — 2026-09-23

Extend existing `sdlc pr` / `sdlc merge`, with no new landing command or persistent receipt file. Use `workspace.Resolve` to distinguish durable primary/numbered slots from ordinary feature worktrees and private dependency checkouts. The project contract applies equally to :0 and :N: landing returns to existing `main` or `main-slotN` without advancing it. Ordinary feature-worktree removal and dependency publication retain their existing, explicitly tested behavior. Operator approved the design and dependency-first landing sequence on 2026-09-23.

Alternatives considered: only skipping worktree removal would still pull/archive through another checkout and could destroy newer branch work; a per-slot transaction journal would add a second durable authority to reconcile with Git and GitHub. Prefer authoritative Git/PR evidence plus an explicit retry branch (ARCH-DRY, ARCH-FUNERAL).

**One target.** Durable `pr` and `merge` derive the destination from the resting branch's single configured named remote tracking `refs/heads/main`, as in #245. Resolve that remote's GitHub repository explicitly; do not infer origin or another checkout's branch. Missing/ambiguous/non-main configuration refuses with guidance. `pr` pushes the issue branch to that target and creates its main PR, with the normal issue links and gates. `merge` never uses the primary checkout to pull, archive or clean up. Same-repository PRs are the initial supported path; mismatched/fork PR identity refuses rather than guessing.

**Integration evidence.** Extend the existing GitHub seam to return structured PR state, exact head/base repository identity, head branch and full head/base commit IDs, and integration commit. Query errors remain errors. Require a unique matching PR targeting the configured main. For an open PR, the local issue head, freshly observed remote issue head and PR head must match before the existing deterministic publish gates permit integration. Merge server-side with an expected-head check; do not use gh's `--delete-branch`, which may perform local checkout/deletion itself. After every submission, re-query the PR and confirm MERGED plus reachable integration; successful queue admission or command exit is not integration. Preserve the current merge-commit strategy for newly submitted merges; recognize already-completed merge, squash and rebase integration through exact PR-head evidence plus its integration commit being reachable from freshly fetched remote main. Unknown/missing evidence blocks cleanup; ancestry of the original issue commits or a matching branch name alone is not proof.

**Scoped remote archive.** After confirmed integration, use existing `gitx.TrunkFile.UpdateMany` to publish the done transition and issue/plan/review moves as one conditional remote-main commit. Derive candidate issues from close evidence in the PR's own commit set (`head` minus commits reachable from its recorded base), selecting records that are codecomplete at the immutable PR head and whose close anchor belongs to that commit set. Do not use a raw base-tree/head-tree diff or sweep remote-main codecomplete files: neither unrelated base changes nor independently published issue-body copies should change ownership of this archive. Parameterize the existing close-anchor/issue readers by commit rather than applying current-worktree reads to a resting checkout. Read current remote bodies on every attempt, preserve compatible concurrent documentation updates, and refuse reopened/identity-conflicting records or occupied archive destinations. Reuse/extract existing issue lifecycle and archive naming/plan-selection rules rather than inventing another archive format. The archive commit records the exact GitHub repository/PR number and integrated head in provenance trailers; its tree diff records the complete issue/plan/review move set. Retry confirms that commit is reachable from fresh remote main, that its complete expected moves and done transitions match this PR, and that the active sources remain absent and archived generation remains intact. A same-name/status destination without this evidence, partial moves, later conflicting archive edits or a reopened record refuses rather than pretending cleanup completed. Matching confirmed archives are a no-op; ambiguous push acknowledgment stops local cleanup and retry re-observes the remote. No local feature, resting or primary ref/index/files are modified by archive publication. Existing close-time project updates remain where they belong; no recursive peer publication is added.

**Return and remove.** Only after integration and archive are confirmed, switch the current durable checkout to its existing resting branch with ignored-file overwrite and submodule recursion disabled. Require no ongoing Git operation; tracked dirt refuses before initial integration, including tracker files (do not inherit the legacy tracker-dirt exception here). Preserve noncolliding untracked/ignored files, and refuse collisions without stash/reset/auto-commit. Revalidate branch, HEAD, resting ref/upstream and topology around mutation. Once on rest, require the completed issue ref still equals the PR's integrated head and is not checked out in any other worktree; remove it with expected-SHA comparison. New commits, another checkout using the ref, uncertain evidence or dirty collisions leave work intact and print the next action. Keep the containing environment and all sibling clones untouched. Do not broaden branch removal to remote branch deletion in this change; GitHub repository auto-deletion remains independent.

**Retry without a journal.** Add optional `sdlc merge --branch <issue-branch> --yes` to resume while already on the resting branch. Print that recovery command before the first irreversible effect. Without the flag, merge continues to use the current issue branch. An explicit branch may select only the current issue branch or a completed branch while this checkout is at rest, never another active workspace's work; `main` and reserved `main-slotN` resting refs cannot be selected as issue branches. Re-observe GitHub integration and the selected issue ref; do not run fresh issue-head gates against resting HEAD. Open-PR merging still requires the issue checked out and normal gates. A merged PR skips merge submission and continues archive, switch, deletion. A missing local branch is complete only when a unique merged PR, reachable integration and scoped archive completion are confirmed; ambiguous reused names refuse. Failed probes and possibly successful network requests never become assumed failure/absence.

| Observed phase | Permitted next action | Failure / interruption |
|---|---|---|
| Open PR, matching checked-out head, gates passed | Expected-head server merge | Re-query PR before another submission |
| Merged PR, exact integrated head, reachable integration | Scoped conditional archive | Preserve branch and checkout; re-observe archive on retry |
| Matching archive confirmed | Safe switch to unchanged rest | Preserve files/ref; retry from issue branch |
| At rest, completed ref still at integrated head | Occupancy check and expected-SHA ref deletion | Preserve ref on changes; retry with `--branch` |
| At rest, ref absent, integration/archive confirmed | Report complete | No publication or checkout mutation |

The phase decisions are pure; Git/GitHub/archive effects stay behind existing IO seams. One invocation owns its effects, uses the existing common-directory SDLC lock, and starts no background worker. Git/editor processes outside that lock can still race: rechecks, Git collision checks and expected-SHA writes provide bounded protection, not an atomic lock over arbitrary external actors (ARCH-PURE, ARCH-STATE).

**Verification and boundaries.** Reuse local Git fixtures and add a stateful GitHub fake backed by a bare remote, exercising actual integration as merge/squash/rebase. Inject failures after merge, archive publication, return-to-rest and ref deletion; retries must not duplicate integration/archive. Assert byte/ref/config preservation for :0, a dirty primary, another active slot and independent dependency siblings. Test deleted remote issue branch, post-integration local commits, wrong/ambiguous PR target, queued-but-unmerged responses, query failure, uncertain publication, tracked/untracked/ignored collisions and occupancy races. Archive fixtures include independently published issue copies, unrelated codecomplete base changes, absent/wrong provenance, partial/conflicting moves and exact complete retry. Check configured non-origin remotes in both pr and merge. Keep legacy ordinary-worktree and dependency publication regressions, including dependency-first then parent landing. README, embedded help, atlas and project records describe the changed :0 behavior and explicit recovery.

Operating envelope: one selected repository/PR per invocation, serial bounded observations and existing bounded trunk-publication retries; no fleet traversal beyond local Git worktree occupancy, no recursive dependency operations and no persistent new files. Bound structured GitHub reads and reject incomplete/truncated evidence instead of broadening queries indefinitely. Trust local Git identity plus the configured GitHub target, validate returned identities/SHAs and artifact paths, and never log credentials (ARCH-CONSTRAINTS, ARCH-SECURE, ARCH-MOCK). Completed issue branches are removed only with integration evidence; existing archival artifacts retain their established history lifecycle (ARCH-FUNERAL).

## Done when

- A slot completes review/publication and lands; its path survives and its resting SHA/upstream are unchanged.
- Primary and another dirty active slot retain their branch, files, and refs throughout landing.
- Unpublished additions, dirty switch collisions, and uncertain integration evidence block destructive cleanup and explain recovery.
- Integration followed by interrupted cleanup can be retried safely, including the repository’s actual merge strategy.
- Issue archiving/project updates complete through existing gates without requiring a clean idle primary; ordinary worktree behavior has regression coverage.

- Landing and interrupted-cleanup retry preserve the enclosing environment and dependency clones, including unpublished issue records and code changes.
- A dependency change can be published first through its own normal SDLC flow, then the parent feature can land without recursively publishing or cleaning other repositories.

## Plan

Implementation follows [the durable plan](../plans/000246-slots-v2-durable-slot-landing-plan.md), one atomic delivery and one close review.

- [ ] Add structured exact GitHub integration evidence and expected-head merge.
- [ ] Add scoped remote archive with authoritative retry proof.
- [ ] Integrate durable PR/merge routing, safe return/deletion and interruption tests.
- [ ] Update guidance and complete acceptance verification.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

### 2026-09-23 — Landing audit and proposal

Claimed #246 and ran start-plan. Current merge invokes gh with local branch deletion, assumes origin, pulls/archives in primary and removes all linked worktrees. Read-only audit recommends early identity-based routing, structured exact PR evidence, remote archive through existing TrunkFile, and explicit branch retry instead of a receipt file. Live read-only PR130 evidence retains its deleted head SHA and pre-merge base SHA plus integration commit. No implementation started; design review/approval and full-flow durable planning precede estimates and code.

### 2026-09-23 — Spec review corrections

Fresh review requested precise archive retry proof and PR-owned issue selection. Clarified that close anchors in the PR's own commit set select artifacts even when issue bodies were independently published, and that reachable archive commit provenance plus the complete artifact generation authorizes retry completion. Queue acceptance must be re-observed as merged integration; reserved resting refs cannot be cleanup targets. No receipt file or new landing command added. Fresh re-review approved the revised spec with no remaining important gaps; operator approval is pending before durable implementation planning and code.

### 2026-09-23 — Implementation entry and integration findings

Plan-quality round 2 accepted; estimate-quality accepted 3.83h and change-code created the issue branch. GH adapter isolated tests, parser fuzz and read-only PR130 conformance passed. Root integration fixtures observed missing archive behavior before implementation. Ad-hoc review found legacy origin-bound duplicate checking and body-diff publish enumeration; both are being routed through shared rules with pinned configured-main/PR-owned evidence. Effective URLs and selected PR branch/HEAD are revalidated. No landing or merge has been performed for #246.

## Revisions

### 2026-09-23 — Retain nested environments and publish repositories separately

Reason: operator agreed nested environments, ordinary remote dependency clones and existing per-repository publication. Delta: added the authoritative scope clarification and acceptance criteria above; original task context remains as provenance. Added #243 as a prerequisite for the nested identity contract. No implementation or lifecycle-status change is claimed by this revision.

### 2026-09-23 — Proposed durable landing and recovery

Reason: trace the existing merge/PR/archive behavior against the agreed stable-workspace contract. Delta: propose equal :0/:N resting-baseline preservation, configured-remote PR/merge targeting, scoped remote archiving and an explicit branch retry flag without a journal. Preserve legacy ordinary/dependency cleanup paths. Pending operator approval.

### 2026-09-23 — Implementation authorized

Operator approved the design, including Ariadne dependency first and Pair parent second. Delta: add the durable implementation plan and concrete task checklist; keep ordinary dependency flow and no recursive landing. Full-flow estimate follows plan-quality acceptance.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Calibration is tagged stale; the result is provisional.

Issue/spec work uses 0.8 design and 0.2 × 0.4 implementation hours. GH integration uses 2 × 0.2 design and 1.5 × 0.4 implementation. Remote archive uses greenfield design 2 × 0.5 (existing TrunkFile/issue/vocabulary libraries) × 0.2 (settled plan) and implementation 0.8 × 0.4. Landing orchestration uses 1 × 0.2 design and 0.8 × 0.4 implementation; no outside library replaces its repository policy. The remaining items cover integration fixtures (0.2 × 0.2 design, 0.5 × 0.4 implementation), docs (0.1 × 0.2, 0.2 × 0.4), one review (0, 0.5 × 0.4), and read-only API conformance (0, 0.3 × 0.4). Familiar Git/Go code uses 1.0 familiarity and the thorough plan earns the 15% design buffer. No vendor propagation overhead; Pair only receives the project record.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.8 impl=0.08
item: api-integration design=0.4 impl=0.6
item: greenfield-go-module design=0.2 impl=0.32
item: greenfield-go-module design=0.2 impl=0.32
item: smaller-go-module design=0.04 impl=0.2
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0 impl=0.2
item: real-api-discovery design=0 impl=0.12
design-buffer: 0.15
total: 3.83
```
