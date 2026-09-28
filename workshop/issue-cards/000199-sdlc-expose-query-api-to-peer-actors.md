---
id: '000199'
status: punt
created: 2026-08-21
updated: 2026-09-02
---

# sdlc: expose query API to peer actors

## Problem

`sdlc`'s command line is the API between an agent and its own harness. In the
actor model being built in `pair#145` (couch), agents also need to answer
*queries from other actors* — "what is the state of `ariadne#111`" — without the
caller reaching into the callee's repo and without spending an LLM turn.

Two things are missing. There is no way for a command to declare that it is safe
to answer from outside, so today the choice is all-or-nothing. And there is no
way for a caller to discover what a given actor will answer, which matters
because binaries skew across repos and per-repo policy may differ.

The constraint that shapes the design: **the dialect is shared, the authority is
not.** Peer ariadne-based repos all understand the same vocabulary, but actor B
invoking `sdlc close` inside A's repo is permission laundering and breaks the
rule that only an actor interprets its own state. Test for the split: a call
that changes the *caller's* world is a query; one that changes the *callee's*
world is a command. Only queries get exposed.

Design context:
`brain/workshop/pensive/2026-08-20-01-pensive-couch-agent-switcher.md`.
