---
collaboration-mode: collaboration
process-instruction: "[free-form]"
---

# Notes

## Locking

Couch takes one lock per file, because a file is the unit it hashes for the compare-and-swap write, while a
section's boundaries move as the operator edits.
