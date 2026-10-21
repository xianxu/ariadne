---
id: 000512
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '0af13f2978dc6d694ed67a9bf5a4f188b516e0ce' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T00:14:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Reproduce two-car cleanup after the sibling archives first, without a transient train outcome; cover already-merged retry and publication interleaving.
- [ ] Route all archive checks through the existing exact tracked-archive proof when per-PR provenance is absent; preserve invalid/ambiguous provenance and mismatched completion refusals.
- [ ] Run targeted regression/proof controls and the recorded full suite, then close and land.

Actions: retry `sdlc merge --yes`.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Landed with exact matching durable archive | Cleanup completes on the existing resting branch | Retry merge | None; completion ownership stays unchanged |
| Active details missing without matching durable proof | Refusal, checkout preserved | Retry merge after evidence is reconciled | None; no speculative cleanup |
| Archive published between cleanup checks | Same proven cleanup | Retry merge | None; each check uses its pinned main snapshot |

## Log

### 2026-10-20

- Reproduced again in pair:4 after pair#733 and pair#736 landed together in train208, train PR504 (component PR497). GitHub confirms MERGED at 2026-10-21T05:38:30Z, merge commit `8db1a277ecfe534753530ab80c1ed8d0a028e5fb`. The combined suite passed in 11m2s (log SHA-256 prefix `17c1eb7e57ca`), followed by passing final-head conformance and all CI checks. After waiting for another merge's repo lock, the conductor refused `active issue missing without confirmed archive` for pair#733. The documented `sdlc merge --branch … --yes` retry returned the same refusal. `sdlc issue show 733 --json` independently reports status done, landing outcome landed, the matching landed commit, and `workshop/history/issues/000733-couch-notifications-are-ephemeral-a-restart-starts-with-an-empty-list-dismiss-and-clear-all-remove-items-old-items-age-out.md`; tracker reconcile reports no unfinished operations. The clean checkout remained on issue head `787ebf42ff3388df7041eda7720e7ced5c8bac7a`, with no commits ahead of main and the active details still present locally. Per TL instruction, preserved that landed branch and switched to the existing main-slot4 resting branch. No archive proof, card fields, or completion ownership were edited.

### 2026-10-21 — design

- Recovered the dispatched issue after #518 landed. Source tracing found that `landing.go` permits generic tracked-archive proof only after a transient waiter outcome. Conductor cleanup and a fresh already-merged retry cannot enter it. `archiveLandingPR` also rechecks refreshed main without that fallback, exposing the same refusal if a sibling archives between checks.
- ARCH-DRY / ARCH-ORDER: put proof selection in the shared archive checker, using the same pinned main snapshot and selected completion generation. Reuse the existing exact archive transition verifier; a notification is not durable authority. No new state or recovery operation.
- Quick plan posted to the hub, proceeding unless objected. This is semantic evidence validation rather than a mechanically lintable API misuse; real-Git ordering and negative-proof regressions are the guard.
