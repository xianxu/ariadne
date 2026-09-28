---
id: '000105'
status: done
created: 2026-06-15
updated: 2026-06-15
estimate_hours: 2
actual_hours: 0.43
---

# continuation datatype: connective-narrative procedure (flush-to-pensive, thread-arc/user-model, open-questions, lessons)

## Problem

The `continuation` datatype (`construct/datatype/continuation.md`) distills a
session's human-meaningful state into a durable handoff doc. Dogfooding (pair#61)
surfaced five gaps against what a high-quality handoff needs:

1. **No flush-first step.** The procedure assumes the session's durable artifacts
   already exist; nothing first captures un-recorded key exchanges, so insights
   die in the transcript.
2. **No lessons.** It records *this-work* decisions/dead-ends but not the
   transferable **lessons** learned in the session.
3. **Lists, doesn't connect.** "Pointers" enumerates files/issues but never
   explains their **history, reasoning, and connections** — the continuation
   isn't built *around* the durable artifacts.
4. **No thread arc / user model.** It captures no **arc of the thread** (where we
   started, the pivots, their underlying connection), no **model of the user's
   mental model** / latent intention, and no channel for the **open questions**
   that model leaves unresolved (to ask on resume).
5. **NEXT ACTION stands alone** — concrete, but not connected to the arc/lessons.

Relationships: mirror of **pair#61** (the dogfood origin). **ariadne#103** wants
the user-model *maintained live every turn*; this issue **persists/checkpoints**
that model into the continuation (the two are complementary halves). **ariadne#90**
(docflow suspend/resume + auto-summary) is the sibling-domain analog.
