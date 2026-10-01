# Boundary Review — ariadne#277 (milestone M2)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | e53c075bd9ae04f71535872f44d124abf2839d4b..83dcbe83cb500b503ac24078d594e2cb1b97bb89 |
| command | sdlc milestone-close --issue 277 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-01T14:34:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 does what it set out to do at the gates. `requireIssueOwnership` is a single helper, and start-plan, change-code and `computeClose` all call it (both close modes, before any judge is dispatched). `claim --adopt`, the set-status stamping and the post-move relocation are wired, and the help text and atlas are updated. Nothing here is a crash or data-loss bug. Two things should be fixed before shipping:

- **Relocation can take ownership with no move behind it.** It treats "the recorded worktree is not on the issue branch" as proof that a move happened. That condition is also true right after a claim and before start-plan, and whenever the owner has simply switched branches.
- **Several Plan and Verb-contract cells aren't delivered or tested.** The `sdlc move` success path that actually relocates the owner is never exercised. There is no adopt race test and no adopt table. set-status quietly acts as a second adopt path.

There are also a README gap and a gofmt failure, both cheap.

## 1. Strengths
- **One gate helper, called before the review.** `cmd/sdlc/close.go:510-520` runs the check for both modes ahead of `prepareTrackerClose`, which matches PQ-1. The test asserts the judge is never dispatched on a refusal (`ownership_test.go:75-104`).
- **Move relocates only after the local move is complete.** The relocation runs after both switches are verified (`move.go:87`). If it fails, it warns, names the convergent `sdlc claim` repair, and leaves the branch safely moved. `moveRelocation` is a narrow seam the test can make fail.
- **`worktreeHoldsBranch` fails safely** (`claimant.go:178-189`). A vanished checkout counts as not holding the branch. An unreadable one is an error, never treated as evidence that the branch left.
- **set-status refuses a foreign takeover even under `--force`,** and that is pinned by `TestStatusDecisionRecordsOrRefusesTheClaimant`.
- **The restart-survival test uses the real built binary** with the real host identity, so no impersonation seam is involved (`TestOwnershipSurvivesRestart`).

## 2. Critical
None.

## 3. Important

**I-1: relocation treats "not on the branch" as proof of a move** (ARCH-ORDER, ARCH-PURPOSE).
- `relocatable` (`cmd/sdlc/claimant.go:162`) and `issue.RelocationAllowed` let any worktree on the same machine take a working card. It only has to be on a branch named `<details stem>` while the recorded worktree is on some other branch. That is also true when the owner:
  - has claimed but not yet run start-plan (the branch never existed there), or
  - has temporarily switched to main or rest.
- **Scenario:** :1 runs `sdlc claim --issue N` on rest. Then :2 runs `git switch -c 000NNN-slug` followed by `sdlc claim --issue N`, and the claim succeeds and repoints ownership to :2.
- **Contract hit:** this breaks the Done-when "Repeated claims cannot take another workspace's working issue". It also means the gate's own refusal message invites people to do exactly this.
- **Fix sketch:** require positive evidence of a move instead of the absence of a contrary observation.
  - Before the network step, `sdlc move` writes a local relocation marker in the destination worktree (for example, git config `sdlc.relocated.<id>=<source worktree>`).
  - The claim repair and the gate hint accept only when that marker names the recorded worktree.
  - Add a negative test: claim in A, create the branch in B, run claim in B → refused.
  - The operator approved relocation in PQ-8, so record a Revision if they still prefer the weaker rule.

**I-2: the Plan and Verb-contract cells for adopt and set-status are not enumerated in tests, and they drift** (family `plan-code-contract-drift`; this is its 3rd finding).
- What's missing or drifting:
  - The plan's "one two-clone adopt race" test does not exist.
  - The adopt "table over status × claimant × flag" was not written. The decision lives in the IO function `adoptClaim` (`claim.go:561`) rather than in pure `claimDecision`, which is also an ARCH-PURE miss.
  - Adopt on a `blocked` card is never tested.
  - `set-status working` on an already-working card with no owner silently stamps an owner (`setstatus.go:144`, reached because `checkTransitionGuards` skips current==next). The Verb contract marks that cell "—" and the spec says adoption is operator-directed via `--adopt`.
