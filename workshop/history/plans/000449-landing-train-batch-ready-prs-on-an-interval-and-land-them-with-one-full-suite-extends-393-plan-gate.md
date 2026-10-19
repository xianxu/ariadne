---
reviewer_runs:
    - version: 1
      intent:
        id: c734e5c7-cd3b-43fc-8c42-3754fed73698
        repository: github.com/xianxu/ariadne
        issue_id: "000449"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: beba015dae7e935995c457b231d5d72452ba2860
        slot: ariadne:2
        requested_at: "2026-10-19T07:41:25.853116Z"
      phase: finished
      pid: 41424
      started_at: "2026-10-19T07:41:26.025352Z"
      finished_at: "2026-10-19T07:41:56.541325Z"
      session_id: 0b8dd07a-725f-46d5-9d61-2f1c3a639ae1
      outcome: returned
gate: plan-quality
issue: 449
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: c89501a1-4004-4a58-9e28-d0fb9f7fa808
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: f76d84ed5d1867a0dac6f9f8f02291246ecbe850
            slot: ariadne:2
            requested_at: "2026-10-19T07:37:19.863711Z"
          phase: finished
          pid: 72469
          started_at: "2026-10-19T07:37:20.076694Z"
          finished_at: "2026-10-19T07:38:20.003631Z"
          session_id: ce283a71-a4ff-4b46-a444-7e71d73538b9
          outcome: returned
      "n": 1
      timestamp: "2026-10-19T00:38:20-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: The Revision says .sdlc/ is not guarded, but local-gates-guard.yml protects '.sdlc/*'
          detail: .github/workflows/local-gates-guard.yml lists '.sdlc/*' just after the cited lines 56-68, and the workflow is base-layer seeded (construct/base.manifest:138). A pilot PR that adds .sdlc/landing.json fails the guard and needs the operator's local-gates-verified label. Either say that the pilot and every interval change are operator-verified local-gates changes, or move the config and justify it.
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: A tickets/<seq>.ready sidecar is deleted by Line.scan on the next poll because nobody holds a flock on it
          detail: scan (internal/landingline/line.go:261-301) tries a flock on every file in tickets/ and removes any file it can lock (:274-278). It checks the name format only afterwards (:291) and skips only .tmp- files before that. Put readiness in its own directory with its own sweep, or hold a flock on it, and add a test that the sidecar survives a scan while its ticket is held.
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: unverifiable
          id: PQ-3
          severity: Important
          title: Car resume requires pr.MergeOID to be an ancestor of main, but mergeCommit for a PR merged by reachability is unverified and the fake assumes the optimistic value
          detail: landing.go:581 checks merge-base --is-ancestor pr.MergeOID main. GitHub's mergeCommit can be null (ghlanding.go:114), and its value for a PR closed by reachability is what Q7 left open. Conformance comes only from the first pilot train, after landing. Make resume accept HeadOID reachable from main when MergeOID is empty or not the train commit, and have the fake exercise the null and head-commit variants.
          family: fake-encodes-unverified-external-behavior
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 5499d218-356a-4b8f-83c4-16e9f8ed4217
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: b1dbeec8eba3f22f1dbbf71be18106318e7630b3
            slot: ariadne:2
            requested_at: "2026-10-19T07:39:04.815391Z"
          phase: finished
          pid: 51929
          started_at: "2026-10-19T07:39:05.139525Z"
          finished_at: "2026-10-19T07:39:35.856659Z"
          session_id: 3859e47f-fab7-4809-b00b-39545e065c4d
          outcome: returned
      "n": 2
      timestamp: "2026-10-19T00:39:36-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Revision now states .sdlc/* is guarded (local-gates-guard.yml:69) and pilot/interval changes are operator-verified local-gates changes.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Sidecar moved to its own ready/ directory with a ticket-absent sweep and a survives-scan test; scan only reads tickets/.
          round: 2
        - id: PQ-3
          disposition: not-addressed
          note: Only the landing.go:581 gate was swept. landingarchive.go:162 (landingOIDValid(MergeOID) refuses on null), landingarchive.go:634 (MergeOID ancestor of tip), and trackercompletion.go:249 (card landed commit = MergeOID) still consume MergeOID on the car resume path that the plan says runs unchanged. Enumerate all MergeOID consumers, route them through one resolved landed commit from landedOnMain, and drive the fake's null/head variants through archive and card completion.
          round: 2
      findings:
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Sidecar lock named line.lock in the integration table but line.mutex in the lifecycle section
          detail: Same lock (line.go:21); use one name so the implementer does not look for two.
          family: plan-internal-inconsistency
          round: 2
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 51919696-49ec-4275-97ea-b8ab8bcd4613
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 6539030fef6c2fe3979431d8ca71132e18124819
            slot: ariadne:2
            requested_at: "2026-10-19T07:40:02.032942Z"
          phase: finished
          pid: 82836
          started_at: "2026-10-19T07:40:02.254023Z"
          finished_at: "2026-10-19T07:40:18.496954Z"
          session_id: ed119f61-f721-4c11-b1f4-f27f6ddaae4c
          outcome: returned
      "n": 3
      timestamp: "2026-10-19T00:40:19-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-3
          disposition: addressed
          note: A pure landedCommit resolver feeds LandedOID; all 4 non-test MergeOID consumers (verified by grep) route through it; the fake covers the null, car-head and train-merge variants end to end, and a grep test enforces the rule.
          round: 3
        - id: PQ-4
          disposition: addressed
          note: Both sections now say the line mutex (line.lock), matching line.go:21.
          round: 3
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 6288d619-e240-4e11-976a-17e9f745b22e
            repository: github.com/xianxu/ariadne
            issue_id: "000449"
            recipe: plan-quality
            round: 4
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: beba015dae7e935995c457b231d5d72452ba2860
            slot: ariadne:2
            requested_at: "2026-10-19T07:41:03.648228Z"
          phase: finished
          pid: 18597
          started_at: "2026-10-19T07:41:03.843583Z"
          finished_at: "2026-10-19T07:41:21.854747Z"
          session_id: ea85e2ea-80c1-424b-b626-ad5588a39352
          outcome: returned
      "n": 4
      timestamp: "2026-10-19T00:41:22-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Prior round resolved; the Revisions record .sdlc/* as guarded.
          round: 4
        - id: PQ-2
          disposition: addressed
          note: The sidecar now lives in ready/, outside tickets/, and scan sweeps it.
          round: 4
        - id: PQ-3
          disposition: addressed
          note: 'Verified: the landedCommit resolver list matches grep of the non-test MergeOID consumers exactly.'
          round: 4
        - id: PQ-4
          disposition: addressed
          note: The lock is named line.lock consistently.
          round: 4
      blocked: false
