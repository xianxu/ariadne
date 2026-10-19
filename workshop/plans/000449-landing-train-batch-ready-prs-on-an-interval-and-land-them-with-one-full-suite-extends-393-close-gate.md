---
gate: boundary-review
issue: 449
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2b877597-7f69-4ea8-a79f-543a37ca2461
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/739c47422e6f2298
            reviewed_head: 4babdef455239df1cb2354f0ce56d58425d204f0
            slot: ariadne:2
            requested_at: "2026-10-19T08:40:06.232542Z"
          phase: finished
          pid: 33940
          started_at: "2026-10-19T08:40:06.463357Z"
          finished_at: "2026-10-19T08:43:52.200262Z"
          session_id: 416d5331-7349-441f-b806-74aadb741ffd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 4babdef455239df1cb2354f0ce56d58425d204f0
        tracker_head: 88f6c859a827b0d9846bc7a7ad980090f04151b4
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T01:38:30-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T01:38:30-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T01:38:30-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T01:38:30-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T01:38:30-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T01:38:30-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T01:38:30-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c2dced6d894b0673800cf19e706a3f64026ac9fe74ed688b8e8fdf161408d077
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 275411294
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T01:38:30-07:00"
      "n": 1
      timestamp: "2026-10-19T01:43:53-07:00"
      dispatched: "2026-10-19T01:40:06-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Data race on Handle.file between Vouch (vouch goroutine) and dropTicket (Wait/admission)
          detail: 'line.go:265 reads h.file in Vouch while line.go:383 writes it via tryAdmit; only an flock separates them, which the Go memory model does not see. A temporary probe test (Vouch in a loop alongside Wait) is flagged by go test -race. vouchWhileWaiting runs exactly this in production; it also prints a misleading "could not record readiness" warning after admission. Fix: a sync.Mutex on Handle covering file/dropTicket/Vouch, treat the post-admission error as a quiet stop, and keep the probe as a regression test.'
          family: shared-handle-unsynchronized
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Transition(Built) silently drops cars that are in neither Kept nor Conflicts
          detail: trainstate.go:181-188 removes such cars from n.Cars with no Hold/Release/Refuse effect, breaking the every-car-ends-once invariant. The random-history generator always partitions the cars, so the invariant test cannot reach this input. Reject a non-partitioning Built (return an error) or Release the unaccounted cars, and make the generator produce such inputs.
          family: state-machine-event-partition
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: Board builds Car.Branch from the sidecar's branch without checking it against the ticket's branch
          detail: train.go:293 uses r.Branch, and Readiness.Valid never compares it with Ticket.Branch. A hand-edited or buggy sidecar can steer the M2 conductor to a branch the waiting merge does not own, and the tip re-check passes because it reads the same untrusted branch. Refuse a mismatch in Board (add it to Why) or use t.Branch, and add a table test.
          family: untrusted-sidecar-cross-check
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: Plan Core concepts and state lists do not match the code (TrainState, Held, event names, DepartureDue signature)
          detail: The plan names TrainState, a Held state and the events CarConflict/CarGone/TipMoved/FlakeRerunPass/CIGreen/CIRed/ConductorDead, plus DepartureDue(now, lastDeparture, sched, line). The code has Train{Phase} with Reintegrating/Merging/Refused, the events Boarded/Built/SuiteDone/CIDone/Reintegrated/Probed, and DepartureDue(now, s, granted, events) reading next_due from the log. Add a Revisions entry and update the table.
          family: plan-core-concepts-drift
          round: 1
        - kind: defect
          id: BR-5
          severity: Important
          title: README update appears missing for the .sdlc/landing.json config keys and the train lines in landing status
          detail: train_interval_minutes and train_max_cars are a new user-facing config surface, and status already points users to the file; README.md is unchanged in the range. Add a short README note, or record the deferral to M3 explicitly in the plan's Revisions.
          family: docs-gate-readme
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: trainAdmission reads the whole events.jsonl once per live ticket per poll, inside the line mutex
          detail: Read the events log once per tryAdmit and pass it in.
          family: repeated-io-in-hot-loop
          round: 1
        - kind: defect
          id: BR-7
          severity: Minor
          title: The grant bypass is written in both trainAdmission and DepartureDue, and "main" has two definitions (main@{upstream} vs remote/main)
          detail: trainAdmission always passes granted=false, so DepartureDue's granted branch is dead in production. landingMainRef (status) and runDurableMerge (t.Remote+"/main") resolve main differently.
          family: duplicated-rule
          round: 1
        - kind: defect
          id: BR-8
          severity: Minor
          title: The schedule is read from an unfetched remote-tracking main before the wait and held for the whole wait
          detail: A missing remote/main ref now refuses even interval-0 merges, and a config change made during a long wait is not seen.
          family: stale-config-read
          round: 1
        - kind: defect
          id: BR-9
          severity: Minor
          title: Not-ready tickets skipped by Admissible are missing from OnAdmit's displaced audit
          detail: NextHead runs over the filtered candidates, so the tickets passed over are not recorded.
          family: audit-displaced-filtered
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d58ee59d-3e8d-4879-9140-3a040c2dea60
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e5305a206114e446
            reviewed_head: 452ccde50bfc1f2304db66a1530d11bfa83446bd
            slot: ariadne:2
            requested_at: "2026-10-19T09:09:30.136993Z"
          phase: finished
          pid: 14812
          started_at: "2026-10-19T09:09:30.243292Z"
          finished_at: "2026-10-19T09:57:26.154955Z"
          session_id: f7f523a8-fecb-4230-a343-b3686b0ea28c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 452ccde50bfc1f2304db66a1530d11bfa83446bd
        tracker_head: 88f6c859a827b0d9846bc7a7ad980090f04151b4
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T01:49:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T01:49:41-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T01:49:41-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T01:49:41-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T01:49:41-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T01:49:41-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T01:49:41-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 0db7edf3c7dd895a25811e1d37474f47517030e35e0b4d7e552f113e02b0fb8e
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 278526008
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T01:49:41-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T02:57:26-07:00"
      dispatched: "2026-10-19T02:09:30-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Handle.mu guards file/head; removing it from Vouch makes TestVouchConcurrentWithAdmission fail with DATA RACE under -race.
          round: 2
        - id: BR-2
          disposition: addressed
          note: partitions() rejects non-partitioning Built; reverting fails TestTrainRejectsIllegalEvents and the random-history invariant.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Readiness.For checks the branch and Board uses t.Branch; reverting fails TestBoard and TestTrainAdmission.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Plan Revisions entry M1 as built (BR-4) matches trainstate.go/train.go names, events, effects and the DepartureDue signature.
          round: 2
        - id: BR-5
          disposition: addressed
          note: README.md:491-498 documents train_interval_minutes, train_max_cars, the local-gates guard, and the status lines.
          round: 2
        - id: BR-6
          disposition: addressed
          note: Options.Admission is built once per attempt; TestTrainAdmission counts one events read across two tickets.
          round: 2
        - id: BR-7
          disposition: addressed
          note: The grant branch is gone from DepartureDue; one landingMainRef(git, remote) is used by both merge and status, and is tested.
          round: 2
        - id: BR-8
          disposition: addressed
          note: landingScheduleSource re-reads on each attempt and falls back to serial; TestLandingScheduleSourceFailsToSerial and TestVouchingStartsWhenTrainsTurnOn cover it.
          round: 2
        - id: BR-9
          disposition: addressed
          note: The audit walks the full order; reverting to NextHead's displaced fails TestAdmissibleGatesTheHead.
          round: 2
      findings:
        - kind: defect
          id: BR-10
          severity: Important
          title: Data race on the warned flag in landingScheduleSource, shared by the vouching goroutine and the admission poll
          detail: 'landingqueue.go:206-207. enterLandingLine passes one sched closure to both openLandingLine (Admission, Wait goroutine) and vouchWhileWaiting; a two-goroutine probe with an unreadable config is flagged by go test -race. This is the 2nd finding in this family. The rule: everything handed to both vouchWhileWaiting and Line.Wait must be safe for concurrent use. Swept: the line (flock and Handle.mu), readiness (vouch goroutine only), stderr (os.File) are fine; sched is the only remaining instance. Fix with sync.Once or atomic.Bool and a concurrent -race test.'
          family: shared-handle-unsynchronized
          round: 2
        - kind: defect
          id: BR-11
          severity: Minor
          title: landing status reads readiness sidecars without Readiness.For, so a vouch for another branch shows as ready
          detail: 'landingcmd.go:199-203 uses readinessLine and Current only. This is the 2nd finding in this family; the rule is that every sidecar reader goes through Readiness.For(ticket, now). Status also prints trains: unknown when the config is unreadable, while merges then land serially.'
          family: untrusted-sidecar-cross-check
          round: 2
        - kind: defect
          id: BR-12
          severity: Minor
          title: Every waiter runs two git processes each 2s poll under the line flock to re-read the schedule, even with trains off
          detail: trainAdmission calls sched() (git ls-tree and git show) inside tryAdmit's mutex on every attempt. Before this change the serial default ran none. This is the 2nd finding in this family; the rule is that admission inputs are read outside the mutex and slow-changing ones are cached (time limit or main's commit id).
          family: repeated-io-in-hot-loop
          round: 2
        - kind: defect
          id: BR-13
          severity: Minor
          title: The audit's passed-over rule is restated in line.go while NextHead's displaced return has no production caller
          detail: line.go:244-257 duplicates order.go:95-99. This is the 2nd finding in this family; the rule is one pure function in order.go computes who was passed over, and NextHead returns only the head.
          family: duplicated-rule
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: aa544094-7b48-4f49-a757-d30d2ca2d389
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/5c4c5ec244a460f9
            reviewed_head: 96d1a9d01bb99b19bac8f2f3c9fd9666e46142c6
            slot: ariadne:2
            requested_at: "2026-10-19T10:35:22.460561Z"
          phase: finished
          pid: 3970
          started_at: "2026-10-19T10:35:22.690705Z"
          finished_at: "2026-10-19T10:50:07.782477Z"
          session_id: ac6ce1ff-52cd-4ab8-95e5-9a0927bf0faa
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 96d1a9d01bb99b19bac8f2f3c9fd9666e46142c6
        tracker_head: ec71ed794e86f9299192aff52dd5d59f35c73c85
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T02:59:27-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T02:59:27-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T02:59:27-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T02:59:27-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T02:59:27-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T02:59:27-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T02:59:27-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 89064105774cb61d5ec254385202bf4f6b13420221f104edd303baba7829f5d4
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 281398554
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T02:59:27-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T02:09:30-07:00"
              finalized: "2026-10-19T02:57:26-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-19T03:50:39-07:00"
      dispatched: "2026-10-19T03:35:22-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-10
          disposition: addressed
          note: Mutex in landingScheduleSource; with the lock removed, TestLandingScheduleSourceIsSharedAndCached fails under -race with 3 DATA RACE reports and 4 reads.
          round: 3
        - id: BR-11
          disposition: addressed
          note: landingcmd.go:201-206 uses r.For(t, now); the test vouches with another ticket's branch and expects "not counted"; the unreadable-config line now says merges land serially.
          round: 3
        - id: BR-12
          disposition: addressed
          note: Admission() is now built before withMutex; the schedule is cached for 30s, so the serial default runs no git on most polls; the cache test pins one read per TTL.
          round: 3
        - id: BR-13
          disposition: addressed
          note: Passed() in order.go:96-112 is the only rule; NextHead returns (head, ok); line.go calls Passed; TestPassedIncludesTicketsThatWereNotAdmissible.
          round: 3
      findings:
        - kind: defect
          id: BR-14
          severity: Minor
          title: Admission reads the events log (the departure ledger) before the line mutex and head.lock, so M2 can admit a second train on a stale "due"
          detail: 'line.go:188-193 builds admissible from Events() outside withMutex. Once M2 records EventTrain and then releases head.lock, a waiter that read the events just before can take the head and depart early, skipping the interval. Keep the cached schedule outside the mutex, but read the events and decide DepartureDue under the mutex after head.lock is acquired (Events() does not take the mutex). Latent in M1: nothing writes EventTrain yet.'
          family: admission-decision-toctou
          round: 3
        - kind: defect
          id: BR-15
          severity: Minor
          title: The openLandingLine doc and plan Revisions still say the schedule is re-read on every admission attempt; the code caches it for 30s
          detail: 'landingqueue.go:39 and plan line 205. 2nd in this family. The rule: a commit that changes behavior described by the "M1 as built" Revisions entry or a function doc updates that text in the same commit; check those passages as part of every answer-review commit.'
          family: plan-core-concepts-drift
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 96d1a9d01bb99b19bac8f2f3c9fd9666e46142c6
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 67360e34-42b2-4c6f-b75d-384157d88394
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/728725127a45d5d1
            reviewed_head: d82c4f30c30bf529b0b3cce2ad5c26bb16d3e969
            slot: ariadne:2
            requested_at: "2026-10-19T11:58:25.229904Z"
          phase: finished
          pid: 17703
          started_at: "2026-10-19T11:58:25.344938Z"
          finished_at: "2026-10-19T12:02:37.196507Z"
          session_id: 4d4bb86f-80f9-42f3-a9c9-b765603e5060
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d82c4f30c30bf529b0b3cce2ad5c26bb16d3e969
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T04:31:03-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T04:31:03-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T04:31:03-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T04:31:03-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T04:31:03-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T04:31:03-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T04:31:03-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 336decb3b40c6140501ad2ecc9e0d9476c8cbd683d6623b1d5af1c1cdb803b08
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 292818926
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T04:31:03-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T02:09:30-07:00"
              finalized: "2026-10-19T02:57:26-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T03:35:22-07:00"
              finalized: "2026-10-19T03:50:39-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-19T05:02:37-07:00"
      dispatched: "2026-10-19T04:58:25-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-16
          severity: Critical
          title: With trains on, a waiting car whose PR is already MERGED can never be admitted or settled and hangs
          detail: 'carReadiness (landingcar.go:56) returns not ready for a merged PR, the train admission rule (landingcar.go:156-175) admits only granted or vouched-ready tickets, and Wait returns only on a final outcome. A conductor that dies after its merge lands, a Settle that fails (landingtrain.go:495 wrongly says the car resumes), or an abort after an unconfirmed merge all leave cars waiting until an operator interrupts. This contradicts the plan''s state-table row (the next head or the car''s resume settles every MERGED car). Fix: make "own PR is MERGED" an exit from Wait (resume) and add a killed-conductor-after-merge sequence test.'
          family: uncertain-outcome-collapsed
          round: 4
        - kind: defect
          id: BR-17
          severity: Important
          title: The conductor's merge() treats its own landed train merge as an outside move of main, then aborts and releases merged cars
          detail: 'After LandingMerge is accepted but the probe lags, Probed{false} re-sends Merge. fetchMain then differs from c.base because main contains the train merge, so the conductor reintegrates, publish times out waiting for an OPEN train PR, and the abort defer releases every car (which then hits the hang above). Fix: in merge() (landingtrain.go:399-418), treat c.heads[k-1] being an ancestor of the fetched main as Merged before emitting MainMoved; add a fake mode where the listing lags.'
          family: uncertain-outcome-collapsed
          round: 4
        - kind: defect
          id: BR-18
          severity: Important
          title: A merged train's landing-train/<id> branch is never deleted from the remote
          detail: 'realGH.LandingMerge (ghlanding.go:172) has no --delete-branch, the conductor deletes nothing after Merged, and reclaimTrains lists only OPEN PRs, so every landed train leaves a branch on the remote. The plan says the merge deletes it. Fix: run a leased deleteRemoteBranch on the train branch after LandCars, and assert in landTrainOfThree that no landing-train ref remains.'
          family: artifact-without-removal
          round: 4
        - kind: defect
          id: BR-19
          severity: Minor
          title: The abort defer sets tr.Phase directly, outside Transition, and its released cars miss tr.Released, so NextDue waits a full interval
          detail: landingtrain.go:104-112. Model the abort as a Transition event so released cars make the next train depart at once.
          family: state-machine-bypass
          round: 4
        - kind: defect
          id: BR-20
          severity: Minor
          title: waitCI and maybeConductTrain use the global ghClient instead of the injected trainGH seam
          detail: landingtrain.go:385 and landingtrain.go:591.
          family: seam-bypass-global-client
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b70e3700-f4b8-4293-bbb3-19baadb4369f
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/ac572036146ee30f
            reviewed_head: a395730bed6d00be7b996616f427d8b8594ff447
            slot: ariadne:2
            requested_at: "2026-10-19T12:37:54.938341Z"
          phase: finished
          pid: 75404
          started_at: "2026-10-19T12:37:55.034639Z"
          finished_at: "2026-10-19T13:12:05.608555Z"
          session_id: 5ba1e5bb-a8ef-45e1-a772-1c9e5c5bd13d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a395730bed6d00be7b996616f427d8b8594ff447
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T05:06:58-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T05:06:58-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T05:06:58-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T05:06:58-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T05:06:58-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T05:06:58-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T05:06:58-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 1641c61a6d549e9a012a23fd172b5df1654bce83db15ad84a21c12175be8351b
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 296072069
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T05:06:58-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T02:09:30-07:00"
              finalized: "2026-10-19T02:57:26-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T03:35:22-07:00"
              finalized: "2026-10-19T03:50:39-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-19T04:58:25-07:00"
              finalized: "2026-10-19T05:02:37-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-19T06:12:06-07:00"
      dispatched: "2026-10-19T05:37:54-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-16
          disposition: addressed
          note: carReadiness reports merged and vouchWhileWaiting ends the wait; TestLandingCarMergedWithoutAWordFinishes goes red (hang, 121s) with the check reverted.
          round: 5
        - id: BR-17
          disposition: addressed
          note: merge() treats head-on-main as Merged (landingtrain.go:428); TestLandingTrainLaggingListingIsItsOwnMerge goes red with the ancestor check reverted.
          round: 5
        - id: BR-18
          disposition: addressed
          note: deleteTrainBranch on PhaseLanded plus the reclaim sweep; landTrainOfThree fails in all 3 variants (landing-train/2 remains) with the call reverted.
          round: 5
        - id: BR-19
          disposition: addressed
          note: Aborted is a Transition event that appends to Released; pinned by TestTrainAbortReleasesEveryRidingCar and the random-sequence test.
          round: 5
        - id: BR-20
          disposition: addressed
          note: waitCI uses c.ci and maybeConductTrain uses the injected lgh/ci; wiring only, checked by reading the code (the sibling at landingcar.go:100 is raised separately).
          round: 5
      findings:
        - kind: defect
          id: BR-21
          severity: Minor
          title: An abort while merging records the train abandoned and its cars released although the merge may have landed
          detail: 'landingtrain.go:110-128 with trainstate.go:171-183: a probe IO error after MergeUnknown aborts into PhaseAbandoned and closes the PR. This is the 3rd finding in this family. Rule: a terminal train record needs a confirming observation; from PhaseMerging, Aborted should record merge-unknown, not abandoned.'
          family: uncertain-outcome-collapsed
          round: 5
        - kind: defect
          id: BR-22
          severity: Minor
          title: productionCarGates still calls the global ghClient for the car CI gate
          detail: 'landingcar.go:100, a sibling of BR-20. Rule: no landing*.go code reads ghClient except where the merge is wired up (landing.go:360); pass the seam in.'
          family: seam-bypass-global-client
          round: 5
        - kind: defect
          id: BR-23
          severity: Minor
          title: Remote train-branch deletion is hand-written three times, and only closePR's is unleased
          detail: 'landingtrain.go:532, 553 and 619. This is the 3rd finding in this family. Rule: one leased deleteRemoteBranch(branch, tip) helper for every train-branch delete.'
          family: duplicated-rule
          round: 5
        - kind: defect
          id: BR-24
          severity: Minor
          title: The atlas conductor section does not describe landed-branch deletion, the leftover-branch sweep, or the merged-while-waiting resume
          detail: atlas/workflow/sdlc-binary.md:431-436 still says only that open leftover PRs are closed.
          family: docs-gate-atlas
          round: 5
        - kind: defect
          id: BR-25
          severity: Minor
          title: The plan's list of events the code implements leaves out Aborted
          detail: 'Plan line 197 lists events without Aborted (trainstate.go:102). This is the 3rd finding in this family. Rule: that list mirrors every type that implements trainEvent and trainEffect.'
          family: plan-core-concepts-drift
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: a395730bed6d00be7b996616f427d8b8594ff447
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3a4528e5-92fc-4783-922e-e83d4b9c8f04
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/f7d3e9b01f1175d5
            reviewed_head: 0db9e31fed4559d1859f1640624bbd5118e4c632
            slot: ariadne:2
            requested_at: "2026-10-19T14:06:38.81747Z"
          phase: finished
          pid: 95470
          started_at: "2026-10-19T14:06:39.089237Z"
          finished_at: "2026-10-19T14:22:34.130251Z"
          session_id: 6a98b83d-0ceb-4c8e-9d3c-9af442be4c13
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 0db9e31fed4559d1859f1640624bbd5118e4c632
        tracker_head: c859465254df59903a291f0f6ed200ca9dcdafe1
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T06:31:44-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T06:31:44-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T06:31:44-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T06:31:44-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T06:31:44-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T06:31:44-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T06:31:44-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 73cc7b4fff334781c482f6488e62e7933cfc595100fc1ac81053a5e1e9e1bab1
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 301908529
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T06:31:44-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T02:09:30-07:00"
              finalized: "2026-10-19T02:57:26-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T03:35:22-07:00"
              finalized: "2026-10-19T03:50:39-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-19T04:58:25-07:00"
              finalized: "2026-10-19T05:02:37-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-19T05:37:54-07:00"
              finalized: "2026-10-19T06:12:06-07:00"
              finalizing: true
      "n": 6
      timestamp: "2026-10-19T07:22:34-07:00"
      dispatched: "2026-10-19T07:06:38-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-26
          severity: Important
          title: README train paragraph not updated for M3 surface (splits, holds, refusals, alone, per-train report)
          detail: 'This is the 2nd finding in family docs-gate-readme. The plan''s M3 docs task names README, but README.md:491-501 still ends at the readiness line and omits the bisect split with the held culprit, conflict holds, main-red and train-CI-red refusal, local-gates-label lands alone, resume on a merged PR with no word, and the last-trains report in sdlc landing status. All of these are in helptext/landing.md and atlas. Rule: every user-visible train behavior in helptext LANDING TRAINS must also appear in README''s train paragraph; sweep the whole list.'
          family: docs-gate-readme
          round: 6
        - kind: defect
          id: BR-27
          severity: Minor
          title: The "is issue granted now" lookup is written out at three sites
          detail: 'This is the 4th finding in family duplicated-rule. GrantEntries plus ValidGrants(...)[issue] appears at landingqueue.go:50, landingqueue.go:178 and landingtrain.go:665. Rule: grant status is one Line method that every caller uses, so an unreadable log means the same thing everywhere.'
          family: duplicated-rule
          round: 6
        - kind: defect
          id: BR-28
          severity: Minor
          title: Board ignores grants, so a waiter granted after admission can board a train
          detail: Readiness.Alone is set only from the local-gates label (landingcar.go:73). A grant added between the head's admission and boarding lets the granted car ride a train that another car's red CI can refuse, contrary to "lands alone, outside the train".
          family: grant-bypass-boarding
          round: 6
        - kind: defect
          id: BR-29
          severity: Minor
          title: TestTrainCodeTakesItsGitHubSeam checks a hand-maintained list of four files
          detail: A new landing*.go file escapes the seam check. Glob the landing files, excluding landing.go, the wiring site.
          family: contract-test-hand-list
          round: 6
      boundary: M3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: bacf2150-dcf8-4516-94c0-3df12edaa04b
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/acacff334121b161
            reviewed_head: b1ab3a4cec0a2314e38c8a344fee2a9956cee0c8
            slot: ariadne:2
            requested_at: "2026-10-19T15:07:20.518007Z"
          phase: finished
          pid: 55761
          started_at: "2026-10-19T15:07:20.793915Z"
          finished_at: "2026-10-19T15:22:30.264475Z"
          session_id: 11e611e1-1520-4aaa-9ee9-0af9e6ef36f2
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: b1ab3a4cec0a2314e38c8a344fee2a9956cee0c8
        tracker_head: 3adcfba4745aa1327f4d85e699615d15b16d7019
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T07:24:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T07:24:08-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T07:24:08-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T07:24:08-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T07:24:08-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T07:24:08-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T07:24:08-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2694bfa4d5d1c431f9b306c69b7bd7c2099efe4f0c2ddcd284867b426f5ac69e
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 304845287
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T07:24:08-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T02:09:30-07:00"
              finalized: "2026-10-19T02:57:26-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T03:35:22-07:00"
              finalized: "2026-10-19T03:50:39-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-19T04:58:25-07:00"
              finalized: "2026-10-19T05:02:37-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-19T05:37:54-07:00"
              finalized: "2026-10-19T06:12:06-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-19T07:06:38-07:00"
              finalized: "2026-10-19T07:22:34-07:00"
              finalizing: false
      "n": 7
      timestamp: "2026-10-19T08:22:43-07:00"
      dispatched: "2026-10-19T08:07:20-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-26
          disposition: addressed
          note: README.md:491-520 now lists vouching, departure, conducting, holds/bisect, flake/main-red/CI-red refusals, alone (grant and local-gates label), resume, and the last-trains status, matching helptext/landing.md LANDING TRAINS one-for-one.
          round: 7
        - id: BR-27
          disposition: addressed
          note: Line.GrantsNow (log.go:89) is used at landingqueue.go:50, landingqueue.go:178 and landingtrain.go:665; the remaining direct ValidGrants use at landingcmd.go:171 needs the skipped count for status.
          round: 7
        - id: BR-28
          disposition: addressed
          note: Board takes grants and passes over a granted waiter (train.go:227); TestBoard granted-waiter subtest fails when the check is disabled (verified by revert) and passes with it.
          round: 7
        - id: BR-29
          disposition: addressed
          note: TestTrainCodeTakesItsGitHubSeam now globs landing*.go, excluding landing.go and tests, with a checked<5 guard against an outdated glob; passes.
          round: 7
      boundary: M3
      recipe: milestone-review
      isolation: fenced
      reviewed: b1ab3a4cec0a2314e38c8a344fee2a9956cee0c8
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 419ddf89-913b-4340-88b8-f996621aca3f
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: milestone-review
            round: 8
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/dea1dca8d8f2007b
            reviewed_head: 74171bfb3611d7c69ee0def1dbb914d7c2b33136
            slot: ariadne:2
            requested_at: "2026-10-19T15:52:57.472418Z"
          phase: finished
          pid: 28029
          started_at: "2026-10-19T15:52:57.570684Z"
          finished_at: "2026-10-19T16:17:08.384877Z"
          session_id: 2946c7c0-22f2-4b17-a6d8-158e61b68f27
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 74171bfb3611d7c69ee0def1dbb914d7c2b33136
        tracker_head: 950af54aa57af8a2ba8b5e748fd19b1677daf0c5
        branch_point: e279f713f66ebde2d11e8662e05d4e481369d34c
        since: "2026-10-19T00:24:20-07:00"
        until: "2026-10-19T08:22:45-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T00:24:20-07:00"
              end: "2026-10-19T08:22:45-07:00"
              began: "2026-10-19T00:24:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3d99778b04a5abfcad9c52903f374e96319ca817fd257eef5893cf415eaa2517
              started: "2026-10-19T07:41:27.89Z"
              spawned: false
              bytes: 392572
              windows:
                - start: "2026-10-19T07:41:27.89Z"
                  end: "2026-10-19T08:22:45-07:00"
            - harness: claude
              id: 3859e47f-fab7-4809-b00b-39545e065c4d
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9e7b79ccc03b5a83546f49fdb9ea8267d34e5ef1958759092e3826439da67440
              started: "2026-10-19T07:39:08.194Z"
              spawned: false
              bytes: 515896
              windows:
                - start: "2026-10-19T07:39:08.194Z"
                  end: "2026-10-19T08:22:45-07:00"
            - harness: claude
              id: ce283a71-a4ff-4b46-a444-7e71d73538b9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f4547a086362f57592441b8ebe51ccc25025e5807e92b9ae941f19565d34db65
              started: "2026-10-19T07:37:21.957Z"
              spawned: false
              bytes: 568734
              windows:
                - start: "2026-10-19T07:37:21.957Z"
                  end: "2026-10-19T08:22:45-07:00"
            - harness: claude
              id: ea85e2ea-80c1-424b-b626-ad5588a39352
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00ba0b4ef6fcb0b4fa7599cbc5b7b3b9f7a747009b08fefc1d56ab99fd9c60f
              started: "2026-10-19T07:41:05.217Z"
              spawned: false
              bytes: 498662
              windows:
                - start: "2026-10-19T07:41:05.217Z"
                  end: "2026-10-19T08:22:45-07:00"
            - harness: claude
              id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 32f1c1354f947b813042e42af826f14db41d5bf479efab13555853520d6330f7
              started: "2026-10-19T07:40:04.612Z"
              spawned: false
              bytes: 502132
              windows:
                - start: "2026-10-19T07:40:04.612Z"
                  end: "2026-10-19T08:22:45-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a875a5b3cbf8a54fa4c2e82ba56ef3b05687013f49666172d1cb8b01eceb37be
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 306893721
              windows:
                - start: "2026-10-19T00:24:20-07:00"
                  end: "2026-10-19T08:22:45-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-19T01:40:06-07:00"
              finalized: "2026-10-19T01:43:53-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T02:09:30-07:00"
              finalized: "2026-10-19T02:57:26-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-19T03:35:22-07:00"
              finalized: "2026-10-19T03:50:39-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-19T04:58:25-07:00"
              finalized: "2026-10-19T05:02:37-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-19T05:37:54-07:00"
              finalized: "2026-10-19T06:12:06-07:00"
              finalizing: true
            - boundary: M3
              dispatched: "2026-10-19T07:06:38-07:00"
              finalized: "2026-10-19T07:22:34-07:00"
              finalizing: false
            - boundary: M3
              dispatched: "2026-10-19T08:07:20-07:00"
              finalized: "2026-10-19T08:22:43-07:00"
              finalizing: true
      "n": 8
      timestamp: "2026-10-19T09:17:09-07:00"
      dispatched: "2026-10-19T08:52:57-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-14
          disposition: addressed
          note: line.go:240-247 reads Events() inside withMutex after head.lock is acquired; only the schedule rule is built outside.
          round: 8
        - id: BR-15
          disposition: addressed
          note: landingqueue.go:37-40 doc and plan line 205 now describe the 30s cached schedule and the in-mutex ledger read.
          round: 8
        - id: BR-21
          disposition: addressed
          note: trainstate.go:181-186 maps Aborted from PhaseMerging to PhaseUnconfirmed with no ClosePR; pinned by trainstate_test.go:325.
          round: 8
        - id: BR-22
          disposition: addressed
          note: productionCarGates(r, ci ghCaller, ...) takes the seam (landingcar.go:90); glob-based TestTrainCodeTakesItsGitHubSeam passes.
          round: 8
        - id: BR-23
          disposition: not-addressed
          note: deleteTrainRef is the leased helper, but the opened-PR close path still deletes via unleased gh pr close --delete-branch (ghlanding.go:244), contradicting the atlas's one-helper claim.
          round: 8
        - id: BR-24
          disposition: addressed
          note: atlas/workflow/sdlc-binary.md conductor section now covers branch cleanup, leftover-branch sweep, and merged-without-a-word resume.
          round: 8
        - id: BR-25
          disposition: addressed
          note: Plan line 210 lists the full event set including Aborted, matching trainstate.go:113-122.
          round: 8
      recipe: milestone-review
      isolation: fenced
      reviewed: 74171bfb3611d7c69ee0def1dbb914d7c2b33136
      blocked: false
