---
reviewer_runs:
    - version: 1
      intent:
        id: 73865fc5-23c0-4b57-a39f-b00277b7e200
        repository: github.com/xianxu/ariadne
        issue_id: "000393"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: 74cac2268f04e67e835bf3ef4caa9a77aa6c3882
        slot: ariadne:2
        requested_at: "2026-10-16T06:57:36.969325Z"
      phase: finished
      pid: 56260
      started_at: "2026-10-16T06:57:37.222241Z"
      finished_at: "2026-10-16T06:57:58.267051Z"
      session_id: 998fb2c5-cc40-4ff2-b389-9a0687cf6168
      outcome: returned
gate: plan-quality
issue: 393
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: cd5fee2f-073e-44d6-af6d-458a1b26b2c7
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 49974826f4b8aea3df76b06b1dd2c6e7651d8e08
            slot: ariadne:2
            requested_at: "2026-10-16T06:53:27.049349Z"
          phase: finished
          pid: 28687
          started_at: "2026-10-16T06:53:27.275688Z"
          finished_at: "2026-10-16T06:54:36.355705Z"
          session_id: 42a78a01-c26e-45f2-abad-7a5f8c34d21a
          outcome: returned
      "n": 1
      timestamp: "2026-10-15T23:54:37-07:00"
      agent: claude
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Admission happens after PR selection; the admitted merge acts on a PR/head observation taken before an unbounded wait
          detail: The plan places line.Wait inside the pr.State != MERGED branch (landing.go:378), after LandingPRs and selectIntegrationPR (landing.go:372-376). A duplicate that waited while the original landed, or a PR pushed or closed during the wait, then hits revalidateLanding or the remote-head mismatch (landing.go:413-415) instead of resuming as MERGED. State that after admission merge re-reads LandingPRs and re-runs selectIntegrationPR, taking the MERGED path if the PR landed meanwhile, and test a duplicate whose original lands during the wait.
          family: stale-observation-after-wait
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: issueOfBranch is said to reuse a parse in landingIssueBranch, which parses no issue ID; the helper is issue.IDFromFilename
          detail: landing.go:155-174 has no issue parse. tickOnLanding uses issue.IDFromFilename(branch) (projectlandtick.go:224). Key tickets and grants on the normalised int so --issue 0391 and 391 agree.
          family: unbacked-existing-code-claim
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: New line_unix.go flock wrapper duplicates internal/reviewbox/lock_unix.go lockExclusive (ARCH-DRY)
          detail: 'Lift the reviewbox flock helper (and its lock_other.go counterpart) into a shared place instead of writing a second one. Also note that #370''s repo lock is mkdir-based (repolock.go:161), not flock.'
          family: reuse-existing-helper
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Grant authority is checked only by the grant verb; ValidGrants trusts any well-formed grants.jsonl line
          detail: A worker can append to the log directly, or run from the :0 checkout without COUCH_TREE. State that the lane is cooperative, or have ValidGrants reject entries whose recorded grantor is a numbered slot, so the "unauthorized grants confer nothing" Done-when holds as stated.
          family: write-side-only-authorization
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: c641e01b-100f-44b7-860b-3d181bbb3544
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: ef566aa63f9e724fac80cc5c66489eafd2ab0a42
            slot: ariadne:2
            requested_at: "2026-10-16T06:55:35.360058Z"
          phase: finished
          pid: 45858
          started_at: "2026-10-16T06:55:35.73447Z"
          finished_at: "2026-10-16T06:56:09.929153Z"
          session_id: e2446d7b-722e-4c66-8c46-6cf6e8d8dfdd
          outcome: returned
      "n": 2
      timestamp: "2026-10-15T23:56:11-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Task 4 moves enqueue/wait before the head read, LandingPRs and selectIntegrationPR (landing.go:357-368), states the class rule, and tests duplicate-while-waiting and pushed-while-waiting.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Ticket.Issue uses issue.IDFromFilename (scaffold.go:41), normalised int.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Task 2 lifts reviewbox lock_unix.go/lock_other.go into internal/flock; atlas notes repolock is mkdir-based.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: ValidGrants rejects numbered-slot grantors on read; cooperative trust model stated; read-side test in Task 2.
          round: 2
      findings:
        - id: PQ-5
          severity: Minor
          title: Plan code sketches disagree on grant map key type, Grantor check and Line.Wait signature
          detail: NextHead takes map[string]Grant while ValidGrants returns map[int]Grant and Ticket.Issue is int; the ValidGrants filter compares the Grantor struct to ""; Line.Wait is described with grantsFn and poll args but called as line.Wait(ctx, h) in Task 4. Pick one signature per entity.
          family: sketch-signature-drift
          round: 2
        - id: PQ-6
          severity: Minor
          title: Interactive confirmation (landing.go:372) would run after admission, holding the head while waiting for an answer
          detail: Without --yes on a TTY, the operator is prompted only after the possibly long wait, while the line's head is held. Ask before enqueueing, or state that holding the head during the prompt is acceptable.
          family: stale-observation-after-wait
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 071df965-3862-4b3c-8b00-f13bbf4dfdac
            repository: github.com/xianxu/ariadne
            issue_id: "000393"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 74cac2268f04e67e835bf3ef4caa9a77aa6c3882
            slot: ariadne:2
            requested_at: "2026-10-16T06:57:11.55969Z"
          phase: finished
          pid: 52661
          started_at: "2026-10-16T06:57:11.751063Z"
          finished_at: "2026-10-16T06:57:35.226944Z"
          session_id: 2d848119-cc8d-4167-80b4-d976b280907d
          outcome: returned
      "n": 3
      timestamp: "2026-10-15T23:57:35-07:00"
      agent: claude
      dispose:
        - id: PQ-5
          disposition: addressed
          note: 'Fixed: NextHead, ValidGrants and Ticket.Issue all use map[int]Grant and an int issue; the grantor check uses Grantor.Address; Wait(ctx, h) takes its options from Open.'
          round: 3
        - id: PQ-6
          disposition: addressed
          note: Task 4 moves the prompt ahead of the enqueue. The prompt text at landing.go:373 is fixed and doesn't depend on PR selection.
          round: 3
      blocked: false
