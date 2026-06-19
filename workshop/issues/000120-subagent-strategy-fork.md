---
id: 000120
status: open
deps: []
github_issue:
created: 2026-06-19
updated: 2026-06-19
estimate_hours:
---

# Subagent strategy: add the fork-for-implementation pattern (context inheritance)

## Problem

The subagent strategy — AGENTS.md §3 and the adapted superpowers skills
(`construct/adapted/superpowers-dispatching-parallel-agents/SKILL.md`,
`construct/adapted/superpowers-subagent-driven-development/SKILL.md`) — frames
subagent use around one axis: *"is the context I need capturable as a prompt?
→ subagent."* That implicitly assumes a **fresh** subagent (starts blank, sees
only the prompt). It never mentions the **fork** option, where the subagent
inherits the parent's *full conversation transcript* and runs on the parent
model.

That omission misses a high-value pattern. Forking substantial implementation
work lets it run with the entire design arc in context, in its **own** context
budget, and return only a digest — so the main thread keeps the design +
conclusion without accumulating the implementation exhaust (every edit, test
run, debug cycle). The user's framing: *"a stack of context"* — push full
context in, pop a digest out; the exhaust stays on the fork's frame and is
discarded on return. Validated this session (pair #66): M2 was forked and spent
~680k tokens on the fork while adding near-zero to the main thread.

## Spec

Update the subagent strategy to make **context inheritance** an explicit axis,
naming three modes with their when-to-use:

- **fork** (inherits full transcript, parent model) — substantial,
  fully-specified implementation that rides on the accumulated thread. The
  default for milestone-sized implementation. The prompt to a fork is *marching
  orders* (restate THIS task's load-bearing constraints), not context —
  inheritance gives knowledge but not focus.
- **fresh** (sees only the prompt) — bounded reads and **reviews**. The
  fresh-eyes / no-inherited-confirmation-bias property *is* the point; already
  mandated for plan reviewers (`superpowers-writing-plans` says "never your
  session history"). Keep that mandate and generalize the rationale.
- **inline** (main thread) — small, warm-context, or **you-in-the-loop** work
  (a fork runs autonomously to completion — it cannot stop and ask mid-stream).

Also capture:
- **Cost trade-off:** a fork re-processes the whole transcript — worth it when
  the work dwarfs the inheritance, overkill for a one-liner.
- **Verify the return:** a fork's report is a *claim*; the parent re-runs the
  tests / checks the artifacts (trust-but-verify), same as any subagent.

Touch points (confirm in implementation):
- AGENTS.md §3 (Subagent Strategy) — the constitution section; add the
  fork/fresh/inline trichotomy alongside the existing "capturable as a prompt"
  test (which becomes the *fresh*-vs-inline test; fork bypasses it).
- `construct/adapted/superpowers-dispatching-parallel-agents/SKILL.md` and
  `construct/adapted/superpowers-subagent-driven-development/SKILL.md` — note
  fork as the inheritance-carrying alternative; clarify that parallel-fan-out
  and review uses stay *fresh*.
- Keep this DRY: prefer one canonical statement (likely §3) that the skills
  reference, rather than restating the trichotomy in each.

Source: `brain/.../memory/feedback_fork_for_implementation.md` (the captured
preference) — port the durable guidance, drop the session-specific detail.

## Done when

- The subagent-strategy guidance names fork / fresh / inline with the
  when-to-use for each, and states fork-for-substantial-implementation as the
  default.
- The "capturable as a prompt?" test is reframed as the *fresh*-vs-inline
  decision (fork inherits, so the test doesn't gate it).
- Review/fan-out uses are explicitly kept *fresh* (fresh-eyes preserved).
- The cost caveat and the verify-the-return rule are present.
- No duplicated trichotomy across the skills — one canonical home, referenced.

## Plan

- [ ]

## Log

### 2026-06-19

- Filed from a pair #66 session: M2 was implemented via a fork (full-context
  inheritance), which surfaced that the subagent strategy only documents the
  fresh-subagent path. Captured the preference in brain memory
  (`feedback_fork_for_implementation.md`); this issue ports it into the shared
  guidance.
