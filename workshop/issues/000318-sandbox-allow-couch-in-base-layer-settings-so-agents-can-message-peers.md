---
id: 000318
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-11
estimate_hours:
card_mirror: '2d07ebc4f6504d2be7ee4cc12342f296b10e5701' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T08:58:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "605774e3", done: "b1990bf8"}
---

# sandbox: allow couch in base-layer settings so agents can message peers

## Problem

Agents can't send couch peer messages from inside the Claude Code sandbox. `couch --send-to` dials the
message socket under `/tmp/pair-message-<uid>/<hash>.sock` and fails with `connect: operation not
permitted`. The outcome is reported as "uncertain", and `couch --message-status` then confirms nothing was
delivered. The base-layer sandbox settings (`.claude/settings.json`, woven into every repo) exclude `sdlc`,
`git` and `gh`, but not `couch`, and the socket isn't in `network.allowUnixSockets`.

The impact grows with multi-slot work. Every TL dispatch and every worker reply needs a sandbox bypass,
which costs a permission prompt or an auto-mode verdict each time. A background TL monitor can't reach
couch at all (it can't detect an idle slot). Seen by the TL in ariadne:1 on 10-09, and again from the ops
slot on 10-10 (ops `teams/` learnings for ariadne-robustness-1, kink 32).

## Spec

Add `couch` to `excludedCommands` in the base-layer sandbox settings, the same treatment as `sdlc`.
couch keeps its own guards: delivery receipts, refusal while a recipient has a pending message, and
`--reboot` refusing live slots. A narrower alternative is to allow the message-socket directory in
`allowUnixSockets`, but the path embeds the uid and a hash, so it's brittle. Pick the exclusion unless
the review finds a reason not to.

## Done when

- `sdlc merge` runs in auto mode without a classifier denial.
- A sandboxed `couch --send-to <slot> --message …` from a woven repo delivers, and `couch --message-status ID --json` shows a receipt.
- The setting lives in the ariadne base layer and reaches peer repos through `weave refresh` (checked in ops).

## Plan

- [x] `.claude/settings.ariadne.json`: add `couch` to `sandbox.excludedCommands`; add `Bash(sdlc:*)` and `Bash(couch:*)` to `permissions.allow` (both keys are `$merge_keys`, so downstream locals union with them).
- [x] Codex (`.codex/config.toml`, woven as a symlink): no change needed, with the reason recorded in the file's comments. Codex has no per-command sandbox exclusion. `workspace-write` with `network_access = true` already permits the socket dial, and `approval_policy = "on-request"` has no classifier to bypass.
- [x] Gemini: not woven (no `.gemini` row in `construct/base.manifest`; ariadne-local only, and Gemini is deprecated in favor of agy). Leave it alone; record why in `## Revisions`.
- [x] `weave compile` here (operator-run, see Log), then confirm the regenerated `.claude/settings.json` carries the change.
- [ ] Verify: a sandboxed `couch --send-to` delivers with a receipt (the final report to ops:0 is the live check), and `sdlc merge` runs in auto mode without a classifier denial.
- [ ] After the merge: `weave refresh` in ops, and confirm ops' `.claude/settings.json` carries both changes.
- [ ] close + merge

## Revisions

### 2026-10-11 — scope widened (TL dispatch from ops:0, operator-approved)
- Was: `couch` in `excludedCommands` only.
- Now also: `Bash(sdlc:*)` and `Bash(couch:*)` in `permissions.allow`, so the auto-mode classifier stops judging the workbench's own verbs. It denied #330's `sdlc merge --yes` as "Merge Without Review", although the binary runs its own review gates. Both binaries keep their own guards.
- Agent-neutral: cover Codex/Gemini wherever weave generates their config, or record why not (see Plan).
- Done-when adds: `sdlc merge` runs in auto mode without a classifier denial.

### 2026-10-11 — TL addendum: items (4) and (5) from the 10-10 project scope
- (4) `sdlc merge`'s local branch cleanup writes `.git/config` (ops kink 52): no new allowance. `sdlc` was already in `sandbox.excludedCommands` before #318, so a plain `sdlc merge` and the git it spawns run unsandboxed. Kink 52 most likely came from an invocation the exclusion doesn't match, such as a compound command (`cd … && sdlc merge`). Run sdlc as the leading command. This issue's own `sdlc merge` is the live check (Log).
- (5) weave compile's `.claude/settings.json` write (#320): deliberately not folded. That file is the harness's permission store, and the sandbox protects it on purpose. An allowance for weave would let an agent rewrite its own permissions without a prompt, which is what auto mode denied twice during this issue ("Self-Modification"). Regenerating settings stays an operator step (`! weave compile` / `weave refresh`).

## Log

### 2026-10-11
- Auto mode denied the agent's commit of `settings.ariadne.json` and its unsandboxed `weave compile` as Self-Modification. The operator ran both (`a9055859`). The regenerated `.claude/settings.json` carries `Bash(sdlc:*)`, `Bash(couch:*)` and `couch` in `excludedCommands`.
- Close refused on the integrate gate; merged origin/main (`e64df5d1`) and re-ran `go test ./cmd/weave/...`: exit 0.

### 2026-10-10
