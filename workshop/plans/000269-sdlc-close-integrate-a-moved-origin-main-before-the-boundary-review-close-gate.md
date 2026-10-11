---
gate: boundary-review
issue: 269
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T01:06:28-07:00"
      dispatched: "2026-10-11T00:59:31-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Legacy-repo integrate fetch runs without a context or timeout
          detail: integrateMainRef's legacy path calls gitx.RunGit("fetch", ...) through exec.Command with no deadline and ignores the ctx it receives, so an unresponsive remote hangs close indefinitely (ARCH-CONSTRAINTS). Pass ctx to the fetch or give it a deadline.
          family: unbounded-external-call
          round: 1
        - id: BR-2
          severity: Minor
          title: integrateBranchPushed treats any @{upstream} as pushed
          detail: An unpushed issue branch created with -b x origin/main tracks main as its upstream, so the gate names git merge instead of git rebase. Ignore an upstream equal to main's ref.
          family: push-state-heuristic
          round: 1
      recipe: milestone-review
      reviewed: 92b6bf38fe85138a0b0a28547e3e607092d4c52e
      blocked: false
---

# Gate ledger — ariadne#269 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T01:06:28-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `unbounded-external-call` Legacy-repo integrate fetch runs without a context or timeout
  integrateMainRef's legacy path calls gitx.RunGit("fetch", ...) through exec.Command with no deadline and ignores the ctx it receives, so an unresponsive remote hangs close indefinitely (ARCH-CONSTRAINTS). Pass ctx to the fetch or give it a deadline.
- **BR-2** [Minor] `push-state-heuristic` integrateBranchPushed treats any @{upstream} as pushed
  An unpushed issue branch created with -b x origin/main tracks main as its upstream, so the gate names git merge instead of git rebase. Ignore an upstream equal to main's ref.

## Open findings

- **BR-1** [Minor] `unbounded-external-call` Legacy-repo integrate fetch runs without a context or timeout
- **BR-2** [Minor] `push-state-heuristic` integrateBranchPushed treats any @{upstream} as pushed
