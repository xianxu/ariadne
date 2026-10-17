---
id: 000418
status: done
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '8ac9a0f4f1d6c0c75d920e69fa21c1c14c671670' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T20:02:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: operator, spec: "1951640a", done: "575d4647"}
actual_hours: 3.61
---

# sdlc merge: a reused branch name matches the earlier milestone's MERGED PR too and refuses ('found 2')

## Problem

## Spec

Reported by pair:2 (pair#618 M2), 2026-10-16. `sdlc merge` refuses with: "need one exact matching PR for 000618-slots-...; found 2".

How it happened:
1. M1 landed from the issue branch as PR #380, now MERGED. `sdlc merge` then deleted the branch.
2. `start-plan` recreated the same branch name for M2, and M2's PR #387 is OPEN at 9917e65f.
3. PR matching by branch name finds both PRs and refuses.

Every multi-milestone issue that lands milestone by milestone hits this.

Rule: match only PRs whose head is this branch's current head, or that are OPEN. A MERGED or CLOSED PR whose head commit is not the branch head is history, not a candidate. Two OPEN PRs for one branch still refuse.

## Done when

- With a fake GitHub holding one MERGED PR (an older head) and one OPEN PR (the current head) for the same branch, `sdlc merge` picks the OPEN one; a test covers it.
- Two OPEN PRs for the branch still refuse, naming both.
- In the live pair#618 M2 check, `sdlc merge` selects OPEN PR #387 despite historical MERGED PR #380 sharing its branch and advances into the pre-merge local suite, with no manual gh merge. Pair:2 owns the eventual landing separately (TL acceptance revision below).

## Plan

- [x] Add fake-GitHub regressions through the real landing flow for completed older heads beside the open PR and ambiguous open PRs; confirm red.
- [x] Reuse one head-aware candidate rule in strict selection, interrupted integration and head-registration polling (ARCH-DRY); preserve identity validation and stale-open push guidance.
- [x] Run targeted tests, stage the binary for pair#618 live verification through the hub, record the full suite, then close and land.

## Log

### 2026-10-16
- 2026-10-16: closed — BR-1 corrected: current Done when now matches TL-accepted live PR selection, original criterion preserved in Revisions. Fake-GitHub regressions red then green across fresh, retry, recovery, delayed registration and queued landing; two OPEN refuse naming both. Corrected recorded suite passed 1525 SDLC tests plus 66 packages, then integrated-main recorded suite passed 1525 tests plus 66 packages in 493.2s (log sha256:5663b9463967). No code changed since that run. Live pair:2 selected OPEN PR387 beside historical MERGED PR380 and entered its suite; TL accepted this proof and directed continuation. Shared root restored to clean main; all seven binary revisions verified modified=false. Atlas and lessons updated.; review verdict: SHIP

### 2026-10-16 — bounded diagnosis and plan

selectIntegrationPR's recovery fallback and syncIntegrationPR call selectLandingPR with an empty head, so every old merged PR becomes live again. Strict selection also permits a matching open PR beside another stale open one. Use one pure candidate rule across those callers: preserve every OPEN candidate (even if stale), ignore completed mismatched heads, and refuse multiple candidates with their PR numbers. Strict selection retains its existing named push remedy for one stale open PR. CLOSED PRs remain ineligible under the existing policy. No new external seam, state or lock: selection operates on the immutable GitHub observation, while the existing landing line and expected-head merge protect effects (ARCH-PURE, ARCH-DRY). The absence of competing PRs cannot be a local lock proof because GitHub owns them; bounded structured evidence and the existing expected-head operation remain authoritative.

The existing landingFakeGH and real Git/bare-origin fixture exercise production merge, integration and cleanup. Tests own the scenario matrix. Post the small plan to the hub and self-approve under tl-worker quick-flow rules.

### 2026-10-16 — red reproduction

Targeted fake-GitHub regressions failed as expected (63.498s): old MERGED heads block fresh, interrupted and pushed-but-unregistered integration; the reused-branch queue fixture refuses after admission; two OPEN PRs fail to identify both numbers. CLOSED history controls pass. Applying the single candidate rule through the existing pure selection functions; no added state or external API.

TL approval receipt c1e7dd46-bed9-4d58-a8ec-486bdada872c also requires an explicit ariadne:0 grant before staging, because #417 shares the target. Request that grant after targeted verification; the TL coordinates pair:2's real landing.

### 2026-10-16 — targeted green

New fake-GitHub regressions and existing stale-head/interrupted-integration controls passed in 124.900s. The single pure candidate selector now retains all OPEN candidates, filters older completed heads against the known branch head, and identifies ambiguous PR numbers. Both recovery fallback and registration polling pass that head instead of broadening to a name-only match. Atlas updated. No new IO seam, state, lock or permission rule.

### 2026-10-16 — full-suite diagnostic expectation

The recorded suite on 13d40cf05375e9ad8ffbe9ce699d4fe76d64bf39 surfaced TestIntegrationPRSync/merged: the existing test expected "state changed" for a MERGED record at the older head. Under the approved #418 rule that record is historical, so zero candidates is the correct refusal. Updating that control to expect "found 0" and adding a MERGED-at-current-head control requiring the existing state-change refusal. No production code changes. The first suite continues through all shards and its failed attempt/log will be retained; a recorded run of the corrected test head will follow.

## Revisions

### 2026-10-16 — align polling diagnostics with head-qualified history

The old polling test assumed a name-only MERGED record remained eligible. The approved rule deliberately removes completed older heads. Split its test into historical-head (no candidate) and current-head (state change) controls, keeping refusal and no-merge assertions in both. This changes no acceptance criterion or production mechanism.

### 2026-10-16 — first suite retained; live stage and corrected run pending

The first recorded suite completed all 12 shards over 1,525 SDLC tests plus 66 packages (565.2s execution). Its sole failure was TestIntegrationPRSync/merged's obsolete diagnostic expectation; no unrelated failures were reported. Retained failed log: .git/sdlc/test-stats/failed/cab81ce40ebedccd0ad82da2e1e71892.log, sha256:76b1b467c262. This is a deterministic test-contract correction, not a flake. The split historical/current-head polling controls and pure candidate controls pass (0.433s). A new recorded suite on the corrected test commit is queued; production code is unchanged.

Live-check state: TL grant 07a351da-5846-47d9-93c3-9c96a9e04e71 authorizes shared ariadne:0 for #418. It is staged at 13d40cf05375e9ad8ffbe9ce699d4fe76d64bf39; manual make sdlc-build repaired the known #417 skip. bin/sdlc revision equals that Git HEAD with vcs.modified=false. The TL received proof (807ca3aa-46db-4415-83ac-489d321d3dd4, submitted) and pair:2 launched the staged binary's merge for PR #387. Await its result. The later change is test/documentation only; do not restage or interrupt the in-flight live check.

After pair#618 PR #387 lands, clear the stage, run make build sdlc-build in /Users/xianxu/workspace/ariadne, verify every bin/* revision equals its restored main HEAD with modified=false, and tell the TL ariadne:0 is free. This restoration is explicitly authorized; do not refresh main or change other slots. Source PR is #276. Close and land #418 only after live acceptance and a recorded green suite.

### 2026-10-16 — live PR selection confirmed; pair suite blocks landing

Verified submitted peer receipt 6d25a454-5e7d-401f-bd63-24495feeca50 from pair:2: the staged binary selected OPEN PR #387 despite historical MERGED PR #380 sharing its branch, advancing into the pre-merge suite where the prior binary refused with found 2. That suite failed on pair-side test causes (a timing bug and couchtty flake tracked as pair#637). There is no merge commit yet. Pair:2 will retry after its corrected recorded suite and explicitly requests keeping the #418 stage. This satisfies the live matching observation but not the issue's final live-landing acceptance criterion. Our corrected recorded suite remains queued.

### 2026-10-16 — corrected full suite green; stage checkout superseded

Recorded suite on 21951f780aeeffcf1f648c999e768086f4eb503a passed all 1,525 SDLC tests and 66 other packages across 12 shards: 656.7s execution after 4,150s queue wait. Exit 0; suite.log SHA256 prefix ade094b7676b. Log: .git/sdlc/local-checks/21951f780aeeffcf1f648c999e768086f4eb503a/suite.log. The first failed attempt remains retained and classified above.

Verified TL receipt ced37b04-2d55-4a2a-ad4c-e1d0a4974d4c: #416 superseded the shared :0 checkout without a grant. Read-only verification confirms its Git HEAD is e7a8db930a2b2bbffbb903cc4b4743ef693415e8 while bin/sdlc still has #418 revision 13d40cf05375e9ad8ffbe9ce699d4fe76d64bf39 and modified=false. Do not rebuild or restage until pair:2 lands. The TL now explicitly directs cleanup as checkout main, rebuild every bin/* from main, verify revisions, and report completion. PR #387 is still open.

### 2026-10-16 — integrate current main before close

Main gained #414 and #415 code while the live landing was blocked. Merged origin/main cleanly into the issue branch and pushed combined head 59b68b975178173b451f0e33337e70676a078201. No #418 production changes or conflict resolutions. Started a new recorded suite for that combined head before close; this is required code-integration verification, not a retry of the already green test head. The staged sdlc remains untouched for pair:2.

### 2026-10-16 — TL accepts live matching evidence and releases shared stage

Verified TL receipt b9130677-9b77-414b-80ab-413c7a9b41af: pair:2 now uses the TL's separate copy at /Users/xianxu/.local/share/ops-tl/sdlc-418, and the TL states our live evidence stands because pair:2 passed PR matching, directing release of :0 now and continuation of #418's suite and landing. This revises the live acceptance dependency: successful live PR selection is accepted for #418; the actual pair merge remains with pair:2 rather than blocking this fix on pair-side test failures. The original Done when records the intended full-flow scenario; no claim is made that PR #387 has merged yet.

Restored ariadne:0 with git checkout main and make build sdlc-build (no refresh). Clean Git HEAD f5927dfedbfb2e6dbc7d8c91ac56085732181b7b. All seven bin/* tools report exactly that vcs.revision and vcs.modified=false. Reported :0 FREE to the TL in submitted receipt 63e791d2-19f7-4815-9592-6bac431927dc. No further stage ownership is retained by this item.

### 2026-10-16 — combined-head verification green; entering close and landing

Recorded suite on integrated head 59b68b975178173b451f0e33337e70676a078201 passed all 1,525 SDLC tests and 66 other packages across 12 shards, exit 0 in 493.2s after 3,822s queue wait. Saved .git/sdlc/local-checks/59b68b975178173b451f0e33337e70676a078201/suite.log, sha256 prefix 5663b9463967. No code changed since that run. The verification work in the final Plan row is complete; its close/land clause is the workflow now being entered, with final outcome to be recorded by the gates. The original failed attempt, corrected green attempt, live matching proof, TL acceptance revision and all-bin restoration remain recorded above. PR #387 remains open at this checkpoint; do not claim its merge.

### 2026-10-16 — BR-1: synchronize the current acceptance contract

The original live criterion was: "pair#618 M2's PR #387 lands through sdlc merge with no manual gh merge." The TL's later acceptance (receipt b9130677-9b77-414b-80ab-413c7a9b41af) accepts successful live selection and directs this issue to continue independently of pair-side suite failures. BR-1 correctly found that the unchanged current criterion contradicted that revision. Updated the final Done when clause to the observable accepted check: OPEN #387 selected beside historical MERGED #380 and progression into the pre-merge suite. Pair:2 retains responsibility for eventual landing; its submitted receipt 6d25a454-5e7d-401f-bd63-24495feeca50 supplies the live observation. Original criterion and rationale are retained here. No production or test changes; full-suite evidence still applies. The first close did not finalize because the ledger requires explicit BR-1 disposition; next action is to rerun close with this correction.
