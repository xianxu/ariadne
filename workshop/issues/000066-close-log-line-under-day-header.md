---
id: 000066
status: done
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 0.5
actual_hours: 0.75
---

# sdlc close: file the dated log line under its matching ### <date> day header, not orphaned above it

## Problem

`sdlc close` appends its `- <date>: closed — …` log line at the **top of the
`## Log` section** (`insertLogLine`, close.go:175). But issue templates seed a
`### <date>` day header right under `## Log`, and agents log their narrative
*under* that header. So the close line lands **above** the same-date header:

```
## Log

- 2026-06-02: closed — …      ← orphaned above
### 2026-06-02
- Filed …
- Implemented …
```

Both pre-merge plan-completeness judges (on #63 and #65) flagged this as a
cosmetic ordering nit. The dated close line reads as out-of-place sitting above
a header carrying the very same date.

## Spec

When the log line carries a leading date (`- YYYY-MM-DD: …`) and the `## Log`
section already contains a matching `### YYYY-MM-DD` day header, insert the line
**immediately beneath that header** (top of the day's group) instead of at the
top of the section. Otherwise keep the existing behavior byte-for-byte (top of
`## Log`, or create the section if absent).

Inserting at the top of the day's group (not the bottom) preserves the existing
**newest-first** convention `insertLogLine` already uses at the section level
(see `TestInsertLogLine_ExistingSection`: a newer entry is prepended above an
older one). Restrict the day-header search to *after* the `## Log` header so a
stray `### <date>` under some other section (e.g. `## Revisions`) can't capture
it. Match the day header with `[ \t]*$` (not `\s*$`) so the regex doesn't eat
the trailing newline and blank lines.

No behavior change when there's no matching day header — the three existing
`insertLogLine` tests must stay green unchanged.

## Done when

- A close on an issue whose `## Log` has a `### <today>` header files the line
  directly under that header, not above it.
- The three existing `insertLogLine` tests pass unchanged (fallback path is
  byte-for-byte identical).
- New test covers the day-header case (line lands under `### <date>`).

## Plan

- [x] Rework `insertLogLine` to prefer the matching `### <date>` day header
  (date parsed from the log line; search scoped after `## Log`; `[ \t]*$`
  header match), falling back to the current top-of-section insert. Add a test
  for the day-header case; confirm the existing three stay green.

## Log

### 2026-06-02
- 2026-06-02: closed — log-line lands under matching ### <date> in the real (last) ## Log section; 6 insertLogLine tests green (3 orig unchanged + 3 new incl. dogfound prose-false-match); review SOUND; --force = pure bugfix, no atlas surface

- Reworked `insertLogLine` (close.go): added `logLineDateRE` to parse the leading
  date off the log line; if a matching `### <date>` header exists in the Log
  section (searched after `## Log`, matched strictly with `[ \t]*$`), insert the
  line right under it; else the original top-of-section path, byte-for-byte.
- Honored plan-quality INFO: comment notes the `[ \t]*$` match is intentionally
  strict (don't loosen to `.*$` or the newline-eating bug returns).
- Tests: added `TestInsertLogLine_UnderMatchingDayHeader` (lands under header)
  and `TestInsertLogLine_DayHeaderDateMismatch_FallsBack` (06-02 line not
  misfiled under a 05-25 header). The three pre-existing `insertLogLine` tests
  pass **unchanged** (fallback path identical). `go test ./cmd/sdlc/...` + `go
  vet` green.
- Note: `sdlc` is a shell function that rebuilds `bin/sdlc` from `cmd/sdlc` each
  call, so closing this issue runs the fixed code — a live dogfood.
- **Dogfood caught a deeper pre-existing bug.** The first close attempt filed the
  verification line into this issue's OWN `## Problem` code-block example, not the
  real `## Log` section — because `insertLogLine` matched the **first** `## Log` /
  `### <date>` in the body, and this meta-issue literally quotes both inside a
  fenced block. Both the old and new code shared this "first-match-wins, even in
  prose" weakness; #66's self-referential body just exposed it.
  - **Fix (folded in, same function/concern):** anchor on the **last** `## Log`
    header (the real Log section is conventionally the final `##` section); take
    all offsets relative to it so both the day-header and fallback inserts target
    the real section. Single-`## Log` bodies are unchanged (last == first).
  - Added `TestInsertLogLine_IgnoresEarlierLogHeaderInProse` (quoted `## Log` in
    an earlier fenced block left untouched; line lands in the real section).
  - Restored the misfiled issue file to its pre-close commit and re-closed with
    the hardened binary (clean dogfood).
