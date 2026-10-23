---
type: side-doc
parent: ../design.md
marker: "why?"
---

# Why is the lock released on submit?

_From [Locks](../design.md), section "Files"._

A submit is the operator saying the text is ready for the agent; holding the lock past it would block the very write
the submit asks for.
