---
id: 000035
status: open
created: 2026-05-26
updated: 2026-05-26
estimate_hours:
---

# sdlc postmortem subcommand — structured per-issue retrospective

## Problem

Stage 7 (Postmortem) in the SDLC arc currently has two surfaces, both narrow:

1. **`xx-introspect`** — mines accumulated transcripts across all recent sessions for cross-cutting agent-behavior patterns; weekly/biweekly cadence; produces auto-loading taste skills. Operates on the agent-evolution axis, not per-issue reflection.
2. **`workshop/lessons.md`** — single global file of rules-to-prevent-recurrence. Operator-curated, ad-hoc.

Neither answers the per-issue postmortem question: *for this specific issue we just shipped, what's the structured retrospective?* Timeline, scope shifts, velocity, surprises, learnings, crystallization candidates — none of these are systematically gathered or surfaced.

Today this happens in operator's head (or doesn't). High-quality issues get rich Log entries; lower-attention issues just ship and move on. Patterns worth crystallizing into targets get noticed only when the operator happens to remember them across sessions. Lessons worth recording in `lessons.md` get missed.

The discipline gap: every issue close should produce evidence — what worked, what didn't, what we learned, what might become a durable commitment.
