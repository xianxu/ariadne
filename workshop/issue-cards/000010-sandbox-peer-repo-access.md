---
id: '000010'
status: wontfix
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
