---
id: 000065
status: working
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 0.5
---

# single-pass atomic work should use plain checkboxes not an M1 tag — avoids redundant milestone-close + issue-close double-log

## Problem

Closing #63 (a single-pass, atomic task) produced two near-identical `## Log`
lines:

```
- 2026-06-02: closed — <verified>          (sdlc close)
- 2026-06-02: closed M1 — <verified>       (sdlc milestone-close --milestone M1)
```

Root cause is **guidance, not a tool bug**. Both verbs funnel through
`runClose` (`milestoneclose.go:99` → `close.go:361-368`), which appends one
`- <date>: closed [Mx] — <verified>` line per call when `--verified` is set.
#63's Plan carried a single `- [x] M1 — …` row; the issue-close guard requires
*each milestone listed in `## Plan`* to carry a `Review-Verdict:` trailer, and
only `milestone-close` produces that trailer. So tagging the one-shot task `M1`
*forced* a `milestone-close M1` (log line #1) before the issue `close` (log
line #2). For a multi-milestone issue this reads as a clean progression
(`closed M1`, `closed M2`, … then `closed`); with exactly one milestone the two
lines carry the same verified text and sit adjacent → reads as a dup.

AGENTS.md §3 actively invites this: *"Single-pass work → one milestone, not
three"* — which says to tag atomic work as `M1`, the very thing that causes the
double-log.

## Spec

Tighten the guidance so genuinely atomic single-pass work uses **plain `## Plan`
checkboxes (no `Mx` tag)** and closes in a single `sdlc close` gesture (one
review boundary, one log line). Reserve `Mx` tags for work with ≥2 boundaries
you'll genuinely `milestone-close` separately. Make clear the mandatory
fresh-eyes review (§3) still happens for un-tagged work — its boundary is the
issue close itself, not a milestone close.

Two base-layer doc touch-points (both propagate via `base.manifest`):
- **AGENTS.md §3** — rewrite the "Single-pass work → one milestone" line.
- **cmd/sdlc/helptext/close.md** — note next to the milestone-row guard that a
  plain-checkbox Plan has no `Mx` rows, so the Review-Verdict requirement
  doesn't apply and the issue closes in one gesture.

No code change — the tool behavior is correct; the guidance was steering wrong.

## Done when

- AGENTS.md §3 no longer tells you to tag single-pass work `M1`; it says
  plain checkboxes + one `close`, and explains why (`Mx` = a separate
  milestone-close boundary with its own trailer + log line).
- §3 still makes clear un-tagged single-pass work gets the mandatory fresh-eyes
  review at issue close (review isn't milestone-gated).
- close.md help notes the plain-checkbox case (no `Mx` rows → no trailer
  requirement → single close).
- A reader following the updated guidance would have closed #63 with one log
  line.

## Plan

- [ ] Rewrite AGENTS.md §3 single-pass/milestone guidance + clarify review
  boundary for un-tagged work; add the plain-checkbox note to close.md help.
  Verify by re-reading: the guidance, followed literally, yields one close
  gesture for atomic work.

## Log

### 2026-06-02