- **Rule:** the Verb contract table in the plan is the source of truth. One table-driven test should enumerate every (situation × verb) cell and assert its outcome. Any cell the code handles differently becomes a Revision, not silent drift.
- **Fix:**
  - Lift the adopt decision into a pure function beside `claimDecision`.
  - Make `statusDecision` refuse working→working on a card with no owner, pointing to `--adopt`.
  - Write that table test and add the adopt race test.

**I-3: move's successful relocation and its no-owner branch are never exercised** (family `missing-branch-unit-test`; this is its 2nd finding).
- `TestRelocationAfterMoveAndRepair` covers two cases: the old worktree still holding the branch, and a stubbed failure followed by the claim repair.
- The success return of `moveRelocation` → `relocateClaimant` (`move.go:140`) and the Unknown return (`move.go:131`) never run.
- The plan's "real-git move test: claim in :1, `sdlc move :0`, claimant names :0, gates pass" is not delivered. The test performs the switches by hand instead.
- **Rule:** every return path of a new effectful function gets a fixture, with real `runMove` for at least the success path.
- **Fix:**
  - Drive `runMove` across a slot pair (or call the unstubbed `moveRelocation` after manual switches) and assert the owner moved.
  - Add an Unknown card case and assert it warns toward `--adopt` and writes nothing.

**I-4: the README doesn't mention ownership, `--adopt`, or the rebuild-on-landing flag day.**
- README.md:15-27 ("Concurrent issue work") describes claim and start-plan in a tracker repository but says nothing about ownership.
- Fix: add one sentence: a claim records the owning workspace; only that workspace continues the issue; legacy working cards need `sdlc claim --adopt`; every `sdlc` binary must be rebuilt when this lands.

## 4. Minor
- `cmd/sdlc/move.go` is not gofmt-clean: in the import block, `"io"` comes after `"path/filepath"`.
- **ARCH-DRY:** the same OWNERSHIP paragraph is copied into four help texts (start-plan, change-code, close, milestone-close). A shared `{{OWNERSHIP}}` placeholder (the help files already use `{{LIFECYCLE}}`) would keep them from drifting. Also, in change-code.md the paragraph sits between "…in any checkout:" and the numbered list it introduces.
- `requireIssueOwnership` (`claimant.go:154`) ignores a `relocatable` error and falls back to the plain "owned by" message, so the operator never sees why the relocation probe failed.
- **ARCH-CONSTRAINTS:** a tracker-era `close` now opens the tracker and takes a snapshot twice (the ownership check, then `prepareTrackerClose`). Passing the env or snapshot through would avoid the second fetch.
- Same-machine Foreign refusals point only to "coordinate / #278". When the owner is this same machine, the message could name the move path (start-plan in the owning slot, then `sdlc move`).

## 5. Test coverage notes
- **Gate coverage is good.** The mutation check is credible: there is one helper and four callers, and each subtest asserts both the refusal text and that the judge was not dispatched.
- **Owner passes close?** For close and milestone-close, "the owner passes" is not asserted in the ownership test; it relies on the `closeReady` fixtures elsewhere. That's acceptable.
- **Relocation:** the only success path tested is the claim repair; move's own success path is untested (see I-3).
- **No ordering or race coverage for adopt.** Adopt is a CAS on the card it read, and the planned two-clone race test would be the oracle for "a loser publishes nothing". Without it, the adopt CAS is checked under only a single ordering (family `test-oracle-single-interleaving`, noted here rather than raised separately).

