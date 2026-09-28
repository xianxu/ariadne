---
id: '000107'
status: done
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 12
actual_hours: 2.9
---

# weave: Option B — per-harness skill-dir lowering (`.claude/skills` + `.agents/skills`), retire the menu, prose entry files

## Problem

`weave compile --target T` should leave a repo holding EXACTLY one target's
artifacts. Today it doesn't — switching targets STRANDS the previous target's
artifacts. Verified on ariadne (24 live `.claude/skills/` links):

```
weave compile --target claude   → 24 .claude/skills/<name> symlinks + prose-only AGENTS.md
weave compile --target codex    → AGENTS.md WITH the `## Skills` menu, and the 24
                                   .claude/skills links LEFT UNTOUCHED (no prune)
```

After `claude → codex` the repo carries BOTH skill faces: the menu in AGENTS.md
AND the 24 stale `.claude/skills/` symlinks.

**Root cause — the prune is target-myopic.** `plan/prune.go`'s scan set is
`ManagedLocations(actions)` = "dirs weave produced a symlink into THIS run." A
backend the current target stops emitting (codex produces nothing under
`.claude/skills`) isn't a managed location → `ScanManagedSymlinks` never looks
there → the claude-era links are orphaned-but-unscanned. The prune can only GC
locations the CURRENT target actively writes into.

**Aggravators:** the stale links are gitignored (`EnsureGitignore` owns
`/.claude/skills/`) → invisible to `git status`; the staleness compounds (a
renamed/removed skill leaves a dangling link until you switch back); and a
menu-session reader that scans `.claude/skills` would see every skill twice.
