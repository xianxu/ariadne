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

## Open findings

- **BR-9** [Minor] `plan-table-drift` (kind: defect) Plan Task 1 sketch still passes now to NextHead, and Task 2 gives a Status signature the code doesn't have
- **BR-10** [Minor] `malformed-input-visibility` (kind: defect) Status reports no head when head.lock is held but its contents don't parse
