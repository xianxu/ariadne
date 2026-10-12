# #308 `issue new` publishes card and details in one go: Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Filing an issue, and ticking a peer repo's project at close, write to origin through checkout-free compare-and-swap commits and never into a local working tree.

**Architecture:** The tracker's `creation` receipt operation keeps its reservation stage, and its second stage changes from "materialize details locally" to "publish details to main" (the same trunk candidate `transfer.main` builds, extracted into one helper). From a numbered slot, a claim follows in the same verb through the existing `cardsPublish` path, but without claim's local refresh. Close's peer-project path replaces the in-place write plus `planPeerWrites` auto-commit with a `TrunkFile.Update` against the peer's origin main. The transform re-derives the tick from the fetched bytes, so a moved peer is recomputed, never overwritten. A `tl:` owner seen on the fetched bytes refuses.

**Tech Stack:** Go (`cmd/sdlc`), git plumbing through `internal/gitx` (`TrunkFile`, `UpdateMany`), the tracker receipt engine (`internal/tracker`), CUE vocabulary (`construct/vocabulary/issue.cue`).

---

## Scope (after the 10-12 TL decisions)

| Spec item | Disposition |
|---|---|
| D-1: card + details to origin, no working-tree write | **M1**: `creation.main` stage |
| Claimant = creating slot `:N`; unclaimed from `:0` or a plain clone | **M1**: claim after creation, no local refresh |
| Point 3: identical bytes on the issue branch | **Dropped (TL 10-12, option a):** it assumed a content-based Behind that #274 did not build. With the ancestry-based guard, a separate identical-bytes commit is refused as Behind, and git merges a later edit as an add/add conflict. Details reach a branch with its next main merge, and later edits go through `issue publish`. |
| `move-detail` retires; `publish` keeps later edits | **M1**: `move-detail` is a hidden, deprecated alias of `issue publish --issue N`. Publish's first-publication path stays only to *finish an incomplete creation* (a pre-#308 card, or a creation whose main stage failed). Wording drops "escape hatch". |
| Skeleton placeholders | **M1**: section presence is status-aware. An `open` issue needs frontmatter and the sections present; Plan and Done-when content are required from `working` on. |
| F1 (landing gate on the branch's changed files) | **Verify only.** It landed with #322 (`runLandingInstanceGate` uses `merge-base..head`). M1 adds one regression row: a skeleton on main plus an unrelated branch lands. |
| C5: toplevel guard, origin required | **M1**: refuse when the issues dir is not `<git toplevel>/<issues-dir>`. Tracked repositories already require main's upstream (`openTracker`). |
| C1: peer project tick via trunk write | **M2** |
| `--no-project` honest semantics | **M2**: skips the whole project update |
| `tl:` projects skipped (#326) | **Verify only**, plus the re-check inside M2's transform |

Out of scope: B5 (cutover mismatch in a stale peer checkout). It becomes moot for `issue new`, which no longer reads the local checkout's files, and is noted in the Log if it is still reproducible.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `creationStages` | `cmd/sdlc/internal/tracker/creation.go` | modified |
| `DetailPublication` (main candidate builder) | `cmd/sdlc/internal/tracker/detailpublish.go` | new |
| `CheckSectionsPresence` (status-aware) | `cmd/sdlc/internal/issue/structural.go` | modified |
| `creatorClaims` (slot → claim decision) | `cmd/sdlc/issuenewclaim.go` | new |
| `projectCloseEdit` (per-project tick/detail-block edit) | `cmd/sdlc/projecttick.go` | new (extracted from `computeClose`) |
| `planPeerWrites` / `PeerWriteDecision` / `RepoGitState` | `cmd/sdlc/peerwrite.go` | deleted |

- **creationStages**: `{"creation.reserve", PublishCard}`, `{"creation.main", PublishMain}`. `NeedsSourceCheckout` for `creation` becomes `false`, since no stage acts locally.
  - **Legacy receipts:** a receipt written by an older binary, at stage 1 with a prepared `creation.detail` candidate (a blob, not a commit), cannot drive `creation.main`. `validateReceipt` keeps the shape valid, and recovery detects `stage==1 && candidate is not a commit` and refuses with "card #N is reserved; finish its creation with `sdlc issue publish --issue N`". Test: `TestRecoveryLegacyCreationReceiptPointsAtPublish`.
- **DetailPublication**: `func DetailPublication(main *gitx.TrunkFile, spec ReceiptSpec, detail []byte, msg string) (gitx.Candidate, error)`, the body of today's `transfer.main` Prepare (refuse `ErrDestinationExists`, `ExactBytes: true`). Used by `TransferOp` and `CreationOp`.
  - **DRY rationale:** two first-publication paths must refuse an existing destination identically.
- **CheckSectionsPresence**: reads `status:` from frontmatter. When `vocab.Issue().IsOpen(status)`, it runs frontmatter-present and spec-present only. Otherwise it runs all four checks, as today. `CheckStructural` (change-code) is unchanged in effect, since an issue is `working` by change-code. The doc comment states the rule.
  - **Relationships:** single source for `sdlc issue validate`, the landing instance gate, and change-code. That is why one edit covers all three.
- **creatorClaims**: `func creatorClaims(id workspace.Identity) bool`: true iff `HomeSlot()` parses to a slot `N ≥ 1`. `:0`, a plain clone (`HomeSlot()==""`), or an unparsable slot gives false. Pure over the identity value.
- **projectCloseEdit**: `func projectCloseEdit(pt string, in projectEditInput) (newPT string, applied, warns []string, err error)`, the milestone/issue branches of the loop at `close.go:822-873`, moved verbatim. A missing detail block without a waiver returns a typed `errMissingDetailBlock` carrying the skeleton text. `computeClose` calls it on the local bytes (to report, gate and dry-run), and M2's transform calls it on the fetched bytes.
  - **DRY rationale:** the local preview and the remote write must compute the same edit.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `CreationOp` | `cmd/sdlc/internal/tracker/createop.go` | modified | git (tracker + main refs) |
| `runIssueNew` | `cmd/sdlc/issue.go` | modified | tracker env, claim publish |
| `issueDirGuard` | `cmd/sdlc/issue.go` | new | `git rev-parse --show-toplevel` |
| `publishPeerProject` | `cmd/sdlc/peerproject.go` | new | `gitx.TrunkFile` on the peer |
| `newIssueMoveDetailCmd` | `cmd/sdlc/issuemovedetail.go` | modified (hidden alias) | — |

- **CreationOp**: gains a `main *gitx.TrunkFile`. `creation.main` Prepare reads the detail blob and calls `DetailPublication`. Apply and Probe use the `mainPublisher` + `publishEvent`/`probeEvent` pair `TransferOp` uses. `materialize` and `Checkout.Root` file writes are deleted. Tests: real git with a bare origin (the existing `createop_test.go` / `tracker_e2e_test.go` fixtures), not mocks.
- **runIssueNew**: (1) `issueDirGuard`, run **first**, before the `repositoryTracked` split, so the legacy path (C5's home-repo case) is covered too; (2) drive creation (reserve, then main); (3) if `creatorClaims`, `cardsPublish(env, []id, operationToken(tracker.OpClaim), nil, decide, nil)` with `claimSetDecision`, and **no** `refreshAfterClaim`/`refreshLocalMirror`; (4) print `Published #N (card + details on origin)` plus `claimed by <slot>` or `unclaimed`. The working tree and branch are never touched: delete the `onRest` hint and `commitOnly`. A claim failure after a successful creation is a warning with the next action (`sdlc claim --issue N`), not a failure, since the issue exists.
- **issueDirGuard**: `root := gitx.RepoTopLevel()`; for a relative `--issues-dir`, require `gitx.InsideRoot(root, dir) == filepath.Clean(dir)`. Otherwise refuse with "issue new files into the repository whose root you run it from: <cwd> is inside <root>, so <dir> would be <rel>; cd to the repository root". Absolute `--issues-dir` is explicit and not checked. Applies to the tracked and legacy paths (C5's emma-private case was the home repo).
- **publishPeerProject** (M2): `func publishPeerProject(ctx, peerRoot, projectPath string, in projectEditInput) (published bool, err error)`:
  - resolve `EvalSymlinks(projectPath)` and require it inside `EvalSymlinks(peerRoot)` (refuse otherwise);
  - refuse a brain (`gitx.IsBrainRepo`);
  - `gitx.NewTrunkFileContext(ctx, peerRoot, "origin", "main")`, refusing if there is no origin;
  - `Update(rel, "project: close-time update (<repo>#<id>)", transform)`, where `transform(old)` re-checks `project.TLOwners(old)` (refuse if owned now), runs `projectCloseEdit(old)`, and returns the old bytes when unchanged (no commit).
  - Bounded by `maxUpdateAttempts` (3). Every attempt re-derives from fetched bytes, and nothing is a blind replay.
  - The peer's checked-out files and index are never read for writing or touched.
  - `TLOwners(old)` returning an error refuses (fail closed, matching `close.go:813`).
  - **Ordering:** `publishPeerProject` runs last in `applyClose`, after the local issue/project writes, so the peer is touched only once local state is written. The close commit itself follows later in the close flow, so a close that dies after the publish leaves the tick on the peer for an issue not yet landed. A retry is safe: the transform re-derives from fetched bytes, and an already-ticked row with present detail fields returns the bytes unchanged (no commit).
  - A refusal or failure is a warning with the manual next action; the close does not fail, the same contract as today's report-only.

### ARCH notes
- **ARCH-CONSTRAINTS:** `issue new` makes up to three remote round trips: the card reservation, the main publication, and the claim (from `:N` only). Each is a fetch plus a push, so a few seconds interactively. Each is bounded by `tracker.MaxPublicationAttempts` (allocation races) or the receipt engine's retries. Exhaustion before the main stage leaves a recorded receipt (`issue recovery reconcile`). Exhaustion in the claim is a warning naming `sdlc claim --issue N`; the issue already exists unclaimed. The peer tick is one fetch plus a push per peer project, bounded by `maxUpdateAttempts` (3).
- **ARCH-FUNERAL:** creates nothing durable beyond what exists: receipts are the existing family, removed on completion, and the peer tick is one commit on the peer's main. `peerwrite.go` and its report-only "left uncommitted" residue are deleted.
- **ARCH-ORDER:** the CAS transform decides mergeability on fresh bytes. Card then main ordering is unchanged: a card without details is "incomplete creation", which recovery finishes.
- **ARCH-DRY:** `DetailPublication`, `projectCloseEdit` and `claimSetDecision` are each shared by two callers.

---

## Chunk 1: M1, issue new publishes (tracker, CLI, validator, guard)

### Task 1: Shared main-publication candidate
**Files:** create `cmd/sdlc/internal/tracker/detailpublish.go`; modify `transferop.go:76-90`; test `detailpublish_test.go`.
- [ ] Failing test `TestDetailPublicationRefusesExistingDestination` (bare origin with the file present → `ErrDestinationExists`) and `TestDetailPublicationWritesExactBytes` (CRLF/no-trailing-newline bytes round-trip).
- [ ] Extract the body of the `transfer.main` case into `DetailPublication`, and have `TransferOp` call it. Run `go test ./cmd/sdlc/internal/tracker/ -run 'DetailPublication|Transfer'` → PASS.
- [ ] Commit `#308: tracker: one main-publication candidate for creation and transfer`.

### Task 2: `creation.main` replaces `creation.detail`
**Files:** modify `creation.go`, `createop.go`, `receipt.go` (`NeedsSourceCheckout`, `MaterializeDetail` removed if unused), `drive.go` if it names stages; tests `creation_test.go`, `createop_test.go`.
- [ ] Failing tests:
  - `TestCreationPublishesDetailsToMainWithoutWritingTheCheckout`: after Drive, origin/main has the exact bytes at `workshop/issues/<id>-<slug>.md`, the worktree has no such file, and `git status --porcelain` is empty.
  - `TestCreationMainRaceRecomputesOnMovedMain`: main moves between prepare and push, and the publish lands on the new tip.
  - `TestCreationMainUncertainIsRecoverable`: the push transport fails, giving `ErrOperationUncertain`, and a later Drive from the receipt finishes it.
- [ ] Implement. `NewCreationOp` takes `main *gitx.TrunkFile`; update the callers `issue.go:322` and `issuerecovery.go:152`.
- [ ] Legacy receipt: `TestRecoveryLegacyCreationReceiptPointsAtPublish` (a hand-built v1 receipt at stage 1 with a blob candidate → recovery refuses, naming `issue publish --issue N`). Implement the check in `issuerecovery.go`.
- [ ] `go test ./cmd/sdlc/internal/tracker/ ./cmd/sdlc/ -run 'Creation|Recovery|IssueNew'` → PASS. Commit.

### Task 3: `runIssueNew`: claim from a slot, nothing local, guard
**Files:** modify `cmd/sdlc/issue.go:232-370`; create `cmd/sdlc/issuenewclaim.go` (+ `_test.go`); tests in `issuenew_test.go`.
- [ ] Failing tests:
  - `TestCreatorClaims` (table: `ariadne:4` → true, `ariadne:0` → false, `""` → false, `pair:1` from an environment host → true).
  - `TestIssueNewFromSlotClaimsAndLeavesTreeClean`: slot fixture; the card has the claimant with `Workspace=repo:N`; a tracker commit has a `Tracker-Operation: claim-…` token (so `ClaimTimes` sees it); main has the details; the worktree is clean and HEAD unchanged.
  - `TestIssueNewFromRestingZeroIsUnclaimed`: the card has no claimant; the resting branch HEAD is unchanged; the tree is clean.
  - `TestIssueNewOnIssueBranchDoesNotCommitThere`: on an issue branch, HEAD and the tree are unchanged.
  - `TestIssueNewRefusesWhenCwdIsNotTheRepoRoot` (nested non-repo dir under a git repo → refusal naming the root; nothing published).
- [ ] Implement per the integration-point notes. Remove the `onRest` hint, the `commitOnly` call, and the `env.branch == ""` refusal (it existed only for the local write; a detached HEAD can file). Stdout contract: it still prints the repo-relative details path, now meaning "where the details are on main". Grep stdout consumers (`skills`, `construct/`, scripts) and update any that open the file locally; update `--help` Long text (publishes to origin; claimed from a slot).
- [ ] Run targeted tests → PASS. Commit.

### Task 4: Status-aware section presence (skeletons)
**Files:** modify `cmd/sdlc/internal/issue/structural.go:75-107`; tests `structural_test.go`, `landinggate_instance_test.go`.
- [ ] Failing tests:
  - `TestSectionsPresenceOpenSkeletonPasses` (rendered scaffold with `status: open` → no failures).
  - `TestSectionsPresenceWorkingRequiresPlanAndDoneWhen` (same text, `status: working` → plan-present and done-when-present).
  - `TestLandingInstanceGateIgnoresMainSkeleton` (F1 regression: main gains a skeleton after the branch point; the branch changes something else; the gate passes).
- [ ] Implement; update the doc comment and `helptext/issue.md` where it documents section requirements.
- [ ] Run `go test ./cmd/sdlc/internal/issue/ ./cmd/sdlc/ -run 'Sections|Structural|LandingInstance|IssueValidate'` → PASS. Commit.

### Task 5: Retire `move-detail`; publish wording
**Files:** `issuemovedetail.go` (command becomes `Hidden: true`, `Deprecated: "use sdlc issue publish --issue N"`, RunE → `runTrackedPublish`), `issuepublish.go` (Short/Long: "later edits; also finishes an incomplete creation"), `helptext/*.md`, `atlas/workflow/*.md`, `AGENTS.base.md` §2 bullet (#284 text about first publication), recovery catalog lines.
- [ ] Failing test `TestNoHelpCallsPublicationAnEscapeHatch`: the scan pattern from `TestNoTextSaysOnlyTheOwnerEditsDetails` (same file set), rejecting `escape hatch` near `move-detail`/`publish`, and `move-detail` in any non-hidden help.
- [ ] Grep `move-detail|first publication|escape hatch` across `cmd/sdlc`, `atlas`, `AGENTS*.md`, `README.md`, `construct/`, and fix every hit (lessons: the policy's words live in many places).
- [ ] Run targeted tests → PASS. Commit.

### Task 6: M1 close
- [ ] `make test` green; record the evidence.
- [ ] Live check: from ariadne:4 (`sdlc issue new "scratch #308 live check" --dry-run` first, then a real one), confirm card + details on origin, claimed by ariadne:4, `git status` clean; then `sdlc abandon --as wontfix` the scratch issue. Log it.
- [ ] Atlas: `atlas/workflow/` issue-tracker page: creation publishes; `move-detail` deprecated.
- [ ] `sdlc milestone-close --issue 308 --milestone M1`.

## Chunk 2: M2, close ticks peer projects on their origin

### Task 7: Extract `projectCloseEdit`
**Files:** create `cmd/sdlc/projecttick.go` (+ `_test.go`); modify `close.go:822-873`.
- [ ] Characterization tests first: milestone tick + detail block, issue tick of N rows, a missing block (unwaived → `errMissingDetailBlock` with the skeleton; waived → a warning). They pass against the extracted function and the existing close tests stay green.
- [ ] Commit `#308: close: extract the per-project edit`.

### Task 8: `publishPeerProject` and the apply path
**Files:** create `cmd/sdlc/peerproject.go` (+ `_test.go`); modify `close.go` (`applyClose` 950-1000, `printCloseDryRun`, `captureCloseReviewSnapshot` 1631: skip peer edits); delete `peerwrite.go`, `peerwrite_test.go`, `peerwrite_apply_test.go` (port any still-meaningful rows).
- [ ] Failing tests, real git with a peer repo plus a bare origin:
  - `TestPeerProjectTickLandsOnOriginAndLeavesCheckoutUntouched`: peer checkout HEAD, index and file bytes unchanged; origin/main has the ticked row.
  - `TestPeerProjectTickResolvesSymlink` (ops layout: `workshop/projects/x.md` → `teams/…/project.md`; the commit changes the real path and is not empty).
  - `TestPeerProjectTickRecomputesOnMovedOrigin` (origin gained an unrelated edit to the same file → both edits present).
  - `TestPeerProjectTickRefusesWhenOriginGainedTL` (origin's copy now has `tl:` → no commit, warning).
  - `TestPeerProjectTickRefusesBrainAndNoOrigin`.
  - `TestCloseReviewSnapshotExcludesPeerProjects` (a peer project edited after review does not stale the SHIP).
- [ ] Implement: `computeClose` keeps a local preview for peers (messages, gate) but marks `projectEdit.peer = true`. `applyClose` writes only same-repo edits locally and calls `publishPeerProject` for peers. The dry run prints `Would publish to <peer> origin/main: <rel>`.
- [ ] Commit.

### Task 9: `--no-project` skips the whole project update
**Files:** `close.go` (flag help at 200; the sweep at 780 returns early when `f.NoProject`; `--force` keeps only waiving the detail-block requirement); the gate catalog entry; `milestoneclose.go` if it shares flags.
- [ ] Failing test `TestNoProjectSkipsTickAndPeerWrite` (a project with a matching row; `--no-project` → row unticked locally and on the peer's origin; info line says skipped).
- [ ] Implement; update the help/catalog text and `TestHelpFlags` fixtures. Rewrite the missing-detail-block refusal (`close.go:860`): the precise ways out are authoring the block, or `--no-project` (which now skips the project update entirely, the tick included). Accepted trade-off (TL decision 10-12): a track-only milestone that wants the tick but no prose block authors a one-line block instead of waiving. Commit.

### Task 10: Close #308
- [ ] Verify #326 (`tl:` skip) still holds: `go test ./cmd/sdlc/ -run 'TLOwned|ProjectClose'`.
- [ ] `make test` green.
- [ ] Live check of Done-when 4 against a scratch peer (a temp clone pair, not ops), logged.
- [ ] Atlas: close flow page (peer tick via origin; `--no-project`). Lessons entry only if the review finds something not code-enforced.
- [ ] `sdlc close --issue 308 --verified '…'`, then `sdlc pr`, merge main, `sdlc merge --yes`.
