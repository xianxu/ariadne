---
id: 000096
status: open
deps: [ariadne#95]
github_issue:
created: 2026-06-14
updated: 2026-06-14
estimate_hours:
---

# weave: prune stale lowered symlinks (.claude/skills GC)

## Problem

weave (like `setup.sh`/`sync-local-skills.sh` before it) *creates* lowered symlinks — `.claude/skills/<prefix><name>` — but never *prunes* stale ones. When a skill is renamed (`xx-fix` → `xx-repair`) or the local prefix changes (`xx-` → `yy-`), weave writes the new lowered symlink and leaves the old one behind (dangling, or pointing at a gone target). Pre-existing gap inherited from the bash; surfaced during the weave (#95) cutover.

Note (a non-issue): the *source* dirs `construct/local` / `construct/adapted` are symlinked **as whole dirs**, so a source-side rename auto-propagates downstream via the dir symlink. The staleness is *only* in the per-name **lowered** symlinks weave emits.

## Spec

A prune/GC pass over weave's lowered locations (`.claude/skills/`, and any future lowered dir weave owns):
- Basic logic: a symlink in a lowered location whose target is **gone/dangling** → delete.
- Better: weave knows the *full intended set* of lowered symlinks for a repo, so remove any lowered symlink it did **not** (re)produce this run (orphan removal) — covers the rename/prefix-change case even when the old target still exists elsewhere.
- Scope to lowered dirs weave manages; never touch non-weave files in those dirs (track weave-owned entries).

## Done when

- Renaming a skill or changing the local prefix, then re-running weave, leaves **no** stale `.claude/skills/<old>` symlink.
- A dangling lowered symlink is removed on the next weave run.
- weave never deletes a lowered entry it doesn't own.

## Plan

- [ ]

## Log

### 2026-06-14