---

# Gate ledger — ariadne#449 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T01:43:53-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `shared-handle-unsynchronized` (kind: defect) Data race on Handle.file between Vouch (vouch goroutine) and dropTicket (Wait/admission)
  line.go:265 reads h.file in Vouch while line.go:383 writes it via tryAdmit; only an flock separates them, which the Go memory model does not see. A temporary probe test (Vouch in a loop alongside Wait) is flagged by go test -race. vouchWhileWaiting runs exactly this in production; it also prints a misleading "could not record readiness" warning after admission. Fix: a sync.Mutex on Handle covering file/dropTicket/Vouch, treat the post-admission error as a quiet stop, and keep the probe as a regression test.
- **BR-2** [Important] `state-machine-event-partition` (kind: defect) Transition(Built) silently drops cars that are in neither Kept nor Conflicts
  trainstate.go:181-188 removes such cars from n.Cars with no Hold/Release/Refuse effect, breaking the every-car-ends-once invariant. The random-history generator always partitions the cars, so the invariant test cannot reach this input. Reject a non-partitioning Built (return an error) or Release the unaccounted cars, and make the generator produce such inputs.
- **BR-3** [Important] `untrusted-sidecar-cross-check` (kind: defect) Board builds Car.Branch from the sidecar's branch without checking it against the ticket's branch
  train.go:293 uses r.Branch, and Readiness.Valid never compares it with Ticket.Branch. A hand-edited or buggy sidecar can steer the M2 conductor to a branch the waiting merge does not own, and the tip re-check passes because it reads the same untrusted branch. Refuse a mismatch in Board (add it to Why) or use t.Branch, and add a table test.
