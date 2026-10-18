---
id: 000447
status: done
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '4eebdac02f2991f9a7ddf95163536fe566b0be51' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T13:57:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.09
---

# review checkouts lack main ref and remote config, so repo tests needing main authority can't run in review

## Problem

The boundary review runs in a disposable checkout that lacks the main branch ref, its upstream and remote configuration. Repository tests that need main authority (pair's Go lint catalog finds tracked files and the merge base against main) correctly refuse there, so the reviewer cannot run them and raises an "unverifiable" finding. It hit pair#665 M1 (BR-4) and pair#666 (BR-1) on 2026-10-18; each slot had to capture host evidence and dispute it.

## Spec

Give the review checkout the same main authority a slot checkout has: the main ref at the reviewed base, and the remote and upstream configuration (read-only; the reviewer never pushes). Any test that runs in the slot then runs the same way in review.

## Done when

- A repository test that resolves main and the merge base passes inside the review checkout; a regression test reproduces the pair#665 shape.
- The review checkout still cannot push or change the real repository.

## Plan

**Cause (measured):** `reviewbox.populate` runs `git clone --shared --no-checkout SRC` and then `git remote remove origin`. Removing the remote also deletes every `refs/remotes/origin/*` ref and the `remote.origin.*` config. The `branch.main.*` keys a slot carries (`remote origin`, `merge refs/heads/main`) never reach the clone either. pair's repolint (`mainRefWithRepository`) reads exactly those: `branch.main.remote` and `branch.main.merge` lead to `refs/remotes/<R>/main`, and failing that, the one `refs/remotes/*/main`. With neither, it refuses. `git rev-parse main@{upstream}` also needs `remote.<R>.fetch` to map `merge` to the remote ref.

**Design: copy the slot's main authority, not its remote.** After the clone, `populate` reads, from SRC:
- the `branch.main.remote`/`branch.main.merge` config;
- each tracked remote's `remote.<R>.fetch`;
- every `refs/remotes/*/main` ref and the local `refs/heads/main`, with their SHAs.

It writes the same refs and the same keys into the checkout. The objects are already there through the clone's alternates.

`remote.<R>.url` and `remote.<R>.pushurl` are set to an inert `readonly://review-checkout` URL, so no remote can push or fetch: git refuses with "unable to find remote helper", and nothing in the checkout points back at SRC or GitHub. SRC's real URL, credentials, `insteadOf` and push settings are never copied. A SRC with no main authority gives a checkout with none, so a test refuses in review exactly as it would in the slot.

- [x] **Pure:** `mainAuthority(config, refs) authority` in `reviewbox/authority.go` parses SRC's `git config --get-regexp` lines and the `for-each-ref` rows into the keys and refs to write. Only the allowed keys pass: `branch.main.{remote,merge}` and `remote.<R>.fetch` for a remote named by branch.main or by a `refs/remotes/<R>/main` ref. A malformed row refuses. Tests cover a slot shape, no authority, two remotes, a `.` remote, and a hostile key (`remote.origin.url`, `url.*.insteadOf`) never passing.
- [x] **IO:** `populate` reads SRC (two git calls), applies the authority with `git config` and `git update-ref`, and sets the inert URLs. Reset reuses populate, so a retried attempt gets the same authority.
- [x] **Regression, the pair#665 shape** (real git, temp repos, as the existing checkout tests do): SRC is a clone of a bare `origin` whose main has moved, on an issue branch. In the checkout:
  - pair's three facts resolve: `branch.main.*` is configured, `refs/remotes/origin/main` equals SRC's, and `git merge-base --all HEAD refs/remotes/origin/main` gives one base equal to SRC's;
  - `git rev-parse main@{upstream}` works;
  - `git push origin HEAD:refs/heads/x` and `git fetch origin` both fail, and SRC's and the bare origin's refs are unchanged after them;
  - a SRC with no remote gives a checkout with no main authority.
- [x] Atlas (review isolation) says the checkout carries the slot's main authority, read-only.

Design notes:
- ARCH-SECURE: SRC's config is our own but copied through an allowlist, never wholesale. The checkout's remote URL is inert, so the "cannot push or change the real repository" guarantee is structural, not a convention. The fence and confinement are unchanged.
- ARCH-FUNERAL: creates nothing durable beyond refs and config inside the checkout, which Close/Sweep already remove.
- ARCH-ORDER: holds no state between events; the copy happens once per populate (create or reset).
- ARCH-CONSTRAINTS: two extra git reads and a handful of writes per review round (milliseconds), next to a clone that is already there.
- ARCH-MOCK: real git in temp repos is the existing seam for reviewbox; no network.
- Size: about 60 added code lines, quick flow.

## Log

### 2026-10-18
- 2026-10-18: closed — Close Minors fixed: remoteMain accepts only a one-segment remote name (a branch feature/main under origin adds no remote; parser row); the config read is --local (TestCheckoutCopiesOnlyTheSlotsConfig with GIT_CONFIG_GLOBAL, which fails without --local). reviewbox tests pass; suite green on 12b3eb3f (8m39s). (fixed-then-ship); review verdict: SHIP
- 2026-10-18: closed — Suite green on 126eb88b (sdlc-local/suite, 8m27s). reviewbox: TestMainAuthority (allowlist; URL, insteadOf and pushurl refused; malformed rows refuse), TestCheckoutCarriesMainAuthorityReadOnly in the pair#665 shape (branch.main config, origin/main, a single merge base and main@{upstream} equal the slot; push and fetch fail against the inert readonly URL; slot snapshot and bare origin refs unchanged), which fails with copyMainAuthority disabled; TestCheckoutWithoutRemoteHasNoRemoteMain. cmd/sdlc isolation tests pass. Atlas sdlc-binary updated.; review verdict: SHIP
- 2026-10-18: flow upgraded quick → full — 143 added lines in code files (limit 100)
- 2026-10-18: the hub approved the plan. Implemented as planned: `reviewbox/authority.go` (`mainAuthority`, `authorityKeys`, `inertURL`) and `Checkout.copyMainAuthority`, called by `populate` after the detached checkout, so Reset gets it too.
- Verified: `go test ./cmd/sdlc/internal/reviewbox` passes. That covers `TestMainAuthority` (nine rows, hostile keys refused), `TestCheckoutCarriesMainAuthorityReadOnly` (the pair#665 shape: config, origin/main, a single merge base and main@{upstream} all equal the slot's; push and fetch fail; slot and bare origin unchanged) and `TestCheckoutWithoutRemoteHasNoRemoteMain`. The regression fails with the copy disabled (`config --get-regexp ^branch.main.` exits 1) and passes with it restored. The cmd/sdlc isolation tests pass.
- Close review (SHIP, two Minors, fixed in the same round): only a one-segment `refs/remotes/<R>/main` names a remote (`remoteMain`), so `origin/feature/main` adds no stray remote; the config read is `--local`, so user and system config are not copied as the slot's (`TestCheckoutCopiesOnlyTheSlotsConfig`, plus a parser row).
