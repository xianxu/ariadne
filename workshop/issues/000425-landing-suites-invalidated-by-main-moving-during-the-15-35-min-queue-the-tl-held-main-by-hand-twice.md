---
id: 000425
status: codecomplete
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7967170a34305d2de4ec5b92e31f75590f7adc47' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T06:17:40-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c8e01053", done: "6c610154"}
actual_hours: N/A
---

# Landing suites invalidated by main moving during the 15–35 min queue; the TL held main by hand twice

## Problem

## Spec

2026-10-16/17: twice the TL had to hold every other ariadne landing by hand. First for #417, whose integration retry was about to run out, then for #423, which ran its full suite three times because main moved under it. Each recorded suite waits 15–35 minutes in the machine test queue and runs 8–11 minutes. A landing that integrates main, then queues its suite, is therefore invalidated whenever another landing reaches main during that wait. With five ariadne slots landing in one night this happens repeatedly, and `sdlc merge` gives up after three integration attempts.

The landing line already serializes merges. The gap is that a landing's suite runs against a main that can still move.

Options to evaluate:
1. **The line holds main through the suite:** the line-holder's integrated head is the next thing allowed onto main. Other merges queue behind it, and nobody else's merge advances main while its suite runs. This is already true for `sdlc merge` landings; check the issue-publish path and any other main writers (`issue new`, publish, project ticks).
2. **Docs-only drift doesn't invalidate:** if main moved only in issue or plan prose (`workshop/`), integrate without rerunning the suite. The #417 landing already did this once.
3. **Measure first:** how many reruns over the last N landings came from code drift vs docs-only drift.

## Done when

- The measurement (option 3) is recorded, split by cause.
- With the chosen fix, a landing whose suite is running is not invalidated by another slot's landing; a test simulates two landings.
- Docs-only drift during a landing's suite does not force a rerun, if that option is chosen and justified by the measurement.

- Worker and TL guidance distinguish branch verification from authoritative integration: neither the hub nor workers manually chase newer main or hold main through pre-close suites; the existing freshness gate and real code-change verification remain in force.

## Plan

- [x] Measure the last ten successful landings, separating integration-only code drift, docs-only drift, branch changes/mixed cases and failed-run recovery.
- [x] Characterize the existing protected phase with a deterministic two-landing integration test in cmd/sdlc/landingqueue_test.go: hold the first landing's suite, enqueue the second, prove it cannot integrate/test/merge until admission, then prove each integrates current main and tests its own integrated tree. Reuse the real-Git fixture and stateful GitHub fake (extend its per-branch PR state as needed); use channels at the suite seam for ordering, not scheduler timing as an oracle.
- [x] Characterize docs-only advancement during the first landing's suite using the actual main-publication fixture, and assert the same recorded suite is inherited after integration. Exercise localRecordsOnAncestors with adversarial cumulative code/config/embedded-help changes to mechanically prevent unsafe inheritance. Reuse publishGateHasCodeSurface; add no second classifier or fingerprint.
- [x] Update cmd/tl-lead/help.md (LANDING AND SMOKE TESTS), cmd/tl-worker/help.md (TESTS/LANDING), cmd/sdlc/helptext/local-check.md and atlas/workflow/sdlc-binary.md to distinguish pre-close branch verification from authoritative current-main integration under the landing turn: retain the required pre-close recorded suite; after it passes, close that tested branch rather than manually chasing newer main with repeated pre-close suites. Branch code fixes still require their own tests and suite; the existing noon freshness gate still applies. On SHIP enter sdlc merge promptly, which integrates current main under admission and reruns only when the integrated code requires it. Do not hold other slots or issue filing by hand for ordinary prose drift.
- [x] Run focused queue/CI/integration tests, then the required recorded full suite. Stage a bounded native help check only with a fresh hub grant.

After verification: close through SDLC, land on SHIP and report verified hashes/PR.


## Log

### 2026-10-17
- 2026-10-17: closed — Post-review recorded suite 02d000f565da8adcdb52e1be66d98d53 passed on 82e1b8a1: 12 shards, 1552 SDLC tests, 66 packages, 487.289s; focused help contracts green and all three reject pre-feature guidance. BR-1 README passages synchronized; BR-2 worker/lead/local-check rendered checks added. Earlier landing tests passed including reviewer race check; staged native help passed, all seven tools restored. No actual: open #420 counts measured queue waiting as active work. Atlas updated; close tested branch, merge owns current-main integration.; review verdict: SHIP

