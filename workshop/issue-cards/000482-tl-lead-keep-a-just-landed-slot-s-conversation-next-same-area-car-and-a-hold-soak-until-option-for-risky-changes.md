---
id: 000482
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# tl-lead: keep a just-landed slot's conversation (next same-area car), and a 'hold: soak until' option for risky changes

## Problem

When a landed change later breaks something, the slot whose conversation wrote it has the freshest context, but
tl-lead has no rule about keeping it: slots get prepared with a fresh conversation, rebooted, or dispatched into
unrelated code right after landing. Operator idea, 10-19 (ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-16-landing-queue-status.md`).
