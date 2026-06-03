---
id: 000073
status: working
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 0.25
---

# insertLogLine date-header matcher misses suffixed headers (### DATE — note)

## Problem

#66 made `sdlc close` file the dated log line under its matching `### YYYY-MM-DD`
day header, but the matcher is too strict: `(?m)^### <date>[ \t]*$` only accepts
a **bare** date line. The established log convention routinely uses **suffixed**
day headers — e.g. `### 2026-05-30 — session summary`, `### 2026-06-02 — closeout`.
Those don't match, so insertLogLine falls back to top-of-`## Log` and the close
line orphans above the day headers — the exact cosmetic #66 set out to fix.

Found closing #49 (its `### 2026-06-02 — closeout` header → close line landed at
the top; tidied by hand). The #66 plan-quality judge even predicted this
("if a header carries trailing text … it won't match").

## Spec

Loosen the day-header matcher to anchor on the date **prefix**, allowing an
optional ` — suffix`: `(?m)^### <date>([ \t].*)?$`. This matches both the bare
date and `### <date> — anything`, while still rejecting `### <date>x` (no
separator) so a date can't accidentally prefix-match a longer token.

Safe re the #66 newline concern: that warning was about `\s*$` (where `\s`
matches `\n`); here the separator is `[ \t]` and `.` does NOT match newline in
Go RE2 (no `(?s)`), and `$` under `(?m)` matches before `\n` — so the match stays
within the single header line and the insert offset (end-of-header-line) is
unchanged. Update the doc comment (which currently says the match is
"intentionally strict") to describe the prefix+optional-suffix semantics.

## Done when

- A close on an issue whose latest day header is `### <today> — <suffix>` files
  the line directly under that header, not at the top of `## Log`.
- The existing insertLogLine tests stay green; a new test covers a suffixed
  day header.

## Plan

- [x] Loosen `dayRE` in `insertLogLine` (close.go) to `([ \t].*)?$`; update the
  doc comment; add `TestInsertLogLine_UnderSuffixedDayHeader`. Confirm the other
  insertLogLine tests stay green.

## Log

### 2026-06-02

- Changed `dayRE` in `insertLogLine` (close.go) from `[ \t]*$` to `([ \t].*)?$`
  (anchors on the date prefix, optional ` — suffix`); rewrote the doc comment to
  explain the prefix+suffix semantics and why it's still one-line-bounded.
- Added `TestInsertLogLine_UnderSuffixedDayHeader` (exact-string: line lands
  directly under `### 2026-05-25 — closeout`, above the existing bullet). The
  bare-header test (`TestInsertLogLine_UnderMatchingDayHeader`) and the other 5
  insertLogLine tests stay green — bare-date placement and the prose-false-match
  guard are unaffected. `go test ./cmd/sdlc/...` + `go vet` clean.