- **BR-4** [Important] `plan-core-concepts-drift` (kind: defect) Plan Core concepts and state lists do not match the code (TrainState, Held, event names, DepartureDue signature)
  The plan names TrainState, a Held state and the events CarConflict/CarGone/TipMoved/FlakeRerunPass/CIGreen/CIRed/ConductorDead, plus DepartureDue(now, lastDeparture, sched, line). The code has Train{Phase} with Reintegrating/Merging/Refused, the events Boarded/Built/SuiteDone/CIDone/Reintegrated/Probed, and DepartureDue(now, s, granted, events) reading next_due from the log. Add a Revisions entry and update the table.
- **BR-5** [Important] `docs-gate-readme` (kind: defect) README update appears missing for the .sdlc/landing.json config keys and the train lines in landing status
  train_interval_minutes and train_max_cars are a new user-facing config surface, and status already points users to the file; README.md is unchanged in the range. Add a short README note, or record the deferral to M3 explicitly in the plan's Revisions.
- **BR-6** [Minor] `repeated-io-in-hot-loop` (kind: defect) trainAdmission reads the whole events.jsonl once per live ticket per poll, inside the line mutex
  Read the events log once per tryAdmit and pass it in.
- **BR-7** [Minor] `duplicated-rule` (kind: defect) The grant bypass is written in both trainAdmission and DepartureDue, and "main" has two definitions (main@{upstream} vs remote/main)
  trainAdmission always passes granted=false, so DepartureDue's granted branch is dead in production. landingMainRef (status) and runDurableMerge (t.Remote+"/main") resolve main differently.
