---
id: 000440
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5eebd5d71d198ddc64fff20583207d990ed1af93' # card fields mirrored from issue-cards; edit via sdlc
---

# lint: a wire-message annotation and a rule against strict decoding into wire types

## Problem

Strict JSON decoding of cross-version peer or wire messages (`DisallowUnknownFields`) breaks mixed-version fleets (pair#594), but the same call is right for local files. A syntactic rule can't tell them apart, so #407 deferred it.

## Spec

## Done when

- An annotation convention marks wire message types (for example a `// wire:` comment), and a pkg/golint rule flags strict decoding into a marked type, with flag/no-flag tests citing pair#594.

## Plan

- [ ]

## Log

### 2026-10-18
