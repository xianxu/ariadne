---
id: '000073'
status: done
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 0.25
actual_hours: 0.25
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
