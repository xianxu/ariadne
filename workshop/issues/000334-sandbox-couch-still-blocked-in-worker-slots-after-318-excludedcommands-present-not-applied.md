---
id: 000334
status: codecomplete
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-12
estimate_hours:
card_mirror: '056203e9d3ebec380ea349a9f2d710f3b3ad0ea8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T23:35:37-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.42
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

## Revisions

- 2026-10-12, after the first close: scope widened again. `gh` is exempt with arguments (`gh *`, 280f2768, operator-approved), in the base and in the user's `~/.claude/settings.json`. `git`, `docker`, `npm` and `npx` stay exact matches. This supersedes the Log's first "left bare" note. Close review round 3 asked for a full close.
- 2026-10-11, close review round 1: the Done-when line "in both bypass-permissions and auto mode" was met live for **auto mode only**. Bypass-permissions is inferred from the docs: matching is independent of mode, and mode only changes the approval flow after a match. It wasn't observed. Scope widened by operator approval (964c47a1) from sdlc/couch to all ariadne (`cmd/*`) and pair binaries.

## Plan

- [x] Reproduce in this (freshly prepared, auto-mode) session
- [x] Find the matching rule (docs + the 2.1.296 binary)
- [x] Guard test `TestBaseSettingsExcludedVerbsMatchArguments`
- [x] `.claude/settings.ariadne.json`: ariadne + pair binaries as `X *` (TL, 964c47a1, operator-approved)
- [x] Live check: `weave compile` in ariadne:1, then sandboxed bare calls plus negative controls
- [x] Atlas: agent-facing rule in `atlas/workflow/sandbox.md`

## Log


- 2026-10-12: closed — Root cause: excludedCommands is exact-match Bash rule syntax; bare couch exempted only a no-arg call. Fixed as X-star entries for ariadne cmd binaries, pair binaries and gh (operator-approved, Revisions); git/docker/npm/npx stay exact. Live in ariadne:1 (auto mode) after weave compile, same session, no restart: bare couch --broadcast-list --json and sdlc state ran sandbox-free, and a later couch --send-to delivered sandboxed; cd-elsewhere, full-path, command-substitution and pipe forms stayed sandboxed. Bypass mode inferred from docs. TestBaseSettingsExcludedVerbsMatchArguments passes at tip (red pre-fix).; review verdict: SHIP
### 2026-10-11
- 2026-10-11: closed — Root cause: excludedCommands is exact-match Bash(...) syntax; bare 'couch' exempts only a no-arg call. Fixed as 'X *' for ariadne cmd/* + pair binaries (964c47a1, operator-approved). Live in ariadne:1 (auto mode) after weave compile, same session, no restart: sandboxed bare 'couch --broadcast-list --json' -> {"broadcast":null}, 'sdlc state' ok; controls (cd elsewhere &&, full path, $(...), pipe) stay sandboxed. Bypass-permissions mode inferred from docs, not observed (see Revisions). TestBaseSettingsExcludedVerbsMatchArguments passes (red pre-fix); go test ./cmd/weave/... green. Round-1 review findings all fixed.; review verdict: SHIP
- **Reproduced** in ariadne:1 (Claude Code 2.1.296, auto mode, freshly prepared): a bare, leading `couch --broadcast-list --json` → `dial unix /tmp/pair-message-501/….sock: connect: operation not permitted`. It wasn't compound and didn't use a full path.
- **Cause:** `sandbox.excludedCommands` entries use `Bash(...)` rule syntax; a pattern without a wildcard is an EXACT match. The docs (code.claude.com/docs/en/sandboxing, excludedCommands section) say "`docker` matches only `docker` with no arguments. `docker *` matches `docker` with or without arguments", so `"couch"` exempts only a bare `couch`. It wasn't hypothesis 1 (settings read at start) or 3 (permission mode). Repository entries would be dropped only under an admin-required sandbox (a managed or `--settings` `allowUnsandboxedCommands:false`), which isn't the case here. A `git -c alias…` probe confirmed `git` with arguments also runs sandboxed; `git`/`gh`/`docker`/`npm`/`npx` were left bare at first, as a separate decision; `gh` was later widened (see Revisions).
- The auto-mode classifier refused my edit to the settings file, and the commit of this log, as [Self-Modification]. I stopped and reported. The operator approved exempting all ariadne and pair binaries; the TL committed 964c47a1 (settings + test + atlas note).
- **Verified live** in ariadne:1: `weave compile` (the sandbox denies writing `.claude/settings.json`; it ran unsandboxed through the permission gate) produced `sdlc *`, `couch *`, … in the compiled settings. In the SAME running session, with no restart, sandboxed bare `couch --broadcast-list --json` → `{"broadcast":null}` and `sdlc state` worked. So excludedCommands edits apply live. Negative controls stayed sandboxed and failed, as the docs say: `cd <other dir> && couch …`, `/…/pair/bin/couch …`, and `X=$(couch …)`. (`cd <cwd> && couch …` passed; a `cd` into the current directory looks like a no-op.) Bypass-permissions mode wasn't tested live (this session is auto); the docs make matching independent of mode, which only changes the approval flow after a match.
- **Check:** `go test ./cmd/weave/internal/settingsx/ -run TestBaseSettings` passes; it was red on the pre-fix base.
- Close review round 1 (FIX-THEN-SHIP): fixed BR-1 (`atlas/workflow/sdlc-binary.md` said slot agents run sdlc sandboxed, so the review fence was always refused; now that's true only for sandboxed call forms), and the guard test now derives ariadne's binaries from `cmd/*` plus an explicit pair list. Atlas no longer overclaims: redirects are from the docs, while pipes were seen live (`sdlc close … | tail` ran sandboxed). Also updated the `$comment` and the atlas `permissions.allow` line, with an ARCH-SECURE note that `weave *` unsandboxed can rewrite `.claude/settings.json`, gated by the permission flow. Bypass mode is recorded in Revisions.
- Close review round 3 (REWORK): BR-7, `gh *` was a scope change after close; now recorded in Revisions, with a full re-close. BR-8, stale `gh` mentions; swept the atlas, the `$comment` and this Log in one commit. That follows the review's rule for this family: a commit that changes a setting greps for and updates every description of it in the same commit.
