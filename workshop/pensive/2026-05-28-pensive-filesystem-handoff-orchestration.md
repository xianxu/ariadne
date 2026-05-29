---
type: pensive
date: 2026-05-28
topic: Filesystem handoffs make multi-agent work human-inspectable
mode: thoughts
description: The artifact is the interface — handing agent work off through git/files (commit-per-turn) keeps a human able to inspect and intervene mid-loop, and is the only channel that works cross-stack. A core ariadne workflow primitive.
references: [AGENTS.md]
---

# Pensive: Filesystem handoffs make multi-agent work human-inspectable

This came out of running a multi-round cross-stack review loop — Codex reviewing Claude's work on the `you-decide` substrate, me checking both in between. The most valuable property wasn't the review quality. It was that **every handoff went through the filesystem and git**, so I could inspect, correct, or revert at every step. The artifact — a review report, edited files, a signed commit — *is* the interface between agents, not an opaque value passed back in memory. ariadne is about how to best use AI and the workflow around it, and this feels like a load-bearing workflow primitive worth naming.

Why it matters:

- **Human-in-the-loop by construction.** A subagent that does work and returns a value hides its reasoning the moment it returns — I can only inspect the end product. A subagent (or peer stack) that writes a file and commits leaves a diff I can read *between* turns. The commit boundary is the inspection point. The ceremony I'd normally see as overhead is exactly what creates the audit trail.
- **Cross-stack only works this way.** Codex and Claude don't share a context window — they share git. The filesystem is the one channel both speak. So cross-stack collaboration *forces* the inspectable pattern; the constraint is a feature.
- **Resumable / durable.** A killed or paused loop leaves a clean checkpoint on disk; the next actor — human or either stack — resumes from committed state, not a volatile buffer.

The trade-off is ceremony: writing artifacts and committing each turn costs more than returning a value. So the heuristic is — filesystem-handoff when a human may want to stay in the loop, or when the output is itself an artifact (reviews, edits, plans, migrations); ephemeral return-value subagents for quick fan-out lookups where no mid-loop inspection is wanted. The commit-per-handoff review loop in `you-decide/review.md` is the worked instance that surfaced this.

## Open questions

- Is there tooling worth building to make the handoff first-class — a thin "write artifact + commit signed + hand to next actor" wrapper, or a `handoff` datatype — or is plain git + the commit conventions (§12) already enough?
- Where does the ceremony stop paying? A 3-turn loop is clearly worth it; a 30-step pipeline of micro-commits could drown the signal. The line probably tracks "would a human plausibly want to look here?"
- This wants to become a **target** — a grounding commitment like *"agent handoffs go through inspectable artifacts, not opaque returns, whenever a human may want to intervene."* Promote once the pattern's been exercised across more task types than review (implementation, migration, research), not just this one.

## References

- `you-decide/review.md` (peer repo) — the commit-per-handoff review loop + "Spawning the other stack" cross-stack mechanics; the instance this realization came from.
- `AGENTS.md` §3 (subagent strategy) and §12 (commit conventions / signing) — the existing primitives this builds on.
