---
id: 000228
status: open
created: 2026-09-15
updated: 2026-09-15
estimate_hours:
github_issue:
---

# Align local superpowers with SDLC review gates

## Problem

The local adaptation of the Superpowers brainstorming and writing-plans skills
requires review loops that overlap with Ariadne's SDLC gates. In particular,
brainstorming requires an automated spec-document review, and writing-plans
requires a plan-document review. `sdlc change-code` then performs the mandatory
fresh-context plan-quality review before implementation, with persistent findings
and disposition tracking.

This creates duplicate review authority and makes the intended workflow unclear.
The local skills should preserve useful design discipline without requiring agents
to pass multiple substantially overlapping automated approvals.
