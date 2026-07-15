---
id: 000176
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours:
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

## Plan

- [ ] inventory brain's workshop/issues backlog (what is engineering vs capture; migration targets)
- [ ] design the two guards (done-issue warn on change-code/start-plan; brain/non-SDLC-repo refusal) as one gate family — lifecycle-only, reads exempt
- [ ] brain AGENTS.local.md override (charter + positive path leads the merged constitution)

## Log

### 2026-07-14

Filed from #172 M4 (T3 findings 3–4).
