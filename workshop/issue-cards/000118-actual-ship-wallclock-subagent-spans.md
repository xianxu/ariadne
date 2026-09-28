---
id: '000118'
status: done
started: 2026-06-18T10:30:48-07:00
created: 2026-06-18
updated: 2026-06-18
estimate_hours: 0.96
actual_hours: 0.72
---

# Measure ship wall-clock: fill subagent spans in active-time so actual matches the estimate's unit

## Problem

`sdlc actual` (active-time) measures the operator's **interaction time** — it sums
gaps between events in the operator's main transcript, capping each at 15 min to
delete idle. But the estimate (`estimate_logic-v2` build-effort) targets the thing
we actually want: **wall-clock for one engineer + AI to ship**. These are different
units, and the gap is the bulk of the apparent "v2 overshoots ~3.5×" finding —
*a measurement artifact, not an estimation error.*

The unit mismatch has a precise cause: **subagent execution spans are wrongly
truncated as idle.** When the main agent dispatches a Task subagent, the subagent
runs in its own transcript (outside the dirs active-time reads), so the operator's
main session shows `dispatch → … → return` as one big gap — capped at 15 min. A
2-hour subagent build registers as ~15 min. But that span is **active project work**
(the AI is shipping), and it is **bounded by observable events** (the Task tool-use
and its tool-result are both in the operator's transcript, timestamped). So we can
distinguish "subagent grinding" from "operator at lunch" — the blanket 15-min cap
conflates them.

Reframe (operator, 2026-06-18): the estimate exists for **launch predictability**
(when can we ship), not for tracking operator-attention. So `actual` should measure
**ship wall-clock**, and operator-attention is *not* the right unit — it belongs one
level up, as the parallelism/throughput limit (one human ≈ 2 concurrent sessions),
not as the per-issue measure. This supersedes #112's operator-attention-model
direction (parked → effectively moot).
