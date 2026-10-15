# Universal review skill

`construct/local/universal-review/` (lowered as `xx-universal-review`) answers the
operator's document review requests in any slot (ariadne#399, S2a of the ops project
universal-agent-input-interface). Parley's Alt+Return sends only `submitted: PATH`; the
skill carries the rest.

- **`SKILL.md`** — the trigger and three pointers; it restates no protocol.
- **`round.md`** — the harness-neutral round: the frontmatter `process-instruction:
  [mode][mode] text`, mode resolution (exact name, else a unique prefix; the first mode's
  flags govern), the flags as directives, marker handling, the reply. It names no Couch
  verb, so it works outside Couch. §4 is reserved for S2's section modes.
- **`modes/*.md`** — the six stock modes, canonical here since #399; parley snapshots them.
- **`adapters/file-edit.md`** — editing the file. A document is edited directly only on a
  confirmed untracked lookup; a `type: session-sync` file or a registered one goes through
  `couch --session-sync` status → diff → `read --all` (the snapshot, hash H) →
  `write --all --if-hash H --fold <the IDs read in status>`. Exit 3 (lock) defers to the
  next submit; exit 6 (stale) recomputes.
- **`review-convention.md`** — a symlink to `construct/local/fix/review-convention.md`, the
  one authored marker grammar.
- **`export.json`** — the versioned manifest (`ariadne/universal-review-export` v1) parley's
  sync script snapshots: round, grammar (with `source`), adapters, and dense mode records,
  each with the resolved file's bytes and sha256, in a total order. Built by
  `pkg/reviewexport`; `TestExportCurrent` fails when it is stale, and
  `UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport` regenerates it.

Composition: an agent in Couch loads the skill (round + grammar + modes + file-edit).
Parley's standalone prompt is round + grammar + the selected modes + its own
`propose_edits` adapter; it leaves out `SKILL.md` and `adapters/file-edit.md` (parley.nvim#319).
