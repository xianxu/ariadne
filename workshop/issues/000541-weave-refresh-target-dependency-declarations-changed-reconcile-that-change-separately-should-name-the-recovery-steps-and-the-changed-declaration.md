---
id: 000541
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'bc0563811226d1b8d0e57b596e470e16e84e882c' # card fields mirrored from issue-cards; edit via sdlc
---

# weave refresh: 'target dependency declarations changed; reconcile that change separately' should name the recovery steps and the changed declaration

## Problem

## Spec

From ops-echo:0 (10-21 evening): after echo#1 landed (it changed echo's `construct/deps` ariadne source from SSH to
HTTPS, operator-approved), Couch's prepare/refresh on echo:1 failed with
`…/echo-slot1/echo: target dependency declarations changed; reconcile that change separately`
(`cmd/weave/internal/refresh/refresh.go:157`). Nobody knew what "reconcile separately" meant; echo:0 got through by
fast-forwarding first. The refusal is right (refresh won't apply a dependency change), but the error must name the
step.

Fix: the error names the exact recovery for the common case, a clean resting branch behind origin/main:
"dependency declarations change between HEAD and origin/main (construct/deps …): fast-forward this checkout
(`git merge --ff-only origin/main`) and run ordinary setup (`./bootstrap.sh`), then retry", listing which
declarations changed (names and fields, never credential-bearing URLs). Document it in the setup atlas.

## Done when

- A fixture where origin/main changes a deps source yields the new message, naming the changed declaration and the
  steps (test).
- The atlas setup page documents the recovery.

## Plan

- [ ]

## Log

### 2026-10-21
