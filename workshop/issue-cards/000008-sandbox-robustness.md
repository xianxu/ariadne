---
id: '000008'
status: done
created: 2026-04-21
updated: 2026-04-21
actual_hours: N/A
---

# 000008 — Sandbox Robustness & Troubleshooting

## Problem

`make sandbox` fails with cascading errors when prerequisites aren't met. The script doesn't pre-check requirements, so failures appear mid-flow with opaque messages.
