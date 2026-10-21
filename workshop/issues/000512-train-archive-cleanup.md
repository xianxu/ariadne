---
id: 000512
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd3debca151e6c7d37cf9d3fb7f4bd87c36442c4d' # card fields mirrored from issue-cards; edit via sdlc
---

# Train-car cleanup refuses confirmed archived issue after landing

## Problem

After pair train197 landed both cars, the conductor for pair#727 failed local cleanup with: active issue missing without confirmed archive: workshop/issues/000727-term-wrapper-one-unsafe-cell-kills-the-presenter-unsafe-cell-12-and-a-0x0-launch-geometry-exits-the-operator-s-right-pane-died.md. A documented sdlc merge --yes recovery returned the same error.

PR486 and train PR495 are MERGED. Git verified car61323a50502774fae0cad086fab45000249e3978 and train mergee7727b341be4bcc53acfdf85a572266b02d95f13 are ancestors of origin/main. sdlc issue show727 --json confirms status done, landing outcome landed, the same landed_commit and archived path under workshop/history/issues. origin/main contains the archive and all archived plan/review files. The clean slot remained on its old issue head e0686637455999591b6b8c6028bb7f53c04acf9e, where the active issue still exists. No local dirt or Git operation remained.

The conductor had waited for another slot's sdlc merge repo lock before the refusal. Causality is unproven. No archive proof/card was edited. Worker preserved the landed issue branch, switched to its existing resting main-slot7 and fast-forwarded to verified main so its next dispatched item can proceed.

## Spec

Resume post-train cleanup from confirmed durable landing/archive evidence, including when another car archived first. Keep the missing-without-proof refusal for genuinely unconfirmed archives. Do not repeat server merge or rewrite completion ownership to recover.

## Done when

- A deterministic two-car cleanup-order regression reproduces and fixes the refusal.
- A confirmed landed/archive card plus matching main archive allows cleanup to the existing resting branch.
- A missing active issue without matching durable proof still refuses; another car/issue cannot satisfy the proof.

## Plan

- [ ]

## Log

### 2026-10-20
