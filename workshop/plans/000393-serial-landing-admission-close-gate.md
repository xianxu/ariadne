---
gate: boundary-review
issue: 393
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 44c551a2-f8fb-4d94-842e-f5cef1741ce2
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/565e06589b4d0944
            reviewed_head: 17b06388290cedca6a6b94f80652583481b448df
            slot: ariadne:2
            requested_at: "2026-10-16T07:05:18.296777Z"
          phase: finished
          pid: 14769
          started_at: "2026-10-16T07:05:18.498697Z"
          finished_at: "2026-10-16T07:07:20.669233Z"
          session_id: 9218fbae-1512-4e98-9fae-cd38013d4370
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 17b06388290cedca6a6b94f80652583481b448df
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T00:04:48-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T00:04:48-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T00:04:48-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5e9e0a723ffa7f59bad1b9b9b6fd203cd41f2a8e4b73691b46da79e4a825c7fb
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 12837356
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T00:04:48-07:00"
      "n": 1
      timestamp: "2026-10-16T00:07:33-07:00"
      dispatched: "2026-10-16T00:05:18-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Plan Core concepts table contradicts the M1 code (DelayStats/stats.go, NextHead signature, Ticket.PR, missing Order)
          detail: The plan table lists DelayStats in stats.go (absent; the code has Delays in log.go), NextHead taking now (it doesn't), and Ticket with PR (it doesn't), and omits the pure Order. The deviations are only in the issue Log. Add a plan Revisions entry and update the table and Task 3's file list.
          family: plan-table-drift
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: ValidGrants accepts future-dated grants, and the sweep never removes them
          detail: 'order.go:59-60 checks now.Before(Expires) and the 24 h span but not Granted <= now, so Granted=2099 confers priority until 2099. The sweep test now.Sub(t) <= Retention (log.go:122) is always true for future times, so the entry is never removed. A probe test confirmed it. Fix: require !Granted.After(now), drop future-dated entries in the sweep, and add a table case.'
          family: untrusted-log-entry-validation
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: AddGrant doesn't stamp Granted, so a revoke with a zero Granted is swept and its grant becomes valid again
          detail: 'Record stamps At when it is zero but AddGrant (log.go:52) doesn''t stamp Granted, and the sweep ages grants by Granted. A revoke appended with a zero Granted is dropped by the next daily sweep, and the earlier grant, still within 24 h, becomes the latest entry and valid again (reproduced with a probe). Fix: stamp Granted from l.o.Now() when it is zero, and test grant, revoke, sweep.'
          family: untrusted-log-entry-validation
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: scan trusts the Seq in a ticket's JSON over its file name
          detail: line.go:262-272 uses the parsed Ticket's Seq; a hand-edited ticket can claim another position. Use the file name as the authoritative Seq.
          family: untrusted-log-entry-validation
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: sweep silently drops malformed log lines instead of counting them
          detail: log.go:122 discards unparsable lines without a count, which contradicts the plan's skipped-and-counted rule.
          family: malformed-input-visibility
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: Wait on a handle whose ticket is gone polls forever until its context ends
          detail: 'line.go:150-157: if h''s Seq isn''t among the live tickets, pos is never 0. Return an error instead.'
          family: wait-termination
          round: 1
        - kind: defect
          id: BR-7
          severity: Minor
          title: Line tests rely on wall-clock timing (150 ms notAdmitted) instead of the injected Sleep seam
          detail: The interleaving isn't controllable or reproducible (ARCH-ORDER), and the tests may be flaky under load.
          family: ordering-test-seam
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 76cf42fa-0e32-47a3-bf0e-8b9039b6f1de
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2ca82bc177b46e19
            reviewed_head: 45eea86e7073b7f34a82638e4293aba92d2e2d5d
            slot: ariadne:2
            requested_at: "2026-10-16T07:10:49.616945Z"
          phase: finished
          pid: 68483
          started_at: "2026-10-16T07:10:49.915698Z"
          finished_at: "2026-10-16T07:13:09.394268Z"
          session_id: 99c7596e-ab46-4003-94ac-79e39e672cee
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 45eea86e7073b7f34a82638e4293aba92d2e2d5d
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T00:10:20-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T00:10:20-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T00:10:20-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f6d4ee0b398d7c92cd32cc6ce4ffcdd7e3f3da29cefbb9cdb5d6be177dffad36
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 13051550
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T00:10:20-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-16T00:05:18-07:00"
              finalized: "2026-10-16T00:07:33-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-16T00:13:11-07:00"
      dispatched: "2026-10-16T00:10:49-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan table now lists Order, Delays in log.go, NextHead without now, Ticket without PR, plus a Revisions entry. Residual Task 1 sketch and Status signature drift is noted as Minor.
          round: 2
        - id: BR-2
          disposition: addressed
          note: order.go:61 adds the Granted-not-after-now check; reverting it turns TestValidGrants red. The sweep drops future-dated entries, but that fix interacts with BR-3's Expires key and caused the new Critical finding.
          round: 2
        - id: BR-3
          disposition: addressed
          note: AddGrant stamps Granted (log.go:59); reverting it turns TestSweptRevokeKeepsTheGrantRevoked red. The revoke-outlives-grant goal is met; the new Critical finding covers the keying regression.
          round: 2
        - id: BR-4
          disposition: addressed
          note: line.go:283-291 takes Seq from the file name; TestLineTicketNumberComesFromItsName covers both a lying JSON body and garbage.
          round: 2
        - id: BR-5
          disposition: addressed
          note: The sweep stamp records malformed=N; TestLogSweepDropsFutureDatedAndCountsMalformed checks it.
          round: 2
        - id: BR-6
          disposition: addressed
          note: tryAdmit errors on a released or vanished ticket (line.go:156-169); TestLineReleasedTicketCantWait covers the released path.
          round: 2
        - id: BR-7
          disposition: addressed
          note: Line tests step tryAdmit directly and inject Sleep; helper tests sync on stdout readiness and cmd.Wait, not timers.
          round: 2
      findings:
        - kind: defect
          id: BR-8
          severity: Critical
          title: 'The grants sweep deletes every grant still in force: retention keys on Expires, and the sweep drops any entry dated in the future'
          detail: 'log.go:67-70 dates a grant by Expires when Expires > Granted, and log.go:144 keeps an entry only when 0 <= now - t <= Retention. A grant still in force has Expires > now, so whenever the daily sweep runs on an AddGrant call (a renewal, another issue''s grant, a revoke), all grants still in force are erased. Probe: AddGrant(issue 1) at t0, AddGrant(issue 391, 24 h) at t0+23h, AddGrant(issue 7) at t0+25h, then ValidGrants has no 391 (FAIL). The atlas states the same rule. This is the 4th finding in family untrusted-log-entry-validation (BR-2 and BR-3 fixed instances, and this one came from the interaction of those fixes). Do not patch this instance alone. Rule: an entry''s removal is decided by a pure keep(entry, now) function in the core, beside ValidGrants, and must preserve the invariant ValidGrants(sweep(L, now), now) == ValidGrants(L, now) for every log L. For grants: drop when malformed, when Granted is after now (forged), or when max(Granted, Expires) + Retention is before now. Events keep the At-based rule. Test the invariant with a table or sequence test over a grant still in force, an expired grant, a revoke, a future-dated grant and a zero-Granted entry. Also apply the rule to the untrusted seq file (line.go:303, where a parse error is silently ignored), and fix the atlas paragraph.'
          family: untrusted-log-entry-validation
          round: 2
        - kind: defect
          id: BR-9
          severity: Minor
          title: Plan Task 1 sketch still passes now to NextHead, and Task 2 gives a Status signature the code doesn't have
          detail: 'This is the 2nd finding in family plan-table-drift. Rule: the Core concepts table and the Revisions entry are authoritative, and every signature elsewhere in the plan either matches them or is marked superseded. Here that covers the sketch at plan:80 and Status() ([]Ticket, headPID int), against the code''s Status() (*Ticket, []Ticket, error).'
          family: plan-table-drift
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 605efb05-bcd4-4f96-bd0c-56e3c60e4cef
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e70d424df0778b09
            reviewed_head: 60c6724cea01c7460dff564a5adf2af2097788bb
            slot: ariadne:2
            requested_at: "2026-10-16T07:15:25.276429Z"
          phase: finished
          pid: 26691
          started_at: "2026-10-16T07:15:25.486136Z"
          finished_at: "2026-10-16T07:17:24.822424Z"
          session_id: 2fb5275e-2a0c-426c-8c42-c1f2b73d72be
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 60c6724cea01c7460dff564a5adf2af2097788bb
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T00:14:50-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T00:14:50-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T00:14:50-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 7cf1fe7c643bc0b86d18b0ce07d4ad4a5609068df76f6372954d7f7c1516af42
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 13154679
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T00:14:50-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-16T00:05:18-07:00"
              finalized: "2026-10-16T00:07:33-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:10:49-07:00"
              finalized: "2026-10-16T00:13:11-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-16T00:17:35-07:00"
      dispatched: "2026-10-16T00:15:25-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: retention.go KeepGrants ages by Granted and drops an issue's whole history when its latest entry is stale. Reverting to the old expiry-based rule makes TestSweepKeepsGrantsInForce, TestKeepGrantsPreservesValidity and TestKeepGrantsCases fail. nextSeq now refuses a malformed seq (TestLineCounterIsUntrustedInput), and the atlas is fixed.
          round: 3
        - id: BR-9
          disposition: not-addressed
          note: 'Fixed: the plan:81 sketch is marked superseded and the Task 2 Status signature matches the code. Still contradicting the code: the Core concepts Integration row and paragraph for Log (Append(file, v) and Sweep(file, now) don''t exist; the methods are on Line), tickets/<seq>.json (the code uses %020d with no extension), Open(dir) (the code takes Options), and Task 3''s "Sweep drops entries older than 30 days" (superseded by KeepGrants). Apply the rule to the whole plan, not just the named sites.'
          round: 3
      findings:
        - kind: defect
          id: BR-10
          severity: Minor
          title: Status reports no head when head.lock is held but its contents don't parse
          detail: 'This is the 2nd finding in this family. In line.go:205-206, readTicket returns nil for a held but unparseable head.lock, so Status (and OnWait''s head, line.go:174) says "no head" while one is running. Rule: a held lock whose contents are unreadable is reported as present but unreadable, never as absent (scan already does this for tickets). Fix: return a non-nil Ticket with an unreadable marker. Test: hold head.lock with garbage and assert head != nil.'
          family: malformed-input-visibility
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      reviewed: 60c6724cea01c7460dff564a5adf2af2097788bb
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a30690e7-c055-4edb-9b15-eb23055729a8
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2d8e850921a514d3
            reviewed_head: 6619f0b44d6b0513d5554ac09b0406f642973dbf
            slot: ariadne:2
            requested_at: "2026-10-16T08:18:13.571086Z"
          phase: finished
          pid: 48633
          started_at: "2026-10-16T08:18:14.375818Z"
          finished_at: "2026-10-16T08:23:06.776724Z"
          session_id: 73d6a3de-6b94-4d08-adc5-1656a874146c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 6619f0b44d6b0513d5554ac09b0406f642973dbf
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T00:40:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T00:40:37-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T00:40:37-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2a0b1ebe6ec75ca8d366d303154c53e170448e567b870fb97722a918a6ee56fd
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 13741374
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T00:40:37-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-16T00:05:18-07:00"
              finalized: "2026-10-16T00:07:33-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:10:49-07:00"
              finalized: "2026-10-16T00:13:11-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:15:25-07:00"
              finalized: "2026-10-16T00:17:35-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-16T01:23:08-07:00"
      dispatched: "2026-10-16T01:18:13-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-11
          severity: Important
          title: anyMergedPR lets a branch with an old merged PR plus a new open PR skip the landing line and land outside it
          detail: 'landing.go:366-375 skips the line if any PR for the branch is MERGED, but gh pr list --state all also returns old merged PRs, and selectLandingPR (landing.go:191) later picks the OPEN PR at the current head and runs integrate/suite/merge without a turn, breaking Done-when 1 when a branch name is reused. Fix: skip only when the PR the landing would select (selectLandingPR with the local head) is MERGED, or when no OPEN PR exists; add a fake-GitHub test with an old merged PR and an open PR on the same branch that waits behind a held line.'
          family: gate-skip-predicate-divergence
          round: 4
        - kind: defect
          id: BR-12
          severity: Important
          title: Plan still claims a landingLineOpen seam and landing_line_test.go, and omits landingqueue.go and an M2 Revisions entry
          detail: '3rd finding in this family, so fix the rule, not this instance: every entity and seam the diff adds or renames appears in the Core concepts tables at its built path, and every as-built deviation gets a dated Revisions entry in the same milestone. Instances: plan line 54 names a package variable landingLineOpen that does not exist (the seams are landingLinePoll and the real common dir); Task 4 names landing_line_test.go but the tests are in landingqueue_test.go; landingqueue.go (enterLandingLine, landingTurn, openLandingLine, landingLineDir) is missing from the Integration points table.'
          family: plan-table-drift
          round: 4
        - kind: defect
          id: BR-13
          severity: Minor
          title: sdlc landing status prints grants in random map order
          detail: landingcmd.go loops over the ValidGrants map directly; sort by issue number.
          family: deterministic-output
          round: 4
        - kind: defect
          id: BR-14
          severity: Minor
          title: Errors from landing-line event and grant writes are ignored with no warning
          detail: The enqueue, OnAdmit and leave writes in landingqueue.go discard errors from Record and AddGrant, so a disk that stays full or unwritable silently loses all audit and pilot data; warn once on stderr.
          family: audit-write-failure-visibility
          round: 4
        - kind: defect
          id: BR-15
          severity: Minor
          title: A duplicate merge admitted after the original landed logs a second landed event and skews the delay p90
          detail: leave() logs EventLanded whenever err is nil, including when the admitted run only resumed cleanup of a PR that was already MERGED; use a distinct kind or skip it.
          family: pilot-metric-fidelity
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a06f3992-9a80-4711-abea-018a15274d9b
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/3613a7274340dd53
            reviewed_head: 5f7a88ff5e5bf298e12f774bcdd6cc09c5f01c4f
            slot: ariadne:2
            requested_at: "2026-10-16T08:29:17.033609Z"
          phase: finished
          pid: 69936
          started_at: "2026-10-16T08:29:17.364371Z"
          finished_at: "2026-10-16T08:59:56.216273Z"
          session_id: e51e4dd9-c6a0-4a65-8d8d-98c38923666a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 5f7a88ff5e5bf298e12f774bcdd6cc09c5f01c4f
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T01:28:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T01:28:10-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T01:28:10-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6279ed7ac64a7705cee5ba2a4b6ee66ec685b2805a43e30ecec68cc764f8a669
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 13901290
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T01:28:10-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-16T00:05:18-07:00"
              finalized: "2026-10-16T00:07:33-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:10:49-07:00"
              finalized: "2026-10-16T00:13:11-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:15:25-07:00"
              finalized: "2026-10-16T00:17:35-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-16T01:18:13-07:00"
              finalized: "2026-10-16T01:23:08-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-16T02:00:00-07:00"
      dispatched: "2026-10-16T01:29:16-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: addressed
          note: anyOpenPR replaces anyMergedPR; reverting it makes TestLandingLineReusedBranchStillQueues fail (no 1 waiting merges), restored it passes.
          round: 5
        - id: BR-12
          disposition: addressed
          note: Integration points table now lists landingqueue.go entities and landingcmd.go; Injected-into says seams are landingLinePoll and landingNow; dated M2 Revisions entry present.
          round: 5
        - id: BR-13
          disposition: not-addressed
          note: slices.Sorted added at landingcmd.go:182, but no status test has two or more grants, so nothing fails without the sort.
          round: 5
        - id: BR-14
          disposition: not-addressed
          note: logLine and the AddGrant warning exist but no test forces a write failure, and the GrantEntries read error in leave() still silently skips ending the grant.
          round: 5
        - id: BR-15
          disposition: addressed
          note: EventResumed logged when the selected PR was MERGED; TestLandingLineDuplicateResumesAsMerged asserts resumed==1 and landed==0, which would fail without the fix.
          round: 5
      findings:
        - kind: defect
          id: BR-16
          severity: Important
          title: Line skip is still decided on the pre-read; a turnless run can land an OPEN PR selected from the second read
          detail: '2nd finding in this family, so state the rule: a run without a turn may only act on a PR that is MERGED when selected. landing.go decides the turn from pre (first LandingPRs) but selectIntegrationPR uses prs (second read); an OPEN PR in the second read lands with turn nil. Fix: after selection, refuse (or enter the line) when turn is nil and pr.State is not MERGED; test with a fake whose first LandingPRs omits the open PR.'
          family: gate-skip-predicate-divergence
          round: 5
        - kind: defect
          id: BR-17
          severity: Important
          title: README update appears missing for sdlc landing status/grant/revoke and for merge waiting in the landing line
          detail: README.md:414-425 describes sdlc merge landing and concurrency but not the line, its waiting output, or the new sdlc landing subcommand; the atlas subcommand table near sdlc-binary.md:43 also lacks a landing row.
          family: user-surface-docs
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b9ec6ee7-cdb2-48cf-a4c7-df9e6511bec9
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/1d6afce90fb8869f
            reviewed_head: 2c61503b65a1ee4d777cbbda9f55d2f93959409a
            slot: ariadne:2
            requested_at: "2026-10-16T09:05:53.249075Z"
          phase: finished
          pid: 53389
          started_at: "2026-10-16T09:05:53.51383Z"
          finished_at: "2026-10-16T09:18:55.069056Z"
          session_id: 7775c767-ff06-4eda-88cb-0ca1c2fd14aa
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2c61503b65a1ee4d777cbbda9f55d2f93959409a
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T02:05:05-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T02:05:05-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T02:05:05-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2a58ed802d7fed603c626f00f4e648cc04f3afd2c958d82f8eb618034cf580ce
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 14041553
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T02:05:05-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-16T00:05:18-07:00"
              finalized: "2026-10-16T00:07:33-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:10:49-07:00"
              finalized: "2026-10-16T00:13:11-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:15:25-07:00"
              finalized: "2026-10-16T00:17:35-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-16T01:18:13-07:00"
              finalized: "2026-10-16T01:23:08-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-16T01:29:16-07:00"
              finalized: "2026-10-16T02:00:00-07:00"
              finalizing: false
      "n": 6
      timestamp: "2026-10-16T02:19:06-07:00"
      dispatched: "2026-10-16T02:05:53-07:00"
      agent: claude
      dispose:
        - id: BR-13
          disposition: addressed
          note: 'landingcmd.go:177 iterates slices.Sorted(maps.Keys(grants)); landingcmd_test.go asserts #12 precedes #391 across repeated runs.'
          round: 6
        - id: BR-14
          disposition: addressed
          note: lineLog.record (landingqueue.go:95) warns once per merge for enqueue/admit/leave writes; leave warns on grant read and AddGrant failures; TestLandingLineLogFailuresWarn pins it.
          round: 6
        - id: BR-16
          disposition: addressed
          note: Rule enforced on the selected PR at landing.go:394; reverting the guard turns TestLandingLineLaggingListingRefusesOutsideTheLine red (err nil); empty listing now queues.
          round: 6
        - id: BR-17
          disposition: addressed
          note: README.md:423-431 covers the line, waiting output and sdlc landing status/grant/revoke; atlas sdlc-binary.md command table gains a landing row.
          round: 6
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      reviewed: 2c61503b65a1ee4d777cbbda9f55d2f93959409a
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: df00763e-3916-4bd4-a9e7-9271a5976258
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: milestone-review
            round: 7
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4bffd9a6ba7776f9
            reviewed_head: 36b7cb950a0ee92184939d5f88da33c4b484ec5a
            slot: ariadne:2
            requested_at: "2026-10-16T09:38:51.974656Z"
          phase: finished
          pid: 6354
          started_at: "2026-10-16T09:38:53.678844Z"
          finished_at: "2026-10-16T09:54:08.909256Z"
          session_id: fd9fd84c-8c04-45ea-b801-6a93c05ff35d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 36b7cb950a0ee92184939d5f88da33c4b484ec5a
        tracker_head: 7605116fdd0dddc1b6d6215484a89e883956bee6
        branch_point: 655bda2e67d4fcb701adda3c88f3bc526fae1f05
        since: "2026-10-15T23:44:34-07:00"
        until: "2026-10-16T02:19:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T23:44:34-07:00"
              end: "2026-10-16T02:19:08-07:00"
              began: "2026-10-15T23:44:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 57542ed79e3a75603be304fa9a05a575f7ce08a4e3902f3ca08d54e8f50f9042
              started: "2026-10-16T06:53:29.832Z"
              spawned: false
              bytes: 600670
              windows:
                - start: "2026-10-16T06:53:29.832Z"
                  end: "2026-10-16T02:19:08-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8a2543ca4cf459ee2a92a35b940de048983b5ec788d76a30c7d329d9230ec65e
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 14081335
              windows:
                - start: "2026-10-15T23:44:34-07:00"
                  end: "2026-10-16T02:19:08-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-16T00:05:18-07:00"
              finalized: "2026-10-16T00:07:33-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:10:49-07:00"
              finalized: "2026-10-16T00:13:11-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-16T00:15:25-07:00"
              finalized: "2026-10-16T00:17:35-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-16T01:18:13-07:00"
              finalized: "2026-10-16T01:23:08-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-16T01:29:16-07:00"
              finalized: "2026-10-16T02:00:00-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-16T02:05:53-07:00"
              finalized: "2026-10-16T02:19:06-07:00"
              finalizing: true
      "n": 7
      timestamp: "2026-10-16T02:54:12-07:00"
      dispatched: "2026-10-16T02:38:51-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: Plan lines 47-59 are now marked "as built" and give Status() (head *Ticket, waiting []Ticket, err error), matching line.go. Task 1's sketch at plan:88 is marked superseded (NextHead takes no now), and Task 2 line 125 uses the as-built Status signature.
          round: 7
        - id: BR-10
          disposition: addressed
          note: 'line.go heldTicket returns Ticket{Unreadable: true} for a held but unparseable head.lock and is used in both Status and tryAdmit. TestLineUnreadableHeadIsStillAHead fails without the fix (it checks got != nil and Unreadable) and passes at head.'
          round: 7
      recipe: milestone-review
      isolation: checkout
      reviewed: 36b7cb950a0ee92184939d5f88da33c4b484ec5a
      blocked: false