- **BR-8** [Minor] `stale-config-read` (kind: defect) The schedule is read from an unfetched remote-tracking main before the wait and held for the whole wait
  A missing remote/main ref now refuses even interval-0 merges, and a config change made during a long wait is not seen.
- **BR-9** [Minor] `audit-displaced-filtered` (kind: defect) Not-ready tickets skipped by Admissible are missing from OnAdmit's displaced audit
  NextHead runs over the filtered candidates, so the tickets passed over are not recorded.

## Round 2 — 2026-10-19T02:57:26-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Handle.mu guards file/head; removing it from Vouch makes TestVouchConcurrentWithAdmission fail with DATA RACE under -race.
- BR-2 — addressed — partitions() rejects non-partitioning Built; reverting fails TestTrainRejectsIllegalEvents and the random-history invariant.
- BR-3 — addressed — Readiness.For checks the branch and Board uses t.Branch; reverting fails TestBoard and TestTrainAdmission.
- BR-4 — addressed — Plan Revisions entry M1 as built (BR-4) matches trainstate.go/train.go names, events, effects and the DepartureDue signature.
- BR-5 — addressed — README.md:491-498 documents train_interval_minutes, train_max_cars, the local-gates guard, and the status lines.
- BR-6 — addressed — Options.Admission is built once per attempt; TestTrainAdmission counts one events read across two tickets.
- BR-7 — addressed — The grant branch is gone from DepartureDue; one landingMainRef(git, remote) is used by both merge and status, and is tested.
- BR-8 — addressed — landingScheduleSource re-reads on each attempt and falls back to serial; TestLandingScheduleSourceFailsToSerial and TestVouchingStartsWhenTrainsTurnOn cover it.
- BR-9 — addressed — The audit walks the full order; reverting to NextHead's displaced fails TestAdmissibleGatesTheHead.

