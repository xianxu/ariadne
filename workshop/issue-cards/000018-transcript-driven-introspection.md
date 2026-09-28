---
id: '000018'
status: done
created: 2026-04-30
updated: 2026-05-27
actual_hours: 8.0
---

# Transcript-Driven Introspection Skill

## Problem

Claude session transcripts (the JSONL files under `~/.claude/projects/`) are
counterfactual-rich training data — they record what was proposed-and-rejected,
redirected, or quietly accepted by a tasteful judge (the user). Today this
signal is lost: corrections happen in flight, but no system distills them into
reusable skills, rules, or memories. `workshop/lessons.md` was meant to fill
this role but atrophies in practice because it asks for *synthesis* mid-session,
which is the expensive step.

This issue proposes an **ariadne base-layer skill** (working name
`/xx-introspect-extract` or similar) that any ariadne-styled repo's user can invoke
to run a post-hoc extraction pass over their accumulated transcripts and
produce activity-typed `introspect-<activity>` skills (introspect-code-review,
introspect-brainstorming, introspect-planning, introspect-debugging, …) plus updates to lessons.md,
AGENTS.md, permissions, and the skill inventory of the repo it's invoked from.

Cadence is user-driven: invoke biweekly, or whenever the user feels the agent
isn't picking up on patterns. Each version of the extracted introspect skills should make
future sessions feel a little more like working with someone who knows the
user's taste, without overfitting to past work.
