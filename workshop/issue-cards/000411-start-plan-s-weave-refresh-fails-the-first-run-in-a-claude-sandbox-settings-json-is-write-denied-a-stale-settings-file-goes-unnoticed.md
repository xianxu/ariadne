---
id: 000411
status: open
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
---

# start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed

## Problem

In a Claude slot (parley.nvim:4, 2026-10-16), every first `sdlc start-plan` fails: its weave refresh tries to write
`.claude/settings.json`, which Claude Code's sandbox write-denies (`operation not permitted`; settings files are in the
sandbox's deny-within-allow list by design, so an agent can't rewrite its own permissions). Re-running start-plan
then goes through, so the slot's `.claude/settings.json` may lag the woven one without anyone noticing.
