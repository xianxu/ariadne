---
id: '000247'
status: done
started: 2026-09-23T22:39:51-07:00
created: 2026-09-23
updated: 2026-09-23
estimate_hours: 4.24
actual_hours: 1.47
---

# Explicitly refresh repositories and dependencies with Weave

## Problem

Numbered slots own private dependency clones, but those clones have no separate
Couch thread and are easy to leave stale. In :0 the operator can directly manage
peer repositories. Agents need an explicit way to refresh the host and its dependency sources
before planning or implementation relies on them, without coupling that action
to issue reservation.
