---
id: 000408
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '4e779e543ce434dec6a229294bb6df08705b3ecc' # card fields mirrored from issue-cards; edit via sdlc
---

# issue publish of an archived issue creates an active twin under workshop/issues (pair#579)

## Problem

2026-10-16: pair:3 ran `sdlc issue publish --issue 579` to add a Log entry to pair#579, which is done and archived
(`workshop/history/issues/000579-*.md`). With the default issues dir, publish created a NEW active twin
`workshop/issues/000579-*.md` (47 lines) on pair main (commit 7af5c7b1). Re-running with
`--issues-dir workshop/history/issues` published correctly to the archive. No verb removes the stray active copy.

## Spec

- `issue publish` resolves the issue's details file wherever it lives (active or archived), the same way
  `issue show` does, and refuses to create an active file for an issue whose card is done or archived.
- If both an active and an archived file exist for one issue, publish refuses and names both.

## Done when

- Tests: publishing an archived issue updates the archive file and creates nothing under `workshop/issues/`; a
  done-card issue with no details refuses; the twin case refuses.

## Plan

- [ ]

## Log

### 2026-10-16
