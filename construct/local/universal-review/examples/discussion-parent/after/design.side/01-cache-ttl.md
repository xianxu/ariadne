---
type: side-doc
parent: ../design.md
marker: "how long should the cache keep entries?"
collaboration-mode: discussion
process-instruction: "[free-form]"
---

# How long should the cache keep entries?

_From [Lock design](../design.md), section "Cache"._

A day is longer than any round needs: a snapshot is stale once the operator submits again, and that happens within
the hour. **Proposal:** keep entries for an hour.

Folded into the parent 2026-10-16, as a proposal to accept or reject: the cache keeps entries for an hour.
