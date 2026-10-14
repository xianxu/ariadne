---
name: tl-lead
description: Use when acting as the TL (technical lead) in a Couch hub — supervising worker slots, dispatching issues, reviewing plans as product steward, deciding landing and smoke tests, keeping the project file and the operator's status file — or when taking over or resuming a TL session.
---

# tl-lead — the TL in a Couch hub

The `tl-lead` binary is the single source of the TL's procedure. This skill is a
static pointer and carries no copy of it, so it can never drift.

**Load the procedure:** run `tl-lead --help` at the start of a TL session and
after a compaction, and follow what it prints. Read it fresh each time.

**If `tl-lead` is not installed** (`command -v tl-lead` finds nothing), tell the
operator that the binary is missing from this environment, and stop before
dispatching anything.
