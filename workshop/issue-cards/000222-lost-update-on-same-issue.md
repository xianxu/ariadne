---
id: 000222
status: wontfix
created: 2026-09-11
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
tracker:
    version: 1
    abandoned: {}
---

# Same-issue concurrent trunk edits are last-writer-wins

## Problem

Publishing issue files moved to the trunk in ariadne#207. The route it replaced
computed a merge base and REFUSED when the same issue file had changed on both
sides; the object-database route has no such check, so two agents editing one
issue file is now **last-writer-wins, silently**.

The concrete case is a stale double-claim: agent A claims issue N and publishes
`status: working`; agent B, on a branch cut before that, claims the same issue
and publishes — B's body overwrites A's, and neither is told. Issue files are
the fleet's coordination surface precisely so this cannot happen.

**Why #207 did not simply restore the old check.** A naive three-way against the
merge base would refuse ordinary republishing. Trunk commits are built
out-of-tree and never land on the branch, so after any publish the branch's
merge base with the trunk predates its own last publication: the trunk looks
"changed since we forked" on every subsequent sync. The old arm had the same
false-positive pressure and papered over it by dropping byte-identical files
before conflict detection, which only covers the no-op case.

Recorded rather than hidden: `sdlc claim --help` and
`atlas/workflow/issue-sync.md` both now state the last-writer-wins behaviour.
