---
id: 000289
status: done
deps: [ariadne#288]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours: 2.25
card_mirror: 'd5a105ad87b1ddfbecdf000aa312319d8703f044' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T14:54:16-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.24
---

# Slot readiness in sdlc fleet inventory

## Problem

After a restart the operator needs to know, per local slot, whether it can take
new work, holds work to resume, or needs recovery. `sdlc fleet inventory
--json` already gathers most inputs for every worktree under the fleet root
(branch, detached, ahead/behind, `dirty_count` including untracked files, and
the issue its branch prefix names with that issue's status), but it reports
worktrees, not slots. It has no `pair:N` address, does not know that a slot's
product checkout and its private substrate checkouts (e.g.
`worktree/pair-slot1/pair` and `worktree/pair-slot1/ariadne`) form one unit, and
gives raw counts rather than a verdict. Couch (pair#363, pair#367) should
consume one per-slot verdict instead of re-deriving git state.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for
operator review; no implementation is authorized by this issue creation.

Extend the read-only fleet observation with slots: group each workspace's
checkouts (the product worktree plus the substrate dependencies weave declares
for it, not every sibling directory), attach the workspace address (`repo:N`,
`:0`) and resting branch, and compute one readiness verdict per slot over all
its checkouts:

- **ready** — every checkout clean (no modified or untracked files), no active
  Git operation, on its resting branch or on a branch already merged to main.
- **holds-work** — clean, but a checkout is on a branch with unlanded commits
  or one whose name points at a non-terminal issue (even with zero commits).
  Claim-awareness (a slot named on a claim while still on its resting branch)
  comes from ariadne#288's observation.
- **needs-recovery** — modified or untracked files, an active merge/rebase, or a
  detached HEAD; name each checkout and the paths behind the verdict.
- **missing** / **unknown** — a declared checkout is absent, or a probe failed.
  Unknown is never ready.

Read-only, versioned JSON, registered in the #280 recovery catalog. The
verdict is an observation: an action that reuses a slot (pair#363's reboot,
then `weave refresh`) must re-check it at action time.

## Done when

- Fleet inventory reports one row per slot with address, member checkouts and
  a verdict; tested with stateful fixtures covering ready, holds-work (unlanded
  commits; issue branch with zero commits), needs-recovery (modified,
  untracked, active rebase, detached) in the product checkout and in a substrate
  checkout, missing checkout, and a failed probe.
- needs-recovery rows give the reasons per checkout (dirty count, active
  operation, detached HEAD); paths are not listed (operator, 2026-10-02).
- Membership comes from weave's declared dependencies, tested against a slot
  with an undeclared sibling directory.
- The command's recovery entry states read-only, with its proof.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module        design=0.1  impl=0.16
item: greenfield-go-module     design=0.15 impl=0.24
item: greenfield-go-module     design=0.2  impl=0.28
item: smaller-go-module        design=0.1  impl=0.2
item: atlas-docs               design=0.05 impl=0.05
item: milestone-review         design=0.0  impl=0.14
item: milestone-review         design=0.0  impl=0.14
item: scope-pivot              design=0.3  impl=0.0
design-buffer: 0.15
total: 2.25
```

M1: shared operation detector + facts (smaller), `JudgeCheckout` (greenfield,
settled by the plan). M2: membership + slot discovery + assembly + contract
(greenfield), dependency-clone rows + tracker aliasing + real-git fixtures
(smaller). The scope pivot (reasons not paths; the #290 detour's redesign) is
design already spent after the claim.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000289-slot-readiness-in-sdlc-fleet-inventory-plan.md`.

- [x] M1 — per-checkout readiness facts and verdict (shared file-based `pkg/workspace` operation detector, operation on facts, pure `JudgeCheckout`)
- [x] M2 — fleet inventory reports one readiness row per slot (slots from paths, declared membership via `layergraph.ParseRows`, dependency clones as rows, `AssembleSlots`, versioned `slots`, real-git fixtures, recovery entry, help, atlas)

## Log

### 2026-10-02
- 2026-10-02: closed — Done-when met. (1) fleet inventory --json has slots[]: address, member checkouts, verdict per slot — TestFleetInventorySlotReadiness on a real fleet covers ready, holds-work (unlanded commits; zero-commit open-issue branch; claim), needs-recovery (modified host, untracked clone, MERGE_HEAD operation, detached clone), missing clone, and an unreadable/malformed declaration (unknown); TestBrokenDependencyCloneIsReported (unknown + diagnostic). (2) needs-recovery members give reasons (dirty, operation:<marker>, detached), no paths (operator revision). (3) membership from construct/deps (layergraph.ParseRows, weave placement rule shared) — an undeclared dirty sibling does not change the verdict; symlink-elsewhere member is outside. (4) sdlc help recovery: fleet inventory [read-only] with TestFleetInventorySlotReadiness/TestJudgeCheckout proofs. make test green (processgroup sandbox-only); weave + pkg green. Live: 39 slots in 6.3-6.8s (#290 baseline 5.9s), clones add no tracker read.; review verdict: FIX-THEN-SHIP
- 2026-10-02: closed M2 — make test green (processgroup sandbox-only); weave + pkg green. BR-8: reader (layergraph.ReadDeclaration) and placement rule (workspace.ValidSlotDependency, weave Policy.validate calls it) shared; DeclaredMembers judges declared+canonical paths (symlink-elsewhere case tested -> outside); walk kept separate from acquire.Restore by design (recorded in plan). TestFleetInventorySlotReadiness all match; live 39 slots in 6.8s.; review verdict: FIX-THEN-SHIP
- 2026-10-02: closed M1 — make test green (904 cmd/sdlc tests; processgroup fails only in sandbox); go test ./cmd/weave/... ./pkg/... green. BR-4: atlas + issue Log name workspace.OperationMarkers (pkg/workspace) and list weave refresh; no gitx.* detector name remains (grep). Ambiguous match now ErrAmbiguousIssue -> TreeRow.IssuesError -> probe:issue (association table + property test). Rebase e2e test moved to pkg/workspace with plain git.; review verdict: SHIP

Filed from pair#384's design review at the operator's direction (pair#384 closed
wontfix; its slot facts belong to sdlc). To be picked up after ariadne#288.
Consumers: pair#363 (bulk resume/reboot), pair#367 (slot view and recovery).

Details were left untracked in pair-slot1/ariadne by the filing session;
published to main with `sdlc issue move-detail` from that checkout at the
operator's direction, then claimed. Design findings: membership is declared in
`construct/deps` (`pkg/layergraph.ParseRows`, transitive); weave records no
clone set, so membership is re-derived; dependency clones rest on `main`;
operation-marker detection exists twice with different lists (to be unified).

M1 implemented: `workspace.OperationMarkers`/`ActiveOperation`/`WorktreeGitDir`
(`pkg/workspace`, file reads; `move-detail`, `move`, landing and weave's
refresh switched to it via one `rev-parse --absolute-git-dir`), `MeasuredFacts.Operation`/`OperationError`
(no added git process, asserted by call count), pure `JudgeCheckout` + verdict
order (`TestJudgeCheckout`, 18 cases). Precedence refined: needs-recovery wins
whenever its facts were read.

M2 implemented: slots from row paths (`workspace.RestingBranch` exported),
`DeclaredMembers` over `layergraph.ParseRows`, dependency clones as rows with
tracker reads aliased to the fleet primary by publication identity,
`AssembleSlots` worst-of fold, versioned contract (`schema_version` 1,
`slots`), render, catalog proof, help, atlas. `TestFleetInventorySlotReadiness`
(real fleet: 11 numbered slots, one cause each in host or clone, plus `:0`)
all match. Live on this machine: 39 slots in 6.3s (#290 baseline 5.9s), e.g.
pair:1 holds-work (claims #363/#367), parli:1 and tools:1 missing their
declared ariadne clone. Found while measuring: origin spellings differ
(SSH vs HTTPS), so origins compare as identities.

## Revisions

- 2026-10-02 — operator: drop path listing from needs-recovery (reasons only).
  Dependency clones join inventory as rows (they are not rows today). Paused
  until the inventory slowdown #288 introduced is fixed in its own issue.
