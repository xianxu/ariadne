---
id: 000455
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'decd1217108ace0eb24b7e0221771ca71910ed89' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: operator-instruction answers become side docs of the session file (and migrate qa/ answers)

## Problem

The hub answers each Operator instructions bullet in a separate doc under `teams/<team>/qa/<operator>/`, linked from the bullet with `tl-lead status instructions link`. parley's attention finder lists an entry's children by scanning its `.side/` folder for `type: side-doc` files, so these answers never appear in the finder (operator, 2026-10-18; ops qa 2026-10-18-24-finder-answer-docs.md). The two conventions never met. The operator: "there might even be a task to migrate from qa/ to xxx.side/ folder for TL's session file?"; none existed.

## Spec

Answers to the session file become real side docs of it: `<session-file-stem>.side/NN-slug.md`, with frontmatter `type: side-doc`, `parent:` (relative path to the session file), `marker:`, `collaboration-mode`, `process-instruction`, and the backlink line under the title, the same contract parley's side docs use. `tl-lead status instructions link` accepts and validates that target; tl-lead's help and the TL skill say answers go there. A one-time migration command moves existing `qa/<operator>/` answer docs into the side folder: it rewrites their frontmatter, rewrites each `→ [answer](…)` link and the `discussions:` index in the session file, keeps side-doc subfolders (`*.side/`) attached, and moves registered files through `couch --session-sync rename`, so Couch tracking survives. It is dry-run first and refuses on any conflict. Gitignore follows wherever qa/ was ignored.

## Done when

- `link` produces side-doc answers; a new answer shows in parley's attention finder under the session file (live check).
- The migration moves the ops answers with every link and the discussions index intact, and Couch still tracks the registered ones (dry-run, then live).

## Plan

- [ ]

## Log

### 2026-10-18
