---
id: 000483
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# Continuing an issue whose branch already landed: refuse with 'close it and file a follow-up' (pair#622 case)

## Problem

pair#622 (10-19): the issue's code landed (PR 406), but its live acceptance stayed open and needed a little more
work. `sdlc change-code --issue 622` insists on the canonical branch, which already merged; `--name` and `--issue`
are mutually exclusive; so there is no supported way to continue an issue whose branch has landed. The worker
improvised a fresh branch outside the gates.
