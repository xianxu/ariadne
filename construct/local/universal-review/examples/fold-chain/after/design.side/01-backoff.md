---
type: side-doc
parent: ../design.md
marker: "how long should it back off?"
collaboration-mode: discussion
---

# How long should it back off?

_From [Retries](../design.md), section "Backoff"._

Exponential, from one second.

Decided (operator): cap the backoff at one minute. → [discussion](02-cap.md) (folded 2026-10-17)
