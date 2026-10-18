---
id: 000446
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ea580e5ffc03ca6e426b0a948f56d8620643f3b6' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc change-code: help says --name overrides --issue, but the combination refuses

## Problem

`sdlc change-code --help` says `--name` overrides `--issue`, but passing both refuses them as mutually exclusive, and the source confirms the refusal. pair:1 hit this on 2026-10-18 starting pair#665 M2 on a fresh branch: it had to run with `--issue` and rename the branch afterwards.

## Spec

Make help and behavior agree. Either `--name` really overrides the branch name while `--issue` still identifies the issue (the useful case: a milestone continued on a fresh branch from landed main), or the help states they are exclusive and names the supported way to continue an issue on a new branch. Prefer the first if it needs no new state.

## Done when

- Help text and behavior agree, with a test of the combination.
- The milestone-on-a-fresh-branch case has one documented command.

## Plan

- [ ]

## Log

### 2026-10-18
