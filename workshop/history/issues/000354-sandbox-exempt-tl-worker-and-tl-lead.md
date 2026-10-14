---
id: 000354
status: done
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '5ca729c24c83ab4788cb411d0d3fd8aa7284394b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T17:36:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "27755c00", done: "7c397d52"}
actual_hours: 0.03
---

# add tl-worker * and tl-lead * to .claude/settings.ariadne.json sandbox.excludedCommands (local gates; unblocks #350)

## Problem

ariadne#350 adds two binaries, `cmd/tl-worker` and `cmd/tl-lead`. `cmd/weave/internal/settingsx` `TestBaseSettingsExcludedVerbsMatchArguments` requires every `cmd/` binary to be listed in `.claude/settings.ariadne.json` `sandbox.excludedCommands` as `X *` (#334). That file is local-gates protected, so per AGENTS §5 the change is its own issue with the operator's verification.

## Spec

Add `"tl-worker *"` and `"tl-lead *"` to `sandbox.excludedCommands` in `.claude/settings.ariadne.json`, the same `X *` form as the other cmd binaries; mention both in its `$comment` list if it enumerates them. `tl-lead` needs it for real (it runs `couch --session-sync` and `sdlc` as subprocesses); `tl-worker` only prints but follows the rule. The PR touches a local-gates path, so it needs the operator's `local-gates-verified` label.

## Done when

- The two entries are present; `TestBaseSettingsExcludedVerbsMatchArguments` passes on a tree with `cmd/tl-worker` and `cmd/tl-lead` (check in a scratch copy with the two dirs); the operator labels the PR and it lands before #350's M1 closes.

## Plan

- [x] Add `"tl-worker *"` and `"tl-lead *"` to `sandbox.excludedCommands`. The
      `$comment` names the class ("ariadne's cmd/*"), not each binary, so it
      needs no change.
- [x] Verify `TestBaseSettingsExcludedVerbsMatchArguments` in a scratch copy
      that has `cmd/tl-worker` and `cmd/tl-lead`.

## Log

### 2026-10-13
- 2026-10-13: closed — make test green at 519c0609: 12 shards, 1136 cmd/sdlc tests plus 47 packages, 0 failed. TestBaseSettingsExcludedVerbsMatchArguments passes in a scratch copy (git archive HEAD) with empty cmd/tl-worker and cmd/tl-lead dirs; the control with main's settings fails with lacks "tl-lead *" and lacks "tl-worker *". Settings-only change, so --no-atlas: there is no new architectural surface, and atlas/workflow/sandbox.md already states the X * rule for every cmd binary.; review verdict: SHIP

- Dispatched by the TL (ops:0); claimed in ariadne:4.
- Scratch check (`git archive HEAD` plus this settings file, with empty
  `cmd/tl-worker` and `cmd/tl-lead`): the test passes. Control, the same tree
  with main's settings: it fails with `lacks "tl-lead *"` and `lacks
  "tl-worker *"`. On the real tree (no such dirs yet) the package passes too.
- Local-gates path: the PR waits for the operator's `local-gates-verified`
  label before `sdlc merge`.
