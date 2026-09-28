---
id: '000101'
status: done
created: 2026-06-15
updated: 2026-06-16
estimate_hours: 1
actual_hours: N/A
---

# weave: support no-prefix skill lowering (bare names when localPrefix is empty)

## Problem

The local-skill namespace prefix is configured in a layer's `construct/config.json` as `{"localPrefix": "xx-"}`. weave's skill lowering (`walk/skill_symlinks.go`, `walk/skills.go`) reads `localPrefix(fs, layer.Path)` and lowers `construct/local/<dir>` → `.claude/skills/<prefix><dir>` (and `construct/adapted/<dir>` → bare). The default is `"xx-"`, mirroring the retired `sync-local-skills.sh`'s `PREFIX="${PREFIX:-xx-}"` fallback — which treats an **empty** prefix the same as **absent** (both → `xx-`). So there is currently **no way to lower local skills with bare names** (`fix`, `pensive`, `voice-apply`) instead of prefixed (`xx-fix`, …).
