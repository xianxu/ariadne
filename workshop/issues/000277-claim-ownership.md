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

- [x] M1 — record and publish ownership: the `claimant` card kind, a mirrored
      field and a pure match; the machine fingerprint plus identity seam; claim
      stamps ownership in its CAS; race and mirror tests; docs.
- [x] M2 — enforce at start-plan, change-code, close and milestone-close; add
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
- 2026-10-01: closed — Claim records claimant (operator, fingerprinted OS machine ID, readable name, optional slot label, worktree, repository) in its status CAS, mirrored to details; two-clone claim and adopt races have one winner with full record, loser publishes none; start-plan/change-code/close/milestone-close refuse foreign (naming owner) and unattributed (toward --adopt) before review — mutation-checked; ownership survives restart (built binary); sdlc move relocates owner only with its own local record (takeover regression mutation-checked); TestVerbContractTable covers every cell; help/atlas/README/vocab consistent (TestEveryFlagAppearsInItsHelp); full sharded suite green twice (processgroup ps failure is sandbox-only).; review verdict: SHIP
- 2026-10-01: closed M2 — ownership gate on start-plan/change-code/close/milestone-close (mutation-checked); relocation requires sdlc move's local record (BR-9 regression mutation-checked; record removed when nothing moved; unreadable record surfaces); TestVerbContractTable covers every cell; adopt race; real sdlc move tests; TestEveryFlagAppearsInItsHelp + help/atlas/plan/README restatements; set-status uses one tracker env; full sharded suite twice: all cmd/sdlc green; actual = 1.84 h measured minus M1's 0.76h; --no-project: pair's project tracks ariadne#277 at issue granularity; review verdict: SHIP
- 2026-10-01: closed M1 — claimant card kind (fail-closed, mirrored), identity seam (fingerprinted OS machine ID, slot label only where pkg/workspace says the layout is in use), claim stamps owner in its CAS; owner repeat no-op (dry-run writes nothing), foreign refused naming owner, unattributed -> --adopt; race test accepts all three loser refusals and checks the winner's full claimant; review round-1 findings all fixed; actual = sdlc actual measurement (first milestone); --no-project: pair's project tracks ariadne#277 at issue granularity; review verdict: SHIP

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
- M1 boundary review round 1 returned FIX-THEN-SHIP. I fixed all seven
  findings in this round:
  - Race-test oracle: it now accepts the late-reader "claimed by" refusal
    (BR-1).
  - BR-2: the plan's Revisions now match the fixed-key fingerprint. Added a
    domain-separation test.
  - Minors:
    - an owner's dry-run repeat claim writes nothing;
    - the flag-day rollout paragraph moved into M1's atlas;
    - the race test skips without a host machine ID;
    - the slot-layout question moved into
      `pkg/workspace.Identity.UsesSlotLayout`;
    - added a `claimDecision` ownership table test.
  - Verified the real tracker has no claimant cards. Claims go through
    :0's pre-#277 binary until this lands.
- M1 advisory (BR-8): reordered the slot-label test so the no-slots state is exercised first.
- M2 in progress:
  - `requireIssueOwnership` is wired into start-plan, change-code and
    `computeClose` (both modes, before the review).
  - `claim --adopt` records an owner only on unattributed working/blocked
    cards.
  - A repeat claim repairs relocated work (`relocatable` /
    `relocateClaimant`).
  - `set-status` into working records the claimant, and refuses on a foreign
    card even under `--force`.
  - The planning-review fixture now seeds its working card with the host
    identity.
- M2:
  - `sdlc move` relocates the owner after its verified switches
    (`relocateAfterMove` → `moveRelocation`). It is skipped in pre-tracker
    repositories, which have no cutover marker. On failure it warns and names
    the `sdlc claim` repair.
  - `ownership_test.go` covers:
    - each of the four gates against Foreign and Unknown, with the judge never
      dispatched, then adopt restoring the owner;
    - adopt only on unattributed work;
    - relocation with a forced re-stamp failure, its repair, and refusals for
      another machine and for a still-held branch;
    - restart survival through the built binary.
  - Mutation check: disabling `requireIssueOwnership` turns all four gate
    subtests red.
