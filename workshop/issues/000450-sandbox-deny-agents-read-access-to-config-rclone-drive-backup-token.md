---
id: 000450
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'dc99fc81f1a1eeb249915fb511f9a77354ea05bd' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T15:38:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: operator, spec: "1f6d4f79", done: "c54c9586"}
---

# sandbox: deny agents read access to ~/.config/rclone (Drive backup token)

## Problem

pair's history backup will upload to Google Drive through rclone (operator decision 2026-10-18). rclone keeps its OAuth token in ~/.config/rclone/rclone.conf, which the agents' Claude Code sandbox can read today. The sandbox read-deny list in .claude/settings.ariadne.json covers ~/.aws, ~/.config/gcloud and others, but not ~/.config/rclone.

## Spec

Add "~/.config/rclone" to sandbox.filesystem.denyRead in .claude/settings.ariadne.json, right after "~/.config/gcloud". One line; nothing else changes. This is a local-gates path: the operator verifies the diff and applies the local-gates-verified label; agents never apply it.

## Done when

- The PR diff is exactly that one line, and the operator has labeled it local-gates-verified.
- After weave refresh, ~/.claude/settings.json denyRead includes ~/.config/rclone.

## Plan

- [ ]

## Log

### 2026-10-18
