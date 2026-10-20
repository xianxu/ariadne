---
id: 000504
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3df0eebfb4c70ec89dff964f3e5b685bc77d31cf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T09:56:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Implement shared omitted-FILE defaults while preserving explicit paths.
- [ ] Implement pure copy-then-switch planning and fixture link rebasing.
- [ ] Implement guarded copying, Couch registration switch, verification and retryable cleanup.
- [ ] Update help/README/atlas with session.md, session.side/ and the decisions rule; verify Parley attention compatibility.
- [ ] Run focused checks, recorded suite and the SDLC boundary review.
- [ ] TL and operator migrate both live hubs at a quiet moment; worker does not run live migrations.

Durable plan: [session file move](../plans/000504-session-file-move-plan.md).

## Log

### 2026-10-20

- Claimed in ariadne:4 and entered planning with `sdlc start-plan`. No implementation edits or live hub mutations.
- TL accepted unambiguous legacy discovery, optional FILE defaults, compatibility and lock/collision/symlink refusals. TL directed copy-then-switch without a recovery journal. Sources remain until copies and registration switch are verified; after switch, cleanup failures use the new session and rerun cleanup. No editor lease exists, so apply requires a quiet window and rechecks observed locks (ARCH-ORDER).
- Reuse `SessionSync.Rename` to preserve registry identity, baselines and owed events; reuse sidedoc link/index transformations (ARCH-DRY). Do not mark unread edits seen during migration.
- Read-only delegated Parley check: root-parent side-doc discovery is path agnostic (`lua/parley/attention.lua`, `attention_model.lua`, `document_move_plan.lua`); `tests/integration/attention_finder_spec.lua` covers root discovery and nested documents on subsequent opens. No Parley change indicated. Tests not yet run.
- Complete plan drafted for TL approval before change-code. Manual migration remains in Done when and will block completion until the TL/operator supplies evidence.

## Revisions

### 2026-10-20 — planning decisions from the TL

- Added the executable plan and explicit defaults/compatibility semantics without changing the existing Spec or Done when. Chose copy-then-switch instead of a recovery journal at the TL's direction. Live migrations remain the last manual acceptance step, owned by the TL and operator.
