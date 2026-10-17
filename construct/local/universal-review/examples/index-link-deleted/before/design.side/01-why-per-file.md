---
type: side-doc
parent: ../design.md
marker: "why per file?"
collaboration-mode: discussion
---

# Why one lock per file?

_From [Locks](../design.md), section "Granularity"._

A section lock would let two writers edit one file, but the merge would need a three-way diff Couch doesn't have.
