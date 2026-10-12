---
name: couch
description: Use when coordinating with another live Couch slot — handing off authorized independent work, asking a peer slot a question — or when interpreting an incoming Couch message. The `couch` binary owns the peer-messaging protocol; this skill loads it.
---

# couch — peer messages between live slots

Couch connects live agent slots under one supervisor. The `couch` binary is the
single source of truth for its protocol. That covers slot discovery, addressing,
receipts, safe delivery and loop-control conventions. This skill is a static
pointer and intentionally carries no copy of the protocol, so it can never drift.

**Load the protocol:** run `couch --skill` and follow the instructions it prints.
Read it fresh each time; do not rely on memory of an earlier version.

Loading this skill sends nothing and starts no work. Message a peer only when
your task calls for it, and only as `couch --skill` describes.

**In a Couch slot**, the constitution's §15 (Couch Slots) applies on top of the
protocol: report every item's outcome to the hub (`couch --send-to hub`), ask the
hub rather than an in-pane question prompt, and call `couch` as a plain leading
command (nothing piped or chained after it) with no `$(`, `<(` or backticks in its arguments.

**If `couch` is not installed** (`command -v couch` finds nothing), stop. Tell
the operator that Couch peer messaging is unavailable in this environment
because the `couch` binary from the pair repository is not on `PATH`. Do not
improvise a substitute channel: no shared files, no other sessions' terminals,
no ad-hoc sockets.
