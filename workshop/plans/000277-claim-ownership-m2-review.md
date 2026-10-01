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

---

## Re-review — 2026-10-01T15:02:20-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | e53c075bd9ae04f71535872f44d124abf2839d4b..b368bb3fc5862a8fb94216ebb0112c962339c6cf |
| command | sdlc milestone-close --issue 277 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-01T15:02:20-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Round 3 closes all three open findings, and I confirmed each with tests rather than taking the commit messages at face value. BR-15 and BR-18 each have a regression test. I reverted each fix in a scratch copy and both tests went red:
- `TestGateSurfacesAnUnreadableMoveRecord` fails with the plain "owned by" refusal.
- `TestMoveFirstSwitchFailureLeavesNoRecord` fails because the record is left behind.

BR-17's atlas rewrite and plan Revision now describe the evidence rule and the record's lifecycle. The targeted ownership, move, relocation and verb-contract tests pass.

One Important gap remains, and it is the same family again: `sdlc claim --help` was never updated in M2. It doesn't list or describe `--adopt`. It also still says "a repeat claim by any other workspace is refused", but the relocation repair is a successful repeat claim from another worktree.

### 1. Strengths
- **Both regression tests go red without their fixes.** BR-15's test hits the "malformed relocation record" path through the real gate. BR-18's test makes the first switch fail for real by locking the index, not by stubbing.
- **The record lifecycle is right on every path.** `move.go:78-81` removes the record when nothing moved. The second-switch failure (`move.go:83-86`) keeps it on purpose, as evidence for the `git switch` + `sdlc claim` repair, and the atlas states this.
- **The Atlas Record bullets in `atlas/workflow/issue-tracker.md:51-67`** cover every write and removal path, plus the bound (ARCH-FUNERAL).
- **The rule behind BR-17 is now in `workshop/lessons.md`**, along with the "absence is not evidence" rule. Both are stated as classes, not as single instances.

### 2. Critical
None.

### 3. Important
**I-1: `cmd/sdlc/helptext/claim.md` doesn't document `--adopt` or the relocation repair** (family `docs-surface-gap`).
- `--adopt` is registered at `claim.go:78` but is missing from the FLAGS list (lines 43-48). The prose only names it as the target of a refusal; it never says what it does or that it refuses an issue that already has an owner.
- Lines 34-36 say "A repeat claim by any other workspace is refused". `claim.go:143-165` contradicts that: a repeat claim at a move destination, with a move record present, relocates and succeeds. `move.md` and the gate both point operators at exactly this command.

> **This is the 3rd finding in family `docs-surface-gap`.** Earlier rounds fixed instances (README in BR-12, atlas/plan in BR-17). The rule: **every flag a command registers appears in its rendered `--help`, and the help page of each verb whose behavior changes is part of the "every restatement" sweep in lessons.md.**
>
> The flag half can be enforced mechanically and should be: add a test next to `helpflags_test.go`. It should walk every subcommand's non-hidden `Flags()` with `VisitAll` and assert that `--<name>` appears in `renderLong(page)`. That catches this class in every verb, not just `claim`.
>
> The prose half (the relocation exception) can't be tested. It is covered by the existing lessons rule, applied to the helptext of every verb the diff touches: here `claim.go`, where `claim.md` was left out of the sweep.

### 4. Minor
- `workshop/plans/000277-claim-ownership-plan.md:87`: the "Superseded by Revision 'M2 review round 1'" note points to a heading that doesn't exist. The Revision is titled "M2 review rounds 1–2".
- The `cmd/sdlc/helptext/change-code.md:3-4` double blank line was noted last round and is still there.
- `gofmt -l` flags only `reviewsidecar.go`, which is outside this window.

### 5. Test coverage notes
- BR-15: `TestGateSurfacesAnUnreadableMoveRecord` covers `readRelocation`'s malformed branch and the wrapped gate error. It is red without the fix.
- BR-18: `TestMoveFirstSwitchFailureLeavesNoRecord` is red without the fix. One weakness in its oracle: a `writeRelocation` failure would also produce "nothing was moved" and no record, so the test could pass for the wrong reason. Asserting the error text starts with `switch ` would pin which path ran.
- No test enumerates registered flags against the help pages. I-1's fix adds one.

