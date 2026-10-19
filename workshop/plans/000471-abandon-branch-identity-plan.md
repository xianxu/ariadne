---
collaboration-mode: discussion
---

# #471: Bind branch effects to the named issue

> **For agentic workers:** Consult AGENTS.md Section 3 for the execution approach. Follow the red-green steps below and keep this plan on the issue branch.

**Goal:** `sdlc abandon --issue A` must never archive or delete B's branch, and other issue-numbered verbs must not commit A's evidence on B's HEAD.

**Architecture:** The card path is the existing authority for an issue's canonical branch (`issue.BranchName`). A small pure branch-admission decision rejects another branch before any local commit, card publication or remote push. `abandon` retains its recorded ref/head recovery, but remote branch deletion is allowed only after a leased update brings that branch to the archived head; deletion itself leases that exact head.

**Tech stack:** Go, `cmd/sdlc`, the existing `trackerRepo` real-Git fixture and bare test remote.

## Core concepts

| Pure entity | Lives in | Status |
| --- | --- | --- |
| Issue-branch admission | `cmd/sdlc/trackerbranch.go` | new |

Issue-branch admission compares the checked-out branch with `issue.BranchName(card.Path)` and returns an error naming issue, expected branch and actual branch. It has no Git or network dependency (ARCH-PURE, ARCH-DRY). A started issue requires its branch; an open issue has no branch and remains card-only. A terminal abandon record may resume from its recorded branch or the resting branch, never another work branch.

| Integration point | Lives in | Status | Wraps |
| --- | --- | --- | --- |
| Branch effect preflight | `cmd/sdlc/abandon.go`, `closetracker.go`, `close.go`, `issueattach.go` | modified | tracker card, checkout identity |
| Abandoned remote branch retirement | `cmd/sdlc/abandon.go` | modified | Git remote refs and leased push/delete |

Use the existing `trackerEnv.git`/`remoteRefTip` and `trackerRepo` fixture. The bare test remote is the stateful Git double; no new shell seam is needed (ARCH-MOCK). No persisted format changes or new durable artifacts are introduced (ARCH-EVOLVE, ARCH-FUNERAL).

## State and recovery

| Internal state / failure | Visible state | Recovery | Possible loss |
| --- | --- | --- | --- |
| Open card, no issue branch | Card-only abandon allowed | Rerun same command if response lost | None |
| Started card, checkout on canonical branch | Abandon admitted | Rerun same command after interruption | None; archive ref keeps work |
| Started card, checkout on another branch or detached/rest | Refused before effects; names both branches | Switch to named issue branch | None |
| Terminal card with abandon record, on recorded branch or rest | Recovery admitted | Rerun same command | None |
| Terminal card with abandon record, on another work branch | Refused before archive/delete | Switch to recorded branch or rest | None |
| Remote branch moved away from archived head, or a CAS loses a race | Refused with observed ref and archive head | Inspect/reconcile the branch, then rerun | No branch deleted; archive remains |

The repo transaction lock covers local checkout/card steps; the remote can move concurrently. Read the remote branch tip before updating, require it to be an ancestor of the archived head, lease the advance to that head, then delete with a lease on that exact archived head. A concurrent push makes either lease fail. If the remote branch is absent, skip both; a repeat recognizes the terminal card and archived ref. Preflight the named branch and any currently divergent remote tip before the archive-ref push. This addresses the most dangerous interrupted ordering: archive/card publication succeeds while branch retirement fails (ARCH-SECURE, ARCH-STATE).

## `--issue`/HEAD audit

| Verb | HEAD effect and decision |
| --- | --- |
| `abandon` | Commits the log and deletes a branch; add the canonical branch guard before effects, and guard terminal retry. |
| `close`, `milestone-close` | Compute/review/commit evidence from HEAD for a named issue; require the canonical branch for both paths before review or mutation. |
| `issue attach` | Writes the named issue's details/assets in this checkout; require the canonical branch for started issues. Open issues may be attached on a resting branch before `start-plan`. |
| `start-plan`, `change-code`, `unclaim` started-work handoff, `finding dispute` | Already compare the current branch with the card-derived branch; retain tests and behavior. |
| `claim`, card setters/`set-status`, `issue publish`, `issue new`, `reclaim`, `actual correction`, `flake register`, landing priority | Operate on selected card/main/ledger or intentionally publish selected details; they do not use HEAD as the target issue's branch. Local mirror refresh is already limited to the card-derived branch. |
| `issue restore` | Intentionally restores *another* issue's details on the current branch after a transfer-guard refusal; its branch choice is the command's purpose. |
| Legacy `issue sync` | Explicitly commits selected issue details on the current branch in repositories without issue tracker; tracker-era use refuses. |
| `issue recovery reconcile` | Uses receipt source branch and refuses a foreign checkout; no new guard. |
| `judge`, `actual`, `usage`, `issue show/list/lint`, `issue attach` read side | Read-only or explicitly selected paths; no HEAD branch retirement. |

The audit is about mistaken branch ownership, not every use of `--issue`. Any path found to commit or delete the current branch on a named issue without these guards joins this change before close (ARCH-PURPOSE).

## Tasks

### 1. Red regressions

- [ ] In `cmd/sdlc/abandon_test.go`, use two started cards and branches in the existing bare-remote fixture. From B's branch, abandon A; assert refusal names A's expected branch and B's actual branch, and assert HEAD, both local/remote branch refs, archive ref, main and both cards are unchanged. Test A's branch success, open A on rest/card-only, and terminal retry from rest versus B.
- [ ] Test a remote A tip that is older than the note commit: the successful path advances it with a lease and then deletes it. Test a divergent/raced remote tip remains undeleted; archive ref is never mistaken for permission to delete a new tip.
- [ ] In close/milestone fixture tests, attempt `--issue A` from B's branch and assert no evidence commit or card change; in attach fixture, assert no asset/detail write. Run the targeted tests and observe the expected failures before implementation (TDD).

### 2. Minimal guard and remote CAS

- [ ] Implement the pure branch admission with a table covering correct, wrong, rest and detached names. Invoke it in `runAbandon` before `keepAbandonedWork`, in whole-issue and milestone close before review, and in started-issue attach before file writes. Preserve the explicit rest-only terminal retry.
- [ ] Tighten `dropAbandonedBranch`: inspect remote tip, reject divergence, lease an advance of an older tip to `rec.Head`, then lease deletion against exactly `rec.Head`. Keep the local ref's existing exact-head check. Make a failed remote probe an error, not proof of absence.
- [ ] Run the targeted tests green. Update `sdlc abandon --help`, the atlas pointer and `workshop/lessons.md` to state the branch-identity rule.

### 3. Verify and land

- [ ] Run `make -k build`, `make sdlc-build`, focused `go test`, `git diff --check` and the recorded `sdlc local-check run --name suite` on the pushed head. Do a short :0 staged live check only with hub grant; it must refuse an A-on-B abandon without changing refs.
- [ ] Run `sdlc close` and its built-in fresh review, open/update the PR, then `sdlc merge --yes`. Report the verified PR and merge SHA to the hub.
