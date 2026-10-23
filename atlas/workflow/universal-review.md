# Universal review skill

`construct/local/universal-review/` (lowered as `xx-universal-review`) answers the
operator's document review requests in any slot (ariadne#399, S2a of the ops project
universal-agent-input-interface). Parley's Alt+Return sends only `submitted: PATH`; the
skill carries the rest.

- **`SKILL.md`** — the trigger and three pointers; it restates no protocol.
- **`round.md`** — the harness-neutral round: single-valued `processing-mode` (exact name or unique prefix),
  separate `processing-instruction` free text, per-region scope/frontier and effective `presentation-mode`,
  marker handling and the reply. Deprecated composite `process-instruction` remains readable; automatic fill
  refuses a multi-mode legacy value rather than dropping a brief. It names no Couch verb, so it works outside
  Couch. §4 resolves settings and routes each ask (#412, #562).
- **`collaboration-modes.md`** — owns the three-key contract (#562): `collaboration-mode`, `presentation-mode`,
  `processing-mode`; scalar frontmatter, heading comments, precedence, defaults, aliases and legacy-map migration.
  Where an ask goes: **discussion** (the default; an imperative is an
  edit, an instruction that only acts is honored, a decision becomes one `Decided (operator): …` line with any notes in a
  linked side doc, a question opens a linked side doc), **collaboration** (answered in
  place), **chat** (each list item gets a dated answer doc). It holds the heading matching and its error cases, the
  side-doc contract (new `<name>.side/YYYY-MM/YYYY-MM-DD-NN-slug.md` paths, old flat names remain valid,
  frontmatter, the write order: side docs, then a folded parent, then the submitted document; retries relink an
  unlinked doc by `parent:` + `marker:` + section, never one already linked), and the state table. Design: ops universal-agent-input-interface, S2.
  **S3 (#424):** a side doc has no closed state; the parent's `discussions:` frontmatter is the operator's review
  queue, under a one-way rule (deleting a link drops its entry; removing an entry means reviewed); fold-in is asked
  from the parent by 🤖[fold in] at the link (an approval inside a side doc only records its Decided line); and §5
  is the move contract (the `.side/` tree moves with its document, a name change rewrites the parent's references,
  a taken folder name suffixes the orphan). `pkg/sidedoc` is the Go home of the links, the index and the tree, shared
  by the checker, tl-lead and sdlc's archiving. It includes `→ [proposal](…)`
  links for large rewrites.
  **Per-type defaults (#415):** every managed type declares its default where the type is defined. Issues and
  projects get theirs from `scaffold.collaboration_mode` in their CUE model, which `sdlc issue new` and
  `sdlc project new` write. Datatypes declare it in their prototype's Frontmatter shape table
  ([Data Artifacts](data-artifacts.md)). Durable plans get it from the writing-plans header (discussion). Side docs
  and answer docs follow §4 of collaboration-modes.md (discussion). The TL session file follows tl-lead's help.
  `vocab.CollaborationModes` is the Go copy of the three values. Under #562's fill contract, creators
  (Parley, Couch and tl-lead's side-doc helpers) write all three mode keys explicitly, preserving set values.
  Legacy-only choices are materialized under canonical keys without rewriting aliases; there is no blanket
  rewrite of existing documents. Runtime adoption belongs to the corresponding consumer implementation.
  Ariadne's `cmd/tl-lead/answers.go` validates answer-link targets before index/section publication: new
  month paths require a real filename date matching the sole month directory and a parent resolved from
  that directory; legacy dated flat targets remain accepted. It does not fill or rewrite answer documents.
- **`modes/*.md`** — the six stock modes, canonical here since #399; parley snapshots them.
- **`examples/`** — reference rounds, one per case: `before/` and `after/` trees plus `case.md` (the
  submitted document, the outcome, the changes, the reply; a move case adds `move: OLD NEW`). A live check runs a round
  on a copy of `before/` and compares it with `after/`. `pkg/reviewexport`'s `TestExamples` holds every case to the
  side-doc contract and the index rules (a test-only checker, mutation-tested), and a move case to an exact projection
  of the move contract. Older cases deliberately preserve legacy aliases/maps/flat names. New editorial cases
  cover heading overrides, explicit defaults, preserved existing/legacy choices and lossless multi-mode refusal.
  The checker validates month directories, cross-layout numbering and retry identity as well as legacy files.
  Only the move cases are exported, as parley's fixtures.
- **`adapters/file-edit.md`** — editing the file. A document is edited directly only on a
  confirmed untracked lookup; a `type: session-sync` file (tracked even outside a slot) or a registered one goes through
  `couch --session-sync` status → diff → `read --all` (the snapshot, hash H) →
  atomic `apply` records with per-edit `explain` and answered fold IDs → `round --json`
  for the reply. Exit 1 invalid/proposal refusals either rebuild the full batch or open a
  linked side doc; exit 3 (lock) defers to the next submit without polling; exit 6 (stale)
  re-reads and rebuilds. Under effective `presentation-mode: propose`, a record replacing the
  operator's ready marker with a side-doc link, or answering a fold-in marker at
  that link, applies directly; ordinary prose edits still become proposals. Its
  matching `discussions:` entry rides in the same atomic apply. A marker-free
  link can be a pending proposal while that entry applies directly; the shared
  side-doc parser extracts every Pair-escaped link inside a pending marker so the index
  remains until the operator accepts or rejects. Untracked `propose` documents
  refuse a direct write; untracked direct/highlight rounds report an observed
  before/after diff instead of claiming a Couch round.
  Generated content and frontmatter-only edits retain `write`.
  "Changes since your last read" never hold the agent's own edits: Couch's diff baseline moves on its write, and an untracked file is
  diffed against the copy the agent saved after its last write (operator requirement).
- **`review-convention.md`** — a symlink to `construct/local/fix/review-convention.md`, the
  one authored marker grammar. **#444:** agents write at most one commentary
  turn per marker; an unclear ask or continued exchange becomes a linked side
  doc. Older chains remain readable and resolve by their last human turn.
  `TestExamples` rejects a newly appended turn in its before/after cases.
- **`export.json`** — the versioned manifest (`ariadne/universal-review-export` v3) parley's
  sync script snapshots: round, grammar (with `source`), collaboration-modes (exactly one, #412), adapters, dense mode
  records (without the retired mode-level deletion knob), and (v2, #424) the move cases' files as `fixture` members for parley's mover (parley.nvim#373),
  each with the resolved file's bytes and sha256, in a total order. Built by
  `pkg/reviewexport`; `TestExportCurrent` fails when it is stale, and
  `UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport` regenerates it.

Composition: an agent in Couch loads the skill (round + grammar + collaboration modes + modes + file-edit).
Parley's standalone prompt is round + grammar + the selected modes + its own
`propose_edits` adapter; it leaves out `SKILL.md` and `adapters/file-edit.md` (parley.nvim#319).
The hosted Pair review workbench instead follows `construct/local/fix/SKILL.md`:
`pair review apply` owns its file write and returns the exact Body for a changed
reply's immediate docflow round commit. Every completed review calls apply, even
with `{"records":[]}`; an empty reply creates no agent commit (#453). Explanations
and questions stay in the session response. Pair #690 owns empty-round artifacts,
clearing the pane's spinner/pending block, and the live pane check; older Pair
accepts the producer's empty apply without a round artifact. There is no handoff
or landed-file step (#437).
An uncertain apply response is resolved with `pair review round FILE --json`,
whose `Current` flag says whether its stored `Body` still describes the file.
Before committing that Body, also match `Round.base_hash` to the attempted H
and read back the result and spans: an older current round is not proof the
last request reached Pair.