## 6. Architecture notes
- **ARCH-DRY: pass**, apart from the duplicated help paragraphs. There is one gate helper and one `ownership()` used by claim, the gate and move.
- **ARCH-PURE: flag.** The adopt decision is embedded in `adoptClaim`'s IO (I-2). `statusDecision` stays pure, and `RelocationAllowed` is pure and tested.
- **ARCH-PURPOSE: flag.** I-1 weakens the core purpose: a working card must not be taken without evidence.
- **ARCH-MOCK: pass.** The tests use real git and the real tracker repository. The machine-ID seam is a package variable used only in-process, and the built-binary tests use the real identity.
- **ARCH-CONSTRAINTS: pass, with a minor.** Close does two snapshots, and move adds one network CAS only after the local move succeeds.
- **ARCH-SECURE: pass.** The raw machine ID is never published, and a malformed claimant on the card fails closed via `CardClaimant`.
- **ARCH-ORDER: flag.** Relocation infers an ownership transfer from a single observation that something is absent (I-1). The CAS ordering itself is sound.
- **ARCH-FUNERAL: pass.** The claimant is replaced in place on its card; nothing new is left behind. I-1's suggested marker would need a removal rule: delete it when the relocation completes.

## 7. Plan revision recommendations
- **Under M2:** record that the adopt decision lives outside `claimDecision`, and either deliver the table and race tests or record dropping them along with the reason.
- **Verb contract:** add the working→working-with-no-owner row for set-status (refuse toward `--adopt`).
- **Relocation:** record the evidence rule chosen for I-1 (a move-written marker, or an explicit operator acceptance of the weaker same-machine rule along with the "no start-plan yet" exposure).

```findings
findings:
  - id: new
    severity: Important
    family: absence-as-authority
    title: |
      Relocation reassigns a working card on any same-machine worktree whose owner merely is not on the issue branch
    detail: |
      relocatable/RelocationAllowed (cmd/sdlc/claimant.go:162) accept "recorded worktree does not hold the branch" as evidence of a move; that is also true right after claim (before start-plan) or when the owner switched away, so `git switch -c 000NNN-slug && sdlc claim` in another slot takes the card. Require positive move evidence (e.g. a local relocation marker written by sdlc move before its CAS) and add the negative test.
  - id: new
    severity: Important
    family: plan-code-contract-drift
    title: |
      Verb-contract cells for adopt/set-status untested and drifting (no adopt race/table; set-status working on unattributed working card silently adopts)
    detail: |
      3rd finding in this family. Rule: the plan's Verb contract table is the source; one table-driven test enumerates every situation x verb cell. Lift adopt's decision out of IO adoptClaim (claim.go:561) into a pure function, make statusDecision refuse working->working on an unattributed card toward --adopt (setstatus.go:144), add the planned two-clone adopt race.
  - id: new
    severity: Important
    family: missing-branch-unit-test
    title: |
      moveRelocation success path and Unknown branch never exercised; planned real move test not delivered
    detail: |
      2nd finding in this family. Rule: every return path of a new effectful function gets a fixture. TestRelocationAfterMoveAndRepair covers only the still-held refusal and a stubbed failure; the move.go:140 relocateClaimant success and move.go:131 Unknown return are unreached. Drive runMove (or unstubbed moveRelocation) to a success and an Unknown case.
  - id: new
    severity: Important
    family: docs-surface-gap
    title: |
      README Concurrent issue work omits claim ownership, --adopt and the rebuild flag day
    detail: |
      README.md:15-27 describes claim/start-plan in tracker repos; add one sentence on the owning workspace, --adopt for legacy cards, and rebuilding sdlc after landing.
  - id: new
    severity: Minor
    family: gofmt-clean
    title: |
      cmd/sdlc/move.go import block is not gofmt-sorted
  - id: new
    severity: Minor
    family: duplicated-help-prose
    title: |
      Identical OWNERSHIP paragraph copied into four helptexts; use a shared template placeholder
    detail: |
      ARCH-DRY; also the change-code.md paragraph splits "in any checkout:" from its numbered list.
  - id: new
    severity: Minor
    family: swallowed-probe-error
    title: |
      requireIssueOwnership discards relocatable's error and reports a plain Foreign refusal
  - id: new
    severity: Minor
    family: repeated-tracker-snapshot
    title: |
      Tracker-era close opens the tracker and snapshots twice (ownership gate, then prepareTrackerClose)
```
