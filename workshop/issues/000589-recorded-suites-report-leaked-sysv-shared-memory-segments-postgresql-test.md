---
id: 000589
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-24
estimate_hours:
card_mirror: 'c4f647edebc1cf26fdf1a071b3aab01aa9793348' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T03:44:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "0a3b6838", done: "bf85bccb"}
---

# Recorded suites report leaked SysV shared-memory segments; PostgreSQL test harnesses must not leak them

## Problem

## Spec

**Leaked System V shared-memory segments fill the machine's limit and break PostgreSQL tests (echo:0, 10-23).** macOS
allows 32 SysV segments (`kern.sysv.shmmni`). On 10-23 31 were in use, every one with 0 attached processes and no
PostgreSQL running: leftovers from test PostgreSQL clusters killed without a clean shutdown (likely suites cancelled,
timed out, or killed by teardown). echo#26's `TestDaemonPostgresLeaseLossExits` then failed with
`shmget ... No space left on device`.

Fix, in two parts:
1. **The recorded suite notices leaks:** `sdlc local-check` records the slot owner's SysV segment count before and
   after a run, and a run that leaves new segments with no attached process reports them in its result (warning, with
   the ids), so the leak is pinned to the run that caused it.
2. **Guidance for harnesses that start PostgreSQL:** stop the cluster (or use a shared-memory-free configuration, such
   as `shared_memory_type=mmap` / `dynamic_shared_memory_type=posix`) in every exit path, including cancellation; the
   repos that start PostgreSQL (echo first) get their own issue.

## Done when

- A suite that leaks a segment gets a warning naming it in its recorded result (test with a fixture that creates and
  abandons a segment).
- The guidance is in the workbench docs, and echo has a filed issue for its harness.

## Plan

Durable design: [suite SysV warning plan](../plans/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test-plan.md). Approved by ops:0 on 2026-10-24; implementation gate completed.

- [x] Approve the warning, attribution, persistence and qualification boundaries.
- [x] Implement pure observations/delta and stateful-fake managed lifecycle tests.
- [x] Qualify native single-owned-segment behavior after operator-cleared capacity, document harness guidance and existing echo#32, and pass recorded pre-close.

Mandatory close review and landing follow these completed implementation and verification steps.

## Log

### 2026-10-23

