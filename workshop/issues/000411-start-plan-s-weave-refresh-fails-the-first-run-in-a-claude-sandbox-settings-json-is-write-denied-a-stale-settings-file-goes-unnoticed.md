---
id: 000411
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: 'dd1b481dac7bea2207d42fa101201efdcc8aff50' # card fields mirrored from issue-cards; edit via sdlc
---

# start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed

## Problem

In a Claude slot (parley.nvim:4, 2026-10-16), every first `sdlc start-plan` fails: its weave refresh tries to write
`.claude/settings.json`, which Claude Code's sandbox write-denies (`operation not permitted`; settings files are in the
sandbox's deny-within-allow list by design, so an agent can't rewrite its own permissions). Re-running start-plan
then goes through, so the slot's `.claude/settings.json` may lag the woven one without anyone noticing.

## Spec

- Find why the first run fails and the second passes (the refresh's change detection, or a partial write).
- A gate must not fail on a write the agent's sandbox forbids by design. Either skip the protected settings file when
  unwritable and report "settings.json is stale: run `weave refresh` from your shell" as a one-line notice (one state,
  one action), or move the settings write out of the gates entirely into an operator-run step.
- Never retry the write unsandboxed.

## Done when

- A test with an unwritable `.claude/settings.json` shows start-plan succeeding on the first run with the stale-settings
  notice, and the notice clearing once the file is refreshed.

## Plan

- [ ]

## Log

### 2026-10-16
