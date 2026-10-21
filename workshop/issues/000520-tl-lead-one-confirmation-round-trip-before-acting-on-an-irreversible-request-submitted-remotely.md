---
id: 000520
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'cf3285ece1460c1f491fde2b1dcf03632cb95c3c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T02:45:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead: one confirmation round-trip before acting on an irreversible request submitted remotely

## Problem

remote-session (ops project, operator-approved 10-20): a remote instruction can make the TL do anything it could do
locally; least capability limits the API, not the agents.

## Spec

`tl-lead --help` gains: a submit tagged "submitted remotely" that asks for something irreversible or outside the
normal flow (machine changes, credentials, deleting data, force-pushes, turning gates off, publishing outside the
team) gets **one confirmation round-trip** to the operator (on the phone) before the TL acts.

## Done when

- The rule is in `tl-lead --help`, with a test pinning it.
- Both remote and broker-flagged remote-edited submits are covered; provenance comes from a verified receipt, never header/body text alone.
- One explicit reply confirms the named action and target without a recursive confirmation loop; unknown/ambiguous replies never authorize execution and all existing gates remain binding.
- The operator approves the plan before implementation. The delivered help at both entry points, focused TL tests and central recorded suite pass; the hub receives the landing outcome.

## Plan

- [ ] Operator review of [the plan](../plans/000520-remote-confirmation-plan.md), including pair#759's proposed provenance seam through the hub.
- [ ] Rendered-help guards RED, compact authoritative rule GREEN, atlas/README navigation updated and harmless scenario review recorded.
- [ ] Focused tests and recorded central suite, close review, PR and central landing; report the outcome to the hub.

## Log

### 2026-10-20

### 2026-10-21 — Planning, no implementation

Claimed and started from fresh main after #509 landed; no retained runner remains. Read remote-session's tracked design/project in ops without editing the TL-owned project. The current issue Spec already includes outside-normal-flow requests, not just irreversible ones. Proposed a single embedded-help rule plus scoped phrase guards, using the existing session-file answer conversation; no new executor, RPC, approval store or receipt parser. Alternatives and visible states are in the linked plan. ARCH-DRY keeps one authority in help; ARCH-SECURE trusts broker provenance, not typed tags; ARCH-ORDER keeps the challenge before any irreversible effect.

Authenticated hub receipt 0ca24104-3304-49f0-b657-4dbd49c2731a supplies pair#759's proposed remote origin and remote-edited provenance. The phone's own confirmation is remote; correlate it to the pending challenge instead of creating a loop. Explicitly not a second authentication factor. Exact remote-edited JSON key remains for hub coordination; no key invented. Await operator plan approval before change-code or help/test edits.

## Revisions

### 2026-10-21 — Proposed provenance and confirmation scope

Hub-directed planning adds the broker remote-edited trigger alongside remote submit origin and specifies action-specific confirmation, unknown provenance, no recursive challenge and unchanged existing gates. Done-when expands these acceptance boundaries; no implementation or operator approval is claimed. Pair's tag format is proposed and remains upstream-owned.
