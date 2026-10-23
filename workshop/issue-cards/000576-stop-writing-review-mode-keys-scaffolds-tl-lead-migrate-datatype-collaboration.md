---
id: 000576
status: open
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
github_issue:
---

# Stop writing review mode keys: scaffolds, tl-lead migrate, datatype collaboration-mode rows

## Problem

ariadne#575 rewrites universal-review to one collaborative editing mode with a `@@chat@@` section carve-out. The old mode keys stay readable as aliases, and nothing new writes them. ariadne's own creators still write them:
- `sdlc issue new` and handoff (`cmd/sdlc/internal/issue/scaffold.go`, `handoff.go`) and the project scaffold (`cmd/sdlc/internal/project/scaffold.go`) write `collaboration-mode: <datatype default>` into every new issue and project.
- `tl-lead` migrate (`cmd/tl-lead/migrate.go`) writes `collaboration-mode: discussion` into side docs.
- The datatype definitions (`construct/datatype/*.md`) carry a `collaboration-mode` row in their Frontmatter shape table. `cmd/datatype/collabmode.go` parses that row, and `pkg/vocab` `CollaborationModes` backs it.
