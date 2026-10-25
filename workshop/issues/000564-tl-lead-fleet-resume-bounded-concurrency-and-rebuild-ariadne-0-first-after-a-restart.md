---
id: 000564
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-24
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd59fb8ad95be0ee2162932218ad6961dc2095420' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T01:28:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ebecbae0", done: "66a225b1"}
---

# tl-lead fleet resume: bounded concurrency, and rebuild ariadne:0 first after a restart

## Problem

## Spec

From the 10-22 full-shutdown recovery (ops review: `session.side/2026-10-22-30-recovery-review.md`). Two tl-lead
procedure gaps:

1. **`tl-lead fleet resume --apply` is strictly serial,** waiting for each slot's terminal receipt (~1.5 min) before
   the next: 15 slots took ~25 minutes. Allow a small bounded concurrency (e.g. 3 at a time, configurable), keeping
   the per-slot terminal-outcome reporting and stopping on a failure class that suggests a shared cause.
2. **The post-restart procedure doesn't rebuild ariadne:0's tools first.** After a toolchain change (10-22: Go 1.27.1 →
   1.27.3) every slot's `sdlc` is still ariadne:0's old binary until the hub rebuilds it. Add "rebuild ariadne:0
   (`make tools`, verify `go version -m bin/sdlc`)" as the first post-restart step in tl-lead's restart/fleet-resume
   guidance, before the fleet notice.


