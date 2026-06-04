---
id: 000021
status: done
deps: []
created: 2026-05-05
updated: 2026-06-03
---

# claude warning some skill's not loaded

skills seem to large or sth

## Done when

- No skill fails to load due to an over-length `description:`.

## Spec

Root cause (operator): a skill's frontmatter `description:` was too long, which
Claude Code rejects → the skill silently didn't load ("skill's not loaded"
warning). Fix is just trimming the offending description under the cap.

## Plan

- [x] Resolved — the over-length description was trimmed. Verified 2026-06-03:
  every `construct/local` + `construct/adapted` SKILL.md `description:` is now
  well under Claude Code's ~1024-char cap (longest 334). No skill blocked.

## Log

### 2026-06-03 — done (verified)
- 2026-06-03: closed — skill-not-loading was an over-length frontmatter description:; verified fixed — all construct/local+adapted SKILL.md descriptions now well under Claude Code ~1024 cap (longest 334). Validation close: no attributable #21 commit (--no-actual), doc-only (--no-judge/--no-atlas).; review verdict: not-run
- Operator: this was the over-length `description:` case. Confirmed fixed —
  scanned all skill `description:` fields, longest is 334 chars (vs the ~1024
  cap), so nothing fails to load. No attributable #21 commit (trimmed
  incidentally); validation close.

### 2026-05-05

