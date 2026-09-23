---
id: 000245
status: working
deps: [ariadne#242, ariadne#243]
github_issue:
created: 2026-09-22
updated: 2026-09-23
estimate_hours:
started: 2026-09-23T14:55:42-07:00
flow: {kind: quick, provenance: inferred, spec: "06f4ce42", done: "8d9b76a4"}
---

# Slots v2: branch from a workspace and refresh

## Problem

Agents need precise instructions for branching from :0/:1/:2 and refreshing a resting branch without implicit synchronization or hidden work loss.

## Spec

Project: `pair/workshop/projects/couch-slots-v2.md`. Fresh task derived from the current v2 contract; historical task bodies are not prerequisites or implementation plans.

Support the instruction “In :2, create an issue branch from :1’s current commit.” Resolve addresses within the repo, require a clean locally committed source and a destination ready to switch, pin the source commit, and record source address plus SHA in the issue workflow evidence. Branch the whole snapshot; do not infer issue-specific files or transfer claims. Neither resting branch nor its upstream moves. Subsequent source changes do not propagate.

Starting independent work on a slot branches from its current baseline without fetching or refreshing implicitly. Refresh is separately requested: fetch the configured remote main and fast-forward the resting branch when safe. Local planning commits/divergence require a separate explicit reconciliation choice, never an automatic reset/stash/commit. Refuse refresh of an active issue workspace unless the chosen explicit procedure first safely returns to rest. Define untracked/ignored-file readiness and source-change-during-capture behavior in the design.

Audit change-code and existing Git/procedure surfaces; prefer documented Git operations where adequate. A new adopt command or duplicate Git framework is not required. ARCH-DRY: one address resolver and existing branch/gate mechanisms.

### Agreed scope — 2026-09-23

This section takes precedence over earlier conflicting layout or policy text.

Resolve numbered main checkouts at `/workspace/worktree/<repo>-slotN/<repo>` through the shared contract updated in #243. Branch-from captures the selected repository's committed snapshot, not a multi-repository snapshot of its enclosing environment. Refresh changes only the explicitly targeted repository/resting branch; neither operation clones, switches, refreshes or resets sibling dependency repositories. Dependency revisions are controlled explicitly by operator/agent using ordinary Git. Cross-repository development is available from every slot and does not grant :0 special branch privileges.

### Proposed engineering design — 2026-09-23

Use documented ordinary Git procedures, resolved through `sdlc workspace ADDRESS --json`, rather than adding a branch/adopt/refresh command. The procedures prepare the checkout; existing `claim`, `start-plan`, `change-code` and review gates continue to own the issue lifecycle. Operator approved this design on 2026-09-23, with the direction to keep it simple.

Alternatives considered: a `change-code --from` flag would couple source selection to planning review and require moving its present sync-before-branch sequence; separate workspace mutation commands would duplicate Git orchestration. Neither is needed for the agreed agent-driven workflow (ARCH-DRY, ARCH-PURPOSE).

**Branch from a workspace.** From the destination, resolve its identity and the requested source address through the existing resolver. Both must be addressable slots of the same Git repository (`repo_identity` equal), including :0. Require destination on its resting branch. The target issue must already be allocated: derive the new branch name from its exact canonical issue-file stem, as change-code does. Reject an existing local branch instead of reusing it. If no issue is allocated yet, create it through the existing issue workflow first, then repeat preparation using its actual allocated filename; do not guess the next ID or create a provisional branch in this procedure. Detached/unborn/unresolvable workspaces, ongoing Git merge/rebase/cherry-pick/revert/bisect operations, and unsafe readiness refuse before switching. The command is preparation only; it neither acquires nor transfers an issue claim.

Capture the source's full commit SHA and symbolic branch. Require source and destination clean: no staged or unstaged tracked changes, dirty submodules, or nonignored untracked files. Ignored build outputs may remain, but must never be overwritten by switching. Re-resolve and repeat source HEAD/branch/readiness checks immediately before accepting the capture, along with destination identity/HEAD/resting/readiness and branch-name checks. An observed source movement during capture refuses and asks for a fresh capture. Once accepted, the full SHA is the immutable start point; subsequent source movement is harmless and does not propagate. These are bounded observations, not an atomic reservation against an unrelated Git process or editor; agents must stop concurrent writers in the affected checkout while preparing it.

Create the new branch in the destination using ordinary Git:

```sh
git -C "$destination" -c submodule.recurse=false switch --no-track --no-overwrite-ignore -c "$issue_branch" "$source_sha"
```

Use quoted, validated arguments and the full captured SHA, never a moving branch name. Do not use force/reset flags. Git is the final branch-name/collision guard. Confirm destination HEAD equals the captured SHA before any workflow edits. Source checkout, both resting refs, and their upstream configurations remain unchanged. The new issue branch deliberately has no inferred upstream.

After switching, append source canonical address and full SHA to the selected issue's `## Log` and checkpoint it through existing issue sync. The initial branch tip is exactly the captured SHA; provenance and later planning checkpoints are subsequent explicit commits on the issue branch. If the allocated issue does not exist in that snapshot, explicitly bring in that exact issue record on the prepared branch and verify its canonical stem still matches the branch before gates; never allocate a replacement issue, claim a reserved issue again, or import an arbitrary destination plan automatically. Claim an open issue through the existing remote reservation check before design; for an already-working issue, require established authorization to continue it rather than treating the branch as ownership. Run planning and `change-code` against the final issue/plan content on this branch. A failed provenance checkpoint leaves the prepared branch visible, reports the missing step, and does not roll back or claim implementation readiness.

**Independent issue start.** Apply the same preparation order using the destination's current committed resting HEAD as the start point. For an already allocated issue, branch before claim/design/checkpoint work when preserving the resting ref matters. Issue allocation is a separate prerequisite, not part of the branch operation. There is no implicit fetch, pull or refresh to select the baseline. Explicit claim and normal change-code documentation publication still contact remote main for their existing purposes; they do not update the selected resting baseline. Existing planning commits already on the resting branch remain part of that baseline; they are not silently removed. Ordinary pre-existing change-code behavior outside this documented slot procedure remains compatible.

**Explicit refresh.** Resolve only the requested slot and require it currently checked out on its identified resting branch, clean under the same readiness policy, with no ongoing Git operation. Determine the configured remote from that resting branch's upstream; require a named remote and `refs/heads/main` tracking configuration. Missing, local-dot, or non-main upstreams refuse with configuration guidance rather than guessing origin. Record the current resting SHA and upstream configuration, fetch remote main explicitly with submodule recursion disabled into `FETCH_HEAD`, and immediately capture its full SHA. Recheck workspace identity, resting branch/HEAD/configuration and readiness after fetch. Fetch failure or changed evidence stops before any local branch movement.

Require the old resting commit to be an ancestor of the fetched commit. Equal commits are a no-op; a behind branch can fast-forward. Ahead or divergent branches, including local planning commits, refuse and explain that the operator must explicitly choose how to preserve/reconcile them. Use `git merge --ff-only --no-autostash --no-overwrite-ignore` with the pinned fetched SHA and `submodule.recurse=false`; never reset, stash, auto-commit, infer a merge/rebase, or fall back to another remote. Fetch may update remote bookkeeping, but the resting ref moves only on a successful fast-forward. An active issue branch refuses; returning to rest is a separate explicit action after preserving its work.

**Isolation and evidence.** Both procedures operate on the selected repository only. They never traverse or update sibling dependency clones. Document :0 and :N examples, failure recovery, and the ordinary-workflow compatibility distinction in an atlas procedure linked from workspace identity and change-code help. Add real local Git fixtures which execute the documented Git operations and readiness sequence; test source movement at the capture boundary deterministically, not with timing sleeps. Reuse existing workspace fixtures/runner seams where suitable rather than introducing a second resolver or a metadata file. No slot-to-issue metadata is added: the issue branch and its issue Log carry the association and provenance.

**Acceptance coverage.** Fixtures cover :0 and numbered source/destination combinations; same-repository rejection; tracked, staged, nonignored-untracked, ignored-collision and dirty-submodule readiness; in-progress Git operations; an existing branch; source movement before/after accepted capture; exact initial SHA and recorded provenance; existing destination-only planning artifacts; missing allocated issue records and canonical branch-name binding (including refusal to guess or replace an issue ID); old baseline without refresh; configured non-origin remote; clean equal/behind/ahead/divergent resting branches; failed fetch; active issue refresh refusal; and nested sibling dependency SHAs, branches, dirty/untracked files and unpublished commits remaining unchanged. Verify existing change-code works on the prepared branch and its planning sync advances that branch, not the resting ref.

## Done when

- Branching :2 from clean :1 starts at the captured SHA, preserves source and both resting refs/upstreams, and records provenance.
- Dirty source, unsafe destination, branch-name conflict, and source movement during capture have tested safe outcomes; no auto-stash/commit occurs.
- Issue start leaves an intentionally old baseline unchanged; explicit refresh fast-forwards a clean behind resting branch.
- Dirty/divergent/planning-commit cases preserve all work and expose explicit recovery; configured remotes other than origin are covered.
- Agent guidance uses branch-from wording and :0/:N shorthand, with ordinary primary workflow regression coverage.

- Branch-from and explicit main-worktree refresh preserve sibling dependency checkout branches, SHAs and dirty/unpublished work in nested-environment fixtures.

## Plan

Small docs-and-tests change, inside the quick-flow shell; no separate plan or new runtime mechanism (ARCH-DRY).

- [x] Add focused real-Git regression fixtures in `cmd/sdlc/workspace_procedure_test.go` for the approved capture/branch/refresh commands, preservation/refusal cases, and existing change-code on a prepared issue branch. Run `go test ./cmd/sdlc -run 'TestWorkspaceProcedure' -count=1`.
- [x] Write one concise `atlas/workflow/workspace-branching.md` guide. Link it from `atlas/index.md`, workspace identity and `cmd/sdlc/helptext/change-code.md`; make embedded help carry the essential procedure so downstream agents can find it.
- [x] Run focused workspace/change-code tests and `git diff --check`; keep sibling dependency operations and existing runtime behavior unchanged. Standard close review and PR publication follow this implementation checklist.

## Log

### 2026-09-22 — fresh v2 task

Created from the agreed workspace/UI contract and the request for a clean task breakdown. Implementation has not started; estimates follow design approval.

### 2026-09-23 — Design audit

Claimed #245 and ran start-plan. Audited workspace resolution, change-code planning review/sync ordering and branch creation with a read-only peer audit. Proposed ordinary Git preparation avoids reviewing the wrong snapshot and avoids placing new planning checkpoints on resting refs. An unrelated untracked stale #244 issue copy is preserved untouched.

### 2026-09-23 — Spec review correction

Fresh review identified that arbitrary branch names or allocating an issue after preparation could make change-code switch to a second branch. Tightened the procedure to require an already allocated issue and its canonical filename stem; missing records are imported explicitly without reallocating IDs. A provisional-branch lifecycle is intentionally outside this procedure. Fresh re-review approved the corrected spec with no remaining important findings. Issue schema validation and diff whitespace checks passed; implementation and durable implementation planning await operator approval of this design.

### 2026-09-23 — Implementation and verification

Added one ordinary-Git guide, discovery links and 39 lines of embedded change-code guidance; no new CLI, runtime branch behavior or metadata. Added real-Git fixtures for pinned source/provenance, :0/:N symmetry, existing change-code reentry, cleanliness/collisions, capture movement, old baseline/planning preservation, non-origin refresh and sibling state preservation. Agent checks remain explicitly procedural, not claimed as binary enforcement.

Verification: procedure suite passed (19.554s); removing ignored-file protection made the collision test fail, then restored test passed. `go test ./pkg/workspace/... -count=1` passed (20.924s); related `TestWorkspace|TestChangeCode|TestCreateInPlaceBranch|TestResolveBranchName` command tests passed (52.224s). Prepared-branch reentry passed again without a structural bypass (7.567s). `go build -o bin/sdlc ./cmd/sdlc` and diff whitespace checks passed; rebuilt help displays the procedure. Published earlier selected design commits in order, resolving change-code's missing-prerequisite publication warning without publishing unrelated history. The untracked stale #244 file remains untouched.

## Revisions

### 2026-09-23 — Nested slots retain per-repository branch and refresh scope

Reason: operator agreed nested environments, ordinary remote dependency clones and existing per-repository publication. Delta: added the authoritative scope clarification and acceptance criteria above; original task context remains as provenance. Added #243 as a prerequisite for the nested identity contract. No implementation or lifecycle-status change is claimed by this revision.

### 2026-09-23 — Proposed ordinary-Git procedure design

Reason: implementation audit found change-code reviews destination artifacts and checkpoints planning before its current branch-creation step. Delta: propose explicit branch preparation before lifecycle gates, plus separately requested fast-forward refresh, using existing workspace identity and ordinary Git. Specify readiness, bounded capture semantics, provenance, configured remote handling and regression coverage without adding mutation commands. Pending operator design approval; no implementation or estimate yet.

### 2026-09-23 — Approved, keep implementation small

Operator confirmed this is agent branching guidance backed by tests and approved execution. Delta: replace the task outline with a compact implementation checklist; use quick flow because tests/docs are excluded from its production-code limit and the design is below 500 lines. No separate plan or additional approval needed.
