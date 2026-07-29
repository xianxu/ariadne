---
id: 000190
status: working
deps: []
github_issue:
created: 2026-07-29
updated: 2026-07-29
estimate_hours:
started: 2026-07-29T16:23:09-07:00
---

# active-time mention-fallback attributes replay work to the replayed issue

## Problem

The active-time engine's mention-fallback attributes work to whatever issue the session
TALKS about, so replaying or postmortem-ing issue X charges the time to X rather than to the
issue doing the replaying.

Measured live at ariadne#187's close: **46.1m/77% attributed to pair#127** by
`mention fallback without issue commit boundary`, because #187's Task 14 replayed pair#127
and its commits, prose and evidence file cite `#127` constantly. #187 measured 2.29h; the
true figure is higher, and pair#127 — long closed — gained 46 minutes it did not spend.

Both directions are wrong in a way that matters, because these numbers feed velocity
calibration: the replaying issue looks cheaper than it was, and a CLOSED issue's actual
silently grows after the fact.

This is a general hazard, not a #187 quirk. Any issue whose work is *about* another issue —
replays, postmortems, migrations, "fix the thing #N introduced" — hits it. The engine
already knows the difference in principle: it warns `without issue commit boundary`, meaning
it had no commit anchoring the segment and fell back to text mentions.

## Spec

- An issue-referencing **commit boundary** should outrank a text mention within the same
  window. The warning already identifies the weak case; the fix is to let the strong signal
  win rather than blending them.
- Attribution to an issue whose status is terminal (`done`/`wontfix`) is the loudest signal
  that a mention fallback is wrong — a closed issue is not accruing work. Treat it as at
  minimum a refusal to attribute, or a hard warning.
- **Do not** solve this by asking the operator to hand-correct hours: a typed actual is
  exactly what the close gate exists to prevent (#178). The engine must attribute correctly,
  or say it cannot.
- Whatever lands must be checkable against #187's window, where the right answer is known:
  the 46.1m belongs to #187.

## Done when

- A window whose commits all anchor `#A` while the prose mentions `#B` attributes to `#A`.
  Commit boundaries outrank mention fallback; the engine already distinguishes them in its
  warning text.
- Mention-fallback attribution to a **terminal-status** issue (`done`/`wontfix`) refuses or
  warns loudly — a closed issue is not accruing work, and silently growing its actual after
  the fact corrupts calibration history.
- **Regression check against a known answer:** re-measuring ariadne#187's real window
  returns the 46.1m currently charged to pair#127 back to #187.
- `sdlc actual` output states which rule attributed each segment, so a wrong number is
  diagnosable without reading the engine.
- No operator-facing workaround is introduced: hand-typing a corrected actual stays exactly
  as forbidden as it is today (#178).
- `atlas/workflow/` documents the precedence rule, since attribution now has one.

## Plan

- [ ] Failing test: a window whose commits all anchor `#A` while the prose mentions `#B`
      attributes to `#A`
- [ ] Prefer commit-boundary attribution over mention fallback when both are present
- [ ] Refuse (or loudly warn) on mention-fallback attribution to a terminal-status issue
- [ ] Regression check against #187's real window: the pair#127 46.1m returns to #187
- [ ] Note in `atlas/workflow/` how attribution resolves, since it now has a precedence rule

## Log

### 2026-07-29
