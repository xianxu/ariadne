---
id: 000504
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c22406b030aac8bd74f9f14de57fc1bd845d26fe' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule

## Problem

The TL session file lives at `teams/<team>/tl-status-<operator>.md` (side docs in `tl-status-<operator>.side/`). Since
introduce-team settled on one team per ops repo, the team folder adds nothing (the repo is the team), and the
operator's name adds nothing either (the file is gitignored, so it's per machine and per operator already).
Operator, 10-20 (ops side doc `2026-10-20-03-session-file-everywhere.md`): move it to the repo root.

Also stated there, and worth writing into tl-lead: the session file and its side docs are the **ephemeral TL/human
interface**; durable decisions that workers can act on live in the issue system (tracked, managed by sdlc), so a
decision made in a side doc gets copied into the tracked artifact (issue, project Log, charter) when it's made.

## Spec

- tl-lead's default hub session file becomes **`session.md` at the hub repo's root** (gitignored), side docs in
  **`session.side/`**; `tl-lead --help` and every `tl-lead status …` default follow. The `operator:` frontmatter
  stays (defaulting from git's user.name) for tooling that wants it.
- **`tl-lead status move [OLD]`**: one-shot migration that renames the file and its side docs, rewrites the side docs'
  parent links and the status file's links into them, updates Couch's registration (`couch --session-sync rename`),
  and updates `.gitignore`. Refuses under the operator's lock; dry run by default.
- tl-lead rule: decisions get copied from side docs into tracked artifacts when made (TL vs worker: workers take
  instructions from the issue system, with occasional draft-pane or TL input; the TL multiplexes and keeps the
  temporal picture in the session file).
- Check parley's attention finder still lists side docs under the new parent (adjust in parley.nvim if not).

## Done when

- Tests: `tl-lead status move` on a fixture renames file and side docs, rewrites links, and refuses under a lock.
- `tl-lead --help` documents the new location and the decisions rule.
- **Last step, manual:** migrate the two live hubs, ops:0 (`teams/devinfra/tl-status-xian-xu.md`) and ops-echo:0
  (`teams/echo/tl-status-xian-xu.md`), with the operator at a quiet moment.

## Plan

- [ ]

## Log

### 2026-10-20
