---
id: '000145'
status: done
started: 2026-07-05T15:01:33-07:00
created: 2026-06-30
updated: 2026-07-05
estimate_hours: 1.08
actual_hours: 0.72
---

# sdlc issue new: derive on-disk template from the issue.cue model

## Problem

`sdlc issue new`'s on-disk template is a **hardcoded Go renderer**
(`cmd/sdlc/internal/issue/scaffold.go:Render`), independent of the
`construct/vocabulary/issue.cue` model. Its own doc comment calls it "the single
source of truth for the on-disk template" — including literal `status: open` and
the `## Spec` / `## Done when` / `## Plan` / `## Log` section list, none of it
derived from cue.

So the issue noun has **two representations kept consistent by hand**:

| Concern | Source |
|---|---|
| Creation (template written to disk) | hardcoded Go — `scaffold.go:Render` |
| Validation (`sdlc issue validate` vs `#Issue`) | cue `#Issue` definition |
| Lifecycle / status (`CanTransition`, `AllStatuses`, `IsOpen`) | cue → `issue.json` → `pkg/vocab` |

The validate gate (`CheckStructural` at `sdlc change-code`) catches drift, but
creation and the cue model don't *share* a source. This blocks true
single-sourcing of the issue shape, and means downstream consumers that want the
issue structure from cue can't get the *creation* template that way (they must
either delegate to `sdlc issue new` or re-derive). Surfaced while designing
parley.nvim#116 (parley's ariadne-support discovery subsystem): parley sources
issue **status** from `construct/generated/vocabulary/issue.json` already, but
for **creation** it can only delegate to `sdlc issue new` because the template
isn't in the cue model.
