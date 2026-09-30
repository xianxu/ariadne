---
id: 000273
status: open
deps: []
github_issue:
created: 2026-09-29
updated: 2026-09-29
estimate_hours:
card_mirror: '99bd10bbc669a0e44f2bd1081d2025ac2986d2cf' # card fields mirrored from issue-cards; edit via sdlc
---

# Rename issue move-detail to publish-detail

## Problem

`sdlc issue move-detail` names its original, narrower mechanism: moving
details created on a feature branch to main. Its purpose is to publish an
issue's initial details to main so the issue becomes claimable, including
from a resting slot, where nothing "moves". Its help also frames it as an
escape hatch to consult the operator about, yet filing an issue in a free
slot and publishing its details is the routine path (e.g. pair#353's
cross-slot dispatch, where the sender files in the target repo's free slot).

## Spec

- Rename the verb to `sdlc issue publish-detail`; keep `move-detail` as a
  deprecated alias that works and points at the new name (downstream repos
  and agent memories carry the old one; ariadne is the base layer).
- Reframe the help: publishing initial details is the normal way to make an
  issue filed off-branch claimable; keep the refusals (details already on
  main, merge in progress, source changed) unchanged.
- Update refusal/next-action messages, atlas pages, and skills that name it.

## Done when

- `sdlc issue publish-detail` does what `move-detail` did; `move-detail`
  still works and prints a deprecation note; tests cover both names.
- No message, help text, atlas page or skill still recommends `move-detail`.

## Plan

## Log

### 2026-09-29

Filed from ariadne#272's brainstorm at the operator's request.
