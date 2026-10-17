---
collaboration-mode: {default: discussion}
process-instruction: "[free-form]"
---

# Lock design

## Locking

Couch takes one lock per file. The lock is held for as long as the operator is editing the file, and then it is released when they submit, so the agent can write the file after that and only after that. 🤖[tighten this paragraph]

A second writer is refused with exit 3. 🤖[why do we lock per file, not per section?]

## Numbers

A round costs little. 🤖[add the numbers]

## Retries

A refused write is retried after the next submit. 🤖[the retry story feels off]

## Cache

The cache keeps entries for a day. → [discussion](design.side/01-cache-ttl.md) 🤖[fold in]
