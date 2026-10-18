---
id: 000446
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# sdlc change-code: help says --name overrides --issue, but the combination refuses

## Problem

`sdlc change-code --help` says `--name` overrides `--issue`, but passing both refuses them as mutually exclusive, and the source confirms the refusal. pair:1 hit this on 2026-10-18 starting pair#665 M2 on a fresh branch: it had to run with `--issue` and rename the branch afterwards.
