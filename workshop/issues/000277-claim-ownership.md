---
id: 000277
status: working
deps: []
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: 'c5b96f7820a325963bb81a0d13a7460b8edc406a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-01T13:19:29-07:00
---

# Record claimant ownership atomically with issue reservation

## Problem

Claims reserve an issue atomically, but do not durably identify the responsible operator, machine and workspace. Humans currently remember assignments, and a working status alone cannot justify another agent resuming the task.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Publish claimant identity in the issue-tracker card in the same compare-and-swap transaction as open → working. Capture operator name, machine identity/readable name, qualified slot (e.g. pair:1) and canonical worktree path, with repository identity sufficient to resolve the assignment. Keep the record after process exit, parking or shutdown; process liveness is not responsibility.

The operator proposed a MAC address for stable machine identity. Evaluate its identity source, selection, stability, replacement and collision behavior before choosing the representation; do not silently make a network interface label the authority. No Couch ID in durable ownership: local singleton is the project direction. Git provides claim chronology; do not add a duplicate claim timestamp. A distinct claim ID is not a requirement: first establish whether the tracker/card revision already distinguishes claim/reclaim generations.

Enforce claim-before-new-work centrally in SDLC. Continuing work requires a matching assignment, not merely status=working. Cover ordinary callers outside Couch explicitly rather than inventing a slot. Legacy working cards without attribution remain unknown until operator-assisted adoption/reclaim. Do not equate matching operator/machine alone with a matching worktree.

ARCH-DRY: one tracker authority for assignment; ARCH-PURPOSE: ownership is responsibility, not an assertion of instantaneous process state.

## Done when

- Concurrent claims from separate clones have one winner, whose complete ownership record is atomically visible; loser publishes no ownership.
- Repeated claims cannot take another workspace’s working issue; SDLC implementation-entry/resume gates check the intended ownership contract.
- Ownership survives restart/parking; historical or missing claimant metadata is exposed as unknown with an operator-directed recovery path.
- Machine identity and outside-Couch behavior have an explicit reviewed contract and tests; no redundant claim ID/timestamp added without a demonstrated need.
- Vocabulary, tracker mirrors, CLI help and tests expose the new record consistently.

## Plan

Implementation plan to be designed after issue claim and start-plan; these are requirements, not an approved implementation plan.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-01 (implementation session)

- Operator authorized the work ("work on #277"). Claimed in ariadne:1 and
  ran start-plan. Branch `000277-claim-ownership` sits at main. Mapping the
  claim CAS, the card vocabulary and the gates that read `working` before
  designing.
- Code map (explorer):
  - Claim is `claimDecision` (claimdecision.go) followed by a
    `tracker.UpdateCard` blob-OID CAS. There are 3 attempts, and a race on
    the same card yields `ErrCardChanged`. A repeat claim already refuses
    with "not open".
  - Nothing on the card distinguishes a claim from a reclaim, beyond the
    card's blob OID and the `Tracker-Operation: claim-…` trailer.
  - Card fields come from `issue.cue` `card.fields`. They are scalar only
    (`validateCardScalar`), and mirroring is span-based.
  - The `tracker:` envelope (handoff/completion) decodes without KnownFields.
    An older binary re-marshalling it would silently drop a new member.
  - Workspace identity: `pkg/workspace.Identity` has
    `Address` (`repo:N`), `Slot` and `WorktreeRoot`. `openTrackerAt` keeps
    only the root and branches. There is no operator/machine helper.
  - Gates that read `working`:
    - start-plan requires `working` (startplan.go:293).
    - change-code checks only the branch.
    - close refuses only `done`.
    - milestone-close and move-detail check no status.
    - `issue set-status working` can claim through its own CAS, bypassing
      claim's readiness check.
- Machine identity sources here: macOS `IOPlatformUUID`, ComputerName and
  hostname. There is no `/etc/machine-id` on macOS; Linux has one. A MAC
  address needs an interface choice (en0?), Wi-Fi private-address
  randomization makes it unstable, and it changes when a NIC is replaced.