- 2026-10-23 (ops:0, from echo#32's investigation): PostgreSQL always allocates one 56-byte SysV interlock segment,
  even with `shared_memory_type=mmap`; its id is on `postmaster.pid` line 7. A clean or SIGQUIT stop removes it; a
  SIGKILLed postmaster leaks it (that was the 31 of 32). A leak report can name the culprit: `ipcs -m -p` gives each
  segment's creator PID (CPID), the postmaster, so record the process name and the run with the ids. echo#32 stops
  clusters with SIGQUIT first, SIGKILL only after 5 s, and tests that each cluster's own segment is gone.


### 2026-10-24 — claimed and design checkpoint

Ran tl-worker help, claimed and started #589 from main after #591 landed. No own check/merge is running; #564 stays held, #590 is next after this issue. Read the managed resource/batch cleanup, attempt publication and saved-pass proof seams. Read echo local rules (no MEMORY.md found) and inspected echo#32: owned by echo:2, codecomplete/SHIP, not landed; no duplicate issue or peer change.

Read-only native ipcs shows32 current owner segments,56 bytes each, NATTCH0; most CTIME displays00:43:21–00:43:47. No segment was created, attached, removed or signalled. All32 slots are occupied, so no native allocation test runs now. Ops:0 approved honest not-attributed wording for concurrent observations unless the creator is witnessed in the suite's own descendants/group (receiptc3fee6b9-4430-4614-9f8c-9e5bb9dc2b57), and directed fakes first, native qualification after operator cleanup.

The proposed plan uses two admitted-run boundary snapshots, at most one optional terminal group witness, and a diagnostic footer in the existing digest-verified log. No new attempt schema, cleanup mechanism or failure state. ARCH-PURE separates parsing/delta from host IO; ARCH-DRY shares warning rendering between fresh and saved-result publication. The native monitor must never infer that every same-UID allocation came from the running suite.

## Revisions

- 2026-10-24: Corrected the planning claim that a before/after count pins the culprit: parallel suites share the UID, and exited creators cannot be named reliably. Keep observed/unattributed warnings explicit; name known creator PIDs/process witnesses only. Existing echo#32 satisfies the requested harness issue; mmap does not remove PostgreSQL's56-byte interlock.

### 2026-10-24 — resumed implementation

Retained change-code completed successfully and committed accepted design (Git state reports 9078882e). Previous recovery screen is unavailable; recovery used the issue, durable plan, Git state and retained process exit. Full plan approval receipt: 38e93c9b-c77c-42ff-9f67-fedbfc22b3af. No native IPC allocation or cleanup performed. Independently filed hub-requested staging-attribution follow-up #594 unclaimed; it does not change this scope.

### 2026-10-24 — first implementation checkpoint

Observed red then green for pure delta/native table parsing and bounded-reader fake controls. The managed fake abandoned-segment regression is green: exactly two snapshots and one terminal witness hook, warning in stderr and digest-verified saved log, successful status unchanged. Parent final summary overrides earlier child-looking summaries. Production local-check, resume, single merge, train and baseline constructors explicitly inject native observation; unconfigured unit-test dependencies do not probe live machine state. Native creator names/group/incarnations are read from one kernel record after the bounded terminal PID/group snapshot; exited or moved rows remain unknown. No real IPC allocated or removed. Native qualification request sent to hub, receipt 1f398835-ab92-43f2-a13c-9893367f0ce9 submitted. More adversarial lifecycle, saved-replay, limits, docs and native qualification remain before full suite/close.

### 2026-10-24 — adversarial controls and qualification boundary

Focused SysV managed-run controls passed normally and with race detection (21.766s): exact saved digest/replay and mutation refusal, old-log compatibility, Go failure-parser immunity, pass/failure/startup refusal/timeout/early batch failure, cancellation, observation unavailability, cleanup-unconfirmed ordering, one/eight command probe counts, admission/log refusal before observation, and VM exclusion. The stateful fake table is changed by the managed command itself. A new red control showed unknown snapshots printing zero counts; renderer now says owner-before/after unknown. Pure package race checks passed; Linux/amd64 package cross-compiled; tracked-Go catalog passed.

Read-only native capture still has32 unattached56-byte segments; saved as the Darwin parser incident fixture, with no IPC allocation or removal. Added an explicitly opt-in native child-allocation test, restricted cleanup by exact ID plus independent creator/UID verification; not run pending operator clearance. Added an opt-in clean private-executable conformance harness (fake IPC and GitHub; private queue/lock/scratch), still to run. Echo:2 guidance notification submitted (receipt e407d782-d9a9-41ee-a8ba-881ebc4fdbfb), linking existing echo#32.

### 2026-10-24 — private executable qualification passed

Built clean private executable /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-cli.K9QhoB/sdlc with Go1.27.3, embedded revision c3270bea8f7aec91aaeb2c9b826e50bf8d235be3, modified=false (go version -m verified). TestSysVPrivateCLIConformance passed in1.982s: actual CLI production run against a private stateful fake IPC table, fake GitHub, private queue/lock/scratch; warning preserved advisory success; deleting the fake IPC table before republish did not trigger a new probe and both published descriptions retained SysV warning. Latest catalog check passed0.905s. No real IPC allocated or removed. Recorded pre-close suite next; native one-owned-segment qualification remains blocked on operator-cleared capacity.

### 2026-10-24 — retained recorded suite and native hold

PR429 is open: https://github.com/xianxu/ariadne/pull/429 . Central sdlc local-check run --pre-close resolved to suite and is retained on pushed head d4575ac91, shell session31958, lease-355567331, attempt543d32f001f5bf36a8cb7d885ca80d4d. Detached checkout /Volumes/TestScratch/local-checks/sdlc-local-72563481; raw log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/inflight/543d32f001f5bf36a8cb7d885ca80d4d/run.log . Initially position2, behind ariadne-slot5; pair-slot1 and pair-slot7 hold capacity. Do not restart or cancel it. No close or merge runner started.

Hub outcome report receipt5c82e70a-9ebd-46a6-af30-6e9fd2254767 is submitted. Native allocation remains held until the operator clears capacity; no native IPC allocation or cleanup has run. Continue by polling retained session31958 and checking for the hub's native-qualification clearance. After native qualification and the pre-close pass, close through mandatory review, then land. #594 is unclaimed; #564 is still held for operator review; #590 remains queued. Previous TTY recovery capture was unavailable; issue/plan commits, Git state and retained exits were the recovery evidence.

### 2026-10-24 — original suite classified; context fix

Original attempt543d32f001f5bf36a8cb7d885ca80d4d ended normally with exit1 after27m22s (2464 cmd/sdlc tests and78 other packages), no retained runner remains. Two failure families: own deterministic TestVerbContextsReachTrackerReads identified fresh background contexts in captureSysV and the terminal witness hook; TestQueueRealHelperCompletion/test-lock-exit.test.py raced reading a mutable operation JSON record.

Preserved unchanged raw failure log under common Git sdlc/test-stats/evidence/000589/original-full-suite.log. Exact failed-head d4575ac91559a6a8d41d9f781f3e52dfb8465fe1 detached classification passed TestQueueRealHelperCompletion4.504s, HEAD and status unchanged before/after; evidence in queue-completion-classification.md alongside the original log. Filed #595 unclaimed and registered only the failing Python subtest with same-head fail/pass evidence. Ops:0 confirmed it is the same race already fixed by ariadne:5 in #592; that owner closes #595. No independent fixture fix here.

After preserving that classification, reused its private worktree /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-classify.3b7IeV/repo for this issue's branch, leaving #590's source checkout and running review untouched. A new stateful context control reproduced lost caller values at both snapshots and the witness hook. Each observation now derives WithoutCancel from the caller and adds its own existing two-second deadline. This retains values while keeping post-cancellation observation independently bounded. No added background-context exemption: the existing guard prevents recurrence.

Focused SysV lifecycle/context-guard checks passed5.331s plus pure parser/delta/native-fake package0.129s; race counterparts passed8.164s and1.151s. New context controls cover normal and canceled runs, value continuity, independent deadline and both snapshot/witness boundaries. Native allocation tests remain opt-in and skipped: no real IPC allocation/removal occurred and operator clearance is still required. The corrected branch needs a fresh pushed-head recorded pre-close; the previous failing full suite is not reused. #590's fresh pre-close and second close review remain retained separately.

Refreshed private executable proof on clean Git revision12a3aa07e72041451cf8a66582578a3329394dac: /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-fixed-cli.mO4tQg/sdlc, Go1.27.3, build metadata modified=false. The isolated real-CLI fake-table conformance passed1.909s (test1.55s), including advisory success and republish after IPC table removal. No live queue/lock/scratch or real IPC allocation was driven. Private worktree may return to detached state after this checkpoint; the issue branch remains pushed and the main slot stays on #590 during its boundary review.

### 2026-10-24 — corrected recorded pre-close retained

Returned this slot to the pushed corrected #589 branch only after #590's full review finalized SHIP. Fresh central pre-close on dbcf68d4a9037093235e0a62eb639a39a59d2f87 is retained in session61339, lease-1305896631, attempt1c1de399ced497b5a7d2b21050622b99, detached checkout /Volumes/TestScratch/local-checks/sdlc-local-723624371. Initial position3, behind #590's separate retained session14119 / lease-792232057. The same-check notice names two different issues, so preserve both waiters. No close/review for #589 until native qualification is cleared and completed.

A read-only ipcs capture at05:51:24 PDT still showed32 unattached segments; no allocation/removal or proof of ownership was inferred from those rows. Native qualification remains held for explicit operator clearance. #564 also stays held; ops:0 and pair:6 report pair#830's producer now landed in PR589, so the asleep_at interface is available once Couch is rebuilt from main. This is producer availability, not authorization to implement #564.

### 2026-10-24 — corrected recorded suite passed; native qualification held

Session61339 / lease-1305896631 finished exit0, attempt1c1de399ced497b5a7d2b21050622b99 on code-equivalent pushed head dbcf68d4a9037093235e0a62eb639a39a59d2f87. Eight shards over2465 cmd/sdlc tests and78 other packages, zero failures/reruns,19m02s after3726s capacity wait. Saved log: common Git sdlc/local-checks/dbcf68d4a9037093235e0a62eb639a39a59d2f87/suite.log, digest prefix9ff76340ffdd. Both hosted CI checks are also green. The context-guard defect is fixed and now passed the full suite; no blind full rerun or flake acceptance was needed.

Native one-owned-segment qualification remains held pending operator clearance. No real IPC allocation/removal was performed; do not close yet. When cleared, run the opt-in owned-segment qualification, preserve proof, then mandatory close review and normal central landing. #590 remains in its retained central merge session80121: current integrated Git head c294828f1e6524cc280910f0a950c75817efcd16, integrated CI passed, integration suite attempt6324e7c9f8ac91dd4de8aeafd78a020f / lease-2284728774 queued (currently position5). Its source checkout must remain untouched while that runner owns the turn. This Log update is isolated in the existing private worktree.

### 2026-10-24 — #590 landed; native-only next step

#590's retained central merge finished exit0, PR430. Git verified merge fb48cafb2bf86942fc31a4266624c47633bbe3e5 is on fetched main; its authoritative integration passed2470 command tests plus77 packages, zero failures/reruns,21m46s. No #590 suite/merge runner remains. Returned this source checkout to this #589 branch after landing; private classification worktree is detached and clean.

Read-only ipcs at08:29:59 PDT still showed all32 shared-memory segments unattached and occupied. Operator clearance has not arrived. This issue's recorded suite is already passed; do not repeat it, allocate native IPC, or remove existing segments while held. Exact next step after clearance: env WF_SYSV_QUALIFY=1 go test ./cmd/sdlc -run '^TestNativeOwnedSysVQualification$' -count=1 -v -timeout 3m; preserve its native owned-segment evidence, then mandatory close review and central merge. Current private executable conformance and full-suite evidence remain linked above. #564 remains held despite pair#830 producer availability. Previous TTY capture was unavailable; recovery evidence was Git, these issue Logs and retained command exits.

### 2026-10-24 — native command package correction

The immediately preceding command targeted the wrong package: TestNativeOwnedSysVQualification lives in cmd/sdlc/internal/sysvshm/native_unix_test.go, not cmd/sdlc. That old command would run no matching test and must not be used as qualification evidence. The authoritative next command, only after operator clearance, is: env WF_SYSV_QUALIFY=1 go test ./cmd/sdlc/internal/sysvshm -run '^TestNativeOwnedSysVQualification$' -count=1 -v -timeout 3m. Source lookup and a nonexecuting go test -list verified the exact test exists in that package. No native test or IPC allocation/removal ran. The full suite/private CLI proof are unchanged. Hub and pair:3 landing-report receipts are both submitted. No owned suite or merge remains running.

### 2026-10-24 — operator-cleared native qualification passed

Verified submitted ops:0 receipt b2809aee-c5d4-4c49-b816-91dcddc8f5c4 authorized the opt-in one-owned-segment qualification after the operator cleared capacity. In the existing private worktree, on Git-verified 1d1b4e2186d7463f6c8999fa78751cee47603590, ran:

    env WF_SYSV_QUALIFY=1 go test ./cmd/sdlc/internal/sysvshm -run '^TestNativeOwnedSysVQualification$' -count=1 -v -timeout 3m

Session19225 exited0: TestNativeOwnedSysVQualification PASS0.02s, package0.156s. Exact output: observed owned abandoned segment id=13893663 creator=9368; exact-ID cleanup registered. Read-only ipcs -m before (10:22:01 PDT) and after (10:22:36 PDT) both showed an empty table. The test verifies the child creator PID and UID before registering cleanup for only that returned ID; no foreign segment was removed. Raw evidence preserved under common Git sdlc/test-stats/evidence/000589/native-owned-qualification.txt. The corrected recorded full-suite pass and clean private CLI proof above remain valid; subsequent changes are documentation only.

#594 retains the main slot's central merge runner (session56512), still alive after the Couch restart; no second runner or checkout switch. Finish this issue's mandatory close review and central landing when the source slot is free. #564 is now operator-approved with a machine-wide third scheduling wave; its revised plan is pushed separately and external scheduler seam filed unclaimed as pair#874.

### 2026-10-24 — close measurement refusal

Close session50486 exited1 before review: measurement branch proof: measurement snapshot does not serialize: json: error calling MarshalJSON for type *gatestate.MeasurementInputs: json: error calling MarshalJSON for type *gatestate.MeasurementBranchEvidence: measurement exceeds 65536 bytes. No window was printed. Central sdlc reports revision945a8168578afcbd94ed27965c283d0a6483aeb1 modified=false. No review or completion generation was created. Ops:0 then supplied standing authorization in receipt88cf17d2-3cbb-4238-9fbd-14610da9c3e2 for this exact 64-KiB refusal until #593 lands: close with --no-actual, exact refusal and printed window if any; never typed hours. The next close uses that precise acknowledgment and unchanged passed verification.

### 2026-10-24 — noon freshness integration

Authorized no-actual close session46109 stopped at the noon freshness gate before review, naming main code through d71859146. Followed its explicit integrate-first instruction: merged origin/main cleanly (no conflicts), preserving the issue patch. Focused integrated SysV/context controls passed in cmd/sdlc9.353s and sysvshm0.125s, count1; no full run outside the queue. The old branch pass is historical; queue one fresh central pre-close on this integrated head, then close with the exact measurement refusal above and no typed hours. The mandatory review may overlap that queued run, but merge waits for both pass and final verdict.

### 2026-10-24 — close round1 REWORK; stop and re-plan

Close session61542 completed REWORK, no finalization. BR-1 advisory observation consumes command timeout; BR-2 current PID/start does not prove creator incarnation; BR-3 CTIME is mutable; BR-4 savedAttemptRecord was reused rather than modified; BR-5 resumed child-only warnings are omitted; BR-6 README lacks the new usage pointer. Verified the affected production paths before responding. Canceled only own still-queued suite lease-2371815346 (attempta86c82b2c9bafba095f03084c59fd64b); session76023 ended interrupted without starting a full suite, preserving machine capacity.

Appended one consolidated re-plan to the durable plan: isolate diagnostic time, remove unsupported optional creator attribution rather than add a monitor, compare observed numeric IDs conservatively, union verified parent/child warnings via shared saved proof, correct concept status and README. This is the third-mechanism stop rule, not permission to implement silently. Commit review evidence and re-plan, send to hub, and await approval before code; #564 remains next after this issue lands.

### 2026-10-24 — approved round1 corrections implemented

Verified submitted approvaldcf4a2bc-390c-41a0-b00b-4f93dcbcca96 covers the consolidated re-plan. Observed regressions red before fixes: CTIME changed existing ID became a candidate; completed success became timed out after diagnostics; current-PID name was falsely attributed; child-only warning was missing. Replaced the initial short-timer repro with a controlled execution-context factory in localRunDeps; it reproduced both timer-before-probe ordering and post-probe timeout on the old placement without relying on startup speed. Final controls preserve successful and exit7 commands when after-observation advances the completed clock. Genuine timeout/cancellation controls remain covered.

Removed WitnessGroup, process/kernel witness adapters and types; numeric CPIDs always remain unattributed. Delta compares numeric before IDs, not mutable metadata or guessed replacements. Added CTIME/size/creator/key/UID/attachment mutation controls. Shared resolvedSavedRecord now derives the child warning from its verified footer using the executed parent check name; a real fresh affected resume and subsequent saved replay both publish SysV warning, with only two captures total. Child log mutation still refuses proof. Current plan tables, README, help, atlas and lessons describe the reduced mechanism; savedAttemptRecord is existing/reused. No persisted schema or policy change.

Focused normal controls passed cmd/sdlc7.929s and sysvshm0.132s; expanded fresh/replay controls passed3.241s. Focused race controls passed cmd/sdlc22.156s and sysvshm1.218s. Tracked-Go catalog passed1.241s. All were targeted count1 runs, serial per slot. Native single-owned-segment qualification on the corrected working patch passed0.154s (test0.02s), observed ID196149249 / creator34542, and exact-ID cleanup completed. Independent ipcs -m was empty at13:28:18 and13:28:30 PDT; no foreign IPC removed. Refresh private executable conformance on a clean commit, then one newly recorded pre-close and close review; obsolete attempt a86c82b2c9bafba095f03084c59fd64b was canceled before admission.

### 2026-10-24 — corrected private executable proof

Built clean private executable /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-review-cli.W5inHq/sdlc from Git-verified9a999428f99397db4d5972cd00b29654705b9eb4; go version -m independently confirms Go1.27.3, same revision, modified=false. TestSysVPrivateCLIConformance passed2.206s (test1.76s): stateful fake allocation, advisory success, verified saved republish after fake IPC removal; isolated queue/lock/scratch and no live IPC during this executable test. Native corrected-patch proof above is separate.

Round1 dispositions for the next mandatory review: BR-1 controlled execution-clock regression and isolated budget/outcome fix; BR-2 numeric creator always unknown and witness machinery removed; BR-3 CTIME/metadata ID controls and conservative delta; BR-4 existing/reused concept correction; BR-5 child-only fresh affected resume plus saved replay, actual posted descriptions and child digest tamper refusal; BR-6 README pointer with matching help/atlas. New code requires one fresh recorded pre-close; review may overlap but landing waits for pass. No full run outside the SDLC queue.

### 2026-10-24 — round2 REWORK, complete outcome boundary re-plan

Close session36015 ended REWORK. BR-1/3/4/5/6 were independently verified addressed with reversal overlays; BR-2 production removal is sound but its replacement fixture stays green if the old optional renderer is restored. BR-7 shows caller cancellation during after-observation still turns a completed command into interrupted. Canceled only the obsolete still-queued suite attempt600eccc730443e406ad96fb76da71c81 via local-check cancel; session45275 exited interrupted before admission. No suite or reviewer remains.

Stopped under the third-delta rule and appended the durable class-wide re-plan: confirmed command/group completion retires the existing queue lease via RetireOutcome before after-probe, joining cancellation readers and freezing caller/signal/durable-queue decisions at the one existing lock boundary. Unconfirmed cleanup stays unknown. Strengthen BR-2 with an overlay-sensitive fixture that populates a restored former Witness field. Await hub approval before changing the completion lifecycle.

### 2026-10-24 — approved completion boundary implemented

Verified submitted ops:0 approval342f86e2-5c7d-4748-abb7-349fefe1fc3b. The new controlled matrix first failed on all late caller/queue exit0/7 cases, while before/during cancellation stayed interrupted. Registered the diagnostic defer first so it executes last, after all named-outcome cleanup, cancellation consumers and queue RetireOutcome/unknown abandonment; retained the original log descriptor through that final append. This also covers strict scratch/typed-owner cleanup without another outcome state or resource. Caller/queue matrix is green for ordinary and strict runs; unconfirmed cleanup remains unknown. Initial focused family green6.877s, expanded strict/unknown controls7.384s. Reused the shared testfix.DeadlineOnCue rather than maintaining a duplicate test clock.

BR-2 restoration proof retained at /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-unsafe-overlay.0Y620n: a private Go overlay restores the former Witness-bearing Candidate and exact old renderer, with no production file mutation. The strengthened same test failed with the expected unrelated-reused-pid attribution; the normal renderer stays green. Actual owned-subprocess SIGINT/SIGTERM controls during the final after-probe passed alongside deadline controls3.013s, proving default termination and no fabricated interruption/active lease. These subprocess fixtures allocate no native IPC or live queue work.

#599 was inspected through tracker: codecomplete, not landed. Asked owning ariadne:2 to report its landing; integrate main then and verify deadline/retirement composition before the next full check and close. No own full suite or review is running.

The final pre-integration cancellation/deadline/default-signal controls passed16.007s after adding assertions that the operation and local run lease are already absent inside the after-probe; the tracked-Go catalog passed1.534s. These are targeted count1 checks only. The new signal helper runs exclusively as an owned child process under a private temp/queue root and has no native IPC allocation. Next: focused race and existing signal/retirement controls, then integrate #599 once its owner confirms landing; refresh private executable proof on the composed clean head and run one recorded pre-close plus close round3. No full run is active or queued.

### 2026-10-24 — completion correction race verification; waiting for #599

Serial targeted race command (-p1, count1) passed sysvshm1.159s, queuegate1.214s, cmd/sdlc40.369s. Coverage included SysV lifecycle/deadline/retirement controls, fresh/saved resume warnings, existing real CLI signal boundaries, and TestQueueRetireOutcomeSerializesCompletionAndCancellation. No own test, review or full-suite waiter remains. Central landing status names #599 active; asked its owner ariadne:2 for the landing (submitted receipt22287b98-66e9-43d9-ad96-787332ae8da9).

Read its committed localcheck_run.go patch to prepare composition: #599 captures stoppedBy(runErr, runCtx.Err()) immediately after the batch, before its ran phase hook. On integration reuse that single deadline/outcome classifier, keep this issue's diagnostic as the last defer after all existing retirement/cleanup, and preserve the controlled execution-context seam. Resolve only after main includes #599; do not stack on its branch. Then run focused composition checks, catalog/private executable conformance, a single queued full pre-close and close round3.

### 2026-10-24 — #599 composed and verified

Verified submitted peer landing50dea274-daee-4a86-904f-891ec3f63d46 and hub instruction09947e6b-2fbf-4e7e-82d3-5c93a36c9e49. Git fetched main and verified merge2e7b1937fb7cad32c252829ea81e75f18d4e7e1a is its ancestor. Merged main; resolved the sole localcheck_run.go conflict using #599 stoppedBy as the execution deadline authority, retaining setup/ran hooks, controlled deadline seam, captured end timestamp and final diagnostic defer after all cleanup/retirement. Git-verified composed implementation53473347d328551949d051d3e3c8913ac019303f.

Combined normal controls passed sysvshm0.130s, queuegate0.197s and cmd/sdlc32.733s. Combined serial race controls passed1.151s,1.191s and47.136s respectively, including #599 setup/bookkeeping/real-overrun and graph/make stop semantics, SysV cancellation and strict cleanup, real signal boundaries and saved/fresh resume warnings. Tracked-Go catalog passed1.101s. Built clean private /Volumes/TestScratch/tmp/ariadne-slot3/ariadne589-composed-cli.xnU3pK/sdlc: Go1.27.3, build revision matches Git53473347d328551949d051d3e3c8913ac019303f, modified=false. TestSysVPrivateCLIConformance passed1.840s (test1.48s), stateful fake allocation, advisory success and digest-verified replay without IPC access; no live queue or native IPC driven by that test. Earlier exact-owned native allocation proof remains recorded separately.

Fresh tracker read still reports #593 codecomplete/not-landed, so ops:0's precise 64-KiB no-actual authorization remains applicable; exact prior refusal/no printed window is retained above. Queue one fresh recorded pre-close on the pushed composed head and close round3 alongside it. Do not enter merge before the check passes and review finalizes; do not chase main further unless a gate requires it.
