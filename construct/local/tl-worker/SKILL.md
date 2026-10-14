---
name: tl-worker
description: Use when you work an item in a Couch slot whose hub is a TL — a dispatched issue, a question to the hub, staging for a live check, landing on SHIP. The `tl-worker` binary owns the worker's rules; this skill loads them.
---

# tl-worker — a worker slot in a TL-run fleet

The `tl-worker` binary is the single source of the rules a worker slot follows
under a TL hub. This skill is a static pointer and carries no copy of them, so
it can never drift.

**Load the rules:** run `tl-worker --help` at the start of each dispatched item
and follow what it prints. Read it fresh each time.

**If `tl-worker` is not installed** (`command -v tl-worker` finds nothing), tell
the hub (`couch --send-to hub`) that the binary is missing from this
environment, and continue under AGENTS §15.