### Raised

- **BR-10** [Important] `shared-handle-unsynchronized` (kind: defect) Data race on the warned flag in landingScheduleSource, shared by the vouching goroutine and the admission poll
  landingqueue.go:206-207. enterLandingLine passes one sched closure to both openLandingLine (Admission, Wait goroutine) and vouchWhileWaiting; a two-goroutine probe with an unreadable config is flagged by go test -race. This is the 2nd finding in this family. The rule: everything handed to both vouchWhileWaiting and Line.Wait must be safe for concurrent use. Swept: the line (flock and Handle.mu), readiness (vouch goroutine only), stderr (os.File) are fine; sched is the only remaining instance. Fix with sync.Once or atomic.Bool and a concurrent -race test.
- **BR-11** [Minor] `untrusted-sidecar-cross-check` (kind: defect) landing status reads readiness sidecars without Readiness.For, so a vouch for another branch shows as ready
  landingcmd.go:199-203 uses readinessLine and Current only. This is the 2nd finding in this family; the rule is that every sidecar reader goes through Readiness.For(ticket, now). Status also prints trains: unknown when the config is unreadable, while merges then land serially.
- **BR-12** [Minor] `repeated-io-in-hot-loop` (kind: defect) Every waiter runs two git processes each 2s poll under the line flock to re-read the schedule, even with trains off
  trainAdmission calls sched() (git ls-tree and git show) inside tryAdmit's mutex on every attempt. Before this change the serial default ran none. This is the 2nd finding in this family; the rule is that admission inputs are read outside the mutex and slow-changing ones are cached (time limit or main's commit id).
