---
type: side-doc
parent: ../../notes.md
marker: "why per file and not per section?"
---

# Why per file and not per section?

_From [Notes](../../notes.md), section "Locking"._

Couch takes one lock per file, because a file is the unit it hashes for the compare-and-swap write, while a
section's boundaries move as the operator edits.
