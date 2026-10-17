---
collaboration-mode: {default: discussion}
process-instruction: "[free-form]"
discussions:
  - doc: design.side/02-lock-granularity.md
  - doc: design.side/03-retry-story.md
---

# Lock design

## Locking

Couch takes one lock per file. 🤖<The lock is held for as long as the operator is editing the file, and then it is released when they submit, so the agent can write the file after that and only after that.>{The operator holds it while editing and releases it on submit; only then can the agent write.}

A second writer is refused with exit 3. → [discussion](design.side/02-lock-granularity.md)

## Numbers

A round costs little. 🤖[add the numbers]{the latency numbers or the cost numbers?}

## Retries

A refused write is retried after the next submit. → [discussion](design.side/03-retry-story.md)

## Cache

The cache keeps entries for 🤖<a day>{an hour}. → [discussion](design.side/01-cache-ttl.md) (folded 2026-10-16)
