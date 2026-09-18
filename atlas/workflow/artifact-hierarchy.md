# Artifact Hierarchy

## Principle

Work artifacts live close to the issue, then graduate to permanent locations or get archived.

## Locations

| Path | Purpose | Lifecycle |
|------|---------|-----------|
| `workshop/issues/` | Active issue files | Archived to history when done |
| `workshop/plans/` | Detailed designs for issues outside the quick-flow shell (#231), authored via the `superpowers-writing-plans` skill (the canonical plan path, #72); also boundary-review sidecars (`-close-review.md` / `-m<x>-review.md`, #136) | Archived with issue |
| `workshop/history/` | Completed issue files | Permanent archive, low-signal |
| `workshop/staging/` | Work-in-progress scratch | Temporary |
| `docs/vision/` | Pensive docs, brainstorms | Permanent thinking artifacts |
| `atlas/` | Sketch-level documentation | Permanent, updated with code |

## Rules

- **Simple case** (inside the quick-flow shell, #231): everything lives in the single issue file, and change-code infers the quick flow
- **Complex case**: issue file + detailed plan in `workshop/plans/` — whose presence makes change-code infer the full flow — (same `NNNNNN-slug` filename with `-plan` suffix), authored via the `superpowers-writing-plans` skill — version-controlled, never the harness builtin's ephemeral `~/.claude/plans/` (#72)
- **When done**: the issue + every `workshop/plans/NNNNNN-*` artifact sharing its id prefix (durable plan + boundary-review sidecars) move to `workshop/history/` — issues into `history/issues/`, plans + sidecars into `history/plans/` (#181; per-kind subdirs derived by `vocab.ArchiveSubdirs`, reads tolerate the pre-#181 flat layout) — swept together at `sdlc merge`/`push` (#143)
- **Atlas**: updated during pre-merge checks to reflect what was built; never exhaustive
- **History**: avoid reading unless explicitly asked — it's archive, not reference
