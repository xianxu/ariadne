---
id: '000209'
status: done
started: 2026-09-07T16:11:22-07:00
created: 2026-09-02
updated: 2026-09-07
estimate_hours: 3.06
actual_hours: 4.49
---

# queue datatype for advisory work ordering

## Problem

**Sequence has nowhere to land.** The existing nouns each hold something else:
an issue holds *what*, a project holds *scope*, `deps:` holds *hard blocking*,
a roadmap holds the month-level aggregate, a project's `## Breakdown` holds
ordering *within* that project. Nothing holds "of the loose issues that could
be done in any order, which one is next, and why."

So it lives in conversation and dies with it. On 2026-09-02 a single advisory
session produced five issues across two repos — `pair#170`, `#171`, `#172`,
`ariadne#206`, `#207` — with a real order between them (`#207` after `#206`,
same dispatch; `#172` after `#170`, needs its switch semantics; `#171`'s
measurement before `#171`'s design). None of that ordering is recorded anywhere.
Meanwhile `sdlc state` prints ariadne's 16 open issues in ID order with no
ranking at all.
