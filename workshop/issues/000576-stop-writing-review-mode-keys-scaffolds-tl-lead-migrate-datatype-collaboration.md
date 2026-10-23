---
id: 000576
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4b5d05fdc51d25c4c95a808ddbb16166ee1c854b' # card fields mirrored from issue-cards; edit via sdlc
---

# Stop writing review mode keys: scaffolds, tl-lead migrate, datatype collaboration-mode rows

## Problem

ariadne#575 rewrites universal-review to one collaborative editing mode with a `@@chat@@` section carve-out. The old mode keys stay readable as aliases, and nothing new writes them. ariadne's own creators still write them:
- `sdlc issue new` and handoff (`cmd/sdlc/internal/issue/scaffold.go`, `handoff.go`) and the project scaffold (`cmd/sdlc/internal/project/scaffold.go`) write `collaboration-mode: <datatype default>` into every new issue and project.
- `tl-lead` migrate (`cmd/tl-lead/migrate.go`) writes `collaboration-mode: discussion` into side docs.
- The datatype definitions (`construct/datatype/*.md`) carry a `collaboration-mode` row in their Frontmatter shape table. `cmd/datatype/collabmode.go` parses that row, and `pkg/vocab` `CollaborationModes` backs it.

## Spec

Remove every ariadne writer of the mode keys (`collaboration-mode`, `presentation-mode`/`presentation`, `processing-mode`, `processing-instruction`, `process-instruction`), as well as the datatype rows and the vocab list that exist only to feed them. Old documents must stay readable. Issue, project and side-doc files that already carry the keys still parse and round-trip unchanged: no reader refuses a document because a mode key is present, and no rewrite drops one. The `@@chat@@` carve-out is the only routing signal new documents carry. Land right after #575.

## Done when

- New issues, projects (`sdlc issue new`, handoff, the project scaffold) and tl-lead side docs carry no mode keys. Tests assert their absence.
- The datatype tables have no `collaboration-mode` row, and nothing in Go reads one; `vocab.CollaborationModes` is gone or reduced to a legacy alias reader.
- Fixtures with legacy keys (an issue with `collaboration-mode: collaboration`, a side doc with all three keys) still parse, and survive every rewrite verb byte for byte.

## Plan

- [ ]

## Log

### 2026-10-23