- Measurement scope: last 10 successful local landing-line events (PRs 272,273,275,276,277,278,274,280,282,281), 25 completed recorded suites, from 62 retained attempt records. No unreadable attempt records or missing Git objects. Group attempts by nearest issue-prefixed first-parent commit; pin each issue's main comparison to its PR merge's first parent (today's main would incorrectly contain the issue's own commits). Compare successive attempts, identifying main-base advancement separately from branch code edits; prior failures are a separate cause, and mixed branch/main changes are not attributed solely to main. Classifier matches publishGateHasCodeSurface: tests/config/embedded cmd markdown count as code; ordinary documentation/process paths do not.

| Issue | Completed suites | Integration-only code drift repeats | Docs-only drift repeats | Branch changes/mixed | Previous failure |
|---|---:|---:|---:|---:|---:|
| #413 | 1 | 0 | 0 | 0 | 0 |
| #414 | 1 | 0 | 0 | 0 | 0 |
| #415 | 2 | 0 | 0 | 1 | 0 |
| #418 | 3 | 1 | 0 | 0 | 1 |
| #417 | 3 | 1 | 0 | 1 | 0 |
| #419 | 1 | 0 | 0 | 0 | 0 |
| #416 | 5 | 1 | 0 | 2 | 1 |
| #421 | 3 | 0 | 0 | 1 | 1 |
| #422 | 1 | 0 | 0 | 0 | 0 |
| #423 | 5 | 2 | 0 | 2 | 0 |

