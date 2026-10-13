---
id: 000347
status: open
deps: [ariadne#337]
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '421470c65215d9078906013a359c8a4349ea58da' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T13:06:13-07:00
---

# Cross-repo issue migration carries asset bundles after deferred issue renumbering support

## Problem

The asset lifecycle in #337 deliberately covers one repository. Cross-repo
`sdlc migrate` currently refuses issue-family artifacts because allocating a
destination ID and reconciling the trackers was deferred by #179. When that
capability exists, moving only issue prose would orphan its evidence.

## Spec

After #337 lands and the issue-migration design deferred by #179 is resolved,
carry the issue body and its indexed asset bundle through the same cross-repo
transaction. Re-key asset paths for the destination ID, rewrite links, enforce
the destination's asset policy before copying, and preserve exact bytes.
An uncertain or interrupted destination publication must not authorize source
deletion. Design this together with tracker identity/renumbering and recovery;
do not bypass the existing safe refusal by manually copying an issue.

## Done when

- Cross-repo issue migration exists with explicit destination ID allocation and tracker ownership semantics.
- Referenced assets move with the issue, links resolve at the destination, and destination policy is enforced before effects.
- Retry, interrupted publication, destination conflicts and refused policy preserve source evidence and never leave a published dangling link.

## Plan

- [ ]

## Log

### 2026-10-13

Filed at ops:0's direction as the explicit migration scope cut from #337.
Depends on #337; also requires resolving #179's deferred issue-family migration.
Left open and unassigned; no design or implementation started.