- **BR-13** [Minor] `duplicated-rule` (kind: defect) The audit's passed-over rule is restated in line.go while NextHead's displaced return has no production caller
  line.go:244-257 duplicates order.go:95-99. This is the 2nd finding in this family; the rule is one pure function in order.go computes who was passed over, and NextHead returns only the head.

## Round 3 — 2026-10-19T03:50:39-07:00 (claude) — passed

### Disposed

- BR-10 — addressed — Mutex in landingScheduleSource; with the lock removed, TestLandingScheduleSourceIsSharedAndCached fails under -race with 3 DATA RACE reports and 4 reads.
- BR-11 — addressed — landingcmd.go:201-206 uses r.For(t, now); the test vouches with another ticket's branch and expects "not counted"; the unreadable-config line now says merges land serially.
- BR-12 — addressed — Admission() is now built before withMutex; the schedule is cached for 30s, so the serial default runs no git on most polls; the cache test pins one read per TTL.
- BR-13 — addressed — Passed() in order.go:96-112 is the only rule; NextHead returns (head, ok); line.go calls Passed; TestPassedIncludesTicketsThatWereNotAdmissible.

### Raised

- **BR-14** [Minor] `admission-decision-toctou` (kind: defect) Admission reads the events log (the departure ledger) before the line mutex and head.lock, so M2 can admit a second train on a stale "due"
  line.go:188-193 builds admissible from Events() outside withMutex. Once M2 records EventTrain and then releases head.lock, a waiter that read the events just before can take the head and depart early, skipping the interval. Keep the cached schedule outside the mutex, but read the events and decide DepartureDue under the mutex after head.lock is acquired (Events() does not take the mutex). Latent in M1: nothing writes EventTrain yet.
