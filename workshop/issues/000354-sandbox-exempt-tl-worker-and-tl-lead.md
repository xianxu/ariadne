---
id: 000354
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '0afc3e8690e0135ce048c88e891510c0fcd2e09a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T17:36:58-07:00
---

# add tl-worker * and tl-lead * to .claude/settings.ariadne.json sandbox.excludedCommands (local gates; unblocks #350)

## Problem

ariadne#350 adds two binaries, `cmd/tl-worker` and `cmd/tl-lead`. `cmd/weave/internal/settingsx` `TestBaseSettingsExcludedVerbsMatchArguments` requires every `cmd/` binary to be listed in `.claude/settings.ariadne.json` `sandbox.excludedCommands` as `X *` (#334). That file is local-gates protected, so per AGENTS §5 the change is its own issue with the operator's verification.

## Spec

Add `"tl-worker *"` and `"tl-lead *"` to `sandbox.excludedCommands` in `.claude/settings.ariadne.json`, the same `X *` form as the other cmd binaries; mention both in its `$comment` list if it enumerates them. `tl-lead` needs it for real (it runs `couch --session-sync` and `sdlc` as subprocesses); `tl-worker` only prints but follows the rule. The PR touches a local-gates path, so it needs the operator's `local-gates-verified` label.

## Done when

- The two entries are present; `TestBaseSettingsExcludedVerbsMatchArguments` passes on a tree with `cmd/tl-worker` and `cmd/tl-lead` (check in a scratch copy with the two dirs); the operator labels the PR and it lands before #350's M1 closes.

## Plan

- [ ]

## Log

### 2026-10-13
