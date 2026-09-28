---
id: '000125'
status: done
started: 2026-06-25T14:33:20-07:00
created: 2026-06-25
updated: 2026-06-25
estimate_hours: 1.56
actual_hours: 0.79
---

# sdlc embeds a help-text fragment GENERATED from the issue vocabulary (stop hand-maintaining lifecycle prose)

## Problem

`sdlc`'s embedded help text (`cmd/sdlc/helptext/*.md`) hand-restates lifecycle facts and
**drifts**. #122 M4 made `set-status.md`'s "All other transitions are allowed without
guards" *false* (the lifecycle gate now refuses non-modeled flips); nothing caught it
automatically — the #122 whole-issue fresh-eyes review did (FIX-THEN-SHIP), and it was
patched by hand. So the operator-facing prose is a **hand-maintained shadow** of the
model, not derived — the exact drift the vocabulary layer (#122) exists to kill, for the
one consumer we never wired. We already generate-from-source where the target is
*templated* — code → `issue.json` via `go:embed`; the vocabulary skill `SKILL.md` via the
`.dynamic-skill` renderer — the gap is the *free-form help prose*.
