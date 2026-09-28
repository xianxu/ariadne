---
id: '000119'
status: wontfix
started: 2026-06-18T23:32:37-07:00
created: 2026-06-18
updated: 2026-09-27
---

# Multi-agent benchmark harness

## Problem

Ariadne has been developed and tested almost entirely against Claude. To claim
the ecosystem is genuinely *multi-agent ready*, we need a controlled way to
measure how well a different coding agent (Codex, Antigravity, Gemini, …) does
**real ariadne work** — not a synthetic coding benchmark, but the actual tasks
in our backlog, judged on both the quality of the result *and* how cleanly the
agent rides the ariadne workflow (claims the issue, plans, creates the right
artifacts, reasons about design subtleties). There is no such harness today.
