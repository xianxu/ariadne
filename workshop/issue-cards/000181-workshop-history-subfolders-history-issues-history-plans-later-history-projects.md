---
id: '000181'
status: done
started: 2026-07-15T16:23:08-07:00
created: 2026-07-15
updated: 2026-07-15
estimate_hours: 0.94
actual_hours: 0.38
---

# workshop/history subfolders: history/issues, history/plans, later history/projects

## Problem

`workshop/history/` is FLAT: 258 files mixing archived issues, their plan
docs, and review sidecars in one directory. It doesn't scale visually or
navigationally, and #180 is about to add a third artifact kind (archived
projects, if archive-on-done is chosen there). The archive layout is
consumed by real machinery, so restructuring is a code change, not a mkdir:

- the merge/push archive step moves the id-keyed family
  (`issues|plans/ → history/`, #160) as one flat destination;
- `sdlc resolve`'s `familyFiles` globs the archive dir from the vocab model
  (`issue.cue` `discovery: archive: "workshop/history"`) — the single
  source consumers derive from;
- the #163 filename-grammar helper and any history-scanning tooling assume
  the flat layout.