**Operator refinement (10-22, night), with pair#830:** after a full shutdown the operator wakes only the asleep hubs;
each hub's fleet resume then wakes its own asleep workers **pending-work first, then the remaining asleep slots**
(using pair#830's asleep flag once it lands).


**Interface (TL, 10-23, from pair#830):** the asleep set comes from the existing `couch --recover-plan-from-sdlc` report:
each per-slot row gains `asleep_at` (omitempty), next to its sdlc claim, issue branch, unlanded and dirty evidence, so
one store read gives both the asleep set and the pending-work order. There is no `couch --wake` CLI (after a full
shutdown no live slot exists to call it); the operator's switcher key wakes the asleep hubs, and each woken hub runs
this resume for its own asleep workers.

## Done when

- Fleet resume with concurrency N resumes N slots at once, reports each outcome, and stops on repeated identical
  failures (test with a fake resume verb).
- tl-lead's restart guidance lists the ariadne:0 rebuild first, and startup is a hubs-only quiet hold until explicit operator-requested recovery.
- Only explicit valid asleep marks authorize worker resume; pending-work wave 1 finishes before remaining-asleep wave 2.
- Every hub reports its worker-wave outcomes and completion handoff. The machine steward alone resumes scheduled tasks after both waves complete on every hub, or explicit operator go; worker resume never enables automation/scheduling itself. The external scheduler part is linked as pair#874.

## Plan

Durable design: [fleet-resume plan](../plans/000564-tl-lead-fleet-resume-bounded-concurrency-and-rebuild-ariadne-0-first-after-a-restart-plan.md). Operator approval is recorded in Revisions; change-code remains the implementation gate.

- [x] Review/approve the options and recovery state table through the hub; operator approval and third-wave clarification recorded below.
- [x] Implement the explicit-asleep planner and bounded receipt-driven scheduler with adversarial tests.
- [x] Update restart guidance and qualify privately; enter the normal recorded pre-close, close and landing gates.

## Log

### 2026-10-22


### 2026-10-24 — recovery and design checkpoint

#580 landed before starting this issue (PR 425); no retained suite/merge remains here. Claimed #564, ran start-plan, and read current fleet/resume/receipt code and recovery guidance. #533 remains parked.

pair:6 confirmed pair#830 is pre-close, not landed. Its schema-1 report carries the optional RFC3339 marker at rows[].agent.rows[].asleep_at; omission cannot distinguish an old producer from an awake thread and authorizes no wake. Normal resume accepts a marked worker without a claim and clears the marker. The current hub-only wake verb does not wake workers. Peer source and tests were read without modifying Pair. Pair's native full-fleet test remains skipped pending pair#863; no native wake or live recovery was run.

The durable plan proposes default concurrency 3 per invocation, pending-phase completion before idle wake, and stopping new launches on the second matching stable receipt failure (immediately for uncertain/unclassified failures), joining calls already started. Shared-tool rebuild remains a steward procedure before the fleet notice. ARCH-DRY keeps one recovery snapshot/classifier; ARCH-RESOURCE keeps a permit for the whole call. Await operator review via the hub before change-code.

### 2026-10-24 — operator approval and scheduler ownership

ops:0 relayed operator approval (submitted receipt e44fc3b4-971d-4160-9bb4-269af41d0540): after a full shutdown only hubs start, automation off and workers asleep; hold until explicit operator-requested recovery. Wave 1 wakes pending-work asleep workers, wave 2 remaining asleep workers, wave 3 resumes scheduled tasks. Clarification 25d26f1d-9e80-483b-829f-7499e93e9f04 makes wave 3 machine-wide: each hub reports both worker waves complete; the machine steward acts only after all hubs report or the operator says go. No hub toggles it alone. Current help's first-operator-message automation toggle must move to that steward step.

Read Pair AGENTS.local.md (root MEMORY.md is absent), current automation owner and scheduler issue evidence without changing source. pair#737 is working/not landed; current Couch exposes per-hub automation controls but tl-lead owns no scheduler. Filed unclaimed pair#874 for the missing scheduler-side machine-wide third-wave contract, with dependencies pair#737 and ariadne#564; report receipt 2dc4144a-4d41-4b42-9715-9314f2cf60cb is submitted. Primary Pair was detached and creation refused before effects; filed from clean resting pair:1 with no checkout writes, status clean afterward. No live automation, scheduling, wake or recovery call was made.

The #564 revision is isolated at /Volumes/TestScratch/tmp/ariadne-slot3/ariadne564-plan.PDsTx9/repo. #594's retained merge (terminal 56512) still owns the durable slot checkout; it is first waiter behind #592. Do not switch that checkout or change its head. Once it is released, detach this private worktree before checking #564 out in the slot, then run change-code and its applicable design gates before any implementation. #589 native qualification remains held.


## Revisions

- 2026-10-24: Corrected the planning interface from the older top-level-marker/no-wake description to pair:6's verified nested asleep marker and hub-only wake semantics. Added explicit options, partial-failure/restart states, counted tests and private-conformance limits in the durable plan. This records a proposed design, not permission to execute fleet recovery.

2026-10-24 — Operator-approved quiet startup and global third wave replace the initial two-wave-only proposal. Keep worker execution scoped to the owning hub; add an explicit handoff and steward-only third wave, with Couch-owned missing part tracked as pair#874. Durable plan renamed canonically and pure/IO classifications and function-level adversarial strategies made explicit. No permission to run a live fleet recovery is implied.

### 2026-10-24 — read-only producer qualification preparation

While #589's fresh suite and mandatory close are retained, inspected the Pair producer without mutation or running a native fleet. Pair origin/main is Git-verified503ba31b9fadffa4e965105f805391eaccdc0af8. RecoverPlan schema1 exposes rows[].agent.rows[].asleep_at exactly as planned; TestRecoverPlanCarriesTheAsleepMark covers positive mark and absent plain-parked omission. However TestDisposableCouchFleetWake on that origin/main still unconditionally skips as blocked by pair#863, even though its landing was reported. This is an explicit native qualification gap, not a passing acceptance result; revisit the owning shared harness before this issue closes, or obtain the hub's approval of the private executable limit as planned. No live worker, automation or scheduler control was called. Existing old-implementation Resume/Candidates baseline passed5.888s before implementation; it is not proof of the new behavior. #589 must land before this issue's code starts.

Ops:0 clarified in verified submitted receiptfee9e7f5-e79d-4403-8e78-d3cae12045a6: pair#870, owned by pair:6 after pair#854, lifts the native wake skip. Do not wait for it to start this issue's code. Private conformance is approved for now; before close, rerun the native fleet test if #870 has landed, otherwise explicitly state that qualification limit. No duplicate filing or native matrix is needed. #589's landing still precedes this issue's code under the dispatch order.

### 2026-10-24 — approved external wave-three seam

Verified submitted ops:0 receipt dd147d9c-989b-4a5b-9420-412cbbdf3ce6. Pair#874's approved plan now defines fresh-Couch scheduling hold and a generation-bound main-hub couch --schedule resume, with all-handoffs or operator-go reason and status reconciliation for lost replies. Per-hub automation stays a separate veto. Updated the durable plan's current seam and appended its Revision: this command only completes worker waves 1 and 2 and reports to the steward; it never releases scheduling. Pair#874 waits on pair#737, and its landing is explicitly not a dependency of #564 close. No live scheduling or recovery call made. #589's central merge runner remains retained in the root checkout; no #564 code started.

### 2026-10-24 — implementation gate and first red control

#589 landed as PR429; Git independently confirmed merge865ad6462c4740e3027cc3023b80876ceaf96cd4 on origin/main and the archive. Its branch and integrated full suites passed2,492 SDLC tests plus78 other packages, no failures/reruns. The first hub send was refused before admission during a Couch restart; the renewed landing report81733db4-df46-4c8a-aa1f-48b7f6295263 is submitted. Detached the private plan worktree, checked #564 into this owned slot, and integrated current main cleanly. No retained suite, merge or reviewer remains.

Ran tl-worker help, arch-principles and change-code --issue564 --worktree=no. The binary inferred quick from the current design size and accepted the checkpoint; it owns any full-review upgrade when the implementation exceeds the shell. No operator flow pin or gate bypass was invented. TestFleetResumeDoesNotInferAsleepFromClaims first failed on the existing behavior: it attempted pair:2 and pair:6 without any positive asleep mark. An initial misspelled timeout flag did not execute the test and is not red evidence; corrected count1 run produced the expected behavioral failure.

The existing strict JSON authority is currently cmd/sdlc/internal/jsonstrict, whose Go visibility excludes tl-lead. Reuse it by moving implementation and its existing tests to cmd/internal/jsonstrict and retaining the former package as a forwarding compatibility seam. This is extraction of the same validator, with no scanner fork, key-semantics change or new format. Add decoder/admission controls next, then the approved bounded scheduler; no live recovery or scheduling controls are authorized.

### 2026-10-24 — recovery projection and admission unit

Implemented the typed schema-1 recovery projection and wired fleetResume to one Couch recovery snapshot instead of the SDLC inventory. Explicit nested nonzero RFC3339 asleep mark plus one parked thread is required; all slot-zero hubs and live actors stay excluded. Missing marks are visibly skipped (not an assertion of awake/idle). Invalid/ambiguous marks or unavailable pending evidence remain unknown. Incomplete proven-candidate replies never admit. One shared validator now lives at cmd/internal/jsonstrict; SDLC's former import path forwards to it, and the existing validator tests moved unchanged. No new recursive scanner or format writer exists.

Focused recovery/envelope/planner tests passed0.890s, existing resume/receipt family passed2.904s after deliberately updating the incomplete-candidate and hub-exclusion expectations, and report-adapter controls passed0.697s. The latter uses only a private executable/payload, exact producer argv and cwd, bounded oversized output and old-provider refusal; no live Couch report or recovery was driven. The adapter reuses existing owned-group teardown, two-minute process budget,2-MiB stdout and bounded stderr. Counted8/16 fixtures assert one recovery capture, one actor capture, N candidate reads, and no SDLC inventory seam. Staged tracked-Go catalog passed1.567s.

This is an implementation checkpoint, not completion: execution is still the previous serial loop. Next write the pure batch transition red tests and bounded stateful-fake barriers, then replace that loop with pending/remaining waves, failure-key stop and joined cancellation. Guidance, private executable conformance, any available native harness qualification, recorded pre-close and full close/landing still remain. No own background run is retained at this checkpoint.

### 2026-10-24 — bounded worker-wave coordinator

Implemented ResumeBatch with one transition owner for admission, result and cancellation, a permit held through terminal outcome, pending/remaining phase barrier, and exact untouched accounting. The effect shell drains ready results before requesting one new launch, stops dispatch on caller cancellation, uncertainty/unclassified failure or the second matching (Status,Code,Diagnostic), and joins every started goroutine before returning. Distinct stable failures may continue but still yield a nonzero incomplete handoff. Unknown/tied owned evidence blocks the completion handoff; --all-only unclaimed rows remain excluded. No scheduler/automation calls are added.

The receipt identity regression failed first because refusal codes were flattened into text; structured failures now retain them. The CLI regression failed first on missing --concurrency; controlled barriers now prove N2 and the wave barrier, with N1/N3 core controls, out-of-order/duplicate results, repeated/distinct codes and joined caller/uncertain/unclassified stops. Additional observed red controls prevent a local pipe error from proving success and missing/null candidate completeness from authorizing resume. Candidate reads reuse bounded owned-group IO and stop future reads after observed caller cancellation. Command output buffers are bounded; truncated proof stays uncertain.

One fixture hang was diagnosed from the existing3-minute stack: channel barriers were born outside synctest's bubble, so Wait could not consider them durably blocked. Moving their creation inside the bubble made the same check pass0.258s, without a timeout change. I also started a broader targeted check while that first run was still active; stopped only that second owned tool session immediately, kept the bounded original for diagnosis, then resumed one-at-a-time verification. That interrupted duplicate is not test evidence.

Full targeted recovery/resume/model/receipt family passed7.815s, then the same focused race family passed7.663s, count1 only. The four runtime stop/join cases independently passed0.103s. No own test/review/suite/merge remains active now. Execution code is implemented, but this is not code-complete: update the dry-run wave preview and help/README/atlas/phrase guards, extend isolated real-Couch conformance, check the owning native fleet harness availability, then normal recorded pre-close and full review/landing.

### 2026-10-24 — guidance and executable qualification checkpoint

Updated dry-run wave preview, help, README and atlas with bounded worker waves, rebuild-first quiet startup and steward-only scheduling handoff. Root recovery source coverage now refuses absent/unavailable fleet or Couch evidence before dispatch; its missing-metadata control failed first. Scope output explicitly does not observe or change scheduler state. Full tl-lead package passed87.141s and shared strict JSON tests passed0.142s. The new opt-in stateful fake executable fixture compiles and the batch controls passed0.223s; opt-in execution is still pending. It uses a private backend and no live broker, workers or scheduler. Next qualify cancellation through the actual private CLI, then real-Couch receipt codes. No full suite or merge is running.

### 2026-10-24 — actual CLI cancellation regression

Clean private tl-lead build at Git e2fddd0469aaff5d8b11217f32d8461334f8730d independently matched build metadata with modified=false. The opt-in codex/signal case failed as intended: SIGTERM terminated the CLI while its two admitted calls remained unjoined. Its fixture released and observed its owned fake processes; no live Couch operation was used. Added a signal context scoped to fleet resume so caller SIGINT/SIGTERM reaches the existing stop-dispatch/join path, without canceling admitted remote work. Extend the real-Couch isolated receipt fixture to verify Code/Diagnostic identity before accepting qualification.

### 2026-10-24 — isolated executable and receipt qualification

Clean private build at Git60bc2ed36f5eef4cddfc99cf6316a7f83af6ca0a matched its embedded revision with modified=false. All six actual tl-lead executable cases passed6.23s: Claude/Codex/Qoder-shaped success, repeated stable refusal, lost reply after applied effect, and real SIGTERM. They prove N2 lifetime permits, pending-before-remaining, stale/live and absent-mark exclusion, explicit replay without duplicate resumes, partial outcomes and joined admitted calls. The stateful fake is a private executable/backend, not a production Couch or agent qualification. No automation/scheduling call is allowed by its argv authority.

The separate real deployed Couch CLI fixture initially refused because its isolation socket now belongs under the selected root, and the inherited t.TempDir path exceeded103 bytes. Read the current producer transport; updated only the fixture to a short canonical caller-owned root and its isolated m/ socket. A long-TMPDIR control covers the bound. Corrected check passed3.383s: real queued/running/succeeded, refused and failed receipts preserve normal exits and stable Code/Diagnostic identity. A transient compile typo in the named return was corrected before execution and is not behavioral evidence.

Fresh Pair tracker read at2026-10-25T01:14:57Z still shows pair#870 open/not-landed, and TestDisposableCouchFleetWake remains unconditionally skipped behind the known producer gap. Per the hub's already-approved qualification boundary, native fleet wake is not claimed; private executable and real CLI receipt conformance are the delivered proof. Pair#874 remains a documented external seam, not this close's dependency. Next final package/race/catalog checks, recorded pre-close, mandatory boundary review and landing. No live recovery or scheduler command was run.

### 2026-10-24 — signal registration ownership control

The final package run exposed a real fixture/runtime interaction after89.106s: Go's process-global signal initialization touched a synctest-owned channel from outside its bubble. Do not hide it with a timeout or remove the executable SIGTERM qualification. Move signal-source construction into main's injected resume-context factory, invoked only around fleet resume; in-process deterministic callers retain their supplied context. The scoped-lifecycle control and same concurrent-wave/model family now pass0.608s. The production executable still installs SIGINT/SIGTERM and must pass its actual CLI qualification after rebuilding. No remote-call cancellation or new supervisor is added.

### 2026-10-24 — final targeted verification

Final clean private executable at Git6f1f57d4160f2a4acf7beebc98f4beb0c59d1a3a matched embedded revision and modified=false. With opt-in executable/real-Couch conformance enabled, shared strict JSON and the entire tl-lead package passed (0.093s and141.559s respectively); no old-path wrapper tests remain because their unchanged tests moved with the implementation. Focused recovery/model/coordinator/receipt family passed under race6.366s. The catalog first flagged an unused suppression comment in the short-root fixture; removing its directive while keeping the ordinary explanation resolved it. Final catalog passed1,713 tracked Go files against19 rules2.502s; strengthened duplicate-address coverage now uses a complete valid source envelope, and that/short-root controls passed0.198s. No behavior changed after the final package/race proof, only that strengthened test and explanatory comments.

Private/native limits remain as recorded above. The complete repository suite still needs its normal recorded queued run; mandatory close review and landing have not run yet. All owned targeted runners are finished.

### 2026-10-24 — recorded pre-close queued; boundary review entry

Central sdlc local-check run --pre-close selected suite (trains off, no affected declaration). It owns run lease-2983636541 at exact Git85f6faa4818a305ac990819548fbaa090fb0f83d, detached checkout /Volumes/TestScratch/local-checks/sdlc-local-3508870854; raw inflight log is commonGit/sdlc/test-stats/inflight/af1ccfba336e806ae987cc523e3fcb8c/run.log. It is queued for both machine slots, initially position6, not a passing full suite. Keep this waiter unchanged. Enter mandatory boundary review on the completed targeted/private evidence while the recorded suite waits, with this limit explicit. The eventual merge still requires a verified passing suite; no pre-close pass, codecomplete or landing is claimed here.

### 2026-10-24 — close blocked on new measurement size bound

The close runner ended before review with exact refusal: measurement branch proof: measurement snapshot is 232727 bytes, over the version 4 bound of 131072 (72 sources, 1 branch records, 1296 branch intervals). No measurement window was printed. No actual was adopted and no review or codecomplete was claimed. The earlier64-KiB waiver ended after#593; the still-standing128-source waiver is a different condition and does not authorize bypassing this size bound. Asked the hub via receipt b559ec75-c6fd-4527-a143-f87ede99c34b whether precise --no-actual is authorized or close should wait for#603. Preserve that one queued message and inspect its receipt rather than resend. The existing recorded suite lease-2983636541 is still queued unchanged.

### 2026-10-24 — precise measurement waiver authorized

Verified submitted ops:0 receipt72405515-d6b7-44fd-af1b-98fefa2545ba: until ariadne#603 lands, the v4 size refusal may close with precise --no-actual, exact refusal in --verified and citation to#603; never typed hours. Fresh tracker read at2026-10-25T01:31:19Z confirms#603 working/not-landed. Retry only the close with that precise authorization. All other gates remain in force and the recorded suite waiter remains unchanged.

### 2026-10-24 — review SHIP rejected by foreign-home-entry guard; safe rerun

The full reviewer returned SHIP/high confidence with no findings and independently passed package, focused race, six private executable cases and real-Couch receipts. SDLC correctly did not finalize: its filesystem guard detected new entry /Users/xianxu/couch-profile-871-restage-20261024 outside review directories. Close exited1, leaving the slot checkout clean; that verdict is not an accepted close. Stat showed birth18:38:12 PDT.

Hub provenance question15f3559f-2f81-42b5-abdd-fa0a0e18d46a was submitted. Verified submitted ops:0 reply37a216f7-a1a4-4543-b4bd-76d45a14d738 confirms it is the operator's COUCH_PROFILE_DIR output for tonight's combined check, started18:38; leave it untouched and rerun close. The parallel pair:7 questionf8fc9fc5-0f7d-4e01-b0e1-98e019a26c5d remained queued and needs no resend now that provenance is answered. No foreign directory was edited/removed and no guard bypass is used. The post-restart suite runner25050/lease-2983636541 remains alive and queued unchanged. Rerun the same full close with the already-authorized precise measurement waiver; no code change is required.

### 2026-10-24 — BR-1 root cause and red/green correction

The accepted round1 boundary ledger is REWORK with BR-1 Critical (recovery-evidence-integrity): absent marker returned before agent evidence validation, so an unreadable owned worker could falsely complete the handoff. Reproduced before production edits: the decoder table failed for empty/unavailable/unknown/unusable/orphaned/malformed/conflicting agent evidence, while positive legacy/known-state omission controls passed; the in-process CLI returned success for unavailable evidence. Independently the actual old clean private executable failed the new unavailable case1.793s, printing skipped plus a successful handoff.

One recoveryAgentProblem authority now validates agent quality/count/thread identity/state before marker interpretation and retains agent Error/thread Reason diagnostics. Known none/live/detached/parked/busy evidence must be complete and consistent; unresolved states are unknown. Marked rows retain parked/time validation without duplicating generic shape checks. Twenty decoder cases and the CLI handoff regression passed with existing recovery controls0.751s before the final redundancy simplification; qualify the final version again. Documented the same boundary in help and atlas, and appended the plan Revision.

Canceled only owned queued old-head run with sdlc local-check cancel --run lease-2983636541. Its terminal25050 ended1 with suite lock interrupted: context canceled before acquiring capacity; no suite ran, no failed test or flake is claimed. The corrected head needs its own new recorded pre-close run. Both review and old check are finished, and no background runner remains at this checkpoint.

### 2026-10-24 — corrected-head qualification complete; fresh gates next

Final recovery/coordinator/receipt/help family passed4.300s after the single agent-evidence authority and marker-only simplification; focused race passed5.351s. Built private tl-lead at Git e6cf6c1fb6da7a050796104274daea5f220e2699; its embedded revision independently matched with modified=false. All seven actual private executable cases passed5.58s, including unavailable agent evidence now returning nonzero, a named unknown diagnostic and incomplete handoff with zero candidate/resume calls. Real deployed Couch queued/running/terminal receipts again passed3.14s; combined conformance package8.835s. Catalog passed1,713 tracked Go files against19 rules1.664s. No live agent or scheduler was driven.

These are corrected-head targeted/private proofs. The earlier full tl-lead package proof predates BR-1 and is not presented as a current full-suite pass. No full module suite ran on the canceled old head. Push this checkpoint, start one fresh normal recorded pre-close run, then re-enter the required full review (BR-1 remains open until the binary disposes it). Native pair#870 remains the documented unavailable qualification boundary; recheck before close.
