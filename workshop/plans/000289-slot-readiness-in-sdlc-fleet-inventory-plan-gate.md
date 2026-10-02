---
gate: plan-quality
issue: 289
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-02T15:50:46-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Dependency clones are not inventory rows; the member-to-row join makes every :N slot unknown
          detail: CollectInventory enumerates only FleetRepoDirs siblings (worktree/ excluded, discover.go:75-80) and their `git worktree list`; weave's dependency clones (worktree/pair-slot1/ariadne, own .git dir) never become rows. The plan must pick and state the collection seam for non-row members (extend rows vs. slot-local CollectFacts + branch/issue association) and drop the "member not in rows -> unknown" test as written.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-2
          severity: Minor
          title: Unioning the operation-marker lists changes landing and move-detail refusal behaviour without saying so
          detail: landing gains REBASE_HEAD/BISECT_LOG, move-detail gains sequencer/BISECT_START, and Stat vs Lstat differ; state this as an intended change in the plan.
          family: unstated-behavior-change
          round: 1
        - id: PQ-3
          severity: Minor
          title: Dependency clones rest on main only when weave's scoped policy applies
          detail: acquire.go:87,152 passes --branch main only if scoped; otherwise the clone is on the remote default branch. Either derive the resting branch or document the assumption.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-4
          severity: Minor
          title: Tasks 3-6 enumerate test cases in prose; reduce to one strategy line per risky function
          family: test-prose-enumeration
          round: 1
        - id: PQ-5
          severity: Minor
          title: Substrate paths read from member construct/deps are not confined to the environment root
          family: untrusted-path-confinement
          round: 1
        - id: PQ-6
          severity: Minor
          title: issueOpen should derive terminality from the vocab model via rows' DeclaredStatus, not an ad-hoc predicate
          family: single-source-vocab
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-02T16:23:43-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Dependency clones are collected as rows via collectInventoryRepo; member-not-a-row now means collection failed, so unknown is correct.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Decisions state the landing and move-detail behavior change and that all markers use lstat; matches issuemovedetail.go:71 and landing.go:126.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: environment.go:43 always sets Policy, so the scoped branch at acquire.go:87,152 applies to weave slots.
          round: 2
        - id: PQ-4
          disposition: not-addressed
          note: The M1 Verdict and M2 Wiring bullets still enumerate case cross-products in prose instead of one strategy line per risky function.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: A member outside the environment root is unknown, checked in DeclaredMembers.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Uses vocab.Issue().IsTerminal(DeclaredStatus) on the existing row association.
          round: 2
      findings:
        - id: PQ-7
          severity: Minor
          title: Goal line still promises needs-recovery paths, contradicting the reasons-not-paths decision
          detail: The plan header Goal says "with the paths behind a needs-recovery verdict"; Decisions and Done-when say reasons only. Update the Goal so implementers do not reintroduce path lists.
          family: plan-self-consistency
          round: 2
      blocked: false
content_hash: e8a77bdfe61ee028b0d6dd22320c5caf6d4ebf3b01347429b6e404b251521919
---

# Gate ledger — ariadne#289 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T15:50:46-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` Dependency clones are not inventory rows; the member-to-row join makes every :N slot unknown
  CollectInventory enumerates only FleetRepoDirs siblings (worktree/ excluded, discover.go:75-80) and their `git worktree list`; weave's dependency clones (worktree/pair-slot1/ariadne, own .git dir) never become rows. The plan must pick and state the collection seam for non-row members (extend rows vs. slot-local CollectFacts + branch/issue association) and drop the "member not in rows -> unknown" test as written.
- **PQ-2** [Minor] `unstated-behavior-change` Unioning the operation-marker lists changes landing and move-detail refusal behaviour without saying so
  landing gains REBASE_HEAD/BISECT_LOG, move-detail gains sequencer/BISECT_START, and Stat vs Lstat differ; state this as an intended change in the plan.
- **PQ-3** [Minor] `unbacked-existing-behavior-claim` Dependency clones rest on main only when weave's scoped policy applies
  acquire.go:87,152 passes --branch main only if scoped; otherwise the clone is on the remote default branch. Either derive the resting branch or document the assumption.
- **PQ-4** [Minor] `test-prose-enumeration` Tasks 3-6 enumerate test cases in prose; reduce to one strategy line per risky function
- **PQ-5** [Minor] `untrusted-path-confinement` Substrate paths read from member construct/deps are not confined to the environment root
- **PQ-6** [Minor] `single-source-vocab` issueOpen should derive terminality from the vocab model via rows' DeclaredStatus, not an ad-hoc predicate

## Round 2 — 2026-10-02T16:23:43-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Dependency clones are collected as rows via collectInventoryRepo; member-not-a-row now means collection failed, so unknown is correct.
- PQ-2 — addressed — Decisions state the landing and move-detail behavior change and that all markers use lstat; matches issuemovedetail.go:71 and landing.go:126.
- PQ-3 — addressed — environment.go:43 always sets Policy, so the scoped branch at acquire.go:87,152 applies to weave slots.
- PQ-4 — not-addressed — The M1 Verdict and M2 Wiring bullets still enumerate case cross-products in prose instead of one strategy line per risky function.
- PQ-5 — addressed — A member outside the environment root is unknown, checked in DeclaredMembers.
- PQ-6 — addressed — Uses vocab.Issue().IsTerminal(DeclaredStatus) on the existing row association.

### Raised

- **PQ-7** [Minor] `plan-self-consistency` Goal line still promises needs-recovery paths, contradicting the reasons-not-paths decision
  The plan header Goal says "with the paths behind a needs-recovery verdict"; Decisions and Done-when say reasons only. Update the Goal so implementers do not reintroduce path lists.

## Open findings

- **PQ-4** [Minor] `test-prose-enumeration` Tasks 3-6 enumerate test cases in prose; reduce to one strategy line per risky function
- **PQ-7** [Minor] `plan-self-consistency` Goal line still promises needs-recovery paths, contradicting the reasons-not-paths decision