---

# Gate ledger — ariadne#393 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T00:07:33-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `plan-table-drift` (kind: defect) Plan Core concepts table contradicts the M1 code (DelayStats/stats.go, NextHead signature, Ticket.PR, missing Order)
  The plan table lists DelayStats in stats.go (absent; the code has Delays in log.go), NextHead taking now (it doesn't), and Ticket with PR (it doesn't), and omits the pure Order. The deviations are only in the issue Log. Add a plan Revisions entry and update the table and Task 3's file list.
- **BR-2** [Important] `untrusted-log-entry-validation` (kind: defect) ValidGrants accepts future-dated grants, and the sweep never removes them
  order.go:59-60 checks now.Before(Expires) and the 24 h span but not Granted <= now, so Granted=2099 confers priority until 2099. The sweep test now.Sub(t) <= Retention (log.go:122) is always true for future times, so the entry is never removed. A probe test confirmed it. Fix: require !Granted.After(now), drop future-dated entries in the sweep, and add a table case.
- **BR-3** [Important] `untrusted-log-entry-validation` (kind: defect) AddGrant doesn't stamp Granted, so a revoke with a zero Granted is swept and its grant becomes valid again
  Record stamps At when it is zero but AddGrant (log.go:52) doesn't stamp Granted, and the sweep ages grants by Granted. A revoke appended with a zero Granted is dropped by the next daily sweep, and the earlier grant, still within 24 h, becomes the latest entry and valid again (reproduced with a probe). Fix: stamp Granted from l.o.Now() when it is zero, and test grant, revoke, sweep.