content_hash: eef871fdf10282935bfb7a474a0774b1a0e9817672dd6eda5c70658efdf56b2a
---

# Gate ledger — ariadne#449 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T00:38:20-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` (kind: defect) The Revision says .sdlc/ is not guarded, but local-gates-guard.yml protects '.sdlc/*'
  .github/workflows/local-gates-guard.yml lists '.sdlc/*' just after the cited lines 56-68, and the workflow is base-layer seeded (construct/base.manifest:138). A pilot PR that adds .sdlc/landing.json fails the guard and needs the operator's local-gates-verified label. Either say that the pilot and every interval change are operator-verified local-gates changes, or move the config and justify it.
- **PQ-2** [Important] `unbacked-existing-behavior-claim` (kind: defect) A tickets/<seq>.ready sidecar is deleted by Line.scan on the next poll because nobody holds a flock on it
  scan (internal/landingline/line.go:261-301) tries a flock on every file in tickets/ and removes any file it can lock (:274-278). It checks the name format only afterwards (:291) and skips only .tmp- files before that. Put readiness in its own directory with its own sweep, or hold a flock on it, and add a test that the sidecar survives a scan while its ticket is held.
- **PQ-3** [Important] `fake-encodes-unverified-external-behavior` (kind: unverifiable) Car resume requires pr.MergeOID to be an ancestor of main, but mergeCommit for a PR merged by reachability is unverified and the fake assumes the optimistic value
  landing.go:581 checks merge-base --is-ancestor pr.MergeOID main. GitHub's mergeCommit can be null (ghlanding.go:114), and its value for a PR closed by reachability is what Q7 left open. Conformance comes only from the first pilot train, after landing. Make resume accept HeadOID reachable from main when MergeOID is empty or not the train commit, and have the fake exercise the null and head-commit variants.

## Round 2 — 2026-10-19T00:39:36-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — Revision now states .sdlc/* is guarded (local-gates-guard.yml:69) and pilot/interval changes are operator-verified local-gates changes.
- PQ-2 — addressed — Sidecar moved to its own ready/ directory with a ticket-absent sweep and a survives-scan test; scan only reads tickets/.
- PQ-3 — not-addressed — Only the landing.go:581 gate was swept. landingarchive.go:162 (landingOIDValid(MergeOID) refuses on null), landingarchive.go:634 (MergeOID ancestor of tip), and trackercompletion.go:249 (card landed commit = MergeOID) still consume MergeOID on the car resume path that the plan says runs unchanged. Enumerate all MergeOID consumers, route them through one resolved landed commit from landedOnMain, and drive the fake's null/head variants through archive and card completion.

### Raised

- **PQ-4** [Minor] `plan-internal-inconsistency` (kind: defect) Sidecar lock named line.lock in the integration table but line.mutex in the lifecycle section
  Same lock (line.go:21); use one name so the implementer does not look for two.

## Round 3 — 2026-10-19T00:40:19-07:00 (claude) — passed

### Disposed

- PQ-3 — addressed — A pure landedCommit resolver feeds LandedOID; all 4 non-test MergeOID consumers (verified by grep) route through it; the fake covers the null, car-head and train-merge variants end to end, and a grep test enforces the rule.
- PQ-4 — addressed — Both sections now say the line mutex (line.lock), matching line.go:21.

## Round 4 — 2026-10-19T00:41:22-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Prior round resolved; the Revisions record .sdlc/* as guarded.
- PQ-2 — addressed — The sidecar now lives in ready/, outside tickets/, and scan sweeps it.
- PQ-3 — addressed — Verified: the landedCommit resolver list matches grep of the non-test MergeOID consumers exactly.
- PQ-4 — addressed — The lock is named line.lock consistently.

## Open findings

(none — every finding has been disposed)
