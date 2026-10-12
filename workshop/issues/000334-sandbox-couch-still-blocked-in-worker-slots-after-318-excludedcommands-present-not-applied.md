---
id: 000334
status: open
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-11
estimate_hours:
card_mirror: 'a398abaf830c46ada3481958b348c426780d9d7a' # card fields mirrored from issue-cards; edit via sdlc
---

# sandbox: couch still blocked in worker slots after #318 (excludedCommands present, not applied)

## Problem

ariadne#318 (PR #195) put `couch` (with `sdlc`, `git`, `gh` and others) into the compiled `.claude/settings.json` `sandbox.excludedCommands`, so worker agents could message peers. It isn't taking effect. Evidence, 2026-10-11: pair:4 (Claude Code 2.1.296, bypass-permissions mode), working pair#453 after `couch --prepare` (weave refresh), ran `couch --send-to ops:0` inside the sandbox: `dial unix /tmp/pair-message-501/...sock: operation not permitted`, never delivered; the resend with the sandbox disabled worked. The slot's `.claude/settings.json` (mtime 20:08; the session process started 20:00) lists `couch` in `excludedCommands`. #318's own finding in ariadne:1 (auto mode) was the same: a bare `sdlc` and `couch --send-to` both ran sandboxed. Every report from a worker currently needs an unsandboxed retry, which the auto-mode classifier may refuse.

## Spec

Find why `excludedCommands` doesn't apply, then fix it at the root. Hypotheses to test, cheapest first:
1. **Settings read at process start:** the file changed after the session started (weave refresh during `couch --prepare` rewrites it; `--reload-context` keeps the process). Test: a cold-started session with the file already in place.
2. **Matching:** how Claude Code matches `excludedCommands` (first word? full command string? glob?). In this environment `couch` is a shell function in the user's profile that rebuilds and then runs `~/workspace/pair/bin/couch`; agents also call the full path. Test bare `couch`, the full path, and a compound command.
3. **Mode:** whether bypass-permissions or auto mode changes how exclusions apply.

Fix whichever it is in weave's compiled settings (patterns that match how agents actually invoke the commands) and/or in `couch --prepare` (restart the agent process when settings changed, if they are read only at start). Add a check that would have caught it: from a freshly prepared slot, a sandboxed `couch --send-to` to a test peer succeeds.

## Done when

- The cause is identified with a reproduction, recorded in the Log.
- From a freshly `couch --prepare`d worker slot, `couch --send-to` and `sdlc` run inside the sandbox succeed with no unsandboxed retry, in both bypass-permissions and auto mode.
- A check covers it (test or documented live check).

## Plan

- [ ]

## Log

### 2026-10-11
