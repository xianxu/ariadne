---
id: 000551
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '31dde955af6a4f109be46c791e3593249d1d9091' # card fields mirrored from issue-cards; edit via sdlc
---

# review box gives the reviewer no dependency checkouts; consumer reviews can't verify substrate claims (echo#9 BR-1 unverifiable)

## Problem

## Spec

From ops-echo:0 (10-22 ~05:30): echo#9's boundary review returned REWORK with one Critical, BR-1 "unverifiable",
because the reviewer couldn't check echo against ariadne's fixtures and templates: it looked for
`/Volumes/TestScratch/review-checkouts/ariadne`, which doesn't exist (`review-checkouts` is the review box's root for
the disposable clone of the repo **under review**, `cmd/sdlc/internal/reviewbox/checkout.go`). No product defect was
found. Every echo close (echo#9, #10, #3, #5) touches the ariadne-seeded workflow, so each would hit the same wall.

The review box gives the reviewer only the repo under review, never its declared dependencies, so a consumer repo's
review can't verify claims that depend on the substrate. Same class as ariadne#545 (suites couldn't find
dependencies).

Fix: the review box also provides read-only, disposable checkouts of the repo's `construct/deps` dependencies (the
substrate at the exact pinned revision or main, as the deps row says), placed beside the review clone, with their
paths told to the reviewer in the review prompt (mirroring #545's SDLC_DEP_* names). Same lifecycle as the review
clone (ARCH-FUNERAL), never writable, bounded size and time.

## Done when

- A review of a consumer repo whose deps name ariadne sees a read-only ariadne checkout at the documented path, and
  the prompt names it (test).
- Removed with the review clone; repos with no deps are unchanged (test).

## Plan

- [ ]

## Log

### 2026-10-22
