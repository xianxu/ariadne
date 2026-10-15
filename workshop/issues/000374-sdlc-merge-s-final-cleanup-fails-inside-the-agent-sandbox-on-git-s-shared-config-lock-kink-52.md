---
id: 000374
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '10d1c7097fe7a344245ab8abadaa61b1fba88308' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T23:06:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc merge's final cleanup fails inside the agent sandbox on git's shared config lock (kink 52)

## Problem

`sdlc merge`'s final cleanup (after landing) fails inside the agent's sandbox on git's shared config lock (`.git/config.lock`), seen 3+ times (TL kink 52); slots rerun the printed recovery command unsandboxed. Retro item (e); not covered by ariadne#370, which landed first.

## Spec

Find which cleanup step writes git's shared config (branch deletion with upstream config, worktree prune, `git config` edits) and either avoid the shared-config write in the sandboxed path (e.g. no branch config to remove, or defer it to the next resting-branch command), or make the step idempotent and resumable so the next sdlc run completes it without the operator. No sandbox widening.

## Done when

- A full `sdlc merge --yes` inside the default agent sandbox completes cleanup without a manual rerun (live check on a scratch issue), and an interrupted cleanup completes on the next run (test).

## Plan

- [ ] Keep ref deletion and identity/occupancy/CAS guards mandatory; make only obsolete branch-config removal a non-blocking deferred cleanup, with the affected config path and retry owner reported.
- [ ] Reuse the existing done-card merge/reconcile sweep to retry configuration cleanup for absent, unoccupied landed issue branches; preserve live/recreated refs and unrelated config (ARCH-DRY, ARCH-FUNERAL).
- [ ] Add real-Git regressions for repeated config-lock denial, interruption before/after ref deletion, automatic later cleanup when writable, and live/recreated branch protection. Exercise the full merge caller with the existing stateful GitHub fake, plus a native denied-config sandbox check.
- [ ] Update help/recovery/atlas for deferred metadata cleanup, run focused and recorded full-suite verification, close and land on SHIP.

## Log

### 2026-10-14

### 2026-10-14 — root cause and proposed design

Claimed by ariadne:5 and started from current main. The durable path explicitly runs git config --local --remove-section branch.<issue> after CAS update-ref deletion (cmd/sdlc/landing.go:321-348). The command acquires the Git common directory's config.lock, even from a linked worktree. A native macOS sandbox scratch experiment denied writes only to shared config/config.lock: update-ref exited 0; remove-section exited 255 with could-not-lock-config; the branch was absent and its upstream config persisted. Current recovery repeats the same permanently denied metadata write. Existing interruption tests cover one transient failure, not permanent denial. Legacy merge uses git branch -D, whose config cleanup is already warning-only; its caller returns success rather than this fatal result.

Recommended design: distinguish mandatory branch retirement from ancillary obsolete-upstream metadata. A failed metadata write warns without invalidating a confirmed landing. The existing done-card sweep on merge/reconcile retries only obsolete configuration belonging to done issues with no local ref or checkout. Cards plus existing config entries supply retry discovery, so no new receipts, queue, or sandbox privileges are needed. A persistent sandbox denial retains only existing metadata (typically the remote/merge pair per landed issue); the next writable merge/reconcile removes it automatically. Ref CAS, occupancy, identity checks and operational failures in critical cleanup remain blocking. The thin Git IO helper stays behind the existing runner (ARCH-PURE); tests use real scratch Git and the stateful GitHub fake (ARCH-MOCK).

Alternatives considered: making all final cleanup fatal preserves the present retry loop under a persistent sandbox restriction; permanently dropping configuration cleanup accumulates residue with no owner. Writing the protected shared config by another path is neither safe nor a permitted sandbox workaround. The proposed deferred sweep supplies an explicit lifecycle for the residue (ARCH-FUNERAL).

The original live acceptance asks for a scratch issue in the default agent sandbox. The full merge regression will cover a scratch issue with real Git and the production merge path plus stateful fake GitHub, including a real denied-config sandbox. Ask the hub how to arrange the additional default-Claude sandbox live landing without overlapping another slot's active issue; do not silently replace that acceptance with a simulation.
