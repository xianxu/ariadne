---
id: 000279
status: open
deps: [ariadne#277]
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '847bcd74596c491522b5ad71cb73fd6298086f92' # card fields mirrored from issue-cards; edit via sdlc
---

# Expose authoritative workflow observations for agents

## Problem

Agents and humans reconstruct workflow progress through ad hoc issue, Git and filesystem inspection. This makes effect verification and recovery inconsistent and duplicates SDLC’s interpretation of its own state.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Extend existing read-only SDLC inspection commands with a coherent machine-readable contract. Expose claim/assignment, current slot/worktree/branch association, workflow checkpoints, working-tree/recent-commit activity, milestone gate verdict/evidence and publication/merge outcomes. Derive from their existing authorities; do not introduce a second workflow state store.

For each observation expose its source/revision and freshness or collection time, and distinguish absent, stale and failed/unknown reads. Separate authoritative milestones from weak activity signals: file changes are not proof of correct progress. A working card proves claim state, not execution or completion. Support direct local invocation against another worktree, including a stopped/parked slot, without asking that slot’s agent to respond. Do not add Couch RPC proxies for hypothetical multi-machine support.

## Done when

- An agent can answer who owns an issue, where it is assigned, what checkpoint passed and whether it landed through documented SDLC queries.
- Structured outputs identify evidence/source and distinguish unknown/unreadable/stale from absent; queries do not mutate state.
- Tests query another local worktree, including parked/no-agent and conflicting/missing worktree cases.
- Activity versus authoritative progress and tracker freshness semantics are documented and covered by contract tests.

## Plan

Implementation plan to be designed after issue claim and start-plan; these are requirements, not an approved implementation plan.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.
