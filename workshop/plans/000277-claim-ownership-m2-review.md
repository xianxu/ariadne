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

---

## Re-review — 2026-10-01T14:56:04-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | e53c075bd9ae04f71535872f44d124abf2839d4b..8bb514f1a838ed5e96a4841e34ae25449321c7a9 |
| command | sdlc milestone-close --issue 277 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-01T14:56:04-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

# Review of ariadne#277 M2, round 2: FIX-THEN-SHIP. One Important finding: the atlas and plan still describe the old relocation rule

Round 2 fixed the round-1 problems. It closes BR-9's takeover hole with positive evidence: `sdlc move` now writes a record in the git common dir, and `RelocationAllowed` and `relocatable` both require a record whose source is the recorded owner and whose destination is this worktree. The adopt decision is now pure (`adoptDecision`). `statusDecision` refuses working→working on a card with no owner. A real `runMove` test covers both the success path and the no-owner path, and there is now a two-clone adopt race.

I ran the targeted tests and both packages pass. I also removed two fixes in scratch copies to see whether tests catch it:
- Disabling set-status's no-owner refusal turns `TestVerbContractTable` red.
- Deleting the record check from `RelocationAllowed` alone leaves `TestRelocationAfterMoveAndRepair` green, because `relocatable`'s own nil check still guards it. The pure `TestRelocationAllowed` "no move record" case covers that path instead, so the protection is doubled, not missing.

What keeps this from SHIP:
- **Important:** the BR-9 fix changed the relocation contract, but the atlas and the plan still describe the old "absence of the branch is enough" rule, and the new on-disk record family is documented in neither.
- **Minor:** the move record outlives a move that never happened.

## 1. Strengths
- **Two independent guards on positive evidence.** `relocatable` (`cmd/sdlc/claimant.go`) refuses when there is no record, and `issue.RelocationAllowed` also requires From/To to match. The pure table covers "no move record", "move from another worktree" and "move to another destination".
- **`relocation.go` writes down its own lifecycle (ARCH-FUNERAL)** and treats a malformed record as an error, never as "no move" (ARCH-SECURE).
- **`raceBuiltBinary` (`claimremote_test.go`)** pulls the pre-push barrier out into one helper that the claim race and the adopt race share (ARCH-DRY). The oracle accepts both legal loser outcomes.
- **`TestVerbContractTable`** checks every situation × verb cell against the pure decisions, so the plan's table is now enforced rather than restated.
- **One tracker read per close:** `requireCardOwnership` runs inside `prepareTrackerClose` (`closetracker.go:338`), and only milestone mode opens the tracker separately.

## 2. Critical
None.

## 3. Important
**I-1: the atlas and plan still describe the pre-BR-9 relocation rule, and the new relocation-record file family is undocumented.**

This is the 2nd finding in family `docs-surface-gap`. The rule that covers both instances: **when a fix round changes a contract, grep every restatement of that contract and update them in the same commit.** The restatements are atlas, README, help text, the plan's step text and verb table, and plan Revisions.

For this contract (search for `RelocationAllowed` and "relocat"), the stale or missing places are:
- **`atlas/workflow/issue-tracker.md:51-53`** says relocation needs only "the current checkout is on the issue branch and the recorded worktree no longer holds it". That is exactly the rule BR-9 showed to be a takeover hole. The atlas should say:
  - relocation requires `sdlc move`'s record at `<git-common-dir>/sdlc/relocations/<id>.json`;
  - when the record is written and when it is removed.
- **The plan's M2 "Relocation" step** still describes the pure table over machine, repository, branch and holder, with no record. The plan has no M2 `## Revisions` entry for:
  - the evidence rule (round 1 §7 asked for this);
  - set-status working→working on a card with no owner now refusing toward `--adopt`. The Verb contract table still shows "—" for that cell, while the code and `TestVerbContractTable` assert a refusal.

## 4. Minor
- **The move record outlives a move that never happened** (`move.go:72-79`, new family `effect-record-outlives-effect`, ARCH-FUNERAL).
  - The record is written before the first `git switch`. If that switch fails ("nothing was moved"), the function returns without calling `removeRelocation`.
  - The leftover record still names this move. If the owner's slot later leaves the branch and the destination checks it out by hand, a plain `sdlc claim` there relocates the card with no move behind it.
  - Fix: remove the record on that return. Keep it on the second-switch failure, where it is the intended repair evidence.
- **Leftover double blank line** in `cmd/sdlc/helptext/change-code.md:3-4` after "in any checkout:".
- `gofmt -l cmd/sdlc` flags only `reviewsidecar.go`, which is outside this window.