- **BR-15** [Minor] `plan-core-concepts-drift` (kind: defect) The openLandingLine doc and plan Revisions still say the schedule is re-read on every admission attempt; the code caches it for 30s
  landingqueue.go:39 and plan line 205. 2nd in this family. The rule: a commit that changes behavior described by the "M1 as built" Revisions entry or a function doc updates that text in the same commit; check those passages as part of every answer-review commit.

## Round 4 — 2026-10-19T05:02:37-07:00 (claude) — BLOCKED

### Raised

- **BR-16** [Critical] `uncertain-outcome-collapsed` (kind: defect) With trains on, a waiting car whose PR is already MERGED can never be admitted or settled and hangs
  carReadiness (landingcar.go:56) returns not ready for a merged PR, the train admission rule (landingcar.go:156-175) admits only granted or vouched-ready tickets, and Wait returns only on a final outcome. A conductor that dies after its merge lands, a Settle that fails (landingtrain.go:495 wrongly says the car resumes), or an abort after an unconfirmed merge all leave cars waiting until an operator interrupts. This contradicts the plan's state-table row (the next head or the car's resume settles every MERGED car). Fix: make "own PR is MERGED" an exit from Wait (resume) and add a killed-conductor-after-merge sequence test.
- **BR-17** [Important] `uncertain-outcome-collapsed` (kind: defect) The conductor's merge() treats its own landed train merge as an outside move of main, then aborts and releases merged cars
  After LandingMerge is accepted but the probe lags, Probed{false} re-sends Merge. fetchMain then differs from c.base because main contains the train merge, so the conductor reintegrates, publish times out waiting for an OPEN train PR, and the abort defer releases every car (which then hits the hang above). Fix: in merge() (landingtrain.go:399-418), treat c.heads[k-1] being an ancestor of the fetched main as Merged before emitting MainMoved; add a fake mode where the listing lags.
- **BR-18** [Important] `artifact-without-removal` (kind: defect) A merged train's landing-train/<id> branch is never deleted from the remote
  realGH.LandingMerge (ghlanding.go:172) has no --delete-branch, the conductor deletes nothing after Merged, and reclaimTrains lists only OPEN PRs, so every landed train leaves a branch on the remote. The plan says the merge deletes it. Fix: run a leased deleteRemoteBranch on the train branch after LandCars, and assert in landTrainOfThree that no landing-train ref remains.
- **BR-19** [Minor] `state-machine-bypass` (kind: defect) The abort defer sets tr.Phase directly, outside Transition, and its released cars miss tr.Released, so NextDue waits a full interval
  landingtrain.go:104-112. Model the abort as a Transition event so released cars make the next train depart at once.
- **BR-20** [Minor] `seam-bypass-global-client` (kind: defect) waitCI and maybeConductTrain use the global ghClient instead of the injected trainGH seam
  landingtrain.go:385 and landingtrain.go:591.

## Round 5 — 2026-10-19T06:12:06-07:00 (claude) — passed

### Disposed

