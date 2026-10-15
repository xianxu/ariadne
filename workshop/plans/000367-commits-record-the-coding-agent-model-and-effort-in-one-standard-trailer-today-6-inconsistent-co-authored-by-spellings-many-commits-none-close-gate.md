---
gate: boundary-review
issue: 367
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a1ca04c516becee90f7d9f9f38104baee5d6cdea
        tracker_head: f2ce7c5302e7e055d5746fd427ce5a898614c9dc
        branch_point: 18915f6c17d9b4aeb5e618d3f736705ec355d2b3
        since: "2026-10-14T20:40:03-07:00"
        until: "2026-10-14T22:39:09-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T20:40:03-07:00"
              end: "2026-10-14T22:39:09-07:00"
              began: "2026-10-14T20:40:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 62c9c8aef57f0e3f2f1b851e1b5efdd09fd28096545300745aa661d2a93a8f4d
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 8444805
              windows:
                - start: "2026-10-14T20:40:03-07:00"
                  end: "2026-10-14T22:39:09-07:00"
      "n": 1
      timestamp: "2026-10-14T22:47:13-07:00"
      dispatched: "2026-10-14T22:39:52-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: gitx calls agentid.StampCurrent (env + transcript reads) instead of taking the identity as a parameter, so gitx now depends on judge/reviewbox/estimate
          detail: 'commitpublication.go:217 and refbootstrap.go:178 call agentid.StampCurrent, against the plan''s "gitx gets the identity as a parameter (no env reads inside gitx)". go list -deps for gitx went from processgroup+workspace to also judge, reviewbox, reviewprepare, estimate, agentid, which also stops agentid reusing transcripts'' roots (cycle). Fix: inject a StampMessage func or Identity on TrunkFile from trackerenv, or split agentid''s pure core from detection so gitx imports no judge (ARCH-PURE).'
          family: io-seam-injected-not-reached-for
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Detector.Detect rebuilds judge.AgentDefaultEnv and hardcodes transcript roots already defined in judge and transcripts
          detail: detect.go:657-663 duplicates judge.CurrentAgentDefaultEnv field by field; add judge.AgentDefaultEnvFrom(getenv) and use it from both. The ~/.claude/projects and ~/.codex/sessions roots duplicate transcripts/defaults.go (ARCH-DRY).
          family: shared-helper-not-reused
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: 'Plan text drifts from the code: 256 KiB vs 1 MiB tail, gh pr merge body listed as stamped, testdata fixtures, gitx parameter, unticked tasks'
          detail: Add a Revisions section to workshop/plans/000367-agent-trailer-plan.md recording what the code actually does.
          family: plan-claims-drift-from-code
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Plan Task 4 promised trailer assertions for claim sync, changecode checkpoint and the landing merge message; none were added
          detail: The sweep test covers stamping at the source-line level, but there is no behavioural assertion for these three sites, and none for the close.go:663 wiring of reportMissingAgent.
          family: plan-promised-test-missing
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a6da38026389acf2358c0132581d61e9cbf6bd63
        tracker_head: a088fcec554386d7b949887787e7514ebd3b923b
        branch_point: 18915f6c17d9b4aeb5e618d3f736705ec355d2b3
        since: "2026-10-14T20:40:03-07:00"
        until: "2026-10-14T22:52:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T20:40:03-07:00"
              end: "2026-10-14T22:52:11-07:00"
              began: "2026-10-14T20:40:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ef7294f56c294e65cfab0c4276dbd49f1ddf4cf6c32a5ac931efc0b103711a25
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 8830527
              windows:
                - start: "2026-10-14T20:40:03-07:00"
                  end: "2026-10-14T22:52:11-07:00"
        cuts:
            - dispatched: "2026-10-14T22:39:52-07:00"
              finalized: "2026-10-14T22:47:13-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-14T23:39:58-07:00"
      dispatched: "2026-10-14T23:28:42-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: agentid is now standard-library only, and go list -deps for gitx adds only agentid (judge, reviewbox, reviewprepare and estimate are gone). The parameter-vs-StampCurrent decision is recorded with its reasons in the plan Revisions, and tests are hermetic through testfix.PinAgent in the gitx, tracker and cmd/sdlc TestMains.
          round: 2
        - id: BR-2
          disposition: addressed
          note: agentid.EnvFrom is the single agent-env parser (judge.CurrentAgentDefaultEnv delegates to it, env.go:24). ClaudeProjects and CodexSessions are used by transcripts/defaults.go, processmanual.go, processmanual/session.go and processmanual/memory.go, and grep finds no remaining production literals.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The plan Revisions now record the 1 MiB tail, the gh merge exemption, the inline fixtures, gitx using StampCurrent, and PAIR_AGENT-only shells reading as sdlc. All tasks are ticked, and each claim matches detect.go, env.go and agentstamp_test.go.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: 'The change-code checkpoint (verified red when its stamp is reverted), landing merge and close-report wiring now have assertions. The claim-sync stamp at claim.go:597 does not: with it removed, -run ''Sync|Claim|IssueNew|Publish'' showed no related failures, because the issue-sync test checks origin/main, which gitx.commitTree stamps again.'
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: a6da38026389acf2358c0132581d61e9cbf6bd63
      blocked: false
