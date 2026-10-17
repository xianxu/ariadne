# Universal review skill

`construct/local/universal-review/` (lowered as `xx-universal-review`) answers the
operator's document review requests in any slot (ariadne#399, S2a of the ops project
universal-agent-input-interface). Parley's Alt+Return sends only `submitted: PATH`; the
skill carries the rest.

- **`SKILL.md`** — the trigger and three pointers; it restates no protocol.
- **`round.md`** — the harness-neutral round: the frontmatter `process-instruction:
  [mode][mode] text`, mode resolution (exact name, else a unique prefix; the first mode's
  flags govern), the flags as directives, marker handling, the reply. It names no Couch
  verb, so it works outside Couch. §4 resolves the collaboration modes and routes each ask (#412).
- **`collaboration-modes.md`** — where an ask's answer goes, per section, by frontmatter `collaboration-mode:`
  (separate from the review mode, which says how an edit is made): **discussion** (the default; an imperative is an
  edit, an instruction that only acts is honored, a question opens a linked side doc), **collaboration** (answered in
  place), **chat** (each list item gets a dated answer doc). It holds the heading matching and its error cases, the
  side-doc contract (one flat `<name>.side/` folder per document, `NN-slug` and `YYYY-MM-DD-NN-slug` names,
  frontmatter, the write order: side docs, then a folded parent, then the submitted document; retries relink an
  unlinked doc by `parent:` + `marker:` + section, never one already linked), and the state table. Design: ops universal-agent-input-interface, S2.
- **`modes/*.md`** — the six stock modes, canonical here since #399; parley snapshots them.
- **`examples/`** — ten reference rounds, one per case: `before/` and `after/` trees plus `case.md` (the submitted
  document, the outcome, the changes, the reply). A live check runs a round on a copy of `before/` and compares it
  with `after/`. `pkg/reviewexport`'s `TestExamples` holds every case to the side-doc contract (a test-only checker,
  mutation-tested); the examples are not exported.
- **`adapters/file-edit.md`** — editing the file. A document is edited directly only on a
  confirmed untracked lookup; a `type: session-sync` file (tracked even outside a slot) or a registered one goes through
  `couch --session-sync` status → diff → `read --all` (the snapshot, hash H) →
  `write --all --if-hash H --fold <the IDs from status that this write answers>`. Exit 3 (lock) defers to the
  next submit; exit 6 (stale) recomputes. "Changes since your last read" never hold the
  agent's own edits: Couch's diff baseline moves on its write, and an untracked file is
  diffed against the copy the agent saved after its last write (operator requirement).
- **`review-convention.md`** — a symlink to `construct/local/fix/review-convention.md`, the
  one authored marker grammar.
- **`export.json`** — the versioned manifest (`ariadne/universal-review-export` v1) parley's
  sync script snapshots: round, grammar (with `source`), collaboration-modes (exactly one, #412), adapters, and dense mode records,
  each with the resolved file's bytes and sha256, in a total order. Built by
  `pkg/reviewexport`; `TestExportCurrent` fails when it is stale, and
  `UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport` regenerates it.

Composition: an agent in Couch loads the skill (round + grammar + collaboration modes + modes + file-edit).
Parley's standalone prompt is round + grammar + the selected modes + its own
`propose_edits` adapter; it leaves out `SKILL.md` and `adapters/file-edit.md` (parley.nvim#319).
