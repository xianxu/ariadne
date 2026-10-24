# Examples: reference rounds for collaborative editing

Each directory here is one **case**: one round, meaning one submitted document and the writes that round makes. The
cases illustrate [collaboration-modes.md](../collaboration-modes.md). They are not instructions: a round never reads
them.

- `before/` is the round's input: the documents as the operator submitted them, with each document's `<name>.side/`
  folder.
- `after/` is what the round leaves, laid out the same way.
- `case.md` gives `submitted:` (the submitted document, relative to the tree), `outcome:` (`ok`, or `fails` when the
  round must change nothing), the changes since the agent's last read when the case depends on them, and the reply.
  A move case adds `move: OLD NEW` (collaboration-modes.md §5).

Links resolve relative to the file inside its own tree, exactly as on disk.

**A live check against a case:** copy its `before/` tree to a scratch directory, run a round on the `submitted:`
document as if its changes were the operator's, and compare the tree with `after/`. Wording may differ. Structure
must not: which asks became edits, side docs or answer docs; the link texts and paths; the side
docs' names and frontmatter; and which files are unchanged.

`pkg/reviewexport`'s `TestExamples` checks every case's side docs against the contract in collaboration-modes.md §4:
- names, frontmatter, parent links and backlinks;
- that an unlinked side doc for the same ask (parent, marker, section) is relinked, never duplicated, and that a
  fold marks the parent's link `(folded <date>)`;
- that no side doc is pruned and new ones take the next free number;
- that a side doc a round creates carries only `type`, `parent` and `marker` (and later its own index), never a
  mode key, while a pre-existing side doc keeps its legacy keys byte for byte (#575);
- the index (§4, "The index"): every `discussions:` entry names a side doc whose `parent:` is that document; a round
  never removes a link, drops an entry only when its link is gone, leaves no entry without its link, and adds one
  only for a side doc it wrote or linked;
- that a failed round leaves `after/` identical to `before/`;
- that a round never adds a second commentary turn to a marker, while unchanged legacy chains remain readable;
- that a move case's `after/` is exactly the move contract applied to `before/`;
- the chat grammar, through `pkg/sidedoc`'s `ChatSections` (`checkChat`): every answer link a round adds sits in a
  chat section, every numbered item of a `@@chat@@` section ends the round answered, and the reply names each
  misplaced tag.

**Exported fixtures.** `export.json` lists the files of the move cases and of the cases `exportedCases` names in
`pkg/reviewexport/export.go`, for pair and parley to pin:

| Cases | Show |
|---|---|
| `chat-tag`, `chat-tag-misplaced` | the `@@chat@@` carve-out: new items, a follow-up, a marker in the section's prose; a tag that is ordinary text |
| `chat-answer-invalid`, `frontmatter-unread-tag` | #581: a chat item whose link has no valid destination is still open; a tag and heading inside an unread header declare nothing |
| `ask-question`, `ask-command`, `ask-command-unclear`, `ask-unsure` | a marker as a question (side doc) or a command (edit), and the unclear and unsure cases |
| `scope-paragraph`, `scope-section`, `scope-document` | scope rules s1, s2 (with a brief) and s3 |
| `legacy-chat-scalar`, `legacy-chat-map`, `legacy-chat-comment`, `legacy-keys-ignored` | §7's legacy keys: the forms that still mean chat, and every retired key read around |
| `move-orphan`, `move-rename` | §5's moves, parley's mover fixtures |

**Older cases** keep legacy keys, maps and flat side-doc names in their documents as compatibility fixtures: their
`before/` files are what older documents look like, not what new writers emit. `chat` is the comprehensive legacy
map case (tl-lead's tests read it); `legacy-collaboration` shows the retired `collaboration` value no longer
answering in place. A reply to a document carrying legacy keys ends with §7's notice lines, exactly as `ChatSections` words them;
`checkChat` holds every case to that.

The work the cases file or dispatch (an issue number, a slot move) is illustrative.
