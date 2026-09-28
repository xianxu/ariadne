---
id: '000122'
status: done
started: 2026-06-24T13:13:10-07:00
created: 2026-06-24
updated: 2026-06-25
estimate_hours: 10
actual_hours: 4.67
---

# Formal schema layer: nouns + lifecycle as a contract-bearing model, compiled to consumers

## Problem

The system's **nouns** (data schema) and **verbs** (lifecycle / state machine) are
defined implicitly and duplicated across repos and languages. The `issue` noun is
the worst case and the oldest: its model lives as prose in the base `AGENTS.md`
status enumeration, as *scattered string literals* in sdlc Go (`isTerminalStatus`
hardcodes `done|wontfix|punt`; `claim.go` compares `prev != "open"`; `startplan.go`
compares `== "working"` — there is **no `Issue` struct and no central status enum**),
and as a status cycle in parley.nvim's Lua. The state machine is real but smeared
across those comparisons; nothing is the source, so drift is structural.

More broadly: "architecture in the agentic age" wants the nouns + verbs formalized
**once**, as an authoritative source compiled to every consumer (LLM prose, the
deterministic shell, application code, tests), so the LLM generates *from* one source
of truth instead of re-deriving the model and duplicating it.

Design captured in
`workshop/pensive/2026-06-24-01-pensive-cue-schema-layer-nouns-verbs.md`.
