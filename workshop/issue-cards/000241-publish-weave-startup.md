---
id: 000241
status: working
created: 2026-09-20
updated: 2026-09-28
estimate_hours: 1.23
github_issue:
started: 2026-09-28T18:52:27-07:00
---

# Publish weave and cut over startup

## Problem

Weave must be available as the standalone entrypoint for ariadne-style repos.
The startup implementation in #239 needs to merge before a public release and
consumer cutover can be verified. Tracking delivery here avoids making #239's
close depend on a release that itself depends on #239 being merged.