- M2 docs:
  - Ownership notes went into the help for start-plan, change-code, close,
    milestone-close, move and set-status.
  - The atlas Claimant section gained Enforcement (gate, adopt, set-status,
    relocation, #278/#279 boundary) and updated verb rows.
  - The process manual needed no change: it catalogs only each help text's
    first paragraph.
  - `TestVerbContextsReachTrackerReads` caught move's relocation using
    `context.Background`; the command context now threads through `runMove`.
- M2 review round 1 returned FIX-THEN-SHIP with four Important findings.
  - **BR-9 (real takeover hole):** relocation accepted the old worktree's
    absence from the branch as proof of a move. Another slot could
    `git switch -c <issue-branch> && sdlc claim` to take the card. The fix
    requires *positive* evidence: a local relocation marker that `sdlc move`
    writes in the git common dir, naming from/to. Relocation (move's re-stamp,
    claim's repair, the gate's hint) requires a marker matching the recorded
    owner and this destination. Added a negative test.
  - **BR-10:** the verb contract becomes one table test over pure decisions.
    The adopt decision is now pure. set-status working→working on an
    unattributed card refuses toward `--adopt` (it silently adopted before).
    Added a two-clone adopt race.
  - **BR-11:** exercise moveRelocation's success and Unknown paths, and a real
    `runMove` when the fixture allows it.
  - **BR-12:** a README sentence.
  - Minors: gofmt; the shared help paragraph; surface relocatable's error;
    open and snapshot the tracker once per close.
- M2 round-1 fixes, part 1:
  - **BR-9 closed.** `RelocationAllowed` requires `sdlc move`'s record
    (`issue.Relocation`, stored in `<git-common-dir>/sdlc/relocations/<id>.json`)
    naming the recorded owner as the source and this worktree as the
    destination.
    - Move writes the record before switching and retires it on success or
      when relocation doesn't apply. A failed re-stamp keeps it for the
      `sdlc claim` repair, which removes it.
    - The regression test (a hand-switched branch plus claim is refused) was
      mutation-checked: it goes red when the record requirement is removed.
  - **BR-10.** `adoptDecision` is pure. set-status working→working on an
    unattributed card refuses toward `--adopt`. `TestVerbContractTable` covers
    every situation × verb cell of the plan.
  - **BR-11.** A real `sdlc move` over a tracker slot fixture relocates its
    owner and retires the record. The unattributed path warns toward
    `--adopt` and leaves no record.
  - Minors: relocatable's probe error is surfaced; one tracker read per
    close (`requireCardOwnership` inside `prepareTrackerClose`).
- M2 round-1 fixes, part 2:
  - BR-10's two-clone adopt race now exists, using a shared
    `raceBuiltBinary` helper that the claim race uses too.
  - BR-12: README sentence on owners, `--adopt`, move and the rebuild.
  - Minor: the help is single-sourced through `{{OWNERSHIP_GATE}}`
    (`ownershipGateHelp`), and change-code's note moved out of its numbered
    list.
  - Full suite: every cmd/sdlc test passes. Two load-timing flakes in a first
    run (`TestCLISignalCancelsOwnedReviewer`,
    `TestPlanningReviewConcurrencySchedules`) passed in isolation and in the
    rerun.
  - Process note: a parallel batch ran in the wrong directory and appended to
    a stray root `ownership_test.go`. I moved its content into the right file
    and removed the stray.
- M2 review round 2: one Important finding (BR-17), atlas and plan
  restatements still described the pre-BR-9 rule. Fixes:
  - Updated the atlas Relocation entry (evidence rule plus record lifecycle),
    move help, plan Revisions and the plan's Verb-contract set-status cell.
  - Added two lessons: update every restatement when a contract changes; and
    absence is not evidence of a transfer.
  - BR-15 is now tested: an unreadable move record surfaces at the gate.
  - BR-18: a move whose first switch fails removes its record. Tested, and
    mutation-checked.
- M2 review round 3: BR-19, `claim --help` lacked `--adopt` and contradicted
  the relocation repair (third docs finding). Fixes:
  - Fixed the class: `TestEveryFlagAppearsInItsHelp` requires every
    registered, non-hidden flag to appear in its page's FLAGS section. Gate
    flags are exempt because the catalog renders them.
  - The test found pre-existing gaps: change-code's `--sandbox`,
    `--issues-dir` and `--plans-dir`, and merge/push `--plans-dir`. Fixed
    those. Hid the retired `claim --no-start`.
  - The claim help now states the relocation exception.
  - Fixed the plan citation, and widened the restatement lesson to include
    help pages.
- Full suite: `TestPlanningReviewConcurrencySchedules` timed out (2s budget) under load in two runs. The cause was mine: set-status working opened the tracker a second time to resolve the identity. `cardUpdate` decisions now receive the setter's own `trackerEnv`, so there is one open.
- M2 advisory (BR-21): start-plan and change-code now judge ownership on the card they already hold (`requireCardOwnership`, plus `refreshMirrorFrom`). The verdict and the mirror read the same card version, with no second snapshot.
