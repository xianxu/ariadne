---
id: 000176
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.65
started: 2026-07-14T17:44:20-07:00
---

# spine guards for off-workflow invocations: change-code on done issues + non-SDLC repos

## Problem

#172's friction audit found the spine gets invoked where the workflow state
says it shouldn't be, with no guard until close-time:

1. **Working on a `done` issue is un-gated.** `change-code` (and every other
   verb) runs silently against a done issue; only re-close is guarded.
   Firing-order measured 11 change-code-after-close inversions (8 brain,
   2 pair, 1 ariadne-dogfood).
2. **brain runs the spine against its own charter.** brain is a Drive-like
   capture repo, not an SDLC repo, yet it concentrates 19 bypasses and 8/11
   firing-order anomalies — every gate becomes noise to route around there,
   and it pollutes cross-repo friction measurement.

## Spec

**Root cause for the brain half (post-filing discussion, 2026-07-14):** the
"don't run sdlc in brain" rule lives only in Claude-side agent memory — codex
can't see it, and brain's own merged AGENTS.md carries the base §0 ("`sdlc`
manages the development life cycle — READ IT NOW"). Agents in brain sessions
are OBEYING brain's constitution. So the fix is two-layer: the binary owns the
gate (#69 pattern), and brain's constitution stops instructing the opposite.

Candidates:
- `change-code` (and `start-plan`) warn or refuse when the issue is
  `status: done`: "reopen via the REWORK path or open a new issue".
- **brain guard:** spine lifecycle verbs (claim, start-plan, change-code,
  milestone-close, close, merge, push) REFUSE (not warn — the audit shows
  refusals convert 149/158) when `.brain/config.md` exists (the constitution's
  own canonical brain test), with a next-action message: captures go through
  datatype flows (project/roadmap/pensive) + plain git; engineering work
  belongs in a peer repo.
- **Precision — guard the lifecycle, not the reads:** sdlc legitimately READS
  brain (estimate-source calibration docs, ledger, project files). The guard
  must not break sdlc-reads-brain, only sdlc-lifecycle-in-brain. brain's
  gcrypt+GPG remote is untouched by all of this.
- **Constitution layer:** brain's `AGENTS.local.md` overrides base §0 so the
  merged AGENTS.md leads with the charter + the positive path (agent-neutral —
  fixes codex too, unlike Claude memory).
- Generic fallback for other non-SDLC repos: refuse when the SDLC layout
  (`workshop/issues/`) is absent.

**Prerequisite: inventory before enforcing.** brain has issue files up to
~#128 and close rationales like "pkg/... build+vet+test green" — real
engineering-shaped work happened inside brain under sdlc. Triage that backlog
first (migrate genuinely-engineering issues/code to a peer repo; archive the
rest as capture), else the guard strands in-flight work and day-one demands a
bypass flag — recreating the exact pattern #172 measured.

## Done when

- Running `sdlc change-code --issue N` on a done issue produces a
  warn/refusal naming the reopen path.
- Spine verbs in a non-SDLC repo (brain) refuse/warn instead of half-working.
- Re-measure: brain's bypass + anomaly concentration drops out of the
  friction report.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.2  impl=0.3
item: milestone-review     design=0.0  impl=0.12
design-buffer: 0.15
total: 0.65
```

Σdesign 0.2 × 1.15 + Σimpl 0.42 × 1.0 = 0.65. One shared guard func + 7
one-line call sites + done-issue check + command-tree drift test + docs +
brain constitution edit; milestone-review = the close-time boundary review.
*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000176-spine-off-workflow-guards-plan.md`.
Single review boundary (no Mx tags).

- [x] inventory brain's workshop/issues backlog — done, recorded in Log: 1 open
  + 9 archived (all June-era, already done); the audit's brain#116/#128
  anomalies were cwd-label noise; nothing to migrate
- [ ] Task 1 — guardSpineRepo (brain marker + no-workshop/issues arms, env escape
  with cwarn ACK) + done-issue guard on start-plan/change-code + WorkflowVerbs
  drift test + gatesig no-collision (TDD)
- [ ] Task 2 — docs (helptext, AGENTS.base.md, atlas) + brain AGENTS.local.md
  charter + weave + live verify + close

## Log

### 2026-07-14 — brain inventory (plan prerequisite)

Read-only sweep of ~/workspace/brain: **1 open issue** (#12 bootstrap-mac
learner env) + **9 archived** (#1–#11: June-era sandbox/gmail/tooling/
extraction work — real engineering, but already done+archived → nothing to
migrate). No nested repos. The audit's "brain #116/#128" firing-order
anomalies turn out to be **cwd-label noise**: sessions that started in a brain
cwd then `cd`'d to pair/parley — the transcript stays brain-labeled while the
commands ran elsewhere (the #172 drill even shows `[claude/brain/merge] cd
…/pair sdlc merge`). So brain's true lifecycle usage is the #1–#11 era; the
guard + constitution fix prevent recurrence, and brain#12 stays as an inert
file (guards block verbs, not files).

### 2026-07-14

Filed from #172 M4 (T3 findings 3–4).