### 6. Architectural notes
- **ARCH-DRY: pass.** The ownership help is single-sourced, and `relocationPath` is the one place that knows the record layout.
- **ARCH-PURE: pass.** The decisions are pure. Record IO is confined to `relocation.go`.
- **ARCH-PURPOSE: flag (I-1).** The behavior is delivered, but claim's own help page, the first place an operator looks for `--adopt`, doesn't teach it.
- **ARCH-MOCK: pass.** Real git and real lock-induced failures, no stubs.
- **ARCH-CONSTRAINTS: pass.**
- **ARCH-SECURE: pass.** A malformed record fails closed and the error reaches the operator.
- **ARCH-ORDER: pass.** Every exit after the record is written has a defined disposition: removed when nothing moved, kept on the second-switch failure, removed on success or when relocation doesn't apply, removed by the claim repair.
- **ARCH-FUNERAL: pass.** BR-18 closed the remaining stale-record path.

### 7. Plan revision recommendations
- Fix the "Superseded by" pointer so it names "M2 review rounds 1–2".

```findings
dispose:
  - id: BR-15
    disposition: addressed
    note: |
      TestGateSurfacesAnUnreadableMoveRecord drives a malformed record through the gate; reverting the wrapped error in a scratch copy turns it red.
  - id: BR-17
    disposition: addressed
    note: |
      atlas/workflow/issue-tracker.md:51-67 states the record evidence rule and full lifecycle; plan has the M2 rounds 1-2 Revision, superseded note, and the updated Verb contract cell; lesson recorded.
  - id: BR-18
    disposition: addressed
    note: |
      move.go removes the record on the first-switch failure; TestMoveFirstSwitchFailureLeavesNoRecord is red with the removal reverted; second-switch failure intentionally keeps it.
findings:
  - id: new
    severity: Important
    family: docs-surface-gap
    title: |
      claim --help omits --adopt and still says any other workspace's repeat claim is refused, contradicting the relocation repair
    detail: |
      3rd in family. helptext/claim.md FLAGS lacks --adopt (registered at claim.go:78) and lines 34-36 contradict claim.go:143-165. Rule: every registered flag appears in its rendered help (enforce with a VisitAll-over-subcommands test beside helpflags_test.go), and the restatement sweep in lessons.md includes the help page of every verb whose behavior changed.
  - id: new
    severity: Minor
    family: plan-code-contract-drift
    title: |
      Plan superseded-note cites Revision "M2 review round 1" but the entry is titled "M2 review rounds 1-2"
```

---

## Re-review — 2026-10-01T15:20:37-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 277 — Record claimant ownership atomically with issue reservation |
| repo | ariadne |
| issue file | workshop/issues/000277-claim-ownership.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | e53c075bd9ae04f71535872f44d124abf2839d4b..8c2da75301d15d2be5148833d31d5052774c7ebd |
| command | sdlc milestone-close --issue 277 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-01T15:20:37-07:00 |
| verdict | SHIP |

## Review

BR-19, BR-20 and the targeted tests all check out, and I've confirmed that slots share one git common dir. Writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

The round-3 fixes hold up. `claim --help` now lists `--adopt`. Its repeat-claim paragraph now matches the relocation branch in `claim.go:144-165`, which refuses unless `sdlc move`'s record names this checkout as the destination. A new test, `TestEveryFlagAppearsInItsHelp` (`helpflags_test.go:53`), walks every subcommand from `buildRoot()`. It enforces the BR-19 rule across the whole family, not one page: it already caught the missing `--sandbox`/`--issues-dir`/`--plans-dir` on change-code and the missing `--plans-dir` on merge and push, and all three are fixed in this window. The base `claim.md` had a FLAGS section without `--adopt`, so this test would fail without the fix. The plan's superseded note now cites the correct Revision title. The targeted tests (`Ownership|Adopt|Relocat|VerbContract|SetStatus|Move|Claimant|GateFlag|EveryFlag`) and `internal/issue` pass. `gofmt -l` reports only `reviewsidecar.go`, which this window doesn't touch. Nothing blocks the boundary.