content_hash: 16c7530a3a7af78e4ff32d7ce57dc3562361c861c359b89877d860894a1b60af
---

# Gate ledger — ariadne#393 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T23:54:37-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `stale-observation-after-wait` (kind: defect) Admission happens after PR selection; the admitted merge acts on a PR/head observation taken before an unbounded wait
  The plan places line.Wait inside the pr.State != MERGED branch (landing.go:378), after LandingPRs and selectIntegrationPR (landing.go:372-376). A duplicate that waited while the original landed, or a PR pushed or closed during the wait, then hits revalidateLanding or the remote-head mismatch (landing.go:413-415) instead of resuming as MERGED. State that after admission merge re-reads LandingPRs and re-runs selectIntegrationPR, taking the MERGED path if the PR landed meanwhile, and test a duplicate whose original lands during the wait.
- **PQ-2** [Minor] `unbacked-existing-code-claim` (kind: defect) issueOfBranch is said to reuse a parse in landingIssueBranch, which parses no issue ID; the helper is issue.IDFromFilename
  landing.go:155-174 has no issue parse. tickOnLanding uses issue.IDFromFilename(branch) (projectlandtick.go:224). Key tickets and grants on the normalised int so --issue 0391 and 391 agree.
- **PQ-3** [Minor] `reuse-existing-helper` (kind: defect) New line_unix.go flock wrapper duplicates internal/reviewbox/lock_unix.go lockExclusive (ARCH-DRY)
  Lift the reviewbox flock helper (and its lock_other.go counterpart) into a shared place instead of writing a second one. Also note that #370's repo lock is mkdir-based (repolock.go:161), not flock.
- **PQ-4** [Minor] `write-side-only-authorization` (kind: defect) Grant authority is checked only by the grant verb; ValidGrants trusts any well-formed grants.jsonl line
  A worker can append to the log directly, or run from the :0 checkout without COUCH_TREE. State that the lane is cooperative, or have ValidGrants reject entries whose recorded grantor is a numbered slot, so the "unauthorized grants confer nothing" Done-when holds as stated.

## Round 2 — 2026-10-15T23:56:11-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Task 4 moves enqueue/wait before the head read, LandingPRs and selectIntegrationPR (landing.go:357-368), states the class rule, and tests duplicate-while-waiting and pushed-while-waiting.
- PQ-2 — addressed — Ticket.Issue uses issue.IDFromFilename (scaffold.go:41), normalised int.
- PQ-3 — addressed — Task 2 lifts reviewbox lock_unix.go/lock_other.go into internal/flock; atlas notes repolock is mkdir-based.
- PQ-4 — addressed — ValidGrants rejects numbered-slot grantors on read; cooperative trust model stated; read-side test in Task 2.

### Raised

- **PQ-5** [Minor] `sketch-signature-drift` Plan code sketches disagree on grant map key type, Grantor check and Line.Wait signature
  NextHead takes map[string]Grant while ValidGrants returns map[int]Grant and Ticket.Issue is int; the ValidGrants filter compares the Grantor struct to ""; Line.Wait is described with grantsFn and poll args but called as line.Wait(ctx, h) in Task 4. Pick one signature per entity.
- **PQ-6** [Minor] `stale-observation-after-wait` Interactive confirmation (landing.go:372) would run after admission, holding the head while waiting for an answer
  Without --yes on a TTY, the operator is prompted only after the possibly long wait, while the line's head is held. Ask before enqueueing, or state that holding the head during the prompt is acceptable.

## Round 3 — 2026-10-15T23:57:35-07:00 (claude) — passed

### Disposed

- PQ-5 — addressed — Fixed: NextHead, ValidGrants and Ticket.Issue all use map[int]Grant and an int issue; the grantor check uses Grantor.Address; Wait(ctx, h) takes its options from Open.
- PQ-6 — addressed — Task 4 moves the prompt ahead of the enqueue. The prompt text at landing.go:373 is fixed and doesn't depend on PR selection.

## Open findings

(none — every finding has been disposed)
