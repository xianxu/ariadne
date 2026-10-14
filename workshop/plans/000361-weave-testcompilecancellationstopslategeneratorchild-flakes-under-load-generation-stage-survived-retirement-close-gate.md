---
gate: boundary-review
issue: 361
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T22:02:30-07:00"
      dispatched: "2026-10-13T21:47:24-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: procgroup package doc claims to be "the one place" weave and sdlc read the kernel's kill answer
          detail: pkg/procgroup/procgroup.go:4 overstates it. weave staging/stage.go:85 and sdlc repolock.go:176 also read kill(pid,0) answers, for single-process liveness checks where EPERM means alive. Narrow the claim to process-group signals. No other count or "only" claim in the window is wrong; the atlas text at atlas/workflow/weave.md:353-358 checks out.
          family: doc-claim-accuracy
          round: 1
        - id: BR-2
          severity: Minor
          title: Plan item 2 names a killGroup seam and EPERM-equivalent fake kill that were never built
          detail: The finishProducers test uses a real zombie group through procgroup.Kill, not a seam. Add a Revisions entry so the plan describes the delivered design.
          family: plan-matches-code
          round: 1
        - id: BR-3
          severity: Minor
          title: ZombieGroup doc comment has one line much longer than the rest
          detail: pkg/procgroup/procgrouptest/zombie.go:16-17. It is the only such line in the window.
          family: comment-formatting
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 1f174297427fced9cc9138068030fd1a315bc80a
      blocked: false
---

# Gate ledger — ariadne#361 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T22:02:30-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-accuracy` procgroup package doc claims to be "the one place" weave and sdlc read the kernel's kill answer
  pkg/procgroup/procgroup.go:4 overstates it. weave staging/stage.go:85 and sdlc repolock.go:176 also read kill(pid,0) answers, for single-process liveness checks where EPERM means alive. Narrow the claim to process-group signals. No other count or "only" claim in the window is wrong; the atlas text at atlas/workflow/weave.md:353-358 checks out.
- **BR-2** [Minor] `plan-matches-code` Plan item 2 names a killGroup seam and EPERM-equivalent fake kill that were never built
  The finishProducers test uses a real zombie group through procgroup.Kill, not a seam. Add a Revisions entry so the plan describes the delivered design.
- **BR-3** [Minor] `comment-formatting` ZombieGroup doc comment has one line much longer than the rest
  pkg/procgroup/procgrouptest/zombie.go:16-17. It is the only such line in the window.

## Open findings

- **BR-1** [Minor] `doc-claim-accuracy` procgroup package doc claims to be "the one place" weave and sdlc read the kernel's kill answer
- **BR-2** [Minor] `plan-matches-code` Plan item 2 names a killGroup seam and EPERM-equivalent fake kill that were never built
- **BR-3** [Minor] `comment-formatting` ZombieGroup doc comment has one line much longer than the rest
