---
type: side-doc
parent: ../design.md
marker: "why do we lock per file, not per section?"
collaboration-mode: discussion
process-instruction: "[free-form]"
---

# Why lock per file, not per section?

_From [Lock design](../design.md), section "Locking"._

A section's boundaries move while the operator edits: a new heading splits one section into two, and a deleted one
merges two. A lock per section would have to follow those moves, and two locks could end up covering the same
lines. The file is the one unit whose identity doesn't change under editing, and it is what Couch hashes for the
compare-and-swap write.
