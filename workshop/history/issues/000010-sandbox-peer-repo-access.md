---
id: 000010
status: wontfix
deps: []
github_issue:
created: 2026-04-23
updated: 2026-06-03
---

# Sandbox: support accessing peer projects

## Problem

The current sandbox mounts a single repo at `/sandbox/repo`. Peer repos on the host (sibling directories under `~/workspace/`) are invisible. This breaks:

- **Ariadne skill symlinks**: skills point to `../../../ariadne/.claude/skills/` — broken inside sandbox
- **Cross-repo references**: e.g. brain needs to read patterns from parley.nvim
- **Shared credentials**: e.g. Google OAuth setup in one repo, needed by another
- **Ariadne base layer itself**: the symlink target (`../ariadne/`) doesn't exist in sandbox

05/27/2026: this likely is outdated, we overhaul how ariadne decedents' dependencies work.

## Use Cases

1. AI agent in brain sandbox wants to reference code in parley.nvim
2. Symlinked ariadne base layer files need their target to exist
3. Shared tooling/config across ariadne-family repos
4. Cross-repo credential sharing (OAuth tokens etc.)

## Constraints

- OpenShell mounts are configured at sandbox creation (static filesystem policy)
- Mutagen syncs can be added dynamically
- Network policy is separate from filesystem
- Each sandbox is scoped to one repo today

## Spec

*(to be filled after brainstorming)*

## Plan

- [ ]

## Log

### 2026-06-03 — wontfix (superseded)
- Superseded by the construct base-layer overhaul (the 2026-05-27 note above):
  `construct/setup.sh --vendor` **copies** base-layer files into a repo instead of
  symlinking into a sibling `../ariadne/`, so the load-bearing breakage — "the
  symlink target `../ariadne/` doesn't exist in the sandbox" — no longer exists.
  A sandbox/CI repo with no visible peer uses vendor mode (see
  `atlas/workflow/base-layer.md`).
- Never implemented as framed (sandbox mounting sibling repos): Spec was empty,
  plan empty, no commits. The other use cases (arbitrary cross-repo reads,
  shared sandbox creds) aren't being pursued — the architecture moved away from
  needing them; if any resurfaces it's a fresh, narrower issue, not this one.

### 2026-04-23
- Originated in brain repo (issue 000001), moved to ariadne since this is base layer infrastructure
