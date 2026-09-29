---
id: 000268
status: working
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '3086404002851105088a772279b7f1350bd76bbc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T20:57:50-07:00
flow: {kind: quick, provenance: inferred, spec: "8de1116f", done: "155eab07"}
---

# Bind hosted review producer effects to activation context

## Problem

Pair branch restoration (pair#341) needs the cooperative producer to preserve activation identity through delayed handoffs and Git effects. Current instructions commit on acknowledgment without checking whether the checkout changed.

## Spec

Extend the authoritative hosted xx-fix instructions with the Pair context envelope `{repo, branch, file, activation}`. Echo request context unchanged, preserve commit-body record encoding, and verify context plus the live checkout immediately before human/agent rounds and ship. On mismatch or failed checks, preserve artifacts and stop. Legacy unscoped traffic remains restricted to uninterrupted legacy activations. This is the producer half of the approved pair#341 plan; Pair owns executable transport tests. ARCH-DRY: change the shared skill, not its Pair symlink. ARCH-ORDER: acknowledgment is evidence for its original activation only.

## Done when

- Hosted instructions require the scoped handoff envelope and revalidation before every Git effect; stale or mismatched artifacts remain intact.
- Scenario probes cover late acknowledgments, human/ship context mismatch, and exact envelope echo; SDLC boundary review accepts the documentation change.

## Plan

- [x] Run baseline producer scenarios, update authoritative instructions, and rerun scenarios.
- [x] Verify the Pair skill resolves to this source; submit the completed change to SDLC close review.

## Log

### 2026-09-28

User approved pair#341 implementation including this linked peer change. Ariadne :0 was clean on main; other work is isolated in sibling slots.

Baseline fresh-context Codex probe with the old skill could not produce the envelope: “I cannot specify exactly how the supplied context must accompany this record without inventing the contract.” It conservatively blocked branch/activation mismatches. The updated-skill probe produced the exact envelope, preserved artifacts and refused agent/human/ship effects on mismatches, and allowed the initial human round with validated active-pane context before matching history exists. Probe outputs: `/tmp/ariadne-268-red.txt`, `/tmp/ariadne-268-green-accepted.txt`. Pair skill realpath resolves to this authoritative file; `git diff --check` passes. Updated the existing docflow atlas entry; executable transport/concurrency regressions remain in pair#341.
