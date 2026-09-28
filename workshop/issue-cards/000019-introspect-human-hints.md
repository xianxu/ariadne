---
id: '000019'
status: done
created: 2026-05-01
updated: 2026-05-01
actual_hours: N/A
---

# Human Hints as Strong Signal in Introspection

## Problem

The introspection pipeline today only consumes `~/.claude/projects/*.jsonl`
transcripts. Patterns reach the deployed `introspect-<activity>` skills only if
they show up *inside* a transcript as a redirect/endorsement, and only after
clearing the ≥2-distinct-segments threshold during the cluster pass.

This means:
1. A user who *knows* a rule should apply has no clean way to inject it.
   They must wait for the rule to manifest organically across two sessions and
   survive the cluster pass — slow, lossy, and easy to miss.
2. The current heuristic blind spots (rules the LLM extractor under-weights,
   subtle preferences that don't surface as redirects) have no escape hatch.

We want a side-channel: **explicit human hints** that the next pipeline run
treats as strong signal, each hint worth its own cluster, bypassing the
≥2-segment threshold. Hints are still subject to user review and retirement —
they're hints, not commandments — but they're authoritative by default.

While we're here, consolidate the scattered `~/.claude/introspect-cache/` and
(planned) `~/.claude/introspect-versions/` paths under a single
`~/.claude/introspect/` root for tidiness.
