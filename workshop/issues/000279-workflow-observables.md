---
id: 000279
status: working
deps: [ariadne#277]
github_issue:
created: 2026-10-01
updated: 2026-10-01
estimate_hours:
card_mirror: '61002c0ce009892b53729176aef6fe04456996a6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-01T23:06:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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

Durable plan: `workshop/plans/000279-workflow-observables-plan.md`.

- [ ] M1 — the contract and its tracker sections: `internal/observe` types,
      strict JSON and golden; `Assemble` for card, assignment, completion and
      landing; `issue show --json`; parked-slot, stale-tracker and no-mutation
      tests.
- [ ] M2 — checkpoints (flow, plan, review verdicts, open blocking),
      workspaces and activity, `--repo`; conflicting, missing and
      other-machine cases; docs.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-01 (implementation session)

- Operator authorized the work ("work on #279"). Claimed in ariadne:1 (the
  claim records the claimant) and ran start-plan. Branch
  `000279-workflow-observables` sits at main, which includes #277, #278 and
  #275. Mapping the existing read-only inspection commands (`state`,
  `issue show/list`, `fleet`, `workspace`, `reclaim` inspect) before
  designing, so this extends them rather than adding a second state store.

- Explorer map of the existing read-only surfaces:
  - `state --json`: workspace, issues, worktrees, recent commits, drift and
    `tracker_stale`. No timestamp in JSON; current repo only.
  - `issue show`/`list`: text only.
  - `fleet inventory --json`: strict JSON with `available` + `error`
    discriminators and per-tree facts (head, dirty_count, ahead/behind), plus
    a branch-prefix issue association.
  - `workspace --json`: `schema_version` 2, the only verb taking an address.
  - `reclaim` inspect: text; its revision is the card blob OID.
  - Nothing exposes the flow record, gate-ledger decisions, verdict
    trailers, the completion binding or landing. No output carries
    observed-at or a source revision.
  - Fleet's JSON conventions (discriminator plus error, strict decode,
    goldens) are the precedent to reuse.
- Operator decisions:
  1. **Surface:** `sdlc issue show N --json`, a versioned observation; the
     text view gains the same sections. `state` and `fleet` are unchanged.
  2. **Freshness:** fetch the tracker, falling back to the last fetch marked
     `stale` with the error. Every answer carries the tracker commit and
     `observed_at`.
  3. **Targeting:** the query is repo-wide from any checkout, finding the
     local worktrees that hold the issue branch through git (including
     parked or no-agent slots). `--repo <path>` queries another repository.
- Wrote the durable plan (M1/M2). Awaiting operator approval.