---

# Gate ledger — ariadne#367 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T22:47:13-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `io-seam-injected-not-reached-for` (kind: defect) gitx calls agentid.StampCurrent (env + transcript reads) instead of taking the identity as a parameter, so gitx now depends on judge/reviewbox/estimate
  commitpublication.go:217 and refbootstrap.go:178 call agentid.StampCurrent, against the plan's "gitx gets the identity as a parameter (no env reads inside gitx)". go list -deps for gitx went from processgroup+workspace to also judge, reviewbox, reviewprepare, estimate, agentid, which also stops agentid reusing transcripts' roots (cycle). Fix: inject a StampMessage func or Identity on TrunkFile from trackerenv, or split agentid's pure core from detection so gitx imports no judge (ARCH-PURE).
- **BR-2** [Minor] `shared-helper-not-reused` (kind: defect) Detector.Detect rebuilds judge.AgentDefaultEnv and hardcodes transcript roots already defined in judge and transcripts
  detect.go:657-663 duplicates judge.CurrentAgentDefaultEnv field by field; add judge.AgentDefaultEnvFrom(getenv) and use it from both. The ~/.claude/projects and ~/.codex/sessions roots duplicate transcripts/defaults.go (ARCH-DRY).
- **BR-3** [Minor] `plan-claims-drift-from-code` (kind: defect) Plan text drifts from the code: 256 KiB vs 1 MiB tail, gh pr merge body listed as stamped, testdata fixtures, gitx parameter, unticked tasks
  Add a Revisions section to workshop/plans/000367-agent-trailer-plan.md recording what the code actually does.
- **BR-4** [Minor] `plan-promised-test-missing` (kind: defect) Plan Task 4 promised trailer assertions for claim sync, changecode checkpoint and the landing merge message; none were added
  The sweep test covers stamping at the source-line level, but there is no behavioural assertion for these three sites, and none for the close.go:663 wiring of reportMissingAgent.

## Round 2 — 2026-10-14T23:39:58-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — agentid is now standard-library only, and go list -deps for gitx adds only agentid (judge, reviewbox, reviewprepare and estimate are gone). The parameter-vs-StampCurrent decision is recorded with its reasons in the plan Revisions, and tests are hermetic through testfix.PinAgent in the gitx, tracker and cmd/sdlc TestMains.
- BR-2 — addressed — agentid.EnvFrom is the single agent-env parser (judge.CurrentAgentDefaultEnv delegates to it, env.go:24). ClaudeProjects and CodexSessions are used by transcripts/defaults.go, processmanual.go, processmanual/session.go and processmanual/memory.go, and grep finds no remaining production literals.
- BR-3 — addressed — The plan Revisions now record the 1 MiB tail, the gh merge exemption, the inline fixtures, gitx using StampCurrent, and PAIR_AGENT-only shells reading as sdlc. All tasks are ticked, and each claim matches detect.go, env.go and agentstamp_test.go.
- BR-4 — not-addressed — The change-code checkpoint (verified red when its stamp is reverted), landing merge and close-report wiring now have assertions. The claim-sync stamp at claim.go:597 does not: with it removed, -run 'Sync|Claim|IssueNew|Publish' showed no related failures, because the issue-sync test checks origin/main, which gitx.commitTree stamps again.

## Open findings

- **BR-4** [Minor] `plan-promised-test-missing` (kind: defect) Plan Task 4 promised trailer assertions for claim sync, changecode checkpoint and the landing merge message; none were added