- **BR-4** [Minor] `untrusted-log-entry-validation` (kind: defect) scan trusts the Seq in a ticket's JSON over its file name
  line.go:262-272 uses the parsed Ticket's Seq; a hand-edited ticket can claim another position. Use the file name as the authoritative Seq.
- **BR-5** [Minor] `malformed-input-visibility` (kind: defect) sweep silently drops malformed log lines instead of counting them
  log.go:122 discards unparsable lines without a count, which contradicts the plan's skipped-and-counted rule.
- **BR-6** [Minor] `wait-termination` (kind: defect) Wait on a handle whose ticket is gone polls forever until its context ends
  line.go:150-157: if h's Seq isn't among the live tickets, pos is never 0. Return an error instead.
- **BR-7** [Minor] `ordering-test-seam` (kind: defect) Line tests rely on wall-clock timing (150 ms notAdmitted) instead of the injected Sleep seam
  The interleaving isn't controllable or reproducible (ARCH-ORDER), and the tests may be flaky under load.

## Round 2 — 2026-10-16T00:13:11-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Plan table now lists Order, Delays in log.go, NextHead without now, Ticket without PR, plus a Revisions entry. Residual Task 1 sketch and Status signature drift is noted as Minor.
- BR-2 — addressed — order.go:61 adds the Granted-not-after-now check; reverting it turns TestValidGrants red. The sweep drops future-dated entries, but that fix interacts with BR-3's Expires key and caused the new Critical finding.
- BR-3 — addressed — AddGrant stamps Granted (log.go:59); reverting it turns TestSweptRevokeKeepsTheGrantRevoked red. The revoke-outlives-grant goal is met; the new Critical finding covers the keying regression.
- BR-4 — addressed — line.go:283-291 takes Seq from the file name; TestLineTicketNumberComesFromItsName covers both a lying JSON body and garbage.
- BR-5 — addressed — The sweep stamp records malformed=N; TestLogSweepDropsFutureDatedAndCountsMalformed checks it.
- BR-6 — addressed — tryAdmit errors on a released or vanished ticket (line.go:156-169); TestLineReleasedTicketCantWait covers the released path.
- BR-7 — addressed — Line tests step tryAdmit directly and inject Sleep; helper tests sync on stdout readiness and cmd.Wait, not timers.