## 5. Test coverage notes
- BR-15's fix (the probe error is now wrapped into the refusal) has no fixture. Nothing writes a malformed relocation record and asserts the "checking whether it was moved here failed" message, so I dispose it as not-addressed. A malformed `relocations/<id>.json` plus a gate call would cover it, along with `readRelocation`'s malformed branch.
- A Foreign card with no move record on the same machine is pinned at the integration level, and the pure From/To mismatch cases are pinned at the unit level.
- The move tests rely on slots being linked worktrees that share a common dir, as the fixture builds them. That assumption should go in the atlas line for I-1.

## 6. Architecture notes
- **ARCH-DRY: pass.** The help text is single-sourced through `{{OWNERSHIP_GATE}}`, the race helper is shared, and one `ownership()` serves claim, the gate and move.
- **ARCH-PURE: pass.** `adoptDecision`, `statusDecision` and `RelocationAllowed` are pure. The IO is in `relocatable`, `relocation.go` and `adoptClaim`.
- **ARCH-PURPOSE: flag (I-1).** The behavior is delivered, but the documented contract still teaches the weaker rule.
- **ARCH-MOCK: pass.** Tests use real git and the real tracker repository. Identity is injected only in-process, and the built binary runs with the real host identity.
- **ARCH-CONSTRAINTS: pass.** Move's network step runs after the switches are verified, and close reads the tracker once.
- **ARCH-SECURE: pass.** The record is local and never published, a malformed record fails closed, and another machine is never probed.
- **ARCH-ORDER: pass, with one gap.** Relocation now needs positive evidence. The remaining gap is that a failed switch leaves stale evidence (Minor above).
- **ARCH-FUNERAL: pass, with the same Minor.** The record family is one file per issue with a removal path. The exception is the first-switch failure, which leaves the record behind.

## 7. Plan revision recommendations
Add an M2 round-1 `## Revisions` entry recording:
- **The relocation evidence rule:** the record at `<git-common-dir>/sdlc/relocations/<id>.json`, written by move before switching and removed on success, when relocation doesn't apply, or by the claim repair. `RelocationAllowed` gains a `*Relocation` parameter.
- **`adoptDecision`** as the pure adopt core.
- **The Verb contract cell** "working, Unknown × set-status → working" changing from "—" to "refuse (use --adopt)", plus the "blocked, Unknown/Foreign" reopen rows that `TestVerbContractTable` asserts.

```findings
dispose:
  - id: BR-9
    disposition: addressed
    note: |
      RelocationAllowed + relocatable require sdlc move's record naming from/to; negative integration test plus pure "no move record" case; red when the record check is removed.
  - id: BR-10
    disposition: addressed
    note: |
      adoptDecision is pure; TestVerbContractTable covers every cell (red with the working->working refusal disabled); TestAdoptRaceHasExactlyOneWinner added.
  - id: BR-11
    disposition: addressed
    note: |
      TestMoveRelocatesItsOwner drives real runMove to success; TestMoveLeavesAnUnattributedIssueUnknown covers the Unknown path; both pass.
  - id: BR-12
    disposition: addressed
    note: |
      README.md now states owner recording, --adopt, move carrying ownership and the rebuild flag day.
  - id: BR-13
    disposition: addressed
    note: |
      gofmt -l no longer lists move.go.
  - id: BR-14
    disposition: addressed
    note: |
      {{OWNERSHIP_GATE}} rendered from ownershipGateHelp in main.go renderLong; paragraph moved out of change-code's numbered list (a stray double blank line remains).
  - id: BR-15
    disposition: not-addressed
    note: |
      Code wraps the probe error, but no test reaches it (no malformed relocation record fixture); the fix is unverified.
  - id: BR-16
    disposition: addressed
    note: |
      Whole-issue close checks ownership inside prepareTrackerClose over the card it already reads; only milestone mode opens the tracker in computeClose.
findings:
  - id: new
    severity: Important
    family: docs-surface-gap
    title: |
      Atlas and plan still state the pre-BR-9 relocation rule; relocation record family undocumented
    detail: |
      2nd in family. Rule: a fix round that changes a contract greps and updates every restatement (atlas, README, help, plan step/table, Revisions) in the same commit. atlas/workflow/issue-tracker.md:51-53 omits the required move record and its <git-common-dir>/sdlc/relocations/<id>.json lifecycle; the plan has no M2 Revision for the evidence rule or the set-status working->working refusal the Verb contract table still shows as "-".
  - id: new
    severity: Minor
    family: effect-record-outlives-effect
    title: |
      sdlc move leaves its relocation record when the first switch fails and nothing moved
    detail: |
      move.go writes the record before switching; the "nothing was moved" return never removes it, leaving stale evidence that later authorizes a claim-relocation after hand switching. Remove it on that path (keep it on the second-switch failure, where it is the repair evidence).
```
