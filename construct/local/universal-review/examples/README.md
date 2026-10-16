# Examples: reference rounds for the collaboration modes

Each directory here is one **case**: one round, meaning one submitted document and the writes that round makes. The
cases illustrate [collaboration-modes.md](../collaboration-modes.md). They are not instructions: a round never reads
them.

- `before/` is the round's input: the documents as the operator submitted them, with each document's `<name>.side/`
  folder.
- `after/` is what the round leaves, laid out the same way.
- `case.md` gives `submitted:` (the submitted document, relative to the tree), `outcome:` (`ok`, or `fails` when the
  round must change nothing), the changes since the agent's last read when the case depends on them, and the reply.

Links resolve relative to the file inside its own tree, exactly as on disk.

**A live check against a case:** copy its `before/` tree to a scratch directory, run a round on the `submitted:`
document as if its changes were the operator's, and compare the tree with `after/`. Wording may differ. Structure
must not: which asks became edits, side docs, answer docs or clarifying turns; the link texts and paths; the side
docs' names and frontmatter; and which files are unchanged.

`pkg/reviewexport`'s `TestExamples` checks every case's side docs against the contract in collaboration-modes.md §4:
- names, frontmatter, parent links and the uniqueness of `(parent, marker)`;
- that no side doc is pruned and new ones take the next free number;
- that a failed round leaves `after/` identical to `before/`.

The work the cases file or dispatch (an issue number, a slot move) is illustrative.