### Raised

- **BR-8** [Critical] `untrusted-log-entry-validation` (kind: defect) The grants sweep deletes every grant still in force: retention keys on Expires, and the sweep drops any entry dated in the future
  log.go:67-70 dates a grant by Expires when Expires > Granted, and log.go:144 keeps an entry only when 0 <= now - t <= Retention. A grant still in force has Expires > now, so whenever the daily sweep runs on an AddGrant call (a renewal, another issue's grant, a revoke), all grants still in force are erased. Probe: AddGrant(issue 1) at t0, AddGrant(issue 391, 24 h) at t0+23h, AddGrant(issue 7) at t0+25h, then ValidGrants has no 391 (FAIL). The atlas states the same rule. This is the 4th finding in family untrusted-log-entry-validation (BR-2 and BR-3 fixed instances, and this one came from the interaction of those fixes). Do not patch this instance alone. Rule: an entry's removal is decided by a pure keep(entry, now) function in the core, beside ValidGrants, and must preserve the invariant ValidGrants(sweep(L, now), now) == ValidGrants(L, now) for every log L. For grants: drop when malformed, when Granted is after now (forged), or when max(Granted, Expires) + Retention is before now. Events keep the At-based rule. Test the invariant with a table or sequence test over a grant still in force, an expired grant, a revoke, a future-dated grant and a zero-Granted entry. Also apply the rule to the untrusted seq file (line.go:303, where a parse error is silently ignored), and fix the atlas paragraph.
- **BR-9** [Minor] `plan-table-drift` (kind: defect) Plan Task 1 sketch still passes now to NextHead, and Task 2 gives a Status signature the code doesn't have
  This is the 2nd finding in family plan-table-drift. Rule: the Core concepts table and the Revisions entry are authoritative, and every signature elsewhere in the plan either matches them or is marked superseded. Here that covers the sketch at plan:80 and Status() ([]Ticket, headPID int), against the code's Status() (*Ticket, []Ticket, error).

## Round 3 — 2026-10-16T00:17:35-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — retention.go KeepGrants ages by Granted and drops an issue's whole history when its latest entry is stale. Reverting to the old expiry-based rule makes TestSweepKeepsGrantsInForce, TestKeepGrantsPreservesValidity and TestKeepGrantsCases fail. nextSeq now refuses a malformed seq (TestLineCounterIsUntrustedInput), and the atlas is fixed.
- BR-9 — not-addressed — Fixed: the plan:81 sketch is marked superseded and the Task 2 Status signature matches the code. Still contradicting the code: the Core concepts Integration row and paragraph for Log (Append(file, v) and Sweep(file, now) don't exist; the methods are on Line), tickets/<seq>.json (the code uses %020d with no extension), Open(dir) (the code takes Options), and Task 3's "Sweep drops entries older than 30 days" (superseded by KeepGrants). Apply the rule to the whole plan, not just the named sites.

### Raised

- **BR-10** [Minor] `malformed-input-visibility` (kind: defect) Status reports no head when head.lock is held but its contents don't parse
  This is the 2nd finding in this family. In line.go:205-206, readTicket returns nil for a held but unparseable head.lock, so Status (and OnWait's head, line.go:174) says "no head" while one is running. Rule: a held lock whose contents are unreadable is reported as present but unreadable, never as absent (scan already does this for tickets). Fix: return a non-nil Ticket with an unreadable marker. Test: hold head.lock with garbage and assert head != nil.

## Round 4 — 2026-10-16T01:23:08-07:00 (claude) — BLOCKED

### Raised

- **BR-11** [Important] `gate-skip-predicate-divergence` (kind: defect) anyMergedPR lets a branch with an old merged PR plus a new open PR skip the landing line and land outside it
  landing.go:366-375 skips the line if any PR for the branch is MERGED, but gh pr list --state all also returns old merged PRs, and selectLandingPR (landing.go:191) later picks the OPEN PR at the current head and runs integrate/suite/merge without a turn, breaking Done-when 1 when a branch name is reused. Fix: skip only when the PR the landing would select (selectLandingPR with the local head) is MERGED, or when no OPEN PR exists; add a fake-GitHub test with an old merged PR and an open PR on the same branch that waits behind a held line.
- **BR-12** [Important] `plan-table-drift` (kind: defect) Plan still claims a landingLineOpen seam and landing_line_test.go, and omits landingqueue.go and an M2 Revisions entry
  3rd finding in this family, so fix the rule, not this instance: every entity and seam the diff adds or renames appears in the Core concepts tables at its built path, and every as-built deviation gets a dated Revisions entry in the same milestone. Instances: plan line 54 names a package variable landingLineOpen that does not exist (the seams are landingLinePoll and the real common dir); Task 4 names landing_line_test.go but the tests are in landingqueue_test.go; landingqueue.go (enterLandingLine, landingTurn, openLandingLine, landingLineDir) is missing from the Integration points table.
- **BR-13** [Minor] `deterministic-output` (kind: defect) sdlc landing status prints grants in random map order
  landingcmd.go loops over the ValidGrants map directly; sort by issue number.
- **BR-14** [Minor] `audit-write-failure-visibility` (kind: defect) Errors from landing-line event and grant writes are ignored with no warning
  The enqueue, OnAdmit and leave writes in landingqueue.go discard errors from Record and AddGrant, so a disk that stays full or unwritable silently loses all audit and pilot data; warn once on stderr.
- **BR-15** [Minor] `pilot-metric-fidelity` (kind: defect) A duplicate merge admitted after the original landed logs a second landed event and skews the delay p90
  leave() logs EventLanded whenever err is nil, including when the admitted run only resumed cleanup of a PR that was already MERGED; use a distinct kind or skip it.

## Round 5 — 2026-10-16T02:00:00-07:00 (claude) — BLOCKED

### Disposed

- BR-11 — addressed — anyOpenPR replaces anyMergedPR; reverting it makes TestLandingLineReusedBranchStillQueues fail (no 1 waiting merges), restored it passes.
- BR-12 — addressed — Integration points table now lists landingqueue.go entities and landingcmd.go; Injected-into says seams are landingLinePoll and landingNow; dated M2 Revisions entry present.
- BR-13 — not-addressed — slices.Sorted added at landingcmd.go:182, but no status test has two or more grants, so nothing fails without the sort.
- BR-14 — not-addressed — logLine and the AddGrant warning exist but no test forces a write failure, and the GrantEntries read error in leave() still silently skips ending the grant.
- BR-15 — addressed — EventResumed logged when the selected PR was MERGED; TestLandingLineDuplicateResumesAsMerged asserts resumed==1 and landed==0, which would fail without the fix.

### Raised

- **BR-16** [Important] `gate-skip-predicate-divergence` (kind: defect) Line skip is still decided on the pre-read; a turnless run can land an OPEN PR selected from the second read
  2nd finding in this family, so state the rule: a run without a turn may only act on a PR that is MERGED when selected. landing.go decides the turn from pre (first LandingPRs) but selectIntegrationPR uses prs (second read); an OPEN PR in the second read lands with turn nil. Fix: after selection, refuse (or enter the line) when turn is nil and pr.State is not MERGED; test with a fake whose first LandingPRs omits the open PR.
- **BR-17** [Important] `user-surface-docs` (kind: defect) README update appears missing for sdlc landing status/grant/revoke and for merge waiting in the landing line
  README.md:414-425 describes sdlc merge landing and concurrency but not the line, its waiting output, or the new sdlc landing subcommand; the atlas subcommand table near sdlc-binary.md:43 also lacks a landing row.

## Round 6 — 2026-10-16T02:19:06-07:00 (claude) — passed

### Disposed

- BR-13 — addressed — landingcmd.go:177 iterates slices.Sorted(maps.Keys(grants)); landingcmd_test.go asserts #12 precedes #391 across repeated runs.
- BR-14 — addressed — lineLog.record (landingqueue.go:95) warns once per merge for enqueue/admit/leave writes; leave warns on grant read and AddGrant failures; TestLandingLineLogFailuresWarn pins it.
- BR-16 — addressed — Rule enforced on the selected PR at landing.go:394; reverting the guard turns TestLandingLineLaggingListingRefusesOutsideTheLine red (err nil); empty listing now queues.
- BR-17 — addressed — README.md:423-431 covers the line, waiting output and sdlc landing status/grant/revoke; atlas sdlc-binary.md command table gains a landing row.

## Round 7 — 2026-10-16T02:54:12-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — Plan lines 47-59 are now marked "as built" and give Status() (head *Ticket, waiting []Ticket, err error), matching line.go. Task 1's sketch at plan:88 is marked superseded (NextHead takes no now), and Task 2 line 125 uses the as-built Status signature.
- BR-10 — addressed — line.go heldTicket returns Ticket{Unreadable: true} for a held but unparseable head.lock and is used in both Status and tryAdmit. TestLineUnreadableHeadIsStillAHead fails without the fix (it checks got != nil and Unreadable) and passes at head.

## Open findings

(none — every finding has been disposed)
