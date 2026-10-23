# Universal review skill

`construct/local/universal-review/` (lowered as `xx-universal-review`) answers the
operator's document review requests in any slot (ariadne#399, S2a of the ops project
universal-agent-input-interface). Parley's Alt+Return sends only `submitted: PATH`; the
skill carries the rest.

- **`SKILL.md`** — the trigger and three pointers; it restates no protocol.
- **`round.md`** — the harness-neutral round (#575): read the chat tags and legacy keys, load a brief when an ask
  names one, scope each command (s1–s3), route each marker as a question or a command and each chat item to an
  answer doc, handle markers, and reply with one notice line for legacy keys. It names no Couch verb, so it works
  outside Couch.
- **`collaboration-modes.md`** (titled "Collaborative editing"; the path is kept because `pkg/sidedoc`, tl-lead and
  sdlc cite its §4/§5) — the one-mode model of #575, which replaced #562's three keys (54 combinations): direct
  edits are content; a marker is a **question** (a linked side doc; the default when unsure) or a **command** (an
  edit, a one-line `Decided (operator): …` record, a fold-in, or an action elsewhere); §2 scope rules s1–s3; §3 the
  `@@chat@@` carve-out (a tag directly above a heading; numbered items, each with a dated answer doc; follow-ups,
  deletions, relinking); §4 the side-doc contract (new `<name>.side/YYYY-MM/YYYY-MM-DD-NN-slug.md` paths, old flat
  names valid, frontmatter of only `type`/`parent`/`marker`, the write order, retries relinking an unlinked doc by
  `parent:` + `marker:` + section); §5 moves; §6 the state table; §7 the legacy keys, read and never written: the
  forms that meant chat still mean chat, everything else is a reply notice, never a failed round. The operator
  decided (10-23) that the one mode applies to issues, plans and projects too: `collaboration-mode: collaboration`
  no longer answers in place.
  **S3 (#424):** a side doc has no closed state; the parent's `discussions:` frontmatter is the operator's review
  queue, under a one-way rule (deleting a link drops its entry; removing an entry means reviewed); fold-in is asked
  from the parent by 🤖[fold in] at the link; and §5 is the move contract (the `.side/` tree moves with its document,
  a name change rewrites the parent's references, a taken folder name suffixes the orphan). `pkg/sidedoc` is the Go
  home of the links, the index, the tree and now the chat grammar (`ChatSections`, `RetiredKeys`), shared by the
  checker, tl-lead and sdlc's archiving.
  **Remaining writers:** sdlc's issue and project scaffolds, tl-lead's side-doc migration, the datatype
  `collaboration-mode` rows (with `vocab.CollaborationModes`) and the writing-plans header still write mode keys;
  ariadne#576 removes them. Ariadne's `cmd/tl-lead/answers.go` validates answer-link targets before index/section
  publication.
- **`briefs/*.md`** — five named editing briefs (copy-editing, line-editing, proofreading, developmental,
  fact-check), loaded when an ask calls for one; name-only frontmatter (#575 retired the v3 order/scope/frontier
  flags and the free-form mode). Parley snapshots them.
- **`examples/`** — reference rounds, one per case: `before/` and `after/` trees plus `case.md` (the submitted
  document, the outcome, the changes, the reply; a move case adds `move: OLD NEW`). A live check runs a round on a
  copy of `before/` and compares it with `after/`. `pkg/reviewexport`'s `TestExamples` holds every case to the
  side-doc contract, the index rules and a move case's exact projection, and (`checkChat`) to the chat grammar via
  `ChatSections`; a new side doc may carry no mode key. The move cases and the 13 cases `exportedCases` names
  (chat tag, question/command, scope, legacy) are exported as fixtures for pair and parley; older cases keep legacy
  documents as compatibility fixtures. `TestNoSpecWritesModeKeys` guards that no skill text tells a writer to write
  a mode key.
- **`adapters/file-edit.md`** — editing the file. A document is edited directly only on a
  confirmed untracked lookup; a `type: session-sync` file (tracked even outside a slot) or a registered one goes through
  `couch --session-sync` status → diff → `read --all` (the snapshot, hash H) →
  atomic `apply` records with per-edit `explain` and answered fold IDs → `round --json`
  for the reply. Exit 1 invalid/proposal refusals either rebuild the full batch or open a
  linked side doc; exit 3 (lock) defers to the next submit without polling; exit 6 (stale)
  re-reads and rebuilds. New documents are highlight; the propose passages describe Couch until pair#838,
  for an older document whose legacy `presentation-mode: propose` it still honors. There, a record replacing the
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
- **`export.json`** — the versioned manifest (`ariadne/universal-review-export` **v4**, #575) parley's sync script
  snapshots: round, grammar (with `source`), collaboration-modes (exactly one, #412), adapters, flagless `brief`
  members, and the `fixture` members (the move cases and `exportedCases`), each with the resolved file's bytes and
  sha256, in a total order (role, name, path). v4 is a deliberate break: parley's sync allows v1–v3 and refuses v4
  before writing, so an installed parley keeps its shipped v2 snapshot until parley.nvim#419 migrates consumer and
  snapshot together; no other repo reads it. Built by `pkg/reviewexport`; `TestExportCurrent` fails when it is
  stale, and `UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport` regenerates it.

Composition: an agent in Couch loads the skill (round + grammar + collaborative editing + briefs as asked +
file-edit). Parley's standalone prompt is round + grammar + the briefs + its own
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
