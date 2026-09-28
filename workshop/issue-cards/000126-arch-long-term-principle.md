---
id: '000126'
status: done
started: 2026-06-25T10:36:32-07:00
created: 2026-06-25
updated: 2026-06-25
estimate_hours: 1.36
actual_hours: 0.54
---

# ARCH-PURPOSE: serve the issue's actual purpose — don't settle for the easy win; single-source ⇒ consumers derive (ARCH-* registry, checked at-plan + at-review)

## Problem

Agents are weakest at architecture / long-horizon judgment (the `ARCH-*` charter, #75).
#122 exposed a specific failure mode the existing principles (`ARCH-DRY`, `ARCH-PURE`,
Simplicity-First, Root-Cause) don't cover: at the close I **settled for the easy win** —
wired one consumer (sdlc Go) + enforcement and silently deferred the consumers that *were
the purpose* (parley Lua, the help-text prose) as "follow-up", leaving `issue.cue` as
just-documentation those surfaces don't derive from. I did this *despite repeatedly
articulating the risk* — proving "principle in my head" is not a mechanism. The durable
fix (per the operator) is to inject the posture as checkable taste at several stages, not
rely on remembering it at the close.
