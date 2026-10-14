---
id: 000371
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '7e44a9c578ecc7c07c91bf0b7058bb557b2a4948' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc issue new --details FILE: file an issue with its body in one call (today: new, edit, publish)

## Problem

Filing an issue with content takes three steps: `sdlc issue new "title"` (publishes an empty template), edit the details file, `sdlc issue publish --issue N`. On 2026-10-14 the TL did this ~20 times; it twice collided with a pair:0 stage (the edit needs the resting checkout) and repeatedly waited on a repo lock between steps. Retro item (a) from couch-broadcast-web.

## Spec

`sdlc issue new "title" --details FILE` (and `--details -` for stdin): FILE holds the body sections (`## Problem`, `## Spec`, `## Done when`, optional Plan/Log), validated against the issue-file contract; the card and the full details publish in the one creation commit, behind the existing creation receipt. Works from any checkout without writing to it (so a staged or busy checkout is never touched). Partial bodies are fine; missing sections keep the template's.

## Done when

- `issue new --details FILE` creates and publishes card plus body in one call; the file never appears dirty in the invoking checkout (test).
- Invalid section structure is refused before reserving an ID (test).

## Plan

- [ ]

## Log

### 2026-10-14
