---
id: 000445
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b0e311d39e5724f4834d29dba8e37140ea0366e8' # card fields mirrored from issue-cards; edit via sdlc
---

# lint strict wire decodes in expression-form if initializers

## Problem

The syntax-only `wire-strict-decode` rule currently checks assignment-form `if` initializers, but skips a direct expression initializer such as `if d.Decode(&m); true {}`. It also misses `if d.DisallowUnknownFields(); true {}` before a later direct decode. These are missed findings under the conservative direct-chain rule discovered in ariadne#440's seventh close review (BR-6).

## Spec

Extend the supported direct-call positions to expression-form `if` initializers without inferring calls nested inside other expressions or branches. Route supported initializer statements through one shared ordered handler so strictness and Decode use the same provenance rules as top-level statements.

## Done when

- Red fixtures show the missed Decode and DisallowUnknownFields expression-initializer cases against the prior matcher, with assignment-initializer and strictness-order controls.
- The rule flags a proven direct strict decode in either initializer form, stays silent before strictness or when decoder ownership is ambiguous, and documents the expanded position.
- Focused tests and the recorded full suite pass before close.

## Plan

- Add the regression matrix, then use a shared direct-initializer handler in `pkg/golint/wire.go`.
- Update the lint's coverage documentation and verify through the SDLC gates.

## Log

- Filed from ariadne#440 BR-6 after the hub fixed that issue's scope at assignment-form `if` initializers; this is the follow-up for expression-form coverage.