- BR-16 — addressed — carReadiness reports merged and vouchWhileWaiting ends the wait; TestLandingCarMergedWithoutAWordFinishes goes red (hang, 121s) with the check reverted.
- BR-17 — addressed — merge() treats head-on-main as Merged (landingtrain.go:428); TestLandingTrainLaggingListingIsItsOwnMerge goes red with the ancestor check reverted.
- BR-18 — addressed — deleteTrainBranch on PhaseLanded plus the reclaim sweep; landTrainOfThree fails in all 3 variants (landing-train/2 remains) with the call reverted.
- BR-19 — addressed — Aborted is a Transition event that appends to Released; pinned by TestTrainAbortReleasesEveryRidingCar and the random-sequence test.
- BR-20 — addressed — waitCI uses c.ci and maybeConductTrain uses the injected lgh/ci; wiring only, checked by reading the code (the sibling at landingcar.go:100 is raised separately).

### Raised

- **BR-21** [Minor] `uncertain-outcome-collapsed` (kind: defect) An abort while merging records the train abandoned and its cars released although the merge may have landed
  landingtrain.go:110-128 with trainstate.go:171-183: a probe IO error after MergeUnknown aborts into PhaseAbandoned and closes the PR. This is the 3rd finding in this family. Rule: a terminal train record needs a confirming observation; from PhaseMerging, Aborted should record merge-unknown, not abandoned.
- **BR-22** [Minor] `seam-bypass-global-client` (kind: defect) productionCarGates still calls the global ghClient for the car CI gate
  landingcar.go:100, a sibling of BR-20. Rule: no landing*.go code reads ghClient except where the merge is wired up (landing.go:360); pass the seam in.
- **BR-23** [Minor] `duplicated-rule` (kind: defect) Remote train-branch deletion is hand-written three times, and only closePR's is unleased
  landingtrain.go:532, 553 and 619. This is the 3rd finding in this family. Rule: one leased deleteRemoteBranch(branch, tip) helper for every train-branch delete.
- **BR-24** [Minor] `docs-gate-atlas` (kind: defect) The atlas conductor section does not describe landed-branch deletion, the leftover-branch sweep, or the merged-while-waiting resume
  atlas/workflow/sdlc-binary.md:431-436 still says only that open leftover PRs are closed.
- **BR-25** [Minor] `plan-core-concepts-drift` (kind: defect) The plan's list of events the code implements leaves out Aborted
  Plan line 197 lists events without Aborted (trainstate.go:102). This is the 3rd finding in this family. Rule: that list mirrors every type that implements trainEvent and trainEffect.

## Round 6 — 2026-10-19T07:22:34-07:00 (claude) — BLOCKED

### Raised

- **BR-26** [Important] `docs-gate-readme` (kind: defect) README train paragraph not updated for M3 surface (splits, holds, refusals, alone, per-train report)
  This is the 2nd finding in family docs-gate-readme. The plan's M3 docs task names README, but README.md:491-501 still ends at the readiness line and omits the bisect split with the held culprit, conflict holds, main-red and train-CI-red refusal, local-gates-label lands alone, resume on a merged PR with no word, and the last-trains report in sdlc landing status. All of these are in helptext/landing.md and atlas. Rule: every user-visible train behavior in helptext LANDING TRAINS must also appear in README's train paragraph; sweep the whole list.
- **BR-27** [Minor] `duplicated-rule` (kind: defect) The "is issue granted now" lookup is written out at three sites
  This is the 4th finding in family duplicated-rule. GrantEntries plus ValidGrants(...)[issue] appears at landingqueue.go:50, landingqueue.go:178 and landingtrain.go:665. Rule: grant status is one Line method that every caller uses, so an unreadable log means the same thing everywhere.
- **BR-28** [Minor] `grant-bypass-boarding` (kind: defect) Board ignores grants, so a waiter granted after admission can board a train
  Readiness.Alone is set only from the local-gates label (landingcar.go:73). A grant added between the head's admission and boarding lets the granted car ride a train that another car's red CI can refuse, contrary to "lands alone, outside the train".
- **BR-29** [Minor] `contract-test-hand-list` (kind: defect) TestTrainCodeTakesItsGitHubSeam checks a hand-maintained list of four files
  A new landing*.go file escapes the seam check. Glob the landing files, excluding landing.go, the wiring site.

## Round 7 — 2026-10-19T08:22:43-07:00 (claude) — passed

### Disposed

- BR-26 — addressed — README.md:491-520 now lists vouching, departure, conducting, holds/bisect, flake/main-red/CI-red refusals, alone (grant and local-gates label), resume, and the last-trains status, matching helptext/landing.md LANDING TRAINS one-for-one.
- BR-27 — addressed — Line.GrantsNow (log.go:89) is used at landingqueue.go:50, landingqueue.go:178 and landingtrain.go:665; the remaining direct ValidGrants use at landingcmd.go:171 needs the skipped count for status.
- BR-28 — addressed — Board takes grants and passes over a granted waiter (train.go:227); TestBoard granted-waiter subtest fails when the check is disabled (verified by revert) and passes with it.
- BR-29 — addressed — TestTrainCodeTakesItsGitHubSeam now globs landing*.go, excluding landing.go and tests, with a checked<5 guard against an outdated glob; passes.

## Round 8 — 2026-10-19T09:17:09-07:00 (claude) — passed

### Disposed

- BR-14 — addressed — line.go:240-247 reads Events() inside withMutex after head.lock is acquired; only the schedule rule is built outside.
- BR-15 — addressed — landingqueue.go:37-40 doc and plan line 205 now describe the 30s cached schedule and the in-mutex ledger read.
- BR-21 — addressed — trainstate.go:181-186 maps Aborted from PhaseMerging to PhaseUnconfirmed with no ClosePR; pinned by trainstate_test.go:325.
- BR-22 — addressed — productionCarGates(r, ci ghCaller, ...) takes the seam (landingcar.go:90); glob-based TestTrainCodeTakesItsGitHubSeam passes.
- BR-23 — not-addressed — deleteTrainRef is the leased helper, but the opened-PR close path still deletes via unleased gh pr close --delete-branch (ghlanding.go:244), contradicting the atlas's one-helper claim.
- BR-24 — addressed — atlas/workflow/sdlc-binary.md conductor section now covers branch cleanup, leftover-branch sweep, and merged-without-a-word resume.
- BR-25 — addressed — Plan line 210 lists the full event set including Aborted, matching trainstate.go:113-122.

## Open findings

- **BR-23** [Minor] `duplicated-rule` (kind: defect) Remote train-branch deletion is hand-written three times, and only closePR's is unleased
