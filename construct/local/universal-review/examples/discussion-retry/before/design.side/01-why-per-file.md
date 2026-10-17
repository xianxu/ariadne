---
type: side-doc
parent: ../design.md
marker: "why per file?"
collaboration-mode: discussion
---

# Why per file?

_From [Locks](../design.md), section "Locks"._

A file is the unit Couch hashes for its compare-and-swap write; a section's boundaries move while the operator edits.