1. **Strengths**
   - `issue.RelocationAllowed` (`internal/issue/claimant.go:189`) now needs positive evidence: move's own `Relocation{From,To}` record. A branch merely being absent from the old worktree is no longer enough. Lessons.md records this rule.
   - `relocateAfterMove` (`move.go:105`) runs only after both switches are verified, and it tells apart "does not apply" (record removed) from "failed" (record kept, repair named). This bounds the effect cleanly.
   - `adoptDecision` is a pure compare-and-swap. The race test (`ownership_test.go:402`) shows a single winner, and the verb×situation table (`verbcontract_test.go`) pins the plan's contract cells against production decisions.
   - Moving set-status's claimant lookup into `decide(env, …)` (`setstatus.go:87`) removes the second tracker open without changing the decision's purity.
   - Relocation records live in the shared git common dir. That's correct for slot worktrees: I confirmed the slot's common dir is `/Users/xianxu/workspace/ariadne/.git`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `startplan.go:296` already holds `card` from a snapshot, then calls `requireIssueOwnership(env, id)`, which takes a second snapshot. `changecode.go:338` followed by `refreshMirror` does the same. **This is the 2nd finding in family `repeated-tracker-snapshot`.** The rule is: a verb that already holds a card judges ownership with `requireCardOwnership(env, card)`; `requireIssueOwnership` is only for callers that hold no card. Apply it to start-plan; change-code can share the snapshot with `refreshMirror`. As a side benefit, status and owner are then judged from the same card version.
   - `helptext/claim.md:39`: the relocation sentence is not re-wrapped, leaving one line of about 150 characters. Same for the new lessons.md bullet.

5. **Test coverage notes:** the help test, verb-contract table, adopt race, relocation success and failure paths, and restart persistence together cover the bug classes this diff could ship. There is no built-binary check that `set-status working` stamps the claimant through `runCardUpdate`. The pure test plus the unchanged seam are enough for this window.

6. **Architectural notes**
   - ARCH-DRY: pass. The `{{OWNERSHIP_GATE}}` placeholder single-sources the help text, and `requireCardOwnership` is shared.
   - ARCH-PURE: pass. The decisions live in `internal/issue` and `claimdecision.go`; the IO stays in `claimant.go` and `relocation.go`.
   - ARCH-PURPOSE: pass. The help-flag rule is enforced class-wide by the test.
   - ARCH-MOCK: pass. Tests use real git worktrees and tracker fixtures.
   - ARCH-CONSTRAINTS: pass, apart from the extra snapshot noted above.
   - ARCH-SECURE: pass. Machine IDs are fingerprinted, and a malformed relocation record is an error rather than being read as "no move".
   - ARCH-ORDER: pass. The record is written before the switch, removed if the first switch fails, and the relocation runs only after both switches are verified.
   - ARCH-FUNERAL: pass. `relocation.go`'s header gives the lifecycle and the size bound.
   - Upcoming work: reclaim (#278) should reuse `MatchClaimant`/`SetCardClaimant` and use the same CAS-on-read-card shape as `adoptClaim`.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-19
    disposition: addressed
    note: |
      claim.md now lists --adopt and states the move-record exception matching claim.go:144-165; TestEveryFlagAppearsInItsHelp walks all subcommands (would fail on base claim.md) and drove the change-code/merge/push FLAGS fixes; lessons.md sweep rule names verb help pages.
  - id: BR-20
    disposition: addressed
    note: |
      Plan superseded-note now cites Revision "M2 review rounds 1–2".
findings:
  - id: new
    severity: Minor
    family: repeated-tracker-snapshot
    title: |
      start-plan and change-code re-snapshot the tracker for the ownership gate while already holding the card
    detail: |
      2nd in family. startplan.go:296 holds card yet calls requireIssueOwnership (second Snapshot); changecode.go:338 snapshots for ownership then again in refreshMirror. Rule: a verb holding a card judges with requireCardOwnership(env, card); requireIssueOwnership only where no card is held — also judges status and owner on one card version.
```
