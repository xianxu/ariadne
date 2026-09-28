---
id: '000066'
status: done
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
