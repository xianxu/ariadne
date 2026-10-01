---
id: 000277
status: working
deps: []
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours: 4.09
card_mirror: '15223d7e7c44c0f427a6e4667bb121e124bcd929' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-01T13:19:29-07:00
flow: {kind: full, provenance: inferred}
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

Durable plan: `workshop/plans/000277-claim-ownership-plan.md`.

- [ ] M1 — record and publish ownership: the `claimant` card kind, a mirrored
      field and a pure match; the machine fingerprint plus identity seam; claim
      stamps ownership in its CAS; race and mirror tests; docs.
- [ ] M2 — enforce at start-plan, change-code, close and milestone-close; add
      `claim --adopt` for unattributed working cards; make set-status into
      working stamp or refuse; make `sdlc move` re-stamp the owner's own
      relocation; restart-survival test; docs, process manual and rollout note.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.8 impl=0.05
item: greenfield-go-module   design=0.5 impl=0.22
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.2 impl=0.14
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.1 impl=0.14
item: smaller-go-module      design=0.2 impl=0.14
item: cross-cutting-refactor design=0.2 impl=0.14
item: atlas-docs             design=0.1 impl=0.05
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 4.09
```

Items, in order:
- issue-spec: brainstorm, operator decisions and three plan-quality rounds.
- greenfield-go-module: pure `claimant.go` (parse/validate, match,
  fingerprint, relocation).
- smaller-go-module ×5:
  1. card kind, `sameCardValue` and the vocabulary;
  2. the identity seam (ioreg, machine-id, names);
  3. claim decision, `--adopt` and set-status;
  4. `requireOwnership` at four gates;
  5. move relocation and the claim repair.
- cross-cutting-refactor: existing claim, close and gate tests gain an
  identity.
- atlas-docs, then two milestone reviews (M1, M2).

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* The calibration source is flagged stale (#127), so treat the numbers as provisional.

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
- Operator decisions, 2026-10-01. Asked after noting that the ariadne
  tracker branch is public:
  1. **Machine identity:** the OS machine ID (macOS `IOPlatformUUID`, Linux
     `/etc/machine-id`), hashed with an ariadne-specific key, plus a readable
     name. The raw hardware ID is never published. MAC address was rejected.
  2. **Storage:** a structured `claimant` card field in the vocabulary,
     mirrored into the details frontmatter. Older binaries refuse such cards,
     which fails closed.
  3. **Legacy unattributed `working` cards:** ownership gates refuse them.
     `sdlc claim --issue N --adopt` records an owner on an unattributed card
     only. Reassigning an owned card stays with #278.
- The operator asked why the machine ID is hashed. The tracker is public, and
  the raw `IOPlatformUUID` / `machine-id` is a permanent fingerprint that its
  owners treat as confidential. Matching only needs equality, which a keyed
  hash preserves, the same way systemd's app-specific IDs work.
- Wrote the durable plan with milestones M1 and M2. The Spec requirements are
  unchanged.
- Operator review of the plan: `workspace` (a slot such as `ariadne:1`) is a
  Couch-specific concept, and ariadne must work without Couch. The field is
  now optional and informational only, read from `pkg/workspace` when
  resolvable, and ownership matching ignores it. Added a plain-clone test.
  Plan Revisions updated.
- change-code plan-quality round 1 returned four Important findings:
  - milestone-close bypasses `prepareTrackerClose`;
  - `sdlc move` would orphan the owner;
  - the built-binary tests can't reach a package-var seam;
  - an older binary aborts the whole snapshot on one claimant card.
  Operator decisions: **move re-stamps** the claimant for the owner's own
  relocation, and a **documented flag-day** rollout. The plan has been
  revised (see its Revisions).
- Plan-quality round 2: round 1's findings were disposed as addressed. One
  new Important finding: move's re-stamp ordering. The plan now generalizes
  the re-stamp into a convergent same-machine *relocation*. It runs after
  move's switches, and `sdlc claim` at the destination repairs a failure.

- Plan-quality cleared after 3 rounds. Estimate derived with v3.1 (4.09h);
  the calibration source is stale per #127.
- change-code passed (estimate-quality: reasonable; relocation is the likeliest overrun). Starting M1.
- M1 step done: added the `claimant` card kind (issue.cue plus the
  regenerated `pkg/vocab/issue.json`) and pure `internal/issue/claimant.go`.
  - Parse and validation fail closed: exact keys, nonempty one-line strings,
    and `machine` must be a 32-hex fingerprint.
  - `SetCardClaimant` / `CardClaimant`, `MatchClaimant` (keyed on repository,
    machine and worktree), `RelocationAllowed` and `MachineFingerprint`.
  - The mirror needed only `sameCardValue` to compare mappings. Table tests
    pass.
- M1 step done: the identity seam `cmd/sdlc/claimant.go`.
  - Reads git `user.name`; the macOS `ioreg` IOPlatformUUID or Linux
    `/etc/machine-id` (then dbus), fingerprinted; ComputerName or hostname;
    the canonical worktree; and the publication repository.
  - `workspace` is recorded only for a slot, or for a primary with slot
    worktrees beside it. A plain clone gets none: a `primary` classifies as
    `repo:0`, which would have mislabeled every plain clone.
  - Parser tests plus a live host test, which ran rather than skipped here.
- M1 step done: claim stamps the claimant in its existing CAS.
  - `claimDecision` takes `*issue.Claimant`; a pre-tracker repository passes
    nil and keeps the old behavior.
  - An owner's repeat claim is a no-op success. Another workspace's is
    refused, naming the owner. An unattributed working card is refused,
    pointing to `--adopt`.
  - Tests: the two-clone race (winner's full claimant; the loser's worktree
    is absent from tracker history), the repeat claim, and the mirror showing
    the owner in details with no slot label for a plain clone.
- M1 docs: added an ownership contract to the claim help and a Claimant
  section plus the claim row to the atlas `issue-tracker.md`.
  - Full sharded suite: all 864 cmd/sdlc tests pass.
  - `internal/processgroup` `TestCancellationKillsDescendants` still fails;
    the sandbox blocks `/bin/ps` there, and the package is untouched.
  - Running M1's milestone-close next.
