---
id: '000250'
status: done
started: 2026-09-24T20:15:40-07:00
created: 2026-09-24
updated: 2026-09-24
actual_hours: 0.09
---

# Seeded merge-check builds weave from source while the tap is unpublished

## Problem

The seeded `.github/workflows/merge-check.yml` (manifest `seed`) hands consumers
to `bootstrap.sh`, which runs `brew install xianxu/ariadne/weave`. The tap repo
`xianxu/homebrew-ariadne` is not published yet (#241), so every consumer that has
committed the post-#239 seed fails `merge-check` before any check runs:
`fatal: could not read Username for 'https://github.com'` (Homebrew cloning a
repo that does not exist). pair is the first: pair#323, every PR since
2026-09-24T02:43Z. The other consumers (parley.nvim, tools, …) still carry the
pre-#239 workflow and will break the same way on their next re-seed. ariadne's
own CI is unaffected because it builds a candidate `weave` from source.
