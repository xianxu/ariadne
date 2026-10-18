---
id: 000445
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# lint strict wire decodes in expression-form if initializers

## Problem

The syntax-only `wire-strict-decode` rule currently checks assignment-form `if` initializers, but skips a direct expression initializer such as `if d.Decode(&m); true {}`. It also misses `if d.DisallowUnknownFields(); true {}` before a later direct decode. These are missed findings under the conservative direct-chain rule discovered in ariadne#440's seventh close review (BR-6).
