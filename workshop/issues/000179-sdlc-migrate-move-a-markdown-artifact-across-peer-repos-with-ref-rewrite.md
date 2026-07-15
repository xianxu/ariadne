---
id: 000179
status: open
deps: []
github_issue:
created: 2026-07-15
updated: 2026-07-15
estimate_hours:
---

# sdlc migrate: move a markdown artifact across peer repos with ref rewrite

## Problem

#171's direction (peer-repo `repo#id` addressing; artifact residency = soft
center-of-gravity default) makes moving a markdown artifact between repos a
NORMAL operation — a project file follows the work, an SDLC artifact leaves
brain. But a move today is a hand job with a silent correctness trap: **bare
`#NNN` refs inside the file are repo-relative.** Moved verbatim, they
re-resolve against the destination repo's issue numbering — pointing at
unrelated issues without any error. The rewrite rules are fixed patterns (the
formal ref grammar `sdlc resolve` already owns), so this belongs in the
binary, not in agent judgment.

## Spec

`sdlc migrate <path-to-file> <dest-repo-dir>` (e.g.
`sdlc migrate data/project/metis-v2.md ../kbench`) — deterministic, no LLM:

1. **Rewrite outbound refs.** Bare `#NNN` in the file body is source-relative
   → qualify to `<source-repo>#NNN` so it resolves identically from the
   destination. Refs already qualified with the DESTINATION repo
   (`<dest>#MMM`) may normalize to bare `#MMM` (they become local). Qualified
   refs to third repos pass through untouched. Skip fenced code blocks (the
   #66 meta-document lesson: docs about the ref grammar quote it literally).
2. **Verify before writing.** Every ref the rewrite touches must resolve —
   `sdlc resolve` machinery against the live fleet; a dangling ref aborts the
   migration with the offending ref named (fail-closed, nothing half-moved).
3. **Move + commit mechanics.** Write the rewritten file at the destination
   (same relative path by default, `--dest-path` to override), remove the
   source, and commit BOTH sides scoped to exactly the files touched
   (`git add -- <file>`, per the lessons.md broad-add rules) — or
   `--no-commit` to stage only. Refuse on a dirty destination (the
   propagate-base Rule-4 posture).
4. **Report inbound refs, don't rewrite them (v1).** Other files across the
   fleet pointing AT the moved artifact (by path, or `<source>#NNN` when the
   artifact is an issue) keep working only if the ref grammar is
   location-independent — issue refs are; PATH references are not. v1: scan
   the fleet (super-repo grep over the fixed patterns) and PRINT the inbound
   sites for the operator/agent to judge; no automatic fleet-wide rewrite.

Design questions (settle at plan time):
- **Issue artifacts renumber.** Issue IDs are per-repo sequences; migrating
  an issue file needs a fresh dest-side ID (allocate via the `issue new`
  counter) + a tombstone/redirect line in the source repo's Log or history.
  Possibly v2 — the #171 driver (project files, slug-named) needs no
  renumbering.
- Whether `migrate` refuses in a brain repo (it's a lifecycle-ish verb;
  probably yes via the #176 guard set — spine `WorkflowVerbs()` membership).
- Frontmatter touch-ups: bump `updated:`; anything else per datatype.

Related: supports #171 (not a blocking dep — the project-gate lift can land
first; migrate makes the residency default cheap to change).

## Done when

- `sdlc migrate <file> <dest-repo>` moves the file with all outbound bare
  refs correctly qualified, dest-local refs normalized, fenced blocks
  untouched, and every touched ref verified to resolve — or refuses naming
  the dangling ref.
- Both repos end with scoped commits (or staged with `--no-commit`); dirty
  destination refuses.
- Inbound references across the fleet are reported with file:line.
- A round-trip (migrate there, migrate back) is ref-stable — no rewrite
  churn.

## Plan

- [ ] design at start-plan: ref-rewrite rules as a pure entity over the
      existing ref grammar (ARCH-DRY with `sdlc resolve`), thin IO for
      move/commit; fixture repos for the round-trip test

## Log

### 2026-07-15

Filed from the #171 brainstorm (operator): moves become normal under
peer-repo addressing, rules are fixed patterns → binary-owned. Bare-ref
requalification + existence verification are the core; inbound refs are
report-only in v1.
