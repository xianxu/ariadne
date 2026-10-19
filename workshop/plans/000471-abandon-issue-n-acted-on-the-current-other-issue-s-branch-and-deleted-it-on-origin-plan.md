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

Close has one additional recovery admission: a differently named branch is the same work only when its HEAD exactly equals the canonical issue branch's local tip, or a fresh remote tip if that local ref is absent. This admits a renamed branch created at the issue tip after an outside PR merge (#285/#365), before a new close-evidence commit is made. A different head, even under a suggestive name, refuses. Abandon and attach retain exact-name admission. The equality proof does not authorize branch deletion.

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

The repo transaction lock covers local checkout/card steps; the remote can move concurrently. Read the remote branch tip before updating, require it to be an ancestor of the local tip before any archive push, and of the archived head before retirement. If the remote has commits beyond local work, refuse and name both tips; never advance past those commits. Lease the advance to the archived head, then delete with a lease on that exact archived head. A concurrent push makes either lease fail. If the remote branch is absent, skip both; a repeat recognizes the terminal card and archived ref. Preflight the named branch and any currently divergent remote tip before the archive-ref push. This addresses the most dangerous interrupted ordering: archive/card publication succeeds while branch retirement fails (ARCH-SECURE, ARCH-STATE).

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

- [x] In `cmd/sdlc/abandon_test.go`, use two started cards and branches in the existing bare-remote fixture. From B's branch, abandon A; assert refusal names A's expected branch and B's actual branch, and assert HEAD, both local/remote branch refs, archive ref, main and both cards are unchanged. Test A's branch success, open A on rest/card-only, and terminal retry from rest versus B.
- [x] Test a remote A tip that is older than the note commit: the successful path advances it with a lease and then deletes it. Test a divergent remote tip remains undeleted; both pushes use leases so a concurrent update refuses, and the archive ref is never mistaken for permission to delete a new tip.
- [x] In close/milestone fixture tests, attempt `--issue A` from B's branch and assert no evidence commit or card change; in attach fixture, assert no asset/detail write. Run the targeted tests and observe the expected failures before implementation (TDD).

### 2. Minimal guard and remote CAS

- [x] Implement the pure branch admission with a table covering correct, wrong, rest and detached names. Invoke it in `runAbandon` before `keepAbandonedWork`, in whole-issue and milestone close before review, and in started-issue attach before file writes. Preserve the explicit rest-only terminal retry.
- [x] Tighten `dropAbandonedBranch`: inspect remote tip, reject divergence, lease an advance of an older tip to `rec.Head`, then lease deletion against exactly `rec.Head`. Keep the local ref's existing exact-head check. Make a failed remote probe an error, not proof of absence.
- [x] Run the targeted tests green. Update `sdlc abandon --help`, the atlas pointer and `workshop/lessons.md` to state the branch-identity rule.

### 3. Verify and land

- [ ] Run `make -k build`, `make sdlc-build`, focused `go test`, `git diff --check` and the recorded `sdlc local-check run --name suite` on the pushed head. Run an A-on-B abandon with this branch's binary in a disposable clone and bare origin; prove all local and origin refs stay unchanged.
- [ ] Run `sdlc close` and its built-in fresh review, open/update the PR, then `sdlc merge --yes`. Report the verified PR and merge SHA to the hub.

## Revisions

- 2026-10-19, hub approval: the remote branch may be advanced with a lease to the archived head only after proving its current tip is contained by the local issue tip. A remote tip with commits beyond local work refuses before the archive push and names the divergent tips; a later concurrent push fails the lease.
- 2026-10-19, full-suite regression and hub decision: the existing #285/#365 recovery closes from a hand-made branch at the canonical issue tip after an outside PR merge. Tracker cards record no PR before close, so a PR-number join cannot prove this case. Keep the exact-name guard for abandon and attach; for whole and milestone close, also admit exact HEAD equality with the canonical local branch tip (or the fresh remote tip only when local is absent). Add a different-head negative. The unrelated `dev-aliases.test.sh` added/deleted-file failures are the registered #465 flake and are not changed here.
- 2026-10-19, hub live-check decision: do not stage on ariadne:0 because its `sdlc` binary serves all fleet slots. Run the same wrong-branch command with this branch's rebuilt binary in a disposable clone and bare origin, comparing every relevant ref before and after. The probe passed without touching the shared checkout.
- 2026-10-19, close review round 1: the earlier local-tip check can go stale when a post-checkout hook moves the branch before `git branch -D`; this deletes a tip absent from the archive. First add a deterministic real-Git regression that moves the local ref in that gap and requires refusal with the moved ref intact. Then replace local deletion with atomic `git update-ref -d refs/heads/<branch> <recorded head>`, retaining the existing preflight for a clear early error. Inspect the other local branch-retirement sites for the same class and record their disposition. Update README's attach and abandon paragraphs with the checkout requirements: started attach and abandon require the canonical issue branch; an open issue may attach from rest; an interrupted terminal abandon may resume from its recorded branch or rest. Re-run targeted tests, the recorded suite on the new code head and the close review.
- 2026-10-19, CAS implementation check: unlike `git branch -D`, `git update-ref -d` will remove a branch checked out in another linked worktree. Preserve the prior occupancy refusal by reusing `landingUnoccupied` immediately before the CAS. A real-Git interrupted-abandon fixture proved the regression red, then green with the guard; the expected-tip CAS still protects a branch moved by a checkout hook.
