---
name: xx-universal-review
description: Use when a message reads `submitted: PATH` (or `review request: PATH`) for a markdown document — the operator's Alt+Return in parley — or when asked to review a document, including an older one whose frontmatter carries legacy review mode keys.
---

# xx-universal-review — answer a document review request

The request carries only the document's path. Everything else comes from the
document and the files beside this one, so read them each round rather than
from memory:

1. **The round:** read [round.md](round.md) and follow it. It applies the one
   collaborative mode of [collaboration-modes.md](collaboration-modes.md) (read
   that too): direct edits are content, each 🤖 marker is a question (a side
   doc) or a command (an edit, scoped s1–s3), and a `@@chat@@` section's items
   always get answer docs. It loads a brief from [briefs/](briefs/) when an ask
   calls for one. Legacy mode keys are read, noticed and never written.
2. **The grammar:** [review-convention.md](review-convention.md) is the marker
   grammar. Read it before writing any marker.
3. **The adapter:** you edit the file yourself, so follow
   [adapters/file-edit.md](adapters/file-edit.md). For a Couch-tracked
   document it goes through the operator's lock and a compare-and-swap apply;
   never edit one directly.

`export.json` lists these files with their hashes, for parley's standalone
copy; it is generated, so don't edit it by hand.