- Totals: 15 repeat runs = 5 integration-only code drift, 0 docs-only drift, 7 branch-change/mixed, 3 after failures. The 5 integration-only repeats consumed 46.197 minutes execution plus 231.472 minutes queue waiting. These costs describe observed repeats, not guaranteed recoverable savings. In all 5, the preceding passed suite was a pre-close local-check run, outside landing admission. Newly incorporated main code landed during the prior queue in 3 cases (#418, #416, first #423), during execution in 1 (second #423), and after the suite in 1 (#417); some bases were already stale before queueing as well. Seven sampled suites ran after landing admission; none has a demonstrated docs-only integration rerun in this sample.
- #423 specifically had five passing suites: two repeated solely for main code (#421 then #422), and two for BR-1/BR-2 branch fixes. Its two integration-only repeats account for 18.095 execution minutes and 56.038 queue minutes. The assertion that all five were invalidated by main would be incorrect.
- Existing boundaries: local-check never takes a landing turn (cmd/sdlc/localcheck_cmd.go:76–144); merge holds the turn from admission through its suite, server merge and archive (cmd/sdlc/landing.go:380–644). Existing first-parent ancestor reuse already accepts docs-only cumulative differences through publishGateHasCodeSurface (cmd/sdlc/cigate.go:480–535). Thus extending only main-publication exclusion inside merge does not protect the pre-close suites that produced the measured repeats.

## Revisions

### 2026-10-17 — proposed choice after measurement (awaiting TL approval)

Choose option 1's existing owned landing phase as the place that integrates current main and runs the authoritative integration suite; correct the pre-close workflow that repeatedly did that integration outside the line. Preserve option 2's existing conservative ancestor reuse and prove it under docs-only main publication. This is a proposed workflow/guidance correction plus adversarial integration coverage, not a new cross-command reservation.

Alternatives: extending publication exclusion to every main writer inside merge would not protect the five measured predecessor suites, because they ran outside merge. A new reservation from local-check through close/merge would protect more time, but needs a cross-command owner, handoff, cancellation, abandonment and death recovery; it would also freeze unrelated work through reviews. Docs-only equivalence alone cannot explain the five code-drift repeats and already exists. Deferring current-main integration to the already-owned merge phase uses its existing flock lifetime (ARCH-DRY/ARCH-ORDER), needs no daemon/lease state (ARCH-FUNERAL), and preserves the full-suite-before-close rule.

Bounds and non-goals: this does not promise zero suites per issue or exemption from real code integration. One branch suite plus one current-main integration suite may still be necessary; branch fixes and a freshness-cutoff rollover still require verification. Ordinary independent Git pushes are outside the cooperative landing line; publication CAS still guards lost updates, and existing integration revalidation/retry still guards unexpected main movement. Do not introduce a global main-write freeze or change the machine test queue. If the two-landing/docs-publication characterization uncovers a binary defect, bring its counterexample and any mechanism change back to the TL before implementing it. Test coverage uses existing portable real-Git fixtures and the stateful GitHub seam (ARCH-MOCK); matching/reuse stays pure with one classifier (ARCH-PURE/ARCH-SECURE). Capacity is unchanged; deterministic tests coordinate bounded actors through channels and finish all owned workers (ARCH-CONSTRAINTS). This addresses the measured class, rather than adding a lock around an unobserved phase (ARCH-PURPOSE).

### Measurement coverage notes

Only completed retained attempt records are counted; cancelled/unrecorded runs are not inferred. Main timing uses Git commit timestamps and the recorded queue/run boundary, so seconds near that boundary are approximate: the first #423 main merge is stamped two seconds before recorded execution start. Cause classification does not claim five avoidable runs: moving the final integration into merge may coalesce repeated pre-close chasing, but testing genuinely changed integrated code remains required.

| Issue | Attempt | Tested commit (Git-resolved) | Outcome | Inside landing |
|---|---|---|---|---|
| #413 | 93c674dedb3a38a1784c253884b2aa4b | 6786ebd9f3ed | 0 | no |
| #414 | dab34d0a3afe74a57350882bc99f2b9f | a86268283f77 | 0 | yes |
| #415 | e1f8ce54f93eff6d7b0028653471b5c7 | 614b8f11f0eb | 0 | no |
| #415 | e4d805ccd8e12c5659a5efbc4a50cacd | aac5272d09fc | 0 | yes |
| #418 | cab81ce40ebedccd0ad82da2e1e71892 | 13d40cf05375 | 2 | no |
| #418 | 6028057602809077155525210a46fb8f | 21951f780aee | 0 | no |
| #418 | c96402794185bbc20f4dbae48ea91d15 | 59b68b975178 | 0 | no |
| #417 | 0d38a0a9a5a38a2120a8a0fcbc597114 | a7f66f308e6a | 0 | no |
| #417 | ff5798bab69f4a3e4351dec76355f629 | 448595c97a12 | 0 | no |
| #417 | 204d36b2ec7ddce0a934a0e347b19637 | 049841de6625 | 0 | yes |
| #419 | da26e71c86c35f097925fc09821221c1 | c3b33d6f4226 | 0 | yes |
| #416 | d5ecb880e188a182590c4815cf749537 | 005587765f51 | 0 | no |
| #416 | 0bbae3af1b6112bb7cb13acbc505a1ca | 1b31e6fc1f28 | 0 | no |
| #416 | ef42644f9ff6b93a2aa217acc8a0dbc2 | e7a8db930a2b | 2 | no |
| #416 | c069e6e8e842f9cf532752c50aec50b7 | 6fe57a8f910e | 0 | no |
| #416 | 62f17897eb25d2bd2b4d72b8c6e39776 | 95218fb672b6 | 0 | yes |
| #421 | 0ca1aa1d35f23b0d32caa429bc6688a5 | d260ae08c4d6 | 2 | no |
| #421 | e52d0859a8f834165786745f28ad224b | 42c2184e109e | 0 | no |
| #421 | 58552a40d707579edcedc727290f8a78 | 753556e8a210 | 0 | yes |
| #422 | 8345ca54e8a1979369ce084aba1fe7d1 | 6f9593168e7f | 0 | yes |
| #423 | 26619cadc5bf3d0eec52a5e8bd03c4f0 | ac4f0623f608 | 0 | no |
| #423 | 2e0137236f6c98802b3328f0539ac922 | 0227eeb1ddb6 | 0 | no |
| #423 | c896f6a83d2fce9a92d66e0558871d4d | bfcfa3954220 | 0 | no |
| #423 | 566a8fde2ee05a22964093a6b59cc9da | 2cf29e61ee88 | 0 | no |
| #423 | b001e0f5ace18ddda5bab3e855d83785 | 5aa359483460 | 0 | no |

### 2026-10-17 — TL approval

TL approved the measured choice: retain the branch suite before close, close without chasing newer main, and leave authoritative current-main integration/suite to merge's existing turn. No new lease or freeze. Added the TL-side instruction in cmd/tl-lead/help.md: do not hold main by hand or direct pre-close integration. Quick flow admitted; production changes are embedded workflow guidance plus controlled characterization tests, not a new synchronization mechanism.

- Guidance implementation: worker TESTS/LANDING, lead LANDING AND SMOKE TESTS, local-check help, and landing atlas now state the approved phase ownership and exceptions. Existing rendered help-contract tests for tl-worker/tl-lead pass (0.132s/0.142s); no local-gates configuration or ops files changed. Concurrent behavior coverage is being added separately before any cmd/sdlc test run.

- Characterization confirms existing ownership: two genuine concurrent merge invocations share a real bare origin and stateful per-PR GitHub/CI fake; while the first suite subprocess is blocked at its readiness/release protocol, the second remains queued with unchanged branch heads and no early suite marker. After release, each landing runs once on an integrated head containing its required main. Real TrunkFile.UpdateMany/CAS prose publication during an active suite causes integration but no second suite; cumulative code/config/embedded-help controls reject record reuse. No binary defect was found, so no production synchronization was changed. Broader focused landing-line/ancestor/pre-close-record tests pass (109.726s); worker/lead rendered-help contracts also passed.

- TL-granted native help verification passed at staged da619b6fa6ec89eb8c94aebadedbf922ed4f82d6: tl-lead, tl-worker and sdlc local-check all render the approved phase ownership and conservative reuse rule. All seven tools matched that revision with vcs.modified=false. Cleared :0, rebuilt and verified all seven at clean main f5fba332587102796bd9920bbe150d439aac072c; hub notified free. Recorded suite f025289ea1136b36efefdd8e3cdb59a4 is queued on the implementation head; close will use that branch evidence rather than manually integrating newer main.

- Final branch verification: recorded suite f025289ea1136b36efefdd8e3cdb59a4 passed on da619b6fa6ec89eb8c94aebadedbf922ed4f82d6: 12 shards, 1,551 cmd/sdlc top-level tests and 66 other packages; 487.546 seconds execution after 1,334.178 seconds queued, log SHA256 92089774ebd95f62ec89d55e56b896dd7ad3abe88d4cdec05d36b88f0c6c42fe. Only issue evidence changed after that tested code head. Close omits calibration with --no-actual because #420 remains working and this session repeatedly polled a 22-minute queue; the known attribution defect would count that waiting as active work. No hours are guessed or manually subtracted.

### 2026-10-17 — close-review corrections (BR-1, BR-2)

Synchronize both README landing-line and pre-close-suite passages, and add rendered-help regression checks for worker, lead and local-check instructions. Checks preserve branch-suite → close → admitted merge, no manual chase/freeze, freshness and branch-fix exceptions, and conservative prose-only reuse. Correct the implementation reference from the planned landingqueue_test.go to landingeconomy_test.go: real suite subprocesses use bounded readiness/release files, with channels only for worker completion. These are coverage/documentation corrections to the approved design; no production synchronization changes.

- BR-1/BR-2 correction verification: all three new rendered-help checks passed, then all three failed against the pre-feature embedded help at 146796d8, then the broader help/flag contracts passed after restoring the approved guidance. README now documents both pre-close branch verification and owned integration, and lessons records the missing guidance-regression class. A new recorded suite will verify the additional tests before re-close.

- Post-review recorded suite 02d000f565da8adcdb52e1be66d98d53 passed on 82e1b8a10c6bd8713c0164801fbf2995b2d0565f: 12 shards, 1,552 cmd/sdlc tests and 66 other packages; 487.289 seconds execution after 2022.101 seconds queued. Log SHA256 5906d4e428457807e0918fae860dc07dc9b6df3b5c0068d55298c5a04530453b. No source changes after this pass. Re-close asks the ledger to dispose BR-1 (README) and BR-2 (rendered-help regression checks with demonstrated pre-feature RED).
