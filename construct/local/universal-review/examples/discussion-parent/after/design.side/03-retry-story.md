---
type: side-doc
parent: ../design.md
marker: "the retry story feels off"
---

# What is off in the retry story?

_From [Lock design](../design.md), section "Retries"._

The sentence covers the lock (exit 3) but not a stale write (exit 6). A stale write isn't retried after the next
submit: the agent rereads at once and recomputes. **Proposal:** "A write refused by the lock waits for the next
submit; a stale write is recomputed at once."
